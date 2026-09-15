package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// Ceremony is a WebAuthn challenge in flight: an identifier naming the row,
// and the options the browser needs.
type Ceremony struct {
	ID          string
	OptionsJSON string
}

// BeginEnrolment starts the registration ceremony for a redeemed link.
//
// A discoverable credential and user verification are both required. Without
// the first, the member could enrol a passkey they can never be offered at
// sign-in, and the only way back would be another link from the operator.
func (s *Service) BeginEnrolment(ctx context.Context, ticketValue string) (Ceremony, store.Member, error) {
	ticket, m, err := s.db.EnrolmentTicketByHash(ctx, hashSecret(ticketValue))
	if errors.Is(err, store.ErrNotFound) {
		return Ceremony{}, store.Member{}, ErrEnrolmentLinkUnusable
	}
	if err != nil {
		return Ceremony{}, store.Member{}, err
	}

	existing, err := s.db.CredentialsForMember(ctx, m.ID)
	if err != nil {
		return Ceremony{}, store.Member{}, err
	}
	user, err := newWebAuthnUser(m, existing)
	if err != nil {
		return Ceremony{}, store.Member{}, err
	}

	creation, session, err := s.wa.BeginRegistration(user,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			RequireResidentKey: protocol.ResidentKeyRequired(),
			UserVerification:   protocol.VerificationRequired,
		}),
		// credProps is what reports back whether a discoverable credential was
		// actually created. Unrequested, its output would fail the ceremony.
		webauthn.WithExtensions(webauthn.WithExtensionCredProps()),
		webauthn.WithExclusions(webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()),
	)
	if err != nil {
		return Ceremony{}, store.Member{}, fmt.Errorf("beginning enrolment: %w", err)
	}

	row, err := s.storeCeremony(ctx, store.Ceremony{
		Kind:            store.CeremonyRegistration,
		MemberID:        &m.ID,
		EnrolmentLinkID: &ticket.LinkID,
	}, session)
	if err != nil {
		return Ceremony{}, store.Member{}, err
	}

	options, err := json.Marshal(creation.Response)
	if err != nil {
		return Ceremony{}, store.Member{}, fmt.Errorf("encoding creation options: %w", err)
	}
	return Ceremony{ID: row.ID, OptionsJSON: string(options)}, m, nil
}

// FinishEnrolment records the passkey and signs the member in.
//
// The ceremony has just performed user verification, so asking for a second
// one seconds later re-proves the same fact and costs a prompt.
func (s *Service) FinishEnrolment(ctx context.Context, ceremonyID, credentialJSON string) (store.Member, Session, error) {
	row, session, err := s.consumeCeremony(ctx, ceremonyID, store.CeremonyRegistration)
	if err != nil {
		return store.Member{}, Session{}, err
	}
	if row.MemberID == nil || row.EnrolmentLinkID == nil {
		return store.Member{}, Session{}, ErrEnrolmentLinkUnusable
	}

	m, err := s.db.MemberByID(ctx, *row.MemberID)
	if err != nil {
		return store.Member{}, Session{}, err
	}
	existing, err := s.db.CredentialsForMember(ctx, m.ID)
	if err != nil {
		return store.Member{}, Session{}, err
	}
	user, err := newWebAuthnUser(m, existing)
	if err != nil {
		return store.Member{}, Session{}, err
	}

	parsed, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(credentialJSON))
	if err != nil {
		return store.Member{}, Session{}, fmt.Errorf("%w: %w", ErrNotDiscoverable, err)
	}

	credential, err := s.wa.CreateCredential(user, session, parsed)
	if err != nil {
		// Nothing is written and the link is untouched, so the member can try
		// again on another device.
		return store.Member{}, Session{}, fmt.Errorf("%w: %w", ErrNotDiscoverable, err)
	}

	// credProps reports three states. Explicit false is a refusal; nil means
	// the client did not report, which is not the same thing and is not
	// grounds to reject a credential that may well be discoverable.
	if credential.Extensions.RK != nil && !*credential.Extensions.RK {
		return store.Member{}, Session{}, ErrNotDiscoverable
	}

	data, err := json.Marshal(credential)
	if err != nil {
		return store.Member{}, Session{}, fmt.Errorf("encoding a passkey: %w", err)
	}
	if _, err := s.db.CreateCredential(ctx, store.Credential{
		MemberID:     m.ID,
		CredentialID: credential.ID,
		Data:         data,
	}); err != nil {
		return store.Member{}, Session{}, err
	}

	// The link closes only now, on a passkey that was actually saved.
	if err := s.db.MarkLinkEnrolled(ctx, *row.EnrolmentLinkID); err != nil && !errors.Is(err, store.ErrNotFound) {
		return store.Member{}, Session{}, err
	}

	sess, err := s.createSession(ctx, m.ID)
	if err != nil {
		return store.Member{}, Session{}, err
	}
	return m, sess, nil
}

