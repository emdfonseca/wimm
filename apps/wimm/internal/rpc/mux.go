package rpc

import (
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1/identityv1connect"
)

// OperatorHandler returns the operator service behind the operator credential.
//
// It is a separate handler on a separate listener rather than a guarded method
// on the public service: the distinction that matters is which port answers at
// all, and a method-level check is one forgotten annotation away from being
// public (ADR 0016).
func OperatorHandler(log *slog.Logger, srv identityv1connect.OperatorServiceHandler, credential string) (string, http.Handler) {
	return identityv1connect.NewOperatorServiceHandler(srv,
		// Auth before anything that touches the message, so an unauthenticated
		// caller cannot probe the schema (api-contract: interceptor order).
		connect.WithInterceptors(AccessLog(log), OperatorAuth(credential)),
	)
}

// PublicHandler returns the service the browser reaches.
func PublicHandler(log *slog.Logger, srv identityv1connect.PublicServiceHandler) (string, http.Handler) {
	return identityv1connect.NewPublicServiceHandler(srv,
		connect.WithInterceptors(AccessLog(log)),
	)
}
