## Context

See proposal.md — Why. Three facts about what exists constrain everything below.

**The orphan is reachable at every layer and refused at none.** The proto marks
`SetAccountOwners.member_ids` as "may be empty"; the handler accepts it; the
service checks only that the caller is an owner; the store deletes every owner
row and inserts none; and `account_owners` has no constraint on how many rows an
account has. The web chooser's "mine" checkbox posts `memberIds: []` when
unticked and `memberIds: [self]` when ticked, so unticking orphans and ticking
silently removes a joint account's other owner. `requireOwner` and
`requireOwnerOnConnection` each carry an escape hatch letting the connection's
`connected_by` member reclaim an orphan, added because without it disowning your
last account locks you out of the screen that could undo it.

**The account upsert is where a household name would be lost.** `insertAccounts`
conflicts on `(bank_id, gateway_ref)` and its `do update set` includes `name`,
so connect, restore and reconnect all overwrite it from the gateway.

**The layout system has three breakpoint families and one of them is nobody's.**
`tokens.css` flips at 768, 1200 and 1800 — the declared device axis. Both shells
flip at 1024, which is between medium and wide. Four leaf components flip at
599, which is 168 short of the compact boundary. Every per-regime layout token
(`layout-sidebar-width`, `layout-subnav-width`, `layout-page-gutter`,
`layout-header-height`, `layout-inspector-width`, `layout-drawer-width`) is
consumed by nothing; the shell hard-codes 264, 56 and 60, and uses `space-8` for
the gutter at every size. `layout-content-max` is used on two screens of nine.

## Language

- **owner** — a member who sees an account in full and may change who else does.
  Ownership is a set, never a level and never a degree. Never "sharer".
- **level** — what a non-owner has been granted on one account: *balance* or
  *details*. Absence of a grant is **hidden**. Unchanged from ADR 0019.
- **left out** — an account the household keeps a record of and wimm does not
  read. A property of the account, not of a member's view of it. Column
  `left_out_at`; RPC `SetAccountLeftOut`; the screen says "Leave out" and "Bring
  back". Never "hidden" — that word is a grant level and reusing it would make
  two different things share a name. Never "archived", "excluded", "disabled",
  "ignored" or "untracked".
- **household name** — the name an owner gave an account, shown to everybody who
  may see it. Column `household_name`. Never "nickname", "alias", "label",
  "custom name" or "display name".
- **bank name** — what the bank calls the account, column `name`, written only by
  the gateway upsert and never by a member.
- **regime** — one of *compact*, *medium*, *wide*, *ultra*. A **breakpoint** is
  the width at which one regime becomes the next; a regime is not a breakpoint
  and the two words are not interchangeable.
- **month** — an entry in the ledger's list of months, each one a place to start
  reading. Never a "filter" and never a "period": a filter would cut the list
  down, and this does not.
- **step** — one screen of connecting a bank. There are four; the hand-off to
  the bank is the one that commits.

## Goals / Non-Goals

**Goals**

- Zero owners is not a state the database can hold, whatever calls it.
- The privacy affordance ADR 0018 exists for survives, reversibly.
- The ledger can be moved through in fewer than one click per page.
- One layout system, four regimes, every screen drawn before it is built.

**Non-Goals**

- No sharing surface spanning banks. ADR 0019 named it as the shape and left it
  out deliberately; nothing here changes that.
- No categories, budgets or search on the ledger. The month list is navigation,
  not a query language.
- Nothing fills the inspector region at ultra. The library already holds the
  Drawer that ADR 0002 meant, and already draws it promoted in
  `Transactions / Ultra 1920`; what is missing is a screen in the app that opens
  a drawer at all. The Page template carries the region and no screen passes
  one; see the decision below.
- No member deletion. There is no route to delete a member today, and this
  change does not add one.
- No change to what a grant level shows. Transactions stay owner-only (ADR
  0021), and *details* stays the top level.

## Decisions

### The minimum-owner guard is a deferred constraint trigger

`create constraint trigger … deferrable initially deferred`, on
`account_owners` after delete and update, and on `accounts` after insert. Each
fires per row, checks the account still exists, and raises if that account has
no owner row.

Deferred is not a preference. `SetAccountOwners` deletes every owner row and
then inserts the new set inside one transaction, so a non-deferred trigger fires
in the gap between the two and refuses a change that is legal. Deferring moves
the check to commit, which is the only moment the question has a stable answer.

