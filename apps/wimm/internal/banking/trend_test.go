package banking_test

import (
	"context"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/banking/bankingtest"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// trendAccount connects one owned account at Montepio and gives it a current
// balance. A connection creates the account; only a read (skipRead=false)
// records its balance, the way BalanceTrend and totals() both require one.
func trendAccount(t *testing.T, svc *banking.Service, gw *bankingtest.Gateway, st *memStore, currency string, balance int64) (accountID, connectionID string) {
	t.Helper()
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Conta à Ordem", Currency: currency})
	gw.SetBalance(montepio().ID, "hash-1", banking.Balance{Money: banking.Money{Minor: balance, Currency: currency}, Kind: "CLAV"})
	done := connectBank(t, svc, gw, st)
	if _, err := svc.Accounts(context.Background(), ada, false); err != nil {
		t.Fatalf("Accounts (recording the balance): %v", err)
	}
	return done.Accounts[0].ID, done.Connection.ID
}

// seedTransaction writes one booked transaction directly onto the store,
// which is the only way to give a test a known, controlled history: the
// bound the trend walks back to is the earliest one of these, never
// transactions_synced_through (an exhausted sync sets that to today however
// far back the history it exhausted actually reaches).
func seedTransaction(st *memStore, id, accountID string, amountMinor int64, currency string, bookingDate time.Time) {
	st.transactions[accountID] = append(st.transactions[accountID], store.Transaction{
		ID: id, AccountID: accountID, Status: store.StatusBooked,
		AmountMinor: amountMinor, Currency: currency, BookingDate: bookingDate,
	})
}

func TestTrendWalksBackOnlyAsFarAsItsOldestTransaction(t *testing.T) {
	svc, gw, st := newService(t)
	accountID, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)

	bound := st.now.AddDate(0, 0, -30)
	seedTransaction(st, "t0", accountID, -1, "EUR", bound)
	seedTransaction(st, "t1", accountID, -20_000, "EUR", st.now.AddDate(0, 0, -10))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatalf("BalanceTrend: %v", err)
	}
	if len(trend.Currencies) != 1 {
		t.Fatalf("got %d currencies, want 1", len(trend.Currencies))
	}
	points := trend.Currencies[0].Points
	if len(points) != 31 {
		t.Fatalf("got %d points, want one a day from the bound to today, 31", len(points))
	}

	// One point per UTC day, each at the end of its day, the last being now.
	// The oldest is the end of the day of the account's earliest transaction.
	wantFirst := time.Date(bound.Year(), bound.Month(), bound.Day(), 23, 59, 59, 0, time.UTC)
	if !points[0].Date.Equal(wantFirst) {
		t.Errorf("oldest point = %v, want the end of %v", points[0].Date, wantFirst)
	}
	if !points[len(points)-1].Date.Equal(st.now) {
		t.Errorf("newest point = %v, want now (%v)", points[len(points)-1].Date, st.now)
	}
	for i := 1; i < len(points); i++ {
		if got := points[i].Date.Sub(points[i-1].Date); i < len(points)-1 && got != 24*time.Hour {
			t.Fatalf("points %d and %d are %v apart, want a day", i-1, i, got)
		}
	}

	// The newest point is the current balance; the -10-day debit has not
	// been undone yet. It is offset by the 1-minor seed transaction dated
	// exactly at the bound, which the walk must count as already applied.
	if got := points[len(points)-1].Money.Minor; got != 500_000 {
		t.Errorf("newest balance = %d, want 500000", got)
	}
	// A point before the -10-day debit has it undone: 500000 - (-20000) =
	// 520000, the balance before that debit posted.
	if got := points[0].Money.Minor; got != 520_000 {
		t.Errorf("oldest balance = %d, want 520000 (the debit undone)", got)
	}
	if trend.Currencies[0].ShortHistory {
		t.Error("ShortHistory set on a month of history")
	}
}

