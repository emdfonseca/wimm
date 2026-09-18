## Journey

- **J03 · Connect a bank** — existing. Gains Medium and Ultra, a step indicator
  and one action footer on every step. Stories S6, S7.
- **J07 · See where the money went** — existing. Gains Medium and Ultra, the
  month strip on every frame, a state for jumping to a month, and the state a
  member sees while a screen is still arriving. Stories S4, S7, S8.
- **J09 · Decide who owns an account** — new. Making another member an owner,
  handing an account on, the refusal that stops the last owner stepping back,
  leaving an account out and bringing it back. Stories S1, S2, S3.
- **J10 · Name an account** — new. Stories S5.
- **J11 · Set how wimm looks** — new. Stories S9, S7.

J01 and J02 are unchanged except for a note; see *States drawn*.

## File

- `apps/web/design/02-banking.pen` — J03, J09, J10
- `apps/web/design/03-transactions.pen` — J07
- `apps/web/design/04-settings.pen` — J11, new, created by `just pen-import`
- `apps/web/design/01-access.pen` — a note only
- `packages/ui/design/product-ui.lib.pen` — the components below

## Saved to disk

All five files saved by `just pen-exec` on clean runs and confirmed with
`git status`:

```text
 M apps/web/design/01-access.pen          12:22
 M apps/web/design/02-banking.pen         12:17
 M apps/web/design/03-transactions.pen    12:26
?? apps/web/design/04-settings.pen        12:21
 M packages/ui/design/product-ui.lib.pen
 M packages/ui/design/library-manifest.tsv
```

`just pen-manifest verify` passes: nothing renamed, retyped, reparented or
removed that was not the change being made. `just check packages/ui` passes.

**Pen.app is holding `03-transactions.pen` open.** Its in-memory copy predates
this work and does not know the new library components. Close that document
**without saving**, or reopen it, before touching it in the app — its next save
would overwrite everything recorded here for that file.

## Frames

### 02-banking.pen · J03 (zone 10, MEDIUM 1024 and ULTRA 1920 rows)

| Frame | ID | Zone |
| --- | --- | --- |
| 01 · Overview / Medium / No banks connected | `EdyGN` | 10 |
| 02 · Choose a bank / Medium / Default | `uHGYx` | 10 |
| 03 · What wimm will see / Medium / Default | `acQ78` | 10 |
| J03.A / 04 · Choose accounts / Medium / As it opens | `IlHjD` | 10 |
| J03.A / 04 · Choose accounts / Medium / One disowned, one granted | `v6TfH` | 10 |
| 05 · Overview / Medium / Accounts connected | `ciEsZ` | 10 |
| 05 · Overview / Medium / Two currencies | `Mqjcx` | 10 |
| 01 · Overview / Ultra / No banks connected | `xzxSZ` | 10 |
| 02 · Choose a bank / Ultra / Default | `ipk89` | 10 |
| 03 · What wimm will see / Ultra / Default | `O8KHK` | 10 |
| J03.A / 04 · Choose accounts / Ultra / As it opens | `c2fVI` | 10 |
| J03.A / 04 · Choose accounts / Ultra / One disowned, one granted | `pwgpr` | 10 |
| 05 · Overview / Ultra / Accounts connected | `tQ7v5` | 10 |
| 05 · Overview / Ultra / Two currencies | `pIOen` | 10 |

Changed in place, all four regimes: `kEjpg` `A4owxP` `uHGYx` `ipk89` (step 1) and
`PobEK` `wWf7W` `acQ78` `O8KHK` (step 2) each gain a step indicator and the one
action footer. Nine `Finish` buttons across the file gained the label they never
had — see *Contracts*.

### 02-banking.pen · J09 and J10 (zone 20)

