package banking_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/banking/bankingtest"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

func gwTx(ref string, day int, minor int64, status banking.TransactionStatus) banking.Transaction {
	return banking.Transaction{
		Ref: ref, Status: status,
		Amount:           banking.Money{Minor: minor, Currency: "EUR"},
		BookingDate:      time.Date(2026, time.September, day, 0, 0, 0, 0, time.UTC),
		CounterpartyName: "Padaria Ribeiro",
	}
}

// ledgerFor connects a bank holding history and returns the service ready to
// read it.
func ledgerFor(t *testing.T, txs ...banking.Transaction) (*banking.Service, *bankingtest.Gateway, *memStore) {
	t.Helper()
	svc, gw, st := newService(t)
	gw.AddBank(montepio(), banking.Account{Ref: "hash-1", Name: "Conta à Ordem", Currency: "EUR"})
	connectBank(t, svc, gw, st)
	gw.SetTransactions(montepio().ID, "hash-1", txs...)
	return svc, gw, st
}

// The first fill asks for the longest the bank offers, and every sync after it
// asks from the synced-through date less the overlap window.
func TestAFirstFillAsksForEverythingAndTheNextSyncAsksFromWhereItStopped(t *testing.T) {
	svc, gw, _ := ledgerFor(t,
		gwTx("ref-a", 10, -1190, banking.StatusBooked),
		gwTx("ref-b", 12, 240_000, banking.StatusBooked))

	ctx := context.Background()
	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(ledger.Page.Transactions) != 2 {
		t.Fatalf("the first fill stored %d transactions, want 2", len(ledger.Page.Transactions))
	}
	if synced := gw.AccountsSynced(); len(synced) != 1 {
		t.Errorf("the first fill made %d reads, want 1", len(synced))
	}

	// A second arrival inside the interval asks the bank nothing at all.
	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada}); err != nil {
		t.Fatalf("second arrival: %v", err)
	}
	if synced := gw.AccountsSynced(); len(synced) != 1 {
		t.Errorf("a second arrival inside the interval made %d reads, want none", len(synced)-1)
	}

	// Refresh is the member asking in as many words, and is not bound by it.
	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada, Refresh: true}); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if synced := gw.AccountsSynced(); len(synced) != 2 {
		t.Errorf("Refresh made %d reads in total, want 2", len(synced))
	}
}

// A transaction dated in none of the three fields cannot be filed anywhere in
// a ledger that orders and groups by one date, so the sync drops it rather
// than inventing today — and drops only it: the rest of what the bank
// returned is still stored. This is wimm's own fallback and refusal
// (identity.EffectiveDate, toStoreTransaction), exercised here through the
// generic sync path against bankingtest rather than any one gateway's client.
func TestATransactionWithNoDateAtAllIsDroppedNotStored(t *testing.T) {
	dateless := gwTx("ref-dateless", 1, -500, banking.StatusBooked)
	dateless.BookingDate = time.Time{}

	svc, _, _ := ledgerFor(t, dateless, gwTx("ref-dated", 10, -1190, banking.StatusBooked))

	ledger, err := svc.Transactions(context.Background(), banking.LedgerRequest{MemberID: ada})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(ledger.Page.Transactions) != 1 {
		t.Fatalf("the ledger holds %d transactions, want 1 (the dateless one dropped)", len(ledger.Page.Transactions))
	}
	if ledger.Page.Transactions[0].DedupKey != "ref-dated" {
		t.Errorf("the stored transaction is %q, want the dated one", ledger.Page.Transactions[0].DedupKey)
	}
}

// The fallback order is value date then transaction date, the same order
// identity.go's digest already uses — asserted here against the sync rather
// than against EffectiveDate alone, so a regression in either is caught by the
// path a member actually reads through.
func TestATransactionWithNoBookingDateFallsBackToTheValueDate(t *testing.T) {
	tx := gwTx("ref-fallback", 1, -500, banking.StatusBooked)
	tx.BookingDate = time.Time{}
	tx.ValueDate = time.Date(2026, time.September, 12, 0, 0, 0, 0, time.UTC)

	svc, _, _ := ledgerFor(t, tx)

	ledger, err := svc.Transactions(context.Background(), banking.LedgerRequest{MemberID: ada})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(ledger.Page.Transactions) != 1 {
		t.Fatalf("the ledger holds %d transactions, want 1", len(ledger.Page.Transactions))
	}
	if got := ledger.Page.Transactions[0].BookingDate; !got.Equal(tx.ValueDate) {
		t.Errorf("BookingDate = %s, want the value date %s", got, tx.ValueDate)
	}
}

