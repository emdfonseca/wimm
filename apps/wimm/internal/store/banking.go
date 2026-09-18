package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrPendingConnectionSpent is a return that was already exchanged. It is
// distinct from ErrNotFound on purpose: "this was used" and "this never
// existed" are different answers, and only one of them is true when a member
// reloads the page they landed on (ADR 0018).
var ErrPendingConnectionSpent = errors.New("pending connection already consumed")

// ErrOwnerHoldsAGrant refuses a member being both an owner and a grantee of
// one account. Ownership outranks every level, so a grant beside it is a second
// answer to a settled question, and a row that is ignored is a row that will one
// day be believed (ADR 0019).
var ErrOwnerHoldsAGrant = errors.New("a member cannot hold both ownership and a level on one account")

// ErrAccountWouldHaveNoOwner is the deferred constraint trigger's refusal
// (ADR 0022): a change that would leave an account with no owner at all.
// Leaving the account out is the way to the same end that remains open.
var ErrAccountWouldHaveNoOwner = errors.New("an account cannot be left with no owner")

// Sealed is a value held encrypted, as its two columns. The store moves the
// pair around and never opens it: the key lives outside the database and
// outside this package.
type Sealed struct {
	Ciphertext []byte
	KeyID      string
}

// BankConnection is one household's read access to one bank.
type BankConnection struct {
	ID               string
	Gateway          string
	GatewayRef       Sealed
	BankID           string
	BankName         string
	BankLogoURL      string
	ConnectedBy      string
	CreatedAt        time.Time
	ConsentExpiresAt time.Time
	ExpiredAt        *time.Time
	DisconnectedAt   *time.Time
}

// Live reports whether this connection can still read. It is a property of the
// row rather than a comparison here: the query that loaded it decided, against
// database time.
func (c BankConnection) Live() bool { return c.ExpiredAt == nil && c.DisconnectedAt == nil }

// PendingBankConnection is an authorisation begun and not yet returned from.
type PendingBankConnection struct {
	ID          string
	Gateway     string
	GatewayRef  string
	BankID      string
	BankName    string
	RedirectURL string
	Restores    *string
	StartedBy   string
	ExpiresAt   time.Time
}

// Source says where an account's figures come from.
type Source string

const (
	// SourceGateway is an account an open-banking gateway exposes.
	SourceGateway Source = "gateway"
	// SourceManual is an account the household keeps itself. Nothing in this
	// change creates one; the value exists because the distinction is the
	// reason accounts are not owned by connections.
	SourceManual Source = "manual"
)

// Account is one account the household has. It is deliberately not "a bank
// account": where its figures come from is Source, and an account with no
// gateway behind it is the same kind of thing — owned, shared and totalled
// identically.
//
// It carries no flag saying who may see it: that is account_owners and
// account_grants.
type Account struct {
	ID     string
	Source Source
	// ConnectionID is empty for an account no gateway sources.
	ConnectionID  *string
	GatewayRef    string
	GatewayUID    Sealed
	Name          string
	NumberSuffix  string
	AccountType   string
	HolderName    string
	Currency      string
	BalanceMinor  *int64
	BalanceReadAt *time.Time
	// HouseholdName is the name an owner gave the account. Empty means nobody
	// has; Name, the bank's own, is what shows instead (ADR 0022).
	HouseholdName string
	// LeftOutAt is set when the household keeps a record of this account and
	// wimm does not read it. Nil means the account is in wimm (ADR 0022).
	LeftOutAt *time.Time
}

// LeftOut reports whether this account is left out of wimm.
func (a Account) LeftOut() bool { return a.LeftOutAt != nil }

// Level is what one member may see of one account they do not own. Hidden is
// the absence of a grant rather than a value, so it is not in this list.
type Level string

const (
	LevelBalance Level = "balance"
	LevelDetails Level = "details"
)

// Valid reports whether l is a level the schema's check constraint accepts.
func (l Level) Valid() bool { return l == LevelBalance || l == LevelDetails }

