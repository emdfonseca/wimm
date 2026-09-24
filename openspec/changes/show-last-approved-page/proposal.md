## Why

When a page reads "Changed since approval" on the design canvas, the only way to
see what moved is to remember what it looked like when it was approved. Nothing
stores that look: an approval holds a fingerprint of the markup, and a
fingerprint cannot be drawn. So a person re-checks the whole page from memory,
which is the chore approvals were meant to remove.

## What Changes

- `just approve` also takes a picture of the page story it approves, at every
  size the story is drawn at, in the light theme at comfortable density. The
  pictures are taken in the same run that settles the implemented version, so
  they show exactly the version approved.
- Only the last approved version keeps pictures. Approving a story replaces its
  earlier pictures; nothing older is kept.
- On an artboard that reads "Changed since approval", viewed light and
  comfortable, a control switches the artboard between the page as implemented
  and the picture of the last approved version.
- A story approved before this change has no picture. Its artboard says so
  instead of offering the switch, until someone approves it again.
- `just check apps/storybook` refuses a picture that does not belong to its
  story's last approved version.
- Declining the approve prompt, or any refusal, writes no picture.
- `just approve` takes several stories, or `--needs` for every page story that
  needs approval, and approves them in one run with one answer. Any story
  refused refuses the whole run.

## Capabilities

### New Capabilities

- `design-canvas/last-approved-look`: pictures of a page story's last approved
  version, taken on approval, and the switch on the canvas between that picture
  and the page as implemented.

### Modified Capabilities

- `design-canvas/page-approvals`: the approve command takes several stories or
  `--needs`, regenerates once and asks once, and refuses the whole run when any
  story is refused.

## Impact

- `apps/storybook/scripts/approvals.mjs`, `apps/storybook/.storybook/preview.ts`,
  `apps/storybook/vitest.config.ts`: capture during the approve run.
- New folder `apps/storybook/canvas/approved/`, committed PNG files. The first
  binary images tracked in the repository.
- `apps/storybook/canvas/canvas.js`, `approvals-ui.js`, `lib.js`, `canvas.css`:
  the switch.
- `CLAUDE.md`: the agent rule covers the pictures as well as the approvals
  record.
- New ADR amending 0027: an approval now also leaves pictures of the version
  approved.
- Depends on `version-and-approve-page-stories` being applied first.
