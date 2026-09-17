package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
	"github.com/xuuid/wimm/apps/wimm/internal/identity"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// fakeSweepStore records what ran and fails what it is told to. A fake rather
// than a mock: these tests assert which deletes happened and what they
// reported, not the shape of the calls.
type fakeSweepStore struct {
	mu sync.Mutex

	ceremonies, sessions, tickets, pending int
	removed                                int64

	failSessions error
	// failPending fails only the fourth statement, so the other three can be
	// shown to run anyway.
	failPending error
	// failFirstSweep fails every table on the first pass only, so a later
	// sweep can be shown to run normally afterwards.
	failFirstSweep error

	// started signals the beginning of each sweep, so a test waits for a pass
	// rather than sleeping for one.
	started chan int
}

func (f *fakeSweepStore) DeleteExpiredCeremonies(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ceremonies++
	if f.started != nil {
		select {
		case f.started <- f.ceremonies:
		default:
		}
	}
	if err := f.firstSweepFailure(); err != nil {
		return 0, err
	}
	return f.removed, nil
}

func (f *fakeSweepStore) DeleteFinishedSessions(context.Context, time.Duration) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.sessions++
	if err := f.firstSweepFailure(); err != nil {
		return 0, err
	}
	return f.removed, f.failSessions
}

func (f *fakeSweepStore) DeleteExpiredEnrolmentTickets(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.tickets++
	if err := f.firstSweepFailure(); err != nil {
		return 0, err
	}
	return f.removed, nil
}

func (f *fakeSweepStore) DeleteAbandonedBankConnections(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.pending++
	if f.failPending != nil {
		return 0, f.failPending
	}
	if err := f.firstSweepFailure(); err != nil {
		return 0, err
	}
	return f.removed, nil
}

// firstSweepFailure is held while f.mu is locked; the ceremony delete is the
// first of the three, so its count is the pass number.
func (f *fakeSweepStore) firstSweepFailure() error {
	if f.failFirstSweep != nil && f.ceremonies == 1 {
		return f.failFirstSweep
	}
	return nil
}

func (f *fakeSweepStore) counts() (int, int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ceremonies, f.sessions, f.tickets
}

func (f *fakeSweepStore) pendingCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.pending
}

// syncBuffer is a log destination a test goroutine can read while Run is
// writing to it.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *syncBuffer) logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(b, nil))
}

// lines parses what was logged. JSON, because the assertions are about
// attributes, which a text handler would bury in a formatted string.
func (b *syncBuffer) lines(t *testing.T) []map[string]any {
	t.Helper()

	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(b.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("log line %q is not JSON: %v", line, err)
		}
		out = append(out, entry)
	}
	return out
}

// sweepConfig builds configuration through the real loader, so a default that
// stopped existing would fail here too rather than being restated in a fixture.
func sweepConfig(t *testing.T, overrides map[string]string) config.Config {
	t.Helper()

	e := map[string]string{
		"WIMM_OPERATOR_CREDENTIAL": "credential",
		"WIMM_DATABASE_URL":        "postgres://localhost/wimm",
	}
	for k, v := range overrides {
		e[k] = v
	}

	cfg, err := config.Load(func(k string) string { return e[k] })
	if err != nil {
		t.Fatalf("building test configuration: %v", err)
	}
	return cfg
}

// A sweep that gave up on the first failure would leave two tables growing
// because a third one broke. There is no invariant spanning them, so there is
// no reason for one to stop the others.
func TestAFailureInOneTableDoesNotSkipTheOthers(t *testing.T) {
	boom := errors.New("the sessions delete failed")
	fake := &fakeSweepStore{removed: 4, failSessions: boom}
	log := &syncBuffer{}

	r := identity.NewSweeper(fake, log.logger(), sweepConfig(t, nil)).Sweep(context.Background())

	ceremonies, sessions, tickets := fake.counts()
	if ceremonies != 1 || sessions != 1 || tickets != 1 {
		t.Errorf("ran %d ceremony, %d session and %d ticket deletes, want one of each",
			ceremonies, sessions, tickets)
	}
	if !errors.Is(r.Sessions.Err, boom) {
		t.Errorf("Sessions.Err = %v, want the sessions failure", r.Sessions.Err)
	}
	if r.Ceremonies.Err != nil || r.Tickets.Err != nil {
		t.Errorf("the working tables reported errors: %v, %v", r.Ceremonies.Err, r.Tickets.Err)
	}
	if r.Ceremonies.Removed != 4 || r.Tickets.Removed != 4 {
		t.Errorf("the working tables removed %d and %d, want 4 each",
			r.Ceremonies.Removed, r.Tickets.Removed)
	}
	if !r.Failed() {
		t.Error("a sweep with a failing table does not report as failed")
	}

	// The two that worked are still reported, so the failure does not hide
	// what the sweep did manage.
	line := findLine(t, log.lines(t), "swept expired identity rows")
	if line["ceremonies"] != float64(4) || line["tickets"] != float64(4) {
		t.Errorf("the sweep line is %v, want the working tables' counts", line)
	}
}

