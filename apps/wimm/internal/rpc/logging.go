package rpc

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
)

// AccessLog records that a call happened and how it ended.
//
// It records the procedure and never the messages. An enrolment link and a
// session identifier both travel inside request and response bodies, and a log
// that prints bodies is a log that hands them to whoever reads it (ADR 0016).
func AccessLog(log *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, req)

			attrs := []any{"procedure", req.Spec().Procedure}
			if err != nil {
				code := connect.CodeOf(err)
				// Built fresh rather than appended onto attrs: appending to a
				// slice you did not create is how two log lines end up sharing
				// a backing array.
				refused := []any{"procedure", req.Spec().Procedure, "code", code.String()}

				// Every refusal names its condition, not only the unmapped
				// ones. Codes collide by design — consent having run out and a
				// bank exposing no accounts are both FailedPrecondition,
				// because a caller shows the same kind of surface for each —
				// so a log carrying the code alone cannot say which happened.
				//
				// This is wimm's own error text, never a request or response
				// body: an enrolment link and a session identifier both travel
				// inside those (ADR 0016), and none of this comes from one.
				//
				// An Internal code is the exception in the other direction:
				// its message is deliberately "something went wrong", so the
				// cause is read off the replacement instead. This interceptor
				// runs outside the handler, so err is already the opaque one.
				var failure *internalFailure
				if errors.As(err, &failure) && failure.cause != nil {
					refused = append(refused, "cause", failure.cause.Error())
				} else {
					refused = append(refused, "cause", err.Error())
				}

				log.InfoContext(ctx, "call refused", refused...)
			} else {
				log.InfoContext(ctx, "call served", attrs...)
			}
			return res, err
		}
	}
}
