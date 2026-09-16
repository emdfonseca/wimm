package identity

import (
	"context"
	"log/slog"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
)

// SweepStore is the part of the store a sweep uses: three deletes, each its
// own statement against its own table.
type SweepStore interface {
	DeleteExpiredCeremonies(ctx context.Context) (int64, error)
	DeleteExpiredEnrolmentTickets(ctx context.Context) (int64, error)
	DeleteFinishedSessions(ctx context.Context, revokedRetention time.Duration) (int64, error)
}

// Swept is what one table's delete removed, and what stopped it.
//
// Both, not one or the other: a delete that failed part way through still
// removed the rows it got to, and reporting zero for them would describe a
// database that does not exist.
type Swept struct {
	Removed int64
	Err     error
}

// SweepResult is one pass over all three tables.
//
// There is no combined error. Each table fails on its own, because there is no
// invariant spanning them that a partial sweep could break — and a single error
// would make a sweep that cleared two tables and lost one look like a sweep
// that did nothing.
type SweepResult struct {
	Ceremonies Swept
	Sessions   Swept
	Tickets    Swept
}

// Failed reports whether any table's delete returned an error.
func (r SweepResult) Failed() bool {
	return r.Ceremonies.Err != nil || r.Sessions.Err != nil || r.Tickets.Err != nil
}

// Sweeper removes identity rows that can no longer be used.
//
// It is one goroutine beside the listeners rather than a scheduler. Nothing
// here is a job runner, and the next background task in this product should
// arrive as its own goroutine too rather than as a framework.
type Sweeper struct {
	store SweepStore
	log   *slog.Logger

	interval         time.Duration
	revokedRetention time.Duration
}

// NewSweeper builds the sweep from validated configuration.
func NewSweeper(store SweepStore, log *slog.Logger, cfg config.Config) *Sweeper {
	return &Sweeper{
		store:            store,
		log:              log,
		interval:         cfg.SweepInterval,
		revokedRetention: cfg.RevokedSessionRetention,
	}
}

// Sweep runs one pass and reports it. Each table is deleted from
// independently, so a failure on one does not skip the others.
//
// One log line per sweep, always, including when every count is zero: zero is
// how a deployment tells a sweep that is working from a process that is merely
// running. The line carries counts and nothing else — no row identifier, and
// nothing that was hashed for a reason.
func (s *Sweeper) Sweep(ctx context.Context) SweepResult {
	var r SweepResult

	r.Ceremonies.Removed, r.Ceremonies.Err = s.store.DeleteExpiredCeremonies(ctx)
	r.Sessions.Removed, r.Sessions.Err = s.store.DeleteFinishedSessions(ctx, s.revokedRetention)
	r.Tickets.Removed, r.Tickets.Err = s.store.DeleteExpiredEnrolmentTickets(ctx)

	s.log.InfoContext(ctx, "swept expired identity rows",
		"ceremonies", r.Ceremonies.Removed,
		"sessions", r.Sessions.Removed,
		"tickets", r.Tickets.Removed)

	for _, f := range []struct {
		kind  string
		swept Swept
	}{
		{"ceremonies", r.Ceremonies},
		{"sessions", r.Sessions},
		{"tickets", r.Tickets},
	} {
		if f.swept.Err != nil {
			s.log.ErrorContext(ctx, "sweeping failed", "kind", f.kind, "error", f.swept.Err)
		}
	}

	return r
}

// Run sweeps until ctx is cancelled, and returns as soon as it is.
//
// The first sweep happens immediately: an instance that has been accumulating
// rows since its last restart should not wait out an interval before the log
// says so. A failure is logged and the next sweep still runs — storage
// housekeeping is not worth an outage, and a sweep that cannot reach the
// database has nothing to retry harder about.
func (s *Sweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}

		s.Sweep(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
