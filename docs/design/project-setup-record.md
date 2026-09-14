# Project Setup Record — wimm

Project-level design decisions. Journey files carry a short summary and their own
`00 · JOURNEY OVERVIEW`; they never duplicate this record.

## Product

**Problem / in scope / out of scope**
Personal money management: where the money is, where it went, and whether the plan
holds. In scope: accounts and balances, transactions, categorisation, budgets,
recurring items, reporting. Out of scope until decided: multi-user households,
investment performance, tax.

**Primary actor + permission differences**
A single person managing their own money. One role, no permission tiers.
OPEN DECISION — shared household access.

**Primary journey + success outcome**
Not yet scoped. The library is being built ahead of the first journey.
OPEN DECISION — first journey (assumption: connect/add an account, then review and
categorise transactions).

**Failure / recovery paths to design**
Per journey. Domain-wide ones known already: account sync failure and stale data,
duplicate transaction, failed categorisation rule, budget exceeded.

## Navigation and surfaces

**Entry route + navigation hierarchy**
OPEN DECISION — the route map. Assumption: a top-level area per noun
(`/accounts`, `/transactions`, `/budgets`, `/reports`, `/settings`), each with a
list and a detail route.

**Back / cancel / success destinations**
Per journey.

**Task surfaces**

```text
Add account            → Full page   / route-backed   (multi-step, resumable)
Add transaction        → Modal       / ephemeral      (short, blocking)
Edit transaction       → Drawer, non-modal / route-backed (context stays visible)
Recategorise           → Inline edit
Split transaction      → Full page   / route-backed   (likely to grow)
Edit budget            → Drawer, non-modal / route-backed
Delete anything money-bearing → Destructive confirmation dialog naming the record
```

Destructive actions never reuse the editing drawer.

## Responsive

**Regimes drawn**

```text
Compact → 390 px representative frame   (width < 768 CSS px)
Wide    → 1440 px representative frame  (768 ≤ width < 1800 CSS px)
Ultra   → 1920 px representative frame  (width ≥ 1800 CSS px)
Medium  → no separate canvas row. Shares Wide structure; the sidebar renders as an
          icon rail below 1200 CSS px. Recorded here rather than drawn.
```

Ultra is a deviation from the pen-design standard's canonical three regimes
(§5.3.1). It earns a fourth regime because one structure genuinely changes: the
editing drawer stops being an overlay and becomes a persistent inspector pane, so
a transaction can be edited without losing the list.

**Structural change points (CSS px)**
768 — Compact stacked layout becomes the shell with sidebar.
1200 — icon rail becomes the full labelled sidebar.
1800 — the overlay drawer becomes a persistent inspector pane.

**What Ultra does not do**
The content column never exceeds `layout-content-max` (1200). Extra width becomes
gutter, never longer lines or wider table cells. Ultra buys a second pane, not
bigger controls: control heights and type sizes resolve identically to Wide.
A surface with nothing to put in a second pane is not drawn at Ultra — it uses the
Wide composition centred, and that mapping is recorded rather than duplicated.

## Themes

**Color axis** — `Light`, `Dark`. Light is the document default.

**Control heights** — 30 / 36 / 44 on pointer regimes, 36 / 44 / 52 on Compact.
36 is a comfortable mouse target and buys back four pixels on every control in a
filter bar; a thumb needs the larger size. Both clear the 24 px AA floor.

**Density axis** — yes: `Comfortable`, `Compact`. Independent of colour and device.
Drives row height, cell padding, stack and section gaps. In a ledger the row height
is the most consequential number in the system: at 1440 × 900 roughly 684 px of
table area shows 12 comfortable rows or 17 compact ones.

`Density = Compact` is never honoured on touch. A 40 px row is a fine pointer
target and a poor thumb target, so where the input is coarse the density values
revert to comfortable whatever the setting says, and the control that sets them is
hidden rather than shown disabled.

**Touch is an input capability, not a width.** The CSS keys this to
`any-pointer: coarse`, not to a breakpoint: a large tablet is a wide viewport with
coarse input, and a narrow desktop window is a small viewport with fine input.
Keying it to width gets both backwards. Layout regimes stay width-driven — that is
what they describe — but the target floor does not.

This is stricter than WCAG requires: 40 px already clears the 24 px AA minimum.
The floor exists because a ledger row is a frequent, precise target and 44 px is
the product's own standard, not because the smaller value fails.

**Device axis** — yes: `Compact`, `Medium`, `Wide`, `Ultra`. Drives page gutter,
section gap, header height, sidebar width, inspector width, page-title and
hero-amount size, and control height.
It is not a breakpoint; a frame's width and its `Device` value are set independently.

## Accessibility

WCAG 2.2 AA is the baseline for the default experience, never an axis or a mode.

Project-specific requirements:

- Amount sign is never carried by colour alone. Every amount shows `+`/`−` and a
  currency, so income/expense read without colour vision.
- Amounts use tabular figures and align on the decimal in any column.
- Deleting or altering a money-bearing record is a consequential submission
  (SC 3.3.4): review-and-confirm naming the record, or a reversal.
