# 0025 · Rules for recurring payments, unusual payments and what moved a day

## Status

Accepted

## Context

Overview, the balance chart and Transactions must mark the same payments for the
same reasons, and a member must be able to check any mark against a stated rule.
A standing-order or direct-debit marker from the bank is not usable: card
subscriptions, the ones people forget, carry none. Amounts are multiplicative,
so €640 against a usual €60 and €6,400 against €600 are the same surprise. A
payment being judged must not drag its own baseline up.

## Decision

Three pure functions in `apps/wimm/internal/banking`, each a table of named
constants. Nothing is stored: marks are derived on read, like the merchant name,
because the rules will change and a stored mark needs a backfill each time.

**Recurrence** (`RecurringPayments`), booked money out only, per currency.
`recurLookbackMonths` 25; `recurMinOccurrences` 3; `recurMinOccurrencesYearly`
2; `recurAmountTolerance` 0.10 (a run's amounts, sorted, each no more than 10%
above the smallest). Cadences: weekly gap 6 to 8 days, slack 2; monthly 27 to
34, slack 5; yearly 351 to 379, slack 14. Group on lower-cased merchant name and
amount run; the newest gap sets the cadence; extend while gaps stay in range. A
run is dropped once today is past expected plus slack, and late while past
expected within the slack. Yearly stands at two because three need more than two
years of ledger, which few banks give, and the same merchant at the same amount
a year apart is a strong sign; a yearly run of exactly two is `likely`, and a
third clears it. Weekly and monthly stay at three: two payments a week apart are
a coincidence. Money in is never recurring, because a salary is not a
commitment.

**Unusual** (`UnusualPayments`), booked rows, each direction judged only against
its own: money in against money in, or every salary would be unusual. Score is
`unusualMADScale` (0.6745) times (ln|amount| minus the median of ln|amount|)
over the MAD, unusual above `unusualScore` (3.5): the modified z-score on log
amounts, with median and MAD so the judged payment does not move its own
baseline. Baseline is the party's own rows in the lookback when there are
`unusualMinPartyPayments` (5), otherwise every row in that direction in the
`unusualGeneralDays` (90) up to the row, and not judged under five rows. A zero
MAD marks a row at `unusualFlatRatio` (3.0) times the median. A first payment
to a merchant, money out only, is unusual above `unusualFirstPaymentShare`
(0.10) of median monthly money out, because the general baseline is wide and a
garage paid once would otherwise need to be dozens of times a usual payment.
`unusualFloorShare` (0.01) of median monthly money in the same direction is a
floor under which nothing is unusual; with no full month, nothing is. Typical
is exp(median), and a first payment carries none. Judged rows are those inside
the `historyMonths` (13) shown; older rows are baseline only. `minFullMonths`
(3) full months are needed before a typical month is stated.

**Explains the day** (`DayMovers`): rows in descending size until they reach
`dayCoverShare` (0.80) of the day's gross movement, at most `dayMaxRows` (5); the
rest are counted as smaller.

A percentile was rejected: the top twentieth is unusual by construction in every
month, including one where nothing was. Exempting recurring payments was
rejected: one rule was the brief, and a yearly premium can be both.

## Consequences

Every threshold is one edit and one table test. Marks cannot disagree between
screens because one function makes them. A merchant paid two to four times is
marked only when far above payments in general, or on its first payment by
share. Tuning happens in these terms against real data. Marks are recomputed on
every read from up to 25 months of rows.
