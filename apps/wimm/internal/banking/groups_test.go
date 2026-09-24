package banking_test

import (
	"context"
	"testing"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

const linus = "33333333-3333-4333-8333-333333333333"

// twoAccounts connects a bank with a joint-to-be account (400000) and a
// personal one (100000), both owned by ada.
func twoAccounts(t *testing.T) (*banking.Service, *memStore, banking.Completed) {
	t.Helper()
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: eur(400_000), Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: eur(100_000), Kind: "CLAV"})
	return svc, st, connectBank(t, svc, gw, st)
}

func groupOf(t *testing.T, view banking.View, accountID string) banking.Group {
	t.Helper()
	g, ok := view.Groups[accountID]
	if !ok {
		t.Fatalf("account %s has no group", accountID)
	}
	return g
}

func sum(totals []banking.Total) map[string]int64 {
	out := map[string]int64{}
	for _, tot := range totals {
		out[tot.Money.Currency] += tot.Money.Minor
	}
	return out
}

func viewOf(t *testing.T, svc *banking.Service, member string) banking.View {
	t.Helper()
	v, err := svc.Accounts(context.Background(), member, false)
	if err != nil {
		t.Fatalf("Accounts(%s): %v", member, err)
	}
	return v
}

func TestAJointAccountIsHouseholdMoneyForBothAndOwnForNeither(t *testing.T) {
	svc, _, done := twoAccounts(t)
	ctx := context.Background()
	joint, personal := done.Accounts[0], done.Accounts[1]
	if err := svc.SetOwners(ctx, ada, joint.ID, []string{ada, grace}); err != nil {
		t.Fatal(err)
	}

	for _, member := range []string{ada, grace} {
		v := viewOf(t, svc, member)
		if got := groupOf(t, v, joint.ID); got != banking.GroupHousehold {
			t.Errorf("%s: joint group = %v, want household", member, got)
		}
		if sum(v.HouseholdTotals)["EUR"] != 400_000 {
			t.Errorf("%s: household = %+v, want 400000", member, v.HouseholdTotals)
		}
	}
	if got := sum(viewOf(t, svc, ada).OwnTotals)["EUR"]; got != 100_000 {
		t.Errorf("ada's own = %d, want only the personal account", got)
	}
	if v := viewOf(t, svc, ada); groupOf(t, v, personal.ID) != banking.GroupOwn {
		t.Error("the personal account is not ada's own")
	}
	if len(viewOf(t, svc, grace).OwnTotals) != 0 {
		t.Error("grace has an own figure over no accounts")
	}
}

func TestAnAccountSharedInFullWithEveryoneIsHouseholdMoney(t *testing.T) {
	svc, _, done := twoAccounts(t)
	if err := svc.SetLevel(context.Background(), ada, done.Accounts[0].ID, grace, store.LevelDetails); err != nil {
		t.Fatal(err)
	}
	for _, member := range []string{ada, grace} {
		if got := sum(viewOf(t, svc, member).HouseholdTotals)["EUR"]; got != 400_000 {
			t.Errorf("%s: household = %d, want 400000", member, got)
		}
	}
}

func TestAPersonalAccountMovesNobodyElsesFigures(t *testing.T) {
	svc, _, _ := twoAccounts(t)
	v := viewOf(t, svc, grace)
	if len(v.Accounts) != 0 || len(v.HouseholdTotals) != 0 || len(v.OwnTotals) != 0 {
		t.Errorf("grace sees %+v", v)
	}
}

func TestEveryMemberSeesTheSameHouseholdMoney(t *testing.T) {
	svc, _, done := twoAccounts(t)
	if err := svc.SetOwners(context.Background(), ada, done.Accounts[0].ID, []string{ada, grace}); err != nil {
		t.Fatal(err)
	}
	a, g := viewOf(t, svc, ada).HouseholdTotals, viewOf(t, svc, grace).HouseholdTotals
	if len(a) != 1 || len(g) != 1 || a[0] != g[0] {
		t.Errorf("ada %+v, grace %+v", a, g)
	}
}

func TestAHouseholdOfOneHasNoHouseholdMoney(t *testing.T) {
	svc, st, _ := twoAccounts(t)
	st.members = []store.Member{{ID: ada, FirstName: "Ada"}}

	v := viewOf(t, svc, ada)
	if len(v.HouseholdTotals) != 0 {
		t.Errorf("household = %+v, want none", v.HouseholdTotals)
	}
	if got := sum(v.OwnTotals)["EUR"]; got != 500_000 {
		t.Errorf("own = %d, want everything they own", got)
	}
	for _, a := range v.Accounts {
		if v.Groups[a.ID] != banking.GroupOwn {
			t.Errorf("%s is %v, want own", a.Name, v.Groups[a.ID])
		}
	}
}

