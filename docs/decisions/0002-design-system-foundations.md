# 0002 · Design system foundations

## Status

Accepted

## Context

`packages/ui/design/product-ui.lib.pen` is the design library every journey file
imports and every UI component implements. Its typefaces are runtime dependencies,
its token names are a public contract consumed by `tokens.css` and by component
code, and its theme axes decide what is expressible at all. Two of the axis
choices depart from the pen-design standard the repo otherwise follows, so they
need a recorded decision rather than a local one.

## Decision

**Typefaces.** Schibsted Grotesk for interface text, IBM Plex Mono for amounts,
account identifiers and table dates. Both are Google Fonts, self-hosted.
`type-family-display` resolves to the interface family — display is weight 600 at
−2% tracking, not a third font.

**Token naming.** Kebab-case (`color-bg-canvas`), mapping 1:1 to CSS custom
properties (`--color-bg-canvas`). This departs from the dotted vocabulary in the
pen-design standard's token reference; roles and structure are unchanged.

**Palette.** Pine & Signal. Deep pine green `#14402F` is the brand and primary
action; blue `#2563EB` is the highlight for links, selection and focus. Green and
red are reserved for money direction and feedback, never for branding, so an
amount's colour means exactly one thing.

**Theme axes.** Three, independent:

- `color` — light, dark.
- `device` — compact, medium, wide, **ultra**. A fourth regime beyond the
  standard's three, justified by one structural change at ≥1800 CSS px: the
  editing drawer stops being an overlay and becomes a persistent inspector pane.
- `density` — comfortable, compact. A third axis beyond the standard's two,
  justified by row height being the most consequential number in a ledger: at
  1440 × 900, roughly 684 px of table area shows 12 comfortable rows or 17
  compact ones.

## Consequences

- Adding a typeface, a theme axis, or an axis value is a new ADR.
- In dark theme the brand mint `#9FE0C0` and income green `#3DD68C` separate at
  only 1.24:1. Both alternatives fail: a deep pine button reaches 2.16:1 against
  the near-black canvas, below the 3:1 that SC 1.4.11 requires for a control to be
  identifiable; desaturating far enough to separate by lightness washes the brand
  out. Context therefore carries the distinction as a rule — brand green is
  fill-only, income green is text-only, never adjacent in dark theme, and charts
  never use income green. Automated contrast checks compare text against
  background and cannot catch a violation of this.
- `density = compact` is never applied at `device = compact`: touch floors at
  `density-row-min-touch` (44) and the control is hidden rather than disabled.
- The canvas stores sizes as literals because this pen build silently drops
  variable references on `width` and `height`. Tokens are authoritative in code;
  where a drawn size and its token disagree, the token wins.
- `tokens.css` is generated from the library by a manual pencil-MCP step. CI
  cannot read `.pen` files, so drift is caught by a committed token-name snapshot
  rather than by regenerating.
