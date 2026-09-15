package rpc_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	identityv1 "github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1"

	"github.com/xuuid/wimm/apps/wimm/internal/identity"
)

func TestRegisterMemberIssuesOneLinkWithItsExpiry(t *testing.T) {
	f := newOperatorFixture(t)

	res, err := f.client.RegisterMember(context.Background(),
		withCredential(registerRequest("ada@example.com", "Ada", "Lovelace"), testCredential))
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	m := res.Msg.GetMember()
	if m.GetFirstName() != "Ada" || m.GetLastName() != "Lovelace" || m.GetEmail() != "ada@example.com" {
		t.Errorf("member = %v, want the name and address the operator gave", m)
	}

	link := res.Msg.GetEnrolmentLink()
	if !strings.Contains(link.GetUrl(), identity.EnrolmentPath) {
		t.Errorf("link URL = %q, want it to carry %q", link.GetUrl(), identity.EnrolmentPath)
	}

	// The operator is told when it stops working, so they can say so when they
	// send it.
	expires := link.GetExpiresAt().AsTime()
	if expires.IsZero() {
		t.Fatal("the link carries no expiry")
	}
	if d := time.Until(expires); d < 23*time.Hour || d > 25*time.Hour {
		t.Errorf("link expires in %s, want about the 24h default", d)
	}

	if n := f.linkCount(t); n != 1 {
		t.Errorf("%d links issued, want exactly one", n)
	}
}

func TestRegisterMemberRefusesAMissingName(t *testing.T) {
	for _, tc := range []struct{ name, first, last, wants string }{
		{"no first name", "", "Lovelace", "first name"},
		{"no last name", "Ada", "", "last name"},
		{"neither", "  ", "\t", "both"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newOperatorFixture(t)

			_, err := f.client.RegisterMember(context.Background(),
				withCredential(registerRequest("ada@example.com", tc.first, tc.last), testCredential))
			if err == nil {
				t.Fatal("a registration with a missing name was accepted")
			}
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want %v", got, connect.CodeInvalidArgument)
			}
			// The operator is told which part is missing.
			if !strings.Contains(err.Error(), tc.wants) {
				t.Errorf("error %q does not say which part is missing (%q)", err, tc.wants)
			}

			if n := f.memberCount(t); n != 0 {
				t.Errorf("%d members registered by a refused registration", n)
			}
			if n := f.linkCount(t); n != 0 {
				t.Errorf("%d links issued by a refused registration", n)
			}
		})
	}
}

func TestRegisterMemberRefusesAnAlreadyRegisteredEmail(t *testing.T) {
	f := newOperatorFixture(t)
	ctx := context.Background()

	if _, err := f.client.RegisterMember(ctx,
		withCredential(registerRequest("ada@example.com", "Ada", "Lovelace"), testCredential)); err != nil {
		t.Fatalf("registering the first time: %v", err)
	}

	// Case folding is the only normalisation: one address, one person.
	_, err := f.client.RegisterMember(ctx,
		withCredential(registerRequest("Ada@Example.com", "Someone", "Else"), testCredential))
	if err == nil {
		t.Fatal("a second registration under the same address was accepted")
	}
	if got := connect.CodeOf(err); got != connect.CodeAlreadyExists {
		t.Errorf("code = %v, want %v", got, connect.CodeAlreadyExists)
	}
	// The operator is trusted and has to act on this, so they are told plainly.
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("error %q does not say the address is in use", err)
	}

	if n := f.memberCount(t); n != 1 {
		t.Errorf("%d members exist, want the original one only", n)
	}
	if n := f.linkCount(t); n != 1 {
		t.Errorf("%d links exist, want the original one only", n)
	}

	// The registered person keeps their name.
	m, err := f.db.MemberByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("reading the original member: %v", err)
	}
	if m.FirstName != "Ada" || m.LastName != "Lovelace" {
		t.Errorf("the existing member was changed: %v", m)
	}
}

func issueRequest(email string) *connect.Request[identityv1.IssueEnrolmentLinkRequest] {
	return connect.NewRequest(&identityv1.IssueEnrolmentLinkRequest{
		Email: &identityv1.EmailAddress{Value: email},
	})
}
