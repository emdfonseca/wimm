## Context

See `proposal.md` — Why.

### What was checked, and what was not

**Verified against the API reference.** The read is
`GET /accounts/{account_id}/transactions`, taking `date_from`, `date_to`,
`continuation_key`, `transaction_status` and `strategy`. A response carries a
`continuation_key` when more pages exist. Each transaction carries
`entry_reference`, `transaction_id`, `transaction_amount`,
`credit_debit_indicator`, `status`, `booking_date`, `value_date`,
`transaction_date`, `creditor`, `debtor`, `creditor_account`, `debtor_account`,
`remittance_information`, `bank_transaction_code`, `merchant_category_code`,
`balance_after_transaction`, `reference_number`, `exchange_rate` and `note`.
`TransactionStatus` is `BOOK`, `PEND` or `OTHR` — **`PEND`, not `PDNG`**.
`entry_reference` is optional in the schema.

**Verified against the changelog.** `strategy=default`, which is also the
behaviour when the parameter is omitted, returns `WRONG_TRANSACTIONS_PERIOD`
when the requested period is unavailable, and is recommended for updating a
feed already fetched. `strategy=longest` asks the API to find the earliest
available transaction and fetch everything after it; `date_to` is ignored and
`date_from` becomes a hint for where to start searching. This is a near-exact
match for the two reads this change needs, and it is why they are two calls
rather than one with a wider window.

**Checked and absent.** There is no ASPSP-level field describing how far back
transactions reach — nothing equivalent to `maximum_consent_validity` for
history. **So wimm cannot state in advance how much history a bank will give.**
It asks for the longest available and records what came back; the screen says
the date it actually reaches rather than a promise made before the read.

**Not verified, and priced as unknown.** Whether `strategy=longest` is
supported at Revolut, Montepio and ActivoBank specifically; the real page size
and page count of a first fill; and whether a bank amends a booked transaction
after returning it. The first two are answered by the sandbox in the first task
and change no spec. The third is handled by the identity rule below either way.

### What the repo already settles

**The consent asked for balances and nothing else.**
`enablebanking/client.go:288` sends `"balances": true`, with a comment naming
this change as the reason. `access` scopes accounts, balances and transactions
separately, so there is no reading of an existing grant that produces
transactions.

**Disconnection already does not rely on the cascade.**
`store/banking.go:374` runs an explicit `delete from accounts where
connection_id = $1`, and `visibleAccountsQuery` at `banking_accounts.go:177`
already excludes accounts whose connection is disconnected. So Overview and
every total keep their current behaviour with no change at all: removing the
delete changes what is stored, not what is shown.

**One consent lasts one day at one of this household's banks.** ADR 0018 and
the archived design record `maximum_consent_validity` of 90 days at Revolut and
Montepio and **1 day at ActivoBank**. For balances that made "access has run
out" a daily event. For transactions it also makes the *first fill* a daily
risk: a connection that dies every night must not re-fetch its whole history
every morning. This is the single strongest argument for storing the ledger
against the account rather than the connection.

**Rate limits do not bite when the member is present.** The roughly four-a-day
ASPSP cap applies to fetching with nobody there. Every sync in this change is
triggered by a member on a page, with PSU headers set, which is the same
position balances already took.

**The failure taxonomy exists.** `internal/banking/errors.go` already carries
bank unavailable, gateway unavailable, consent declined, consent expired, no
accounts and rate limited. This change adds no member-facing member to it.

## Language

Binding on spec text, interface copy and code identifiers.

- **Transaction** — one entry on one account. Never "entry", "movement",
  "item" or "payment".
- **Booked** — a transaction the bank has settled. wimm's word for `BOOK`.
  Never "cleared" or "posted".
- **Pending** — a transaction the bank has not settled. wimm's word for `PEND`.
  Never "uncleared" or "unsettled" in code; *unsettled* may appear in interface
  copy where "pending" would read as "waiting for wimm".