Checking that the account still exists is the other half. Deleting an account —
or deleting a connection, which cascades to its accounts — cascades to its owner
rows, and a trigger that did not look would refuse every disconnection.

*Alternative: a check in Go.* Rejected, and the reason is on disk: the exclusion
between owning and holding a level is already guarded in Go
(`ErrOwnerHoldsAGrant`) and is enforced on the other write path by silently
deleting the grant instead. Two call sites, two behaviours, one rule. A guard
the database holds has one behaviour.

*Alternative: `accounts.owner_id not null`.* Rejected: it would undo the
many-to-many that ADR 0019 exists for, and a joint account is the case the model
was built for.

*Alternative: `check (exists …)`.* Not available — a check constraint cannot
read another table.

**The escape hatches come out.** `requireOwner` and `requireOwnerOnConnection`
each let the `connected_by` member act on an account nobody owns. That state
becomes unreachable, so the branches become dead code that would silently grant
authority if the invariant ever broke. Their tests become tests of the refusal.

**Member deletion, if it is ever added, will be refused by this trigger** where
the member is an account's last owner. That is the correct failure: it surfaces
at the moment the decision is made rather than producing an account nobody can
reach. A test asserts it rather than leaving it to be discovered.

### Left out is a nullable timestamp on the account

`accounts.left_out_at timestamptz`. Null means the account is in wimm; a value
means it is not, and when it stopped being. Clearing it brings the account back.

It belongs on the account rather than in a table because it is a statement about
the account's presence in the product, not about one member's view of it. The
per-member version of the same idea already exists and is called *hidden* — it
is the absence of a grant. Two mechanisms for one question is what ADR 0019
refused when it made ownership and level distinct, and the same argument holds
here.

A timestamp rather than a boolean, so the screen can say when without a second
column, and so the value is the same shape as every other lifetime in the
schema.

**One predicate replaces the two-part test.** `ReadableAccounts` currently ends
with `exists (owner) or exists (grant)`. Because an account now always has an
owner, that whole clause reduces to `left_out_at is null`, which is the boundary
moving from "unowned and ungranted" to "left out" exactly as the proposal says.
The transactions scope gains the same clause.

**Redaction stays in the handler**, per ADR 0019. `visibleAccountsQuery` returns
a left-out account to its owners and to nobody else; the handler omits the
balance and its read time rather than the query nulling them, so there remains
one place where what a member may see is decided.

**Bringing an account back announces its grants first.** The grants survive
being left out, so bringing an account back can re-expose it to a member who
will not have been asked again. The screen names who will see it and at what
level before the account comes back. This is the only place in the change where
a confirmation is required rather than offered, and it is required because the
consequence lands on somebody who is not in the room.

### The household name is a second column the upsert does not name

`accounts.household_name text`. The gateway upsert's `do update set` list is
where a name gets lost, and the fix is that the column is absent from it. That
is the entire mechanism, and it is why this cannot be a value written into
`name`.

Reads become `coalesce(household_name, name)` at the three sort orders that
currently order by `a.name` and at every place a name is serialised. The bank's
own name stays available beside it.

*Alternative: a name per member, following `account_grants`' shape.* Rejected.
Two members calling one account different things makes the conversation the
product exists to support harder, not easier — "did you see the four hundred
leave the joint account" needs one name. The grantee case that motivated the
per-member shape is already answered: a grantee sees the household's name,
because it reveals nothing the bank's name did not.

*Alternative: let the member edit `name` and stop the upsert overwriting it.*
Rejected: the bank's name is evidence of what the bank said, and a rename would
destroy it. Both facts are worth keeping and they are two facts.

### The month list is one aggregate, and a month jump is a seek

```sql
select date_trunc('month', booking_date)::date as month, count(*)
  from transactions
 where account_id in (<the member's owned, not-left-out accounts>)
 group by month
 order by month desc
```

The existing `transactions_account_date_idx` on
`(account_id, booking_date desc, id desc)` covers the account predicate and the
column being grouped, so this is an index-only scan over the same rows the list
already pages through. It runs with the ledger page, in the same request, and
its result is the set of months offered — which is why a month with nothing in
it cannot be offered: it is not in the result.

