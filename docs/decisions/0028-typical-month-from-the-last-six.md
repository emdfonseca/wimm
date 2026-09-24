# 0028 · The typical month is drawn from the last six full months

## Status

Accepted

## Context

Overview states a typical month (the median of each full month's net) and an
average month (the mean) beside it, with the same pair again with unusual
payments set aside. `wimmd` computes all four in
`apps/wimm/internal/banking/history.go` and sends them on `CurrencyHistory`
(ADR 0024).

A full month is one before the current month and not before the earliest ledger
held in scope. Banks do not return a year of ledger for every account, so a full
month can still lack an account that began later. In the household's All view,
Casa and Emanuel begin on 19 Mar and Casa CC on 19 Sep: half of twelve full
months leave out Casa and Emanuel, and pull both figures toward a household that
is not the one being looked at.

The exact rule — only months in which every account in scope holds a ledger —
would state no typical month in the All view until Casa CC held three full
months, and every newly connected bank would blank the figure for three months.

Three, the minimum, makes the median one month, and one bad month is what the
typical month exists to shrug off.

## Decision

The typical month, the average month and their set-aside pair are drawn from the
newest six full months held, or from every full month held when fewer. Six is
`typicalMonths` in `unusual.go`, beside `historyMonths` (13) and `minFullMonths`
(3); the computation walks the months newest first and takes the first six
marked full. Six is the largest count that holds the household's later accounts,
and half the months shown.

`CurrencyHistory.typical_months` (field 11) says how many full months the figures
were drawn from, unset when there are none. The web client writes
`From the last {n} full months.` when it is less than `full_months`, and
`From {n} full months.` otherwise. It never computes the six itself: a second
copy of the constant, in another language, drifts the first time it moves.

The months shown, the chart's span, the three-month threshold, the unusual-payment
baselines (including the median monthly money out behind the first-payment and
floor rules) and the merchants' usual month behind a month's rise are unchanged
and keep every full month shown, under ADR 0025. The chart's dashed typical line
and its spoken summary read the same field as the tile, so they follow it.

## Consequences

- An exceptional month older than the last six still shows among the months and
  no longer moves either figure.
- The last six can still miss part of an account: March misses 18 days of Casa
  and Emanuel, and every month before 19 Sep misses Casa CC. `late_ledgers`
  sentences name them under the figures.
- Six recent months move more between visits than twelve would.
- A connected bank never blanks the figure; the exact every-account rule stays
  available for when it becomes the better trade.
- The field is additive on a Connect message consumed only by `apps/web`.
