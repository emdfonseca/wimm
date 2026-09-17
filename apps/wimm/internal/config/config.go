// Package config reads everything that differs between a laptop and a
// deployment. Nothing here has a value compiled in that a deployment would
// want to change (ADR 0016).
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// Config is wimmd's startup configuration, already validated.
type Config struct {
	// PublicAddr is the address the browser reaches, through the web app's
	// /api proxy.
	PublicAddr string

	// OperatorEnabled turns the operator listener on. Off, wimmctl cannot
	// reach this instance at all.
	OperatorEnabled bool
	// OperatorAddr is that listener's address. Loopback by default.
	OperatorAddr string
	// OperatorCredential is the bearer credential wimmctl presents. Compared
	// in constant time, never logged.
	OperatorCredential string

	// RelyingPartyID is the WebAuthn relying-party identifier. Changing it
	// invalidates every passkey ever registered against it, with no migration
	// path, which is why it is configuration rather than a constant.
	RelyingPartyID string
	// RelyingPartyName is what the browser shows in its prompt.
	RelyingPartyName string
	// Origins are the origins a ceremony may arrive from.
	Origins []string

	// EnrolmentLinkLifetime is how long a link works for.
	EnrolmentLinkLifetime time.Duration
	// SessionLifetime is a session's absolute expiry.
	SessionLifetime time.Duration
	// CeremonyLifetime is how long a WebAuthn challenge stays valid.
	CeremonyLifetime time.Duration

	// SweepInterval is how often wimmd removes identity rows that can no
	// longer be used. Nothing depends on it being short: an expired row is
	// already refused by the check that reads it, so a slow sweep costs
	// storage and nothing else.
	SweepInterval time.Duration
	// RevokedSessionRetention is how long a revoked session is kept after
	// revocation, so that "was this session actually cut off" has an answer.
	// It is not an audit trail, and the default is short for that reason.
	RevokedSessionRetention time.Duration

	// BaseURL is the origin enrolment links are built against.
	BaseURL string

	// DatabaseURL names the Postgres instance wimmd owns.
	DatabaseURL string

	// BankingGateway names the open-banking adapter. Empty disables
	// connecting entirely: the accounts screen shows its empty state, which
	// is what lets this deploy before anyone holds gateway credentials.
	BankingGateway string
	// BankingEncryptionKeyPath is the file holding the keys that seal what
	// could reach a bank. A path, never the material: an environment
	// variable carrying a key reaches every child process and every crash
	// report (ADR 0018).
	BankingEncryptionKeyPath string
	// EnableBankingApplicationID is the registered application.
	EnableBankingApplicationID string
	// EnableBankingPrivateKeyPath is the file holding the request signing
	// key. A path, for the same reason.
	EnableBankingPrivateKeyPath string
	// EnableBankingRedirectURL must match one registered in the gateway's
	// control panel, or every hand-off is refused at the bank.
	EnableBankingRedirectURL string

	// BalanceStaleAfter is how old a reading may be before it is shown as
	// stale. It never hides a reading: a figure with its age is the contract.
	BalanceStaleAfter time.Duration
}

// ErrMissingOperatorCredential is the refusal to start with the operator
// surface reachable and nothing guarding it.
var ErrMissingOperatorCredential = errors.New(
	"WIMM_OPERATOR_CREDENTIAL is not set and the operator listener is enabled: " +
		"set the credential, or disable the listener with WIMM_OPERATOR_ENABLED=false")

// Defaults that a laptop can run on unchanged.
const (
	DefaultPublicAddr            = "127.0.0.1:9467"
	DefaultOperatorAddr          = "127.0.0.1:9468"
	DefaultRelyingPartyID        = "localhost"
	DefaultRelyingPartyName      = "wimm"
	DefaultBaseURL               = "http://localhost:9466"
	DefaultEnrolmentLinkLifetime = 24 * time.Hour
	DefaultSessionLifetime       = 14 * 24 * time.Hour
	DefaultCeremonyLifetime      = 5 * time.Minute
	DefaultBalanceStaleAfter     = 24 * time.Hour

	// GatewayEnableBanking is the only adapter that exists. The value is
	// stored on every connection, so a later gateway is new connections
	// rather than a reinterpretation of old rows.
	GatewayEnableBanking = "enablebanking"

	// Hourly is often enough that a household instance never notices the
	// backlog, and rare enough that the deletes never compete with a request.
	DefaultSweepInterval = time.Hour
	// A week covers the span in which anyone asks whether a device was cut
	// off. Past that the row answers a question nobody is still asking.
	DefaultRevokedSessionRetention = 7 * 24 * time.Hour
)

