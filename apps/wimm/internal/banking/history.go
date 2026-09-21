package banking

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// riserCount is the most merchants listed as having taken more than usual.
const riserCount = 5

// History is what Overview's Month by month and Recurring payments sections
// need, per currency, within one scope.
type History struct {
	Scope     Scope
	Available []Scope
	// Counted and NotCounted are set only when Household counts fewer accounts
	// than the household money figure does.
	Counted, NotCounted []string
	Currencies          []CurrencyHistory
	// Labels names the account each payment left.
	Labels map[string]store.AccountLabel
}

// CurrencyHistory is one currency's months, newest first.
type CurrencyHistory struct {
	Currency   string
	Months     []HistoryMonth
	FullMonths int
	// Typical* and Average* are nil under minFullMonths full months. The Usual
	// pair is the same with every unusual payment set aside.
	TypicalNet, AverageNet           *Money
	TypicalNetUsual, AverageNetUsual *Money
	UnusualCount                     int
	Recurring                        []RecurringPayment
	// LateLedgers are accounts missing from at least one full month shown: their
	// ledger begins after the earliest one, and after the oldest full month
	// began.
	LateLedgers []LateLedger
}

// LateLedger is a scoped account whose ledger begins part way through the
// months shown, so the months before From do not include it.
type LateLedger struct {
	AccountName string
	From        time.Time
}

// ledger is one scoped account's reach: where its oldest booked row is.
type ledger struct {
	accountID string
	begins    time.Time
}

// HistoryMonth is one calendar month. In and Out are positive; Net is In minus
// Out.
type HistoryMonth struct {
	Start        time.Time
	In, Out, Net Money
	// NetUsual is set only when the month held an unusual payment.
	NetUsual *Money
	Full     bool
	SoFar    bool
	// HeldFrom is set when a contributing ledger begins inside the month.
	HeldFrom *time.Time
	// Risers is empty for a month that is not full, and wherever no usual can be
	// stated.
	Risers  []MerchantRise
	Unusual []UnusualPayment
	// TransfersLeftOut is how many transfers between the member's own accounts
	// this month's figures left out, and TransfersTotal what they came to as a
	// positive amount. A pair is counted in the month its money left, so the
	// months' counts add up to the number of pairs (ADR 0026).
	TransfersLeftOut int
	TransfersTotal   Money
}

// MerchantRise is a merchant that took more in a month than in its own middle
// month. A zero Usual means it is not usually paid at all.
type MerchantRise struct {
	Name     string
	Total    Money
	Usual    Money
	Payments int
}

// UnusualPayment is a booked payment the one rule marks, with what it was
// measured against.
type UnusualPayment struct {
	Transaction store.Transaction
	// Typical is nil for a first payment.
	Typical      *Money
	FirstPayment bool
}

// patternsUnderAll judges the payments Transactions is about to show against
// every account the member owns, whatever the Overview scope: a row's mark must
// not depend on a control on another screen. Rows older than the months shown
// are baseline only, so a page wholly older than them is not judged at all.
//
// Under All every pair is inside the scope by construction, so no labelled row
// is ever also marked unusual here.
func (s *Service) patternsUnderAll(ctx context.Context, memberID string, oldest, newest time.Time) (Patterns, error) {
	now, err := s.store.Now(ctx)
	if err != nil {
		return Patterns{}, err
	}
	today := utcDay(now.T)
	current := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstShown := current.AddDate(0, -(historyMonths - 1), 0)
	lookbackFrom := current.AddDate(0, -(recurLookbackMonths - 1), 0)
	to := today.AddDate(0, 0, 1)

	scope, err := s.scopedAccounts(ctx, memberID, ScopeAll)
	if err != nil {
		return Patterns{}, err
	}

	// One read covering both the lookback the marks are judged over and the
	// page itself, which may reach further back than the lookback does. A row
	// whose partner is on screen beside it must be labelled, and the label
	// does not depend on a month being shown.
	from, until := lookbackFrom, to
	if reach := oldest.AddDate(0, 0, -ownTransferWindowDays); reach.Before(from) {
		from = reach
	}
	if reach := newest.AddDate(0, 0, ownTransferWindowDays); reach.After(until) {
		until = reach
	}
	rows, err := s.store.OwnedBooked(ctx, memberID, scope.OwnedIDs(), from, until)
	if err != nil {
		return Patterns{}, err
	}
	out := Patterns{Transfers: OwnTransfers(rows, scope.Owned)}

	if newest.Before(firstShown) {
		// Nothing on this page is judged: it is wholly older than the months
		// shown, so those rows are baseline and nothing else.
		return out, nil
	}
	accounts, err := s.store.OwnedAccountsForTrend(ctx, memberID, scope.AccountIDs)
	if err != nil {
		return Patterns{}, err
	}
	// Under All every pair is inside the scope, so the counted rows are the
	// lookback's rows less every pair.
	var lookback []store.Transaction
	for _, r := range rows {
		if !r.BookingDate.Before(lookbackFrom) && r.BookingDate.Before(to) {
			lookback = append(lookback, r)
		}
	}
	counted, _ := countedRows(lookback, out.Transfers, scope.AccountIDs)
	out.Unusual = marksAcross(accounts, counted, current, firstShown)
	return out, nil
}

