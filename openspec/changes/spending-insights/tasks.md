## 1. Before anything else

- [x] 1.1 Confirm `make-overview-a-useful-dashboard` is finished: every task
      ticked, its Seen written, and the change archived so
      `openspec/specs/banking/overview/spec.md` holds the month summary, top
      spending and the balance chart. This change builds on its
      `GetMonthSummary`, `MerchantName`, `BalanceChart` and ADR 0024, and is not
      started until that is true. Verify: `just openspec list` no longer lists
      it, and `just check apps/wimm apps/web packages/ui` passes on the
      starting tree.

## 2. wimmd: the rules

- [x] 2.1 Record the decision with `/adr`: the detection rules. Recurrence, the
      one unusual-payment rule and the explains-the-day rule, every constant
      in design.md by name (`recurLookbackMonths`, `recurMinOccurrences`,
      `recurMinOccurrencesYearly`, `recurAmountTolerance`, the cadence ranges
      and slacks, `unusualScore`, `unusualMADScale`, `unusualMinPartyPayments`,
      `unusualGeneralDays`, `unusualFlatRatio`, `unusualFirstPaymentShare`,
      `unusualFloorShare`, `dayCoverShare`, `dayMaxRows`, `historyMonths`,
      `minFullMonths`), why yearly stands at two and says likely, why money in
      is judged only against money in, why a first payment has its own share,
      why log amounts with median and MAD, why not a
      percentile, and why marks are derived on read and never stored. Verify:
      `just adr-index` regenerates `.claude/rules/decisions.md` with the new
      record.
- [x] 2.2 Test-first (`/tdd`), `banking.RecurringPayments` in
      `apps/wimm/internal/banking/recurring.go`, with design.md's constants as
      a `const` block and the cadence table as a `var`. Verify: one table case
      per scenario of "What wimm calls a recurring payment", including the
      price that rose, the weekly supermarket, only twice, the weekend shift,
      two runs at one merchant, the cancelled one, the late one, a salary, a
      pending third payment, a yearly payment seen twice coming back `likely`
      and a third clearing it; each case's failure was seen before its code.
- [x] 2.3 Test-first (`/tdd`), `banking.UnusualPayments` in
      `apps/wimm/internal/banking/unusual.go`: the modified z-score on log
      amounts in both directions, each against its own, the party baseline at
      five payments, the 90-day general baseline, the zero-MAD ratio, the
      first-payment share, the floor per direction, and `typical`. Verify: one table
      case per scenario of "What wimm calls an unusual payment", including the bonus, the salary that is not unusual, a first payment
      over and under the share, plus a general baseline with a realistic
      spread, a baseline under five rows that is not
      judged, and a case showing the judged payment does not move its own
      median.
- [x] 2.4 Test-first (`/tdd`), `banking.DayMovers` in
      `apps/wimm/internal/banking/movers.go`: the 80% cover, the cap of five
      and the smaller count. Verify: table cases for one row covering the day,
      two rows needed, twelve similar rows giving five and seven, money in both
      directions, and an empty day.

## 3. wimmd: scope, months, chart

- [x] 3.1 In `packages/contracts/proto/wimm/banking/v1/banking.proto` add
      `InsightScope`, `Cadence`, the `GetMonthHistory` RPC with its messages,
      `scope` on `GetMonthSummaryRequest` and `GetBalanceTrendRequest`,
      `TrendPoint.movers` and `smaller`, and `Transaction.unusual`, as
      design.md gives them; export the new types from
      `packages/contracts/src/banking.ts`; run `just gen`. Verify:
      `just check packages/contracts` passes and no file under `gen/` was
      hand-edited.
- [x] 3.2 Give `OwnedWindowSums`, `OwnedOutgoing` and `OwnedAccountsForTrend`
      in `apps/wimm/internal/store` an `accountIDs` argument applied beside the
      `account_owners` join, add `OwnedMonthlySums`, and add `OwnedBooked` for
      both directions with `OwnedOutgoing` a filter over it. Verify: store tests
      against Postgres, including one that passes the id of an account the
      member holds *details* on and does not own and gets nothing for it,
      pending rows excluded, every window bounded by arguments, and
      `OwnedBooked` over 50,000 seeded rows inside the budget the test
      states.
- [x] 3.3 Test-first, `scopedAccounts` in `apps/wimm/internal/banking`: the
      three account sets, `available`, the answered scope, and the
      counted and not-counted names. Apply it to `MonthSummary` and
      `BalanceTrend`. Mirror in `memstore`/`bankingtest`. Verify: one service
      test per scenario of "One scope for everything drawn from transactions"
      that the service decides, including a household of one, nothing of their
      own, an unavailable scope answered as All, and the month summary under
      Household counting the joint account only.
