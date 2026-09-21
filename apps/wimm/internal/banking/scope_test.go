package banking_test

import (
	"context"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// scopeHousehold is a joint account owned by ada and grace and a personal one
// owned by ada, each with one September payment of a different size.
func scopeHousehold(t *testing.T) (*banking.Service, *memStore, string, string) {
	t.Helper()
	svc, st, done := twoAccounts(t)
	joint, personal := done.Accounts[0], done.Accounts[1]
	if err := svc.SetOwners(context.Background(), ada, joint.ID, []string{ada, grace}); err != nil {
		t.Fatal(err)
	}
	seedTransaction(st, "anchor-j", joint.ID, -1, "EUR", on(2026, time.July, 20))
	seedTransaction(st, "anchor-p", personal.ID, -1, "EUR", on(2026, time.July, 20))
	seedPayment(st, "j1", joint.ID, -40_000, on(2026, time.September, 3), "COMPRA PINGO DOCE 111111111111")
	seedPayment(st, "p1", personal.ID, -2_500, on(2026, time.September, 4), "COMPRA COFFEE 222222222222")
	return svc, st, joint.ID, personal.ID
}

func monthUnder(t *testing.T, svc *banking.Service, member string, scope banking.Scope) banking.MonthSummary {
	t.Helper()
	m, err := svc.MonthSummary(context.Background(), member, scope)
	if err != nil {
		t.Fatalf("MonthSummary(%v): %v", scope, err)
	}
	return m
}

func outOf(t *testing.T, m banking.MonthSummary) int64 {
	t.Helper()
	if len(m.Months) != 1 {
		t.Fatalf("got %d months, want 1", len(m.Months))
	}
	return m.Months[0].Out.Minor
}

func TestHouseholdCountsTheJointAccountOnly(t *testing.T) {
	svc, _, _, _ := scopeHousehold(t)
	if got := outOf(t, monthUnder(t, svc, ada, banking.ScopeHousehold)); got != 40_000 {
		t.Errorf("Household out = %d, want the joint account's 40000", got)
	}
}

func TestYoursCountsTheMembersOwnAccountsOnly(t *testing.T) {
	svc, _, _, _ := scopeHousehold(t)
	if got := outOf(t, monthUnder(t, svc, ada, banking.ScopeOwn)); got != 2_500 {
		t.Errorf("Yours out = %d, want the personal account's 2500", got)
	}
}

func TestAllIsWhatTheMonthSummaryCountedBeforeTheControlExisted(t *testing.T) {
	svc, _, _, _ := scopeHousehold(t)
	if got := outOf(t, monthUnder(t, svc, ada, banking.ScopeAll)); got != 42_500 {
		t.Errorf("All out = %d, want both accounts, 42500", got)
	}
	if got := outOf(t, monthUnder(t, svc, ada, banking.Scope(0))); got != 42_500 {
		t.Errorf("an unset scope counted %d, want All, 42500", got)
	}
}

func TestEachScopeAddsUpToAll(t *testing.T) {
	svc, _, _, _ := scopeHousehold(t)
	household := outOf(t, monthUnder(t, svc, ada, banking.ScopeHousehold))
	yours := outOf(t, monthUnder(t, svc, ada, banking.ScopeOwn))
	all := outOf(t, monthUnder(t, svc, ada, banking.ScopeAll))
	if household+yours != all {
		t.Errorf("household %d + yours %d != all %d", household, yours, all)
	}
}

func TestAnAccountHeldOnlyByGrantIsInNoScope(t *testing.T) {
	svc, _, _, personal := scopeHousehold(t)
	if err := svc.SetLevel(context.Background(), ada, personal, grace, store.LevelDetails); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []banking.Scope{banking.ScopeAll, banking.ScopeHousehold, banking.ScopeOwn} {
		if got := outOf(t, monthUnder(t, svc, grace, scope)); got != 40_000 {
			t.Errorf("grace under %v counted %d, want only the joint account she owns, 40000", scope, got)
		}
	}
}

func TestAScopeThatIsNotOnOfferIsAnsweredAsAll(t *testing.T) {
	svc, _, _, _ := scopeHousehold(t)
	// Grace owns only the joint account, so Household and All would count the
	// same accounts and no control is on offer: Yours reads as All.
	if got := outOf(t, monthUnder(t, svc, grace, banking.ScopeOwn)); got != 40_000 {
		t.Errorf("grace under an unavailable Yours counted %d, want All, 40000", got)
	}
}
