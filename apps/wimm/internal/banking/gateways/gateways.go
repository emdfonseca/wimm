// Package gateways builds the adapter a configuration names.
//
// It exists so that **nothing outside internal/banking imports an adapter** —
// not even wiring. A composition root that reached for a provider package
// directly would be the one exception that makes the lint rule advisory, and
// that rule is the only thing standing between a seam and a convention
// (ADR 0018).
//
// It is a package of its own rather than a function in banking because the
// adapters import banking: the port cannot import its own implementations.
package gateways

import (
	"fmt"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/banking/enablebanking"
)

// Config is what building a gateway needs, in wimm's own terms. No field here
// is named after a provider.
type Config struct {
	// Name selects the adapter. It is also stored on every connection, so a
	// later gateway is new connections rather than a reinterpretation of old
	// rows.
	Name string
	// ApplicationID is the registered application.
	ApplicationID string
	// PrivateKeyPath is the request signing key. A path, never the material.
	PrivateKeyPath string
	// RedirectURL must match one registered with the gateway.
	RedirectURL string
}

// New builds the adapter cfg names.
func New(cfg Config) (banking.Gateway, error) {
	switch cfg.Name {
	case "enablebanking":
		return enablebanking.New(enablebanking.Options{
			ApplicationID:  cfg.ApplicationID,
			PrivateKeyPath: cfg.PrivateKeyPath,
			RedirectURL:    cfg.RedirectURL,
		})
	default:
		return nil, fmt.Errorf("banking: no adapter for gateway %q", cfg.Name)
	}
}
