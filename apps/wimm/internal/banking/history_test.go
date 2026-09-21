package banking_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// theTwentieth is the day every history test opens Overview: 20 September 2026.
var theTwentieth = time.Date(2026, time.September, 20, 9, 0, 0, 0, time.UTC)

func seedRow(st *memStore, id, accountID string, minor int64, date time.Time, counterparty string) {
	st.transactions[accountID] = append(st.transactions[accountID], store.Transaction{
		ID: id, AccountID: accountID, Status: store.StatusBooked,
		AmountMinor: minor, Currency: "EUR", BookingDate: date, CounterpartyName: counterparty,
	})
}

// seedMonthly puts one row a month on a day, from first to last inclusive.
func seedMonthly(st *memStore, prefix, accountID string, minor int64, day int, counterparty string, first, last time.Time) {
	for m := first; !m.After(last); m = m.AddDate(0, 1, 0) {
		seedRow(st, prefix+m.Format("-2006-01"), accountID, minor, time.Date(m.Year(), m.Month(), day, 0, 0, 0, 0, time.UTC), counterparty)
	}
}

// yearLedger is one EUR account held since August 2024 with a €2,450.00 salary
// on the 1st and €820.00 of rent on the 6th of every month to September 2026:
// every month nets +€1,630.00.
func yearLedger(t *testing.T) (*banking.Service, *memStore, string) {
	t.Helper()
	svc, gw, st := newService(t)
	st.now = theTwentieth
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedRow(st, "anchor", id, -1, on(2024, time.August, 1), "Anchor")
	first, last := on(2024, time.September, 1), on(2026, time.September, 1)
	seedMonthly(st, "sal", id, 245_000, 1, "Salary", first, last)
	seedMonthly(st, "rent", id, -82_000, 6, "Rent", first, last)
	return svc, st, id
}

func historyOf(t *testing.T, svc *banking.Service, member string, scope banking.Scope) banking.History {
	t.Helper()
	h, err := svc.MonthHistory(context.Background(), member, scope)
	if err != nil {
		t.Fatalf("MonthHistory: %v", err)
	}
	return h
}

func onlyCurrency(t *testing.T, h banking.History) banking.CurrencyHistory {
	t.Helper()
	if len(h.Currencies) != 1 {
		t.Fatalf("got %d currencies, want 1", len(h.Currencies))
	}
	return h.Currencies[0]
}

func monthNamed(t *testing.T, c banking.CurrencyHistory, y int, m time.Month) banking.HistoryMonth {
	t.Helper()
	for _, month := range c.Months {
		if month.Start.Equal(on(y, m, 1)) {
			return month
		}
	}
	t.Fatalf("no month %s %d among %d", m, y, len(c.Months))
	return banking.HistoryMonth{}
}

func TestAYearOfMonthsIsTheMonthSoFarAndTwelveFullOnes(t *testing.T) {
	svc, _, _ := yearLedger(t)
	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))

	if len(c.Months) != 13 || c.FullMonths != 12 {
		t.Fatalf("got %d months, %d full, want 13 and 12", len(c.Months), c.FullMonths)
	}
	if now := c.Months[0]; !now.Start.Equal(on(2026, time.September, 1)) || !now.SoFar || now.Full {
		t.Errorf("newest = %+v, want September so far and not full", now)
	}
	if oldest := c.Months[12]; !oldest.Start.Equal(on(2025, time.September, 1)) || !oldest.Full {
		t.Errorf("oldest = %+v, want September 2025, full", oldest)
	}
	if aug := monthNamed(t, c, 2026, time.August); aug.In.Minor != 245_000 || aug.Out.Minor != 82_000 || aug.Net.Minor != 163_000 {
		t.Errorf("August = %+v, want in 245000, out 82000, net 163000", aug)
	}
	if got := c.Months[0].Net.Minor; got != 163_000 {
		t.Errorf("September so far net = %d, want 163000: both rows are booked by the 20th", got)
	}
}

func TestALedgerFromTheTwelfthOfAprilHoldsApril(t *testing.T) {
	svc, gw, st := newService(t)
	st.now = theTwentieth
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedRow(st, "anchor", id, -1, on(2026, time.April, 12), "Anchor")
	seedMonthly(st, "sal", id, 245_000, 15, "Salary", on(2026, time.April, 1), on(2026, time.September, 1))

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if len(c.Months) != 6 {
		t.Fatalf("got %d months, want April to September and nothing before", len(c.Months))
	}
	april := monthNamed(t, c, 2026, time.April)
	if april.Full || april.HeldFrom == nil || !april.HeldFrom.Equal(on(2026, time.April, 12)) {
		t.Errorf("April = %+v, want held from 12 April and not full", april)
	}
	if c.FullMonths != 4 {
		t.Errorf("FullMonths = %d, want May to August", c.FullMonths)
	}
	if monthNamed(t, c, 2026, time.September).HeldFrom != nil {
		t.Error("September claims a ledger begins inside it")
	}
}

