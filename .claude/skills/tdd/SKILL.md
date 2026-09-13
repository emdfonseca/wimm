---
name: tdd
description: Strict red-green-refactor - one failing test at a time, failure seen before any implementation, smallest change to green, explicit refactor step. Use when the user says /tdd, asks to work test-first, or is about to implement domain logic, error mapping, a parser, a state machine, a backfill, or any behaviour with rules worth pinning - even without the word "test". Not for UI, config, or glue.
argument-hint: "[behaviour to build, or a file/package]"
allowed-tools: Bash(just *) Bash(go *) Bash(pnpm *) Bash(uv *) Bash(git *) Read Edit Write Glob Grep
---

# TDD: red, green, refactor

**The test is written first because it is the only version of the spec the implementation cannot bend to fit.**

Task: $ARGUMENTS

## The cycle

```text
RED       write ONE test for ONE behaviour → run it → read the failure → confirm it fails for the right reason
GREEN     the smallest change that makes it pass — nothing the test does not demand
REFACTOR  improve the design with the test green; run again
```

Never two tests ahead of the code; never code the test did not ask for.

### Red must be seen

Run the test and read the output before writing any implementation. The failure must be the assertion you wrote — not a compile error, a typo, or a fixture that did not load. Report the failing line verbatim.

### Green means minimal

Implement what the test demands and stop. An `if` the current test does not exercise is the next test.

### Refactor is a step, not an option

After green, ask what the code now wants to be: a better name, duplication between the cases just written, a parameter that should be a type. Do it, run the test, then say what was refactored — or "nothing to refactor" so the step is visibly taken.

### Mutation check

Before calling a behaviour done, break the implementation — flip the condition, return the wrong value — and run the test. It must fail.

### One test at a time

Ask "what is the ONE next behaviour?" before every test. Never write six cases and then implement all six. Commits are per behaviour and only when the user asks. Never commit red.

## When the test encodes a decision, pause

Some tests pin a fact the user already gave you (`given X, return Y`). Some pin a decision nobody has made yet — which error code a failure maps to, whether an empty list is an error, how rounding works. Write the test, show it, and stop before implementing when it is the second kind. The failing test is the cheapest possible artefact to disagree with; the implementation is not.

For everything else, keep cycling without asking.

## Where it applies

| Do TDD | Use something else |
|---|---|
| Domain logic in `apps/*/internal/<domain>`, `packages/*` | Route modules, `cmd/main.go` wiring |
| Handler error mapping — one test per contract error code | UI layout and states — Storybook |
| Parsers, validators, state machines, money and time arithmetic | Config, IaC, generated code |
| Backfill jobs — idempotency and resume | Spikes — throw away, then TDD the real thing |
| A bug report: the reproduction *is* the red test | Glue calling two tested things in sequence |

Run: `just test <dir>`; a single test via the native runner. Language specifics: `.claude/rules/<lang>.md`, loaded automatically by path.

## Per-cycle report

```text
RED    TestInvoice_RejectsNegativeTotal — FAIL: expected ErrInvalidTotal, got nil   (invoice_test.go:41)
GREEN  invoice.go: +4 lines, guard in New()   REFACTOR  extracted validateTotal() / none
```
