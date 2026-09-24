package identity

import (
	"context"
	"errors"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// Session is what makes a browser signed in. The value goes into the cookie
// and is never stored; the row it names is what can be revoked.
type Session struct {
	Value     string
	ExpiresAt time.Time
}

// LandingPath is where a member arrives with nothing more specific to return
// to.
const LandingPath = "/"

func (s *Service) createSession(ctx context.Context, memberID string) (Session, error) {
	value, hash, err := mintSecret()
	if err != nil {
		return Session{}, err
	}

	// The identifier is minted fresh, so enrolment does not continue a session
	// that was already on this browser.
	row, err := s.db.CreateSession(ctx, memberID, hash, s.sessionLifetime)
	if err != nil {
		return Session{}, err
	}
	return Session{Value: value, ExpiresAt: row.ExpiresAt}, nil
}

// MemberForSession returns who a session belongs to, and records that it was
// seen. A revoked or expired session is ErrNotSignedIn: there is no third
// answer, and nothing says which of the two it was.
func (s *Service) MemberForSession(ctx context.Context, value string) (store.Member, error) {
	if value == "" {
		return store.Member{}, ErrNotSignedIn
	}

	_, m, err := s.db.SessionByHash(ctx, hashSecret(value))
	if errors.Is(err, store.ErrNotFound) {
		return store.Member{}, ErrNotSignedIn
	}
	if err != nil {
		return store.Member{}, err
	}
	return m, nil
}

// SignOut stops a session working on this browser. Signing out a session that
// is already gone is not an error: the member asked to be signed out and they
// are.
func (s *Service) SignOut(ctx context.Context, value string) error {
	if value == "" {
		return nil
	}
	err := s.db.RevokeSession(ctx, hashSecret(value))
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	return err
}

// safeReturnPath keeps a path only if it is one of ours. Anything else falls
// back to the landing page, so there is no open redirect to defend against.
func safeReturnPath(path string) string {
	if len(path) < 2 || path[0] != '/' || path[1] == '/' || path[1] == '\\' {
		return LandingPath
	}
	return path
}
