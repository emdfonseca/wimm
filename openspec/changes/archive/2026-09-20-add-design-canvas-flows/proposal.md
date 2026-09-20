## Why

The design canvas shows every screen and says nothing about how a member gets
from one to the next. A flow review, connecting a bank from Overview to the
chosen accounts, still needs either the running app with a real bank and a
real passkey, or the `.pen` journey files, whose presentation mode cannot play
them. The screens already take their intent as callbacks and the stories
already fill those callbacks with spies, so a click-through on fixture data is
a lookup away.

## What Changes

- **Flows are declared once**, in `apps/storybook/canvas/flows.js`: a name, the
  stories it passes through in order, the transitions between them (this
  story, this callback, that story), and a routes table (this href, that
  story) so a link in the app shell or on a screen goes somewhere.
- **Flow view.** `?flow=<name>` lays that flow's stories out in one row, in
  order, with a connector between neighbours labelled with what moves the
  member on. The sitemap gains a Flows section above the layers.
- **Play mode.** `?play=<name>` shows one artboard at a time, live. Clicking
  something that calls a mapped callback, or a mapped link, swaps in the next
  story. Something with nowhere to go says so in the bar. Back steps back,
  Escape returns to the flow view, and the size can be switched while
  playing.
- A flow naming a story that does not exist is refused by name when the design
  canvas loads.
- One flow ships with it: connecting a bank, from Overview to Accounts.

Not in this change:

- **Click-through on real routes.** Play mode swaps fixtures; it never runs a
  load function, a form action or the gateway. A fixture mode for `apps/web`
  is its own change and needs a decision on sign-in.
- State carried between screens. Picking Monzo does not make the next story
  say Monzo unless that story already does.
- Arrows across the whole design canvas. A flow is seen in its own view.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. A developer tool with no product behaviour; the change declares
`skip_specs: true`.

## Impact

- New: `apps/storybook/canvas/flows.js`, and flow functions with their tests
  in `canvas/lib.js` and `canvas/lib.test.js`.
- `apps/storybook/canvas/canvas.js`, `canvas.css`, `index.html`: the flow
  view, play mode and the sitemap section.
- `apps/storybook/.storybook/preview.ts`: one decorator, inert unless the
  story's URL carries the play marker.
- `README.md`: one line.
- No new dependency, so no ADR. No story file in `packages/ui` changes.
