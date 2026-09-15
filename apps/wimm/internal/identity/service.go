package identity

import (
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// Service is the identity domain: who a member is, how they prove it, and what
// makes a browser signed in.
type Service struct {
	db *store.DB
	wa *webauthn.WebAuthn

	baseURL               string
	enrolmentLinkLifetime time.Duration
	ticketLifetime        time.Duration
	sessionLifetime       time.Duration
	ceremonyLifetime      time.Duration
}

// TicketLifetime is how long a redeemed enrolment link's ticket lasts. It
// covers one sitting at the enrolment page, not the life of the link.
const TicketLifetime = 30 * time.Minute

// New builds the domain from validated configuration.
//
// The relying-party identifier arrives here and is never adjusted afterwards:
// changing it invalidates every passkey ever registered against it (ADR 0016).
func New(db *store.DB, cfg config.Config) (*Service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          cfg.RelyingPartyID,
		RPDisplayName: cfg.RelyingPartyName,
		RPOrigins:     cfg.Origins,
	})
	if err != nil {
		return nil, fmt.Errorf("configuring WebAuthn: %w", err)
	}

	return &Service{
		db:                    db,
		wa:                    wa,
		baseURL:               cfg.BaseURL,
		enrolmentLinkLifetime: cfg.EnrolmentLinkLifetime,
		ticketLifetime:        TicketLifetime,
		sessionLifetime:       cfg.SessionLifetime,
		ceremonyLifetime:      cfg.CeremonyLifetime,
	}, nil
}
