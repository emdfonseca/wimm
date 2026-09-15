# 0012 · Stories carry the user, specs carry the contract

## Status

Accepted

## Context

The workflow from ADR 0008 goes proposal → specs → design → canvas → tasks. Every
one of those states what the system does. None states who wants it or what they
get.

`### Requirement: The system SHALL...` has no actor and no value. It cannot answer
whether a slice is worth shipping, whether it stands on its own, or whether it is
one piece of work or three. Those are the questions INVEST asks, and they have to
be answered before deciding which capabilities exist — which is to say, before
the specs, not after.

The tempting move is to fold user stories into the specs. That is wrong for a
specific reason: a spec scenario is already an acceptance criterion. `#### Scenario:`
with WHEN/THEN is Gherkin-shaped and testable. Putting criteria in both places
creates two sources of truth that drift, and the one that drifts is always the
one nobody runs.

## Decision

A `stories` artifact sits between proposal and specs. `specs` requires it.

```text
proposal → stories → specs → design → canvas → tasks
```

The two levels split by audience, and the split is the point:

```text
story criterion   what a person can do, checked by using the product
spec scenario     normative system behaviour, checked by a test
```

A criterion naming a function, table, endpoint or component is not a criterion —
it is an implementation note, and it belongs in design or tasks.

**Stories name a real user by role.** "As a developer" and "as the user" are the
tell that there is no actor, and without an actor there is no value to state.

**INVEST is recorded, not performed.** Six lines, each answered. A story that
fails one is not automatically wrong: name the letter, say why shipping it anyway
is right. An honest "Small — no, this covers three screens, and splitting them
would ship a half-usable ledger" is worth more than six unconsidered ticks.

**Every requirement traces to a story.** If one does not, either the story is
missing or the requirement is.

**It is conditional**, like design and canvas. Tooling, refactors and internal
migrations record a one-line skip rather than inventing a user.

## Consequences

- Six artifacts is a lot for a small change. Three of them — stories, design,
  canvas — are conditional, so a tooling change still writes two.
- The traceability rule is only as good as the person applying it; nothing in CI
  checks that a requirement names a story. If that becomes a real problem the
  check is cheap to write, and it is not worth writing before it does.
- Acceptance criteria that cover only the happy path were the common failure this
  is meant to catch. The instruction asks for the unhappy path the user can
  actually reach, and the canvas artifact now draws every state the criteria
  name.
- Stories are numbered S1, S2, … and the canvas and tasks refer to them by
  number, so the thread from value to frame to task is followable in both
  directions.
