## Screens

- **Overview.** `packages/ui/src/pages/Overview.svelte` and
  `packages/ui/src/pages/Overview.stories.svelte`. Both exist and are
  rewritten: two money tiles, a month summary, a balance chart, top spending,
  an accounts list and recent transactions replace one total, a recent list
  and a strip of bars.
- **Transactions.** `packages/ui/src/pages/TransactionsScreen.svelte` and
  `packages/ui/src/pages/TransactionsScreen.stories.svelte`. Both exist and
  are altered: the bank's line under a name that differs from it, the date in
  the day header instead of in every row, the span of dates shown once.

Two components change with them and carry their own state stories:
`packages/ui/src/molecules/LedgerRow.svelte` and
`packages/ui/src/molecules/PageScrubber.svelte`.

## State stories

`Overview` is `packages/ui/src/pages/Overview.stories.svelte`, `Transactions`
is `packages/ui/src/pages/TransactionsScreen.stories.svelte`, `LedgerRow` is
`packages/ui/src/molecules/LedgerRow.stories.svelte` and `PageScrubber` is
`packages/ui/src/molecules/PageScrubber.stories.svelte`. Scenarios are named
by requirement where two share a name.

| Story | File | Fixture | Scenario |
| --- | --- | --- | --- |
| `Populated` | Overview | One currency. Household money over a joint account, your money over two owned accounts, one account shared at *balance* read yesterday, a month summary with prior figures, daily points from 22 Jun to 20 Sep, five merchants, five largest payments, three recent transactions. | `banking/overview`: Recent activity at a glance; A payment is named, not quoted; A month under way; Spending is up; Transfers between own accounts; Several payments to one merchant; The largest single payments; Seeing where each figure sits; An account shared with them; A balance read some time ago; A chart from what has been read; The ends and the extremes are labelled; A member who cannot see the chart; The total is there, the connection controls are not; Ours and mine, side by side; The screen uses the space it has. `banking/household-accounts`: The figures agree with the accounts. |
| `ADayPointedAt` | Overview | `Populated`, with the chart's marker on 3 Aug. | Reading a day off the chart |
| `ChartDoesNotCoverEveryAccount` | Overview | `Populated`, with household money at €9,244.55 and the chart's partial coverage set. | One account connected yesterday; An account with balances only |
| `LedgerStartsThisMonth` | Overview | `Populated`, with no prior figures, `countedFrom` 19 Sep, the chart's short history set and no points. | Last month is not all there; The ledger starts inside this month; Less than a week of history anywhere |
| `NoTransactionHistory` | Overview | The two money tiles and the accounts list, and no summary, chart, top spending or recent transactions. | No transaction history anywhere; An owned account with no transactions yet |
| `NothingSpentYet` | Overview | `Populated`, with money in only: no merchants, no largest payments, money out at €0.00. | Nothing spent yet |
| `ThreeMerchants` | Overview | `Populated`, with three merchants and three largest payments. | Fewer than five |
| `OwnsNoAccount` | Overview | Household money at €4,120.00 over one joint account the member holds *details* on, one account shared at *balance*, and nothing else. | A member who owns nothing sees no section; A member who owns nothing; An account granted at balance. `banking/household-accounts`: Shared at balance only |
| `HouseholdOfOne` | Overview | `Populated`, with no household money and no Household group. | `banking/household-accounts`: A household of one; Nothing in a group |
| `TwoCurrencies` | Overview | EUR and GBP, each with money tiles, a month summary, a chart and top spending. EUR holds more. | Two currencies; Two currencies, two charts. `banking/household-accounts`: Accounts in more than one currency |
| `CurrencyHoldingNothing` | Overview | `Populated`, plus one owned USD account at $0.00 in the accounts list, and no USD tile, summary or chart. | A currency holding nothing |
| `ReachesTheFullLedger` | Overview | `Populated`. | Reaching the full ledger; Looking at the month in full; Going to an account; Getting to account management from Overview |
| `BeforeAnyBankConnected` | Overview | `hasAccounts` false and nothing else. | The first member arrives |
| `MemberMayNothing` | Overview | `hasAccounts` false and nothing else. | A member who may see no account |
| `AsItOpens` | Transactions | The newest page: two days, three rows whose name differs from the bank's line, one salary and one transfer whose name is the bank's. | `banking/transactions`: Checking a name against the statement; The bank names the other party; The same name everywhere |
| `AnOlderPage` | Transactions | A middle page of the scrubber. | None in this change. Holds the footer contract below. |
| `TheOldestPage` | Transactions | The oldest page of the scrubber. | None in this change. Holds the footer contract below. |
| `Compact` | Transactions | `AsItOpens` at 390 wide. | `banking/transactions`: Checking a name against the statement |
| `BanksLineShown` | LedgerRow | A name and a bank's line that differ. | `banking/transactions`: Checking a name against the statement |
| `BanksLineSameAsName` | LedgerRow | A name equal to the bank's line. | `banking/transactions`: The bank names the other party |
| `CompactBanksLineShown` | LedgerRow | `BanksLineShown`, compact. | `banking/transactions`: Checking a name against the statement |
| `No label at rest` | PageScrubber | A middle page, nothing hovered or focused. | None in this change. Holds the scrubber contract below. |
| `Label under the pointer` | PageScrubber | A middle page, one other dot focused. | None in this change. Holds the scrubber contract below. |