- Charts state the value in text near the mark; never colour alone.
- Account numbers are masked by default and revealed by an explicit control.

## System

**Library path + revision** — `packages/ui/design/product-ui.lib.pen`.

**Surfaces** — `docs/design/surfaces.md`: which navigation pattern, task surface
and feedback surface to use for what, plus the rulings specific to money (no toast
for a failed money operation; direction is a control, not a typed minus; the
currency symbol is an affix, not content).

**Conventions** — `docs/design/library-conventions.md`: component block structure
(contract, variants and states together), the specimen-cell contract that makes
alignment structural, and the required state matrices per component.

**Verification** — `docs/design/canvas-audit.md` holds the audit pass: overflow,
dropped size bindings, hugging controls, nested focus-ring radius, text contrast
against the resolved background, and text sized without `textGrowth`. Run it as its
own `execute` call after every canvas change. Results read inside the mutating call
are stale — bounds, `ctx.problems` and screenshots all report the pre-layout frame.

**Known tool limitation** — this pen build silently drops variable references on
`width` and `height`: the property is discarded and the frame falls back to hugging
its content. Every size on the canvas is therefore a literal matching its token
(`control-height-md` drawn as 40, `layout-sidebar-width` as 264). Colour, padding,
corner radius, stroke width and type size bind normally. In code the tokens are
authoritative; where a drawn size and its token disagree, the token wins.

**Token naming** — the library uses kebab-case variable names (`color-bg-canvas`)
rather than the dotted form in the pen-design standard's §8.2 vocabulary, so each
name maps 1:1 to a CSS custom property (`--color-bg-canvas`). Roles and structure
are unchanged; only the separator differs.

**New shared assets / token changes expected**
Domain tokens beyond the standard vocabulary: `color-amount-positive`,
`color-amount-negative`, `color-amount-neutral`, and the `type-size-amount-*` scale.
Accent tokens for the teal family: `color-accent`, `color-accent-hover`,
`color-accent-subtle`.
Layout tokens for the two-pane regime: `layout-inspector-width` (0 until Ultra) and
`layout-drawer-width` (the overlay width at Compact through Wide).

**Typography** — Schibsted Grotesk for the interface: a newspaper-grade grotesque
that stays quiet at 13 px and carries character at 32 px. IBM Plex Mono for amounts,
account identifiers, and dates in tables: slashed zero, unambiguous `1`, and columns
that line up.

`type-family-display` resolves to the same family as the body face. Display is a
treatment, not a second font: weight 600 and −2% tracking, permitted on page titles
and hero amounts only. Hero amounts stay in IBM Plex Mono — a proportional face
would cost the decimal alignment that makes the ledger readable.

**Palette** — Pine & Signal. Deep pine green (`#14402F`, 11.6:1 on white) is the
brand and primary action; blue (`#2563EB`) is the highlight for links, selection and
focus; neutrals are tinted green-grey.

Green and red are reserved for money direction and feedback. Pine is dark and flat
where income green is bright and saturated, and the two separate at 2.05:1 in light
theme.

**The dark-theme constraint.** In dark theme the brand resolves to mint `#9FE0C0`
and income to `#3DD68C`, which separate at only **1.24:1** — effectively
indistinguishable by lightness. Both alternatives fail: a deep pine button reaches
just 2.16:1 against the near-black canvas, below the 3:1 that SC 1.4.11 requires for
the control to be identifiable at all; desaturating far enough to separate by
lightness washes the brand out to near-white.

Context is therefore what keeps them apart, and it is a rule, not a hope:

```text
Brand green   fill only   — buttons, selected nav, chart series 1
Income green  text only   — amounts, always with a + sign
Never adjacent in dark theme.
Charts never use income green; the green slot in the chart ramp is teal.
```

Any design that puts a pine-brand fill beside an income amount in dark theme is
wrong even though every contrast check passes, because the checks measure text
against its background and this is a fill against a fill.

## Constraints

**Routing / data / platform constraints**
SvelteKit with Svelte 5 (ADR 0001). Pages are presentational components taking data
as props and emitting intent as callbacks; a thin route module supplies the data.
Every designed state must be reachable by setting props.

**pen.dev app/extension version** — `.pen` schema 2.14.

## Open decisions

```text
Question:            What is the first journey to design?
Current assumption:  Add an account, then review and categorise transactions.
If wrong:            The component inventory ordering shifts; foundations and atoms
                     are unaffected.

Question:            Is account data manually entered, imported, or bank-synced?
Current assumption:  Manual entry and file import first; sync later.
If wrong:            Sync adds connection status, re-auth, and stale-data states to
                     the shell and the account organism.

Question:            Single currency or multi-currency?
Current assumption:  Single currency per user, formatted by locale.
If wrong:            Amount components need a currency slot and conversion
                     disclosure; the token scale is unaffected.

Question:            Shared household access?
Current assumption:  No. One person, one dataset.
If wrong:            Adds an actor, permission states, and an invite journey.
```