// The overlap re-reads a few days and the identity rule discards what is
// already held, so nothing is stored twice.
func TestTheOverlapReReadsWithoutDuplicating(t *testing.T) {
	svc, gw, _ := ledgerFor(t,
		gwTx("ref-a", 10, -1190, banking.StatusBooked),
		gwTx("ref-b", 12, 240_000, banking.StatusBooked))

	ctx := context.Background()
	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada}); err != nil {
		t.Fatalf("first fill: %v", err)
	}

	// The bank returns the same two again, plus one that is new.
	gw.SetTransactions(montepio().ID, "hash-1",
		gwTx("ref-a", 10, -1190, banking.StatusBooked),
		gwTx("ref-b", 12, 240_000, banking.StatusBooked),
		gwTx("ref-c", 13, -450, banking.StatusBooked))

	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada, Refresh: true})
	if err != nil {
		t.Fatalf("incremental sync: %v", err)
	}
	if len(ledger.Page.Transactions) != 3 {
		t.Errorf("after an overlapping sync the ledger holds %d transactions, want 3", len(ledger.Page.Transactions))
	}
	if ledger.Count != 3 {
		t.Errorf("Count = %d, want 3", ledger.Count)
	}
}

// A first fill at a bank that keeps years of history is unbounded, so it stops
// at the page cap — and records how far it actually reached, so the next sync
// continues rather than restarting.
func TestAFillStopsAtThePageCap(t *testing.T) {
	var history []banking.Transaction
	for i := 1; i <= 20; i++ {
		history = append(history, gwTx("ref-"+string(rune('a'+i-1)), i, int64(-100*i), banking.StatusBooked))
	}

	svc, gw, _ := ledgerFor(t, history...)
	gw.PageSize = 2

	ctx := context.Background()
	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}

	// Three pages of two, and no more: testLedgerOptions caps it at three.
	if ledger.Count != 6 {
		t.Errorf("a capped fill stored %d transactions, want 3 pages of 2", ledger.Count)
	}
	if ledger.ReachesBack == nil {
		t.Fatal("a fill that stored transactions reports no date it reaches back to")
	}
}

// No sync runs without a member. There is no ticker and no path into the sync
// that is not an arrival or a Refresh, because PSU-present headers are what
// exempt these reads from the background-fetch cap and a header cannot honestly
// be set on a call nobody asked for.
func TestNoSyncPathRunsWithoutAMember(t *testing.T) {
	svc, gw, _ := ledgerFor(t, gwTx("ref-a", 10, -1190, banking.StatusBooked))
	ctx := context.Background()

	// Rendering a page the member paged to asks no bank anything: a sync would
	// insert at the newest end while they are reading somewhere else.
	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada, SkipSync: true}); err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if synced := gw.AccountsSynced(); len(synced) != 0 {
		t.Errorf("reading a stored page made %d bank reads, want none", len(synced))
	}

	// And every sync that does run is one a member asked for: both entry points
	// take the member whose accounts are being read.
	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada}); err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if synced := gw.AccountsSynced(); len(synced) != 1 {
		t.Errorf("an arrival made %d reads, want 1", len(synced))
	}
}

// Every failure leaves the stored rows and their time exactly as they were,
// which is the one thing that must not be lost.
func TestEveryFailureLeavesTheStoredLedgerUntouched(t *testing.T) {
	for _, want := range []error{
		banking.ErrBankUnavailable,
		banking.ErrGatewayUnavailable,
		banking.ErrConsentExpired,
		banking.RateLimited(6 * time.Hour),
	} {
		t.Run(want.Error(), func(t *testing.T) {
			svc, gw, _ := ledgerFor(t, gwTx("ref-a", 10, -1190, banking.StatusBooked))
			ctx := context.Background()

			before, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada})
			if err != nil {
				t.Fatalf("first fill: %v", err)
			}
			if before.SyncedAt == nil {
				t.Fatal("the first fill recorded no sync time")
			}
			syncedAt := *before.SyncedAt

			// The bank now holds more, and refuses to hand any of it over.
			gw.SetTransactions(montepio().ID, "hash-1",
				gwTx("ref-a", 10, -1190, banking.StatusBooked),
				gwTx("ref-b", 12, -450, banking.StatusBooked))
			gw.Fail("Transactions", want)

			after, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada, Refresh: true})
			if err != nil {
				t.Fatalf("Transactions: %v", err)
			}

			if after.Count != before.Count {
				t.Errorf("a %v left %d transactions, want the %d that were stored", want, after.Count, before.Count)
			}
			if after.SyncedAt == nil || !after.SyncedAt.Equal(syncedAt) {
				t.Errorf("SyncedAt moved to %v on a failure, want it left at %s", after.SyncedAt, syncedAt)
			}
			if len(after.Failures) != 1 {
				t.Fatalf("the member was told about %d banks, want 1", len(after.Failures))
			}
			if after.Failures[0].BankName != montepio().Name {
				t.Errorf("the failure names %q, want the bank", after.Failures[0].BankName)
			}
		})
	}
}

