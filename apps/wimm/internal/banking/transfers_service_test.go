package banking_test

import (
	"context"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

// transferHousehold is the shape every scenario below needs: a joint account
// ada and grace both own, and a personal one only ada owns, each reaching back
// to July. Nothing is seeded into September, so each test states its own month.
func transferHousehold(t *testing.T) (svc *banking.Service, st *memStore, joint, personal string) {
	t.Helper()
	svc, st, done := twoAccounts(t)
	joint, personal = done.Accounts[0].ID, done.Accounts[1].ID
	if err := svc.SetOwners(context.Background(), ada, joint, []string{ada, grace}); err != nil {
		t.Fatal(err)
	}
	// A connection creates the accounts; only a read records their balances,
	// which the balance chart needs before it draws anything.
	if _, err := svc.Accounts(context.Background(), ada, false); err != nil {
		t.Fatalf("Accounts (recording the balances): %v", err)
	}
	seedTransaction(st, "anchor-j", joint, -1, "EUR", on(2026, time.July, 20))
	seedTransaction(st, "anchor-p", personal, -1, "EUR", on(2026, time.July, 20))
	return svc, st, joint, personal
}

func monthFigures(t *testing.T, svc *banking.Service, member string, scope banking.Scope) (in, out int64, months int) {
	t.Helper()
	m, err := svc.MonthSummary(context.Background(), member, scope)
	if err != nil {
		t.Fatalf("MonthSummary(%v): %v", scope, err)
	}
	if len(m.Months) == 0 {
		return 0, 0, 0
	}
	return m.Months[0].In.Minor, m.Months[0].Out.Minor, len(m.Months)
}

// --- The month so far ---------------------------------------------------

func TestAMonthWithATransferToSavingsCountsNeitherHalf(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	seedPayment(st, "salary", personal, 245_000, on(2026, time.September, 1), "SALARY")
	seedPayment(st, "shopping", personal, -190_000, on(2026, time.September, 5), "COMPRA PINGO DOCE 111111111111")
	// Both accounts are ada's, so under All this pair is inside the scope.
	seedPayment(st, "out", personal, -50_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 50_000, on(2026, time.September, 11), "TRANSFERENCIA")

	in, out, _ := monthFigures(t, svc, ada, banking.ScopeAll)
	if in != 245_000 || out != 190_000 {
		t.Errorf("All = in %d out %d, want 245000 / 190000, not 295000 / 240000", in, out)
	}
}

func TestAPairAcrossTheEndOfTheMonthIsInNeitherMonth(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	// The money leaves on the last day of August and arrives on the 1st of
	// September. The read is widened by the window at each end, so the pair is
	// found even though its halves fall in different months.
	seedPayment(st, "out", personal, -80_000, on(2026, time.August, 31), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 80_000, on(2026, time.September, 1), "TRANSFERENCIA")
	seedPayment(st, "aug", personal, -1_000, on(2026, time.August, 4), "COMPRA COFFEE 111111111111")
	seedPayment(st, "sep", personal, -2_000, on(2026, time.September, 4), "COMPRA COFFEE 222222222222")

	m, err := svc.MonthSummary(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	c := m.Months[0]
	if c.In.Minor != 0 || c.Out.Minor != 2_000 {
		t.Errorf("September = in %d out %d, want 0 / 2000", c.In.Minor, c.Out.Minor)
	}
	if c.PriorOut == nil || c.PriorOut.Minor != 1_000 {
		t.Errorf("August out = %v, want 1000: the transfer is in neither month", c.PriorOut)
	}
}

func TestAMonthHoldingOnlyATransferHasNoSummary(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	seedPayment(st, "out", personal, -50_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 50_000, on(2026, time.September, 10), "TRANSFERENCIA")

	if _, _, months := monthFigures(t, svc, ada, banking.ScopeAll); months != 0 {
		t.Errorf("got %d months, want none: nothing was spent or received", months)
	}
}

// --- A transfer inside the scope is left out of what is counted ---------

// payIntoTheJoint is ada sending 800.00 from her own account to the joint one,
// which is one pair for her and crosses both narrower scopes.
func payIntoTheJoint(t *testing.T, st *memStore, joint, personal string, d int) {
	t.Helper()
	seedPayment(st, "out", personal, -80_000, on(2026, time.September, d), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 80_000, on(2026, time.September, d), "TRANSFERENCIA")
}

func TestPayingIntoTheJointAccountUnderYoursIsMoneyOut(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	payIntoTheJoint(t, st, joint, personal, 3)

	in, out, _ := monthFigures(t, svc, ada, banking.ScopeOwn)
	if out != 80_000 || in != 0 {
		t.Errorf("Yours = in %d out %d, want 0 / 80000: it left the accounts Yours counts", in, out)
	}
}

func TestPayingIntoTheJointAccountUnderHouseholdIsMoneyIn(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	payIntoTheJoint(t, st, joint, personal, 3)

	in, out, _ := monthFigures(t, svc, ada, banking.ScopeHousehold)
	if in != 80_000 || out != 0 {
		t.Errorf("Household = in %d out %d, want 80000 / 0", in, out)
	}
}

func TestPayingIntoTheJointAccountUnderAllIsInNeitherFigure(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	payIntoTheJoint(t, st, joint, personal, 3)
	seedPayment(st, "shop", personal, -2_500, on(2026, time.September, 4), "COMPRA COFFEE 222222222222")

	in, out, _ := monthFigures(t, svc, ada, banking.ScopeAll)
	if in != 0 || out != 2_500 {
		t.Errorf("All = in %d out %d, want 0 / 2500", in, out)
	}
}

// Both owners of the joint account read the same Household figures, whichever
// of them paid in. grace has no pair at all, because she owns one side only.
func TestBothOwnersReadTheSameHouseholdMonth(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	payIntoTheJoint(t, st, joint, personal, 3)
	seedPayment(st, "jshop", joint, -12_000, on(2026, time.September, 6), "COMPRA PINGO DOCE 333333333333")

	adaIn, adaOut, _ := monthFigures(t, svc, ada, banking.ScopeHousehold)
	graceIn, graceOut, _ := monthFigures(t, svc, grace, banking.ScopeHousehold)
	if adaIn != graceIn || adaOut != graceOut {
		t.Errorf("ada reads in %d out %d, grace reads in %d out %d: they must agree",
			adaIn, adaOut, graceIn, graceOut)
	}
	if adaIn != 80_000 || adaOut != 12_000 {
		t.Errorf("Household = in %d out %d, want 80000 / 12000", adaIn, adaOut)
	}
}

func TestALargeTransferIsNotAnUnusualPayment(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	for m := time.July; m <= time.September; m++ {
		seedPayment(st, "coffee-"+m.String(), personal, -1_000, on(2026, m, 4), "COMPRA COFFEE 111111111111")
		seedPayment(st, "shop-"+m.String(), personal, -2_000, on(2026, m, 6), "COMPRA PINGO DOCE 222222222222")
	}
	// Far above anything ada usually pays, and half of a pair inside All.
	seedPayment(st, "out", personal, -600_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 600_000, on(2026, time.September, 10), "TRANSFERENCIA")

	h, err := svc.MonthHistory(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range h.Currencies {
		for _, m := range c.Months {
			for _, u := range m.Unusual {
				if u.Transaction.ID == "out" || u.Transaction.ID == "in" {
					t.Errorf("%s was called unusual, and it is half of a transfer", u.Transaction.ID)
				}
			}
		}
	}
}

func TestAStandingTransferIsNotRecurringUnderAllAndIsUnderYours(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	// 300.00 to the joint account on the 1st of every month, for long enough
	// to be a monthly run.
	for i := range 6 {
		d := on(2026, time.April, 1).AddDate(0, i, 0)
		seedPayment(st, "out-"+d.Format("2006-01"), personal, -30_000, d, "TRANSFERENCIA POUPANCA")
		seedPayment(st, "in-"+d.Format("2006-01"), joint, 30_000, d, "TRANSFERENCIA POUPANCA")
	}

	recurring := func(scope banking.Scope) int {
		t.Helper()
		h, err := svc.MonthHistory(context.Background(), ada, scope)
		if err != nil {
			t.Fatal(err)
		}
		var n int
		for _, c := range h.Currencies {
			n += len(c.Recurring)
		}
		return n
	}

	if got := recurring(banking.ScopeAll); got != 0 {
		t.Errorf("All listed %d recurring payments, want 0: the pair is inside the scope", got)
	}
	if got := recurring(banking.ScopeOwn); got == 0 {
		t.Error("Yours listed none, want the standing transfer: it leaves Yours every month")
	}
}

func TestATransferIsInNeitherTopMerchantsNorLargestPayments(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	seedPayment(st, "shop", personal, -2_500, on(2026, time.September, 4), "COMPRA PINGO DOCE 111111111111")
	seedPayment(st, "out", personal, -90_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 90_000, on(2026, time.September, 10), "TRANSFERENCIA")

	m, err := svc.MonthSummary(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	c := m.Months[0]
	for _, p := range c.LargestPayments {
		if p.ID == "out" {
			t.Error("the transfer is among the largest payments")
		}
	}
	for _, mt := range c.TopMerchants {
		if mt.Total.Minor == 90_000 {
			t.Errorf("the transfer is a top merchant: %+v", mt)
		}
	}
}

func TestOneThatCouldNotBePairedIsMoneyOutInEveryScope(t *testing.T) {
	svc, st, _, personal := transferHousehold(t)
	// Sent to an account wimm does not hold: nothing arrives anywhere.
	seedPayment(st, "sent", personal, -100_000, on(2026, time.September, 10), "TRANSFERENCIA")

	for _, scope := range []banking.Scope{banking.ScopeAll, banking.ScopeOwn} {
		if _, out, _ := monthFigures(t, svc, ada, scope); out != 100_000 {
			t.Errorf("scope %v out = %d, want 100000", scope, out)
		}
	}
}

// --- A month says how many transfers it left out ------------------------

func monthsUnder(t *testing.T, svc *banking.Service, member string, scope banking.Scope) []banking.HistoryMonth {
	t.Helper()
	h, err := svc.MonthHistory(context.Background(), member, scope)
	if err != nil {
		t.Fatalf("MonthHistory(%v): %v", scope, err)
	}
	if len(h.Currencies) == 0 {
		return nil
	}
	return h.Currencies[0].Months
}

func TestAMonthStatesHowManyTransfersItLeftOut(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	seedPayment(st, "out1", personal, -50_000, on(2026, time.September, 3), "TRANSFERENCIA")
	seedPayment(st, "in1", joint, 50_000, on(2026, time.September, 3), "TRANSFERENCIA")
	seedPayment(st, "out2", personal, -90_000, on(2026, time.September, 12), "TRANSFERENCIA")
	seedPayment(st, "in2", joint, 90_000, on(2026, time.September, 12), "TRANSFERENCIA")

	for _, m := range monthsUnder(t, svc, ada, banking.ScopeAll) {
		if !m.Start.Equal(on(2026, time.September, 1)) {
			continue
		}
		if m.TransfersLeftOut != 2 || m.TransfersTotal.Minor != 140_000 {
			t.Errorf("September left out %d transfers totalling %d, want 2 and 140000",
				m.TransfersLeftOut, m.TransfersTotal.Minor)
		}
		return
	}
	t.Fatal("September is not among the months")
}

func TestAMonthWithOneTransferCountsOne(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	payIntoTheJoint(t, st, joint, personal, 3)

	for _, m := range monthsUnder(t, svc, ada, banking.ScopeAll) {
		if !m.Start.Equal(on(2026, time.September, 1)) {
			continue
		}
		if m.TransfersLeftOut != 1 || m.TransfersTotal.Minor != 80_000 {
			t.Errorf("September left out %d totalling %d, want 1 and 80000",
				m.TransfersLeftOut, m.TransfersTotal.Minor)
		}
		return
	}
	t.Fatal("September is not among the months")
}

// A pair is counted in the month its money left, so the months' counts add up
// to the number of pairs even where one straddles a month boundary.
func TestTheMonthsCountsAddUpToTheNumberOfPairs(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	const pairs = 6
	for i := range pairs {
		// The last day of one month, arriving on the first of the next. The
		// run ends in August, so every pair is behind the 17th of September.
		out := on(2026, time.March, 1).AddDate(0, i+1, 0).AddDate(0, 0, -1)
		seedPayment(st, "out-"+out.Format("2006-01-02"), personal, -int64(10_000+i), out, "TRANSFERENCIA")
		seedPayment(st, "in-"+out.Format("2006-01-02"), joint, int64(10_000+i), out.AddDate(0, 0, 1), "TRANSFERENCIA")
	}

	var total int
	for _, m := range monthsUnder(t, svc, ada, banking.ScopeAll) {
		total += m.TransfersLeftOut
	}
	if total != pairs {
		t.Errorf("the months left out %d transfers in total, want %d", total, pairs)
	}
}

func TestAMonthWithNoTransferSaysNothing(t *testing.T) {
	svc, st, _, personal := transferHousehold(t)
	seedPayment(st, "shop", personal, -2_500, on(2026, time.September, 4), "COMPRA COFFEE 111111111111")

	for _, m := range monthsUnder(t, svc, ada, banking.ScopeAll) {
		if m.TransfersLeftOut != 0 || m.TransfersTotal.Minor != 0 {
			t.Errorf("%s left out %d totalling %d, want nothing",
				m.Start.Format("2006-01"), m.TransfersLeftOut, m.TransfersTotal.Minor)
		}
	}
}

// Under Household the pair crosses the scope, so nothing is left out and the
// month says so.
func TestACrossingPairIsLeftOutOfNoMonth(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	payIntoTheJoint(t, st, joint, personal, 3)

	for _, scope := range []banking.Scope{banking.ScopeHousehold, banking.ScopeOwn} {
		for _, m := range monthsUnder(t, svc, ada, scope) {
			if m.TransfersLeftOut != 0 {
				t.Errorf("scope %v, %s left out %d, want 0: the pair crosses the scope",
					scope, m.Start.Format("2006-01"), m.TransfersLeftOut)
			}
		}
	}
}
