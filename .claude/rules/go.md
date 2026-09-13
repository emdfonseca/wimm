---
paths:
  - "**/*.go"
---

# Go conventions

Decisions, not tutorials. Effective Go and the Code Review Comments apply as written; this file only records where this repo picked one option.

- **Logging** — `log/slog`, obtained through `packages/telemetry`, context passed on every call so `trace_id` attaches. Never `fmt.Println` / `log.Printf` / `log.Fatal` outside `main` (enforced by `forbidigo`).
- **Errors** — stdlib only. Wrap with `fmt.Errorf("doing X: %w", err)`. Sentinel `ErrNotFound`-style values for conditions callers branch on; typed errors when data must ride along. Domain code never knows about Connect codes or HTTP status — mapping happens once, in the handler package (see `api-contract`).
- **Wiring** — explicit constructors called from `cmd/<app>/main.go`. No DI framework, no `init()` side effects, no package-level mutable state. If `main` grows past ~100 lines, group wiring into `internal/app`, still hand-written.
- **HTTP** — Connect handlers registered on a `net/http` mux. Public REST uses `net/http` 1.22+ `ServeMux` method patterns (`mux.HandleFunc("GET /v1/invoices/{id}", …)`). No third-party router.
- **Layout** — `cmd/<app>/main.go` is wiring only; everything else under `internal/`. No top-level `pkg/`. Promote out of `internal/` only when a second consumer exists (see `monorepo`).
- **Context** — `ctx context.Context` is the first parameter of anything that does I/O or can block. Propagate it; `context.Background()` appears in `main` and tests only.
- **Interfaces** — declared where consumed, sized to what the consumer uses. Return concrete types.
- **Tests** — stdlib `testing`; table-driven with `t.Run`; external `_test` package by default; fakes over mocks; `go test -race ./...` in `just test`. No assertion library.
- **Handler tests** — `net/http/httptest.NewServer` around the real Connect handler, called through the generated Connect client. No mocks of the transport.
- **Lint** — root `.golangci.yml` inherited by every module; `depguard` encodes the dependency rules; `forbidigo` bans unstructured logging. `GOFLAGS=-mod=readonly` from devbox so builds never rewrite `go.mod`.
- **Generated code** — `*.pb.go`, `*.connect.go`, `*.gen.go` are never edited; change the source and run `just gen`.
