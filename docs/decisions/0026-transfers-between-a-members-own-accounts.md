# 0026 · Transfers between a member's own accounts

## Status

Accepted

## Context

A member who moves money from one account they own to another sees it twice in
wimm: once as money out and once as money in. Neither figure is wrong on its own
and the pair together is, because the household spent nothing and earned nothing.
On the household's ledger — 1,222 booked rows on six owned accounts — this
accounts for 14 movements, one of which is €4,234.18 and seven of which are a
standing monthly €1,867.76.

The rows carry almost nothing to match on. `store.Transaction` holds
`CounterpartyName` and `Remittance` and nothing else about the other party; no
counterparty account number or IBAN is stored, and ADR 0018 keeps the member's
own full account numbers out of the database. What an account has is `name`,
`household_name`, `number_suffix` and `holder_name`.

`spending-insights`' design named this a non-goal. This reverses that.

## Decision

**A transfer between a member's own accounts is two booked rows, in opposite
directions, of the same currency and exactly equal amounts, on two different
accounts the member owns, whose booking dates are at most
`ownTransferWindowDays` (3) apart in either order.** `banking.OwnTransfers` in
`apps/wimm/internal/banking/transfers.go` is a pure function over the rows
already in hand, like ADR 0025's rules, returning both directions of every pair.

The window is three because the ledger's own pairs end at two and a weekend adds
one; four and five find nothing more and double the room for a coincidence.

**Amounts are exact and same-currency.** A tolerance for a fee or a conversion is
where false pairs come from, and exactness is the only rule that needs no
constant.

**Candidates are ranked by date gap, then by evidence, and a tie is refused.**
Two rows pair when each is the other's single best-ranked unpaired candidate;
a row whose best rank is shared by two candidates is not paired in that round and
may be in a later one. Pairing greedily in date order was rejected: with two outs
and two ins of one amount it pairs the first with the first even when the second
is the same-day one, and the result then depends on row order.

**Evidence is a row's text naming the other row's account, and it only breaks a
tie.** A token is that account's `number_suffix` when it is all digits and at
least `ownTransferMinSuffix` (4) long, or its `name` or `household_name` when at
least `ownTransferMinNameRunes` (4) runes long — minus every value a second
account the member owns also carries. A shared value says "mine" and not
"which": three of the household's accounts repeat `holder_name` in `name`, and
the two Revolut accounts share all three fields because nothing distinguishes
them. `holder_name` is therefore never a token, and neither is a value that
behaves like one. A suffix must be digits because a length minimum was standing
in for that: the household's PayPal account has the suffix `.com`.

Requiring evidence for every pair was rejected: Portuguese banks write
`TRF P/ <name>` as often as a number, and two of the household's fourteen pairs —
both recognised — carry none.

**A pair is found across every account the member owns; leaving it out is decided
per scope.** Both halves inside the scope in force means the pair is left out of
that scope's figures; one half outside means the pair is **crossing** and is
counted. So money from a joint account to a personal one is money out under
Household and money in under Yours, and left out only under All, where both
accounts are the member's. Leaving a crossing pair out wherever either half is in
scope was rejected: an owner's Household would lose money the other owner's keeps,
and Yours would stop showing what leaves it every month, which is the most useful
thing Yours says. There is no fraction of an account to move — ADR 0019 makes
ownership whole — so a crossing pair is the only place the change of hands lands.

**A row is labelled whenever it is half of any pair, in every scope, and the
label takes the status slot from `Unusual`.** The wire carries
`Transaction.own_transfer`, `DayMover.own_transfer`, and
`HistoryMonth.transfers_left_out` with `transfers_total`. **No partner id is
sent**: the browser has no way to show the other row, and an id is one more thing
that could point at an account the reader may not own.

**Pairs are derived on read**, for ADR 0025's reason: the window and the evidence
rule will be tuned, a stored pair needs a backfill each time, and a late-arriving
row can turn an unpaired row into a pair.

**An unpaired row whose counterparty is the member's own `holder_name` stays
counted.** It is a likely transfer to an account wimm does not hold, and leaving
it out would change what money out means. On the household's ledger there is one,
€160.80.

## Consequences

Money in and money out fall in any scope that holds both halves of a pair; no net
figure moves, because both halves leave together. On the household's ledger
Household loses nothing, Yours loses four pairs in one month, and All loses all
fourteen. Figures a member has already read change downward with no action of
theirs, so the sentence under them changes in the same deploy and each month
states how many transfers it left out.

ADR 0025's detectors are handed the counted rows of the scope in force, so a
standing transfer is a recurring payment under Yours and nothing under All, and a
large move to savings is unusual in no scope that pairs it. Under Yours a crossing
pair can be both counted and large, so it may be unusual in the month's list while
its ledger row reads `Between your accounts`; the label wins the slot.

The figures are summed in Go from one row set, because SQL cannot leave out a row
whose partner it has not seen. `store.OwnedWindowSums` and `store.OwnedMonthlySums`
go, and `MonthSummary` reads `OwnedBooked` over a window widened by
`ownTransferWindowDays` at each end so a pair across the edge is still found.

A payment and an unrelated arrival of the same amount within three days can be
paired, and a real payment then vanishes from the figures. The closest-date and
ambiguity rules limit it, both rows are labelled where a member will see them, and
the month accounts for the difference. `ownTransferRequireEvidence` is the
fallback if a false pair appears.

A pair whose second half has not arrived is counted for a day or two and then left
out. Only the month so far can show it.

Nothing is stored and no schema moves, so rollback is the previous build and the
figures return to what they were.
