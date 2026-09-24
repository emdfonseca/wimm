One sentence on Overview changes its words; no screen, state, component or layout changes. `Overview.svelte` takes the sentence as its `history.basis` prop and renders it as it does today, so the screen is untouched and only the fixture's words and the stories asserting them move.

## Screens

- `packages/ui/src/pages/Overview.svelte` with `Overview.stories.svelte` and `Overview.fixture.ts`. Exists. Unchanged in code.

## State stories

No new story. Two existing stories hold the two forms of the sentence:

| Story | File | Fixture | Scenario |
| --- | --- | --- | --- |
| `Populated` (changed words) | `packages/ui/src/pages/Overview.stories.svelte` | `basis` in `Overview.fixture.ts`, twelve full months | Living within what comes in |
| `LedgerBeginsPartWayThrough` (unchanged) | `packages/ui/src/pages/Overview.stories.svelte` | its own `basis`, four full months | Six or fewer full months held |
| `TwoFullMonths` (unchanged) | `packages/ui/src/pages/Overview.stories.svelte` | its own `waiting` | Two full months held |

## Words

- More than six full months held: `From the last 6 full months. The typical month is the middle one, so one exceptional month barely moves it.`
- Six or fewer held, unchanged: `From 4 full months. The typical month is the middle one, so one exceptional month barely moves it.`
- Under three held, unchanged: `A typical month and an average month appear once three full months are held.`

## Flow

Stands alone. `pages-overview--populated` is already a step of `see-where-the-money-is` in `apps/storybook/canvas/flows.js`; its place does not change.

## Surfaces

- The sentence: inline, under the typical and average month tiles, as today.

## Components used

`packages/ui/src/pages/Overview.svelte`, unchanged.

## Components missing

None. The sentence already has its slot.

## States left out

- **Exactly six full months held.** It reads `From 6 full months.`, the six-or-fewer form, which `LedgerBeginsPartWayThrough` already shows at four. The boundary is held by the web client's unit test instead.
- **The scenario "An exceptional month older than the last six".** It is arithmetic, not a state a person sees differently: the screen shows the same tiles with other numbers. Held by the Go unit test on the computation.

## Contracts for implementation

- `Populated`'s play function asserts the `From the last 6 full months.` sentence in full with `getByText`.
- `LedgerBeginsPartWayThrough` keeps asserting `/^From 4 full months\./`.

## Seen

Looked at on the design canvas, 23 Sep 2026, light theme.

- `Populated` at compact (390), medium (834), wide (1440) and ultra (1920): `From the last 6 full months. …` sits under the tiles and the typical-month sentence, two lines at compact and one line at the other three, with no horizontal overflow. Its tiles, sentence and the chart's dashed line showed the twelve-month +€189.40 and −€37.80, which contradicted the sentence. The fixture now carries the last six (March to August): typical +€174.45, average −€110.24, and without unusual payments +€211.85 and +€199.76. `AMonthWithUnusualIncomeOpened`'s average is +€1,523.76 for the same reason.
- `LedgerBeginsPartWayThrough` at compact, medium, wide and ultra: `From 4 full months. …` reads the same way, two lines at compact and one line elsewhere, with no overflow. Its tiles still show the twelve-month figures over four months. The story is unchanged by this change, so that is left as it is.
