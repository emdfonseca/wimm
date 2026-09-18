## Why

Six findings came out of using wimm after `read-bank-transactions` shipped. Two
of them are defects in what the product promises; four are the app failing to be
a place a person can move around in.

The first is the serious one. Today a member can remove their own last ownership
of an account with one unconfirmed checkbox. The account then has no owner and no
grant, so it disappears from Overview, its transactions become unreadable by
everyone, and the bank card it sat under disappears with it when it was that
bank's last visible account — taking the "Who sees these" link, which is the only
route back, because that link is gated on owning something at the bank. There is
a reclaim path in Go, but it works only for the member recorded in
`connected_by`, only for an account that still has a connection, and only by
typing the URL from memory. `ADR 0019` designed the end state on purpose —
"an account that reaches the unowned-and-ungranted state is never read again …
goes dark and stays dark" — so this is not a bug against the decision. It is the
decision being wrong, and half of it has to be overturned.

The rest are the app. The ledger can only be walked one page at a time in two
directions. Account names are the bank's and cannot be changed. The connect
wizard puts its buttons somewhere different on each of its four steps. Layout is
worse than untidy: three mutually inconsistent breakpoint families are live at
once, every per-regime layout token is declared and consumed by nothing, and no
journey frame has ever been drawn at Medium or Ultra. And every navigation in the
signed-in app is a blocking server round-trip with no indicator at all, so a slow
one is indistinguishable from a click that did not land.

## What Changes

**An account always has an owner, and "not in wimm" becomes its own fact.**

- Ownership can never fall to zero. A database constraint trigger enforces it, so
  no route, handler or future caller can produce an orphan; the guard is not a
  check in Go that the next writer forgets.
- **BREAKING** (to a recorded decision, not to a released contract): an account
  with no owner and no grant is no longer representable. `ADR 0019`'s
  unowned-and-ungranted state is replaced by **left out** — a reversible fact on
  the account, separate from who owns it. Leaving an account out stops every read
  of it and hides it from everyone but its owners, who keep a row saying it is
  left out and a way to bring it back. The privacy affordance `ADR 0018` exists
  for survives intact, and it gains a route home.
- The chooser stops being able to orphan. Its "mine" checkbox becomes an owner
  set that names every member of the household, so handing an account to the
  person it belongs to is possible in the app for the first time — the spec has
  required it since `connect-bank-accounts` and no control ever existed. Ticking
  it no longer silently drops a joint account's other owner.
- Backfill: every account that is already orphaned is brought back as left out,
  owned by the member who connected its bank.

**The ledger can be moved around in.**

- Jump to the newest page and jump to the oldest page, both native to a keyset
  seek.
- A month scrubber: the months that hold transactions, from one `date_trunc`
  aggregate over the index that already exists, each one seeking to that month's
  newest row. Page numbers are still refused, for the reasons `ADR 0021` gives.
- The pager stops disappearing when everything fits on one page; it keeps the
  span of dates, which is the thing that says where you are.

**Accounts can be named by the household.** A name the household sets sits in its
own column, which the gateway upsert does not touch, so a reconnect or a restore
no longer overwrites it. The bank's own name is kept and shown beneath.

**The connect wizard reads as one flow.** One action placement across all four
steps, one action pattern, a step indicator, and step copy rewritten so the four
screens are four steps of one sentence rather than four unrelated pages.

**One layout system, four regimes, every screen.** One breakpoint family
(768 / 1200 / 1800, the declared device axis); the 1024 and 599 families are
deleted. A `Page` template replaces six hand-rolled copies of the same container
rule, four of which forget the content cap. Every per-regime layout token is
consumed by something. Medium and Ultra frames are drawn for every journey
screen, which is where the pass starts — no screen is changed before its frame
says what it should be.

**The app says when it is working.** One global navigation indicator in the root
layout, driven by SvelteKit's own pending-navigation state, announced rather than
only drawn.

**A place to keep preferences, and density reaches the screen.** A Settings
screen with a subsection list card per `ADR 0005`, carrying appearance (moved off
the bottom of the signed-in layout, where it currently sits after the page
content) and a density control. `data-density` is written nowhere today, so the
whole density axis has been unreachable since it was specified; the control is
hidden where the pointer is coarse, per `ADR 0004`.

## Capabilities

### New Capabilities

- `app/screen-layout`: every screen fits the window it is given, at all four
  device regimes, and the same regime means the same thing everywhere.
- `app/navigation-feedback`: a member can always tell that the app has taken
  their navigation and is working on it.
- `app/settings`: a place a member sets how wimm looks — appearance and density —
  and the rule that a control which cannot apply is not shown.

### Modified Capabilities

- `banking/household-accounts`: an account always has at least one owner; **left
  out** replaces unowned-and-ungranted as the state that stops reads; a household
  name for an account, kept across reconnects; what a total and a list include
  when an account is left out.
- `banking/bank-connections`: the chooser edits an owner set rather than one
  checkbox, cannot orphan, and offers leaving an account out instead; the connect
  flow's steps and their actions.
- `banking/transactions`: jumping to the newest and oldest pages and to a month;
  an account left out stops appearing in the ledger and its rows are not deleted.

## Impact

**Schema.** `apps/wimm/internal/store/migrations/` gains one migration:
`accounts.left_out_at`, `accounts.household_name`, a `date_trunc` index or a
confirmation the existing `transactions_account_date_idx` serves the aggregate,
a constraint trigger on `account_owners` asserting at least one owner per
surviving account, and the backfill of existing orphans.

**Contract.** `packages/contracts/proto/wimm/banking/v1/banking.proto`:
`SetAccountOwners` gains a minimum-one validation and stops accepting an empty
list; new `SetAccountLeftOut` and `SetAccountName`; `ListTransactionsRequest`
gains an oldest-page mode and a month cursor; `Ledger` gains the months that hold
transactions; `Account` gains the household name and the left-out fact.
Regenerated with `just gen`.

**Go.** `apps/wimm/internal/store/banking_accounts.go`, `banking.go`,
`transactions.go`; `apps/wimm/internal/banking/service.go`, `reading.go`,
`ledger.go`; `apps/wimm/internal/rpc/banking.go`, `errors.go`. The `requireOwner`
and `requireOwnerOnConnection` escape hatches for orphaned accounts are removed,
because the state they exist to rescue can no longer occur.

**Design.** `apps/web/design/01-access.pen`, `02-banking.pen`,
`03-transactions.pen` gain Medium and Ultra regimes and a new journey for
Settings; `packages/ui/design/product-ui.lib.pen` gains the components the
screens need. `packages/ui/design/tokens.json` and `just gen` for any token the
Page template needs.

**Web.** A new `Page` template and a new `Settings` route; every screen in
`packages/ui/src/pages/` for the layout pass; `apps/web/src/routes/+layout.svelte`
for the navigation indicator; `apps/web/src/app.html` for `data-density`.

**Checks.** `packages/ui/scripts/check-canvas.py` gains the new frames in
`IMPLEMENTS`; `check-geometry.py` gains the regime measurements. Two gaps in the
canvas checks are closed as part of this: the contract generator's `SCREEN`
pattern misses every frame in the access journey and every happy-path frame of
the connect flow, so `Sign in`, `Enrol a passkey`, `Choose a bank`,
`What wimm will see` and Overview's default state are drawn and asserted by
nothing today.

**Decisions.** One ADR superseding the orphan half of `ADR 0019`. `ADR 0021` is
extended rather than overturned — its refusal of page numbers is the reason the
scrubber is a month list and not a pager.
