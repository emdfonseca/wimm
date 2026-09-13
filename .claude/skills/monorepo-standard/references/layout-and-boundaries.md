# Layout and boundaries

## The app / package / tool decision

| Kind | Lives in | Test | Notes |
|---|---|---|---|
| Deployable unit | `apps/<name>` | Ships somewhere: a container, a binary, a static bundle, a lambda | Never imported by anything else in the repo |
| Importable library | `packages/<name>` | Another directory depends on it | Never deployed on its own |
| Repo-local tooling | `tools/<name>` | Only this repo runs it: generators, migration scripts, lint plugins | Never published; nothing depends on it |
| Infrastructure | `infra/<stack>` | Describes where things run | References built artifacts, never source |

If something seems to be both an app and a package, it is two things: the deployable shell in `apps/`, the reusable core in `packages/`. This split is worth doing at creation time, because retrofitting it means untangling imports that accumulated in the meantime.

## One directory, one language, one owner

A directory contains exactly one language's build. A Go service with a small TS admin UI is two directories (`apps/billing`, `apps/billing-admin`), not one with two toolchains — otherwise every task verb needs conditionals and CI cannot cache either half properly.

Ownership follows directories. Record it in `CODEOWNERS` at the directory level so the boundary is enforced by review, not memory.

## Dependency rules

```text
apps/*      → packages/*        allowed
packages/*  → packages/*        allowed, must stay acyclic
packages/*  → apps/*            never
apps/*      → apps/*            never
tools/*     → anything          allowed
anything    → tools/*           never
infra/*     → built artifacts   never source code
```

Two apps needing the same thing is the signal to extract a package. An app importing an app hardwires a deployment coupling that nothing in the build system will warn you about.

Cycles between packages are a boundary error. The fix is moving the shared piece down into a third package or inverting the dependency with an interface — not a build flag.

## Cross-language sharing

Never hand-copy types across languages. Define the contract once and generate:

```text
packages/contracts/
├── proto/            # or openapi/
├── gen/go/
├── gen/ts/
└── justfile          # just gen regenerates every target
```

Decide once whether generated code is committed. Committing it keeps builds hermetic and makes diffs reviewable, at the cost of regeneration noise; generating on demand keeps the tree clean but makes every consumer depend on the generator toolchain. Committing is the safer default for a small team, and `just gen` plus a CI check that the tree is clean catches drift.

## Depth

Two levels below the root category is the working limit: `packages/ui/src/...` is fine, `packages/frontend/shared/ui/src/...` is a smell. Deep nesting usually encodes a grouping that should be a package boundary or a naming convention instead.

## Files that belong at the root

Only what is genuinely repo-wide: `devbox.json`, `justfile`, workspace manifests (`pnpm-workspace.yaml`, `go.work`), `.gitignore`, `CODEOWNERS`, `README.md`, CI config. Everything else pushes down into the directory it serves — a root-level config that only one package reads is a trap for the next person who greps for it.