| Frame | ID | Zone |
| --- | --- | --- |
| J09.A / 01 · Choose accounts / Wide / Making an owner | `CJ3jD` | 20 |
| J09.A / 02 · Choose accounts / Wide / Handed on | `LvMRW` | 20 |
| J09.A / 03 · Choose accounts / Wide / The last owner cannot step back | `X424Q` | 20 |
| J09.A / 04 · Choose accounts / Wide / Leaving an account out | `BaB36` | 20 |
| J09.A / 05 · Overview / Wide / An account left out | `Oxc2l` | 20 |
| J09.A / 06 · Choose accounts / Wide / Bringing it back | `navZJ` | 20 |
| J10.A / 01 · Choose accounts / Wide / Naming an account | `D9X0CH` | 20 |
| J10.A / 02 · Overview / Wide / The household's name | `zIKSr` | 20 |

Path groups: `waU2O` (J09.A), `WBPqz` (J10.A).

### 03-transactions.pen · J07

| Frame | ID | Zone |
| --- | --- | --- |
| J07.A / 00 · Transactions / Wide / Waiting for the screen | `nTE3A` | 10 |
| J07.A / 05 · Transactions / Wide / Jumped to a month | `aHGv0` | 10 |
| J07.A / 01 · Transactions / Medium / As it opens | `NGeAS` | 10 |
| J07.A / 02 · Transactions / Medium / One account | `P0AY0` | 10 |
| J07.A / 03 · Transactions / Medium / An older page | `vnaNb` | 10 |
| J07.A / 04 · Transactions / Medium / The oldest page | `ilYhK` | 10 |
| J07.A / 01 · Transactions / Ultra / As it opens | `ZCxhb` | 10 |
| J07.A / 02 · Transactions / Ultra / One account | `l1c1TQ` | 10 |
| J07.A / 03 · Transactions / Ultra / An older page | `dnzXn` | 10 |
| J07.A / 04 · Transactions / Ultra / The oldest page | `fNoMH` | 10 |

The month strip was added to all 19 Transactions frames in the file, including
the failure and recovery states, because it is part of the screen rather than a
mode of it.

### 04-settings.pen · J11

| Frame | ID | Zone |
| --- | --- | --- |
| J11.A / 01 · Settings / Wide / Appearance | `aKoCs` | 10 |
| J11.A / 01 · Settings / Medium / Appearance | `YyMbD` | 10 |
| J11.A / 01 · Settings / Ultra / Appearance | `jfrPt` | 10 |
| J11.A / 01 · Settings / Compact / Appearance | `UUFeF` | 10 |
| J11.B / 01 · Settings / Compact / On a touch screen | `hTkMR` | 20 |

Journey overview `kWJZq`; zones `H01ss` `C3UtLc` `LBFZU` `SwCNQ` `eolMa` `fSEVL`.

## Surfaces

```text
Settings                      full page, route-backed, /settings
Who sees these accounts       full page, route-backed, unchanged by this change
Leave an account out          modal dialog, ephemeral — one decision, nothing
                              behind it usable, Back does not close it
Bring an account back         modal dialog, ephemeral — same shape, and it is
                              required rather than offered because the
                              consequence lands on a member who is not there
The last owner refusal        inline, beside the row where the change was made,
                              recoverable immediately, no overlay
Name an account               inline, ephemeral, on the row it belongs to
Month strip                   inline, part of the ledger region, not an overlay
Navigation progress           fixed to the viewport top, reserves no space,
                              owned by the root shell and never by a screen
```

The connect flow keeps its four full pages. Nothing became an overlay and
nothing stopped being one.

## Components used

From `packages/ui/design/product-ui.lib.pen`:

```text
Signed-in landing            IGbQe    the Wide shell
Signed-in landing / compact  pu6qZ    the Compact shell
Account choice row           QU4bL    the chooser's row, extended below
Account selector             vHJTa    an account on Overview
Ledger row                   VE2z9    a transaction
Seek pager                   Hxmzc    Newer and Older, and the span
Dialog                       kEly8    both confirmations
Notice                       WFSg5    the refusal, and being handed an account
Checkbox                     UEy2Z    ownership
Segmented control            e0JfP    a level, and the theme
Segmented control            hcT3D    row height, through the preset below
Button                       W2gOKx   every action
Category chip                RnpZH    a month in the strip
Sidebar nav / in shell       q7dY7c   Wide and Ultra navigation
Sidebar nav/rail             UwIRb    Medium navigation
Bottom nav                   b3MTx    Compact navigation
App header / in shell        s2VNo8   the Page template's header
Empty state                  w5ZouR   a shell with nothing in it yet
```

