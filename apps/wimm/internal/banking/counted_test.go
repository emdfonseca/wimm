package banking

import (
	"slices"
	"testing"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

func TestCountedRows(t *testing.T) {
	t.Parallel()

	// ada owns her own account and the joint one; the savings account is hers
	// too. Every pair below is found across all three.
	accounts := []OwnedAccount{
		{ID: "own", Name: "Ada Current", HouseholdName: "Ada", NumberSuffix: "1111"},
		{ID: "savings", Name: "Ada Savings", HouseholdName: "Savings", NumberSuffix: "2222"},
		{ID: "joint", Name: "Joint Account", HouseholdName: "Joint", NumberSuffix: "3333"},
	}

	cases := []struct {
		name string
		rows []store.Transaction
		// scope is the account ids the scope in force counts.
		scope []string
		// wantCounted are the row ids the scope counts, in order.
		wantCounted []string
		// wantLeftOut is how many pairs the scope left out, and their total.
		wantLeftOut int
		wantTotal   int64
	}{
		{
			name: "a pair inside the scope is left out",
			rows: []store.Transaction{
				tx("salary", "own", 245000, "EUR", 1, "Employer"),
				tx("out", "own", -50000, "EUR", 3, "transfer"),
				tx("in", "savings", 50000, "EUR", 4, "transfer"),
				tx("shop", "own", -1900, "EUR", 5, "A shop"),
			},
			scope:       []string{"own", "savings", "joint"},
			wantCounted: []string{"salary", "shop"},
			wantLeftOut: 1,
			wantTotal:   50000,
		},
		{
			name: "a pair crossing the scope, out row inside",
			rows: []store.Transaction{
				tx("out", "own", -80000, "EUR", 3, "transfer"),
				tx("in", "joint", 80000, "EUR", 3, "transfer"),
			},
			scope:       []string{"own", "savings"},
			wantCounted: []string{"out"},
			wantLeftOut: 0,
		},
		{
			name: "a pair crossing the scope, in row inside",
			rows: []store.Transaction{
				tx("out", "own", -80000, "EUR", 3, "transfer"),
				tx("in", "joint", 80000, "EUR", 3, "transfer"),
			},
			scope:       []string{"joint"},
			wantCounted: []string{"in"},
			wantLeftOut: 0,
		},
		{
			name: "an unpaired row is always counted",
			rows: []store.Transaction{
				tx("sent", "own", -100000, "EUR", 3, "Ada Reis"),
			},
			scope:       []string{"own", "savings", "joint"},
			wantCounted: []string{"sent"},
			wantLeftOut: 0,
		},
		{
			name: "a pair straddling two scopes of one member is counted once each",
			rows: []store.Transaction{
				tx("out", "own", -80000, "EUR", 3, "transfer"),
				tx("in", "joint", 80000, "EUR", 3, "transfer"),
				tx("shop", "joint", -2500, "EUR", 4, "A shop"),
			},
			scope:       []string{"own", "savings", "joint"},
			wantCounted: []string{"shop"},
			wantLeftOut: 1,
			wantTotal:   80000,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			pairs := OwnTransfers(c.rows, accounts)
			counted, leftOut := countedRows(c.rows, pairs, c.scope)

			var got []string
			for _, r := range counted {
				got = append(got, r.ID)
			}
			if !slices.Equal(got, c.wantCounted) {
				t.Errorf("counted = %v, want %v", got, c.wantCounted)
			}
			if leftOut.Pairs != c.wantLeftOut {
				t.Errorf("left out %d pairs, want %d", leftOut.Pairs, c.wantLeftOut)
			}
			if leftOut.Total != c.wantTotal {
				t.Errorf("left-out total = %d, want %d", leftOut.Total, c.wantTotal)
			}
		})
	}
}

// A member who owns one side of a movement has no pair at all, so nothing is
// left out and the row is ordinary money in. A label there would tell them the
// other account exists.
func TestCountedRowsForAMemberWhoOwnsOneSide(t *testing.T) {
	t.Parallel()

	// grace owns the joint account and nothing else.
	graceAccounts := []OwnedAccount{
		{ID: "joint", Name: "Joint Account", HouseholdName: "Joint", NumberSuffix: "3333"},
	}
	rows := []store.Transaction{tx("in", "joint", 80000, "EUR", 3, "transfer")}

	pairs := OwnTransfers(rows, graceAccounts)
	if len(pairs) != 0 {
		t.Fatalf("OwnTransfers() = %v, want no pair for a member who owns one side", pairs)
	}
	counted, leftOut := countedRows(rows, pairs, []string{"joint"})
	if len(counted) != 1 || counted[0].ID != "in" {
		t.Errorf("counted = %v, want the arriving row", counted)
	}
	if leftOut.Pairs != 0 {
		t.Errorf("left out %d pairs, want 0", leftOut.Pairs)
	}
}
