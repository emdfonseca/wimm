---
name: api-contract
description: Shape of every service API - Connect over protobuf for our callers, REST + OpenAPI only for third parties, one error vocabulary, cursor pagination, idempotency keys, deadlines, additive-only versioning via buf breaking. Use whenever someone writes or changes a .proto, an OpenAPI document, a Connect or HTTP handler, or an endpoint's request/response/error shape, or adds an endpoint, however small.
paths:
  - "**/*.proto"
  - "**/openapi.{yaml,yml,json}"
  - "**/buf.yaml"
  - "**/buf.gen.yaml"
---

# API contract

Three surfaces, and no fourth (ADR 0001):

```text
our own services, server to server   Connect over protobuf
our own web client, browser to app   REST over HTTP and JSON, in the app's own server routes
third parties and webhooks           REST + OpenAPI 3.1, JSON, RFC 9457 errors
```

The browser never speaks Connect. The web client's REST is undocumented and unversioned on purpose — its only consumer ships in the same deploy — so rules 3 to 6 below apply to the first and third rows, not to it. A second consumer moves it to the third row, OpenAPI and all.

## Rules

1. Contract first. The `.proto` (or OpenAPI document) is written and generated before the handler. Handlers implement the generated interface.
2. One error vocabulary: Connect codes for both protocols, mapped from domain errors once at the handler boundary. Nothing internal crosses it.
3. Every list pages by cursor. Every retryable mutation takes an idempotency key. Every RPC states its deadline in its contract comment.
4. Additive only within a major. `buf breaking --against '.git#branch=main'` fails CI; breaking → new `v2` package beside `v1`.
5. Naming: package `<org>.<domain>.v1`, service `<Domain>Service`, RPC `VerbNoun`, messages `<Rpc>Request` / `<Rpc>Response`, fields `snake_case` (no `json_name`). REST paths: plural kebab-case under `/v1/`.
6. Validation lives in the contract (`buf validate` / OpenAPI schema), not only in code.

## References

| Read | When |
|---|---|
| `references/connect.md` | Proto layout, `buf.yaml`, handler, interceptor order |
| `references/errors.md` | Code → HTTP mapping, what never leaks |
| `references/pagination-idempotency-deadlines.md` | Field names, idempotency window, deadline comment |
| `references/versioning.md` | Additive vs breaking, new major |
| `references/new-endpoint-checklist.md` | Adding an endpoint |
