## Context

See proposal.md for why. This change is applied after `spending-insights` is
archived; everything below describes the code as that change leaves it.

- `store.Transaction` carries `CounterpartyName` and `Remittance` and nothing
  else about the other party. **No counterparty account number or IBAN is
  stored**, and ADR 0018 keeps the member's own full account numbers out of
  the database too. What an account has is `name`, `household_name`,
  `number_suffix` and `holder_name`.
- `banking.UnusualPayments`, `banking.RecurringPayments` and
  `banking.DayMovers` are pure functions over `[]store.Transaction` (ADR 0025).
  `Service.MonthHistory` reads `store.OwnedBooked` for the scoped accounts over
  25 months and `store.OwnedMonthlySums` for the figures;
  `Service.MonthSummary` reads `store.OwnedWindowSums` twice and
  `store.OwnedOutgoing`; `Service.BalanceTrend` reads `TransactionsSince` per
  contributing account and calls `unusualMarks`; `Service.Transactions` calls
  `marksUnderAll`.
- `scopedAccounts` resolves Household, Yours and All to owned account ids, and
  every owned store query applies them beside the `account_owners` join
  (ADR 0021).
- The sentence `Money moved between your own accounts is counted.` is a
  constant in `apps/web/src/lib/overview.ts` and a second, hard-coded copy in
  `packages/ui/src/pages/Overview.svelte`.
- Measured on the household's local ledger, read-only, on 21 September 2026:
  1,222 booked rows on six owned accounts. The rule pairs 14, all of which the
  member recognises: 5 at a gap of zero days, 7 at one, 2 at two, and none from
  three to five. Twelve carry evidence. The ambiguity rule refused nothing. One
  unpaired out row names the member's own `holder_name`, €160.80, so the
  transfers wimm holds one side of do not outweigh the pairs. Household leaves
  out no pair; Yours leaves out 4, all in one month; All leaves out all 14. No
  net figure moves in any scope, because both halves of a pair leave together.

## Language

- **Transfer between your accounts** — two booked rows wimm has paired as one
  movement between accounts the member owns. Code: `ownTransfer`,
  `OwnTransfers`, wire `own_transfer`. On a row the label is
  **Between your accounts**. Never "internal transfer" (internal to what),
  "self transfer", "contra", "movement", or "excluded": the rows are left out
  of a count, they are not excluded from anything a member can see.
- **Pair** — the two rows of one transfer. Code: `pair`. Never "match" on
  screen.
- **Half** — one row of a pair. Code: `partner` for the other one.
- **Left out** — not counted in a figure. The words on screen are `left out`.
  Never "excluded", "ignored", "removed", "hidden" or "netted".
- **Counted** — the opposite, and the word the old sentence already used.
- **Inside the scope** — both halves are on accounts the scope in force counts.
  Code: `inScope`. A pair with one half outside is **crossing** the scope, and
  is counted.
- **Evidence** — a row's text naming the other account. Code: `evidence`.
  Never shown on screen.

## Goals / Non-Goals

**Goals:**

- One matcher, pure, with its constants named, run once per request over rows
  already in hand, before anything is summed or judged. The detectors of
  ADR 0025 do not change: they are handed fewer rows.
- A pair is a fact about the member's rows; whether it is left out is a fact
  about the scope. The two are computed separately so neither leaks into the
  other.
- Two owners of a household account read the same Household figures.
- Nothing a member can see about a row says anything about an account they do
  not own.

**Non-Goals:**

- Recognising a transfer wimm holds one side of. A row whose counterparty is
  the member's own name, with nothing arriving anywhere wimm can see, stays
  counted. See Open Questions: the read-only run reports how many there are.
- Amounts that differ, by a fee or by a conversion. Exact and same-currency
  only.
- Transfers between two members' accounts (a partner paying a partner). It is
  real money changing hands, and pairing it would need one member's rows to be
  read for another, which ADR 0021 forbids.
- Persisting a pair, or letting a member confirm, reject or make one by hand.
  That is stored state and its own change, as it was for ADR 0025's marks.
- A filter on Transactions for transfers.

## Decisions

**`banking.OwnTransfers(rows, accounts)` in
`apps/wimm/internal/banking/transfers.go`.** Rows are every booked row of every
account the member owns, all currencies; `accounts` carries each owned
account's id, `name`, `household_name`, `number_suffix` and `holder_name`. It
returns `map[transactionID]partnerID`, holding both directions of every pair.

