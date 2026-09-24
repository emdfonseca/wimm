package identity_test

import (
	"context"
	"errors"
	"testing"

	"github.com/emdfonseca/wimm/apps/wimm/internal/identity"
	"github.com/emdfonseca/wimm/apps/wimm/internal/identity/authenticatortest"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// testOrigin matches the WIMM_BASE_URL the test service is configured with,
// which is also its only expected origin.
const testOrigin = "http://localhost:9466"

type enrolled struct {
	svc           *identity.Service
	db            *store.DB
	member        store.Member
	authenticator *authenticatortest.Authenticator
	session       identity.Session
}

// enrol runs a whole enrolment: register, redeem, begin, finish.
func enrol(t *testing.T, email, first, last string) enrolled {
	t.Helper()

	svc, db := newService(t)
	ctx := context.Background()

	m, link, err := svc.RegisterMember(ctx, email, first, last)
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	ticket, shown, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, link))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}
	// The member is shown who the link is for before they are asked to create
	// anything.
	if shown.FirstName != first || shown.LastName != last {
		t.Errorf("redeeming showed %v, want the registered name", shown)
	}

	ceremony, _, err := svc.BeginEnrolment(ctx, ticket.Value)
	if err != nil {
		t.Fatalf("beginning enrolment: %v", err)
	}

	auth := authenticatortest.New(t)
	member, session, err := svc.FinishEnrolment(ctx, ceremony.ID,
		auth.Register(t, ceremony.OptionsJSON, testOrigin))
	if err != nil {
		t.Fatalf("finishing enrolment: %v", err)
	}
	if member.ID != m.ID {
		t.Fatalf("enrolled %s, want %s", member.ID, m.ID)
	}

	return enrolled{svc: svc, db: db, member: member, authenticator: auth, session: session}
}

// Enrolment signs the member in: the ceremony has just verified them, and a
// second one seconds later proves nothing new.
func TestEnrolmentSavesThePasskeyAndSignsTheMemberIn(t *testing.T) {
	e := enrol(t, "ada@example.com", "Ada", "Lovelace")
	ctx := context.Background()

	if e.session.Value == "" {
		t.Fatal("enrolment produced no session")
	}
	m, err := e.svc.MemberForSession(ctx, e.session.Value)
	if err != nil {
		t.Fatalf("the session enrolment created does not work: %v", err)
	}
	if m.FirstName != "Ada" {
		t.Errorf("the session names %v", m)
	}

	creds, err := e.db.CredentialsForMember(ctx, e.member.ID)
	if err != nil {
		t.Fatalf("listing passkeys: %v", err)
	}
	if len(creds) != 1 {
		t.Fatalf("%d passkeys saved, want one", len(creds))
	}
}

// A passkey the member could never be offered at sign-in is refused, and the
// refusal costs them nothing: they are un-enrolled with their link still
// usable, so the next attempt on another device can succeed.
func TestAPasskeyTheDeviceWillNotSaveIsRefused(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	m, link, err := svc.RegisterMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	value := linkValue(t, link)

	ticket, _, err := svc.RedeemEnrolmentLink(ctx, value)
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}
	ceremony, _, err := svc.BeginEnrolment(ctx, ticket.Value)
	if err != nil {
		t.Fatalf("beginning enrolment: %v", err)
	}

	refusing := authenticatortest.New(t)
	refusing.Discoverable = false

	_, _, err = svc.FinishEnrolment(ctx, ceremony.ID,
		refusing.Register(t, ceremony.OptionsJSON, testOrigin))
	if !errors.Is(err, identity.ErrNotDiscoverable) {
		t.Fatalf("err = %v, want ErrNotDiscoverable", err)
	}

	creds, err := db.CredentialsForMember(ctx, m.ID)
	if err != nil {
		t.Fatalf("listing passkeys: %v", err)
	}
	if len(creds) != 0 {
		t.Errorf("%d passkeys saved by a refused enrolment", len(creds))
	}

	// The link still works, on this device or another.
	retryTicket, _, err := svc.RedeemEnrolmentLink(ctx, value)
	if err != nil {
		t.Fatalf("the link stopped working after a refused attempt: %v", err)
	}
	retry, _, err := svc.BeginEnrolment(ctx, retryTicket.Value)
	if err != nil {
		t.Fatalf("beginning a second attempt: %v", err)
	}
	if _, _, err := svc.FinishEnrolment(ctx, retry.ID,
		authenticatortest.New(t).Register(t, retry.OptionsJSON, testOrigin)); err != nil {
		t.Fatalf("the second attempt failed: %v", err)
	}
}

