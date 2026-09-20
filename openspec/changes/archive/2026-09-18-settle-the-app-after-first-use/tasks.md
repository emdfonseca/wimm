## 1. The schema, and the invariant it holds

- [x] 1.1 Write one goose migration in `apps/wimm/internal/store/migrations/`
  adding `accounts.left_out_at timestamptz` and `accounts.household_name text`.
  Verify `just check apps/wimm` passes and `migrate_test.go` shows the columns
  present after up and gone after down.
- [x] 1.2 In the same migration, backfill every account with no owner row: owned
  by its connection's `connected_by`, and `left_out_at = now()`. End the
  backfill with an assertion that no ownerless account remains and fail the
  migration if one does. Verify with a test that seeds three orphans, one of
  them the connection's last account, and asserts all three come back owned and
  left out.
- [x] 1.3 In the same migration, backfill `transactions.booking_date` for any
  row dated at the zero time from its value date, then its transaction date,
  then `alter column booking_date set not null`. Verify with a test that seeds a
  zero-dated row and asserts it ends with the value date and the column is not
  null.
- [x] 1.4 In the same migration, after the backfills, create deferred constraint
  triggers on `account_owners` (after delete and update) and on `accounts`
  (after insert) that refuse a commit leaving an account with no owner, and that
  skip an account that no longer exists. Verify with tests that: setting owners
  to the empty set is refused; delete-then-insert inside one transaction is
  accepted; deleting an account succeeds; deleting a connection succeeds; and
  deleting a member who is an account's last owner is refused.
- [x] 1.5 Regenerate nothing by hand. Verify `just check apps/wimm` passes.

## 2. The contract

- [x] 2.1 In `packages/contracts/proto/wimm/banking/v1/banking.proto`, require at
  least one `member_id` on `SetAccountOwners` with `buf.validate`, and delete the
  comment saying the list may be empty. Verify `just gen` runs clean and
  `just check packages/contracts` passes.
- [x] 2.2 Add `SetAccountLeftOut` and `SetAccountName` to `BankingService`, and
  `left_out_at` and `household_name` to `Account`. Verify `just gen` and
  `just check packages/contracts`.
- [x] 2.3 Add an oldest-page mode and a month cursor to `ListTransactionsRequest`,
  and the months that hold transactions to `Ledger`. Verify `just gen` and
  `just check packages/contracts`.

## 3. The store and the service

- [x] 3.1 `banking_accounts.go`: reduce `ReadableAccounts`' owner-or-grant test to
  `left_out_at is null`, and return a left-out account to its owners only from
  `visibleAccountsQuery`. Verify with store tests that a left-out account is read
  by nobody and appears for its owners and for nobody else.
- [x] 3.2 `banking_accounts.go`: add `SetAccountLeftOut` and `SetAccountName`, and
  make every name read `coalesce(household_name, name)` — including the three
  order-by clauses. Verify with a store test that renames an account, runs
  `ReplaceConnectionAccounts` with the gateway's own name, and asserts the
  household's name survived.
- [x] 3.3 `transactions.go`: add `and a.left_out_at is null` to the owned-accounts
  scope in all four places it appears. Verify with a test that leaving an account
  out removes its rows from the ledger, the count and the span, and that bringing
  it back restores them.
- [x] 3.4 `transactions.go`: add the oldest-page branch (`order by booking_date
  asc, id asc`, reversed in Go) and the month seek (`(booking_date, id) <= (last
  day of month, greatest id that day)`). Verify with tests that a jump lands on a
  page contiguous with the ones either side, with nothing repeated and nothing
  skipped.
- [x] 3.5 `transactions.go`: add the month aggregate, scoped to the member's own
  not-left-out accounts, ordered newest first, returning no empty month. Verify
  with a test that a household with nothing in March is not offered March, and
  measure the query on a seeded ledger of 50,000 rows.
- [x] 3.6 `service.go` and `reading.go`: delete the orphan branches from
  `requireOwner` and `requireOwnerOnConnection`, and turn
  `TestDisowningYourLastAccountDoesNotLockYouOut` into a test of the refusal.
  Verify `just check apps/wimm`.
