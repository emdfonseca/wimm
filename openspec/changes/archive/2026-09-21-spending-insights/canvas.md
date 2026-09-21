## Screens

- **Overview.** `packages/ui/src/pages/Overview.svelte` and
  `packages/ui/src/pages/Overview.stories.svelte`. Both exist and are altered:
  a scope control, a Month by month card and a Recurring payments card are
  added, and the balance chart's pointed day opens a popover.
- **Transactions.** `packages/ui/src/pages/TransactionsScreen.svelte` and
  `packages/ui/src/pages/TransactionsScreen.stories.svelte`. Both exist and are
  altered: a row may carry the word Unusual.

Two existing components change with them and carry their own state stories:
`packages/ui/src/molecules/BalanceChart.svelte` and
`packages/ui/src/molecules/LedgerRow.svelte`. Two are new, under Components
missing.

## State stories

`Overview` is `packages/ui/src/pages/Overview.stories.svelte`, `Transactions`
is `packages/ui/src/pages/TransactionsScreen.stories.svelte`, `BalanceChart`,
`LedgerRow`, `MonthlyNetChart` and `MonthRow` are the `.stories.svelte` beside
each molecule. Scenarios are `banking/overview`'s unless a capability is named.
Today in every fixture is 20 September 2026.

| Story | File | Fixture | Scenario |
| --- | --- | --- | --- |
| `Populated` (changed) | Overview | As it stands, plus: scope control on All; thirteen months, September 2025 to September 2026 so far, twelve full, March holding two unusual payments and December one; typical and average month; August open with three merchants that rose; six recurring payments, one late. | A year of months; A month that ended below zero; Living within what comes in; The two figures disagree; The month so far is a bad one; Fuel cost more in August; A merchant they do not usually pay; Several payments to one merchant; What is due and when; One is late; A month with a repair in it; A month with none; Still in the other lists; As it opens; Pending payments are not counted. `banking/payment-patterns`: every scenario that ends in a payment being listed or not, as fixture rows. |
| `ADayPointedAt` (changed) | Overview | `Populated`, with the chart's marker on 3 Aug: two movers, one unusual, three smaller. | What happened on the day it dropped; One of them is unusual |
| `AMonthWithUnusualPaymentsOpened` | Overview | `Populated`, with March 2026 selected. | Opening that month; Going to the payment; Reading one month |
| `AMonthWithUnusualIncomeOpened` | Overview | `Populated`, with May 2026 holding a €9,804.00 bonus and selected. | A month with a bonus in it; Setting a bonus aside. `banking/payment-patterns`: A bonus |
| `ALikelyYearlyPayment` | Overview | `Populated`, with `Fidelidade` paid twice, a year apart. | A yearly payment seen twice. `banking/payment-patterns`: An insurance premium paid twice |
| `AMonthWithNoneOpened` | Overview | `Populated`, with June 2026 selected: no unusual payments and nothing that rose. | A month with none; Nothing rose |
| `TheMonthSoFarOpened` | Overview | `Populated`, with September 2026 selected. | Reading the month so far |
| `WithoutUnusualPayments` | Overview | `Populated`, with the view switched. | Setting the repair aside; The payments are still there |
| `MoreGoesOutThanComesIn` | Overview | `Populated`, with a typical month of −€212.40 and an average of −€260.15. | More goes out than comes in |
| `TwoFullMonths` | Overview | `Populated`, with a ledger from 1 July: July and August full, September so far, no typical or average, no risers, nothing unusual, nothing recurring. | Two full months held; Nothing to set aside; Nothing recurs |
| `AYoungAccountBesideOlderOnes` | Overview | `Populated`, with an account whose ledger begins 19 September beside older ones. | A second bank connected last month |
| `LedgerBeginsPartWayThrough` | Overview | `Populated`, with a ledger from 12 April: April held from 12 April, four full months. | A shorter ledger; A second bank connected last month |
| `HouseholdScope` | Overview | `Populated` under Household: every section counts Joint account only; three recurring payments. | Are we overspending; The figures do not move; Under another scope |
| `HouseholdScopeIsNarrower` | Overview | `HouseholdScope`, plus `Joint savings` in the Household group, not owned. | Household money they do not own |
| `YoursScope` | Overview | `Populated` under Yours: the two own accounts only. | Am I overspending |
| `HouseholdScopeWithNoHistory` | Overview | `HouseholdScope`, whose accounts hold no full month, no recurring payment and no unusual payment. | A scope with no full month; A scope with no recurring payment; A scope with nothing to set aside |
| `HouseholdScopeWithNothingToSetAside` | Overview | `HouseholdScope`, with a typical month and no unusual payment. | A scope with nothing to set aside |
| `LedgerStartsThisMonth` (changed) | Overview | As it stands. Asserts the absences below. | No full month yet; `banking/payment-patterns`: Too little history |
| `NoTransactionHistory`, `OwnsNoAccount` (changed) | Overview | As they stand. Assert the absences below. | A member who owns nothing (both requirements) |
| `HouseholdOfOne` (changed) | Overview | As it stands, with the new sections. Asserts no scope control. | A household of one; Nothing of their own |
| `TwoCurrencies` (changed) | Overview | As it stands, with months and recurring payments in EUR and GBP. | Two currencies (both requirements) |
| `ADayPointedAt` (changed) | BalanceChart | The marker on 3 Aug, with movers. | What happened on the day it dropped |
| `ADayWithNoTransactions` | BalanceChart | The marker on 9 Aug. | A day when nothing happened |
| `MoneyArriving` | BalanceChart | The marker on 15 Sep. | Money arriving |
| `UnusualIncomeArriving` | BalanceChart | The marker on 28 Aug, the day a bonus arrived. | A bonus arrives |
| `TheFirstDay` | BalanceChart | The marker on 22 Jun. | The requirement's last sentence |
| `ABusyDay` | BalanceChart | The marker on a day with twelve payments. | A busy day |
| `OpensFromTheKeyboard` | BalanceChart | `AsItOpens`; the play function focuses and presses the keys. | By keyboard |
| `OpensFromATap` | BalanceChart | `AsItOpens`; the play function taps the plot, then outside it. | By touch |
| `AsItOpens`, `AMonthPointedAt`, `AMonthBelowZero`, `PartlyHeldMonths`, `NoTypicalMonthYet` | MonthlyNetChart | Thirteen months; the marker on March; two full months. | A year of months; Reading one month; A month that ended below zero; A shorter ledger; Two full months held |
| `Row`, `RowWithUnusual`, `Selected`, `SoFar`, `HeldFrom`, `Choosing` | MonthTable | Thirteen rows, or one, as named. | A month with a repair in it; A month with none; Selecting that month |
| `Detail`, `DetailWithUnusual`, `DetailWithUnusualIncome`, `DetailSoFar` | MonthDetail | One month each. | Selecting that month; A month with a bonus in it |
| `Unusual`, `CompactUnusual`, `UnusualIncome`, `ExpectedDate`, `WithANoteAndALink` | LedgerRow | One row each. | `banking/transactions`: Finding the repair in the ledger; A bonus in the ledger. `banking/overview`: What is due and when; Opening that month |
| `AsItOpens` (changed) | Transactions | As it stands, with `Galp` marked Unusual. | `banking/transactions`: Finding the repair in the ledger; A payment not yet settled |