// An account whose connection was never granted transaction access
// contributes to the total but not to the trend, and the trend says its
// coverage is partial rather than drawing a line as if every account fed it.
func TestAnAccountWithBalancesOnlyContributesToTotalNotTrend(t *testing.T) {
	svc, gw, st := newService(t)

	wideID, _ := trendAccount(t, svc, gw, st, "EUR", 300_000)
	seedTransaction(st, "t0", wideID, -1, "EUR", st.now.AddDate(0, 0, -20))

	// A second account, at a connection whose consent covers balances only.
	gw.AddBank(banking.Bank{ID: "PT:ActivoBank", Name: "ActivoBank", Country: "PT", MaxConsent: 24 * time.Hour},
		banking.Account{Ref: "hash-2", Name: "Savings", Currency: "EUR"})
	gw.SetBalance("PT:ActivoBank", "hash-2", banking.Balance{Money: banking.Money{Minor: 100_000, Currency: "EUR"}, Kind: "CLAV"})
	if _, err := svc.BeginConnection(context.Background(), ada, "PT:ActivoBank"); err != nil {
		t.Fatalf("BeginConnection: %v", err)
	}
	narrow, err := svc.CompleteConnection(context.Background(), ada,
		banking.Callback{Code: "code", State: gw.LastState})
	if err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}
	st.scopes[narrow.Connection.ID] = store.ScopeBalances
	if _, err := svc.Accounts(context.Background(), ada, false); err != nil {
		t.Fatalf("Accounts (recording the second balance): %v", err)
	}

	view, err := svc.Accounts(context.Background(), ada, true)
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	if len(view.OwnTotals) != 1 || view.OwnTotals[0].Money.Minor != 400_000 {
		t.Fatalf("total = %+v, want both accounts summed to 400000", view.OwnTotals)
	}

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatalf("BalanceTrend: %v", err)
	}
	if len(trend.Currencies) != 1 {
		t.Fatalf("got %d currencies, want 1", len(trend.Currencies))
	}
	// Only the wide account's balance, not both.
	newest := trend.Currencies[0].Points[len(trend.Currencies[0].Points)-1]
	if newest.Money.Minor != 300_000 {
		t.Errorf("newest trend balance = %d, want 300000 (the wide account only)", newest.Money.Minor)
	}
	if !trend.PartialCoverage {
		t.Error("PartialCoverage = false, want true: an owned account was excluded")
	}
}

// wimm holds no rates: two currencies never combine into one series.
func TestTwoCurrenciesAreNeverCombinedInTheTrend(t *testing.T) {
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Euro account", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "Sterling account", Currency: "GBP"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: banking.Money{Minor: 200_000, Currency: "EUR"}, Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: banking.Money{Minor: 150_000, Currency: "GBP"}, Kind: "CLAV"})
	done := connectBank(t, svc, gw, st)
	if _, err := svc.Accounts(context.Background(), ada, false); err != nil {
		t.Fatalf("Accounts (recording balances): %v", err)
	}

	bound := st.now.AddDate(0, 0, -15)
	for _, a := range done.Accounts {
		seedTransaction(st, "seed-"+a.ID, a.ID, -1, a.Currency, bound)
	}

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatalf("BalanceTrend: %v", err)
	}
	if len(trend.Currencies) != 2 {
		t.Fatalf("got %d currencies, want 2", len(trend.Currencies))
	}
	seen := map[string]int64{}
	for _, c := range trend.Currencies {
		seen[c.Currency] = c.Points[len(c.Points)-1].Money.Minor
	}
	if seen["EUR"] != 200_000 || seen["GBP"] != 150_000 {
		t.Errorf("trend balances = %+v, want EUR 200000 and GBP 150000, never combined", seen)
	}
}

// An owned account with wide scope but no transaction ever read yet
// contributes no trend for its currency — not a scope problem (no
// PartialCoverage), just nothing to walk.
func TestNoTrendWhenAnAccountHasNeverSyncedAnyTransaction(t *testing.T) {
	svc, gw, st := newService(t)
	trendAccount(t, svc, gw, st, "EUR", 200_000)
	// No transactions seeded: Oldest stays nil for this account.

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatalf("BalanceTrend: %v", err)
	}
	if len(trend.Currencies) != 0 {
		t.Fatalf("got %d currencies, want none", len(trend.Currencies))
	}
	if trend.PartialCoverage {
		t.Error("PartialCoverage = true, want false: the account's scope is wide, it just has no history yet")
	}
}

// A member who owns nothing sees no trend, the same absence as no history at
// all — there is no owned account to walk in the first place.
func TestNoTrendForAMemberWhoOwnsNothing(t *testing.T) {
	svc, gw, st := newService(t)
	trendAccount(t, svc, gw, st, "EUR", 200_000)

	trend, err := svc.BalanceTrend(context.Background(), grace, banking.ScopeAll)
	if err != nil {
		t.Fatalf("BalanceTrend: %v", err)
	}
	if len(trend.Currencies) != 0 {
		t.Fatalf("got %d currencies, want none", len(trend.Currencies))
	}
	if trend.PartialCoverage {
		t.Error("PartialCoverage = true, want false: grace owns nothing to exclude")
	}
}

func day(now time.Time, back int) time.Time {
	d := now.AddDate(0, 0, -back)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
}

