package identity_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
	"github.com/xuuid/wimm/apps/wimm/internal/identity"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
	"github.com/xuuid/wimm/apps/wimm/internal/store/storetest"
)

func newService(t *testing.T) (*identity.Service, *store.DB) {
	t.Helper()

	db := storetest.New(t)

	cfg, err := config.Load(func(k string) string {
		return map[string]string{
			"WIMM_OPERATOR_CREDENTIAL": "credential",
			"WIMM_DATABASE_URL":        storetest.URL(t),
			"WIMM_BASE_URL":            "http://localhost:9466",
		}[k]
	})
	if err != nil {
		t.Fatalf("building test configuration: %v", err)
	}

	svc, err := identity.New(db, cfg)
	if err != nil {
		t.Fatalf("building the identity service: %v", err)
	}
	return svc, db
}

// linkValue pulls the value out of a minted URL, the way the web app pulls it
// out of the path it was opened on.
func linkValue(t *testing.T, link identity.EnrolmentLink) string {
	t.Helper()

	u, err := url.Parse(link.URL)
	if err != nil {
		t.Fatalf("parsing the enrolment URL %q: %v", link.URL, err)
	}
	value := strings.TrimPrefix(u.Path, identity.EnrolmentPath)
	if value == "" || value == u.Path {
		t.Fatalf("enrolment URL %q carries no value under %q", link.URL, identity.EnrolmentPath)
	}
	return value
}

// Issuing a further link is how a member adds a device or replaces a lost one.
// It must close the old way in and leave every existing way in open.
func TestIssuingAFurtherLinkReplacesTheOutstandingOne(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	member, first, err := svc.RegisterMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	// A passkey they enrolled before must keep working across a reissue.
	if _, err := db.CreateCredential(ctx, store.Credential{
		MemberID:     member.ID,
		CredentialID: []byte("credential-from-an-earlier-device"),
		Data:         []byte(`{"id":"ZWFybGllcg","publicKey":"a2V5"}`),
	}); err != nil {
		t.Fatalf("recording an existing passkey: %v", err)
	}

	// Taken before the reissue, so it is outstanding when the link is replaced.
	earlier, _, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, first))
	if err != nil {
		t.Fatalf("redeeming the first link: %v", err)
	}

	_, second, err := svc.IssueEnrolmentLink(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("issuing a further link: %v", err)
	}
	if second.URL == first.URL {
		t.Fatal("the second link is the same URL as the first")
	}

	if _, _, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, first)); !errors.Is(err, identity.ErrEnrolmentLinkUnusable) {
		t.Errorf("the earlier link still works: err = %v", err)
	}

	// And a ticket already taken from that link dies with it. Otherwise
	// reissuing closes the front door and leaves a window open for as long as
	// the ticket lasts.
	if _, _, err := svc.BeginEnrolment(ctx, earlier.Value); !errors.Is(err, identity.ErrEnrolmentLinkUnusable) {
		t.Errorf("a ticket from the replaced link still begins an enrolment: err = %v", err)
	}

	if _, _, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, second)); err != nil {
		t.Errorf("the new link does not work: %v", err)
	}

	creds, err := db.CredentialsForMember(ctx, member.ID)
	if err != nil {
		t.Fatalf("listing passkeys: %v", err)
	}
	if len(creds) != 1 {
		t.Errorf("%d passkeys after a reissue, want the one enrolled earlier to survive", len(creds))
	}
}

func TestIssuingALinkForAnUnknownEmailIsRefused(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	_, _, err := svc.IssueEnrolmentLink(ctx, "nobody@example.com")
	if !errors.Is(err, identity.ErrMemberNotFound) {
		t.Fatalf("err = %v, want ErrMemberNotFound", err)
	}

	var n int
	if err := db.Pool().QueryRow(ctx, "select count(*) from enrolment_links").Scan(&n); err != nil {
		t.Fatalf("counting links: %v", err)
	}
	if n != 0 {
		t.Errorf("%d links issued for an unregistered address", n)
	}
}

// Expired, spent, replaced and never-issued are one outcome, and a caller must
// not be able to tell them apart.
func TestUnusableLinksAreIndistinguishable(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	_, live, err := svc.RegisterMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	_, replaced, err := svc.RegisterMember(ctx, "grace@example.com", "Grace", "Hopper")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	if _, _, err := svc.IssueEnrolmentLink(ctx, "grace@example.com"); err != nil {
		t.Fatalf("replacing a link: %v", err)
	}

	_, expiredMember, err := svc.RegisterMember(ctx, "katherine@example.com", "Katherine", "Johnson")
	if err != nil {
		t.Fatalf("registering: %v", err)
	}
	if _, err := db.Pool().Exec(ctx,
		`update enrolment_links set expires_at = now() - interval '1 second'
		 where member_id = (select id from members where email = 'katherine@example.com')`); err != nil {
		t.Fatalf("expiring a link: %v", err)
	}

	// Spend one properly, so the fourth case is a link that did enrol.
	ticket, _, err := svc.RedeemEnrolmentLink(ctx, linkValue(t, live))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}
	if ticket.Value == "" {
		t.Fatal("redeeming produced no ticket")
	}
	if _, err := db.Pool().Exec(ctx,
		`update enrolment_links set enrolled_at = now()
		 where member_id = (select id from members where email = 'ada@example.com')`); err != nil {
		t.Fatalf("marking a link enrolled: %v", err)
	}

	for _, tc := range []struct {
		name  string
		value string
	}{
		{"expired", linkValue(t, expiredMember)},
		{"already enrolled", linkValue(t, live)},
		{"replaced", linkValue(t, replaced)},
		{"never issued", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.RedeemEnrolmentLink(ctx, tc.value)
			if !errors.Is(err, identity.ErrEnrolmentLinkUnusable) {
				t.Fatalf("err = %v, want ErrEnrolmentLinkUnusable", err)
			}
			if got := err.Error(); got != identity.ErrEnrolmentLinkUnusable.Error() {
				t.Errorf("message = %q, want the same text for every cause (%q)",
					got, identity.ErrEnrolmentLinkUnusable.Error())
			}
		})
	}
}
