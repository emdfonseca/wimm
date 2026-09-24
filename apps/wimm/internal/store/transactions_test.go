package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store/storetest"
)

func day(n int) time.Time { return time.Date(2026, time.March, n, 0, 0, 0, 0, time.UTC) }

func booked(key string, d int, minor int64) store.Transaction {
	return store.Transaction{
		Status: store.StatusBooked, DedupKey: key, AmountMinor: minor,
		Currency: "EUR", BookingDate: day(d), CounterpartyName: "Padaria Ribeiro",
	}
}

func pending(key string, d int, minor int64) store.Transaction {
	t := booked(key, d, minor)
	t.Status = store.StatusPending
	return t
}

// ledger stores a sync and returns the account it was stored against.
func ledger(t *testing.T, ctx context.Context, db *store.DB, owner string, txs ...store.Transaction) string {
	t.Helper()
	_, stored := connect(t, ctx, db, owner, 90*24*time.Hour, account("hash-1", "Conta à Ordem", "0538"))
	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID, txs, day(10)); err != nil {
		t.Fatalf("WriteAccountTransactions: %v", err)
	}
	return stored[0].ID
}

func readPage(t *testing.T, ctx context.Context, db *store.DB, q store.LedgerQuery) store.LedgerPage {
	t.Helper()
	page, err := db.Ledger(ctx, q)
	if err != nil {
		t.Fatalf("Ledger: %v", err)
	}
	return page
}

func keys(page store.LedgerPage) []string {
	out := make([]string, 0, len(page.Transactions))
	for _, t := range page.Transactions {
		out = append(out, t.DedupKey)
	}
	return out
}

