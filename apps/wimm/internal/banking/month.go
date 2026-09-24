package banking

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// topCount is how long each of the two top-spending lists is.
const topCount = 5

// MonthSummary is what Overview's month section needs: per currency, the month
// so far against the same days of last month, and where the most money went.
type MonthSummary struct {
	Months []CurrencyMonth
	// Labels names the account each largest payment left.
	Labels map[string]store.AccountLabel
}

// CurrencyMonth is one currency's month so far. In and Out are both positive;
// Net is In minus Out.
type CurrencyMonth struct {
	Currency string
	In, Out  Money
	Net      Money
	// Prior* are the same days of last month, nil unless every contributing
	// ledger reaches back to the first of last month. A part-month set against
	// a month wimm only partly holds would always look like an improvement.
	PriorIn, PriorOut, PriorNet *Money
	// CountedFrom is set when a contributing ledger begins after the 1st, so
	// the figures are not passed off as the whole month so far. With several
	// ledgers it is the latest beginning: the date from which every one is held.
	CountedFrom *time.Time
	MonthStart  time.Time

	TopMerchants    []MerchantTotal
	LargestPayments []store.Transaction
}

// MerchantTotal is money out to one merchant this month.
type MerchantTotal struct {
	Name     string
	Total    Money
	Payments int
}