// MonthHistory computes the months, the recurring payments and the unusual
// payments of the accounts a member owns within a scope, from stored rows
// alone: it asks no bank. Every window comes from database time.
func (s *Service) MonthHistory(ctx context.Context, memberID string, asked Scope) (History, error) {
	scope, err := s.scopedAccounts(ctx, memberID, asked)
	if err != nil {
		return History{}, err
	}
	out := History{
		Scope: scope.Answered, Available: scope.Available,
		Counted: scope.Counted, NotCounted: scope.NotCounted,
	}
	if len(scope.AccountIDs) == 0 {
		return out, nil
	}

	now, err := s.store.Now(ctx)
	if err != nil {
		return History{}, err
	}
	today := utcDay(now.T)
	current := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstShown := current.AddDate(0, -(historyMonths - 1), 0)
	lookbackFrom := current.AddDate(0, -(recurLookbackMonths - 1), 0)
	to := today.AddDate(0, 0, 1)

	accounts, err := s.store.OwnedAccountsForTrend(ctx, memberID, scope.AccountIDs)
	if err != nil {
		return History{}, err
	}
	// Read every owned account rather than the scoped ones, because a pair is
	// a fact about the member's rows and one half may sit outside the scope.
	// The months are then summed in Go from the counted rows, so there is one
	// source for a figure rather than a SQL sum that cannot see a partner
	// (ADR 0026).
	rows, err := s.store.OwnedBooked(ctx, memberID, scope.OwnedIDs(), lookbackFrom, to)
	if err != nil {
		return History{}, err
	}
	out.Labels, err = s.store.OwnedAccountLabels(ctx, memberID)
	if err != nil {
		return History{}, err
	}
	counted, leftOut := countedRows(rows, OwnTransfers(rows, scope.Owned), scope.AccountIDs)

	rowsBy := map[string][]store.Transaction{}
	for _, r := range counted {
		rowsBy[r.Currency] = append(rowsBy[r.Currency], r)
	}
	leftOutBy := leftOutByMonth(rows, leftOut)
	ledgersBy := ledgersOf(accounts)

	for currency, ledgers := range ledgersBy {
		c := buildCurrencyHistory(currency, ledgers, monthSums(rowsBy[currency], firstShown, to),
			rowsBy[currency], leftOutBy[currency], today, current, firstShown, out.Labels)
		if len(c.Months) == 0 && len(c.Recurring) == 0 {
			continue
		}
		out.Currencies = append(out.Currencies, c)
	}
	slices.SortStableFunc(out.Currencies, func(a, b CurrencyHistory) int {
		if x, y := totalOut(a), totalOut(b); x != y {
			if x > y {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Currency, b.Currency)
	})
	return out, nil
}

func totalOut(c CurrencyHistory) int64 {
	var n int64
	for _, m := range c.Months {
		n += m.Out.Minor
	}
	return n
}

// ledgersOf groups the scoped accounts that hold a readable ledger by currency.
func ledgersOf(accounts []store.TrendAccount) map[string][]ledger {
	out := map[string][]ledger{}
	for _, a := range accounts {
		if a.Scope.ReadsTransactions() && a.Oldest != nil {
			out[a.Currency] = append(out[a.Currency], ledger{accountID: a.ID, begins: utcDay(*a.Oldest)})
		}
	}
	return out
}

func beginsOf(ledgers []ledger) []time.Time {
	out := make([]time.Time, len(ledgers))
	for i, l := range ledgers {
		out[i] = l.begins
	}
	return out
}

// monthSums sums counted rows into months over [from, to). It replaces a SQL
// sum, which cannot leave out a row whose partner it has not seen.
func monthSums(counted []store.Transaction, from, to time.Time) map[time.Time]store.MonthSum {
	out := map[time.Time]store.MonthSum{}
	for _, r := range counted {
		if r.BookingDate.Before(from) || !r.BookingDate.Before(to) {
			continue
		}
		start := time.Date(r.BookingDate.Year(), r.BookingDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		m := out[start]
		m.Currency, m.Month = r.Currency, start
		if r.AmountMinor < 0 {
			m.OutMinor += -r.AmountMinor
		} else {
			m.InMinor += r.AmountMinor
		}
		out[start] = m
	}
	return out
}

// monthLeftOut is how many transfers one month's figures left out, and what
// they came to.
type monthLeftOut struct {
	Pairs int
	Total int64
}

// leftOutByMonth files each left-out pair under the month its money left, so
// the months' counts add up to the number of pairs even where a pair straddles
// two of them.
func leftOutByMonth(rows []store.Transaction, leftOut LeftOut) map[string]map[time.Time]monthLeftOut {
	byID := make(map[string]store.Transaction, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}
	out := map[string]map[time.Time]monthLeftOut{}
	for _, id := range leftOut.OutRowIDs {
		r, ok := byID[id]
		if !ok {
			continue
		}
		start := time.Date(r.BookingDate.Year(), r.BookingDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		if out[r.Currency] == nil {
			out[r.Currency] = map[time.Time]monthLeftOut{}
		}
		m := out[r.Currency][start]
		m.Pairs++
		m.Total += -r.AmountMinor
		out[r.Currency][start] = m
	}
	return out
}

func buildCurrencyHistory(
	currency string, ledgers []ledger, sums map[time.Time]store.MonthSum,
	rows []store.Transaction, leftOut map[time.Time]monthLeftOut,
	today, current, firstShown time.Time,
	labels map[string]store.AccountLabel,
) CurrencyHistory {
	money := func(minor int64) Money { return Money{Minor: minor, Currency: currency} }

	months := shownMonths(currency, beginsOf(ledgers), sums, current, firstShown)
	for i := range months {
		if l, ok := leftOut[months[i].Start]; ok {
			months[i].TransfersLeftOut = l.Pairs
			months[i].TransfersTotal = money(l.Total)
		}
	}
	c := CurrencyHistory{Currency: currency, Months: months}
	var oldestFull time.Time
	for _, m := range months {
		if m.Full {
			c.FullMonths++
			oldestFull = m.Start // newest first, so the last one seen is the oldest
		}
	}
	if c.FullMonths > 0 {
		earliest := slices.MinFunc(beginsOf(ledgers), time.Time.Compare)
		for _, l := range ledgers {
			if l.begins.After(earliest) && l.begins.After(oldestFull) {
				c.LateLedgers = append(c.LateLedgers, LateLedger{AccountName: labels[l.accountID].Name, From: l.begins})
			}
		}
		slices.SortFunc(c.LateLedgers, func(a, b LateLedger) int {
			if d := a.From.Compare(b.From); d != 0 {
				return d
			}
			return strings.Compare(a.AccountName, b.AccountName)
		})
	}
	marks := marksFor(months, rows, firstShown)

	usualBy := merchantMedians(rows, months)
	byMonth := map[time.Time]int{}
	for i, m := range months {
		byMonth[m.Start] = i
	}
	totalsBy := map[time.Time]map[string]*MerchantRise{}
	for _, r := range rows { // newest first, which is the order a month lists them in
		start := time.Date(r.BookingDate.Year(), r.BookingDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		i, shown := byMonth[start]
		if !shown {
			continue
		}
		if mark, ok := marks[r.ID]; ok {
			p := UnusualPayment{Transaction: r, FirstPayment: mark.FirstPayment}
			if !mark.FirstPayment {
				typical := money(mark.Typical)
				p.Typical = &typical
			}
			months[i].Unusual = append(months[i].Unusual, p)
			c.UnusualCount++
		}
		if r.AmountMinor < 0 && months[i].Full {
			name := MerchantName(r.CounterpartyName, r.Remittance)
			key := strings.ToLower(name)
			if totalsBy[start] == nil {
				totalsBy[start] = map[string]*MerchantRise{}
			}
			t, ok := totalsBy[start][key]
			if !ok {
				t = &MerchantRise{Name: name, Total: money(0)}
				totalsBy[start][key] = t
			}
			t.Total.Minor -= r.AmountMinor
			t.Payments++
		}
	}

	for i := range months {
		m := &months[i]
		if len(m.Unusual) > 0 {
			var setAside int64
			for _, u := range m.Unusual {
				setAside += u.Transaction.AmountMinor
			}
			net := money(m.Net.Minor - setAside)
			m.NetUsual = &net
		}
		if !m.Full || c.FullMonths < minFullMonths {
			continue
		}
		for key, t := range totalsBy[m.Start] {
			usual := usualBy[key]
			if t.Total.Minor > usual {
				t.Usual = money(usual)
				m.Risers = append(m.Risers, *t)
			}
		}
		slices.SortFunc(m.Risers, func(a, b MerchantRise) int {
			if x, y := a.Total.Minor-a.Usual.Minor, b.Total.Minor-b.Usual.Minor; x != y {
				if x > y {
					return -1
				}
				return 1
			}
			return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		})
		m.Risers = m.Risers[:min(len(m.Risers), riserCount)]
	}

	if c.FullMonths >= minFullMonths {
		var nets, netsUsual []int64
		for _, m := range months {
			if !m.Full {
				continue
			}
			nets = append(nets, m.Net.Minor)
			if m.NetUsual != nil {
				netsUsual = append(netsUsual, m.NetUsual.Minor)
			} else {
				netsUsual = append(netsUsual, m.Net.Minor)
			}
		}
		typical, average := money(medianMinor(nets)), money(meanMinor(nets))
		typicalUsual, averageUsual := money(medianMinor(netsUsual)), money(meanMinor(netsUsual))
		c.TypicalNet, c.AverageNet = &typical, &average
		c.TypicalNetUsual, c.AverageNetUsual = &typicalUsual, &averageUsual
	}
	c.Months = months

	var outgoing []store.Transaction
	for _, r := range rows {
		if r.AmountMinor < 0 {
			outgoing = append(outgoing, r)
		}
	}
	c.Recurring = RecurringPayments(outgoing, today)
	return c
}

// shownMonths lists the months of one currency that are shown, newest first,
// with what each is: full, so far, or held from a date.
func shownMonths(
	currency string, ledgers []time.Time, sums map[time.Time]store.MonthSum, current, firstShown time.Time,
) []HistoryMonth {
	earliest := slices.MinFunc(ledgers, time.Time.Compare)
	money := func(minor int64) Money { return Money{Minor: minor, Currency: currency} }

	var months []HistoryMonth
	for start := current; !start.Before(firstShown); start = start.AddDate(0, -1, 0) {
		next := start.AddDate(0, 1, 0)
		if next.AddDate(0, 0, -1).Before(earliest) {
			break
		}
		sum := sums[start]
		m := HistoryMonth{
			Start: start,
			In:    money(sum.InMinor), Out: money(sum.OutMinor), Net: money(sum.InMinor - sum.OutMinor),
			SoFar: start.Equal(current),
			Full:  start.Before(current) && !start.Before(earliest),
		}
		if earliest.After(start) && earliest.Before(next) {
			held := earliest
			m.HeldFrom = &held
		}
		months = append(months, m)
	}
	return months
}

// marksFor is the unusual payments among rows, judged against the medians of
// the full months among months. With no full month held nothing is unusual.
func marksFor(months []HistoryMonth, rows []store.Transaction, firstShown time.Time) map[string]UnusualMark {
	var fullIn, fullOut []int64
	for _, m := range months {
		if m.Full {
			fullIn = append(fullIn, m.In.Minor)
			fullOut = append(fullOut, m.Out.Minor)
		}
	}
	var medians MonthlyMedians
	if len(fullOut) > 0 {
		medians = MonthlyMedians{Out: medianMinor(fullOut), In: medianMinor(fullIn)}
	}
	return UnusualPayments(rows, medians, firstShown)
}

// Patterns is what one read of a member's rows says about them: which are
// unusual, and which are halves of a transfer between their own accounts. They
// come back together because they are derived from one row set and a row that
// is a transfer is never also unusual.
//
// Ask through IsTransfer and MarkFor rather than reading the maps: the
// precedence between the two marks is the rule, and it lives here so no caller
// can apply it differently (ADR 0026).
type Patterns struct {
	Unusual map[string]UnusualMark
	// Transfers holds both directions of every pair, keyed by transaction id.
	// A row is labelled whenever it is in here, whatever the scope.
	Transfers map[string]string
}

// IsTransfer reports a row being one half of a movement between the member's
// own accounts. The label takes the status slot from unusual (ADR 0026).
func (p Patterns) IsTransfer(id string) bool { _, ok := p.Transfers[id]; return ok }

// MarkFor is the unusual-payment mark on a row, and never one on a row the
// transfer rule already claimed.
func (p Patterns) MarkFor(id string) (UnusualMark, bool) {
	if p.IsTransfer(id) {
		return UnusualMark{}, false
	}
	m, ok := p.Unusual[id]
	return m, ok
}

// patternsFor judges every payment of the scoped accounts, all currencies at
// once. It reads the same rows MonthHistory does, so a mark on the chart is the
// mark in the month, and it pairs across every owned account so a crossing pair
// is still labelled.
func (s *Service) patternsFor(
	ctx context.Context, memberID string, scope scoped, accounts []store.TrendAccount, now time.Time,
) (Patterns, error) {
	today := utcDay(now)
	current := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstShown := current.AddDate(0, -(historyMonths - 1), 0)
	to := today.AddDate(0, 0, 1)

	rows, err := s.store.OwnedBooked(ctx, memberID, scope.OwnedIDs(),
		current.AddDate(0, -(recurLookbackMonths-1), 0), to)
	if err != nil {
		return Patterns{}, err
	}
	pairs := OwnTransfers(rows, scope.Owned)
	counted, _ := countedRows(rows, pairs, scope.AccountIDs)
	return Patterns{Unusual: marksAcross(accounts, counted, current, firstShown), Transfers: pairs}, nil
}

// marksAcross judges the counted rows per currency. The months it measures
// against are summed from those same rows, so a transfer the scope left out
// moves neither a month's figure nor the baseline a payment is judged against.
func marksAcross(
	accounts []store.TrendAccount, counted []store.Transaction, current, firstShown time.Time,
) map[string]UnusualMark {
	rowsBy := map[string][]store.Transaction{}
	for _, r := range counted {
		rowsBy[r.Currency] = append(rowsBy[r.Currency], r)
	}
	to := current.AddDate(0, 1, 0)
	out := map[string]UnusualMark{}
	for currency, ledgers := range ledgersOf(accounts) {
		sums := monthSums(rowsBy[currency], firstShown, to)
		months := shownMonths(currency, beginsOf(ledgers), sums, current, firstShown)
		for id, mark := range marksFor(months, rowsBy[currency], firstShown) {
			out[id] = mark
		}
	}
	return out
}

// merchantMedians is each merchant's median monthly total over the full
// months, a month with no payment to it counting as nothing.
func merchantMedians(rows []store.Transaction, months []HistoryMonth) map[string]int64 {
	full := map[time.Time]bool{}
	for _, m := range months {
		if m.Full {
			full[m.Start] = true
		}
	}
	perMonth := map[string]map[time.Time]int64{}
	for _, r := range rows {
		start := time.Date(r.BookingDate.Year(), r.BookingDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		if r.AmountMinor >= 0 || !full[start] {
			continue
		}
		key := strings.ToLower(MerchantName(r.CounterpartyName, r.Remittance))
		if perMonth[key] == nil {
			perMonth[key] = map[time.Time]int64{}
		}
		perMonth[key][start] -= r.AmountMinor
	}
	out := make(map[string]int64, len(perMonth))
	for key, byMonth := range perMonth {
		totals := make([]int64, 0, len(full))
		for start := range full {
			totals = append(totals, byMonth[start])
		}
		out[key] = medianMinor(totals)
	}
	return out
}

// medianMinor is the middle value of xs, or the mean of the middle two,
// rounded half away from zero.
func medianMinor(xs []int64) int64 {
	if len(xs) == 0 {
		return 0
	}
	s := slices.Clone(xs)
	slices.Sort(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return divRound(s[n/2-1]+s[n/2], 2)
}

func meanMinor(xs []int64) int64 {
	var sum int64
	for _, x := range xs {
		sum += x
	}
	return divRound(sum, int64(len(xs)))
}

// divRound is a/b rounded half away from zero, for positive b.
func divRound(a, b int64) int64 {
	if a >= 0 {
		return (a + b/2) / b
	}
	return -((-a + b/2) / b)
}
