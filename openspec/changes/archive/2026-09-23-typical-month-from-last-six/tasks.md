## 1. Decision

- [x] 1.1 Record the decision with `/adr`: the typical and average month are drawn from the newest six full months held, or every full month held when fewer; `CurrencyHistory.typical_months` carries the count; the unusual-payment baselines and the merchants' usual month keep every full month shown. It covers design.md decisions 1 to 4, the rejected "months every account is in" rule and why (Casa CC begins 19 Sep, so the All view would state nothing until January 2027), and the known limit that the last six can still miss part of an account. Verify: `just adr-index` regenerates `.claude/rules/decisions.md` and `just adr-index-check` passes.

## 2. The figures

- [x] 2.1 Test-first (`/tdd`), in `apps/wimm/internal/banking/history_test.go`: with twelve full months where the eighth newest holds a large net, `TypicalNet` and `AverageNet` equal the median and mean of the newest six nets and `TypicalMonths` is 6; the same holds for `TypicalNetUsual` and `AverageNetUsual`; with four full months both use all four and `TypicalMonths` is 4; with two, all four figures are nil and `TypicalMonths` is 0; `FullMonths` and the months returned are unchanged in every case. Each test seen failing first. Verify: `just check apps/wimm`.
- [x] 2.2 Add `typicalMonths = 6` to `apps/wimm/internal/banking/unusual.go` beside `historyMonths` and `minFullMonths`, and `TypicalMonths int` to `CurrencyHistory`; in `history.go` take the nets of the first `typicalMonths` full months, newest first, for both pairs. Verify: the 2.1 tests pass and `just check apps/wimm` passes.

## 3. The contract

- [x] 3.1 Add `int32 typical_months = 11;` to `CurrencyHistory` in `packages/contracts/proto/wimm/banking/v1/banking.proto`, commented as how many full months the typical and average month are drawn from, unset when they are; run `just gen`. Verify: `git status` shows only generated files and the proto changed under `packages/contracts`, and `just check packages/contracts` passes.
- [x] 3.2 Map it in `apps/wimm/internal/rpc/banking.go` beside `FullMonths`, with a `toProtoCurrencyHistory` test in `banking_transaction_test.go` asserting the value is carried. Verify: `just check apps/wimm`.

## 4. The screen

- [x] 4.1 In `packages/ui/src/pages/Overview.fixture.ts` change `basis` to canvas.md's first Words line, and make `Populated`'s play function in `Overview.stories.svelte` assert it in full. `LedgerBeginsPartWayThrough` stays as it is. Verify: `just check packages/ui` and `just check apps/storybook` pass, and the check names `pages-overview--populated` among the stories with a new version.
- [x] 4.2 Look: open `Populated` and `LedgerBeginsPartWayThrough` on the design canvas (`just canvas`) at compact, medium, wide and ultra, change what is wrong, and write what was seen into canvas.md under Seen. Verify: Seen names both stories at all four regimes.

## 5. Wiring

- [x] 5.1 In `apps/web/src/lib/insights.ts`, write the basis from `typicalMonths`: `From the last ${n} full months.` when it is less than `fullMonths`, `From ${n} full months.` otherwise, each followed by the existing second sentence. Test-first in `apps/web/src/routes/(app)/overview.test.ts`: twelve full months with `typicalMonths` 6 gives the first form; four and four gives the second; six and six gives `From 6 full months.`; the existing under-three case is unchanged. Verify: `just check apps/web`.
- [x] 5.2 Confirm on the running app (`just up`) that Overview's All, Household and Yours views each read `From the last 6 full months.` and that the typical tile, the sentence under it and the chart's dashed line agree. Verify: write the three typical values seen into this task's commit message.
