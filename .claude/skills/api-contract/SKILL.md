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

Our callers → Connect over protobuf. Third parties and webhooks → REST + OpenAPI 3.1, JSON, RFC 9457 errors. No third kind.

## Rules

1. Contract first. The `.proto` (or OpenAPI document) is written and generated before the handler. Handlers implement the generated interface; `gen/` is never edited (protect-generated hook, `monorepo-standard`).
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
