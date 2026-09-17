package banking_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/banking/bankingtest"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// These run against the in-memory gateway and an in-memory store, so the whole
// service is exercised with no network and no database — which is what
// bankingtest exists for (ADR 0018).

const (
	ada   = "11111111-1111-4111-8111-111111111111"
	grace = "22222222-2222-4222-8222-222222222222"
)

func newService(t *testing.T) (*banking.Service, *bankingtest.Gateway, *memStore) {
	t.Helper()

	keys, err := banking.NewKeyring([]banking.Key{{ID: "k1", Material: testKey(3)}})
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	gw := bankingtest.New()
	st := newMemStore()

	svc := banking.NewService(st, gw, keys, slog.New(slog.NewTextHandler(io.Discard, nil)),
		"https://localhost:8765/psd2/callback", 24*time.Hour, testLedgerOptions())
	return svc, gw, st
}

func montepio() banking.Bank {
	return banking.Bank{ID: "PT:Montepio", Name: "Montepio", Country: "PT", MaxConsent: 90 * 24 * time.Hour}
}

// connectBank runs the whole round trip the way a member does.
func connectBank(t *testing.T, svc *banking.Service, gw *bankingtest.Gateway, st *memStore) banking.Completed {
	t.Helper()
	ctx := context.Background()

	if _, err := svc.BeginConnection(ctx, ada, montepio().ID); err != nil {
		t.Fatalf("BeginConnection: %v", err)
	}

	done, err := svc.CompleteConnection(ctx, ada, banking.Callback{Code: "code", State: gw.LastState})
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}
	return done
}

func TestConnectingOwnsEveryAccountAndGrantsNobody(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"})

	done := connectBank(t, svc, gw, st)

	if len(done.Accounts) != 2 {
		t.Fatalf("got %d accounts, want both", len(done.Accounts))
	}
	for _, a := range done.Accounts {
		if !a.Owned || a.Level != store.LevelDetails {
			t.Errorf("%s: owned=%v level=%q, want the connecting member to own it", a.Name, a.Owned, a.Level)
		}
	}

	// And nobody else sees anything.
	view, err := svc.Accounts(context.Background(), grace, true)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	if len(view.Accounts) != 0 {
		t.Errorf("a member who was granted nothing sees %d accounts", len(view.Accounts))
	}
	if len(view.Totals) != 0 {
		t.Error("a member who sees nothing was given a total")
	}
}

// 6.3: two members, one household, different views and different totals.
func TestTwoMembersSeeDifferentTotals(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 400_000, Currency: "EUR"}, Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: banking.Money{Minor: 100_000, Currency: "EUR"}, Kind: "CLAV"})

	done := connectBank(t, svc, gw, st)
	joint := done.Accounts[0]

	ctx := context.Background()
	if err := svc.SetLevel(ctx, ada, joint.ID, grace, store.LevelBalance); err != nil {
		t.Fatalf("SetLevel: %v", err)
	}

	adaView, err := svc.Accounts(ctx, ada, false)
	if err != nil {
		t.Fatalf("Accounts(ada): %v", err)
	}
	graceView, err := svc.Accounts(ctx, grace, true)
	if err != nil {
		t.Fatalf("Accounts(grace): %v", err)
	}

	if len(adaView.Totals) != 1 || adaView.Totals[0].Money.Minor != 500_000 {
		t.Errorf("ada's total = %+v, want 500000", adaView.Totals)
	}
	if len(graceView.Totals) != 1 || graceView.Totals[0].Money.Minor != 400_000 {
		t.Errorf("grace's total = %+v, want only the joint account", graceView.Totals)
	}
	if len(graceView.Accounts) != 1 {
		t.Errorf("grace sees %d accounts, want 1", len(graceView.Accounts))
	}
}

