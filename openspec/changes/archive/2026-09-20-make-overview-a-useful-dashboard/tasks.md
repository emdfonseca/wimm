## 1. wimmd: names, month summary, daily chart

- [x] 1.1 Test-first (`/tdd`), add `banking.MerchantName(counterparty,
      remittance)` in `apps/wimm/internal/banking/merchant.go` with design.md's
      seven rules as a `var` table. Verify: a table test covers the five
      walk-through lines (`Bxv Via Verde`, `Amazon`, `Number 1 Hair`,
      `Rest Botafogo`, `Paypal Converse`), a counterparty name, a line that
      would empty out, and empty input giving `Card payment`.
- [x] 1.2 In `packages/contracts/proto/wimm/banking/v1/banking.proto` add
      `Transaction.display_name`, `CurrencyTrend.short_history`, and the
      `GetMonthSummary` RPC with `CurrencyMonth` and `MerchantTotal` as
      design.md gives them; run `just gen`. Verify: `just check
      packages/contracts` passes and no file under `gen/` was hand-edited.
- [x] 1.3 Set `display_name` in `toProtoTransaction`
      (`apps/wimm/internal/rpc/banking.go`). Verify: an rpc test asserts a
      listed transaction carries the merchant name and still carries the
      bank's line in `remittance`.
- [x] 1.4 Rework `apps/wimm/internal/banking/trend.go`: one point per UTC
      day, at most 90 days, the contributor rule (30-day reach, else the
      single longest), `short_history` under 7 days, `partial_coverage` for
      any owned account left out, and no `CurrencyTrend` for a quiet
      currency. Verify: service tests for each `banking/overview` chart
      scenario, including one account with 90 days beside one with 1 day,
      and a zero-balance account with one old booked row yielding no trend.
- [x] 1.5 Add the month queries to
      `apps/wimm/internal/store/transactions.go`: signed booked sums for two
      windows per account set, and the window's outgoing booked rows. Verify:
      store tests against Postgres, pending rows excluded, both windows
      bounded by arguments and never by a process clock.
- [x] 1.6 Add `Service.MonthSummary` in `apps/wimm/internal/banking`: windows
      from `s.store.Now`, the same-days rule with the shorter-month cap,
      `prior_*` only when every contributing ledger reaches the prior month's
      first day, `counted_from`, top five merchants grouped on the lower-cased
      merchant name, five largest payments, owned accounts only, nothing for a
      quiet currency. Mirror it in `memstore`/`bankingtest`. Verify: one
      service test per scenario under "The month so far" and "Where the most
      money went", including the 31st against a 30-day month and a member who
      owns nothing getting an empty response.
- [x] 1.7 Record the decision with `/adr`: household money and a member's own
      money, superseding ADR 0019's "Totals are per member". Verify:
      `just adr-index` regenerates `.claude/rules/decisions.md` and ADR 0019's
      status names the new record.
- [x] 1.8 Test-first, classify accounts in `apps/wimm/internal/banking`
      (`HOUSEHOLD`, `OWN`, `SHARED`) from one store query counting, per
      account, the members holding an owner row or a *details* grant; replace
      `toProtoTotals` with `household_totals` and `own_totals`; add
      `Account.group` to the proto and run `just gen`. Verify: one service
      test per scenario under "Household money and a member's own money are
      separate figures", including the household of three, the member who
      joins, and a household of one.
- [x] 1.9 Serve `GetMonthSummary` in `apps/wimm/internal/rpc/banking.go`.
      Verify: `just check apps/wimm` passes.

## 2. packages/ui: components before the screen

- [x] 2.1 Balance chart: `molecules/BalanceChart.svelte` as hand-written SVG
      and `BalanceChart.stories.svelte` with stories for as it opens, a day
      pointed at, the coverage sentence, and short history, to canvas.md's
      Contracts. Verify: a play function moves the marker with the arrow
      keys, Home and End, and asserts the readout and the accessible name;
      the short-history story asserts "A balance chart appears once there is
      a week of history."; `just check packages/ui` passes.
- [x] 2.2 `molecules/BudgetMeter.svelte` and `BudgetMeter.stories.svelte`: a
      label, a value and a track filled to a proportion. Verify: stories at
      100%, 45% and 4% fill, each play function asserting the label and the
      value it was given (`Pingo Doce`, `€412.60 · 9 payments`), since no
      other check holds a merchant row's words; `just check packages/ui`
      passes.
