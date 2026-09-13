# Pagination, idempotency, deadlines

The three behaviours every endpoint needs and every second implementation gets subtly wrong. They are rules here so they stop being decisions.

## Pagination

Cursor-based, always. Offset pagination skips or duplicates rows whenever the underlying data changes between pages, which on a live system is always.

```text
request   page_size (int32, server clamps to [1, max]; 0 → default)
          page_token (opaque string; empty → first page)
response  next_page_token (empty → no more)
```

The token is opaque to the client and encodes whatever the server needs — typically the last row's sort key plus the filter, signed or at least versioned so a token from an old release fails cleanly rather than returning wrong rows. Sort order is stable and documented; the cursor is meaningless otherwise.

Filters are explicit request fields, not a free-text query language, unless the product genuinely needs search — in which case that is its own endpoint with its own contract.

## Idempotency

Any mutation a client might reasonably retry — create, charge, send, submit — takes an idempotency key: `idempotency_key` on the request message, `Idempotency-Key` header on REST. Semantics:

- First call with a key executes and stores `(key, request hash, response)`.
- Replay with the same key and same request returns the stored response and status, without re-executing.
- Same key, different request → `invalid_argument` (the client is confused, tell it).
- Keys expire after a documented window (24h is conventional); the store is keyed per caller so tenants cannot collide.

Idempotency is implemented once as middleware or a small package, not per handler. A handler that is "naturally idempotent" (`PUT` semantics, upsert by natural key) can skip the key and say so in its contract comment.

## Deadlines

- Clients set a timeout on every call. There is no such thing as a call with no deadline; there are only calls whose deadline is "whenever the socket dies".
- Servers enforce a per-RPC maximum (a `MaxDeadline` interceptor) so a client cannot ask for an hour.
- `context.Context` flows to every downstream call — database, other services, queues. A handler that starts work and ignores cancellation keeps consuming resources for a caller that has already given up.
- Exceeding the deadline returns `deadline_exceeded`, and the work is either rolled back or made idempotent so the retry is safe. An operation that may have partially completed with unknown result is the scenario the idempotency key exists for.

Timeouts are a per-endpoint decision recorded in the contract comment (`// Deadline: 2s typical, 10s max`), not a global constant discovered in an incident.
