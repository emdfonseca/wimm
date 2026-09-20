## Context

See proposal.md for why. What shapes the approach:

- The design canvas is plain static files under `apps/storybook/canvas/`,
  same origin as `iframe.html`, with pure logic in `lib.js` under a node
  `unit` Vitest project. Artboards are iframes loaded one at a time while the
  canvas is still; the sitemap is built by `sitemapFrom(boards)`.
- Every page in `packages/ui/src/pages/` is presentational: intent leaves as
  `on…` callbacks. Most stories pass `fn()` spies for them in `args`. Two do
  not: `TransactionsScreen` (`onrefresh`) and `WidenConsentScreen`
  (`oncontinue`) declare the prop and no spy.
- Screens and the shell also navigate by plain links: the sidebar, rail and
  bottom bar (`/`, `/accounts`, `/transactions`, `/settings`), and `backHref`,
  `cancelHref`, `connectHref` and friends on screens. In Storybook those hrefs
  lead to a 404 inside the iframe.
- Many page stories have a `play` function that clicks the very buttons a
  flow maps. `Pages/Overview · BeforeAnyBankConnected` clicks "Go to
  Accounts" as soon as it renders.
- Story ids derive from title and name: `Pages/ChooseBankScreen` · `Default`
  is `pages-choosebankscreen--default`.
- `@storybook/svelte`'s `decorateStory` lets a decorator call
  `story({ args })` to replace args. Observed in its source, not yet
  exercised here.

## Language

- **Flow** — a named, ordered walk through stories, declared in `flows.js`.
  Never "journey", which is a pen artifact with an ID, and never "prototype".
- **Step** — one story's place in a flow.
- **Transition** — step, callback name, next step.
- **Route** — an href and the story it stands for while playing.
- **Play mode** — one live artboard that follows transitions and routes.
  Never "preview" or "presentation".

## Goals / Non-Goals

**Goals:**

- A flow costs a few lines in one file and no edit to any story.
- Nothing a test sees changes. The decorator is inert without the marker.
- A dead end is visible: the bar names the callback or href that had nowhere
  to go, so a missing transition reads as missing rather than as a broken
  button.

**Non-Goals:**

- Fidelity to the app's real navigation. Routes are declared, not derived
  from `apps/web`, which `apps/storybook` may not import.
- Carrying state between steps, branching on data, or timing.
- Validating flows in `just check`. The story index exists only on a running
  Storybook; the design canvas refuses a bad flow at load instead.

## Decisions

**One file declares flows, as data.**

```js
export const routes = { '/': 'pages-overview--populated', '/accounts': … };
export const flows = {
  'connect-a-bank': {
    title: 'Connect a bank',
    steps: ['pages-overview--before-any-bank-connected', …],
    transitions: [
      { from: 'pages-overview--before-any-bank-connected',
        on: 'ongotoaccounts', to: 'pages-accountsscreen--empty' },
      …
    ]
  }
};
```

`steps` is the order the flow view draws; `transitions` is what play mode
follows, and may point at a story outside `steps` for a side path. `routes`
is shared by every flow, since `/accounts` means one thing. *Alternative:*
`parameters.flow` on each story. Rejected: a flow would be scattered over
eight story files in `packages/ui`, and reading one would mean opening all of
them. *Alternative:* derive flows from the journey IDs in the `.pen` files.
Rejected: it ties the tool to the artifact `make-code-the-design-of-record`
may retire.

**Pure functions carry the rules.** `flowFrom(flows, name, index)` returns the
flow or throws naming every unknown story id, callback-less transition or
unknown flow name, all at once. `nextStory(flow, storyId, callback)` and
`routeFor(routes, href)` return an id or nothing; `routeFor` ignores query
and hash, so `/transactions?refresh=1` resolves as `/transactions`.
`callbacksFor(flow, storyId)` lists the callbacks play mode must hear on a
step. All TDD'd in `lib.test.js`.

