## Why

Every user-facing change is built twice: drawn in a `.pen` frame, then written
in Svelte, with `check-canvas.py` and `gen-canvas-contract.py` standing between
the two copies to catch drift. They do not catch all of it. Transactions'
frames drew month chips and an Older button for weeks while the built screen
paged with a dotted track, and nothing failed. Drawing is also the slow half:
a `pen-exec` round trip takes 20 to 60 seconds against a second for a Svelte
edit under hot reload, and pen.dev's presentation mode cannot play the zoned
files at all.

## Trial

`add-local-design-canvas` used on `make-overview-a-useful-dashboard`: the
`Pages/Overview` stories at Compact, Medium, Wide and Ultra, read against
`apps/web/design/05-overview.pen`, and three elements pointed at.

**The design canvas showed what the frames did not.**

- The built states, not the intended ones: eight stories at four sizes, each
  at its real width. The frames draw Wide at 1920 only, in one state.
- The screen that exists. The frames draw Top merchants, Largest payments, a
  balance chart and a period comparison; the built Overview has the household
  total, recent transactions and a trend. The gap is the unbuilt half of the
  change, visible at a glance.
- Compact reflow: `Overview` reads `768px` itself, so the recent rows change
  shape between 390 and 834 with nothing in a frame to say so.
- Where a thing comes from. Alt-click on the "Recent transactions" title names
  `packages/ui/src/pages/Overview.svelte:122:5`; on a tile,
  `molecules/MetricTile.svelte:28:0`; on "See all",
  `atoms/Button.svelte:40:0`. An element is named by the file that creates it,
  so a shared atom reports itself and the calling screen appears as an
  enclosing component, not as the answer.

**The frames showed what the design canvas did not.**

- Intent that has no code yet. A frame can be drawn before its component
  exists; a story cannot.
- Frames arranged as a journey with an entry and an outcome (`J12`). A story
  index is flat and says nothing about how a member arrives.
- A whole screen at once. Every artboard in a row is a separate iframe, so
  what the design canvas gives at 32 artboards in one row (8 states × 4
  sizes) is a strip too small to read when fitted. It wants one row per story
  with the sizes across.

**What the trial settles.** Two of the three things only the frames gave are
answered elsewhere. A journey's order is a flow on the design canvas
(`add-design-canvas-flows`). A frame's copy is checked against the screen by
`check-canvas.py`, and the page stories already assert their own copy in play
functions that run in `just ci`, so the story carries the check once there is
no second copy. Intent with no code yet is the one loss, and it is accepted: a
presentational screen on fixtures is cheap enough to be the sketch.

## What Changes

- **BREAKING** for the process: the Svelte screen and its stories become the
  design of record. A design is a presentational screen rendered from
  fixtures in Storybook, seen on the local design canvas, before any wiring
  exists.
- The OpenSpec `canvas` artifact keeps its place in the graph, between design
  and tasks, and changes what it asks for: the state stories to be written,
  the words each shows, and what was seen on the design canvas, instead of
  frame IDs in a `.pen`. Planning writes no code, so the screen and its
  stories are the first tasks, a look at all four sizes follows, and wiring
  comes after.
- Components still come before the screen that uses them. Each becomes two
  deliverables, the Svelte component and its story, not three.
- The frame-to-screen checks are retired, since there is no second copy to
  compare. Copy is asserted by each page story's play function. The layout
  measurements in `check-geometry.py` read only CSS and stay.
- The journey `.pen` files, the pen library, its manifest, the `bin/pen-*`
  tooling, the `pen-design` skill and the pencil MCP entry are deleted. Git
  keeps them. `design/tokens.json` becomes the source of token values, not an
  export of one.
- `make-overview-a-useful-dashboard` is rewritten to the new shape before it
  is applied: its `canvas.md` and the tasks that edit the library and the
  canvas contract.

Not in this change: the canvas itself, which is `add-local-design-canvas`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app/screen-layout`: the requirement that a screen matches the frame drawn
  for its regime, and the regime requirement's frame wording, now name the
  screen's stories as what a regime is designed in.

## Impact

- A new ADR superseding ADR 0020, and the pen halves of ADRs 0008, 0009, 0010
  and 0011. ADR 0002's token naming and ADR 0003's token pipeline are
  untouched: `design/tokens.json` stays the contract.
- `openspec/schemas/wimm/`: the `canvas` instruction and template, and the
  `tasks` instruction's three-deliverable rule.
- `CLAUDE.md`: the four lines about `.pen` files, the library and drawing.
- `.claude/skills/pen-design/`: deleted. `.claude/skills/storybook/`: its
  references to the `.lib.pen` reworded. `docs/design/`: `canvas-audit.md` and
  `library-conventions.md` deleted, `tokens.md`, `surfaces.md` and
  `project-setup-record.md` reworded.
- `.mcp.json`: the pencil server. `openspec/config.yaml`: the pen context and
  the apply and archive guidance about frames.
- `packages/ui/scripts/check-canvas.py`, `gen-canvas-contract.py`,
  `test-canvas.py`, `test-contract.py`, `src/canvas-contract.json`, and their
  place in `just gen` and `just check packages/ui`. `check-geometry.py` stays.
- `apps/web/design/*.pen`, `packages/ui/design/product-ui.lib.pen` and
  `library-manifest.tsv`.
- `bin/pen-exec`, `bin/pen-import`, `bin/pen-save`, `bin/pen-manifest`,
  `bin/pen-verify-tokens`, `bin/test-pen-exec`, their `just` recipes, and
  `pen-exec-test` in `just ci`.
- What a story must now prove that a frame used to: copy and states, stated
  in the `canvas` instruction. See design.md.
