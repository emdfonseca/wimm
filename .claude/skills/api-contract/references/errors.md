# Errors

One vocabulary, mapped once, nothing leaked. The vocabulary is Connect's code set; it is expressive enough for every case we have and it is what the generated clients already understand.

## The mapping table

| Situation | Connect code | HTTP (REST) |
|---|---|---|
| Request fails validation | `invalid_argument` | 400 |
| Caller not authenticated | `unauthenticated` | 401 |
| Caller lacks permission | `permission_denied` | 403 |
| Resource does not exist | `not_found` | 404 |
| Create conflicts with existing | `already_exists` | 409 |
| Precondition or state rule violated | `failed_precondition` | 412 |
| Concurrent modification lost | `aborted` | 409 |
| Rate or quota limit | `resource_exhausted` | 429 |
| Deadline passed | `deadline_exceeded` | 504 |
| Dependency unavailable / retry later | `unavailable` | 503 |
| Not implemented yet | `unimplemented` | 501 |
| Anything else | `internal` | 500 |

Connect performs this HTTP mapping for its own protocol; the REST layer applies the same table so both surfaces agree.

## Map once, at the boundary

```go
// apps/billing/internal/httpx/errors.go — the only file that knows both worlds
func toConnectError(err error) error {
	var nf billing.NotFoundError
	switch {
	case errors.As(err, &nf):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, billing.ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, err)
	}
	return connect.NewError(connect.CodeInternal, errors.New("internal error"))
}
```

Domain code returns typed or sentinel Go errors and knows nothing about codes. The handler package maps them. Anything unmapped becomes `internal` with a generic message — the real error goes to the log with the trace ID, never to the caller.

## Details

Structured detail for things a client can act on:

```go
detail, _ := connect.NewErrorDetail(&errdetails.BadRequest{
	FieldViolations: []*errdetails.BadRequest_FieldViolation{
		{Field: "customer_id", Description: "unknown customer"},
	},
})
return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("validation failed")).WithDetails(detail)
```

Use the standard `google.rpc` detail types (`BadRequest`, `PreconditionFailure`, `RetryInfo`, `ErrorInfo`) before inventing one. A custom detail message belongs in the domain's proto package and needs a reason.

## What never crosses the boundary

Stack traces, SQL, file paths, upstream hostnames, internal IDs that are not part of the contract, and the existence or non-existence of resources the caller may not know about (an unauthenticated caller asking about a private resource gets `unauthenticated`, not `not_found`). The message field is for the human reading the client's error; the log line is for us.

## Retryability is part of the contract

`unavailable`, `deadline_exceeded`, `resource_exhausted` (with `RetryInfo`), and `aborted` are retryable. `invalid_argument`, `not_found`, `permission_denied`, `failed_precondition`, and `internal` are not — an `internal` that a client retries in a loop is an outage amplifier. Clients honour this; servers pick codes knowing clients will.
