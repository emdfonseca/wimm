## Context

See proposal.md for why. This change is applied after
`make-overview-a-useful-dashboard` is finished; everything below describes the
code as that change leaves it.

- `apps/wimm/internal/banking/month.go` computes the month so far from
  `store.OwnedWindowSums` and `store.OwnedOutgoing`, both over **every account
  the member owns**, left-out and disconnected accounts excluded. That is the
  All scope, which is why All is the default: nothing a member already reads
  changes until they touch the control.
- `banking.MerchantName` names a row; `topMerchants` groups on its lower-cased
  form. Both are reused as they are.
- `apps/wimm/internal/banking/groups.go` classifies every visible account as
  `HOUSEHOLD`, `OWN` or `SHARED` (ADR 0024). `Account.owned` and
  `Account.group` are already on the wire.
- `apps/wimm/internal/banking/trend.go` fetches
  `store.TransactionsSince(account, bound)` for each contributing account and
  walks the balance back over those rows. The rows that moved each day are
  already in hand where the chart is built.
- `molecules/BalanceChart.svelte` has one tab stop, arrow keys, Home and End, a
  pointer, a `marked` index and an `aria-live="polite"` readout of
  `date · amount`. It formats nothing.
- Transactions takes `?page=<booking date>.<id>`, read inclusively as that
  page's own newest row, which is a way to land on one payment that already
  exists. It ignores `?month=`, which the month tiles send; that is not this
  change's to fix.
- `time.Now` is banned outside tests (ADR 0017). `s.store.Now(ctx)` is the
  clock, and every store window is an argument.
- `booking_date` is a date. Months and days are UTC, as the chart's are.

## Language

- **Month by month** — the section holding the months. Code: `history`. Never
  "trend" (the balance chart lost that word), "cash flow", "report" or
  "budget".
- **Full month** — a calendar month that has ended and that every contributing
  ledger holds from its first day. Code: `full`. Never "complete".
- **So far** — the current month. **Held from** — a month a ledger begins
  inside. Code: `partial`. Neither is ever called a full month.
- **Typical month** — the median net of the full months shown. Code:
  `typical`. Never "median" on screen, never "normal".
- **Average month** — the mean net of the same months. Code: `average`.
- **More comes in than goes out / More goes out than comes in** — the
  sentence. Never "earn", "spend more than you earn", "deficit" or "surplus":
  a transfer in is not earnings, and Money in and Money out are the words the
  month summary already uses. "Income" appears in exactly one place, the label
  **Unusual income**.
- **Usual** — a merchant's own median month, or an unusual payment's baseline
  median. "€82.40 more than usual", "usually about €40".
- **Recurring payment** — a payment `banking/payment-patterns` recognises.
  Code: `recurring`. Never "subscription" (rent is not one), "bill" or
  "standing order" (that is a bank instruction, which wimm cannot see).
- **Cadence** — `Weekly`, `Monthly`, `Yearly`, and `Likely yearly` for a
  yearly payment seen twice. Never "frequency", "probably" or "maybe".
- **Expected** — the date a recurring payment is next due. Never "due": wimm
  predicts, the merchant decides.
- **Unusual payment** — a booked transaction the one rule marks, in either
  direction. Code: `unusual`. One going out is labelled **Unusual**; one coming
  in is labelled **Unusual income**. Never "outlier", "anomaly", "suspicious",
  "flagged" or "windfall" on screen: it is a statement about size, not about
  wrongdoing or luck.
- **First payment** — an unusual payment to a merchant never paid before. It
  reads "first payment to Auto Reparadora", never "usually about".
- **Without unusual payments** — the view. Never "excluding outliers",
  "adjusted" or "normalised".
- **Scope** — which owned accounts the insights count: **Household**,
  **Yours**, **All**. Code: `scope`, `HOUSEHOLD`, `OWN`, `ALL`. Household and
  Yours are ADR 0024's words. Never "filter", and never "Shared".
