## Why

Overview answers how much the household has. It cannot answer why that number
is lower than it was on Friday, because wimm reads a balance and nothing else.
A balance is a position; the question a household actually argues about is what
happened, and that lives in the transactions behind it.

Everything this needs is already built except the read itself. The gateway
port, the ownership model, the levels, the freshness discipline and the restore
flow all exist, and ADR 0018 shaped `Gateway` against a second provider with
this read explicitly in view. What is missing is a ledger, the consent to fill
it, and somewhere to put it.

## What Changes

- **Transactions become a destination.** `/transactions` lists what a member
  may see, newest first, grouped by date. An account's transactions are that
  same page filtered, per `docs/design/surfaces.md` — one home, never a second
  copy with its own hierarchy.
- **The ledger is read a page at a time, and the page holds still.** Ninety
  days across several accounts is always more than one screen. Paging seeks on
  the date rather than counting an offset, because a sync runs on every arrival
  and inserts at the newest end: with offsets, arriving transactions shift every
  page under the member and they silently re-read some rows and skip others.
  The member is shown the span of dates they are looking at, never a page
  number.
- **A joint account is one that two members own, and nothing else makes it
  joint.** Ownership is the household's own record. wimm never infers an owner
  from the name the bank has on the account, and never edits that name to agree
  with who owns it: the two are allowed to disagree and the disagreement is not
  an error.
- **BREAKING: the consent a member grants widens.** Every connection wimm holds
  today asked its bank for `balances` and nothing else, and the hand-off screen
  promises in as many words that wimm will read account names and balances "and
  not its transactions". Transactions are a separate scope at the bank, so
  there is no migration that produces them — only a member confirming again.
  New connections ask for both and the hand-off copy changes to say so.
- **Existing connections upgrade through the restore flow**, which already
  re-enters at the hand-off, skips the picker and carries owners forward. It
  gains a reason it did not have: the connection is live and its consent is
  narrow, which is neither working nor expired. Waiting for consent to lapse
  would also work — every grant dies inside 90 days — and would make a member
  who connected yesterday wait a quarter.
- **Transactions are seen by an account's owners and by nobody else.** A member
  granted *balance* or *details* sees no transactions at all. This leaves
  ADR 0019's fourth level exactly where it parked it, one `ALTER TABLE` away,
  for the day someone asks to share spending rather than identity.
- **A ledger is stored, which a balance never was.** A balance is one number
  re-read on arrival; ninety days across five accounts is not a page load. So
  stored rows render immediately with the time they were last brought up to
  date, and the sync runs incrementally behind that arrival. This is a weaker
  freshness promise than Overview's and it is stated on the screen rather than
  implied, because ADR 0018's rule — a stale figure presented as live is worse
  than no figure — is the one being bent.
- **Pending transactions are a replaceable set; booked ones accumulate.** Each
  sync discards an account's pending rows and rewrites them, and only booked
  rows are appended. A pending row becoming a booked row with a different
  identifier, amount and date is the hardest correctness problem in bank data,
  and this removes it rather than solving it.
- **BREAKING: disconnecting a bank no longer erases its accounts.** Today
  disconnection deletes them outright. A ledger that a bank will not hand back
  past ninety days cannot be destroyed by a button whose whole purpose is to
  stop wimm reading. Accounts and their transactions survive; Overview and
  every total exclude them exactly as they do now, by the read filter that is
  already there.
- **Reconnecting the same bank re-attaches rather than duplicates.** This was
  free while disconnection deleted, and is now real work: accounts are matched
  on the gateway's cross-session hash across connections, not within one.
  Restore, reconnect and the consent upgrade become three doors into one rule —
  an account is identified by its bank and that hash, for as long as it exists.
- **Transactions becomes the second destination in a sidebar that already
  exists.** `SidebarNav` is built and `SignedInLanding` already renders it; it
  takes a list of destinations and drops any without an address, which is why
  a shell carrying one item has looked like a shell with no navigation. This
  change passes it a second entry rather than building anything.
- `Gateway` gains a transactions read carrying a cursor, implemented for Enable
  Banking and in `bankingtest` beside it.

Not in this change, and deliberately: categories, budgets, reports, search
beyond the account filter, jumping to a numbered page, editing or annotating a
transaction, splitting,
merging duplicates, manually entered transactions, currency conversion,
syncing with no member present, and any way to delete an account or its
history.

## Capabilities

### New Capabilities

- `banking/transactions`: the household's ledger. Covers who may see a
  transaction, what the list shows and how it is ordered and filtered, how far
  back history reaches, how the ledger is brought up to date and what it says
  about its own freshness, how pending and booked entries differ, what happens
  to duplicates wimm cannot resolve, and what the screen says before any
  transaction has ever been read.

### Modified Capabilities

- `banking/bank-connections`: what a member is told wimm will read, and the
  scope actually requested, now include transactions. A connection granted
  before this change gains a state of its own — live, but unable to read
  transactions — with the way to widen it. Restoring and reconnecting both
  match accounts on the gateway's cross-session hash, so an account that
  survived a disconnection is re-attached instead of created a second time.
- `banking/household-accounts`: an account stops appearing and stops counting
  when its bank is disconnected, as it does today, but is no longer erased.
  The requirement says so explicitly, because a reader who concludes deletion
  would find the reconnection behaviour contradicts it.

## Impact

**Schema.** A transactions table owned by `wimmd`, keyed on the account rather
than the connection so it survives both a restore and a disconnection, plus a
per-account sync cursor. The `accounts` row outlives its connection: the
explicit delete in `DisconnectBankConnection` goes, and the sealed per-session
identifier on each account is destroyed in its place, so access ends exactly as
ADR 0018 requires while the record stays.

**Contracts.** `Gateway` gains a method, which changes a port ADR 0018 defined
and therefore needs an ADR of its own. New Connect methods on the public
listener; the browser reaches them through SvelteKit server routes and never
directly, per ADR 0001.

**Rate limits.** The roughly four-a-day cap applies to fetching with nobody
present, and this change does none: every sync is triggered by a member on the
page, with PSU headers set. A refusal leaves the stored ledger on screen with
its original sync time, the same shape balances already use.

**Web.** New routes for the ledger and its account filter, and a new reason
threaded through the existing restore route. The signed-in shell gains a second
destination, which is a list entry rather than a component.

**Design system.** Less is new here than it first appeared. The library already
carries a `Transaction row` molecule, `Transactions` shell templates at every
regime, a sidebar, a page header and an inspector. What it does not carry is a
row shaped like *this* change: the drawn row has a selection control and a
category column, and no account, bank or unsettled marker. A row with fewer
columns is a finding for the canvas, not an assumption to make here. This is
the first screen the `density` axis was justified by — ADR 0002 sized it on
twelve comfortable rows against seventeen compact — and the first that could use
`ultra`'s inspector pane, though nothing in this change edits anything for it to
inspect.

**Operations.** A household that upgrades sees every existing bank asking to be
confirmed again. Nothing breaks if they never do: balances keep working and the
ledger stays empty for that bank, with the screen saying which bank and why.

**Two gaps this change knowingly leaves.** Nothing deletes an account or its
history any more, so a household that connects and disconnects repeatedly
accumulates rows it cannot remove. And Overview's empty state still invites a
household to connect its first bank while a ledger of past transactions sits
behind it.
