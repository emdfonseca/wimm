package banking

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

func day(d int) time.Time {
	return time.Date(2026, 6, d, 0, 0, 0, 0, time.UTC)
}

// tx builds a booked row. id names it in the expectations.
func tx(id, account string, minor int64, currency string, d int, text string) store.Transaction {
	return store.Transaction{
		ID:               id,
		AccountID:        account,
		Status:           store.StatusBooked,
		AmountMinor:      minor,
		Currency:         currency,
		BookingDate:      day(d),
		CounterpartyName: text,
	}
}

// pairsOf renders the result as sorted "a+b" strings so a table case states
// what was paired without depending on map order or on which half is which.
func pairsOf(t *testing.T, got map[string]string) []string {
	t.Helper()
	var out []string
	for a, b := range got {
		if got[b] != a {
			t.Errorf("pairing is not mutual: %s -> %s -> %s", a, b, got[b])
		}
		if a < b {
			out = append(out, a+"+"+b)
		}
	}
	slices.Sort(out)
	return out
}

// current, savings and third are one member's accounts. joint is a second
// account of theirs whose number suffix is too short to be evidence.
var testAccounts = []OwnedAccount{
	{ID: "current", Name: "Current Account", HouseholdName: "Current", NumberSuffix: "1111", HolderName: "Ana Reis"},
	{ID: "savings", Name: "Savings Pot", HouseholdName: "Savings", NumberSuffix: "2222", HolderName: "Ana Reis"},
	{ID: "third", Name: "Third Account", HouseholdName: "Third", NumberSuffix: "3333", HolderName: "Ana Reis"},
	{ID: "shortsuffix", Name: "Wallet", HouseholdName: "Wallet", NumberSuffix: "44", HolderName: "Ana Reis"},
}

