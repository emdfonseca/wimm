// Package rpc holds the Connect handlers and the single point where a domain
// error becomes a Connect code. Nothing internal crosses that boundary.
package rpc

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"strings"

	"connectrpc.com/connect"
)

// OperatorCredentialHeader is where wimmctl puts the credential.
const OperatorCredentialHeader = "Authorization"

const operatorScheme = "Bearer "

// ErrOperatorUnauthenticated is what a wrong or absent credential produces.
// One error for both: telling the two apart is a distinction that only helps
// whoever is guessing.
var ErrOperatorUnauthenticated = errors.New("operator credential missing or incorrect")

// OperatorAuth refuses every call that does not carry the configured operator
// credential.
//
// The comparison is constant time over a hash of each side, so neither the
// credential's length nor a shared prefix is observable from how long a
// refusal takes. It runs before validation, so an unauthenticated caller
// cannot probe the schema (api-contract: interceptor order).
func OperatorAuth(credential string) connect.UnaryInterceptorFunc {
	want := sha256.Sum256([]byte(credential))

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			got := sha256.Sum256([]byte(bearer(req.Header().Get(OperatorCredentialHeader))))

			// An empty configured credential must never match an empty header:
			// the process refuses to start in that state, and this is the
			// second line under it.
			if credential == "" || subtle.ConstantTimeCompare(want[:], got[:]) != 1 {
				return nil, connect.NewError(connect.CodeUnauthenticated, ErrOperatorUnauthenticated)
			}
			return next(ctx, req)
		}
	}
}

func bearer(header string) string {
	if len(header) >= len(operatorScheme) && strings.EqualFold(header[:len(operatorScheme)], operatorScheme) {
		return header[len(operatorScheme):]
	}
	return header
}