// A refusal because wimm asked too often carries when it can next be tried, and
// the transactions on screen stay where they are.
func TestARateLimitedSyncCarriesItsRetryAfter(t *testing.T) {
	svc, gw, _ := ledgerFor(t, gwTx("ref-a", 10, -1190, banking.StatusBooked))
	ctx := context.Background()

	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada}); err != nil {
		t.Fatalf("first fill: %v", err)
	}
	gw.Fail("Transactions", banking.RateLimited(6*time.Hour))

	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada, Refresh: true})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(ledger.Failures) != 1 || ledger.Failures[0].RetryAfter != 6*time.Hour {
		t.Errorf("failures = %+v, want one carrying a 6h retry-after", ledger.Failures)
	}
	if ledger.Count != 1 {
		t.Errorf("a rate-limited refresh left %d transactions, want the 1 already held", ledger.Count)
	}
}

// Access running out does not reset the ledger, and restoring carries on from
// where it stopped — which at a bank granting a single day is the difference
// between a ledger and a screen that empties every night.
func TestAccessRunningOutKeepsTheLedgerAndRestoringCarriesOn(t *testing.T) {
	svc, gw, _ := ledgerFor(t, gwTx("ref-a", 10, -1190, banking.StatusBooked))
	ctx := context.Background()

	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada}); err != nil {
		t.Fatalf("first fill: %v", err)
	}

	gw.Fail("Transactions", banking.ErrConsentExpired)
	out, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada, Refresh: true})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if out.Count != 1 {
		t.Errorf("a bank whose access ran out left %d transactions, want the 1 already read", out.Count)
	}
	if len(out.Failures) != 1 || !errors.Is(out.Failures[0].Err, banking.ErrConsentExpired) {
		t.Errorf("failures = %+v, want the member told which bank is no longer being read", out.Failures)
	}
}

// A member granted a level on an account they do not own reads no transactions
// and no count, and is told they see transactions for accounts that are theirs
// rather than shown an empty list.
func TestAMemberWhoOwnsNoAccountIsToldWhy(t *testing.T) {
	svc, _, _ := ledgerFor(t, gwTx("ref-a", 10, -1190, banking.StatusBooked))
	ctx := context.Background()

	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada}); err != nil {
		t.Fatalf("first fill: %v", err)
	}

	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: grace})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if !ledger.OwnsNothing {
		t.Error("a member who owns no account is not told so")
	}
	if ledger.Count != 0 || len(ledger.Page.Transactions) != 0 {
		t.Errorf("a member who owns nothing read %d transactions and a count of %d", len(ledger.Page.Transactions), ledger.Count)
	}
}

// A bank connected before wimm could read transactions is never synced, is not
// reported as a failure, and is offered as something to widen.
func TestANarrowConnectionIsNotSyncedAndIsOfferedToBeWidened(t *testing.T) {
	svc, gw, st := ledgerFor(t, gwTx("ref-a", 10, -1190, banking.StatusBooked))
	ctx := context.Background()

	// The connection as it would read if it had been made before this change.
	connections, err := st.BankConnections(ctx)
	if err != nil {
		t.Fatalf("BankConnections: %v", err)
	}
	if err := st.SetConnectionScope(ctx, connections[0].ID, "balances"); err != nil {
		t.Fatalf("SetConnectionScope: %v", err)
	}

	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}

	if synced := gw.AccountsSynced(); len(synced) != 0 {
		t.Errorf("a narrow connection was read %d times, want none: the bank was never asked", len(synced))
	}
	if len(ledger.Failures) != 0 {
		t.Errorf("a narrow connection was reported as a failure: %+v", ledger.Failures)
	}
	if len(ledger.Narrow) != 1 || ledger.Narrow[0].BankName != montepio().Name {
		t.Errorf("narrow = %+v, want the one bank that can be widened", ledger.Narrow)
	}
}

