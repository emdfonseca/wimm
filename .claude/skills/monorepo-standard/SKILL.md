---
name: monorepo-standard
paths:
  - "justfile"
  - "**/justfile"
  - "devbox.json"
  - "pnpm-workspace.yaml"
  - "go.work"
  - "pyproject.toml"
  - "CODEOWNERS"
  - ".github/workflows/**"
description: Organizing standard for a polyglot monorepo - Go, TypeScript/Node, Python, and infrastructure living together under apps/ and packages/, with devbox owning the toolchain and just owning every task entry point. Use this whenever someone adds an app, service, package, library, or module to a repo; asks where a file, directory, or shared piece of code should live; sets up or changes devbox.json, justfile, pnpm-workspace.yaml, go.work, or CI workflows; wonders which package may import which, or how to share code between languages; asks about versioning, releasing, or publishing from a monorepo; or is starting a new repo and wants it laid out properly. Reach for it even when the request sounds small ("add a script for X", "where do I put this helper", "new service for billing") - those are exactly the decisions that decide whether the repo stays navigable.
---

# Polyglot monorepo organization

**Core rule: `devbox` owns the toolchain, `just` owns the tasks, the language's native tool owns its own packages.** Each layer does one job, so no tool has to pretend to understand the others.

```text
devbox.json   which versions of which tools exist        (Nix; pinned; identical locally and in CI)
justfile      every command a human or CI ever runs      (uniform verbs across all languages)
pnpm / go.work / uv / terraform   dependency resolution inside one language
```

Anything that blurs those layers is the thing to push back on: task logic in `devbox.json` scripts, toolchain versions in CI YAML, a language's dependency graph reimplemented in the task runner.

## Start here, every time

1. **Is it deployable, or importable?** Deployable → `apps/`. Importable → `packages/`. Nothing is both. An app is never imported by anything; a package is never deployed on its own.
2. **Which language owns it?** One directory = one language = one owner. Cross-language sharing goes through a generated contract in `packages/`, never through copied types.
3. **Does the entry point already exist?** Every package answers the same small verb set. Adding work usually means implementing an existing verb, not inventing a command.

## Repository layout

```text
repo/
├── devbox.json            # every tool + version, for humans and CI alike
├── devbox.lock
├── justfile               # root task entry points; delegates to package justfiles
├── pnpm-workspace.yaml    # TS workspace membership
├── go.work                # Go workspace membership
├── apps/                  # deployable units — one per deployable artifact
│   ├── web/               # TS
│   ├── api/               # Go
│   └── ingest/            # Python
├── packages/              # importable libraries — never deployed alone
│   ├── ui/                # TS
│   ├── telemetry/         # Go
│   └── contracts/         # generated cross-language clients (proto/OpenAPI)
├── infra/                 # IaC, one directory per environment or stack
├── tools/                 # repo-local generators and scripts; never published
└── docs/
```

Deeper nesting (`apps/web/frontend/...`) is where navigability dies. If a directory needs subdivision, that is usually a sign it should be two packages.

## Where to read next

| Read this | When |
|---|---|
| `references/cohesion.md` | Deciding what belongs in one package; sizing or splitting packages; anything tempting you toward a `shared`/`utils` directory; enforcing boundaries with linters. Read this before creating a package. |
| `references/layout-and-boundaries.md` | Placing a new directory; deciding app vs package vs tool; dependency rules and cycles; cross-language sharing. |
| `references/devbox.md` | Adding or pinning a tool; the CI parity contract; why `devbox.json` scripts stay nearly empty. |
| `references/just.md` | Writing or changing tasks; the verb vocabulary; root-to-package delegation; arguments and recipes. |
| `references/typescript.md` | pnpm workspaces, `@repo/*` naming, tsconfig inheritance, build outputs, when Turborepo earns its place. |
| `references/go.md` | `go.work`, module paths, `cmd/` vs `internal/` vs `pkg/`, shared libraries, build and test caching. |
| `references/python.md` | `uv`, src layout, workspace membership, lockfiles. |
| `references/infra.md` | Terraform/k8s layout, environment separation, what infra may reference. |
| `references/ci.md` | Affected-package detection, caching per language, required checks. |
| `references/versioning-and-release.md` | Internal vs published packages, changesets, Go module tags, conventional commits. |
| `references/new-package-checklist.md` | Adding an app or package — the full procedure, end to end. |
| `references/enforcement-hooks.md` | The Claude Code hooks that make the generated-file and justfile rules deterministic instead of advisory; install when the repo is scaffolded. |
| `assets/` | Starter `devbox.json`, root and package `justfile`, `tsconfig.base.json`, workspace manifests, `hooks/` for the enforcement hooks. |