func TestAYoungAccountBesideAnOldOneKeepsMonthsFullAndIsNamedAsMissing(t *testing.T) {
	svc, st, done := twoAccounts(t)
	st.now = theTwentieth
	old, fresh := done.Accounts[0], done.Accounts[1]
	seedRow(st, "anchor", old.ID, -1, on(2025, time.January, 1), "Anchor")
	seedMonthly(st, "sal", old.ID, 245_000, 1, "Salary", on(2025, time.January, 1), on(2026, time.September, 1))
	seedRow(st, "new", fresh.ID, -1_000, on(2026, time.August, 12), "Coffee")

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if c.FullMonths != 12 {
		t.Errorf("FullMonths = %d, want 12: the old ledger holds every month", c.FullMonths)
	}
	if aug := monthNamed(t, c, 2026, time.August); !aug.Full || aug.HeldFrom != nil {
		t.Errorf("August = %+v, want full: the earliest ledger began long before", aug)
	}
	if len(c.LateLedgers) != 1 || c.LateLedgers[0].AccountName != "Personal" ||
		!c.LateLedgers[0].From.Equal(on(2026, time.August, 12)) {
		t.Errorf("late ledgers = %+v, want Personal from 12 August", c.LateLedgers)
	}
}

func TestAnOldAccountAndAYoungOneGiveFullMonthsSinceTheOldOneBegan(t *testing.T) {
	svc, st, done := twoAccounts(t)
	st.now = theTwentieth
	old, young := done.Accounts[0], done.Accounts[1]
	seedRow(st, "anchor", old.ID, -1, on(2026, time.March, 19), "Anchor")
	seedMonthly(st, "sal", old.ID, 245_000, 1, "Salary", on(2026, time.April, 1), on(2026, time.September, 1))
	seedRow(st, "young", young.ID, -900, on(2026, time.September, 19), "Coffee")

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if c.FullMonths != 5 { // April to August
		t.Errorf("FullMonths = %d, want April to August", c.FullMonths)
	}
	if march := monthNamed(t, c, 2026, time.March); march.Full || march.HeldFrom == nil {
		t.Errorf("March = %+v, want held from 19 March", march)
	}
	if len(c.LateLedgers) != 1 || !c.LateLedgers[0].From.Equal(on(2026, time.September, 19)) {
		t.Errorf("late ledgers = %+v, want the young account from 19 September", c.LateLedgers)
	}
}

func TestAnAccountThatBeginsBeforeTheFirstFullMonthIsNotLate(t *testing.T) {
	svc, st, done := twoAccounts(t)
	st.now = theTwentieth
	old, other := done.Accounts[0], done.Accounts[1]
	seedRow(st, "anchor", old.ID, -1, on(2026, time.March, 19), "Anchor")
	seedMonthly(st, "sal", old.ID, 245_000, 3, "Salary", on(2026, time.April, 1), on(2026, time.September, 1))
	seedRow(st, "other", other.ID, -900, on(2026, time.March, 25), "Coffee")

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if len(c.LateLedgers) != 0 {
		t.Errorf("late ledgers = %+v, want none: the second account begins before the first full month", c.LateLedgers)
	}
}

func TestALedgerThatBeginsThisMonthHoldsNoFullMonth(t *testing.T) {
	svc, gw, st := newService(t)
	st.now = theTwentieth
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedRow(st, "a", id, -5_000, on(2026, time.September, 3), "Coffee")

	for _, c := range historyOf(t, svc, ada, banking.ScopeAll).Currencies {
		if c.FullMonths != 0 {
			t.Errorf("FullMonths = %d, want none", c.FullMonths)
		}
	}
}

func TestAMemberWhoOwnsNothingHasNoHistory(t *testing.T) {
	svc, _, _ := yearLedger(t)
	h := historyOf(t, svc, grace, banking.ScopeAll)
	if len(h.Currencies) != 0 || len(h.Available) != 0 {
		t.Errorf("got %+v, want nothing", h)
	}
}

