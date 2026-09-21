package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
	"github.com/xuuid/wimm/apps/wimm/internal/store/storetest"
)

// monthAccount stores a sync for a member on a fresh account and returns it.
func monthAccount(t *testing.T, ctx context.Context, db *store.DB, owner string, txs ...store.Transaction) string {
	t.Helper()
	_, stored := connect(t, ctx, db, owner, 90*24*time.Hour, account("hash-m", "Conta à Ordem", "0538"))
	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID, txs, day(28)); err != nil {
		t.Fatalf("WriteAccountTransactions: %v", err)
	}
	return stored[0].ID
}

func TestWindowSumsSplitMoneyInFromMoneyOutAndSkipPendingRows(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	id := monthAccount(t, ctx, db, ada.ID,
		booked("salary", 2, 200_000),
		booked("rent", 3, -80_000),
		booked("coffee", 5, -350),
		pending("holds", 6, -9_999),
		booked("before", 1, -111),
		booked("after", 20, -222),
	)

	sums, err := db.OwnedWindowSums(ctx, ada.ID, []string{id}, day(2), day(6))
	if err != nil {
		t.Fatalf("OwnedWindowSums: %v", err)
	}
	if len(sums) != 1 {
		t.Fatalf("got %+v, want one currency", sums)
	}
	got := sums[0]
	if got.Currency != "EUR" || got.InMinor != 200_000 || got.OutMinor != 80_350 || got.Rows != 3 {
		t.Errorf("got %+v, want in 200000, out 80350 over 3 booked rows in [2, 6)", got)
	}
}

func TestWindowSumsNeverReadAnotherMembersAccounts(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	id := monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))
	if err := db.SetAccountLevel(ctx, id, grace.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatal(err)
	}

	sums, err := db.OwnedWindowSums(ctx, grace.ID, []string{id}, day(1), day(28))
	if err != nil {
		t.Fatal(err)
	}
	if len(sums) != 0 {
		t.Errorf("a member granted details read sums %+v", sums)
	}
	rows, err := db.OwnedOutgoing(ctx, grace.ID, []string{id}, day(1), day(28))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Errorf("a member granted details read %d outgoing rows", len(rows))
	}
}

func TestOutgoingRowsAreBookedNegativeAndInsideTheWindow(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	id := monthAccount(t, ctx, db, ada.ID,
		booked("salary", 2, 200_000),
		booked("rent", 3, -80_000),
		booked("coffee", 5, -350),
		pending("holds", 6, -9_999),
		booked("late", 9, -1),
	)

	rows, err := db.OwnedOutgoing(ctx, ada.ID, []string{id}, day(1), day(9))
	if err != nil {
		t.Fatalf("OwnedOutgoing: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want the rent and the coffee only", len(rows))
	}
	for _, r := range rows {
		if r.AmountMinor >= 0 || r.Status != store.StatusBooked {
			t.Errorf("row %+v is not a booked payment", r)
		}
	}
}

func TestALeftOutAccountIsInNoWindow(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	id := monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))
	if err := db.SetAccountLeftOut(ctx, id, true); err != nil {
		t.Fatal(err)
	}
	sums, err := db.OwnedWindowSums(ctx, ada.ID, []string{id}, day(1), day(28))
	if err != nil {
		t.Fatal(err)
	}
	if len(sums) != 0 {
		t.Errorf("got %+v, want nothing for a left-out account", sums)
	}
}

