## Journey

- **J03 · Connect a bank** — a member grants access at their bank, then chooses
  which of its accounts the household sees. Stories S1.
- **J04 · Keep the household's balances current** — balances are read on arrival
  and on request. Stories S2.
- **J05 · Disconnect a bank** — a member ends access and the accounts go.
  Stories S3.
- **J06 · Restore access to a bank** — the grant has run out and the member
  confirms again at the bank. Stories S4.

J01 and J02 are taken by `01-access.pen`. These four share one file because they
share one product area.

## File

`apps/web/design/02-banking.pen`, importing `packages/ui/design/product-ui.lib.pen`
as `ui`.

## Saved to disk

Saved. `git status` shows `?? apps/web/design/02-banking.pen`. Every mutation
ran through `just pen-exec`, which saves on a clean run and writes nothing on a
failed one.

## Frames

| Frame | ID | Zone |
| --- | --- | --- |
| J03.A / 01 · Overview / No banks connected / Wide | `nCoPj` | 10 |
| J03.A / 02 · Choose a bank / Wide | `kEjpg` | 10 |
| J03.A / 03 · What wimm will see / Wide | `PobEK` | 10 |
| J03.A / 04 · Choose accounts / Wide / As it opens | `LV6Bi` | 10 |
| J03.A / 04 · Choose accounts / Wide / One disowned, one granted | `a7OaSW` | 10 |
| J03.A / 05 · Overview / Wide / Accounts connected | `W7ixB` | 10 |
| J03.A / 05 · Overview / Wide / Two currencies | `B3Q8Ez` | 10 |
| J03.A / 01 · Overview / Compact / No banks connected | `u5wHxi` | 10 |
| J03.A / 02 · Choose a bank / Compact | `A4owxP` | 10 |
| J03.A / 03 · What wimm will see / Compact | `wWf7W` | 10 |
| J03.A / 04 · Choose accounts / Compact / As it opens | `g9XC9` | 10 |
| J03.A / 04 · Choose accounts / Compact / Levels stacked | `Es1yG` | 10 |
| J03.A / 05 · Overview / Compact / Accounts connected | `Euld4` | 10 |
| J03.B / 01 · Overview / Access not granted | `hmnWx` | 20 |
| J03.C / 01 · Overview / Bank could not be reached | `sk2rn` | 20 |
| J03.D / 01 · Overview / No accounts came back | `Mdnno` | 20 |
| J04.B / 01 · Overview / Refresh refused | `AmfbN` | 20 |
| J05.A / 01 · Overview / Disconnect confirmation | `eKJdq` | 20 |
| J05.A / 02 · Overview / Disconnected | `dpe2m` | 20 |
| J06.A / 01 · Overview / Access has run out | `D8tBDe` | 20 |
| J06.A / 02 · Choose accounts / Something new to choose | `wRHUb` | 20 |
| J06.A / 03 · Overview / Access restored | `zFruq` | 20 |
| QA · Overview / Accounts connected · Color=Dark | `q65ue` | 40 |

Zone frames: `rH5Ip` 00, `eXFbr` 10, `I2ZpC` 20, `TB1k4` 30, `a1xKy` 40,
`wh0sP` 90.

## Surfaces

| Surface | Form | Modality | Routing |
| --- | --- | --- | --- |
| Choose a bank | Full page | n/a | Route-backed, `/connect` |
| What wimm will see | Full page | n/a | Route-backed, `/connect/<bank>` |
| Consent | Off product, at the bank | n/a | Leaves and returns |
| Consent return | No surface | n/a | Exchanges server side, redirects |
| Choose accounts | Full page | n/a | Route-backed, `/connect/<bank>/accounts/<connection>`, reached from Overview |
| Refresh balances | Inline on Overview | n/a | No navigation |
| Restore access | Re-enters the hand-off | n/a | Route-backed, skips the picker |
| Disconnect confirmation | Confirmation dialog | Modal | Ephemeral, not route-backed |

Connecting is a full page because `docs/design/surfaces.md` names exactly this
case: complex, resumable, deep-linkable, likely to grow. The chooser is a full
page for the same reason, and it is route-backed per bank because it is where a
member goes to answer a question about one bank, not a step they pass through
on the way to somewhere else.

**The chooser is not part of connecting.** Returning from the bank goes to
Overview. Every account the bank returned is already owned by the connecting
member and already invisible to everyone else, so the question the chooser asks
has its correct answer already filled in, and asking it there protects nothing.
Closing the tab loses no choice, because the choice that matters has already
been made by default.

