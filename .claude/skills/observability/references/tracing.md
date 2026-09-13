# Tracing

## Propagation first

W3C `traceparent` / `tracestate` via the composite propagator, on every hop: HTTP, Connect, queues (put the context into message headers and restore it in the consumer), scheduled jobs (start a root span with the job name). A trace that ends at a queue is a trace that ends where the interesting failures begin.

## What gets a span

- Every inbound request (from `otelhttp` / `otelconnect` — free).
- Every outbound call: HTTP client via `otelhttp.NewTransport`, database via the driver's instrumentation, queue publish/consume.
- Units of domain work someone might want to time or find: `billing.invoice.create`, `billing.tax.calculate`. A handful per request, not one per function.

Do not span: trivial helpers, loops per item (use one span with a count attribute), anything under a millisecond.

## Naming

`<domain>.<entity>.<operation>` in lowercase, describing the operation from the outside. Constant per operation — never include IDs, paths with parameters, or anything variable. The variable parts are attributes.

```go
ctx, span := tracer.Start(ctx, "billing.invoice.create",
	trace.WithAttributes(attribute.String("billing.customer_id", customerID)))
defer span.End()
```

The tracer is obtained once per package: `otel.Tracer("billing")`, named for the package.

## Attributes, events, status

- **Attributes** carry facts: IDs, counts, choices made. Semantic-convention names where they exist, `<domain>.<name>` otherwise. Same names as the log fields, so a search works across both.
- **Events** mark moments inside a span: `retry`, `cache.miss`. Sparingly.
- **Status**: `span.SetStatus(codes.Error, msg)` and `span.RecordError(err)` on failure. An error that is not recorded on the span is invisible to tail sampling and to error-rate dashboards built on spans.

## Sampling

Head sampling in the SDK is off (`AlwaysSample`) in dev and staging. In prod, sampling is the collector's job: tail-based, keeping 100% of errors and of requests over the latency SLO, and a ratio of the rest. Tail sampling in the collector is why services never need to know the sample rate.

## Baggage

Cross-service context that is not a trace ID — tenant, request origin — goes in baggage and is copied onto spans by a processor. Keep it tiny; it rides on every hop.
