# Tokens

## One stylesheet, no copies

`preview.ts` imports the same token stylesheet the application imports. Not a Storybook-specific copy, not a subset — the real one. A copy drifts, and the drift shows up as "it looked right in Storybook", which is the single most expensive sentence in a design system.

```css
/* packages/ui/src/lib/styles/tokens.css */
:root {
  --color-bg-surface: #ffffff;
  --color-text-primary: #16181d;
  --color-action-primary: #2b5fd9;
  --layout-page-gutter: 32px;
  --radius-control: 6px;
  --focus-ring-color: #2b5fd9;
  --focus-ring-width: 2px;
}

[data-theme='dark'] {
  --color-bg-surface: #14161a;
  --color-text-primary: #e9ecf2;
  --color-action-primary: #7aa2f7;
}
```

Names mirror the design variables (`color.bg.surface` → `--color-bg-surface`), per the design-to-code mapping in the standard. The mapping is a convention, not an automatic conversion — document units where they differ, and note that not every design number is a pixel.

## Components consume semantics

```css
/* ✅ */ background: var(--color-bg-surface);
/* ❌ */ background: #ffffff;
/* ❌ */ background: var(--grey-100);
```

A component reaching for a raw value or a primitive ramp instead of a semantic role will be wrong in the other theme, and the story will not tell you — it renders perfectly in whichever theme you were looking at.

## Responsive token values

Where the design uses `Device`-aware variables, the implementation counterpart is a media query (or container query) redefining the custom property, not a separate component:

```css
:root { --layout-page-gutter: 16px; }
@media (min-width: 768px)  { :root { --layout-page-gutter: 24px; } }
@media (min-width: 1200px) { :root { --layout-page-gutter: 32px; } }
```

The design axis and the media query are different mechanisms describing the same intent — the canvas cannot derive one from the other, which is exactly why the standard insists the boundaries be written down rather than inferred from a frame width.

## Documenting them

A `Foundations/Tokens` MDX page that reads the computed custom properties from the DOM stays correct by construction. A page listing hex values in Markdown is a second source of truth, and it will be wrong within a month.
