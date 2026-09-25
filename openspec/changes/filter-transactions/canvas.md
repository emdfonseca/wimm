Transactions gains a filter bar above the ledger: an account, a search, a month and a direction, with a way to clear them, and one new empty state for a filter that matches nothing. The inline `.chip` that named a narrowed account goes; the account select names it instead.

## Screens

- `packages/ui/src/pages/TransactionsScreen.svelte` with `TransactionsScreen.stories.svelte`. Exists. Changes: the filter bar, the `no-match` empty reason, the filtered oldest-page sentence, the lede, the `.filter` row and the `.chip` removed, and the freshness line moved beside Refresh. Props `filterAccount` and `showAllHref` give way to `filters`, `accounts`, `months`, `clearHref` and `onfilter` (design.md decision 8).

## State stories

Fixtures are the existing ones in `TransactionsScreen.stories.svelte` (`days`, `olderDays`, `pages`, `freshness`, `span`) plus one `accounts` list (Current account · Monzo, Savings · Monzo, Casa CC · Montepio) and one `months` list (September 2026 back to September 2025).

| Story | File | Fixture | Scenario |
| --- | --- | --- | --- |
| `AsItOpens` (changed) | `packages/ui/src/pages/TransactionsScreen.stories.svelte` | existing, plus `accounts` and `months`, no filter in force | The accounts offered |
| `OneAccount` (changed) | same | `filters.account` = Savings, its one day | Narrowing to one account |
| `Searched` (new) | same | `filters.q` = `galp`, `count` 12, two days of Galp rows | Finding a merchant; The count follows the filters |
| `OneMonth` (new) | same | `filters.month` = `2026-08`, `count` 61, August days, `span` 31 August to 18 August 2026, `pages` of August | Jumping to a month; Arriving from Overview |
| `TheOldestPageOfAMonth` (new) | same | `filters.month` = `2026-08`, `atOldest`, August's oldest days | Paging on from a month; The oldest page of a filtered list |
| `MoneyIn` (new) | same | `filters.direction` = `in`, `count` 38, only plus rows, one of them the savings half of a transfer | Only money in; One half of a transfer |
| `FiltersCombined` (new) | same | account Current, `q` = `galp`, month `2026-08`, direction `out`, `count` 3 | Two filters at once; Clearing every filter |
| `NothingMatches` (new) | same | `filters.q` = `plumber`, `days` empty, `count` 0, `freshness` set, one `unreachable` problem | A search that finds nothing; Not mistaken for an empty ledger |
| `AnAccountWithNothing` (new) | same | `filters.account` = Casa CC, `days` empty, `count` 0 | An account with nothing in it |
| `CompactFiltered` (new) | same | `compact`, `q` = `galp`, month `2026-08` | A filtered list says how current it is, at 390 wide |
| `CompactOneAccount` (changed) | same | `compact`, `filters.account` = Savings | Narrowing to one account, on a phone |
| `TheOldestPage` (unchanged words) | same | existing, no filter | The oldest page, unfiltered |
| `TotalsInTwoCurrencies` (new) | same | `filters.month` = `2026-08`, `totals` in euros and dollars | Two currencies |
| `EverythingFitsOnOnePage` (changed) | same | existing, `months` of one | Not enough ledger to need it |
| `NoBankConnected`, `OwnsNoAccount` (changed) | same | existing | No filter bar where there is nothing to filter |

## Words

The header: the title, then `Updated at 09:14. Reaching back to 4 June 2026.` beside Refresh at medium and up, and as a line under the header at compact. There is no lede.

The filter bar, in every state that shows it, in one row at medium and up. Labels are accessible names, not shown: the placeholder, the chosen option and the glyph say what each control is.

- Group name (accessible): `Filter transactions`
- Search field first, label `Search`, placeholder `Who it was with, or the bank's line`, a search glyph; its clear button's accessible name `Clear search`
- Account select, label `Account`: `All accounts`, then each account as `{name} · {bank}`, for example `Savings · Monzo`
- Month select, label `Month`: `All months`, then each month as `August 2026`, newest first
- Direction, label `Direction`: `All`, `Money in`, `Money out`
- `Clear filters`, only while a filter is in force
- At compact: the search, and a `Filters` button whose name adds `{n} chosen` for the account, month and direction chosen, opening a sheet titled `Filters` that holds the three, labelled, with `Clear filters` and `Done`

The totals, on the count's line while a filter is in force: `In +€2,214.99` and `Out −€1,204.55` per currency, only the side a direction asks for, and an info button named `What these figures leave out` that shows the note under the line.

Per story:

