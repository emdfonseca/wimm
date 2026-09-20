# Theme and viewport

## Theme is a global

The `color` axis in `design/tokens.json` maps to the Storybook `theme` global:

```text
axis     Storybook global   DOM
light    light              [data-theme="light"]
dark     dark               [data-theme="dark"]
```

The decorator in `preview.ts` sets `data-theme` on the document element; the tokens stylesheet redefines the custom properties under `[data-theme="dark"]`. Components read semantic tokens and never know which theme is active.

Pin a theme on a story only when the state is genuinely theme-specific:

```svelte
<Story name="Default" globals={{ theme: 'dark' }} />
```

If that is common, the component is hardcoding colours — see `references/tokens.md`.

## Viewports are the four regimes

The presets are declared once, in `apps/storybook/canvas/viewports.js`. `preview.ts` and the design canvas both read them, so a story in the toolbar and the same story on the design canvas are the same width:

```text
viewport key   width    regime
compact        390px    Compact   below 768
medium         834px    Medium    768 and above
wide           1440px   Wide      1200 and above
ultra          1920px   Ultra     1800 and above
```

Pin a viewport where the component only makes sense in one regime:

```svelte
<Story name="CompactNav" globals={{ viewport: { value: 'compact', isRotated: false } }} />
```

Those widths are review conveniences. Real responsive verification happens just below, at, and above each structural boundary, plus the WCAG reflow condition (320 CSS px, or a 1280px viewport at 400% zoom) — a fixed preset proves none of that.

## The one legitimate side-by-side

A component whose theming is genuinely risky — focus rings, elevation, error and disabled treatments, data graphics — may carry a single `ThemeQA` story rendering the same component twice with both themes forced. Label it as a test fixture in its docs description and tag it out of the default view.

One per component, at most.