`TrendDoesNotCoverEveryAccount`, `NoTrend` and `OwnsNoAccountNoRecentSection`
in `Overview.stories.svelte` are replaced by
`ChartDoesNotCoverEveryAccount`, `NoTransactionHistory` and `OwnsNoAccount`.

## Words

Amounts, dates, merchants, account names and bank names are fixture data. They
are written here so a play function asserts the sentence they sit in. No state
shows the words "budget", "income" or "expenses".

**`Populated`.**

- Page title: `Overview`.
- Money tiles: `Household money` over `€6,698.00`, `Your money` over
  `€4,995.55`.
- Month tiles: `Money in` over `€2,450.00` with `€150.00 more` and
  `than 1 to 20 Aug`; `Money out` over `€1,812.64` with `€182.40 more` and
  `than 1 to 20 Aug`; `Net` over `+€637.36` with `€32.40 less` and
  `than 1 to 20 Aug`.
- Summary note: `1 to 20 Sep, against 1 to 20 Aug. Money moved between your
  own accounts is counted.`
- Chart card: title `Balance · EUR`, span `22 Jun to 20 Sep`, scale `€12,118`
  at the top and `€9,204` at the bottom, dates `22 Jun` at the left and
  `20 Sep` at the right.
- Chart's accessible name: `Balance from 22 Jun to 20 Sep. It starts at
  €9,640.00 and ends at €11,693.55. Highest €12,118.00 on 2 Sep. Lowest
  €9,204.00 on 14 Jul.`
- Top merchants card: title `Top merchants`, span `1 to 20 Sep`, then
  `Pingo Doce` with `€412.60 · 9 payments`, `Galp` with
  `€186.40 · 3 payments`, `Amazon` with `€142.18 · 4 payments`, `Via Verde`
  with `€64.30 · 12 payments`, `Netflix` with `€12.99 · 1 payment`. One
  payment reads `1 payment`, never `1 payments`.
- Accounts card: title `Accounts`, button `Manage accounts`. Group `Household`:
  `Joint account`, `Monzo`, `€6,698.00`, `Read just now`. Group `Yours`:
  `Current account`, `Monzo`, `€2,480.55`, `Read just now`; `Conta à Ordem`,
  `Montepio`, `€2,515.00`, `Read yesterday at 18:04`. Group `Shared with you`:
  `Savings`, `Montepio`, `€3,200.00`, `Read just now`.
- Largest payments card: title `Largest payments`, then `Rent`,
  `Joint account · Monzo`, `6 Sep`, `−€820.00`; `Galp`,
  `Current account · Monzo`, `12 Sep`, `−€92.10`; `Pingo Doce`, `14 Sep`,
  `−€88.34`; `Amazon`, `9 Sep`, `−€64.99`; `EDP`, `2 Sep`, `−€58.20`.
- Recent transactions card: title `Recent transactions`, button `See all`,
  then `Pingo Doce`, `17 Sep`, `−€42.18`; `Salary`, `15 Sep`, `+€2,450.00`;
  `Netflix`, `14 Sep`, `−€12.99`.
- Absent: any control named Connect, Restore, Disconnect or Leave out.

**`ADayPointedAt`.** `Populated`'s words, and the readout `3 Aug · €9,870.12`
under the dates.

**`ChartDoesNotCoverEveryAccount`.** `Populated`'s words with
`Household money` over `€9,244.55`, and under the dates: `Not every account is
in this chart. An account is left out when its bank shares balances only, or
when its history is under 30 days.` One sentence covers both reasons, because
the member's question is the same: why the chart and the figures disagree.

**`LedgerStartsThisMonth`.** `Populated`'s words, except: the three month
tiles show `€2,450.00`, `€1,812.64` and `+€637.36` with no `more`, `less` or
`than` under them; the summary note reads `Counted from 19 Sep. Money moved
between your own accounts is counted.`; the chart card holds its title
`Balance · EUR` and `A balance chart appears once there is a week of
history.` and nothing else; the Top merchants span reads `19 to 20 Sep`.

**`NoTransactionHistory`.** `Overview`, `Household money`, `Your money`,
`Accounts`, `Manage accounts` and the three groups as in `Populated`. Absent:
`Money in`, `Balance · EUR`, `Top merchants`, `Largest payments`,
`Recent transactions`, and any sentence about a chart.

**`NothingSpentYet`.** `Populated`'s words with `Money out` over `€0.00`.
Absent: `Top merchants`, `Largest payments`.

**`ThreeMerchants`.** `Populated`'s words with `Pingo Doce`, `Galp` and
`Amazon` only under `Top merchants`, and three rows under
`Largest payments`. Nothing stands in for the missing two.

**`OwnsNoAccount`.** `Overview`; `Household money` over `€4,120.00`;
`Accounts`; group `Household`: `Joint account`, `Monzo`, `€4,120.00`,
`Read just now`; group `Shared with you`: `Savings`, `Montepio`, `€3,200.00`,
`Read just now`, with no account number. Absent: `Your money`, `Yours`,
`Manage accounts`, `Money in`, `Balance · EUR`, `Top merchants`,
`Largest payments`, `Recent transactions`.

**`HouseholdOfOne`.** `Populated`'s words without `Household money` and
without the `Household` group. `Your money` reads `€11,693.55`, its tile
takes the row, and all three owned accounts sit under `Yours`.

**`TwoCurrencies`.** Every title that repeats per currency carries its code:
`Household money · EUR`, `Your money · EUR`, `Money in · EUR`,
`Money out · EUR`, `Net · EUR`, `Balance · EUR`, `Top merchants · EUR`,
`Largest payments · EUR`, and the same eight with `GBP`. Under the money
tiles: `Shown per currency. wimm does not convert between them.` `Accounts`
and `Recent transactions` appear once.

**`CurrencyHoldingNothing`.** `Populated`'s words, and under `Yours` a fourth
account: `Dollar account`, `Monzo`, `$0.00`, `Read just now`. Absent:
`Balance · USD` and any tile holding `$0.00`.

**`ReachesTheFullLedger`.** `Populated`'s words. The play function follows
`See all`, `Manage accounts`, the `Joint account` row and the `Money out`
tile.

**`BeforeAnyBankConnected` and `MemberMayNothing`.** Title `Connect a bank to
see where you stand`, body `Once a bank is connected, your total, your recent
activity and how it is trending show up here.`, button `Go to Accounts`.
Absent: any amount. These words are the ones the two stories hold today.

**Transactions `AsItOpens`.**

- `Transactions`, `Every account you own, newest first.`, `Refresh`,
  `382 transactions`, `Updated at 09:14. Reaching back to 4 June 2026.`
- Column headings `Description`, `Account`, `Amount`. No `Date` heading.
- Day headers `Today, 17 September` and `Yesterday, 16 September`. An older
  day reads `4 August 2026`.
- `Pingo Doce` over `COMPRA PINGO DOCE LISBOA 230002268342127`,
  `Current account · Monzo`, `−€42.18`.
- `Transfer to Ana Reis`, `Joint savings · Montepio`, `Not settled`,
  `−€60.00`, with no line under the name.
- `Galp` over `COMPRA GALP A5 OEIRAS 230002270158934`, `−€71.40`.
- `Salary`, `+€2,180.00`, with no line under the name.
- `NOS` over `DD NOS COMUNICACOES SA 000000234058260`, `−€39.99`.
- Footer: `17 September to 15 September 2026`, once.

**Transactions `AnOlderPage` and `TheOldestPage`.** The span of the page on
screen, once, and the scrubber's ends by their accessible names, `Oldest` and
`Newest`. On `TheOldestPage`, `Oldest` is disabled.

**Transactions `Compact`.** `AsItOpens`'s words without the column headings.

**LedgerRow `BanksLineShown` and `CompactBanksLineShown`.** `Pingo Doce` over
`COMPRA PINGO DOCE LISBOA 230002268342127`.

**LedgerRow `BanksLineSameAsName`.** `Salary`, once.

**PageScrubber `No label at rest`.** No span text is visible.
**`Label under the pointer`.** The focused page's span,
`12 August to 3 August 2026`, and no other.

## Flow

`connect-a-bank` in `apps/storybook/canvas/flows.js` opens on
`pages-overview--before-any-bank-connected` and leaves it on
`ongotoaccounts`. That state keeps its name, its button and its callback, so
the flow is unchanged. `routes` in the same file resolves `/` to
`pages-overview--populated` and `/transactions` to
`pages-transactionsscreen--as-it-opens`, which is why those two stories keep
their names. Every other state here stands alone.

## Surfaces

Overview stays a full page, route-backed (`/`). Nothing on it opens a modal
or a drawer. Every link leaves the page: **Manage accounts** and each account
row go to `/accounts`, **See all** to `/transactions`, a month summary tile
to Transactions at this month.

Transactions stays a full page, route-backed (`/transactions`), and each dot
of the page scrubber is a link to a real page.

## Components used

- `packages/ui/src/templates/SignedInLanding.svelte` and
  `packages/ui/src/templates/Page.svelte`: the shell and the page header,
  unchanged.
- `packages/ui/src/molecules/MetricTile.svelte`: household money, your money
  and the three month figures. The three month tiles use its Delta row
  (arrow, change, period); the two money tiles pass none. Arrow and change are
  `color-text-secondary`: green and red mean money direction only (ADR 0002),
  and "€182.40 more" is not a direction of money.
- `packages/ui/src/molecules/AccountRow.svelte`: one per account, Balance over
  Reading, the same row Accounts shows, so the two screens cannot drift.
- `packages/ui/src/molecules/LedgerRow.svelte`: Largest payments and Recent
  transactions, with no settled marker, the description filling so amounts
  align right in a half-width card. `compact` at Compact.
- `packages/ui/src/atoms/Button.svelte`: **Manage accounts** and **See all**,
  secondary.
- `packages/ui/src/molecules/EmptyState.svelte` and
  `packages/ui/src/atoms/Icon.svelte`: before any bank is connected.
- On Transactions: `packages/ui/src/molecules/PageScrubber.svelte` for the
  track of pages and `packages/ui/src/molecules/SeekPager.svelte` for the
  span. The scrubber changes: it shows a label only for the page under the
  pointer or focus, because the span beside it already names the current
  page.

`packages/ui/src/molecules/TrendSparkline.svelte` is not used by Overview.

## Components missing

**Balance chart** (molecule). `packages/ui/src/molecules/BalanceChart.svelte`
and `BalanceChart.stories.svelte`, with stories for as it opens, a day pointed
at, the coverage sentence and short history. A chart header, a plot row
holding a 72 px scale and an SVG plot with an area and a line, a dates row,
and an optional readout and coverage sentence. Read before deciding:
`molecules/TrendSparkline.svelte` holds the same data as bars from a zero
baseline at a fixed 48 px and has no scale, no dates and no way to mark a day.
`molecules/MetricTile.svelte` holds one figure and a delta. The only other SVG
under `packages/ui/src` is icons and marks (`atoms/Icon.svelte`,
`atoms/Brand.svelte`, `atoms/Notice.svelte`, `molecules/BankRow.svelte`).
Nothing draws a quantity over time.

**Budget meter** (molecule). `packages/ui/src/molecules/BudgetMeter.svelte`
and `BudgetMeter.stories.svelte`: a label, a value, and a track filled to a
proportion, one per top merchant. Read before deciding:
`molecules/TrendSparkline.svelte` draws proportional bars but vertical,
unlabelled and as one series. `atoms/NavigationProgress.svelte` is a bar
fixed to the top of the viewport that reports a navigation and carries no
label, value or proportion. `molecules/MetricTile.svelte` holds a figure with
no proportion. The word "budget" never reaches interface copy.

**A read-only account row.** A configuration of
`packages/ui/src/molecules/AccountRow.svelte`: Balance over Reading, the whole
row a link, no connection controls. `AccountRow.svelte` was read: it renders
the balance and the reading and takes no `href`. Its story goes in
`AccountRow.stories.svelte`.

## States left out

- **A currency at zero that has been used.** The service decides a currency is
  in use; the screen renders it like any other, which is `TwoCurrencies`.
- **The 31st against a shorter month.** The same summary note with other
  dates. The window is the service's, and its test holds it.
- **Pending payments are not counted, Money coming in is not spending, The
  ranking is by money.** What is in a figure or a list is the service's. The
  screen renders what it is given.
- **An account left out.** An absence from a list. Nothing distinguishes it
  from `Populated`.
- **Spending is down.** Both directions are in `Populated`: Money in and
  Money out show "more", Net shows "less".
- **The scenarios of "Household money and a member's own money are separate
  figures" about who counts an account where.** The grouping is the
  service's. `Populated`, `OwnsNoAccount` and `HouseholdOfOne` are the three
  shapes the screen can take.
- **The naming scenarios of `banking/transactions`.** A name is derived by the
  service. The screen shows the name and, where it differs, the bank's line.
- **Going straight to a month.** `banking/transactions` describes it, the page
  scrubber moves by page, and this change alters neither.

## Contracts for implementation

- Body padding `space-8`, gap 24; Compact `space-4`, gap 16.
- Order, every regime: page header, the two money tiles, the three month
  tiles, summary note, Balance chart, Top merchants and Accounts, Largest
  payments and Recent transactions. Compact stacks them in that order with
  **Accounts before Top merchants**, since the balance is the more common
  reason to open the app on a phone. Medium keeps both tile rows and stacks
  Largest payments above Recent transactions at full width; Wide and Ultra put
  each pair side by side with a 24 px gap. Two currencies: the tile row, note,
  chart and top spending repeat per currency, largest total first.
- No section is capped by `layout-content-max`. At Medium, Wide and Ultra the
  chart card's width equals the Main column's content width.
- Money row: two tiles filling the width equally, 82 high. Month row: three
  tiles filling equally, 102 high with the Delta row, 82 without a
  comparison. A money tile with no account in it is absent and the other
  takes the row.
- A month tile is a link to Transactions at this month. Direction is carried
  by the words "more" and "less", never by colour alone.
- Accounts card: a 28 px group heading, 11 px at weight 600 in
  `color-text-secondary`, above each group: `Household`, `Yours`,
  `Shared with you`, in that order, each absent when empty. No rule above the
  first row of a group.
- Accounts: rows 56 high, inset 6, a 1 px `color-border-subtle` rule between
  rows and none above the first. The whole row is the link, and the card
  holds no control that changes anything.
- **Manage accounts** calls `ongotoaccounts` and **See all** calls `onseeall`.
- Balance chart card: padding 16, gap 12, 293 high. Plot row 200 high: scale
  72 wide, gap 8, plot fills the rest (1000 at Wide, 776 Medium, 1456 Ultra,
  246 Compact). Dates row is inset 80 from the left so the start date sits
  under the plot, not under the scale. Line is 2 px `color-chart-1`; area is
  `color-chart-1` at 12% opacity; the plot's bottom edge is a 1 px
  `color-border-subtle` rule.
- **The line is data.** The box, the stroke, the fill, the labels and their
  positions are fixed. The y scale is `[low, high]` padded 8% each side and
  never forced to zero. High and low are the top and bottom of the scale
  column, right-aligned, in the mono family.
- Marker: 1 px, `color-border-strong`, full plot height. Readout is mono,
  `date · amount`, and is an `aria-live="polite"` region. The chart is one
  tab stop; left and right arrows move a day, Home and End go to the ends.
  Its accessible name is the sentence the spec requires: span, balance at
  each end, high and low with their dates. The caller passes it, already
  formatted.
- Top merchants: card header 48, meters inset 16 with a 14 gap, each meter 28
  high with a 7 px track filled in `color-chart-1`. Proportion is against the
  largest merchant, which is always full.
- Ledger rows 56 high; Compact rows 64.
- Transactions rows stay 56 high at Medium, Wide and Ultra with the bank's
  line in them: name 13 px, line 11 px in `color-text-placeholder`, 2 px
  apart. Compact rows with a bank's line are 80 high, the rest 64, in the
  order name and amount, account, bank's line. The bank's line renders only
  when it differs from the name, exactly as the bank wrote it.
- Transactions columns are Description, Account, a 92 px status slot, Amount.
  A row carries no date at Medium and wider; its day header does.
- Transactions footer is 56 high: the span fills the left, the scrubber sits
  right. Compact stacks them, 92 high, span above scrubber. The span appears
  once in the footer.
- Page scrubber: 6 px dots joined by 10 px rules, the current page at 10 px
  in `color-accent`, the two ends 28 × 28 with a 14 px chevron. An end it
  cannot go past is present and disabled. It shows a label for the page under
  the pointer or focus only, never for the current page at rest.
- Money arriving carries `+` wherever an amount is shown: `+€2,180.00`, never
  `€2,180.00`.
- The words around a figure, "more", "less", "than", "against" and
  "Counted from", are asserted by play functions on the fixture's whole
  string, since nothing else holds them.

## Seen

Looked at by Claude on 2026-09-20, at the household member's request, in light
theme and comfortable density. Every story below was rendered from Storybook at
390, 834, 1440 and 1920 wide with the frame grown to the page's full height, and
seen in a screenshot. Heights and widths are measured from the rendered page.
At every story and regime: no horizontal overflow, nothing past the right edge.
Dark theme and compact density were not looked at.

Measured on every Overview story that has the part: money tiles 82 high; month
tiles 102 with a comparison and 82 without (103 at Compact with a comparison);
account rows 56; meters 28; ledger rows 56, and 64 at Compact; the chart card
294 high (the contract says 293), 322 with a day's readout, 338 to 354 with the
coverage sentence; the chart card as wide as the Main column's content at
Medium, Wide and Ultra (714, 1112, 1536); the plot 244, 600, 998 and 1422 wide
(the contract says 246, 776, 1000 and 1456; Medium's 776 does not fit a 714
column).

- `Populated`. Compact: header, two money tiles side by side, the three month
  tiles stacked one per row, note, chart, **Accounts before Top merchants**,
  Largest payments, Recent transactions, each ledger row two lines. Medium:
  both tile rows kept, then chart, Top merchants, Accounts, Largest payments,
  Recent transactions, each at full width. Wide: Top merchants beside Accounts,
  Largest payments beside Recent transactions. Ultra: as Wide, filling 1536.
  Groups read Household, Yours, Shared with you; `+€637.36` and `+€2,450.00`
  carry their sign; "more" and "less" sit beside the arrows.
- `ADayPointedAt`. All four: a marker over 3 Aug, and `3 Aug · €9,870.12`
  under the dates. Otherwise as `Populated`.
- `ChartDoesNotCoverEveryAccount`. All four: household money €9,244.55 and the
  coverage sentence under the chart, on one line at Wide and Ultra, two at
  Medium, three at Compact.
- `LedgerStartsThisMonth`. All four: month tiles with no comparison, "Counted
  from 19 Sep." in the note, the chart card holding only its title and "A
  balance chart appears once there is a week of history.", Top merchants headed
  19 to 20 Sep.
- `NoTransactionHistory`. All four: the two money tiles and the Accounts card
  and nothing else. At Wide and Ultra the Accounts card was half the row beside
  an empty half; fixed, see Changed.
- `NothingSpentYet`. All four: Money out €0.00 with no comparison, no Top
  merchants, no Largest payments. At Wide and Ultra, Accounts and Recent
  transactions each take the full row after the fix below.
- `ThreeMerchants`. All four: three meters, three largest payments.
- `OwnsNoAccount`. All four: one Household money tile taking the row, an
  Accounts card with Household and Shared with you, no Manage accounts, nothing
  else.
- `HouseholdOfOne`. All four: one Your money tile taking the row, no Household
  group; the rest as `Populated`. The fixture still shows a Shared with you
  group, which a household of one cannot have.
- `TwoCurrencies`. All four: the tile rows, note and chart once for EUR and once
  for GBP, EUR first, with "wimm does not convert between them."; Top merchants
  and Largest payments once per currency; one Accounts card. GBP month tiles
  carry no comparison. Wide and Ultra were seen at about half size, so their
  words were not read, only their layout.
- `CurrencyHoldingNothing`. All four: a Dollar account at $0.00 under Yours, and
  no USD tile, note or chart.
- `ReachesTheFullLedger`. All four: as `Populated`; it differs only in what its
  play function clicks.
- `BeforeAnyBankConnected`, `MemberMayNothing`. All four: one card, "Connect a
  bank to see where you stand", its sentence and Go to Accounts, centred in the
  Main column. The two look identical.
- Transactions `AsItOpens`. Compact: rows with a bank's line 80 high, the rest
  64, footer 92. Medium, Wide, Ultra: columns Description, Account, a status
  slot, Amount, no Date; day headers "Today, 17 September" and "Yesterday, 16
  September"; rows 56; a bank's line under Pingo Doce, Galp and NOS and none
  under Salary or the transfer; footer 56 with the span once. At Medium the
  name column is about 150 wide, so a bank's line ends in an ellipsis after
  "COMPRA PINGO DOCE LISBOA 2"; at Wide and Ultra it is whole. At Medium the
  rows' accounts did not line up; fixed, see Changed.
- Transactions `AnOlderPage`. All four: two day headers, the span left and the
  scrubber right with the current page in the middle; Compact stacks span above
  scrubber, 92 high.
- Transactions `TheOldestPage`. All four: "4 June 2026", "Nothing older. This is
  as far back as the bank would go.", the scrubber's older end disabled. At
  Compact the date wrapped to two lines; fixed, see Changed.
- Transactions `Compact`. The story holds the compact form at every width:
  stacked rows 80 and 64, footer 92. At Medium, Wide and Ultra that form is
  stretched across the column, which no member sees.
- LedgerRow `BanksLineShown`, `BanksLineSameAsName`. Medium, Wide, Ultra: 56
  high, the line whole under the name in the first and absent in the second.
  Compact: this is the wide form of the row at 390, which the app never shows;
  the name is cut to "Pi…" there. Both were centred and shrunk to fit, which
  cut "Salary" to "S…" at every regime; fixed, see Changed.
- LedgerRow `CompactBanksLineShown`. All four: 80 high, name and amount, then
  account and date, then the bank's line whole.
- PageScrubber `No label at rest`. All four: two 28 × 28 ends, dots joined by
  rules, the current page larger and in the accent colour, no label.
- PageScrubber `Label under the pointer`. All four: one label above the focused
  dot and none for the current page.

Left as seen, for a person to decide: at Wide and Ultra the shorter card of a
pair leaves a gap under it. At Medium a bank's line of about 40 characters still
loses its last few digits: the name column is 232 wide and the line needs 236
to 259; the whole line is in its tooltip and shows at Wide and Ultra.

The plot widths in Contracts left out the card's 1 px border and misstated two
columns. What the screen draws is the column less the border, the padding, the
scale and the gap: 244, 600, 998 and 1422.

### Changed

- `packages/ui/src/molecules/LedgerRow.svelte`: the account column's width is
  `--ledger-account-basis` (220 px by default) and the name column's basis is
  `--ledger-name-basis` (its words by default).
- `packages/ui/src/pages/Overview.svelte`: a side-by-side pair sets
  `--ledger-account-basis: auto`. At Wide, Largest payments and Recent
  transactions cut names to "Pingo D…", "Sala…" and "Net…" beside an account
  column that was mostly empty; the names are now whole at every regime.
- `packages/ui/src/pages/Overview.svelte`: a card alone in a pair takes the
  full row, as a lone money tile does.
- `packages/ui/src/pages/TransactionsScreen.svelte`: the table sets
  `--ledger-name-basis: 0px`, so a long bank's line no longer pushes the
  account out of its column; at Medium the accounts now line up.
- `packages/ui/src/molecules/AccountRow.svelte`: the figures column has a floor
  of 108 px instead of a fixed width, and the reading does not wrap. "Read
  yesterday at 18:04" wrapped to two lines at every regime, Ultra included.
- `packages/ui/src/molecules/SeekPager.svelte`: the span stays on one line, so
  at Compact the sentence beside it wraps instead of "4 June / 2026".
- `packages/ui/src/molecules/LedgerRow.stories.svelte`: the stories use the
  padded layout, so a row fills its width as it does in a list.

`just check packages/ui`, `just check apps/storybook` (331 tests) and
`just check apps/web` pass after these changes.
- `packages/ui/src/molecules/BalanceChart.svelte`: the dates row has a fixed
  14 px line, so the card is 293 high as Contracts says and not 294.
- `packages/ui/src/pages/TransactionsScreen.svelte`: at Medium the account
  column is 168 wide and not 220, in the header and the rows, which takes the
  name column from about 180 to 232.
- `packages/ui/src/molecules/LedgerRow.svelte`: a bank's line carries its whole
  text as its title.
- `packages/ui/src/pages/Overview.stories.svelte`: `HouseholdOfOne` has no
  shared account and asserts there is no Shared with you group.
- `Compact` and `CompactOneAccount` of Transactions, and `Compact` of the consent
  explainer and of widen consent, carry `size-compact`, and the canvas draws a
  story so tagged at that size only.
