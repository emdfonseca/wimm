## Why

Overview's typical month and average month are drawn from every full month shown, up to twelve. The household's banks did not return a year of ledger for every account: in the All view, Casa and Emanuel start on 19 Mar 2026 and Casa CC on 19 Sep. So half the months behind today's figures leave out Casa and Emanuel, and they pull both figures toward a household that is not the one being looked at. The newest six full months, March to August today, hold both of them for all but eighteen days.

## What Changes

- The typical month and the average month are drawn from the **newest six full months**, not every full month shown. With fewer than six held, they use every full month held, as now.
- Overview's sentence naming the basis says so: `From the last 6 full months.` when more than six are held, and `From 4 full months.` as it is today when six or fewer are held.
- The same six months feed the typical and average month with unusual payments set aside.
- `CurrencyHistory` gains `typical_months`: how many full months the two figures were drawn from. The web client words the sentence from it and never works out the six itself.
- Unchanged: the month-by-month chart and table still show up to thirteen months; three full months are still needed before either figure appears; the merchants behind a month's rise and the unusual-payment rules keep their own baselines.
- One ADR records the six and why.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `banking/overview`: Requirement "A typical month and an average month" changes which full months the two figures are drawn from, and what Overview says about it.

## Impact

- `apps/wimm/internal/banking`: a `typicalMonths` constant beside `historyMonths` and `minFullMonths`; the typical and average computation in `history.go` takes the newest six full months.
- `packages/contracts/proto/wimm/banking/v1/banking.proto`: `CurrencyHistory.typical_months`, regenerated with `just gen`. Additive.
- `apps/wimm/internal/rpc/banking.go`: maps the new field.
- `apps/web/src/lib/insights.ts`: the basis sentence.
- `packages/ui/src/pages/Overview.fixture.ts` and the Overview stories: the words of the basis sentence.
- `docs/decisions/`: one new ADR.
- No migration, no schema change, no new dependency.
