# Making a service observable

In order. Do this before the first deploy; a service that reaches staging without it will have its first incident before its first instrumentation PR.

## Bootstrap

- [ ] `main` calls `telemetry.Init(ctx, service, version)` first and defers `shutdown` with a bounded context.
- [ ] `service.name` equals the app directory name; `service.version` from the build; `deployment.environment.name` from the environment.
- [ ] Exports to the collector via `OTEL_EXPORTER_OTLP_ENDPOINT`; no backend-specific configuration in the service.
- [ ] Composite propagator (`TraceContext` + `Baggage`) set globally.

## Boundaries

- [ ] Inbound HTTP wrapped with `otelhttp.NewHandler`; Connect handlers registered with the `otelconnect` interceptor.
- [ ] Outbound HTTP uses `otelhttp.NewTransport`; database client uses instrumented driver; queue publish/consume propagate context.
- [ ] Boundary interceptor logs one line per request outcome with the required fields.

## Signals

- [ ] `slog` default handler is the `otelslog` bridge; no `fmt.Println` / `log.Printf` (linter enforced).
- [ ] Domain spans named `<domain>.<entity>.<operation>` with attributes, not events, for facts.
- [ ] Errors recorded on spans (`RecordError` + `SetStatus`).
- [ ] RED metrics visible for every endpoint in the local stack.
- [ ] Any custom metric: enumerable labels, declared unit, description.

## Operations

- [ ] SLOs written in the README with owner.
- [ ] Burn-rate alerts instantiated in `infra/alerts/<service>.yaml` with runbook links.
- [ ] Dashboard provisioned from `infra/` with RED + SLO panels.
- [ ] Runbook exists for each alert: meaning, first checks, mitigation, escalation.

## Verify

- [ ] Send one request locally; find it by trace ID in traces, logs, and as an exemplar in metrics. If any of the three is missing, the service is not done.
