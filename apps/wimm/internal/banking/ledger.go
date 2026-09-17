package banking

import (
	"context"
	"errors"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// Ledger is what a member sees on Transactions: the page they are looking at,
// how current it is, and every reason a bank is not contributing to it.
type Ledger struct {
	Page store.LedgerPage
	// Count is how many transactions this member may see. A count of what they
	// may see and not a page count, and it survives paging unchanged.
	Count int
	// SyncedAt is the most recent time any of their accounts was brought up to
	// date, which is what the list header states. Nil means nothing has ever
	// been read.
	SyncedAt *time.Time
	// ReachesBack is the oldest transaction wimm holds for them. wimm states
	// the date the list actually reaches rather than a period, because no bank
	// says in advance how much history it will hand over.
	ReachesBack *time.Time
	// Narrow are the banks connected before wimm could read transactions, each
	// offering to be widened. They are an inline alert on that bank, never a
	// page banner.
	Narrow []store.NarrowConnection
	// Failures are the banks that did not answer this time. The transactions
	// already on screen stay with their original time.
	Failures []BankFailure
	// OwnsNothing distinguishes a member who owns no account at all from one
	// whose accounts are simply empty. The reasons differ and must not be
	// collapsed into one empty list.
	OwnsNothing bool
	// Accounts names every account this member owns, so a row can say which
	// one it came from without a second call per row.
	Accounts map[string]store.AccountLabel
}

// LedgerRequest is what a member asked for.
type LedgerRequest struct {
	MemberID  string
	AccountID string
	Cursor    store.Cursor
	Older     bool
	// Refresh asks the banks again rather than waiting for the interval. It is
	// still bounded by it: a member holding the button does not multiply the
	// calls.
	Refresh bool
	// SkipSync renders what is stored without asking any bank. The screen uses
	// it for a page the member paged to, where a sync would insert at the
	// newest end while they are reading somewhere else.
	SkipSync bool
}

// Transactions returns one page of the ledger, having first brought the
// member's accounts up to date.
//
// The sync runs before the read and never blocks it for long: it is bounded by
// a page cap per account and by a minimum interval between syncs of one
// account. What it cannot do is run without a member — there is no ticker and
// no path into it that is not an arrival or a Refresh, which is what keeps the
// background-fetch cap out of reach (ADR 0021).
func (s *Service) Transactions(ctx context.Context, req LedgerRequest) (Ledger, error) {
	var failures []BankFailure
	if !req.SkipSync {
		failures = s.SyncTransactions(ctx, req.MemberID, req.Refresh)
	}

	page, err := s.store.Ledger(ctx, store.LedgerQuery{
		MemberID: req.MemberID, AccountID: req.AccountID,
		Cursor: req.Cursor, Older: req.Older, Limit: s.pageSize,
	})
	if err != nil {
		return Ledger{}, err
	}

	count, err := s.store.CountLedger(ctx, req.MemberID, req.AccountID)
	if err != nil {
		return Ledger{}, err
	}

	narrow, err := s.store.NarrowConnections(ctx, req.MemberID)
	if err != nil {
		return Ledger{}, err
	}

	state, err := s.store.LedgerState(ctx, req.MemberID, req.AccountID)
	if err != nil {
		return Ledger{}, err
	}

	labels, err := s.store.OwnedAccountLabels(ctx, req.MemberID)
	if err != nil {
		return Ledger{}, err
	}

	return Ledger{
		Accounts: labels,
		Page:     page, Count: count, Narrow: narrow, Failures: failures,
		SyncedAt:    state.SyncedAt,
		ReachesBack: state.ReachesBack,
		// A member who owns no account is told that they see transactions for
		// accounts that are theirs, rather than being shown an empty list: the
		// reasons for having nothing differ and must not be collapsed into one.
		OwnsNothing: state.OwnedAccounts == 0,
	}, nil
}

// SyncTransactions brings every account a member owns up to date, and returns
// the banks that did not answer.
//
// Every call is an arrival or a Refresh. There is no ticker: PSU-present
// headers are what exempt these reads from the roughly four-a-day cap, and a
// header cannot honestly be set on a call nobody asked for.
func (s *Service) SyncTransactions(ctx context.Context, memberID string, refresh bool) []BankFailure {
	accounts, err := s.store.SyncableAccounts(ctx, memberID)
	if err != nil {
		s.log.ErrorContext(ctx, "listing accounts to sync", "error", err)
		return nil
	}

	now, err := s.store.Now(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "reading database time for a sync", "error", err)
		return nil
	}

	// One open per connection rather than one per account: the sealed handle is
	// the same value for every account behind one bank.
	opened := map[string]Connection{}
	failed := map[string]BankFailure{}
	var order []string

	for _, a := range accounts {
		if !a.Scope.ReadsTransactions() {
			// The bank was never asked. This is not a failure and is not
			// reported as one: it is the narrow-scope state, which the screen
			// shows on that bank with the way to widen it.
			continue
		}
		if _, alreadyFailed := failed[a.Connection]; alreadyFailed {
			continue
		}
		if !s.dueForSync(a, now.T, refresh) {
			continue
		}

		live, ok := opened[a.Connection]
		if !ok {
			conn, err := s.store.BankConnectionByID(ctx, a.Connection)
			if err != nil {
				s.log.ErrorContext(ctx, "reading a connection to sync", "error", err)
				continue
			}
			gatewayRef, err := s.open(conn.GatewayRef, conn.ID)
			if err != nil {
				// A connection whose secret cannot be opened cannot be read,
				// and a member experiences that exactly as access having run
				// out.
				s.recordSyncFailure(ctx, failed, &order, a, ErrConsentExpired)
				continue
			}
			live = Connection{GatewayRef: gatewayRef, ExpiresAt: conn.ConsentExpiresAt}
			opened[a.Connection] = live
		}

		if err := s.syncAccount(ctx, live, a, now.T); err != nil {
			s.recordSyncFailure(ctx, failed, &order, a, err)
		}
	}

	out := make([]BankFailure, 0, len(order))
	for _, id := range order {
		out = append(out, failed[id])
	}
	return out
}

