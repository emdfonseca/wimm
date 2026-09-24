# Storybook and the design canvas

Storybook serves `packages/ui` in isolation, and the design canvas at
`/canvas/index.html` lays a screen's state stories side by side. Both start
with `devbox run -- just up`, on http://localhost:9469.

`just canvas` opens the design canvas: one screen at a time, its state stories
at real sizes, one row per kind of story: States, Waiting, Errors, Outcomes,
then Behaviour for the stories that only assert something. A page story says
which with one tag, `tags={['kind-error']}`; the kinds are defined in
`apps/storybook/canvas/lib.js`, and `just check apps/storybook` fails for a page
story with none. A story that pins its own viewport adds `size-compact` and is
drawn at that size only. Pan and zoom as usual. The sidebar lists every screen and
story, and `/` finds one. Alt-click an element to copy its source location. It
needs `just up` running and starts nothing itself. Append `?flow=connect-a-bank`
to `/canvas/index.html` to see a flow as a map: its happy path left to right, and
under each step the ways off it, or `?play=connect-a-bank` to click through the
happy path; play mode swaps fixture screens and runs no route, load function or
bank. Flows are declared in `apps/storybook/canvas/flows.js`, one per journey a
member would name. Every page story that shows something is on a flow or on that
file's `unplaced` list, and `just check apps/storybook` fails for one that is on
neither; a `kind-behaviour` story only asserts something and is on neither.

Every page story has a version: a fingerprint of what it renders, shown as seven
characters and the date first seen. It moves when the markup, a component's
styles, the tokens, base styles or fonts move, and not for a refactor that
changes none of them. `just check apps/storybook` brings
`apps/storybook/canvas/versions.json` up to date and never fails because it was
behind, so it can leave a diff to commit; that diff lists the pages whose look
changed. Each artboard's badge shows the approved and the implemented version:
one line when they are the same, both when the page changed since approval, and
the implemented one alone when nobody approved it. History opens a story's
approved versions, and the sidebar filter `Needs approval` (kept as `?needs=1`)
lists what to look at. To approve a page after looking at it, run
`just approve pages-overview--populated "a note"` in your own terminal. Name
several stories to approve them in one run, or use `just approve --needs` for
every page story that needs approval; either asks once, and one refused story
refuses the run. It refuses
inside an agent session or without a terminal, and an agent never runs it. The
record only grows; two people approving on two branches merge by keeping both
lines.

Approving also keeps a picture of the page at every size it is drawn at, light
theme and comfortable density only, taken from the same run that settled the
version approved. Only a story's last approved version keeps pictures, under
`apps/storybook/canvas/approved/<story>/`; approving again replaces them. Commit
the record and the pictures together, on their own, as the command says:
`apps/storybook/canvas/approvals.jsonl` and
`apps/storybook/canvas/approved/pages-overview--populated/`. On an artboard that
changed since approval, viewed light and comfortable, `Approved look` switches
between the page as implemented and that picture; an approval recorded with no
picture says so instead. Open the canvas with `?data=changed` (or
`changed-pictured`, `approved`, `never`, `mixed`, `history`) to see a state on
fixtures without approving anything.


## Checking it

```bash
devbox run -- just check apps/storybook   # lint, typecheck, stories as tests, versions, approvals
```

The stories' play functions are the tests; they run in Chromium through
Vitest's browser mode. `versions.json` is written from the run that just
passed. `scripts/approvals.mjs --check` refuses a changed or removed approval
line and a picture that does not belong to its story's last approved version;
a story with no picture passes.
