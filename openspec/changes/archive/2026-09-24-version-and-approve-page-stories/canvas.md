This change adds no screen, flow, state or component to wimm. What a person sees
and operates is the design canvas itself, `apps/storybook/canvas/index.html`, so
this is a state plan for that page. It bends the template in one way, stated
once: the canvas is plain JavaScript with no Svelte and no Storybook stories of
its own, so there are no state stories and no play functions here. Its existing
pattern is pure functions in `lib.js` held by node unit tests, plus one browser
test, and this plan follows it. No Svelte component is invented for it.

## Screens

- `apps/storybook/canvas/index.html` with `canvas.js`, `canvas.css`, `lib.js`.
  Exists. Gains a status on each artboard's label, a history control and panel,
  a "Needs approval" filter in the sidebar, a mark on each listed story, and a
  count on each flow.
- `apps/storybook/canvas/approvals-ui.js`. New. The DOM for the badge and the
  history panel, exported as functions that take a document and data and return
  elements, with no work done on import, so a browser test can mount them.
  `canvas.js` runs `main()` on import and cannot be mounted in a test.

## State stories

There are none, for the reason above. Each state is instead a **fixture pair**
under `apps/storybook/canvas/fixtures/<name>/` (`versions.json`,
`approvals.jsonl`), which the canvas reads in place of the real files when
opened with `?data=<name>`. That is how a state is put in front of a person
without approving or changing a real page. The same fixtures feed the unit
tests, so what is looked at and what is asserted are one input.

| State | Fixture | Held by | Scenario |
| --- | --- | --- | --- |
| Approved, one version | `approved` | `statusOf`, `badgeWords` unit tests | The approved version is the implemented one |
| Approved by two | `approved-by-two` | `statusOf`, `badgeWords` | A second household member approves |
| Changed, both versions | `changed` | `statusOf`, `badgeWords` | The approved version and the implemented one differ |
| Never approved, implemented only | `never` | `statusOf`, `badgeWords` | A page nobody approved |
| Exempt | `approved` (its behaviour stories) | `statusOf` | A story that only asserts something |
| No versions | `?data=missing` | `readData` unit test, browser test | The canvas before any versions exist |
| History, approved twice | `history` | `historyOf` unit test, browser test | A page approved twice over its life |
| History, one entry both marks | `approved` | `historyOf`, browser test | The implemented version is the latest approved |
| History, none | `never` | `historyOf`, browser test | A page with no history |
| History, long | `history-long` (14 approved versions) | browser test | A long history |
| Filter on | `mixed` | `needsApproval` unit test | Filtering to what needs a look |
| Filter on, nothing to show | `approved` | `needsApproval`, `filterWords` | Nothing needs a look |
| Filter and find together | `mixed` | `needsApproval` with a query | Filtering and finding together |
| Flow, part approved | `mixed` | `flowStatus`, `flowWords` | A flow with pages still to approve |
| Flow, none changed | `never-some` | `flowStatus`, `flowWords` | Nothing changed, some never approved |
| Flow, all approved | `approved` | `flowStatus`, `flowWords` | A fully approved flow |

## Words

Dates are written `20 Sep 2026`. Names are the first word of the recorded name.
No em or en dash anywhere.

**A version, wherever one is written:** seven characters, a middle dot, the
date, `a3f9c21 · 20 Sep 2026`. For the implemented version the date is when it
was first seen; for the approved version it is when it was last approved.

**Badge.** Both versions, approved and implemented, one line each. One line when
they are the same version.

- Approved, the two are one version: `Approved a3f9c21 · 20 Sep 2026 by Emanuel`
- Approved by two: `Approved a3f9c21 · 20 Sep 2026 by Emanuel and Grace`, the
  date being the later approval, with each date in the history
- Approved by three or more: `... by Emanuel, Grace and 1 other`
- Changed, the two differ, three lines: `Changed since approval`, then
  `Approved 77b0d4a · 12 Sep 2026 by Emanuel`, then
  `Implemented a3f9c21 · 20 Sep 2026`
- Never, two lines: `Never approved`, then `Implemented a3f9c21 · 20 Sep 2026`
- Exempt: `Implemented a3f9c21 · 20 Sep 2026` and no status.

Each badge carries a shape as well as a colour: a tick, a dot, an empty ring.
The colours are `color-feedback-success` with `color-feedback-success-bg`,
`color-feedback-warning` with `color-feedback-warning-bg`, and
`color-text-secondary`, from `packages/ui/design/tokens.json`.

**History control:** a button on the label reading `History`. Accessible name
`History of <story name>`.

