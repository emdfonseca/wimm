# Logging

## The contract

`log/slog`, JSON handler bridged through `otelslog`. The request context is passed on every call; that is what attaches `trace_id` and `span_id`.

## Fields

Set by `packages/telemetry` or the bridge, never hand-added:

```text
time  level  msg  service  service.version  deployment.environment.name  trace_id  span_id
```

Errors go in `err` via `slog.Any("err", err)`, always that name. Domain fields are `<domain>.<name>` in snake_case (`billing.invoice_id`, `auth.subject`). Standard fields use OTel semantic-convention names (`http.request.method`, `http.response.status_code`), never invented `method` / `status`.

## Levels

| Level | Means | Prod default |
|---|---|---|
| `Error` | Failed; a person should look | on |
| `Warn` | Degraded but handled: fallback used, retry succeeded | on |
| `Info` | One line per request outcome, one per job run, lifecycle events | on |
| `Debug` | Step-by-step detail for local development | off |

An `Error` that fires on every request during a known condition is a `Warn`. If the level does not change what someone does, it is wrong.

## One line per request

The boundary interceptor logs the outcome: route, status/code, duration, trace ID, caller identity. Handlers and domain code add no "entering X" lines; spans carry that. A domain component logs only what the span cannot carry: a decision with a reason, or an error swallowed on purpose.

## Never log

Secrets, tokens, API keys, passwords, session cookies, full request or response bodies, card numbers, personal data beyond an opaque identifier. Log the ID or a hash instead. Each package keeps a small allowlist of known-safe fields in its `doc.go`; anything else needs a reason.

## `fmt.Println`, `log.Printf`, `console.log`, `print`

Never. Enforced by `forbidigo` (Go) and `no-console` (TS).
