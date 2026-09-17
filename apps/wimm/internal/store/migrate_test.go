package store_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
	"github.com/xuuid/wimm/apps/wimm/internal/store/storetest"
)

// A down migration that leaves anything behind turns a rollback into a manual
// clean-up, so the round trip is asserted rather than assumed.
func TestMigrationsRoundTripToEmpty(t *testing.T) {
	// Its own database: this test rolls the schema back to nothing, and the
	// rest of the suite is running against the shared one at the same time.
	url := storetest.Scratch(t, "migrations")

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating up: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()

	if got := tables(t, ctx, pool); len(got) == 0 {
		t.Fatal("after migrating up the schema has no tables")
	}

	if err := store.MigrateDownTo(url, 0); err != nil {
		t.Fatalf("migrating down: %v", err)
	}

	if got := tables(t, ctx, pool); len(got) != 0 {
		t.Errorf("after migrating down these tables remain: %v", got)
	}
	if got := enumTypes(t, ctx, pool); len(got) != 0 {
		t.Errorf("after migrating down these types remain: %v", got)
	}
}

func TestDatabaseTimeIsReadable(t *testing.T) {
	db := storetest.New(t)

	now, err := db.Now(context.Background())
	if err != nil {
		t.Fatalf("reading database time: %v", err)
	}
	if now.T.IsZero() {
		t.Error("database time is the zero value")
	}
}

func tables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []string {
	t.Helper()
	return strings(t, ctx, pool, `
		select tablename from pg_tables
		where schemaname = 'public' and tablename <> 'goose_db_version'
		order by 1`)
}

func enumTypes(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []string {
	t.Helper()
	return strings(t, ctx, pool, `
		select t.typname from pg_type t
		join pg_namespace n on n.oid = t.typnamespace
		where n.nspname = 'public' and t.typtype = 'e'
		order by 1`)
}

func strings(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string) []string {
	t.Helper()

	rows, err := pool.Query(ctx, query)
	if err != nil {
		t.Fatalf("querying: %v", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("scanning: %v", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("querying: %v", err)
	}
	return out
}

// Every lifetime is computed as database time plus a duration. Asserting the
// interval in SQL is what proves it: the process clock is never consulted, so
// a skewed instance cannot extend a session.
func TestSessionExpiryIsMeasuredAgainstDatabaseTime(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()

	m, err := db.CreateMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	const lifetime = 72 * time.Hour
	s, err := db.CreateSession(ctx, m.ID, []byte("hash"), lifetime)
	if err != nil {
		t.Fatalf("creating a session: %v", err)
	}

	var seconds float64
	if err := db.Pool().QueryRow(ctx,
		`select extract(epoch from (expires_at - now())) from sessions where id = $1`, s.ID,
	).Scan(&seconds); err != nil {
		t.Fatalf("measuring the session: %v", err)
	}

	got := time.Duration(seconds * float64(time.Second))
	if diff := got - lifetime; diff > 5*time.Second || diff < -5*time.Second {
		t.Errorf("the session lasts %s by the database's own clock, want %s", got, lifetime)
	}
}

// The sweep's indexes are the first migration to land on a schema that already
// has rows in it, so its rollback has to put the schema back rather than
// approximately back. Rolling forward again has to restore it exactly.
func TestSweepIndexesRoundTripToThePreviousSchema(t *testing.T) {
	// Its own database: this test migrates backwards, and the rest of the
	// suite is running against the shared one at the same time.
	url := storetest.Scratch(t, "sweep_indexes")

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating up: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()

	const identityOnly = 1

	// The schema migration 1 alone produces, captured rather than derived.
	// Deriving it as "head minus what this migration adds" would encode the
	// number of migrations that exist today, and would fail on the next one
	// for a reason that has nothing to do with the sweep.
	if err := store.MigrateDownTo(url, identityOnly); err != nil {
		t.Fatalf("rolling back to the identity schema: %v", err)
	}
	atIdentity := indexes(t, ctx, pool)

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating up: %v", err)
	}
	atHead := indexes(t, ctx, pool)

	sweepIndexes := []string{
		"enrolment_tickets_expires_at_idx",
		"sessions_expires_at_idx",
		"sessions_revoked_at_idx",
	}
	for _, name := range sweepIndexes {
		if !slices.Contains(atHead, name) {
			t.Fatalf("after migrating up %s is missing; head has %v", name, atHead)
		}
	}

	if err := store.MigrateDownTo(url, identityOnly); err != nil {
		t.Fatalf("rolling the sweep indexes back: %v", err)
	}

	rolledBack := indexes(t, ctx, pool)
	for _, name := range sweepIndexes {
		if slices.Contains(rolledBack, name) {
			t.Errorf("%s survived the rollback", name)
		}
	}
	// Back to exactly the identity schema: every rollback removed what its
	// migration added, and nothing else.
	if !slices.Equal(rolledBack, atIdentity) {
		t.Errorf("after the rollback the schema has %v, want %v", rolledBack, atIdentity)
	}
	if got := tables(t, ctx, pool); len(got) == 0 {
		t.Error("the rollback took the identity tables with it")
	}

	if err := store.MigrateUp(url); err != nil {
		t.Fatalf("migrating up again: %v", err)
	}
	if got := indexes(t, ctx, pool); !slices.Equal(got, atHead) {
		t.Errorf("after rolling forward again the schema has %v, want %v", got, atHead)
	}
}

func indexes(t *testing.T, ctx context.Context, pool *pgxpool.Pool) []string {
	t.Helper()
	return strings(t, ctx, pool, `
		select indexname from pg_indexes
		where schemaname = 'public' and tablename <> 'goose_db_version'
		order by 1`)
}
