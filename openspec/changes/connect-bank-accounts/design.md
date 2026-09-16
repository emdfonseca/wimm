## Context

See `proposal.md` — Why.

### What was checked, and what was not

The first draft of this design was written from recollection of Enable Banking's
API. It has since been reconciled against their published documentation, and
five decisions changed as a result. Stating which is which, because a claim that
reads as verified and is not is worse than no claim.

**Verified against the API reference and FAQ:** the AIS endpoint set; that
`access` scopes accounts, balances and transactions separately; that
`maximum_consent_validity` is returned per bank in `GET /aspsps` and is 180 days
for the majority; that renewal is a fresh authorisation with no refresh
mechanism and an expired session returns `EXPIRED_SESSION` with 401; that
`POST /sessions` returns account details **shown only once**; that account `uid`
is per-session while `identification_hash` matches an account across sessions;
that ASPSP rate limits of about four fetches a day apply to background fetching
and do **not** apply when PSU headers indicate the member is present, with a 429
carrying `ASPSP_RATE_LIMIT_EXCEEDED`; and that some banks let the member narrow
accounts in their own consent screen while others hand over everything.

**Not verified, and each is a task rather than an assumption:** whether the
target country's banks are covered; whether the sandbox accepts a localhost
redirect; and whether Enable Banking's free Restricted Production tier, which
covers accounts the application owner personally links, extends to a second
household member's bank. That last one decides whether wimm works for a
household or only for one person, and no documentation answers it.

### What the repo already settles

**There is no household table.** One instance serves one household (README), so
"household-visible" is not a scoping rule to implement — it is the absence of
one. `connected_by` exists to answer "whose consent is holding this open", which
is a different question and one the spec asks.

**Identity already solved the bearer-value-in-a-URL problem.** ADR 0016 issues a
one-time value, stores only its hash, exchanges it server-side on first load and
redirects to an address that does not contain it. The consent return is the same
problem with a different issuer.

**There is already a sweeper.** ADR 0017 runs independent statements on a ticker
in bounded batches against database time. Abandoned consent attempts are a
fourth statement, not a new mechanism.

## Language

Binding on spec text, interface copy and code identifiers.

- **Gateway** — the third-party open-banking service wimm reaches banks
  through. Never "provider", "aggregator" or "PSD2 provider".
- **Bank** — the institution a member holds money at. Never "ASPSP" outside a
  gateway adapter's own files, where it is the wire vocabulary and stays there.
- **Connection** — one household's live read access to one bank. Enable Banking
  calls this a *session* and GoCardless a *requisition*; wimm calls it neither,
  because **session** already means a signed-in member's session (ADR 0016) and
  reusing it would make "expired session" ambiguous between "sign in again" and
  "confirm again at your bank".
- **Hand-off** — sending the member out to their bank to consent. Never a
  "link", because **link** already means an enrolment link (ADR 0016).
- **Account** — one account at a connected bank.
- **Shared** — an account the connecting member chose to put in front of the
  household. The opposite is *not shared*, never "hidden" or "private": nothing
  is being concealed from anyone who could otherwise see it.
- **Balance** — a signed amount in minor units with a currency and the time it
  was read. A balance without a read time is not a balance in this system.
- **Reading** — one fetch of balances from the gateway. "Refresh" is the verb a
  member sees; **reading** is the noun in the model.
- **Restore** — confirming again at the bank after access has run out. Never
  "renew" or "refresh": refresh is what happens to balances, and open banking
  has no refresh of consent.
- **Sealed** — a stored value that could reach a bank, held encrypted. The Go
  type is `banking.Sealed`, and a plaintext one never has a name in a struct.

## Goals / Non-Goals

**Goals**

- One seam, at a package boundary a linter enforces, between wimm and any
  gateway.
- A port validated against a second gateway on paper before it is written.
- Balances a member can trust: read when they arrive, carrying their age.
- A member decides what the household sees, separately from what the bank
  granted.
- A connection that can be restored when its access runs out, because every
  connection reaches that point.
- Nothing stored that opens a bank to whoever reads the database.

**Non-Goals**

- Implementing a second gateway. The paper validation is the deliverable.
- Scheduled background syncing. That is the case the rate limits actually bite,
  and it is change 3.
- A port method for transactions. Change 2 adds it; this design only commits to
  not painting it into a corner.
- Currency conversion, of any kind, anywhere.

## Decisions

### 1. The seam is `internal/banking`, and a linter holds it

