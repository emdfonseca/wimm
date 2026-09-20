package banking_test

import (
	"testing"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

func TestMerchantName(t *testing.T) {
	for _, tc := range []struct {
		name         string
		counterparty string
		remittance   string
		want         string
	}{
		{"toll operator", "", "COMPRA BXV VIA VERDE 512345678901", "Bxv Via Verde"},
		{"web shop with reference and order number", "", "COMPRA WWW.AMAZON NM4HU1VZ4 230002268264350", "Amazon"},
		{"a name with a digit in it", "", "COMPRA NUMBER 1 HAIR 987654321", "Number 1 Hair"},
		{"restaurant", "", "COMPRA REST BOTAFOGO 400123456", "Rest Botafogo"},
		{"two leading words", "", "COMPRAS MULTIBANCO PAYPAL CONVERSE 35314369001", "Paypal Converse"},
		{"country and original amount trailer", "", "COMPRA STEAMGAMES - PAIS: LU VALOR ORIG 19,99 EUR", "Steamgames"},
		{"original value trailer without country", "", "COMPRA IKEA VALOR ORIG 45,00 EUR", "Ikea"},
		{"dotted word and domain suffix", "", "PAG WWW.NETFLIX.COM", "Netflix"},
		{"the bank names the other party", "Ana Silva", "TRF MB WAY 123456789012", "Ana Silva"},
		{"a counterparty is cleaned too", "COMPRA PINGO DOCE 123456789", "ignored", "Pingo Doce"},
		{"mixed case is left alone", "Café Central", "", "Café Central"},
		{"nothing left after cleaning", "", "COMPRA 123456789012", "COMPRA 123456789012"},
		{"whitespace is collapsed", "", "COMPRA   EDP    COMERCIAL", "Edp Comercial"},
		{"no line at all", "", "", "Card payment"},
		{"only whitespace", "  ", "  ", "Card payment"},
		{"same merchant, different references", "", "COMPRA AMAZON ZZ9YY8XX7 111111111111", "Amazon"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := banking.MerchantName(tc.counterparty, tc.remittance); got != tc.want {
				t.Errorf("MerchantName(%q, %q) = %q, want %q", tc.counterparty, tc.remittance, got, tc.want)
			}
		})
	}
}