// Load reads the environment and validates it. It returns every problem it
// finds, not only the first, so one restart is enough to see them all.
func Load(env func(string) string) (Config, error) {
	c := Config{
		PublicAddr:            or(env("WIMM_PUBLIC_ADDR"), DefaultPublicAddr),
		OperatorAddr:          or(env("WIMM_OPERATOR_ADDR"), DefaultOperatorAddr),
		OperatorCredential:    env("WIMM_OPERATOR_CREDENTIAL"),
		RelyingPartyID:        or(env("WIMM_RP_ID"), DefaultRelyingPartyID),
		RelyingPartyName:      or(env("WIMM_RP_NAME"), DefaultRelyingPartyName),
		BaseURL:               strings.TrimRight(or(env("WIMM_BASE_URL"), DefaultBaseURL), "/"),
		DatabaseURL:           env("WIMM_DATABASE_URL"),
		EnrolmentLinkLifetime: DefaultEnrolmentLinkLifetime,
		SessionLifetime:       DefaultSessionLifetime,
		CeremonyLifetime:      DefaultCeremonyLifetime,

		SweepInterval:           DefaultSweepInterval,
		RevokedSessionRetention: DefaultRevokedSessionRetention,

		BankingGateway:              strings.TrimSpace(env("WIMM_BANKING_GATEWAY")),
		BankingEncryptionKeyPath:    env("WIMM_BANKING_ENCRYPTION_KEY"),
		EnableBankingApplicationID:  env("WIMM_ENABLEBANKING_APPLICATION_ID"),
		EnableBankingPrivateKeyPath: env("WIMM_ENABLEBANKING_PRIVATE_KEY"),
		EnableBankingRedirectURL:    env("WIMM_ENABLEBANKING_REDIRECT_URL"),
		BalanceStaleAfter:           DefaultBalanceStaleAfter,
	}

	var problems []error

	enabled, err := boolOr(env("WIMM_OPERATOR_ENABLED"), true)
	if err != nil {
		problems = append(problems, fmt.Errorf("WIMM_OPERATOR_ENABLED: %w", err))
	}
	c.OperatorEnabled = enabled

	for _, d := range []struct {
		key   string
		field *time.Duration
	}{
		{"WIMM_ENROLMENT_LINK_LIFETIME", &c.EnrolmentLinkLifetime},
		{"WIMM_SESSION_LIFETIME", &c.SessionLifetime},
		{"WIMM_CEREMONY_LIFETIME", &c.CeremonyLifetime},
		{"WIMM_SWEEP_INTERVAL", &c.SweepInterval},
		{"WIMM_REVOKED_SESSION_RETENTION", &c.RevokedSessionRetention},
		{"WIMM_BALANCE_STALE_AFTER", &c.BalanceStaleAfter},
	} {
		if raw := env(d.key); raw != "" {
			v, err := time.ParseDuration(raw)
			switch {
			case err != nil:
				problems = append(problems, fmt.Errorf("%s: %w", d.key, err))
			case v <= 0:
				problems = append(problems, fmt.Errorf("%s: must be positive, got %s", d.key, raw))
			default:
				*d.field = v
			}
		}
	}

	c.Origins = splitOrigins(env("WIMM_ORIGINS"))
	if len(c.Origins) == 0 {
		c.Origins = []string{c.BaseURL}
	}
	for _, o := range c.Origins {
		if _, err := url.Parse(o); err != nil {
			problems = append(problems, fmt.Errorf("WIMM_ORIGINS: %q is not a URL: %w", o, err))
		}
	}

	if c.RelyingPartyID == "" {
		problems = append(problems, errors.New("WIMM_RP_ID is empty"))
	}
	if c.DatabaseURL == "" {
		problems = append(problems, errors.New("WIMM_DATABASE_URL is not set"))
	}

	problems = append(problems, c.checkBanking()...)

	// The refusal the operator surface turns on. Last, so it reads first among
	// equals when several things are wrong.
	if c.OperatorEnabled && c.OperatorCredential == "" {
		problems = append(problems, ErrMissingOperatorCredential)
	}

	if len(problems) > 0 {
		return Config{}, errors.Join(problems...)
	}
	return c, nil
}

// LoadFromEnv reads the process environment.
func LoadFromEnv() (Config, error) { return Load(os.Getenv) }

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func boolOr(v string, fallback bool) (bool, error) {
	if v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%q is not a boolean", v)
	}
	return b, nil
}

func splitOrigins(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, strings.TrimRight(p, "/"))
		}
	}
	return out
}

// checkBanking refuses a half-configured gateway. Unset is a supported state
// and produces no problems at all; named, every credential it needs must be
// present and every key file must be fit to hold a key.
//
// A gateway that starts without its sealing key would write plaintext session
// identifiers into the database until someone noticed, which is the failure
// this exists to prevent — so it is a refusal to start rather than a warning.
func (c Config) checkBanking() []error {
	if c.BankingGateway == "" {
		return nil
	}

	var problems []error
	if c.BankingGateway != GatewayEnableBanking {
		problems = append(problems, fmt.Errorf(
			"WIMM_BANKING_GATEWAY: %q is not a gateway wimm has an adapter for (known: %s)",
			c.BankingGateway, GatewayEnableBanking))
	}

	for _, r := range []struct{ key, value string }{
		{"WIMM_ENABLEBANKING_APPLICATION_ID", c.EnableBankingApplicationID},
		{"WIMM_ENABLEBANKING_REDIRECT_URL", c.EnableBankingRedirectURL},
	} {
		if r.value == "" {
			problems = append(problems, fmt.Errorf("%s is not set, and %s is enabled", r.key, c.BankingGateway))
		}
	}
	if c.EnableBankingRedirectURL != "" {
		if _, err := url.Parse(c.EnableBankingRedirectURL); err != nil {
			problems = append(problems, fmt.Errorf("WIMM_ENABLEBANKING_REDIRECT_URL: %w", err))
		}
	}

	// Both key files, checked the same way, so neither is the one that was
	// forgotten.
	for _, f := range []struct{ key, path string }{
		{"WIMM_BANKING_ENCRYPTION_KEY", c.BankingEncryptionKeyPath},
		{"WIMM_ENABLEBANKING_PRIVATE_KEY", c.EnableBankingPrivateKeyPath},
	} {
		if f.path == "" {
			problems = append(problems, fmt.Errorf("%s is not set, and %s is enabled", f.key, c.BankingGateway))
			continue
		}
		if err := banking.RequireSecretFile(f.path); err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", f.key, err))
		}
	}
	return problems
}
