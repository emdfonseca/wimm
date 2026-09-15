package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Session is what makes a browser signed in: a row an opaque cookie names.
type Session struct {
	ID        string
	MemberID  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// CreateSession records a session. Its expiry is database time plus lifetime,
// so a skewed process clock cannot extend it.
func (db *DB) CreateSession(ctx context.Context, memberID string, valueHash []byte, lifetime time.Duration) (Session, error) {
	const q = `
		insert into sessions (member_id, value_hash, expires_at)
		values ($1, $2, now() + make_interval(secs => $3))
		returning id, member_id, created_at, expires_at`

	var s Session
	err := db.pool.QueryRow(ctx, q, memberID, valueHash, lifetime.Seconds()).
		Scan(&s.ID, &s.MemberID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return Session{}, fmt.Errorf("creating a session: %w", err)
	}
	return s, nil
}

// SessionByHash returns a live session and its member, and records that the
// session was seen. Liveness is decided in SQL: not revoked, not expired,
// both against database time.
func (db *DB) SessionByHash(ctx context.Context, valueHash []byte) (Session, Member, error) {
	const q = `
		update sessions s
		set last_seen_at = now()
		from members m
		where s.value_hash = $1
		  and s.revoked_at is null
		  and s.expires_at > now()
		  and m.id = s.member_id
		returning s.id, s.member_id, s.created_at, s.expires_at,
		          m.id, m.email, m.first_name, m.last_name, m.created_at`

	var s Session
	var m Member
	err := db.pool.QueryRow(ctx, q, valueHash).Scan(
		&s.ID, &s.MemberID, &s.CreatedAt, &s.ExpiresAt,
		&m.ID, &m.Email, &m.FirstName, &m.LastName, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, Member{}, ErrNotFound
	}
	if err != nil {
		return Session{}, Member{}, fmt.Errorf("reading a session: %w", err)
	}
	return s, m, nil
}

// RevokeSession ends a session. Revocation is the whole recovery story for a
// lost device, so it is a column rather than a delete: the row stays for as
// long as the next change needs it.
func (db *DB) RevokeSession(ctx context.Context, valueHash []byte) error {
	const q = `update sessions set revoked_at = now() where value_hash = $1 and revoked_at is null`
	tag, err := db.pool.Exec(ctx, q, valueHash)
	if err != nil {
		return fmt.Errorf("revoking a session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