The disconnect confirmation is a dialog because it is destructive and
money-bearing, and it names the bank and how many accounts go rather than asking
"Are you sure?" about nothing. It is drawn over the full page, not on a plate of
its own.

**No stepper**, recorded as an open decision in zone 90.

## Components used

| Component | Origin | Used for |
| --- | --- | --- |
| Signed-in landing | `ui:IGbQe` | Every wide screen |
| Signed-in landing / compact | `ui:pu6qZ` | Every compact screen |
| Empty state | `ui:w5ZouR` | No banks connected |
| Notice | `ui:WFSg5` | Every failure, warning and recovery banner |
| Dialog | `ui:kEly8` | Disconnect confirmation |
| Metric tile | `ui:oweC0` | Household total, `Delta` disabled |
| Account selector | `ui:vHJTa` | Every account row on Overview |
| Checkbox | `ui:UEy2Z` | Claiming ownership of a row in the chooser |
| Segmented control | `ui:TBD` | Picking a member's level on a row — **missing, see below** |
| Badge | `ui:iyeM0` | An account newly offered on restore |
| Search field | `ui:n9KlSw` | Searching the bank list |
| Button | `ui:W2gOKx` | Every action, primary, secondary and destructive |
| Icon button | `ui:mRsB3` | Refresh at compact |
| Avatar | `ui:Z2oYAy` | Bank mark, standing in for a logo |

## Components missing

**1 · Bank row** — molecule. One selectable bank in a list: a logo, a name, a
chevron, and hover, focus and pressed states.

Opened and read: **Account selector (`ui:vHJTa`)** holds a mark, a name, a meta
line and a trailing balance; the trailing slot is an amount and the meta line is
an account identifier, so a bank choice would be an account row with half its
parts switched off and a name that lies about what it holds. **Transaction row
(`ui:PviJn`)** is a five-column ledger row with a checkbox, a category chip and
a date. **Nav item (`ui:rEOk1`)** is a sidebar destination carrying current and
hover states for navigation, not a list choice with a mark.

**2 · Account choice row** — molecule. A mark, a name, an identifier, an
optional badge, an ownership control, and one level control per other member of
the household. It carries a balance, because the connecting member owns every
row when the chooser opens and an owner sees their own account in full — and
because choosing who sees an account by its name alone, with no figure, is
choosing blind.

The row is **not** a two-state control. Ownership is a checkbox; each other
member's level is one of three, so the row carries one three-way control per
member. At Compact the level controls stack under the account rather than
sitting beside it, because a household of three would otherwise need four
columns on a phone.

Opened and read: **Transaction row (`ui:PviJn`)** already pairs a checkbox with
a mark and text, and is the closest thing in the library, but its remaining
columns are a category chip, a date and an amount, and its selection is a
transient bulk-action selection rather than a stored choice. **Account selector
(`ui:vHJTa`)** has no checkbox and no room for one.

**2b · Segmented control** — atom. Origin `ui:e0JfP`, `SegmentedControl.svelte`. Three exclusive options in
a row — *nothing*, *balance*, *details* — labelled, keyboard-operable as one
control, and readable at Compact. Opened and read: the library has **Checkbox
(`ui:UEy2Z`)** and **Button (`ui:W2gOKx`)** and nothing that expresses one choice
among three. Three checkboxes would permit zero or two answers to a question with
exactly one; three buttons would not announce themselves as a single control to a
screen reader.

**3 · Account row** — molecule, an **additive extension of Account selector
(`ui:vHJTa`)** rather than a new component. It already holds the right four data
points. What it lacks is a place for the time the balance was read and for a
status badge, and an instance cannot be given a child, so the canvas had to
replace its `Balance` text with a two-line stack to draw the state at all. The
extension adds a `Reading` text and a `Badge` slot, both disabled by default, so
every existing instance renders unchanged.

Opened and read: **Metric tile (`ui:oweC0`)** carries a label, one value and a
delta — one figure, not a list row. **Badge (`ui:iyeM0`)** is the right atom for
the status but has nowhere to sit inside the row as it stands.

**4 · `color-scrim` token** — the wash behind a modal. The library has no token
for it and the canvas uses a raw hex. It surfaced only when the disconnect
dialog was drawn over the whole page rather than on a plate of its own. It ships
in the same task as the components, which means re-exporting the library and
running `just gen`; nothing hand-edits `tokens.css`.

Nothing else is missing. The bank list card, the consent explainer, the chooser
card and the accounts card are compositions of components that already exist.

## States drawn

Beside their step, wide and compact unless noted. The full index with reasons,
including everything deliberately not drawn, is zone 30 on the canvas.

