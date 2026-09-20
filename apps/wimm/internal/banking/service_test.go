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
	return newServiceWithBanks(t, "Montepio")
}

func newServiceWithBanks(t *testing.T, connectable ...string) (*banking.Service, *bankingtest.Gateway, *memStore) {
	t.Helper()

	keys, err := banking.NewKeyring([]banking.Key{{ID: "k1", Material: testKey(3)}})
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	gw := bankingtest.New()
	st := newMemStore()

	svc := banking.NewService(st, gw, keys, slog.New(slog.NewTextHandler(io.Discard, nil)),
		"https://localhost:8765/psd2/callback", 24*time.Hour, connectable, testLedgerOptions())
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
	if len(view.HouseholdTotals) != 0 || len(view.OwnTotals) != 0 {
		t.Error("a member who sees nothing was given a total")
	}
}

// Two members, one household: the owner's figure holds the account, and a
// member granted only its balance has it listed and in no figure.
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

	if len(adaView.OwnTotals) != 1 || adaView.OwnTotals[0].Money.Minor != 500_000 {
		t.Errorf("ada's own = %+v, want 500000", adaView.OwnTotals)
	}
	if len(graceView.OwnTotals) != 0 || len(graceView.HouseholdTotals) != 0 {
		t.Errorf("grace's figures = %+v %+v, want none: the account is shared with her at balance",
			graceView.OwnTotals, graceView.HouseholdTotals)
	}
	if len(graceView.Accounts) != 1 {
		t.Errorf("grace sees %d accounts, want 1", len(graceView.Accounts))
	}
}

// The guarantee ADR 0022 keeps from 0018/0019: a left-out account is never
// read from the bank at all, whatever it held before.
func TestNoBalanceIsEverReadForALeftOutAccount(t *testing.T) {
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
	if err := svc.SetLeftOut(ctx, ada, personal.ID, true); err != nil {
		t.Fatalf("SetLeftOut: %v", err)
	}

	if _, err := svc.Accounts(ctx, ada, false); err != nil {
		t.Fatalf("Accounts: %v", err)
	}

	read := gw.AccountsRead()
	if slices.Contains(read, "hash-2") {
		t.Errorf("a balance was read for a left-out account; reads were %v", read)
	}
	if !slices.Contains(read, "hash-1") {
		t.Errorf("the account still in wimm was not read; reads were %v", read)
	}
}

// The total agrees with what the member is shown: a left-out account holds a
// balance underneath, and it counts in neither the total nor its account count
// (ADR 0022).
func TestTheTotalDropsByALeftOutAccountsBalance(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 400_000, Currency: "EUR"}, Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: banking.Money{Minor: 100_000, Currency: "EUR"}, Kind: "CLAV"})

	done := connectBank(t, svc, gw, st)
	personal := done.Accounts[1]

	ctx := context.Background()
	before, err := svc.Accounts(ctx, ada, false)
	if err != nil {
		t.Fatalf("Accounts before leaving out: %v", err)
	}
	if len(before.OwnTotals) != 1 || before.OwnTotals[0].Money.Minor != 500_000 || before.OwnTotals[0].AccountCount != 2 {
		t.Fatalf("total before leaving out = %+v, want 500000 across 2 accounts", before.OwnTotals)
	}

	if err := svc.SetLeftOut(ctx, ada, personal.ID, true); err != nil {
		t.Fatalf("SetLeftOut: %v", err)
	}

	after, err := svc.Accounts(ctx, ada, true)
	if err != nil {
		t.Fatalf("Accounts after leaving out: %v", err)
	}
	if len(after.OwnTotals) != 1 || after.OwnTotals[0].Money.Minor != 400_000 || after.OwnTotals[0].AccountCount != 1 {
		t.Errorf("total after leaving out = %+v, want 400000 across 1 account", after.OwnTotals)
	}
}