- **`OTHR`** — mapped to **booked**, because a transaction wimm cannot classify
  has already moved money and hiding it is worse than filing it. Stated here
  because it is a silent judgement otherwise.
- **Sync** — one fetch of transactions for one account. The counterpart of
  **reading**, which stays the noun for one fetch of balances. Both are
  **Refresh** to a member: one verb, one meaning — ask the bank again.
- **Synced through** — the date up to which an account's transactions are
  believed complete. It is wimm's own fact, written after a sync finishes.
- **Continuation key** — the gateway's handle for the next page *within* one
  sync. Wire vocabulary; it stays inside the adapter and is never persisted.
  It is not the synced-through date and the two are never called the same thing.
- **Scope** — what a connection's consent covers: `balances`, or
  `balances_and_transactions`. A property of the connection, separate from its
  state. A connection that is live with the narrow scope is described to a
  member as connected *before wimm could read transactions*, never as broken.
- **Re-attach** — matching an account that outlived its connection to a new
  connection at the same bank, on its gateway reference. Never "merge",
  "import" or "restore" — **restore** is taken, and means confirming again at
  the bank after access ran out.
- **Joint account** — an account more than one member owns. It is joint
  because `account_owners` holds two rows, and for no other reason. Never a
  column, never a flag, never a word the bank supplies. **Not** the same fact
  as `holder_name`, which is what the bank has on the account and which wimm
  copies and never edits.
- **Page** — one read of the ledger, bounded by a cursor rather than an offset.
  A member is never shown a page number: the position is a span of dates.
  "Older" and "Newer" are the words on the controls, never "Previous" and
  "Next", which say nothing about which way time runs.
- **Ledger** — prose only. It is not a code identifier and not an interface
  label; the destination is **Transactions**, because the product itself is a
  ledger and the word cannot mean both.

## Goals / Non-Goals

**Goals:**

- A member sees the transactions of accounts they own, current enough to act on
  and honest about how current that is.
- A history that survives a restore, a disconnection and a reconnection, because
  no bank hands it back.
- A first fill that happens once per account, not once per consent, at a bank
  whose consent lives one day.
- Identity rules that produce two rows for two identical coffees and one row for
  one transaction seen twice.

**Non-Goals:**

- **No duplicate detection.** `surfaces.md` describes a row-level "looks like a
  duplicate" alert; that is for a later change where transactions arrive from
  more than one source. Here every row comes from one bank, and the occurrence
  index below is what keeps two identical coffees from collapsing into one.
- **No reconciliation of amended transactions.** See Risks.
- No syncing with nobody present, which is also what keeps the ASPSP cap out of
  scope.
- No transaction is editable, annotatable, splittable or deletable.

## Decisions

### 1. The ledger is stored, and Overview's freshness promise does not carry

A balance is one number and is re-read on arrival before the screen renders.
Transactions cannot be: the first fill is unbounded pages across every account,
and even an incremental sync is one round trip per account. So the Transactions
screen renders **stored rows immediately**, states when it was last synced, and
runs the sync behind that arrival, updating when it lands.

This is a weaker promise than `banking/household-accounts` makes about
balances, and the weakening is the decision. It is made survivable by saying it
on the screen rather than implying it: the list carries its synced-through date
the way a balance carries its read time. ADR 0018's rule — a stale figure
presented as live is worse than no figure — is satisfied by the statement, not
by the freshness.

*Alternative:* block the render on the sync, as Overview does. Rejected: at
ActivoBank, where consent dies nightly, that is a first-fill-shaped wait on a
routine morning visit.

### 2. Two reads, because the gateway has two strategies

```text
first fill    strategy=longest, no date_from
              -> the earliest the bank will give, forward
incremental   strategy=default, date_from = synced through - overlap
              -> what changed since
```

The overlap exists because a bank can book a transaction with a booking date
earlier than the day wimm last synced. Re-reading a few days and relying on the
identity rule to discard what is already held is cheaper and more correct than
trusting a watermark. The overlap is configuration with a default of seven days.

