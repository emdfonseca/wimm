# 0017 · Sweeping expired identity rows

## Status

Accepted

## Context

Three identity tables accumulate rows that can no longer be used.
`ceremony_challenges` grows fastest: `/signin` mints one per page view, so it
tracks unauthenticated traffic rather than use. Expired sessions and unused
enrolment tickets accumulate the same way, and nothing has ever deleted a
session — `RevokeSession` sets a column.

None of it is a correctness problem. Every lifetime is compared against
database time in SQL (ADR 0016), so an expired row is refused by the check that
reads it whether or not anything removes it afterwards. It is storage growing
without an upper bound anyone chose.

That makes the sweep free to be late, to skip a run, or to fail. What it is not
free to do is delay shutdown, take the service down with it, or make a question
about a revoked session unanswerable.

`wimmd` runs exactly one thing today: `server.Run` with its listeners, under a
signal context. There is no scheduler and no job runner.

Options for where the sweep runs: a goroutine in `wimmd`; a `wimmctl sweep`
subcommand driven by system cron; `pg_cron`.

## Decision

`wimmd` sweeps on a ticker from one goroutine started beside the listeners,
sharing their signal context. One pass is three independent statements, one per
table, each reporting its own count and its own error; a failure in one does not
skip the others, because no invariant spans them. Each delete runs in bounded
batches of `store.DeleteBatchSize`, repeating until a batch comes back short, so
the lock is a function of the batch rather than of the backlog. Every comparison
is `now()` in SQL, enforced by `forbidigo` banning `time.Now` outside tests; the
interval between sweeps is a Go ticker, which is not a lifetime.

An expired session is removed at once. A revoked one is kept for
`WIMM_REVOKED_SESSION_RETENTION` after `revoked_at`, because revocation is the
whole recovery story for a lost device and a row deleted on revocation cannot
answer "was this actually cut off". Both that window and `WIMM_SWEEP_INTERVAL`
are configuration with defaults a household instance never sets, refused at
startup when non-positive like every other lifetime.

Each sweep logs one line carrying a count per kind, including when every count
is zero, and nothing else: no row identifier, and nothing that is stored hashed.

Sessions and enrolment tickets gain indexes on the columns the sweep scans —
`expires_at` on both, and a partial index on `sessions.revoked_at` for the
branch that reads it — built `CONCURRENTLY` in a `NO TRANSACTION` migration.
`ceremony_challenges` already had its own.

## Consequences

Storage is bounded by what is live on an instance nobody administers, without a
second thing needing to be installed: a cron entry nobody added would reproduce
the defect this removes, silently, and `pg_cron` would put the retention window
somewhere no one reading Go would find it.

Two instances sweeping at once is harmless — the deletes are idempotent and
target rows disjoint by time — so there is no leader election and no advisory
lock. `wimmd` still holds no request-spanning state.

The first sweep on an established instance removes a large backlog and logs a
large count. The lock it takes is bounded by the batch size; the count is worth
seeing.

This bounds the growth without fixing its cause. `/signin` keeps minting a
challenge per page view, and removing that means the sign-in page can no longer
hold its challenge ready — the click handler would have to `await` a fetch
before `navigator.credentials.get()`, and an intervening `await` can lose the
transient user activation Safari requires. That is a change to the one control
this product has, and it needs its own device testing.

A revoked session is answerable for a week and not after. This is not an audit
trail, and an instance that needs one needs a different change. Whether expired
sessions should wait out the same window is open: they ended on their own, where
a revoked one was ended by someone, and that difference is worth revisiting when
revocation becomes something a member can do.

One goroutine is not a scheduler. The next background task in `wimmd` should
arrive as its own goroutine rather than as a framework this one gets folded
into.
