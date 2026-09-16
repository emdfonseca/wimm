package rpc

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/xuuid/wimm/apps/wimm/internal/identity"
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

	default:
		return connect.NewError(connect.CodeInternal, errors.New("something went wrong"))
	}
}
