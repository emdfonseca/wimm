# 0006 · A more vivid palette, and separating the two greens

## Status

Accepted

Revises the palette recorded in ADR 0002. The typefaces, token naming, theme axes
and the reservation of green and red for money direction are unchanged.

## Context

The brand green was asked to be more vivid in the same direction. Measuring the
candidates turned up something the vividness question was sitting on top of.

**Brand green and income green are the same colour.** In light theme the brand is
hue 157 and income hue 155; in dark, 150 and 151. They are also close in
lightness — 1.24:1 in dark. Two channels, no separation on either. This is what
the fill-only / text-only rule in ADR 0002 exists to work around, and the
workaround was doing all of the work.

Lightening the brand makes it worse before it makes it better. At `#0A6B4C` the
light-theme brand/income ratio falls from 2.05 to 1.15, and at `#0C7A57` to 1.07 —
so a lighter brand alone would extend the dark-theme collision into light theme
as well.

## Decision

Three things move together, because moving any one of them alone is worse than
moving none.

```text
brand    #14402F / #9FE0C0  →  #0C7A57 / #5FE7B8
accent   #2563EB / #6699FF  →  #1F5FF0 / #7AA8FF
income   #067647 / #3DD68C  →  #127A33 / #4ADB6E
```

`#0C7A57` is the lightest the brand can be: white on it measures 5.33, and one
step lighter (`#0E8A62`) drops the button label to 4.35 and fails. Income moves
off the brand's hue rather than its lightness — 139 and 135 against the brand's
161 and 159, a gap of 22° and 24° where it was 2° and 1°. Every threshold holds:
white on brand 5.33, brand on canvas 4.95, accent as text 5.31 and 4.92, income
text 5.44 and 9.64, and the dark equivalents higher.

Sixteen tokens move, because the brand, accent and income families each carry
derivatives — hover and active states, the chart ramp's first two slots, the
focus ring, the info and success feedback pairs, and the hero gradient.

**The fill-only / text-only rule stays.** It is tempting to retire it now that the
greens differ by hue, and that would be wrong: a 22° separation inside the green
band is close to invisible under deuteranopia, which is the condition the rule was
protecting against in the first place. Hue separation helps typical vision and
does nothing for the case that mattered. What actually carries direction is the
sign: every amount shows `+` or `−`, so no reading of the ledger depends on
telling two greens apart.

Two tokens are added rather than adjusted. `color-text-on-brand` and
`color-text-on-brand-secondary` replace two colours that were hard-coded onto the
balance card, and which the new gradient broke — the card was carrying its own
palette and nothing said so until the contrast pass failed on it.

The hero gradient does not follow the brand all the way. `#0E6E4F` rather than
`#12805C`, because at the lighter value no secondary tone clears 4.5 against it:
the lightest candidate reached 4.44. The gradient is still far more saturated
than it was — 78% against 51% — at about the same lightness.

## Alternatives rejected

**Keep today's palette.** Rejected only because a change was asked for; it is the
most conservative option and the one with the largest brand/income lightness
separation in light theme, at 2.05.

**Vivid brand at its original darkness** (`#0A6B4C` / `#5FE7B8`, income
unchanged). More vivid, all thresholds clear, and leaves the brand/income
collision exactly where it was — 6° of hue in light, 8° in dark. Rejected for
spending the change without buying the fix.

Both are drawn beside the adopted palette in `90 · QA / Palette · options
considered`, hard-coded so they stay legible now that the tokens have moved.

## Consequences

- The contrast pass over both themes reports 30 failures, all
  `color-text-disabled` on disabled specimens, exempt under SC 1.4.3 — the same
  30 as before the palette moved.
- The balance card no longer carries hard-coded colour.
- Anything that assumed income green and success green are `#067647` needs
  regenerating rather than editing; both come from tokens.