- [x] 3.7 `rpc/banking.go`: omit the balance and its read time for a left-out
  account in the handler, not in the query, and serialise the household name.
  Verify with a handler test that the field is absent from the response rather
  than null.
- [x] 3.8 The Enable Banking adapter falls back through value date then
  transaction date before writing a transaction, and refuses one with no date at
  all. Verify with an adapter test against `bankingtest`.

## 4. Library components, before the screens that use them

Each task below is one component in three places: the origin in
`product-ui.lib.pen`, the Svelte component in `packages/ui/src/<layer>/`, and its
`.stories.svelte` beside it. The origins already exist — they were drawn as part
of this change's canvas — so each task is the Svelte half and the story.

- [x] 4.1 `packages/ui/src/atoms/NavigationProgress.svelte` from origin `Mdi1X`,
  with `Atoms/NavigationProgress.stories.svelte`. Story covers waiting, reduced
  motion, and the delay and minimum-visible timings. Verify `just check
  packages/ui`.
- [x] 4.2 `packages/ui/src/molecules/StepIndicator.svelte` from `W8V14`, with its
  story. Verify `just check packages/ui`.
- [x] 4.3 `packages/ui/src/molecules/StepActions.svelte` from `o1elL` and
  `prZYD`, with its story covering the row and the stacked form. Verify
  `just check packages/ui`.
- [x] 4.4 `packages/ui/src/molecules/MonthScrubber.svelte` from `W8ig0`, with its
  story covering a middle month, the newest, the oldest, and one month only.
  Verify `just check packages/ui`.
- [x] 4.5 `packages/ui/src/molecules/DensityControl.svelte` from `Yt1rq`, with its
  story covering fine and coarse pointers. Verify `just check packages/ui`.
- [x] 4.6 `packages/ui/src/molecules/AccountName.svelte` from `FguRX`, with its
  story covering a named account, an unnamed one, and clearing. Verify
  `just check packages/ui`.
- [x] 4.7 `packages/ui/src/templates/Page.svelte` from `BPW3V`, with its story
  covering all four regimes and the two optional slots. Verify `just check
  packages/ui`.
- [x] 4.8 `packages/ui/src/templates/SignedInLanding.svelte` gains the rail and
  ultra compositions from `WMlvF` and `NEIet`, and its story gains a case per
  regime. Verify `just check packages/ui`.
- [x] 4.9 `packages/ui/src/organisms/BottomNav.svelte` gains the Settings
  destination from `wCboc`. Verify `just check packages/ui`.
- [x] 4.10 `packages/ui/src/molecules/AccountChoiceRow.svelte` gains the owner set
  and the leave-out action from `QU4bL`'s `o0Tc2`, and its story gains the
  refusal. The ownership control must add an owner without removing any, which is
  the defect it has today. Verify `just check packages/ui`.
- [x] 4.11 `packages/ui/src/molecules/AccountRow.svelte` gains the left-out state
  from `NZhzh`, and its story gains it. Verify `just check packages/ui`.

## 5. One layout system

- [x] 5.1 Delete the 1024 breakpoint from `SignedInLanding.svelte` and
  `AuthShell.svelte`, and the four 599s from `AccountsOverview.svelte`,
  `ConsentExplainerScreen.svelte`, `AccountChoiceRow.svelte` and
  `DisconnectBankDialog.svelte`. Every one becomes 768, 1200 or 1800. Verify with
  a check that greps `packages/ui/src` and `apps/web/src` for any `@media
  (min-width` or `(max-width` not in {767, 768, 1199, 1200, 1799, 1800} and fails
  on a hit, and that the check is itself tested against a file carrying 1024.
- [x] 5.2 Replace the hard-coded 264, 56, 60 and `space-8` in
  `SignedInLanding.svelte`, `SidebarNav.svelte` and `BottomNav.svelte` with
  `--layout-sidebar-width`, `--layout-header-height` and `--layout-page-gutter`.
  Verify by asserting in a test that each token's declared value per regime is
  the one the component renders.
