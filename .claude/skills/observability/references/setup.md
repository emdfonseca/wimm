# SDK setup

## Principles

- One `telemetry` package per language in `packages/` (`packages/telemetry` for Go), called once from `main`. Services do not configure exporters themselves.
- Configuration by environment, following the standard OTel variables: `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_RESOURCE_ATTRIBUTES`. No service-specific env names.
- Export to a collector, never directly to a backend. The collector does batching, retries, tail sampling, redaction, and routing, and it is the only component that knows which backend exists.

## Go

```go
// packages/telemetry/telemetry.go
func Init(ctx context.Context, service, version string) (shutdown func(context.Context) error, err error) {
	res, err := resource.New(ctx,
		resource.WithFromEnv(),      // OTEL_RESOURCE_ATTRIBUTES, incl. deployment.environment.name
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(service),
			semconv.ServiceVersion(version),
		),
	)
	// traces
	traceExp, err := otlptracegrpc.New(ctx)
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp), sdktrace.WithResource(res))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	// metrics
	metricExp, err := otlpmetricgrpc.New(ctx)
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)), sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)
	// logs
	logExp, err := otlploggrpc.New(ctx)
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)), sdklog.WithResource(res))
	global.SetLoggerProvider(lp)
	slog.SetDefault(slog.New(otelslog.NewHandler(service)))

	return func(ctx context.Context) error { /* shut down tp, mp, lp in order */ }, nil
}
```

`main` calls `Init`, defers `shutdown` with a bounded context, and wraps its HTTP mux with `otelhttp.NewHandler` and its Connect handlers with the `otelconnect` interceptor. Nothing else in the service imports the SDK packages; they import `go.opentelemetry.io/otel` (the API) only.

A version note worth knowing: `otelhttp` and the `otelslog` bridge are published from the contrib repo's *experimental* module sets (v0.x), while the core API and SDK are stable v1. Pin them in `go.mod` and expect occasional minor-version churn; the API you call is small and stable in practice.

## TypeScript (SvelteKit server)

`@opentelemetry/sdk-node` with the OTLP exporters, initialised in `hooks.server.ts` before anything else, resource attributes from the same env variables. Browser-side tracing is opt-in and separate; do not ship the SDK to the client by default.

## Python

`opentelemetry-sdk` + `opentelemetry-exporter-otlp`, initialised at process start, `opentelemetry-instrument` auto-instrumentation acceptable for frameworks, resource attributes from env.

## The collector

Lives in `infra/`, one config, deployed alongside the services. Responsibilities, in this order: receive OTLP, batch, redact (the attribute processor drops anything on the denylist), tail-sample traces (keep all errors and slow requests, sample the rest), export. Changing the backend is a change to the exporter section and nothing else.

## Local development

Run the collector plus a local all-in-one (Grafana LGTM image, or Jaeger + Prometheus) via `just dev` in `infra/local`. Developers should see their own traces from day one; observability that only exists in prod is observability nobody has looked at.
