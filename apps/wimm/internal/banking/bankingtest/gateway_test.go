package bankingtest_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/banking/bankingtest"
)

func montepio() banking.Bank {
	return banking.Bank{
		ID: "montepio", Name: "Caixa Económica Montepio Geral",
		Country: "PT", MaxConsent: 90 * 24 * time.Hour,
	}
}

// activobank grants a single day, which is not a variation on ninety: it makes
// restoring the most-used path in the product rather than its failure handling.
func activobank() banking.Bank {
	return banking.Bank{ID: "activobank", Name: "ActivoBank", Country: "PT", MaxConsent: 24 * time.Hour}
}

func account(ref, name string) banking.Account {
	return banking.Account{Ref: ref, Name: name, NumberSuffix: ref[len(ref)-4:], Currency: "EUR"}
}

// connect runs the whole round trip and returns what the caller would store.
func connect(t *testing.T, g *bankingtest.Gateway, b banking.Bank) (banking.Connection, []banking.Account) {
	t.Helper()
	ctx := context.Background()

	handoff, err := g.BeginConnection(ctx, banking.BeginRequest{
		Bank: b, RedirectURL: "https://localhost:8765/psd2/callback", State: "state-1",
	})
	if err != nil {
		t.Fatalf("BeginConnection: %v", err)
	}

	conn, accounts, err := g.CompleteConnection(ctx,
		banking.PendingConnection{Bank: b, GatewayRef: handoff.GatewayRef, State: "state-1"},
		banking.Callback{Code: "code", State: "state-1"},
	)
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}
	return conn, accounts
}

func TestTheRoundTripReturnsAccountsAndABalance(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	g.SetBalance("montepio", "acct-6580", banking.Balance{
		Money: banking.Money{Minor: 420_000, Currency: "EUR"}, ReadAt: bankingtest.Epoch, Kind: "available",
	})

	conn, accounts := connect(t, g, montepio())
	if len(accounts) != 1 {
		t.Fatalf("got %d accounts, want 1", len(accounts))
	}
	if accounts[0].GatewayUID == "" {
		t.Error("the account carries no per-session identifier")
	}

	balances, err := g.Balances(context.Background(), conn, accounts[0])
	if err != nil {
		t.Fatalf("Balances: %v", err)
	}
	if balances[0].Money.Minor != 420_000 {
		t.Errorf("Minor = %d", balances[0].Money.Minor)
	}
}

// Consent length comes from the bank, not from wimm, and the gap between these
// two is the whole reason the member is shown a real date.
func TestConsentLengthComesFromTheBank(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	g.AddBank(activobank(), account("acct-5594", "Conta ActivoBank"))

	long, _ := connect(t, g, montepio())
	short, _ := connect(t, g, activobank())

	if want := bankingtest.Epoch.Add(90 * 24 * time.Hour); !long.ExpiresAt.Equal(want) {
		t.Errorf("Montepio expires %s, want %s", long.ExpiresAt, want)
	}
	if want := bankingtest.Epoch.Add(24 * time.Hour); !short.ExpiresAt.Equal(want) {
		t.Errorf("ActivoBank expires %s, want %s", short.ExpiresAt, want)
	}
}

// The property account identity rests on: a restore issues new per-session
// identifiers while the cross-session ref stays put. A schema keyed on the uid
// would lose every sharing choice on the first restore.
func TestRestoringChangesTheUIDAndKeepsTheRef(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))

	_, before := connect(t, g, montepio())
	_, after := connect(t, g, montepio())

	if before[0].Ref != after[0].Ref {
		t.Errorf("the cross-session ref changed: %q then %q", before[0].Ref, after[0].Ref)
	}
	if before[0].GatewayUID == after[0].GatewayUID {
		t.Error("the per-session uid survived a restore, which the real gateway's does not")
	}
}

// The two restore cases 3.4 names.
func TestABankThatOffersANewAccountOnRestore(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	_, before := connect(t, g, montepio())

	g.OffersOnRestore("montepio", account("acct-6580", "Conta à Ordem"), account("acct-7712", "Poupança"))
	_, after := connect(t, g, montepio())

	if len(before) != 1 || len(after) != 2 {
		t.Fatalf("before %d accounts, after %d; want 1 then 2", len(before), len(after))
	}
}

func TestABankThatWithdrawsAnAccountOnRestore(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"), account("acct-7712", "Poupança"))
	_, before := connect(t, g, montepio())

	g.OffersOnRestore("montepio", account("acct-6580", "Conta à Ordem"))
	_, after := connect(t, g, montepio())

	if len(before) != 2 || len(after) != 1 {
		t.Fatalf("before %d accounts, after %d; want 2 then 1", len(before), len(after))
	}
	if after[0].Ref != "acct-6580" {
		t.Errorf("the wrong account was withdrawn: %q remains", after[0].Ref)
	}
}

