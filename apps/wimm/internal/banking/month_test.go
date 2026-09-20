package banking_test

import (
	"context"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

func on(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func seedPayment(st *memStore, id, accountID string, minor int64, date time.Time, remittance string) {
	st.transactions[accountID] = append(st.transactions[accountID], store.Transaction{
		ID: id, AccountID: accountID, Status: store.StatusBooked,
		AmountMinor: minor, Currency: "EUR", BookingDate: date, Remittance: remittance,
	})
}

// monthLedger is one EUR account with rows in July, August and September
// 2026, so the ledger reaches back well before last month. Now is the 17th.
func monthLedger(t *testing.T) (*banking.Service, *memStore, string) {
	t.Helper()
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "anchor", id, -1, "EUR", on(2026, time.July, 20))

	seedPayment(st, "a1", id, 300_000, on(2026, time.August, 1), "SALARY")
	seedPayment(st, "a2", id, -50_000, on(2026, time.August, 10), "COMPRA RENT 111111111111")
	seedPayment(st, "a3", id, -99_999, on(2026, time.August, 20), "COMPRA TOO LATE 111111111111")
	seedPayment(st, "a4", id, -1_000, on(2026, time.August, 17), "COMPRA COFFEE 111111111111")

	seedPayment(st, "s1", id, 320_000, on(2026, time.September, 1), "SALARY")
	seedPayment(st, "s2", id, -40_000, on(2026, time.September, 3), "COMPRA PINGO DOCE 111111111111")
	seedPayment(st, "s3", id, -2_500, on(2026, time.September, 17), "COMPRA COFFEE 222222222222")
	return svc, st, id
}

func monthOf(t *testing.T, svc *banking.Service, member string) banking.MonthSummary {
	t.Helper()
	m, err := svc.MonthSummary(context.Background(), member)
	if err != nil {
		t.Fatalf("MonthSummary: %v", err)
	}
	return m
}

func TestAMonthUnderWayIsSetAgainstTheSameDaysOfLast(t *testing.T) {
	svc, _, _ := monthLedger(t)
	m := monthOf(t, svc, ada)

	if len(m.Months) != 1 {
		t.Fatalf("got %d months, want 1", len(m.Months))
	}
	c := m.Months[0]
	if c.In.Minor != 320_000 || c.Out.Minor != 42_500 || c.Net.Minor != 277_500 {
		t.Errorf("this month = in %d out %d net %d, want 320000 / 42500 / 277500", c.In.Minor, c.Out.Minor, c.Net.Minor)
	}
	// The 1st to the 17th of August: the rent and the coffee, not the payment
	// on the 20th.
	if c.PriorIn == nil || c.PriorOut == nil || c.PriorNet == nil {
		t.Fatal("no comparison, want one: the ledger reaches back to July")
	}
	if c.PriorIn.Minor != 300_000 || c.PriorOut.Minor != 51_000 || c.PriorNet.Minor != 249_000 {
		t.Errorf("last month = in %d out %d net %d, want 300000 / 51000 / 249000", c.PriorIn.Minor, c.PriorOut.Minor, c.PriorNet.Minor)
	}
	if c.CountedFrom != nil {
		t.Errorf("CountedFrom = %v, want none: the ledger began before the month", c.CountedFrom)
	}
	if !c.MonthStart.Equal(on(2026, time.September, 1)) {
		t.Errorf("MonthStart = %v", c.MonthStart)
	}
}

func TestTheThirtyFirstIsSetAgainstTheWholeOfAShorterMonth(t *testing.T) {
	svc, st, id := monthLedger(t)
	st.now = time.Date(2026, time.October, 31, 9, 0, 0, 0, time.UTC)
	seedPayment(st, "o1", id, -7_000, on(2026, time.October, 31), "COMPRA OCT 333333333333")
	seedPayment(st, "s9", id, -800, on(2026, time.September, 30), "COMPRA LAST DAY 444444444444")

	c := monthOf(t, svc, ada).Months[0]
	// September has 30 days; the prior window is all of it, the 30th included.
	if c.PriorOut == nil || c.PriorOut.Minor != 40_000+2_500+800 {
		t.Errorf("prior out = %v, want all of September: 43300", c.PriorOut)
	}
}

func TestNoComparisonWhenLastMonthIsNotAllHeld(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedPayment(st, "a1", id, -100, on(2026, time.August, 14), "COMPRA X 111111111111")
	seedPayment(st, "s1", id, -200, on(2026, time.September, 3), "COMPRA Y 111111111111")

	c := monthOf(t, svc, ada).Months[0]
	if c.PriorIn != nil || c.PriorOut != nil || c.PriorNet != nil {
		t.Errorf("got a comparison %v %v %v, want none: the ledger begins on 14 August", c.PriorIn, c.PriorOut, c.PriorNet)
	}
	if c.Out.Minor != 200 {
		t.Errorf("out = %d, want this month's figure alone", c.Out.Minor)
	}
}

func TestTheSummarySaysWhereALedgerBeginsInsideTheMonth(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedPayment(st, "s1", id, -200, on(2026, time.September, 10), "COMPRA Y 111111111111")

	c := monthOf(t, svc, ada).Months[0]
	if c.CountedFrom == nil || !c.CountedFrom.Equal(on(2026, time.September, 10)) {
		t.Errorf("CountedFrom = %v, want 10 September", c.CountedFrom)
	}
}

func TestPendingPaymentsAreInNeitherFigure(t *testing.T) {
	svc, st, id := monthLedger(t)
	st.transactions[id] = append(st.transactions[id], store.Transaction{
		ID: "p1", AccountID: id, Status: store.StatusPending, AmountMinor: -77_000,
		Currency: "EUR", BookingDate: on(2026, time.September, 16), Remittance: "COMPRA HOLD 111111111111",
	})
	c := monthOf(t, svc, ada).Months[0]
	if c.Out.Minor != 42_500 {
		t.Errorf("out = %d, want the pending payment left out", c.Out.Minor)
	}
}

func TestEachCurrencyHasItsOwnSummary(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Euro", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Sterling", Currency: "GBP"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: eur(1000), Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: banking.Money{Minor: 1000, Currency: "GBP"}, Kind: "CLAV"})
	done := connectBank(t, svc, gw, st)
	if _, err := svc.Accounts(context.Background(), ada, false); err != nil {
		t.Fatal(err)
	}
	st.transactions[done.Accounts[0].ID] = []store.Transaction{{
		ID: "e", AccountID: done.Accounts[0].ID, Status: store.StatusBooked, AmountMinor: -300,
		Currency: "EUR", BookingDate: on(2026, time.September, 2), Remittance: "COMPRA A 111111111111"}}
	st.transactions[done.Accounts[1].ID] = []store.Transaction{{
		ID: "g", AccountID: done.Accounts[1].ID, Status: store.StatusBooked, AmountMinor: -900,
		Currency: "GBP", BookingDate: on(2026, time.September, 2), Remittance: "COMPRA B 111111111111"}}

	m := monthOf(t, svc, ada)
	if len(m.Months) != 2 {
		t.Fatalf("got %d months, want one per currency", len(m.Months))
	}
	if m.Months[0].Currency != "GBP" || m.Months[0].Out.Minor != 900 || m.Months[1].Out.Minor != 300 {
		t.Errorf("got %+v, want the currency with the most money out first, never combined", m.Months)
	}
}