`WRONG_TRANSACTIONS_PERIOD` on an incremental sync is retried **once** with
`strategy=longest` inside the adapter, and only a second failure surfaces. A
member cannot act on "the window you asked for is unavailable", so it is not a
member-facing failure and the taxonomy does not grow.

### 3. Pending is a replaceable set, booked is append-only

Each sync deletes an account's pending rows and writes the pending transactions
the bank just returned. Booked rows are only ever inserted or left alone.

This deletes the hardest problem in bank data rather than solving it. A pending
transaction becomes a booked one under a different `entry_reference`, a
different amount — the tip lands later — and a different date. Matching the two
is guesswork that is wrong in exactly the cases a household argues about. Under
this rule the pending row simply ceases to exist at the next sync and the
booked row arrives on its own.

The visible cost is that a pending row has no stable identity across syncs, so
nothing can be attached to one. Nothing in this change attaches anything to a
transaction, and when something does, it attaches to booked rows.

### 4. Booked identity is the bank's reference, then a digest and an occurrence

```text
dedup_key = entry_reference                     where the bank gives one
          = hex(sha256(booking_date, amount,    where it does not
                       currency, counterparty,
                       remittance))
occurrence = 1, 2, 3 … for rows sharing a key within one account
```

Unique on `(account_id, dedup_key, occurrence)` among booked rows. Two
identical coffees on the same day at the same shop are two real transactions and
get occurrence 1 and 2; the same transaction seen in two overlapping syncs
matches occurrence 1 and is not written twice.

The occurrence index is assigned by counting how many rows with that key a sync
returned against how many are already stored, not by a global counter — so a
sync returning two and finding one stored inserts exactly one.

*Alternative:* trust `transaction_id`, which the API also returns. Rejected
without evidence it is stable across sessions; `entry_reference` is the field
the standard defines for this and the one the account-identity decision's
reasoning already applies to.

### 5. Scope is a column, and a narrow connection is not a broken one

`bank_connections.scope` defaults to `balances`, which is true of every row that
exists. New connections write `balances_and_transactions`. Widening is the
restore flow with a third reason, which already re-enters at the hand-off,
skips the picker and carries owners forward.

Text with a check rather than a Postgres enum, for the reason `level` already
is: a later value must not need `ALTER TYPE … ADD VALUE`, which cannot run
inside a transaction.

The state is surfaced as an **inline alert on that bank**, never a page banner.
This is the ActivoBank lesson applied before it bites: a household with one
narrow bank and two wide ones must not see a page-level warning, or the warning
is furniture within a week.

### 6. An account outlives its connection, and its bank moves onto it

The explicit delete in `DisconnectBankConnection` is replaced by nulling each
account's sealed per-session identifier, so access ends exactly as ADR 0018
requires while the record stays. Both existing check constraints already permit
this: `accounts_gateway_columns_match_source` asks for `connection_id` and
`gateway_ref`, and both survive.

**`bank_id` moves onto `accounts`.** Which bank an account is at is a permanent
fact, and the account is now the permanent thing. It also makes the real
invariant expressible:

```sql
create unique index accounts_bank_gateway_ref_key
    on accounts (bank_id, gateway_ref)
    where gateway_ref is not null;
```

That index is what guarantees there is never a second copy of an account, which
is the promise `bank-connections` already makes and which was free only while
disconnection deleted. Re-attach then has nothing to disambiguate: there is at
most one row to find. The bank's display name still comes from the connection
row, which is soft-deleted and survives.

*Alternative:* enforce single-copy inside the re-attach transaction. Rejected —
this repo consistently prefers a state that cannot be represented over a state
that is checked, and a missed re-attach would present a member with two of
their own accounts, one holding the history.

### 7. Transactions are owner-only

An account's transactions are visible to its owners and to nobody else. A
member granted *balance* or *details* sees no transactions and no count of
them.

This leaves ADR 0019's fourth level exactly where it parked it. That ADR's
argument is the whole reason: a member who granted *details* before this change
consented to showing an account number, not to showing their spending, and
those are different sentences. Riding transactions on *details* would widen a
grant that a person already made, silently, in a deploy.

