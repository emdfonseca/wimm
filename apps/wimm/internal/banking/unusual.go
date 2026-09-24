package banking

import (
	"math"
	"slices"
	"strings"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// The unusual-payment rule (ADR 0025).
const (
	unusualScore             = 3.5
	unusualMADScale          = 0.6745
	unusualMinPartyPayments  = 5
	unusualGeneralDays       = 90
	unusualFlatRatio         = 3.0
	unusualFirstPaymentShare = 0.10
	unusualFloorShare        = 0.01

	historyMonths = 13
	minFullMonths = 3
	typicalMonths = 6
)

// MonthlyMedians are the median monthly money out and money in over the full
// months held, as positive minor units. Zero means no full month is held, and
// then nothing in that direction is unusual.
type MonthlyMedians struct {
	Out, In int64
}

// UnusualMark says why a payment is unusual and what it was measured against.
type UnusualMark struct {
	TransactionID string
	In            bool // money coming in
	// FirstPayment: the merchant was never paid before, and Typical is unset.
	FirstPayment bool
	// Typical is the baseline's median as a positive amount in minor units.
	Typical int64
}

// UnusualPayments marks the booked payments in rows that are far above their
// baseline, keyed by transaction ID. rows are one currency and carry the whole
// lookback; only those booked on or after judgedFrom are judged, the older
// ones serve as baseline. Each direction is judged against its own.
func UnusualPayments(rows []store.Transaction, medians MonthlyMedians, judgedFrom time.Time) map[string]UnusualMark {
	type party struct {
		name string
		in   bool
	}
	var booked []store.Transaction
	byParty := map[party][]int{}
	byDirection := map[bool][]int{}
	for _, r := range rows {
		if r.Status != store.StatusBooked || r.AmountMinor == 0 {
			continue
		}
		booked = append(booked, r)
		i := len(booked) - 1
		p := party{strings.ToLower(MerchantName(r.CounterpartyName, r.Remittance)), r.AmountMinor > 0}
		byParty[p] = append(byParty[p], i)
		byDirection[p.in] = append(byDirection[p.in], i)
	}

	out := map[string]UnusualMark{}
	for i, r := range booked {
		if r.BookingDate.Before(judgedFrom) {
			continue
		}
		in := r.AmountMinor > 0
		monthly := medians.Out
		if in {
			monthly = medians.In
		}
		mag := math.Abs(float64(r.AmountMinor))
		if monthly <= 0 || mag < unusualFloorShare*float64(monthly) {
			continue
		}

		p := party{strings.ToLower(MerchantName(r.CounterpartyName, r.Remittance)), in}
		mark := UnusualMark{TransactionID: r.ID, In: in}

		if !in && isFirstPayment(booked, byParty[p], i) && mag > unusualFirstPaymentShare*float64(monthly) {
			mark.FirstPayment = true
			out[r.ID] = mark
			continue
		}

		var baseline []float64
		others := without(byParty[p], i)
		if len(others) >= unusualMinPartyPayments {
			baseline = magnitudes(booked, others)
		} else {
			since := r.BookingDate.AddDate(0, 0, -unusualGeneralDays)
			for _, j := range without(byDirection[in], i) {
				d := booked[j].BookingDate
				if d.After(since) && !d.After(r.BookingDate) {
					baseline = append(baseline, math.Abs(float64(booked[j].AmountMinor)))
				}
			}
			if len(baseline) < unusualMinPartyPayments {
				continue
			}
		}

		typical, unusual := judge(mag, baseline)
		if !unusual {
			continue
		}
		mark.Typical = int64(math.Round(typical))
		out[r.ID] = mark
	}
	return out
}

func isFirstPayment(booked []store.Transaction, party []int, self int) bool {
	for _, j := range party {
		if j != self && booked[j].BookingDate.Before(booked[self].BookingDate) {
			return false
		}
	}
	return true
}

func without(idx []int, self int) []int {
	out := make([]int, 0, len(idx))
	for _, j := range idx {
		if j != self {
			out = append(out, j)
		}
	}
	return out
}

func magnitudes(booked []store.Transaction, idx []int) []float64 {
	out := make([]float64, len(idx))
	for k, j := range idx {
		out[k] = math.Abs(float64(booked[j].AmountMinor))
	}
	return out
}

// judge applies the modified z-score to log amounts, and the flat-baseline
// ratio where every baseline amount is alike. typical is exp(median of logs).
func judge(mag float64, baseline []float64) (typical float64, unusual bool) {
	logs := make([]float64, len(baseline))
	for i, b := range baseline {
		logs[i] = math.Log(b)
	}
	med := median(logs)
	typical = math.Exp(med)

	devs := make([]float64, len(logs))
	for i, l := range logs {
		devs[i] = math.Abs(l - med)
	}
	mad := median(devs)
	if mad == 0 {
		return typical, mag >= unusualFlatRatio*median(baseline)
	}
	return typical, unusualMADScale*(math.Log(mag)-med)/mad > unusualScore
}

func median(xs []float64) float64 {
	s := slices.Clone(xs)
	slices.Sort(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}
