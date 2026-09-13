# Adding an app or package

Work through this in order. The steps that get skipped — workspace registration, the verb set, ownership — are precisely the ones that make a directory invisible to CI and to the next person.

## 1. Decide what it is

- [ ] Deployable → `apps/`; importable → `packages/`; repo-local tooling → `tools/`.
- [ ] Name is a domain noun in kebab-case, not a technology (`billing`, not `billing-svc-go`).
- [ ] It is not one of the banned names (`utils`, `common`, `shared`, `core`, `helpers`) — see `references/cohesion.md`.
- [ ] You can state its one reason to change in a sentence.
- [ ] It is genuinely not an existing package. Two packages that will always change together should be one.

## 2. Create it

- [ ] One language only.
- [ ] Manifest name matches the directory (`@repo/<dir>`, `<prefix>/<path>`, `<org>-<dir>`).
- [ ] Public surface is explicit: Go `internal/`, TS `exports` map, Python src layout.
- [ ] README states what it is for and what it deliberately does not do.

## 3. Register it

- [ ] Workspace manifest updated: `pnpm-workspace.yaml`, `go.work`, or the uv workspace members.
- [ ] Added to the root justfile's package list.
- [ ] `CODEOWNERS` entry.
- [ ] Any new tool it needs is pinned in `devbox.json`.

## 4. Wire the verbs

- [ ] `justfile` implements `build`, `test`, `lint`, `fmt`, `check`, `dev`, `clean`.
- [ ] Verbs that do not apply fail with a clear message rather than being absent.
- [ ] `just check <dir>` passes from a clean clone.

## 5. Check the boundaries

- [ ] Dependency direction is legal (`packages/` never depends on `apps/`).
- [ ] No cycle introduced.
- [ ] Nothing shared with another package by copy — extract or generate instead.
- [ ] Boundary linter rules updated if the package introduces a new layer.

## 6. Confirm

- [ ] `devbox run -- just ci` passes.
- [ ] Build outputs are gitignored.
- [ ] It appears in `just --list` output the way a newcomer would expect to find it.
