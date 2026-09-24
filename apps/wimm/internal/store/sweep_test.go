package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store/storetest"
)

// revokedRetention is what the sweep tests hold a revoked session for. It is
// not the configured default: a test that borrowed that constant would keep
// passing if the default changed to zero.
const revokedRetention = 7 * 24 * time.Hour

// The backlog on an instance that has been running for months is larger than
// one batch, so the delete has to repeat until it comes up short. A single
// pass would leave everything past the first batch behind and report a total
// that looked like success.
func TestExpiredCeremoniesAreDeletedPastTheFirstBatch(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()

	const expired = store.DeleteBatchSize + 7
	seedCeremonies(t, ctx, db, expired, -time.Minute)
	seedCeremonies(t, ctx, db, 3, time.Hour)

	n, err := db.DeleteExpiredCeremonies(ctx)
	if err != nil {
		t.Fatalf("sweeping ceremonies: %v", err)
	}
	if n != expired {
		t.Errorf("removed %d challenges, want %d", n, expired)
	}
	if got := count(t, ctx, db, "ceremony_challenges"); got != 3 {
		t.Errorf("%d challenges remain, want the 3 that have not expired", got)
	}
}

func TestAnExpiredEnrolmentTicketGoesAndALiveOneStays(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()

	m := member(t, ctx, db, "ada@example.com")
	dead := seedTicket(t, ctx, db, m.ID, "dead", -time.Second)
	live := seedTicket(t, ctx, db, m.ID, "live", time.Hour)

	n, err := db.DeleteExpiredEnrolmentTickets(ctx)
	if err != nil {
		t.Fatalf("sweeping tickets: %v", err)
	}
	if n != 1 {
		t.Errorf("removed %d tickets, want 1", n)
	}
	if exists(t, ctx, db, "enrolment_tickets", dead) {
		t.Error("the expired ticket is still there")
	}
	if !exists(t, ctx, db, "enrolment_tickets", live) {
		t.Error("a ticket still inside its lifetime was removed")
	}
}

// Revocation is the recovery story for a lost device. A row deleted the
// instant it is revoked cannot answer "was this actually cut off", which is
// the only question anyone asks afterwards.
func TestARevokedSessionIsKeptForTheRetentionWindow(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()

	m := member(t, ctx, db, "ada@example.com")
	live := seedSession(t, ctx, db, m.ID, "live", time.Hour, nil)
	expired := seedSession(t, ctx, db, m.ID, "expired", -time.Minute, nil)
	justRevoked := seedSession(t, ctx, db, m.ID, "just-revoked", time.Hour, ptr(-time.Minute))
	longRevoked := seedSession(t, ctx, db, m.ID, "long-revoked", time.Hour, ptr(-2*revokedRetention))

	n, err := db.DeleteFinishedSessions(ctx, revokedRetention)
	if err != nil {
		t.Fatalf("sweeping sessions: %v", err)
	}
	if n != 2 {
		t.Errorf("removed %d sessions, want 2", n)
	}

	for _, tc := range []struct {
		name string
		id   string
		want bool
	}{
		{"a live session", live, true},
		{"an expired session", expired, false},
		{"a session revoked a minute ago", justRevoked, true},
		{"a session revoked longer ago than the window", longRevoked, false},
	} {
		if got := exists(t, ctx, db, "sessions", tc.id); got != tc.want {
			t.Errorf("%s: present = %t, want %t", tc.name, got, tc.want)
		}
	}
}

// Every lifetime in wimm is compared against database time (ADR 0016), and the
// sweep is no exception: a skewed instance must not be able to delete a live
// session or spare a dead one. The seeded offsets are read back in SQL, so the
// assertion below is about the database's clock rather than this process's.
func TestTheSweepMeasuresAgainstDatabaseTime(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()

	m := member(t, ctx, db, "ada@example.com")
	const margin = 30 * time.Minute
	inside := seedSession(t, ctx, db, m.ID, "inside", time.Hour, ptr(margin-revokedRetention))
	outside := seedSession(t, ctx, db, m.ID, "outside", time.Hour, ptr(-margin-revokedRetention))

	// The two rows straddle the window by the same margin, measured by the
	// database rather than asserted from here.
	if age := revokedAge(t, ctx, db, inside); age > revokedRetention {
		t.Fatalf("the surviving row was revoked %s ago, already past the %s window", age, revokedRetention)
	}
	if age := revokedAge(t, ctx, db, outside); age < revokedRetention {
		t.Fatalf("the removed row was revoked %s ago, still inside the %s window", age, revokedRetention)
	}

	if _, err := db.DeleteFinishedSessions(ctx, revokedRetention); err != nil {
		t.Fatalf("sweeping sessions: %v", err)
	}

	if !exists(t, ctx, db, "sessions", inside) {
		t.Error("a session revoked inside the window by database time was removed")
	}
	if exists(t, ctx, db, "sessions", outside) {
		t.Error("a session revoked outside the window by database time survived")
	}
}