- **What moved the day** — the rows in the chart's popover. Code: `movers`.
  Never "top transactions": the rule is about explaining the day, not ranking.

## Goals / Non-Goals

**Goals:**

- Everything is computed in `wimmd` from stored rows, with no bank read, no
  migration and no new dependency.
- One detector per pattern, each a pure function over rows with its thresholds
  as named constants, so Overview, the chart and Transactions cannot disagree
  and a threshold is one edit and one table test.
- Every scope is a subset of the accounts the member owns. No code path adds up
  transactions from an account the member merely holds a grant on.

**Non-Goals:**

- Categories, budgets, forecasts, conversion between currencies.
- Recognising transfers between a member's own accounts. Same reasoning as the
  month summary: it is guesswork, and the section says transfers are counted.
- Calling money coming in recurring. A salary is not a commitment.
- Persisting any pattern. Recurring and unusual are derived on read, like the
  merchant name: the rules will change, and a stored mark needs a backfill
  each time.
- Widening any scope to accounts a member holds *details* on and does not own.
  ADR 0021 settles it; see Open Questions.
- A way to dismiss, confirm or correct a mark. That is stored state and its own
  change.
- Fixing `?month=` on Transactions.

## Decisions

**One new RPC, `GetMonthHistory`, and a scope on the three that draw from
transactions.**

```text
enum InsightScope { UNSPECIFIED (read as ALL), HOUSEHOLD, OWN, ALL }

GetMonthHistoryRequest  { InsightScope scope }
GetMonthSummaryRequest  { InsightScope scope }     field added
GetBalanceTrendRequest  { InsightScope scope }     field added

GetMonthHistoryResponse {
  InsightScope scope                     the scope actually answered
  repeated InsightScope available        empty when no control is to be shown
  repeated string household_counted      account names, set only when
  repeated string household_not_counted  Household is narrower than the figure
  repeated CurrencyHistory histories
}
CurrencyHistory {
  currency
  repeated HistoryMonth months           newest first, at most 13
  int32 full_months
  Money typical_net, average_net                 unset under 3 full months
  Money typical_net_usual, average_net_usual     the same, unusual set aside
  int32 unusual_count                            across the months shown
  repeated RecurringPayment recurring            soonest expected first
}
HistoryMonth {
  Timestamp month_start
  Money in, out, net
  Money net_usual                        set only when the month held one
  bool so_far
  Timestamp held_from                    set when a ledger begins inside it
  repeated MerchantRise risers           at most 5; empty for a partial month
  repeated UnusualPayment unusual        newest first
}
MerchantRise     { name, Money total, Money usual, int32 payments }
UnusualPayment   { Transaction transaction, Money typical }
UnusualPayment   += bool first_payment      typical is unset when it is set
RecurringPayment { name, Money amount, Cadence cadence, bool likely,
                   Timestamp expected, bool late, string account_id }
TrendPoint       += repeated DayMover movers, int32 smaller
DayMover         { display_name, Money amount, bool unusual }
Transaction      += bool unusual
```

One RPC for the months, recurring and unusual together because they read one
row set: 25 months of booked rows, both directions, on the scoped accounts
(`store.OwnedBooked`, which `OwnedOutgoing` becomes a filter over). Splitting
them would fetch it three times. *Alternative:* fold everything into
`GetMonthSummary`. Rejected: the month summary is a two-window sum that a
member with a three-week ledger gets in full, and it should not start costing a
25-month read.

**A scope is an account set, resolved once in the service and enforced again in
the store.** `scopedAccounts(ctx, member, scope)` returns the ids of the
member's owned accounts whose group is `HOUSEHOLD`, `OWN`, or either. The
owned-account store queries (`OwnedWindowSums`, `OwnedOutgoing`,
`OwnedAccountsForTrend`, the new `OwnedMonthlySums`) gain an `accountIDs`
argument applied as `a.id = any($n)` **beside** their existing join on
`account_owners`, never instead of it. An id the member does not own therefore
matches nothing even if a caller passes it, which is the ADR 0021 line held in
SQL and asserted by a store test that passes a granted account's id.

