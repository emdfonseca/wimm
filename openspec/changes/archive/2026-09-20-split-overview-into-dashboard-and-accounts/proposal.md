## Why

Overview is the screen a member lands on after signing in, and today it is
entirely bank-connection management: per-bank cards, restore/disconnect
controls, who-sees-this. It shows a total, but nothing about what changed
recently or which way a balance is moving. A member checking wimm to see
where their money stands has to read past account administration to find the
one number they came for, and gets no sense of recent activity or direction
without a detour to Transactions.

## What Changes

- Overview keeps the household total and drops the account list, the bank
  cards, and every connection action (connect, restore, disconnect,
  who-sees-this).
- Overview gains a recent slice of the member's own transactions (owner-only,
  per ADR 0021) and a per-currency balance trend, both derived from the
  stored ledger.
- A new **Accounts** screen, reachable from the primary navigation, takes
  over everything Overview drops: the account list, per-bank cards, and every
  connection action.
- The primary navigation gains a fourth destination.

## Capabilities

### New Capabilities
- `banking/overview`: what the landing screen shows once a bank is
  connected — the total, a recent slice of the member's own transactions, and
  a per-currency balance trend — and what it shows before any bank is
  connected.

### Modified Capabilities
- `banking/household-accounts`: the account list, the arrival balance read,
  and the "before any bank is connected" empty state move from Overview to
  the new Accounts screen; Accounts becomes a primary navigation destination.
- `banking/bank-connections`: every reference to a connection's status or a
  connection action appearing "on Overview" moves to "on Accounts".

## Impact

- `apps/web/src/routes/(app)/+page.svelte` and `+page.server.ts`: rewritten
  for the dashboard; the account/bank logic they carry moves to a new
  `apps/web/src/routes/(app)/accounts/` route.
- `packages/ui/src/pages/AccountsOverview.svelte`: split into a new Overview
  dashboard page and an Accounts page; the account/bank-card markup moves
  largely as-is.
- `packages/ui/src/destinations.ts`: a fourth destination, and a fourth icon
  value in the `Destination['icon']` union.
- New UI: a recent-transactions list and a trend chart, likely new
  `packages/ui` components (canvas will say which).
- No schema change: the trend is derived from `transactions` and the current
  `accounts.balance_minor`/`balance_read_at`, both already stored.
