package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Member is a person registered to use this instance.
type Member struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
}

const uniqueViolation = "23505"

// CreateMember registers a person. The email is stored as the operator typed
// it; uniqueness is on its lower-case form.
func (db *DB) CreateMember(ctx context.Context, email, firstName, lastName string) (Member, error) {
	const q = `
		insert into members (email, first_name, last_name)
		values ($1, $2, $3)
		returning id, email, first_name, last_name, created_at`

	m, err := scanMember(db.pool.QueryRow(ctx, q, email, firstName, lastName))
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == uniqueViolation && pg.ConstraintName == "members_email_key" {
			return Member{}, ErrEmailAlreadyRegistered
		}
		return Member{}, fmt.Errorf("registering a member: %w", err)
	}
	return m, nil
}

// MemberByEmail finds a member by their address, case-insensitively.
func (db *DB) MemberByEmail(ctx context.Context, email string) (Member, error) {
	const q = `
		select id, email, first_name, last_name, created_at
		from members where lower(email) = lower($1)`

	m, err := scanMember(db.pool.QueryRow(ctx, q, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	if err != nil {
		return Member{}, fmt.Errorf("finding a member by email: %w", err)
	}
	return m, nil
}

// MemberByID finds a member by their identifier.
func (db *DB) MemberByID(ctx context.Context, id string) (Member, error) {
	const q = `
		select id, email, first_name, last_name, created_at
		from members where id = $1`

	m, err := scanMember(db.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	if err != nil {
		return Member{}, fmt.Errorf("finding a member by id: %w", err)
	}
	return m, nil
}

func scanMember(row pgx.Row) (Member, error) {
	var m Member
	err := row.Scan(&m.ID, &m.Email, &m.FirstName, &m.LastName, &m.CreatedAt)
	return m, err
}
