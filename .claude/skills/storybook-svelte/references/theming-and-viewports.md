# Theme and viewport

## Theme is a global

pen `Color` axis values map to the Storybook `theme` global:

```text
pen      Storybook global   DOM
Light    light              [data-theme="light"]
Dark     dark               [data-theme="dark"]
```

The decorator in `preview.ts` sets `data-theme` on the document element; the tokens stylesheet redefines the custom properties under `[data-theme="dark"]`. Components read semantic tokens and never know which theme is active.

Pin a theme on a story only when the state is genuinely theme-specific:

```svelte
<Story name="Default" globals={{ theme: 'dark' }} />
```

If that is common, the component is hardcoding colours — see `references/tokens.md`.

## Viewports mirror the design frames

The presets in `preview.ts` use the regime names and widths from the design standard, so a Storybook screenshot and a `.pen` frame are comparable:

```text
viewport key   width    pen frame / shell / Device value
compact        390px    Compact
medium         768px    Medium
wide           1440px   Wide
```

Pin a viewport where the component only makes sense in one regime:

```svelte
<Story name="CompactNav" globals={{ viewport: { value: 'compact', isRotated: false } }} />
```

Those widths are review conveniences. Real responsive verification happens just below, at, and above each structural boundary, plus the WCAG reflow condition (320 CSS px, or a 1280px viewport at 400% zoom) — a fixed preset proves none of that.

## The one legitimate side-by-side

A component whose theming is genuinely risky — focus rings, elevation, error and disabled treatments, data graphics — may carry a single `ThemeQA` story rendering the same component twice with both themes forced. It mirrors the design's QA zones: library zone 90 for library components, journey zone 40 for Pages. Label it as a test fixture in its docs description and tag it out of the default view.

One per component, at most.
