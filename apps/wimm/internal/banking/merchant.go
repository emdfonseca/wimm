package banking

import (
	"regexp"
	"strings"
	"unicode"
)

// fallbackName names a transaction the bank said nothing about.
const fallbackName = "Card payment"

// trailerMarkers start a trailer about country or original amount; everything
// from the marker on is cut. Upper case, matched case-insensitively.
var trailerMarkers = []string{" - PAIS:", " VALOR ORIG"}

// transactionWords are the bank's word for the kind of transaction, at the
// start of a line. Portuguese because the banks connected today are: a new
// bank's words are a row here and a case in merchant_test.go.
var transactionWords = map[string]bool{
	"COMPRA": true, "COMPRAS": true, "PAGAMENTO": true, "PAG": true, "PG": true,
	"TRF": true, "TRANSF": true, "TRANSFERENCIA": true, "DD": true,
	"LEVANTAMENTO": true, "MULTIBANCO": true,
}

var domainSuffixes = []string{".COM", ".PT", ".EU", ".NET"}

var (
	longDigits  = regexp.MustCompile(`^[0-9]{9,}$`)
	hasDigit    = regexp.MustCompile(`[0-9]`)
	hasLetter   = regexp.MustCompile(`\p{L}`)
	allNonSpace = regexp.MustCompile(`\S`)
)

// MerchantName is the name wimm shows for a transaction: the counterparty
// where the bank gave one, otherwise the statement line, with the bank's own
// bookkeeping removed. It is never empty, and it is the one place a name is
// derived, so grouping and display cannot disagree.
func MerchantName(counterparty, remittance string) string {
	source := counterparty
	if strings.TrimSpace(source) == "" {
		source = remittance
	}
	if name := cleanMerchant(source); name != "" {
		return name
	}
	if line := strings.TrimSpace(remittance); line != "" {
		return line
	}
	if line := strings.TrimSpace(counterparty); line != "" {
		return line
	}
	return fallbackName
}

func cleanMerchant(line string) string {
	upper := strings.ToUpper(line)
	for _, marker := range trailerMarkers {
		if i := strings.Index(upper, marker); i >= 0 {
			line = line[:i]
			upper = upper[:i]
		}
	}

	tokens := strings.Fields(line)

	for len(tokens) > 0 {
		last := tokens[len(tokens)-1]
		if !longDigits.MatchString(last) && !isReference(last) {
			break
		}
		tokens = tokens[:len(tokens)-1]
	}

	for len(tokens) > 0 && transactionWords[strings.ToUpper(strings.TrimRight(tokens[0], ".-"))] {
		tokens = tokens[1:]
	}

	if len(tokens) > 0 && strings.HasPrefix(strings.ToUpper(tokens[0]), "WWW.") {
		tokens[0] = tokens[0][len("WWW."):]
	}
	if len(tokens) > 0 {
		last := tokens[len(tokens)-1]
		for _, suffix := range domainSuffixes {
			if strings.HasSuffix(strings.ToUpper(last), suffix) {
				tokens[len(tokens)-1] = last[:len(last)-len(suffix)]
				break
			}
		}
	}

	name := strings.Join(tokens, " ")
	if !allNonSpace.MatchString(name) {
		return ""
	}
	if name == strings.ToUpper(name) {
		name = titleCase(name)
	}
	return name
}

// isReference is a token of 8 or more characters mixing letters and digits,
// the shape of an order or terminal reference.
func isReference(token string) bool {
	return len([]rune(token)) >= 8 && hasDigit.MatchString(token) && hasLetter.MatchString(token)
}

func titleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}
