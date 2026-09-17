package banking_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/banking/bankingtest"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// Task 10.1, as a committed test rather than an inspection.
//
// What could reach a bank is the gateway's session handle and each account's
// per-session identifier. A log line carrying either hands a reader this
// household's accounts for as long as the grant lasts, and the whole reason
// those values are sealed is that a copy of the data must not be enough
// (ADR 0018).
//
// Read at its most verbose, and across the whole flow: connecting, reading,
// restoring, failing and disconnecting. An inspection proves the code as it was
// read; this proves it on every run.

func TestNoLogLineCarriesAnythingThatCouldReachABank(t *testing.T) {
	var logged bytes.Buffer

	keys, err := banking.NewKeyring([]banking.Key{{ID: "k1", Material: testKey(5)}})
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	gw := bankingtest.New()
	st := newMemStore()

	// Debug: the most verbose level anything in this change writes at.
	svc := banking.NewService(st, gw, keys,
		slog.New(slog.NewJSONHandler(&logged, &slog.HandlerOptions{Level: slog.LevelDebug})),
		"https://localhost:8765/psd2/callback", 24*time.Hour)

	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"})
	for _, ref := range []string{"hash-1", "hash-2"} {
		gw.SetBalance("PT:Montepio", ref, banking.Balance{
			Money: banking.Money{Minor: 1000, Currency: "EUR"}, Kind: "CLAV",
		})
	}

	ctx := context.Background()

	// The whole flow, including the paths that only run when something fails.
	if _, err := svc.BeginConnection(ctx, ada, montepio().ID); err != nil {
		t.Fatalf("BeginConnection: %v", err)
	}
	done, err := svc.CompleteConnection(ctx, ada, banking.Callback{Code: "code", State: gw.LastState})
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}
	if _, err := svc.Accounts(ctx, ada, false); err != nil {
		t.Fatalf("Accounts: %v", err)
	}

	// A failure, so the error paths write whatever they write.
	gw.Fail("Balances", banking.RateLimited(6*time.Hour))
	if _, err := svc.Accounts(ctx, ada, false); err != nil {
		t.Fatalf("Accounts after a refusal: %v", err)
	}
	gw.Fail("Balances", banking.ErrConsentExpired)
	if _, err := svc.Accounts(ctx, ada, false); err != nil {
		t.Fatalf("Accounts after an expiry: %v", err)
	}

	// Restoring, which reseals every identifier.
	if _, err := svc.RestoreConnection(ctx, ada, done.Connection.ID); err != nil {
		t.Fatalf("RestoreConnection: %v", err)
	}
	if _, err := svc.CompleteConnection(ctx, ada, banking.Callback{Code: "code", State: gw.LastState}); err != nil {
		t.Fatalf("CompleteConnection on restore: %v", err)
	}
	if err := svc.Disconnect(ctx, ada, done.Connection.ID); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}

	// Everything the fake gateway ever handed out, in every encoding a value
	// could plausibly reach a log in.
	live := gw.IssuedSecrets()
	if len(live) == 0 {
		t.Fatal("the fixture produced no secrets, so this test asserts nothing")
	}

	output := logged.String()
	for _, secret := range live {
		for encoding, needle := range map[string]string{
			"raw":    secret,
			"hex":    hex.EncodeToString([]byte(secret)),
			"base64": base64.StdEncoding.EncodeToString([]byte(secret)),
		} {
			if needle == "" {
				continue
			}
			if strings.Contains(output, needle) {
				t.Errorf("a log line carries %s as %s:\n%s", secret, encoding, output)
			}
		}
	}

	// And the sealed values themselves, as ciphertext: harmless alone, and
	// pointless to log beside the row id they are bound to.
	for _, ciphertext := range st.everyCiphertext() {
		if len(ciphertext) == 0 {
			continue
		}
		for _, needle := range []string{
			string(ciphertext),
			hex.EncodeToString(ciphertext),
			base64.StdEncoding.EncodeToString(ciphertext),
		} {
			if strings.Contains(output, needle) {
				t.Errorf("a log line carries a sealed value:\n%s", output)
			}
		}
	}

	// Today this flow logs almost nothing, so the output checked is often
	// empty — which would make this test pass for the wrong reason if it were
	// the only evidence. It is not: reverting any redaction, or adding a
	// `"gateway_ref", gatewayRef` attribute to one log call, turns it red. Its
	// job is to stay red the moment someone adds logging that carries one.
	t.Logf("checked %d secrets against %d bytes of debug-level output", len(live), len(output))
}

// A sealed value inside a store row must not render either, however a row is
// formatted on its way into a log.
func TestAStoredConnectionDoesNotRenderItsSecret(t *testing.T) {
	keys, err := banking.NewKeyring([]banking.Key{{ID: "k1", Material: testKey(5)}})
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	sealed, err := keys.Seal([]byte("gateway-session-abc123"), "row-1")
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// The store's Sealed is two columns, not banking.Sealed: it deliberately
	// has no redaction of its own, because it holds ciphertext rather than a
	// value. What must never appear is the plaintext.
	row := store.BankConnection{
		ID:         "row-1",
		Gateway:    "bankingtest",
		GatewayRef: store.Sealed{Ciphertext: sealed.Ciphertext(), KeyID: sealed.KeyID()},
		BankName:   "Montepio",
	}

	var logged bytes.Buffer
	slog.New(slog.NewJSONHandler(&logged, nil)).Info("connected", "connection", row)

	if strings.Contains(logged.String(), "gateway-session-abc123") {
		t.Errorf("the plaintext reached a log through a stored row:\n%s", logged.String())
	}
}