## Components missing

Every entry below was built during this canvas work, in the library, and each
one records the components that were opened and read first.

**Navigation progress** — origin `Mdi1X`, block `C9DT1w`, 20 · ATOMS.
Read `Notice` (`WFSg5`), which is a message about something that happened rather
than a report that something is happening, and carries a title, a body and an
action that this must not have; and `Empty state` (`w5ZouR`), which fills a
region rather than sitting above one. Searched the library for anything holding
"work in progress" and found `PendingButton`'s equivalent only inside `Button`'s
own states, which is per-control and cannot describe a navigation.

**Step indicator** — origin `W8V14`, block `EJBTW`, 30 · MOLECULES.
Read `Breadcrumbs` (`L4bVT`), which names ancestors rather than position in a
sequence, and `Tab` (`CtsNw`), which is a chooser between peers and not a
progression. Neither can say "two of two".

**Step actions** — origins `o1elL` and `prZYD` (compact), block `Onn81`,
30 · MOLECULES. Read `Dialog`'s footer (`jbhwl`), which is the right arrangement
and belongs to the dialog, and `Drawer`'s footer (`rtEgS`), the same. Lifting
either would mean instancing a dialog to get a page's action row. This is a
preset of two `Button` instances and redraws no chrome.

**Month scrubber** — origin `W8ig0`, block `Z2ra6`, 30 · MOLECULES.
Read `Seek pager` (`Hxmzc`), which moves one page at a time and whose contract
is explicit that where you are is a span of dates; `Pagination` (`HlRZX`), which
is the offset control ADR 0021 refuses; and `Account selector` (`vHJTa`), in
case the library already had a horizontal chooser of values — it is a row, not a
chooser. The months themselves are `Category chip` instances with the dot
switched off, so the chip chrome is not redrawn.

**Density control** — origin `Yt1rq`, block `PD8Hj`, 30 · MOLECULES.
A preset wrapping one `hcT3D` instance. Read `ThemeToggle`'s library
counterpart and found none: the theme control exists in code and has never had
an origin, which is why `Segmented control` is instanced directly for it here.

**Account name** — origin `FguRX`, block `ggCvV`, 30 · MOLECULES.
Read `Form field` (`gJtKG`), which is the right shape for a field in a form and
carries a label above and help below with no room for an inline save; and
`Search field` (`n9KlSw`), which is a filter. This is an `Input` instance plus a
`Button` instance and a helper line.

**Page** — origin `BPW3V`, 50 · TEMPLATES.
Read `List scaffold` (`eykQg`), which has no page header at all; `Detail
scaffold` (`LHBGo`), which has one with breadcrumbs enabled for a nested record;
and both Settings templates (`H0APd`, `Z9Dpm`), which hand-roll a page header
rather than instancing `App header`. Three near-misses and no origin, which is
exactly the gap the six hand-rolled `.screen` rules in code grew into.

**Signed-in landing / rail** — origin `WMlvF`, 50 · TEMPLATES.
Read `Signed-in landing` (`IGbQe`) and `Signed-in landing / compact` (`pu6qZ`),
the only two shells that exist. `Sidebar nav/rail` (`UwIRb`) and `Nav item/rail`
(`Ogizy`) were already in the library and instanced by exactly one illustration
template; nothing composed them into a shell. This does.

**Signed-in landing / ultra** — origin `NEIet`, 50 · TEMPLATES.
Same two shells read. Ultra differs from Wide by the sidebar width and the
gutter, both of which are literals in a shell because pen discards a size bound
to a variable, so it cannot be an instance override of the Wide shell.

