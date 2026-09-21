package banking_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// signed is a magnitude with the direction its party implies: employers pay in.
func signed(name string, minor int64) int64 {
	if strings.Contains(name, "Employer") {
		return minor
	}
	return -minor
}

// history is n booked rows to one party, one every step days, newest ending on
// end, each with the next of amounts (cycling). IDs are unique per party.
func history(name string, end time.Time, step, n int, amounts ...int64) []store.Transaction {
	out := make([]store.Transaction, 0, n)
	for i := range n {
		date := end.AddDate(0, 0, -step*(n-1-i))
		r := row(name, date, signed(name, amounts[i%len(amounts)]))
		r.ID = fmt.Sprintf("%s-%d", name, i)
		out = append(out, r)
	}
	return out
}

func target(name string, date time.Time, minor int64) store.Transaction {
	r := row(name, date, signed(name, minor))
	r.ID = "target"
	return r
}

func TestUnusualPayments(t *testing.T) {
	medians := banking.MonthlyMedians{Out: 210000, In: 245000}
	sep := on(2026, time.September, 10)
	from := on(2025, time.September, 1)

	type want struct {
		unusual  bool
		in       bool
		first    bool
		typical  int64
		typicalW int64 // tolerance in minor units
	}

	for _, tc := range []struct {
		name    string
		rows    []store.Transaction
		medians banking.MonthlyMedians
		want    want
	}{
		{
			name: "far above what that merchant usually costs",
			rows: append(history("Galp", on(2026, time.August, 30), 14, 12, 5800, 6000, 6200, 5900, 6100),
				target("Galp", sep, 64000)),
			medians: medians,
			want:    want{unusual: true, typical: 6000, typicalW: 200},
		},
		{
			name: "a first payment that is a large part of a month",
			rows: append(history("Pingo Doce", sep, 3, 40, 2500, 4000, 6100, 1800),
				target("Auto Reparadora", sep, 165000)),
			medians: medians,
			want:    want{unusual: true, first: true},
		},
		{
			name: "a first payment of an ordinary size",
			rows: append(history("Pingo Doce", sep, 3, 40, 2500, 4000, 6100, 1800),
				target("Restaurante", sep, 4800)),
			medians: medians,
			want:    want{unusual: false},
		},
		{
			name: "a merchant paid a few times, judged against payments in general",
			rows: append(append(history("Pingo Doce", sep, 3, 25, 2500, 4000, 6100, 1800, 3300, 1200),
				history("Ferragens", on(2026, time.August, 20), 10, 3, 3000)...),
				target("Ferragens", sep, 90000)),
			medians: medians,
			want:    want{unusual: true, typical: 3000, typicalW: 1500},
		},
		{
			name: "large, and what that merchant always costs",
			rows: append(history("Landlord", on(2026, time.August, 3), 30, 5, 82000),
				target("Landlord", on(2026, time.September, 2), 82000)),
			medians: medians,
			want:    want{unusual: false},
		},
		{
			name: "far above usual, and still small",
			rows: append(history("Cafe", sep, 1, 30, 120),
				target("Cafe", sep, 900)),
			medians: medians,
			want:    want{unusual: false},
		},
		{
			name: "a fixed price that tripled",
			rows: append(history("Streaming", on(2026, time.August, 14), 30, 6, 1299),
				target("Streaming", sep, 3897)),
			medians: medians,
			want:    want{unusual: true, typical: 1299, typicalW: 1},
		},
		{
			name: "a quiet month",
			rows: append(history("Galp", on(2026, time.August, 30), 14, 12, 5800, 6000, 6200, 5900, 6100),
				target("Galp", sep, 6300)),
			medians: medians,
			want:    want{unusual: false},
		},
		{
			name: "too little history: no full month held",
			rows: append(history("Galp", on(2026, time.August, 30), 14, 12, 5800, 6000, 6200, 5900, 6100),
				target("Galp", sep, 64000)),
			medians: banking.MonthlyMedians{},
			want:    want{unusual: false},
		},
		{
			name: "a bonus",
			rows: append(history("Employer Lda", on(2026, time.August, 25), 30, 12, 245000),
				target("Employer Lda", sep, 980400)),
			medians: medians,
			want:    want{unusual: true, in: true, typical: 245000, typicalW: 1},
		},
		{
			name: "the salary itself is not judged against money going out",
			rows: append(history("Pingo Doce", sep, 1, 60, 2500, 4000, 6100, 1800),
				target("New Employer", sep, 245000)),
			medians: medians,
			want:    want{unusual: false},
		},
		{
			name: "a baseline under five rows is not judged",
			rows: append(append(history("Ferragens", on(2026, time.August, 20), 10, 2, 3000),
				history("Padaria", on(2026, time.August, 25), 10, 2, 400)...),
				target("Ferragens", sep, 9000)),
			medians: medians,
			want:    want{unusual: false},
		},
		{
			name: "the party baseline applies at five payments",
			rows: append(append(history("Ginasio", on(2026, time.August, 30), 20, 5, 3000, 3100, 2950, 3050, 3000),
				history("Loja", sep, 3, 30, 800, 5000, 12000, 1500, 25000)...),
				target("Ginasio", sep, 9000)),
			medians: medians,
			want:    want{unusual: true, typical: 3000, typicalW: 100},
		},
		{
			name: "four payments fall back to the general baseline",
			rows: append(append(history("Ginasio", on(2026, time.August, 30), 20, 4, 3000, 3100, 2950, 3050),
				history("Loja", sep, 3, 30, 800, 5000, 12000, 1500, 25000)...),
				target("Ginasio", sep, 9000)),
			medians: medians,
			want:    want{unusual: false},
		},
		{
			name: "general baseline stops at 90 days",
			rows: append(history("Old", on(2026, time.May, 1), 1, 30, 2000),
				target("Ferragens", sep, 9000)),
			medians: medians,
			want:    want{unusual: false},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := banking.UnusualPayments(tc.rows, tc.medians, from)
			mark, ok := got["target"]
			if ok != tc.want.unusual {
				t.Fatalf("unusual = %v, want %v (%+v)", ok, tc.want.unusual, mark)
			}
			if !ok {
				return
			}
			if mark.In != tc.want.in || mark.FirstPayment != tc.want.first {
				t.Errorf("got %+v, want in=%v first=%v", mark, tc.want.in, tc.want.first)
			}
			if tc.want.first {
				if mark.Typical != 0 {
					t.Errorf("a first payment carries no typical, got %d", mark.Typical)
				}
				return
			}
			if d := mark.Typical - tc.want.typical; d < -tc.want.typicalW || d > tc.want.typicalW {
				t.Errorf("typical = %d, want %d ±%d", mark.Typical, tc.want.typical, tc.want.typicalW)
			}
		})
	}
}

