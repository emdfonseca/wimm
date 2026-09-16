## 1. Data

- [x] 1.1 Add a goose migration indexing `sessions` and `enrolment_tickets` by the column the sweep scans, since the sweep is the first query to read either by time; verify an up/down round trip returns the schema to its previous state and `just check apps/wimm` passes.

## 2. Store

- [x] 2.1 Make `DeleteExpiredCeremonies` delete in bounded batches, repeating until a batch comes back short, and return the total; verify a test seeds more rows than one batch and asserts every expired row goes and the count is right.
- [x] 2.2 Add the same batched delete for expired enrolment tickets; verify a test asserts an expired unused ticket is removed and one still inside its lifetime is not.
- [x] 2.3 Add the same for sessions: expired ones go at once, revoked ones only once they are older than the retention window; verify a test asserts a just-revoked session survives, a long-revoked one does not, and a live one is untouched.
- [x] 2.4 Assert every delete compares against database time, not the process clock; verify by reading `expires_at - now()` back in SQL, the way the session-lifetime test already does, and by `forbidigo` continuing to refuse `time.Now` outside tests.

## 3. Configuration

- [x] 3.1 Load the sweep interval and the revoked-session retention window, both with defaults a household instance never has to set, and refuse to start on a non-positive value the way the other lifetimes already do; verify tests cover the defaults, an override, and the refusal.

## 4. The sweep

- [x] 4.1 Implement one sweep: three deletes, each failing independently, returning a per-table count and a per-table error; verify a test drives a failure in one table and asserts the other two still ran and reported.
- [x] 4.2 Run it on a ticker from `cmd/wimmd/main.go` beside the listeners, sharing the signal context; verify a test asserts the loop exits when the context is cancelled and does not wait out the interval.
- [x] 4.3 Log one line per sweep with a count per kind, including when every count is zero; verify a test asserts a sweep that removed nothing still logs, and that no row identifier or secret value appears in the line.
- [x] 4.4 Assert a failing sweep neither stops the service nor prevents the next one; verify a test makes one sweep fail and asserts a later sweep runs normally.

## 5. Verification

- [x] 5.1 Confirm the sweep does not change any outcome: with an expired challenge, an expired session and an expired ticket all present, assert every request that would have been refused before the sweep is refused identically after it.
- [x] 5.2 Run the stack, load `/signin` repeatedly to mint challenges, and confirm a sweep removes them while a live session and an in-flight enrolment survive; verify against a running instance, not a unit test.
- [x] 5.3 Confirm `just pen-manifest verify` reports nothing added or changed, since this change touches no design file.