```text
ownTransferWindowDays   = 3    booking dates at most 3 days apart, either order
ownTransferMinSuffix    = 4    a number suffix shorter than this is not a token
ownTransferMinNameRunes = 4    nor is an account name shorter than this

candidate   = out row o and in row i with o.currency == i.currency,
              o.amount == −i.amount, o.account != i.account,
              |o.booking_date − i.booking_date| <= ownTransferWindowDays,
              both booked
tokens(a)   = an owned account's number_suffix when it is all digits and at
              least ownTransferMinSuffix long, plus its name and its
              household_name when at least ownTransferMinNameRunes long, each
              lower cased with spaces collapsed; MINUS every value carried by
              a second account the member owns
evidence    = the text of either row (counterparty name and remittance, lower
              cased, spaces collapsed) contains a token of the OTHER row's
              account
rank        = smaller date gap first; at equal gap, evidence before none
pairing     = repeat until nothing changes: pair (o, i) when i is o's single
              best-ranked unpaired candidate AND o is i's; remove both.
              A row whose best rank is shared by two candidates is not paired
              in that round, and may be in a later one once one of them has
              been paired elsewhere. What is left is unpaired
```

**A token is dropped when a second owned account carries it**, which is the
`holder_name` exclusion stated as the property rather than as a column. Three of
the household's accounts carry the same string in `name` as in `holder_name`, so
excluding the column and admitting the value said "mine" and not "which" all the
same; the two Revolut accounts share `name`, `household_name` and
`number_suffix`, and nothing distinguishes them because nothing does. **A number
suffix must be digits** for the same reason a length minimum was reached for: the
household's PayPal account has the suffix `.com`, which is four characters and is
not an account number. Neither rule moves a pair on the household's ledger —
evidence only breaks a rank tie, and the ambiguity rule refused nothing there.

The window is three because the ledger's own pairs end at two and a weekend
adds one; five would have found nothing more and doubles the room for a
coincidence. Buckets are keyed on `(currency, |amount|)`, so the work is a sort
per bucket and the 25-month row set costs milliseconds. *Alternative:* pair
greedily in date order. Rejected: with two outs and two ins of one amount it
pairs the first out with the first in even when the second is the same-day one,
and the result depends on row order. *Alternative:* require evidence for every
pair. Rejected for now: Portuguese banks write `TRF P/ ...` with a name as
often as a number, the household's rows will show how often, and the read-only
run in tasks.md reports every pair with and without evidence so this can be
made a constant (`ownTransferRequireEvidence`) in the ADR's terms if a false
pair shows up.

**`holder_name` finds the transfers wimm cannot pair, and only reports them.**
An out row whose counterparty is the member's own `holder_name` and which has
no partner is a likely transfer to an account wimm does not hold. It stays
counted, as the spec says. The read-only run lists them with their monthly
total, because on the household's data they may matter more than the pairs do.

**Pairs are found across All; leaving out is decided per scope.** Every service
method resolves two account sets: `owned` (the All set, always) and `scoped`.
Rows are read for `owned`, `OwnTransfers` runs once, and then:

```text
inScope(pair)   = both rows' accounts are in scoped
counted rows    = scoped rows, minus every row of an inScope pair
label           = every row of any pair, whatever the scope
```

So ada's €800.00 from her own account to the joint account is one pair for
ada. Under Yours the out row is scoped and its partner is not: counted, money
out. Under Household the in row is scoped and its partner is not: counted,
money in. Under All both are scoped: left out. For grace, who owns only the
joint account, there is no pair at all and the in row is money in. Household
therefore reads +€800.00 for both, which is S3, and a service test compares the
two owners' Household months figure by figure. *Alternative:* leave a pair out
wherever either half is in scope. Rejected: ada's Household would lose the
€800.00 grace's keeps, and Yours would stop showing that €800.00 leaves it
every month, which is the most useful thing Yours says.

**The figures are summed in Go from one row set.** A SQL sum cannot leave out a
row whose partner it has not seen. `MonthHistory` already holds
`OwnedBooked` over 25 months; it now reads it for `owned` rather than `scoped`
and sums months from the counted rows, and `OwnedMonthlySums` goes.
`MonthSummary` reads `OwnedBooked` from `priorStart − ownTransferWindowDays` to
`windowEnd + ownTransferWindowDays` (so a pair across the edge of a window is
still found), sums both windows from the counted rows, and `OwnedWindowSums`
goes; `OwnedOutgoing` stays a filter over the same rows. That keeps the month
summary a read of about two months, which is the property `spending-insights`
protected when it refused to fold everything into one RPC. `WindowSum.Rows`
becomes a count of counted rows, so a month holding only a transfer has no
summary, which is right: nothing was spent or received. *Alternative:* keep the
SQL sums and subtract the pairs in Go. Rejected: two sources for one figure is
the fault the generated-artifacts rule names, and the subtraction has to
re-derive which window each half fell in.

**A pair across two months is left out of both, and counted once in the month
its money left.** Each row is dropped from its own month's figures. The
left-out line (`transfers_left_out`, `transfers_total`) counts a pair in the out
row's month, so twelve months' counts add up to the number of pairs.

