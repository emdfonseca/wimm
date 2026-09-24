package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/config"
)

func env(pairs map[string]string) func(string) string {
	return func(k string) string { return pairs[k] }
}

func valid() map[string]string {
	return map[string]string{
		"WIMM_OPERATOR_CREDENTIAL": "s3cret",
		"WIMM_DATABASE_URL":        "postgres://localhost/wimm",
	}
}

// The refusal that matters: an operator surface reachable with nothing
// guarding it must stop the process, not open it.
func TestRefusesToStartWithOperatorListenerAndNoCredential(t *testing.T) {
	e := valid()
	delete(e, "WIMM_OPERATOR_CREDENTIAL")

	_, err := config.Load(env(e))
	if err == nil {
		t.Fatal("configuration with the operator listener enabled and no credential was accepted")
	}
	if !errors.Is(err, config.ErrMissingOperatorCredential) {
		t.Errorf("error is not ErrMissingOperatorCredential: %v", err)
	}
	if !strings.Contains(err.Error(), "WIMM_OPERATOR_CREDENTIAL") {
		t.Errorf("error does not name the missing credential: %v", err)
	}
}

func TestOperatorCredentialNotNeededWhenListenerIsOff(t *testing.T) {
	e := valid()
	delete(e, "WIMM_OPERATOR_CREDENTIAL")
	e["WIMM_OPERATOR_ENABLED"] = "false"

	c, err := config.Load(env(e))
	if err != nil {
		t.Fatalf("configuration with the listener off was refused: %v", err)
	}
	if c.OperatorEnabled {
		t.Error("OperatorEnabled is true with WIMM_OPERATOR_ENABLED=false")
	}
}

func TestDefaults(t *testing.T) {
	c, err := config.Load(env(valid()))
	if err != nil {
		t.Fatalf("valid configuration was refused: %v", err)
	}

	for _, tc := range []struct{ name, got, want string }{
		{"PublicAddr", c.PublicAddr, config.DefaultPublicAddr},
		{"OperatorAddr", c.OperatorAddr, config.DefaultOperatorAddr},
		{"RelyingPartyID", c.RelyingPartyID, config.DefaultRelyingPartyID},
		{"BaseURL", c.BaseURL, config.DefaultBaseURL},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
	if c.EnrolmentLinkLifetime != config.DefaultEnrolmentLinkLifetime {
		t.Errorf("EnrolmentLinkLifetime = %s, want %s", c.EnrolmentLinkLifetime, config.DefaultEnrolmentLinkLifetime)
	}
	// With no explicit origin list, the ceremony accepts exactly the origin
	// links are built against.
	if len(c.Origins) != 1 || c.Origins[0] != config.DefaultBaseURL {
		t.Errorf("Origins = %v, want [%s]", c.Origins, config.DefaultBaseURL)
	}
}

func TestReportsEveryProblemAtOnce(t *testing.T) {
	_, err := config.Load(env(map[string]string{
		"WIMM_RP_ID":                   "",
		"WIMM_ENROLMENT_LINK_LIFETIME": "yesterday",
	}))
	if err != nil {
		for _, want := range []string{"WIMM_DATABASE_URL", "WIMM_ENROLMENT_LINK_LIFETIME", "WIMM_OPERATOR_CREDENTIAL"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error does not mention %s: %v", want, err)
			}
		}
		return
	}
	t.Fatal("configuration with three problems was accepted")
}

func TestRejectsNonPositiveLifetimes(t *testing.T) {
	e := valid()
	e["WIMM_SESSION_LIFETIME"] = "0s"

	if _, err := config.Load(env(e)); err == nil {
		t.Fatal("a zero session lifetime was accepted")
	}
}

func TestParsesOverrides(t *testing.T) {
	e := valid()
	e["WIMM_RP_ID"] = "wimm.example"
	e["WIMM_ORIGINS"] = "https://wimm.example, https://www.wimm.example/"
	e["WIMM_SESSION_LIFETIME"] = "72h"

	c, err := config.Load(env(e))
	if err != nil {
		t.Fatalf("valid overrides were refused: %v", err)
	}
	if c.RelyingPartyID != "wimm.example" {
		t.Errorf("RelyingPartyID = %q", c.RelyingPartyID)
	}
	if len(c.Origins) != 2 || c.Origins[1] != "https://www.wimm.example" {
		t.Errorf("Origins = %v, want the trailing slash trimmed", c.Origins)
	}
	if c.SessionLifetime != 72*time.Hour {
		t.Errorf("SessionLifetime = %s", c.SessionLifetime)
	}
}

