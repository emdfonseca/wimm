## Purpose

Rows that record something in flight — a challenge, a session, an enrolment
ticket — stop being usable the moment they expire, because every lifetime is
checked against database time when the row is read. This capability is what
removes them afterwards, so storage is bounded by what is live rather than by
everything that ever happened.

It is not user-facing and `stories.md` records a deliberate skip, so the
requirements below carry no `**Story**` line: there is no actor, and the
behaviour is the system's own.

## ADDED Requirements

### Requirement: Removing rows that can no longer be used
The system SHALL remove ceremony challenges, sessions and enrolment tickets
that can no longer be used, on a recurring interval, for as long as it is
running. Removal SHALL NOT change the outcome of any request: a row removed by
this was already refused by the check that reads it.

#### Scenario: An abandoned challenge is removed once it expires
- **WHEN** a sign-in or enrolment ceremony is begun and never finished, and its
  challenge has passed its expiry
- **THEN** the challenge is removed
- **AND** no request that would previously have been refused is now accepted

#### Scenario: An expired session is removed
- **WHEN** a session has passed its expiry
- **THEN** it is removed
- **AND** the browser holding its cookie is still asked to sign in, exactly as
  it was before the removal

#### Scenario: An unused enrolment ticket is removed
- **WHEN** an enrolment ticket has passed its expiry without being used
- **THEN** it is removed

#### Scenario: Live rows are left alone
- **WHEN** a sweep runs while a member has a live session, an enrolment is part
  way through, and a ticket is still within its lifetime
- **THEN** none of those rows is removed
- **AND** the member's session still works, and the enrolment can still be
  completed

### Requirement: Keeping a revoked session long enough to account for it
A session that was revoked SHALL be kept for a stated retention window after
revocation rather than removed immediately, so that the question "was this
session actually cut off" can be answered for that window. After the window it
SHALL be removed.

#### Scenario: A just-revoked session is still on record
- **WHEN** a member signs out and a sweep runs immediately afterwards
- **THEN** the revoked session is still recorded as revoked
- **AND** it still gives no access

#### Scenario: A long-revoked session is removed
- **WHEN** a session was revoked longer ago than the retention window
- **THEN** it is removed

### Requirement: Reporting what each sweep removed
Each sweep SHALL report how many rows it removed, per kind, so that an operator
reading the logs can tell a sweep that is working from a process that is merely
running. A sweep that removes nothing SHALL still report.

#### Scenario: A sweep that removed rows says so
- **WHEN** a sweep removes expired rows
- **THEN** it reports a count for each kind of row it removed

#### Scenario: A sweep that removed nothing still reports
- **WHEN** a sweep runs and finds nothing to remove
- **THEN** it reports zero rather than staying silent

### Requirement: Surviving its own failures and shutdown
A sweep that fails SHALL NOT stop the service or prevent later sweeps, because
storage housekeeping is not worth an outage. Sweeping SHALL stop when the
service is shutting down, and SHALL NOT delay shutdown.

#### Scenario: A failing sweep does not take the service with it
- **WHEN** a sweep cannot reach the database
- **THEN** the failure is reported
- **AND** the service keeps serving requests, and a later sweep runs normally

#### Scenario: Shutdown is not delayed
- **WHEN** the service is asked to shut down
- **THEN** it stops sweeping and shuts down without waiting for the next
  interval
