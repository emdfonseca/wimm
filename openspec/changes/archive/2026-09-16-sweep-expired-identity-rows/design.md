## Context

See proposal.md · Why. The relevant current state is that three tables
accumulate dead rows and one cleanup function already exists with no caller:

```text
ceremony_challenges   ConsumeCeremony deletes on finish; abandoned rows stay.
                      /signin mints one per page view, so this grows with
                      unauthenticated traffic, not with use.
sessions              nothing deletes a session, ever. RevokeSession sets a
                      column.
enrolment_tickets     deleted when their link is enrolled; expired-unused
                      rows stay.
```

Every lifetime is already compared against database time in SQL (ADR 0016), so
correctness does not depend on this change at all. That is what makes the design
simple: the sweep can be late, can skip a run, or can fail, and nothing becomes
wrong — only larger.

`wimmd` today runs exactly one thing: `server.Run` with its listeners, cancelled
by a signal context. There is no scheduler, no job runner, and no background
work of any kind.

## Language

- **Sweep** — one pass over all three tables. Never *cleanup*, never *GC*.
- **Retention window** — how long a revoked session is kept after revocation.
  Distinct from a *lifetime*, which is how long a row is usable.
- **Expired** — past its `expires_at`. A row can be expired and still present;
  that is precisely the state this change removes.

## Goals / Non-Goals

**Goals:**

- Storage bounded by what is live, on an instance nobody administers.
- A failed sweep is a log line, never an outage.
- The first sweep on an instance that has been running for months does not
  block anything else.

**Non-Goals:**

- Removing the per-page-view minting. Stated in the proposal; it changes the
  sign-in control's behaviour and needs device testing.
- Revoking sessions or credentials on demand. Still absent from
  `identity/passkey-sign-in`.
- A general job runner. One goroutine is not a scheduler, and pretending
  otherwise invites the next background task to arrive as a framework.

## Decisions

### A goroutine in wimmd, not a cron or a separate binary

The sweep starts beside the listeners and stops with them, sharing the signal
context `server.Run` already uses.

*Alternative:* a `wimmctl sweep` subcommand driven by system cron. Rejected —
it makes correct operation depend on a second thing being installed, on a
self-hosted box whose operator is a household member. A deployment that forgets
the cron entry gets the bug this change is fixing, silently.

*Alternative:* `pg_cron`. Rejected — it puts application policy in the database
and makes the retention window invisible to anyone reading Go.

Two instances both sweeping is harmless: the deletes are idempotent and target
disjoint-by-time rows. No leader election, no advisory lock.

### One pass, three statements, failures isolated per table

Each table is its own statement, and a failure on one does not skip the others.
A sweep reports per-table counts and a per-table error.

The alternative — one transaction over all three — buys atomicity nothing wants.
There is no invariant spanning these tables that a partial sweep could break.

### Deletes are batched, and the first sweep is the reason

An instance that has been running for months has a large backlog, and a single
unbounded `delete from ceremony_challenges where expires_at <= now()` takes a
lock proportional to it. The delete runs in bounded batches, repeating until a
batch comes back short.

```sql
delete from ceremony_challenges
where id in (
  select id from ceremony_challenges where expires_at <= now() limit $1
)
```

`ceremony_challenges` already has an index on `expires_at`. Sessions and tickets
do not, and the sweep is the first query to scan them by time — that is a
migration this change carries, not an optimisation to defer.

### A revoked session is kept for a retention window

Revocation is the whole recovery story for a lost device, and deleting the row
the moment it is revoked makes "was this cut off" unanswerable. It is kept for a
window after `revoked_at`, then removed.

The window is short by default. This is not an audit log — it is the difference
between a question you can answer this week and one you cannot answer at all.
An instance that wants a real audit trail needs a different change.

### Everything is compared against database time

`now()` in SQL, not the process clock, for the same reason every other lifetime
is (ADR 0016) — and enforced the same way, by `forbidigo` banning `time.Now`
outside tests. The interval between sweeps is a Go ticker, which is not a
lifetime and does not need database time.

## Risks / Trade-offs

- **A large first sweep on an established instance.** → Batched, so the lock is
  bounded by batch size rather than backlog. The first run logs a large count;
  that is worth seeing.
- **A sweep can delete a row a request is about to read.** → The request would
  have refused it anyway: both use the same `expires_at` comparison, and the
  loser of that race gets the same answer either way.
- **The interval is a guess.** → Configurable, and wrong in the safe direction:
  too slow only means more dead rows, which is today's behaviour.
- **This bounds growth without fixing its cause.** Stated plainly because it is
  the honest position: `/signin` will keep minting a challenge per page view,
  and the sweep will keep removing them. The sweeper is worth having regardless,
  and is not an argument against fixing the cause.

## Migration Plan

Two indexes, on `sessions` and `enrolment_tickets`, by the column the sweep
scans. No data migration; nothing to backfill. Rollback is removing the
goroutine — the rows simply accumulate again.

## Open Questions

- Whether the retention window should apply to expired sessions as well as
  revoked ones. They differ: an expired session ended on its own, a revoked one
  was ended by someone. Treating them alike is simpler; treating them
  differently is what an audit trail would want. Decided as written — expired
  goes at once, revoked waits — and worth revisiting when revocation becomes a
  feature a member can use.