- [x] 3.4 Add `Service.MonthHistory` in `apps/wimm/internal/banking`: windows
      from `s.store.Now`, fullness, `held_from`, typical and average with and
      without unusual payments, risers against each merchant's median month,
      each month's unusual payments in both directions with `typical` or
      `first_payment`, `net_usual` with both removed, and the recurring list
      with `likely`.
      Verify: one service test per scenario of "Month by month", "A typical
      month and an average month", "The merchants behind a month's rise" and "A
      month says how many unusual payments it held", including a ledger from 12
      April, a second bank from August, two full months, a month with a bonus, and the median of an
      even number of months in minor units.
- [x] 3.5 In `apps/wimm/internal/banking/trend.go`, group each contributing
      account's rows by day into `movers` and `smaller`, label unusual ones,
      and leave an account the chart drops out of them. Verify: service tests
      for a day with two movers and three smaller, an empty day, a day holding
      money in, and a dropped account's rows appearing on no day.
- [x] 3.6 Set `Transaction.unusual` in `ListTransactions`
      (`apps/wimm/internal/banking/ledger.go`), judged under All, skipped for a
      page older than the months shown. Serve `GetMonthHistory` and the scope
      fields in `apps/wimm/internal/rpc/banking.go`. Verify: an rpc test that
      the payments `GetMonthHistory` counts in March under All are exactly the
      rows `ListTransactions` marks; `just check apps/wimm` passes.

## 4. packages/ui: components before the screen

- [x] 4.1 Monthly net chart: `molecules/MonthlyNetChart.svelte` as hand-written
      SVG and `MonthlyNetChart.stories.svelte` with `AsItOpens`,
      `AMonthPointedAt`, `AMonthBelowZero`, `PartlyHeldMonths` and
      `NoTypicalMonthYet`, to canvas.md's Contracts. Verify: a play function
      moves the marker with the arrows, Home and End and asserts the readout
      and the accessible name; `just check packages/ui` passes.
- [x] 4.2 Month row: `molecules/MonthRow.svelte` and `MonthRow.stories.svelte`
      with `Closed`, `ClosedWithUnusual`, `Open`, `OpenWithUnusual`, `OpenWithUnusualIncome`, `SoFar` and `HeldFrom`. Verify: play functions assert `aria-expanded` before and
      after Enter and Space, that focus stays on the button, and the words
      canvas.md gives each, `1 unusual` and `without it` included.
- [x] 4.3 In `molecules/BalanceChart.svelte`, turn the pointed-day readout into
      the popover and carry all of it in the live region. Stories:
      `ADayPointedAt` changed, and `ADayWithNoTransactions`, `MoneyArriving`,
      `UnusualIncomeArriving`, `TheFirstDay`, `ABusyDay`, `OpensFromTheKeyboard` and `OpensFromATap`
      added. Verify: `OpensFromTheKeyboard` presses left, Home and End and
      asserts the popover and the live region's whole string each time;
      `OpensFromATap` asserts it opens on a tap and clears on a tap outside;
      no story reaches the popover by hover alone.
- [x] 4.4 In `molecules/LedgerRow.svelte`, add `unusual`, `note` and `href`,
      and let the date column fit `Was expected 18 Sep`. Stories: `Unusual`,
      `CompactUnusual`, `UnusualIncome`, `ExpectedDate`, `WithANoteAndALink`. Verify: the
      existing LedgerRow stories still pass; `WithANoteAndALink` finds one link
      holding the name, the date, the amount and the note; rows stay 56 high
      at Wide.

## 5. packages/ui: the presentational screens and their state stories

- [x] 5.1 In `packages/ui/src/pages/Overview.svelte`, on fixtures, add the
      scope control, the Month by month card and the Recurring payments card
      in canvas.md's order, composing the components from group 4. In
      `Overview.stories.svelte` write or change one state story per Overview
      row of canvas.md's State stories, each play function asserting the words
      and the absences canvas.md gives it. Verify: `just check packages/ui`
      passes; removing "They are still listed in their months." from the screen
      fails `WithoutUnusualPayments`; `HouseholdOfOne` fails if the scope
      control renders.
- [x] 5.2 In `packages/ui/src/pages/TransactionsScreen.svelte`, pass each row
      its `unusual`. Bring `AsItOpens` in `TransactionsScreen.stories.svelte`
      to canvas.md's words. Verify: `AsItOpens` finds `Unusual` on the `Galp`
      row, exactly once on the page, and `Not settled` still on the transfer.

