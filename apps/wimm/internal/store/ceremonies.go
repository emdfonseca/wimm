package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// CeremonyKind distinguishes a registration ceremony from an authentication
// one. It is the enum the schema declares.
type CeremonyKind string

const (
	CeremonyRegistration   CeremonyKind = "registration"
	CeremonyAuthentication CeremonyKind = "authentication"
)

// Ceremony is one WebAuthn challenge in flight. It lives in Postgres with a
// lifetime in seconds and is deleted on use, so wimmd holds no
// request-spanning state and needs no sticky routing (ADR 0016).
type Ceremony struct {
	ID              string
	Kind            CeremonyKind
	MemberID        *string
	EnrolmentLinkID *string
	SessionData     []byte
	IntendedPath    *string
	ExpiresAt       time.Time
}

// CreateCeremony records a challenge.
func (db *DB) CreateCeremony(ctx context.Context, c Ceremony, lifetime time.Duration) (Ceremony, error) {
	const q = `
		insert into ceremony_challenges (kind, member_id, enrolment_link_id, session_data, intended_path, expires_at)
		values ($1, $2, $3, $4, $5, now() + make_interval(secs => $6))
		returning id, expires_at`

	err := db.pool.QueryRow(ctx, q,
		string(c.Kind), c.MemberID, c.EnrolmentLinkID, c.SessionData, c.IntendedPath, lifetime.Seconds()).
		Scan(&c.ID, &c.ExpiresAt)
	if err != nil {
		return Ceremony{}, fmt.Errorf("recording a ceremony challenge: %w", err)
	}
	return c, nil
}

// ConsumeCeremony reads a challenge and deletes it in the same statement, so a
// challenge is answered at most once even with two instances running.
func (db *DB) ConsumeCeremony(ctx context.Context, id string, kind CeremonyKind) (Ceremony, error) {
	const q = `
		delete from ceremony_challenges
		where id = $1 and kind = $2 and expires_at > now()
		returning id, kind, member_id, enrolment_link_id, session_data, intended_path, expires_at`

	var c Ceremony
	var kindText string
	err := db.pool.QueryRow(ctx, q, id, string(kind)).
		Scan(&c.ID, &kindText, &c.MemberID, &c.EnrolmentLinkID, &c.SessionData, &c.IntendedPath, &c.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Ceremony{}, ErrNotFound
	}
	if err != nil {
		return Ceremony{}, fmt.Errorf("consuming a ceremony challenge: %w", err)
	}
	c.Kind = CeremonyKind(kindText)
	return c, nil
}

// DeleteExpiredCeremonies clears challenges nobody answered.
func (db *DB) DeleteExpiredCeremonies(ctx context.Context) (int64, error) {
	tag, err := db.pool.Exec(ctx, `delete from ceremony_challenges where expires_at <= now()`)
	if err != nil {
		return 0, fmt.Errorf("clearing expired ceremony challenges: %w", err)
	}
	return tag.RowsAffected(), nil
}
