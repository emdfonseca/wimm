---
name: api-contract
description: The shape every service API in this repo follows - Connect/gRPC (protobuf) for service-to-service and first-party clients, REST with OpenAPI only for the public surface, one error vocabulary across both, cursor pagination, idempotency keys, deadline propagation, and additive-only versioning enforced by buf breaking. Use this whenever someone writes or changes a .proto file, an OpenAPI document, a Connect or HTTP handler, a generated client, or an endpoint's request/response/error shape; asks how a service should report errors, page results, time out, or handle retries; decides whether something should be REST or RPC; or adds a new endpoint or service. Reach for it even for a "quick internal endpoint" - the internal ones are where the inconsistencies start.
paths:
  - "**/*.proto"
  - "**/openapi.{yaml,yml,json}"
  - "**/buf.yaml"
  - "**/buf.gen.yaml"
  - "**/packages/contracts/**"
---

# API contract

**Default to Connect over protobuf. REST is a product decision for third parties, not a convenience for us.** Every service speaks the same error vocabulary, pages the same way, and evolves additively — so a client written against one service already knows how to talk to the next.

```text
service ↔ service         Connect (gRPC-compatible), protobuf
first-party web / CLI     Connect via connect-es / connect-go client
third parties, webhooks   REST + OpenAPI 3.1, JSON, RFC 9457 errors
```

Two protocols are already one more than ideal. A third — ad-hoc JSON endpoints that are neither — is the thing this skill exists to prevent.

## Start here, every time

1. **Who calls it?** Only us → Connect. External developers → REST, and the REST surface is a deliberate, documented product with its own review. Never "REST because it was quicker".
2. **Does the contract exist before the code?** The `.proto` or OpenAPI document is written, reviewed, and generated from first. Handlers implement generated interfaces; they do not define the shape.
3. **Is it additive?** Within a major version, fields are added, never removed, renamed, or retyped. `buf breaking` in CI enforces it for protobuf; the same discipline is manual for OpenAPI.

## Where to read next

| Read this | When |
|---|---|
| `references/connect.md` | Writing services, RPCs, and handlers; proto package layout; buf config; connect-es on the client. |
| `references/rest-public.md` | Designing or changing the public REST surface; OpenAPI conventions; RFC 9457 problem details. |
| `references/errors.md` | Mapping domain failures to codes; what goes in details; what must never leak. The one vocabulary both protocols share. |
| `references/pagination-idempotency-deadlines.md` | List endpoints, mutating endpoints that can be retried, and timeouts — the three things every endpoint gets wrong without a rule. |
| `references/versioning.md` | What counts as breaking, how majors are introduced, deprecation, `buf breaking`. |
| `references/new-endpoint-checklist.md` | Adding an RPC or route end to end. |

Where contracts live in the repo and how `just gen` produces clients is the `monorepo-standard` skill's business, not this one's.

## Rules that are constantly needed

### Contract first, generated second, handwritten last

```text
packages/contracts/proto/<org>/<domain>/v1/<domain>.proto     ← the source
packages/contracts/gen/{go,ts,py}/...                         ← just gen; never edited
apps/<service>/internal/<domain>/handler.go                   ← implements the generated interface
```

A handler that accepts a shape the contract does not describe is a contract change that skipped review. The generated-file hook in `monorepo-standard` blocks edits to `gen/`; the point is not bureaucracy, it is that the contract is the only place two languages can agree.

### One error vocabulary

Connect's codes are the vocabulary for both protocols — `invalid_argument`, `not_found`, `already_exists`, `permission_denied`, `unauthenticated`, `failed_precondition`, `resource_exhausted`, `unavailable`, `internal`. REST maps them to HTTP status and carries the code in the problem-details body. Domain errors are mapped to a code exactly once, at the handler boundary; everything below returns plain Go errors. Details are structured (`errdetails.BadRequest` for field violations), messages are for humans, and nothing internal — stack traces, SQL, upstream hostnames — crosses the boundary. `references/errors.md` has the mapping table.

### Naming

Proto packages are `<org>.<domain>.v1`; services are `<Domain>Service`; RPCs are `VerbNoun` (`CreateProject`, `ListProjects`); messages are `<Rpc>Request` / `<Rpc>Response`. Fields are `snake_case` in proto and appear as `lowerCamelCase` in JSON via protojson — do not fight that with `json_name`. REST paths are plural nouns, lowercase, kebab-case, under `/v1/`.

### Every list pages, every mutation can be retried, every call has a deadline

- Lists take `page_size` (bounded server-side) and `page_token`, return `next_page_token`. Offset pagination is not offered; it goes wrong on any table that changes while being read.
- Create/mutate RPCs that a client might retry accept an idempotency key and return the original result on replay. On REST it is the `Idempotency-Key` header.
- Servers enforce a maximum deadline and propagate `context.Context` to every downstream call. A handler that ignores the context is a handler that keeps working for a client that has gone.

### Breaking means a new major

Removing or renaming a field, changing a type, changing the meaning of an existing value, tightening validation — all breaking. The response is `v2` alongside `v1`, with a documented migration and a sunset date, never an in-place edit. Deprecate with `[deprecated = true]` and a comment saying what replaces it. `buf breaking --against` runs in CI against `main`, so this is a failing check rather than a review comment.

## Checkpoints

- **A `.proto` change** — run `buf lint` and `buf breaking` locally before the PR; if breaking is intended, the PR is a `v2` PR and says so.
- **A new endpoint** — walk `references/new-endpoint-checklist.md`. Error mapping, pagination, idempotency, deadline, auth, and telemetry attributes are the items people skip when the happy path works.
- **Anything on the public REST surface** — OpenAPI document updated in the same PR, examples included, and the change reviewed as a product change. External clients cannot be regenerated by `just gen`.
- **A handler doing validation the contract could express** — move it into the proto (`buf validate` constraints) or the OpenAPI schema. Validation in two places disagrees eventually.
