## Screens

- **Overview.** `packages/ui/src/pages/Overview.svelte` and
  `packages/ui/src/pages/Overview.stories.svelte`. Both exist and are altered:
  the transfers sentence becomes a string the screen is given, a selected month
  may say how many transfers it left out, and a listed row may carry
  `Between your accounts`.
- **Transactions.** `packages/ui/src/pages/TransactionsScreen.svelte` and
  `packages/ui/src/pages/TransactionsScreen.stories.svelte`. Both exist and are
  altered: a row may carry `Between your accounts`.

Three existing molecules change with them and carry their own state stories:
`molecules/LedgerRow.svelte`, `molecules/MonthDetail.svelte` and
`molecules/BalanceChart.svelte`. Nothing is new.

## State stories

`Overview` is `packages/ui/src/pages/Overview.stories.svelte`, `Transactions`
is `packages/ui/src/pages/TransactionsScreen.stories.svelte`, and `LedgerRow`,
`MonthDetail` and `BalanceChart` are the `.stories.svelte` beside each
molecule. Scenarios are `banking/overview`'s unless a capability is named.
Today in every fixture is 20 September 2026, as it already is.

| Story | File | Fixture | Scenario |
| --- | --- | --- | --- |
| `Populated` (changed) | Overview | As it stands, under All, with the new sentence in the month summary and under the months, and `Transfer to savings` among recent transactions, labelled. | Transfers between own accounts; Transfers are left out, and the section says so; In recent transactions |
| `AMonthWithTransfersLeftOut` | Overview | `Populated`, with June 2026 selected: two transfers left out. | A month with two transfers. `banking/own-transfers`: A month with a transfer to savings |
| `AMonthWithNoneOpened` (changed) | Overview | As it stands. Asserts the absence below. | A month with none (of `A month says how many transfers it left out`) |
| `YoursScope` (changed) | Overview | As it stands, with the narrower sentence, `Joint account` at `−€800.00` among largest payments, labelled, and `Joint account` among recurring payments. | Under a narrower scope; Among largest payments; Under a scope that holds only one side. `banking/own-transfers`: Paying into the joint account, under Yours; A standing transfer that crosses the scope |
| `HouseholdScope` (changed) | Overview | As it stands, with the narrower sentence. | Under a narrower scope. `banking/own-transfers`: Paying into the joint account, under Household |
| `ADayATransferLeft` | Overview | `Populated`, with the chart's marker on 1 Sep: one mover, a transfer. | The day the money left |
| `TwoCurrencies` (changed) | Overview | As it stands. The sentence appears once per currency, as the old one did. | Transfers between own accounts |
| `ATransferLeaving` | BalanceChart | The marker on 1 Sep. | The day the money left |
| `DetailWithTransfersLeftOut`, `DetailWithOneTransferLeftOut` | MonthDetail | June 2026 with two; July 2026 with one. | A month with two transfers; One transfer |
| `BetweenYourAccounts`, `CompactBetweenYourAccounts`, `BetweenYourAccountsArriving` | LedgerRow | One row each. | `banking/transactions`: Both halves are labelled |
| `AsItOpens` (changed) | Transactions | As it stands, plus the two halves of one transfer on 1 Sep. | `banking/transactions`: Both halves are labelled; A large transfer; The count does not change |
| `OneAccount` (changed) | Transactions | As it stands, reading `Savings`, with the arriving half. | `banking/transactions`: Reading one account |

## Words

Amounts, dates and account names are fixture data, written here so a play
function asserts the sentence they sit in. No state shows the words "internal",
"excluded", "ignored", "hidden", "netted" or "matched".

**`Populated`.** Its words as they stand, except:

