package enablebanking

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

func ledgerAccount() banking.Account {
	return banking.Account{Ref: "hash-1", GatewayUID: "uid-1", Currency: "EUR"}
}

// Recorded response bodies, kept whole rather than minimised: the shapes this
// adapter gets wrong are the ones a trimmed fixture no longer has.
const (
	// A debit with an entry reference, and a credit without one — the field is
	// optional in the schema, and the digest fallback exists because of it.
	onePage = `{"transactions":[
		{"entry_reference":"ref-a","transaction_id":"txn-a",
		 "transaction_amount":{"amount":"11.90","currency":"EUR"},
		 "credit_debit_indicator":"DBIT","status":"BOOK",
		 "booking_date":"2026-03-04","value_date":"2026-03-04","transaction_date":"2026-03-03",
		 "creditor":{"name":"Padaria Ribeiro"},"debtor":{"name":"Ada Lovelace"},
		 "remittance_information":["Pão e café"],
		 "bank_transaction_code":{"description":"Card payment"}},
		{"transaction_amount":{"amount":"2400.00","currency":"EUR"},
		 "credit_debit_indicator":"CRDT","status":"BOOK",
		 "booking_date":"2026-03-01","value_date":"2026-03-01",
		 "debtor":{"name":"Acme Lda"},
		 "remittance_information":["Salário Março"]},
		{"entry_reference":"ref-c",
		 "transaction_amount":{"amount":"4.50","currency":"EUR"},
		 "credit_debit_indicator":"DBIT","status":"PEND",
		 "booking_date":"2026-03-05",
		 "creditor":{"name":"Metro"}}
	]}`

	firstOfTwo = `{"transactions":[
		{"entry_reference":"ref-a","transaction_amount":{"amount":"1.00","currency":"EUR"},
		 "credit_debit_indicator":"DBIT","status":"BOOK","booking_date":"2026-03-04"}
	],"continuation_key":"page-2"}`

	secondOfTwo = `{"transactions":[
		{"entry_reference":"ref-b","transaction_amount":{"amount":"2.00","currency":"EUR"},
		 "credit_debit_indicator":"DBIT","status":"BOOK","booking_date":"2026-03-03"}
	]}`
)

// A zero From is "as far back as this bank goes", which is the one strategy
// that finds the earliest transaction available. A set From is the incremental
// read the changelog recommends for a feed already fetched.
func TestTheStrategyFollowsWhetherAFromWasGiven(t *testing.T) {
	for _, tc := range []struct {
		name   string
		req    banking.TransactionsRequest
		want   map[string]string
		absent []string
	}{
		{
			name:   "a first fill asks for the longest available",
			req:    banking.TransactionsRequest{},
			want:   map[string]string{"strategy": "longest"},
			absent: []string{"date_from"},
		},
		{
			name: "an incremental sync asks from a date",
			req:  banking.TransactionsRequest{From: time.Date(2026, time.February, 25, 0, 0, 0, 0, time.UTC)},
			want: map[string]string{"strategy": "default", "date_from": "2026-02-25"},
		},
		{
			name: "a cursor drives paging",
			req:  banking.TransactionsRequest{Cursor: "page-2"},
			want: map[string]string{"strategy": "longest", "continuation_key": "page-2"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, c := newServer(t)
			s.on("GET /accounts/uid-1/transactions", http.StatusOK, `{"transactions":[]}`)

			if _, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(), tc.req); err != nil {
				t.Fatalf("Transactions: %v", err)
			}

			query := s.last().Query
			for key, value := range tc.want {
				if !strings.Contains(query, key+"="+value) {
					t.Errorf("query %q does not carry %s=%s", query, key, value)
				}
			}
			for _, key := range tc.absent {
				if strings.Contains(query, key+"=") {
					t.Errorf("query %q carries %s, which a first fill must not send", query, key)
				}
			}
		})
	}
}