// The sweep runs on an instance nobody administers, so its two durations have
// to be right without being set. A default that is zero or missing turns the
// ticker into a busy loop or the retention window into an immediate delete.
func TestSweepDefaults(t *testing.T) {
	c, err := config.Load(env(valid()))
	if err != nil {
		t.Fatalf("valid configuration was refused: %v", err)
	}

	if c.SweepInterval != config.DefaultSweepInterval {
		t.Errorf("SweepInterval = %s, want %s", c.SweepInterval, config.DefaultSweepInterval)
	}
	if c.RevokedSessionRetention != config.DefaultRevokedSessionRetention {
		t.Errorf("RevokedSessionRetention = %s, want %s",
			c.RevokedSessionRetention, config.DefaultRevokedSessionRetention)
	}
}

func TestParsesSweepOverrides(t *testing.T) {
	e := valid()
	e["WIMM_SWEEP_INTERVAL"] = "15m"
	e["WIMM_REVOKED_SESSION_RETENTION"] = "48h"

	c, err := config.Load(env(e))
	if err != nil {
		t.Fatalf("valid overrides were refused: %v", err)
	}
	if c.SweepInterval != 15*time.Minute {
		t.Errorf("SweepInterval = %s, want 15m", c.SweepInterval)
	}
	if c.RevokedSessionRetention != 48*time.Hour {
		t.Errorf("RevokedSessionRetention = %s, want 48h", c.RevokedSessionRetention)
	}
}

// A zero interval is a ticker that panics and a retention window that deletes
// a session the instant it is revoked. Both are refusals at startup, the way
// every other lifetime already is.
func TestRejectsNonPositiveSweepDurations(t *testing.T) {
	for _, key := range []string{"WIMM_SWEEP_INTERVAL", "WIMM_REVOKED_SESSION_RETENTION"} {
		t.Run(key, func(t *testing.T) {
			e := valid()
			e[key] = "0s"

			_, err := config.Load(env(e))
			if err == nil {
				t.Fatalf("a zero %s was accepted", key)
			}
			if !strings.Contains(err.Error(), key) {
				t.Errorf("the refusal does not name %s: %v", key, err)
			}
		})
	}
}

// secretFile writes a fixture fit to hold key material and returns its path.
func secretFile(t *testing.T, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "key")
	if err := os.WriteFile(path, []byte("2026-01 AAAA\n"), mode); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	return path
}

// withBanking is a valid environment with the gateway switched on.
func withBanking(t *testing.T) map[string]string {
	t.Helper()
	e := valid()
	e["WIMM_BANKING_GATEWAY"] = config.GatewayEnableBanking
	e["WIMM_BANKING_ENCRYPTION_KEY"] = secretFile(t, 0o600)
	e["WIMM_ENABLEBANKING_PRIVATE_KEY"] = secretFile(t, 0o600)
	e["WIMM_ENABLEBANKING_APPLICATION_ID"] = "16560d8b-2dc5-4b4d-ac41-8266f62e719b"
	e["WIMM_ENABLEBANKING_REDIRECT_URL"] = "https://localhost:8765/psd2/callback"
	return e
}

// Unset is a supported state, and the one a fresh checkout is in. It must not
// be a warning or a degraded start — it is simply a wimm that cannot connect a
// bank yet.
func TestNoGatewayIsNotAProblem(t *testing.T) {
	c, err := config.Load(env(valid()))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.BankingGateway != "" {
		t.Errorf("BankingGateway = %q, want empty", c.BankingGateway)
	}
}

func TestAFullyConfiguredGatewayLoads(t *testing.T) {
	c, err := config.Load(env(withBanking(t)))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.BankingGateway != config.GatewayEnableBanking {
		t.Errorf("BankingGateway = %q", c.BankingGateway)
	}
	if c.BalanceStaleAfter != config.DefaultBalanceStaleAfter {
		t.Errorf("BalanceStaleAfter = %s, want the default", c.BalanceStaleAfter)
	}
}

