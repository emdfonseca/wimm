## Why

Identity rows that are dead are never removed. A WebAuthn challenge is a row, and
`/signin` mints one on every page view, so the table grows with unauthenticated
traffic rather than with use — ten page loads that touched nothing added ten
rows. Expired sessions and unused enrolment tickets accumulate the same way.
`DeleteExpiredCeremonies` was written for exactly this and has never had a
caller.

None of it is a correctness problem: an expired row cannot be consumed, because
every lifetime is checked in SQL. It is a storage problem that grows without
bound and has no upper limit anyone chose.

## What Changes

- **`wimmd` sweeps expired identity rows on an interval.** One goroutine beside
  the listeners, shutting down with them. It deletes expired ceremony
  challenges, expired and revoked sessions, and expired enrolment tickets.
- **The sweep is observable.** One log line per run with a count per table, so
  a deployment can tell the difference between working and merely running.
- **Interval and retention are configuration**, with defaults a household
  instance never has to think about.
- **A revoked session is kept for a retention window, not deleted at once.**
  Revocation is the whole recovery story for a lost device, and a row deleted
  the instant it is revoked cannot answer "was this actually cut off". The
  window is what makes that answerable without keeping rows forever.
- **Nothing about what a member or the operator can do changes.** No new
  surface, no new response, no behaviour a person can observe.

Not in this change: the per-page-view minting that makes `ceremony_challenges`
the fastest-growing of the three. Removing it means the sign-in page cannot hold
its challenge ready, so the click handler must `await` a fetch before calling
`navigator.credentials.get()`, and an intervening `await` can lose the transient
user activation Safari requires. That is a behaviour change to the one control
this product has, and it needs its own change and its own device testing. The
sweeper is correct either way, and bounds the growth in the meantime.

Also not in this change: revoking a session or removing a credential on demand,
which `identity/passkey-sign-in` still does not offer.

## Capabilities

### New Capabilities

- `identity/expired-row-sweeping`: expired ceremony challenges, sessions and
  enrolment tickets are removed on an interval, and the removal is reported.

### Modified Capabilities

None. The three identity capabilities describe what a person can do, and this
changes none of it — the rows being removed are already unusable, so no
scenario's outcome differs.

## Impact

**Code.** `apps/wimm/internal/store` gains delete queries for expired sessions
and expired tickets beside the ceremony one that already exists.
`apps/wimm/internal/identity` gains the sweep itself. `cmd/wimmd/main.go` starts
it alongside the listeners, and `internal/config` carries the interval and the
retention window.

**Data.** No schema change. `sessions.revoked_at` and every `expires_at` column
the sweep reads are already there, and `ceremony_challenges` already has an
index on `expires_at`.

**Operations.** A deployment that has been running for a while will see a large
first sweep. Nothing depends on those rows, but the delete is worth knowing
about rather than discovering in a lock wait.

**Dependencies.** None. No new package, no new service.
