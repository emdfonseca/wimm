# Public REST

REST exists for people who cannot consume protobuf: third-party developers, webhooks, anyone integrating without our client libraries. It is a product surface, reviewed as one. Nothing internal is REST merely because it was faster.

## Shape

- OpenAPI 3.1 document in `packages/contracts/openapi/<product>.yaml`, the source of truth; server routes and any generated SDKs derive from it.
- Paths: `/v1/<plural-noun>`, `/v1/<plural-noun>/{id}`, kebab-case, no verbs. Actions that do not fit CRUD are sub-resources (`POST /v1/invoices/{id}/send`), not query parameters.
- JSON bodies, `lowerCamelCase` keys, RFC 3339 timestamps in UTC, strings for IDs and money.
- Every response has `Content-Type: application/json`; errors use `application/problem+json`.

## Errors: RFC 9457 problem details

```json
{
  "type": "https://api.example.com/problems/invalid-argument",
  "title": "Invalid argument",
  "status": 400,
  "detail": "customerId must be a valid customer identifier",
  "instance": "/v1/invoices",
  "code": "invalid_argument",
  "violations": [{ "field": "customerId", "description": "unknown customer" }]
}
```

`code` is the Connect code name — the same vocabulary as internal services, so a public failure and an internal one are described the same way in logs and dashboards. `references/errors.md` has the status mapping.

## What the public surface also needs

- `Idempotency-Key` on every `POST` that creates or charges; replay returns the original response and status.
- Cursor pagination: `?pageSize=&pageToken=` in, `nextPageToken` out. Same semantics as Connect, different casing.
- Rate limits surfaced with `RateLimit-*` headers and `429` + `resource_exhausted`.
- Explicit deprecation: `Deprecation` and `Sunset` headers on retiring routes, a dated notice in the OpenAPI description, and the route kept alive until the sunset date.
- Examples in the OpenAPI document for every operation. A public API without examples is one every integrator reverse-engineers from error messages.

## Review

Changes to the public surface get the same review as a pricing change: someone outside the immediate team, a check that the OpenAPI diff is additive, and a note in the changelog that integrators will read. Regeneration does not save external clients, so the bar is higher here than for Connect.
