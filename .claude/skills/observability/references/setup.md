# SDK setup

## Principles

- One `telemetry` package per language in `packages/` (`packages/telemetry` for Go), called once from `main`. Services do not configure exporters themselves.
- Configuration by the standard OTel variables only: `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_RESOURCE_ATTRIBUTES`. No service-specific env names.
- Export to a collector, never directly to a backend. The collector is the only component that knows which backend exists.

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
	otel.SetTextMapPropagator(propagation.TraceContext{})
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

`main` calls `Init`, defers `shutdown` with a bounded context, wraps its HTTP mux with `otelhttp.NewHandler` and its Connect handlers with the `otelconnect` interceptor. Nothing else in the service imports the SDK packages; they import `go.opentelemetry.io/otel` (the API) only.

`otelhttp` and `otelslog` ship from the contrib repo's experimental module sets (v0.x); pin them in `go.mod`.

## TypeScript (SvelteKit server)

`@opentelemetry/sdk-node` with the OTLP exporters, initialised in `hooks.server.ts` before anything else, resource attributes from the same env variables. No SDK in the browser bundle by default.

## Python

`opentelemetry-sdk` + `opentelemetry-exporter-otlp`, initialised at process start; `opentelemetry-instrument` auto-instrumentation is acceptable for frameworks.