func TestAMonthThatEndedBelowZeroReadsWithAMinus(t *testing.T) {
	svc, st, id := yearLedger(t)
	seedRow(st, "big", id, -300_000, on(2026, time.March, 20), "Landlord Two")

	march := monthNamed(t, onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll)), 2026, time.March)
	if march.Net.Minor != 163_000-300_000 {
		t.Errorf("March net = %d, want %d", march.Net.Minor, 163_000-300_000)
	}
}

func TestTwoFullMonthsStateNoTypicalOrAverageMonth(t *testing.T) {
	svc, gw, st := newService(t)
	st.now = theTwentieth
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedRow(st, "anchor", id, -1, on(2026, time.July, 1), "Anchor")
	seedMonthly(st, "sal", id, 245_000, 1, "Salary", on(2026, time.July, 1), on(2026, time.September, 1))

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if c.FullMonths != 2 {
		t.Fatalf("FullMonths = %d, want 2", c.FullMonths)
	}
	if c.TypicalNet != nil || c.AverageNet != nil || c.TypicalNetUsual != nil || c.AverageNetUsual != nil {
		t.Error("a typical or average month was stated from two full months")
	}
	for _, m := range c.Months {
		if len(m.Risers) != 0 {
			t.Errorf("%s lists risers with no usual to measure them against", m.Start.Format("Jan 2006"))
		}
	}
}

func TestTheTypicalMonthOfAnEvenNumberOfMonthsIsTheMeanOfTheMiddleTwoRoundedAwayFromZero(t *testing.T) {
	svc, gw, st := newService(t)
	st.now = theTwentieth
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedRow(st, "anchor", id, -1, on(2026, time.May, 1), "Anchor")
	for i, minor := range []int64{100, 201, 300, 401} { // May to August
		seedRow(st, "m"+string(rune('a'+i)), id, minor, on(2026, time.Month(5+i), 10), "Client")
	}

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if c.FullMonths != 4 {
		t.Fatalf("FullMonths = %d, want 4", c.FullMonths)
	}
	// The anchor is May's −1, so the nets are 99, 201, 300 and 401.
	if c.TypicalNet.Minor != 251 { // (201+300)/2 = 250.5
		t.Errorf("typical = %d, want 251", c.TypicalNet.Minor)
	}
	if c.AverageNet.Minor != 250 { // 1001/4 = 250.25
		t.Errorf("average = %d, want 250", c.AverageNet.Minor)
	}
}

func TestABonusIsUnusualIncomeAndComesOffTheNet(t *testing.T) {
	svc, st, id := yearLedger(t)
	seedRow(st, "bonus", id, 980_400, on(2026, time.May, 22), "Salary")

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	may := monthNamed(t, c, 2026, time.May)
	if len(may.Unusual) != 1 {
		t.Fatalf("May holds %d unusual payments, want the bonus", len(may.Unusual))
	}
	u := may.Unusual[0]
	if u.Transaction.ID != "bonus" || u.FirstPayment || u.Typical == nil || u.Typical.Minor != 245_000 {
		t.Errorf("got %+v, want the bonus, usually about 245000", u)
	}
	if may.Net.Minor != 163_000+980_400 || may.NetUsual == nil || may.NetUsual.Minor != 163_000 {
		t.Errorf("May net %d, without %v, want %d and 163000", may.Net.Minor, may.NetUsual, 163_000+980_400)
	}
	if c.UnusualCount != 1 {
		t.Errorf("UnusualCount = %d, want 1", c.UnusualCount)
	}
	if c.TypicalNet.Minor != 163_000 {
		t.Errorf("typical = %d, want 163000: one month barely moves the middle one", c.TypicalNet.Minor)
	}
	if want := int64((11*163_000 + 163_000 + 980_400 + 6) / 12); c.AverageNet.Minor != want {
		t.Errorf("average = %d, want %d", c.AverageNet.Minor, want)
	}
	if c.AverageNetUsual.Minor != 163_000 || c.TypicalNetUsual.Minor != 163_000 {
		t.Errorf("without unusual payments: typical %d, average %d, want 163000 for both", c.TypicalNetUsual.Minor, c.AverageNetUsual.Minor)
	}
}

func TestTheSalaryItselfIsNotUnusual(t *testing.T) {
	svc, _, _ := yearLedger(t)
	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if c.UnusualCount != 0 {
		t.Errorf("UnusualCount = %d, want none in a ledger of salaries and rent", c.UnusualCount)
	}
	for _, m := range c.Months {
		if m.NetUsual != nil {
			t.Errorf("%s states a net without unusual payments having held none", m.Start.Format("Jan 2006"))
		}
	}
}

