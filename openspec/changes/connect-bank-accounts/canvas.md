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
| J03.A / 04 · Choose accounts / Wide / Nothing chosen | `LV6Bi` | 10 |
| J03.A / 04 · Choose accounts / Wide / Two chosen | `a7OaSW` | 10 |
| J03.A / 05 · Overview / Wide / Accounts connected | `W7ixB` | 10 |
| J03.A / 05 · Overview / Wide / Two currencies | `B3Q8Ez` | 10 |
| J03.A / 01 · Overview / Compact / No banks connected | `u5wHxi` | 10 |
| J03.A / 02 · Choose a bank / Compact | `A4owxP` | 10 |
| J03.A / 03 · What wimm will see / Compact | `wWf7W` | 10 |
| J03.A / 04 · Choose accounts / Compact / Nothing chosen | `g9XC9` | 10 |
| J03.A / 04 · Choose accounts / Compact / Two chosen | `Es1yG` | 10 |
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
| Choose accounts | Full page | n/a | Route-backed, `/connect/<bank>/accounts` |
| Refresh balances | Inline on Overview | n/a | No navigation |
| Restore access | Re-enters the hand-off | n/a | Route-backed, skips the picker |
| Disconnect confirmation | Confirmation dialog | Modal | Ephemeral, not route-backed |

Connecting is a full page because `docs/design/surfaces.md` names exactly this
case: complex, resumable, deep-linkable, likely to grow. The chooser is a full
page for the same reason and because it is route-backed — a member who closes
the tab mid-choice has already granted access, and must be able to come back to
it rather than lose the connection.

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
| Checkbox | `ui:UEy2Z` | Every row in the chooser |
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

**2 · Account choice row** — molecule. A checkbox, a mark, a name, an
identifier, an optional badge and a sharing status. Distinct from the Bank row
above because it is a persistent two-state control rather than a navigation
choice, and distinct from the account row below because it carries no balance —
by design, since wimm reads no balance for an account nobody has shared.

Opened and read: **Transaction row (`ui:PviJn`)** already pairs a checkbox with
a mark and text, and is the closest thing in the library, but its remaining
columns are a category chip, a date and an amount, and its selection is a
transient bulk-action selection rather than a stored choice. **Account selector
(`ui:vHJTa`)** has no checkbox and no room for one.

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
- Choose accounts with nothing chosen, Finish disabled.
- Choose accounts with two of three chosen.
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

**Nothing is shared until it is chosen.** The chooser's Finish control is
disabled with nothing ticked, and its helper text says why rather than leaving a
dead control unexplained. Selecting all is one action.

**No balance is read for an unshared account**, at connect, on arrival, or on
refresh. This is what makes leaving an account out meaningful rather than
cosmetic.

**Balances are read when the member arrives.** The read happens in the
background of the request: previous readings render immediately and update in
place, so a bank having a bad day never blocks the screen.

**Restoring keeps what was shared.** Accounts are matched across sessions by the
gateway's cross-session hash, never by its per-session identifier, so a member
who restores does not re-make every sharing choice. An account newly offered
arrives unshared and carries a badge; one no longer offered goes, and the member
is told which.

**Focus.** Returning from the bank moves focus to the chooser's heading.
Returning to a failure moves focus to the page banner. The disconnect dialog
traps focus, focuses the safe action first, and returns focus to the control it
came from whether confirmed or dismissed.

**Announcements.** Refreshing announces its outcome in a live region: the new
read time on success, the retry time when a bank refuses, and which bank did not
answer when only some succeeded. The bank search announces the number of
matches. Ticking an account in the chooser announces the running count, because
the Finish control's label changes with it.

**Keyboard completion.** Every step is completable from the keyboard alone,
including picking a bank, ticking accounts, and dismissing the dialog.

**Colour never carries direction alone.** Overdrawn balances carry a sign. A
reading that is no longer being updated changes colour and its text changes with
it.

**Themes.** One dark comparison, of the connected Overview — the screen carrying
the most colour decisions in this change. Checked against ADR 0014: content sits
one step above the chrome, separation is carried by border rather than fill, and
the primary action is the inverted mint fill rather than white on green.
