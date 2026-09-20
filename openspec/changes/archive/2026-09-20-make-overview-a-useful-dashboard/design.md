## Context

See proposal.md for why. What shapes the approach:

- Overview's load (`apps/web/src/routes/(app)/+page.server.ts`) makes three
  Connect calls in parallel and never asks a bank. That stays true.
- `apps/wimm/internal/banking/trend.go` walks each owned account backward
  from its current balance over booked rows. `trendPoints = 12` evenly spaced
  timestamps between `bound` and now, where `bound` is the latest of the
  contributing accounts' earliest booked dates. A one-day ledger therefore
  yields twelve points about two hours apart on rows dated at midnight, which
  is the twelve equal bars. An account at zero with one old booked row yields
  twelve zeros, which is the flat USD strip.
- `molecules/TrendSparkline.svelte` draws bars from a zero baseline at 48 px,
  so a 1% move on a large balance is under a pixel. The card is fixed at
  360 px.
- A transaction's text is `counterpartyName || remittance || 'Card payment'`,
  written twice: Overview's load and Transactions' load. Nothing cleans it.
- `ListAccounts` already returns every visible account with balance,
  `read_at`, `stale`, bank name and `left_out_at`. Overview loads it today and
  throws the accounts away. Its `totals` sum every visible account per
  currency (`reading.go:189`), granted ones included, which is the figure
  being replaced.
- `time.Now` is banned outside tests (ADR 0017); `s.store.Now(ctx)` is the
  clock.
- Nothing under `packages/ui/src` draws a quantity over time as a line. The
  only SVG there is icons and marks.

## Language

- **Household money** — the sum of accounts every member owns or has
  *details* on. Code: `household`. Never "household total", "shared money"
  or "joint".
- **Your money** — the sum of accounts a member owns that are not household
  money. Code: `own`. Never "personal" or "private"; the group heading is
  **Yours**.
- **Shared with you** — accounts a member may see that count in neither
  figure. Code: `shared`. It is a group in a list, never a figure.
- **Month summary** — money in, money out and net for the month so far. Never
  "budget", "cash flow" or "report".
- **Money in / Money out / Net** — the three figures. Never "income",
  "expenses", "profit": income is a token family for amounts' direction (ADR
  0006) and a transfer in is not income.
- **Month so far** — the 1st through today. Never "this month" on screen
  beside a comparison, because last month's figure is not the whole of last
  month.
- **Balance chart** — the daily line. Never "trend" or "sparkline" for this
  thing. `TrendSparkline` keeps its name for the small bars it still is.
- **Merchant name** — the name wimm derives for a transaction. Code:
  `display_name`. Never "description", "label" or "clean name".
- **Bank's line** — the statement text as the bank wrote it. Never "raw".
- **Top spending** — the section. **Top merchants** and **Largest payments**
  are its two lists.
- **Accounts list** — Overview's read-only list. Never "accounts strip" or
  "cards"; the cards are on Accounts.

## Goals / Non-Goals

**Goals:**

- Everything is computed in `wimmd` from stored rows, in one round of
  parallel calls, with no bank read.
- One home for the merchant name, in Go, so grouping and display cannot
  disagree.
- No new dependency and no schema change.

**Non-Goals:**

- Recognising transfers between a member's own accounts. Matching equal and
  opposite amounts is the same guesswork ADR 0021 refused for pending rows.
  The summary says transfers are counted.
- A range picker, categories, budgets, conversion.
- Learning merchant names, or a per-household rename. The rule set is fixed
  and table-tested.
- Persisting the merchant name. It is derived on read; the rules will change
  and a stored name would need a backfill each time.

## Decisions

**The balance chart is daily points over at most 90 days.**
`GetBalanceTrend` keeps its name and shape (`CurrencyTrend.points`) and
returns one `TrendPoint` per day, end of day, the last being now. `walkBack`
is unchanged in principle; `evenDates` becomes day boundaries in UTC, which
is what `booking_date` is. *Alternative:* keep twelve points and draw a line
through them. Rejected: a member pointing at the chart is asking about a day,
and a point every 7.5 days answers a different question.

**Which accounts set the span.** Per currency, among owned accounts with
transaction scope and a non-nil `Oldest`:

