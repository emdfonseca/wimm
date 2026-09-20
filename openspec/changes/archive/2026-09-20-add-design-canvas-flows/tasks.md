## 1. Prove the hook

- [x] 1.1 A preview decorator in `apps/storybook/.storybook/preview.ts` that
      does nothing without `play=` in `location.search`, and with it replaces
      each named arg by a function calling the original and posting
      `{ type: 'wimm-canvas-callback', name }` to `window.parent`, only while
      `navigator.userActivation.isActive`. Verify: `just check apps/storybook`
      passes unchanged; opening
      `iframe.html?id=pages-overview--before-any-bank-connected&viewMode=story&play=ongotoaccounts`
      inside a test page posts once on a real click and not at all from the
      story's own play function. If `story({ args })` does not reach the
      component, switch to assigning onto `context.args` and record which in
      design.md.

## 2. Flow rules, test-first

- [x] 2.1 `/tdd`: `flowFrom(flows, name, index)` in `canvas/lib.js`. Verify:
      tests cover a valid flow, an unknown flow name refused by name, unknown
      step and transition ids all reported in one error, a transition with no
      `on`, and docs entries not counting as stories.
- [x] 2.2 `/tdd`: `nextStory`, `callbacksFor` and `routeFor`. Verify: a mapped
      and an unmapped callback, callbacks listed once per step, a route
      matched with query and hash ignored, an absolute same-origin href, and
      an unmapped href returning nothing.
- [x] 2.3 `apps/storybook/canvas/flows.js` with `routes` for `/`, `/accounts`,
      `/transactions`, `/settings` and the `connect-a-bank` flow. Verify: a
      unit test runs `flowFrom` over the real file against a fixture index
      built from the ids it names, and the ids match the running Storybook's
      `/index.json`.

## 3. Flow view

- [x] 3.1 `?flow=<name>`: the flow's steps in one row in order, a labelled
      connector between neighbours, a bad flow shown in the bar by name.
      Verify: `?flow=connect-a-bank` shows six steps in order at Compact and
      Wide, and `?flow=nope` names `nope`.
- [x] 3.2 A Flows section at the top of the sitemap, each entry opening its
      flow view, with a Play control. Verify: it lists `Connect a bank` and
      both links work.

## 4. Play mode

- [x] 4.1 `?play=<name>`: one artboard at 100%, its URL carrying the step's
      callbacks, swapped on a posted callback through `nextStory`; the bar
      shows the flow, the step and "fixture flow". Verify: clicking through
      Overview to Accounts (connected) reaches every step with real clicks.
- [x] 4.2 Links: a capturing click listener resolves `a[href]` through
      `routeFor`; Alt-click still points. Verify: the sidebar's Transactions
      link swaps to the routed story and the iframe never navigates to a 404.
- [x] 4.3 Dead ends, Back, Escape and size keys. Verify: an unmapped callback
      and an unmapped href each name themselves in the bar; `[` and the
      browser's back return to the previous step; Escape returns to the flow
      view; `c` and `w` switch size on the current step.

## 5. Verification

- [x] 5.1 README line for flows and play mode, stating it swaps fixtures and
      runs no route. Verify: the documented URLs open.
- [x] 5.2 `just check apps/storybook` and `just check packages/ui` pass, and
      `git diff --stat packages/ui` is empty.