// Shutdown must not wait out an interval. An hour is the default, and a
// service that took an hour to stop is a service nobody can deploy.
func TestRunStopsWhenTheContextIsCancelled(t *testing.T) {
	fake := &fakeSweepStore{started: make(chan int, 4)}
	sweeper := identity.NewSweeper(fake, (&syncBuffer{}).logger(),
		sweepConfig(t, map[string]string{"WIMM_SWEEP_INTERVAL": "1h"}))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sweeper.Run(ctx)
		close(done)
	}()

	// Let the first sweep happen, so what is being timed is the wait for the
	// next tick rather than a race with startup.
	waitForSweep(t, fake, 1)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within five seconds of the context being cancelled, with an hour to the next tick")
	}
}

// Zero is the interesting count, not the large one: a deployment can only tell
// a sweep that is working from a process that is merely running if the quiet
// runs say so too.
func TestASweepThatRemovedNothingStillReports(t *testing.T) {
	log := &syncBuffer{}
	identity.NewSweeper(&fakeSweepStore{removed: 0}, log.logger(), sweepConfig(t, nil)).
		Sweep(context.Background())

	line := findLine(t, log.lines(t), "swept expired identity rows")
	for _, kind := range []string{"ceremonies", "sessions", "tickets"} {
		got, ok := line[kind]
		if !ok {
			t.Errorf("the sweep line has no count for %s: %v", kind, line)
			continue
		}
		if got != float64(0) {
			t.Errorf("%s = %v, want 0", kind, got)
		}
	}
}

// Housekeeping is not worth an outage, and a sweep that cannot reach the
// database has nothing to retry harder about. It logs and the next one runs.
func TestAFailingSweepDoesNotPreventTheNext(t *testing.T) {
	boom := errors.New("the database is unreachable")
	fake := &fakeSweepStore{removed: 2, failFirstSweep: boom, started: make(chan int, 8)}
	log := &syncBuffer{}

	sweeper := identity.NewSweeper(fake, log.logger(),
		sweepConfig(t, map[string]string{"WIMM_SWEEP_INTERVAL": "10ms"}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		sweeper.Run(ctx)
		close(done)
	}()

	waitForSweep(t, fake, 2)
	cancel()
	<-done

	lines := log.lines(t)
	if line := findLine(t, lines, "sweeping failed"); line["error"] != boom.Error() {
		t.Errorf("the failure line is %v, want the database error", line)
	}

	var swept int
	for _, line := range lines {
		if line["msg"] == "swept expired identity rows" {
			swept++
		}
	}
	if swept < 2 {
		t.Fatalf("%d sweeps were reported, want a later one after the failure", swept)
	}
	// The later sweep did the work the failing one could not.
	last := lines[len(lines)-1]
	if last["msg"] != "swept expired identity rows" || last["ceremonies"] != float64(2) {
		t.Errorf("the last line is %v, want a sweep that removed rows normally", last)
	}
}

func waitForSweep(t *testing.T, fake *fakeSweepStore, n int) {
	t.Helper()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case got := <-fake.started:
			if got >= n {
				return
			}
		case <-deadline:
			t.Fatalf("sweep %d did not start within five seconds", n)
		}
	}
}

func findLine(t *testing.T, lines []map[string]any, msg string) map[string]any {
	t.Helper()

	for _, line := range lines {
		if line["msg"] == msg {
			return line
		}
	}
	t.Fatalf("nothing logged %q; the log holds %v", msg, lines)
	return nil
}

// A log line is the one part of this that leaves the process, so it carries
// counts and nothing else. A row identifier would name a member's session in
// whatever collects these logs; a value hash is the stored half of a bearer
// credential and has no business in a line about housekeeping.
func TestTheSweepLineCarriesNoIdentifierOrSecret(t *testing.T) {
	_, db := newService(t)
	ctx := context.Background()

	m, err := db.CreateMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	const secret = "a-value-hash-nothing-should-log"
	var sessionID string
	err = db.Pool().QueryRow(ctx, `
		insert into sessions (member_id, value_hash, expires_at)
		values ($1, $2, now() - make_interval(secs => 60))
		returning id`, m.ID, []byte(secret)).Scan(&sessionID)
	if err != nil {
		t.Fatalf("seeding an expired session: %v", err)
	}

	log := &syncBuffer{}
	r := identity.NewSweeper(db, log.logger(), sweepConfig(t, nil)).Sweep(ctx)
	if r.Failed() {
		t.Fatalf("the sweep failed: %+v", r)
	}
	if r.Sessions.Removed != 1 {
		t.Fatalf("removed %d sessions, want the one that had expired", r.Sessions.Removed)
	}

	written := log.String()
	for _, banned := range []struct{ what, value string }{
		{"the session's identifier", sessionID},
		{"the session's value hash", secret},
		{"the member's identifier", m.ID},
		{"the member's address", m.Email},
	} {
		if strings.Contains(written, banned.value) {
			t.Errorf("the sweep logged %s: %s", banned.what, written)
		}
	}
}

