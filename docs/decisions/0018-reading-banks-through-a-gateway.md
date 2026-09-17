# 0018 · Reading banks through a gateway

## Status

Accepted; sharing superseded by 0019

Accounts have owners and each member sees a level, in place of one `shared`
boolean. The gateway port, account identity, sealing, money and the chooser's
purpose are unchanged.

## Context

wimm signs a household in and shows it nothing. The ledger the identity work
protects does not exist, and the product's own question is "where is my money".

The honest first answer is not a form. `docs/design/surfaces.md` was written on
the assumption that money arrives from a bank rather than a keyboard: it names a
stepper for connecting an account, a badge for a transaction pending *at the
bank*, an inline alert for an account that has not synced in six days. Building
manual entry first would make those rules dead and would make every hard part —
external identifiers, deduplication, booked against pending — a retrofit onto a
model that never anticipated them.

It is also a dependency on a company. Gateways are repriced, acquired, withdrawn
from countries and occasionally shut. A household instance that cannot follow its
owner to another gateway has an expiry date.

**What was checked.** An earlier draft of this record was written from
recollection and presented an endpoint table as though it had been verified. It
had not. The decisions below are now reconciled against Enable Banking's
published API reference and FAQ, and five of them changed as a result. Three
things remain unverified and are tasks rather than assumptions: whether the
target country's banks are covered, whether the sandbox accepts a localhost
redirect, and whether the free Restricted Production tier — which covers
accounts the application owner personally links — extends to a second household
member's bank. That last one decides whether wimm serves a household or one
person, and no documentation answers it.

## Decision

**Enable Banking is the first gateway, and nothing outside one package knows its
name.** `apps/wimm/internal/banking` holds a `Gateway` port in wimm's own
vocabulary; `internal/banking/enablebanking` implements it; `internal/banking/
bankingtest` implements it in memory. An import restriction in the lint
configuration that already bans `time.Now` refuses any import of an adapter from
outside `internal/banking/...`. Coupling a check can catch is the only kind
still absent in six months.

**The port is validated against a second gateway on paper.** Writing a second
adapter to prove the abstraction is disproportionate; asserting it without
checking is how wrapper-shaped interfaces get written.

```text
port method            Enable Banking            GoCardless Bank Account Data
---------------------  ------------------------  ----------------------------
Banks(country)         GET /aspsps?country=      GET /institutions/?country=
                       maximum_consent_validity  max_historical_days
BeginConnection        POST /auth -> url +       agreement + requisition
                       authorization_id          -> link + id
CompleteConnection     POST /sessions {code}     GET /requisitions/{id}
                       -> session_id + accounts  -> account ids, then fetch
                       (details returned once)
Balances(account)      GET /accounts/{uid}/      GET /accounts/{id}/balances/
                       balances
EndConnection          DELETE /sessions/{id}     DELETE /requisitions/{id}
Transactions (ch. 2)   GET .../transactions      GET .../transactions/
                       continuation_key          date window
```

Two findings came out of the exercise and both changed the signature.
`CompleteConnection` takes **both** the row written at begin and the callback,
because Enable Banking's truth is in the callback and GoCardless's is in the row.
It **returns accounts**, even though GoCardless needs a second round-trip to
produce them; the extra fetch lives in the adapter rather than leaking the
cheaper gateway's shape upward.

**Redirect-based gateways only, stated rather than designed around.** Enable
Banking, GoCardless, TrueLayer, Yapily and Tink all redirect. Plaid's Link step
is client-side and does not fit `BeginConnection -> URL`. Naming the constraint
beats a variant nothing exercises.

**Account identity is the gateway's cross-session hash, never its per-session
id.** Enable Banking issues a new `uid` for every account on every
authorisation; `identification_hash` is what matches an account across them. A
schema keyed on `uid` would lose every sharing choice the moment a member
restored a connection, which at 180 days is certain rather than possible.

**Everything the bank returns is stored.** `POST /sessions` returns account
details **once**, and no endpoint lists them again. So an earlier plan — store
only the accounts a member wants seen, re-read the list if they change their
mind — is not implementable: changing your mind would mean going back to your
bank. Every account the session returns is stored, and **no balance is ever read
for an account nobody may see**, so wimm knows an account exists and does not
know what is in it.

*Who may see it is superseded by ADR 0019*, which replaces the `shared` boolean
with owners and a per-member level. The reason this decision exists — details
are returned once, so everything is stored — is unchanged, and so is the rule
that wimm never reads what nobody may see.

**The account chooser is not a consent step.** Some banks let the member narrow
accounts in their own consent screen; others hand over everything, and wimm
cannot know which happened or narrow what was granted. By the time the chooser
appears, access exists either way. It therefore asks one question — which of
these should the household see — and never implies it restricts the bank.
Nothing is preselected, because the whole reason the step exists is that
clicking through should not share a personal account; sharing everything is one
action, so the fast case stays fast.