## 6. The look

- [x] 6.1 Open every state story of canvas.md on the design canvas
      (`just canvas`) at compact, medium, wide and ultra. Change what is wrong,
      settle the bar width and whether Compact shows every month row, then
      write what was seen and what changed into canvas.md under Seen. Verify:
      Seen names every state story at every regime.

## 7. apps/web: wiring

- [x] 7.1 In `apps/web/src/routes/(app)/+page.server.ts` and `+page.svelte`:
      read `scope` from the address, pass it to `GetMonthSummary`,
      `GetBalanceTrend` and the new `GetMonthHistory` call in the parallel
      load, redirect to the bare address when the answered scope is not the
      one asked for, and navigate on `onscope`. Verify: Vitest cases for each
      scope reaching all three calls, for an unavailable scope redirecting, and
      for no control when `available` is empty; `just check apps/web` passes.
- [x] 7.2 In the same load, format the months: every amount and date, the
      chart's accessible-name sentence, the typical sentence in both
      directions and both views, `N unusual` with its singular, `usually about`, `a usual payment is about`
      and `first payment to`, `Likely yearly`, the narrower
      Household sentence, each unusual payment's `/transactions?page=` address,
      and each day's `change`, movers and `smaller`. Verify: a Vitest case per
      sentence in canvas.md's Words that the load builds, including `1 unusual
      · … without it`, `No change from 8 Aug` and `and 1 smaller`.
- [x] 7.3 `apps/web/src/routes/(app)/transactions/+page.server.ts` passes
      `unusual` to the screen. Verify: a Vitest case that
      `?page=<date>.<id>` lands with that payment as the first row;
      `just check apps/web` passes.

## 7b. Sections that stay in place across scopes

- [x] 7b.1 Spec scenarios and canvas state (`HouseholdScopeWithNoHistory`)
      revised. Verify: the three new scenarios exist under
      `banking/overview`.
- [x] 7b.2 wimmd returns a currency with no full month or no recurring
      payment when the scope control is on offer, flagged so the web can tell.
      Verify: service tests per new scenario; `just check apps/wimm`.
- [x] 7b.3 The web load builds the empty-state lines and the view control's
      presence. Verify: Vitest cases; `just check apps/web`.
- [x] 7b.4 Overview renders them; story `HouseholdScopeWithNoHistory` with a
      play function. Verify: `just check packages/ui apps/storybook`.

## 8. Verification

- [x] 8.1 `just check apps/wimm`, `just check apps/web`,
      `just check packages/ui`, `just check packages/contracts` and
      `just check apps/storybook` all pass.
- [x] 8.2 Walk the signed-in app with the household's real data at Compact and
      Wide: the months agree with Transactions for one month added up by hand;
      every recurring payment listed is one the household recognises and none
      they know of is missing without a reason the rules give; every unusual
      payment is one a person would call unusual, and the constants are tuned
      in the ADR's terms if not; each scope's month summary adds up to All's;
      the popover opens by mouse, by keyboard and on a phone; and both owners
      of a joint account see the same figures under Household.

## 7c. A young ledger does not hold the months back

- [x] 7c.1 Spec, design and canvas revised: full months follow the earliest
      ledger; accounts missing from a full month are named. Verify: the
      scenario `A second bank connected last month` and the canvas state
      `AYoungAccountBesideOlderOnes` exist.
- [x] 7c.2 `MonthHistory` uses the earliest ledger and returns `late_ledgers`.
      Verify: service tests for a young account beside an old one, the
      2026-03-19 / 2026-09-19 case, and an account that begins before the first
      full month; `just check apps/wimm`.
- [x] 7c.3 The web load words one line per late ledger and Overview renders it.
      Verify: Vitest cases; story `AYoungAccountBesideOlderOnes` with a play
      function; `just check apps/web packages/ui apps/storybook`.

## 7d. Months are a table with one detail panel

- [x] 7d.1 Spec, design and canvas revised: a month is selected, not opened;
      the chart's marker and the table share one selection. Verify: canvas.md
      names `MonthTable` and `MonthDetail` stories.
- [x] 7d.2 `MonthRow` becomes `MonthTable` and `MonthDetail`, wired into
      Overview, stories with play functions asserting `aria-pressed`, the
      selection shared with the chart, and the words of each month. Verify:
      `just check packages/ui apps/storybook`.
- [ ] 7d.3 Look at the changed stories on the design canvas at Compact, Medium,
      Wide and Ultra and write what was seen. Verify: canvas.md Seen names every
      changed story at every regime.