// MonthSummary computes the month so far for the accounts a member owns, from
// stored rows alone: it asks no bank.
//
// Windows come from database time. The prior window is the same day numbers of
// last month, capped at that month's length, so the 31st is set against the
// whole of a 30-day month.
func (s *Service) MonthSummary(ctx context.Context, memberID string, asked Scope) (MonthSummary, error) {
	scope, err := s.scopedAccounts(ctx, memberID, asked)
	if err != nil {
		return MonthSummary{}, err
	}
	now, err := s.store.Now(ctx)
	if err != nil {
		return MonthSummary{}, err
	}
	today := utcDay(now.T)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	windowEnd := today.AddDate(0, 0, 1)

	priorStart := monthStart.AddDate(0, -1, 0)
	priorDays := min(today.Day(), daysIn(priorStart))
	priorEnd := priorStart.AddDate(0, 0, priorDays)

	// One read of every owned account, widened by the transfer window at each
	// end so a pair straddling the edge of a window is still found. The sums
	// are then taken in Go, because SQL cannot leave out a row whose partner it
	// has not seen (ADR 0026). It stays a read of about two months, which is
	// the property that kept this a separate RPC.
	rows, err := s.store.OwnedBooked(ctx, memberID, scope.OwnedIDs(),
		priorStart.AddDate(0, 0, -ownTransferWindowDays),
		windowEnd.AddDate(0, 0, ownTransferWindowDays))
	if err != nil {
		return MonthSummary{}, err
	}
	counted, _ := countedRows(rows, OwnTransfers(rows, scope.Owned), scope.AccountIDs)

	current := windowSums(counted, monthStart, windowEnd)
	if len(current) == 0 {
		return MonthSummary{}, nil
	}
	prior := windowSums(counted, priorStart, priorEnd)

	// Outgoing is the same rows filtered, so a transfer the scope left out is
	// in neither top merchants nor largest payments.
	var outgoing []store.Transaction
	for _, t := range counted {
		if t.AmountMinor < 0 && !t.BookingDate.Before(monthStart) && t.BookingDate.Before(windowEnd) {
			outgoing = append(outgoing, t)
		}
	}

	accounts, err := s.store.OwnedAccountsForTrend(ctx, memberID, scope.AccountIDs)
	if err != nil {
		return MonthSummary{}, err
	}
	labels, err := s.store.OwnedAccountLabels(ctx, memberID)
	if err != nil {
		return MonthSummary{}, err
	}

	priorByCurrency := map[string]store.WindowSum{}
	for _, w := range prior {
		priorByCurrency[w.Currency] = w
	}

	var months []CurrencyMonth
	for _, w := range current {
		if w.Rows == 0 {
			continue
		}
		c := CurrencyMonth{
			Currency:   w.Currency,
			In:         Money{Minor: w.InMinor, Currency: w.Currency},
			Out:        Money{Minor: w.OutMinor, Currency: w.Currency},
			Net:        Money{Minor: w.InMinor - w.OutMinor, Currency: w.Currency},
			MonthStart: monthStart,
		}

		var ledgers []time.Time
		for _, a := range accounts {
			if a.Currency == w.Currency && a.Scope.ReadsTransactions() && a.Oldest != nil {
				ledgers = append(ledgers, utcDay(*a.Oldest))
			}
		}
		reachesPrior, latest := len(ledgers) > 0, time.Time{}
		for _, begins := range ledgers {
			if begins.After(priorStart) {
				reachesPrior = false
			}
			if begins.After(latest) {
				latest = begins
			}
		}
		if reachesPrior {
			p := priorByCurrency[w.Currency]
			in := Money{Minor: p.InMinor, Currency: w.Currency}
			out := Money{Minor: p.OutMinor, Currency: w.Currency}
			net := Money{Minor: p.InMinor - p.OutMinor, Currency: w.Currency}
			c.PriorIn, c.PriorOut, c.PriorNet = &in, &out, &net
		}
		if latest.After(monthStart) {
			c.CountedFrom = &latest
		}

		var payments []store.Transaction
		for _, t := range outgoing {
			if t.Currency == w.Currency {
				payments = append(payments, t)
			}
		}
		c.TopMerchants = topMerchants(payments, w.Currency)
		c.LargestPayments = largestPayments(payments)
		months = append(months, c)
	}

	slices.SortStableFunc(months, func(a, b CurrencyMonth) int {
		if a.Out.Minor != b.Out.Minor {
			if a.Out.Minor > b.Out.Minor {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Currency, b.Currency)
	})
	return MonthSummary{Months: months, Labels: labels}, nil
}

func daysIn(monthStart time.Time) int {
	return int(monthStart.AddDate(0, 1, 0).Sub(monthStart).Hours() / 24)
}

// windowSums sums counted rows over [from, to) per currency, the currency with
// the most money out first. Rows is the count of counted rows, so a window
// holding nothing but a transfer has no summary: nothing was spent or received.
func windowSums(counted []store.Transaction, from, to time.Time) []store.WindowSum {
	byCurrency := map[string]*store.WindowSum{}
	var order []string
	for _, t := range counted {
		if t.BookingDate.Before(from) || !t.BookingDate.Before(to) {
			continue
		}
		w, ok := byCurrency[t.Currency]
		if !ok {
			w = &store.WindowSum{Currency: t.Currency}
			byCurrency[t.Currency] = w
			order = append(order, t.Currency)
		}
		if t.AmountMinor < 0 {
			w.OutMinor += -t.AmountMinor
		} else {
			w.InMinor += t.AmountMinor
		}
		w.Rows++
	}

	out := make([]store.WindowSum, 0, len(order))
	for _, c := range order {
		out = append(out, *byCurrency[c])
	}
	slices.SortStableFunc(out, func(a, b store.WindowSum) int {
		if a.OutMinor != b.OutMinor {
			if a.OutMinor > b.OutMinor {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Currency, b.Currency)
	})
	return out
}

// topMerchants groups outgoing payments on the lower-cased merchant name and
// ranks by money. Display keeps the first-seen casing; payments arrive newest
// first, so that is the most recent.
func topMerchants(payments []store.Transaction, currency string) []MerchantTotal {
	byName := map[string]*MerchantTotal{}
	var order []string
	for _, t := range payments {
		name := MerchantName(t.CounterpartyName, t.Remittance)
		key := strings.ToLower(name)
		m, ok := byName[key]
		if !ok {
			m = &MerchantTotal{Name: name, Total: Money{Currency: currency}}
			byName[key] = m
			order = append(order, key)
		}
		m.Total.Minor -= t.AmountMinor
		m.Payments++
	}

	out := make([]MerchantTotal, 0, len(order))
	for _, key := range order {
		out = append(out, *byName[key])
	}
	slices.SortStableFunc(out, func(a, b MerchantTotal) int {
		if a.Total.Minor != b.Total.Minor {
			if a.Total.Minor > b.Total.Minor {
				return -1
			}
			return 1
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return out[:min(len(out), topCount)]
}

func largestPayments(payments []store.Transaction) []store.Transaction {
	out := slices.Clone(payments)
	slices.SortStableFunc(out, func(a, b store.Transaction) int {
		switch {
		case a.AmountMinor < b.AmountMinor:
			return -1
		case a.AmountMinor > b.AmountMinor:
			return 1
		}
		return 0
	})
	return out[:min(len(out), topCount)]
}
