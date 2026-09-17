package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// TransactionStatus is whether the bank has settled a transaction. It mirrors
// the check constraint on the column and nothing else is accepted.
type TransactionStatus string

const (
	StatusBooked  TransactionStatus = "booked"
	StatusPending TransactionStatus = "pending"
)

// Transaction is one entry on one account, as the ledger holds it.
type Transaction struct {
	ID        string
	AccountID string
	Status    TransactionStatus
	// DedupKey is the bank's entry reference or a digest standing in for one.
	// The domain computes it; the store never derives identity itself.
	DedupKey string
	// Occurrence tells apart rows that share a key, so the same payment made
	// twice is two transactions.
	Occurrence int
	// AmountMinor is signed: a debit is negative, so direction never depends on
	// a separate column agreeing with the sign.
	AmountMinor int64
	Currency    string
	// BookingDate is the day the ledger orders and groups by. A transaction
	// with no date at all is never stored, so this is never null in practice.
	BookingDate     time.Time
	ValueDate       *time.Time
	TransactionDate *time.Time
	// CounterpartyName is who it was with, where the bank named them.
	CounterpartyName string
	Remittance       string
	FirstSeenAt      time.Time
	LastSeenAt       time.Time
}

// Cursor is a position in the ledger. It is the sort key itself rather than an
// offset, because a sync inserts at the newest end while a member is reading:
// with an offset the rows under them shift and they silently re-read some and
// skip others (ADR 0021).
type Cursor struct {
	BookingDate time.Time
	ID          string
}

// Zero reports a cursor that names no position, which means the newest page.
func (c Cursor) Zero() bool { return c.ID == "" }

// LedgerPage is one read of the ledger.
type LedgerPage struct {
	Transactions []Transaction
	// Oldest and Newest are the span of dates on this page. A member is told
	// where in time they are, never which page they are on: a seek knows no
	// page number, and the span is what a person scanning backwards wants.
	Oldest time.Time
	Newest time.Time
	// HasOlder and HasNewer say which controls to offer. On the oldest page the
	// screen says there is nothing older rather than greying a control.
	HasOlder bool
	HasNewer bool
}

// LedgerQuery is what a member asked to see.
type LedgerQuery struct {
	// MemberID scopes the read. Only accounts this member owns are ever read.
	MemberID string
	// AccountID narrows to one account, and is empty for every account they
	// own. It is the same list filtered, not a different screen.
	AccountID string
	// Cursor is where to read from; the zero cursor is the newest page.
	Cursor Cursor
	// Older reads away from today. False with a set cursor reads back towards
	// it.
	Older bool
	// Limit is the page size.
	Limit int
}

// ownedAccounts is the scope of every ledger read, and the only one.
//
// It joins account_owners and nothing else. An account's transactions are
// visible to its owners and to nobody else: a member granted balance or details
// on an account they do not own sees none of it and is not told how many there
// are, because seeing a balance and seeing what was spent are different
// sentences and only the first was agreed to (ADR 0021).
//
// It is a fragment rather than a repeated predicate so that widening it is one
// edit in one place, which is what the fourth level will be.
const ownedAccounts = `
	select a.id from accounts a
	join account_owners o on o.account_id = a.id and o.member_id = $1
	where ($2::uuid is null or a.id = $2)`

