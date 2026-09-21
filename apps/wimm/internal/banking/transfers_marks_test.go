package banking_test

import (
	"context"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// The line is balances and does not change. A same-day pair on two accounts
// the chart draws still moved the day's money, so the rows stay and are
// labelled instead of dropped.
func TestTheBalanceLineIsUntouchedByAPair(t *testing.T) {
	balances := func(t *testing.T, withPair bool) []int64 {
		t.Helper()
		svc, st, joint, personal := transferHousehold(t)
		seedPayment(st, "shop", personal, -2_500, on(2026, time.September, 4), "COMPRA COFFEE 111111111111")
		if withPair {
			seedPayment(st, "out", personal, -50_000, on(2026, time.September, 10), "TRANSFERENCIA")
			seedPayment(st, "in", joint, 50_000, on(2026, time.September, 10), "TRANSFERENCIA")
		}
		tr, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
		if err != nil {
			t.Fatal(err)
		}
		var out []int64
		for _, c := range tr.Currencies {
			for _, p := range c.Points {
				out = append(out, p.Money.Minor)
			}
		}
		return out
	}

	// The pair moves money between two accounts the chart already draws, so
	// every point is the sum it was.
	with, without := balances(t, true), balances(t, false)
	if len(with) != len(without) {
		t.Fatalf("got %d points with a pair and %d without", len(with), len(without))
	}
	for i := range with {
		if with[i] != without[i] {
			t.Errorf("point %d = %d with a pair, %d without", i, with[i], without[i])
		}
	}
}

// A mover is labelled on the day the money left, and is never also unusual.
func TestAMoverThatIsHalfOfAPairIsLabelledAndNotUnusual(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	for m := time.July; m <= time.September; m++ {
		seedPayment(st, "coffee-"+m.String(), personal, -1_000, on(2026, m, 4), "COMPRA COFFEE 111111111111")
	}
	left := on(2026, time.September, 10)
	seedPayment(st, "out", personal, -600_000, left, "TRANSFERENCIA")
	seedPayment(st, "in", joint, 600_000, left, "TRANSFERENCIA")

	tr, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, c := range tr.Currencies {
		for _, p := range c.Points {
			for _, m := range p.Movers {
				if m.TransactionID != "out" {
					continue
				}
				found = true
				if got := p.Date.Format("2006-01-02"); got != left.Format("2006-01-02") {
					t.Errorf("the mover is on %s, want the day the money left, %s",
						got, left.Format("2006-01-02"))
				}
				if !m.OwnTransfer {
					t.Error("the mover is not labelled a transfer")
				}
				if m.Unusual {
					t.Error("the mover is marked unusual, and a transfer never is")
				}
			}
		}
	}
	if !found {
		t.Fatal("the transfer is not among the day's movers, and the day's money did move")
	}
}

// --- Transactions -------------------------------------------------------

func ledgerRows(t *testing.T, svc *banking.Service, member, accountID string) banking.Ledger {
	t.Helper()
	l, err := svc.Transactions(context.Background(), banking.LedgerRequest{
		MemberID: member, AccountID: accountID, SkipSync: true,
	})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	return l
}

func TestBothHalvesOfAPairAreLabelledInTheLedger(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	seedPayment(st, "out", personal, -50_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 50_000, on(2026, time.September, 11), "TRANSFERENCIA")

	l := ledgerRows(t, svc, ada, "")
	for _, id := range []string{"out", "in"} {
		if !l.Patterns.IsTransfer(id) {
			t.Errorf("%s is not labelled a transfer", id)
		}
	}
}

// Reading one account still marks the half that is in it, because pairing is
// done across every account the member owns and only the page is narrowed.
func TestReadingOneAccountStillMarksTheHalfInIt(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	seedPayment(st, "out", personal, -50_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 50_000, on(2026, time.September, 11), "TRANSFERENCIA")

	l := ledgerRows(t, svc, ada, personal)
	if !l.Patterns.IsTransfer("out") {
		t.Error("the out row is not labelled when only its own account is read")
	}
	for _, r := range l.Page.Transactions {
		if r.ID == "in" {
			t.Error("the partner is on the page, and only one account was asked for")
		}
	}
}

// grace owns the joint account and not ada's, so her row set holds no pair.
// A label there would tell her ada has an account holding at least 500.00.
func TestTheOtherOwnerGetsNoMarkOnTheArrivingRow(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	seedPayment(st, "out", personal, -50_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 50_000, on(2026, time.September, 11), "TRANSFERENCIA")

	l := ledgerRows(t, svc, grace, "")
	if l.Patterns.IsTransfer("in") {
		t.Error("grace's arriving row is labelled a transfer, which names an account she does not own")
	}
	if l.Patterns.IsTransfer("out") {
		t.Error("grace can see a row on an account she does not own")
	}
}

// Under All every pair is inside the scope, so no labelled row is also
// unusual: the label wins the slot.
func TestALabelledLedgerRowIsNeverAlsoUnusual(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	for m := time.July; m <= time.September; m++ {
		seedPayment(st, "coffee-"+m.String(), personal, -1_000, on(2026, m, 4), "COMPRA COFFEE 111111111111")
		seedPayment(st, "shop-"+m.String(), personal, -2_000, on(2026, m, 6), "COMPRA PINGO DOCE 222222222222")
	}
	seedPayment(st, "out", personal, -600_000, on(2026, time.September, 10), "TRANSFERENCIA")
	seedPayment(st, "in", joint, 600_000, on(2026, time.September, 10), "TRANSFERENCIA")

	l := ledgerRows(t, svc, ada, "")
	if !l.Patterns.IsTransfer("out") {
		t.Fatal("the out row is not labelled")
	}
	if _, marked := l.Patterns.MarkFor("out"); marked {
		t.Error("the out row is also marked unusual, and the label takes the slot")
	}
}

// Patterns is asked through its methods so the precedence cannot be applied
// two ways. A row carrying both is a transfer and nothing else.
func TestPatternsGivesTheSlotToTheTransfer(t *testing.T) {
	p := banking.Patterns{
		Unusual:   map[string]banking.UnusualMark{"both": {TransactionID: "both"}, "plain": {TransactionID: "plain"}},
		Transfers: map[string]string{"both": "partner", "partner": "both"},
	}
	if !p.IsTransfer("both") {
		t.Error("both is not a transfer")
	}
	if _, ok := p.MarkFor("both"); ok {
		t.Error("both is still unusual, and the label takes the slot")
	}
	if _, ok := p.MarkFor("plain"); !ok {
		t.Error("plain lost its mark, and it is half of nothing")
	}
}

// A page older than every month shown is judged against nothing, but a row on
// it may still be half of a transfer and is still labelled.
func TestAPageOlderThanTheMonthsShownIsStillLabelled(t *testing.T) {
	svc, st, joint, personal := transferHousehold(t)
	old := on(2024, time.March, 10)
	seedPayment(st, "out", personal, -50_000, old, "TRANSFERENCIA")
	seedPayment(st, "in", joint, 50_000, old, "TRANSFERENCIA")

	l, err := svc.Transactions(context.Background(), banking.LedgerRequest{
		MemberID: ada, SkipSync: true, Oldest: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var onPage bool
	for _, r := range l.Page.Transactions {
		if r.ID == "out" {
			onPage = true
		}
	}
	if !onPage {
		t.Skip("the oldest page does not hold the transfer")
	}
	if !l.Patterns.IsTransfer("out") {
		t.Error("a row older than the months shown lost its label")
	}
}
