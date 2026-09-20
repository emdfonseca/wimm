## Context

See proposal.md for why. What shapes the approach:

- Storybook is its own app, `apps/storybook`, `@storybook/sveltekit` 10.6,
  started by `just up` on `WIMM_STORYBOOK_PORT` (9469). It has no
  `staticDirs` and no manager file.
- Storybook serves `/index.json`, every story's `id`, `title`, `name` and
  `tags`, and `/iframe.html?id=<id>&viewMode=story&globals=…`, one story with
  nothing around it.
- Svelte 5.57's dev build sets `element.__svelte_meta = { parent, loc: { file,
  line, column } }` on every element it creates
  (`svelte/src/internal/client/dev/elements.js`). A `storybook build` has
  none of it. `file` is relative to the compiler's `rootDir` only when the
  path sits under it; Storybook's is `apps/storybook`, so `packages/ui` files
  are expected to arrive absolute. Expected, not yet observed.
- The viewport table lives inline in `preview.ts`: compact 390 × 844, medium
  834 × 1112, wide 1440 × 900, ultra 1920 × 1080. The pen frames draw Medium
  at 1024. The canvas follows Storybook, which is what the stories are
  tested at.
- `theme` and `density` are Storybook globals, applied by a decorator to the
  document root.
- Play functions run in Vitest browser mode under one project, `storybook`.
  There is no plain unit project in `apps/storybook`.

## Language

- **Design canvas** — the page. Never "the canvas" alone in prose about this
  repo, where that still means the pen.dev artifact.
- **Artboard** — one story at one viewport on the design canvas. Never
  "frame", which is a pen node.
- **Point at** — Alt-click an element to get its source location. Never
  "inspect" or "select".
- **Source location** — `path:line:column`, the path relative to the repo
  root.

## Goals / Non-Goals

**Goals:**

- Nothing to maintain per screen. A story that exists is on the design
  canvas.
- One declaration of the viewport sizes.
- The page is plain files with no build step and no dependency, so it cannot
  break the Storybook build it rides on.

**Non-Goals:**

- Editing, comments, saving a layout, sharing. It is local and stateless
  apart from the URL.
- Working from a static `storybook build`. The artboards render there;
  pointing does not, and the page says so rather than failing silently.
- Replacing Storybook's own UI for controls, docs or accessibility results.

## Decisions

**Served by Storybook, as static files.** `staticDirs: [{ from: '../canvas',
to: '/canvas' }, { from: '../../../packages/ui/src', to: '/canvas-ui' }]`, so the
page is `http://localhost:9469/canvas/index.html` (Storybook serves no
directory index) and takes its stylesheet from `/canvas-ui/base.css`. Same origin
as `iframe.html` is the whole reason: pointing at an element reads
`iframe.contentDocument`, which a page on another port cannot. *Alternative:*
a dev-only route in `apps/web`. Rejected: a different origin, and an apps to
apps dependency the monorepo standard bans. *Alternative:* a Storybook
manager addon. Rejected: manager addons are React, and a tab inside the
manager still shows one story's worth of chrome around everything.

**Artboards come from `/index.json`.** Entries of type `story`, grouped by
title, in index order. The default view is every `Pages/` story at `compact` and `wide`;
the URL carries the rest: `?layer=pages`, `?title=Pages/Overview`,
`?sizes=compact,wide`, `?theme=dark`, `?density=compact`. The URL is the only
state, so a view can be pasted into a message. Each artboard is
`iframe.html?id=<id>&viewMode=story&globals=theme:<t>;density:<d>;viewport:<v>`.

**An artboard is as tall as its content.** The iframe starts at the
viewport's height and, once loaded, is resized to its document's
`scrollHeight`, re-measured on a `ResizeObserver`. A dashboard 1470 tall is
shown whole, not behind a scrollbar, which is the difference between this and
Storybook's own viewport tool.

**Pan and zoom are a CSS transform on one container.** Wheel pans, pinch or
Ctrl-wheel zooms about the pointer, `0` fits everything, `1` is 100%. Iframes
swallow wheel events, so while no key is held a transparent cover sits over
each artboard and the design canvas gets the wheel; holding Alt lifts the
covers so clicks reach the story. That one key does both jobs: point, and
interact.

**Pointing walks up to the nearest recorded element.** On Alt-click inside an
artboard: from `event.target` up through `parentElement` until one carries
`__svelte_meta`; its `loc` becomes the source location, and its `parent`
chain gives up to three enclosing components. Alt-hover outlines the element
that would be reported. The result is shown in a bar, and copied as
`packages/ui/src/pages/Overview.svelte:142:5`.

**Paths are made repo-relative by cutting at a known root.** Everything
before the last `/packages/` or `/apps/` segment is dropped. The page never
learns where the repo is, and an absolute path never reaches the clipboard.
A path with neither segment is shown as it came, marked as unrecognised.

**The viewport table moves to `canvas/viewports.js`.** Plain ESM, exporting
the same object `preview.ts` holds today; `preview.ts` imports it and the page
imports it. One declaration, two readers, per the generated-artifacts rule
about paired checks reading one source of truth. It is `.js` because a static
file cannot import TypeScript, and `preview.ts` can import either.

**Logic is pure functions in `canvas/lib.js`, tested without a browser.**
`artboardsFrom(index, query)`, `storyUrl(id, globals)`, `relativeToRepo(path)`
and `nearestLocation(node)`, the last taking any object with `parentElement`
and `__svelte_meta`. `canvas.js` is the thin DOM half. Vitest gains a `unit`
project in node environment beside `storybook`, and `just test apps/storybook`
runs both.

**One check in a real browser.** A Vitest browser test mounts the Overview
screen, finds its "Recent transactions" title, and asserts `nearestLocation`
returns a path ending `Overview.svelte`. An element is named by the file that
creates it, so the screen's heading, which `Page.svelte` renders, reports
`Page.svelte` with Overview as an enclosing component. It is the test that the expectation about
`__svelte_meta` holds under this Storybook, and the first thing to fail if a
Svelte upgrade moves it.

**`just canvas`** checks the Storybook port answers, says to run `just up` if
it does not, and otherwise opens the page. It starts nothing itself.

**Signed-in pages are drawn inside the app shell.** A story opts in with
`parameters.shell`, the path the shell marks as current, and a preview
decorator wraps it in `SignedInLanding` exactly as `apps/web`'s `(app)` layout
does. The shell fills the viewport and scrolls its content inside, so the
artboard is grown until no inner scroller hides anything.

**A sitemap lists every artboard by layer, title and story.** Choosing one
centres its sizes together; the URL hash records it and `/` searches.
Artboards load one at a time, nearest the viewport first and only while the
canvas is still, so gestures never wait on a story booting.

## Risks / Trade-offs

- [`__svelte_meta` is a dev internal, not an API] → One browser test pins it,
  and the page degrades to "no source location recorded" rather than
  throwing.
- [Forty iframes at once is heavy] → Artboards outside the viewport are not
  loaded until they come within one screen of it, and the default view is one
  layer at one size.
- [A story using `globals` of its own overrides the URL's] → Correct: a story
  that pins `viewport` means it. The artboard is labelled with the size it
  actually rendered at.
- [The static page ships in `storybook build`] → Harmless, and it says
  pointing needs the dev server.

## Open Questions

- Whether a pointed-at location should also be written somewhere an agent
  can read without a paste. It needs a local endpoint, which a static page
  does not have. Clipboard first; revisit after use.
