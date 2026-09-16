## Journey

- **J01 · Enrol a passkey** — a registered member turns an enrolment link into a passkey and a signed-in session. Stories S2.
- **J02 · Sign in** — a member with an enrolled passkey gets back in. Stories S3.

Both are new IDs, allocated here and never reused.

**S1 records a deliberate skip.** The operator's surface is `wimmctl`, a terminal.
It has no pixels to design; its contract is the scenarios in
`identity/operator-registration`, and a frame invented for it would be a drawing
nobody builds. Noted on the canvas in zone 30 as well, so a reviewer reading only
the `.pen` sees the same reasoning.

## File

`apps/web/design/01-access.pen`, importing `packages/ui/design/product-ui.lib.pen`
as `ui`.

The journey lives under `apps/web/` because a journey is the app's flow and the
app owns it; `packages/ui` holds the design system it instances. ADR 0001, 0008
and 0010 were corrected to say so, and `bin/pen-import` now accepts a library
anywhere in the repository rather than only beside the journey — the written
import is `../../../packages/ui/design/product-ui.lib.pen`, which pen resolves.

## Saved to disk

Saved. `apps/web/design/01-access.pen` (217600 bytes) and
`packages/ui/design/product-ui.lib.pen`, which carries the four page shells, the
two shell presets and the header-height fix. Every mutation ran through `just
pen-exec`, which saves on a clean run and writes nothing on an error.

No source frame changed during implementation, so the zone 40 dark fixtures were
not re-exported. All 23 frame IDs and all 6 zone IDs in this document were
resolved against the journey file after the work, so the table above still
points at something.

## Frames

| Frame | ID | Zone |
| --- | --- | --- |
| J01.A / 01 · Enrolment invitation / Compact / Default   | `nycwW` | 10 |
| J01.A / 01 · Enrolment invitation / Wide / Default   | `IC4AQ` | 10 |
| J01.A / 01 · Enrolment invitation / Wide / Passkey not saved   | `hOVdC` | 10 |
| J01.A / 01 · Enrolment invitation / Wide / Prompt dismissed   | `h1NPCE` | 10 |
| J01.A / 02 · Creating passkey / Wide / Waiting   | `UWQr3` | 10 |
| J01.A / 03 · Signed in / Compact / Default   | `v8bd7N` | 10 |
| J01.A / 03 · Signed in / Ultra / Default   | `On6yE` | 10 |
| J01.A / 03 · Signed in / Wide / Default   | `c3N3z` | 10 |
| J02.A / 01 · Sign in / Compact / Default   | `kX5Hn` | 10 |
| J02.A / 01 · Sign in / Wide / Default   | `h79SR` | 10 |
| J02.A / 01 · Sign in / Wide / Prompt dismissed   | `xd2g8` | 10 |
| J02.A / 01 · Sign in / Wide / Session expired   | `hsnpo` | 10 |
| J02.A / 02 · Signed in / Compact / Default   | `uQFxV` | 10 |
| J02.A / 02 · Signed in / Wide / Default   | `CErpj` | 10 |
| J01.B / 01 · Link unusable / Compact / Default   | `CRAE1` | 20 |
| J01.B / 01 · Link unusable / Wide / Default   | `d4i2f` | 20 |
| J02.B / 01 · Passkey not recognised / Compact / Default   | `I9oOq6` | 20 |
| J02.B / 01 · Passkey not recognised / Wide / Default   | `lthfY` | 20 |
| J01.A / 01 · Enrolment invitation / Wide / Default [QA · Dark]   | `GNXrg` | 40 |
| J01.A / 01 · Enrolment invitation / Wide / Passkey not saved [QA · Dark]   | `s3ejQ` | 40 |
| J01.A / 03 · Signed in / Wide / Default [QA · Dark]   | `PfG8y` | 40 |
| J01.B / 01 · Link unusable / Wide / Default [QA · Dark]   | `mKm0m` | 40 |
| J02.B / 01 · Passkey not recognised / Wide / Default [QA · Dark] | `mhr0E` | 40 |
Every screen is an instance of a library origin inside an identical **plate**:

```text
44  title bar      frame name, and its regime on the right
40  trigger strip  "Reached when: …" — what puts the member on this screen
    the screen     one border around all three
```

The trigger strip exists because a state with no stated cause is a picture, not a
specification. Every plate carries one, including the happy path.

**Canvas layout is a fixed scale, not a judgement per row.**

```text
48   between states inside one step     alternatives at one point
24   between a step and a transition    176 for the transition column itself
96   between regime rows
160  between path groups
64   zone padding
```

