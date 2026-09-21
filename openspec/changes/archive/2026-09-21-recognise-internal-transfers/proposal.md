## Why

A member whose balance grows every month reads a negative typical month and a
negative average month under Yours. Money they move between their own accounts
is counted as money out of one account and money in to the other, so every
month looks more expensive than it was, a large transfer reads as an unusual
payment, and a standing transfer reads as a recurring one. The months, the
month summary and top spending shipped saying so in a sentence
(`Money moved between your own accounts is counted.`). Saying it does not make
the figures usable, and the household's own data shows the pairs are there to
be found: every candidate pair in the ledger sits within two days of its other
half.

## What Changes

- wimm recognises a **transfer between a member's own accounts**: a booked row
  going out of one account the member owns, paired with a booked row coming in
  to another account they own, in the same currency, for the same amount, a few
  days apart at most. Each row pairs at most once, the closest date wins, the
  other account's name or number in the row's text breaks a tie, and where a
  tie remains nothing is paired. A transfer wimm cannot pair stays an ordinary
  payment.
- A pair is found across every account the member owns. It is **left out of
  what is counted only where both of its rows are inside the scope in force**
  (Household, Yours, All). Money sent from a member's own account to the joint
  account is still money out under Yours and money in under Household, and
  cancels under All.
- Left out means left out of money in, money out and net in the month summary
  and in every month, the typical and average month, the merchants behind a
  rise, top merchants and largest payments, and never called an unusual or a
  recurring payment.
- **Nothing is hidden.** Transactions labels both rows `Between your accounts`,
  the balance chart's pointed day lists the row with that label, and a month
  says how many transfers it left out and for how much.
- The sentence `Money moved between your own accounts is counted.` is
  replaced, in the month summary and under the months, by one that says paired
  transfers are left out and others are counted.
- **BREAKING** for the figures a member has already read: money in and money
  out fall in every month that held a paired transfer. Net is unchanged under
  All, and changes under Household and Yours only where a pair sat wholly
  inside the scope.
- The marks are derived on read and never stored, as ADR 0025's are.
- A decision record, because this reverses a non-goal `spending-insights`
  stated in its design.

Builds on `spending-insights`, which must be finished and archived first: the
deltas below are written against the specs as that change leaves them.

## Capabilities

### New Capabilities

- `banking/own-transfers`: what wimm calls a transfer between a member's own
  accounts, and when such a transfer is left out of what is counted.

### Modified Capabilities

- `banking/overview`: the month summary and the months stop counting paired
  transfers inside the scope and say so; a month says how many it left out; a
  transfer listed on Overview is labelled.
- `banking/transactions`: a row that is one half of a paired transfer is
  labelled in the ledger.

## Impact

- `apps/wimm/internal/banking`: a new pure matcher beside `unusual.go` and
  `recurring.go`; `month.go`, `history.go`, `trend.go` and `ledger.go` run it
  before anything is summed or judged.
- `apps/wimm/internal/store`: the month summary and the months sum in Go from
  `OwnedBooked`, read across every owned account; `OwnedWindowSums` and
  `OwnedMonthlySums` lose their callers and go.
- `packages/contracts`: `Transaction.own_transfer`, `DayMover.own_transfer`,
  and a count and a total of transfers left out on `HistoryMonth`. No new RPC.
- `apps/web`: the two sentences, the month's left-out line, and the label on
  ledger rows and chart movers.
- `packages/ui`: `LedgerRow`, `BalanceChart`, `MonthDetail`, `Overview` and
  `TransactionsScreen` gain the label and the line; no new component.
- No migration, no bank read, no new dependency.
- A new ADR, and ADR 0025 gains nothing: its rules are unchanged, they are
  given fewer rows.