## Words

Amounts, dates, merchants and account names are fixture data, written here so a
play function asserts the sentence they sit in. No state shows the words
"outlier", "anomaly", "subscription", "earn", "expenses", "budget", "filter",
"probably" or "windfall", and "income" appears only in the label
`Unusual income`.

**`Populated`**, beyond what it shows today.

- Scope control, named `Accounts counted`: `Household`, `Yours`, `All`, with
  `All` chosen. No sentence under it.
- Card title `Month by month`, span `September 2025 to September 2026`.
- Tiles: `Typical month` over `+€189.40`; `Average month` over `−€37.80`.
- Under them: `In a typical month €189.40 more comes in than goes out.` and
  `From 12 full months. The typical month is the middle one, so one
  exceptional month barely moves it.`
- View control, named `Payments counted`: `All payments`,
  `Without unusual payments`, with `All payments` chosen.
- Chart's accessible name: `Net by month from September 2025 to September
  2026. 8 of 12 full months ended with more in than out. Highest +€310.00 in
  February 2026. Lowest −€1,480.30 in March 2026. A typical month is
  +€189.40.`
- Month rows, newest first: `September 2026`, `So far`, `+€637.36`;
  `August 2026`, `+€305.40`; `July 2026`, `−€60.45`; `June 2026`, `+€150.20`;
  `May 2026`, `+€225.00`; `April 2026`, `+€198.70`; `March 2026`,
  `−€1,480.30`, `2 unusual · +€379.70 without them`; `February 2026`,
  `+€310.00`; `January 2026`, `+€240.55`; `December 2025`, `−€640.00`,
  `1 unusual · −€160.00 without it`; `November 2025`, `−€95.20`;
  `October 2025`, `+€180.10`; `September 2025`, `+€212.40`. One unusual payment
  reads `1 unusual` and `without it`, never `1 unusuals` or `without them`.
- August is selected: `Money in` `€2,450.00`, `Money out` `€2,144.60`, `Net`
  `+€305.40`; heading `More than usual`; `Galp` with
  `€246.80 · €82.40 more than usual`; `Amazon` with
  `€188.20 · €64.10 more than usual`; `Zara` with
  `€120.00 · not usually paid`.
- Under the rows: `Money moved between your own accounts is counted.`
- Card title `Recurring payments`, then, soonest first: `Spotify`,
  `Monthly · Current account · Monzo`, `Was expected 18 Sep`, `−€9.99`;
  `Limpeza Casa`, `Weekly · Joint account · Monzo`, `Expected 24 Sep`,
  `−€45.00`; `Rent`, `Monthly · Joint account · Monzo`, `Expected 6 Oct`,
  `−€820.00`; `Netflix`, `Monthly · Current account · Monzo`,
  `Expected 14 Oct`, `−€12.99`; `NOS`, `Monthly · Conta à Ordem · Montepio`,
  `Expected 17 Oct`, `−€39.99`; `Fidelidade`,
  `Yearly · Joint account · Monzo`, `Expected 3 Mar 2027`, `−€386.00`.
- Absent: any control named Cancel, Pause, Hide or Dismiss, and any total over
  the recurring payments.

**`ADayPointedAt`** (Overview and BalanceChart). `3 Aug`; `€9,870.12`;
`€805.79 less than 2 Aug`; `Leroy Merlin`, `−€640.00`, `Unusual`; `Galp`,
`−€92.10`; `and 3 smaller`. The live region reads, as one string: `3 Aug.
€9,870.12. €805.79 less than 2 Aug. Leroy Merlin −€640.00, Unusual. Galp
−€92.10. And 3 smaller.` The line `3 Aug · €9,870.12` under the dates is gone:
the popover is that readout.

**`ADayWithNoTransactions`.** `9 Aug`; `€9,064.33`; `No change from 8 Aug`;
`No transactions this day.`

**`MoneyArriving`.** `15 Sep`; `€11,748.72`; `€2,437.01 more than 14 Sep`;
`Salary`, `+€2,450.00`; `and 1 smaller`. One reads `and 1 smaller`.

**`TheFirstDay`.** `22 Jun`; `€9,640.00`; no line naming a day before; then its
rows or `No transactions this day.`

**`UnusualIncomeArriving`.** `28 Aug`; `€12,640.18`; `€9,804.00 more than
27 Aug`; `Employer Lda`, `+€9,804.00`, `Unusual income`. The live region ends
`Employer Lda +€9,804.00, Unusual income.`

**`ABusyDay`.** Five rows, then `and 7 smaller`.

**`AMonthWithUnusualPaymentsOpened`.** `Populated`'s words with March selected
instead of August: `Money in` `€2,450.00`, `Money out` `€3,930.30`, `Net`
`−€1,480.30`, `Without unusual payments` `+€379.70`; heading `More than
usual`: `Auto Reparadora` with `€1,650.00 · not usually paid`, `Galp` with
`€318.40 · €154.00 more than usual`; heading `Unusual payments`: `12 Mar`,
`Auto Reparadora`, `−€1,650.00`, `Unusual`, `first payment to Auto Reparadora`;
`27 Mar`, `Galp`, `−€210.00`, `Unusual`, `usually about €60`. Each of the two
is a link. `usually about` never appears beside a first payment.

**`AMonthWithUnusualIncomeOpened`.** `Populated`'s words, except: `Average
month` over `+€779.20`, `Typical month` still over `+€189.40`; the row
`May 2026`, `+€10,029.00`, `1 unusual · +€225.00 without it`; and May selected:
`Money in` `€12,254.00`, `Money out` `€2,225.00`, `Net` `+€10,029.00`,
`Without unusual payments` `+€225.00`; heading `Unusual payments`: `22 May`,
`Employer Lda`, `+€9,804.00`, `Unusual income`, `usually about €2,450`. The
view control is present, and with it switched the average month reads
`+€157.20` again.

**`ALikelyYearlyPayment`.** `Populated`'s words, with the last recurring
payment reading `Fidelidade`, `Likely yearly · Joint account · Monzo`,
`Expected 3 Mar 2027`, `−€386.00`. Absent: `Yearly ·` on that row.

**`AMonthWithNoneOpened`.** June selected: `Money in` `€2,450.00`, `Money out`
`€2,299.80`, `Net` `+€150.20`; `Nothing took more than usual this month.`
Absent inside the row: `Unusual payments`, `unusual`, `Without unusual
payments`.

**`TheMonthSoFarOpened`.** September selected: `1 to 20 Sep`, `Money in`
`€2,450.00`, `Money out` `€1,812.64`, `Net` `+€637.36`, and `A month under way
is not set against whole months.` Absent: `More than usual`.

**`WithoutUnusualPayments`.** `Populated`'s words, except: `Without unusual
payments` is chosen; `Typical month` over `+€205.55`; `Average month` over
`+€157.20`; `In a typical month €205.55 more comes in than goes out.`; and
under the view control `3 unusual payments are set aside. They are still listed
in their months.` March still reads `2 unusual · +€379.70 without them`. The
chart's accessible name ends `Lowest −€160.00 in December 2025. A typical month
is +€205.55.`

**`MoreGoesOutThanComesIn`.** `Typical month` over `−€212.40`; `Average month`
over `−€260.15`; `More goes out than comes in. In a typical month €212.40 more
goes out than comes in.`

**`TwoFullMonths`.** `Month by month`, span `July 2026 to September 2026`,
three month rows, and `A typical month and an average month appear once three
full months are held.` Absent: `Typical month`, `Average month`, `Payments
counted`, `More than usual`, `Recurring payments`.

**`AYoungAccountBesideOlderOnes`.** `Populated`'s words, and under the tiles
`Months before 19 Sep do not include CLASSIC CEMG.` Absent: `No full month is
held`.

**`LedgerBeginsPartWayThrough`.** Span `April 2026 to September 2026`; the row
`April 2026`, `Held from 12 Apr`; `From 4 full months.`

**`HouseholdScope`.** `Household` chosen. `Recurring payments` holds
`Limpeza Casa`, `Rent` and `Fidelidade` only. `Household money` and
`Your money` read as in `Populated`. No sentence under the control.

**`HouseholdScopeWithNoHistory`.** `Household` chosen. `Month by month` with
`No full month is held for these accounts yet.` Absent: `Typical month`,
`Average month`. `Recurring payments` with `No recurring payments in these
accounts.` `Payments counted` absent, because no typical month is stated. In a scope
holding a typical month and no unusual payment, `Payments counted` is present and
`No unusual payments to set aside.` shows under it.

**`HouseholdScopeIsNarrower`.** `HouseholdScope`'s words, and under the
control: `Household counts Joint account. Joint savings is household money too,
but it is not yours, so its transactions are not counted.`

**`YoursScope`.** `Yours` chosen. `Recurring payments` holds `Spotify`,
`Netflix` and `NOS` only.

**`LedgerStartsThisMonth`, `NoTransactionHistory`, `OwnsNoAccount`.** Their
words as they stand. Absent in all three: `Month by month`, `Recurring
payments`, `Typical month`, `Payments counted`. Absent in the last two:
`Accounts counted`.

**`HouseholdOfOne`.** Its words as they stand, with `Month by month` and
`Recurring payments`. Absent: `Accounts counted`.

**`TwoCurrencies`.** `Month by month · EUR`, `Recurring payments · EUR`, and
both with `GBP`. One scope control, above the first currency.

**MonthlyNetChart.** `AMonthPointedAt`: readout `March 2026 · −€1,480.30`.
`NoTypicalMonthYet`: the accessible name ends at the lowest month and names no
typical month.

**MonthTable.** `Row`: `August 2026`, `+€305.40`. `RowWithUnusual`:
`March 2026`, `−€1,480.30`, `2 unusual · +€379.70 without them`. `SoFar`:
`September 2026`, `So far`, `+€637.36`. `HeldFrom`: `April 2026`,
`Held from 12 Apr`, `+€41.10`. `Selected`: one row has `aria-pressed` true.
`Choosing`: Enter and Space on a row select it and clear the previous one.
**MonthDetail.** `Detail` and `DetailWithUnusual`: the August and March panels
above. `DetailWithUnusualIncome`: the May panel. `DetailSoFar`: the September
panel.

**LedgerRow.** `Unusual` and `CompactUnusual`: `Galp`, `−€210.00`, `Unusual`.
`UnusualIncome`: `Employer Lda`, `+€9,804.00`, `Unusual income`.
`ExpectedDate`: `Spotify`, `Monthly · Current account · Monzo`,
`Was expected 18 Sep`, `−€9.99`. `WithANoteAndALink`: `Galp`, `27 Mar`,
`−€210.00`, `usually about €60`, the whole row one link.

**Transactions `AsItOpens`.** Its words as they stand, and `Unusual` on the
`Galp` row and on no other. `Transfer to Ana Reis` still reads `Not settled`
and not `Unusual`.

## Flow

Every state here stands alone. `connect-a-bank` still opens on
`pages-overview--before-any-bank-connected`, and `routes` in
`apps/storybook/canvas/flows.js` still resolves `/` to
`pages-overview--populated` and `/transactions` to
`pages-transactionsscreen--as-it-opens`, so those two keep their names.

## Surfaces

- Overview stays a full page, route-backed. The scope is part of its address:
  `/` is All, `/?scope=household` and `/?scope=yours` are the others.
- Without unusual payments is inline and ephemeral: component state, gone on
  reload.
- A month is selected in a table beside or below one detail panel: inline,
  non-modal, ephemeral. One is selected at a time, and pointing at a bar
  selects its row.
- The chart's popover is inline, non-modal and ephemeral. It is not a dialog:
  focus stays on the chart and nothing inside it is operable.
- An unusual payment in a month leaves the page for
  `/transactions?page=<date>.<id>`.

## Components used

- `packages/ui/src/atoms/SegmentedControl.svelte`: the scope control and the
  view control. One tab stop, arrow keys, one choice at all times, which is
  what both are. The scope control is given only the choices on offer.
- `packages/ui/src/molecules/MetricTile.svelte`: Typical month and Average
  month, with no Delta row and no link.
- `packages/ui/src/molecules/BudgetMeter.svelte`: the merchants under `More
  than usual`, the track filled against the largest rise in that month.
- `packages/ui/src/molecules/LedgerRow.svelte`: recurring payments
  (`description` the name, `account` the cadence and the account, `date` the
  expected date, `hideStatus`) and a month's unusual payments (`note`, `href`).
  It changes; see below.
- `packages/ui/src/molecules/BalanceChart.svelte`: changes; see below.

**Changed, not missing.**

- `BalanceChart.svelte`: `BalancePoint` gains `change?`, `movers?`
  (`name`, `amount`, `negative`, `unusual`), `smaller?` and `empty?`, all
  already written by the load. The readout under the dates becomes a popover
  beside the marker, and the live region carries the whole of it.
- `LedgerRow.svelte`: `unusual?: boolean` puts `Unusual` in the status slot,
  which a booked row never otherwise uses, or `Unusual income` when the row is
  not `negative`; `note?: string` is a second
  line under the name in the bank's line's place, never beside one; `href?`
  makes the row a link; the date column fits `Was expected 18 Sep`.

## Components missing

**Monthly net chart** (molecule).
`packages/ui/src/molecules/MonthlyNetChart.svelte` and
`MonthlyNetChart.stories.svelte`. One bar per month from a zero rule, above or
below it, a dashed rule at the typical month, month labels, a marker and a live
readout. Read before deciding: `molecules/BalanceChart.svelte` draws one
continuous line over daily points on a scale that is never forced to zero, and
a month's net is a signed quantity whose whole meaning is which side of zero it
is on. `molecules/TrendSparkline.svelte` draws bars from a zero baseline but
upward only, at a fixed 48 px, with no labels, no negative values and nothing
to point at. `molecules/BudgetMeter.svelte` is one horizontal proportion with
no sign. Nothing draws a signed series.

**Month table** (molecule). `packages/ui/src/molecules/MonthTable.svelte` and
`MonthTable.stories.svelte`. A list of button rows with `aria-pressed`, each
holding the month's name, a state word (`So far`, `Held from 12 Apr`), money
in, money out, its net and its unusual line. **Month detail** (molecule).
`molecules/MonthDetail.svelte` and its stories: the selected month's figures,
risers and unusual payments. Read before deciding: `molecules/LedgerRow.svelte`
is one transaction, not operable. `molecules/AccountRow.svelte` is a link that
leaves the page. `molecules/MetricTile.svelte` holds one figure and may be a
link. Nothing selects one of several rows and shows it elsewhere.

## States left out

- **Each rule's cases.** A price that rose a little, only twice so far, a day
  moved by a weekend, two subscriptions with one merchant, a cancelled
  subscription, the third year confirming a yearly payment, a first payment of
  an ordinary size, a merchant paid a few times, the salary itself, large and
  usual, far above usual and small, a fixed price that
  tripled, a quiet month. What is recurring and what is unusual
  is the service's, held by table tests. The screen lists what it is given.
- **A scope that no longer applies** and **Reloading**. Routing, held by the
  load's tests.
- **Coming back** to all payments. It is the default state, which is
  `Populated`.
- **An account the chart leaves out.** An absence from a popover, which nothing
  distinguishes from `ADayPointedAt`.
- **`banking/transactions`: Overview and Transactions agree; Two years ago;
  Arriving from Overview.** The first two are the service's. The third is a
  route, and the page it lands on is `AsItOpens` with other rows.
- **A month below zero as its own Overview story.** March and December in
  `Populated` are below zero.

## Contracts for implementation

- Order at Medium, Wide and Ultra: page header, the two money tiles, **the
  scope control**, the three month tiles, summary note, Balance chart, **Month
  by month**, **Recurring payments**, Top merchants and Accounts, Largest
  payments and Recent transactions. Compact keeps Accounts directly after the
  Balance chart, as it is today, and the two new cards follow it. Two
  currencies repeat the new cards per currency with the rest; the scope control
  appears once.
- The new cards span the Main column and are not capped by
  `layout-content-max`. Body padding and gaps are the page's existing ones:
  `space-8` and 24, Compact `space-4` and 16.
- The scope control calls `onscope` with `household`, `own` or `all`. It
  renders only the choices it is given and is absent when given none. Choosing
  the chosen one calls nothing.
- The view control is absent when no month holds an unusual payment. It swaps
  strings the screen already holds and calls nothing.
- Monthly net chart card: the Balance chart card's own measurements, padding
  16, gap 12, a plot row 200 high with a 72 wide scale and a gap of 8, and a
  labels row inset 80 so the first month sits under the plot. Bars are
  `color-chart-1`; a so-far or held-from bar is `color-chart-1` at 40%
  opacity; the zero rule is 1 px `color-border-strong`; the typical rule is
  1 px dashed `color-text-secondary`. The scale is `[lowest, highest]` padded
  8% and always includes zero. The width of a bar within its month is the
  look's to settle and is written into Seen.
- The chart is one tab stop. Left and right move a month, Home and End go to
  the ends, and moving the marker opens that month's row. Its readout is mono,
  `month · net`, in an `aria-live="polite"` region. Its accessible name is the
  sentence under Words, passed in already written.
- Direction is never colour alone: a bar is above or below the zero rule, and
  every net carries `+` or `−`.
- Month table row: 56 high, Compact 64. It is a `button` with `aria-pressed`
  and `aria-controls` naming the detail panel. Enter and Space select it, which
  clears the previous one. Focus stays on the button. The unusual line is
  `color-text-secondary`. Money in and money out are hidden at Compact.
- Inside the detail panel: the three figures in the mono family; `More than usual`
  as budget meters, 28 high with a 14 gap; `Unusual payments` as ledger rows 56
  high, Compact 80 with the note.
- Balance chart popover: opens when `marked` is set, by pointer move, by tap,
  and by left, right, Home and End; Escape, a tap outside the plot, and the
  pointer leaving clear it, except that a tap keeps it until the next tap. It
  sits beside the marker and flips to the marker's other side rather than leave
  the plot. It is `aria-hidden`: the live region beside it carries the same
  words as one string, in the order date, balance, change, movers, smaller. No
  focus moves into it. `Unusual` and `Unusual income` are text in
  `color-text-secondary`, not a colour.
- A recurring payment row is 56 high, Compact 64, and is not a link. A late one
  differs only in its words.
- The words around a figure, `more than usual`, `not usually paid`, `without
  them`, `without it`, `less than`, `more than`, `No change from`, `and N
  smaller`, `first payment to`, `Likely yearly`, `Unusual income`, `Expected`,
  `Was expected`, `Held from` and `So far`, are asserted
  by play functions on the fixture's whole string, since nothing else holds
  them.

## Seen

Each state story was opened at Compact 390, Medium 834, Wide 1440 and Ultra
1920, on fixtures, in the Storybook iframe at the regime's width. The whole page
was read for `Populated` at all four; every other Overview story was read at the
region it changes (the tiles and control, the month card, the recurring card),
because the rest of its page is `Populated`'s.

**Settled.**

- Bar width: 60% of its month's slot, with a slot never wider than 120 px, so
  three months are three bars and not three slabs. Twelve months at Wide and
  Ultra fill the card.
- Compact shows every month row. Thirteen one-line rows read as one list and
  none needs a way to the rest.
- Compact thins the month labels to every other one when more than eight are
  drawn. Thirteen three-letter labels ran together.
- Overview at every regime reads in the order the Contracts give.

**Changed.**

- `MonthlyNetChart`: header wraps at Compact so the span sits under the title;
  label thinning above; slot cap above.
- `BalanceChart`: at Compact the popover ran off the plot and hid the merchant
  name. It now takes the room on its side of the marker, is sized by its content,
  and wraps a mover's tag under its amount rather than dropping the name.
- `LedgerRow`: the compact date stays on one line, so `Was expected 18 Sep` is
  never split.
- `MonthRow` (since replaced by `MonthDetail`): unusual payments lined up with
  the panel's text, and gave the note room beside a short account so `first
  payment to Auto Reparadora` was whole at Medium.

**Overview.**

- `Populated`: Compact, Medium, Wide, Ultra read whole. Scope control sits under
  the two money tiles; the typical and average tiles, both sentences, the view
  control, the chart, thirteen rows with August open, the recurring card and the
  merchants and accounts cards follow. At Compact the two new cards follow Accounts.
- `ADayPointedAt`: Compact, Medium, Wide, Ultra. The popover shows `3 Aug`,
  `€9,870.12`, the change, `Leroy Merlin` with `Unusual`, `Galp`, `and 3 smaller`.
- `AMonthWithUnusualPaymentsOpened`, `AMonthWithUnusualIncomeOpened`,
  `ALikelyYearlyPayment`, `AMonthWithNoneOpened`, `TheMonthSoFarOpened`: Compact,
  Medium, Wide, Ultra; the words are those of Words.
- `WithoutUnusualPayments`, `MoreGoesOutThanComesIn`, `TwoFullMonths`,
  `LedgerBeginsPartWayThrough`: Compact, Medium, Wide, Ultra.
- `HouseholdScope`, `HouseholdScopeIsNarrower`, `YoursScope`: Compact, Medium,
  Wide, Ultra. The recurring card holds three rows each, as Words gives.
- `LedgerStartsThisMonth`, `NoTransactionHistory`, `OwnsNoAccount`,
  `HouseholdOfOne`, `TwoCurrencies`: Compact, Medium, Wide, Ultra. The absences
  hold; `HouseholdOfOne` has no scope control; `TwoCurrencies` has one, above the
  first currency.

**Molecules.**

- `BalanceChart`: `ADayPointedAt`, `ADayWithNoTransactions`, `MoneyArriving`,
  `UnusualIncomeArriving`, `TheFirstDay`, `ABusyDay`, `OpensFromTheKeyboard`,
  `OpensFromATap` at all four regimes. The two interaction stories end with the
  popover cleared, as their play functions leave them.
- `MonthlyNetChart`: `AsItOpens`, `AMonthPointedAt`, `AMonthBelowZero`,
  `PartlyHeldMonths`, `NoTypicalMonthYet` at all four regimes.
- `LedgerRow`: `Unusual`, `CompactUnusual`, `UnusualIncome`, `ExpectedDate`,
  `WithANoteAndALink` at all four regimes.

**Transactions.** `AsItOpens` at all four regimes: `Unusual` on the `Galp` row
only, `Not settled` on the transfer.

**Left as it is.**

- Non-compact `LedgerRow` stories are drawn with `compact` off, so at the
  Compact regime their rows clip. Overview passes `compact` and reads correctly.
- At Compact a recurring row's account is cut to an ellipsis on the longest
  cadence lines (`Monthly · Current account · …`) so the date stays whole.
- `Populated`'s meters are filled by the merchant's total and listed Galp,
  Amazon, Zara, while Zara's `not usually paid` €120.00 is the largest rise;
  Words lists them in that order, so the fixture was not changed.

**Months as a table with one detail panel.** Looked at on 21 September 2026 as
the canvas's own artboards, the story frame at 390, 834, 1440 and 1920 wide.
The canvas page itself draws nothing in a background tab and the browser
extension could not capture a page holding frames, so `MonthTable` was seen
through the extension at all four regimes and everything else below was
captured with the repository's Playwright at the four artboard sizes and read
from those captures.

- `MonthTable`: `Row`, `RowWithUnusual`, `Selected`, `SoFar`, `HeldFrom`,
  `Choosing` at all four regimes. One fixture under all six: `So far` and
  `Held from 12 Apr` sit beside the month, the unusual line under it, August
  tinted as the selection. Compact shows month and net only; Medium, Wide and
  Ultra add In and Out as right-aligned mono columns, which read evenly at
  Medium. Alone at Wide and Ultra the figures stand far from the name, which
  Overview never shows because the table takes three fifths of the row there.
- `MonthDetail`: `Detail`, `DetailWithUnusual`, `DetailWithUnusualIncome`,
  `DetailSoFar` at all four regimes. Figures in a row, `Without unusual
  payments` wrapping under them at Compact; meters; `first payment to Auto
  Reparadora` and `usually about €60` under their names from Medium up.
- Overview, at all four regimes: `Populated`, `AMonthWithUnusualPaymentsOpened`,
  `AMonthWithUnusualIncomeOpened`, `AMonthWithNoneOpened`,
  `TheMonthSoFarOpened`, `WithoutUnusualPayments`,
  `AYoungAccountBesideOlderOnes`, `HouseholdScopeWithNoHistory`,
  `HouseholdScopeWithNothingToSetAside`, `ReadingMonthsByPointing`,
  `TwoFullMonths`, `LedgerBeginsPartWayThrough`, `TwoCurrencies`. The chart's
  marker, the tinted row and the detail's heading name the same month in every
  one. `AYoungAccountBesideOlderOnes` reads `Months before 19 Sep do not include
  CLASSIC CEMG.` under the typical sentence. `HouseholdScopeWithNothingToSetAside`
  reads `No unusual payments to set aside.` under the view control.

**Changed by this look.**

- The detail now comes before the table, in the markup and on screen. Stacked,
  at Compact and Medium, it had sat under all thirteen rows, some 830 px below
  the row that chose it; it now stays under the chart that selects it.
- Settled: from 1200 px the table stands left at three fifths and the detail
  right at two fifths. The detail is set sticky; a still capture cannot show
  that, so it is not seen.
  Below 1200 px they stack, detail first.
- Beside the table the detail is too narrow for a full ledger row: the account
  was cut to `Current ac…` and the note to `first payment to Auto Re…` at Wide
  and Ultra. There its unusual payments take the stacked row form Compact uses.
- `HouseholdScopeWithNoHistory`: `No full month is held for these accounts yet.`
  and `No recurring payments in these accounts.` sat flush against the card's
  edge; a card holding only a sentence now gives it the card's 16 px inset.
- `TwoFullMonths` selected no month, which the load never produces; it now
  opens on August like the rest, with figures and no risers.
- Compact keeps all thirteen rows at 64 high, about 830 px, below the detail.

**Left as it is.** `MonthDetail`'s own stories are drawn with `compact` off, so
at the Compact regime their unusual payment rows clip, as `LedgerRow`'s do.
Overview passes `compact` and reads correctly. At Wide and Ultra the detail is
shorter than the table and leaves the rest of its column empty.
