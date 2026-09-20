## 1. One viewport table

- [x] 1.1 Move the viewport options out of
      `apps/storybook/.storybook/preview.ts` into
      `apps/storybook/canvas/viewports.js` and import them back. Verify:
      `just check apps/storybook` passes and Storybook's viewport toolbar
      still lists Compact, Medium, Wide and Ultra at the same sizes.

## 2. Pure logic, test-first

- [x] 2.1 Add a `unit` project (node environment) to
      `apps/storybook/vitest.config.ts` and make the package's `test` script
      run both projects. Verify: `just test apps/storybook` reports both.
- [x] 2.2 `/tdd`: `artboardsFrom(index, query)` and `storyUrl(id, globals)` in
      `apps/storybook/canvas/lib.js`. Verify: tests cover the default view
      (every `Pages/` story at wide), `layer`, `title`, `sizes`, `theme` and
      `density`, an unknown size refused by name, and docs entries ignored.
- [x] 2.3 `/tdd`: `relativeToRepo(path)` and `nearestLocation(node)`. Verify:
      an absolute path under `/packages/` and one under `/apps/` become
      repo-relative, a path with neither is returned marked unrecognised, a
      node with no recorded ancestor returns nothing, and the parent chain is
      capped at three.

## 3. The page

- [x] 3.1 `apps/storybook/canvas/index.html`, `canvas.js` and `canvas.css`,
      and `staticDirs` in `.storybook/main.ts`: artboards from `/index.json`,
      labelled with title, story name and rendered size, laid out one row per
      title. Styled from `@wimm/ui` tokens, no raw hex. Verify: with `just up`
      running, `/canvas/index.html` shows every `Pages/` story and a story added to
      `packages/ui` appears after a reload with no other edit.
- [x] 3.2 Content-height artboards, lazy loading within one screen of the
      viewport, pan, zoom about the pointer, `0` to fit and `1` for 100%.
      Verify: `Pages/Overview` is shown whole with no inner scrollbar, and
      the network panel shows off-screen artboards unloaded.
- [x] 3.3 Alt lifts the covers, outlines the element under the pointer, and
      Alt-click shows and copies its source location with up to three
      enclosing components. On a build with no recorded locations the bar
      says pointing needs the dev server. Verify: Alt-clicking the Overview
      screen's "Recent transactions" title copies a path ending
      `packages/ui/src/pages/Overview.svelte` with a line number, and the clipboard never holds an absolute path.
- [x] 3.4 A Vitest browser test that renders a page story and asserts
      `nearestLocation` on its "Recent transactions" title ends `Overview.svelte`. Verify: it
      passes, and fails when `nearestLocation` is made to ignore
      `__svelte_meta`.

## 4. Entry point and verification

- [x] 4.1 `just canvas`: refuse with "run `just up` first" when the Storybook
      port does not answer, otherwise open `/canvas/index.html`. Run it once before
      documenting it in `README.md`. Verify: both branches by hand.
- [x] 4.2 `just check apps/storybook` passes.
- [x] 4.3 Use it on `make-overview-a-useful-dashboard`: view the Overview
      stories at all four sizes beside their pen frames, point at three
      elements, and record in
      `openspec/changes/make-code-the-design-of-record/proposal.md` what the
      design canvas showed that the frames did not, and the reverse.
