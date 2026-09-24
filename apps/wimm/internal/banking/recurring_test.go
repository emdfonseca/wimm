package banking_test

import (
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

func row(name string, date time.Time, minor int64) store.Transaction {
	return store.Transaction{
		ID:               name + date.Format("20060102"),
		AccountID:        "acc-1",
		Status:           store.StatusBooked,
		AmountMinor:      minor,
		Currency:         "EUR",
		BookingDate:      date,
		CounterpartyName: name,
	}
}

func TestRecurringPayments(t *testing.T) {
	today := on(2026, time.September, 20)

	type want struct {
		name     string
		amount   int64
		cadence  banking.Cadence
		likely   bool
		late     bool
		expected time.Time
	}

	for _, tc := range []struct {
		name string
		rows []store.Transaction
		want []want
	}{
		{
			name: "a subscription paid every month",
			rows: []store.Transaction{
				row("Netflix", on(2026, time.June, 14), -1299),
				row("Netflix", on(2026, time.July, 14), -1299),
				row("Netflix", on(2026, time.August, 14), -1299),
				row("Netflix", on(2026, time.September, 14), -1299),
			},
			want: []want{{"Netflix", -1299, banking.CadenceMonthly, false, false, on(2026, time.October, 14)}},
		},
		{
			name: "the price went up a little",
			rows: []store.Transaction{
				row("Netflix", on(2026, time.July, 14), -1299),
				row("Netflix", on(2026, time.August, 14), -1299),
				row("Netflix", on(2026, time.September, 14), -1399),
			},
			want: []want{{"Netflix", -1399, banking.CadenceMonthly, false, false, on(2026, time.October, 14)}},
		},
		{
			name: "the same supermarket most weeks for different amounts",
			rows: []store.Transaction{
				row("Pingo Doce", on(2026, time.September, 18), -3000),
				row("Pingo Doce", on(2026, time.September, 11), -14000),
				row("Pingo Doce", on(2026, time.September, 4), -6100),
				row("Pingo Doce", on(2026, time.August, 28), -9800),
				row("Pingo Doce", on(2026, time.August, 21), -4200),
			},
		},
		{
			name: "only twice so far",
			rows: []store.Transaction{
				row("Gym", on(2026, time.August, 5), -2500),
				row("Gym", on(2026, time.September, 5), -2500),
			},
		},
		{
			name: "weekly, three times",
			rows: []store.Transaction{
				row("Limpeza Casa", on(2026, time.September, 3), -4500),
				row("Limpeza Casa", on(2026, time.September, 10), -4500),
				row("Limpeza Casa", on(2026, time.September, 17), -4500),
			},
			want: []want{{"Limpeza Casa", -4500, banking.CadenceWeekly, false, false, on(2026, time.September, 24)}},
		},
		{
			name: "an insurance premium paid twice is likely yearly",
			rows: []store.Transaction{
				row("Fidelidade", on(2025, time.March, 3), -38600),
				row("Fidelidade", on(2026, time.March, 3), -38600),
			},
			want: []want{{"Fidelidade", -38600, banking.CadenceYearly, true, false, on(2027, time.March, 3)}},
		},
		{
			name: "the third year confirms it",
			rows: []store.Transaction{
				row("Fidelidade", on(2024, time.September, 25), -38600),
				row("Fidelidade", on(2025, time.September, 18), -38600),
				row("Fidelidade", on(2026, time.September, 14), -38600),
			},
			want: []want{{"Fidelidade", -38600, banking.CadenceYearly, false, false, on(2027, time.September, 14)}},
		},
		{
			name: "the day moves around a weekend",
			rows: []store.Transaction{
				row("Rent", on(2026, time.July, 1), -82000),
				row("Rent", on(2026, time.August, 3), -82000),
				row("Rent", on(2026, time.September, 2), -82000),
			},
			want: []want{{"Rent", -82000, banking.CadenceMonthly, false, false, on(2026, time.October, 2)}},
		},
		{
			name: "two subscriptions with one merchant",
			rows: []store.Transaction{
				row("Apple", on(2026, time.July, 9), -499),
				row("Apple", on(2026, time.July, 9), -1799),
				row("Apple", on(2026, time.August, 9), -499),
				row("Apple", on(2026, time.August, 9), -1799),
				row("Apple", on(2026, time.September, 9), -499),
				row("Apple", on(2026, time.September, 9), -1799),
			},
			want: []want{
				{"Apple", -499, banking.CadenceMonthly, false, false, on(2026, time.October, 9)},
				{"Apple", -1799, banking.CadenceMonthly, false, false, on(2026, time.October, 9)},
			},
		},
		{
			name: "a subscription that was cancelled",
			rows: []store.Transaction{
				row("Old Gym", on(2026, time.May, 20), -2500),
				row("Old Gym", on(2026, time.June, 20), -2500),
				row("Old Gym", on(2026, time.July, 20), -2500),
			},
		},
		{
			name: "a few days late is still recurring",
			rows: []store.Transaction{
				row("Spotify", on(2026, time.June, 18), -999),
				row("Spotify", on(2026, time.July, 18), -999),
				row("Spotify", on(2026, time.August, 18), -999),
			},
			want: []want{{"Spotify", -999, banking.CadenceMonthly, false, true, on(2026, time.September, 18)}},
		},
		{
			name: "a salary is not recurring",
			rows: []store.Transaction{
				row("Employer Lda", on(2026, time.July, 25), 245000),
				row("Employer Lda", on(2026, time.August, 25), 245000),
				row("Employer Lda", on(2026, time.September, 25), 245000),
			},
		},
		{
			name: "a pending third payment does not count",
			rows: []store.Transaction{
				row("Netflix", on(2026, time.July, 14), -1299),
				row("Netflix", on(2026, time.August, 14), -1299),
				func() store.Transaction {
					r := row("Netflix", on(2026, time.September, 14), -1299)
					r.Status = store.StatusPending
					return r
				}(),
			},
		},
		{
			name: "the month end is capped at the month's length",
			rows: []store.Transaction{
				row("Water", on(2026, time.June, 30), -3000),
				row("Water", on(2026, time.July, 31), -3000),
				row("Water", on(2026, time.August, 31), -3000),
			},
			want: []want{{"Water", -3000, banking.CadenceMonthly, false, false, on(2026, time.September, 30)}},
		},
		{
			name: "soonest expected first",
			rows: []store.Transaction{
				row("Later", on(2026, time.July, 18), -1000),
				row("Later", on(2026, time.August, 18), -1000),
				row("Later", on(2026, time.September, 18), -1000),
				row("Sooner", on(2026, time.July, 14), -2000),
				row("Sooner", on(2026, time.August, 14), -2000),
				row("Sooner", on(2026, time.September, 14), -2000),
			},
			want: []want{
				{"Sooner", -2000, banking.CadenceMonthly, false, false, on(2026, time.October, 14)},
				{"Later", -1000, banking.CadenceMonthly, false, false, on(2026, time.October, 18)},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := banking.RecurringPayments(tc.rows, today)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d recurring payments %+v, want %d", len(got), got, len(tc.want))
			}
			for i, w := range tc.want {
				g := got[i]
				if g.Name != w.name || g.Amount != w.amount || g.Cadence != w.cadence ||
					g.Likely != w.likely || g.Late != w.late || !g.Expected.Equal(w.expected) {
					t.Errorf("[%d] got %+v, want %+v", i, g, w)
				}
			}
		})
	}
}