// Connecting after this change asks for both, so a household connecting its
// first bank never sees the narrow state.
func TestANewConnectionIsWide(t *testing.T) {
	svc, _, st := ledgerFor(t)
	ctx := context.Background()

	connections, err := st.BankConnections(ctx)
	if err != nil {
		t.Fatalf("BankConnections: %v", err)
	}
	scope, err := st.BankConnectionScope(ctx, connections[0].ID)
	if err != nil {
		t.Fatalf("BankConnectionScope: %v", err)
	}
	if !scope.ReadsTransactions() {
		t.Errorf("a connection made after this change reads as %q, want the wide scope", scope)
	}

	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(ledger.Narrow) != 0 {
		t.Errorf("a member who has only just connected is offered %d banks to widen, want none", len(ledger.Narrow))
	}
}

// Widening is confirming once more at the bank, without disconnecting first and
// without choosing the bank from the list again. Every account keeps the owners
// and the levels it had.
func TestWideningKeepsOwnersAndLevelsAndChangesTheScope(t *testing.T) {
	svc, gw, st := ledgerFor(t, gwTx("ref-a", 10, -1190, banking.StatusBooked))
	ctx := context.Background()

	connections, err := st.BankConnections(ctx)
	if err != nil {
		t.Fatalf("BankConnections: %v", err)
	}
	connectionID := connections[0].ID
	if err := st.SetConnectionScope(ctx, connectionID, "balances"); err != nil {
		t.Fatalf("SetConnectionScope: %v", err)
	}

	accounts, err := st.AccountsForConnection(ctx, connectionID)
	if err != nil {
		t.Fatalf("AccountsForConnection: %v", err)
	}
	if err := st.SetAccountLevel(ctx, accounts[0].ID, grace, "balance", ada); err != nil {
		t.Fatalf("granting a level: %v", err)
	}

	// The member widens: back to the bank, no picker, no disconnection.
	if _, err := svc.RestoreConnection(ctx, ada, connectionID); err != nil {
		t.Fatalf("RestoreConnection: %v", err)
	}
	if _, err := svc.CompleteConnection(ctx, ada, banking.Callback{Code: "code", State: gw.LastState}); err != nil {
		t.Fatalf("CompleteConnection: %v", err)
	}

	scope, err := st.BankConnectionScope(ctx, connectionID)
	if err != nil {
		t.Fatalf("BankConnectionScope: %v", err)
	}
	if !scope.ReadsTransactions() {
		t.Errorf("after widening the connection reads as %q, want the wide scope", scope)
	}

	owners, err := st.AccountOwners(ctx, accounts[0].ID)
	if err != nil {
		t.Fatalf("AccountOwners: %v", err)
	}
	if len(owners) != 1 || owners[0] != ada {
		t.Errorf("owners = %v, want them unchanged by widening", owners)
	}
	grants, err := st.GrantsForConnection(ctx, connectionID)
	if err != nil {
		t.Fatalf("GrantsForConnection: %v", err)
	}
	if len(grants) != 1 || grants[0].MemberID != grace {
		t.Errorf("grants = %+v, want them unchanged by widening", grants)
	}

	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(ledger.Narrow) != 0 {
		t.Errorf("a widened bank is still offered to be widened: %+v", ledger.Narrow)
	}
	if ledger.Count != 1 {
		t.Errorf("after widening the ledger holds %d transactions, want the bank's 1", ledger.Count)
	}
}

// Under a search the count, the pages offered and the rows on the page all
// describe the filtered list, and the accounts and months offered come with it.
func TestTheCountThePagesAndTheRowsAgreeUnderASearch(t *testing.T) {
	var txs []banking.Transaction
	for i := range 70 {
		tx := gwTx(fmt.Sprintf("ref-%02d", i), 1+i%28, -1000, banking.StatusBooked)
		if i%7 != 0 {
			tx.CounterpartyName = "GALP ENERGIA"
		}
		txs = append(txs, tx)
	}
	svc, _, _ := ledgerFor(t, txs...)
	ctx := context.Background()
	if _, err := svc.Transactions(ctx, banking.LedgerRequest{MemberID: ada}); err != nil {
		t.Fatalf("first arrival: %v", err)
	}

	ledger, err := svc.Transactions(ctx, banking.LedgerRequest{
		MemberID: ada, SkipSync: true, LedgerFilter: store.LedgerFilter{Search: "galp"}})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if ledger.Count != 60 {
		t.Errorf("Count = %d, want the 60 that match", ledger.Count)
	}
	if len(ledger.Pages) != 2 {
		t.Errorf("%d pages offered, want 2 pages of the 60 that match", len(ledger.Pages))
	}
	if len(ledger.Page.Transactions) != 50 {
		t.Errorf("the page holds %d rows, want 50", len(ledger.Page.Transactions))
	}
	for _, tx := range ledger.Page.Transactions {
		if tx.CounterpartyName != "GALP ENERGIA" {
			t.Errorf("a search for galp listed %q", tx.CounterpartyName)
		}
	}
	if len(ledger.FilterAccounts) != 1 || ledger.FilterAccounts[0].Name != "Conta à Ordem" ||
		ledger.FilterAccounts[0].BankName == "" {
		t.Errorf("FilterAccounts = %+v, want the one account, named with its bank", ledger.FilterAccounts)
	}
	if want := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC); len(ledger.Months) != 1 || !ledger.Months[0].Equal(want) {
		t.Errorf("Months = %v, want [%v]", ledger.Months, want)
	}
}