// BeginSignIn starts a username-less authentication ceremony.
//
// intendedPath is where the member was heading. It is held against this row
// rather than carried in the URL, so there is nothing to tamper with.
func (s *Service) BeginSignIn(ctx context.Context, intendedPath string) (Ceremony, error) {
	assertion, session, err := s.wa.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		return Ceremony{}, fmt.Errorf("beginning sign-in: %w", err)
	}

	path := safeReturnPath(intendedPath)
	row, err := s.storeCeremony(ctx, store.Ceremony{
		Kind:         store.CeremonyAuthentication,
		IntendedPath: &path,
	}, session)
	if err != nil {
		return Ceremony{}, err
	}

	options, err := json.Marshal(assertion.Response)
	if err != nil {
		return Ceremony{}, fmt.Errorf("encoding request options: %w", err)
	}
	return Ceremony{ID: row.ID, OptionsJSON: string(options)}, nil
}

// FinishSignIn validates the assertion and creates the session, returning the
// path the member was trying to reach.
func (s *Service) FinishSignIn(ctx context.Context, ceremonyID, credentialJSON string) (store.Member, Session, string, error) {
	row, session, err := s.consumeCeremony(ctx, ceremonyID, store.CeremonyAuthentication)
	if err != nil {
		return store.Member{}, Session{}, "", err
	}

	parsed, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(credentialJSON))
	if err != nil {
		return store.Member{}, Session{}, "", ErrPasskeyNotRecognised
	}

	var stored store.Credential
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		c, err := s.db.CredentialByID(ctx, rawID)
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrPasskeyNotRecognised
		}
		if err != nil {
			return nil, err
		}
		stored = c

		m, err := s.db.MemberByID(ctx, string(userHandle))
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrPasskeyNotRecognised
		}
		if err != nil {
			return nil, err
		}
		if m.ID != c.MemberID {
			return nil, ErrPasskeyNotRecognised
		}

		rows, err := s.db.CredentialsForMember(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		return newWebAuthnUser(m, rows)
	}

	user, credential, err := s.wa.ValidatePasskeyLogin(handler, session, parsed)
	if err != nil {
		// Whether the passkey, the member or neither is the problem is not
		// said, here or to the member.
		return store.Member{}, Session{}, "", ErrPasskeyNotRecognised
	}

	m, ok := user.(*member)
	if !ok {
		return store.Member{}, Session{}, "", ErrPasskeyNotRecognised
	}

	// The signature counter and the authenticator flags moved; the stored
	// record has to move with them.
	data, err := json.Marshal(credential)
	if err != nil {
		return store.Member{}, Session{}, "", fmt.Errorf("encoding a passkey: %w", err)
	}
	if err := s.db.UpdateCredential(ctx, stored.ID, data); err != nil {
		return store.Member{}, Session{}, "", err
	}

	sess, err := s.createSession(ctx, m.row.ID)
	if err != nil {
		return store.Member{}, Session{}, "", err
	}

	path := LandingPath
	if row.IntendedPath != nil {
		path = safeReturnPath(*row.IntendedPath)
	}
	return m.row, sess, path, nil
}

func (s *Service) storeCeremony(ctx context.Context, c store.Ceremony, session *webauthn.SessionData) (store.Ceremony, error) {
	data, err := json.Marshal(session)
	if err != nil {
		return store.Ceremony{}, fmt.Errorf("encoding ceremony state: %w", err)
	}
	c.SessionData = data

	row, err := s.db.CreateCeremony(ctx, c, s.ceremonyLifetime)
	if err != nil {
		return store.Ceremony{}, err
	}
	return row, nil
}

// consumeCeremony reads the challenge and deletes it in one statement, so a
// challenge is answered at most once however many instances are running.
func (s *Service) consumeCeremony(ctx context.Context, id string, kind store.CeremonyKind) (store.Ceremony, webauthn.SessionData, error) {
	row, err := s.db.ConsumeCeremony(ctx, id, kind)
	if errors.Is(err, store.ErrNotFound) {
		if kind == store.CeremonyRegistration {
			return store.Ceremony{}, webauthn.SessionData{}, ErrEnrolmentLinkUnusable
		}
		return store.Ceremony{}, webauthn.SessionData{}, ErrPasskeyNotRecognised
	}
	if err != nil {
		return store.Ceremony{}, webauthn.SessionData{}, err
	}

	// Decoded into a fresh value: the library records which extensions were
	// requested, and a reused struct carries the previous ceremony's.
	var session webauthn.SessionData
	if err := json.Unmarshal(row.SessionData, &session); err != nil {
		return store.Ceremony{}, webauthn.SessionData{}, fmt.Errorf("reading ceremony state: %w", err)
	}
	return row, session, nil
}