func TestAMemberWhoOwnsNothingGetsNoSummary(t *testing.T) {
	svc, _, _ := monthLedger(t)
	if m := monthOf(t, svc, grace); len(m.Months) != 0 {
		t.Errorf("got %+v, want nothing", m.Months)
	}
}

func TestACurrencyWithNothingBookedThisMonthGetsNoSummary(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedPayment(st, "a1", id, -100, on(2026, time.August, 14), "COMPRA X 111111111111")
	if m := monthOf(t, svc, ada); len(m.Months) != 0 {
		t.Errorf("got %+v, want nothing", m.Months)
	}
}

func TestSeveralPaymentsToOneMerchantAreOneRowWithTheirTotalAndCount(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "anchor", id, -1, "EUR", on(2026, time.July, 1))
	for i, d := range []int{2, 5, 9, 12} {
		seedPayment(st, "p"+string(rune('a'+i)), id, -10_000, on(2026, time.September, d), "COMPRA PINGO DOCE 55555555555"+string(rune('0'+i)))
	}
	seedPayment(st, "q", id, -250, on(2026, time.September, 6), "pingo doce")

	c := monthOf(t, svc, ada).Months[0]
	if len(c.TopMerchants) != 1 {
		t.Fatalf("got %+v, want one merchant", c.TopMerchants)
	}
	got := c.TopMerchants[0]
	if got.Name != "Pingo Doce" || got.Total.Minor != 40_250 || got.Payments != 5 {
		t.Errorf("got %+v, want Pingo Doce 40250 over 5 payments, the lower-cased name grouping them", got)
	}
}

func TestTheRankingIsByMoneyNotByVisits(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "anchor", id, -1, "EUR", on(2026, time.July, 1))
	seedPayment(st, "big", id, -90_000, on(2026, time.September, 3), "COMPRA LANDLORD 111111111111")
	for i := 0; i < 10; i++ {
		seedPayment(st, "c"+string(rune('a'+i)), id, -2_000, on(2026, time.September, 4), "COMPRA COFFEE 22222222222"+string(rune('0'+i)))
	}
	c := monthOf(t, svc, ada).Months[0]
	if c.TopMerchants[0].Name != "Landlord" || c.TopMerchants[1].Name != "Coffee" {
		t.Errorf("got %+v, want the single large payment above ten small ones", c.TopMerchants)
	}
}

func TestMoneyComingInIsNotSpendingAndFewerThanFiveStandAlone(t *testing.T) {
	svc, _, _ := monthLedger(t)
	c := monthOf(t, svc, ada).Months[0]

	if len(c.TopMerchants) != 2 || len(c.LargestPayments) != 2 {
		t.Fatalf("got %d merchants and %d payments, want the two paid and nothing standing in", len(c.TopMerchants), len(c.LargestPayments))
	}
	for _, m := range c.TopMerchants {
		if m.Name == "Salary" {
			t.Error("a salary is listed as a merchant")
		}
	}
}

func TestTopSpendingIsCappedAtFiveOfEach(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "anchor", id, -1, "EUR", on(2026, time.July, 1))
	for i := 0; i < 8; i++ {
		seedPayment(st, "m"+string(rune('a'+i)), id, int64(-1_000*(i+1)), on(2026, time.September, 2+i), "COMPRA SHOP"+string(rune('A'+i))+" 111111111111")
	}
	c := monthOf(t, svc, ada).Months[0]
	if len(c.TopMerchants) != 5 || len(c.LargestPayments) != 5 {
		t.Fatalf("got %d merchants, %d payments, want five each", len(c.TopMerchants), len(c.LargestPayments))
	}
	if c.LargestPayments[0].AmountMinor != -8_000 || c.LargestPayments[4].AmountMinor != -4_000 {
		t.Errorf("largest payments = %+v, want -8000 down to -4000", c.LargestPayments)
	}
	if m := monthOf(t, svc, ada); m.Labels[c.LargestPayments[0].AccountID].Name == "" {
		t.Error("a payment does not say which account it left")
	}
}
