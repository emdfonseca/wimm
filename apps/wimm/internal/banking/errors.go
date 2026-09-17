package banking

import (
	"errors"
	"fmt"
	"time"
)

// wimm's failure taxonomy. Adapters map their gateway's status codes onto these
// and nothing above this package ever sees an HTTP status or a provider's error
// string — no two gateways agree on them, and the surface a member is shown is
// routed off the kind alone (ADR 0018).
//
// Six kinds, because six different things are shown to a member. A seventh
// sentinel that routes to the same screen as an existing one is not a kind.
var (
	// ErrBankUnavailable is the bank's own systems failing or refusing. The
	// member's recourse is to try later; nothing about wimm is wrong.
	ErrBankUnavailable = errors.New("the bank is not answering")

	// ErrGatewayUnavailable is the service wimm reaches banks through
	// failing. Distinct from the bank being down because the remedy differs:
	// one bank or all of them.
	ErrGatewayUnavailable = errors.New("the service wimm reaches banks through is not answering")

	// ErrConsentDeclined is the member saying no at their bank, which is not a
	// failure at all. The connection is left exactly as it was.
	ErrConsentDeclined = errors.New("consent was declined at the bank")

	// ErrConsentExpired is access having run out, whether at the date wimm
	// set or early. Both are treated identically: the member's experience of
	// "my bank stopped updating" is the same either way.
	ErrConsentExpired = errors.New("access to the bank has run out")

	// ErrNoAccounts is access granted that exposes nothing. Rare and real:
	// it means a completed hand-off with nothing to show for it.
	ErrNoAccounts = errors.New("the bank granted access but exposed no accounts")

	// ErrRateLimited is being told to ask less often. Match with errors.Is;
	// reach the retry-after with errors.As and RateLimitError.
	ErrRateLimited = errors.New("the bank or the service reaching it has been asked too often")
)

// RateLimitError carries the one piece of data a caller must act on: when the
// member can be told to try again. It is a type rather than a sentinel because
// "try again in 90 seconds" and "try again later" are different things to show.
type RateLimitError struct {
	// RetryAfter is how long to wait. Zero means the gateway refused without
	// saying, which HasRetryAfter distinguishes so a zero is never rendered as
	// an instruction.
	RetryAfter time.Duration
}

// RateLimited builds a rate-limit failure. Pass zero when the gateway gave no
// retry-after.
func RateLimited(retryAfter time.Duration) error {
	return &RateLimitError{RetryAfter: retryAfter}
}

func (e *RateLimitError) Error() string {
	if !e.HasRetryAfter() {
		return ErrRateLimited.Error()
	}
	return fmt.Sprintf("%s; retry after %s", ErrRateLimited, e.RetryAfter)
}

// Unwrap is what makes errors.Is(err, ErrRateLimited) work on this type, so a
// caller that only cares which kind it is never has to know the type exists.
func (e *RateLimitError) Unwrap() error { return ErrRateLimited }

// HasRetryAfter reports whether the gateway said when to try again. A zero
// duration is the absence of an answer, not an instruction to retry now.
func (e *RateLimitError) HasRetryAfter() bool { return e.RetryAfter > 0 }
