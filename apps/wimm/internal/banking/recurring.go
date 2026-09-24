package banking

import (
	"slices"
	"strings"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// The recurrence rule (ADR 0025). Every threshold is named here so tuning is
// one edit and one table case.
const (
	recurLookbackMonths       = 25
	recurMinOccurrences       = 3
	recurMinOccurrencesYearly = 2
	// recurAmountTolerance: a run's amounts, sorted ascending, are each no more
	// than this share above the run's smallest.
	recurAmountTolerance = 0.10
)

// Cadence is how often a recurring payment repeats.
type Cadence int

const (
	CadenceUnspecified Cadence = iota
	CadenceWeekly
	CadenceMonthly
	CadenceYearly
)

type cadenceRule struct {
	cadence        Cadence
	minGap, maxGap int
	slack          int
	minOccurrences int
	next           func(time.Time) time.Time
}

var cadenceRules = []cadenceRule{
	{CadenceWeekly, 6, 8, 2, recurMinOccurrences, func(t time.Time) time.Time { return t.AddDate(0, 0, 7) }},
	{CadenceMonthly, 27, 34, 5, recurMinOccurrences, nextMonth},
	{CadenceYearly, 351, 379, 14, recurMinOccurrencesYearly, nextYear},
}

// RecurringPayment is a run of booked money-out to one merchant at about one
// amount on a steady interval.
type RecurringPayment struct {
	Name      string
	Amount    int64 // signed minor units of the newest payment
	Currency  string
	Cadence   Cadence
	Likely    bool // a yearly run of exactly two
	Late      bool // expected has passed, within the cadence's slack
	Expected  time.Time
	AccountID string
}

// RecurringPayments finds the recurring payments in rows, soonest expected
// first. Only booked money out inside the lookback counts; money in never does.
func RecurringPayments(rows []store.Transaction, today time.Time) []RecurringPayment {
	today = utcDay(today)
	from := today.AddDate(0, -recurLookbackMonths, 0)

	type key struct{ name, currency string }
	groups := map[key][]store.Transaction{}
	for _, r := range rows {
		if r.Status != store.StatusBooked || r.AmountMinor >= 0 || r.BookingDate.Before(from) {
			continue
		}
		k := key{strings.ToLower(MerchantName(r.CounterpartyName, r.Remittance)), r.Currency}
		groups[k] = append(groups[k], r)
	}

	var out []RecurringPayment
	for _, group := range groups {
		for _, run := range amountRuns(group) {
			if p, ok := recurringRun(run, today); ok {
				out = append(out, p)
			}
		}
	}
	slices.SortFunc(out, func(a, b RecurringPayment) int {
		if c := a.Expected.Compare(b.Expected); c != 0 {
			return c
		}
		if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
			return c
		}
		return int(b.Amount - a.Amount)
	})
	return out
}

// amountRuns clusters money-out rows so each cluster's amounts sit within the
// tolerance of its smallest.
func amountRuns(rows []store.Transaction) [][]store.Transaction {
	sorted := slices.Clone(rows)
	slices.SortFunc(sorted, func(a, b store.Transaction) int {
		return int(b.AmountMinor - a.AmountMinor) // -100 before -500: smallest magnitude first
	})
	var runs [][]store.Transaction
	var smallest int64
	for _, r := range sorted {
		mag := -r.AmountMinor
		if len(runs) > 0 && float64(mag) <= float64(smallest)*(1+recurAmountTolerance) {
			runs[len(runs)-1] = append(runs[len(runs)-1], r)
			continue
		}
		smallest = mag
		runs = append(runs, []store.Transaction{r})
	}
	return runs
}

func recurringRun(run []store.Transaction, today time.Time) (RecurringPayment, bool) {
	if len(run) < 2 {
		return RecurringPayment{}, false
	}
	slices.SortFunc(run, func(a, b store.Transaction) int {
		if c := b.BookingDate.Compare(a.BookingDate); c != 0 {
			return c
		}
		return strings.Compare(b.ID, a.ID)
	})
	gap := func(i int) int { return int(utcDay(run[i].BookingDate).Sub(utcDay(run[i+1].BookingDate)).Hours() / 24) }

	rule, ok := ruleForGap(gap(0))
	if !ok {
		return RecurringPayment{}, false
	}
	n := 2
	for n < len(run) && gap(n-1) >= rule.minGap && gap(n-1) <= rule.maxGap {
		n++
	}
	if n < rule.minOccurrences {
		return RecurringPayment{}, false
	}

	newest := run[0]
	expected := rule.next(utcDay(newest.BookingDate))
	if today.After(expected.AddDate(0, 0, rule.slack)) {
		return RecurringPayment{}, false
	}
	return RecurringPayment{
		Name:      MerchantName(newest.CounterpartyName, newest.Remittance),
		Amount:    newest.AmountMinor,
		Currency:  newest.Currency,
		Cadence:   rule.cadence,
		Likely:    rule.cadence == CadenceYearly && n == recurMinOccurrencesYearly,
		Late:      today.After(expected),
		Expected:  expected,
		AccountID: newest.AccountID,
	}, true
}

func ruleForGap(days int) (cadenceRule, bool) {
	for _, r := range cadenceRules {
		if days >= r.minGap && days <= r.maxGap {
			return r, true
		}
	}
	return cadenceRule{}, false
}

// nextMonth is the same day next month, capped at that month's length.
func nextMonth(t time.Time) time.Time {
	y, m, d := t.Date()
	first := time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return time.Date(first.Year(), first.Month(), min(d, last), 0, 0, 0, 0, time.UTC)
}

func nextYear(t time.Time) time.Time {
	y, m, d := t.Date()
	first := time.Date(y+1, m, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return time.Date(y+1, m, min(d, last), 0, 0, 0, 0, time.UTC)
}