// One page, read field by field: this is where a wrong sign or a dropped
// counterparty would show up, and both are silent everywhere else.
func TestOnePageIsReadIntoWimmsVocabulary(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /accounts/uid-1/transactions", http.StatusOK, onePage)

	page, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(), banking.TransactionsRequest{})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(page.Transactions) != 3 {
		t.Fatalf("read %d transactions, want 3", len(page.Transactions))
	}
	if page.NextCursor != "" {
		t.Errorf("NextCursor = %q on a response with no continuation key", page.NextCursor)
	}

	debit := page.Transactions[0]
	if debit.Ref != "ref-a" {
		t.Errorf("Ref = %q, want the bank's entry reference", debit.Ref)
	}
	if debit.Amount.Minor != -1190 || debit.Amount.Currency != "EUR" {
		t.Errorf("Amount = %+v, want -1190 EUR: a debit is negative", debit.Amount)
	}
	if debit.Status != banking.StatusBooked {
		t.Errorf("Status = %q, want booked", debit.Status)
	}
	if debit.CounterpartyName != "Padaria Ribeiro" {
		t.Errorf("CounterpartyName = %q, want the creditor on a debit", debit.CounterpartyName)
	}
	if debit.Remittance != "Pão e café" {
		t.Errorf("Remittance = %q", debit.Remittance)
	}
	if want := time.Date(2026, time.March, 4, 0, 0, 0, 0, time.UTC); !debit.BookingDate.Equal(want) {
		t.Errorf("BookingDate = %s, want %s", debit.BookingDate, want)
	}

	credit := page.Transactions[1]
	if credit.Ref != "" {
		t.Errorf("Ref = %q, want empty where the bank gave no entry reference", credit.Ref)
	}
	if credit.Amount.Minor != 240_000 {
		t.Errorf("Amount = %d, want +240000: money arriving is positive", credit.Amount.Minor)
	}
	if credit.CounterpartyName != "Acme Lda" {
		t.Errorf("CounterpartyName = %q, want the debtor on a credit", credit.CounterpartyName)
	}
	if !credit.TransactionDate.IsZero() {
		t.Errorf("TransactionDate = %s, want the zero time where the bank gave none", credit.TransactionDate)
	}

	if page.Transactions[2].Status != banking.StatusPending {
		t.Errorf("PEND read as %q, want pending — the code is PEND, not PDNG", page.Transactions[2].Status)
	}
}

// OTHR is filed as booked: a transaction wimm cannot classify has already moved
// money, and hiding it is worse than filing it.
func TestAnUnclassifiableStatusIsFiledAsBooked(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /accounts/uid-1/transactions", http.StatusOK,
		`{"transactions":[{"entry_reference":"ref-x","transaction_amount":{"amount":"1.00","currency":"EUR"},
		 "credit_debit_indicator":"DBIT","status":"OTHR","booking_date":"2026-03-04"}]}`)

	page, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(), banking.TransactionsRequest{})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if page.Transactions[0].Status != banking.StatusBooked {
		t.Errorf("OTHR read as %q, want booked", page.Transactions[0].Status)
	}
}

// The continuation key is the gateway's handle for the next page within one
// sync. It reaches the caller as an opaque cursor and goes back out unchanged.
func TestTheContinuationKeyDrivesPaging(t *testing.T) {
	s, c := newServer(t)
	s.handlers["GET /accounts/uid-1/transactions"] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("continuation_key") == "page-2" {
			_, _ = w.Write([]byte(secondOfTwo))
			return
		}
		_, _ = w.Write([]byte(firstOfTwo))
	}

	first, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(), banking.TransactionsRequest{})
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if first.NextCursor != "page-2" {
		t.Fatalf("NextCursor = %q, want the continuation key", first.NextCursor)
	}

	second, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(),
		banking.TransactionsRequest{Cursor: first.NextCursor})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if second.NextCursor != "" {
		t.Errorf("NextCursor = %q on the last page, want empty", second.NextCursor)
	}
	if second.Transactions[0].Ref != "ref-b" {
		t.Errorf("the cursor did not advance: got %q", second.Transactions[0].Ref)
	}
}

