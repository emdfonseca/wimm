package store

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"

	// Registers the "pgx" database/sql driver that goose runs against.
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationsFS is the migration set wimmd owns, so a test applies exactly what
// `just db-up` applies.
func MigrationsFS() embed.FS { return migrationsFS }

func migrator(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("opening the database for migration: %w", err)
	}
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("selecting the postgres dialect: %w", err)
	}
	return db, nil
}

// MigrateUp brings the schema at url to head.
func MigrateUp(url string) error {
	db, err := migrator(url)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("migrating up: %w", err)
	}
	return nil
}

// MigrateDownTo rolls the schema at url back to version, 0 being empty.
func MigrateDownTo(url string, version int64) error {
	db, err := migrator(url)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := goose.DownTo(db, "migrations", version); err != nil {
		return fmt.Errorf("migrating down to %d: %w", version, err)
	}
	return nil
}
