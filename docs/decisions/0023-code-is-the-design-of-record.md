# 0023 · Code is the design of record

## Status

Accepted

## Context

Every user-facing change was built twice: drawn as a frame in a `.pen` file, then
written in Svelte, with a generated copy contract and a frame-to-file map
standing between the two copies to catch drift. They did not catch all of it.
Transactions' frames drew month chips and an Older button while the built screen
paged with a dotted track, and nothing failed.

A trial on the Overview screen compared the two. The design canvas in Storybook
showed the built states at four real widths, the reflow between them, and which
file creates any element pointed at. The frames showed one width in one state.
What only the frames gave was intent with no code yet, an order through a flow,
and a whole screen at a glance. A flow on the design canvas carries the order.

*Alternative:* keep the `.pen` files ungated, for sketching. Rejected: an
ungated drawing beside the code is still read as a decision.

*Alternative:* keep the copy contract as a hand-edited list. Rejected: it is the
second copy under another name, and its seventeen accepted divergences show how
it ages.

## Decision

**A page's presentational Svelte component and its state stories are the design
of record.** A design is a screen rendered from fixtures in Storybook and looked
at on the design canvas at compact, medium, wide and ultra, before any wiring
exists. There is no drawing of a screen.

`canvas.md` keeps its place between design and tasks and becomes a state plan:
the state stories a change will write, the words each shows, the flow they join,
and what was seen. Planning writes no code, so apply runs in a fixed order:
components, the presentational screen with its state stories, a look at all four
regimes that writes Seen into `canvas.md`, then wiring.

**The words of a state are held by its story's play function.**
`check-stories.py` refuses a page story that has none, and
`test-stories.py` asserts that refusal.

**`design/tokens.json` is the source of token values.** Nothing exports it. The
validation in ADR 0003 is unchanged.

The journey files, the pen library and its manifest, the `bin/pen-*` tooling,
the `pen-design` skill, the pencil MCP entry and the four frame-check scripts
are deleted. Commit `38355d1` is the last that holds the `.pen` files;
`git show 38355d1:<path>` restores one.

This supersedes ADRs 0020, 0009, 0010 and 0011 whole, the drawing in 0008, and
the library clause of 0003. 0008's artifact, its conditional skip and its
components-before-screens rule stand.

## Consequences

**Copy coverage is reduced, and this is not parity.** The frame contract checked
597 strings. What is checked now is what play functions assert. Every notice
title, notice body and failure message in the contract was read against the page
stories before the contract went, and each got an assertion or a recorded reason.

**The play-function rule cannot see what is asserted.** It proves a page story
has a play function, not that the function names the state's words. That stays a
review matter, read against the Words section of the change's `canvas.md`.

Intent with no code cannot be recorded as a picture. The state plan holds the
words and the states; a presentational screen on fixtures is the sketch.

A component missing from the design system is two deliverables, the Svelte
component and its story, where it was three.

`check-geometry.py` stays as a CSS check. Its numbers are the record, not a copy
of one.

Archived changes name frames that exist only at `38355d1`. An archive is a
record, not a reference, and is not rewritten.
