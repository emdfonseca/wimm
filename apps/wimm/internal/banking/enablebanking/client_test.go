package enablebanking

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// recorded is one captured request, so a test can assert what wimm sent rather
// than only what it did with the reply.
type recorded struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   map[string]any
}

// server stands in for the gateway. Handlers are keyed by "METHOD /path".
type server struct {
	t        *testing.T
	requests []recorded
	handlers map[string]func(w http.ResponseWriter, r *http.Request)
}

func newServer(t *testing.T) (*server, *Client) {
	t.Helper()
	s := &server{t: t, handlers: map[string]func(http.ResponseWriter, *http.Request){}}

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recorded{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Header: r.Header.Clone()}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&rec.Body)
		}
		s.requests = append(s.requests, rec)

		if h, ok := s.handlers[r.Method+" "+r.URL.Path]; ok {
			h(w, r)
			return
		}
		w.WriteHeader(http.StatusNotImplemented)
	}))
	t.Cleanup(httpServer.Close)

	client, err := New(Options{
		ApplicationID:  applicationID,
		PrivateKeyPath: writeKey(t, generateKey(t), false, 0o600),
		RedirectURL:    "https://localhost:8765/psd2/callback",
		BaseURL:        httpServer.URL,
		Now:            func() time.Time { return issuedAt },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, client
}

func (s *server) on(route string, status int, body string) {
	s.handlers[route] = func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func (s *server) last() recorded {
	s.t.Helper()
	if len(s.requests) == 0 {
		s.t.Fatal("no request was made")
	}
	return s.requests[len(s.requests)-1]
}

// Every request carries the signed bearer token.
func TestEveryRequestIsSigned(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /aspsps", http.StatusOK, `{"aspsps":[]}`)

	if _, err := c.Banks(context.Background(), "PT"); err != nil {
		t.Fatalf("Banks: %v", err)
	}

	auth := s.last().Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Fatalf("Authorization = %q", auth)
	}
	if strings.Count(strings.TrimPrefix(auth, "Bearer "), ".") != 2 {
		t.Error("the bearer value is not a three-segment JWT")
	}
}

// 4.2: each bank's own maximum reaches the domain type, because the consent
// screen states a real date and the real dates differ by two orders of
// magnitude in one country.
func TestBanksCarryTheirOwnConsentMaximum(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /aspsps", http.StatusOK, `{"aspsps":[
		{"name":"Montepio","country":"PT","maximum_consent_validity":7776000,"psu_types":["personal"]},
		{"name":"ActivoBank","country":"PT","maximum_consent_validity":86400,"psu_types":["personal"]}
	]}`)

	banks, err := c.Banks(context.Background(), "PT")
	if err != nil {
		t.Fatalf("Banks: %v", err)
	}
	if len(banks) != 2 {
		t.Fatalf("got %d banks, want 2", len(banks))
	}
	if banks[0].MaxConsent != 90*24*time.Hour {
		t.Errorf("Montepio MaxConsent = %s, want 2160h", banks[0].MaxConsent)
	}
	if banks[1].MaxConsent != 24*time.Hour {
		t.Errorf("ActivoBank MaxConsent = %s, want 24h", banks[1].MaxConsent)
	}
	if got := s.last().Query; got != "country=PT" {
		t.Errorf("query = %q", got)
	}
}

// A business-only bank produces a hand-off that fails at the bank, so it is not
// offered to a household.
func TestABusinessOnlyBankIsNotOffered(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /aspsps", http.StatusOK, `{"aspsps":[
		{"name":"Corporate","country":"PT","psu_types":["business"]},
		{"name":"Montepio","country":"PT","psu_types":["business","personal"]}
	]}`)

	banks, err := c.Banks(context.Background(), "PT")
	if err != nil {
		t.Fatalf("Banks: %v", err)
	}
	if len(banks) != 1 || banks[0].Name != "Montepio" {
		t.Errorf("got %v, want only Montepio", banks)
	}
}

// The hand-off carries the bank, the scope and the date the consent will run
// out. What the scope now contains is asserted in transactions_test.go, beside
// the read it exists for.
func TestBeginConnectionSendsTheBankTheScopeAndTheDate(t *testing.T) {
	s, c := newServer(t)
	s.on("POST /auth", http.StatusOK, `{"url":"https://bank.example/c","authorization_id":"auth-1"}`)

	validUntil := issuedAt.Add(24 * time.Hour)
	handoff, err := c.BeginConnection(context.Background(), banking.BeginRequest{
		Bank:       banking.Bank{ID: "PT:ActivoBank", Name: "ActivoBank", Country: "PT"},
		ValidUntil: validUntil,
		State:      "state-1",
	})
	if err != nil {
		t.Fatalf("BeginConnection: %v", err)
	}
	if handoff.URL != "https://bank.example/c" || handoff.GatewayRef != "auth-1" {
		t.Errorf("handoff = %+v", handoff)
	}

	body := s.last().Body
	encoded, _ := json.Marshal(body)

	access, ok := body["access"].(map[string]any)
	if !ok {
		t.Fatalf("no access object: %s", encoded)
	}
	if access["balances"] != true {
		t.Error("the request does not ask for balances")
	}
	if got := access["valid_until"]; got != validUntil.UTC().Format(time.RFC3339) {
		t.Errorf("valid_until = %v, want %s", got, validUntil.UTC().Format(time.RFC3339))
	}
	aspsp, _ := body["aspsp"].(map[string]any)
	if aspsp["name"] != "ActivoBank" || aspsp["country"] != "PT" {
		t.Errorf("aspsp = %v", aspsp)
	}
}

// 4.4: every account the session returns reaches the caller, because the
// details are returned once and no endpoint lists them again.
func TestCompleteConnectionReturnsEveryAccount(t *testing.T) {
	s, c := newServer(t)
	s.on("POST /sessions", http.StatusOK, `{
		"session_id":"sess-1",
		"access":{"valid_until":"2026-12-16T09:00:00Z"},
		"accounts":[
			{"uid":"uid-1","identification_hash":"hash-1","product":"Conta à Ordem","name":"Ada Lovelace",
			 "currency":"EUR","cash_account_type":"CACC","account_id":{"iban":"PT50003601089910006580538"}},
			{"uid":"uid-2","identification_hash":"hash-2","product":"Poupança","name":"Ada Lovelace",
			 "currency":"EUR","cash_account_type":"SVGS","account_id":{"iban":"PT50002300004554466045594"}}
		]}`)

	conn, accounts, err := c.CompleteConnection(context.Background(),
		banking.PendingConnection{State: "s"}, banking.Callback{Code: "code", State: "s"})
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}
	if conn.GatewayRef != "sess-1" {
		t.Errorf("GatewayRef = %q", conn.GatewayRef)
	}
	if len(accounts) != 2 {
		t.Fatalf("got %d accounts, want both", len(accounts))
	}

	// Identity is the cross-session hash, never the per-session uid.
	if accounts[0].Ref != "hash-1" {
		t.Errorf("Ref = %q, want the identification_hash", accounts[0].Ref)
	}
	if accounts[0].GatewayUID != "uid-1" {
		t.Errorf("GatewayUID = %q", accounts[0].GatewayUID)
	}
	// Only the suffix is kept.
	if accounts[0].NumberSuffix != "0538" {
		t.Errorf("NumberSuffix = %q, want 0538", accounts[0].NumberSuffix)
	}
	for _, a := range accounts {
		if strings.Contains(a.NumberSuffix, "PT50") {
			t.Errorf("the full IBAN was kept: %q", a.NumberSuffix)
		}
	}
}

func TestDecliningAtTheBankNeedsNoRequest(t *testing.T) {
	s, c := newServer(t)

	_, _, err := c.CompleteConnection(context.Background(),
		banking.PendingConnection{State: "s"}, banking.Callback{Error: "ACCESS_DENIED", State: "s"})
	if !errors.Is(err, banking.ErrConsentDeclined) {
		t.Fatalf("got %v, want ErrConsentDeclined", err)
	}
	if len(s.requests) != 0 {
		t.Error("a declined consent still called the gateway")
	}
}

func TestAMismatchedCallbackStateIsRefused(t *testing.T) {
	s, c := newServer(t)

	_, _, err := c.CompleteConnection(context.Background(),
		banking.PendingConnection{State: "expected"}, banking.Callback{Code: "c", State: "forged"})
	if err == nil {
		t.Fatal("accepted a callback whose state does not match")
	}
	if len(s.requests) != 0 {
		t.Error("a mismatched state still called the gateway")
	}
}

// 4.5: PSU headers on every member-initiated read. Their presence is what
// exempts the call from the background-fetch rate limit.
func TestBalancesSendsPSUHeaders(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /accounts/uid-1/balances", http.StatusOK, `{"balances":[
		{"balance_type":"CLAV","balance_amount":{"amount":"4200.10","currency":"EUR"}}
	]}`)

	balances, err := c.Balances(context.Background(),
		banking.Connection{GatewayRef: "sess-1"},
		banking.Account{Ref: "hash-1", GatewayUID: "uid-1", Currency: "EUR"})
	if err != nil {
		t.Fatalf("Balances: %v", err)
	}

	if balances[0].Money.Minor != 420_010 {
		t.Errorf("Minor = %d, want 420010", balances[0].Money.Minor)
	}
	if balances[0].ReadAt.IsZero() {
		t.Error("a balance with no read time is not a balance in this system")
	}
	for _, h := range []string{"PSU-IP-Address", "PSU-User-Agent"} {
		if s.last().Header.Get(h) == "" {
			t.Errorf("%s is absent, so the background rate limit applies", h)
		}
	}
}

// The bank list is not a member-initiated read of their data, so it does not
// claim to be one.
func TestTheBankListDoesNotClaimAMemberIsPresent(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /aspsps", http.StatusOK, `{"aspsps":[]}`)

	if _, err := c.Banks(context.Background(), "PT"); err != nil {
		t.Fatalf("Banks: %v", err)
	}
	if s.last().Header.Get("PSU-IP-Address") != "" {
		t.Error("the bank list asserts a member is present")
	}
}

func TestEndConnectionDeletesTheSession(t *testing.T) {
	s, c := newServer(t)
	s.on("DELETE /sessions/sess-1", http.StatusOK, `{}`)

	if err := c.EndConnection(context.Background(), banking.Connection{GatewayRef: "sess-1"}); err != nil {
		t.Fatalf("EndConnection: %v", err)
	}
	if s.last().Method != http.MethodDelete {
		t.Errorf("method = %s", s.last().Method)
	}
}

// A session the gateway has already forgotten is a disconnection that already
// happened. Reporting it would leave a member unable to remove a dead bank.
func TestEndConnectionTreatsAnAlreadyGoneSessionAsDone(t *testing.T) {
	s, c := newServer(t)
	s.on("DELETE /sessions/sess-1", http.StatusUnauthorized, `{"code":"EXPIRED_SESSION"}`)

	if err := c.EndConnection(context.Background(), banking.Connection{GatewayRef: "sess-1"}); err != nil {
		t.Errorf("EndConnection: %v", err)
	}
}

// 4.6: the whole mapping, table driven, fed the responses the gateway actually
// sends.
func TestFailuresMapOntoTheTaxonomy(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		body    string
		retry   string
		want    error
		seconds time.Duration
	}{
		{"an expired session", http.StatusUnauthorized, `{"code":"EXPIRED_SESSION"}`, "", banking.ErrConsentExpired, 0},
		{"the ASPSP rate limit", http.StatusTooManyRequests, `{"code":"ASPSP_RATE_LIMIT_EXCEEDED"}`, "21600", banking.ErrRateLimited, 6 * time.Hour},
		{"a rate limit with no retry-after", http.StatusTooManyRequests, `{"code":"ASPSP_RATE_LIMIT_EXCEEDED"}`, "", banking.ErrRateLimited, 0},
		{"a bare 429", http.StatusTooManyRequests, `{}`, "90", banking.ErrRateLimited, 90 * time.Second},
		{"a malformed retry-after", http.StatusTooManyRequests, `{}`, "next tuesday", banking.ErrRateLimited, 0},
		{"the member declining", http.StatusForbidden, `{"code":"ACCESS_DENIED"}`, "", banking.ErrConsentDeclined, 0},
		{"the bank erroring", http.StatusBadGateway, `{"code":"ASPSP_ERROR"}`, "", banking.ErrBankUnavailable, 0},
		{"a bare 500 from the gateway", http.StatusInternalServerError, `{}`, "", banking.ErrBankUnavailable, 0},
		{"no accounts", http.StatusNotFound, `{"code":"NO_ACCOUNTS"}`, "", banking.ErrNoAccounts, 0},
		{"a bare 404", http.StatusNotFound, `{}`, "", banking.ErrConsentExpired, 0},
		{"a bad request", http.StatusBadRequest, `{}`, "", banking.ErrGatewayUnavailable, 0},
		{"an HTML error page", http.StatusBadGateway, `<html>gateway down</html>`, "", banking.ErrBankUnavailable, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, c := newServer(t)
			s.handlers["GET /aspsps"] = func(w http.ResponseWriter, _ *http.Request) {
				if tc.retry != "" {
					w.Header().Set("Retry-After", tc.retry)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}

			_, err := c.Banks(context.Background(), "PT")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}

			if errors.Is(tc.want, banking.ErrRateLimited) {
				var limited *banking.RateLimitError
				if !errors.As(err, &limited) {
					t.Fatal("no retry-after reachable")
				}
				if limited.RetryAfter != tc.seconds {
					t.Errorf("RetryAfter = %s, want %s", limited.RetryAfter, tc.seconds)
				}
			}
		})
	}
}

// A failure must not carry the gateway's prose upward: it is another
// vocabulary, and everything the caller routes on is the kind.
func TestAFailureCarriesNoGatewayProse(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /aspsps", http.StatusBadGateway,
		`{"code":"ASPSP_ERROR","message":"Das Kreditinstitut ist nicht erreichbar"}`)

	_, err := c.Banks(context.Background(), "PT")
	if err == nil {
		t.Fatal("no error")
	}
	if strings.Contains(err.Error(), "Kreditinstitut") {
		t.Errorf("the gateway's message reached wimm's error: %v", err)
	}
}

func TestAnUnreachableGatewayIsNotABankFailure(t *testing.T) {
	c, err := New(Options{
		ApplicationID:  applicationID,
		PrivateKeyPath: writeKey(t, generateKey(t), false, 0o600),
		RedirectURL:    "https://localhost:8765/psd2/callback",
		// A port nothing is listening on.
		BaseURL: "http://127.0.0.1:1",
		Now:     func() time.Time { return issuedAt },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = c.Banks(context.Background(), "PT")
	if !errors.Is(err, banking.ErrGatewayUnavailable) {
		t.Errorf("got %v, want ErrGatewayUnavailable", err)
	}
}

func TestNewRefusesIncompleteOptions(t *testing.T) {
	key := writeKey(t, generateKey(t), false, 0o600)
	for _, tc := range []struct {
		name string
		opts Options
		want string
	}{
		{"no redirect url", Options{ApplicationID: applicationID, PrivateKeyPath: key}, "no redirect url"},
		{"no application id", Options{PrivateKeyPath: key, RedirectURL: "https://x"}, "no application id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New(tc.opts); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want %q", err, tc.want)
			}
		})
	}
}

// The response that broke a real connection, as Montepio actually sends it.
//
// `account_id.other` is an object, not a string. Declaring it as a string
// decoded every IBAN-bearing account correctly and failed the entire
// connection on the first card account — and because no banking error was
// mapped at the time, the member saw "something went wrong" and the log said
// `code: internal` with nothing else.
//
// Recorded rather than derived: every other fixture in this file comes from
// Enable Banking's published reference, which is what let this through.
func TestAnAccountIdentifiedByACardRatherThanAnIBAN(t *testing.T) {
	s, c := newServer(t)
	s.on("POST /sessions", http.StatusOK, `{
		"session_id":"sess-1",
		"access":{"valid_until":"2026-12-16T09:00:00Z"},
		"accounts":[
			{"uid":"uid-1","identification_hash":"hash-1","product":"Conta Ordem",
			 "name":"Ada Lovelace","currency":"EUR","cash_account_type":"CACC",
			 "account_id":{"iban":"PT50003601089910006580538"}},
			{"uid":"uid-2","identification_hash":"hash-2","product":"Cartão",
			 "name":"Ada Lovelace","currency":"EUR","cash_account_type":"CARD",
			 "account_id":{"other":{"identification":"422240******6438",
			  "scheme_name":"CPAN","issuer":"Montepio"}}}
		]}`)

	conn, accounts, err := c.CompleteConnection(context.Background(),
		banking.PendingConnection{State: "s"}, banking.Callback{Code: "code", State: "s"})
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}
	if conn.GatewayRef != "sess-1" {
		t.Errorf("GatewayRef = %q", conn.GatewayRef)
	}
	if len(accounts) != 2 {
		t.Fatalf("got %d accounts, want both — one account with no IBAN must not lose the other", len(accounts))
	}

	// The IBAN account is unchanged.
	if accounts[0].NumberSuffix != "0538" {
		t.Errorf("IBAN account suffix = %q, want 0538", accounts[0].NumberSuffix)
	}

	// The card account takes its suffix from the generic identification, and
	// keeps only the trailing characters — the masked digits are what tell two
	// cards apart, and the rest is never stored.
	if accounts[1].NumberSuffix != "6438" {
		t.Errorf("card account suffix = %q, want 6438", accounts[1].NumberSuffix)
	}
	if strings.Contains(accounts[1].NumberSuffix, "422240") {
		t.Errorf("the leading digits were kept: %q", accounts[1].NumberSuffix)
	}
	if accounts[1].Ref != "hash-2" {
		t.Errorf("Ref = %q, want the cross-session hash", accounts[1].Ref)
	}
}

// An account with neither an IBAN nor a generic identification still arrives:
// a bank that names nothing is not a reason to lose the account.
func TestAnAccountWithNoIdentifierAtAll(t *testing.T) {
	s, c := newServer(t)
	s.on("POST /sessions", http.StatusOK, `{
		"session_id":"sess-1",
		"accounts":[
			{"uid":"uid-1","identification_hash":"hash-1","product":"Poupança",
			 "currency":"EUR","account_id":{}}
		]}`)

	_, accounts, err := c.CompleteConnection(context.Background(),
		banking.PendingConnection{State: "s"}, banking.Callback{Code: "c", State: "s"})
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got %d accounts, want 1", len(accounts))
	}
	if accounts[0].NumberSuffix != "" {
		t.Errorf("NumberSuffix = %q, want empty rather than invented", accounts[0].NumberSuffix)
	}
}