**Settings, in the bottom bar** — `wCboc` / `H0gal`, inside `Bottom nav`
(`b3MTx`). Read `Bottom nav`'s two tabs and `Sidebar nav` (`fwKnX`), which
already carries a Settings row (`zVJvb`). The bar was the one navigation form
with no way to reach Settings.

**Account choice row, extended** — `QU4bL` gains an `Owners` row (`o0Tc2`) and a
`Mine` label (`hbN3e`). Read the row itself first: its ownership control was a
bare `Checkbox` with no visible label and no way to name anybody else, which is
the control the whole of finding 1 is about. Extended rather than duplicated,
because a second account-choice row would be two components owning one job.
Additive: nothing existing was renamed, retyped or reparented.

**Account selector, left-out state** — `NZhzh`, a state cell beside the existing
ones. `vHJTa` already carries `Reading` and `Badge` as `enabled:false` slots,
which is the library's own additive-extension pattern; this uses them rather
than adding anything.

## States drawn

```text
J03   step indicator and action footer      Wide, Compact, Medium, Ultra
J03   every existing state                  Medium and Ultra added
J07   as it opens / one account /           Wide, Compact (1–2 only), Medium, Ultra
      an older page / the oldest page
J07   jumped to a month                     Wide
J07   waiting for the screen                Wide
J09   making an owner                       Wide
J09   handed on                             Wide
J09   the last owner cannot step back       Wide
J09   leaving an account out                Wide
J09   an account left out, on Overview      Wide
J09   bringing it back                      Wide
J10   naming an account                     Wide
J10   the household's name, on Overview     Wide
J11   appearance                            Wide, Compact, Medium, Ultra
J11   on a touch screen                     Compact
```

**Deliberately not drawn, with reasons:**

- **Medium in `01-access.pen`.** Every screen there is `Auth shell`: a centred
  card with no navigation, identical between 768 and 1200. A Medium frame would
  be a second copy of Wide that somebody has to keep in step forever. Recorded
  as a note in that file's zone 90, which is what the design standard prescribes
  for a regime that adds no structural information. Compact and Ultra are drawn,
  because the card's behaviour does change at both.
- **Medium and Ultra for J09, J10 and every J07 failure state.** These are states
  of screens whose four regimes are drawn on the primary path. A failure notice
  does not rearrange differently from the screen it sits on, and drawing each
  one four times would quadruple the file to say nothing new. This follows the
  file's existing practice: most states have had a Wide frame only since
  `connect-bank-accounts`.
- **The waiting state on more than one screen.** It is one component in the root
  shell, identical everywhere. Drawn once, on the screen with the slowest
  arrival.
- **A saving state, a validation error and an empty state in Settings.** Nothing
  is saved, neither control can be wrong, and a screen of settings is never
  empty.
- **A subsection list card in Settings.** One section exists. ADR 0005 picks the
  form from the content: a list card is for a longer or growing set, tabs for a
  fixed handful of peers, and one section is neither. The library's Settings
  templates draw six rows and their own contract table marks those rows
  provisional; drawing them here would turn an illustration into a promise.
- **Anything filling the inspector region at Ultra.** `Drawer` (`yhPTP`) is in
  the library and its contract already says that at Ultra it becomes the
  persistent inspector pane; `Transactions / Ultra 1920` draws exactly that. No
  screen in wimm opens a drawer, so there is nothing to promote. The `Page`
  origin carries the region, disabled, and the drawn Ultra frames give the extra
  width to the page.

## Contracts for implementation

Every number below was measured on the canvas after layout, in a call separate
from the one that made the change.

**Page** (`BPW3V`) — 1176 × 620 at Wide. Header 1176 × 96. Body 1176 × 524 with
`layout-page-gutter` padding. Content column 1144 × 492, gap
`layout-section-gap`. Subsection slot 212 wide, disabled. Inspector slot 400
wide, disabled.

**Signed-in landing / rail** (`WMlvF`) — 1024 × 760. Rail 72 × 760, right border
1. Main 952 × 760.

