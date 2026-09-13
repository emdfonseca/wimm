# Pagination, idempotency, deadlines

## Pagination

Cursor only, never offset.

```text
request   page_size (int32; server clamps to [1, max]; 0 → default)
          page_token (opaque string; empty → first page)
response  next_page_token (empty → no more)
```

The token is versioned so a token from an old release fails with `invalid_argument`. Sort order is stable and documented. Filters are explicit request fields.

## Idempotency

`idempotency_key` on the request message; `Idempotency-Key` header on REST.

- First call stores `(caller, key, request hash, response)`.
- Same key, same request → stored response, no re-execution.
- Same key, different request → `invalid_argument`.
- Keys expire after 24h.

Implemented once as an interceptor, not per handler. A naturally idempotent RPC (upsert by natural key) skips the key and says so in its contract comment.

## Deadlines

Every RPC records its deadline in the contract comment:

```protobuf
// Deadline: 2s typical, 10s max
rpc CreateInvoice(...) returns (...);
```

A `MaxDeadline` interceptor enforces the max; `ctx` flows to every downstream call; overrun returns `deadline_exceeded`.