## Rules that are constantly needed

### The task verb vocabulary

Every package implements the same verbs, whatever language it is written in. This is the entire reason a polyglot monorepo stays usable — a newcomer runs `just test api` without knowing that `api` is Go.

```text
just build <pkg>     produce artifacts
just test <pkg>      run tests
just lint <pkg>      static analysis
just fmt <pkg>       format in place
just check <pkg>     lint + typecheck + test — what CI runs per package
just dev <pkg>       local run/watch loop
just clean <pkg>     remove build outputs
```

Root aggregates: `just check` runs every package, `just ci` runs what CI runs. A verb a package genuinely cannot support should fail loudly rather than be silently absent — a missing verb is indistinguishable from a broken one otherwise.

### devbox declares tools, never tasks

Toolchain versions live in `devbox.json` and nowhere else — not in CI YAML, not in `.nvmrc`, not in a README instruction. CI runs `devbox run -- just ci`, so the local shell and the CI runner resolve the same binaries.

Keep `devbox.json`'s `shell.scripts` empty except for bootstrap (`init`, at most). Task logic belongs in the justfile because recipes take arguments, compose, declare dependencies, and are readable in one place; devbox scripts do none of that, and splitting tasks across both means nobody can find where a command is defined.

### Cohesion is the point

A package holds things that change for the same reason, and exposes as little as it can get away with. Two habits protect that, and both have to be applied at creation time because neither can be retrofitted cheaply:

**Slice by domain, not by layer.** `packages/billing` (types, logic, client, UI for billing) rather than `packages/types` + `packages/utils` + `packages/api-clients`, where every feature touches every package and nothing has a real owner. Layering *inside* a package is fine — it is the package boundary that creates dependency edges and ownership.

**Refuse the junk-drawer names.** `utils`, `common`, `shared`, `core`, `helpers`, `misc`. A name that states no reason to change attracts everything, and the result is a package depended on by all and owned by none. What wants to go there is almost always a domain concept (put it in that domain), a cross-cutting capability (name it for the capability: `telemetry`, `config`), or one small generic function (inline it — duplicating five lines costs less than a coupling edge).

Enforce with a linter rather than a convention doc: `depguard` for Go, `dependency-cruiser` or `eslint-plugin-boundaries` for TS, `import-linter` for Python, all wired into `just lint`. Details and sizing heuristics are in `references/cohesion.md`.

### Dependency direction

```text
apps/*      → packages/*        allowed
packages/*  → packages/*        allowed, acyclic
packages/*  → apps/*            never
tools/*     → anything          allowed; nothing depends on tools/
infra/*     → nothing in code   infra references built artifacts, not source
```

Two apps needing the same code means extracting a package, never importing app to app. Cycles between packages are a design error, not a build configuration problem — fix the boundary.

### Cross-language sharing

Share **contracts**, not code: protobuf or OpenAPI definitions in `packages/contracts/`, with generated clients per language committed or generated by `just gen`. Hand-copied structs drift silently and the drift surfaces in production, which is why the generation step is worth the friction.

This skill owns *where* contracts live and how they are generated. The `api-contract` skill owns what they *look like* — protocol choice, error shape, pagination, versioning. Read that one before writing a `.proto` or an OpenAPI document.

### Naming

Directories are kebab-case and named for the domain, not the technology (`billing`, not `billing-service-go`). TS packages are `@repo/<dir>`; Go modules are `<module-prefix>/packages/<dir>`; Python distributions are `<org>-<dir>`. The directory name is the canonical identity — every manifest matches it, so grep finds everything.

## Checkpoints worth insisting on

These are the moments where a monorepo either stays organized or quietly stops being one.

- **Adding a directory** — run `references/new-package-checklist.md`. A package that skips workspace registration or the verb set becomes invisible to CI, and nobody notices until it breaks.
- **Adding a tool** — it goes in `devbox.json` with a pinned version, or it does not exist. "Works on my machine" starts here.
- **Adding a script** — a justfile recipe, not a loose `.sh` in a package, not a `package.json` script that only TS people can discover.
- **Second copy of anything** — the moment code is duplicated across two apps, extract it; the cost of extraction only grows.
- **CI drifting from local** — if CI does something no justfile recipe does, the repo has two sources of truth and one of them will rot.

When you find an existing violation while doing unrelated work, name it and offer the fix rather than silently working around it — but keep it separate from the task at hand unless the user asks otherwise.