func TestFullAccessCountsOwnersAndDetailsGrantsButNotBalanceGrants(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	linus := member(t, ctx, db, "linus@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour,
		account("hash-1", "Joint", "0538"), account("hash-2", "Personal", "5594"))
	if err := db.SetAccountOwners(ctx, stored[0].ID, []string{ada.ID, grace.ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAccountLevel(ctx, stored[0].ID, linus.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAccountLevel(ctx, stored[1].ID, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatal(err)
	}

	counts, err := db.FullAccessCounts(ctx)
	if err != nil {
		t.Fatalf("FullAccessCounts: %v", err)
	}
	if counts[stored[0].ID] != 3 {
		t.Errorf("joint = %d, want 3: two owners and one details grant", counts[stored[0].ID])
	}
	if counts[stored[1].ID] != 1 {
		t.Errorf("personal = %d, want 1: a balance grant is not full access", counts[stored[1].ID])
	}
}

func TestAnAccountIdTheMemberDoesNotOwnMatchesNothingInEveryOwnedQuery(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	id := monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))
	if err := db.SetAccountLevel(ctx, id, grace.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatal(err)
	}
	ids := []string{id}

	if sums, err := db.OwnedWindowSums(ctx, grace.ID, ids, day(1), day(28)); err != nil || len(sums) != 0 {
		t.Errorf("OwnedWindowSums = %+v, %v, want nothing", sums, err)
	}
	if months, err := db.OwnedMonthlySums(ctx, grace.ID, ids, day(1), day(28)); err != nil || len(months) != 0 {
		t.Errorf("OwnedMonthlySums = %+v, %v, want nothing", months, err)
	}
	if rows, err := db.OwnedBooked(ctx, grace.ID, ids, day(1), day(28)); err != nil || len(rows) != 0 {
		t.Errorf("OwnedBooked = %d rows, %v, want nothing", len(rows), err)
	}
	if rows, err := db.OwnedOutgoing(ctx, grace.ID, ids, day(1), day(28)); err != nil || len(rows) != 0 {
		t.Errorf("OwnedOutgoing = %d rows, %v, want nothing", len(rows), err)
	}
	if accounts, err := db.OwnedAccountsForTrend(ctx, grace.ID, ids); err != nil || len(accounts) != 0 {
		t.Errorf("OwnedAccountsForTrend = %+v, %v, want nothing", accounts, err)
	}
}

func TestAnEmptyAccountScopeIsAnEmptyScope(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	monthAccount(t, ctx, db, ada.ID, booked("rent", 3, -80_000))

	for _, ids := range [][]string{nil, {}} {
		sums, err := db.OwnedWindowSums(ctx, ada.ID, ids, day(1), day(28))
		if err != nil || len(sums) != 0 {
			t.Errorf("ids %#v: got %+v, %v, want nothing", ids, sums, err)
		}
	}
}

func TestOwnedBookedIsBothDirectionsBookedOnlyAndBoundedByArguments(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	id := monthAccount(t, ctx, db, ada.ID,
		booked("salary", 2, 200_000),
		booked("rent", 3, -80_000),
		pending("holds", 4, -9_999),
		booked("before", 1, -111),
		booked("end", 9, -1),
	)

	rows, err := db.OwnedBooked(ctx, ada.ID, []string{id}, day(2), day(9))
	if err != nil {
		t.Fatalf("OwnedBooked: %v", err)
	}
	if len(rows) != 2 || rows[0].DedupKey != "rent" || rows[1].DedupKey != "salary" {
		t.Fatalf("got %+v, want rent then salary: booked, both directions, in [2, 9)", rows)
	}
}

func TestOwnedMonthlySumsGroupByCurrencyAndCalendarMonth(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	feb := booked("feb", 1, -500)
	feb.BookingDate = time.Date(2026, time.February, 27, 0, 0, 0, 0, time.UTC)
	id := monthAccount(t, ctx, db, ada.ID,
		feb,
		booked("salary", 2, 200_000),
		booked("rent", 3, -80_000),
		pending("holds", 4, -9_999),
		booked("coffee", 28, -350),
	)

	months, err := db.OwnedMonthlySums(ctx, ada.ID, []string{id},
		time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC), day(29))
	if err != nil {
		t.Fatalf("OwnedMonthlySums: %v", err)
	}
	want := []store.MonthSum{
		{Currency: "EUR", Month: time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC), OutMinor: 500},
		{Currency: "EUR", Month: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC), InMinor: 200_000, OutMinor: 80_350},
	}
	if len(months) != len(want) {
		t.Fatalf("got %+v, want %+v", months, want)
	}
	for i := range want {
		if months[i] != want[i] {
			t.Errorf("month %d = %+v, want %+v", i, months[i], want[i])
		}
	}
}

func TestOwnedBookedStaysFastAtFiftyThousandRows(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Conta a Ordem", "0538"))
	accountID := stored[0].ID

	const total = 50_000
	start := time.Date(2024, time.September, 1, 0, 0, 0, 0, time.UTC)
	rows := make([][]any, total)
	for i := range total {
		rows[i] = []any{
			accountID, string(store.StatusBooked), fmt.Sprintf("bulk-%d", i), int64(-100 - i),
			"EUR", start.AddDate(0, 0, i%750),
		}
	}
	if _, err := db.Pool().CopyFrom(ctx,
		pgx.Identifier{"transactions"},
		[]string{"account_id", "status", "dedup_key", "amount_minor", "currency", "booking_date"},
		pgx.CopyFromRows(rows),
	); err != nil {
		t.Fatalf("seeding 50,000 transactions: %v", err)
	}

	ids := []string{accountID}
	to := start.AddDate(0, 0, 800)
	began := time.Now()
	got, err := db.OwnedBooked(ctx, ada.ID, ids, start, to)
	elapsed := time.Since(began)
	if err != nil {
		t.Fatalf("OwnedBooked: %v", err)
	}
	t.Logf("OwnedBooked over %d rows took %s", total, elapsed)
	if len(got) != total {
		t.Errorf("OwnedBooked returned %d rows, want %d", len(got), total)
	}
	if elapsed > 3*time.Second {
		t.Errorf("OwnedBooked took %s over %d rows, want well under 3s", elapsed, total)
	}
}
