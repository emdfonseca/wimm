# Go

## Workspace

```text
// go.work
go <version pinned in devbox.json>

use (
    ./apps/api
    ./packages/telemetry
    ./packages/contracts/gen/go
)
```

Each directory is its own module with its own `go.mod`. No `replace` directives; `go.work.sum` is committed.

## Module paths

`<module-prefix>/<dir-path>`, e.g. `github.com/org/repo/packages/telemetry`. The import path mirrors the filesystem.

Internal layout, tests, lint: `.claude/rules/go.md`.