func TestUnusualPaymentsDoNotMoveTheirOwnBaseline(t *testing.T) {
	rows := history("Galp", on(2026, time.August, 30), 14, 5, 6000)
	a := row("Galp", on(2026, time.September, 2), -60000)
	a.ID = "a"
	b := row("Galp", on(2026, time.September, 4), -62000)
	b.ID = "b"
	rows = append(rows, a, b)

	got := banking.UnusualPayments(rows, banking.MonthlyMedians{Out: 210000, In: 245000}, on(2025, time.September, 1))
	for _, id := range []string{"a", "b"} {
		m, ok := got[id]
		if !ok {
			t.Fatalf("%s should be unusual", id)
		}
		if m.Typical != 6000 {
			t.Errorf("%s typical = %d, want 6000: a large payment moved the baseline", id, m.Typical)
		}
	}
}

func TestUnusualPaymentsBeforeTheMonthsShownAreOnlyBaseline(t *testing.T) {
	old := target("Galp", on(2025, time.June, 1), 64000)
	rows := append(history("Galp", on(2025, time.May, 30), 14, 12, 6000), old)
	got := banking.UnusualPayments(rows, banking.MonthlyMedians{Out: 210000, In: 245000}, on(2025, time.September, 1))
	if len(got) != 0 {
		t.Fatalf("rows before the months shown must not be judged, got %+v", got)
	}
}