- `AsItOpens`: `384 transactions`; no money in or money out figures; `All accounts`, `All months` and `All` chosen; no `Clear filters`.
- `OneAccount`: `204 transactions`; `Savings · Monzo` chosen; `Clear filters`. No `Showing one account`.
- `Searched`: `12 transactions`; the search field holds `galp`; every row named `Galp`; `In +€0.00` and `Out −€612.40`, and no info button.
- `OneMonth`: `61 transactions`; `August 2026` chosen; span `31 August to 18 August 2026`; `In +€2,214.99` and `Out −€1,204.55` on the count's line; the info button shows `Leaves out 1 transfer between your accounts. Does not count 1 payment not yet settled.`
- `TheOldestPageOfAMonth`: `Nothing older matches these filters.` and not `Nothing older. This is as far back as the bank would go.`
- `MoneyIn`: `38 transactions`; `Money in` chosen; every amount starts with `+`; `Transfer from current account` reads `Between your accounts`; `In +€9,034.99` and no `Out`; the info button shows `Leaves out 2 transfers between your accounts.`
- `FiltersCombined`: `Out −€174.60` and no `In`; `3 transactions`; `Current account · Monzo`, `galp`, `August 2026` and `Money out` all chosen; `Clear filters`.
- `NothingMatches`: title `Nothing matches these filters`; body `Change a filter above, or clear them to see every transaction.`; action `Clear filters`; the search field still holds `plumber`; `Refresh` shown; the `did not answer` notice still shown. No `Connect a bank`, no `Nothing read yet`.
- `AnAccountWithNothing`: title `Casa CC · Montepio has no transactions`; body `The bank has sent none for this account.`; action `Show all accounts`.
- `CompactFiltered`: the freshness line, then the search holding `galp` and `Filters 1 chosen` on one row; no select outside the sheet.
- `CompactOneAccount`: `Filters 1 chosen`; the sheet shows `Savings · Monzo` chosen and `Clear filters`; `Done` closes it; no `Showing one account`.
- `TheOldestPage`: unchanged, `Nothing older. This is as far back as the bank would go.`
- `TotalsInTwoCurrencies`: `In +€2,214.99`, `Out −€1,204.55` and `In +US$0.00`, `Out −US$42.00`, each pair together.
- `EverythingFitsOnOnePage`: no `Month` select.
- `NoBankConnected`, `OwnsNoAccount`: no `Filter transactions` group.

## Flow

Joins `read-transactions` in `apps/storybook/canvas/flows.js`. Steps are unchanged. New branches:

- from `pages-transactionsscreen--as-it-opens`: `searched for a payment` to `pages-transactionsscreen--searched`; `filtered to a month` to `pages-transactionsscreen--one-month`; `money in only` to `pages-transactionsscreen--money-in`; `every filter at once` to `pages-transactionsscreen--filters-combined`; `a search that finds nothing` to `pages-transactionsscreen--nothing-matches`; `an account with nothing in it` to `pages-transactionsscreen--an-account-with-nothing`; `on a phone, filtered` to `pages-transactionsscreen--compact-filtered`.
- from `pages-transactionsscreen--as-it-opens`: `a month read to its end` to `pages-transactionsscreen--the-oldest-page-of-a-month`. A branch leaves from a step, and `one-month` is not one.
- from `pages-overview--populated`: `a month followed from Overview` to `pages-transactionsscreen--one-month`.
- from `pages-transactionsscreen--as-it-opens`: `a month spent in two currencies` to `pages-transactionsscreen--totals-in-two-currencies`.

## Surfaces

- The filter bar: inline, above the ledger card, route-backed. Every filter is in the address.
- Nothing matches: inline, in place of the ledger card, the filter bar staying above it.

## Components used

- `packages/ui/src/templates/Page.svelte`
- `packages/ui/src/atoms/Button.svelte`
- `packages/ui/src/atoms/SegmentedControl.svelte` (direction, inside `LedgerFilters`)
- `packages/ui/src/molecules/EmptyState.svelte`
- `packages/ui/src/molecules/ErrorNotice.svelte`, `InfoNotice.svelte`
- `packages/ui/src/molecules/LedgerRow.svelte`
- `packages/ui/src/molecules/SeekPager.svelte`, `PageScrubber.svelte`

## Components missing

