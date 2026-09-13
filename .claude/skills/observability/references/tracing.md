# Tracing

## Propagation

W3C `traceparent` / `tracestate` via the `TraceContext` propagator on every hop: HTTP, Connect, queues (context into message headers, restored in the consumer), scheduled jobs (root span named for the job).

## What gets a span

- Every inbound request (`otelhttp` / `otelconnect`, free).
- Every outbound call: `otelhttp.NewTransport`, the instrumented database driver, queue publish/consume.
- Units of domain work someone might time or search for: `billing.invoice.create`, `billing.tax.calculate`. A handful per request.

No span for trivial helpers, per-item loop bodies (one span with a count attribute), or anything under a millisecond.

## Naming

`<domain>.<entity>.<operation>`, lowercase, constant per operation. Variable parts are attributes.

```go
var tracer = otel.Tracer("billing")   // once per package

ctx, span := tracer.Start(ctx, "billing.invoice.create",
	trace.WithAttributes(attribute.String("billing.customer_id", customerID)))
defer span.End()
```

## Attributes, events, status

- Attributes carry facts: semantic-convention names where they exist, `<domain>.<name>` otherwise, same names as the log fields.
- Events mark moments (`retry`, `cache.miss`). Sparingly.
- On failure: `span.RecordError(err)` and `span.SetStatus(codes.Error, msg)`.