**Transitions carry their trigger.** An arrow appears between every consecutive
pair of steps and is labelled with the action that causes the move — "Member
selects Create a passkey. The browser's own prompt opens." Arrows never appear
between states, which are alternatives at one point rather than a sequence, and
never in J01.B, which has one step. The label is why the transition column is
176 wide rather than 64: the gap carries information instead of air.

Zone frames: `osQRM` 00, `WOsl8` 10, `b523Rz` 20, `aHF93` 30, `buA3J` 40,
`QtODz` 90.

## Surfaces

| Surface | Modality | Routing |
| --- | --- | --- |
| Enrolment invitation | Full page | Route-backed, `/enrol/<link>`. Deep-linked by definition — the link is the route. |
| Link unusable | Full page | Route-backed, same path. A dead end, not an overlay: there is nothing behind it to return to. |
| Sign in | Full page | Route-backed, `/signin`. |
| Signed-in landing | Full page inside the app shell | Route-backed, `/`. |

No modals, drawers or inline edits anywhere in either journey. Every step is a
destination reached from outside the app, by someone with no session, and an
overlay needs a page underneath it to overlay.

`layout-content-max` does not apply: these are single-card pages with no prose
measure to protect (ADR 0005).

## Components used

| Component | Origin | Used for |
| --- | --- | --- |
| **Auth shell** | `kJkV1` | Every unauthenticated page at Wide and Ultra |
| **Auth shell / compact** | `tqiBA` | The same pages at Compact |
| **Signed-in landing** | `IGbQe` | The signed-in page at Wide and Ultra |
| **Signed-in landing / compact** | `pu6qZ` | The same page at Compact |
| Sidebar nav / in shell | `q7dY7c` | The landing's navigation, brand and account |
| Brand | `vGjfy` | Inside the auth shells, inverted on the gradient panel |
| Button | `W2gOKx` | The single action in the card |
| Notice | `WFSg5` | The card's feedback slot |
| Empty state | `w5ZouR` | The landing's body |

**The four page shells were built into `product-ui.lib.pen` as part of this
change**, under `50 · TEMPLATES · Page shells`. The journey instances them; it
draws no page chrome of its own.

**The card has two feedback slots, and which one a message uses is a question
about the message, not about the screen.**

```text
Feedback slot  LqoqA (Wide), Szvqr (Compact), above the heading. Unchanged.
               Why the member is here this time: "You were signed out"

Result slot    ft9RY (Wide), n7iV7P (Compact), between the body and the action
               What happened when they last pressed it: "Your device did not
               save the passkey", "Nothing was created", "Nothing happened",
               "That passkey is not recognised"
```

`Feedback slot` keeps its name. Renaming it to match its narrower job would
have been tidying an existing node while adding a new one, which
`just pen-manifest verify` now refuses.

**`50 · TEMPLATES` holds complete screens and nothing else.** The test is "could
someone screenshot this and call it a screen", not "is it page-level". Three
kinds of thing were in there that failed it, and all three moved to
`40 · ORGANISMS`: the two shell presets, and the `List` and `Detail` scaffolds,
which are drawn at 1176 - the Main column of a 1440 shell, so a region rather
than a screen. `design-system-library.md` said scaffolds belonged in 50 and has
been corrected.

**The four new shells were missing the zone's chrome entirely**: no
`cornerRadius`, no `stroke`, no `strokeWidth`, no `strokeAlignment`, and - the
one that mattered - no `theme` binding. Every other frame in 50 carries
`theme: {device}` matching its width, and without it a shell resolves every
device-axis token at the wrong regime, so `layout-sidebar-width` and its
neighbours are wrong inside a frame that looks correct. They now carry
`wide` at 1440 and `compact` at 390.

**The two shell presets moved out of `50 · TEMPLATES` into `40 · ORGANISMS`.**
`Sidebar nav / in shell` and `App header / in shell` are presets of organisms,
and the zoning the library follows gives TEMPLATES the app shell and the page
scaffolds only. They now sit in their own components' specimen blocks, where
library-method puts a preset. The `Shell presets` wrapper that held them under
TEMPLATES is gone, and `50 · TEMPLATES · Page shells` holds the four page
shells and nothing else.

A single slot above the heading was the first drawing, and it was wrong for
three of the four messages. A result belongs beside the control that produced
it: the member pressed a button at the foot of the card, and answering at the
head of it means scrolling past two paragraphs they have already read. It also
displaced the heading, which is the identity check
`identity/passkey-enrolment` requires them to see *before* being asked to
create anything - so the one state where something had gone wrong was the one
state that hid who the link was for.