```text
reach(a)   = today - a.Oldest, in days
long       = accounts with reach >= 30
contribute = long, if any; otherwise the single account with the longest reach
bound      = max(latest Oldest among contribute, today - 90 days)
no chart   = today - bound < 7 days     -> CurrencyTrend.short_history = true
partial    = any owned account not in contribute
```

`partial_coverage` already exists and already drives the "does not cover
every account" line; short-history exclusion sets the same flag. When nothing
reaches 30 days only the longest account contributes, rather than all of
them, for the reason the rule exists: the shortest would otherwise set the
span. `short_history` is a new bool on `CurrencyTrend`, so the screen can
tell "no chart yet" from "no transaction access", which get different copy.
*Alternative:* let each account join the sum on its own first day. Rejected:
the line would step up by an account's whole balance on the day its history
begins, which reads as money arriving.

**An account's group is decided in `wimmd`, and the figures are sums over
it.** `Account` gains `group` (`HOUSEHOLD`, `OWN`, `SHARED`) and
`ListAccountsResponse.totals` is replaced by `household_totals` and
`own_totals`, each a `CurrencyTotal` per currency. The rule, per visible
account that is not left out:

```text
members    = every member of the household
full(m)    = m owns the account, or holds a details grant on it
HOUSEHOLD    when there are two or more members and full(m) for every one
OWN          when the viewer owns it and it is not HOUSEHOLD
SHARED       otherwise
```

One query gives, per account, the count of members with an owner row or a
*details* grant; the account is household when that count equals the
household's member count. A household of one has no "everyone else", so
everything its member owns is `OWN`. The web load never classifies anything:
a group computed in two places is two answers waiting to differ.
*Alternative:* a household flag an owner sets by hand. Rejected: it is a
second statement of something the grants already say, and the two would
drift the first time somebody changed a level.

Household money is the same figure for every member, which ADR 0019 ruled
out for totals because a shared figure could reveal an account somebody may
not see. It cannot here: an account is household only when every member
already sees it in full.

**The zero-currency rule lives in `wimmd`, once.** A currency is *quiet* when
both its figures are zero and no owned account in it has a booked row inside
the chart window. `GetBalanceTrend` returns no `CurrencyTrend` for it and
`GetMonthSummary` no summary. The money tiles are dropped in the web load, which
is the only place holding both the figures and the trend response: a currency
with zero figures and no trend entry and no summary entry gets no tile.

**One new RPC, `GetMonthSummary`.**

```text
GetMonthSummaryResponse { repeated CurrencyMonth months }
CurrencyMonth {
  currency
  Money in, out, net                    month so far
  Money prior_in, prior_out, prior_net  unset when last month is not all held
  Timestamp counted_from                set when a ledger begins after the 1st
  Timestamp month_start                 for the link to Transactions
  repeated MerchantTotal top_merchants  name, Money total, int32 payments
  repeated Transaction largest_payments
}
```

Summary and top spending share a window and a row set, so they share a call
and a query. Two store queries per owned account set: signed sums for the two
windows in SQL, and this window's outgoing booked rows fetched whole, because
grouping is by merchant name and the name is derived in Go. A household's
month of outgoing rows is hundreds, not thousands. *Alternative:* group in
SQL on `counterparty_name`. Rejected: the rows that most need grouping are
the ones where it is empty and the merchant is inside `remittance`.

Windows come from `s.store.Now(ctx)`: `[first of month, now]` and
`[first of prior month, same day number, capped at that month's length]`.
`prior_*` is set only when every contributing account's `Oldest` is on or
before the first of the prior month.

**The merchant name is a Go function and a proto field.**
`banking.MerchantName(counterparty, remittance string) string` in
`apps/wimm/internal/banking/merchant.go`; `Transaction.display_name` is set in
`toProtoTransaction`. Both web loads read it and the duplicated `||` chain
goes. The first rule set, applied to the counterparty where there is one and
otherwise to the remittance:

```text
1. cut a trailer starting at " - PAIS:" or " VALOR ORIG" (case-insensitive)
2. drop trailing tokens that are 9 or more digits, repeatedly
3. drop a trailing token of 8+ characters mixing letters and digits
4. drop a leading transaction word: COMPRA, COMPRAS, PAGAMENTO, PAG, PG,
   TRF, TRANSF, TRANSFERENCIA, DD, LEVANTAMENTO, MULTIBANCO, with any
   trailing "." or "-"; repeat, so "COMPRAS MULTIBANCO PAYPAL" loses both
5. drop a leading "WWW." and a trailing ".COM", ".PT", ".EU", ".NET"
6. collapse whitespace; title-case a result that is entirely upper case
7. empty result -> the bank's line, trimmed; that empty too -> "Card payment"
```