A month is a **cursor**, resolved on the server. The route gains `?month=YYYY-MM`
beside `?before=` and `?after=`, and the store resolves it to
`(booking_date, id) <= (last day of that month, the largest id on that day)`
ordered descending — inclusive, so the member lands *on* the month's newest page
rather than just past it. Everything after that is the existing older-ward seek,
which is what makes paging on from a month run into the month before it.

*Alternative: `?before=<the month's newest cursor>`.* Rejected: `before` is
strictly older, so landing on a month would mean the client computing the row
one newer than the month's newest, which it cannot know.

*Alternative: a `where date_trunc(...) = month` filter.* Rejected by the spec: a
month is a place to start reading, and a filter would stop the member paging out
of it.

**Newest** is the cursor being absent, which the route already expresses as
`/transactions` with no query — no store change at all. **Oldest** is
`order by booking_date asc, id asc limit n` with the rows reversed in Go, which
is the shape the existing newer-ward branch already has.

### booking_date becomes not null, and gets a fallback before it is written

The column is nullable and nothing writes null — `store.Transaction.BookingDate`
is a `time.Time`, not a pointer — so a bank that returns no booking date writes
the zero time and the row sorts to year 1. The month list would offer "January
0001" to anyone whose bank did that, and the tuple seek behaves differently for
a null than for a zero, which is a difference nobody should have to hold.

So: the adapter falls back through value date then transaction date before
writing, which is the order `identity.go` already uses to build a digest, and
the migration makes the column `not null`. A transaction with no date at all
from any of the three is refused by the sync rather than stored undated.

### One Page template, and the layout tokens are consumed

`packages/ui/src/templates/Page.svelte` takes the page heading, an optional
subsection list, the content, and an optional inspector region. It owns the page
gutter, the section gap, the measure cap for prose and the heading's type — the
six hand-rolled `.screen` rules and the four disagreeing `h1` blocks go.

The shell's 1024 becomes 768 / 1200 / 1800. The four 599s become 768. Nothing in
the product declares a width of its own afterwards.

**The hard-coded sizes were hard-coded for a reason that does not apply to CSS.**
`SignedInLanding.svelte` records that pen discards a size bound to a variable, so
the *drawing* holds 56 as a literal. That is a fact about the canvas tool. The
Svelte component has no such constraint and uses the custom property, and the
frames keep their literals.

**`layout-inspector-width` and `layout-drawer-width` stay unconsumed, on
purpose, and the reason is narrower than it first looks.** The design is not
missing: `Drawer` is in the library, its contract already says that at Ultra it
"stops being an overlay in the other direction and becomes the persistent
inspector pane, with the elevation and the close control removed because nothing
is dismissed", and `Transactions / Ultra 1920` draws exactly that. What is
missing is a screen in the app that opens a drawer at all — nothing in wimm
edits a record in place yet, so there is nothing to promote.

So the Page template carries the region and no screen passes one; at ultra with
no inspector, the extra width goes to the page. Recorded in canvas.md as a
deliberate difference rather than left as two tokens nobody can explain.

**One number in the drawn Ultra template disagrees with the token that names
it.** `Transactions / Ultra 1920` gives the sidebar 264 where
`layout-sidebar-width` says 288 at ultra, and the templates' own contract table
lists "264 / 288 sidebar … they come from layout tokens" as a *contract* rather
than an illustration. The frame is redrawn to 288, which is the direction ADR
0020 prescribes when the two disagree and the token is the one with a stated
reason behind it.

### Navigation feedback is one component in the root layout

SvelteKit's `navigating` state, read in `routes/+layout.svelte`, driving one
indicator fixed to the top of the viewport. It does not shift the page, which is
what `position: fixed` and no reserved space give.

Two timings, both for the same reason — a flicker is worse than no feedback:

```text
150 ms   before it appears, so a fast screen shows nothing
400 ms   minimum on screen once it has appeared
```

Announcement is a polite live region, not an assertive one, because the spec
says it waits its turn. It says the screen is loading on start and names the
screen on arrival; the name comes from the route, so it is the one place in the
app that needs a route-to-name map.

Under reduced motion the indicator appears and does not animate. The motion
duration tokens already collapse to zero, and the component adds nothing that
would survive that.

### Density reaches the document the same way the theme already does