func TestOwnTransfers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		rows     []store.Transaction
		accounts []OwnedAccount
		want     []string
	}{
		{
			name: "moving money to savings, the next day",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "EUR", 4, "transfer"),
			},
			want: []string{"in+out"},
		},
		{
			name: "both sides on the same day",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "EUR", 3, "transfer"),
			},
			want: []string{"in+out"},
		},
		{
			name: "too far apart",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "EUR", 14, "transfer"),
			},
			want: nil,
		},
		{
			name: "at the edge of the window",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "EUR", 3+ownTransferWindowDays, "transfer"),
			},
			want: []string{"in+out"},
		},
		{
			name: "one day past the window",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "EUR", 4+ownTransferWindowDays, "transfer"),
			},
			want: nil,
		},
		{
			name: "not the same amount",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 49900, "EUR", 4, "transfer"),
			},
			want: nil,
		},
		{
			name: "two currencies",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "GBP", 3, "transfer"),
			},
			want: nil,
		},
		{
			name: "the same amount twice in a week",
			rows: []store.Transaction{
				tx("out1", "current", -20000, "EUR", 1, "transfer"),
				tx("in1", "savings", 20000, "EUR", 1, "transfer"),
				tx("out3", "current", -20000, "EUR", 3, "transfer"),
				tx("in3", "savings", 20000, "EUR", 3, "transfer"),
			},
			want: []string{"in1+out1", "in3+out3"},
		},
		{
			name: "one arrival, two possible sources",
			rows: []store.Transaction{
				tx("outA", "current", -20000, "EUR", 1, "transfer"),
				tx("outB", "third", -20000, "EUR", 1, "transfer"),
				tx("in", "savings", 20000, "EUR", 1, "transfer"),
			},
			want: nil,
		},
		{
			name: "the text names the other account",
			rows: []store.Transaction{
				tx("outA", "current", -20000, "EUR", 1, "transfer"),
				tx("outB", "third", -20000, "EUR", 1, "transfer"),
				tx("in", "savings", 20000, "EUR", 1, "from account 3333"),
			},
			want: []string{"in+outB"},
		},
		{
			name: "sent to an account wimm does not hold",
			rows: []store.Transaction{
				tx("out", "current", -100000, "EUR", 3, "Ana Reis"),
			},
			want: nil,
		},
		{
			name: "a refund that happens to match",
			rows: []store.Transaction{
				tx("paid", "current", -5000, "EUR", 3, "A shop"),
				tx("refund", "current", 5000, "EUR", 5, "A shop"),
			},
			want: nil,
		},
		{
			name: "not yet settled",
			rows: []store.Transaction{
				tx("out", "current", -50000, "EUR", 3, "transfer"),
				func() store.Transaction {
					r := tx("in", "savings", 50000, "EUR", 4, "transfer")
					r.Status = store.StatusPending
					return r
				}(),
			},
			want: nil,
		},
		{
			name: "a grant is never either side",
			rows: []store.Transaction{
				tx("out", "granted", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "EUR", 4, "transfer"),
			},
			want: nil,
		},
		{
			name: "a number suffix under the minimum is not evidence",
			rows: []store.Transaction{
				tx("outA", "current", -20000, "EUR", 1, "transfer"),
				tx("outB", "shortsuffix", -20000, "EUR", 1, "transfer"),
				tx("in", "savings", 20000, "EUR", 1, "from 44"),
			},
			want: nil,
		},
		{
			name: "holder_name is not evidence",
			rows: []store.Transaction{
				tx("outA", "current", -20000, "EUR", 1, "transfer"),
				tx("outB", "third", -20000, "EUR", 1, "transfer"),
				tx("in", "savings", 20000, "EUR", 1, "from Ana Reis"),
			},
			want: nil,
		},
		{
			name: "a name a second owned account carries is not evidence",
			rows: []store.Transaction{
				tx("outA", "twinA", -20000, "EUR", 1, "transfer"),
				tx("outB", "twinB", -20000, "EUR", 1, "transfer"),
				tx("in", "savings", 20000, "EUR", 1, "from Revolut"),
			},
			accounts: []OwnedAccount{
				{ID: "savings", Name: "Savings Pot", HouseholdName: "Savings", NumberSuffix: "2222"},
				{ID: "twinA", Name: "Ana Reis", HouseholdName: "Revolut", NumberSuffix: "7928"},
				{ID: "twinB", Name: "Ana Reis", HouseholdName: "Revolut", NumberSuffix: "7928"},
			},
			want: nil,
		},
		{
			name: "a number suffix that is not digits is not evidence",
			rows: []store.Transaction{
				tx("outA", "current", -20000, "EUR", 1, "transfer"),
				tx("outB", "paypal", -20000, "EUR", 1, "transfer"),
				tx("in", "savings", 20000, "EUR", 1, "paid at shop.com"),
			},
			accounts: append(slices.Clone(testAccounts),
				OwnedAccount{ID: "paypal", Name: "Personal", HouseholdName: "PayPal", NumberSuffix: ".com"}),
			want: nil,
		},
		{
			name: "the closer date wins over evidence",
			rows: []store.Transaction{
				tx("near", "current", -20000, "EUR", 3, "transfer"),
				tx("far", "third", -20000, "EUR", 1, "transfer"),
				tx("in", "savings", 20000, "EUR", 3, "from account 3333"),
			},
			want: []string{"in+near"},
		},
		{
			name: "a tie broken elsewhere pairs in a later round",
			rows: []store.Transaction{
				// inTied is a day from both outs, so neither takes it in the
				// first round. inPull is on third, so only outA can be its
				// partner; once that pair is made outB is inTied's only
				// candidate and the tie is gone.
				tx("outA", "current", -20000, "EUR", 2, "transfer"),
				tx("outB", "third", -20000, "EUR", 2, "transfer"),
				tx("inPull", "third", 20000, "EUR", 2, "transfer"),
				tx("inTied", "savings", 20000, "EUR", 3, "transfer"),
			},
			want: []string{"inPull+outA", "inTied+outB"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			accounts := c.accounts
			if accounts == nil {
				accounts = testAccounts
			}
			got := pairsOf(t, OwnTransfers(c.rows, accounts))
			if !slices.Equal(got, c.want) {
				t.Errorf("OwnTransfers() = %v, want %v", got, c.want)
			}
		})
	}
}

// A pair is a fact about the rows, so shuffling them cannot change it.
func TestOwnTransfersDoesNotDependOnRowOrder(t *testing.T) {
	t.Parallel()

	rows := []store.Transaction{
		tx("out1", "current", -20000, "EUR", 1, "transfer"),
		tx("in1", "savings", 20000, "EUR", 1, "transfer"),
		tx("out3", "current", -20000, "EUR", 3, "transfer"),
		tx("in3", "savings", 20000, "EUR", 4, "transfer"),
		tx("outA", "current", -55000, "EUR", 6, "transfer"),
		tx("outB", "third", -55000, "EUR", 6, "transfer"),
		tx("inC", "savings", 55000, "EUR", 6, "transfer"),
		tx("shop", "current", -1250, "EUR", 2, "A shop"),
	}
	want := pairsOf(t, OwnTransfers(rows, testAccounts))
	if len(want) == 0 {
		t.Fatal("the fixture pairs nothing, so order proves nothing")
	}

	r := rand.New(rand.NewPCG(1, 2))
	for i := range 50 {
		shuffled := slices.Clone(rows)
		r.Shuffle(len(shuffled), func(a, b int) {
			shuffled[a], shuffled[b] = shuffled[b], shuffled[a]
		})
		if got := pairsOf(t, OwnTransfers(shuffled, testAccounts)); !slices.Equal(got, want) {
			t.Fatalf("shuffle %d: OwnTransfers() = %v, want %v", i, got, want)
		}
	}
}