The error-summary pattern that does sit above the heading exists for forms with
many fields, where the summary links to each failure. This card has no fields
and one control.

Announcement is unchanged by the move: errors assertive, info polite, neither
taking focus, because the member is returning from a native prompt.

**The auth shell is a split page, not a card on empty canvas.** A brand panel of
560 carries the lockup, the one-line reason a passkey replaces a password, three
supporting points and a footer, on the `gradient-brand-from → gradient-brand-to`
ramp at rotation 160 — the same ramp the balance card uses. Text on it is
`color-text-on-brand` and `color-text-on-brand-secondary`, which exist for exactly
this and had no user until now. The brand mark is inverted there: a white tile
with the glyph in `color-action-primary`, because the stock mark is brand green on
a brand-green ground. At Compact the panel becomes a 200 band above the card,
since a 390 viewport has no room for a column beside anything.

**The landing composes the library's existing shell, and adds nothing to it.** An
earlier version of this canvas introduced an `App header / plain` organism holding
the brand and the signed-in member. That was a duplicate: `Sidebar nav` already
carries the brand at its head and an `Account trigger` in its footer, which is
where this design system puts both. The lesson is in the search, not the fix — the
component list gives names and ids and says nothing about contents, so "the
library has no top bar" was true and irrelevant. The right question was where the
system already shows a signed-in person. `App header / plain` has been deleted.

**The landing carries no page header.** `App header` is a vertical stack of
breadcrumbs, a title row and tabs, sized by its content, with the title at
`type-size-page-title` — 32 at Wide. On this page all three regions are empty
except the title, so the component rendered one word at 32 in a box built for
three rows, repeating what the sidebar already marks as current. A header with
nothing to carry is not a smaller header, it is no header. It returns when the
page gains breadcrumbs, tabs or an action.

**Shell parts drop their card chrome through a preset, not a repeated override.**
`Sidebar nav` and `App header` each define a full border and a corner radius,
right on the tile they are displayed on and wrong against another shell part.
`Sidebar nav / in shell` and `App header / in shell` wrap an instance and trade
that chrome for the single edge a shell needs. See Library findings.

**Overview is marked as the current item**, using the pattern the library already
defines on its own sidebar: `color-accent-subtle` fill with the icon and label in
`color-accent` at weight 600. A page whose nav item is not marked is a page that
does not say where you are.

**The landing shows only what exists.** `Sidebar nav` with Overview enabled and
Accounts, Transactions, Budgets, Reports and Settings disabled, because those
screens do not exist; `App header` with breadcrumbs, tabs, the nav trigger and the
primary action disabled, because none of them has anything to carry. Enabling them
as the product grows is a change to an instance, not a new component.

**Interface copy carries no em dashes and no marketing voice.** Every sentence a
member reads was rewritten to use a full stop or a comma instead. The validator
checks it: an em dash, an en dash, or any of "seamless", "effortless", "elevate",
"unlock the power", "dive in", "robust" inside a screen instance fails the run.
The check was proven by putting an em dash back into one sign-in body, confirming
it was flagged, and removing it.

**The signed-in landing carries no navigation, deliberately.** An earlier version
instanced the library's App header and Sidebar nav, which bring Accounts,
Transactions, Budgets, Reports, breadcrumbs and an Activity/Pending/Uncategorised
tab strip. None of those exist. A canvas describes the product as it will be when
this change ships and nothing further, so the landing is a header carrying the
member's identity and an empty body. Navigation arrives with the first thing to
navigate to.

## Components missing

Each is three deliverables: the origin in `product-ui.lib.pen`, the Svelte
component under `packages/ui/src/<layer>/`, and its `.stories.svelte` beside it.
A component with no story cannot be seen in isolation, in either theme, at any
viewport, and its behavioural contract is asserted nowhere. `just check
packages/ui` refuses a component that has no story, and the schema now says the
same so it is planned rather than remembered.

1. **Button · pending preset** — a wrapper around a Button instance for "an action
   is underway and the browser has taken over": non-interactive, its own label,
   and an accessible status. It redraws none of Button's chrome, so it is a
   preset, not a component. Faked on `UWQr3` today with a disabled fill and a
   substituted label, which is a look without a behaviour.
2. **Notice · tone presets (info, error)** — Notice carries no tonal variant, so
   every use here overrides `fill` and `stroke` inline against `color-feedback-*`.
   Two presets wrapping an instance, not two components.

