package banking

import (
	"slices"
	"strings"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// The explains-the-day rule (ADR 0025).
const (
	dayCoverShare = 0.80
	dayMaxRows    = 5
)

// DayMover is one row that moved a day.
type DayMover struct {
	TransactionID string
	Name          string
	Amount        int64 // signed minor units
	Unusual       bool  // set by the caller, which knows the marks
	// OwnTransfer is set by the caller too. A mover that is half of a transfer
	// between the member's own accounts is never also unusual, and the row
	// stays on the chart whatever the scope: the day's money still moved, so
	// the popover has to say why (ADR 0026).
	OwnTransfer bool
}

// DayMovers takes a day's booked rows in either direction, largest first, until
// they reach dayCoverShare of the day's gross movement, never more than
// dayMaxRows. smaller counts the rows left out.
func DayMovers(rows []store.Transaction) (movers []DayMover, smaller int) {
	sorted := slices.Clone(rows)
	slices.SortFunc(sorted, func(a, b store.Transaction) int {
		if c := abs64(b.AmountMinor) - abs64(a.AmountMinor); c != 0 {
			if c > 0 {
				return 1
			}
			return -1
		}
		return strings.Compare(a.ID, b.ID)
	})

	var gross int64
	for _, r := range sorted {
		gross += abs64(r.AmountMinor)
	}
	var covered int64
	for _, r := range sorted {
		if len(movers) == dayMaxRows || float64(covered) >= dayCoverShare*float64(gross) {
			break
		}
		covered += abs64(r.AmountMinor)
		movers = append(movers, DayMover{
			TransactionID: r.ID,
			Name:          MerchantName(r.CounterpartyName, r.Remittance),
			Amount:        r.AmountMinor,
		})
	}
	return movers, len(sorted) - len(movers)
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