**Flow view is a filtered canvas, not an overlay.** `?flow=<name>` replaces
the artboard list with the flow's steps in order, one row, at the sizes the
URL asks for, and draws a connector between neighbours labelled with the
transition's callback. Everything else, loading, pan, zoom, pointing, is the
existing code. *Alternative:* arrows between artboards where they already sit
on the full canvas. Rejected: steps live in different rows, so every arrow
crosses the layout and none can be followed.

**Play mode hears callbacks through a decorator, and only when asked.** The
artboard URL carries `&play=ongotoaccounts,onseeall`, the step's callbacks
from `callbacksFor`. A preview decorator that finds no `play` parameter in
`location.search` returns `story()` untouched, which is every test and every
ordinary Storybook visit. With it, the decorator replaces each named arg with
a function that calls the original, if the story had one, and then posts
`{ type: 'wimm-canvas-callback', name }` to `window.parent`. Names come from
the URL rather than from wrapping every function arg because two screens
declare callbacks their stories never pass, so there is nothing to wrap.
*Alternative:* the design canvas reaches into the iframe and patches the
component. Rejected: props are not reachable from the DOM.

The decorator also wraps every function-valued `on*` arg the story already
has, so an unmapped callback is heard and named as a dead end. The marker is
`play=` with a possibly empty list. Step swaps use `location.replace`, because
an assigned iframe `src` adds a history entry that Back would undo instead of
the step. `story({ args })` reached the component as hoped; args replace
wholesale, so the decorator merges into `context.args`.

**A story's own play function must not drive the flow.** The wrapped callback
forwards only while `navigator.userActivation.isActive`, which a real click
sets and Testing Library's synthetic events do not. So
`BeforeAnyBankConnected` clicking its own button on render moves nothing.
*Alternative:* wait for Storybook's `storyRendered` channel event before
listening. Rejected as the primary guard: it couples the decorator to the
channel and still lets a slow play function through; user activation is one
condition and states the actual rule, a person clicked.

**Links are caught in the artboard's document.** In play mode the design
canvas adds a capturing `click` listener to the iframe document, as pointing
already does. A click on an `a[href]` is always prevented, then resolved with
`routeFor`. Alt-click still points.

**Dead ends speak.** An unmapped callback or href writes
`No transition for onrefresh on Pages/TransactionsScreen · AsItOpens` or
`No route for /connect/monzo` to the bar. Play mode never fails silently,
because silence is indistinguishable from a broken screen.

**Play mode is one artboard and a history stack.** A step swap sets the
iframe's `src`; the stack gives Back (`[` and the browser's back, via
`history.pushState` with the step id in the hash). Escape returns to
`?flow=<name>`. `c`, `m`, `w`, `u` switch size. The artboard is shown at 100%
and scrolls like a page, since a prototype is used, not surveyed; pan and zoom
are off in play mode.

**The first flow is connecting a bank**, Overview (no bank) → Accounts (empty)
→ Choose bank → Consent → Choose accounts → Accounts (connected), with the
four shell routes mapped. It exercises a callback transition, a link
transition and a step whose story has a clicking play function.

## Risks / Trade-offs

- [`story({ args })` may not reach a Svelte CSF story's component] → It is the
  first task and nothing else is built until a wrapped callback is seen to
  post. The fallback is assigning onto `context.args` before `story()`.
- [Story ids rot when a story is renamed] → `flowFrom` refuses at load and
  names the id. Not caught by `just check`, stated in Non-Goals.
- [A flow implies behaviour the app does not have] → The bar shows
  "fixture flow" throughout play mode, and the proposal's exclusions are in
  the README line.
- [`userActivation` stays active about five seconds after a click] → A play
  function that runs inside that window after a step swap could forward. The
  new document of a swapped iframe starts without activation, so it cannot.

## Open Questions

- Whether flows should later be checked in CI from `storybook index`'s output.
  Revisit once there is more than one flow.