```text
household = owned ∩ group HOUSEHOLD      yours = owned ∩ group OWN
available = [HOUSEHOLD, OWN, ALL]  when both sets are non-empty
          = []                     otherwise
answered  = the requested scope when available, else ALL
narrower  = visible HOUSEHOLD accounts the member does not own, by name
```

All is offered only beside both others: with one set empty, All and the other
count the same accounts and a control with two identical answers is worse than
none. A household of one has no `HOUSEHOLD` group, so it falls out of the same
rule. Availability is decided in `wimmd` although `Account.owned` and
`Account.group` would let the web load derive it, for the reason the group
itself is: a rule computed in two places is two answers.

**The scope is a query parameter, `/?scope=household` or `/?scope=yours`.** All
is the address with no parameter. It changes what three RPCs return, so it is a
load concern and not component state; it survives a reload and can be sent to a
partner. An unknown or unavailable value is answered as All by the service and
the load redirects to the bare address so the control and the address agree.
*Alternative:* remember it per member. Rejected: stored preference is a schema
change for something the address already does.

**The scope applies to the month summary, top spending and the chart as well as
to the new sections.** September's bar in Month by month and the Money out tile
above it are the same number. A control that moved one and not the other would
put two different Septembers on one screen.

**Months: one grouped sum, fullness decided in Go.**
`store.OwnedMonthlySums(member, accountIDs, from, to)` groups booked rows by
currency and `date_trunc('month', booking_date)`. `from` is the first of the
month 12 before the current one and `to` is tomorrow, both from
`s.store.Now`. Per currency, with `ledgers` the `Oldest` of each scoped account
with transaction scope:

```text
begins      = the latest of ledgers            (every ledger is held from here)
month shown   when its last day >= the earliest of ledgers
full          when it has ended and its first day >= begins
held_from     = begins, when begins falls inside the month
so_far        the current month
section       absent when full_months == 0
typical, average   over full months only, unset when full_months < 3
```

This is the month summary's rule for `prior_*`, extended: strict, so a bank
connected last month makes earlier months partly held. *Alternative:* the
chart's contributor rule, dropping short ledgers from the whole history.
Rejected: September's bar would then disagree with the month summary above it,
which counts every account.

The median of an even number of months is the mean of the middle two, in minor
units, rounded half away from zero; the average rounds the same way.

**Usual, for a merchant, is its median monthly total over the full months,
zeros included.** A merchant paid in three months of twelve has a usual of
nothing, and reads "not usually paid" rather than as a rise from zero. Risers
are the five largest positive `total − usual`, per month, from the same
outgoing rows the detectors read.

**Recurrence.** `banking.RecurringPayments(rows, today)` in
`apps/wimm/internal/banking/recurring.go`, over booked outgoing rows of one
currency:

```text
recurLookbackMonths        = 25   three yearly payments need two years and slack
recurMinOccurrences        = 3    weekly and monthly
recurMinOccurrencesYearly  = 2    likely = a yearly run holding exactly 2
recurAmountTolerance = 0.10    a run is amounts, sorted ascending, each no
                               more than 10% above the run's smallest
cadence   gap in days   slack   next expected
Weekly    6 to 8        2       last + 7 days
Monthly   27 to 34      5       same day next month, capped at its length
Yearly    351 to 379    14      same date next year

group on (lower-cased merchant name, amount run); within a group, newest
first, the cadence is the newest gap's; extend the run while each gap falls in
that cadence's range; recurring when the run holds its cadence's minimum
late      = expected < today <= expected + slack
dropped   = today > expected + slack
amount    = the newest payment's; account = the newest payment's
```

*Alternative:* trust a bank's standing-order or direct-debit marker. Rejected:
card subscriptions, which are the ones people forget, carry none.

