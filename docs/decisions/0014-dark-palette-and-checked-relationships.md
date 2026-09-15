# 0014 · A neutral dark palette, and a check that can see it

## Status

Accepted

## Context

The dark theme had never been looked at. Every pair in it passed WCAG 2.2 AA —
most comfortably, several above 10:1 — and `just check packages/ui` reported a
clean run on every commit. It was still wrong, in ways no threshold expresses:

```text
gradient-brand-to  #07120D  ==  color-bg-canvas  #07120D     ratio 1.00
color-action-primary #5FE7B8 against the canvas               ratio 12.36
color-feedback-error-bg against color-bg-surface              ratio 1.04
```

The brand panel's gradient ended on exactly the page background, so a split
layout dissolved into the page. The primary button was the brightest thing on
screen by a distance, where the same button in light sits at 4.95. A notice was
indistinguishable from the card beneath it.

Underneath those three, a fourth problem: every dark surface was green. The
canvas, the cards, the borders and the sidebar all carried the brand hue, so the
blue accent landed as a foreign object and the green button had nothing to stand
out from.

ADR 0003 scopes CI to two questions — is the token document well formed, and does
the stylesheet agree with it. Neither can see any of this, because all of it is
relationships between values rather than properties of one. Contrast was never
checked either.

## Decision

**Dark is neutral. Green and blue are highlights on it, not the ground it is made
of.** Surfaces run from `#0A0B0C` to `#2A2E33` as a near-black grey ramp, and the
only saturated areas on a dark screen are the ones carrying meaning: the brand
panel's gradient, the primary action, the selected row, links, amounts and charts.

```text
bg-canvas       #07120D → #0A0B0C
bg-surface      #0F1D17 → #141618      1.10 → 1.30 against the canvas
bg-subtle       #16281F → #1A1C1F
bg-elevated     #16281F → #1E2124
border-default  #1E3529 → #2E3338      1.27 → 1.62 against the surface
text-primary    #E6F0EA → #E9EBEE
text-secondary  #93AFA2 → #9BA3AB
accent-subtle   #0E1E3D → #17263F      1.01 → 1.29 against the surface
action-primary  #5FE7B8 → #38B48C      12.36 → 6.57 against the canvas
gradient-to     #07120D → #0C2E22      1.00 → 3.05 against the canvas
```

Hover, active, secondary, disabled, control and feedback-background derivatives
follow their families. `color-chart-1` follows the brand, so dark carries one
green rather than two.

**The brand gradient stays green.** It is the one branded surface in the product
and the only large area of colour on a dark screen, which is what makes it read
as branding rather than as decoration.

**The inverted button stays, and it is not a preference.** It is tempting to give
dark the same treatment as light — white label on a deep green fill — and it
cannot be done. A green dark enough to carry white text at 4.5 tops out around
`#12805C`, which measures 3.88 against the dark canvas: dimmer than light's own
button at 4.95, on a page where dimmer means invisible. In dark the fill has to be
light, and a light fill demands a dark label. The remedy for a button that shouts
is a quieter mint, not a return to white on green.

**A fill has a ceiling as well as a floor.** `color-action-primary` must clear 3.0
against the canvas so it reads as an action, and must not exceed 9.5, beyond which
it stops reading as a control and starts reading as a light source. This is the
first threshold in the system with a maximum, and it is the one that would have
caught what shipped.

**Separation is satisfied by fill or by border, not by both.** A card in light is
`#FFFFFF` on `#F4F7F5` — a ratio of 1.08 — and is perfectly legible because its
border does the work. The same rule applied to dark, where luminance does the
work, would fail light for doing it the other way round. Each separation rule
names the candidates that could carry it and requires that one of them does.

**`packages/ui/scripts/check-palette.py` asserts all of this, in every theme the
document declares**, and runs inside `just check packages/ui`: eighteen contrast
pairs, seven separation rules and one loudness ceiling, 52 relationships across
light and dark.

**Its fixture is the palette that shipped.**
`scripts/fixtures/dark-before-0014.json` holds the dark values as they were, and
`scripts/test-palette.py` asserts they are rejected, alongside five constructed
failures. Checking that the current document passes proves nothing about the
checker, because the current document is valid.

This supersedes the dark half of ADR 0006's palette. The light values, the
fill-only / text-only rule, and the reasoning about hue separation under
deuteranopia are unchanged.

## Consequences

- **The pen library and `tokens.json` are updated together.** ADR 0003 makes the
  export the contract and leaves library-to-export agreement to a person; a change
  that edits `tokens.json` without pushing the same values into
  `product-ui.lib.pen` leaves every canvas rendering a palette the product no
  longer has.
- **A constructed fixture derives its values from the document.** A fixture that
  hard-codes "the surface colour" stops reproducing its defect the moment the
  surface moves, and then asserts nothing while still reporting a pass. This
  happened once already, between the green dark palette and this one.
- **A new colour token needs a rule in `check-palette.py` or it is unchecked.**
  The file lists pairs by name, so a token nothing names is invisible to it.
- **The ceiling is a judgement, and 9.5 is where it was set.** No standard has an
  opinion about a button being too bright. Moving it is a decision, and moving it
  to accommodate a value rather than to correct the rule is how a check stops
  meaning anything.
- **The checker cannot see hue.** Every rule here is a luminance ratio, so a
  palette can satisfy all 52 relationships and still put a navy block on a green
  page — which is the defect that prompted the neutral ramp, and which no
  assertion here would have caught.