The read is the existing one narrowed — `account_owners` only, where the
visible-accounts query joins owners and grants.

### 8. Paging is keyset, so the ground does not move

A sync runs on every arrival and every refresh, and it inserts at the newest
end. With offset paging, a member reading page seven when twelve transactions
arrive has every row shift by twelve: they re-read some and skip others, and
nothing tells them it happened. That is not an edge case here, it is what every
visit does.

So the ledger pages by seeking on the key it is already indexed by:

```text
order      (booking_date desc, id desc)
index      transactions_account_date_idx
older      where (booking_date, id) < (cursor_date, cursor_id)  limit N
newer      where (booking_date, id) > (cursor_date, cursor_id)  limit N, reversed
route      /transactions?before=<cursor>  and  ?after=<cursor>
```

The cost is the thing the library's `Pagination` molecule is built to show:
there is no "page 3 of 65" and no "41 to 60 of 1,284", because a seek does not
know either number without a second count that would be wrong by the time it
rendered. What replaces it is better for a ledger anyway — the span of dates on
screen, which is what a person scanning backwards actually wants to know.

The total count stays in the toolbar. It is a count of what the member may see,
not a page count, and it survives paging unchanged.

`N` defaults to 50 and is configuration. At comfortable density roughly twelve
rows are visible and at compact roughly seventeen, so a page is three to four
screens either way, which is the number ADR 0002 sized the density axis on.

**A day is named again when it straddles a boundary.** Grouping is presentation
over a flat ordered list, so a page simply starts mid-day and repeats the
heading. The alternative, paging by whole days, makes page size unpredictable
and a single busy day unbounded.

*Alternative:* offset paging with numbered pages, which the library already
draws. Rejected for the drift above. *Alternative:* infinite scroll. Rejected
twice over: it cannot be deep-linked, which `surfaces.md` requires of a full
page, and it makes returning to a spot a matter of luck.

### 9. Ownership is the household's record, and the bank's holder name is not it

`account_owners` is the only thing that makes an account joint. `holder_name`
is what the bank said, stored as given and never edited.

They are allowed to disagree, and the temptation to reconcile them is the thing
this decision refuses. Matching a holder name against member names to add an
owner would be a guess that silently widens who sees an account, which is the
exact failure ADR 0019 built the ownership model to prevent. A household may
also own an account jointly in wimm that one of them holds alone at the bank,
because that is how they treat it, and wimm has no standing to correct them.

Nothing in this change infers, suggests or defaults an owner from a name.

### 10. Schema

```sql
alter table bank_connections
    add column scope text not null default 'balances'
        check (scope in ('balances', 'balances_and_transactions'));

alter table accounts
    add column bank_id                    text,
    add column transactions_synced_through date,
    add column transactions_synced_at      timestamptz;
-- backfilled from the connection, then set not null for gateway accounts

create table transactions (
    id                uuid        primary key default uuidv7(),
    -- On the account, never on the connection: an account survives a restore,
    -- a disconnection and a reconnection, and its history survives with it.
    account_id        uuid        not null references accounts (id) on delete cascade,
    status            text        not null check (status in ('booked', 'pending')),
    -- The bank's own reference where it gives one, else a digest of the
    -- fields that identify the transaction. Never null: the fallback always
    -- produces a value.
    dedup_key         text        not null,
    -- Which of several rows sharing a key this is, so two identical coffees
    -- are two transactions.
    occurrence        int         not null default 1 check (occurrence >= 1),
    -- Signed minor units. A debit is negative, so direction never depends on
    -- a separate column agreeing with the sign.
    amount_minor      bigint      not null,
    currency          text        not null check (length(currency) = 3),
    booking_date      date,
    value_date        date,
    transaction_date  date,
    counterparty_name text,
    remittance        text,
    first_seen_at     timestamptz not null default now(),
    last_seen_at      timestamptz not null default now()
);

create unique index transactions_booked_key
    on transactions (account_id, dedup_key, occurrence)
    where status = 'booked';

-- The screen's only ordering, and the account filter's index.
create index transactions_account_date_idx
    on transactions (account_id, booking_date desc, id desc);
```