```text
apps/wimm/internal/banking/                  the port: wimm's nouns only
apps/wimm/internal/banking/enablebanking/    the adapter: wire vocabulary
apps/wimm/internal/banking/bankingtest/      an in-memory Gateway
```

Nothing outside `internal/banking/...` may import `enablebanking`, enforced as
an import restriction in the lint configuration that already bans `time.Now`
(ADR 0017). Coupling a check can catch is the only kind still absent in six
months.

### 2. The port is validated against a second gateway on paper

```text
port method            Enable Banking            GoCardless Bank Account Data
---------------------  ------------------------  ----------------------------
Banks(country)         GET /aspsps?country=      GET /institutions/?country=
                       carries maximum_consent   carries max_historical_days
                       _validity per bank
BeginConnection        POST /auth -> url +       agreement + requisition
                       authorization_id          -> link + id
CompleteConnection     POST /sessions {code}     GET /requisitions/{id}
                       -> session_id + accounts  -> account ids, then fetch
                       (details shown once)
Balances(account)      GET /accounts/{uid}/      GET /accounts/{id}/balances/
                       balances
EndConnection          DELETE /sessions/{id}     DELETE /requisitions/{id}
Transactions (ch. 2)   GET .../transactions      GET .../transactions/
                       continuation_key          date window
```

Two findings came out of the exercise and both changed the signature.

**`CompleteConnection` takes both the row written at begin and the callback.**
Enable Banking's truth is in the callback (a `code`); GoCardless's is in the row
(the requisition id), and its callback carries almost nothing.

```go
CompleteConnection(ctx, pending PendingConnection, callback Callback) (Connection, []Account, error)
```

**It returns accounts**, even though GoCardless needs a second round-trip to
produce them. The extra fetch lives in the adapter.

*Alternative:* write the port from the first gateway and refactor when a second
arrives. Rejected — the refactor lands exactly when there is production data
keyed on the first gateway's identifiers.

### 3. Redirect-based gateways only, stated rather than designed around

Enable Banking, GoCardless, TrueLayer, Yapily and Tink all redirect. Plaid's
Link step is client-side and does not fit `BeginConnection -> URL`. Naming the
constraint beats a variant nothing exercises; the day one is wanted, the return
type grows a case and the compiler finds every site.

### 4. Account identity is the gateway's cross-session hash, never its uid

**Corrected from the first draft.** Enable Banking's account `uid` is issued per
session and changes on every restore; `identification_hash` is what matches an
account across sessions. A schema keyed on `uid` would lose every sharing choice
the moment a member restored a connection, which at 180 days is certain.

```text
bank_accounts.gateway_ref   identification_hash   stable, the key
bank_accounts.gateway_uid   uid                   sealed, rewritten per restore
```

Gateway values are never primary keys. Every table has a wimm `uuid` and carries
`(gateway, gateway_ref)` beside it, unique together.

### 5. Everything the bank returns is stored; a member chooses what is shared

**Corrected from the first draft**, which proposed storing only shared accounts
and re-reading the list when a member wanted to change their mind. That is not
possible: `POST /sessions` returns account details **once**, and there is no
endpoint that lists them again. Re-reading would mean a fresh consent, so a
member could not add an account without going back to their bank.

So every account the session returns is stored, with `shared` defaulting to
false, and the chooser reads wimm's own rows. An unshared account holds a name
and an identifier and **no balance is ever read for it**, so wimm knows an
account exists and does not know what is in it.

*Alternative:* discard unshared accounts entirely. Rejected for the reason
above — it makes the choice a one-way door in the wrong direction.

### 6. The chooser is not a consent step, and its copy must not pretend otherwise

Some banks let the member narrow accounts in their own consent screen; others
grant everything. wimm cannot know which happened, and cannot narrow what the
bank granted. By the time the chooser appears, access exists either way.

It therefore asks one question — which of these should the household see — and
never implies it restricts the bank. **Nothing is preselected**, because the
whole reason the step exists is that clicking through should not share a
personal account; and sharing everything is one action, so the fast case stays
fast. Finishing with nothing chosen is refused, since a bank sharing nothing
shows the household nothing and looks broken.

### 7. Balances are read when a member arrives, and on request

**Corrected from the first draft**, which made refresh manual-only on the belief
that balance reads are tightly rate-limited. They are not, in the case that
matters: the roughly four-a-day ASPSP cap applies to background fetching, and
does not apply when PSU headers indicate the member is present. Reading on
arrival is the unthrottled case.

