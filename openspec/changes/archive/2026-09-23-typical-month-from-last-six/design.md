## Context

See proposal.md for why. What shapes the approach:

- `wimmd` decides every figure; no client classifies or sums (ADR 0024). The typical and average month are computed once, in `apps/wimm/internal/banking/history.go`, as the median and mean of each full month's net, and sent on `CurrencyHistory` as `typical_net`, `average_net` and their unusual-payments-set-aside pair.
- A month is full when it is before the current month and not before the earliest ledger held in scope. So a full month can still lack an account that began later; `late_ledgers` names those, and Overview writes one sentence for each.
- `historyMonths` (13) and `minFullMonths` (3) sit in `unusual.go` as named constants, per ADR 0025.
- The web client writes `From {fullMonths} full months.` today, reading `full_months`. That count also drives the chart summary, the risers gate and the section's presence, so it cannot be repurposed.

## Language

- **the last six** — the newest six full months held, the ones the typical and average month are drawn from. In copy: `the last 6 full months`. Never "window", "period", "range" or "rolling".
- `typical_months` / `TypicalMonths` — how many full months the two figures were drawn from: six, or every full month held when that is fewer. Never "sample", "basis count".

## Goals / Non-Goals

**Goals:**

- Both figures and their set-aside pair from the same months, chosen in one place.
- The sentence says truthfully how many months stand behind them.

**Non-Goals:**

- Changing the months shown, their bars, or the chart's span.
- Changing the three-month threshold.
- Changing the unusual-payment rules' own baselines, including the median monthly money out that the first-payment and floor rules measure against, and the merchants' usual month behind a month's rise. Those are ADR 0025's and keep every full month shown.
- Dropping months an account is missing from. See decision 2.

## Decisions

### 1. Six, as a named constant beside the others

`typicalMonths = 6` in `unusual.go`, beside `historyMonths` and `minFullMonths`. The computation walks the months newest first and takes the first six marked full, for both the plain and the set-aside nets.

Six is the household's data, not a statistic: its later accounts begin on 19 Mar, so six full months is the largest count that holds them. It is also half the months shown, so an ordinary year still yields a figure from the recent half.

*Alternative: keep every full month shown.* Rejected: that is the problem.
*Alternative: three, the minimum.* Rejected: a median of three is one month, and one bad month is exactly what the typical month exists to shrug off.

### 2. The last six, not "months every account is in"

*Alternative: only months in which every account in scope holds a ledger.* The exact form of what the member wants. Rejected for now: Casa CC begins on 19 Sep, so the All view would state no typical month until January 2027, and every newly connected bank would blank the figure for three months. A fixed count degrades gracefully and never disappears on a connect. `late_ledgers` keeps saying which months miss which account, so nothing is hidden.

### 3. `typical_months` on the wire

`CurrencyHistory.typical_months` (int32, field 11): how many full months the two figures were drawn from. Unset when neither figure is. The web client writes `From the last {n} full months.` when `typical_months` is less than `full_months`, and `From {n} full months.` otherwise.

*Alternative: the client computes `min(fullMonths, 6)`.* Rejected: a second copy of the six, in another language, that drifts the first time the constant moves. ADR 0024 already rules that the server decides figures.

This is an additive field on a Connect message consumed only by `apps/web`, and it needs an ADR because it changes the contract and a recorded rule's inputs.

### 4. The chart's typical line and summary follow the figure

`MonthlyNetChart` draws the typical month as a dashed line across every bar, and the chart's spoken summary ends `A typical month is +€205.55.` Both keep doing so, at the value from the last six: they are the same figure the tile states, from the same field, so they cannot disagree. Nothing in `packages/ui` changes for this.

## Risks / Trade-offs

- [The last six still include months an account is partly missing from — March misses 18 days of Casa and Emanuel, and every month before 19 Sep misses Casa CC] → Accepted, and stated: `late_ledgers` sentences already name them under the figures. Decision 2 is the exact alternative, recorded for when it becomes the better trade.
- [Six recent months move more between visits than twelve] → Accepted: that is the point, recent months describe the household now.
- [Older full months stay on screen but stop feeding the figures] → The sentence says `the last 6`, so the reader is told which months count.

## Migration Plan

None. Additive field, derived on read; nothing stored. Rollback is reverting the commit.