// twoEuroAccounts connects two EUR accounts, both owned by ada, balances 300000
// and 200000.
func twoEuroAccounts(t *testing.T) (*banking.Service, *memStore, string, string) {
	t.Helper()
	svc, gw, st := newService(t)
	gw.AddBank(montepio(),
		banking.Account{Ref: "hash-1", Name: "Old", Currency: "EUR"},
		banking.Account{Ref: "hash-2", Name: "New", Currency: "EUR"})
	gw.SetBalance("PT:Montepio", "hash-1", banking.Balance{Money: eur(300_000), Kind: "CLAV"})
	gw.SetBalance("PT:Montepio", "hash-2", banking.Balance{Money: eur(200_000), Kind: "CLAV"})
	done := connectBank(t, svc, gw, st)
	if _, err := svc.Accounts(context.Background(), ada, false); err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	return svc, st, done.Accounts[0].ID, done.Accounts[1].ID
}

func TestTheChartCoversAtMostNinetyDays(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "t0", id, -100, "EUR", day(st.now, 200))
	seedTransaction(st, "t1", id, -100, "EUR", day(st.now, 3))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	points := trend.Currencies[0].Points
	if len(points) != 90 {
		t.Fatalf("got %d points, want 90", len(points))
	}
	if !points[len(points)-1].Date.Equal(st.now) || points[len(points)-1].Money.Minor != 500_000 {
		t.Errorf("the chart must end on now at today's balance, got %+v", points[len(points)-1])
	}
}

func TestAnAccountConnectedYesterdayDoesNotShortenTheChart(t *testing.T) {
	svc, st, longID, shortID := twoEuroAccounts(t)
	seedTransaction(st, "l0", longID, -100, "EUR", day(st.now, 90))
	seedTransaction(st, "l1", longID, -5_000, "EUR", day(st.now, 40))
	seedTransaction(st, "s0", shortID, -100, "EUR", day(st.now, 1))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(trend.Currencies) != 1 {
		t.Fatalf("got %d currencies", len(trend.Currencies))
	}
	c := trend.Currencies[0]
	if len(c.Points) < 80 {
		t.Errorf("got %d points, want the long account's ~90 days", len(c.Points))
	}
	if got := c.Points[len(c.Points)-1].Money.Minor; got != 300_000 {
		t.Errorf("newest = %d, want the long account alone, 300000", got)
	}
	if !trend.PartialCoverage {
		t.Error("PartialCoverage = false, want true: the short account was left out")
	}
}

func TestWhenNoAccountReachesThirtyDaysTheLongestAloneContributes(t *testing.T) {
	svc, st, a, b := twoEuroAccounts(t)
	seedTransaction(st, "a0", a, -100, "EUR", day(st.now, 20))
	seedTransaction(st, "b0", b, -100, "EUR", day(st.now, 10))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	c := trend.Currencies[0]
	if len(c.Points) != 21 {
		t.Errorf("got %d points, want the 20-day account's 21", len(c.Points))
	}
	if got := c.Points[len(c.Points)-1].Money.Minor; got != 300_000 {
		t.Errorf("newest = %d, want the longest account alone", got)
	}
	if !trend.PartialCoverage {
		t.Error("PartialCoverage = false, want true")
	}
}

func TestLessThanAWeekOfHistoryDrawsNoChartAndSaysWhy(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "t0", id, -100, "EUR", day(st.now, 6))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(trend.Currencies) != 1 {
		t.Fatalf("got %d currencies, want 1 carrying the reason", len(trend.Currencies))
	}
	if c := trend.Currencies[0]; !c.ShortHistory || len(c.Points) != 0 {
		t.Errorf("got ShortHistory=%v with %d points, want the flag and no points", c.ShortHistory, len(c.Points))
	}
}

func TestAWeekOfHistoryIsEnough(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "t0", id, -100, "EUR", day(st.now, 7))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	if c := trend.Currencies[0]; c.ShortHistory || len(c.Points) != 8 {
		t.Errorf("ShortHistory=%v, %d points, want a chart of 8", c.ShortHistory, len(c.Points))
	}
}

func TestAQuietCurrencyGetsNoTrend(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "USD", 0)
	seedTransaction(st, "old", id, 0, "USD", day(st.now, 200))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(trend.Currencies) != 0 {
		t.Errorf("got %+v, want no trend for a currency at zero with nothing moving", trend.Currencies)
	}
	if trend.PartialCoverage {
		t.Error("a quiet currency must not make the chart claim partial coverage")
	}
}

