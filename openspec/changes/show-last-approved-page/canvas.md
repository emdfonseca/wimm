This change adds no wimm screen. What a person sees is the design canvas,
`apps/storybook/canvas/index.html`, and the approve command's output. As in
`version-and-approve-page-stories`, the canvas is plain JavaScript with no
Svelte and no stories of its own, so each state is a fixture pair opened with
`?data=<name>`, held by unit tests in `lib.test.js` and the browser test in
`approvals.browser.test.js`.

## Screens

- `apps/storybook/canvas/index.html` with `canvas.js`, `canvas.css`, `lib.js`.
  Exists. A changed artboard gains the `Approved look` toggle and one badge
  line.
- `apps/storybook/canvas/approvals-ui.js`. Exists. Gains the toggle and the
  picture element as functions taking a document and data, like the badge.

## State stories

None, for the reason above. Each state is a fixture under
`apps/storybook/canvas/fixtures/<name>/`.

| State | Fixture | Held by | Scenario |
| --- | --- | --- | --- |
| Changed, picture kept | `changed-pictured` (new: `changed` plus `approved/` pictures at compact and wide for Pages/Overview · Populated) | browser test | Switching to the approved look |
| Changed, switched back | `changed-pictured` | browser test | Switching back |
| Changed, no picture | `changed` | `pictureLine` unit test, browser test | Approved before pictures were kept |
| Changed, dark | `changed-pictured` with `theme=dark` | `offersApprovedLook` unit test | Viewed in dark |
| Changed, compact density | `changed-pictured` with `density=compact` | `offersApprovedLook` unit test | Viewed in dark (same rule) |
| Approved | `approved` | `offersApprovedLook` unit test | A page that has not changed |
| Never approved | `never` | `offersApprovedLook` unit test | A page that has not changed |

## Words

No em or en dash anywhere.

**Toggle:** a button on the label, after `History`, reading `Approved look`.
Accessible name `Approved look of <story name> at <size>`. Pressed state is
`aria-pressed`, the words do not change.

**Badge, while the picture shows:** one more line after the badge's lines,
`Showing approved 77b0d4a`.

**Badge, changed with no picture:** one more line, `No picture of the approved
version`.

**Picture:** `alt` reads `<Screen> · <story name> at <size>, approved 77b0d4a`.

**Command line, on recording**, replacing the two lines it prints today:

- `Recorded. Emanuel approved pages-overview--populated at a3f9c21.`
- `Kept a picture at compact, medium, wide and ultra.`
- `Commit apps/storybook/canvas/approvals.jsonl and apps/storybook/canvas/approved/pages-overview--populated/ together, on their own.`

A story drawn at one size reads `Kept a picture at compact.`

**Command, approving several**, before asking, one line per story, then the
question:

- `Pages/Overview · Populated at a3f9c21, first seen 20 Sep 2026`
- `Approve these 3 page stories? yes/no`

On recording:

- `Recorded. Emanuel approved 3 page stories.`
- `Kept pictures of each at the sizes it is drawn at.`
- `Commit apps/storybook/canvas/approvals.jsonl and apps/storybook/canvas/approved/ together, on their own.`

**Command, `--needs` with nothing to approve:** `Nothing needs approval. Nothing recorded.`

**Command, named stories and `--needs` together:** `Name stories or use --needs, not both. Nothing recorded.`

A refused run of several prints one line per refused story, then `Nothing recorded.`

**Check failures**, one line per file:

- `apps/storybook/canvas/approved/pages-overview--populated/77b0d4a…-wide.png is not of the last approved version, a3f9c21.`
- `apps/storybook/canvas/approved/pages-x--y/… belongs to no approved story.`
- `apps/storybook/canvas/approved/pages-overview--populated/…-tablet.png names a size the canvas does not have.`
- `apps/storybook/canvas/approved/pages-overview--populated.next/ is left from an approval that did not finish. Delete it.`

## Flow

Stands alone. Nothing joins `apps/storybook/canvas/flows.js`.

## Surfaces

- Toggle and badge line: inline, on the artboard's label, outside the iframe.
- Picture: in place of the artboard's iframe, same width, its own height.
  Ephemeral; not kept in the address.
- Pictures are kept by the approve command in a terminal. No canvas surface.

## Components used

None from `packages/ui/src`. The canvas imports no Svelte component, as in
`version-and-approve-page-stories`.

## Components missing

None in `packages/ui/src`. Read before deciding: `atoms/Button.svelte` has no
pressed state, and `atoms/SegmentedControl.svelte` switches between two named
views, which fits, but the canvas cannot instance a Svelte component without a
build step it was written to avoid. The toggle is a
`<button>` in `approvals-ui.js` styled like History from the same tokens.

## States left out

- **Dark theme and compact density pictures.** The person chose light and
  comfortable only; those artboards offer no control and no line.
- **Switching every artboard at once.** One artboard at a time was asked for.
- **Pictures of earlier approved versions.** Only the last is kept.
- **Loading the picture.** One local file; the artboard's existing
  "Waiting to render" covers a slow image.

## Contracts for implementation

Held by `approvals.browser.test.js` against `changed-pictured`:

- `Approved look` is a `<button>` with `aria-pressed="false"`, reachable with
  Tab, and its accessible name names the story and size.
- Activating it sets `aria-pressed="true"`, hides the iframe without removing
  it, shows an `<img>` whose `alt` is the words above, and adds the
  `Showing approved` line. Focus stays on the button.
- Activating it again restores the iframe, removes the `<img>` and the line,
  and the iframe was not reloaded (its `src` and document are the same object).
- With `?data=changed` no `Approved look` button exists and the badge reads the
  no-picture line.
- With `?theme=dark` on `changed-pictured`, no `Approved look` button and no
  no-picture line exist.

## Seen

Not looked at yet.
