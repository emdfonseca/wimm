package rpc

import (
	"context"
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
				log.InfoContext(ctx, "call refused", append(attrs, "code", connect.CodeOf(err).String())...)
			} else {
				log.InfoContext(ctx, "call served", attrs...)
			}
			return res, err
		}
	}
}
