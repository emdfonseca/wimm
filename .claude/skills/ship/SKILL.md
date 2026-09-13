---
name: ship
description: The definition-of-done gate. Run before declaring a change complete, opening a PR, or merging - it walks scope, checks, contracts, migrations, observability, design parity, ADRs, and security in a fixed order and reports each as pass, fail, or N/A with a reason. Invoke explicitly with /ship; it does not trigger on its own.
disable-model-invocation: true
argument-hint: "[scope: a package dir, a PR number, or blank for the working tree]"
allowed-tools: Bash(just *) Bash(git *) Read Grep Glob
---

# Ship: the definition-of-done gate

Work is done when every applicable gate below passes or is explicitly N/A with a reason. Not when the tests pass, and not when it "looks right". Report the result as a table — one row per gate — and be blunt: a `fail` with the reason is worth more than a `pass` that turns out to be "did not check".

The argument, if any, scopes the gate: a directory under `apps/` or `packages/`, a PR number, or nothing for the current working tree.

## The gates, in order

Run them in this order because later ones assume earlier ones. Stop and report at the first `fail` that makes the rest meaningless (no scope, `just check` failing); otherwise run all of them and report everything.

### 1. Scope is stated

The change can be described in one sentence: what it does and for whom. If that sentence needs "and", it may be two changes. Confirm which directories are touched (`git diff --name-only <base>`) and that they are the ones the sentence implies.

### 2. Checks pass

`just check <dir>` for each touched directory, or `just ci-affected` if available. Output goes in the report, not in the summary. A skipped test or a `//nolint` added in this change is a `fail` until it has a comment saying why.

### 3. Nothing generated is stale

If `packages/contracts` or any generator input changed: `just gen`, then `git status --porcelain` shows nothing new. Generated code committed by hand-edit is a `fail` (the hook should have caught it; confirm anyway).

### 4. API contract — if a `.proto`, OpenAPI file, or handler changed

Walk `api-contract`'s `references/new-endpoint-checklist.md`. `buf lint` and `buf breaking` clean, or the PR is explicitly a new-major. Every error the contract can return has a test.

### 5. Migrations — if anything under `migrations/` changed

Walk `data-migrations`'s `references/review-checklist.md`. The PR names its expand/contract step. Duration on a production-sized table is stated. `migrate-redo` passes.

### 6. Observability — if a service or endpoint was added

Walk `observability`'s `references/instrument-service-checklist.md`, at minimum the *Verify* item: one request findable by trace ID in traces, logs, and metrics. New alerts have runbooks and owners.

### 7. Design parity — if UI changed

Every design state for the touched component or screen has a story (`storybook-svelte`), the a11y check is clean, and the accessibility contract has a play function. A design state with no story is a `fail`, not a follow-up.

### 8. Architecture is recorded — if a decision was made

A new dependency, a new service, a changed boundary, a protocol choice, a data-model shape: that is a decision, and it gets an ADR via `documentation-and-adrs`. "We can write it up later" is the sentence before it is never written.

### 9. Security — if auth, permissions, secrets, or user data are touched

Run `security-review` on the diff. Any public REST surface change: the second reviewer from outside the team has signed off.

### 10. Commit and PR are honest

Conventional commit with the package as scope. PR description says what, why, which expand/contract step if any, and what was *not* done. Reviewers should learn nothing from the diff that the description hid.

## Report format

```text
SHIP  <scope>

1  scope stated            pass   "Adds invoice currency; apps/billing, packages/contracts"
2  checks                  pass   just check apps/billing — 0 failures
3  generated current       pass
4  api contract            pass   buf breaking clean; 3 error-code tests added
5  migrations              fail   expand step not named in PR; duration on staging not recorded
6  observability           N/A    no new service or endpoint
7  design parity           N/A    no UI change
8  ADR                     pass   ADR-0007 currency handling
9  security                N/A    no auth/data surface touched
10 commit/PR               pass

RESULT: NOT READY — fix gate 5 and rerun /ship
```

`N/A` always carries the reason. `RESULT` is `READY` only when there are no `fail` rows. Never soften a `fail` into a "note" because the user seems in a hurry; the gate exists for exactly that moment.