**History panel**

- Heading: `<Screen> · <story name>`
- Each entry: `a3f9c21 · first seen 20 Sep 2026`, its marks, then one line per
  approval, `Emanuel, 20 Sep 2026`, and the note beneath if there is one
- Marks: `Implemented now` on the first entry; `Latest approved` on the newest
  entry with an approval; both on one entry when they are one version, written
  `Implemented now · Latest approved`
- First entry with no approval: `Implemented now, not approved`
- Earlier approved versions past the latest: no mark
- No approvals at all: `Nobody has approved this page yet.`
- To approve: `To approve this version, run this in your own terminal:` then
  `just approve pages-overview--populated` with a `Copy` button. After copying
  the status bar reads `Copied.`
- Close: a button reading `Close`

**Sidebar**

- Filter: a checkbox labelled `Needs approval` followed by the count, `Needs approval (23)`
- Mark beside a listed story: `changed` or `never`, as text after the name
- Nothing to show: `Every page is approved at its implemented version.`
- Nothing matches filter and find together: `No page needing approval matches that.`

**Flows**

- Some changed, some never: `7 of 12 approved · 3 changed since approval · 2 never approved`
- A part that counts none is left out: `10 of 12 approved · 2 never approved`
- All approved: `Approved`
- A flow with nothing that needs approval: nothing is shown

**Status bar**

- Files unreadable: `Versions could not be read. Run just gen apps/storybook.`
- Approvals unreadable while versions read: `Approvals could not be read. Statuses are hidden.`

**Command line**, since a person reads these too:

- Confirmation: `Approve Pages/Overview · Populated at a3f9c21, first seen 20 Sep 2026? yes/no`
- Recorded: `Recorded. Emanuel approved pages-overview--populated at a3f9c21.` then
  `Commit apps/storybook/canvas/approvals.jsonl on its own.`
- Declined: `Nothing recorded.`
- Agent or no terminal: `Approvals are recorded by a person in their own terminal. Nothing recorded.`
- Unknown story: `No page story "pages-overveiw--populated". Closest: pages-overview--populated.`
- No version yet: `pages-x--y has no version yet. Run just gen apps/storybook, then approve.`
- The page changed since they looked: `pages-x--y now renders a3f9c21, first seen today. Look again, then approve. Nothing recorded.`
- Already approved: `You approved this version on 20 Sep 2026.`
- Behaviour story: `pages-x--y only asserts something and needs no approval.`
- No git identity: `git has no user.name or user.email set. Nothing recorded.`

## Flow

Stands alone. No state here is a page story, so nothing joins a flow in
`apps/storybook/canvas/flows.js`, and nothing is added to its `unplaced` list.

## Surfaces

- Badge and version: inline, on the artboard's label, outside the iframe.
- History: one `<dialog>` owned by the canvas page, opened non-modally and
  anchored to the artboard, ephemeral, not route-backed. Non-modal because a
  person reads the history while still looking at the page it describes.
- Filter: inline in the sidebar, under the find box. Its state is kept in the
  address as `?needs=1` so a reload and a shared link keep it.
- Approving: no surface on the canvas. The panel offers the command.

## Components used

None from `packages/ui/src`. The canvas page imports no Svelte component and
this change keeps it that way; it is styled from the same tokens through
`/canvas-ui/base.css`, as it is today.

## Components missing

None in `packages/ui/src`, because nothing here is composed from it. Read before
deciding that, searching for where the design system shows a status with a tone:
`atoms/Notice.svelte` and `molecules/InfoNotice.svelte` are the two that do, and
either could carry these words. There is no badge or tag component. Instancing a
Svelte component inside a page that has no Svelte runtime means giving the canvas
a build step, which the canvas was written to avoid, and a notice is a block of
prose where this is one line on a label. The badge is a few lines of DOM in
`approvals-ui.js` using the same feedback tokens `Notice` does. If a second
consumer ever appears, that is the moment to reconsider.

## States left out

- **Approving from the canvas.** A static page cannot write a file.
- **A version per size, theme or density.** One version covers a story. See
  design.md, decision 2.
- **A removed story's history.** It is kept in the files and drawn nowhere, since
  there is no artboard to hang it on.
- **What the approved version looked like.** It would need an image stored with
  each approval. A possible later change.
- **A diff between two versions.** The history says that a page changed and when
  it was last agreed; what changed is `git log` on the screen's files.
- **Progress while the files load.** They are two small local files read once.

## Contracts for implementation