func same(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// The same payment made twice is two transactions; the same transaction read
// twice is one. A household that paid twice needs to see that it paid twice,
// and the occurrence index is the whole reason both hold.
func TestOverlappingSyncsKeepARealDuplicateAndDropARereadOne(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	// Two identical coffees, same day, same shop, same amount, and one salary.
	first := []store.Transaction{
		booked("coffee", 4, -250),
		booked("coffee", 4, -250),
		booked("salary", 1, 240_000),
	}
	accountID := ledger(t, ctx, db, ada.ID, first...)

	if got := count(t, ctx, db, "transactions"); got != 3 {
		t.Fatalf("the first sync stored %d transactions, want 3", got)
	}

	// The next sync overlaps: it returns both coffees again, the salary again,
	// and a third coffee that is genuinely new.
	second := []store.Transaction{
		booked("coffee", 4, -250),
		booked("coffee", 4, -250),
		booked("coffee", 4, -250),
		booked("salary", 1, 240_000),
	}
	result, err := db.WriteAccountTransactions(ctx, accountID, second, day(10))
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if result.BookedInserted != 1 {
		t.Errorf("the overlapping sync inserted %d rows, want exactly the 1 that was new", result.BookedInserted)
	}
	if result.BookedUpdated != 3 {
		t.Errorf("the overlapping sync found %d rows already stored, want 3", result.BookedUpdated)
	}
	if got := count(t, ctx, db, "transactions"); got != 4 {
		t.Errorf("%d transactions after two overlapping syncs, want 4", got)
	}
}

// Pending is a replaceable set and booked is append-only, which is what removes
// the pending-to-booked matching problem rather than solving it.
func TestAnUnsettledTransactionIsReplacedAndItsBookedFormArrivesAlone(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	// The bank returns it unsettled: no reference of its own yet, one amount,
	// one date.
	accountID := ledger(t, ctx, db, ada.ID, pending("pend-1", 5, -4500))

	page := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(page.Transactions) != 1 || page.Transactions[0].Status != store.StatusPending {
		t.Fatalf("after the first sync the ledger holds %+v, want one unsettled row", page.Transactions)
	}

	// It settles: a different reference, a different amount — the tip landed —
	// and a different date. Nothing tries to match the two.
	if _, err := db.WriteAccountTransactions(ctx, accountID,
		[]store.Transaction{booked("ref-settled", 6, -5200)}, day(10)); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	page = readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(page.Transactions) != 1 {
		t.Fatalf("the ledger holds %d rows, want 1: the household must not be shown both versions", len(page.Transactions))
	}
	if page.Transactions[0].Status != store.StatusBooked || page.Transactions[0].DedupKey != "ref-settled" {
		t.Errorf("the surviving row is %+v, want the settled one", page.Transactions[0])
	}
}

// A bank that stops returning an unsettled transaction is saying it never
// happened, and it stops appearing.
func TestAnUnsettledTransactionTheBankDropsStopsAppearing(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	accountID := ledger(t, ctx, db, ada.ID, pending("pend-1", 5, -4500), booked("ref-1", 4, -1190))

	if _, err := db.WriteAccountTransactions(ctx, accountID,
		[]store.Transaction{booked("ref-1", 4, -1190)}, day(10)); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	page := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(page.Transactions) != 1 || page.Transactions[0].DedupKey != "ref-1" {
		t.Errorf("the ledger holds %+v, want only the settled transaction", keys(page))
	}
}

// Paging seeks on the sort key in both directions, and returns through the same
// transactions in the same order.
func TestPagingReachesBothEndsWithNothingRepeatedOrSkipped(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	var txs []store.Transaction
	for i := 1; i <= 9; i++ {
		txs = append(txs, booked("ref-"+string(rune('a'+i-1)), i, int64(-100*i)))
	}
	ledger(t, ctx, db, ada.ID, txs...)

	// Newest first, three at a time, all the way to the oldest.
	var forwards [][]string
	page := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 3})
	forwards = append(forwards, keys(page))
	if !page.HasOlder || page.HasNewer {
		t.Errorf("the newest page offers older=%v newer=%v, want older only", page.HasOlder, page.HasNewer)
	}

	for page.HasOlder {
		last := page.Transactions[len(page.Transactions)-1]
		page = readPage(t, ctx, db, store.LedgerQuery{
			MemberID: ada.ID, Limit: 3, Older: true,
			Cursor: store.Cursor{BookingDate: last.BookingDate, ID: last.ID},
		})
		forwards = append(forwards, keys(page))
		if len(forwards) > 5 {
			t.Fatal("paging older did not terminate")
		}
	}

	if len(forwards) != 3 {
		t.Fatalf("read %d pages of 3 over 9 transactions, want 3: %v", len(forwards), forwards)
	}
	if page.HasOlder {
		t.Error("the oldest page still offers older transactions")
	}

	seen := map[string]int{}
	for _, p := range forwards {
		for _, k := range p {
			seen[k]++
		}
	}
	if len(seen) != 9 {
		t.Errorf("paging saw %d distinct transactions, want 9", len(seen))
	}
	for k, n := range seen {
		if n != 1 {
			t.Errorf("%s was read %d times", k, n)
		}
	}

	// And back towards today through the same transactions in the same order.
	for i := len(forwards) - 1; i > 0; i-- {
		first := page.Transactions[0]
		page = readPage(t, ctx, db, store.LedgerQuery{
			MemberID: ada.ID, Limit: 3,
			Cursor: store.Cursor{BookingDate: first.BookingDate, ID: first.ID},
		})
		if !same(keys(page), forwards[i-1]) {
			t.Fatalf("coming back gave %v, want %v", keys(page), forwards[i-1])
		}
	}
	if page.HasNewer {
		t.Error("the newest page reached by going back still offers newer transactions")
	}
}

// The whole reason paging is a seek: a sync inserts at the newest end while a
// member is reading, and the page they are on must not move under them.
func TestAPageIsUnchangedByTransactionsArrivingAtTheNewestEnd(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	var txs []store.Transaction
	for i := 1; i <= 6; i++ {
		txs = append(txs, booked("ref-"+string(rune('a'+i-1)), i, int64(-100*i)))
	}
	accountID := ledger(t, ctx, db, ada.ID, txs...)

	newest := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 3})
	last := newest.Transactions[len(newest.Transactions)-1]
	cursor := store.Cursor{BookingDate: last.BookingDate, ID: last.ID}

	older := readPage(t, ctx, db, store.LedgerQuery{
		MemberID: ada.ID, Limit: 3, Older: true, Cursor: cursor})
	before := keys(older)

	// Twelve arrive at the newest end, which is what every visit does.
	var arriving []store.Transaction
	for i := range 12 {
		arriving = append(arriving, booked("new-"+string(rune('a'+i)), 20+i, -1))
	}
	if _, err := db.WriteAccountTransactions(ctx, accountID, append(txs, arriving...), day(25)); err != nil {
		t.Fatalf("a sync while the member was reading: %v", err)
	}

	after := readPage(t, ctx, db, store.LedgerQuery{
		MemberID: ada.ID, Limit: 3, Older: true, Cursor: cursor})
	if !same(keys(after), before) {
		t.Errorf("the page moved under the member: %v became %v", before, keys(after))
	}
}