Yearly stands at two because three need more than two years of ledger, which
few banks hand over, and the same merchant at the same amount a year apart is
already a strong sign. It is said as `Likely yearly` rather than passed off as
settled, and becomes `Yearly` on the third. *Alternative:* two for every
cadence. Rejected: two payments a week apart are a coincidence.

**One unusual-payment rule.** `banking.UnusualPayments(rows, floor)` in
`apps/wimm/internal/banking/unusual.go`:

```text
x          = ln(|amount|)
score      = unusualMADScale × (x − median(X)) / MAD(X)       0.6745
unusual    = score > unusualScore                              3.5
each direction is judged alone: "party" is the merchant for money out and
the payer for money in, by the same merchant name; "monthly" is money out
for one and money in for the other
baseline X = that party's booked rows in the lookback, same direction, when
             there are at least unusualMinPartyPayments        5
           = otherwise every booked row in that direction in the
             unusualGeneralDays up to and including the row    90
             and not judged when that holds fewer than 5 rows
MAD == 0   → unusual when |amount| >= unusualFlatRatio × median amount   3.0
first      = money out only: no earlier booked payment to that merchant in
             the lookback, and |amount| > unusualFirstPaymentShare ×
             median monthly money out                           0.10
             → unusual, first_payment set, typical unset
floor      = unusualFloorShare × median monthly money, same direction, over
             full months; |amount| < floor is never unusual    0.01
             no full month held → nothing is unusual
typical    = exp(median(X)), shown rounded to the currency unit
net_usual  = net with every unusual row of the month removed, both directions
judged     = payments inside the months shown; older rows are baseline only
```

This is the Iglewicz–Hoaglin modified z-score on log amounts. Log, because
payment sizes are multiplicative: €640 against a usual €60 and €6,400 against
€600 are the same surprise. Median and MAD, because the repair that is being
judged must not drag its own baseline up, which a mean and standard deviation
let it do. Money in is judged only against money in: set against payments
going out, every salary would be unusual. The first-payment share exists
because the general baseline is wide, often a MAD near 1 in log terms, so a
garage paid once would need to be dozens of times a usual payment to be marked
by the score alone, and it is exactly the payment a member expects to see
marked. *Alternative:* flag above the 95th percentile. Rejected: it marks
one payment in twenty by construction, in every month, including a month in
which nothing unusual happened, and the count would mean nothing. *Alternative:*
exempt recurring payments. Rejected: one rule was the brief, and a yearly
insurance premium really does bend its month; it can be both Yearly and
Unusual, and both are true.

Transactions judges under All always, so a row's mark does not depend on a
control on another screen. Overview judges within the scope in force, so under
All the two agree exactly, which the spec requires and a test holds.

**What moved a day.** `banking.DayMovers(rows)` in
`apps/wimm/internal/banking/movers.go`, over one day's booked rows on the
chart's contributing accounts, either direction:

```text
dayCoverShare = 0.80    dayMaxRows = 5
sort by |amount| descending; take rows until their |amounts| reach 80% of the
day's gross movement (the sum of every |amount|), never more than 5;
smaller = the count of rows not taken
```

**Movers ship with the trend response, not per hover.** `currencyTrend` already
holds every contributing account's rows since `bound`, so grouping them by day
costs no query; the ceiling is 90 days × 5 rows per currency. A request per
pointed day would put a network round trip behind an arrow key, and the
announcement would lag the marker. The one added cost is the unusual label:
`BalanceTrend` calls the same `patterns` helper `GetMonthHistory` does, so an
Overview load reads the 25-month row set twice, in parallel. *Alternative:* one
combined RPC. Rejected for now: the two responses are cached and invalidated
differently by nothing yet, and merging them is cheap later if the measurement
in the tasks says it matters.

