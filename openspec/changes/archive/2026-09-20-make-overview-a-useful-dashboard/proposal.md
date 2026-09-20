## Why

Overview answers one question, "how much is there", and then stops. A member
who opens wimm to see how the month is going finds a total, five raw bank
strings and a strip of twelve identical bars, with two thirds of the screen
empty. Walking the built screen with real data also showed the trend is wrong
in two ways that matter: a ledger reaching back one day still draws twelve
bars, and a currency holding nothing still gets a total and a flat trend.

## What Changes

- The single total becomes two figures: **household money**, the accounts
  every member owns or sees in full, and **your money**, the accounts a member
  owns that are not the household's. They never overlap. Anything else a
  member may see is listed and counted in neither. **BREAKING** for
  `banking/household-accounts`: a total is no longer everything a member may
  see, and ADR 0019's "totals are per member" is superseded for the household
  figure, which is the same for everybody.
- Overview gains a **month summary** per currency: money in, money out and
  the net for this month so far, each set against the same days of last
  month.
- The twelve-bar trend becomes a **balance chart**: a line across the full
  content width, one point per day, with dates and amounts a member can read
  off it. An account with only a few days of history no longer shortens the
  chart for every other account, and a span too short to mean anything draws
  nothing.
- Overview gains **top spending** for the month so far: the merchants the
  most money went to, and the largest single payments.
- A transaction is shown under a **merchant name** cleaned of the bank's own
  noise: `COMPRA WWW.AMAZON NM4HU1VZ4 230002268264350` reads as
  `Amazon`. The same name is used wherever a transaction is listed,
  Overview and Transactions alike, and it is what top spending groups by.
- Overview gains an **accounts list**: every account the member may see, with
  its balance and when that balance was read, leading to Accounts, grouped as
  Household, Yours and Shared with you, so each figure can be checked against
  the accounts under it. **BREAKING** for
  `banking/overview` as drafted: that capability said Overview lists no
  account and names no bank. It still performs no connection action.
- A currency with a zero total and no movement gets no tile, no summary and
  no chart on Overview. Its account still appears in the accounts list.
- The screen fills the Main column at Medium, Wide and Ultra (ADR 0005)
  instead of stacking narrow cards down the left edge.

Not in this change: categories, budgets, a date-range picker, currency
conversion, and recognising transfers between a member's own accounts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `banking/overview`: the total requirement stops forbidding an account list
  and gains the zero-currency rule; the trend requirement is replaced by the
  balance chart's; month summary, top spending and the accounts list are
  added.
- `banking/transactions`: adds how a transaction is named on screen.
- `banking/household-accounts`: "The total is the member's own" is replaced by
  household money and a member's own money as separate figures.

## Impact

- `packages/contracts/proto/wimm/banking/v1/banking.proto`: a new
  `GetMonthSummary` RPC, a `display_name` field on `Transaction`, and
  `GetBalanceTrend` returning daily points. Server to server only (ADR 0001),
  one consumer in the same deploy. `ListAccounts` returns the two figures and
  each account's group.
- A new ADR, superseding the totals paragraph of ADR 0019.
- `apps/wimm/internal/banking`: `trend.go` reworked; new month summary and
  merchant-name code; `apps/wimm/internal/store/transactions.go` gains the
  month queries. No schema change, no migration.
- `apps/web/src/routes/(app)/+page.server.ts` and
  `transactions/+page.server.ts`: the duplicated
  `counterpartyName || remittance || 'Card payment'` goes; both read
  `display_name`.
- `packages/ui`: `pages/Overview.svelte` rebuilt with one state story per
  row of canvas.md's state plan; new components as canvas.md lists them.
- No new dependency: the chart is hand-written SVG.