- Overview with no banks connected.
- Bank list, default.
- Consent explainer, carrying the date that bank's access will end.
- Choose accounts as it opens: every row owned by the connecting member, every
  other member at *nothing*.
- Choose accounts with one row disowned and one granted *balance* to a second
  member.
- Choose accounts with a second member at *details* on one row and *nothing* on
  another, showing the two levels side by side.
- Choose accounts at Compact, where the level controls stack under the row.
- Overview with accounts connected.
- Overview with accounts in two currencies (wide only — one total per currency
  is not a regime-dependent rule).

Zone 20, wide only:

- Access not granted at the bank.
- Bank could not be reached.
- No accounts came back.
- Refresh refused by the bank.
- Access has run out, with the last readings kept and marked.
- Restoring, where the bank now offers an account it did not before.
- Access restored.
- Disconnect confirmation, over the page, and the screen after confirming.

**Deliberately not drawn**, each with its reason in zone 30: the bank list with
no match and while loading; balances loading on arrival, which renders the
previous readings and updates in place rather than being a screen of its own;
refresh in progress, which is the Button component's pending state and is
storied there; and compact copies of zone 20's branches, since each is the
compact Overview plus the same Notice and the shell is drawn four times already.

## Contracts for implementation

Every number below was measured on the canvas, not recalled.

**Geometry.** At 1440 the sidebar is 264 and the Main column is 1176. Compact is
390 × 844 with a 56-high app bar and no sidebar. Account rows measure 56 high,
uniform across all rows in both regimes — asserted rather than eyeballed,
because a dropped size binding leaves rows hugging to something shorter and
reads as a tight table rather than a defect.

**The consent return.** The value the bank returns with never reaches an address
a member can bookmark or share. The route exchanges it server side and redirects,
the same shape `/enrol/[link]` already uses, and logs never record the return
path with its value. The canvas cannot execute any of this; it is drawn as a
transition and nothing more.

**Nobody else sees anything until they are given it.** Every other member starts
at *nothing* on every row, and Finish is always enabled — finishing having
granted nothing is a legitimate outcome, because the accounts are the connecting
member's and they can see them. Setting one member to one level across every row
at that bank is one action.

**No balance is read for an account with no owner and no grant**, at connect, on
arrival, or on refresh. The canvas cannot execute that rule; it is a contract on
the routes.

**Every account a bank returns is read once before anyone can disown it**, and
this is a consequence of the chooser leaving the connect flow rather than a gap
in the rule above. Returning goes to Overview, where the connecting member owns
everything the bank returned, so everything is readable and everything is read.
Disowning happens afterwards and stops every later read.

ADR 0019 left this window open deliberately and described it as the case of a
member who abandons the chooser. It is now the only path, so the reading is
certain rather than possible. Closing it would mean showing no figures until a
member had finished choosing — which is the empty screen that argument
rejected — and the account being read is one the connecting member can already
see by logging in at their bank.

**Balances are read when the member arrives.** The read happens in the
background of the request: previous readings render immediately and update in
place, so a bank having a bad day never blocks the screen.

**Restoring keeps owners and levels.** Accounts are matched across sessions by
the gateway's cross-session hash, never by its per-session identifier, so a
member who restores does not re-make every choice. An account newly offered
arrives owned by the restoring member with every other member at *nothing*, and
carries a badge; one no longer offered goes with its owners and grants, and the
member is told which.

**Focus.** Returning from the bank moves focus to Overview's outcome notice,
which names the bank and what happened to it. Returning to a failure moves
focus to the page banner. The disconnect dialog
traps focus, focuses the safe action first, and returns focus to the control it
came from whether confirmed or dismissed.

**Announcements.** Refreshing announces its outcome in a live region: the new
read time on success, the retry time when a bank refuses, and which bank did not
answer when only some succeeded. The bank search announces the number of
matches. Changing a level in the chooser announces the account and the level it
moved to, not a running count: with three levels per member there is no single
number to count, and "Joint current account, balance" is what the member needs to
hear back.

**Keyboard completion.** Every step is completable from the keyboard alone,
including picking a bank, claiming or releasing an account, moving a member
between the three levels with arrow keys inside one control, and dismissing the
dialog.

**Colour never carries direction alone.** Overdrawn balances carry a sign. A
reading that is no longer being updated changes colour and its text changes with
it.

**Themes.** One dark comparison, of the connected Overview — the screen carrying
the most colour decisions in this change. Checked against ADR 0014: content sits
one step above the chrome, separation is carried by border rather than fill, and
the primary action is the inverted mint fill rather than white on green.
