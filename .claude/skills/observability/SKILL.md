---
name: observability
description: The instrumentation contract every service in this repo follows - OpenTelemetry for traces, metrics, and logs exported over OTLP to a collector (backend deliberately undecided), required resource attributes, structured slog logging through the OTel bridge with a fixed field set, span naming, RED metrics per endpoint, cardinality rules, and symptom-based alerts with runbooks. Use this whenever someone adds or changes logging, tracing, metrics, or alerting; creates a new service or endpoint; asks what to log, which fields, at what level, or how to name a span or metric; sets up an OTel SDK, collector config, or exporter; writes or reviews a dashboard or alert rule; or is debugging why a request cannot be found in traces or logs. Reach for it even for "just add a log line" - the field set and the level are the whole point.
paths:
  - "**/otel*.go"
  - "**/telemetry/**"
  - "**/observability/**"
  - "**/otel-collector*.{yaml,yml}"
  - "**/alerts/**"
  - "**/*.alerts.{yaml,yml}"
---

# Observability

**Every service emits traces, metrics, and logs through OpenTelemetry, correlated by trace ID, exported over OTLP to a collector.** The backend behind the collector is undecided on purpose: the collector is the seam, and nothing in a service knows or cares what is on the other side. That is what lets the backend choice change without touching a line of application code.

```text
service ──OTLP──▶ collector ──▶ (whatever backend: Grafana LGTM, Datadog, Honeycomb…)
```

The failure mode this prevents is the usual one: three services logging three different field names for the same request ID, spans named after function names, a `user_id` label on a metric that explodes cardinality, and an alert that fires on CPU while users see errors. None of that is hard to avoid at the start; all of it is expensive to fix once dashboards depend on it.

## Start here, every time

1. **Can this request be found?** Every log line, span, and metric exemplar carries the trace ID. If a support ticket arrives with a request ID and you cannot pull up the whole story across services, the instrumentation is incomplete regardless of how much of it there is.
2. **Which signal is this?** A thing that happened once with context → span event or log. A rate/duration/count to graph → metric. The path a request took → span. Logging a metric or graphing logs is the smell.
3. **Will a stranger understand the name?** `billing.invoice.create` reads. `handleCreateInvoiceV2` does not. Names describe the operation from the outside.

## Where to read next

| Read this | When |
|---|---|
| `references/setup.md` | Bootstrapping the SDK in a Go, TS, or Python service; resource attributes; OTLP exporter; the collector. |
| `references/logging.md` | What to log, the required field set, levels, what must never be logged, slog + `otelslog`. |
| `references/tracing.md` | Span naming, attributes vs events, propagation, sampling, what gets its own span. |
| `references/metrics.md` | RED metrics from instrumentation, custom metric naming, units, and the cardinality rules. |
| `references/alerts-and-slos.md` | Symptom-based alerting, SLOs and burn rate, runbooks, what belongs in `infra/`. |
| `references/instrument-service-checklist.md` | Making a new service or endpoint observable, end to end. |

## Rules that are constantly needed

### Resource attributes are mandatory

Every process sets, at SDK init, from environment:

```text
service.name                 = the app directory name (apps/billing → "billing")
service.version              = git SHA or release tag
deployment.environment.name  = dev | staging | prod
```

These are how everything is filtered and joined. A process without them is unattributable noise in every dashboard. `service.name` equals the directory name so the monorepo, CODEOWNERS, logs, and dashboards all agree on what a thing is called.

### Logs are structured, correlated, and sparse

`slog` with the `otelslog` bridge, JSON, one event per line, always carrying `trace_id` and `span_id` (the bridge does this when the context is passed — so pass the context). Fixed field names across all services:

```text
trace_id  span_id  service  level  msg  err  duration_ms  <domain>.<entity>_id
```

One log line at the boundary per request outcome (the telemetry interceptor emits it), not one per step — steps are span events. `Error` means someone should look; `Warn` means degraded but handled; `Info` is the request-level narrative; `Debug` is off in prod. Never log secrets, tokens, full request bodies, or personal data; the field allowlist in `references/logging.md` exists so this is a lookup rather than a judgment call under pressure.

### Spans are named for the operation, attributes over events

`otelhttp` and `otelconnect` name the server spans; you name the rest as `<domain>.<operation>` — `billing.invoice.create`, `db.invoices.insert`. Put facts on the span as attributes (`billing.invoice_id`, `db.rows_affected`), not in the span name and not as events. Propagate W3C `traceparent` everywhere, including through queues. A span for every function is as useless as no spans; a span per unit of work someone might need to time is right.

### Metrics: RED for free, custom ones named and bounded

Instrumentation gives request rate, error rate, and duration per endpoint. Custom metrics follow `<domain>_<thing>_<unit>` (`billing_invoices_created_total`, `queue_depth_items`), carry units, and never have an unbounded label — no IDs, emails, URLs with parameters, or free text. A label whose value set you cannot enumerate on a whiteboard will eventually take down the metrics backend.

### Alert on symptoms, with a runbook

Alerts fire on what users experience — error ratio, latency SLO burn, availability — not on CPU, memory, or queue depth unless those are the SLO. Every alert links a runbook stating what it means, what to check first, and how to mitigate. An alert without a runbook is a notification; it trains people to acknowledge and ignore.

## Checkpoints

- **New service** — walk `references/instrument-service-checklist.md` before the first deploy. Resource attributes, propagation, and the boundary interceptor are the minimum; a service deployed without them cannot be debugged when it first misbehaves, which is usually the first week.
- **New endpoint** — confirm its RED metrics appear and its server span carries the domain attributes. Free from the interceptor, but only if the handler was registered through it.
- **New metric label** — enumerate the values. If you cannot, it is a span attribute or a log field, not a label.
- **New alert** — it has an SLO or a symptom behind it, a runbook link, and an owner. Otherwise it is a dashboard panel.
- **Reviewing a PR with a `log.Printf` or `fmt.Println`** — that is the signal the request will not be findable. Say so.