func TestBalanceAloneIsSharedForEveryoneElseAndOwnForTheOwner(t *testing.T) {
	svc, _, done := twoAccounts(t)
	joint := done.Accounts[0]
	if err := svc.SetLevel(context.Background(), ada, joint.ID, grace, store.LevelBalance); err != nil {
		t.Fatal(err)
	}

	g := viewOf(t, svc, grace)
	if got := groupOf(t, g, joint.ID); got != banking.GroupShared {
		t.Errorf("grace: group = %v, want shared", got)
	}
	if len(g.HouseholdTotals) != 0 || len(g.OwnTotals) != 0 {
		t.Errorf("grace has figures %+v %+v; a shared account counts in none", g.HouseholdTotals, g.OwnTotals)
	}
	a := viewOf(t, svc, ada)
	if groupOf(t, a, joint.ID) != banking.GroupOwn || len(a.HouseholdTotals) != 0 {
		t.Error("shared at balance only, the account is still ada's own")
	}
}

func TestDetailsForOneMemberOfThreeIsNotHouseholdMoney(t *testing.T) {
	svc, st, done := twoAccounts(t)
	st.members = append(st.members, store.Member{ID: linus, FirstName: "Linus"})
	joint := done.Accounts[0]
	if err := svc.SetLevel(context.Background(), ada, joint.ID, grace, store.LevelDetails); err != nil {
		t.Fatal(err)
	}
	for _, member := range []string{ada, grace} {
		v := viewOf(t, svc, member)
		if len(v.HouseholdTotals) != 0 {
			t.Errorf("%s: household = %+v, want none", member, v.HouseholdTotals)
		}
		if groupOf(t, v, joint.ID) == banking.GroupHousehold {
			t.Errorf("%s: account is household money", member)
		}
	}
}

func TestSomebodyJoiningReturnsHouseholdAccountsToTheirOwners(t *testing.T) {
	svc, st, done := twoAccounts(t)
	joint := done.Accounts[0]
	if err := svc.SetOwners(context.Background(), ada, joint.ID, []string{ada, grace}); err != nil {
		t.Fatal(err)
	}
	if len(viewOf(t, svc, ada).HouseholdTotals) != 1 {
		t.Fatal("precondition: the joint account is household money")
	}

	st.members = append(st.members, store.Member{ID: linus, FirstName: "Linus"})

	v := viewOf(t, svc, ada)
	if len(v.HouseholdTotals) != 0 {
		t.Errorf("household = %+v after a third member joined", v.HouseholdTotals)
	}
	if got := sum(v.OwnTotals)["EUR"]; got != 500_000 {
		t.Errorf("own = %d, want the joint account back in it", got)
	}
}

func TestEachFigureAgreesWithTheAccountsUnderIt(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Overdrawn joint", Currency: "EUR"},
		banking.Account{Ref: "hash-3", Name: "Personal", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: eur(300_000), Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: eur(-50_000), Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-3", banking.Balance{Money: eur(70_000), Kind: "CLAV"})
	done := connectBank(t, svc, gw, st)
	for _, a := range done.Accounts[:2] {
		if err := svc.SetOwners(context.Background(), ada, a.ID, []string{ada, grace}); err != nil {
			t.Fatal(err)
		}
	}

	v := viewOf(t, svc, ada)
	listed := map[banking.Group]int64{}
	for _, a := range v.Accounts {
		listed[v.Groups[a.ID]] += *a.BalanceMinor
	}
	if sum(v.HouseholdTotals)["EUR"] != listed[banking.GroupHousehold] || listed[banking.GroupHousehold] != 250_000 {
		t.Errorf("household = %+v, accounts add to %d", v.HouseholdTotals, listed[banking.GroupHousehold])
	}
	if sum(v.OwnTotals)["EUR"] != listed[banking.GroupOwn] || listed[banking.GroupOwn] != 70_000 {
		t.Errorf("own = %+v, accounts add to %d", v.OwnTotals, listed[banking.GroupOwn])
	}
}

func TestALeftOutAccountIsInNoFigure(t *testing.T) {
	svc, _, done := twoAccounts(t)
	joint := done.Accounts[0]
	if err := svc.SetOwners(context.Background(), ada, joint.ID, []string{ada, grace}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetLeftOut(context.Background(), ada, joint.ID, true); err != nil {
		t.Fatal(err)
	}
	v := viewOf(t, svc, ada)
	if len(v.HouseholdTotals) != 0 {
		t.Errorf("household = %+v, want none", v.HouseholdTotals)
	}
	if got := sum(v.OwnTotals)["EUR"]; got != 100_000 {
		t.Errorf("own = %d, want the personal account only", got)
	}
}

func TestAccountsInTwoCurrenciesAreFiguredSeparately(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Euro", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Sterling", Currency: "GBP"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: eur(1000), Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: banking.Money{Minor: 2000, Currency: "GBP"}, Kind: "CLAV"})
	connectBank(t, svc, gw, st)

	own := viewOf(t, svc, ada).OwnTotals
	if len(own) != 2 {
		t.Fatalf("own = %+v, want one per currency", own)
	}
}