// 6.4, and the guarantee ADR 0019 keeps from 0018: an account nobody may see
// is never read from the bank at all.
func TestNoBalanceIsEverReadForAnAccountNobodyMaySee(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"})
	for _, ref := range []string{"hash-1", "hash-2"} {
		gw.SetBalance("PT:Montepio", ref, banking.Balance{Money: banking.Money{Minor: 1000, Currency: "EUR"}, Kind: "CLAV"})
	}

	done := connectBank(t, svc, gw, st)
	personal := done.Accounts[1]

	ctx := context.Background()
	// Disowned, and granted to nobody: the account exists and nothing may read it.
	if err := svc.SetOwners(ctx, ada, personal.ID, nil); err != nil {
		t.Fatalf("SetOwners: %v", err)
	}

	if _, err := svc.Accounts(ctx, ada, false); err != nil {
		t.Fatalf("Accounts: %v", err)
	}

	read := gw.AccountsRead()
	if slices.Contains(read, "hash-2") {
		t.Errorf("a balance was read for an account nobody may see; reads were %v", read)
	}
	if !slices.Contains(read, "hash-1") {
		t.Errorf("the owned account was not read; reads were %v", read)
	}
}

// One bank failing must not lose the other's readings, and the member is told
// which bank did not answer.
func TestOneBankFailingLeavesTheOthersReadings(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 400_000, Currency: "EUR"}, Kind: "CLAV"})

	connectBank(t, svc, gw, st)
	ctx := context.Background()

	// A good read first, so there is a previous reading to preserve.
	if _, err := svc.Accounts(ctx, ada, false); err != nil {
		t.Fatalf("Accounts: %v", err)
	}

	gw.Fail("Balances", banking.RateLimited(6*time.Hour))
	view, err := svc.Accounts(ctx, ada, false)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}

	if len(view.Failures) != 1 {
		t.Fatalf("got %d failures, want the bank that refused", len(view.Failures))
	}
	if view.Failures[0].RetryAfter != 6*time.Hour {
		t.Errorf("RetryAfter = %s, want 6h", view.Failures[0].RetryAfter)
	}
	if view.Failures[0].BankName != "Montepio" {
		t.Errorf("the member is not told which bank: %q", view.Failures[0].BankName)
	}

	// The reading that was already there is still there.
	if len(view.Accounts) != 1 || view.Accounts[0].BalanceMinor == nil {
		t.Fatal("the previous reading was lost when the refresh failed")
	}
	if *view.Accounts[0].BalanceMinor != 400_000 {
		t.Errorf("balance = %d, want the previous reading", *view.Accounts[0].BalanceMinor)
	}
}

// Access running out is recorded, not only reported: the next arrival must
// offer to restore rather than try again and fail the same way.
func TestAccessRunningOutIsRecorded(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 1000, Currency: "EUR"}, Kind: "CLAV"})

	connectBank(t, svc, gw, st)
	ctx := context.Background()

	gw.Fail("Balances", banking.ErrConsentExpired)
	if _, err := svc.Accounts(ctx, ada, false); err != nil {
		t.Fatalf("Accounts: %v", err)
	}

	if !st.expired {
		t.Error("a connection whose access ran out was not recorded as expired")
	}
}

// Only an owner may change ownership or levels.
func TestOnlyAnOwnerMayChangeAnAccount(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"})
	done := connectBank(t, svc, gw, st)
	id := done.Accounts[0].ID
	ctx := context.Background()

	if err := svc.SetLevel(ctx, grace, id, ada, store.LevelBalance); !errors.Is(err, banking.ErrNotOwner) {
		t.Errorf("SetLevel by a non-owner = %v, want ErrNotOwner", err)
	}
	if err := svc.SetOwners(ctx, grace, id, []string{grace}); !errors.Is(err, banking.ErrNotOwner) {
		t.Errorf("SetOwners by a non-owner = %v, want ErrNotOwner", err)
	}
	if err := svc.Disconnect(ctx, grace, done.Connection.ID); !errors.Is(err, banking.ErrNotOwner) {
		t.Errorf("Disconnect by a non-owner = %v, want ErrNotOwner", err)
	}
}

// Mixed currencies are never summed.
func TestCurrenciesAreTotalledSeparately(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Euro", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Sterling", Currency: "GBP"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 400_000, Currency: "EUR"}, Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: banking.Money{Minor: 250_000, Currency: "GBP"}, Kind: "CLAV"})

	connectBank(t, svc, gw, st)
	view, err := svc.Accounts(context.Background(), ada, false)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}

	if len(view.Totals) != 2 {
		t.Fatalf("got %d totals, want one per currency", len(view.Totals))
	}
	for _, total := range view.Totals {
		switch total.Money.Currency {
		case "EUR":
			if total.Money.Minor != 400_000 {
				t.Errorf("EUR total = %d", total.Money.Minor)
			}
		case "GBP":
			if total.Money.Minor != 250_000 {
				t.Errorf("GBP total = %d", total.Money.Minor)
			}
		default:
			t.Errorf("unexpected currency %q", total.Money.Currency)
		}
	}
}