// Everything fits on one page, so no way to page is offered.
func TestNoPagingIsOfferedWhenEverythingFits(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	ledger(t, ctx, db, ada.ID, booked("ref-a", 1, -100), booked("ref-b", 2, -200))

	page := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if page.HasOlder || page.HasNewer {
		t.Errorf("older=%v newer=%v on a ledger that fits, want neither", page.HasOlder, page.HasNewer)
	}
}

// Where the member is, is a date.
func TestAPageStatesTheSpanOfDatesItCovers(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	ledger(t, ctx, db, ada.ID, booked("ref-a", 1, -100), booked("ref-b", 5, -200), booked("ref-c", 9, -300))

	page := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if !page.Newest.Equal(day(9)) || !page.Oldest.Equal(day(1)) {
		t.Errorf("the page spans %s to %s, want %s to %s", page.Oldest, page.Newest, day(1), day(9))
	}
}

// Transactions are owner-only. A member granted balance or details on an
// account they do not own sees none of it, and is not told how many there are:
// seeing a balance and seeing what was spent are different sentences.
func TestAGrantedMemberReadsNoTransactionsAndNoCount(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour,
		account("hash-1", "Conta à Ordem", "0538"),
		account("hash-2", "Poupança", "5594"))

	if err := db.SetAccountLevel(ctx, stored[0].ID, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatalf("granting balance: %v", err)
	}
	if err := db.SetAccountLevel(ctx, stored[1].ID, grace.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatalf("granting details: %v", err)
	}
	for _, a := range stored {
		if _, err := db.WriteAccountTransactions(ctx, a.ID,
			[]store.Transaction{booked("ref-"+a.ID, 4, -1190)}, day(10)); err != nil {
			t.Fatalf("storing transactions: %v", err)
		}
	}

	page := readPage(t, ctx, db, store.LedgerQuery{MemberID: grace.ID, Limit: 50})
	if len(page.Transactions) != 0 {
		t.Errorf("a member holding balance and details read %d transactions, want none", len(page.Transactions))
	}

	n, err := db.CountLedger(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("CountLedger: %v", err)
	}
	if n != 0 {
		t.Errorf("a member holding balance and details is told there are %d transactions, want 0", n)
	}

	owner := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(owner.Transactions) != 2 {
		t.Errorf("the owner reads %d transactions, want both accounts'", len(owner.Transactions))
	}
}

// Both owners of a joint account see its transactions in full, and neither is
// told what the other sees.
func TestBothOwnersOfAJointAccountReadItsTransactions(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))
	if err := db.SetAccountOwners(ctx, stored[0].ID, []string{ada.ID, grace.ID}); err != nil {
		t.Fatalf("setting owners: %v", err)
	}
	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID,
		[]store.Transaction{booked("ref-a", 4, -1190)}, day(10)); err != nil {
		t.Fatalf("storing transactions: %v", err)
	}

	for _, m := range []struct {
		who string
		id  string
	}{{"ada", ada.ID}, {"grace", grace.ID}} {
		page := readPage(t, ctx, db, store.LedgerQuery{MemberID: m.id, Limit: 50})
		if len(page.Transactions) != 1 {
			t.Errorf("%s reads %d transactions on a jointly owned account, want 1", m.who, len(page.Transactions))
		}
	}
}

// Narrowing to one account is the same list filtered, and an account the member
// does not own yields nothing however it is asked for.
func TestNarrowingToOneAccountIsScopedToOwnership(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour,
		account("hash-1", "Conta à Ordem", "0538"),
		account("hash-2", "Poupança", "5594"))
	for _, a := range stored {
		if _, err := db.WriteAccountTransactions(ctx, a.ID,
			[]store.Transaction{booked("ref-"+a.ID, 4, -1190)}, day(10)); err != nil {
			t.Fatalf("storing transactions: %v", err)
		}
	}

	narrowed := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, AccountID: stored[0].ID, Limit: 50})
	if len(narrowed.Transactions) != 1 || narrowed.Transactions[0].AccountID != stored[0].ID {
		t.Errorf("narrowing gave %+v, want only that account's transactions", narrowed.Transactions)
	}

	widened := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(widened.Transactions) != 2 {
		t.Errorf("widening gave %d transactions, want 2", len(widened.Transactions))
	}

	notTheirs := readPage(t, ctx, db, store.LedgerQuery{MemberID: grace.ID, AccountID: stored[0].ID, Limit: 50})
	if len(notTheirs.Transactions) != 0 {
		t.Errorf("a member read %d transactions of an account they do not own", len(notTheirs.Transactions))
	}
}

