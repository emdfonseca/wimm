## Context

See proposal.md - Why. Today `(app)/+page.svelte` and `+page.server.ts` are
the account-management screen wimmd's `ListAccounts` and `ListConnectionAccounts`
RPCs back (`apps/wimm/internal/rpc/banking.go:261`,`:103`); connect, restore
and disconnect are POSTs to `/api/banking`. `ListAccounts` already computes the
per-member, per-currency total (`toProtoTotals`). `ListTransactions`
(`apps/wimm/internal/rpc/banking.go:338`) already serves the household's
owner-only ledger, keyset-paged newest-first (ADR 0021).

No new storage: `accounts.balance_minor`/`balance_read_at` hold the current
reading, and `transactions` holds the booked ledger — the trend's own bound
per account is `min(booking_date)` over that table, not
`accounts.transactions_synced_through` (a forward completeness watermark: an
exhausted sync sets it to today however far back the history it exhausted
actually reaches, so it never says how far *back* a ledger goes). This
change reads that data differently; it does not add to it.

## Language

- **Overview** — the screen a member lands on after signing in. Shows the
  total, recent transactions, the trend. Never shows an account list, a bank
  name, or a connection action.
- **Accounts** — the screen that holds the account list and every connection
  action (connect, restore, disconnect, who-sees-this). Never shows the
  trend.
- **Trend** — the per-currency balance sparkline (bars, not a line) Overview
  draws from booked transactions. Never called a "chart" in copy (it is one
  word, describing direction) and never a stored value — it is computed on
  read.
- **Recent transactions** — the bounded slice of the ledger Overview shows.
  Never called "activity" or "feed"; the Transactions screen is still "the
  ledger" or "Transactions", never "the dashboard".

## Goals / Non-Goals

**Goals:**
- Move every account-management concern off Overview onto Accounts with no
  behaviour change to that management (connect/restore/disconnect/who-sees
  keep their existing rules, only their screen changes).
- Compute the trend and the recent slice from data already stored, with no
  migration.

**Non-Goals:**
- Currency conversion or a combined total across currencies — unchanged from
  today, and the trend follows the same never-combine rule.
- A configurable trend window (7/30/90 days, etc.) — the trend shows what the
  ledger has, full stop; a picker is a later change if wanted.
- Moving `RefreshBalances` itself. It already reads every account a member
  may see regardless of which screen calls it; only where the button lives
  moves, to Accounts.

## Decisions

**The trend is computed in wimmd, not in the SvelteKit load function.**
Walking each account's booked transactions backward from `balance_minor`,
bounded by that account's own earliest booked transaction, and grouping by
currency, is the same kind of money-shape logic `toProtoTotals` already does
in Go — the one
place ADR 0001 puts server-to-server business logic. A new RPC method (name
and message shape are implementation detail for tasks.md) returns trend
points per currency alongside the existing account/total data. Alternative
considered: compute it in `+page.server.ts` from `ListTransactions` output.
Rejected — it would duplicate the currency-grouping and history-bound rules
`toProtoTotals` and the account/scope checks already enforce in Go, in a
layer ADR 0001 reserves for wiring, not money arithmetic.

**Recent transactions reuse `ListTransactions`, not a new endpoint.** Overview
asks for the first page at a small size; it is the same owner-only,
keyset-paged household ledger `banking/transactions` already specifies, so no
new authorization or pagination logic is written twice.

**The account-list and bank-card markup moves as a block.** The
`{#each banks as bank}` section and its `AccountRow` usage in
`packages/ui/src/pages/AccountsOverview.svelte:319-377` become a new
`AccountsScreen.svelte` largely unchanged; the outcome-notice snippet
(connect/restore/disconnect/declined feedback) moves with it, since every one
of those outcomes is a connection action. `AccountsOverview.svelte` itself is
retired; a new `Overview.svelte` page component is built for the total +
recent + trend layout rather than trimmed down from it, since it shares
almost nothing with what remains.

**The trend is a bar sparkline, not a line chart or a charting library.** A
row of plain, evenly spaced bars (one per period, height proportional to
balance) is a flexbox layout wherever it is built, whereas a line connecting
individual points needs a coordinate system neither the canvas tool nor a
dependency-free Svelte component gets for free. `packages/ui` has no charting
dependency today, and `Icon.svelte`'s comment records the same choice for its
nine glyphs: a small, closed visual built from the token stylesheet costs
less than a dependency whose tree-shaking and theming need verifying for one
component. Alternative considered: a line chart via a charting library.
Rejected on both counts — a line needs positioned points, not layout, and a
dependency is more than this one visual is worth.

**The primary navigation gains a fourth destination, `Accounts`, with a new
icon.** `packages/ui/src/destinations.ts:11`'s `Destination['icon']` union
gains one value (a bank/landmark glyph is the natural fit; the exact lucide
path is a task, added to `Icon.svelte` the same way its existing nine were).
Reachability is specified once, on `banking/household-accounts` (mirroring
how `app/settings` specifies its own reachability), not duplicated in
`app/navigation-feedback`, which governs loading/announcement behaviour only
and never named the destination list.

## Risks / Trade-offs

[A member's Overview trend goes flat or missing the day they widen a
connection's consent scope from balances to balances-and-transactions] →
Expected, specified in `banking/overview`'s "An account with balances only"
scenario, and stated on screen rather than silently drawn as a shorter line.

[Computing the trend by walking transactions backward on every Overview load
costs more than reading a stored balance] → Bounded by how far back each
account's ledger actually reaches, which is itself bounded by what the bank
grants (ADR 0018/0021 — typically well under a year); revisit only if a
sync's history genuinely grows unbounded, which no current gateway does.

## Migration Plan

No data migration. Deploy order: ship the new wimmd RPC method and the
`ListTransactions`-backed recent slice first (additive), then the two-screen
web UI and the navigation change together, since a landed nav change with no
Accounts route would be a dead link. Rollback is reverting the web deploy;
the additive RPC method needs no rollback of its own.
