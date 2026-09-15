# Project Setup Record — wimm

Project-level design decisions. Journey files carry a short summary and their own
`00 · JOURNEY OVERVIEW`; they never duplicate this record.

## Product

**Name**
`wimm` — where is my money? The lock-up ends in the question mark the initialism
stands for, set in the accent blue. The mark alone never carries it: a lone `?`
in a rounded square is a help affordance everywhere else.

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
Edit transaction       → Drawer     / route-backed   (renders per regime, below)
Recategorise           → Inline edit
Split transaction      → Full page   / route-backed   (likely to grow)
Edit budget            → Drawer     / route-backed   (renders per regime, below)
Delete anything money-bearing → Destructive confirmation dialog naming the record
```

Destructive actions never reuse the editing drawer.

**One editing surface, three renderings.** Edit transaction is one route and one
Drawer component throughout, and modality is a property of the rendering rather
than of the component — which is why the table above states the route and defers
the rest to here.

- **Compact** — an in-page replacement. The list is not rendered, so there is
  nothing behind it: no scrim, nothing to make inert, no focus to trap. Back
  returns to the list with the row restored. At 390 there is no list worth
  keeping visible, so keeping it visible is not a goal worth paying for.
- **Medium and Wide** — a right-anchored overlay at `layout-drawer-width` 400,
  **non-modal**: the ledger stays visible *and interactive* behind it, which is
  the whole reason to prefer it over a dialog. Focus moves in on open and returns
  to the row on close.
- **Ultra** — a persistent inspector pane, with the elevation and the close
  control removed because nothing is dismissed.

Only the middle one has anything behind it, so only the middle one has a modality
question to answer. A full-bleed sheet inheriting the non-modal contract would be
the worst of both: covering the page while leaving focus free to wander behind it.
Nothing is reachable at one width and not another.

**View preferences live in the account menu.** Appearance (light / dark / system)
and density (comfortable / compact) are set there, reached from the account
trigger in the sidebar footer. Density is hidden under `any-pointer: coarse`
(ADR 0004). Sidebar collapse is the exception: its control sits in the sidebar
footer beside the account trigger, because it acts on the thing it sits in.
Width sets the default form — rail below 1200 CSS px — and the control overrides
it for the session.

## Responsive

**Regimes drawn**

```text
Compact → 390 px representative frame   (width < 768 CSS px)
Medium  → 1024 px representative frame  (768 ≤ width < 1200 CSS px)
Wide    → 1440 px representative frame  (1200 ≤ width < 1800 CSS px)
Ultra   → 1920 px representative frame  (width ≥ 1800 CSS px)
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
Ultra buys a second pane, not bigger controls: control heights and type sizes
resolve identically to Wide. A surface with nothing to put in a second pane is not
drawn at Ultra — it uses the Wide composition, and that mapping is recorded rather
than duplicated.

`layout-content-max` (1200) caps **text measure only** — prose, help, long-form
settings copy. It does not cap panels, tables or chart regions, which take the
full Main column. Capping everything cost 212 px of a 1920 screen and made the
extra width buy nothing.

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

**What is measured, and what is only promised.** The distinction matters more than
the list: a measured claim is re-checked by `canvas-audit.md` on every pass, and a
promised one is a test waiting for code that does not exist yet.

```text
measured   target size      292 control instances, 0 under 24 × 24, nearest two
                            centres 37 px apart, so the spacing exception is not
                            relied on anywhere
measured   reflow at 320    two 320 px frames carrying the longest realistic
                            strings, 0 nodes crossing either edge
measured   text contrast    1985 text nodes, 30 failures, all color-text-disabled
                            and exempt under SC 1.4.3
derived    focus order      read out of the canvas per surface, in 50 · TEMPLATES
                            → Focus order; stops a frame does not draw are marked
promised   keyboard         40 · ORGANISMS → Keyboard contract: operability, focus
                            transitions, error identification and status
                            announcement, one row per organism, none of it
                            measurable until there is markup
```

**Skip link.** `20 · ATOMS` → Skip link, a preset of Button, first in the tab
order of every shell and off-screen until focused. It exists because the Wide
shell puts twelve tab stops between the top of the page and the first ledger row.
It is deliberately absent from the shell templates: a control that is invisible
until focused, drawn always-visible in a composition template, teaches the
composition wrong. Its position is recorded in the Focus order block instead.

**The Compact page header is its own component.** `App header/compact` drops the
breadcrumb and the tab counts and renders the primary action as an icon button
with the label as its accessible name. At 320 CSS px a 141 px labelled action left
90 px for the page title, which broke mid-word. Repeating those overrides per
screen was the alternative, and it is the drift the one-component rule exists to
prevent.

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
`layout-subnav-width` (212) is the in-page subsection list, identical at every
pointer regime.

**Typography** — Schibsted Grotesk for the interface: a newspaper-grade grotesque
that stays quiet at 13 px and carries character at 32 px. IBM Plex Mono for amounts,
account identifiers, and dates in tables: slashed zero, unambiguous `1`, and columns
that line up.

`type-family-display` resolves to the same family as the body face. Display is a
treatment, not a second font: weight 600 and −2% tracking, permitted on page titles
and hero amounts only. Hero amounts stay in IBM Plex Mono — a proportional face
would cost the decimal alignment that makes the ledger readable.

**Palette** — Pine & Signal. Deep pine green (`#0C7A57`) is the brand and primary
action; blue (`#1F5FF0`) is the highlight for links, selection and focus;
neutrals are tinted green-grey. Revised in ADR 0006 — see it for the values, the
alternatives, and the reasoning; they are not repeated here.

Red is reserved for money direction and feedback, never for branding. Green
carries both jobs and they are kept apart by role: `color-action-primary` is the
brand, `color-amount-positive` is income, and neither is ever used for the
other's purpose.

**The two greens are separated by hue, not by lightness.** Brand green sits at
hue 161/159 and income green at 139/135, a gap of 22° and 24°. By lightness they
are still close — 1.17:1 in dark — so the contextual rule stands unchanged:

```text
Brand green   fill only   — buttons, selected nav, chart series 1
Income green  text only   — amounts, always with a + sign
Never adjacent in dark theme.
Charts never use income green; the green slot in the chart ramp is teal.
```

That rule is not redundant now that the hues differ. A 22° separation inside the
green band is close to invisible under deuteranopia, which is the case the rule
was protecting in the first place — hue separation helps typical vision and does
nothing for that one. What carries direction is the sign: every amount shows `+`
or `−`, so no reading of the ledger depends on telling two greens apart.

**Text on the brand surface has its own tokens.** `color-text-on-brand` and
`color-text-on-brand-secondary` are theme-invariant, because the brand surface is
dark in both themes. The balance card used to hard-code these and the values
broke the moment the gradient moved.

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
