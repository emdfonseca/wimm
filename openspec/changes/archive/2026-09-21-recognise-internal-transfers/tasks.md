## 1. Before anything else

- [x] 1.1 Confirm `spending-insights` is finished and archived, so
      `openspec/specs/banking/overview/spec.md` holds `Month by month, as far
      back as the ledger is whole` and `openspec/specs/banking/payment-patterns`
      exists. Verify: `just openspec list` no longer lists it,
      `just openspec validate recognise-internal-transfers` reports no
      archive warning, and `just check apps/wimm apps/web packages/ui` passes
      on the starting tree.
- [x] 1.2 Run the matcher's rule read-only against the household's local
      database, as SQL or a throwaway script kept in the scratchpad and never
      committed, with nothing written to the database. Report, for windows of
      0 to 5 days: the pairs found, how many carry evidence (the other
      account's `number_suffix`, `name` or `household_name` in the row's text)
      and how many do not, every row left unpaired by the ambiguity rule, and
      every unpaired out row whose counterparty is the member's own
      `holder_name` with its monthly total. Then, per scope and per month, money
      in, money out and net as they are now and as they would be. Verify: the
      report is given to the member, who says whether every pair is one they
      recognise, before any constant in design.md is fixed; if unpaired
      own-name transfers outweigh the pairs, stop and bring that back rather
      than continuing.

## 2. wimmd: the rule

- [x] 2.1 Record the decision with `/adr`: what a transfer between a member's
      own accounts is, every constant in design.md by name
      (`ownTransferWindowDays`, `ownTransferMinSuffix`,
      `ownTransferMinNameRunes`), the ranking and the refusal to pair a tie,
      why amounts are exact, why `holder_name` is not evidence, pairs across
      All with leaving-out per scope and why a crossing pair is counted, why no
      partner id is sent, why derived on read, and that it reverses the
      non-goal in `spending-insights`' design. Include what task 1.2 found.
      Verify: `just adr-index` regenerates `.claude/rules/decisions.md` with
      the new record.
- [x] 2.2 Test-first (`/tdd`), `banking.OwnTransfers` in
      `apps/wimm/internal/banking/transfers.go`, with design.md's constants as
      a `const` block. Verify: one table case per scenario of "What wimm calls
      a transfer between a member's own accounts" that the function decides:
      the next day, the same day, too far apart, not the same amount, two
      currencies, the same amount twice in a week, one arrival with two
      sources unpaired, evidence breaking that tie, a refund on the same
      account, a pending half, plus a result that does not depend on row
      order, a suffix under four characters that is not evidence, and 50,000
      rows inside the budget the test states; each case's failure was seen
      before its code.

## 3. wimmd: counting

- [x] 3.1 In `packages/contracts/proto/wimm/banking/v1/banking.proto` add
      `Transaction.own_transfer`, `DayMover.own_transfer`, and
      `HistoryMonth.transfers_left_out` with `transfers_total`, as design.md
      gives them; run `just gen`. Verify: `just check packages/contracts`
      passes and no file under `gen/` was hand-edited.
- [x] 3.2 Test-first, a `countedRows` helper in `apps/wimm/internal/banking`:
      given the member's owned rows, the pairs and the scoped account ids, the
      rows the scope counts and the pairs it leaves out. Give `scopedAccounts`
      the All set beside the scoped one, and the owned accounts' names and
      suffixes for evidence. Mirror in `memstore`. Verify: service-level table
      cases for a pair inside the scope, a pair crossing it in each direction,
      and no pair for a member who owns one side only.
- [x] 3.3 `Service.MonthSummary` (`month.go`): read `store.OwnedBooked` for
      the owned accounts from `priorStart − ownTransferWindowDays` to
      `windowEnd + ownTransferWindowDays`, sum both windows from counted rows,
      and take top merchants and largest payments from them. Remove
      `store.OwnedWindowSums` and its store and memstore tests, moving the
      granted-account and pending cases onto `OwnedBooked` where they are not
      already. Verify: the existing month tests pass unchanged, plus one test
      per scenario of "The month so far" that changed: a transfer to savings in
      neither figure, a pair across the end of the month in neither month, and
      a month holding only a transfer giving no summary.
- [x] 3.4 `Service.MonthHistory` (`history.go`): read `OwnedBooked` for the
      owned accounts, sum months from counted rows, hand counted rows to
      `marksFor`, `merchantMedians`, the risers and `RecurringPayments`, and
      set each month's `transfers_left_out` and total in the out row's month.
      Remove `store.OwnedMonthlySums` with its tests. Verify: one service test
      per scenario of "A transfer inside the scope is left out of what is
      counted": the month with a transfer to savings, paying into the joint
      account under Yours, under Household and under All, both owners reading
      the same Household month figure by figure, a large transfer not unusual,
      a standing transfer not recurring under All and recurring under Yours;
      and of "A month says how many transfers it left out", including the
      singular's count of one and twelve months' counts adding up to the
      number of pairs.