**The detectors see counted rows only.** `marksFor`, `RecurringPayments`,
`merchantMedians`, the risers, `topMerchants` and `largestPayments` are all
handed the counted rows of the scope in force. So a standing €800.00 to the
joint account is a recurring payment under Yours and nothing under All, and a
€6,000.00 move to savings is unusual nowhere it is a pair in scope. Under
`marksUnderAll` (Transactions) every pair is in scope by construction, so no
labelled row is ever also marked unusual there. On Overview under Yours a
crossing pair can be both counted and large; it may then be unusual in the
month's list, and its ledger row carries `Between your accounts` and not
`Unusual`. **The label wins the slot**: `own_transfer` is set on the wire
whenever the row is half of any pair, and the web load gives the status slot to
it. *Alternative:* never judge a paired row, in any scope. Rejected: under Yours
a first €5,000.00 payment into a new joint account really does bend the month.

**The balance chart keeps the rows and labels them.** The line is balances, and
does not change. `DayMovers` still receives every row of the contributing
accounts: where halves fall on different days, or one account is not on the
chart, the line moved and the popover has to say why. `DayMover` gains
`own_transfer`; a mover that is a pair is never `unusual`. `BalanceTrend`
already calls the shared pattern helper, which now returns pairs beside marks,
so the cost is the row set it already reads. *Alternative:* drop a same-day
pair whose accounts are both on the chart, since the line did not move.
Rejected: the day's gross movement still happened, and "No transactions this
day" beside two rows in Transactions is a contradiction.

**Wire: three fields, no RPC.**

```text
Transaction   += bool own_transfer          half of a pair, judged under All
DayMover      += bool own_transfer
HistoryMonth  += int32 transfers_left_out   pairs in scope whose money left
                 Money transfers_total       this month; total is positive
```

No partner id is sent. The browser has no use for it without a way to show the
other row, and an id is one more thing that could point at an account.

**Marks are per member and never reveal another member's account.** Pairing
reads only `owned` for the calling member, through store queries that join
`account_owners`. Grace's in row on the joint account has no partner in her row
set, so it is unlabelled for her. A test asserts it, because a label there
would tell her ada has an account holding at least €800.00.

**Words.** Status slot: `Between your accounts`, in `LedgerRow`'s existing tag,
`color-text-secondary`, beside `Unusual` and `Not settled` and never with
either. The sentence, built in `apps/web/src/lib/overview.ts` and passed to the
screen, which stops holding a copy:

```text
All, or no control    Money moved between your accounts is left out where wimm
                      holds both. Other transfers are counted.
Household or Yours    Money moved between accounts counted here is left out.
                      Money moved to or from your other accounts is counted.
month detail          2 transfers between your accounts left out · €1,400.00
                      1 transfer between your accounts left out · €500.00
```

**Derived on read, for ADR 0025's reason.** The window and the evidence rule
will be tuned, and a stored pair needs a backfill each time. A late-arriving
row can also turn an unpaired row into a pair, which a stored mark would have
to notice.

**One ADR.** The rule and its constants, pairs across All with leaving-out per
scope, why crossing pairs are counted, why exact amounts, why `holder_name` is
not evidence, why no partner id is sent, and that it reverses the non-goal in
`spending-insights`' design. It supersedes nothing: ADR 0025's rules stand.

## Risks / Trade-offs

- [A payment and an unrelated arrival of the same amount in another account
  within three days are paired, and a real payment vanishes from the figures] →
  The closest-date and ambiguity rules limit it, both rows are labelled in
  Transactions where a member will see it, the month says how many were left
  out, and the read-only run lists every pair without evidence before the
  constants are fixed. `ownTransferRequireEvidence` is the fallback.
- [The member's complaint is mostly transfers to accounts wimm does not hold,
  and this change does not move those] → The run reports them. If they
  dominate, that is a finding to bring back before the rest is built, not
  after.
- [Figures a member has already read change, downward, with no action of
  theirs] → The sentence under the figures changes in the same deploy, and the
  month's left-out line accounts for the difference.
- [A pair whose second half has not arrived yet is counted for a day or two,
  then left out] → True, and unavoidable without guessing. The month so far is
  the only month it can touch.
- [A transfer with a fee, or between currencies, is never paired] → Exact is
  the only rule that does not need a tolerance, and a tolerance is where false
  pairs come from.
- [Under Yours a row reads `Between your accounts` in Transactions and is still
  in the figures] → The narrower-scope sentence says exactly that.
- [`MonthHistory` under Household now reads every owned account's rows rather
  than the scoped ones] → The same row set All already reads, inside the
  budget `OwnedBooked`'s 50,000-row test states.
- [Removing `OwnedWindowSums` and `OwnedMonthlySums` removes store tests that
  held ADR 0021's line in SQL] → `OwnedBooked` keeps the same join and the same
  test; the granted-account case moves onto it.

## Migration Plan

No schema and no stored state. Deploy is the code; rollback is the previous
build, and the figures return to what they were.

## Open Questions

- Whether an unpaired row whose counterparty is the member's own name should
  one day be left out as a transfer to an account wimm does not hold. It
  changes what "money out" means (money that left the accounts wimm sees,
  against money that left the member), so it is its own decision. Task 1.1
  measures it; nothing in this change depends on the answer.
- Whether `Between your accounts` fits the status slot at Compact. The look
  settles it; the fallback is the same words on the row's second line.
