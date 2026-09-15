# 0015 · The toolchain tracks the latest stable release

## Status

Accepted

## Decision

Every package in `devbox.json` is pinned to the highest **stable** major
available from the devbox search index, and every one of them carries a major
pin — `postgresql@18`, never `postgresql` and never `postgresql@18.6`.

Stable excludes anything carrying `rc`, `beta`, `alpha` or a `-pre` suffix.
Python is the case that makes the word do work: `3.15.0rc1` is published and
`3.14` is the pin.

Bumping is part of any change that touches `devbox.json`. A change adding a
package checks every other pin in the same edit, so drift is caught by the work
already in flight rather than by a scheduled sweep nobody runs.

## Context

The repo pins nine tools and adds more as capabilities arrive. Three of them had
drifted a major behind the index — `nodejs@24` against 26, `pnpm@10` against 11 —
without anything noticing, because a pin that resolves is a pin that looks fine.

Drift is cheapest to pay down while it is one major. Two majors of Node is a
migration; one is a version string. Deferring the bump converts a diff into a
project.

The alternative is a lower bound with no upper — `nodejs` unpinned, resolving to
whatever the index holds that day. Rejected: `devbox.lock` would then be the only
statement of what the toolchain is, and a reviewer reading `devbox.json` would
learn nothing. The major pin is what makes the bump a reviewable line.

The other alternative is a conservative track — latest minus one major, or an LTS
line where the project publishes one. Rejected for a single-household product
with one contributor: there is no fleet to stagger, no customer to protect from a
regression, and the cost of being wrong is a `devbox install` away from being
undone.

## Consequences

- A new major appears and the pin moves in one line. `devbox install` and
  `just ci` are the whole verification.
- Breakage arrives early, one tool at a time, with the change that bumped it
  standing next to it in the diff. This is the trade being made: a small,
  attributable break instead of a large, undated one.
- The patch level is `devbox.lock`'s business. It moves without a decision and
  without a diff in `devbox.json`.
- Nothing enforces this. There is no check that compares a pin against the index,
  because the index is a network call and `just ci` does not make one. The rule
  holds by being read at the moment `devbox.json` is edited.
