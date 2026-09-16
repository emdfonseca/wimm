package config_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
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
