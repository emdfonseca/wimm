# 0013 · Acceptance criteria are scenarios

## Status

Accepted

## Context

ADR 0012 gave the `stories` artifact both an INVEST narrative and a list of
user-testable acceptance criteria, and split it from the specs by audience: a
criterion is checked by using the product, a scenario is checked by a test.

That line is too thin to hold. A criterion and a `#### Scenario:` are the same
kind of statement, so the same behaviour gets written twice at slightly different
altitudes — the duplication ADR 0012 set out to avoid, one level down.

Merging the other way, so that a story simply *is* a requirement, works
mechanically and fails slowly. The CLI ignores unknown `##` sections and copies
each requirement's raw markdown verbatim through archive, so an `### INVEST`
block does reach the main spec. But `MODIFIED` replaces a whole block, and while a
dropped scenario is an error, a dropped narrative is only a warning. Story text
would erode silently on the next change that touches the requirement — and the
permanent spec would accumulate work-sizing notes from past work, which is history
in a state document.

## Decision

The spec's scenarios are the acceptance criteria. There is no second list.

The line between the two artifacts is **lifetime**, not audience:

```text
openspec/specs/<cap>/spec.md   living. Accumulates across changes.
stories.md                     change-scoped. Archives with the change.
```

A requirement is a persistent behaviour contract; a story is a unit of work.
INVEST asks whether the work item is Independent, Estimable, Small — questions
with no meaning about a permanent capability.

So `stories.md` keeps only what a spec cannot hold: the actor, the value, the
INVEST answer, and `Satisfied by`. The user-testable bar moves onto the scenarios
themselves — a scenario for a user-facing capability must be observable by a
person, and one naming a function, table, endpoint or component is an
implementation note.

**The link is a metadata line.** Every requirement carries `**Story**: S<n>`
under its header. `**Key**: value` is the one pattern the parser recognises: it is
excluded from the requirement body when other text is present, and survives
archive verbatim. Free prose would not.

## Consequences

- **A MODIFIED block must repeat its `**Story**:` line.** A dropped scenario is an
  error and a dropped metadata line is not caught at all, so this is the one place
  traceability can rot unnoticed.
- Verified end to end on a throwaway capability: `**Story**: S1` reached
  `openspec/specs/smoke/spec.md` intact, `stories.md` archived with the change,
  and the permanent spec carried no INVEST or As-a text. The run also exercised
  propose → stories → specs → design → canvas → tasks → archive for the first
  time.
- **Scenario shape is a house rule the tool will not enforce.** The CLI counts any
  `####` header as a scenario and never inspects the bullets — WHEN/THEN appears
  nowhere in its parser. A malformed scenario passes validation silently.
- `--strict` only promotes warnings to failures; its practical effect here is
  making a requirement without SHALL or MUST fatal.
- Nothing checks that every requirement names a story. That check is cheap and
  worth writing once there is a real change to run it against.
