package rpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	identityv1 "github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1"
	"github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1/identityv1connect"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
	"github.com/xuuid/wimm/apps/wimm/internal/identity"
	"github.com/xuuid/wimm/apps/wimm/internal/identity/authenticatortest"
	"github.com/xuuid/wimm/apps/wimm/internal/rpc"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
	"github.com/xuuid/wimm/apps/wimm/internal/store/storetest"
)

type publicFixture struct {
	db       *store.DB
	svc      *identity.Service
	client   identityv1connect.PublicServiceClient
	operator identityv1connect.OperatorServiceClient
}

func newPublicFixture(t *testing.T, origins string) *publicFixture {
	t.Helper()

	db := storetest.New(t)

	cfg, err := config.Load(func(k string) string {
		return map[string]string{
			"WIMM_OPERATOR_CREDENTIAL": testCredential,
			"WIMM_DATABASE_URL":        storetest.URL(t),
			"WIMM_BASE_URL":            "http://localhost:9466",
			"WIMM_ORIGINS":             origins,
		}[k]
	})
	if err != nil {
		t.Fatalf("building test configuration: %v", err)
	}

	svc, err := identity.New(db, cfg)
	if err != nil {
		t.Fatalf("building the identity service: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle(rpc.PublicHandler(testLogger(t), rpc.NewPublicServer(svc, cfg.Origins)))
	mux.Handle(rpc.OperatorHandler(testLogger(t), rpc.NewOperatorServer(svc), testCredential))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return &publicFixture{
		db:       db,
		svc:      svc,
		client:   identityv1connect.NewPublicServiceClient(srv.Client(), srv.URL),
		operator: identityv1connect.NewOperatorServiceClient(srv.Client(), srv.URL),
	}
}

func (f *publicFixture) register(t *testing.T, email, first, last string) string {
	t.Helper()

	res, err := f.operator.RegisterMember(context.Background(),
		withCredential(registerRequest(email, first, last), testCredential))
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	u, err := url.Parse(res.Msg.GetEnrolmentLink().GetUrl())
	if err != nil {
		t.Fatalf("parsing the enrolment URL: %v", err)
	}
	return strings.TrimPrefix(u.Path, identity.EnrolmentPath)
}

// The session cookie has to be unreadable by script, must not travel
// cross-site, and must be marked Secure exactly where the origin is HTTPS —
// a Secure cookie on http://localhost is simply never stored.
func TestSessionCookieAttributes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		origins    string
		wantSecure bool
	}{
		{"http origin", "http://localhost:9466", false},
		{"https origin", "https://wimm.example", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPublicFixture(t, tc.origins)
			ctx := context.Background()

			value := f.register(t, "ada@example.com", "Ada", "Lovelace")
			res, err := f.client.RedeemEnrolmentLink(ctx,
				connect.NewRequest(&identityv1.RedeemEnrolmentLinkRequest{LinkValue: value}))
			if err != nil {
				t.Fatalf("redeeming: %v", err)
			}

			cookie := parseSetCookie(t, res.Header(), rpc.EnrolmentCookie)
			if !cookie.HttpOnly {
				t.Error("the enrolment cookie is not HttpOnly")
			}
			if cookie.SameSite != http.SameSiteLaxMode {
				t.Errorf("SameSite = %v, want Lax", cookie.SameSite)
			}
			if cookie.Secure != tc.wantSecure {
				t.Errorf("Secure = %v, want %v for origins %q", cookie.Secure, tc.wantSecure, tc.origins)
			}
		})
	}
}