Without this step, connecting a bank to share a joint account also shares every
personal account at that bank with everyone in the house. That is the failure
this decision exists to prevent.

**Balances are read when a member arrives, and again on request.** An earlier
draft made reading manual-only on the belief that balance reads are tightly
rate-limited. They are not in the case that matters: the roughly four-a-day ASPSP
cap applies to background fetching, and does not apply when PSU headers indicate
the member is present. There is no background reading at all in this change. A
429 carrying `ASPSP_RATE_LIMIT_EXCEEDED` is still possible and is handled by
leaving the previous readings on screen with their original times, which is the
one thing that must not be lost.

**Restoring is the whole flow again, because open banking has no renewal.**
Consent is valid until a date wimm sets, capped by that bank's
`maximum_consent_validity` — 180 days at most banks. A session can also expire
early, surfacing as `EXPIRED_SESSION` with 401, and both are treated identically
because the member's experience of "my bank stopped updating" is the same. wimm
requests the bank's maximum and shows the member the resulting date before the
hand-off, because it is that bank's limit and not a wimm policy.

Restoring re-enters the flow at the hand-off, skipping the picker. On return,
accounts are matched by hash so sharing choices carry forward; an account newly
offered arrives unshared, and one no longer offered goes with the member told
which.

This is in the first change rather than a later one. Every connection reaches
this state, so shipping without it means shipping a screen that tells a member
their bank stopped working and offers nothing.

**Nothing stored opens a bank on its own.** wimm never receives a member's
banking credentials and never holds the bank's OAuth tokens — the gateway holds
those and refreshes them internally. What wimm does hold is the gateway
`session_id`, which reads that household's accounts until the grant runs out,
and the per-account identifiers used with it. Those are live credentials and a
column is not where they belong.

They are sealed with AES-256-GCM before reaching the database. The key is read at
startup from a **path**, never from an environment variable holding the
material. The owning row's uuid is the additional authenticated data, so a
ciphertext lifted from one row cannot be replayed into another, and a key id is
stored beside each value so a key can be rotated without a flag day.

*Alternative:* `pgcrypto`, or full-disk encryption. Both rejected for the same
reason: the threat is a copy of the data — a dump, a backup, a replica, a stolen
volume — and in each of those the key travels with the rows, or the SQL carrying
it reaches the query log. Encrypting in Go keeps the key out of the database
entirely.

A sealed value has a type that cannot be printed: `String`, `MarshalJSON` and
`slog.LogValuer` all return a redaction, and plaintext exists only as a local
inside the adapter for the length of one call. Sealed values are destroyed when
access ends — on disconnection and on a grant running out — while the connection
row survives for the audit question. The application signing key is the master
credential and is treated the same way, with startup refused if its file is
group or world readable.

**Failures are wimm's taxonomy**: bank unavailable, gateway unavailable, consent
declined, consent expired, no accounts, rate limited with a retry-after.
`docs/design/surfaces.md` routes a page banner, an inline alert and a field error
off what kind of failure it is, and no two gateways agree on status codes.

**`bankingtest` is part of the port, not a test helper.** It keeps `just check`
off the network, and it is a second implementation written the same week as the
first, which is the cheapest pressure test for whether the interface abstracts
anything.

**Money is `int64` minor units plus an ISO 4217 code.** Never a float, never the
gateway's decimal string, never a bare number without its currency. Totals are
per currency and mixed currencies are never summed: wimm holds no rates, and
inventing one would invent the number a household trusts most.

**Superseded by ADR 0019.** This decision said accounts are household-visible
once shared, with no scoping rule beyond the flag itself. An account now has
owners and each other member a level. `connected_by` survives unchanged: it
records whose consent is holding a connection open and who will have to restore
it, and it confers no authority over who sees what.

**The full account number is never stored.** Only enough trailing characters to
tell two accounts at one bank apart, which is all the behaviour requires.

## Consequences

A new third-party dependency, a new outbound network path, two new secrets at
rest, and three new tables owned by `wimmd`. The sweeper gains a fourth statement
for abandoned consent attempts.

The household's balances sit at rest in wimm, for the accounts its members chose
to share and no others. The privacy boundary is the household by design and by
the product's shape: one household per instance, run by its own operator.

The abstraction is a claim until a second gateway is implemented. The paper
table and `bankingtest` reduce that risk; the residual is that a third gateway is
shaped unlike either, which no design prevents.

Deploying with `WIMM_BANKING_GATEWAY` unset is a no-op beyond an empty accounts
screen. Rollback is the goose down migration, losing connections, sharing choices
and readings. Reconnecting recovers the accounts, since identity is the gateway's
cross-session hash, but the sharing choices are made again.