So: read on arrival, a control to read again, and no background reading at all
in this change. A 429 with `ASPSP_RATE_LIMIT_EXCEEDED` is still possible and is
handled — the previous readings stay on screen with their original times, which
is the one thing that must not be lost. The documented remedy of retrying after
six hours belongs to background fetching and therefore to change 3.

No balance is read for an unshared account, at arrival or ever.

### 8. Restoring is the whole flow again, because open banking has no renewal

Consent is valid until a date wimm sets at authorisation, capped by that bank's
`maximum_consent_validity`, 180 days at most banks. There is no refresh: renewal
is a fresh authorisation. A session can also expire early, surfacing as
`EXPIRED_SESSION` with 401, which is treated identically — the member's
experience of "my bank stopped updating" is the same either way.

wimm requests the bank's maximum, and shows the member the resulting date before
the hand-off, because it is that bank's limit and not a wimm policy.

Restoring re-enters the flow at the hand-off, skipping the picker: the bank is
already known. On return, accounts are matched by `identification_hash`, so
`shared` carries forward. An account newly offered arrives unshared and is
surfaced as something to choose; an account no longer offered stops appearing
and the member is told which.

### 9. Nothing stored opens a bank on its own

wimm never receives a member's banking credentials, and never holds the bank's
OAuth tokens — the gateway holds those and refreshes them internally. What wimm
does hold is the gateway `session_id`, which reads that household's accounts
until the grant runs out, and the per-account `uid`s used with it. Those are
live credentials and a column is not where they belong.

**They are sealed with AES-256-GCM before they reach the database.** The key is
read at startup from the path in `WIMM_BANKING_ENCRYPTION_KEY`, never from an
environment variable holding the material. The connection's wimm `uuid` is the
additional authenticated data, so a ciphertext lifted from one row cannot be
replayed into another. A key id is stored beside each value so a key can be
rotated without a flag day.

*Alternative:* `pgcrypto`, or full-disk encryption. Both rejected for the same
reason: the threat is a copy of the data — a dump, a backup, a replica, a stolen
volume — and in each of those the key travels with the rows or the SQL carrying
it reaches the query log. Encrypting in Go keeps the key out of the database
entirely.

**A sealed value has a type that cannot be printed.** `banking.Sealed`
implements `String`, `MarshalJSON` and `slog.LogValuer` to return a redaction,
and plaintext exists only as a local inside the adapter for the length of one
call. This is the mechanical half of "never logged"; the lint rule is the other.

**The application signing key is the master credential** and is treated the same
way: a path, not a value, so it stays out of process listings, crash dumps and
container inspection; refused at startup if its file is group or world readable.

**Sealed values are destroyed when access ends.** Disconnecting, and a
connection reaching the end of its grant, both clear them rather than leaving
them behind. The connection row survives for the audit question; what could open
the bank does not.

### 10. Failures are wimm's taxonomy

```text
ErrBankUnavailable     the bank itself is down or refused the gateway
ErrGatewayUnavailable  wimm could not reach the gateway at all
ErrConsentDeclined     the member said no at the bank
ErrConsentExpired      EXPIRED_SESSION, or the grant's date has passed
ErrNoAccounts          consent granted, nothing readable came back
ErrRateLimited         ASPSP_RATE_LIMIT_EXCEEDED, carrying a retry-after
```

`docs/design/surfaces.md` routes a page banner, an inline alert and a field
error off what kind of failure it is, and no two gateways agree on status codes.
The mapping is table-driven and tested per adapter.

### 11. `bankingtest` is part of the port, not a test helper

An in-memory `Gateway` the whole service runs against, following the existing
`identity/authenticatortest` pattern. It keeps `just check` off the network, and
it is a second implementation of the interface written the same week as the
first, which is the cheapest pressure test for whether the interface abstracts
anything.

### 12. Money is `int64` minor units plus an ISO 4217 code

Never a float, never the gateway's decimal string, never a bare number without
its currency. Totals are per currency; mixed currencies are never summed. wimm
holds no rates, and inventing one would invent the number a household trusts
most.

### 13. Schema

Owned by `wimmd`, goose, one migration. Every lifetime compared against database
time.

```sql
bank_connections
  id, gateway, gateway_ref_sealed, key_id, bank_id, bank_name, bank_logo_url,
  connected_by -> members(id), created_at, consent_expires_at, expired_at,
  disconnected_at

pending_bank_connections
  id, state_hash, gateway, gateway_ref, bank_id, restores -> bank_connections(id),
  started_by -> members(id), created_at, expires_at, consumed_at

bank_accounts
  id, connection_id -> bank_connections(id) on delete cascade,
  gateway_ref, gateway_uid_sealed, key_id,
  name, number_suffix, currency, shared,
  balance_minor, balance_read_at, raw jsonb,
  unique (connection_id, gateway_ref)
```