// Every error in 3.2 must be scriptable, or the service's failure paths cannot
// be tested at all.
func TestEveryFailureInTheTaxonomyIsScriptable(t *testing.T) {
	for _, want := range []error{
		banking.ErrBankUnavailable,
		banking.ErrGatewayUnavailable,
		banking.ErrConsentDeclined,
		banking.ErrConsentExpired,
		banking.ErrNoAccounts,
		banking.ErrRateLimited,
	} {
		t.Run(want.Error(), func(t *testing.T) {
			g := bankingtest.New()
			g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
			g.Fail("Banks", want)

			if _, err := g.Banks(context.Background(), "PT"); !errors.Is(err, want) {
				t.Errorf("got %v, want %v", err, want)
			}
		})
	}
}

// Rate limiting carries a retry-after through the fake unchanged, because the
// screen shows it to the member.
func TestARateLimitKeepsItsRetryAfter(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	conn, accounts := connect(t, g, montepio())
	g.Fail("Balances", banking.RateLimited(6*time.Hour))

	_, err := g.Balances(context.Background(), conn, accounts[0])
	var limited *banking.RateLimitError
	if !errors.As(err, &limited) {
		t.Fatalf("got %v, want a rate limit", err)
	}
	if limited.RetryAfter != 6*time.Hour {
		t.Errorf("RetryAfter = %s", limited.RetryAfter)
	}
}

// Failures queue, so "fails once then works" — the partial-failure path on
// Overview — is expressible.
func TestAScriptedFailureIsConsumedOnce(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	g.SetBalance("montepio", "acct-6580", banking.Balance{Money: banking.Money{Minor: 1, Currency: "EUR"}})
	conn, accounts := connect(t, g, montepio())

	g.Fail("Balances", banking.ErrBankUnavailable)
	if _, err := g.Balances(context.Background(), conn, accounts[0]); err == nil {
		t.Fatal("the scripted failure did not fire")
	}
	if _, err := g.Balances(context.Background(), conn, accounts[0]); err != nil {
		t.Fatalf("the failure fired twice: %v", err)
	}
}

// A connection cut off early surfaces as consent expired, which is how
// EXPIRED_SESSION reaches the member as "my bank stopped updating".
func TestAConnectionCutOffEarlyReadsAsExpired(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	conn, accounts := connect(t, g, montepio())
	g.Expire(conn)

	if _, err := g.Balances(context.Background(), conn, accounts[0]); !errors.Is(err, banking.ErrConsentExpired) {
		t.Errorf("got %v, want ErrConsentExpired", err)
	}
}

func TestDecliningAtTheBank(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))

	_, _, err := g.CompleteConnection(context.Background(),
		banking.PendingConnection{Bank: montepio(), GatewayRef: "h", State: "s"},
		banking.Callback{Error: "access_denied", State: "s"},
	)
	if !errors.Is(err, banking.ErrConsentDeclined) {
		t.Errorf("got %v, want ErrConsentDeclined", err)
	}
}

func TestAccessGrantedThatExposesNoAccounts(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio()) // no accounts

	_, _, err := g.CompleteConnection(context.Background(),
		banking.PendingConnection{Bank: montepio(), GatewayRef: "h", State: "s"},
		banking.Callback{Code: "c", State: "s"},
	)
	if !errors.Is(err, banking.ErrNoAccounts) {
		t.Errorf("got %v, want ErrNoAccounts", err)
	}
}

// AccountsRead is what "no balance is read for an account nobody may see" is
// asserted against, so it has to record exactly what was read and nothing else.
func TestAccountsReadRecordsOnlyBalanceReads(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"), account("acct-7712", "Poupança"))
	g.SetBalance("montepio", "acct-6580", banking.Balance{Money: banking.Money{Minor: 1, Currency: "EUR"}})
	conn, accounts := connect(t, g, montepio())

	if _, err := g.Balances(context.Background(), conn, accounts[0]); err != nil {
		t.Fatalf("Balances: %v", err)
	}

	read := g.AccountsRead()
	if len(read) != 1 || read[0] != "acct-6580" {
		t.Errorf("AccountsRead = %v, want [acct-6580]", read)
	}
}

func tx(ref string, day int, minor int64, status banking.TransactionStatus) banking.Transaction {
	return banking.Transaction{
		Ref:         ref,
		Status:      status,
		Amount:      banking.Money{Minor: minor, Currency: "EUR"},
		BookingDate: time.Date(2026, time.March, day, 0, 0, 0, 0, time.UTC),
	}
}

