## Why

Storybook shows one story at a time. Judging a screen means seeing its states
and its four widths side by side, which today exists only in the `.pen`
files, a second copy of every screen that drifts from the built one and that
pen.dev's presentation mode cannot play. And when something on a screen is
wrong, saying which thing takes a sentence of description where a file and a
line would do.

## What Changes

- A **design canvas**: one page, served by the Storybook dev server, that
  lays stories out as artboards at real sizes, side by side, with pan and
  zoom. It lists itself from Storybook's own story index, so a new story
  appears on it with nobody adding it.
- **Pointing at an element.** Holding Alt and clicking anything inside an
  artboard names its Svelte source, `packages/ui/src/pages/Overview.svelte:142`,
  and copies it, using the location Svelte's dev build already records on
  every element.
- The viewport sizes are declared once and read by both Storybook's toolbar
  and the canvas.
- `just canvas` opens it.

Nothing is removed. Pen, the frame checks and every ADR stay as they are;
whether they should is `make-code-the-design-of-record`, decided after this
has been used.

Not in this change:

- **Click-through on fixture data.** The running app is the click-through
  prototype, and today it runs on real data only: the gateway factory accepts
  `enablebanking` and nothing else, there is no seed data, and sign-in is a
  real passkey ceremony. A fixture mode is its own change, with its own
  argument about a fake gateway being selectable by configuration.
- Any editing surface. Edits are made in the source and hot reload shows them.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. A developer tool with no product behaviour; the change declares
`skip_specs: true`.

## Impact

- `apps/storybook/.storybook/main.ts`: a `staticDirs` entry.
- `apps/storybook/.storybook/preview.ts`: imports the viewport table instead
  of declaring it.
- New under `apps/storybook/canvas/`: the page, its script, the viewport
  table and the pure functions with their tests.
- `apps/storybook/vitest.config.ts`: a second project for plain unit tests.
- `justfile`: a `canvas` recipe.
- No new dependency, so no ADR.