- In the month summary's note, in place of `Money moved between your own
  accounts is counted.`: `Money moved between your accounts is left out where
  wimm holds both. Other transfers are counted.`
- Under the months, the same sentence.
- Among recent transactions: `Transfer to savings`, `1 Sep`, `−€500.00`,
  `Between your accounts`.
- Absent anywhere on the page: `is counted.` as the end of the old sentence,
  that is, the string `Money moved between your own accounts is counted.`

**`AMonthWithTransfersLeftOut`.** `Populated`'s words with June selected:
`Money in` `€2,450.00`, `Money out` `€2,299.80`, `Net` `+€150.20`, and under
the figures `2 transfers between your accounts left out · €1,400.00`.

**`AMonthWithNoneOpened`.** Its words as they stand. Absent inside the detail:
`left out`.

**`YoursScope`.** Its words as they stand, except the sentence, in both places:
`Money moved between accounts counted here is left out. Money moved to or from
your other accounts is counted.` Among largest payments: `Joint account`,
`1 Sep`, `−€800.00`, `Between your accounts`. `Recurring payments` holds
`Spotify`, `Netflix`, `NOS` and `Joint account`, the last with
`Monthly · Current account · Monzo`, `Expected 1 Oct`, `−€800.00`. Absent in
the selected month's detail: `left out`.

**`HouseholdScope`.** Its words as they stand, with the same narrower sentence.

**`ADayATransferLeft`** (Overview) and **`ATransferLeaving`** (BalanceChart).
`1 Sep`; `€10,912.40`; `€500.00 less than 31 Aug`; `Transfer to savings`,
`−€500.00`, `Between your accounts`. The live region reads, as one string:
`1 Sep. €10,912.40. €500.00 less than 31 Aug. Transfer to savings −€500.00,
Between your accounts.` Absent: `Unusual`.

**MonthDetail.** `DetailWithTransfersLeftOut`: `June 2026`, its three figures,
and `2 transfers between your accounts left out · €1,400.00`.
`DetailWithOneTransferLeftOut`: `July 2026` and `1 transfer between your
accounts left out · €500.00`, never `1 transfers`.

**LedgerRow.** `BetweenYourAccounts` and `CompactBetweenYourAccounts`:
`Transfer to savings`, `Current account`, `1 Sep`, `−€500.00`,
`Between your accounts`. `BetweenYourAccountsArriving`: `Transfer from current
account`, `Savings`, `1 Sep`, `+€500.00`, `Between your accounts`.

**Transactions `AsItOpens`.** Its words as they stand, including `Unusual` on
the `Galp` row and `Not settled` on `Transfer to Ana Reis`. Added: `Transfer to
savings`, `−€500.00`, `Between your accounts`, and `Transfer from current
account`, `+€500.00`, `Between your accounts`. `Between your accounts` appears
exactly twice. `Transfer to Ana Reis` does not carry it: it is a payment to
somebody else, and not settled. The count of transactions in the header is the
fixture's count with the two rows in it.

**Transactions `OneAccount`.** Its words as they stand, and `Transfer from
current account`, `+€500.00`, `Between your accounts`, exactly once.

## Flow

Every state here stands alone. `AMonthWithTransfersLeftOut` and
`ADayATransferLeft` are added to `unplaced` in
`apps/storybook/canvas/flows.js`, which the flow check requires of a page story
no flow walks. `routes` still resolves `/` to `pages-overview--populated` and
`/transactions` to `pages-transactionsscreen--as-it-opens`, so those two keep
their names.

## Surfaces

- Nothing new. Overview stays a full page, route-backed, with the scope in its
  address; Transactions stays a full page. The label is text on a row, the
  left-out line is text in the month detail, and neither is operable.

## Components used

- `packages/ui/src/molecules/LedgerRow.svelte`: ledger rows, recent
  transactions, largest payments. Changes; see below.
- `packages/ui/src/molecules/MonthDetail.svelte`: the selected month. Changes.
- `packages/ui/src/molecules/BalanceChart.svelte`: the popover's movers.
  Changes.
- `packages/ui/src/molecules/MonthTable.svelte`,
  `packages/ui/src/molecules/MonthlyNetChart.svelte`,
  `packages/ui/src/atoms/SegmentedControl.svelte`: as they are.

**Changed, not missing.**

- `LedgerRow.svelte`: `transfer?: boolean` puts `Between your accounts` in the
  status tag, the slot `Unusual` and `Not settled` use. `Not settled` wins over
  it, and it wins over `unusual`; the load never sends two.
- `MonthDetail.svelte`: `leftOut?: string`, one line under the figures in
  `color-text-secondary`, already written by the load.
- `BalanceChart.svelte`: `BalanceMover` gains `transfer?: boolean`; the tag and
  the live region say `Between your accounts` where they would say `Unusual`.
- `Overview.svelte`: `CurrencySection` gains `transferNote: string`, and the
  hard-coded sentence goes. `RecentTransaction` and the largest-payment entry
  gain `transfer?`.
- `TransactionsScreen.svelte`: an entry gains `transfer?`, passed to its row.

## Components missing

None. Read before deciding: `molecules/LedgerRow.svelte` already has one status
tag that carries `Not settled`, `Unusual` and `Unusual income`, which is where
a fact about one row belongs; `molecules/MonthDetail.svelte` already has a
`note` line, used for the month so far, and a second line beside it is a prop
and not a component; `molecules/BalanceChart.svelte` already tags a mover.
`atoms/Notice.svelte` was read and is not used: it is a block message about a
screen or a section, and a fact about one row in a list of fifty is a tag. No
atom under `packages/ui/src/atoms` is a tag or a badge; the pill in
`LedgerRow` is the one that exists, and this reuses it.

## States left out

- **Each rule's cases.** Too far apart, not the same amount, two currencies,
  the same amount twice in a week, one arrival with two possible sources, the
  text naming the other account, a refund that happens to match, not yet
  settled, an account held only by a grant, a transfer across the end of the
  month. What is a pair is the service's, held by table tests. The screen
  labels what it is told to.
- **The other owner of a joint account** and **A partner's own account.** An
  absence of a label, which nothing distinguishes from an ordinary row. Held by
  a service test.
- **Both owners read the same household month.** Two sessions, held by a
  service test.
- **One that could not be paired.** An ordinary row, which every existing story
  already shows.
- **A growing balance no longer reads as overspending.** A figure, which
  `Populated` already shows above zero; whether it is right is task 8.2's walk
  with real data.
- **Paying into the joint account, under All.** An absence from two figures.
  Held by a service test.

## Contracts for implementation

- The label is text in `LedgerRow`'s existing pill tag (`color-text-secondary`
  on a `color-border-default` outline) and `BalanceChart`'s existing mover
  tag. It is never colour alone and never an
  icon alone.
- A row shows at most one of `Not settled`, `Between your accounts`, `Unusual`
  or `Unusual income`, in that order of precedence. A play function on
  `LedgerRow` asserts that a row given both `transfer` and `unusual` shows
  `Between your accounts` and not `Unusual`.
- `LedgerRow` stays 56 high at Wide and 64 at Compact with the label. Whether
  the label fits the tag's column at Compact is the look's to settle; if it
  does not, it moves to the row's second line and Seen says so.
- The balance chart's live region carries the label in the same place it
  carries `Unusual`: after the amount, after a comma, before the full stop.
- The left-out line is not a link and not a control. It is absent, not empty,
  when the load sends no `leftOut`.
- The transfers sentence is passed in already written. `Overview.svelte` holds
  no copy of it, and a play function on `YoursScope` asserts the narrower
  sentence and the absence of the All one.
- The words `Between your accounts`, `left out`, `1 transfer` and
  `N transfers` are asserted by play functions on the fixture's whole string,
  since nothing else holds them.

## Seen

Not looked at yet.
