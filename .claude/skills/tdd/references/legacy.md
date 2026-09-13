# Legacy code

Code with no tests and no seams. The greenfield cycle does not apply directly because you cannot write a meaningful failing test against something you cannot instantiate in isolation. The sequence changes:

```text
1. CHARACTERIZE   pin current behaviour, right or wrong, as a safety net
2. SEAM           the smallest change that lets you test one piece in isolation
3. TDD            new behaviour, with the greenfield cycle, on the now-testable piece
```

## Characterization tests

Call the code as it is, with real-ish inputs, and assert on whatever it returns — including the bug you are about to fix. The test's job is to fail if a refactor changes behaviour you did not mean to change. Name it for what it captures: `TestLegacyPricing_CurrentBehaviour`. When the intended fix lands, the assertion changes, deliberately, in its own cycle.

Approval/golden tests are the fast way to characterise large outputs: capture once, diff on every run.

## Finding a seam

The place where behaviour can be changed without editing the code there. In practice:

- **Extract a function** from the middle of a long one, so it can be called directly.
- **Introduce a parameter** for the hidden dependency (the clock, the database, the HTTP client) with the old behaviour as the default, so callers do not change.
- **Extract an interface** for the collaborator, sized to what this code uses, and pass a fake.
- **Subclass-and-override / wrap** when nothing else is possible — a temporary measure to get the net in place.

Each of these is a tiny commit made with the characterization test green. Never combine "make it testable" with "change what it does".

## Commit sequence

```text
test(billing): characterize current proration behaviour
refactor(billing): extract prorate() from Invoice.Recalculate
refactor(billing): inject clock into prorate()
feat(billing): prorate by calendar days with test        ← greenfield cycle starts here
```

## When to stop

The goal is not to test the whole legacy module. It is to get a net under the part you need to change, change it safely, and leave it more testable than you found it. Resist the rewrite; the characterization tests you would need to make a rewrite safe are the same ones that make the incremental path cheap.