func TestAZeroCurrencyThatMovedRecentlyIsShown(t *testing.T) {
	svc, gw, st := newService(t)
	id, _ := trendAccount(t, svc, gw, st, "USD", 0)
	seedTransaction(st, "in", id, 5_000, "USD", day(st.now, 30))
	seedTransaction(st, "out", id, -5_000, "USD", day(st.now, 2))

	trend, err := svc.BalanceTrend(context.Background(), ada, banking.ScopeAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(trend.Currencies) != 1 || len(trend.Currencies[0].Points) == 0 {
		t.Fatalf("got %+v, want a chart for a balance that reached zero", trend.Currencies)
	}
}

func pointOn(t *testing.T, trend banking.Trend, date time.Time) banking.TrendPoint {
	t.Helper()
	if len(trend.Currencies) != 1 {
		t.Fatalf("got %d currencies, want 1", len(trend.Currencies))
	}
	for _, p := range trend.Currencies[0].Points {
		if p.Date.Year() == date.Year() && p.Date.YearDay() == date.YearDay() {
			return p
		}
	}
	t.Fatalf("no point on %s", date.Format("2 Jan"))
	return banking.TrendPoint{}
}

func TestADayNamesTheRowsThatMovedItAndCountsTheSmallerOnes(t *testing.T) {
	svc, gw, st := newService(t)
	accountID, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "anchor", accountID, -1, "EUR", st.now.AddDate(0, 0, -40))
	day := on(2026, time.September, 5)
	for id, minor := range map[string]int64{"big": -60_000, "mid": -30_000, "s1": -1_000, "s2": -900, "s3": -800} {
		seedTransaction(st, id, accountID, minor, "EUR", day)
	}

	p := pointOn(t, mustTrend(t, svc, ada), day)
	if len(p.Movers) != 2 || p.Movers[0].Amount != -60_000 || p.Movers[1].Amount != -30_000 || p.Smaller != 3 {
		t.Errorf("got %+v and %d smaller, want the two large ones and 3 smaller", p.Movers, p.Smaller)
	}
}

func TestADayWithNothingOnItHasNoMovers(t *testing.T) {
	svc, gw, st := newService(t)
	accountID, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "anchor", accountID, -1, "EUR", st.now.AddDate(0, 0, -40))

	p := pointOn(t, mustTrend(t, svc, ada), on(2026, time.September, 9))
	if len(p.Movers) != 0 || p.Smaller != 0 {
		t.Errorf("got %+v and %d smaller, want an empty day", p.Movers, p.Smaller)
	}
}

func TestADayHoldingMoneyInNamesItWithAPlus(t *testing.T) {
	svc, gw, st := newService(t)
	accountID, _ := trendAccount(t, svc, gw, st, "EUR", 500_000)
	seedTransaction(st, "anchor", accountID, -1, "EUR", st.now.AddDate(0, 0, -40))
	seedTransaction(st, "salary", accountID, 245_000, "EUR", on(2026, time.September, 15))
	seedTransaction(st, "tip", accountID, -500, "EUR", on(2026, time.September, 15))

	p := pointOn(t, mustTrend(t, svc, ada), on(2026, time.September, 15))
	if len(p.Movers) != 1 || p.Movers[0].Amount != 245_000 || p.Smaller != 1 {
		t.Errorf("got %+v and %d smaller, want the salary and 1 smaller", p.Movers, p.Smaller)
	}
}

func TestAnAccountTheChartDropsAppearsOnNoDay(t *testing.T) {
	svc, st, done := twoAccounts(t)
	long, short := done.Accounts[0], done.Accounts[1]
	seedTransaction(st, "anchor", long.ID, -1, "EUR", st.now.AddDate(0, 0, -40))
	seedTransaction(st, "short-start", short.ID, -1, "EUR", st.now.AddDate(0, 0, -10))
	seedTransaction(st, "short-big", short.ID, -70_000, "EUR", on(2026, time.September, 10))
	if _, err := svc.Accounts(context.Background(), ada, false); err != nil {
		t.Fatal(err)
	}

	trend := mustTrend(t, svc, ada)
	for _, p := range trend.Currencies[0].Points {
		for _, m := range p.Movers {
			if m.Amount == -70_000 {
				t.Errorf("a dropped account's payment moved %s", p.Date.Format("2 Jan"))
			}
		}
	}
}

func mustTrend(t *testing.T, svc *banking.Service, member string) banking.Trend {
	t.Helper()
	trend, err := svc.BalanceTrend(context.Background(), member, banking.ScopeAll)
	if err != nil {
		t.Fatalf("BalanceTrend: %v", err)
	}
	return trend
}
