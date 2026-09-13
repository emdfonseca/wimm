# Logging

## The contract

`log/slog`, JSON handler bridged through `otelslog` so records become OTel log records with trace context. Loggers are obtained from context (`slog.Default()` with the request context passed on every call) — passing the context is what attaches `trace_id` and `span_id`. A log call without a context is a log line nobody can correlate.

## Required fields

Set by the telemetry package or the bridge; never hand-added:

```text
time  level  msg  service  service.version  deployment.environment.name  trace_id  span_id
```

Domain fields follow `<domain>.<name>` with `snake_case`: `billing.invoice_id`, `auth.subject`, `http.route`. Standard fields use OTel semantic convention names where one exists (`http.request.method`, `http.response.status_code`, `db.system`) rather than inventing `method` / `status`.

Errors go in `err` via `slog.Any("err", err)` — the field is always `err`, so `err != null` is a filter that works everywhere.

## Levels

| Level | Means | Prod default |
|---|---|---|
| `Error` | Something failed that a person should look at; usually paired with a non-2xx or a failed job | on |
| `Warn` | Degraded but handled — fallback used, retry succeeded, deprecated path taken | on |
| `Info` | The request-level narrative: one line per request outcome, one per job run, lifecycle events | on |
| `Debug` | Step-by-step detail for local development | off |

An `Error` that fires on every request during a known condition is a `Warn`. A `Warn` nobody would ever act on is `Info`. Levels are for triage; if the level does not change what someone does, it is wrong.

## One line per request

The boundary interceptor logs the outcome: route, status/code, duration, trace ID, caller identity. Handlers and domain code do not add "entering X" / "leaving X" lines — that is what spans are for. A domain component logs when it has something the span cannot carry: a decision with a reason (`Warn`: fallback chosen because …), or an error it is swallowing on purpose.

## Never log

Secrets, tokens, API keys, passwords, session cookies, full request or response bodies, card numbers, personal data beyond an opaque identifier. The collector's redaction processor is a backstop, not the control. If a value might be sensitive, log its ID or a hash.

The allowlist approach: a small set of known-safe fields per domain, documented in that package's `doc.go`. Anything not on it needs a reason before it is logged.

## `fmt.Println`, `log.Printf`, `console.log`, `print`

Do not. They bypass the bridge, lose the trace context, and produce unstructured lines the pipeline cannot parse. The linter rule (`forbidigo` for Go, `no-console` for TS) enforces it; a PR that adds one is a PR that will be undebuggable at 3am.
