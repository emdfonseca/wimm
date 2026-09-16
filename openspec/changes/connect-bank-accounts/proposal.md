## Why

wimm signs a household in and shows them nothing. Identity is complete —
register, enrol, sign in, sweep — and the ledger it protects does not exist, so
every session ends on an empty Overview.

The product's question is "where is my money", and the honest first answer is
not a form. It is the household's real accounts, with their real balances, on
one screen, instead of three separate bank logins. That is reachable now
because open-banking gateways expose exactly this, and it is the smallest slice
a member can use without anyone typing a number in by hand.

## What Changes

- A member picks their bank from a list, is sent to the bank to consent, and
  comes back to wimm with that bank's accounts.
- **The member then chooses which of those accounts the household sees.** This
  is not a consent step: access is already granted by the time it appears. Some
  banks let a member narrow the accounts in their own consent screen and some
  hand over everything, so this is the only place the choice reliably exists.
  Without it, connecting a bank to share a joint account also shares every
  personal account at that bank.
- Accounts the member shares are visible to **every** member of the household.
  The connection records who made it, for audit and for who has to act when it
  expires.
- Overview stops being empty. It lists the shared accounts with their balances
  and a household total.
- **Balances are read when a member opens Overview**, not only when a bank is
  connected, with a control to read them again. A member sitting in front of
  the screen is exactly the case gateways do not throttle.
- **Access expires, and the member can restore it.** Consent lasts at most what
  each bank allows, typically 180 days. There is no refresh mechanism in open
  banking: renewal is the whole authorisation flow again, which this change is
  already building. Shipping without it would mean a screen that tells a member
  their bank stopped working and offers nothing.
- A member can disconnect a bank, which removes its accounts from Overview.
- `wimmd` gains a provider-agnostic banking port, `internal/banking`, with
  Enable Banking behind it as the first implementation and an in-memory fake
  beside it. Nothing outside that package names a gateway; a lint rule enforces
  it rather than a convention.
- The gateway is a stored column on the connection, not a build-time choice, so
  a later gateway is new connections rather than a reinterpretation of old rows.

Not in this change, and deliberately: transactions and the ledger table,
scheduled background syncing, categories, budgets, reports, and any currency
conversion.

## Capabilities

### New Capabilities

- `banking/bank-connections`: a member connects a bank through an open-banking
  gateway, chooses which of its accounts the household sees, restores access
  when it expires, and disconnects. Covers the consent hand-off and return,
  what a connection carries, and what a member is told when the bank refuses or
  is unreachable.
- `banking/household-accounts`: what the household sees once a bank is
  connected. Covers household visibility of shared accounts, the account list
  and its balances, the freshness of a balance and how it is brought up to
  date, the household total, and the empty state.

### Modified Capabilities

None. Identity's three specs are untouched; this change consumes the session
they already establish.

## Impact

**New dependency and new external service.** Enable Banking is the first
third-party service wimm calls outbound, and the first secret at rest beyond
the operator credential. Needs an ADR: dependency, service, schema and secret.

**Commercial.** Enable Banking's free Restricted Production tier covers accounts
the application owner personally links, which is the shape a self-hosted
household instance has. Whether a second household member's bank counts as
personally linked is not answered in their documentation and has to be asked
before this ships.

**Schema.** New tables for connections, pending connections and accounts, owned
by `wimmd`, with goose migrations. Accounts are keyed on the gateway's
cross-session identity hash, not on a per-session identifier. Every lifetime is
compared against database time, per ADR 0017.

**Contracts.** New Connect service definition on the public listener. The
browser never sees it: SvelteKit server routes call Connect and hand the web
client JSON, per ADR 0001.

**Web.** New routes for the bank picker, the consent hand-off, the return, and
the account chooser. The return path carries a one-time value, so it is
exchanged server-side and redirected away from, the same shape ADR 0016 already
uses for enrolment links.

**Design system.** Overview gains its first real content. It stays the only
destination: Overview is the accounts screen in this change, and a second nav
item showing the same thing is a product nobody agreed to build. Several
library components this needs do not exist yet and are built first, per ADR 0008.

**Operations.** A household that connects no bank sees the empty state, so the
change is safe to deploy before anyone holds gateway credentials.
