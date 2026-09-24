package enablebanking

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

// Gateways send amounts as decimal strings. wimm holds int64 minor units and a
// currency, so the conversion happens here, once, on the way in — never a
// float, at any point, because a float is how "4200.10" becomes 4200.099999
// and a household's balance stops matching their bank's.
//
// The exponent comes from the currency. Most have two minor digits; these are
// the exceptions that exist in practice. An unknown currency is not guessed at
// two — it is refused, because silently treating a JPY amount as having two
// decimals is off by a factor of a hundred.
var minorDigits = map[string]int{
	"EUR": 2, "GBP": 2, "USD": 2, "CHF": 2, "SEK": 2, "NOK": 2, "DKK": 2,
	"PLN": 2, "CZK": 2, "RON": 2, "BGN": 2, "HUF": 2, "CAD": 2, "AUD": 2,

	// No minor unit at all.
	"JPY": 0, "ISK": 0, "CLP": 0, "KRW": 0,
	// Three.
	"BHD": 3, "JOD": 3, "KWD": 3, "OMR": 3, "TND": 3,
}

// parseAmount converts a gateway's decimal string and ISO 4217 code into
// wimm's Money.
func parseAmount(decimal, currency string) (banking.Money, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return banking.Money{}, errors.New("an amount with no currency")
	}
	exponent, known := minorDigits[currency]
	if !known {
		return banking.Money{}, fmt.Errorf("currency %q has no known minor unit", currency)
	}

	text := strings.TrimSpace(decimal)
	if text == "" {
		return banking.Money{}, errors.New("an empty amount")
	}

	negative := false
	switch text[0] {
	case '-':
		negative, text = true, text[1:]
	case '+':
		text = text[1:]
	}

	// A sign on its own is not zero. Stripping it above can empty the string,
	// and an empty string must not reach the "" -> "0" default below.
	if text == "" {
		return banking.Money{}, fmt.Errorf("%q is not a number", decimal)
	}

	whole, fraction, hasFraction := strings.Cut(text, ".")
	if whole == "" {
		whole = "0"
	}

	// Both halves must be digits before anything is measured against them.
	// Cut splits at the first point, so "1.2.3" would otherwise reach the
	// precision check with a fraction of "2.3" and be refused for the wrong
	// reason — a misleading error about rounding, on input that is not a
	// number at all.
	if !isDigits(whole) || (hasFraction && !isDigits(fraction)) {
		return banking.Money{}, fmt.Errorf("%q is not a number", decimal)
	}

	// More decimals than the currency has is a real loss of value, not a
	// rounding opportunity: refuse rather than silently truncate.
	if len(fraction) > exponent {
		if strings.Trim(fraction[exponent:], "0") != "" {
			return banking.Money{}, fmt.Errorf(
				"%q has more precision than %s holds, and rounding it would change the amount", decimal, currency)
		}
		fraction = fraction[:exponent]
	}
	fraction += strings.Repeat("0", exponent-len(fraction))

	// Concatenating rather than multiplying keeps this exact for any magnitude
	// int64 can hold.
	minor, err := strconv.ParseInt(whole+fraction, 10, 64)
	if err != nil {
		return banking.Money{}, fmt.Errorf("%q does not fit in a 64-bit amount", decimal)
	}
	if negative {
		minor = -minor
	}
	return banking.Money{Minor: minor, Currency: currency}, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