// A gateway reports "XXX" — ISO 4217's own "no currency" — for an account it
// cannot express as a single currency, such as a multi-currency wallet
// aggregated as one account. Its 0 minor units is not a real balance, so it
// carries no total of its own and does not appear in another currency's.
func TestAnAccountWithNoCurrencyCarriesNoTotal(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Current", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Wallet", Currency: "XXX"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 400_000, Currency: "EUR"}, Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: banking.Money{Minor: 0, Currency: "XXX"}, Kind: "CLAV"})

	connectBank(t, svc, gw, st)

	view, err := svc.Accounts(context.Background(), ada, false)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	if len(view.OwnTotals) != 1 || view.OwnTotals[0].Money.Currency != "EUR" || view.OwnTotals[0].Money.Minor != 400_000 {
		t.Errorf("totals = %+v, want only the EUR total", view.OwnTotals)
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

	if len(view.OwnTotals) != 2 {
		t.Fatalf("got %d totals, want one per currency", len(view.OwnTotals))
	}
	for _, total := range view.OwnTotals {
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

// Disowning your last account used to lock a real member out: unchecking the
// last account they owned left nobody owning anything on that connection, so
// the chooser refused them — and the chooser was the only screen that could
// undo it.
//
// ADR 0022 closes this a different way: an account always has an owner, so
// disowning your last one is refused outright rather than produced and then
// rescued. This is now a test of that refusal.
func TestDisowningYourLastAccountIsRefused(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "CLASSIC CEMG", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Conta à Ordem", Currency: "EUR"})

	done := connectBank(t, svc, gw, st)
	ctx := context.Background()

	// Releasing either is refused: both are still owned by nobody but ada.
	for _, account := range done.Accounts {
		if err := svc.SetOwners(ctx, ada, account.ID, nil); err == nil {
			t.Errorf("disowning %s's only owner succeeded, want a refusal", account.Name)
		}
	}

	// Both accounts are untouched, and the chooser still shows them.
	shown, err := svc.ConnectionAccounts(ctx, ada, done.Connection.ID)
	if err != nil {
		t.Fatalf("ConnectionAccounts: %v", err)
	}
	if len(shown) != 2 {
		t.Errorf("shows %d accounts, want both", len(shown))
	}
	view, err := svc.Accounts(ctx, ada, true)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	if len(view.Accounts) != 2 {
		t.Errorf("sees %d accounts after a refused disowning, want both", len(view.Accounts))
	}
}

// The connecting member has no standing power over other people's accounts
// (ADR 0019): owning a bank's connection is not owning what it exposes.
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

	// Grace neither owns this account nor connected the bank.
	if _, err := svc.ConnectionAccounts(ctx, grace, done.Connection.ID); !errors.Is(err, banking.ErrNotOwner) {
		t.Errorf("ConnectionAccounts = %v, want ErrNotOwner", err)
	}
}

func revolut() banking.Bank {
	return banking.Bank{ID: "PT:Revolut", Name: "Revolut", Country: "PT", MaxConsent: 90 * 24 * time.Hour}
}

// Banks offers only wimm's curated set; a bank outside it is treated the same
// as one the gateway does not offer at all (design.md - Decisions).
func TestBanksOffersOnlyTheConfiguredSet(t *testing.T) {
	svc, gw, _ := newServiceWithBanks(t, "montepio") // case-insensitive
	gw.AddBank(montepio())
	gw.AddBank(revolut())

	banks, err := svc.Banks(context.Background(), "PT")
	if err != nil {
		t.Fatalf("Banks: %v", err)
	}
	if len(banks) != 1 || banks[0].Name != "Montepio" {
		t.Errorf("Banks = %v, want only Montepio", banks)
	}
}

// A bank later dropped from the curated set must not orphan a connection
// already restoring against it: RestoreConnection keeps resolving any bank
// the gateway knows.
func TestRestoreConnectionResolvesABankNoLongerConnectable(t *testing.T) {
	svc, gw, _ := newServiceWithBanks(t, "Montepio")
	gw.AddBank(revolut(), banking.Account{Ref: "hash-1", Name: "Personal", Currency: "EUR"})

	if _, err := svc.BeginConnection(context.Background(), ada, revolut().ID); err != nil {
		t.Fatalf("BeginConnection: %v", err)
	}
	done, err := svc.CompleteConnection(context.Background(), ada, banking.Callback{Code: "code", State: gw.LastState})
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}

	// Revolut is not in the curated set, so it would not appear in the picker...
	banks, err := svc.Banks(context.Background(), "PT")
	if err != nil {
		t.Fatalf("Banks: %v", err)
	}
	for _, b := range banks {
		if b.Name == "Revolut" {
			t.Fatalf("Banks includes Revolut, which is outside the curated set")
		}
	}

	// ...but restoring the existing connection still works.
	if _, err := svc.RestoreConnection(context.Background(), ada, done.Connection.ID); err != nil {
		t.Fatalf("RestoreConnection: %v", err)
	}
}

// OwnersOnConnection is what the chooser's "Also owned by" checkboxes read.
// It must report every owner of every account on the connection, not only
// the calling member's own ownership — the gap that let a co-owner added
// through SetOwners never show as checked again.
func TestOwnersOnConnectionReportsEveryAccountsFullOwnerSet(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"})

	done := connectBank(t, svc, gw, st)
	joint, personal := done.Accounts[0].ID, done.Accounts[1].ID
	ctx := context.Background()

	if err := svc.SetOwners(ctx, ada, joint, []string{ada, grace}); err != nil {
		t.Fatalf("SetOwners: %v", err)
	}

	owners, err := svc.OwnersOnConnection(ctx, ada, done.Connection.ID)
	if err != nil {
		t.Fatalf("OwnersOnConnection: %v", err)
	}

	got := map[string][]string{}
	for _, o := range owners {
		got[o.AccountID] = append(got[o.AccountID], o.MemberID)
	}
	if !slices.Contains(got[joint], ada) || !slices.Contains(got[joint], grace) {
		t.Errorf("joint account owners = %v, want both ada and grace", got[joint])
	}
	if slices.Contains(got[personal], grace) {
		t.Errorf("personal account owners = %v, want grace absent", got[personal])
	}
}
