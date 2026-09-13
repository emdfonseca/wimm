# Theme and viewport

## Theme is a global

One story, rendered in whichever theme the toolbar selects. The decorator in `preview.ts` sets `data-theme` on the document element; the tokens stylesheet redefines the custom properties under `[data-theme="dark"]`. Components read semantic tokens and therefore never know which theme is active — the same rule as the canvas, where components reference Color-aware variables instead of being duplicated per theme.

Pin a theme on a story only when the state is genuinely theme-specific:

```svelte
<Story name="Default" globals={{ theme: 'dark' }} />
```

That should be rare. If it is common, the component is probably hardcoding colours somewhere — see `references/tokens.md`.

## Viewports mirror the design frames

The presets in `preview.ts` use the same widths as the representative frames in the design standard, so a Storybook screenshot and a `.pen` frame are comparable:

```text
compact  390px   ↔  Compact shell   ↔  Device = Mobile
medium   768px   ↔  Medium shell    ↔  Device = Tablet
wide     1440px  ↔  Wide shell      ↔  Device = Desktop
```

Set a default per story where the component only makes sense in one regime:

```svelte
<Story name="CompactNav" globals={{ viewport: { value: 'compact', isRotated: false } }} />
```

Those widths are review conveniences, not a claim the component works only there. Real responsive verification happens just below, at, and above each structural boundary, plus the WCAG reflow condition (320 CSS px, or a 1280px viewport at 400% zoom) — none of which a fixed preset proves. Storybook shows you the three regimes; it does not test the ranges between them.

## The one legitimate side-by-side

Mirroring zone 40 of a journey file, a component whose theming is genuinely risky — focus rings, elevation, error and disabled treatments, data graphics — may carry a single `ThemeQA` story rendering the same component twice with both themes forced. Label it as a test fixture in its docs description and tag it so it does not clutter the default view.

One per component, at most. The moment `ThemeQA` stories multiply, the matrix has come back through the side door.