func TestAFirstPaymentThatIsALargePartOfAMonthIsUnusual(t *testing.T) {
	svc, st, id := yearLedger(t)
	seedRow(st, "garage", id, -165_000, on(2026, time.March, 12), "Auto Reparadora")

	march := monthNamed(t, onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll)), 2026, time.March)
	if len(march.Unusual) != 1 {
		t.Fatalf("March holds %d unusual payments, want the garage", len(march.Unusual))
	}
	u := march.Unusual[0]
	if !u.FirstPayment || u.Typical != nil {
		t.Errorf("got %+v, want a first payment with no typical amount", u)
	}
	if march.NetUsual == nil || march.NetUsual.Minor != 163_000 {
		t.Errorf("March without it = %v, want 163000", march.NetUsual)
	}
}

func TestMerchantsThatTookMoreThanUsualAreListedLargestRiseFirst(t *testing.T) {
	svc, st, id := yearLedger(t)
	seedMonthly(st, "galp", id, -16_440, 10, "Galp", on(2024, time.September, 1), on(2026, time.September, 1))
	seedRow(st, "galp-extra", id, -8_240, on(2026, time.August, 25), "Galp")
	seedRow(st, "zara", id, -8_000, on(2026, time.August, 12), "Zara")
	seedRow(st, "zara-2", id, -100, on(2026, time.August, 13), "zara")

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	august := monthNamed(t, c, 2026, time.August)
	if len(august.Risers) != 2 {
		t.Fatalf("got %+v, want Galp and Zara", august.Risers)
	}
	galp, zara := august.Risers[0], august.Risers[1]
	if galp.Name != "Galp" || galp.Total.Minor != 24_680 || galp.Usual.Minor != 16_440 {
		t.Errorf("first riser = %+v, want Galp, 24680 against a usual 16440", galp)
	}
	if zara.Total.Minor != 8_100 || zara.Usual.Minor != 0 || zara.Payments != 2 {
		t.Errorf("second riser = %+v, want Zara once, both payments, not usually paid", zara)
	}
	if got := monthNamed(t, c, 2026, time.September).Risers; len(got) != 0 {
		t.Errorf("the month so far lists %+v, want none", got)
	}
}

func TestAMonthWhereNothingRoseListsNoRisers(t *testing.T) {
	svc, _, _ := yearLedger(t)
	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if got := monthNamed(t, c, 2026, time.June).Risers; len(got) != 0 {
		t.Errorf("June lists %+v, want none", got)
	}
}

func TestOnlyTheFiveLargestRisesAreListed(t *testing.T) {
	svc, st, id := yearLedger(t)
	for i, name := range []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot", "Golf"} {
		seedRow(st, "r"+name, id, -int64(1_000*(i+1)), on(2026, time.August, 14), name)
	}
	august := monthNamed(t, onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll)), 2026, time.August)
	if len(august.Risers) != 5 || august.Risers[0].Name != "Golf" {
		t.Errorf("got %+v, want the five largest, Golf first", august.Risers)
	}
}

func TestRecurringPaymentsCarryWhenTheNextIsExpected(t *testing.T) {
	svc, _, _ := yearLedger(t)
	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	if len(c.Recurring) != 1 {
		t.Fatalf("got %+v, want the rent", c.Recurring)
	}
	rent := c.Recurring[0]
	if !strings.EqualFold(rent.Name, "Rent") || rent.Cadence != banking.CadenceMonthly ||
		rent.Amount != -82_000 || !rent.Expected.Equal(on(2026, time.October, 6)) || rent.Late {
		t.Errorf("got %+v, want monthly rent of 82000 expected 6 October", rent)
	}
}

func TestAnInsurancePremiumPaidTwiceIsLikelyYearly(t *testing.T) {
	svc, st, id := yearLedger(t)
	seedRow(st, "fid-1", id, -38_600, on(2025, time.March, 3), "Fidelidade")
	seedRow(st, "fid-2", id, -38_600, on(2026, time.March, 3), "Fidelidade")

	c := onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll))
	for _, r := range c.Recurring {
		if strings.EqualFold(r.Name, "Fidelidade") {
			if r.Cadence != banking.CadenceYearly || !r.Likely || !r.Expected.Equal(on(2027, time.March, 3)) {
				t.Errorf("got %+v, want likely yearly, expected 3 March 2027", r)
			}
			return
		}
	}
	t.Error("Fidelidade is not listed")
}