// The sweep is allowed to be late, to skip a run, or to fail, because it
// decides nothing: every lifetime is already checked in SQL when the row is
// read. This is that claim as a test — the same three requests, refused the
// same way either side of a sweep that removes all three rows.
func TestTheSweepChangesNoOutcome(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	// An enrolment part way through, whose ticket then expires.
	_, link, err := e.svc.RegisterMember(ctx, "grace@example.com", "Grace", "Hopper")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	ticket, _, err := e.svc.RedeemEnrolmentLink(ctx, linkValue(t, link))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}

	// A sign-in begun and never finished.
	ceremony, err := e.svc.BeginSignIn(ctx, "")
	if err != nil {
		t.Fatalf("beginning sign-in: %v", err)
	}

	expire(t, ctx, e.db, `update sessions set expires_at = now() - interval '1 second'`)
	expire(t, ctx, e.db, `update enrolment_tickets set expires_at = now() - interval '1 second'`)
	expire(t, ctx, e.db, `update ceremony_challenges set expires_at = now() - interval '1 second'`)

	refusals := func() [3]string {
		_, sessionErr := e.svc.MemberForSession(ctx, e.session.Value)
		_, _, enrolmentErr := e.svc.BeginEnrolment(ctx, ticket.Value)
		// The ceremony is consumed before anything parses the credential, so
		// the body never has to be a real one.
		_, _, _, signInErr := e.svc.FinishSignIn(ctx, ceremony.ID, "{}")
		return [3]string{errText(sessionErr), errText(enrolmentErr), errText(signInErr)}
	}

	before := refusals()
	for i, got := range before {
		if got == "" {
			t.Fatalf("request %d was accepted before the sweep, with nothing left to compare", i)
		}
	}

	r := identity.NewSweeper(e.db, (&syncBuffer{}).logger(), sweepConfig(t, nil)).Sweep(ctx)
	if r.Failed() {
		t.Fatalf("the sweep failed: %+v", r)
	}
	if r.Ceremonies.Removed < 1 || r.Sessions.Removed < 1 || r.Tickets.Removed < 1 {
		t.Fatalf("the sweep removed %d challenges, %d sessions and %d tickets; it was meant to remove all three",
			r.Ceremonies.Removed, r.Sessions.Removed, r.Tickets.Removed)
	}

	if after := refusals(); after != before {
		t.Errorf("after the sweep the same requests answer %v, before they answered %v", after, before)
	}
}

func expire(t *testing.T, ctx context.Context, db *store.DB, statement string) {
	t.Helper()

	tag, err := db.Pool().Exec(ctx, statement)
	if err != nil {
		t.Fatalf("running %q: %v", statement, err)
	}
	if tag.RowsAffected() == 0 {
		t.Fatalf("%q changed nothing: the row it was meant to expire is not there", statement)
	}
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// The fourth statement is independent of the other three, which is the whole
// reason each table gets its own delete and its own error.
func TestTheOtherThreeStatementsRunWhenTheFourthFails(t *testing.T) {
	boom := errors.New("pending_bank_connections is unreachable")
	fake := &fakeSweepStore{removed: 2, failPending: boom}

	var logged syncBuffer
	sweeper := identity.NewSweeper(fake, logged.logger(), sweepConfig(t, nil))

	result := sweeper.Sweep(context.Background())

	if !errors.Is(result.PendingBankConnections.Err, boom) {
		t.Errorf("PendingBankConnections.Err = %v, want the failure", result.PendingBankConnections.Err)
	}
	if !result.Failed() {
		t.Error("a failed fourth statement did not make the sweep report a failure")
	}
	for name, swept := range map[string]identity.Swept{
		"ceremonies": result.Ceremonies,
		"sessions":   result.Sessions,
		"tickets":    result.Tickets,
	} {
		if swept.Err != nil {
			t.Errorf("%s failed alongside the fourth statement: %v", name, swept.Err)
		}
		if swept.Removed != 2 {
			t.Errorf("%s removed %d rows, want 2", name, swept.Removed)
		}
	}
}

// The log line carries a count per kind, including the new one, and including
// when every count is zero.
func TestTheSweepLineCarriesTheFourthCount(t *testing.T) {
	fake := &fakeSweepStore{}

	var logged syncBuffer
	sweeper := identity.NewSweeper(fake, logged.logger(), sweepConfig(t, nil))
	sweeper.Sweep(context.Background())

	line := logged.String()
	for _, key := range []string{"ceremonies", "sessions", "tickets", "pending_bank_connections"} {
		if !strings.Contains(line, `"`+key+`":0`) {
			t.Errorf("the sweep line does not carry %s as zero: %s", key, line)
		}
	}
	if fake.pendingCount() != 1 {
		t.Errorf("the fourth statement ran %d times, want 1", fake.pendingCount())
	}
}
