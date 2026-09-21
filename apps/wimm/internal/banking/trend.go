package banking

import (
	"context"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

const (
	// chartDays is the most days a balance chart covers, today included.
	chartDays = 90
	// minChartSpan is the fewest days between a chart's first point and today
	// for it to mean anything.
	minChartSpan = 7
	// longReach is how far back an account's ledger must reach to be one that
	// sets the chart's span on its own.
	longReach = 30
)

// CurrencyTrend is one currency's balance history, oldest point first.
type CurrencyTrend struct {
	Currency string
	Points   []TrendPoint
	// ShortHistory is set when no contributing ledger reaches back a week.
	// There are no Points then: the screen says a chart appears once there is
	// one, which is different from a currency with no transaction access.
	ShortHistory bool
}

// TrendPoint is the member's summed balance, in one currency, at the end of one
// UTC day; the last is now.
type TrendPoint struct {
	Date  time.Time
	Money Money
	// Movers are the rows that moved the day, largest first, and Smaller counts
	// the rest. Only rows of the accounts the chart is drawn from.
	Movers  []DayMover
	Smaller int
}

// Trend is what Overview's balance chart needs.
type Trend struct {
	Currencies []CurrencyTrend
	// PartialCoverage is true when an owned account counted in a figure is not
	// in the chart: its connection was never granted transaction access, or its
	// history is too short beside another's. The chart says so rather than
	// drawing a line as if it covered every account (`banking/overview`).
	PartialCoverage bool
}

// utcDay is midnight UTC of t's day, which is what booking_date is.
func utcDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int {
	return int(utcDay(to).Sub(utcDay(from)).Hours() / 24)
}

// BalanceTrend walks each of a member's owned accounts' booked transactions
// backward from its current balance, one point per UTC day, per currency. It
// covers only as far back as the accounts that set the span actually reach —
// a chart never draws a line across a gap it has no transactions for, and one
// account with a short history never shortens it for the rest
// (`banking/overview`).
func (s *Service) BalanceTrend(ctx context.Context, memberID string, asked Scope) (Trend, error) {
	scope, err := s.scopedAccounts(ctx, memberID, asked)
	if err != nil {
		return Trend{}, err
	}
	accounts, err := s.store.OwnedAccountsForTrend(ctx, memberID, scope.AccountIDs)
	if err != nil {
		return Trend{}, err
	}

	now, err := s.store.Now(ctx)
	if err != nil {
		return Trend{}, err
	}

	figured, err := s.currenciesWithFigures(ctx, memberID)
	if err != nil {
		return Trend{}, err
	}
	pats, err := s.patternsFor(ctx, memberID, scope, accounts, now.T)
	if err != nil {
		return Trend{}, err
	}

	byCurrency := map[string][]store.TrendAccount{}
	var order []string
	for _, a := range accounts {
		if a.BalanceMinor == nil || a.Currency == "" || a.Currency == "XXX" {
			continue
		}
		if _, ok := byCurrency[a.Currency]; !ok {
			order = append(order, a.Currency)
		}
		byCurrency[a.Currency] = append(byCurrency[a.Currency], a)
	}

	windowStart := utcDay(now.T).AddDate(0, 0, -(chartDays - 1))

	var out Trend
	for _, currency := range order {
		group := byCurrency[currency]

		quiet, err := s.quietCurrency(ctx, currency, group, figured, windowStart)
		if err != nil {
			return Trend{}, err
		}
		if quiet {
			continue
		}

		trend, partial, err := s.currencyTrend(ctx, currency, group, now.T, windowStart, pats)
		if err != nil {
			return Trend{}, err
		}
		out.PartialCoverage = out.PartialCoverage || partial
		if trend.ShortHistory || len(trend.Points) > 0 {
			out.Currencies = append(out.Currencies, trend)
		}
	}
	return out, nil
}

// currenciesWithFigures is every currency in which the member's household or
// own figure is not zero.
func (s *Service) currenciesWithFigures(ctx context.Context, memberID string) (map[string]bool, error) {
	visible, err := s.store.VisibleAccounts(ctx, memberID, "")
	if err != nil {
		return nil, err
	}
	groups, err := s.groupAccounts(ctx, visible)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, g := range []Group{GroupHousehold, GroupOwn} {
		for _, t := range totals(visible, groups, g) {
			if t.Money.Minor != 0 {
				out[t.Money.Currency] = true
			}
		}
	}
	return out, nil
}

// quietCurrency is a currency whose figures are zero and in which nothing was
// booked in the chart window. It gets no chart, no summary and no tile: a flat
// line over a currency nobody uses is not information.
func (s *Service) quietCurrency(
	ctx context.Context, currency string, accounts []store.TrendAccount,
	figured map[string]bool, windowStart time.Time,
) (bool, error) {
	if figured[currency] {
		return false, nil
	}
	for _, a := range accounts {
		if !a.Scope.ReadsTransactions() || a.Oldest == nil {
			continue
		}
		txs, err := s.store.TransactionsSince(ctx, a.ID, windowStart)
		if err != nil {
			return false, err
		}
		if len(txs) > 0 {
			return false, nil
		}
	}
	return true, nil
}

// currencyTrend picks which accounts set the span, then sums each one's own
// walk-back onto one point per day.
//
// Among accounts with transaction access and a ledger, those reaching back 30
// days or more contribute; where none does, only the single longest does, so
// the shortest never sets the span. Joining each account on its own first day
// instead would make the line step up by a whole balance the day a history
// begins, which reads as money arriving.
func (s *Service) currencyTrend(
	ctx context.Context, currency string, accounts []store.TrendAccount, now, windowStart time.Time,
	pats Patterns,
) (CurrencyTrend, bool, error) {
	partial := false
	var candidates []store.TrendAccount
	for _, a := range accounts {
		switch {
		case !a.Scope.ReadsTransactions():
			partial = true
		case a.Oldest == nil:
			// Synced and holding nothing, or never synced: nothing to walk.
		default:
			candidates = append(candidates, a)
		}
	}
	if len(candidates) == 0 {
		return CurrencyTrend{Currency: currency}, partial, nil
	}

	var contribute []store.TrendAccount
	for _, a := range candidates {
		if daysBetween(*a.Oldest, now) >= longReach {
			contribute = append(contribute, a)
		}
	}
	if len(contribute) == 0 {
		longest := candidates[0]
		for _, a := range candidates[1:] {
			if a.Oldest.Before(*longest.Oldest) {
				longest = a
			}
		}
		contribute = []store.TrendAccount{longest}
	}
	if len(contribute) < len(accounts) {
		partial = true
	}

	bound := utcDay(*contribute[0].Oldest)
	for _, a := range contribute[1:] {
		if d := utcDay(*a.Oldest); d.After(bound) {
			bound = d
		}
	}
	if bound.Before(windowStart) {
		bound = windowStart
	}
	if daysBetween(bound, now) < minChartSpan {
		return CurrencyTrend{Currency: currency, ShortHistory: true}, partial, nil
	}

	dates := dailyDates(bound, now)
	sums := make([]int64, len(dates))
	byDay := make([][]store.Transaction, len(dates))
	for _, a := range contribute {
		txs, err := s.store.TransactionsSince(ctx, a.ID, bound)
		if err != nil {
			return CurrencyTrend{}, false, err
		}
		for _, t := range txs {
			if i := daysBetween(bound, t.BookingDate); i >= 0 && i < len(byDay) {
				byDay[i] = append(byDay[i], t)
			}
		}
		for i, balance := range walkBack(*a.BalanceMinor, txs, dates) {
			sums[i] += balance
		}
	}

	points := make([]TrendPoint, len(dates))
	for i, d := range dates {
		points[i] = TrendPoint{Date: d, Money: Money{Minor: sums[i], Currency: currency}}
		points[i].Movers, points[i].Smaller = DayMovers(byDay[i])
		for j := range points[i].Movers {
			id := points[i].Movers[j].TransactionID
			points[i].Movers[j].OwnTransfer = pats.IsTransfer(id)
			_, points[i].Movers[j].Unusual = pats.MarkFor(id)
		}
	}
	return CurrencyTrend{Currency: currency, Points: points}, partial, nil
}

// dailyDates is the end of every UTC day from first's day through today, the
// last replaced by now itself.
func dailyDates(first, now time.Time) []time.Time {
	n := daysBetween(first, now) + 1
	dates := make([]time.Time, n)
	for i := range dates {
		dates[i] = utcDay(first).AddDate(0, 0, i+1).Add(-time.Second)
	}
	dates[n-1] = now
	return dates
}

// walkBack reconstructs one account's balance at each of dates from its
// current balance, undoing transactions newer than each date in turn. txs
// must be sorted oldest first and dates oldest first — the same order
// TransactionsSince and dailyDates already return.
func walkBack(current int64, txs []store.Transaction, dates []time.Time) []int64 {
	out := make([]int64, len(dates))
	balance := current
	// i is how many of txs, counted from the start, have not yet been undone.
	// Walking dates newest to oldest, each date only ever undoes the
	// transactions between it and the previous (newer) date.
	i := len(txs)
	for d := len(dates) - 1; d >= 0; d-- {
		for i > 0 && txs[i-1].BookingDate.After(dates[d]) {
			i--
			balance -= txs[i].AmountMinor
		}
		out[d] = balance
	}
	return out
}
