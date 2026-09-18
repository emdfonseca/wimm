package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
	"github.com/xuuid/wimm/apps/wimm/internal/store/storetest"
)

// The schema version this migration lands on top of: 00005_transactions.
const beforeAccountOwnership = 5

func TestAccountOwnershipColumnsRoundTrip(t *testing.T) {
	url := storetest.Scratch(t, "account_ownership_columns")

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating up: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()

	if got := columns(t, ctx, pool, "accounts"); !contains(got, "left_out_at") || !contains(got, "household_name") {
		t.Fatalf("accounts columns = %v, want left_out_at and household_name", got)
	}

	if err := store.MigrateDownTo(url, beforeAccountOwnership); err != nil {
		t.Fatalf("migrating down: %v", err)
	}
	if got := columns(t, ctx, pool, "accounts"); contains(got, "left_out_at") || contains(got, "household_name") {
		t.Fatalf("accounts columns = %v, want neither left_out_at nor household_name after rollback", got)
	}
}

func columns(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) []string {
	t.Helper()
	rows, err := pool.Query(ctx, `
		select column_name from information_schema.columns
		where table_schema = 'public' and table_name = $1`, table)
	if err != nil {
		t.Fatalf("reading columns of %s: %v", table, err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scanning a column of %s: %v", table, err)
		}
		out = append(out, name)
	}
	return out
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// The backfill runs against orphans that predate the invariant, so it is
// asserted against orphans that predate it: the schema is taken back to
// before this migration, three accounts are left with no owner row — one of
// them a connection's only account — and the migration is asked to bring
// every one of them back, owned and left out.
func TestOrphanBackfillOwnsAndLeavesOutEveryOrphan(t *testing.T) {
	url := storetest.Scratch(t, "orphan_backfill")

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating up: %v", err)
	}
	if err := store.MigrateDownTo(url, beforeAccountOwnership); err != nil {
		t.Fatalf("rolling back to the schema before account ownership: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()

	var memberID, otherMemberID string
	if err := pool.QueryRow(ctx, `
		insert into members (email, first_name, last_name)
		values ('ada@example.com', 'Ada', 'Lovelace') returning id`).Scan(&memberID); err != nil {
		t.Fatalf("registering a member: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		insert into members (email, first_name, last_name)
		values ('grace@example.com', 'Grace', 'Hopper') returning id`).Scan(&otherMemberID); err != nil {
		t.Fatalf("registering a second member: %v", err)
	}

	var connID string
	if err := pool.QueryRow(ctx, `
		insert into bank_connections (gateway, bank_id, bank_name, connected_by, consent_expires_at)
		values ('enablebanking', 'PT:Montepio', 'Montepio', $1, now() + interval '90 days')
		returning id`, memberID).Scan(&connID); err != nil {
		t.Fatalf("recording a connection: %v", err)
	}

	// This connection's only account, and an orphan: nobody owns it.
	var lastAccountID string
	if err := pool.QueryRow(ctx, `
		insert into accounts (source, connection_id, bank_id, gateway_ref, name, currency)
		values ('gateway', $1, 'PT:Montepio', 'hash-only', 'Conta a Ordem', 'EUR')
		returning id`, connID).Scan(&lastAccountID); err != nil {
		t.Fatalf("recording the connection's only account: %v", err)
	}

	var connID2 string
	if err := pool.QueryRow(ctx, `
		insert into bank_connections (gateway, bank_id, bank_name, connected_by, consent_expires_at)
		values ('enablebanking', 'PT:Revolut', 'Revolut', $1, now() + interval '90 days')
		returning id`, otherMemberID).Scan(&connID2); err != nil {
		t.Fatalf("recording a second connection: %v", err)
	}

	// Two more orphans on a connection that also has an owned account, so the
	// backfill has to find every orphan rather than stopping at the first.
	var orphanTwo, orphanThree, owned string
	if err := pool.QueryRow(ctx, `
		insert into accounts (source, connection_id, bank_id, gateway_ref, name, currency)
		values ('gateway', $1, 'PT:Revolut', 'hash-two', 'Savings', 'EUR') returning id`,
		connID2).Scan(&orphanTwo); err != nil {
		t.Fatalf("recording an orphan: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		insert into accounts (source, connection_id, bank_id, gateway_ref, name, currency)
		values ('gateway', $1, 'PT:Revolut', 'hash-three', 'ISA', 'EUR') returning id`,
		connID2).Scan(&orphanThree); err != nil {
		t.Fatalf("recording an orphan: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		insert into accounts (source, connection_id, bank_id, gateway_ref, name, currency)
		values ('gateway', $1, 'PT:Revolut', 'hash-owned', 'Current', 'EUR') returning id`,
		connID2).Scan(&owned); err != nil {
		t.Fatalf("recording an owned account: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`insert into account_owners (account_id, member_id) values ($1, $2)`, owned, otherMemberID); err != nil {
		t.Fatalf("owning an account: %v", err)
	}

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating the ownership invariant in over existing orphans: %v", err)
	}

	for _, tc := range []struct {
		account string
		want    string
	}{
		{lastAccountID, memberID},
		{orphanTwo, otherMemberID},
		{orphanThree, otherMemberID},
	} {
		var owner string
		if err := pool.QueryRow(ctx,
			`select member_id from account_owners where account_id = $1`, tc.account).Scan(&owner); err != nil {
			t.Fatalf("reading the backfilled owner of %s: %v", tc.account, err)
		}
		if owner != tc.want {
			t.Errorf("account %s is owned by %s, want %s", tc.account, owner, tc.want)
		}

		var leftOut *time.Time
		if err := pool.QueryRow(ctx,
			`select left_out_at from accounts where id = $1`, tc.account).Scan(&leftOut); err != nil {
			t.Fatalf("reading left_out_at for %s: %v", tc.account, err)
		}
		if leftOut == nil {
			t.Errorf("account %s came back from the backfill without left_out_at set", tc.account)
		}
	}

	var leftOutOwned *time.Time
	if err := pool.QueryRow(ctx,
		`select left_out_at from accounts where id = $1`, owned).Scan(&leftOutOwned); err != nil {
		t.Fatalf("reading left_out_at for the already-owned account: %v", err)
	}
	if leftOutOwned != nil {
		t.Error("the backfill left out an account that already had an owner")
	}
}

// A zero-dated row predates the fallback, so it is asserted against a
// zero-dated row: the schema goes back to before this migration, a row is
// written the way old code wrote one with no booking date, and the migration
// is asked to recover it from the value date.
func TestBookingDateBackfillFallsBackToTheValueDate(t *testing.T) {
	url := storetest.Scratch(t, "booking_date_backfill")

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating up: %v", err)
	}
	if err := store.MigrateDownTo(url, beforeAccountOwnership); err != nil {
		t.Fatalf("rolling back to the schema before account ownership: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()

	var memberID string
	if err := pool.QueryRow(ctx, `
		insert into members (email, first_name, last_name)
		values ('ada@example.com', 'Ada', 'Lovelace') returning id`).Scan(&memberID); err != nil {
		t.Fatalf("registering a member: %v", err)
	}
	var connID string
	if err := pool.QueryRow(ctx, `
		insert into bank_connections (gateway, bank_id, bank_name, connected_by, consent_expires_at)
		values ('enablebanking', 'PT:Montepio', 'Montepio', $1, now() + interval '90 days')
		returning id`, memberID).Scan(&connID); err != nil {
		t.Fatalf("recording a connection: %v", err)
	}
	var accountID string
	if err := pool.QueryRow(ctx, `
		insert into accounts (source, connection_id, bank_id, gateway_ref, name, currency)
		values ('gateway', $1, 'PT:Montepio', 'hash-1', 'Conta a Ordem', 'EUR') returning id`,
		connID).Scan(&accountID); err != nil {
		t.Fatalf("recording an account: %v", err)
	}
	if _, err := pool.Exec(ctx, `insert into account_owners (account_id, member_id) values ($1, $2)`,
		accountID, memberID); err != nil {
		t.Fatalf("owning the account: %v", err)
	}

	var txID string
	if err := pool.QueryRow(ctx, `
		insert into transactions (account_id, status, dedup_key, amount_minor, currency, booking_date, value_date)
		values ($1, 'booked', 'ref-a', -1190, 'EUR', '0001-01-01', '2026-03-04')
		returning id`, accountID).Scan(&txID); err != nil {
		t.Fatalf("recording a zero-dated transaction: %v", err)
	}

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating the booking_date backfill in over an existing row: %v", err)
	}

	var got time.Time
	if err := pool.QueryRow(ctx, `select booking_date from transactions where id = $1`, txID).Scan(&got); err != nil {
		t.Fatalf("reading the backfilled booking_date: %v", err)
	}
	want := time.Date(2026, time.March, 4, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("booking_date = %s, want %s", got, want)
	}
}

// The trigger is the invariant, so it is exercised directly against the
// database rather than through the Go store methods that will later call it.
func TestAccountOwnerConstraintTrigger(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()

	member := func(email string) string {
		t.Helper()
		var id string
		if err := db.Pool().QueryRow(ctx,
			`insert into members (email, first_name, last_name) values ($1, 'A', 'B') returning id`,
			email).Scan(&id); err != nil {
			t.Fatalf("registering %s: %v", email, err)
		}
		return id
	}
	connection := func(owner string) string {
		t.Helper()
		var id string
		if err := db.Pool().QueryRow(ctx, `
			insert into bank_connections (gateway, bank_id, bank_name, connected_by, consent_expires_at)
			values ('enablebanking', 'PT:Test', 'Test Bank', $1, now() + interval '90 days')
			returning id`, owner).Scan(&id); err != nil {
			t.Fatalf("recording a connection: %v", err)
		}
		return id
	}
	// One transaction: the insert trigger is deferred to commit, but a bare
	// Exec per statement would commit the account with no owner in between,
	// which is exactly the state the trigger exists to refuse.
	account := func(connID string, ref string, owners ...string) string {
		t.Helper()
		tx, err := db.Pool().Begin(ctx)
		if err != nil {
			t.Fatalf("beginning a transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		var id string
		if err := tx.QueryRow(ctx, `
			insert into accounts (source, connection_id, bank_id, gateway_ref, currency)
			values ('gateway', $1, 'PT:Test', $2, 'EUR') returning id`, connID, ref).Scan(&id); err != nil {
			t.Fatalf("recording an account: %v", err)
		}
		for _, o := range owners {
			if _, err := tx.Exec(ctx,
				`insert into account_owners (account_id, member_id) values ($1, $2)`, id, o); err != nil {
				t.Fatalf("owning an account: %v", err)
			}
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("committing a new account: %v", err)
		}
		return id
	}

	t.Run("setting owners to the empty set is refused", func(t *testing.T) {
		ada := member("ada-empty@example.com")
		conn := connection(ada)
		acc := account(conn, "ref-empty", ada)

		// A bare Exec runs as its own single-statement transaction, so the
		// deferred trigger fires at its implicit commit and the call itself
		// reports the refusal.
		if _, err := db.Pool().Exec(ctx, `delete from account_owners where account_id = $1`, acc); err == nil {
			t.Fatal("deleting an account's only owner succeeded, want a refusal")
		}
	})

	t.Run("delete-then-insert inside one transaction is accepted", func(t *testing.T) {
		ada := member("ada-swap@example.com")
		grace := member("grace-swap@example.com")
		conn := connection(ada)
		acc := account(conn, "ref-swap", ada)

		tx, err := db.Pool().Begin(ctx)
		if err != nil {
			t.Fatalf("beginning a transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if _, err := tx.Exec(ctx, `delete from account_owners where account_id = $1`, acc); err != nil {
			t.Fatalf("clearing owners: %v", err)
		}
		if _, err := tx.Exec(ctx,
			`insert into account_owners (account_id, member_id) values ($1, $2)`, acc, grace); err != nil {
			t.Fatalf("inserting the replacement owner: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Errorf("a delete-then-insert swap inside one transaction was refused: %v", err)
		}
	})

	t.Run("deleting an account succeeds", func(t *testing.T) {
		ada := member("ada-delete-account@example.com")
		conn := connection(ada)
		acc := account(conn, "ref-delete-account", ada)

		if _, err := db.Pool().Exec(ctx, `delete from accounts where id = $1`, acc); err != nil {
			t.Errorf("deleting an account with an owner was refused: %v", err)
		}
	})

	t.Run("deleting a connection succeeds", func(t *testing.T) {
		ada := member("ada-delete-connection@example.com")
		conn := connection(ada)
		account(conn, "ref-delete-connection", ada)

		if _, err := db.Pool().Exec(ctx, `delete from bank_connections where id = $1`, conn); err != nil {
			t.Errorf("deleting a connection was refused: %v", err)
		}
	})

	t.Run("deleting a member who is an account's last owner is refused", func(t *testing.T) {
		// Ownership is handed to grace before the delete: ada is also this
		// connection's connected_by, and deleting her would be refused by that
		// foreign key regardless of the trigger under test here. Grace holds no
		// connection, which isolates the invariant this subtest names.
		ada := member("ada-delete-member@example.com")
		grace := member("grace-delete-member@example.com")
		conn := connection(ada)
		acc := account(conn, "ref-delete-member", ada)

		if _, err := db.Pool().Exec(ctx,
			`update account_owners set member_id = $2 where account_id = $1`, acc, grace); err != nil {
			t.Fatalf("handing the account to grace: %v", err)
		}

		if _, err := db.Pool().Exec(ctx, `delete from members where id = $1`, grace); err == nil {
			t.Error("deleting a member who is an account's last owner succeeded, want a refusal")
		}
	})
}