`app.html` already applies the stored theme in an inline, blocking script,
because setting it from a component runs after first paint. Density has exactly
the same requirement and gets exactly the same treatment: `wimm-density` in
local storage, `data-density` on the document element, set in the same script.

Copying an existing mechanism rather than inventing a second one is the point.
A cookie read on the server was considered and rejected: the theme does not use
one, and two preferences applied by two different mechanisms is how one of them
comes to be applied a frame late.

**The control is hidden, not disabled, where the pointer is coarse** — ADR 0004,
unchanged. `tokens.css` already reverts density under `any-pointer: coarse`, so
the token half is done; the missing half is the control, which is hidden by the
same media query. A member who chose compact on a laptop keeps that choice; the
phone simply does not honour it and does not offer to.

### The canvas checks grow to cover what they currently miss

`gen-canvas-contract.py` recognises a screen by `^J\d+\.\w+ / \d+`, which the
access journey's frames and the connect flow's happy-path frames do not match —
so Sign in, Enrol a passkey, Choose a bank, What wimm will see and Overview's
default state are drawn, built, and asserted by nothing. This change adds frames
to those screens, so it closes the gap rather than widening it: the pattern is
extended, `IMPLEMENTS` gains the screens, and `test-contract.py` gains a fixture
for a frame in each newly-covered shape.

That is a generator change, and ADR 0020 is explicit about which way a generator
fails: quietly. So the fixtures come first, asserting that a frame of each shape
is extracted, before the pattern is touched.

## Risks / Trade-offs

**The trigger refuses a member deletion that has no other route** → There is no
member-deletion path today, so nothing can hit it. If one is added, the refusal
is what should happen, and a test pins it so the next person meets it in a test
rather than in production.

**The backfill needs somebody to give the orphans to** → Existing orphans are
gateway accounts, whose connection carries `connected_by`. A manual account
cannot currently be orphaned because it cannot be created — there is no route,
only test fixtures. The migration asserts it found an owner for every row it
touched and fails rather than leaving one behind.

**Bringing an account back re-exposes it to a member who was not asked** → The
confirmation names them and their level before it happens. The alternative —
dropping grants when an account is left out — was rejected because it turns a
reversible act into a destructive one and makes the member re-grant from memory.

**The month aggregate on a large ledger** → Bounded by the member's own accounts
and served by the index the list already uses. Measured on a seeded ledger before
the screen is built; if it does not hold, the fallback is a stored per-account
month summary, which is a bigger change and is not started speculatively.

**Redrawing every screen at two new regimes is most of this change's wall clock**
→ It is also the gate on almost everything else, so it goes first and the
component work runs beside it. The screens themselves cannot start before their
frames exist, and ADR 0020 is what makes that a rule rather than a preference.

**The checks cannot see whether a layout is good** → Stated rather than
mitigated. `check-canvas.py` holds copy and `check-geometry.py` holds
measurements; that Overview reads well at medium is a judgement a person makes
against the frame. The checks are a floor, which is what ADR 0020 already says
about them.

## Migration Plan

One goose migration, in this order within the transaction:

1. `alter table accounts add column left_out_at timestamptz` and
   `add column household_name text`.
2. Backfill: every account with no owner row becomes owned by its connection's
   `connected_by` member and gets `left_out_at = now()`. Assert the count of
   ownerless accounts is zero afterwards, and fail the migration if it is not.
3. Backfill `transactions.booking_date` for any zero-dated row from its value
   date or transaction date, then `alter column booking_date set not null`.
4. Create the constraint triggers. They are created after the backfill so the
   backfill is not fighting them.

Rolling back is the reverse: drop the triggers, drop the two columns, restore
the column's nullability. Nothing is lost — an account that was left out becomes
an account that is simply read again, which is the pre-change behaviour.

The proto changes are additive except one: `SetAccountOwners` stops accepting an
empty `member_ids`. Its only caller ships in the same deploy.

## Open Questions

- How far back the month list should reach when a household has years of
  ledger. Answering it changes a limit, not the approach: the aggregate, the
  route and the screen are the same whether it offers every month or the last
  thirty-six. Measure first.
- Whether the Page template should carry the inspector region at all while no
  screen opens a drawer. It is one prop and one column, the library has already
  designed what goes in it, and the decision is cheap either way — better made
  with the redrawn Ultra frames in front of us than now.
