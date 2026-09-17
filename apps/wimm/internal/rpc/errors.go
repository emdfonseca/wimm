package rpc

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/identity"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// toConnectError is the single place a domain error becomes a Connect code.
// Anything unrecognised becomes Internal with a fixed message: an unexpected
// error must not describe itself to a caller.
func toConnectError(err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, identity.ErrNameMissing):
		return connect.NewError(connect.CodeInvalidArgument, err)

	case errors.Is(err, identity.ErrEmailAlreadyRegistered):
		return connect.NewError(connect.CodeAlreadyExists, err)

	case errors.Is(err, identity.ErrMemberNotFound):
		return connect.NewError(connect.CodeNotFound, err)

	// Expired, spent, replaced and never-issued all arrive here as the same
	// error and leave as the same code, so nothing about the outcome can be
	// inferred from the response.
	case errors.Is(err, identity.ErrEnrolmentLinkUnusable):
		return connect.NewError(connect.CodePermissionDenied, err)

	case errors.Is(err, identity.ErrNotDiscoverable):
		return connect.NewError(connect.CodeFailedPrecondition, err)

	case errors.Is(err, identity.ErrPasskeyNotRecognised):
		return connect.NewError(connect.CodePermissionDenied, err)

	case errors.Is(err, identity.ErrNotSignedIn):
		return connect.NewError(connect.CodeUnauthenticated, err)

	// Banking. Every one of these is a condition the web client routes a
	// surface off, so falling through to Internal does not merely lose a
	// message — it makes the taxonomy tasks 3.2 and 4.6 built unreachable, and
	// the reason unloggable.

	case errors.Is(err, banking.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, err)

	// The member said no at their bank. Not a failure, and not retryable.
	case errors.Is(err, banking.ErrConsentDeclined):
		return connect.NewError(connect.CodePermissionDenied, err)

	// Access has run out. The remedy is restoring, which is the whole flow
	// again, so it must be distinguishable from a bank having a bad minute.
	case errors.Is(err, banking.ErrConsentExpired):
		return connect.NewError(connect.CodeFailedPrecondition, err)

	case errors.Is(err, banking.ErrRateLimited):
		return connect.NewError(connect.CodeResourceExhausted, err)

	case errors.Is(err, banking.ErrBankUnavailable),
		errors.Is(err, banking.ErrGatewayUnavailable):
		return connect.NewError(connect.CodeUnavailable, err)

	case errors.Is(err, banking.ErrNoAccounts):
		return connect.NewError(connect.CodeFailedPrecondition, err)

	case errors.Is(err, banking.ErrBankNotFound):
		return connect.NewError(connect.CodeNotFound, err)

	// A return that was already exchanged is not a return that never existed,
	// and the two land the member in different places.
	case errors.Is(err, store.ErrPendingConnectionSpent):
		return connect.NewError(connect.CodeAlreadyExists, err)

	case errors.Is(err, store.ErrOwnerHoldsAGrant):
		return connect.NewError(connect.CodeInvalidArgument, err)

	case errors.Is(err, store.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)

	default:
		// The caller is told nothing, and the cause is kept for the log.
		return connect.NewError(connect.CodeInternal, &internalFailure{cause: err})
	}
}

// internalFailure is an unrecognised error on its way out.
//
// Its message is fixed, because an unexpected error must not describe itself to
// a caller — that message is what reaches the wire. Its cause is reachable with
// errors.As, which is how AccessLog records what actually happened.
//
// Replacing the error outright is what hid two bugs for a whole session: the
// response said "something went wrong" and so did the log, because the
// interceptor runs outside the handler and only ever sees what this returns.
type internalFailure struct {
	cause error
}

func (f *internalFailure) Error() string { return "something went wrong" }

// Unwrap keeps errors.Is and errors.As working through the replacement, so a
// caller inside the process can still ask what it was.
func (f *internalFailure) Unwrap() error { return f.cause }
