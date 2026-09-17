# 0021 · A stored ledger, and the consent that fills it

## Status

Accepted

Extends the gateway port ADR 0018 defined with one method. The port's shape,
account identity on the gateway's cross-session hash, sealing values when access
ends, money as `int64` minor units and the failure taxonomy are unchanged.
ADR 0019's owners and levels are unchanged, and its fourth level stays unbuilt.

## Context

wimm reads a balance. A balance is a position, and the question a household
argues about is what happened, which lives in the transactions behind it.

Four facts shape how that ledger is held. Transactions are a separate scope at
the bank, so no reading of a grant wimm already holds produces them. A sync
inserts at the newest end and runs on every arrival. A transaction the bank has
not settled changes its reference, its amount and its date when it settles. And
no bank hands history back indefinitely — one of this household's banks grants
access for a single day — so anything wimm read once and dropped is gone.

## Decision

**The consent widens, and every existing connection needs a member to act.**
`bank_connections.scope` is `balances` or `balances_and_transactions`, text with
a check rather than a Postgres enum so a later value needs no `ALTER TYPE`.
Every row that exists reads as `balances` and is correct. New connections ask
the bank for both. Widening is the restore flow with a third reason — the
connection is live and its consent is narrow, which is neither working nor
expired — re-entering at the hand-off, skipping the picker and carrying owners
forward. A narrow connection is described as connected before wimm could read
transactions, never as broken, and the state is an inline alert on that bank
rather than a page banner: a household with one narrow bank and two wide ones
must not be warned about the whole product.

**The ledger is keyset-paged, because a sync inserts at the newest end.** The
list orders by `(booking_date desc, id desc)` and seeks on that key in both
directions; the routes are `?before=<cursor>` and `?after=<cursor>`. With offset
paging, a member reading a page while twelve transactions arrive re-reads some
rows and skips others with nothing telling them. That is not an edge case here,
it is what every visit does. The price is that there is no page number and no
"41 to 60 of 1,284", because a seek knows neither without a count that would be
wrong by the time it rendered. What replaces it is the span of dates on screen,
which is what a person scanning backwards wants. A day that straddles a page
boundary is named again at the top of the next page.

**Pending is a replaceable set; booked is append-only.** Each sync deletes an
account's pending rows and writes the ones the bank just returned; booked rows
are only inserted or left alone. This deletes the hardest problem in bank data
rather than solving it — matching a pending row to the booked row it becomes is
guesswork that is wrong in exactly the cases a household argues about. A booked
row is identified by the bank's `entry_reference`, or by a digest over booking
date, amount, currency, counterparty and remittance where the bank gives none,
plus an occurrence index assigned by counting how many rows with that key a sync
returned against how many are stored. Two identical coffees are two rows; one
transaction read by two overlapping syncs is one.

**An account outlives its connection.** Disconnecting nulls each account's
sealed per-session identifier instead of deleting the row, so access ends
exactly as ADR 0018 requires while the record stays. `bank_id` moves onto
`accounts`, because which bank an account is at is permanent and the account is
now the permanent thing, and a unique index on `(bank_id, gateway_ref)` where
`gateway_ref` is not null makes a second copy of an account unrepresentable
rather than merely checked. Re-attach then has nothing to disambiguate: there is
at most one row to find. Ending wimm's access to a bank and destroying the record
of what it read are different decisions, and disconnection makes only the first.

**Transactions are owner-only.** An account's transactions are visible to its
owners and to nobody else; a member granted *balance* or *details* sees none and
is not told how many there are. Riding transactions on *details* would widen a
grant a person already made, silently, in a deploy.

**The screen states its own freshness rather than waiting for it.** Stored rows
render immediately with the date they are synced through, and the sync runs
behind the arrival. This is a weaker promise than the one balances make, and the
weakening is survivable only because it is written on the screen: ADR 0018's rule
that a stale figure presented as live is worse than no figure is satisfied by the
statement, not by the freshness.

## Consequences

Nothing deletes an account or its history any more, so a household that connects
and disconnects repeatedly accumulates rows it cannot remove. This is the direct
price of the guarantee and is paid by a deliberate delete-with-confirmation in a
later change, not by weakening this one.

Where wimm fell back to a digest, a bank amending a booked transaction produces a
new key: the amended row is inserted and the superseded one lingers. `last_seen_at`
is written on every sync so a later change can find rows the bank stopped
returning.

A pending row has no stable identity across syncs, so nothing can be attached to
one. When something attaches to a transaction, it attaches to booked rows.

The rollback to reach for is the code, not the migration. The schema is additive
and safe to leave in place; rolling it back drops the transactions table and the
history in it, and an older binary rolling forward over surviving accounts
re-deletes them.

`Gateway` grows `Transactions`, carrying a `From` whose zero value means "as far
back as this bank goes" and an opaque cursor that is never persisted. Enable
Banking expresses the zero value as `strategy=longest`; GoCardless would clamp it
to that institution's `max_historical_days`. Neither gateway's shape reaches the
caller, which is the test ADR 0018 set for this port. `bankingtest` implements
the method beside the real adapter, not later.