## States drawn

| Path | Step | States |
| --- | --- | --- |
| J01.A | 01 · Enrolment invitation | Default, Passkey not saved, Prompt dismissed |
| J01.A | 02 · Creating passkey | Waiting |
| J01.A | 03 · Signed in | Default |
| J01.B | 01 · Link unusable | Default — expired, spent, replaced and never-issued all resolve here |
| J02.A | 01 · Sign in | Default, Prompt dismissed, Passkey not recognised, Session expired |

**Branch versus local state.** J01.B is the only branch, because it is the only
place the goal cannot be completed and the recovery leaves the product entirely.
Everything else keeps the member on one step with an immediate retry, so it sits
beside that step rather than forking the path.

**Deliberately not drawn:**

- **The browser's passkey prompt.** Native chrome. Nothing on the canvas styles,
  positions or controls it. `pEQxs` shows only what wimm renders while it is up.
- **Medium and Ultra.** Both auth pages are one centred card on empty canvas, so
  a wider viewport changes nothing structurally, and the only screen with a
  sidebar is already drawn at Wide. Recorded as a note in zone 90 rather than as
  four more frames of the same thing.
- **A loading state for the landing.** There is no data behind it yet; a skeleton
  would be drawing a promise this change does not make.
- **J02.A / 02 · Signed in.** It is J01.A / 03, not a copy of it. A second frame
  would be a second thing to keep in step.

## Contracts for implementation

The canvas states intent; none of the following is executed by it.

**Measured on the canvas.** Wide frames 1440 × 900, Compact 390 × 844, verified
from resolved bounds across all 18 frames. Card measure 480 at Wide, 342 at
Compact, inside 24 px page padding — the Compact card clears the 320 CSS px
reflow floor with the page gutter intact. Button origin height is 36 (read from
`W2gOKx`), which clears the 24 px AA target floor of SC 2.5.8; nothing in these
journeys reduces it.

**Focus.** On load, focus moves to the card's heading, not to the action — the
member must read who the page thinks they are before committing. The action is
the first tab stop after it. Every journey here is completable with one key.

**Announcement.** The error notice on "Passkey not saved" and on "Passkey not
recognised" is assertive; the info notice on "Prompt dismissed" and "You were
signed out" is polite. Neither moves focus: the member is returning from a native
prompt and a focus jump loses their place. The pending state announces that the
browser has taken over, and reverts when the ceremony resolves either way.

**Routing.** The four unusable-link causes render one identical response, so
nothing about the outcome can be inferred from what is shown, from the status
code, or from response timing. Sign-in after an expired session returns the
member to the path they were trying to reach; that path is held server-side
against the attempt and never appears in the URL.

**The dark frames in zone 40 are fixtures, not specifications.** Each names its
source frame. They exist so the `color` axis is reviewed on the four screens that
carry feedback colour, and they are re-copied when their source changes.

**Verified by rendering, not by assertion alone.** Every frame was exported and
read back at 1:1. Zero zone overlaps and zero clipping inside any frame, measured
with fonts loaded — an earlier pass reported the same numbers with the font host
unreachable, and was wrong by up to 459 px on a single zone because unfetched
fonts collapse every text-driven height. `bin/pen-exec` now fails closed when an
asset does not load, so a measurement taken in that state cannot be mistaken for
a clean one.

**No frame shows origin placeholder text.** All 20 instances were checked against
the shells' default slot content, ignoring disabled nodes; the check was proven to
fail by restoring one heading to its placeholder and confirming it was flagged.
This matters because `descendants` overrides on an imported component are dropped
**silently** when the keys are not alias-qualified — `ui:lfPvg`, not `lfPvg`, and
every segment of a nested path. A dropped map produces a plausible-looking frame
carrying the library's sample copy, with no error anywhere.

**The brand mark is `ttoEX`'s own two paths.** `Get` elides path data by
default and returns `"..."`, which reads like an empty node rather than a
withheld one; `includePathGeometry: true` returns it. Rebuilding the mark from
its bounding box instead produced a passable likeness that was visibly not the
logo. Both paths share a scale of 9.80035 against the 16 frame.

**An expired session is the page's own subject, so the heading carries it.**
The first drawing put "You were signed out" in a notice above a "Welcome back"
heading, which made the card commiserate and welcome in the same breath and
restated what the heading was for. `hsnpo` now reads "Your session ended" with
the return promise in the body and no notice at all. The two remaining sign-in
notices are results of an attempt, which is why they sit in the Result slot.