// totalsLedger is ada's two accounts in August: fuel, a salary, €500.00 moved
// from one to the other, a payment not yet settled and one in dollars.
func totalsLedger(t *testing.T) (*banking.Service, string, string) {
	t.Helper()
	svc, st, c := twoAccounts(t)
	joint, personal := c.Accounts[0].ID, c.Accounts[1].ID
	aug := func(d int) time.Time { return time.Date(2026, time.August, d, 0, 0, 0, 0, time.UTC) }
	seedRow(st, "fuel", joint, -4500, aug(3), "Galp")
	seedRow(st, "to-personal", joint, -50000, aug(10), "Transfer to Personal")
	seedRow(st, "from-joint", personal, 50000, aug(10), "Transfer from Joint")
	seedRow(st, "salary", joint, 250000, aug(25), "Empresa")
	st.transactions[joint] = append(st.transactions[joint],
		store.Transaction{ID: "pending", AccountID: joint, Status: store.StatusPending,
			AmountMinor: -1200, Currency: "EUR", BookingDate: aug(28), CounterpartyName: "Galp"},
		store.Transaction{ID: "dollars", AccountID: joint, Status: store.StatusBooked,
			AmountMinor: -1000, Currency: "USD", BookingDate: aug(12), CounterpartyName: "Galp US"})
	return svc, joint, personal
}

func totalsOf(t *testing.T, svc *banking.Service, f store.LedgerFilter) banking.Ledger {
	t.Helper()
	l, err := svc.Transactions(context.Background(), banking.LedgerRequest{MemberID: ada, SkipSync: true, LedgerFilter: f})
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	return l
}

// What a filtered list adds up to is settled money only, per currency, with a
// transfer between two of the member's accounts in scope left out — the same
// figure Overview gives the month.
func TestAFilteredListAddsUpItsSettledRowsLeavingOutTransfers(t *testing.T) {
	svc, joint, _ := totalsLedger(t)
	august := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name       string
		f          store.LedgerFilter
		want       []banking.FilterTotal
		transfers  int
		notSettled int
	}{
		{"a month, both accounts in scope", store.LedgerFilter{Month: august},
			[]banking.FilterTotal{{Currency: "EUR", In: 250000, Out: 4500}, {Currency: "USD", Out: 1000}}, 1, 1},
		{"one account: the transfer left it", store.LedgerFilter{Month: august, AccountID: joint},
			[]banking.FilterTotal{{Currency: "EUR", In: 250000, Out: 54500}, {Currency: "USD", Out: 1000}}, 0, 1},
		{"money in: its partner is still the member's", store.LedgerFilter{Direction: store.MoneyIn},
			[]banking.FilterTotal{{Currency: "EUR", In: 250000}}, 1, 0},
		{"a search", store.LedgerFilter{Search: "galp"},
			[]banking.FilterTotal{{Currency: "EUR", Out: 4500}, {Currency: "USD", Out: 1000}}, 0, 1},
		{"nothing matches", store.LedgerFilter{Search: "plumber"}, nil, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := totalsOf(t, svc, tc.f)
			if !slices.Equal(l.Totals, tc.want) {
				t.Errorf("Totals = %+v, want %+v", l.Totals, tc.want)
			}
			if l.TransfersLeftOut != tc.transfers {
				t.Errorf("TransfersLeftOut = %d, want %d", l.TransfersLeftOut, tc.transfers)
			}
			if l.NotSettled != tc.notSettled {
				t.Errorf("NotSettled = %d, want %d", l.NotSettled, tc.notSettled)
			}
		})
	}
}

// With no filter in force there is nothing to sum: the whole ledger is not a
// period anybody asked about.
func TestTheWholeLedgerHasNoTotals(t *testing.T) {
	svc, _, _ := totalsLedger(t)
	if l := totalsOf(t, svc, store.LedgerFilter{}); l.Totals != nil {
		t.Errorf("Totals = %+v with no filter, want none", l.Totals)
	}
}
