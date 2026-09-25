package rpc_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	bankingv1 "github.com/emdfonseca/wimm/packages/contracts/gen/go/wimm/banking/v1"
	"github.com/emdfonseca/wimm/packages/contracts/gen/go/wimm/banking/v1/bankingv1connect"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/banking/bankingtest"
	"github.com/emdfonseca/wimm/apps/wimm/internal/config"
	"github.com/emdfonseca/wimm/apps/wimm/internal/identity"
	"github.com/emdfonseca/wimm/apps/wimm/internal/rpc"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store/storetest"
)

const sessionValue = "session-value-under-test"

// ledgerFixture is a member signed in over a real banking handler, owning one
// account whose ledger holds money in and money out across two months.
type ledgerFixture struct {
	client    bankingv1connect.BankingServiceClient
	accountID string
}

func newLedgerFixture(t *testing.T) ledgerFixture {
	t.Helper()
	ctx := context.Background()
	db := storetest.New(t)

	cfg, err := config.Load(func(k string) string {
		return map[string]string{
			"WIMM_OPERATOR_CREDENTIAL": testCredential,
			"WIMM_DATABASE_URL":        storetest.URL(t),
			"WIMM_BASE_URL":            "http://localhost:9466",
			"WIMM_ORIGINS":             "http://localhost:9466",
		}[k]
	})
	if err != nil {
		t.Fatalf("building test configuration: %v", err)
	}
	ids, err := identity.New(db, cfg)
	if err != nil {
		t.Fatalf("building the identity service: %v", err)
	}
	keys, err := banking.NewKeyring([]banking.Key{{ID: "k1", Material: bytes.Repeat([]byte{3}, banking.KeySize)}})
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	svc := banking.NewService(db, bankingtest.New(), keys, testLogger(t),
		"http://localhost:9466/psd2/callback", 24*time.Hour, nil,
		banking.LedgerOptions{Overlap: 7 * 24 * time.Hour, SyncInterval: time.Hour, MaxPages: 3, PageSize: 50})

	ada, err := db.CreateMember(ctx, "ada@example.com", "Ada", "Lovelace")
	if err != nil {
		t.Fatalf("CreateMember: %v", err)
	}
	hash := sha256.Sum256([]byte(sessionValue))
	if _, err := db.CreateSession(ctx, ada.ID, hash[:], time.Hour); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	now, err := db.Now(ctx)
	if err != nil {
		t.Fatalf("Now: %v", err)
	}
	_, stored, err := db.CreateBankConnection(ctx, store.BankConnection{
		Gateway: "enablebanking", GatewayRef: store.Sealed{Ciphertext: []byte("sealed"), KeyID: "k1"},
		BankID: "PT:Montepio", BankName: "Montepio", ConnectedBy: ada.ID,
		ConsentExpiresAt: now.T.Add(90 * 24 * time.Hour),
	}, []store.Account{{
		GatewayRef: "hash-1", GatewayUID: store.Sealed{Ciphertext: []byte("uid"), KeyID: "k1"},
		Name: "Conta à Ordem", NumberSuffix: "0538", AccountType: "CACC", Currency: "EUR",
	}}, ada.ID)
	if err != nil {
		t.Fatalf("CreateBankConnection: %v", err)
	}
	on := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID, []store.Transaction{
		{Status: store.StatusBooked, DedupKey: "a", AmountMinor: -4500, Currency: "EUR", BookingDate: on(time.August, 31), CounterpartyName: "GALP ENERGIA"},
		{Status: store.StatusBooked, DedupKey: "b", AmountMinor: 250000, Currency: "EUR", BookingDate: on(time.August, 25), CounterpartyName: "Empresa"},
		{Status: store.StatusBooked, DedupKey: "c", AmountMinor: -3000, Currency: "EUR", BookingDate: on(time.July, 3), CounterpartyName: "Galp"},
	}, on(time.September, 1)); err != nil {
		t.Fatalf("WriteAccountTransactions: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle(rpc.BankingHandler(testLogger(t), rpc.NewBankingServer(svc, ids)))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return ledgerFixture{
		client:    bankingv1connect.NewBankingServiceClient(srv.Client(), srv.URL),
		accountID: stored[0].ID,
	}
}

func (f ledgerFixture) list(msg *bankingv1.ListTransactionsRequest) (*bankingv1.Ledger, error) {
	msg.SkipSync = true
	req := connect.NewRequest(msg)
	req.Header().Set("Cookie", rpc.SessionCookie+"="+sessionValue)
	res, err := f.client.ListTransactions(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res.Msg.GetLedger(), nil
}

func TestAFilterWimmCannotUseIsRefused(t *testing.T) {
	f := newLedgerFixture(t)

	for _, tc := range []struct {
		name string
		msg  *bankingv1.ListTransactionsRequest
	}{
		{"a month that is not YYYY-MM", &bankingv1.ListTransactionsRequest{Month: "2026-13"}},
		{"a month written another way", &bankingv1.ListTransactionsRequest{Month: "August 2026"}},
		{"a search over 100 characters", &bankingv1.ListTransactionsRequest{Search: strings.Repeat("a", 101)}},
		{"a direction that does not exist", &bankingv1.ListTransactionsRequest{Direction: bankingv1.LedgerDirection(7)}},
		{"an account id that is not a uuid", &bankingv1.ListTransactionsRequest{AccountId: "not-a-uuid"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.list(tc.msg)
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v (%v), want %v", got, err, connect.CodeInvalidArgument)
			}
		})
	}
}

func TestAFilteredReadCarriesTheAccountsAndMonthsOnOffer(t *testing.T) {
	f := newLedgerFixture(t)

	ledger, err := f.list(&bankingv1.ListTransactionsRequest{
		AccountId: f.accountID,
		Search:    " galp ",
		Month:     "2026-08",
		Direction: bankingv1.LedgerDirection_LEDGER_DIRECTION_OUT,
	})
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if n := len(ledger.GetTransactions()); n != 1 || ledger.GetTotalCount() != 1 {
		t.Errorf("listed %d with a count of %d, want August's one Galp payment", n, ledger.GetTotalCount())
	}
	accounts := ledger.GetFilterAccounts()
	if len(accounts) != 1 || accounts[0].GetAccountId() != f.accountID ||
		accounts[0].GetName() != "Conta à Ordem" || accounts[0].GetBankName() != "Montepio" {
		t.Errorf("filter_accounts = %v, want the one account with its bank", accounts)
	}
	if got := ledger.GetTotals(); len(got) != 1 || got[0].GetCurrency() != "EUR" ||
		got[0].GetMoneyIn().GetMinor() != 0 || got[0].GetMoneyOut().GetMinor() != -4500 {
		t.Errorf("totals = %v, want EUR money out of −4500 and nothing in", got)
	}
	// The months follow every filter but the month: the Galp money out is in
	// August and July.
	if got := ledger.GetMonths(); len(got) != 2 || got[0] != "2026-08" || got[1] != "2026-07" {
		t.Errorf("months = %v, want [2026-08 2026-07]", got)
	}
}