`on delete cascade` from `accounts` is correct and is not the cascade this
change breaks: an account is now never deleted, so the cascade never fires. It
stays so that the day something does delete an account, its history does not
outlive it as orphans.

### 11. The port grows one method

```go
Transactions(ctx context.Context, conn Connection, account Account,
    req TransactionsRequest) (TransactionsPage, error)

type TransactionsRequest struct {
    // From is where to start. A zero From means "as far back as this bank
    // goes", which the adapter expresses as strategy=longest.
    From time.Time
    // Cursor is the gateway's own page handle, opaque above the adapter and
    // never stored.
    Cursor string
}

type TransactionsPage struct {
    Transactions []Transaction
    NextCursor   string // empty when the last page has been returned
}
```

A zero `From` meaning "the longest this bank offers" keeps the second gateway
honest: GoCardless has no strategy parameter and a `max_historical_days` per
institution, so its adapter clamps a zero `From` to that. Neither gateway's
shape reaches the caller, which is the test ADR 0018 set for this port.

`bankingtest` implements it in the same task, not a later one.

### 12. Compact navigation is a bottom bar, not a hamburger

The library holds a half-answer inherited from a generic build rather than a
decision: `App header` carries a `Nav trigger` that is disabled everywhere and
that `App header/compact` switches on, and nothing was ever drawn for what it
opens. The shell this product actually uses, `App bar / compact`, has no
trigger at all. So there is no precedent to follow, only a default to reject.

**The switch is the product.** Overview answers how much and Transactions
answers what happened, and they are two halves of one question: a member who
reads a balance and wants to know why moves between them immediately and
repeatedly. Navigation that is used constantly and bidirectionally has to be
visible, and a hamburger puts a tap and a mental model in front of two items.

**ADR 0005 already settled the analogous question and its reasoning carries.**
It chose the form of secondary navigation from the content and said so
explicitly: a fixed handful of peer views gets tabs *at every size*, a growing
set gets a list card *at every size*, and "width plays no part in the choice".
A hamburger on small screens and a sidebar on large ones is exactly the
width-decides answer that ADR rejected. A fixed handful of top-level areas is
persistent at every size: a sidebar where there is a column for one, a bottom
bar where there is not.

**It fits the set this product is heading for.** The library's sidebar draws
Overview, Accounts, Transactions, Budgets, Reports and Settings. Accounts is
already redundant, because Overview is the accounts screen. Reports is
low-frequency and Settings belongs behind the account avatar, which the compact
app bar already carries. That leaves Overview, Transactions, Budgets and
Reports: four, under the five a bottom bar holds before it degrades.

**What it costs.** 60 px, against a 64 px row. Slightly less than one
transaction out of roughly twelve on screen, for a switch that is one tap from
the easiest place on the screen to reach rather than two taps from the hardest.

The current item is marked by an accent top edge, an accent icon and an accent
label at 600. The edge matters: weight and colour alone would put the whole
signal in hue, and the sidebar's accent-subtle fill does not survive the change
of scale — half a phone screen filled with accent is a different object from a
240 px pill inside a 264 px column.

*Alternative:* a segmented control under the page header. Rejected twice: it is
already spoken for by `surfaces.md` as peer views of one entity, which these
are not, and it does not survive a third destination. *Alternative:* the
hamburger the library left switched on. Rejected as above.

### 13. Surfaces

- **The sidebar already exists and already takes a list.** `SidebarNav` renders
  the destinations it is given and drops any without an address, which is why a
  shell with one destination has read as a shell with no navigation. Adding
  Transactions is a second entry in that list. Nothing is built.
- **`/transactions`** is the home; `?account=<id>` is the same page filtered
  with a removable chip and no breadcrumb, per `surfaces.md`.
