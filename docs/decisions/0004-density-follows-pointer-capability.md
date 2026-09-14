# 0004 · Density availability follows pointer capability

## Status

Accepted

## Context

ADR 0002 introduced the density axis and recorded, as a consequence, that
`density = compact` is never applied at `device = compact` — tying the rule to a
viewport width. The implementation now keys it to input capability instead, and
the two statements contradict each other.

Width is a poor proxy for touch in both directions. A large tablet is a wide
viewport with coarse input and would receive pointer-sized rows. A narrow desktop
window is a small viewport with fine input and would lose a density setting that
suits it perfectly well.

## Decision

Layout regimes stay width-driven: the `device` axis describes available space,
which is what a width query measures.

The touch floor does not. `density = compact` reverts to comfortable under
`@media (any-pointer: coarse)`, and the control that sets density is hidden where
it does not apply rather than shown disabled — a greyed control invites someone to
work out how to enable it.

This supersedes the density consequence in ADR 0002. The three-axis decision
itself is unchanged.

## Consequences

- A hybrid device with both a trackpad and a touchscreen reports coarse pointer
  availability and therefore gets the comfortable floor. That is the safe
  direction to be wrong in: the cost is a few rows per screen, against a target
  that is hard to hit.
- The floor is stricter than WCAG requires. 40 px already clears the 24 px AA
  minimum; 44 px is the product's own standard for a frequent, precise target in
  a ledger, not a conformance threshold.
- Any future capability-driven behaviour follows the same split: describe space
  with width, describe input with pointer queries, and never let one stand in for
  the other.
