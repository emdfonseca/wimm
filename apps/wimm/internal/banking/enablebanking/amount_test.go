package enablebanking

import (
	"strings"
	"testing"
)

func TestParseAmount(t *testing.T) {
	for _, tc := range []struct {
		name     string
		decimal  string
		currency string
		want     int64
	}{
		{"a plain balance", "4200.10", "EUR", 420_010},
		{"no fractional part", "4200", "EUR", 420_000},
		{"a trailing zero", "4200.00", "EUR", 420_000},
		{"one decimal where two are held", "4200.1", "EUR", 420_010},
		{"an overdraft", "-1234.56", "EUR", -123_456},
		{"an explicit plus", "+12.34", "EUR", 1_234},
		{"zero", "0.00", "EUR", 0},
		{"negative zero is just zero", "-0.00", "EUR", 0},
		{"under one unit", "0.07", "EUR", 7},
		{"no leading zero", ".07", "EUR", 7},
		{"a currency with no minor unit", "1234", "JPY", 1_234},
		{"a currency with three", "1.234", "KWD", 1_234},
		{"surrounding space", "  55.50  ", "EUR", 5_550},
		{"a lowercase currency", "10.00", "eur", 1_000},
		{"excess precision that is only zeros", "12.3400", "EUR", 1_234},
		{"a large balance", "92233720368547.75", "EUR", 9_223_372_036_854_775},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAmount(tc.decimal, tc.currency)
			if err != nil {
				t.Fatalf("parseAmount(%q, %q): %v", tc.decimal, tc.currency, err)
			}
			if got.Minor != tc.want {
				t.Errorf("Minor = %d, want %d", got.Minor, tc.want)
			}
			if got.Currency != strings.ToUpper(strings.TrimSpace(tc.currency)) {
				t.Errorf("Currency = %q", got.Currency)
			}
		})
	}
}

// Every one of these would otherwise become a wrong number on a household's
// screen, which is worse than an error.
func TestParseAmountRefuses(t *testing.T) {
	for _, tc := range []struct {
		name              string
		decimal, currency string
		want              string
	}{
		{"no currency", "10.00", "", "no currency"},
		{"a currency with no known minor unit", "10.00", "XYZ", "no known minor unit"},
		{"an empty amount", "", "EUR", "empty amount"},
		{"words", "abc", "EUR", "not a number"},
		{"a thousands separator", "1,234.56", "EUR", "not a number"},
		{"two decimal points", "1.2.3", "EUR", "not a number"},
		{"precision the currency cannot hold", "10.005", "EUR", "more precision"},
		{"decimals on a currency with none", "1234.5", "JPY", "more precision"},
		{"larger than int64", "99999999999999999999.00", "EUR", "does not fit"},
		{"a bare sign", "-", "EUR", "not a number"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAmount(tc.decimal, tc.currency)
			if err == nil {
				t.Fatalf("accepted %q %q as %d", tc.decimal, tc.currency, got.Minor)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not say %q", err, tc.want)
			}
		})
	}
}

// The defect this file exists to prevent: 4200.10 through a float is
// 4200.099999999999, and a household's balance stops matching their bank's.
func TestAmountsThatAFloatWouldGetWrong(t *testing.T) {
	for _, tc := range []struct {
		decimal string
		want    int64
	}{
		{"4200.10", 420_010},
		{"0.29", 29},
		{"1.005", 0}, // refused, not rounded
		{"70.07", 7_007},
		{"8.11", 811},
	} {
		t.Run(tc.decimal, func(t *testing.T) {
			got, err := parseAmount(tc.decimal, "EUR")
			if tc.want == 0 && err == nil {
				t.Fatalf("accepted %q, which cannot be held exactly", tc.decimal)
			}
			if tc.want == 0 {
				return
			}
			if err != nil {
				t.Fatalf("parseAmount: %v", err)
			}
			if got.Minor != tc.want {
				t.Errorf("Minor = %d, want %d", got.Minor, tc.want)
			}
		})
	}
}