Against the five lines in the walk-through: `Bxv Via Verde`, `Amazon`,
`Number 1 Hair`, `Rest Botafogo`, `Paypal Converse`. Rule 4's word list is
Portuguese because the banks connected today are. It is a `var` table, and a
new bank's words are a row and a test case, not a design change.

**Top merchants group on the lower-cased merchant name.** Display uses the
first-seen casing. `Paypal Converse` and `Paypal Steam` stay separate, which
is right: the member paid two merchants.

**The chart is hand-written SVG.** `molecules/BalanceChart.svelte`: a
`viewBox` polyline over an area fill in `--color-chart-1`, the y-axis scaled
to `[min, max]` padded by 8% and never forced to zero, since a baseline at
zero is what flattened the bars. Four text labels: start date, end date, high
and low. One focusable group with `role="img"` and a summary sentence as its
accessible name, left and right arrows moving a marker a day at a time, a
pointer doing the same, and the marker's date and balance in a
`aria-live="polite"` readout. Formatting of amounts and dates is done by the
load, which passes strings beside numbers, so `packages/ui` formats no money.
*Alternative:* a chart library. Rejected: one line chart does not justify a
dependency and the ADR it would need, and every candidate brings its own
colour and type decisions into a system that has tokens for both.

**The chart's geometry is a contract, and its line is data.** The box,
padding, stroke, fill, label positions and copy are fixed numbers and words in
canvas.md, under Contracts and Words, and the component's state stories hold
them. The line itself comes from the points a story or the load passes in.

**Layout.** Wide and Ultra, top to bottom: household money and your money
as a row of two tiles; the month summary as a row of three; the balance chart across the full Main column; top spending
and the accounts list side by side; recent transactions across the full
column. Medium keeps the order and both tile rows. Compact stacks
everything in that order. Two currencies repeat the tile row and the chart
per currency; top spending shows the currency with the most money out first.
No section is capped by `layout-content-max`, which protects text measure
only (ADR 0005).

**Transactions shows the bank's line under the name when they differ.** It
sits under the name in `molecules/LedgerRow.svelte`, and the row state is a
state story in `LedgerRow.stories.svelte`, as canvas.md plans it.

**One ADR, for the figures.** It supersedes the "Totals are per member"
paragraph of ADR 0019 and records the household rule, the disjointness, and
why a figure every member shares leaks nothing. The RPC and the chart need
none: no dependency, service, schema or public contract.

## Risks / Trade-offs

- [Transfers between own accounts inflate both figures, and paying a credit
  card from a current account is the common case] → The summary says so in
  its own copy. Recognising transfers is its own change, with its own
  argument about guesswork.
- [A rule strips something that was the merchant] → Rule 7 never yields an
  empty name, Transactions shows the bank's line beside any name that differs
  from it, and the rule table is tested on real lines from each connected
  bank.
- [Dropping the shortest account makes the chart's last point disagree with
  the total] → Already true for balance-only accounts, and already said on
  screen by the partial-coverage line.
- [The month comparison disappears for a household that connected a second
  bank three weeks ago] → Correct, and it returns on its own once that ledger
  covers a whole prior month.
- [A member joining with nothing granted empties household money for
  everyone until they are given *details*] → Correct by the rule and stated
  as a scenario. The accounts move to their owners' own money rather than
  vanishing, so no balance leaves the screen.
- [90 daily points across several accounts means more rows walked per load]
  → One `TransactionsSince` per contributing account, bounded at 90 days; the
  walk is linear in rows.
- [UTC day boundaries put a late-evening payment on the wrong day for a
  member west of UTC] → `booking_date` is the bank's date, not an instant.
  The chart agrees with Transactions, which is the agreement that matters.

## Open Questions

- Whether `TrendSparkline` stays in `packages/ui/src` once Overview stops
  using it. Overview is its only user, and `packages/ui/src/index.ts`
  exports it.
