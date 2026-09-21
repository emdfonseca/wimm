## Why

Overview says how this month is going against last month and nothing about the
year around it, so a household cannot tell whether it spends more than comes
in, which merchants are behind a rise, what leaves every month without being
asked, or whether one large payment is bending the picture. Each of those is
answered today by adding up Transactions by hand. The balance chart has the
same gap one level down: it shows that a day moved and not what moved it.

This builds on `make-overview-a-useful-dashboard` (the month summary, the
merchant name, household money and a member's own money) and is applied only
after that change is finished.

## What Changes

- **Month by month.** Overview shows money in, money out and net for the month
  so far and up to the 12 full calendar months before it, per currency, from
  booked transactions of owned accounts, with windows taken from database
  time. It states a typical month (the median net) beside the average month,
  only once three full months are held, and says plainly when more goes out
  than comes in. The month so far and any partly held month are drawn and are
  in neither figure.
- **Where a month rose.** Pointing at a month names the merchants that took
  more than they usually do, grouped on the merchant name that already exists.
  Categories are out of scope.
- **Recurring payments.** wimm recognises a payment to one merchant, at about
  one amount, at a steady weekly, monthly or yearly interval, seen at least
  three times, or twice for a yearly one, which reads as likely yearly until a
  third confirms it, and lists each with its amount, its cadence and the date it is
  next expected.
- **Unusual payments.** One rule, used everywhere, marks a payment that is far
  above what is usual, going out or coming in, the second labelled unusual
  income. A first payment to a merchant is also marked when it alone is over a
  tenth of a typical month's money out. Each month says how many it held and its net with and
  without them, and opens to list them, each leading to the payment in
  Transactions, where it is marked too. The typical and average month can be read without them,
  as a view; nothing is hidden or deleted.
- **One scope for the insights.** A single control on Overview, Household,
  Yours and All, applies to everything drawn from transactions. Every scope is
  owned accounts only (ADR 0021); an account a member merely holds a grant on
  is in none. All is the default, since it is what the month cards count
  today. Where Household is narrower than the household money figure, the
  screen says which accounts it counts.
- **What moved a day.** Pointing at a day on the balance chart opens, from the
  pointer, a tap or the keyboard, that day's balance, the change from the day
  before and the transactions that explain the day, with an unusual one
  labelled. It extends the chart's existing readout and follows the scope.
- **An ADR for the detection rules**: recurrence, the unusual-payment rule and
  the explains-the-day rule, with their thresholds as named constants.

## Capabilities

### New Capabilities

- `banking/payment-patterns`: what wimm calls a recurring payment and what it
  calls an unusual payment, stated once so Overview, the balance chart and
  Transactions cannot disagree.

### Modified Capabilities

- `banking/overview`: adds the month-by-month history with its typical and
  average month, the merchants behind a rise, the recurring payments list, the
  unusual payments list and the view without them, the scope control, and what
  moved a day on the balance chart.
- `banking/transactions`: a payment wimm calls unusual is marked in the ledger.

## Impact

- `packages/contracts/proto/wimm/banking/v1/banking.proto`: a new
  `GetMonthHistory` RPC; a `scope` request field on `GetMonthHistory`,
  `GetMonthSummary` and `GetBalanceTrend`; day movers on `TrendPoint`;
  `Transaction.unusual`. Regenerated with `just gen`.
- `apps/wimm/internal/banking`: a recurrence detector, an unusual-payment
  detector, the explains-the-day rule, a month history service, and scope
  applied to `MonthSummary` and `BalanceTrend`. `apps/wimm/internal/store`: a
  monthly sums query and an account filter on the owned-account queries. No
  migration, no new dependency.
- `packages/ui`: a monthly net chart component, a popover on
  `molecules/BalanceChart.svelte`, an unusual marker on
  `molecules/LedgerRow.svelte`, and new state stories on Overview and
  Transactions.
- `apps/web`: Overview's load gains one call and a `scope` query parameter;
  Transactions' load passes the marker through.
- `docs/decisions`: one ADR, for the detection rules.
