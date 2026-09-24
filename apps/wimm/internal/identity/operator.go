package identity

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// EnrolmentLink is what the operator is handed: a URL to pass on and the
// moment it stops working, so they can say so when they send it.
//
// The URL carries the only copy of the value. It is returned once and cannot
// be read back.
type EnrolmentLink struct {
	URL       string
	ExpiresAt time.Time
}

// RegisterMember registers a person and mints their first enrolment link.
//
// The email, first name and last name are operator assertions. Nothing
// verifies that the address belongs to the person named.
func (s *Service) RegisterMember(ctx context.Context, email, firstName, lastName string) (store.Member, EnrolmentLink, error) {
	firstName, lastName = strings.TrimSpace(firstName), strings.TrimSpace(lastName)
	if firstName == "" || lastName == "" {
		return store.Member{}, EnrolmentLink{}, missingNames(firstName, lastName)
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return store.Member{}, EnrolmentLink{}, errors.New("an email address is required")
	}

	m, err := s.db.CreateMember(ctx, email, firstName, lastName)
	if errors.Is(err, store.ErrEmailAlreadyRegistered) {
		// No link is issued, and the registered person is untouched.
		return store.Member{}, EnrolmentLink{}, ErrEmailAlreadyRegistered
	}
	if err != nil {
		return store.Member{}, EnrolmentLink{}, err
	}

	link, err := s.mintEnrolmentLink(ctx, m.ID)
	if err != nil {
		return store.Member{}, EnrolmentLink{}, err
	}
	return m, link, nil
}

// IssueEnrolmentLink hands a further link to someone already registered, for a
// new device or one they have lost.
//
// Any outstanding link stops working. Passkeys they have already enrolled keep
// working: this adds a way in, it does not close the existing ones.
func (s *Service) IssueEnrolmentLink(ctx context.Context, email string) (store.Member, EnrolmentLink, error) {
	m, err := s.db.MemberByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, store.ErrNotFound) {
		return store.Member{}, EnrolmentLink{}, ErrMemberNotFound
	}
	if err != nil {
		return store.Member{}, EnrolmentLink{}, err
	}

	link, err := s.mintEnrolmentLink(ctx, m.ID)
	if err != nil {
		return store.Member{}, EnrolmentLink{}, err
	}
	return m, link, nil
}

func (s *Service) mintEnrolmentLink(ctx context.Context, memberID string) (EnrolmentLink, error) {
	value, hash, err := mintSecret()
	if err != nil {
		return EnrolmentLink{}, err
	}

	row, err := s.db.IssueEnrolmentLink(ctx, memberID, hash, s.enrolmentLinkLifetime)
	if err != nil {
		return EnrolmentLink{}, err
	}

	return EnrolmentLink{
		URL:       s.baseURL + EnrolmentPath + url.PathEscape(value),
		ExpiresAt: row.ExpiresAt,
	}, nil
}

// EnrolmentPath is where an enrolment link points. The value follows it.
const EnrolmentPath = "/enrol/"

func missingNames(firstName, lastName string) error {
	switch {
	case firstName == "" && lastName == "":
		return fmt.Errorf("%w: both are missing", ErrNameMissing)
	case firstName == "":
		return fmt.Errorf("%w: the first name is missing", ErrNameMissing)
	default:
		return fmt.Errorf("%w: the last name is missing", ErrNameMissing)
	}
}
