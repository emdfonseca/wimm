// Package storetest hands a test a migrated, empty database.
//
// It never skips. A store test that quietly passes because no database was
// reachable is a check that cannot fail, which is worse than no check.
package storetest

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

var migrateMu sync.Mutex

// URL is the database the tests run against. `just test apps/wimm` sets
// WIMM_TEST_DATABASE_URL from `bin/db url`; CI starts the same cluster.
func URL(t *testing.T) string {
	t.Helper()

	u := strings.TrimSpace(os.Getenv("WIMM_TEST_DATABASE_URL"))
	if u == "" {
		t.Fatal("WIMM_TEST_DATABASE_URL is not set: run these tests through `just test apps/wimm`, with the database started by `just db-up`")
	}
	return u
}

// New returns a migrated database with every table empty. It truncates rather
// than recreating, so tests share one cluster and stay fast; they must not run
// in parallel against it.
func New(t *testing.T) *store.DB {
	t.Helper()

	url := URL(t)

	migrateMu.Lock()
	err := store.MigrateUp(url)
	migrateMu.Unlock()
	if err != nil {
		t.Fatalf("migrating the test database: %v\n\nstart it with `just db-up`.", err)
	}

	ctx := context.Background()
	db, err := store.Open(ctx, url)
	if err != nil {
		t.Fatalf("opening the test database: %v\n\nstart it with `just db-up`.", err)
	}
	t.Cleanup(db.Close)

	truncateAll(t, ctx, db.Pool())
	return db
}

func truncateAll(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	rows, err := pool.Query(ctx, `
		select tablename from pg_tables
		where schemaname = 'public' and tablename <> 'goose_db_version'`)
	if err != nil {
		t.Fatalf("listing tables: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("reading table name: %v", err)
		}
		names = append(names, `"`+name+`"`)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("listing tables: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("the test database has no tables: migrations did not apply")
	}

	stmt := "truncate " + strings.Join(names, ", ") + " restart identity cascade"
	if _, err := pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("emptying the test database: %v", err)
	}
}