// Ledger reads one page of the transactions a member may see.
//
// The seek is on (booking_date, id), which is the index the table is ordered
// by. Reading newer-ward asks for the same key the other way round and reverses
// the result, so both directions return newest first and the page a member sees
// is the same page whichever control brought them to it.
func (db *DB) Ledger(ctx context.Context, q LedgerQuery) (LedgerPage, error) {
	if q.Limit <= 0 {
		return LedgerPage{}, fmt.Errorf("a ledger page of %d transactions", q.Limit)
	}

	// One row more than asked for, which is how "is there another page that
	// way" is answered without a second count — and a count would be wrong by
	// the time it rendered anyway.
	probe := q.Limit + 1

	var query string
	args := []any{q.MemberID, nullString(q.AccountID)}

	switch {
	case q.Cursor.Zero():
		query = `
			select id, account_id, status, dedup_key, occurrence, amount_minor, currency,
			       booking_date, value_date, transaction_date,
			       coalesce(counterparty_name, ''), coalesce(remittance, ''),
			       first_seen_at, last_seen_at
			from transactions
			where account_id in (` + ownedAccounts + `)
			order by booking_date desc, id desc
			limit $3`
		args = append(args, probe)
	case q.Older:
		query = `
			select id, account_id, status, dedup_key, occurrence, amount_minor, currency,
			       booking_date, value_date, transaction_date,
			       coalesce(counterparty_name, ''), coalesce(remittance, ''),
			       first_seen_at, last_seen_at
			from transactions
			where account_id in (` + ownedAccounts + `)
			  and (booking_date, id) < ($4::date, $5::uuid)
			order by booking_date desc, id desc
			limit $3`
		args = append(args, probe, q.Cursor.BookingDate, q.Cursor.ID)
	default:
		// Ascending, then reversed below: "the 50 nearest rows on the newer
		// side" is a different question from "the 50 newest rows", and only the
		// first one comes back towards the member through the same
		// transactions in the same order.
		query = `
			select id, account_id, status, dedup_key, occurrence, amount_minor, currency,
			       booking_date, value_date, transaction_date,
			       coalesce(counterparty_name, ''), coalesce(remittance, ''),
			       first_seen_at, last_seen_at
			from transactions
			where account_id in (` + ownedAccounts + `)
			  and (booking_date, id) > ($4::date, $5::uuid)
			order by booking_date asc, id asc
			limit $3`
		args = append(args, probe, q.Cursor.BookingDate, q.Cursor.ID)
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return LedgerPage{}, fmt.Errorf("reading the ledger: %w", err)
	}
	defer rows.Close()

	out := make([]Transaction, 0, probe)
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return LedgerPage{}, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return LedgerPage{}, fmt.Errorf("reading the ledger: %w", err)
	}

	readingNewer := !q.Cursor.Zero() && !q.Older
	more := len(out) > q.Limit
	if more {
		out = out[:q.Limit]
	}
	if readingNewer {
		reverse(out)
	}

	page := LedgerPage{Transactions: out}
	if len(out) > 0 {
		page.Newest = out[0].BookingDate
		page.Oldest = out[len(out)-1].BookingDate
	}
	switch {
	case q.Cursor.Zero():
		page.HasOlder = more
	case q.Older:
		page.HasOlder = more
		page.HasNewer = true
	default:
		page.HasNewer = more
		page.HasOlder = true
	}
	return page, nil
}

func reverse(t []Transaction) {
	for i, j := 0, len(t)-1; i < j; i, j = i+1, j-1 {
		t[i], t[j] = t[j], t[i]
	}
}

func scanTransaction(rows pgx.Rows) (Transaction, error) {
	var t Transaction
	if err := rows.Scan(&t.ID, &t.AccountID, &t.Status, &t.DedupKey, &t.Occurrence,
		&t.AmountMinor, &t.Currency, &t.BookingDate, &t.ValueDate, &t.TransactionDate,
		&t.CounterpartyName, &t.Remittance, &t.FirstSeenAt, &t.LastSeenAt); err != nil {
		return Transaction{}, fmt.Errorf("reading a transaction: %w", err)
	}
	return t, nil
}

// CountLedger is how many transactions a member may see, for the toolbar. It is
// a count of what they may see and not a page count, and it survives paging
// unchanged.
func (db *DB) CountLedger(ctx context.Context, memberID, accountID string) (int, error) {
	query := `select count(*) from transactions where account_id in (` + ownedAccounts + `)`
	var n int
	if err := db.pool.QueryRow(ctx, query, memberID, nullString(accountID)).Scan(&n); err != nil {
		return 0, fmt.Errorf("counting the ledger: %w", err)
	}
	return n, nil
}

// SyncResult is what one account's sync wrote.
type SyncResult struct {
	BookedInserted int
	BookedUpdated  int
	PendingWritten int
}

