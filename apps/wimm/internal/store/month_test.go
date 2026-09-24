package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store/storetest"
)

// monthAccount stores a sync for a member on a fresh account and returns it.
func monthAccount(t *testing.T, ctx context.Context, db *store.DB, owner string, txs ...store.Transaction) string {
	t.Helper()
	_, stored := connect(t, ctx, db, owner, 90*24*time.Hour, account("hash-m", "Conta à Ordem", "0538"))
	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID, txs, day(28)); err != nil {
		t.Fatalf("WriteAccountTransactions: %v", err)
	}
	return stored[0].ID
}

func TestFullAccessCountsOwnersAndDetailsGrantsButNotBalanceGrants(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	linus := member(t, ctx, db, "linus@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour,
		account("hash-1", "Joint", "0538"), account("hash-2", "Personal", "5594"))
	if err := db.SetAccountOwners(ctx, stored[0].ID, []string{ada.ID, grace.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAccountLevel(ctx, stored[0].ID, linus.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAccountLevel(ctx, stored[1].ID, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatal(err)
	}

	counts, err := db.FullAccessCounts(ctx)
	if err != nil {
		t.Fatalf("FullAccessCounts: %v", err)
	}
	if counts[stored[0].ID] != 3 {
		t.Errorf("joint = %d, want 3: two owners and one details grant", counts[stored[0].ID])
	}
	if counts[stored[1].ID] != 1 {
		t.Errorf("personal = %d, want 1: a balance grant is not full access", counts[stored[1].ID])
	}
}

func TestAnAccountIdTheMemberDoesNotOwnMatchesNothingInEveryOwnedQuery(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	id := monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))
	if err := db.SetAccountLevel(ctx, id, grace.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatal(err)
	}
	ids := []string{id}

	if rows, err := db.OwnedBooked(ctx, grace.ID, ids, day(1), day(28)); err != nil || len(rows) != 0 {
		t.Errorf("OwnedBooked = %d rows, %v, want nothing", len(rows), err)
	}
	if accounts, err := db.OwnedAccountsForTrend(ctx, grace.ID, ids); err != nil || len(accounts) != 0 {
		t.Errorf("OwnedAccountsForTrend = %+v, %v, want nothing", accounts, err)
	}
}

func TestOwnedBookedIsBothDirectionsBookedOnlyAndBoundedByArguments(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	id := monthAccount(t, ctx, db, ada.ID,
		booked("salary", 2, 200_000),
		booked("rent", 3, -80_000),
		pending("holds", 4, -9_999),
		booked("before", 1, -111),
		booked("end", 9, -1),
	)

	rows, err := db.OwnedBooked(ctx, ada.ID, []string{id}, day(2), day(9))
	if err != nil {
		t.Fatalf("OwnedBooked: %v", err)
	}
	if len(rows) != 2 || rows[0].DedupKey != "rent" || rows[1].DedupKey != "salary" {
		t.Fatalf("got %+v, want rent then salary: booked, both directions, in [2, 9)", rows)
	}
}

func TestOwnedBookedStaysFastAtFiftyThousandRows(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Conta a Ordem", "0538"))
	accountID := stored[0].ID

	const total = 50_000
	start := time.Date(2024, time.September, 1, 0, 0, 0, 0, time.UTC)
	rows := make([][]any, total)
	for i := range total {
		rows[i] = []any{
			accountID, string(store.StatusBooked), fmt.Sprintf("bulk-%d", i), int64(-100 - i),
			"EUR", start.AddDate(0, 0, i%750),
		}
	}
	if _, err := db.Pool().CopyFrom(ctx,
		pgx.Identifier{"transactions"},
		[]string{"account_id", "status", "dedup_key", "amount_minor", "currency", "booking_date"},
		pgx.CopyFromRows(rows),
	); err != nil {
		t.Fatalf("seeding 50,000 transactions: %v", err)
	}

	ids := []string{accountID}
	to := start.AddDate(0, 0, 800)
	began := time.Now()
	got, err := db.OwnedBooked(ctx, ada.ID, ids, start, to)
	elapsed := time.Since(began)
	if err != nil {
		t.Fatalf("OwnedBooked: %v", err)
	}
	t.Logf("OwnedBooked over %d rows took %s", total, elapsed)
	if len(got) != total {
		t.Errorf("OwnedBooked returned %d rows, want %d", len(got), total)
	}
	if elapsed > 3*time.Second {
		t.Errorf("OwnedBooked took %s over %d rows, want well under 3s", elapsed, total)
	}
}

// The rules that used to be asserted against OwnedWindowSums and
// OwnedMonthlySums now ride on OwnedBooked, which is the one query the domain
// sums from (ADR 0026). None of them is about summing.

func TestOwnedBookedNeverReadsAnotherMembersAccount(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	id := monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))
	if err := db.SetAccountLevel(ctx, id, grace.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatal(err)
	}

	rows, err := db.OwnedBooked(ctx, grace.ID, []string{id}, day(1), day(28))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Errorf("a member granted details read %d rows", len(rows))
	}
}

func TestALeftOutAccountIsInNoOwnedRead(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	id := monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))
	if err := db.SetAccountLeftOut(ctx, id, true); err != nil {
		t.Fatal(err)
	}

	rows, err := db.OwnedBooked(ctx, ada.ID, []string{id}, day(1), day(28))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Errorf("got %d rows, want nothing for a left-out account", len(rows))
	}
}

func TestAnEmptyAccountScopeIsAnEmptyScope(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))

	for _, ids := range [][]string{nil, {}} {
		rows, err := db.OwnedBooked(ctx, ada.ID, ids, day(1), day(28))
		if err != nil || len(rows) != 0 {
			t.Errorf("ids %#v: got %d rows, %v, want nothing", ids, len(rows), err)
		}
	}
}