// A client that does not implement credProps reports nothing, which is not the
// same as reporting false and is not grounds to refuse.
func TestAClientThatDoesNotReportCredPropsIsAccepted(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()

	_, link, err := svc.RegisterMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	ticket, _, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, link))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}
	ceremony, _, err := svc.BeginEnrolment(ctx, ticket.Value)
	if err != nil {
		t.Fatalf("beginning enrolment: %v", err)
	}

	quiet := authenticatortest.New(t)
	quiet.ReportCredProps = false

	if _, _, err := svc.FinishEnrolment(ctx, ceremony.ID,
		quiet.Register(t, ceremony.OptionsJSON, testOrigin)); err != nil {
		t.Fatalf("a client that does not report credProps was refused: %v", err)
	}
}

// The link is single-use in the sense that matters: one passkey.
func TestTheLinkCannotEnrolASecondPasskey(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()

	_, link, err := svc.RegisterMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	value := linkValue(t, link)

	ticket, _, err := svc.RedeemEnrolmentLink(ctx, value)
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}
	ceremony, _, err := svc.BeginEnrolment(ctx, ticket.Value)
	if err != nil {
		t.Fatalf("beginning enrolment: %v", err)
	}
	if _, _, err := svc.FinishEnrolment(ctx, ceremony.ID,
		authenticatortest.New(t).Register(t, ceremony.OptionsJSON, testOrigin)); err != nil {
		t.Fatalf("finishing enrolment: %v", err)
	}

	if _, _, err := svc.RedeemEnrolmentLink(ctx, value); !errors.Is(err, identity.ErrEnrolmentLinkUnusable) {
		t.Errorf("the link worked a second time: err = %v", err)
	}
}

// A challenge is answered at most once, whatever else is running.
func TestACeremonyCannotBeAnsweredTwice(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()

	_, link, err := svc.RegisterMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	ticket, _, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, link))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}
	ceremony, _, err := svc.BeginEnrolment(ctx, ticket.Value)
	if err != nil {
		t.Fatalf("beginning enrolment: %v", err)
	}

	auth := authenticatortest.New(t)
	response := auth.Register(t, ceremony.OptionsJSON, testOrigin)

	if _, _, err := svc.FinishEnrolment(ctx, ceremony.ID, response); err != nil {
		t.Fatalf("finishing enrolment: %v", err)
	}
	if _, _, err := svc.FinishEnrolment(ctx, ceremony.ID, response); err == nil {
		t.Error("the same challenge was answered twice")
	}
}

// A link enrols one passkey. The ticket it became has its own lifetime, so
// without this it outlives the link it came from: the browser that just
// finished enrolling still holds a usable ticket cookie and can enrol a second
// passkey through a link the operator was told is single use.
func TestASpentLinksTicketCannotEnrolAgain(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	_, link, err := svc.RegisterMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	ticket, _, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, link))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}
	ceremony, member, err := svc.BeginEnrolment(ctx, ticket.Value)
	if err != nil {
		t.Fatalf("beginning enrolment: %v", err)
	}
	if _, _, err := svc.FinishEnrolment(ctx, ceremony.ID,
		authenticatortest.New(t).Register(t, ceremony.OptionsJSON, testOrigin)); err != nil {
		t.Fatalf("finishing enrolment: %v", err)
	}

	// The same ticket, still inside its own lifetime.
	if _, _, err := svc.BeginEnrolment(ctx, ticket.Value); !errors.Is(err, identity.ErrEnrolmentLinkUnusable) {
		t.Fatalf("the spent link's ticket still begins an enrolment: err = %v", err)
	}

	creds, err := db.CredentialsForMember(ctx, member.ID)
	if err != nil {
		t.Fatalf("listing passkeys: %v", err)
	}
	if len(creds) != 1 {
		t.Errorf("%d passkeys enrolled through a single-use link", len(creds))
	}
}