**Signed-in landing / ultra** (`NEIet`) — 1920 × 860. Sidebar 288 × 860. Main
1632 × 860. `Transactions / Ultra 1920` in the library draws its sidebar at 264,
which disagrees with `layout-sidebar-width` at Ultra and with the templates'
own contract table; the token is right and that frame is a task to redraw.

**Navigation progress** (`Mdi1X`) — 3 px tall, full viewport width, `color-accent`
on `color-bg-subtle`. Fixed to the viewport, reserving no space. **150 ms before
it appears; at least 400 ms on screen once it has.** Announced through a polite
live region: that a screen is loading on start, and which screen arrived on
finish. Under reduced motion it appears and does not animate. None of the three
timings, the announcement or the politeness is drawable; all are verified in
code.

**Step indicator** (`W8V14`) — 360 wide, segments 3 px with a 4 px gap, label
`type-size-body-sm`. Two steps, because the flow has two steps inside wimm
before the hand-off. Choosing who sees an account is not one of them: the
specification says it is reachable at any time and is not a step a member passes
through, so counting it would promise a flow that does not end where it says.

**Step actions** (`o1elL`, `prZYD`) — 36 high at Wide, Medium and Ultra; 44 high
and full width at Compact. **The lesser action comes first and the primary last**,
which is the order `Dialog`'s own footer already uses and the opposite of what
three of the four connect steps do today. At Compact the pair stacks and the
primary goes on top, because a column reverses which end is nearest the thumb.

**Month scrubber** (`W8ig0`) — 744 × 44 in a 760 slot. Chips 28 high. The strip
clips and scrolls horizontally; it does not wrap, because flex here has no wrap
and a wrapped strip would change the ledger's height as the member paged. It is
**the one place in the product where movement carries meaning**: the strip slides
to bring the chosen month under the pointer so the member can see which way they
travelled. Under reduced motion it jumps. Newest is hidden on the newest page and
Oldest on the oldest, rather than disabled, matching `Seek pager`'s own rule that
nothing is offered that would do nothing.

**Density control** (`Yt1rq`) — 36 high. **Absent, not disabled, under
`any-pointer: coarse`.** The token set already reverts density there; the control
is the half that was missing. The choice has to reach the document element before
first paint, by the same inline script that already applies the stored theme —
a component would run after first paint and the product would visibly rearrange.

**Account name** (`FguRX`) — 420 wide. The bank's name stays in the helper line
beneath the field. Clearing the field returns the account to the bank's name; it
does not leave it blank.

**Account choice row** (`QU4bL`) — 177 high with the owners row, up from 149.
Three accounts no longer fit above the fold at Compact, which is the reason the
level controls stack there. Ownership is a set: ticking a member adds them and
removes nobody, which is the defect the current screen has — its checkbox posts
only the current member and silently drops a joint account's other owner.

**The refusal is inline and beside the row**, not a dialog. It is immediately
recoverable and it has to name the account it is about, which a dialog over the
whole list would have to repeat.

**Nine `Finish` buttons in `02-banking.pen` carried no label override**, so every
one of them rendered `Button`'s own default, "Add account", on the chooser. The
screen says "Finish". The frames were wrong and are corrected. Nothing caught
this because none of those frames is in the canvas contract — which is the gap
the change closes.

**Five drawn screens are in no canvas contract at all.** The generator recognises
a screen by `^J\d+\.\w+ / \d+`, and the connect flow's happy-path frames are
named `01 · Overview`, `02 · Choose a bank`, `03 · What wimm will see`,
`05 · Overview` — no journey prefix — while every frame in `01-access.pen` uses
`·` where the pattern wants `/`. Sign in, Enrol a passkey, Choose a bank, What
wimm will see and Overview's default state are therefore drawn, built, and
asserted by nothing. Renaming them and extending the pattern is a task, and it
lands with the screen work rather than on its own, because the contract will
name copy the screens do not yet carry.
