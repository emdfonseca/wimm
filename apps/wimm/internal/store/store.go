// Package store owns wimmd's database: the connection pool, the migrations it
// applies, and the queries the domain runs. Every lifetime in the schema is
// compared against database time rather than a process clock (ADR 0016).
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is a pool of connections to the wimm database.
type DB struct {
	pool *pgxpool.Pool
}

// Open connects to url and verifies the connection before returning.
func Open(ctx context.Context, url string) (*DB, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("connecting to the database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("reaching the database: %w", err)
	}
	return &DB{pool: pool}, nil
}

// Close releases every connection in the pool.
func (db *DB) Close() { db.pool.Close() }

// Pool exposes the underlying pool to the queries in this package.
func (db *DB) Pool() *pgxpool.Pool { return db.pool }

// Now reads the database's clock. Nothing in wimmd compares a stored timestamp
// against time.Now: a skewed instance must not be able to extend a link or a
// session.
func (db *DB) Now(ctx context.Context) (t Time, err error) {
	err = db.pool.QueryRow(ctx, "select now()").Scan(&t.T)
	if err != nil {
		return Time{}, fmt.Errorf("reading database time: %w", err)
	}
	return t, nil
}