func ptr(d time.Duration) *time.Duration { return &d }

func member(t *testing.T, ctx context.Context, db *store.DB, email string) store.Member {
	t.Helper()

	m, err := db.CreateMember(ctx, email, "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering %s: %v", email, err)
	}
	return m
}

// seedCeremonies inserts n challenges whose expiry is offset from database
// time. Offsets are applied in SQL for the same reason the production queries
// compare in SQL.
func seedCeremonies(t *testing.T, ctx context.Context, db *store.DB, n int, offset time.Duration) {
	t.Helper()

	_, err := db.Pool().Exec(ctx, `
		insert into ceremony_challenges (kind, session_data, expires_at)
		select 'authentication', '{}'::jsonb, now() + make_interval(secs => $2)
		from generate_series(1, $1)`, n, offset.Seconds())
	if err != nil {
		t.Fatalf("seeding %d challenges: %v", n, err)
	}
}

func seedTicket(t *testing.T, ctx context.Context, db *store.DB, memberID, label string, offset time.Duration) string {
	t.Helper()

	link, err := db.IssueEnrolmentLink(ctx, memberID, []byte("link-"+label), time.Hour)
	if err != nil {
		t.Fatalf("issuing a link for %s: %v", label, err)
	}

	var id string
	err = db.Pool().QueryRow(ctx, `
		insert into enrolment_tickets (enrolment_link_id, member_id, value_hash, expires_at)
		values ($1, $2, $3, now() + make_interval(secs => $4))
		returning id`, link.ID, memberID, []byte("ticket-"+label), offset.Seconds()).Scan(&id)
	if err != nil {
		t.Fatalf("seeding the %s ticket: %v", label, err)
	}
	return id
}

// seedSession records a session expiring at an offset from database time, and
// revoked at another offset when revokedOffset is not nil.
func seedSession(t *testing.T, ctx context.Context, db *store.DB, memberID, label string, offset time.Duration, revokedOffset *time.Duration) string {
	t.Helper()

	var revoked *float64
	if revokedOffset != nil {
		seconds := revokedOffset.Seconds()
		revoked = &seconds
	}

	var id string
	err := db.Pool().QueryRow(ctx, `
		insert into sessions (member_id, value_hash, expires_at, revoked_at)
		values ($1, $2, now() + make_interval(secs => $3),
		        case when $4::float8 is null then null
		             else now() + make_interval(secs => $4) end)
		returning id`, memberID, []byte("session-"+label), offset.Seconds(), revoked).Scan(&id)
	if err != nil {
		t.Fatalf("seeding the %s session: %v", label, err)
	}
	return id
}

// revokedAge is how long ago a session was revoked, by the database's clock.
func revokedAge(t *testing.T, ctx context.Context, db *store.DB, id string) time.Duration {
	t.Helper()

	var seconds float64
	err := db.Pool().QueryRow(ctx,
		`select extract(epoch from (now() - revoked_at)) from sessions where id = $1`, id).Scan(&seconds)
	if err != nil {
		t.Fatalf("measuring when %s was revoked: %v", id, err)
	}
	return time.Duration(seconds * float64(time.Second))
}

func count(t *testing.T, ctx context.Context, db *store.DB, table string) int {
	t.Helper()

	var n int
	// The table name is a literal from this file, never a value.
	if err := db.Pool().QueryRow(ctx, `select count(*) from `+table).Scan(&n); err != nil {
		t.Fatalf("counting %s: %v", table, err)
	}
	return n
}

func exists(t *testing.T, ctx context.Context, db *store.DB, table, id string) bool {
	t.Helper()

	var present bool
	if err := db.Pool().QueryRow(ctx,
		`select exists (select 1 from `+table+` where id = $1)`, id).Scan(&present); err != nil {
		t.Fatalf("looking for %s in %s: %v", id, table, err)
	}
	return present
}
