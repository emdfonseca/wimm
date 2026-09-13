# Connect

Connect (connectrpc.com) speaks gRPC, gRPC-Web, and its own HTTP/JSON-friendly protocol from one handler. That is why it is the default: Go services talk to each other over gRPC, the SvelteKit app calls the same handlers from the browser with `connect-es`, and nobody maintains a gateway.

## Layout

```text
packages/contracts/
├── buf.yaml                 # lint + breaking config
├── buf.gen.yaml             # go, connect-go, es, python targets
├── proto/<org>/billing/v1/
│   ├── billing.proto        # messages + service
│   └── errors.proto         # domain error detail messages, if any
└── gen/                     # generated; never edited (hook-enforced)
```

One proto package per domain per major version. A domain that needs two services still lives in one package; a package that is only messages shared by several domains is a smell — it is usually the `shared` junk drawer from the cohesion rules, wearing a `.proto` extension.

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

`buf lint` enforces naming (`<org>.<domain>.v1`, `VerbNoun` RPCs, request/response suffixes). `buf breaking --against '.git#branch=main'` in CI is what makes "additive only" a failing check rather than a review opinion. Both belong in the `packages/contracts` justfile's `lint` verb.

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
  string idempotency_key = 2;   // see pagination-idempotency-deadlines.md
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

Reserve removed field numbers (`reserved 4;`) so they are never reused with a different meaning by accident.

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

Handlers convert, call the domain, convert back, and map errors. Business logic in a handler is business logic that cannot be tested without the transport.

## Interceptors

One chain, declared once per service, in this order: recovery, telemetry (`otelconnect`), auth, validation, then the handler. Auth before validation so unauthenticated callers cannot probe the schema. The observability skill owns what the telemetry interceptor records.

## Client side

`connect-es` generates TypeScript clients from the same protos; the SvelteKit app depends on `@repo/contracts` and never hand-writes request types. Go callers use the generated `v1connect.NewBillingServiceClient` with an `otelhttp`-wrapped transport. Python uses the generated stubs from the same `buf.gen.yaml`.