- [x] 5.3 Replace the six hand-rolled `.screen` rules and the four disagreeing
  `h1` blocks with `Page`. Verify each screen renders through it and
  `just check packages/ui` passes.
- [x] 5.4 Pass `compact` to `TransactionsScreen` from `transactions/+page.svelte`,
  or make the screen read the regime itself. Its compact layout is written and
  has never been switched on. Verify by loading the route at 390 and seeing
  stacked rows.
- [x] 5.5 Go through every screen at all four regimes against its frame, and
  record each deliberate difference with its reason — in `ACCEPTED` for copy, in
  this change's `canvas.md` otherwise. Verify `just check packages/ui`.

## 6. The screens

- [x] 6.1 The chooser: owner set, the refusal, leaving out, bringing back with the
  confirmation that names who will see it again. Route actions for each. Verify
  against frames `CJ3jD` `LvMRW` `X424Q` `BaB36` `navZJ` and with route tests
  covering the refusal and both confirmations.
- [x] 6.2 Overview: the left-out row, and the total that no longer counts it.
  Verify against `Oxc2l` and with a route test asserting the total drops by the
  account's balance.
- [x] 6.3 The rename, on the chooser and wherever an account's name is read.
  Verify against `D9X0CH` and `zIKSr`.
- [x] 6.4 Transactions: the month strip, and the Newest and Oldest jumps, wired
  to `?month=`, no query, and the oldest-page mode. Verify against `aHGv0` and
  with route tests that a month jump is contiguous with the pages either side.
- [x] 6.5 The pager keeps the span on a single page and offers no controls there.
  Verify against the spec scenario and a route test.
- [x] 6.6 The connect wizard: a step indicator and one action footer on all four
  steps, the lesser action first. Verify against `kEjpg` `PobEK` and their three
  regime twins.
- [x] 6.7 `apps/web/src/routes/+layout.svelte` gains the navigation indicator,
  driven by SvelteKit's `navigating`, with the 150 ms delay and the 400 ms floor
  and a polite live region. Verify against `nTE3A` and with a test that a fast
  navigation renders nothing and a slow one announces.
- [x] 6.8 A `/settings` route rendering appearance and density, with the theme
  control moved off the bottom of the signed-in layout. Verify against `aKoCs`
  `YyMbD` `jfrPt` `UUFeF` `hTkMR`.
- [x] 6.9 `apps/web/src/app.html` applies the stored density in the same inline
  script that already applies the theme. Verify by loading with a stored compact
  density and asserting no reflow after first paint.

## 7. The checks that missed all of this

- [x] 7.1 Add a fixture to `packages/ui/scripts/test-contract.py` for a frame
  named like the connect flow's happy path and one named like the access
  journey's, asserting the generator extracts their copy. Watch both fail.
- [x] 7.2 Widen `gen-canvas-contract.py`'s `SCREEN` pattern until both fixtures
  pass, and rename the unqualified frames in `02-banking.pen` and `01-access.pen`
  to the fully qualified form. Verify `just check packages/ui`.
- [x] 7.3 Add every newly covered frame to `check-canvas.py`'s `IMPLEMENTS`, and
  fix the screens until the check passes. This is where Sign in, Enrol a passkey,
  Choose a bank, What wimm will see and Overview's default state are asserted for
  the first time. Verify `just check packages/ui`.
- [x] 7.4 Add the Medium and Ultra measurements to `check-geometry.py`, including
  the rail at 72 and the sidebar at 288, and redraw `Transactions / Ultra 1920`
  from 264 to 288. Verify `just check packages/ui`.
- [x] 7.5 Run `just check` across every directory and `devbox run -- just ci`.
  Verify both pass.

## 8. The record

- [x] 8.1 Confirm `docs/decisions/0022-an-account-always-has-an-owner.md` still
  describes what was built, and regenerate the index with `just adr-index`.
  Verify `just adr-index-check` passes.
- [x] 8.2 Confirm every `.pen` file is saved and committed, and that Pen.app is
  not holding a stale copy of any of them. Verify `git status` shows no
  unexpected design changes.
