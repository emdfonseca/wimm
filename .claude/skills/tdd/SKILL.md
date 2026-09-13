---
name: tdd
description: Strict red-green-refactor for this repo - one failing test at a time, the failure seen and read before any implementation, the smallest change to green, an explicit refactor step, one commit per cycle; characterization tests first on untested legacy code. Use whenever the user says /tdd, asks to work test-first, or is about to implement domain logic, a handler's error mapping, a parser, a state machine, a backfill, or any behaviour with rules worth pinning - even if they did not say "test". Reach for it especially when the temptation is to write the code and "add tests after": with an agent that produces tests that merely confirm what was written. Not for UI layout, scaffolding, config, or glue; Storybook play functions and e2e cover those.
argument-hint: "[behaviour to build, or a file/package]"
allowed-tools: Bash(just *) Bash(go *) Bash(pnpm *) Bash(uv *) Bash(git *) Read Edit Write Glob Grep
---

# TDD: red, green, refactor

**The test is written first because it is the only version of the spec the implementation cannot bend to fit.** Code written and then tested produces tests that pass by construction — they describe what the code does, not what it should do. That risk is higher, not lower, when an agent writes both halves in one sitting, which is why the discipline below is strict about *seeing* red rather than assuming it.

Task: $ARGUMENTS

## The cycle

```text
RED       write ONE test for ONE behaviour → run it → read the failure → confirm it fails for the right reason
GREEN     the smallest change that makes it pass — nothing the test does not demand
REFACTOR  improve the design with the test green; run again
COMMIT    test + implementation + any doc touched, one commit: feat(<pkg>): <behaviour> with test
```

Then the next behaviour. Never two tests ahead of the code; never code the test did not ask for.

### Red must be seen

Run the test and read the output before writing any implementation. "It would fail" is not red. Two things count:

- In Go, the first red is often a compile error — the function does not exist. That is a legitimate red, but it is not the *interesting* one. Add the stub, run again, and get the **assertion** failure. Only then does the test prove it checks behaviour.
- The failure has to be for the *right reason*: the assertion you wrote, not a typo, a missing import, or a fixture that did not load. A test that fails for the wrong reason and then passes has taught you nothing.

Report the failing line verbatim. The user should be able to see, in the transcript, that the test was red.

### Green means minimal

Implement what the test demands and stop. Hard-coding a return value to pass the first test is fine — the second test removes it. Generalising early is how untested branches appear. If you find yourself writing an `if` the current test does not exercise, that is the next test, not this implementation.

### Refactor is a step, not an option

After green, ask what the code now wants to be: a name that reads better, duplication between the two cases just written, a parameter that should be a type. Do it, run the test, then say what was refactored — or say "nothing to refactor" so the step is visibly taken rather than silently skipped. Agents skip this step more than any other; naming it is what keeps it.

### One test at a time, one commit per cycle

Ask "what is the ONE next behaviour?" before every test. Refuse to write a test file with six cases and then implement all six; the value of the cycle is in the small steps, and batching removes it. Each commit is a complete cycle: `feat(billing): reject negative invoice totals with test`. Never commit red.

## When the test encodes a decision, pause

Some tests pin a fact the user already gave you (`given X, return Y`). Some pin a decision nobody has made yet — which error code a failure maps to, whether an empty list is an error, how rounding works. Write the test, show it, and stop before implementing when it is the second kind. The failing test is the cheapest possible artefact to disagree with; the implementation is not.

For everything else, keep cycling without asking.

## Where it applies in this repo

| Do TDD | Use something else |
|---|---|
| Domain logic in `apps/*/internal/<domain>` and `packages/*` | Route modules (`+page.svelte`, `cmd/main.go` wiring) |
| Handler error mapping — one test per code the contract can return (`api-contract`) | UI layout and visual states — Storybook stories and play functions (`storybook-svelte`) |
| Parsers, validators, state machines, money and time arithmetic | Config, IaC, generated code |
| Backfill jobs — idempotency and resume (`data-migrations`) | Exploratory spikes — throw the spike away, then TDD the real thing |
| Anything with a bug report: the reproduction *is* the red test | Glue that only calls two well-tested things in sequence |

Tests live beside the code (`_test.go`, `.test.ts`) or in the package's `tests/` for Python, per `monorepo-standard`. Run them through the verb: `just test <dir>`; for a single test during the cycle use the native runner directly (`go test -run`, `pnpm vitest -t`, `uv run pytest -k`).

## Legacy code: characterization first

Untested code that is hard to test gets a different first move: a **characterization test** that captures what it does today, right or wrong, as a safety net. Then refactor for testability — extract, inject, separate — in tiny steps with that net green. Then TDD the new behaviour. Never refactor untested code without the net; never TDD on top of code you cannot run in isolation. `references/legacy.md` has the techniques.

## Where to read next

| Read this | When |
|---|---|
| `references/test-quality.md` | Naming, one-behaviour-per-test, fakes over mocks, determinism, the "would it fail if I broke the code?" check. Read once; apply always. |
| `references/go.md` | Table-driven tests, `t.Run`, testing Connect handlers, `httptest`, fakes at the consumer's interface, golden files. |
| `references/typescript.md` | Vitest conventions, `vi.fn`, what belongs in a unit test versus a Storybook play function. |
| `references/python.md` | pytest fixtures, `parametrize`, property tests with Hypothesis for parsers and arithmetic. |
| `references/legacy.md` | Characterization tests, finding seams, Feathers' techniques, approval tests. |

## Per-cycle report

Keep it short and keep it honest:

```text
RED       TestInvoice_RejectsNegativeTotal — FAIL: expected ErrInvalidTotal, got nil   (apps/billing/internal/invoice/invoice_test.go:41)
GREEN     invoice.go: +4 lines, guard in New()
REFACTOR  extracted validateTotal(); none needed / …
COMMIT    feat(billing): reject negative invoice totals with test
```

Behavioural prompts to keep asking yourself: *Is this greenfield or legacy? What is the one behaviour? Did the test fail for the right reason? Does the implementation do more than the test asked? What does the code want to be now that it is green?*
