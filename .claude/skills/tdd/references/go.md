# Go

## Layout

`foo_test.go` beside `foo.go`, package `foo_test` (external test package) by default so tests exercise the public surface. Use the internal package `foo` only when testing an unexported piece that genuinely cannot be reached otherwise — and treat that as a smell to revisit in the refactor step.

## Table-driven

```go
func TestInvoice_Total(t *testing.T) {
	tests := []struct {
		name  string
		items []LineItem
		want  Money
		err   error
	}{
		{name: "sums line items", items: []LineItem{{Cents: 100}, {Cents: 250}}, want: Money{Cents: 350}},
		{name: "rejects negative line", items: []LineItem{{Cents: -1}}, err: ErrNegativeAmount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Total(tt.items)
			if !errors.Is(err, tt.err) {
				t.Fatalf("err = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("Total() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

Add one row per cycle. The row name is the behaviour sentence. Run a single row while cycling: `go test ./... -run 'TestInvoice_Total/rejects_negative'`.

## Handlers

Test Connect handlers through the generated client against an `httptest.Server`, so the interceptor chain and the error mapping are exercised:

```go
mux := http.NewServeMux()
mux.Handle(v1connect.NewBillingServiceHandler(srv))
ts := httptest.NewServer(mux)
defer ts.Close()
client := v1connect.NewBillingServiceClient(http.DefaultClient, ts.URL)

_, err := client.CreateInvoice(ctx, connect.NewRequest(&v1.CreateInvoiceRequest{}))
if connect.CodeOf(err) != connect.CodeInvalidArgument { t.Fatalf("code = %v", connect.CodeOf(err)) }
```

One test per error code the contract can return — that is the `api-contract` rule, and it is a natural TDD sequence: red on `not_found`, green, red on `permission_denied`, green.

## Fakes

Define the interface where it is consumed and keep it small:

```go
// in package invoice
type Store interface {
	Save(ctx context.Context, inv Invoice) error
	Get(ctx context.Context, id string) (Invoice, error)
}
```

The test uses an in-memory map implementation in `invoice_test.go`. No mocking framework; a generated mock of a 30-method interface is the thing this avoids.

## Time, IDs, randomness

Inject `func() time.Time` and an ID generator; the fake returns fixed values. `time.Now()` inside domain code is untestable by definition.

## Backfills and migrations

A backfill test runs the job twice against a fake store and asserts the second run changes nothing (idempotent) and that stopping after N batches and resuming reaches the same end state. `just migrate-redo` is the migration's own test and belongs in the `data-migrations` cycle, not here.

## Useful flags

`go test -race ./...` in `just test` — always. `-count=1` to defeat the cache when a flaky test is suspected. `-run` and `-v` while cycling; never `-v` in CI output.
