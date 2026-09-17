package store

import (
	"context"
	"fmt"
	"time"
)

// DeleteBatchSize bounds how many rows one delete statement removes.
//
// An instance that has been running for months has a backlog proportional to
// its uptime, and a single unbounded delete takes a lock proportional to that
// backlog. Batching makes the lock a function of this constant instead, at the
// cost of a statement per batch — which is the right way round for work that
// nothing is waiting on.
const DeleteBatchSize = 1000

// DeleteExpiredCeremonies clears challenges nobody answered.
func (db *DB) DeleteExpiredCeremonies(ctx context.Context) (int64, error) {
	const q = `
		delete from ceremony_challenges
		where id in (
			select id from ceremony_challenges where expires_at <= now() limit $1
		)`

	n, err := db.deleteInBatches(ctx, q)
	if err != nil {
		return n, fmt.Errorf("clearing expired ceremony challenges: %w", err)
	}
	return n, nil
}

// DeleteExpiredEnrolmentTickets clears tickets whose sitting at the enrolment
// page ended without a passkey. A ticket that did produce one is already gone:
// DeleteEnrolmentTicketsForLink spends every ticket the link made.
func (db *DB) DeleteExpiredEnrolmentTickets(ctx context.Context) (int64, error) {
	const q = `
		delete from enrolment_tickets
		where id in (
			select id from enrolment_tickets where expires_at <= now() limit $1
		)`

	n, err := db.deleteInBatches(ctx, q)
	if err != nil {
		return n, fmt.Errorf("clearing expired enrolment tickets: %w", err)
	}
	return n, nil
}

// DeleteFinishedSessions clears sessions that can no longer sign anyone in.
//
// The two ways a session finishes are not the same event. An expired one ended
// on its own and goes at once. A revoked one was ended by a person, and
// revocation is the whole recovery story for a lost device — so it is kept for
// revokedRetention afterwards, which is what makes "was this session actually
// cut off" a question with an answer.
func (db *DB) DeleteFinishedSessions(ctx context.Context, revokedRetention time.Duration) (int64, error) {
	const q = `
		delete from sessions
		where id in (
			select id from sessions
			where expires_at <= now()
			   or (revoked_at is not null
			       and revoked_at <= now() - make_interval(secs => $2))
			limit $1
		)`

	n, err := db.deleteInBatches(ctx, q, revokedRetention.Seconds())
	if err != nil {
		return n, fmt.Errorf("clearing finished sessions: %w", err)
	}
	return n, nil
}

// deleteInBatches runs query until a batch comes back short, and returns the
// total removed. The batch size is always $1; anything else the query needs
// follows it.
//
// A count is returned alongside an error rather than instead of it: the rows
// removed before a failure are gone, and a sweep that reported zero for them
// would be describing a database that does not exist.
func (db *DB) deleteInBatches(ctx context.Context, query string, args ...any) (int64, error) {
	params := append([]any{DeleteBatchSize}, args...)

	var total int64
	for {
		tag, err := db.pool.Exec(ctx, query, params...)
		if err != nil {
			return total, err
		}

		removed := tag.RowsAffected()
		total += removed
		if removed < DeleteBatchSize {
			return total, nil
		}
	}
}

// DeleteAbandonedBankConnections clears authorisations that were begun and
// never returned from.
//
// Most pending connections end this way: a member who closes the tab at their
// bank leaves one behind, so this table accumulates faster than the three
// above it. A consumed row goes too, once it is past its expiry — it is kept
// until then so that a replayed return is refused as spent rather than as
// never-issued, which are different answers and only one of them is true.
func (db *DB) DeleteAbandonedBankConnections(ctx context.Context) (int64, error) {
	const q = `
		delete from pending_bank_connections
		where id in (
			select id from pending_bank_connections where expires_at <= now() limit $1
		)`

	n, err := db.deleteInBatches(ctx, q)
	if err != nil {
		return n, fmt.Errorf("clearing abandoned bank connections: %w", err)
	}
	return n, nil
}