// A gateway that starts without its sealing key writes plaintext bank
// credentials until someone notices. Every one of these is a refusal to start.
func TestRefusesToStartWithAHalfConfiguredGateway(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(t *testing.T, e map[string]string)
		want   string
	}{
		{
			"no sealing key",
			func(_ *testing.T, e map[string]string) { delete(e, "WIMM_BANKING_ENCRYPTION_KEY") },
			"WIMM_BANKING_ENCRYPTION_KEY is not set",
		},
		{
			"no signing key",
			func(_ *testing.T, e map[string]string) { delete(e, "WIMM_ENABLEBANKING_PRIVATE_KEY") },
			"WIMM_ENABLEBANKING_PRIVATE_KEY is not set",
		},
		{
			"a sealing key that does not exist",
			func(t *testing.T, e map[string]string) {
				e["WIMM_BANKING_ENCRYPTION_KEY"] = filepath.Join(t.TempDir(), "absent")
			},
			"does not exist",
		},
		{
			"a group-readable sealing key",
			func(t *testing.T, e map[string]string) {
				e["WIMM_BANKING_ENCRYPTION_KEY"] = secretFile(t, 0o640)
			},
			"readable beyond its owner",
		},
		{
			"a world-readable signing key",
			func(t *testing.T, e map[string]string) {
				e["WIMM_ENABLEBANKING_PRIVATE_KEY"] = secretFile(t, 0o644)
			},
			"readable beyond its owner",
		},
		{
			"an unreadable sealing key",
			func(t *testing.T, e map[string]string) {
				e["WIMM_BANKING_ENCRYPTION_KEY"] = secretFile(t, 0o200)
			},
			"cannot be opened",
		},
		{
			"no application id",
			func(_ *testing.T, e map[string]string) { delete(e, "WIMM_ENABLEBANKING_APPLICATION_ID") },
			"WIMM_ENABLEBANKING_APPLICATION_ID is not set",
		},
		{
			"no redirect url",
			func(_ *testing.T, e map[string]string) { delete(e, "WIMM_ENABLEBANKING_REDIRECT_URL") },
			"WIMM_ENABLEBANKING_REDIRECT_URL is not set",
		},
		{
			"a gateway nothing implements",
			func(_ *testing.T, e map[string]string) { e["WIMM_BANKING_GATEWAY"] = "plaid" },
			"not a gateway wimm has an adapter for",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := withBanking(t)
			tc.mutate(t, e)

			_, err := config.Load(env(e))
			if err == nil {
				t.Fatal("started")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not say %q", err, tc.want)
			}
		})
	}
}

// One restart should be enough to see everything that is wrong, the way the
// rest of Load already behaves.
func TestEveryBankingProblemIsReportedAtOnce(t *testing.T) {
	e := withBanking(t)
	delete(e, "WIMM_BANKING_ENCRYPTION_KEY")
	delete(e, "WIMM_ENABLEBANKING_APPLICATION_ID")
	e["WIMM_ENABLEBANKING_PRIVATE_KEY"] = secretFile(t, 0o644)

	_, err := config.Load(env(e))
	if err == nil {
		t.Fatal("started")
	}
	for _, want := range []string{
		"WIMM_BANKING_ENCRYPTION_KEY is not set",
		"WIMM_ENABLEBANKING_APPLICATION_ID is not set",
		"readable beyond its owner",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q: %v", want, err)
		}
	}
}

func TestConnectableBanksDefaultsWhenUnset(t *testing.T) {
	c, err := config.Load(env(valid()))
	if err != nil {
		t.Fatalf("valid configuration was refused: %v", err)
	}
	if !slices.Equal(c.ConnectableBanks, config.DefaultConnectableBanks) {
		t.Errorf("ConnectableBanks = %v, want %v", c.ConnectableBanks, config.DefaultConnectableBanks)
	}
}

func TestConnectableBanksParsesCustomList(t *testing.T) {
	e := valid()
	e["WIMM_BANKING_CONNECTABLE_BANKS"] = "Revolut, N26 ,  , Wise"

	c, err := config.Load(env(e))
	if err != nil {
		t.Fatalf("valid overrides were refused: %v", err)
	}
	want := []string{"Revolut", "N26", "Wise"}
	if !slices.Equal(c.ConnectableBanks, want) {
		t.Errorf("ConnectableBanks = %v, want %v", c.ConnectableBanks, want)
	}
}
