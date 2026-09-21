package banking

import (
	"slices"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// LeftOut is how many transfers between the member's own accounts a scope left
// out of its figures, and what they came to. A pair is counted in the month its
// money left, so a set of months' counts adds up to the number of pairs.
type LeftOut struct {
	Pairs int
	// Total is positive minor units, one currency, being the amount that
	// moved once per pair rather than twice.
	Total int64
	// OutRowIDs are the out rows of the pairs left out, so a caller filing a
	// pair under a month can ask which month its money left.
	OutRowIDs []string
}

// countedRows splits a scope's rows into the ones its figures count and the
// transfers it leaves out. A pair is left out only when both of its rows are on
// accounts the scope counts; one half outside means the pair crosses the scope
// and is counted as what it is there, money that left those accounts or money
// that arrived in them (ADR 0026).
//
// rows are the member's rows for every account they own, which is what
// OwnTransfers was given, and pairs is its result. Only the rows on scope are
// returned, in the order they came in.
func countedRows(rows []store.Transaction, pairs map[string]string, scope []string) ([]store.Transaction, LeftOut) {
	inScope := make(map[string]bool, len(scope))
	for _, id := range scope {
		inScope[id] = true
	}
	account := make(map[string]string, len(rows))
	for _, r := range rows {
		account[r.ID] = r.AccountID
	}

	// A pair is left out when the scope counts both of its accounts. Asking
	// the question per pair rather than per row keeps the count and the rows
	// one decision, so neither can drift from the other.
	leftOutRow := map[string]bool{}
	var out LeftOut
	for id, partner := range pairs {
		if !inScope[account[id]] || !inScope[account[partner]] {
			continue
		}
		leftOutRow[id] = true
	}

	var counted []store.Transaction
	for _, r := range rows {
		switch {
		case !inScope[r.AccountID]:
		case leftOutRow[r.ID]:
			// Count the pair once, under the row whose money left.
			if r.AmountMinor < 0 {
				out.Pairs++
				out.Total += -r.AmountMinor
				out.OutRowIDs = append(out.OutRowIDs, r.ID)
			}
		default:
			counted = append(counted, r)
		}
	}
	slices.Sort(out.OutRowIDs)
	return counted, out
}
