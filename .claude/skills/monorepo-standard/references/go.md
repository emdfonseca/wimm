# Go

## Workspace

```text
// go.work
go 1.23

use (
    ./apps/api
    ./apps/worker
    ./packages/telemetry
    ./packages/contracts/gen/go
)
```

Each directory is its own module with its own `go.mod`. `go.work` makes local edits resolve to local source without `replace` directives scattered through every manifest. `go.work.sum` is committed.

A single module for the whole repo is the simpler alternative, and reasonable if everything is Go — but in a polyglot repo the per-directory module keeps ownership and release boundaries aligned with directories.

## Module paths

`<module-prefix>/<dir-path>` — for example `github.com/org/repo/packages/telemetry`. The import path mirrors the filesystem, so a reader can find any import instantly.

## Internal layout

```text
apps/api/
├── go.mod
├── justfile
├── cmd/api/main.go     # entry point: wiring only, no logic
└── internal/           # everything else — compiler-enforced privacy
    ├── http/
    ├── store/
    └── billing/
```

For packages:

```text
packages/telemetry/
├── go.mod
├── justfile
├── telemetry.go        # the public surface, at the module root
└── internal/           # implementation detail
```

`internal/` is the most effective cohesion tool the repo has: anything under it is unreachable from outside the module and can be refactored without coordination. Default to putting code there and promote it out only when a second consumer genuinely needs it. Avoid a top-level `pkg/` — it adds a directory level and signals nothing that the module path does not already say.

## Tasks

`go build` and `go test` maintain their own content-addressed cache, which is why the task layer stays thin here — do not wrap them in extra caching logic. Keep `GOFLAGS=-mod=readonly` set in devbox so a build never silently rewrites `go.mod`.

Lint with `golangci-lint`, configured once at the repo root and inherited; enable `depguard` to encode the dependency rules from `references/layout-and-boundaries.md`.

## Build outputs

Binaries land in a root `dist/`, named for the app. Nothing else writes there, so `just clean` and CI artifact collection both stay trivial.
