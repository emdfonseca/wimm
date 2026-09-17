package banking_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// The taxonomy exists so a caller can route a page banner, an inline alert or a
// field error off what kind of failure it is, without knowing a gateway's
// status codes. So the property under test is exactly that: each kind is
// distinguishable from every other by a caller holding only this package.
func TestEveryFailureIsDistinguishable(t *testing.T) {
	kinds := []error{
		banking.ErrBankUnavailable,
		banking.ErrGatewayUnavailable,
		banking.ErrConsentDeclined,
		banking.ErrConsentExpired,
		banking.ErrNoAccounts,
		banking.ErrRateLimited,
	}

	for i, sentinel := range kinds {
		t.Run(sentinel.Error(), func(t *testing.T) {
			// Wrapped the way an adapter would wrap it, to prove errors.Is
			// survives the context an adapter adds.
			wrapped := fmt.Errorf("reading balances at Montepio: %w", sentinel)

			if !errors.Is(wrapped, sentinel) {
				t.Fatal("a wrapped failure no longer matches its own kind")
			}
			for j, other := range kinds {
				if i == j {
					continue
				}
				if errors.Is(wrapped, other) {
					t.Errorf("%v also matches %v, so a caller cannot tell them apart", sentinel, other)
				}
			}
		})
	}
}

// Rate limiting is the one failure carrying data the caller must act on: the
// member is told when it can next be tried, and the balances already on screen
// stay where they are.
func TestRateLimitedCarriesItsRetryAfter(t *testing.T) {
	err := banking.RateLimited(90 * time.Second)

	if !errors.Is(err, banking.ErrRateLimited) {
		t.Fatal("a rate-limit error does not match ErrRateLimited")
	}

	var limited *banking.RateLimitError
	if !errors.As(err, &limited) {
		t.Fatal("errors.As cannot reach the retry-after")
	}
	if limited.RetryAfter != 90*time.Second {
		t.Errorf("RetryAfter = %s, want 1m30s", limited.RetryAfter)
	}
}

// A gateway that refuses without saying when is still a rate limit. The caller
// must be able to tell "try again in 90 seconds" from "try again later"
// without a zero masquerading as an instruction.
func TestRateLimitedWithNoRetryAfter(t *testing.T) {
	var limited *banking.RateLimitError
	if !errors.As(banking.RateLimited(0), &limited) {
		t.Fatal("errors.As failed")
	}
	if limited.HasRetryAfter() {
		t.Error("a zero retry-after is reported as an instruction to the member")
	}
}

// Wrapping must survive an adapter adding its own context, or the mapping in
// 4.6 would have to be done at the call site instead.
func TestAdapterContextSurvivesWrapping(t *testing.T) {
	err := fmt.Errorf("enablebanking: POST /sessions: %w", banking.ErrConsentExpired)

	if !errors.Is(err, banking.ErrConsentExpired) {
		t.Fatal("the kind was lost")
	}
	if errors.Is(err, banking.ErrConsentDeclined) {
		t.Fatal("expired and declined are not distinguishable")
	}
}

// Declined and expired are the pair most likely to be collapsed by mistake, and
// they route to different screens: one offers the flow again, the other says
// the member said no.
func TestDeclinedAndExpiredAreNotTheSameThing(t *testing.T) {
	if errors.Is(banking.ErrConsentDeclined, banking.ErrConsentExpired) {
		t.Error("declining at the bank is being treated as consent running out")
	}
}
