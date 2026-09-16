package rpc_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	identityv1 "github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1"
	"github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1/identityv1connect"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
	"github.com/xuuid/wimm/apps/wimm/internal/identity"
	"github.com/xuuid/wimm/apps/wimm/internal/rpc"
	"github.com/xuuid/wimm/apps/wimm/internal/store/storetest"
)

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewJSONHandler(&testWriter{t: t}, nil))
}

type testWriter struct {
	t  *testing.T
	mu sync.Mutex
}

func (w *testWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.t.Log(strings.TrimSpace(string(p)))
	return len(p), nil
}

// An enrolment link is a bearer credential: whoever reads it becomes that
// member. A log line carrying one hands the account to whoever reads the logs.
func TestNoLogLineCarriesAnEnrolmentLinkValue(t *testing.T) {
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

	var logged bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logged, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mux := http.NewServeMux()
	mux.Handle(rpc.PublicHandler(log, rpc.NewPublicServer(svc, cfg.Origins)))
	mux.Handle(rpc.OperatorHandler(log, rpc.NewOperatorServer(svc), testCredential))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ctx := context.Background()
	operator := identityv1connect.NewOperatorServiceClient(srv.Client(), srv.URL)
	public := identityv1connect.NewPublicServiceClient(srv.Client(), srv.URL)

	registered, err := operator.RegisterMember(ctx,
		withCredential(registerRequest("ada@example.com", "Ada", "Lovelace"), testCredential))
	if err != nil {
		t.Fatalf("registering: %v", err)
	}

	link := registered.Msg.GetEnrolmentLink().GetUrl()
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parsing the enrolment URL: %v", err)
	}
	value := strings.TrimPrefix(u.Path, identity.EnrolmentPath)
	if value == "" {
		t.Fatal("the enrolment URL carries no value")
	}

	redeemed, err := public.RedeemEnrolmentLink(ctx,
		connect.NewRequest(&identityv1.RedeemEnrolmentLinkRequest{LinkValue: value}))
	if err != nil {
		t.Fatalf("redeeming: %v", err)
	}

	// A never-issued value too, so the refusal path is covered as well.
	_, _ = public.RedeemEnrolmentLink(ctx,
		connect.NewRequest(&identityv1.RedeemEnrolmentLinkRequest{LinkValue: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}))

	out := logged.String()
	if out == "" {
		t.Fatal("nothing was logged at all, so this proves nothing")
	}
	for _, secret := range []struct{ name, value string }{
		{"the enrolment link value", value},
		{"the enrolment URL", link},
		{"the enrolment ticket", redeemed.Msg.GetEnrolmentTicket()},
	} {
		if secret.value != "" && strings.Contains(out, secret.value) {
			t.Errorf("%s reached the log", secret.name)
		}
	}
	// The log does have to be useful, or the assertion above is satisfied by
	// logging nothing.
	if !strings.Contains(out, "RedeemEnrolmentLink") {
		t.Error("the log does not record that RedeemEnrolmentLink was called")
	}
}