// A member cannot act on "the window you asked for is unavailable", so it never
// becomes one of their failures: the adapter retries once with the widest
// strategy and surfaces only a second failure.
func TestAWrongPeriodIsRetriedOnceWithTheLongestStrategy(t *testing.T) {
	s, c := newServer(t)

	var strategies []string
	s.handlers["GET /accounts/uid-1/transactions"] = func(w http.ResponseWriter, r *http.Request) {
		strategies = append(strategies, r.URL.Query().Get("strategy"))
		w.Header().Set("Content-Type", "application/json")
		if len(strategies) == 1 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":"WRONG_TRANSACTIONS_PERIOD"}`))
			return
		}
		_, _ = w.Write([]byte(onePage))
	}

	page, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(),
		banking.TransactionsRequest{From: time.Date(2026, time.February, 25, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(page.Transactions) != 3 {
		t.Errorf("the retry returned %d transactions, want 3", len(page.Transactions))
	}
	if want := []string{"default", "longest"}; len(strategies) != 2 ||
		strategies[0] != want[0] || strategies[1] != want[1] {
		t.Errorf("strategies = %v, want exactly %v", strategies, want)
	}
}

// Once, and no more. A gateway that always refuses the period must not be
// retried in a loop, and the second failure is mapped like any other.
func TestASecondWrongPeriodSurfacesAndAddsNoTaxonomyMember(t *testing.T) {
	s, c := newServer(t)

	attempts := 0
	s.handlers["GET /accounts/uid-1/transactions"] = func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"WRONG_TRANSACTIONS_PERIOD"}`))
	}

	_, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(),
		banking.TransactionsRequest{From: time.Date(2026, time.February, 25, 0, 0, 0, 0, time.UTC)})
	if err == nil {
		t.Fatal("a gateway refusing twice reported success")
	}
	if attempts != 2 {
		t.Errorf("attempted %d times, want exactly 2", attempts)
	}
	if !errors.Is(err, banking.ErrGatewayUnavailable) {
		t.Errorf("got %v, want the existing taxonomy rather than a new member", err)
	}
}

// The read is a member-present call, which is what keeps the background-fetch
// cap out of reach.
func TestTheTransactionsReadCarriesThePSUHeaders(t *testing.T) {
	s, c := newServer(t)
	s.on("GET /accounts/uid-1/transactions", http.StatusOK, `{"transactions":[]}`)

	if _, err := c.Transactions(context.Background(), banking.Connection{}, ledgerAccount(), banking.TransactionsRequest{}); err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if s.last().Header.Get("PSU-IP-Address") == "" {
		t.Error("the read did not claim the member was present")
	}
}

// The consent asked for at the bank now covers both, because transactions are a
// separate scope and no reading of a balances-only grant produces them.
func TestTheConsentAsksForTransactionsAlongsideBalances(t *testing.T) {
	s, c := newServer(t)
	s.on("POST /auth", http.StatusOK, `{"url":"https://bank.example/consent","authorization_id":"auth-1"}`)

	if _, err := c.BeginConnection(context.Background(), banking.BeginRequest{
		Bank:       banking.Bank{ID: "PT:Montepio", Name: "Montepio", Country: "PT"},
		ValidUntil: time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		State:      "state-1",
	}); err != nil {
		t.Fatalf("BeginConnection: %v", err)
	}

	access, ok := s.last().Body["access"].(map[string]any)
	if !ok {
		t.Fatalf("the request carried no access object: %v", s.last().Body)
	}
	if access["balances"] != true {
		t.Error("the consent stopped asking for balances")
	}
	if access["transactions"] != true {
		t.Error("the consent does not ask for transactions")
	}
}