- [x] 2.3 Confirm `molecules/AccountRow.svelte` renders Balance over Reading
      as a link with no connection controls; add that read-only form and its
      story if it does not. Verify: a story 56 high whose play function finds
      one link holding the name, the bank, the balance and the reading, and
      no button.
- [x] 2.4 In `molecules/MetricTile.svelte`, render the Delta row (arrow,
      change, period, under the value) in `color-text-secondary` in both
      directions, and let the tile be a link when given an `href`. Verify:
      stories with and without a delta, and one as a link; the tile fills a
      102-high row with a delta and an 82-high row without.
- [x] 2.5 Give `molecules/LedgerRow.svelte` the bank's line under the name,
      shown only when it differs, and drop its date at Medium and wider.
      Verify: the stories `BanksLineShown`, `BanksLineSameAsName` and
      `CompactBanksLineShown` with the words canvas.md gives them; 56 high
      either way at Wide, 80 and 64 at Compact.
- [x] 2.6 Page scrubber: in `molecules/PageScrubber.svelte` show the label
      only for the page under the pointer or focus, never for the current
      page. Verify: the stories at newest, middle and oldest still pass; `No
      label at rest` asserts no span text is visible and `Label under the
      pointer` asserts the focused page's span and no other;
      `just check packages/ui` passes.

## 3. packages/ui: the presentational screens and their state stories

- [x] 3.1 Rebuild `packages/ui/src/pages/Overview.svelte` on fixtures, to
      canvas.md's Contracts, composing the components from group 2. Write one
      state story in `Overview.stories.svelte` per Overview row of canvas.md's
      State stories, each play function asserting the words canvas.md gives
      that state, merchant rows included, and the absences it names. Verify:
      `just check packages/ui` passes, and removing "Money moved between your
      own accounts is counted." from the screen fails `Populated`.
- [x] 3.2 In `packages/ui/src/pages/TransactionsScreen.svelte`, pass each
      row its bank's line, drop the Date column, put the date in the day
      header ("Today, 17 September"), and show the span once in the footer.
      Bring `AsItOpens`, `AnOlderPage`, `TheOldestPage` and `Compact` in
      `TransactionsScreen.stories.svelte` to the words canvas.md gives them.
      Verify: `AsItOpens` asserts a bank's line under `Pingo Doce`, none
      under `Salary`, no `Date` heading, and the span exactly once;
      `just check packages/ui` passes.

## 4. The look

- [x] 4.1 Open every state story of canvas.md on the design canvas
      (`just canvas`) at compact, medium, wide and ultra. Change what is
      wrong, then write what was seen and what changed into canvas.md under
      Seen. Verify: Seen names every state story at every regime.

## 5. apps/web: wiring

- [x] 5.1 Rewrite `apps/web/src/routes/(app)/+page.server.ts` and
      `+page.svelte`: add the `GetMonthSummary` call to the parallel load,
      pass the accounts through grouped as Household, Yours and Shared with
      you with the two figures, drop the tiles of a currency with zero figures
      and neither a trend nor a summary, format every amount and date
      in the load, build the chart's summary sentence, read `display_name`.
      Verify: a Vitest case for the tile drop and one for the chart's summary
      sentence; `just check apps/web`.
- [x] 5.2 `apps/web/src/routes/(app)/transactions/+page.server.ts` reads
      `display_name` and passes the bank's line to the screen. Verify: the
      `counterpartyName || remittance` chain exists nowhere under
      `apps/web/src`; `just check apps/web` passes.
- [x] 5.3 Money arriving carries `+` wherever an amount is formatted
      (`apps/web/src/lib/money.ts`). Verify: a Vitest case for a positive, a
      negative and a zero amount; Transactions shows `+€151.99`.

## 6. Verification

- [x] 6.1 `just check apps/wimm`, `just check apps/web`, `just check
      packages/ui` all pass.
- [x] 6.2 Walk the signed-in app with the household's real data at Compact
      and Wide: no `$0.00` tile or flat chart for the unused currency, the
      EUR chart spans more than a day or says why not, the five walk-through
      payments read as design.md lists them, each group of accounts adds up to its figure, household money reads the
      same signed in as either member, and the page fills the Main column.
