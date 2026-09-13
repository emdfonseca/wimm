# Connect

## Layout

```text
packages/contracts/
├── buf.yaml                 # lint + breaking config
├── buf.gen.yaml             # go, connect-go, es, python targets
├── proto/<org>/billing/v1/
│   ├── billing.proto        # messages + service
│   └── errors.proto         # domain error detail messages, if any
└── gen/                     # generated
```

One proto package per domain per major version. A package that is only messages shared by several domains is the `common` junk drawer wearing a `.proto` extension.

## buf

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
lint:
  use: [STANDARD]
breaking:
  use: [FILE]
```

`buf lint` and `buf breaking --against '.git#branch=main'` both run from the package's `lint` verb.

## Service definition

```protobuf
syntax = "proto3";
package org.billing.v1;

service BillingService {
  rpc CreateInvoice(CreateInvoiceRequest) returns (CreateInvoiceResponse);
  rpc ListInvoices(ListInvoicesRequest) returns (ListInvoicesResponse);
}

message CreateInvoiceRequest {
  string customer_id = 1;
  string idempotency_key = 2;
  repeated LineItem items = 3;
}

message ListInvoicesRequest {
  string customer_id = 1;
  int32 page_size = 2;          // server clamps; 0 means default
  string page_token = 3;
}

message ListInvoicesResponse {
  repeated Invoice invoices = 1;
  string next_page_token = 2;   // empty means no more
}
```

Reserve removed field numbers (`reserved 4;`).

## Handler

```go
func (s *Server) CreateInvoice(ctx context.Context, req *connect.Request[v1.CreateInvoiceRequest]) (*connect.Response[v1.CreateInvoiceResponse], error) {
	inv, err := s.billing.Create(ctx, fromProto(req.Msg))
	if err != nil {
		return nil, toConnectError(err)   // the single mapping point — see errors.md
	}
	return connect.NewResponse(&v1.CreateInvoiceResponse{Invoice: toProto(inv)}), nil
}
```

Handlers convert, call the domain, convert back, map errors. No business logic.

## Interceptors

One chain per service, in this order: recovery, telemetry (`otelconnect`), auth, validation, handler. Auth before validation so unauthenticated callers cannot probe the schema.

## Clients

`connect-es` generates the TypeScript client; consumers import generated types from `@repo/contracts` and never hand-write request types. Go callers use the generated `v1connect.New<Domain>ServiceClient` with an `otelhttp`-wrapped transport. Python uses the stubs from the same `buf.gen.yaml`.