// WriteAccountTransactions applies one sync to one account, in one transaction.
//
// Booked rows are append-only: they are inserted, or found and brought up to
// date, and never removed. Pending rows are a replaceable set — the account's
// are discarded and the ones the bank just returned are written in their place.
//
// That asymmetry deletes the hardest problem in bank data rather than solving
// it. A pending transaction becomes a booked one under a different reference, a
// different amount and a different date; matching the two is guesswork that is
// wrong in exactly the cases a household argues about. Under this rule the
// pending row ceases to exist at the next sync and the booked row arrives on its
// own (ADR 0021).
//
// syncedThrough is the date the account's transactions are believed complete
// through, and is written even when the fill was partial, so the next sync
// continues rather than starting the history again.
func (db *DB) WriteAccountTransactions(
	ctx context.Context, accountID string, txs []Transaction, syncedThrough time.Time,
) (SyncResult, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("storing transactions: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var result SyncResult

	if _, err := tx.Exec(ctx,
		`delete from transactions where account_id = $1 and status = 'pending'`, accountID); err != nil {
		return SyncResult{}, fmt.Errorf("clearing unsettled transactions: %w", err)
	}

	// The occurrence is the ordinal of this row among the ones sharing its key
	// in what the bank just returned. Written this way the count comparison the
	// decision describes falls out and cannot race: a sync returning two and
	// finding one stored writes ordinals 1 and 2, the first is found and
	// updated, and exactly one row is inserted.
	seen := map[string]int{}

	const upsertBooked = `
		insert into transactions
			(account_id, status, dedup_key, occurrence, amount_minor, currency,
			 booking_date, value_date, transaction_date, counterparty_name, remittance)
		values ($1, 'booked', $2, $3, $4, $5, $6, $7, $8, $9, $10)
		on conflict (account_id, dedup_key, occurrence) where status = 'booked'
		do update set
			-- A bank amending a transaction it already returned is found here,
			-- where it gave an entry reference. Where wimm fell back to a
			-- digest the amendment produces a different key and arrives as a
			-- new row; last_seen_at is what a later change will use to find the
			-- one the bank stopped returning.
			amount_minor      = excluded.amount_minor,
			currency          = excluded.currency,
			booking_date      = excluded.booking_date,
			value_date        = excluded.value_date,
			transaction_date  = excluded.transaction_date,
			counterparty_name = excluded.counterparty_name,
			remittance        = excluded.remittance,
			last_seen_at      = now()
		returning (xmax = 0) as inserted`

	const insertPending = `
		insert into transactions
			(account_id, status, dedup_key, occurrence, amount_minor, currency,
			 booking_date, value_date, transaction_date, counterparty_name, remittance)
		values ($1, 'pending', $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	for _, t := range txs {
		if t.BookingDate.IsZero() {
			// Unfileable: the ledger orders and groups by this one date, and
			// inventing today would put the transaction on the wrong day.
			continue
		}
		seen[string(t.Status)+"\x1f"+t.DedupKey]++
		occurrence := seen[string(t.Status)+"\x1f"+t.DedupKey]

		args := []any{
			accountID, t.DedupKey, occurrence, t.AmountMinor, t.Currency,
			t.BookingDate, t.ValueDate, t.TransactionDate,
			nullString(t.CounterpartyName), nullString(t.Remittance),
		}

		if t.Status == StatusPending {
			if _, err := tx.Exec(ctx, insertPending, args...); err != nil {
				return SyncResult{}, fmt.Errorf("storing an unsettled transaction: %w", err)
			}
			result.PendingWritten++
			continue
		}

		var inserted bool
		if err := tx.QueryRow(ctx, upsertBooked, args...).Scan(&inserted); err != nil {
			return SyncResult{}, fmt.Errorf("storing a transaction: %w", err)
		}
		if inserted {
			result.BookedInserted++
		} else {
			result.BookedUpdated++
		}
	}

	const recordSync = `
		update accounts
		set transactions_synced_through = $2, transactions_synced_at = now()
		where id = $1`
	if _, err := tx.Exec(ctx, recordSync, accountID, syncedThrough); err != nil {
		return SyncResult{}, fmt.Errorf("recording the sync: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return SyncResult{}, fmt.Errorf("storing transactions: %w", err)
	}
	return result, nil
}

// AccountSync is where an account's ledger has reached.
type AccountSync struct {
	// SyncedThrough is the date the transactions are believed complete through.
	// Nil means nothing has ever been read, which is a different state from an
	// account read and found empty: the screen says so rather than showing one
	// empty list for both.
	SyncedThrough *time.Time
	// SyncedAt is when that was last established.
	SyncedAt *time.Time
	// Oldest is the date the stored ledger actually reaches back to. wimm
	// states this rather than a period, because no bank says in advance how
	// much history it will hand over.
	Oldest *time.Time
}

// AccountSyncState reads where one account's ledger has reached.
func (db *DB) AccountSyncState(ctx context.Context, accountID string) (AccountSync, error) {
	const query = `
		select a.transactions_synced_through, a.transactions_synced_at,
		       (select min(t.booking_date) from transactions t where t.account_id = a.id)
		from accounts a where a.id = $1`

	var s AccountSync
	err := db.pool.QueryRow(ctx, query, accountID).Scan(&s.SyncedThrough, &s.SyncedAt, &s.Oldest)
	if err != nil {
		return AccountSync{}, fmt.Errorf("reading an account's sync state: %w", err)
	}
	return s, nil
}

// SyncableAccount is one account a member owns, with everything a sync needs:
// where its ledger has reached, and the live connection that can read it.
//
// It exists because a sync's scope is ownership, not visibility: an account
// with no owner is never read, and an account a member merely has a level on is
// never read for them.
type SyncableAccount struct {
	Account
	BankID        string
	SyncedThrough *time.Time
	SyncedAt      *time.Time
	// Connection is the live connection reading this account. Named rather
	// than taken from the embedded Account's nullable column, because a
	// syncable account always has one.
	Connection       string
	BankName         string
	ConsentExpiresAt time.Time
	// Scope is what the connection's consent covers. An account at a narrow
	// connection is never synced: the bank was never asked.
	Scope ConnectionScope
}

// ConnectionScope is what a connection's consent covers. It is a property of
// the connection and separate from its state: a connection that is live with
// the narrow scope is working, and is described to a member as connected before
// wimm could read transactions rather than as broken.
type ConnectionScope string

const (
	ScopeBalances                ConnectionScope = "balances"
	ScopeBalancesAndTransactions ConnectionScope = "balances_and_transactions"
)

// ReadsTransactions reports whether this consent covers the ledger.
func (s ConnectionScope) ReadsTransactions() bool { return s == ScopeBalancesAndTransactions }

// SyncableAccounts lists the accounts a member owns at live connections, for
// syncing. Accounts at a disconnected or expired connection are absent: nothing
// new arrives for them, which is the whole point of disconnecting.
func (db *DB) SyncableAccounts(ctx context.Context, memberID string) ([]SyncableAccount, error) {
	const query = `
		select a.id, a.source, a.connection_id, coalesce(a.gateway_ref, ''),
		       a.gateway_uid_sealed, coalesce(a.key_id, ''),
		       coalesce(a.name, ''), coalesce(a.number_suffix, ''), coalesce(a.account_type, ''),
		       coalesce(a.holder_name, ''), a.currency, a.balance_minor, a.balance_read_at,
		       coalesce(a.bank_id, ''), a.transactions_synced_through, a.transactions_synced_at,
		       c.id, c.bank_name, c.consent_expires_at, c.scope
		from accounts a
		join account_owners o on o.account_id = a.id and o.member_id = $1
		join bank_connections c on c.id = a.connection_id
		where c.disconnected_at is null and c.expired_at is null
		order by c.bank_name, a.name`

	rows, err := db.pool.Query(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("listing accounts to sync: %w", err)
	}
	defer rows.Close()

	var out []SyncableAccount
	for rows.Next() {
		var s SyncableAccount
		if err := rows.Scan(&s.ID, &s.Source, &s.ConnectionID, &s.GatewayRef,
			&s.GatewayUID.Ciphertext, &s.GatewayUID.KeyID,
			&s.Name, &s.NumberSuffix, &s.AccountType, &s.HolderName, &s.Currency,
			&s.BalanceMinor, &s.BalanceReadAt,
			&s.BankID, &s.SyncedThrough, &s.SyncedAt,
			&s.Connection, &s.BankName, &s.ConsentExpiresAt, &s.Scope); err != nil {
			return nil, fmt.Errorf("listing accounts to sync: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// NarrowConnection is a bank that is live and cannot be read for transactions,
// which is neither working nor expired. It is surfaced as an inline alert on
// that bank and never as a page banner: a household with one narrow bank and
// two wide ones must not be warned about the whole product.
type NarrowConnection struct {
	ConnectionID string
	BankName     string
}

// NarrowConnections lists the live connections a member owns an account at
// whose consent does not cover transactions.
func (db *DB) NarrowConnections(ctx context.Context, memberID string) ([]NarrowConnection, error) {
	const query = `
		select distinct c.id, c.bank_name
		from bank_connections c
		join accounts a on a.connection_id = c.id
		join account_owners o on o.account_id = a.id and o.member_id = $1
		where c.disconnected_at is null and c.scope <> 'balances_and_transactions'
		order by c.bank_name`

	rows, err := db.pool.Query(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("listing narrow connections: %w", err)
	}
	defer rows.Close()

	var out []NarrowConnection
	for rows.Next() {
		var n NarrowConnection
		if err := rows.Scan(&n.ConnectionID, &n.BankName); err != nil {
			return nil, fmt.Errorf("listing narrow connections: %w", err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// SetConnectionScope records what a connection's consent covers. A first
// connection writes the wide scope; widening an existing one rewrites it after
// the member has confirmed again at their bank.
func (db *DB) SetConnectionScope(ctx context.Context, connectionID string, scope ConnectionScope) error {
	tag, err := db.pool.Exec(ctx,
		`update bank_connections set scope = $2 where id = $1`, connectionID, string(scope))
	if err != nil {
		return fmt.Errorf("recording a connection's scope: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// BankConnectionScope reads what one connection's consent covers.
func (db *DB) BankConnectionScope(ctx context.Context, connectionID string) (ConnectionScope, error) {
	var scope ConnectionScope
	err := db.pool.QueryRow(ctx,
		`select scope from bank_connections where id = $1`, connectionID).Scan(&scope)
	if err != nil {
		return "", fmt.Errorf("reading a connection's scope: %w", err)
	}
	return scope, nil
}

// LedgerState is what the Transactions screen says about itself: how current it
// is, how far back it reaches, and whether this member owns any account at all.
type LedgerState struct {
	// SyncedAt is the most recent time any account in scope was brought up to
	// date. Nil means nothing has ever been read, which the screen states
	// rather than implying with an empty list.
	SyncedAt *time.Time
	// ReachesBack is the oldest transaction held in scope. wimm states the date
	// the list actually reaches rather than a period, because no bank says in
	// advance how much history it will hand over.
	ReachesBack *time.Time
	// OwnedAccounts is how many accounts this member owns, in scope. Zero is a
	// different empty state from an account read and found empty.
	OwnedAccounts int
}

// LedgerState reads the three facts the list header and the empty states need,
// scoped exactly as the ledger itself is.
func (db *DB) LedgerState(ctx context.Context, memberID, accountID string) (LedgerState, error) {
	const query = `
		with owned as (` + ownedAccounts + `)
		select (select max(a.transactions_synced_at) from accounts a where a.id in (select id from owned)),
		       (select min(t.booking_date) from transactions t where t.account_id in (select id from owned)),
		       (select count(*) from owned)`

	var s LedgerState
	err := db.pool.QueryRow(ctx, query, memberID, nullString(accountID)).
		Scan(&s.SyncedAt, &s.ReachesBack, &s.OwnedAccounts)
	if err != nil {
		return LedgerState{}, fmt.Errorf("reading the ledger's state: %w", err)
	}
	return s, nil
}

// AccountLabel is what a ledger row says it came from: the account's name and
// its bank. The list is every account a member owns, so a row has to name which
// one it is on.
type AccountLabel struct {
	Name     string
	BankName string
}

// OwnedAccountLabels names every account a member owns, including accounts at a
// bank that has been disconnected: their transactions are still listed, so
// their rows still have to say where they came from.
func (db *DB) OwnedAccountLabels(ctx context.Context, memberID string) (map[string]AccountLabel, error) {
	const query = `
		select a.id, coalesce(a.name, ''), coalesce(c.bank_name, '')
		from accounts a
		join account_owners o on o.account_id = a.id and o.member_id = $1
		left join bank_connections c on c.id = a.connection_id`

	rows, err := db.pool.Query(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("naming accounts: %w", err)
	}
	defer rows.Close()

	out := map[string]AccountLabel{}
	for rows.Next() {
		var id string
		var label AccountLabel
		if err := rows.Scan(&id, &label.Name, &label.BankName); err != nil {
			return nil, fmt.Errorf("naming accounts: %w", err)
		}
		out[id] = label
	}
	return out, rows.Err()
}