- [x] 3.5 `Service.BalanceTrend` (`trend.go`) and `Service.Transactions`
      (`ledger.go`): the shared pattern helper returns pairs beside marks; a
      mover and a ledger row that is half of a pair carries the transfer mark
      and never the unusual one; the balance line is untouched. Serve the new
      fields in `apps/wimm/internal/rpc/banking.go`. Verify: a trend test that
      the points' balances are byte-identical with and without a pair in the
      ledger, a mover labelled on the day the money left, a ledger test that
      both halves are marked and that reading one account still marks the
      half in it, a test that the other owner of a joint account gets no mark
      on the arriving row, and an rpc test for the three fields;
      `just check apps/wimm` passes.

## 4. packages/ui: components before the screen

- [x] 4.1 In `molecules/LedgerRow.svelte`, add `transfer` to the status tag
      with canvas.md's precedence. Stories: `BetweenYourAccounts`,
      `CompactBetweenYourAccounts`, `BetweenYourAccountsArriving`. Verify: the
      existing LedgerRow stories still pass; a play function gives a row both
      `transfer` and `unusual` and finds `Between your accounts` and not
      `Unusual`; `just check packages/ui` passes.
- [x] 4.2 In `molecules/MonthDetail.svelte`, add `leftOut`. Stories:
      `DetailWithTransfersLeftOut`, `DetailWithOneTransferLeftOut`. Verify:
      play functions assert both whole strings, and `Detail` asserts the
      absence of `left out`.
- [x] 4.3 In `molecules/BalanceChart.svelte`, add `transfer` to
      `BalanceMover`, in the tag and in the live region. Story:
      `ATransferLeaving`. Verify: the play function asserts the live region's
      whole string from canvas.md and the absence of `Unusual`.

## 5. packages/ui: the presentational screens and their state stories

- [x] 5.1 In `packages/ui/src/pages/Overview.svelte`, take `transferNote` per
      currency section and drop the hard-coded sentence, pass `transfer` to
      recent transactions and largest payments, and pass `leftOut` to the
      selected month. In `Overview.stories.svelte` and `Overview.fixture.ts`
      write or change one state story per Overview row of canvas.md's State
      stories, each play function asserting the words and the absences
      canvas.md gives it; add the two new stories to `unplaced` in
      `apps/storybook/canvas/flows.js`. Verify: `just check packages/ui` and
      `just check apps/storybook` pass; `grep -rn "own accounts is counted"
      packages/ui/src` finds nothing; `YoursScope` fails if it is given the
      All sentence.
- [x] 5.2 In `packages/ui/src/pages/TransactionsScreen.svelte`, pass each row
      its `transfer`. Bring `AsItOpens` and `OneAccount` in
      `TransactionsScreen.stories.svelte` to canvas.md's words. Verify:
      `AsItOpens` finds `Between your accounts` exactly twice, `Unusual`
      still once on `Galp`, and `Not settled` still on `Transfer to Ana Reis`.

## 6. The look

- [x] 6.1 Open every state story of canvas.md on the design canvas
      (`just canvas`) at compact, medium, wide and ultra. Change what is
      wrong, settle whether `Between your accounts` fits the tag at Compact or
      moves to the row's second line, then write what was seen and what
      changed into canvas.md under Seen. Verify: Seen names every state story
      at every regime.

## 7. apps/web: wiring

- [x] 7.1 In `apps/web/src/lib/overview.ts` and `insights.ts`, build the
      transfers sentence from the answered scope (the All sentence when no
      control is on offer), the month's `leftOut` line with its singular, and
      `transfer` on chart movers, recent transactions and largest payments,
      the label taking the slot from `unusual`. Verify: a Vitest case per
      sentence in canvas.md's Words that the load builds, including
      `1 transfer between your accounts left out · €500.00`, both scope
      sentences, and a row sent both marks coming out with `transfer` only;
      `grep -rn "own accounts is counted" apps/web/src` finds nothing;
      `just check apps/web` passes.
- [x] 7.2 `apps/web/src/routes/(app)/transactions/+page.server.ts` passes
      `transfer` to the screen from `own_transfer`. Verify: a Vitest case;
      `just check apps/web` passes.

## 8. Verification

- [x] 8.1 `just check apps/wimm`, `just check apps/web`,
      `just check packages/ui`, `just check packages/contracts` and
      `just check apps/storybook` all pass.
- [x] 8.2 Walk the signed-in app with the household's real data at Compact and
      Wide: every row labelled `Between your accounts` is a transfer the member
      made and none they know of between two connected accounts is missing
      without a reason the rule gives; one month added up by hand from
      Transactions, leaving the labelled rows out, agrees with that month
      under All; Household plus Yours still adds up to All for money in and
      money out once crossing transfers are accounted for by the narrower
      sentence; the typical month under Yours is one the member believes; and
      both owners of the joint account see the same figures under Household.
