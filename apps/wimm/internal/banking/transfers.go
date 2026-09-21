package banking

import (
	"cmp"
	"slices"
	"strings"
	"unicode"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// The rule for a transfer between a member's own accounts (ADR 0026).
const (
	// ownTransferWindowDays is how far apart two booking dates may be, in
	// either order. Three because the household's own pairs end at two and a
	// weekend adds one; four and five find nothing more and double the room
	// for a coincidence.
	ownTransferWindowDays = 3
	// ownTransferMinSuffix is the shortest number suffix that can be a token.
	ownTransferMinSuffix = 4
	// ownTransferMinNameRunes is the shortest account name that can be one.
	ownTransferMinNameRunes = 4
)

// OwnedAccount is what the rule knows about an account the member owns. It is
// only ever an owned account: one the member merely holds a grant on is never
// either half of a transfer (ADR 0021).
type OwnedAccount struct {
	ID            string
	Name          string
	HouseholdName string
	NumberSuffix  string
	HolderName    string
}

// OwnTransfers pairs the booked rows that are two halves of one movement
// between accounts the member owns, and returns both directions of every pair
// keyed by transaction ID. rows are every booked row of every account the
// member owns, in every currency; whether a pair is then left out of a figure
// is a question about the scope in force, decided by the caller.
//
// It is pure and runs once per request over rows already in hand, so tuning
// the rule never needs a backfill (ADR 0025's reason, kept).
func OwnTransfers(rows []store.Transaction, accounts []OwnedAccount) map[string]string {
	owned := make(map[string]bool, len(accounts))
	for _, a := range accounts {
		owned[a.ID] = true
	}
	tokens := discriminatingTokens(accounts)

	// A pair is two equal opposite amounts of one currency, so the search is
	// per bucket and never across the whole ledger.
	type bucket struct {
		currency string
		minor    int64
	}
	type candidateSet struct{ outs, ins []int }
	buckets := map[bucket]*candidateSet{}
	for i, r := range rows {
		if r.Status != store.StatusBooked || r.AmountMinor == 0 || !owned[r.AccountID] {
			continue
		}
		key := bucket{r.Currency, max(r.AmountMinor, -r.AmountMinor)}
		set := buckets[key]
		if set == nil {
			set = &candidateSet{}
			buckets[key] = set
		}
		if r.AmountMinor < 0 {
			set.outs = append(set.outs, i)
		} else {
			set.ins = append(set.ins, i)
		}
	}

	paired := map[string]string{}
	for _, set := range buckets {
		pairBucket(rows, set.outs, set.ins, tokens, paired)
	}
	return paired
}

// candidate is one row's view of a possible partner: how it ranks, and which
// row it is. Smaller is better.
type candidate struct {
	gap        int
	noEvidence int // 0 when a row's text names the other account
	row        int
}

// pairBucket pairs rows of one currency and one absolute amount. Two rows pair
// when each is the other's single best-ranked unpaired candidate. A row whose
// best rank is shared by two candidates is left for a later round, where it
// may pair once one of them has gone elsewhere; what is left over is unpaired.
//
// Pairing greedily in date order was rejected: with two outs and two ins of one
// amount it takes the first of each even when the second is the same-day one,
// so the answer would depend on the order rows arrived in.
func pairBucket(rows []store.Transaction, outs, ins []int, tokens map[string][]string, paired map[string]string) {
	outs, ins = slices.Clone(outs), slices.Clone(ins)
	byDate := func(a, b int) int { return rows[a].BookingDate.Compare(rows[b].BookingDate) }
	slices.SortFunc(outs, byDate)
	slices.SortFunc(ins, byDate)

	for {
		best := map[int]candidate{}
		tied := map[int]bool{}
		// Both sides are in date order, so the ins that could partner an out
		// are one window-wide run. Walking its edges forward makes the scan
		// linear in the rows within a window rather than in the bucket, which
		// is what keeps a 25-month ledger off a quadratic path.
		lo, hi := 0, 0
		for _, o := range outs {
			from := rows[o].BookingDate.AddDate(0, 0, -ownTransferWindowDays)
			to := rows[o].BookingDate.AddDate(0, 0, ownTransferWindowDays)
			for lo < len(ins) && rows[ins[lo]].BookingDate.Before(from) {
				lo++
			}
			for hi < len(ins) && !rows[ins[hi]].BookingDate.After(to) {
				hi++
			}
			for _, i := range ins[lo:hi] {
				if rows[o].AccountID == rows[i].AccountID {
					continue
				}
				// Either order: an arrival is sometimes booked before the
				// departure it came from.
				gap := max(daysBetween(rows[o].BookingDate, rows[i].BookingDate),
					daysBetween(rows[i].BookingDate, rows[o].BookingDate))
				evidence := namesAccount(rows[o], tokens[rows[i].AccountID]) ||
					namesAccount(rows[i], tokens[rows[o].AccountID])
				noEvidence := 1
				if evidence {
					noEvidence = 0
				}
				consider(best, tied, o, candidate{gap, noEvidence, i})
				consider(best, tied, i, candidate{gap, noEvidence, o})
			}
		}

		done := map[int]bool{}
		for _, o := range outs {
			c, ok := best[o]
			if !ok || tied[o] || tied[c.row] {
				continue
			}
			if back, ok := best[c.row]; ok && back.row == o {
				paired[rows[o].ID] = rows[c.row].ID
				paired[rows[c.row].ID] = rows[o].ID
				done[o], done[c.row] = true, true
			}
		}
		if len(done) == 0 {
			return
		}
		outs = slices.DeleteFunc(outs, func(o int) bool { return done[o] })
		ins = slices.DeleteFunc(ins, func(i int) bool { return done[i] })
	}
}

// consider keeps row's best candidate, and records that its best rank is
// shared. A row wimm cannot tell apart is worse paired than left alone,
// because pairing it removes a real payment from the figures.
func consider(best map[int]candidate, tied map[int]bool, row int, c candidate) {
	current, ok := best[row]
	if !ok {
		best[row] = c
		return
	}
	switch cmp.Or(cmp.Compare(c.gap, current.gap), cmp.Compare(c.noEvidence, current.noEvidence)) {
	case -1:
		best[row] = c
		tied[row] = false
	case 0:
		tied[row] = true
	}
}

// discriminatingTokens is, per owned account, the text that names that account
// and no other of the member's. A value a second owned account carries says
// "mine" and not "which", so it cannot tell two of them apart: holder_name is
// the same on every account the member holds, and an account name or a number
// suffix repeated across two accounts is no better (ADR 0026).
func discriminatingTokens(accounts []OwnedAccount) map[string][]string {
	raw := make(map[string][]string, len(accounts))
	seen := map[string]int{}
	for _, a := range accounts {
		var tokens []string
		// A number suffix must be digits: a length minimum was standing in for
		// that, and admits a value like ".com".
		if s := normaliseText(a.NumberSuffix); len([]rune(s)) >= ownTransferMinSuffix && isDigits(s) {
			tokens = append(tokens, s)
		}
		for _, name := range []string{a.Name, a.HouseholdName} {
			if s := normaliseText(name); len([]rune(s)) >= ownTransferMinNameRunes {
				tokens = append(tokens, s)
			}
		}
		// HolderName is never a token, by the rule above.
		tokens = slices.Compact(slices.Sorted(slices.Values(tokens)))
		raw[a.ID] = tokens
		for _, t := range tokens {
			seen[t]++
		}
	}

	out := make(map[string][]string, len(accounts))
	for id, tokens := range raw {
		kept := slices.DeleteFunc(tokens, func(t string) bool { return seen[t] > 1 })
		if len(kept) > 0 {
			out[id] = kept
		}
	}
	return out
}

// namesAccount reports a row's text carrying one of the other account's
// tokens. Evidence only breaks a tie between candidates that are equally close
// in date; it is never required, because a bank writes a name as often as a
// number and a pair with neither is still a pair.
func namesAccount(row store.Transaction, tokens []string) bool {
	if len(tokens) == 0 {
		return false
	}
	text := normaliseText(row.CounterpartyName + " " + row.Remittance)
	for _, t := range tokens {
		if strings.Contains(text, t) {
			return true
		}
	}
	return false
}

func normaliseText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func isDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}
