//go:build integration

// These run against the real gateway and are not part of `just check`: they
// need credentials, they reach the network, and a suite that cannot run
// offline is a suite that stops being run. Build with the tag and supply the
// same configuration wimmd takes:
//
//	WIMM_ENABLEBANKING_APPLICATION_ID=... \
//	WIMM_ENABLEBANKING_PRIVATE_KEY=/path/to/key.pem \
//	go test -tags=integration -run TestAgainstTheSandbox ./internal/banking/enablebanking/
//
// What they are for is the half of an adapter unit tests cannot reach: that the
// token this code mints is one the gateway actually accepts, and that the field
// names in the response are the ones this code reads. Both are assumptions
// taken from documentation everywhere else in this package.
package enablebanking

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

func sandboxClient(t *testing.T) *Client {
	t.Helper()

	id, keyPath := os.Getenv("WIMM_ENABLEBANKING_APPLICATION_ID"), os.Getenv("WIMM_ENABLEBANKING_PRIVATE_KEY")
	if id == "" || keyPath == "" {
		t.Skip("set WIMM_ENABLEBANKING_APPLICATION_ID and WIMM_ENABLEBANKING_PRIVATE_KEY to run this")
	}

	c, err := New(Options{
		ApplicationID:  id,
		PrivateKeyPath: keyPath,
		RedirectURL: cmpOr(os.Getenv("WIMM_ENABLEBANKING_REDIRECT_URL"),
			"https://localhost:8765/psd2/callback"),
		BaseURL: os.Getenv("WIMM_ENABLEBANKING_BASE_URL"),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// 4.2 against the real thing: the token is accepted, the response parses, and
// the consent maximums are the ones this household's banks actually grant.
func TestAgainstTheSandboxBanks(t *testing.T) {
	c := sandboxClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	banks, err := c.Banks(ctx, "PT")
	if err != nil {
		t.Fatalf("Banks: %v", err)
	}
	if len(banks) == 0 {
		t.Fatal("no banks in PT, which means either the request was rejected or the response shape changed")
	}

	for _, b := range banks {
		if b.Name == "" {
			t.Error("a bank came back with no name")
		}
		if b.MaxConsent <= 0 {
			t.Errorf("%s has no consent maximum, so the consent screen has no date to state", b.Name)
		}
		if b.ID != bankID(b.Name, b.Country) {
			t.Errorf("%s has id %q, which will not round-trip into a hand-off", b.Name, b.ID)
		}
	}

	// The finding that shaped the design: consent length varies by two orders
	// of magnitude inside one country. If this stops being true, the emphasis
	// on restoring can be revisited.
	t.Logf("%d banks in PT", len(banks))
	for _, b := range banks {
		if want := map[string]bool{"Activo Bank": true, "Caixa Económica Montepio Geral": true, "Revolut": true}[b.Name]; want {
			t.Logf("%-24s consent %s", b.Name, b.MaxConsent)
		}
	}
}

// 4.5 against the real thing: a read with PSU headers reaches the gateway and
// is refused for a reason wimm recognises rather than an unmapped one. It
// cannot assert a balance without a live consent, so it asserts the failure is
// in the taxonomy — an unmapped error here is the defect 4.6 exists to prevent.
func TestAgainstTheSandboxBalancesWithoutASession(t *testing.T) {
	c := sandboxClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.Balances(ctx,
		banking.Connection{GatewayRef: "00000000-0000-0000-0000-000000000000"},
		banking.Account{Ref: "none", GatewayUID: "00000000-0000-0000-0000-000000000000"})
	if err == nil {
		t.Fatal("reading a balance with no session succeeded")
	}

	for _, known := range []error{
		banking.ErrConsentExpired,
		banking.ErrBankUnavailable,
		banking.ErrGatewayUnavailable,
		banking.ErrRateLimited,
		banking.ErrNoAccounts,
		banking.ErrConsentDeclined,
	} {
		if errors.Is(err, known) {
			t.Logf("mapped to %v", known)
			return
		}
	}
	t.Errorf("the gateway returned something outside wimm's taxonomy: %v", err)
}