// A first fill is unbounded pages, so the fake has to page or nothing that
// consumes it is testing what it will actually meet.
func TestTransactionsPageToExhaustion(t *testing.T) {
	g := bankingtest.New()
	g.PageSize = 2
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	g.SetTransactions("montepio", "acct-6580",
		tx("ref-1", 5, -450, banking.StatusBooked),
		tx("ref-2", 4, -1190, banking.StatusBooked),
		tx("ref-3", 3, 240_000, banking.StatusBooked),
		tx("ref-4", 2, -2500, banking.StatusBooked),
		tx("ref-5", 1, -700, banking.StatusPending),
	)

	conn, accounts := connect(t, g, montepio())

	var read []string
	var pages int
	cursor := ""
	for {
		page, err := g.Transactions(context.Background(), conn, accounts[0],
			banking.TransactionsRequest{Cursor: cursor})
		if err != nil {
			t.Fatalf("Transactions: %v", err)
		}
		pages++
		for _, transaction := range page.Transactions {
			read = append(read, transaction.Ref)
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
		if pages > 10 {
			t.Fatal("paging did not terminate")
		}
	}

	if pages != 3 {
		t.Errorf("read %d pages of 2 over 5 transactions, want 3", pages)
	}
	want := []string{"ref-1", "ref-2", "ref-3", "ref-4", "ref-5"}
	if len(read) != len(want) {
		t.Fatalf("read %v, want %v", read, want)
	}
	for i := range want {
		if read[i] != want[i] {
			t.Fatalf("read %v, want %v — nothing repeated and nothing skipped", read, want)
		}
	}
}

// A zero From is "as far back as this bank goes"; a set From is the incremental
// read, and the overlap window depends on it filtering by booking date.
func TestASetFromNarrowsToWhatWasBookedOnOrAfterIt(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	g.SetTransactions("montepio", "acct-6580",
		tx("ref-1", 5, -450, banking.StatusBooked),
		tx("ref-2", 3, -1190, banking.StatusBooked),
		tx("ref-3", 1, 240_000, banking.StatusBooked),
	)

	conn, accounts := connect(t, g, montepio())

	all, err := g.Transactions(context.Background(), conn, accounts[0], banking.TransactionsRequest{})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(all.Transactions) != 3 {
		t.Errorf("a zero From read %d transactions, want every one of them", len(all.Transactions))
	}

	since := banking.TransactionsRequest{From: time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC)}
	page, err := g.Transactions(context.Background(), conn, accounts[0], since)
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(page.Transactions) != 2 {
		t.Fatalf("read %d transactions from the 3rd, want 2 — the boundary day is included", len(page.Transactions))
	}
	if page.Transactions[1].Ref != "ref-2" {
		t.Errorf("the boundary day was dropped: got %v", page.Transactions)
	}
}

// A sync that fails part way through a fill must be expressible, because
// leaving the stored rows and their time untouched is the behaviour that then
// has to be asserted.
func TestASyncCanFailMidPage(t *testing.T) {
	g := bankingtest.New()
	g.PageSize = 1
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	g.SetTransactions("montepio", "acct-6580",
		tx("ref-1", 5, -450, banking.StatusBooked),
		tx("ref-2", 4, -1190, banking.StatusBooked),
	)

	conn, accounts := connect(t, g, montepio())

	first, err := g.Transactions(context.Background(), conn, accounts[0], banking.TransactionsRequest{})
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if first.NextCursor == "" {
		t.Fatal("the fake did not offer a second page")
	}

	g.Fail("Transactions", banking.ErrBankUnavailable)
	_, err = g.Transactions(context.Background(), conn, accounts[0],
		banking.TransactionsRequest{Cursor: first.NextCursor})
	if !errors.Is(err, banking.ErrBankUnavailable) {
		t.Errorf("got %v, want the scripted failure", err)
	}
}

// Reading transactions on a connection the bank has cut off is consent expired,
// exactly as reading a balance on one is: the member's experience of both is
// that their bank stopped updating.
func TestSyncingAnExpiredConnectionSaysConsentExpired(t *testing.T) {
	g := bankingtest.New()
	g.AddBank(montepio(), account("acct-6580", "Conta à Ordem"))
	g.SetTransactions("montepio", "acct-6580", tx("ref-1", 5, -450, banking.StatusBooked))

	conn, accounts := connect(t, g, montepio())
	g.Expire(conn)

	if _, err := g.Transactions(context.Background(), conn, accounts[0], banking.TransactionsRequest{}); !errors.Is(err, banking.ErrConsentExpired) {
		t.Errorf("got %v, want ErrConsentExpired", err)
	}
}
