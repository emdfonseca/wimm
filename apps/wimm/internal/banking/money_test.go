package banking_test

import (
	"testing"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

func eur(minor int64) banking.Money { return banking.Money{Minor: minor, Currency: "EUR"} }

// The rule the whole totals requirement rests on: wimm holds no rates, so two
// currencies never add. A plausible wrong total is worse than no total.
func TestMixedCurrenciesNeverAdd(t *testing.T) {
	sum, ok := eur(1000).Add(banking.Money{Minor: 1000, Currency: "GBP"})
	if ok {
		t.Fatalf("added EUR to GBP and got %v", sum)
	}
	if sum != (banking.Money{}) {
		t.Errorf("a refused sum returned %v, want the zero value", sum)
	}
}

func TestAddingWithinOneCurrency(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b banking.Money
		want int64
	}{
		{"two positives", eur(1050), eur(2500), 3550},
		{"an overdraft against a balance", eur(5000), eur(-1200), 3800},
		{"two overdrafts", eur(-100), eur(-250), -350},
		{"a zero balance", eur(4200), eur(0), 4200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.a.Add(tc.b)
			if !ok {
				t.Fatal("refused to add two amounts in one currency")
			}
			if got.Minor != tc.want {
				t.Errorf("Minor = %d, want %d", got.Minor, tc.want)
			}
			if got.Currency != "EUR" {
				t.Errorf("Currency = %q, want EUR", got.Currency)
			}
		})
	}
}

// Zero and negative are shown, never hidden, so the predicates the UI branches
// on have to mean exactly what they say.
func TestZeroAndNegativeAreDistinct(t *testing.T) {
	if !eur(0).IsZero() {
		t.Error("zero is not reported as zero")
	}
	if eur(0).Negative() {
		t.Error("zero is reported as negative")
	}
	if !eur(-1).Negative() {
		t.Error("an overdraft is not reported as negative")
	}
	if eur(-1).IsZero() {
		t.Error("an overdraft is reported as zero")
	}
}

// An amount with no currency must not silently add to a euro amount.
func TestAnAmountWithNoCurrencyDoesNotAddToOneWithA(t *testing.T) {
	if _, ok := eur(100).Add(banking.Money{Minor: 100}); ok {
		t.Error("added a currency-less amount to euros")
	}
}
