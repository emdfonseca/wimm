package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Credential is an enrolled passkey. "Credential" is the WebAuthn word and
// stays in code; a member reads "passkey".
//
// Data is the library's own credential record, stored whole and handed back
// verbatim on validation.
type Credential struct {
	ID           string
	MemberID     string
	CredentialID []byte
	Data         []byte
	CreatedAt    time.Time
	LastUsedAt   *time.Time
}

// CreateCredential records a passkey against a member.
func (db *DB) CreateCredential(ctx context.Context, c Credential) (Credential, error) {
	const q = `
		insert into credentials (member_id, credential_id, data)
		values ($1, $2, $3)
		returning id, created_at`

	if err := db.pool.QueryRow(ctx, q, c.MemberID, c.CredentialID, c.Data).
		Scan(&c.ID, &c.CreatedAt); err != nil {
		return Credential{}, fmt.Errorf("recording a passkey: %w", err)
	}
	return c, nil
}

// CredentialByID finds a passkey by the identifier the authenticator sent.
// ErrNotFound means this instance has no record of it — which is all the
// caller learns, and all the member is told.
func (db *DB) CredentialByID(ctx context.Context, credentialID []byte) (Credential, error) {
	const q = `
		select id, member_id, credential_id, data, created_at, last_used_at
		from credentials where credential_id = $1`

	c, err := scanCredential(db.pool.QueryRow(ctx, q, credentialID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Credential{}, ErrNotFound
	}
	if err != nil {
		return Credential{}, fmt.Errorf("finding a passkey: %w", err)
	}
	return c, nil
}

// CredentialsForMember lists a member's passkeys, oldest first.
func (db *DB) CredentialsForMember(ctx context.Context, memberID string) ([]Credential, error) {
	const q = `
		select id, member_id, credential_id, data, created_at, last_used_at
		from credentials where member_id = $1 order by created_at`

	rows, err := db.pool.Query(ctx, q, memberID)
	if err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	defer rows.Close()

	var out []Credential
	for rows.Next() {
		c, err := scanCredential(rows)
		if err != nil {
			return nil, fmt.Errorf("reading a passkey: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	return out, nil
}

// UpdateCredential writes back what a successful assertion changed: the
// signature counter and the authenticator flags, both inside Data.
func (db *DB) UpdateCredential(ctx context.Context, id string, data []byte) error {
	const q = `update credentials set data = $2, last_used_at = now() where id = $1`
	tag, err := db.pool.Exec(ctx, q, id, data)
	if err != nil {
		return fmt.Errorf("recording a passkey use: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanCredential(row pgx.Row) (Credential, error) {
	var c Credential
	err := row.Scan(&c.ID, &c.MemberID, &c.CredentialID, &c.Data, &c.CreatedAt, &c.LastUsedAt)
	return c, err
}
