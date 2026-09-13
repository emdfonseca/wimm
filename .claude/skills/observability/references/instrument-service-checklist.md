# Making a service observable

Do this before the first deploy.

## Bootstrap

- [ ] `main` calls `telemetry.Init(ctx, service, version)` first and defers `shutdown` with a bounded context.
- [ ] `service.name` equals the app directory name; `service.version` from the build; `deployment.environment.name` from the environment.
- [ ] Exports via `OTEL_EXPORTER_OTLP_ENDPOINT`; no backend-specific configuration in the service.
- [ ] `TraceContext` propagator set globally.

## Boundaries

- [ ] Inbound HTTP wrapped with `otelhttp.NewHandler`; Connect handlers registered with the `otelconnect` interceptor.
- [ ] Outbound HTTP uses `otelhttp.NewTransport`; database client uses the instrumented driver; queue publish/consume propagate context.
- [ ] Boundary interceptor logs one line per request outcome with the required fields.

## Signals

- [ ] `slog` default handler is the `otelslog` bridge; no `fmt.Println` / `log.Printf` (linter enforced).
- [ ] Domain spans named `<domain>.<entity>.<operation>`; facts as attributes, not events.
- [ ] Errors recorded on spans (`RecordError` + `SetStatus`).
- [ ] RED metrics visible for every endpoint in the local stack.
- [ ] Any custom metric: `<domain>.<entity>.<what>`, enumerable labels, declared unit, description.

## Verify

- [ ] Send one request locally; find it by trace ID in traces, logs, and as an exemplar in metrics. If any of the three is missing, the service is not done.