// A member is not told a bank is unreachable when they cannot see any of its
// accounts: that would reveal the bank is connected at all.
func TestAFailureIsNotReportedToAMemberWhoSeesNothingOfThatBank(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Personal", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 1000, Currency: "EUR"}, Kind: "CLAV"})

	connectBank(t, svc, gw, st)
	gw.Fail("Balances", banking.ErrBankUnavailable)

	view, err := svc.Accounts(context.Background(), grace, false)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	if len(view.Failures) != 0 {
		t.Errorf("a member who sees nothing of that bank was told it failed: %+v", view.Failures)
	}
}

// Disowning your last account must stay reversible.
//
// This locked a real member out: unchecking the last account they owned left
// nobody owning anything on that connection, so the chooser refused them —
// and the chooser is the only screen that could undo it. The check asked for
// an owner, and there was none to be.
func TestDisowningYourLastAccountDoesNotLockYouOut(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "CLASSIC CEMG", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Conta à Ordem", Currency: "EUR"})

	done := connectBank(t, svc, gw, st)
	ctx := context.Background()

	// Release both: now nobody owns anything at this bank.
	for _, account := range done.Accounts {
		if err := svc.SetOwners(ctx, ada, account.ID, nil); err != nil {
			t.Fatalf("disowning %s: %v", account.Name, err)
		}
	}

	// The chooser still opens for the member who connected the bank.
	shown, err := svc.ConnectionAccounts(ctx, ada, done.Connection.ID)
	if err != nil {
		t.Fatalf("the chooser locked out the member who connected the bank: %v", err)
	}
	if len(shown) != 2 {
		t.Errorf("shows %d accounts, want both", len(shown))
	}

	// And they can take one back.
	if err := svc.SetOwners(ctx, ada, done.Accounts[0].ID, []string{ada}); err != nil {
		t.Fatalf("reclaiming an orphaned account: %v", err)
	}
	view, err := svc.Accounts(ctx, ada, true)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	if len(view.Accounts) != 1 {
		t.Errorf("sees %d accounts after reclaiming one, want 1", len(view.Accounts))
	}
}

// The escape hatch is only for an account nobody owns. It is not the standing
// power over other people's accounts that ADR 0019 withholds.
func TestTheConnectingMemberCannotTouchAnAccountSomebodyElseOwns(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"})

	done := connectBank(t, svc, gw, st)
	id := done.Accounts[0].ID
	ctx := context.Background()

	// Handed to Grace entirely.
	if err := svc.SetOwners(ctx, ada, id, []string{grace}); err != nil {
		t.Fatalf("handing it over: %v", err)
	}

	// Ada connected the bank and still may not touch it: it has an owner.
	if err := svc.SetOwners(ctx, ada, id, []string{ada}); !errors.Is(err, banking.ErrNotOwner) {
		t.Errorf("SetOwners = %v, want ErrNotOwner", err)
	}
	if err := svc.SetLevel(ctx, ada, id, grace, store.LevelBalance); !errors.Is(err, banking.ErrNotOwner) {
		t.Errorf("SetLevel = %v, want ErrNotOwner", err)
	}
}

// And a member who neither owns anything nor connected the bank is still out.
func TestAStrangerToTheConnectionIsStillRefused(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Personal", Currency: "EUR"})

	done := connectBank(t, svc, gw, st)
	ctx := context.Background()

	if err := svc.SetOwners(ctx, ada, done.Accounts[0].ID, nil); err != nil {
		t.Fatalf("disowning: %v", err)
	}

	// Orphaned, but Grace did not connect this bank.
	if _, err := svc.ConnectionAccounts(ctx, grace, done.Connection.ID); !errors.Is(err, banking.ErrNotOwner) {
		t.Errorf("ConnectionAccounts = %v, want ErrNotOwner", err)
	}
}