**The popover is the readout, made visible.** `BalanceChart`'s live region
stays the single source of what is announced; the popover renders the same
fields beside the marker. It is not a dialog: no focus moves into it, nothing
in it is operable, Escape or a tap outside clears `marked`. The load passes
every string (`change`, each mover's name and amount, `smaller`), so the
component still formats nothing.

**A month's unusual payment links to `/transactions?page=<booking
date>.<id>`.** `page_start` reads inclusively from a cursor, so the payment is
the first row of the page it lands on. It is not a page of the index, and the
scrubber then shows no current dot until the member pages; that is the existing
behaviour for any `before`/`after` cursor.

**Month by month is a chart and a list over one selection.**
`molecules/MonthlyNetChart.svelte` draws one bar per month from a zero rule,
above it or below it, with a dashed rule at the typical month; so-far and
held-from bars are drawn at reduced opacity. It follows `BalanceChart`'s
contract: one tab stop, arrows, Home, End, a marker, a live readout, an
accessible name that is a sentence. Below it, one disclosure row per month
carries the net, the unusual count and the net without them, and opens to the
month's figures, its risers and its unusual payments. Pointing at a bar opens
that month's row: one selection, two views of it. The last full month is open
as the screen loads. *Alternative:* the list alone, with an inline bar per row.
Rejected: a trend across thirteen rows is read by scrolling, and the typical
month has nowhere to be drawn.

**Without unusual payments is component state.** Both sets of figures arrive in
one response, the control swaps which strings are shown, and nothing is
requested. It resets on load by design: the default reading is the true one.

**One ADR, for the detection rules.** Recurrence with yearly at two, the unusual-payment rule in both directions
with its first-payment share, and the explains-the-day rule, with every
constant above, why log and MAD, why not
a percentile, and why the marks are derived and never stored. The scope needs
none: it applies ADR 0021 and ADR 0024 and decides nothing new.

## Risks / Trade-offs

- [The general baseline is wide, so a merchant paid two to four times is marked
  only when a payment is dozens of times a usual one] → The first payment is
  caught by its own share of a month; the gap is the second to fourth, and the
  table test carries a realistic spread so the constants can be tuned.
- [A first payment of rent, or a deposit, is marked as a first payment] → True
  and useful: it is over a tenth of a month and new.
- [A transfer in from a member's own savings reads as unusual income] → The
  same limit the months already state: transfers between own accounts are
  counted. Setting it aside in the view is the right reading of that month.
- [A new landlord's rent is unusual for its first four months, until the
  merchant has five payments] → True, and it is a large new payment. It reads
  "Unusual" and, from the third, also Monthly.
- [A variable bill such as electricity misses the 10% tolerance and is not
  recurring] → A tolerance wide enough for it would call a weekly supermarket
  recurring. It is one constant.
- [Two unrelated payments to one merchant a year apart read as likely yearly]
  → They must also be within 10% of each other, and the words say likely.
- [Connecting a second bank turns earlier months into partly held ones and can
  drop the typical month for up to three months] → The bars stay, the sentence
  says why, and it returns on its own. It is the month summary's existing
  behaviour.
- [An Overview load reads 25 months of booked rows twice] → A store test
  seeds 50,000 rows and asserts a budget, as the page index's does. One
  combined read is the fallback.
- [Transactions runs the detector on every page] → Only when the page's dates
  fall inside the months Overview shows; older pages skip it.
- [Transfers between own accounts inflate money in and money out in every
  month, and under Household a transfer to a member's own account reads as
  money out] → Said on screen, as the month summary says it. Under Household it
  is also simply true: the money left the household's accounts.
- [Household scope is narrower than household money for a member who holds
  details on an account they do not own] → Named on screen, account by
  account.
- [Thirteen disclosure rows are long at Compact] → Closed rows are one line;
  the look decides whether Compact shows six and a way to the rest.

## Open Questions

- Whether Household should one day count household accounts a member holds
  *details* on and does not own. It would need ADR 0021 reopened and is
  deliberately not planned here.