// CreatePendingBankConnection records an authorisation about to be handed off.
// The expiry is database time plus lifetime.
func (db *DB) CreatePendingBankConnection(
	ctx context.Context, p PendingBankConnection, stateHash []byte, lifetime time.Duration,
) (PendingBankConnection, error) {
	const insert = `
		insert into pending_bank_connections
			(state_hash, gateway, gateway_ref, bank_id, bank_name, redirect_url, restores, started_by, expires_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, now() + make_interval(secs => $9))
		returning id, gateway, gateway_ref, bank_id, bank_name, redirect_url, restores, started_by, expires_at`

	var out PendingBankConnection
	err := db.pool.QueryRow(ctx, insert,
		stateHash, p.Gateway, p.GatewayRef, p.BankID, p.BankName, p.RedirectURL,
		p.Restores, p.StartedBy, lifetime.Seconds(),
	).Scan(&out.ID, &out.Gateway, &out.GatewayRef, &out.BankID, &out.BankName,
		&out.RedirectURL, &out.Restores, &out.StartedBy, &out.ExpiresAt)
	if err != nil {
		return PendingBankConnection{}, fmt.Errorf("recording a pending connection: %w", err)
	}
	return out, nil
}

// ConsumePendingBankConnection exchanges a return exactly once.
//
// The update is the check: only a row that is unconsumed and unexpired is
// claimed, both measured against database time. A second return finds nothing
// to claim, and the follow-up read says whether that is because it was spent or
// because it never existed.
func (db *DB) ConsumePendingBankConnection(ctx context.Context, stateHash []byte) (PendingBankConnection, error) {
	const consume = `
		update pending_bank_connections set consumed_at = now()
		where state_hash = $1 and consumed_at is null and expires_at > now()
		returning id, gateway, gateway_ref, bank_id, bank_name, redirect_url, restores, started_by, expires_at`

	var p PendingBankConnection
	err := db.pool.QueryRow(ctx, consume, stateHash).Scan(
		&p.ID, &p.Gateway, &p.GatewayRef, &p.BankID, &p.BankName,
		&p.RedirectURL, &p.Restores, &p.StartedBy, &p.ExpiresAt)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return PendingBankConnection{}, fmt.Errorf("consuming a pending connection: %w", err)
	}

	const spent = `select consumed_at is not null from pending_bank_connections where state_hash = $1`
	var wasConsumed bool
	switch err := db.pool.QueryRow(ctx, spent, stateHash).Scan(&wasConsumed); {
	case errors.Is(err, pgx.ErrNoRows):
		return PendingBankConnection{}, ErrNotFound
	case err != nil:
		return PendingBankConnection{}, fmt.Errorf("checking a pending connection: %w", err)
	case wasConsumed:
		return PendingBankConnection{}, ErrPendingConnectionSpent
	default:
		// The row exists and was not consumed, so it expired.
		return PendingBankConnection{}, ErrNotFound
	}
}

