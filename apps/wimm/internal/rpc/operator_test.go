package rpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	identityv1 "github.com/emdfonseca/wimm/packages/contracts/gen/go/wimm/identity/v1"
	"github.com/emdfonseca/wimm/packages/contracts/gen/go/wimm/identity/v1/identityv1connect"

	"github.com/emdfonseca/wimm/apps/wimm/internal/config"
	"github.com/emdfonseca/wimm/apps/wimm/internal/identity"
	"github.com/emdfonseca/wimm/apps/wimm/internal/rpc"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store/storetest"
)

const testCredential = "operator-credential-under-test"

// operatorFixture stands the real handler behind an httptest server and calls
// it through the generated client. No mock of the transport (.claude/rules/go.md).
type operatorFixture struct {
	db     *store.DB
	client identityv1connect.OperatorServiceClient
}

func newOperatorFixture(t *testing.T) *operatorFixture {
	t.Helper()

	db := storetest.New(t)

	cfg, err := config.Load(func(k string) string {
		return map[string]string{
			"WIMM_OPERATOR_CREDENTIAL": testCredential,
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

	mux := http.NewServeMux()
	mux.Handle(rpc.OperatorHandler(testLogger(t), rpc.NewOperatorServer(svc), testCredential))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return &operatorFixture{
		db:     db,
		client: identityv1connect.NewOperatorServiceClient(srv.Client(), srv.URL),
	}
}

func registerRequest(email, first, last string) *connect.Request[identityv1.RegisterMemberRequest] {
	return connect.NewRequest(&identityv1.RegisterMemberRequest{
		Email:     &identityv1.EmailAddress{Value: email},
		FirstName: first,
		LastName:  last,
	})
}

func withCredential[T any](req *connect.Request[T], credential string) *connect.Request[T] {
	req.Header().Set(rpc.OperatorCredentialHeader, "Bearer "+credential)
	return req
}

func (f *operatorFixture) memberCount(t *testing.T) int {
	t.Helper()
	var n int
	if err := f.db.Pool().QueryRow(context.Background(), "select count(*) from members").Scan(&n); err != nil {
		t.Fatalf("counting members: %v", err)
	}
	return n
}

func (f *operatorFixture) linkCount(t *testing.T) int {
	t.Helper()
	var n int
	if err := f.db.Pool().QueryRow(context.Background(), "select count(*) from enrolment_links").Scan(&n); err != nil {
		t.Fatalf("counting enrolment links: %v", err)
	}
	return n
}

// An operator surface that answers without the credential is the failure this
// whole separation exists to prevent, so both ways of getting it wrong are
// asserted, and so is the absence of any side effect.
func TestOperatorCredentialIsRequired(t *testing.T) {
	for _, tc := range []struct {
		name   string
		header func(*connect.Request[identityv1.RegisterMemberRequest]) *connect.Request[identityv1.RegisterMemberRequest]
	}{
		{"absent", func(r *connect.Request[identityv1.RegisterMemberRequest]) *connect.Request[identityv1.RegisterMemberRequest] {
			return r
		}},
		{"wrong", func(r *connect.Request[identityv1.RegisterMemberRequest]) *connect.Request[identityv1.RegisterMemberRequest] {
			return withCredential(r, "not-the-credential")
		}},
		{"empty bearer", func(r *connect.Request[identityv1.RegisterMemberRequest]) *connect.Request[identityv1.RegisterMemberRequest] {
			return withCredential(r, "")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newOperatorFixture(t)

			_, err := f.client.RegisterMember(context.Background(),
				tc.header(registerRequest("ada@example.com", "Ada", "Lovelace")))
			if err == nil {
				t.Fatal("the call was accepted without the configured operator credential")
			}
			if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
				t.Errorf("code = %v, want %v", got, connect.CodeUnauthenticated)
			}

			if n := f.memberCount(t); n != 0 {
				t.Errorf("%d members were registered by a refused call", n)
			}
			if n := f.linkCount(t); n != 0 {
				t.Errorf("%d enrolment links were issued by a refused call", n)
			}
		})
	}
}

func TestOperatorCredentialAccepted(t *testing.T) {
	f := newOperatorFixture(t)

	res, err := f.client.RegisterMember(context.Background(),
		withCredential(registerRequest("ada@example.com", "Ada", "Lovelace"), testCredential))
	if err != nil {
		t.Fatalf("registering with the configured credential: %v", err)
	}
	if res.Msg.GetMember().GetFirstName() != "Ada" {
		t.Errorf("member = %v", res.Msg.GetMember())
	}
}
