# Cohesion: keeping the repo tied together

A monorepo fails slowly. It rarely breaks; it just becomes a place where nobody can tell what depends on what, and every change touches six directories. Cohesion is the property that prevents that, and it is a design decision made at package-creation time — not something a build tool can recover later.

**High cohesion:** everything in a package changes for the same reason.
**Low coupling:** packages know as little about each other as the work allows.

## Slice by domain, not by layer

The single most damaging habit is organizing packages by technical layer:

```text
packages/
├── types/          ← changes for every feature
├── utils/          ← changes for every feature
├── api-clients/    ← changes for every feature
└── components/     ← changes for every feature
```

Every feature touches all four, every package depends on all the others, and no package has an owner who understands all of it. Slice by domain instead, so a feature lands in one place:

```text
packages/
├── billing/        ← types, logic, client, and UI for billing
├── identity/
├── catalog/
└── telemetry/      ← genuinely cross-cutting, and only because it is
```

Layer organization *inside* a package is fine and usually good (`billing/src/domain`, `billing/src/http`). The rule is about package boundaries, which is where ownership and dependency edges are created.

## The banned names

`utils`, `common`, `shared`, `core`, `helpers`, `misc`, `lib`. Not because the code is bad, but because the name states no reason to change — so everything accretes there, and the package becomes a dependency of everything and an owner of nothing.

When you want to reach for one of those names, the honest question is what the code actually is. `shared` usually turns out to be three real things: a domain concept (goes in that domain's package), a cross-cutting capability (`telemetry`, `config`, `http-client` — named for the capability), and one genuinely generic function (inline it; two copies of a five-line helper cost less than a coupling edge).

## Sizing a package

Use these as evidence, not thresholds:

- **One reason to change.** List the last ten changes to the package. Several unrelated reasons means it should be split; a package that only ever changes alongside another means they should be merged.
- **Someone can own it.** If no single person or team can review any change to it, it is too big or too incoherent.
- **The name is a noun a domain expert recognizes.** If describing the package needs "and", suspect two packages.
- **Its public surface is small relative to its inside.** A package exporting almost everything it contains is a folder, not a module.

Splitting too early is also a real cost — two packages that always change together add ceremony without buying isolation. When unsure, keep it together and split at the second consumer.

## Make the public surface explicit

Coupling is to what a package *exposes*, so control that deliberately:

- **Go:** put everything not meant for consumers under `internal/`. The compiler enforces it — this is the strongest boundary available in the repo, so use it aggressively.
- **TypeScript:** declare `exports` in `package.json` with a single entry (plus explicit subpaths when genuinely needed) and never let consumers deep-import `src/...`. Without an `exports` map every file is public API.
- **Python:** keep a src layout and export through `__init__.py`; treat underscore-prefixed modules as private.

A package with one well-chosen entry point can be refactored freely. A package whose internals are reachable cannot be changed without a repo-wide search.

## Enforce it, do not just document it

Conventions decay unless something fails. Wire at least one check per language into `just lint`, so violations surface in the pull request that creates them rather than in a cleanup quarter later:

- **Go:** `depguard` (in golangci-lint) to forbid disallowed import paths; `internal/` does the rest for free.
- **TypeScript:** `dependency-cruiser` or `eslint-plugin-boundaries` with rules matching `references/layout-and-boundaries.md` — no `packages/* → apps/*`, no cycles, no deep imports.
- **Python:** `import-linter` contracts for layered and independent package rules.

A `just graph` recipe that renders the current dependency graph is worth the twenty minutes: an edge nobody can justify is easy to spot in a picture and invisible in a diff.

## Signals to act on

- A change routinely touches three or more packages → the boundary is in the wrong place.
- A package is imported by everything → it is either a genuine foundation (rare, keep it tiny and stable) or a junk drawer.
- Two packages are always edited together → merge them.
- A package has no owner in `CODEOWNERS` → nobody is defending its coherence.
- Cycles appear → an interface belongs on the other side of the edge.

Treat these as design feedback rather than cleanup chores. The repo is telling you where its real seams are.
