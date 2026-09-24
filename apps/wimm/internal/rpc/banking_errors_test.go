package rpc

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// Every banking condition the web client routes a surface off must arrive as a
// code it can branch on.
//
// This exists because it did not. The whole taxonomy tasks 3.2 and 4.6 built
// stopped at this function: nothing mapped it, everything fell through to
// Internal with a fixed message, and the reason was discarded on the way out.
// A member saw "something went wrong" whatever had happened, and the server log
// said `code: internal` and nothing else.
func TestEveryBankingFailureArrivesAsSomethingTheClientCanRouteOn(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want connect.Code
	}{
		{"not an owner", banking.ErrNotOwner, connect.CodePermissionDenied},
		{"declined at the bank", banking.ErrConsentDeclined, connect.CodePermissionDenied},
		{"access has run out", banking.ErrConsentExpired, connect.CodeFailedPrecondition},
		{"asked too often", banking.ErrRateLimited, connect.CodeResourceExhausted},
		{"the bank is down", banking.ErrBankUnavailable, connect.CodeUnavailable},
		{"the gateway is down", banking.ErrGatewayUnavailable, connect.CodeUnavailable},
		{"no accounts exposed", banking.ErrNoAccounts, connect.CodeFailedPrecondition},
		{"no such bank", banking.ErrBankNotFound, connect.CodeNotFound},
		{"a return already used", store.ErrPendingConnectionSpent, connect.CodeAlreadyExists},
		{"an owner given a level", store.ErrOwnerHoldsAGrant, connect.CodeInvalidArgument},
		{"a row that is not there", store.ErrNotFound, connect.CodeNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Wrapped the way a handler wraps it, because that is how it
			// actually arrives.
			wrapped := fmt.Errorf("completing a connection: %w", tc.err)

			got := connect.CodeOf(toConnectError(wrapped))
			if got != tc.want {
				t.Errorf("code = %s, want %s", got, tc.want)
			}
			if got == connect.CodeInternal {
				t.Error("an unmapped failure: the client cannot tell it from a bug, and the reason is discarded")
			}
		})
	}
}

// A rate limit carries a retry-after the member is shown. Mapping it must not
// flatten the typed error into a bare sentinel.
func TestARateLimitKeepsItsRetryAfterThroughMapping(t *testing.T) {
	mapped := toConnectError(fmt.Errorf("reading balances: %w", banking.RateLimited(6*time.Hour)))

	if code := connect.CodeOf(mapped); code != connect.CodeResourceExhausted {
		t.Fatalf("code = %s", code)
	}

	var limited *banking.RateLimitError
	if !errors.As(mapped, &limited) {
		t.Fatal("the retry-after did not survive mapping")
	}
	if limited.RetryAfter != 6*time.Hour {
		t.Errorf("RetryAfter = %s, want 6h", limited.RetryAfter)
	}
}

// Anything genuinely unrecognised still says nothing, because an unexpected
// error must not describe itself to a caller.
func TestAnUnrecognisedErrorStaysOpaque(t *testing.T) {
	mapped := toConnectError(errors.New("a column called password_plaintext does not exist"))

	if code := connect.CodeOf(mapped); code != connect.CodeInternal {
		t.Errorf("code = %s, want Internal", code)
	}
	if got := mapped.Error(); got != "internal: something went wrong" {
		t.Errorf("message = %q, want the fixed one", got)
	}
}

// The wire stays opaque and the log does not.
//
// The first attempt at this logged err.Error() from the interceptor, which runs
// outside the handler and therefore sees the already-replaced error: the log
// line read `"error":"internal: something went wrong"` and the cause was gone.
// This asserts both halves so that cannot recur.
func TestAnUnrecognisedErrorIsOpaqueOnTheWireAndNamedInTheLog(t *testing.T) {
	cause := errors.New("dial tcp 10.0.0.7:5432: connection refused")

	mapped := toConnectError(fmt.Errorf("completing a connection: %w", cause))

	// What a caller sees.
	if got := mapped.Error(); got != "internal: something went wrong" {
		t.Errorf("the wire message is %q, want the fixed one", got)
	}
	if strings.Contains(mapped.Error(), "10.0.0.7") {
		t.Error("the wire message describes the failure to the caller")
	}

	// What the log can reach.
	var failure *internalFailure
	if !errors.As(mapped, &failure) {
		t.Fatal("the cause is not reachable, so a log line cannot name it")
	}
	if !errors.Is(failure.cause, cause) {
		t.Errorf("cause = %v, want the original", failure.cause)
	}
}

// A mapped failure carries its own message, because it is wimm's own wording
// and a caller acts on it.
func TestAMappedFailureKeepsItsMessage(t *testing.T) {
	mapped := toConnectError(fmt.Errorf("reading balances: %w", banking.ErrConsentExpired))

	if !strings.Contains(mapped.Error(), "access to the bank has run out") {
		t.Errorf("message = %q, want the domain's own wording", mapped.Error())
	}

	var failure *internalFailure
	if errors.As(mapped, &failure) {
		t.Error("a recognised failure was wrapped as an unrecognised one")
	}
}
