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
- **The member then says who owns each account and who may see it.** This is
  not a consent step: access is already granted by the time it appears. Some
  banks let a member narrow the accounts in their own consent screen and some
  hand over everything, so this is the only place the choice reliably exists.
  Without it, connecting a bank to show a joint account also shows every
  personal account at that bank.
- **Seeing an account is per member and comes in three levels**: nothing at all,
  the balance, or the balance plus the account's identifying details. An account
  belongs to one or more members — a joint account has two — and an owner always
  sees their own account in full. The member who connected the bank owns
  everything it returned until they hand it over or let it go, and nobody else
  sees anything until they are given it.
- Overview stops being empty. It lists the accounts that member may see, with
  their balances and a total of exactly those — so two members of one household
  can land on the same screen and correctly see different numbers.
- The connection records who made it, for audit and for who has to act when it
  expires.
- **Balances are read when a member opens Overview**, not only when a bank is
  connected, with a control to read them again. A member sitting in front of
  the screen is exactly the case gateways do not throttle.
- **Access expires, and the member can restore it.** Consent lasts at most what
  each bank allows — 90 days at the banks this household uses. There is no refresh mechanism in open
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
  gateway, says who owns each of its accounts and what each other member may see
  of them, restores access when it expires, and disconnects. Covers the consent hand-off and return,
  what a connection carries, and what a member is told when the bank refuses or
  is unreachable.
- `banking/household-accounts`: what each member sees once a bank is
  connected. Covers ownership and the three levels, the account list
  and its balances, the freshness of a balance and how it is brought up to
  date, the household total, and the empty state.

### Modified Capabilities

None. Identity's three specs are untouched; this change consumes the session
they already establish.

## Impact

**New dependency and new external service.** Enable Banking is the first
third-party service wimm calls outbound, and the first secret at rest beyond
the operator credential. Needs an ADR: dependency, service, schema and secret.

**Commercial.** Enable Banking's free Restricted Production tier reaches only
accounts the Control Panel user links as themselves — their Terms of Service
refuse account information that does not belong to that user. So one member
connects the banks they can authenticate at, joint accounts included, and wimm's
ownership and levels decide what the rest of the household sees of them. Lifting
that limit is a contract and KYB, which needs a company; it is not a reason to
change gateway.

**Schema.** New tables for connections, pending connections and accounts, plus
an owners table and a per-member grants table, owned by `wimmd`, with goose
migrations. An account carries no `shared` flag and no single owner: both are
questions with one answer per member. Accounts are keyed on the gateway's
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
