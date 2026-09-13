# Metrics

## RED comes from instrumentation

`otelhttp` and `otelconnect` emit request count, error count, and duration histograms per route with semantic-convention attributes (`http.request.method`, `http.route`, `http.response.status_code`, `rpc.method`). Confirm they appear for a new endpoint; do not re-implement.

## Custom metrics

Only for what RED does not say: business events, queue depths, cache hit ratios, batch sizes.

```go
meter := otel.Meter("billing")
invoicesCreated, _ := meter.Int64Counter("billing.invoices.created",
	metric.WithUnit("{invoice}"),
	metric.WithDescription("Invoices successfully created"))

invoicesCreated.Add(ctx, 1, metric.WithAttributes(attribute.String("billing.plan", plan)))
```

Name: `<domain>.<entity>.<what>`, OTel dotted form; exporters derive `billing_invoices_created_total` for Prometheus-style backends. Units are declared, never encoded in the name.

| Want | Use |
|---|---|
| Things that happen | Counter |
| Current level (queue depth, connections) | UpDownCounter or Gauge |
| Distribution (sizes, durations) | Histogram, explicit buckets if the defaults do not fit |

## Cardinality

A label's value set must be enumerable in advance and small. Plan tier, region, status code, method: fine. User ID, customer ID, email, URL with parameters, error text, anything user-supplied: never; those go on spans and logs.

Per-endpoint labels use the route template (`/v1/invoices/{id}`), never the concrete path. `otelhttp` does this when the router supplies the template; confirm it does.
