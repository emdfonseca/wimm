---
name: ship
description: The definition-of-done gate. Run before declaring a change complete, opening a PR, or merging - it scopes the diff, runs the deterministic checks, invokes code review, and checks contract, migration, observability, security, and ADR coverage, reporting each gate as pass, fail, or N/A with a reason. Invoke explicitly with /ship; it does not trigger on its own.
disable-model-invocation: true
argument-hint: "[base ref or package dir; blank for the working tree]"
allowed-tools: Bash(just *) Bash(git *) Read Grep Glob Agent Skill
---

# Ship

Scope: `git diff --name-only <base>` ($ARGUMENTS, or the working tree). No files changed → stop, report "nothing to ship".

1. **Checks** — `just check` for each touched package that has a `justfile`. If `.proto`, OpenAPI, or `migrations/` files are touched: `just gen && git status --porcelain` must print nothing. Report the single output line that decides pass/fail.
2. **Code review** — always: `/code-review low` on the diff.
3. **Security** — `/security-review` only if the diff touches auth, secrets, session, crypto, or input-parsing files. Otherwise N/A.
4. **Standards** — `api-contract`, `data-migrations`, `observability` load by path. When their files are in scope, walk the matching `references/*-checklist.md` (new-endpoint, review, instrument-service); report unmet items.
5. **ADR** — a change that adds a dependency, a service, a public contract, or a schema needs a `docs/decisions/NNNN-*.md` in the diff. Missing → fail; point to `/adr`.

Output only a table: `gate | pass/fail/N/A | one-line reason`. Fail if any row fails.