Held by `apps/storybook/canvas/approvals.browser.test.js`, which mounts the
functions of `approvals-ui.js` against a fixture:

- The history button is a `<button>`, reachable with Tab from the sidebar without
  holding Alt, and its accessible name includes the story's name.
- Activating it opens the panel and moves focus to the panel's heading. Escape
  and `Close` both close it and return focus to the button that opened it.
- Opening a second story's history closes the first. There is never more than
  one panel.
- While the panel is open and focus is inside it, the canvas's single-key
  shortcuts (`0`, `1`, `m`, `/`) do nothing, as they already do nothing while
  the find box has focus.
- The badge's status is in its text. A test reads the text with the stylesheet
  removed and still finds it.
- The count in `Needs approval (n)` is a live region, polite, so turning the
  filter on is announced once.
- `Copy` puts exactly the command on the clipboard, with no prompt character and
  no trailing newline.
- With `?data=missing` the page draws every artboard and no badge, and the status
  bar carries the unreadable message.

Held by node unit tests in `lib.test.js`: every string under Words that a
function returns, against the fixtures named in the table above.

## Seen

Looked at in Chrome against the running Storybook at two widths: a window 1440
wide, and the canvas in a 600-wide frame, since the Chrome window would not
narrow. The four regimes are the artboards' own sizes, which this change does
not alter. The canvas chrome has one theme; light and dark are the artboards'
`theme`.

Screen view is Pages/Overview; flow view is See where the money is unless named.
The flow views of `changed`, `approved-by-two`, `history`, `history-long` and
`missing` were opened and their words read from the page rather than looked at.

- `approved`: screen, light, and flow, dark. Every badge one line with a tick,
  every flow reads `Approved`, `Needs approval (0)`.
- `approved-by-two`: screen and flow. `Approved 964f836 · 21 Sep 2026 by
  Emanuel and Grace`; the filter on shows the everything-approved sentence and
  puts `needs=1` in the address.
- `changed`: screen, and flow reading `0 of 14 approved · 14 changed since
  approval`. Three lines with a dot, History on every label.
- `never`: screen, and flow Enrol a passkey reading `0 of 6 approved · 6 never
  approved`. Two lines with a ring.
- `never-some`: screen, sidebar open and closed, and flow Sign in reading `5 of
  6 approved · 1 never approved`.
- `mixed`: screen, light, measured with every artboard level whatever its
  badge's length, and flow with the filter on, reading `9 of 14
  approved · 3 changed since approval · 2 never approved`.
- `history`: screen, dark, the panel listing three versions with marks,
  approvers, notes, the command, Copy and Close; flow reading `0 of 14 approved
  · 1 changed since approval · 13 never approved`.
- `history-long`: screen, dark. The panel stays inside the window and scrolls;
  the command, Copy and Close are reached by scrolling it. Flow as `history`.
- `missing`: screen and flow. Every artboard draws, no badge, no filter, and
  the status bar reads the unreadable sentence.

At 600 wide:

- `approved`: screen. Tick badge, every flow `Approved`, `Needs approval (0)`.
- `approved-by-two`: flow. Tick badge `Approved 964f836 · 21 Sep 2026 by ...`,
  running past the frame's edge as every label does.
- `changed`: flow. Three-line badge with a dot; each sidebar flow count wraps
  to two lines and stays readable.
- `never`: screen, sidebar closed. Two-line badge with a ring.
- `never-some`: flow Sign in reading `5 of 6 approved · 1 never approved`.
- `mixed`: screen and flow with the filter on, `Needs approval (30)`.
- `history`: screen, dark, the panel open with three versions, notes, the
  command, Copy and Close, all inside the frame.
- `history-long`: screen, the panel inside the frame and scrolling.
- `missing`: screen. The status bar wraps to three lines and the find box sits
  below it.

Changed after looking:

- A one-line or two-line badge filled three lines of tint. The badge still
  takes three lines of room so artboards in a row stay level, but the tint now
  covers only its lines (`canvas.css`).
- Every sidebar link dropped `data` and `needs`, so leaving a screen for a flow
  left the fixture and the filter. `screenUrl` and `flowUrl` carry both, and
  turning the filter on or off readdresses the sidebar's links (`lib.js`,
  `canvas.js`, one unit test).
- At 600 wide the status bar wraps after load and covered the top of the find
  box, because its height was measured once. It is now watched with a
  `ResizeObserver` (`canvas.js`).
- At 600 wide the history panel opened at its button's left edge, which is past
  the window's right edge on a long label, so the panel was off screen. It is
  now kept inside the window (`approvals-ui.js`, one browser test).