// A capped first fill records how far it actually reached, so the next sync
// continues rather than starting the history again — which at a bank granting
// one day of access is the difference between a ledger and a screen that
// empties every night.
func TestAPartialFillRecordsHowFarItReached(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Conta à Ordem", "0538"))

	before, err := db.AccountSyncState(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("AccountSyncState: %v", err)
	}
	if before.SyncedThrough != nil || before.SyncedAt != nil {
		t.Error("an account nothing has been read from reports a sync")
	}

	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID,
		[]store.Transaction{booked("ref-a", 4, -1190)}, day(4)); err != nil {
		t.Fatalf("a capped first fill: %v", err)
	}

	after, err := db.AccountSyncState(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("AccountSyncState: %v", err)
	}
	if after.SyncedThrough == nil || !after.SyncedThrough.Equal(day(4)) {
		t.Errorf("SyncedThrough = %v, want the date the fill actually reached", after.SyncedThrough)
	}
	if after.SyncedAt == nil {
		t.Error("a finished sync recorded no time")
	}
	if after.Oldest == nil || !after.Oldest.Equal(day(4)) {
		t.Errorf("Oldest = %v, want the date the ledger reaches back to", after.Oldest)
	}
}

// Nothing infers, suggests or defaults an owner from the name the bank has on
// an account, even where it matches a member's name exactly. Matching them
// would silently widen who sees an account, which is the failure the ownership
// model exists to prevent.
func TestOwnershipIsNeverInferredFromTheHolderName(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	// The bank says the account is Grace's. Ada connected it.
	held := account("hash-1", "Joint", "0538")
	held.HolderName = grace.FirstName + " " + grace.LastName
	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, held)

	owners, err := db.AccountOwners(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("AccountOwners: %v", err)
	}
	if len(owners) != 1 || owners[0] != ada.ID {
		t.Errorf("owners = %v, want only the member who connected the bank", owners)
	}

	// And the holder name is returned exactly as the bank gave it, beside
	// whoever the household says owns the account.
	if err := db.SetAccountOwners(ctx, stored[0].ID, []string{ada.ID, grace.ID}); err != nil {
		t.Fatalf("setting owners: %v", err)
	}
	after, err := db.AccountByID(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("AccountByID: %v", err)
	}
	if after.HolderName != held.HolderName {
		t.Errorf("HolderName = %q, want it verbatim as the bank gave it: %q", after.HolderName, held.HolderName)
	}
	both, err := db.AccountOwners(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("AccountOwners: %v", err)
	}
	if len(both) != 2 {
		t.Errorf("a joint account reports %d owners, want 2", len(both))
	}
}

// Leaving an account out says it is not in wimm (ADR 0022), and a ledger still
// listing its rows would contradict that: it drops out of the list, the count
// and the span, none of it deleted, all of it back the moment it is un-left.
func TestLeavingAnAccountOutRemovesItFromTheLedgerAndBringingItBackRestoresIt(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	accountID := ledger(t, ctx, db, ada.ID,
		booked("ref-a", 1, -100), booked("ref-b", 5, -200), booked("ref-c", 9, -300))

	page := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(page.Transactions) != 3 {
		t.Fatalf("before leaving out: %d transactions, want 3", len(page.Transactions))
	}

	if err := db.SetAccountLeftOut(ctx, accountID, true); err != nil {
		t.Fatalf("SetAccountLeftOut: %v", err)
	}

	page = readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(page.Transactions) != 0 {
		t.Errorf("a left-out account's ledger holds %d transactions, want 0", len(page.Transactions))
	}
	if n, err := db.CountLedger(ctx, ada.ID, ""); err != nil || n != 0 {
		t.Errorf("CountLedger = %d, %v, want 0, nil", n, err)
	}
	state, err := db.LedgerState(ctx, ada.ID, "")
	if err != nil {
		t.Fatalf("LedgerState: %v", err)
	}
	if state.ReachesBack != nil {
		t.Errorf("ReachesBack = %v for a left-out account's ledger, want nil", state.ReachesBack)
	}

	if err := db.SetAccountLeftOut(ctx, accountID, false); err != nil {
		t.Fatalf("bringing it back: %v", err)
	}
	page = readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 50})
	if len(page.Transactions) != 3 {
		t.Errorf("after bringing it back: %d transactions, want 3 restored", len(page.Transactions))
	}
	if n, err := db.CountLedger(ctx, ada.ID, ""); err != nil || n != 3 {
		t.Errorf("CountLedger = %d, %v, want 3, nil", n, err)
	}
}