func TestEnrolmentSetsASessionCookie(t *testing.T) {
	f := newPublicFixture(t, "https://wimm.example")
	ctx := context.Background()

	value := f.register(t, "ada@example.com", "Ada", "Lovelace")
	redeemed, err := f.client.RedeemEnrolmentLink(ctx,
		connect.NewRequest(&identityv1.RedeemEnrolmentLinkRequest{LinkValue: value}))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}

	begun, err := f.client.BeginEnrolment(ctx,
		connect.NewRequest(&identityv1.BeginEnrolmentRequest{
			EnrolmentTicket: redeemed.Msg.GetEnrolmentTicket(),
		}))
	if err != nil {
		t.Fatalf("beginning enrolment: %v", err)
	}

	auth := authenticatortest.New(t)
	finished, err := f.client.FinishEnrolment(ctx,
		connect.NewRequest(&identityv1.FinishEnrolmentRequest{
			CeremonyId:     begun.Msg.GetCeremonyId(),
			CredentialJson: auth.Register(t, begun.Msg.GetCreationOptionsJson(), "https://wimm.example"),
		}))
	if err != nil {
		t.Fatalf("finishing enrolment: %v", err)
	}
	if finished.Msg.GetMember().GetFirstName() != "Ada" {
		t.Errorf("member = %v", finished.Msg.GetMember())
	}

	session := parseSetCookie(t, finished.Header(), rpc.SessionCookie)
	if !session.HttpOnly || session.SameSite != http.SameSiteLaxMode || !session.Secure {
		t.Errorf("session cookie = %+v, want HttpOnly, SameSite=Lax and Secure", session)
	}
	if session.Value == "" {
		t.Error("the session cookie carries no value")
	}

	// The enrolment ticket is spent, so it is cleared rather than left behind.
	cleared := parseSetCookie(t, finished.Header(), rpc.EnrolmentCookie)
	if cleared.MaxAge >= 0 {
		t.Errorf("the enrolment cookie was not cleared: %+v", cleared)
	}
}

// Expired, spent, replaced and never-issued must be one response. A caller who
// can tell them apart can probe for which links ever existed.
func TestUnusableLinksAreOneResponse(t *testing.T) {
	f := newPublicFixture(t, "http://localhost:9466")
	ctx := context.Background()

	expired := f.register(t, "katherine@example.com", "Katherine", "Johnson")
	if _, err := f.db.Pool().Exec(ctx,
		`update enrolment_links set expires_at = now() - interval '1 second'`); err != nil {
		t.Fatalf("expiring a link: %v", err)
	}

	replaced := f.register(t, "grace@example.com", "Grace", "Hopper")
	if _, err := f.operator.IssueEnrolmentLink(ctx,
		withCredential(issueRequest("grace@example.com"), testCredential)); err != nil {
		t.Fatalf("replacing a link: %v", err)
	}

	spent := f.register(t, "ada@example.com", "Ada", "Lovelace")
	if _, err := f.db.Pool().Exec(ctx,
		`update enrolment_links set enrolled_at = now()
		 where member_id = (select id from members where email = 'ada@example.com')`); err != nil {
		t.Fatalf("marking a link enrolled: %v", err)
	}

	cases := map[string]string{
		"expired":      expired,
		"replaced":     replaced,
		"spent":        spent,
		"never issued": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}

	type outcome struct {
		code    connect.Code
		message string
	}
	seen := map[string]outcome{}
	timings := map[string]time.Duration{}

	for name, value := range cases {
		start := time.Now()
		_, err := f.client.RedeemEnrolmentLink(ctx,
			connect.NewRequest(&identityv1.RedeemEnrolmentLinkRequest{LinkValue: value}))
		timings[name] = time.Since(start)

		if err == nil {
			t.Fatalf("%s: the link was accepted", name)
		}
		seen[name] = outcome{code: connect.CodeOf(err), message: err.Error()}
	}

	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)

	first := seen[names[0]]
	for _, n := range names[1:] {
		if seen[n] != first {
			t.Errorf("%s answered %v %q, but %s answered %v %q — the four causes are distinguishable",
				n, seen[n].code, seen[n].message, names[0], first.code, first.message)
		}
	}

	// Timing is the third channel. The bound is loose on purpose: this is here
	// to catch a case doing visibly more work, such as an extra query or a
	// hash, not to measure the network.
	var slowest, fastest time.Duration
	for _, d := range timings {
		if slowest == 0 || d > slowest {
			slowest = d
		}
		if fastest == 0 || d < fastest {
			fastest = d
		}
	}
	if fastest > 0 && slowest > 10*fastest {
		t.Errorf("the slowest case took %s and the fastest %s: %v", slowest, fastest, timings)
	}
}

func parseSetCookie(t *testing.T, header http.Header, name string) *http.Cookie {
	t.Helper()

	for _, raw := range header.Values("Set-Cookie") {
		parsed, err := http.ParseSetCookie(raw)
		if err != nil {
			t.Fatalf("parsing Set-Cookie %q: %v", raw, err)
		}
		if parsed.Name == name {
			return parsed
		}
	}
	t.Fatalf("no %s cookie in %v", name, header.Values("Set-Cookie"))
	return nil
}
