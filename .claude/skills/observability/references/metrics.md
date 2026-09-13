# Metrics

## RED comes from instrumentation

`otelhttp` and `otelconnect` emit request count, error count, and duration histograms per route, with the semantic-convention attributes (`http.request.method`, `http.route`, `http.response.status_code`, `rpc.method`). Do not re-implement these. Confirm they appear for a new endpoint and move on.

## Custom metrics

Only for things RED does not say: business events, queue depths, cache hit ratios, batch sizes.

```go
meter := otel.Meter("billing")
invoicesCreated, _ := meter.Int64Counter("billing.invoices.created",
	metric.WithUnit("{invoice}"),
	metric.WithDescription("Invoices successfully created"))

invoicesCreated.Add(ctx, 1, metric.WithAttributes(attribute.String("billing.plan", plan)))
```

Naming: `<domain>.<thing>.<what>` in the OTel dotted form; exporters convert to `billing_invoices_created_total` for Prometheus-style backends. Units are declared, not encoded in the name. Instrument kinds:

| Want | Use |
|---|---|
| Things that happen | Counter |
| Current level (queue depth, connections) | UpDownCounter or Gauge |
| Distribution (sizes, durations) | Histogram, with explicit buckets if the default ones do not fit |

## Cardinality

The rule that matters most, because breaking it takes down the metrics backend for everyone:

**A label's value set must be enumerable in advance and small.** Plan tier, region, status code, method: fine. User ID, customer ID, email, URL with parameters, error message text, anything user-supplied: never. Those go on spans and logs.

Per-endpoint labels use the *route template* (`/v1/invoices/{id}`), never the concrete path. `otelhttp` does this when the router provides the template; confirm it does.

## Dashboards

Built from RED per service plus the SLO panels. Custom metrics get a panel only when someone would act on it. Dashboards live in `infra/` as code (Grafana JSON, provisioned), reviewed like anything else — a dashboard hand-edited in the UI is a dashboard that vanishes on the next redeploy.