func TestHouseholdOffersAllThreeChoicesAndNamesWhatItDoesNotCount(t *testing.T) {
	svc, gw, st := newService(t)
	st.now = theTwentieth
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Joint", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Personal", Currency: "EUR"},
		banking.Account{Ref: "hash-3", Name: "Joint savings", Currency: "EUR"})
	for _, ref := range []string{"hash-1", "hash-2", "hash-3"} {
		gw.SetBalance("PT:Montepio", ref, banking.Balance{Money: eur(100_000), Kind: "CLAV"})
	}
	done := connectBank(t, svc, gw, st)
	ctx := context.Background()
	if err := svc.SetOwners(ctx, ada, done.Accounts[0].ID, []string{ada, grace}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetOwners(ctx, ada, done.Accounts[2].ID, []string{grace}); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetLevel(ctx, grace, done.Accounts[2].ID, ada, store.LevelDetails); err != nil {
		t.Fatal(err)
	}

	h := historyOf(t, svc, ada, banking.ScopeHousehold)
	if h.Scope != banking.ScopeHousehold {
		t.Errorf("answered %v, want Household", h.Scope)
	}
	if got := len(h.Available); got != 3 {
		t.Errorf("available = %v, want Household, Yours and All", h.Available)
	}
	if len(h.Counted) != 1 || h.Counted[0] != "Joint" || len(h.NotCounted) != 1 || h.NotCounted[0] != "Joint savings" {
		t.Errorf("counted %v, not counted %v, want Joint and Joint savings", h.Counted, h.NotCounted)
	}
	if h := historyOf(t, svc, ada, banking.ScopeAll); len(h.Counted) != 0 || len(h.NotCounted) != 0 {
		t.Errorf("All names %v and %v, want no such sentence", h.Counted, h.NotCounted)
	}
}

func TestNoControlIsOfferedWhenOneChoiceWouldRemain(t *testing.T) {
	svc, _, _ := yearLedger(t)
	h := historyOf(t, svc, ada, banking.ScopeHousehold)
	if len(h.Available) != 0 || h.Scope != banking.ScopeAll {
		t.Errorf("answered %v with %v on offer, want All and nothing: ada owns one account of her own", h.Scope, h.Available)
	}
}

func TestTheChartLabelsTheSamePaymentsTheMonthDoes(t *testing.T) {
	svc, st, id := yearLedger(t)
	seedRow(st, "bonus", id, 980_400, on(2026, time.August, 28), "Salary")

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	p := pointOn(t, trend, on(2026, time.August, 28))
	if len(p.Movers) == 0 || p.Movers[0].TransactionID != "bonus" || !p.Movers[0].Unusual {
		t.Errorf("got %+v, want the bonus first and unusual", p.Movers)
	}
	for _, m := range p.Movers[1:] {
		if m.Unusual {
			t.Errorf("%+v is labelled unusual", m)
		}
	}
}

func TestTransactionsMarksExactlyTheRowsTheMonthsCount(t *testing.T) {
	svc, st, id := yearLedger(t)
	seedRow(st, "garage", id, -165_000, on(2026, time.March, 12), "Auto Reparadora")
	seedRow(st, "bonus", id, 980_400, on(2026, time.May, 22), "Salary")
	ctx := context.Background()

	counted := map[string]bool{}
	for _, m := range onlyCurrency(t, historyOf(t, svc, ada, banking.ScopeAll)).Months {
		for _, u := range m.Unusual {
			counted[u.Transaction.ID] = true
		}
	}
	if len(counted) != 2 {
		t.Fatalf("the months count %v, want the garage and the bonus", counted)
	}

	marked := map[string]bool{}
	req := banking.LedgerRequest{MemberID: ada, SkipSync: true}
	for range 3 {
		l, err := svc.Transactions(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		for _, tx := range l.Page.Transactions {
			if _, ok := l.Patterns.MarkFor(tx.ID); ok {
				marked[tx.ID] = true
			}
		}
		if !l.Page.HasOlder {
			break
		}
		last := l.Page.Transactions[len(l.Page.Transactions)-1]
		req.Cursor, req.Older = store.Cursor{BookingDate: last.BookingDate, ID: last.ID}, true
	}
	if len(marked) != len(counted) || !marked["garage"] || !marked["bonus"] {
		t.Errorf("Transactions marks %v, the months count %v", marked, counted)
	}
}
