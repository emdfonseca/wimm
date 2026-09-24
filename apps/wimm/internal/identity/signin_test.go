package identity_test

import (
	"context"
	"errors"
	"testing"

	"github.com/emdfonseca/wimm/apps/wimm/internal/identity"
	"github.com/emdfonseca/wimm/apps/wimm/internal/identity/authenticatortest"
)

// Coming back after closing the browser: no email, no username, nothing typed.
func TestSignInWithAnEnrolledPasskey(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	ceremony, err := e.svc.BeginSignIn(ctx, "")
	if err != nil {
		t.Fatalf("beginning sign-in: %v", err)
	}

	m, session, returnPath, err := e.svc.FinishSignIn(ctx, ceremony.ID,
		e.authenticator.Authenticate(t, ceremony.OptionsJSON, testOrigin, e.member.ID))
	if err != nil {
		t.Fatalf("signing in: %v", err)
	}
	if m.ID != e.member.ID {
		t.Errorf("signed in as %s, want %s", m.ID, e.member.ID)
	}
	if m.FirstName != "Ada" {
		t.Errorf("the member's own name is not shown: %v", m)
	}
	if returnPath != identity.LandingPath {
		t.Errorf("returnPath = %q, want the landing page", returnPath)
	}

	if _, err := e.svc.MemberForSession(ctx, session.Value); err != nil {
		t.Errorf("the session sign-in created does not work: %v", err)
	}
}

// A passkey for this site that this instance has never seen signs nobody in,
// and says nothing about whether the passkey, the member or neither is the
// problem.
func TestSignInWithAPasskeyThisInstanceDoesNotKnow(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	ceremony, err := e.svc.BeginSignIn(ctx, "")
	if err != nil {
		t.Fatalf("beginning sign-in: %v", err)
	}

	stranger := authenticatortest.New(t)
	_, _, _, err = e.svc.FinishSignIn(ctx, ceremony.ID,
		stranger.Authenticate(t, ceremony.OptionsJSON, testOrigin, e.member.ID))
	if !errors.Is(err, identity.ErrPasskeyNotRecognised) {
		t.Fatalf("err = %v, want ErrPasskeyNotRecognised", err)
	}
}

// Sign-out is what makes the sign-in story testable without waiting for a
// session to expire, and it is the recovery story for a shared computer.
func TestSignOutStopsTheSessionWorking(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	if _, err := e.svc.MemberForSession(ctx, e.session.Value); err != nil {
		t.Fatalf("the session does not work before signing out: %v", err)
	}

	if err := e.svc.SignOut(ctx, e.session.Value); err != nil {
		t.Fatalf("signing out: %v", err)
	}

	if _, err := e.svc.MemberForSession(ctx, e.session.Value); !errors.Is(err, identity.ErrNotSignedIn) {
		t.Errorf("the session still works after signing out: err = %v", err)
	}
}

// The member is returned to what they were looking at, not to a generic
// starting page — and the path comes from the server's own record of the
// attempt, so a client cannot choose it.
func TestSignInReturnsToTheIntendedPath(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	ceremony, err := e.svc.BeginSignIn(ctx, "/accounts/current")
	if err != nil {
		t.Fatalf("beginning sign-in: %v", err)
	}

	_, _, returnPath, err := e.svc.FinishSignIn(ctx, ceremony.ID,
		e.authenticator.Authenticate(t, ceremony.OptionsJSON, testOrigin, e.member.ID))
	if err != nil {
		t.Fatalf("signing in: %v", err)
	}
	if returnPath != "/accounts/current" {
		t.Errorf("returnPath = %q, want the path the member was heading to", returnPath)
	}
}

// Anything that is not one of our own paths falls back to the landing page,
// so there is no open redirect to defend against.
func TestAnIntendedPathThatIsNotOursFallsBackToTheLanding(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	for _, path := range []string{
		"https://evil.example/steal",
		"//evil.example/steal",
		"/\\evil.example",
		"",
		"not-a-path",
	} {
		t.Run(path, func(t *testing.T) {
			ceremony, err := e.svc.BeginSignIn(ctx, path)
			if err != nil {
				t.Fatalf("beginning sign-in: %v", err)
			}
			_, _, returnPath, err := e.svc.FinishSignIn(ctx, ceremony.ID,
				e.authenticator.Authenticate(t, ceremony.OptionsJSON, testOrigin, e.member.ID))
			if err != nil {
				t.Fatalf("signing in: %v", err)
			}
			if returnPath != identity.LandingPath {
				t.Errorf("returnPath = %q, want the landing page", returnPath)
			}
		})
	}
}

// A dismissed prompt sends nothing, so nothing changes and the member can try
// again: the ceremony they abandoned simply expires.
func TestADismissedPromptLeavesTheMemberAbleToTryAgain(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	if _, err := e.svc.BeginSignIn(ctx, ""); err != nil {
		t.Fatalf("beginning an abandoned sign-in: %v", err)
	}

	second, err := e.svc.BeginSignIn(ctx, "")
	if err != nil {
		t.Fatalf("beginning sign-in again: %v", err)
	}
	if _, _, _, err := e.svc.FinishSignIn(ctx, second.ID,
		e.authenticator.Authenticate(t, second.OptionsJSON, testOrigin, e.member.ID)); err != nil {
		t.Fatalf("the second attempt failed: %v", err)
	}
}