- **Amounts carry a sign** and are never distinguished by colour alone, per
  ADR 0006.
- **Pending is a badge on the row.** `surfaces.md` already rules this.
- **A narrow-scope bank is an inline alert on that bank**, per decision 5.
- **The list header** carries the synced-through date and Refresh. A refusal
  leaves the rows and their date exactly as they are, the shape balances use.
- **The ledger footer carries the paging.** It is a configuration of the
  library's `Pagination` molecule, not a new component: the range text becomes
  a span of dates and the page-number text is switched off.

## Risks / Trade-offs

**A bank amends a booked transaction after returning it** → Where the bank
gives an `entry_reference` the row is found and updated, which is most banks.
Where wimm fell back to a digest, an amended amount or description produces a
new key: the amended transaction is inserted and the superseded one lingers.
Accepted and not mitigated in this change — reconciling it needs a rule for
deciding which of two rows is the real one, which is the duplicate machinery
this change explicitly excludes. `last_seen_at` is written on every sync so a
later change can find rows the bank stopped returning.

**The first fill is unbounded** → `strategy=longest` can return years across
several pages, on the first visit after a member widens a consent. Bounded two
ways: paging stops at a configured maximum number of pages per account per
sync, and `transactions_synced_through` records how far back the fill actually
reached, so a partial fill is a stated fact and the next sync continues rather
than restarting.

**A member with five accounts triggers five syncs on arrival** → The page does
not wait for them. A per-account minimum interval between syncs means a member
reloading repeatedly does not multiply the calls, and the ASPSP cap stays out
of reach because every one of these carries PSU-present headers.

**Every existing connection needs a member to act** → Nothing breaks if they
never do: balances keep working and that bank's transactions stay empty, with
the reason on screen. The cost is a release that is invisible until acted upon,
which is why the alert sits on the bank rather than in a settings page nobody
opens.

**Nothing deletes an account or its history any more** → A household that
connects and disconnects repeatedly accumulates rows it cannot remove. Named in
the proposal, not solved here. It is the direct price of the guarantee this
change is buying and should be paid by a deliberate delete-with-confirmation in
a later change, not by weakening this one.

**Owner-only is not the final answer** → A household where one partner owns the
joint account and the other was granted *details* will want the other to see
the spending. That is the fourth level, and it is deliberately not here. The
risk is that it reads as an omission rather than a decision, so the screen says
what a member sees and why rather than showing an empty list.

## Migration Plan

1. `00005_transactions.sql`: the transactions table and its indexes,
   `bank_connections.scope` defaulting to `balances`, and the three new
   `accounts` columns.
2. Backfill `accounts.bank_id` from each account's connection in the same
   migration, then constrain it for gateway-sourced accounts.
3. Deploy. Every existing connection reads as `balances` and is correct:
   Overview is unchanged, and Transactions shows each bank's state with the way
   to widen it.
4. New connections request both scopes from the first deploy, so a household
   connecting its first bank afterwards never sees the narrow state.

**Rollback.** The schema is additive and safe to leave in place while the code
rolls back; the only behavioural change to existing code is the removal of the
delete in `DisconnectBankConnection`, and an older binary rolling forward over
surviving accounts re-deletes them. Rolling the migration back drops the
transactions table and the history in it, so the code rollback is the one to
reach for.

## Open Questions

- The default page cap for a first fill, and the default minimum interval
  between syncs of one account. Both are configuration with a default, both are
  answered by watching a real first fill, and neither changes a spec.
- Whether `strategy=longest` is honoured at Revolut, Montepio and ActivoBank.
  The sandbox answers this in the first task; if one bank refuses it, that
  adapter path falls back to a dated window and the screen still states what it
  reached.
- What the Transactions screen does at the `ultra` regime. ADR 0002 gave it a
  persistent inspector pane, the library carries both the pane and a
  `Transactions / Ultra 1920` template that places it, and this change has
  nothing to put in it. The canvas decides whether `ultra` simply widens the
  list; instancing a pane with no content would draw a product nobody agreed to
  build.