`gateway_ref` on an account is the cross-session hash and is not sealed: it
identifies an account and opens nothing. `gateway_uid_sealed` and the
connection's `gateway_ref_sealed` are the session-scoped values from decision 9.

`number_suffix` holds the last few characters only. The full number is never
stored: the spec requires telling two accounts apart and nothing more, and an
IBAN at rest is a liability with no use here.

`raw` keeps the gateway's account payload, bounded at one row per account and
overwritten on each reading, with anything sealed removed before it is written.
It is the difference between mapping a field we missed and going back to the
member for a fresh consent — which, given details are returned once, is the only
other way to get it.

`pending_bank_connections.restores` is null for a first connection and names the
connection being restored otherwise, which is how the return knows to match by
hash rather than create.

Disconnection sets `disconnected_at`, clears the sealed values, and the accounts
cascade away. The connection row stays: "which bank did we read, who authorised
it, when did it end" is worth being able to answer.

`pending_bank_connections` gains the sweeper's fourth statement, with an
`expires_at` index built `CONCURRENTLY` in a `NO TRANSACTION` migration.

### 14. Configuration

```text
WIMM_BANKING_GATEWAY              which adapter; unset disables connecting
WIMM_ENABLEBANKING_APPLICATION_ID the registered application
WIMM_ENABLEBANKING_PRIVATE_KEY    path to the signing key, never its contents
WIMM_BANKING_ENCRYPTION_KEY       path to the sealing key, never its contents
WIMM_BALANCE_STALE_AFTER          default 24h
```

`wimmd` refuses to start when a gateway is named and any of its credentials are
absent, or when either key file is group or world readable — the same shape of
rule ADR 0016 applies to the operator listener. Unset is supported: the accounts
screen shows its empty state and connecting is unavailable, so the change
deploys safely before anyone holds gateway credentials.

### 15. Surfaces

New Connect service `wimm.banking.v1` on the public listener, with protovalidate
constraints like the identity protos, carrying no gateway vocabulary and no
sealed value in any message. The browser never sees it: SvelteKit server routes
call Connect and hand the web client JSON (ADR 0001). The hand-off return lands
on a SvelteKit server route that exchanges and redirects, mirroring
`/enrol/[link]`.

## Risks / Trade-offs

**The free tier may not cover a household.** Restricted Production covers
accounts the application owner personally links. If a second member's bank does
not count, wimm either needs a contract per household or works for one person.
→ Ask Enable Banking before building. It is the one open item that can
invalidate the gateway choice rather than delay it.

**The port is validated on paper.** → The table covers change 2's method as well
as this change's, and `bankingtest` is a genuine second implementation. Residual
risk: gateway three is shaped unlike either, which no design prevents.

**Reading on arrival means a member's page load depends on their bank.** → Read
in the background of the request and render the previous readings immediately
if the bank is slow, so a bank having a bad day never blocks the screen.

**Every connection expires, most at 180 days.** → Restoring is in this change,
which is why it is in this change.

**wimm holds a live credential to a household's bank.** → Sealed with a key that
lives outside the database, destroyed when access ends, never logged, never in a
message, and typed so it cannot be printed by accident.

**A gateway outage would make disconnecting impossible if wimm waited on it.**
→ `EndConnection` failing does not block disconnection; wimm ends its own access
and tells the member to withdraw at the bank too.

## Migration Plan

Additive throughout. New tables, a new Connect service, new routes; nothing
existing changes shape. Deploying with `WIMM_BANKING_GATEWAY` unset is a no-op
beyond an empty accounts screen, so the schema and the gateway credentials can
land in separate deploys.

Rollback is the goose down migration. The data lost is connections, sharing
choices and balance readings. Reconnecting recovers the accounts, since identity
is the gateway's cross-session hash, but the sharing choices are gone and the
member makes them again.

## Open Questions

- **Whether the free tier covers a second member's bank.** Answered by Enable
  Banking, not by documentation. Task 1.
- **Which country and which banks.** Needed before the adapter is exercised.
  Task 1.
- **Whether the bank list needs grouping beyond search** — by popularity, or by
  recently used. Deferred safely: nothing in the schema or the port changes if
  grouping is added later.
