package storetest

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Scratch returns the URL of a database of its own, dropped when the test
// ends.
//
// Go runs packages' tests in parallel against one cluster, so a test that
// migrates the schema down would pull it out from under everything else. A
// test that changes the shape of the database needs its own.
func Scratch(t *testing.T, name string) string {
	t.Helper()

	base := URL(t)
	dbname := "wimm_test_" + sanitise(name)

	admin := replaceDatabase(t, base, "postgres")
	exec(t, admin, `drop database if exists "`+dbname+`" with (force)`)
	exec(t, admin, `create database "`+dbname+`"`)

	t.Cleanup(func() {
		exec(t, admin, `drop database if exists "`+dbname+`" with (force)`)
	})

	return replaceDatabase(t, base, dbname)
}

func replaceDatabase(t *testing.T, raw, dbname string) string {
	t.Helper()

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parsing %q: %v", raw, err)
	}
	u.Path = "/" + dbname
	return u.String()
}

func exec(t *testing.T, dsn, statement string) {
	t.Helper()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting to %q: %v", dsn, err)
	}
	defer func() { _ = conn.Close(ctx) }()

	if _, err := conn.Exec(ctx, statement); err != nil {
		t.Fatalf("running %q: %v", statement, err)
	}
}

func sanitise(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return fmt.Sprintf("%d", len(name))
	}
	return b.String()
}
