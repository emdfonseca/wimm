# Errors

## Code → HTTP mapping

Connect applies this table for its own protocol; the REST layer applies the same table so both surfaces agree.

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

## Map once, at the boundary

```go
// apps/<service>/internal/<transport>/errors.go — the only file that knows both worlds
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

Domain code returns typed or sentinel Go errors and knows nothing about codes. Anything unmapped becomes `internal` with a generic message; the real error goes to the log with the trace ID.

## Details

Field violations go in `errdetails.BadRequest` via `connect.NewErrorDetail(...)` and `.WithDetails(detail)`. Use the standard `google.rpc` detail types (`BadRequest`, `PreconditionFailure`, `RetryInfo`, `ErrorInfo`) before inventing one; a custom detail message lives in the domain's proto package.

## Never crosses the boundary

Stack traces, SQL, file paths, upstream hostnames, internal IDs outside the contract, and the existence of resources the caller may not see (an unauthenticated caller gets `unauthenticated`, not `not_found`).
