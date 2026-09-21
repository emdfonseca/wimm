package banking_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

func TestDayMovers(t *testing.T) {
	d := on(2026, time.August, 3)
	mk := func(minors ...int64) []store.Transaction {
		out := make([]store.Transaction, len(minors))
		for i, m := range minors {
			out[i] = row(fmt.Sprintf("Shop %d", i), d, m)
			out[i].ID = fmt.Sprintf("t%d", i)
		}
		return out
	}

	for _, tc := range []struct {
		name        string
		rows        []store.Transaction
		wantAmounts []int64
		wantSmaller int
	}{
		{"an empty day", nil, nil, 0},
		{"one row covers the day", mk(-64000, -9210, -300, -200), []int64{-64000}, 3},
		{"two rows are needed", mk(-500, -5000, -4000, -500), []int64{-5000, -4000}, 2},
		{"twelve similar rows give five and seven", mk(-100, -100, -100, -100, -100, -100, -100, -100, -100, -100, -100, -100),
			[]int64{-100, -100, -100, -100, -100}, 7},
		{"money in and out, each by size", mk(5000, -4000, 500), []int64{5000, -4000}, 1},
		{"a single row is all of the day", mk(-1234), []int64{-1234}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			movers, smaller := banking.DayMovers(tc.rows)
			if smaller != tc.wantSmaller || len(movers) != len(tc.wantAmounts) {
				t.Fatalf("got %d movers, %d smaller; want %d, %d", len(movers), smaller, len(tc.wantAmounts), tc.wantSmaller)
			}
			for i, m := range movers {
				if m.Amount != tc.wantAmounts[i] {
					t.Errorf("mover %d amount = %d, want %d", i, m.Amount, tc.wantAmounts[i])
				}
			}
		})
	}
}