- **`SearchField`**, atom, `packages/ui/src/atoms/SearchField.svelte` and its story. A labelled `type="search"` input with a clear button, `maxlength` 100, that reports a 300 ms pause in typing, Enter and clearing as one `onsearch(text, { live })` with the text trimmed, and only when the trimmed text changes. Read: the search in `pages/ChooseBankScreen.svelte` is inline markup owned by that page, filters as the member types and has no clear button, so there is nothing to instance; `molecules/AccountName.svelte` is a rename field bound to a save button, with the bank's name beneath it. ChooseBankScreen keeps its own markup in this change.
- **`SelectField`**, atom, `packages/ui/src/atoms/SelectField.svelte` and its story. A labelled native `<select>` in the control chrome, reporting `onchange(value)`. Read: `atoms/SegmentedControl.svelte` is one tab stop over a fixed handful of options split evenly across its width, and cannot hold fourteen months or a household's accounts; `molecules/PageScrubber.svelte` chooses a page to move to and says in its own comment it is not a filter; `molecules/DensityControl.svelte` and `ThemeToggle.svelte` are fixed two-way preferences. Nothing in `packages/ui/src` renders a `<select>`.
- **`LedgerFilters`**, molecule, `packages/ui/src/molecules/LedgerFilters.svelte` and its story. A `role="search"` GET form composing `SearchField`, `SelectField` (account), `SelectField` (month, absent when fewer than two months are offered and none is chosen), `SegmentedControl` (direction) and the `Clear filters` link, emitting one `onfilter(next)`. At compact the three choices move into a native `<dialog>` sheet behind a `Filters` button. Read: the `.filter` row and `.chip` in `pages/TransactionsScreen.svelte` name an account and cannot choose one, and are removed; `molecules/PageScrubber.svelte` moves through pages without cutting the list, which is the opposite job.

## States left out

- **A filter the address gets wrong.** The route drops it and redirects, so the screen never renders it. Held by the route test.
- **An account in the address that is not theirs.** Same: the route drops it. Held by the route test, which asserts the three cases redirect identically.
- **Changing a filter starts at the newest, Back undoes a filter, Returning to a filtered view, Paging keeps the filters.** Routing, not a state. Held by the route test on every link's address and on the address a filter change produces.
- **Nothing is found that the member may not see; A left-out account is not searched.** Nothing to show. Held by the store tests.
- **Case, accents and a literal `%`.** The screen looks the same whatever matched. Held by the store tests.
- **The months follow the search, the account and the direction.** Other entries in the same select. Held by the store test on `LedgerMonths`.
- **Syncing while filtered.** `SyncingWithRowsAlreadyHeld` already shows the only difference, the Refresh button's words.

## Contracts for implementation

- `LedgerFilters` is one `role="search"` landmark named `Filter transactions`. Tab order at medium and up: Search, the search's clear button when it holds text, Account, Month, Direction (one tab stop, arrow keys inside, as `SegmentedControl` already does), `Clear filters`. At compact: Search, its clear button, `Filters`; the sheet traps focus while open and `Done` returns it to `Filters`. Asserted in `LedgerFilters`' story.
- A pause in typing calls `onfilter` once with `q` trimmed and `live` true, never once per letter; Enter calls it at once with `live` false and the pause adds nothing; the clear button calls `onfilter` with `q` empty and returns focus to the field. Choosing an Account or Month option, or a Direction, calls `onfilter` once with every other value unchanged. Asserted with `fn()` spies in the `LedgerFilters` story.
- When `onfilter` is absent the form submits as a GET to `/transactions` with fields named `account`, `q`, `month`, `direction`. Asserted by reading the form's `action`, `method` and field names.
- `TransactionsScreen`'s existing polite status region reads `{n} transactions` while a filter is in force and rows exist, and `Nothing matches these filters` when none do, so a member whose focus stayed on the control hears the outcome. Asserted in `Searched` and `NothingMatches`.
- While `NothingMatches` shows its `Clear filters` action, the bar's own `Clear filters` is not rendered, so there is one link with that name. Asserted with `getAllByRole('link', { name: 'Clear filters' })` having length 1.
- `Clear filters` and `Show all accounts` are links to `clearHref`, and on activation move focus to the count heading, as `Show all accounts` does today. Asserted in `FiltersCombined`.
- The route applies a filter with `goto(url, { keepFocus: true, noScroll: true })`. Asserted by the route test on the address; focus is checked by hand in the wiring task.
- Medium and up lay the bar in one row that wraps, with `--space-2` between controls; compact (below 768, ADR 0002) is the search and the `Filters` button on one row, the sheet stacking its controls with `--space-3` between them. SegmentedControl keeps its own 32 height.

## Seen

The totals (tasks 7.x) and the lighter header (tasks 8.x) have not been looked at yet.

Looked at by a person on 2026-09-24: every story in State stories at compact, medium, wide and ultra, in light and dark, and the filter bar's height at 390 wide above the scrolling ledger. Nothing was wrong and nothing changed. The same versions were approved with `just approve`.
