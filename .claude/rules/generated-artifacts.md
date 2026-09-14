# Generated artifacts

Rules for any file a script writes from another file — stylesheets from tokens,
clients from schemas, indexes from directories. Every one of these was learned
from a defect that shipped past a green check.

## Validate before writing, not after

A generator that writes first and validates later leaves bad output on disk when
it fails. The build exits 0, the stale-looking file is newer than the source, and
the damage outlives the run that caused it.

Refuse to write. Validate the input, report every problem, exit non-zero, touch
nothing. A build step that cannot produce correct output should produce none.

## Paired checks must read one source of truth

Completeness and capability are different questions, and asking them against
different lists is how a value passes both and still vanishes.

A document declaring three branches where the generator renders two satisfies a
completeness check (every token supplies all three) and a name check (the axis is
known) while the third branch is silently dropped. The fix is not a third check —
it is one declaration of what the generator supports, exported by the generator
and imported by the validator, asserted in both directions: nothing declared that
cannot be rendered, nothing renderable left undeclared.

## Keep the fixture, not just the fix

When a check misses something, fixing the check is half the work. The other half
is a committed test that feeds it the bad input and asserts rejection.

Checking that the *current* document passes proves nothing about any of it,
because the current document is valid. A regression suite asserts the opposite:
these specific malformed things are refused, and refused for the stated reason.

Assert two properties per case, not one: the run failed, **and** the output file
was not modified. A generator that writes garbage and then exits non-zero still
corrupted the artifact.

## Test the promise, not only the values

Value validation catches a malformed colour. It says nothing about whether the
documented *behaviour* reached the output — and behaviour is what the
documentation actually promises.

A token set can be entirely valid while the stylesheet has no reduced-motion
branch, no explicit `color-scheme` per theme, or only one of two density
selectors. Every value checks out; every documented guarantee is absent.

So assert the contract against the generated artifact: the branch exists, the
tokens it should override are overridden, the selectors the documentation names
are present. These tests are short and they are the only thing standing between a
written policy and a file that quietly does not implement it.

## Name the guarantee precisely

"CI verifies the design system" is the kind of claim that survives until someone
depends on it. State what the check actually covers and what it cannot.

Where a source is unreadable by CI — an encrypted file, a hosted service, a
desktop-only export — say so in the decision record, name the human step that
substitutes for it, and make that step part of the same commit. Provenance is not
proof of synchronisation.

## One command shape, and make it work

A documented command that does not run is worse than an undocumented one: it
sends a reader looking for their own mistake. Pick the public shape, implement
it, use it in every repair message, and run it once before writing it down.

Chain the steps so a failure propagates. `a; b` reports b's status and hides a's.
