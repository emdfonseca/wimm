package identity

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// Ticket is what a redeemed enrolment link becomes. The value goes into an
// HttpOnly cookie; the link value it replaced never travels again.
type Ticket struct {
	Value     string
	ExpiresAt time.Time
}

// RedeemEnrolmentLink exchanges a link value for a ticket and returns the
// member it belongs to, so the page can show them the name they were
// registered under before asking them to create anything.
//
// Expired, spent, replaced and never-issued all return
// ErrEnrolmentLinkUnusable. Nothing distinguishes them.
func (s *Service) RedeemEnrolmentLink(ctx context.Context, linkValue string) (Ticket, store.Member, error) {
	linkValue = strings.TrimSpace(linkValue)
	if linkValue == "" {
		return Ticket{}, store.Member{}, ErrEnrolmentLinkUnusable
	}

	ticketValue, ticketHash, err := mintSecret()
	if err != nil {
		return Ticket{}, store.Member{}, err
	}

	row, member, err := s.db.RedeemEnrolmentLink(ctx, hashSecret(linkValue), ticketHash, s.ticketLifetime)
	if errors.Is(err, store.ErrNotFound) {
		return Ticket{}, store.Member{}, ErrEnrolmentLinkUnusable
	}
	if err != nil {
		return Ticket{}, store.Member{}, err
	}

	return Ticket{Value: ticketValue, ExpiresAt: row.ExpiresAt}, member, nil
}