// Newest is the cursor being absent; oldest and a month are the same seek
// anchored elsewhere. A jump must land on a page contiguous with the ones
// either side of it, with nothing repeated and nothing skipped.
func TestOldestJumpIsContiguousWithThePageEitherSide(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	var txs []store.Transaction
	for i := 1; i <= 9; i++ {
		txs = append(txs, booked("ref-"+string(rune('a'+i-1)), i, int64(-100*i)))
	}
	ledger(t, ctx, db, ada.ID, txs...)

	oldest := readPage(t, ctx, db, store.LedgerQuery{MemberID: ada.ID, Limit: 3, Oldest: true})
	if oldest.HasOlder {
		t.Error("the oldest page offers something older")
	}
	if !oldest.HasNewer {
		t.Error("the oldest page offers nothing newer, but six transactions are")
	}
	if want := []string{"ref-c", "ref-b", "ref-a"}; !same(keys(oldest), want) {
		t.Fatalf("the oldest page holds %v, want %v", keys(oldest), want)
	}

	// Newer, from the oldest page's own newest row: must reach the segment
	// immediately above it with nothing repeated and nothing skipped.
	first := oldest.Transactions[0]
	newer := readPage(t, ctx, db, store.LedgerQuery{
		MemberID: ada.ID, Limit: 3, Cursor: store.Cursor{BookingDate: first.BookingDate, ID: first.ID}})
	if want := []string{"ref-f", "ref-e", "ref-d"}; !same(keys(newer), want) {
		t.Errorf("paging newer from the oldest page gave %v, want %v", keys(newer), want)
	}
}

// A page start is a cursor resolved on the server: it must land on precisely
// that page, contiguous with the page before and after it — the property a
// calendar month cannot promise once a page spans more of them than one.
func TestPageStartIsContiguousWithThePageEitherSide(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	on := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	txOn := func(key string, when time.Time, minor int64) store.Transaction {
		return store.Transaction{
			Status: store.StatusBooked, DedupKey: key, AmountMinor: minor,
			Currency: "EUR", BookingDate: when, CounterpartyName: "Padaria Ribeiro",
		}
	}

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Conta a Ordem", "0538"))
	txs := []store.Transaction{
		txOn("jan-a", on(time.January, 1), -100),
		txOn("jan-b", on(time.January, 5), -200),
		txOn("feb-a", on(time.February, 2), -300),
		txOn("feb-b", on(time.February, 10), -400),
		txOn("mar-a", on(time.March, 3), -500),
		txOn("mar-b", on(time.March, 7), -600),
	}
	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID, txs, on(time.March, 10)); err != nil {
		t.Fatalf("WriteAccountTransactions: %v", err)
	}

	index, err := db.LedgerPageIndex(ctx, ada.ID, "", 2)
	if err != nil {
		t.Fatalf("LedgerPageIndex: %v", err)
	}
	if len(index) != 3 {
		t.Fatalf("LedgerPageIndex = %+v, want 3 pages of 2", index)
	}
	middle := index[1] // mar-b, mar-a is index[0]; feb-b, feb-a is index[1].

	page := readPage(t, ctx, db, store.LedgerQuery{
		MemberID: ada.ID, Limit: 2, PageStart: &middle.Cursor})
	if want := []string{"feb-b", "feb-a"}; !same(keys(page), want) {
		t.Fatalf("the page named by the index holds %v, want %v", keys(page), want)
	}
	if !page.HasOlder {
		t.Error("this page offers nothing older, but January is")
	}
	if !page.HasNewer {
		t.Error("this page offers nothing newer, but March is")
	}

	last := page.Transactions[len(page.Transactions)-1]
	older := readPage(t, ctx, db, store.LedgerQuery{
		MemberID: ada.ID, Limit: 2, Older: true,
		Cursor: store.Cursor{BookingDate: last.BookingDate, ID: last.ID}})
	if want := []string{"jan-b", "jan-a"}; !same(keys(older), want) {
		t.Errorf("paging older from this page gave %v, want January: %v", keys(older), want)
	}

	first := page.Transactions[0]
	newer := readPage(t, ctx, db, store.LedgerQuery{
		MemberID: ada.ID, Limit: 2, Cursor: store.Cursor{BookingDate: first.BookingDate, ID: first.ID}})
	if want := []string{"mar-b", "mar-a"}; !same(keys(newer), want) {
		t.Errorf("paging newer from this page gave %v, want March: %v", keys(newer), want)
	}
}

