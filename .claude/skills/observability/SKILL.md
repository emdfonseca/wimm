---
name: observability
description: Instrumentation contract for every service - OpenTelemetry traces, metrics, and logs over OTLP to a collector, resource attributes, slog via otelslog with a fixed field set, span and metric naming, cardinality. Use whenever someone adds or changes logging, tracing, or metrics; creates a service or endpoint; asks what to log or how to name a span or metric; sets up the OTel SDK. Even one log line.
paths:
  - "**/otel*.go"
  - "**/telemetry/**"
  - "**/otel-collector*.{yaml,yml}"
---

# Observability

Every service emits traces, metrics, and logs through OpenTelemetry over OTLP to a collector. The backend is undecided; services never know it.

## Rules

1. Resource attributes at SDK init: `service.name` = app directory name, `service.version` = git SHA, `deployment.environment.name` = `dev`|`staging`|`prod`.
2. Logs: `log/slog` through `otelslog`, JSON, context on every call. Fields: `time level msg service service.version deployment.environment.name trace_id span_id`; errors in `err`. Domain fields `<domain>.<name>`, snake_case. One line per request outcome, from the boundary interceptor. Never `fmt.Println` / `log.Printf` / `console.log`.
3. Spans: `otelhttp` / `otelconnect` name server spans; domain spans are `<domain>.<entity>.<operation>` (`billing.invoice.create`), constant per operation. Facts are attributes, named like log fields.
4. Metrics: RED comes from the interceptors. Custom metrics are OTel dotted `<domain>.<entity>.<what>` (`billing.invoices.created`) with a declared unit. Label values enumerable in advance; IDs, emails, URLs, free text never.
5. Propagate W3C `traceparent` on every hop, queues included.

## References

| Read | When |
|---|---|
| `references/setup.md` | `packages/telemetry` Init, env vars, TS, Python |
| `references/logging.md` | Levels, never-log list |
| `references/tracing.md` | What gets a span, attributes, status |
| `references/metrics.md` | Instrument kinds, route labels |
| `references/instrument-service-checklist.md` | Checklist |
