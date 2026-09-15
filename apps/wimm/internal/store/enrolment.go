package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// EnrolmentLink is a link row. The value itself is never stored and never read
// back: only its hash is kept, so a copy of the table yields no working links.
type EnrolmentLink struct {
	ID        string
	MemberID  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// EnrolmentTicket is what a redeemed link becomes: short-lived, opaque, and
// carried in a cookie rather than a path.
type EnrolmentTicket struct {
	ID        string
	LinkID    string
	MemberID  string
	ExpiresAt time.Time
}

// IssueEnrolmentLink invalidates any live link the member holds and records a
// new one, in a single transaction, so a member never holds two.
//
// The expiry is computed as database time plus lifetime: a skewed process
// clock cannot extend a link.
func (db *DB) IssueEnrolmentLink(ctx context.Context, memberID string, valueHash []byte, lifetime time.Duration) (EnrolmentLink, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return EnrolmentLink{}, fmt.Errorf("issuing an enrolment link: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const invalidate = `
		update enrolment_links set invalidated_at = now()
		where member_id = $1 and enrolled_at is null and invalidated_at is null`
	if _, err := tx.Exec(ctx, invalidate, memberID); err != nil {
		return EnrolmentLink{}, fmt.Errorf("invalidating outstanding links: %w", err)
	}

	const insert = `
		insert into enrolment_links (member_id, value_hash, expires_at)
		values ($1, $2, now() + make_interval(secs => $3))
		returning id, member_id, created_at, expires_at`

	var l EnrolmentLink
	err = tx.QueryRow(ctx, insert, memberID, valueHash, lifetime.Seconds()).
		Scan(&l.ID, &l.MemberID, &l.CreatedAt, &l.ExpiresAt)
	if err != nil {
		return EnrolmentLink{}, fmt.Errorf("recording an enrolment link: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return EnrolmentLink{}, fmt.Errorf("issuing an enrolment link: %w", err)
	}
	return l, nil
}

// RedeemEnrolmentLink exchanges a link value for a ticket, marking the link
// redeemed. Usability is decided in SQL against database time.
//
// It returns ErrNotFound for a link that has expired, has been used to enrol,
// has been replaced, or was never issued — the caller cannot tell which, which
// is the point.
func (db *DB) RedeemEnrolmentLink(ctx context.Context, linkHash, ticketHash []byte, ticketLifetime time.Duration) (EnrolmentTicket, Member, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return EnrolmentTicket{}, Member{}, fmt.Errorf("redeeming an enrolment link: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const claim = `
		update enrolment_links
		set redeemed_at = coalesce(redeemed_at, now())
		where value_hash = $1
		  and enrolled_at is null
		  and invalidated_at is null
		  and expires_at > now()
		returning id, member_id`

	var linkID, memberID string
	err = tx.QueryRow(ctx, claim, linkHash).Scan(&linkID, &memberID)
	if errors.Is(err, pgx.ErrNoRows) {
		return EnrolmentTicket{}, Member{}, ErrNotFound
	}
	if err != nil {
		return EnrolmentTicket{}, Member{}, fmt.Errorf("claiming an enrolment link: %w", err)
	}

	const insert = `
		insert into enrolment_tickets (enrolment_link_id, member_id, value_hash, expires_at)
		values ($1, $2, $3, now() + make_interval(secs => $4))
		returning id, enrolment_link_id, member_id, expires_at`

	var tkt EnrolmentTicket
	err = tx.QueryRow(ctx, insert, linkID, memberID, ticketHash, ticketLifetime.Seconds()).
		Scan(&tkt.ID, &tkt.LinkID, &tkt.MemberID, &tkt.ExpiresAt)
	if err != nil {
		return EnrolmentTicket{}, Member{}, fmt.Errorf("recording an enrolment ticket: %w", err)
	}

	const member = `
		select id, email, first_name, last_name, created_at from members where id = $1`
	m, err := scanMember(tx.QueryRow(ctx, member, memberID))
	if err != nil {
		return EnrolmentTicket{}, Member{}, fmt.Errorf("reading the member behind a link: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return EnrolmentTicket{}, Member{}, fmt.Errorf("redeeming an enrolment link: %w", err)
	}
	return tkt, m, nil
}

// EnrolmentTicketByHash returns a ticket that has not expired, with its member.
func (db *DB) EnrolmentTicketByHash(ctx context.Context, ticketHash []byte) (EnrolmentTicket, Member, error) {
	const q = `
		select t.id, t.enrolment_link_id, t.member_id, t.expires_at,
		       m.id, m.email, m.first_name, m.last_name, m.created_at
		from enrolment_tickets t
		join members m on m.id = t.member_id
		where t.value_hash = $1 and t.expires_at > now()`

	var tkt EnrolmentTicket
	var m Member
	err := db.pool.QueryRow(ctx, q, ticketHash).Scan(
		&tkt.ID, &tkt.LinkID, &tkt.MemberID, &tkt.ExpiresAt,
		&m.ID, &m.Email, &m.FirstName, &m.LastName, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return EnrolmentTicket{}, Member{}, ErrNotFound
	}
	if err != nil {
		return EnrolmentTicket{}, Member{}, fmt.Errorf("finding an enrolment ticket: %w", err)
	}
	return tkt, m, nil
}

// MarkLinkEnrolled closes a link for good. This is what makes it single-use: a
// refused attempt never reaches here, so the link stays usable.
func (db *DB) MarkLinkEnrolled(ctx context.Context, linkID string) error {
	const q = `update enrolment_links set enrolled_at = now() where id = $1 and enrolled_at is null`
	tag, err := db.pool.Exec(ctx, q, linkID)
	if err != nil {
		return fmt.Errorf("closing an enrolment link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
