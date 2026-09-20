# Project Setup Record — wimm

Project-level design decisions. A flow's own decisions live in its spec under
`openspec/specs/`; they never duplicate this record.

## Product

**Name**
`wimm` — where is my money? The lock-up ends in the question mark the initialism
stands for, set in the accent blue. The mark alone never carries it: a lone `?`
in a rounded square is a help affordance everywhere else.

**Problem / in scope / out of scope**
Household money management: where the money is, where it went, and whether the plan
holds. In scope: accounts and balances, transactions, categorisation, budgets,
recurring items, reporting. Out of scope until decided: investment performance, tax.

**Primary actor + permission differences**
A household member. The operator who runs the instance registers members; nobody
signs themselves up (ADR 0016, `openspec/specs/identity/operator-registration/spec.md`).
An account has owners, who see it in full. Every other member sees it at a level
set per account: hidden, balance or details (ADR 0019). Transactions are
owner-only (ADR 0021).

**Primary flow + success outcome**
Connect a bank, choose which accounts the household sees, then read balances on
Overview and transactions in the ledger
(`openspec/specs/banking/bank-connections/spec.md`,
`openspec/specs/banking/overview/spec.md`,
`openspec/specs/banking/transactions/spec.md`).

**Data source**
Accounts, balances and transactions are read from banks through a gateway
(ADR 0018, ADR 0021). Nothing is entered by hand or imported from a file.

**Currency**
Money is minor units plus an ISO 4217 code. Totals are per currency and mixed
currencies are never summed (ADR 0018).

**Failure / recovery paths to design**
Per flow. Domain-wide ones known already: account sync failure and stale data,
duplicate transaction, failed categorisation rule, budget exceeded.

## Navigation and surfaces

**Entry route + navigation hierarchy**
Four destinations, one list for every form of the navigation
(`packages/ui/src/destinations.ts`): Overview `/`, Accounts `/accounts`,
Transactions `/transactions`, Settings `/settings`. Connecting a bank is a flow
under `/connect`. The navigation is a bottom bar at Compact, a rail at Medium and
a labelled sidebar at Wide and Ultra (`openspec/specs/app/screen-layout/spec.md`).

**Back / cancel / success destinations**
Per flow.

**Task surfaces**

```text
Connect a bank         → Full page   / route-backed   (multi-step, resumable)
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
- **Ultra** — a persistent inspector pane, with no elevation and no close
  control because nothing is dismissed.

Only the middle one has anything behind it, so only the middle one has a modality
question to answer. A full-bleed sheet inheriting the non-modal contract would be
the worst of both: covering the page while leaving focus free to wander behind it.
Nothing is reachable at one width and not another.

**View preferences live in Settings.** Appearance (light / dark / system) and
density (comfortable / compact) are set there and nowhere else
(`openspec/specs/app/settings/spec.md`). Every choice takes effect at once, with
no saving step. Density is hidden under `any-pointer: coarse` (ADR 0004).

## Responsive

**Regimes**

```text
Compact → width < 768 CSS px            (checked at 390)
Medium  → 768 ≤ width < 1200 CSS px     (checked at 1024)
Wide    → 1200 ≤ width < 1800 CSS px    (checked at 1440)
Ultra   → width ≥ 1800 CSS px           (checked at 1920)
```

Ultra is a fourth regime, see ADR 0002. It earns its place because one structure
genuinely changes: the editing drawer stops being an overlay and becomes a
persistent inspector pane, so a transaction can be edited without losing the list.

`just canvas` shows every screen and its state stories at all four regimes.

**Structural change points (CSS px)**
768 — Compact stacked layout becomes the shell with sidebar.
1200 — icon rail becomes the full labelled sidebar.
1800 — the overlay drawer becomes a persistent inspector pane.

**What Ultra does not do**
Ultra buys a second pane, not bigger controls: control heights and type sizes
resolve identically to Wide. A surface with nothing to put in a second pane uses
the Wide composition at Ultra.

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
hero-amount size, and control height. Its values switch at the structural change
points above.

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

**Skip link.** A skip link is first in the tab order of every shell and
off-screen until focused. It exists because the Wide shell puts twelve tab stops
between the top of the page and the first ledger row.

**The Compact page header is its own component.** It drops the breadcrumb and the
tab counts and renders the primary action as an icon button with the label as its
accessible name. At 320 CSS px a 141 px labelled action left 90 px for the page
title, which broke mid-word. Repeating those overrides per screen was the
alternative, and it is the drift the one-component rule exists to prevent.

## System

**Design of record** — code (ADR 0023). A screen's design is its presentational
Svelte component under `packages/ui/src/pages` plus its state stories in
Storybook. Components live in
`packages/ui/src/{atoms,molecules,organisms,templates,pages}`.

**Tokens** — `packages/ui/design/tokens.json` is the hand-edited source of token
values. `packages/ui/src/tokens.css` is generated from it by `just gen`
(ADR 0003).

**Surfaces** — `docs/design/surfaces.md`: which navigation pattern, task surface
and feedback surface to use for what, plus the rulings specific to money (no toast
for a failed money operation; direction is a control, not a typed minus; the
currency symbol is an affix, not content).

**Token naming** — kebab-case (`color-bg-canvas`), so each name maps 1:1 to a CSS
custom property (`--color-bg-canvas`).

**Domain tokens beyond the standard vocabulary**
Amounts: `color-amount-positive`, `color-amount-negative`, `color-amount-neutral`,
and the `type-size-amount-*` scale.
Accent: `color-accent`, `color-accent-hover`, `color-accent-subtle`.
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
action; blue (`#1F5FF0`) is the highlight for links, selection and focus; light
neutrals are tinted green-grey and the dark ground is a cool near-black. See
ADR 0006 for the light values and ADR 0014 for the dark ones, with the
alternatives and the reasoning; they are not repeated here.

Red is reserved for money direction and feedback, never for branding. Green
carries both jobs and they are kept apart by role: `color-action-primary` is the
brand, `color-amount-positive` is income, and neither is ever used for the
other's purpose.

**The two greens are separated by hue, not by lightness.** Brand green sits at
hue 161/159 and income green at 139/135, a gap of 22° and 24°. By lightness they
are still close — 1.17:1 in dark — so the contextual rule stands:

```text
Brand green   fill only   — buttons, selected nav, chart series 1
Income green  text only   — amounts, always with a + sign
Never adjacent in dark theme.
Charts never use income green; the green slot in the chart ramp is teal.
```

That rule is not redundant given that the hues differ. A 22° separation inside the
green band is close to invisible under deuteranopia, which is the case the rule
protects — hue separation helps typical vision and does nothing for that one. What
carries direction is the sign: every amount shows `+` or `−`, so no reading of the
ledger depends on telling two greens apart.

**Text on the brand surface has its own tokens.** `color-text-on-brand` and
`color-text-on-brand-secondary` are theme-invariant, because the brand surface is
dark in both themes. The balance card takes its text colours from them and never
hard-codes its own.

## Constraints

**Routing / data / platform constraints**
SvelteKit with Svelte 5 (ADR 0001). Pages are presentational components taking data
as props and emitting intent as callbacks; a thin route module supplies the data.
Every designed state must be reachable by setting props.