// The date window makes the candidate scan linear in the rows near each other
// rather than in the bucket. It is an optimisation, so it is pinned against
// the reading of the rule that does not have it.
func TestOwnTransfersMatchesAnUnwindowedScan(t *testing.T) {
	t.Parallel()

	accounts := []OwnedAccount{
		{ID: "a", Name: "Account A", HouseholdName: "Alpha", NumberSuffix: "1111"},
		{ID: "b", Name: "Account B", HouseholdName: "Bravo", NumberSuffix: "2222"},
		{ID: "c", Name: "Account C", HouseholdName: "Charlie", NumberSuffix: "3333"},
	}
	texts := []string{"transfer", "from 1111", "from 2222", "from 3333", "A shop"}

	r := rand.New(rand.NewPCG(5, 6))
	for trial := range 200 {
		rows := make([]store.Transaction, 0, 40)
		for i := range 40 {
			minor := int64(r.IntN(3)+1) * 10000
			if r.IntN(2) == 0 {
				minor = -minor
			}
			rows = append(rows, tx(
				fmt.Sprintf("t%d", i),
				accounts[r.IntN(len(accounts))].ID,
				minor, "EUR", 1+r.IntN(12), texts[r.IntN(len(texts))],
			))
		}
		want := pairsOf(t, unwindowedOwnTransfers(rows, accounts))
		if got := pairsOf(t, OwnTransfers(rows, accounts)); !slices.Equal(got, want) {
			t.Fatalf("trial %d: OwnTransfers() = %v, unwindowed = %v", trial, got, want)
		}
	}
}

// unwindowedOwnTransfers is the rule stated plainly: every out against every
// in, filtered by the window rather than seeking it. It exists only to be
// compared against.
func unwindowedOwnTransfers(rows []store.Transaction, accounts []OwnedAccount) map[string]string {
	owned := map[string]bool{}
	for _, a := range accounts {
		owned[a.ID] = true
	}
	tokens := discriminatingTokens(accounts)

	var outs, ins []int
	for i, r := range rows {
		switch {
		case r.Status != store.StatusBooked || r.AmountMinor == 0 || !owned[r.AccountID]:
		case r.AmountMinor < 0:
			outs = append(outs, i)
		default:
			ins = append(ins, i)
		}
	}

	paired := map[string]string{}
	for {
		best := map[int]candidate{}
		tied := map[int]bool{}
		for _, o := range outs {
			for _, i := range ins {
				if rows[o].Currency != rows[i].Currency ||
					rows[o].AmountMinor != -rows[i].AmountMinor ||
					rows[o].AccountID == rows[i].AccountID {
					continue
				}
				gap := max(daysBetween(rows[o].BookingDate, rows[i].BookingDate),
					daysBetween(rows[i].BookingDate, rows[o].BookingDate))
				if gap > ownTransferWindowDays {
					continue
				}
				noEvidence := 1
				if namesAccount(rows[o], tokens[rows[i].AccountID]) ||
					namesAccount(rows[i], tokens[rows[o].AccountID]) {
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
			return paired
		}
		outs = slices.DeleteFunc(outs, func(o int) bool { return done[o] })
		ins = slices.DeleteFunc(ins, func(i int) bool { return done[i] })
	}
}

// Every service method runs this over the rows it already holds, so it has to
// stay far below the request it sits inside.
func TestOwnTransfersHandlesAWholeLedger(t *testing.T) {
	t.Parallel()

	const (
		rowCount = 50_000
		budget   = 2 * time.Second
	)

	accounts := make([]OwnedAccount, 6)
	for i := range accounts {
		accounts[i] = OwnedAccount{
			ID:            fmt.Sprintf("account-%d", i),
			Name:          fmt.Sprintf("Account %d", i),
			HouseholdName: fmt.Sprintf("House %d", i),
			NumberSuffix:  fmt.Sprintf("%04d", 1000+i),
		}
	}

	r := rand.New(rand.NewPCG(3, 4))
	rows := make([]store.Transaction, 0, rowCount)
	for i := range rowCount {
		// Amounts repeat heavily, so buckets are large and the ranking runs,
		// spread over the 25 months MonthHistory reads.
		minor := int64(r.IntN(200)+1) * 500
		if i%2 == 0 {
			minor = -minor
		}
		rows = append(rows, tx(
			fmt.Sprintf("t%d", i),
			accounts[r.IntN(len(accounts))].ID,
			minor, "EUR", 1+r.IntN(25*30), "transfer",
		))
	}

	start := time.Now()
	OwnTransfers(rows, accounts)
	if elapsed := time.Since(start); elapsed > budget {
		t.Errorf("OwnTransfers() over %d rows took %s, budget %s", rowCount, elapsed, budget)
	}
}
