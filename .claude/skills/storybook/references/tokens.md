# Tokens

## One stylesheet, no copies

`preview.ts` imports the same token stylesheet the application imports — `packages/ui/src/lib/styles/tokens.css`, not a Storybook copy or subset.

```css
/* packages/ui/src/lib/styles/tokens.css */
:root {
  --color-bg-surface: <light surface>;
  --color-text-primary: <light text>;
  --color-action-primary: <brand>;
  --layout-page-gutter: <px>;
  --radius-control: <px>;
  --focus-ring-color: <brand>;
  --focus-ring-width: <px>;
}

[data-theme='dark'] {
  --color-bg-surface: <dark surface>;
  --color-text-primary: <dark text>;
  --color-action-primary: <brand, adjusted for dark>;
}
```

## Design ↔ code naming

Dotted pen variable → kebab-case CSS custom property. The mapping is a convention, not an automatic conversion; document units where they differ (canvas px vs rem), and not every design number is a pixel.

```text
pen.dev                    code
------------------------------------------------
color.bg.surface     ↔     --color-bg-surface
color.text.primary   ↔     --color-text-primary
layout.pageGutter    ↔     --layout-page-gutter
radius.control       ↔     --radius-control
```

Layer vocabulary stays the same across design and code:

```text
Design taxonomy          Implementation concept
--------------------------------------------------------------
Foundation          ↔    tokens / styles / shared constraints
Atom                ↔    small reusable control
Molecule            ↔    focused composition of controls
Organism            ↔    reusable section / complex interaction
Template            ↔    page-level scaffold / shell
Page                ↔    presentational screen component + the route that wires it
Journey             ↔    cross-view user flow and behavior
```

## Components consume semantics

```css
/* ✅ */ background: var(--color-bg-surface);
/* ❌ */ background: #ffffff;
/* ❌ */ background: var(--grey-100);
```

A component using a raw value or a primitive ramp is wrong in the other theme, and the story will not tell you.

## Responsive token values

Where the design uses `Device`-aware variables (Compact / Medium / Wide), the implementation counterpart is a media or container query redefining the custom property, not a separate component:

```css
:root { --layout-page-gutter: 16px; }                       /* Compact */
@media (min-width: 768px)  { :root { --layout-page-gutter: 24px; } }  /* Medium */
@media (min-width: 1200px) { :root { --layout-page-gutter: 32px; } }  /* Wide */
```

The canvas cannot derive one mechanism from the other; boundaries are written down in the Project Setup Record.

## Documenting them

A `Foundations/Tokens` MDX page that reads the computed custom properties from the DOM stays correct by construction. A page listing hex values in Markdown is a second source of truth.