// The page index buckets every transaction into a real page, oldest page
// last and never larger than the page size — a calendar aggregate could
// offer a range with a hundred rows in it, and none of them would be a page
// a member could actually land on.
func TestThePageIndexBucketsIntoRealPages(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	on := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Conta a Ordem", "0538"))
	txs := []store.Transaction{
		{Status: store.StatusBooked, DedupKey: "jan-a", AmountMinor: -100, Currency: "EUR", BookingDate: on(time.January, 1)},
		{Status: store.StatusBooked, DedupKey: "jan-b", AmountMinor: -200, Currency: "EUR", BookingDate: on(time.January, 20)},
		{Status: store.StatusBooked, DedupKey: "mar-a", AmountMinor: -300, Currency: "EUR", BookingDate: on(time.March, 3)},
	}
	if _, err := db.WriteAccountTransactions(ctx, stored[0].ID, txs, on(time.March, 10)); err != nil {
		t.Fatalf("WriteAccountTransactions: %v", err)
	}

	index, err := db.LedgerPageIndex(ctx, ada.ID, "", 2)
	if err != nil {
		t.Fatalf("LedgerPageIndex: %v", err)
	}
	if len(index) != 2 {
		t.Fatalf("LedgerPageIndex = %+v, want 2 pages of at most 2", index)
	}
	if !index[0].Newest.Equal(on(time.March, 3)) || !index[0].Oldest.Equal(on(time.January, 20)) {
		t.Errorf("newest page = %+v, want March 3 to January 20", index[0])
	}
	if !index[1].Newest.Equal(on(time.January, 1)) || !index[1].Oldest.Equal(on(time.January, 1)) {
		t.Errorf("oldest page = %+v, want January 1 alone", index[1])
	}
}

// The page index runs on the same index the ledger already pages through, so
// it is measured here rather than assumed: a scan over 50,000 rows must stay
// well clear of anything a member would notice.
func TestThePageIndexStaysFastAtFiftyThousandRows(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Conta a Ordem", "0538"))
	accountID := stored[0].ID

	const total = 50_000
	start := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)
	rows := make([][]any, total)
	for i := range total {
		rows[i] = []any{
			accountID, string(store.StatusBooked), fmt.Sprintf("bulk-%d", i), int64(-100 - i),
			"EUR", start.AddDate(0, 0, i%2000),
		}
	}
	if _, err := db.Pool().CopyFrom(ctx,
		pgx.Identifier{"transactions"},
		[]string{"account_id", "status", "dedup_key", "amount_minor", "currency", "booking_date"},
		pgx.CopyFromRows(rows),
	); err != nil {
		t.Fatalf("seeding 50,000 transactions: %v", err)
	}

	began := time.Now()
	index, err := db.LedgerPageIndex(ctx, ada.ID, "", 50)
	elapsed := time.Since(began)
	if err != nil {
		t.Fatalf("LedgerPageIndex: %v", err)
	}
	t.Logf("LedgerPageIndex over %d rows took %s", total, elapsed)
	if elapsed > 2*time.Second {
		t.Errorf("LedgerPageIndex took %s over %d rows, want well under 2s", elapsed, total)
	}
	if len(index) != total/50 {
		t.Errorf("LedgerPageIndex returned %d pages, want %d", len(index), total/50)
	}
}