// dueForSync decides whether to ask this account's bank again.
//
// An arrival is bounded by the per-account interval, so a member with five
// accounts reloading the page repeatedly does not multiply the calls. Refresh
// is not: it is the member asking in as many words, and a control that silently
// does nothing is worse than one that costs a request.
//
// An account nothing has ever been read from is always due — that is the first
// fill, and it happens once per account rather than once per consent.
func (s *Service) dueForSync(a store.SyncableAccount, now time.Time, refresh bool) bool {
	if a.SyncedAt == nil || refresh {
		return true
	}
	return now.Sub(*a.SyncedAt) >= s.syncInterval
}

// recordSyncFailure keeps one failure per bank, in the order they happened. A
// member is told which bank did not answer, not which account.
func (s *Service) recordSyncFailure(
	ctx context.Context, failed map[string]BankFailure, order *[]string,
	a store.SyncableAccount, err error,
) {
	if _, seen := failed[a.Connection]; seen {
		return
	}

	var limited *RateLimitError
	retry := time.Duration(0)
	if errors.As(err, &limited) {
		retry = limited.RetryAfter
	}

	// Access having run out is recorded, not just reported: the next arrival
	// must offer to restore rather than try again and fail.
	if errors.Is(err, ErrConsentExpired) {
		if markErr := s.store.MarkBankConnectionExpired(ctx, a.Connection); markErr != nil {
			s.log.ErrorContext(ctx, "recording an expired connection", "error", markErr)
		}
	}

	failed[a.Connection] = BankFailure{
		ConnectionID: a.Connection, BankName: a.BankName, Err: err, RetryAfter: retry,
	}
	*order = append(*order, a.Connection)
}

// syncAccount reads one account's transactions and stores them.
//
// A first fill asks for the longest the bank offers; every sync after it asks
// from the synced-through date less the overlap window, because a bank can book
// a transaction with a booking date earlier than the day wimm last synced.
// Re-reading a few days and letting the identity rule discard what is already
// held is cheaper and more correct than trusting a watermark.
//
// Nothing is written until the whole read finishes. Every failure therefore
// leaves the stored rows and their sync time exactly as they were, which is the
// one thing that must not be lost.
func (s *Service) syncAccount(
	ctx context.Context, live Connection, a store.SyncableAccount, now time.Time,
) error {
	uid, err := s.open(a.GatewayUID, a.ID)
	if err != nil {
		return ErrConsentExpired
	}
	account := Account{Ref: a.GatewayRef, GatewayUID: uid, Currency: a.Currency}

	var from time.Time
	if a.SyncedThrough != nil {
		from = a.SyncedThrough.Add(-s.overlap)
	}

	var collected []store.Transaction
	cursor := ""
	exhausted := false
	for page := 0; page < s.maxPages; page++ {
		read, err := s.gateway.Transactions(ctx, live, account,
			TransactionsRequest{From: from, Cursor: cursor})
		if err != nil {
			return err
		}
		for _, t := range read.Transactions {
			row, ok := toStoreTransaction(t)
			if !ok {
				continue
			}
			collected = append(collected, row)
		}
		if read.NextCursor == "" {
			exhausted = true
			break
		}
		cursor = read.NextCursor
	}

	// How far back the fill actually reached, which is a stated fact rather
	// than a promise: a capped fill records where it stopped, and the next sync
	// continues from there instead of starting the history again.
	syncedThrough := syncedThroughFor(collected, from, now, exhausted)

	if _, err := s.store.WriteAccountTransactions(ctx, a.ID, collected, syncedThrough); err != nil {
		return err
	}
	return nil
}

// syncedThroughFor is the date an account's transactions are believed complete
// through.
//
// A sync that reached the end of what the bank offers is complete up to today.
// One that stopped at the page cap is complete only as far as its oldest row,
// because everything older is still unread.
func syncedThroughFor(rows []store.Transaction, from, now time.Time, exhausted bool) time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if exhausted {
		return today
	}
	if len(rows) == 0 {
		// Nothing came back and there is more to come: the watermark does not
		// move, so the next sync asks the same question rather than skipping
		// forward over transactions it never saw.
		if from.IsZero() {
			return time.Time{}
		}
		return from
	}
	oldest := rows[0].BookingDate
	for _, r := range rows {
		if r.BookingDate.Before(oldest) {
			oldest = r.BookingDate
		}
	}
	return oldest
}

// toStoreTransaction converts what the gateway returned into a row, computing
// the identity the store writes against.
//
// A transaction the bank dated in none of its three date fields is dropped: the
// ledger orders and groups by one date, and inventing today would file it on
// the wrong day.
func toStoreTransaction(t Transaction) (store.Transaction, bool) {
	date, ok := EffectiveDate(t)
	if !ok {
		return store.Transaction{}, false
	}

	row := store.Transaction{
		Status:           store.TransactionStatus(t.Status),
		DedupKey:         DedupKey(t),
		AmountMinor:      t.Amount.Minor,
		Currency:         t.Amount.Currency,
		BookingDate:      date,
		CounterpartyName: t.CounterpartyName,
		Remittance:       t.Remittance,
	}
	if !t.ValueDate.IsZero() {
		v := t.ValueDate
		row.ValueDate = &v
	}
	if !t.TransactionDate.IsZero() {
		d := t.TransactionDate
		row.TransactionDate = &d
	}
	return row, true
}