// CreateBankConnection records live access and every account it exposed, in one
// transaction. The accounts come back from the gateway once and cannot be
// listed again, so a partial write here loses them for good.
//
// owner is made the owner of every account: the member linked the bank as
// themselves, and they disown from the chooser (ADR 0019).
func (db *DB) CreateBankConnection(
	ctx context.Context, c BankConnection, accounts []Account, owner string,
) (BankConnection, []Account, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return BankConnection{}, nil, fmt.Errorf("recording a connection: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertConnection = `
		insert into bank_connections
			(gateway, gateway_ref_sealed, key_id, bank_id, bank_name, bank_logo_url,
			 connected_by, consent_expires_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id, gateway, bank_id, bank_name, coalesce(bank_logo_url, ''),
		          connected_by, created_at, consent_expires_at, expired_at, disconnected_at`

	var out BankConnection
	err = tx.QueryRow(ctx, insertConnection,
		c.Gateway, nullBytes(c.GatewayRef.Ciphertext), nullString(c.GatewayRef.KeyID),
		c.BankID, c.BankName, c.BankLogoURL, c.ConnectedBy, c.ConsentExpiresAt,
	).Scan(&out.ID, &out.Gateway, &out.BankID, &out.BankName, &out.BankLogoURL,
		&out.ConnectedBy, &out.CreatedAt, &out.ConsentExpiresAt, &out.ExpiredAt, &out.DisconnectedAt)
	if err != nil {
		return BankConnection{}, nil, fmt.Errorf("recording a connection: %w", err)
	}
	out.GatewayRef = c.GatewayRef

	stored, err := insertAccounts(ctx, tx, out.ID, out.BankID, accounts, owner)
	if err != nil {
		return BankConnection{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return BankConnection{}, nil, fmt.Errorf("recording a connection: %w", err)
	}
	return out, stored, nil
}

// insertAccounts writes what a bank returned, and is also the re-attach: an
// account is identified by its bank and the gateway's cross-session hash for as
// long as it exists, so a row that survived a disconnection is found and
// re-pointed at the new connection rather than created a second time beside it
// (ADR 0021). Its owners, grants and transactions come with it, because nothing
// about the row changes except which connection reads it.
func insertAccounts(
	ctx context.Context, tx pgx.Tx, connectionID, bankID string, accounts []Account, owner string,
) ([]Account, error) {
	const insertAccount = `
		insert into accounts
			(source, connection_id, bank_id, gateway_ref, gateway_uid_sealed, key_id, name,
			 number_suffix, account_type, holder_name, currency)
		values ('gateway', $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		-- The predicate is repeated because the index is partial: an account no
		-- gateway sources has no reference to be unique about.
		on conflict (bank_id, gateway_ref) where gateway_ref is not null
		do update set
			connection_id      = excluded.connection_id,
			gateway_uid_sealed = excluded.gateway_uid_sealed,
			key_id             = excluded.key_id,
			name               = excluded.name,
			number_suffix      = excluded.number_suffix,
			account_type       = excluded.account_type,
			holder_name        = excluded.holder_name,
			currency           = excluded.currency
		returning id, source, connection_id, coalesce(gateway_ref, ''), coalesce(name, ''),
		          coalesce(number_suffix, ''), coalesce(account_type, ''),
		          coalesce(holder_name, ''), currency, balance_minor, balance_read_at,
		          coalesce(household_name, ''), left_out_at`

	const own = `
		insert into account_owners (account_id, member_id)
		values ($1, $2) on conflict do nothing`

	stored := make([]Account, 0, len(accounts))
	for _, a := range accounts {
		var out Account
		err := tx.QueryRow(ctx, insertAccount,
			connectionID, bankID, a.GatewayRef,
			nullBytes(a.GatewayUID.Ciphertext), nullString(a.GatewayUID.KeyID),
			a.Name, a.NumberSuffix, a.AccountType, a.HolderName, a.Currency,
		).Scan(&out.ID, &out.Source, &out.ConnectionID, &out.GatewayRef, &out.Name, &out.NumberSuffix,
			&out.AccountType, &out.HolderName, &out.Currency, &out.BalanceMinor, &out.BalanceReadAt,
			&out.HouseholdName, &out.LeftOutAt)
		if err != nil {
			return nil, fmt.Errorf("recording an account: %w", err)
		}
		out.GatewayUID = a.GatewayUID

		if owner != "" {
			if _, err := tx.Exec(ctx, own, out.ID, owner); err != nil {
				return nil, fmt.Errorf("recording ownership: %w", err)
			}
		}
		stored = append(stored, out)
	}
	return stored, nil
}

// BankConnectionByID loads one connection.
func (db *DB) BankConnectionByID(ctx context.Context, id string) (BankConnection, error) {
	const query = `
		select id, gateway, gateway_ref_sealed, coalesce(key_id, ''), bank_id, bank_name,
		       coalesce(bank_logo_url, ''), connected_by, created_at, consent_expires_at,
		       expired_at, disconnected_at
		from bank_connections where id = $1`

	var c BankConnection
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Gateway, &c.GatewayRef.Ciphertext, &c.GatewayRef.KeyID, &c.BankID, &c.BankName,
		&c.BankLogoURL, &c.ConnectedBy, &c.CreatedAt, &c.ConsentExpiresAt,
		&c.ExpiredAt, &c.DisconnectedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return BankConnection{}, ErrNotFound
	}
	if err != nil {
		return BankConnection{}, fmt.Errorf("reading a connection: %w", err)
	}
	return c, nil
}

// BankConnections lists every connection the household holds, newest first.
// One instance serves one household, so there is no scoping argument.
func (db *DB) BankConnections(ctx context.Context) ([]BankConnection, error) {
	const query = `
		select id, gateway, gateway_ref_sealed, coalesce(key_id, ''), bank_id, bank_name,
		       coalesce(bank_logo_url, ''), connected_by, created_at, consent_expires_at,
		       expired_at, disconnected_at
		from bank_connections
		where disconnected_at is null
		order by created_at desc`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing connections: %w", err)
	}
	defer rows.Close()

	var out []BankConnection
	for rows.Next() {
		var c BankConnection
		if err := rows.Scan(&c.ID, &c.Gateway, &c.GatewayRef.Ciphertext, &c.GatewayRef.KeyID,
			&c.BankID, &c.BankName, &c.BankLogoURL, &c.ConnectedBy, &c.CreatedAt,
			&c.ConsentExpiresAt, &c.ExpiredAt, &c.DisconnectedAt); err != nil {
			return nil, fmt.Errorf("listing connections: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ExpireBankConnectionsPastConsent marks every connection whose consent has run
// out, against database time, and destroys what could reach a bank. The row
// survives for the audit question; nothing openable remains on it (ADR 0018).
func (db *DB) ExpireBankConnectionsPastConsent(ctx context.Context) (int64, error) {
	const update = `
		update bank_connections
		set expired_at = now(), gateway_ref_sealed = null, key_id = null
		where expired_at is null and disconnected_at is null and consent_expires_at <= now()`

	tag, err := db.pool.Exec(ctx, update)
	if err != nil {
		return 0, fmt.Errorf("expiring connections: %w", err)
	}
	return tag.RowsAffected(), nil
}

// MarkBankConnectionExpired records a connection the gateway refused before its
// date, which a member experiences identically to one that ran out.
func (db *DB) MarkBankConnectionExpired(ctx context.Context, id string) error {
	const update = `
		update bank_connections
		set expired_at = coalesce(expired_at, now()), gateway_ref_sealed = null, key_id = null
		where id = $1 and disconnected_at is null`
	tag, err := db.pool.Exec(ctx, update, id)
	if err != nil {
		return fmt.Errorf("expiring a connection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DisconnectBankConnection ends wimm's access to a bank. The connection row
// stays so the audit question has an answer, and every sealed value — on the
// connection and on each of its accounts — is destroyed, so nothing that
// remains can open a bank.
//
// The accounts themselves stay. Ending access and destroying the record of what
// that access read are different decisions, and only the first is being made
// here (ADR 0021): no bank hands history back indefinitely, so a ledger a button
// can erase is a ledger that is gone. They stop being shown and stop counting
// exactly as before, by the disconnected_at filter already in
// visibleAccountsQuery.
func (db *DB) DisconnectBankConnection(ctx context.Context, id string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("disconnecting: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const update = `
		update bank_connections
		set disconnected_at = coalesce(disconnected_at, now()),
		    gateway_ref_sealed = null, key_id = null
		where id = $1`
	tag, err := tx.Exec(ctx, update, id)
	if err != nil {
		return fmt.Errorf("disconnecting: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	const disarm = `
		update accounts set gateway_uid_sealed = null, key_id = null
		where connection_id = $1`
	if _, err := tx.Exec(ctx, disarm, id); err != nil {
		return fmt.Errorf("ending access to accounts: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("disconnecting: %w", err)
	}
	return nil
}

func nullBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