**The Compact landing had no brand and no account, and that was a defect the
drawing hid.** At Wide the sidebar carries both; Compact drops the sidebar and
the first version dropped both with it, leaving a member on a phone with a
centred empty state and no way to tell which app they were in or who they were
signed in as. The caption even described it as a virtue - "the body takes the
whole frame".

`App bar / compact` (`EpRPP`, in `40 · ORGANISMS`) carries them: the lockup on
the left, the account on the right, 56 high. It is not a page header - no title,
no breadcrumbs, no tabs - so the reasoning that deleted `App header / plain`
still holds: at Wide the sidebar owns this content and the bar is hidden, so
neither is ever a duplicate of the other.

**A ticket is only as good as the link behind it.** The exchange gives the
browser a ticket with its own 30-minute lifetime, and the first version checked
only that lifetime. So a member who had just enrolled still held a usable
ticket, and could enrol a second passkey through a link the operator had been
told was single use — the link was correctly refused, and the cookie in the
browser was not. `EnrolmentTicketByHash` now joins the link and requires it
unspent, unreplaced and unexpired, and finishing an enrolment deletes every
ticket that link produced, including on another device it was opened on.

**Sign-out is the footer's icon button, not a text link.** The canvas draws the
sidebar footer as `Account trigger` + an `Icon button` instance — 36 square, no
fill, no border, a 16 glyph in `color-text-secondary` — and that instance was
Collapse. Collapsing a sidebar with one destination does nothing, and
`identity/passkey-sign-in` requires signing out, so that slot carries `log-out`
instead. The compact bar carries the same control beside the account.

The first attempt put a text link there. It matched nothing in the library: an
icon button is what the design system already uses for a control in that
position, and inventing a second shape for it was a shape nobody had agreed to.

**Two tooling traps found here are not yet in the skill.**
`.claude/skills/pen-design/` is not writable from this environment, so they are
recorded here instead: the silent `descendants` drop above, and the fontless
measurement. Both belong in `references/tooling-traps.md`.

## Library findings

All three were raised by this change and all three are fixed in the same commit.

A fourth was raised during implementation and is fixed there rather than in the
library: the canvas records `page padding 24`, which is true at Compact and wrong
at Wide. `kJkV1`'s card area and brand panel both carry `padding: 64`, and the
signed-in landing uses neither value — `IGbQe` is `$space-8` and `pu6qZ` is
`$space-4`. The built components follow the file. Card padding is `$space-8` and
card gap `$space-6` in both auth shells; the only card geometry that differs
between Wide and Compact is the 480/342 measure.

Two other numbers in Contracts for implementation needed the same correction.
The card heading is `$type-size-heading-lg`, not heading-md. The action instances
in both shells override Button's 36 to **44**, so the drawn control is the large
step rather than the origin's medium one.

1. **`layout-header-height` had no consumer.** The token reached `tokens.css`
   (56 Compact, 64 elsewhere) and no component referenced it: `App header` hugged
   its content, so its bar height was whatever the title row happened to total.
   The title row is now 64, and 56 inside `App header/compact`, with the title
   vertically centred. The canvas holds these as literals rather than as
   `$layout-header-height` because this pen build **discards `width` and `height`
   when bound to a variable** and leaves no trace that it did. The token stays
   authoritative in code; the canvas mirrors its value.

2. **Shell parts carried specimen chrome.** Eleven instances across
   `Transactions / Wide`, `Transactions / Wide 1440 · editing`, `Transactions /
   Ultra 1920`, `Settings / Wide`, `Settings / Ultra 1920`, `Detail scaffold /
   Wide` and this journey's landing each repeated `strokeWidth {right: 1}` or
   `{bottom: 1}` with `cornerRadius: 0`. Two presets now carry it:

```text
q7dY7c  Sidebar nav / in shell    wraps fwKnX, right edge only
s2VNo8  App header / in shell     wraps uC4nV, bottom edge only
```

   Every one of the eleven was repointed at the preset with its content overrides
   preserved, re-pathed under the preset's inner instance. The base components are
   unchanged, so the specimen tiles that display them standalone keep their card
   look.

3. **The dark theme had never been checked by anything.** Every pair passed WCAG
   AA and the theme was still wrong: the brand gradient ended on exactly the page
   background, the primary button measured 12.36 against the canvas where light's
   measures 4.95, and a notice was indistinguishable from the card under it.
   Twenty-four dark values move, and `check-palette.py` now asserts 52
   relationships across both themes inside `just check packages/ui`, with the
   palette as it shipped kept as a rejection fixture. See ADR 0014.
