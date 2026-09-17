package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// VisibleAccount is one account as one member may see it. The level is what
// that member holds; an owner's is LevelDetails and Owned is true.
//
// Nothing here is filtered by the caller: a member who may not see an account
// does not get a row for it, so there is no chance of a field surviving into a
// response because a check was forgotten upstream.
type VisibleAccount struct {
	Account

	// Owned reports that this member owns the account, which outranks every
	// level and is why an owner always sees the identifying fields.
	Owned bool
	// Level is what this member may see. Owners carry LevelDetails.
	Level Level

	// Connection is where this account's figures come from, and is absent for
	// an account no gateway sources. Every field on it is meaningless without
	// it, which is why they are one optional struct rather than six nullable
	// columns on the account.
	Connection *AccountConnection
}

// AccountConnection is the gateway connection behind a sourced account.
type AccountConnection struct {
	ID               string
	BankID           string
	BankName         string
	BankLogoURL      string
	ConnectedBy      string
	ConsentExpiresAt time.Time
	Live             bool
}

// SetAccountOwners replaces an account's owners.
//
// Ownership and a grant on one account are mutually exclusive, so any grant
// held by a new owner is removed in the same transaction rather than left to be
// noticed later.
func (db *DB) SetAccountOwners(ctx context.Context, accountID string, memberIDs []string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("setting owners: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `delete from account_owners where account_id = $1`, accountID); err != nil {
		return fmt.Errorf("clearing owners: %w", err)
	}

	const own = `insert into account_owners (account_id, member_id) values ($1, $2)`
	const dropGrant = `delete from account_grants where account_id = $1 and member_id = $2`
	for _, m := range memberIDs {
		if _, err := tx.Exec(ctx, own, accountID, m); err != nil {
			return fmt.Errorf("recording ownership: %w", err)
		}
		if _, err := tx.Exec(ctx, dropGrant, accountID, m); err != nil {
			return fmt.Errorf("clearing a grant superseded by ownership: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("setting owners: %w", err)
	}
	return nil
}

// AccountOwners lists who owns an account.
func (db *DB) AccountOwners(ctx context.Context, accountID string) ([]string, error) {
	rows, err := db.pool.Query(ctx,
		`select member_id from account_owners where account_id = $1 order by created_at`, accountID)
	if err != nil {
		return nil, fmt.Errorf("reading owners: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("reading owners: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// SetAccountLevel gives a member a level on an account, or removes their grant
// when level is empty — hidden is the absence of a row.
//
// An owner cannot hold a grant: the refusal is here rather than a resolution
// rule, because a row that is ignored is a row that will one day be believed.
func (db *DB) SetAccountLevel(ctx context.Context, accountID, memberID string, level Level, grantedBy string) error {
	if level == "" {
		if _, err := db.pool.Exec(ctx,
			`delete from account_grants where account_id = $1 and member_id = $2`,
			accountID, memberID); err != nil {
			return fmt.Errorf("removing a level: %w", err)
		}
		return nil
	}
	if !level.Valid() {
		return fmt.Errorf("%q is not a level", level)
	}

	const owns = `select exists (
		select 1 from account_owners where account_id = $1 and member_id = $2)`
	var isOwner bool
	if err := db.pool.QueryRow(ctx, owns, accountID, memberID).Scan(&isOwner); err != nil {
		return fmt.Errorf("checking ownership: %w", err)
	}
	if isOwner {
		return ErrOwnerHoldsAGrant
	}

	const upsert = `
		insert into account_grants (account_id, member_id, level, granted_by)
		values ($1, $2, $3, $4)
		on conflict (account_id, member_id) do update
			set level = excluded.level, granted_by = excluded.granted_by`
	if _, err := db.pool.Exec(ctx, upsert, accountID, memberID, string(level), grantedBy); err != nil {
		return fmt.Errorf("setting a level: %w", err)
	}
	return nil
}

// MemberOwnsAccount reports whether a member may change an account's owners and
// levels. Any owner may; the member who connected the bank has no standing
// power once ownership is assigned (ADR 0019).
func (db *DB) MemberOwnsAccount(ctx context.Context, accountID, memberID string) (bool, error) {
	const query = `select exists (
		select 1 from account_owners where account_id = $1 and member_id = $2)`
	var owns bool
	if err := db.pool.QueryRow(ctx, query, accountID, memberID).Scan(&owns); err != nil {
		return false, fmt.Errorf("checking ownership: %w", err)
	}
	return owns, nil
}

// visibleAccountsQuery is the one place "what may this member see" is decided.
//
// The join is what does the work: an owner row or a grant row produces a result
// and their absence produces nothing, so a hidden account cannot be returned by
// forgetting a filter. Ownership is coalesced to details because an owner sees
// their account in full.
//
// The connection is joined left, not inner. An account is a thing the household
// has; being sourced from a bank is a property some accounts have and others do
// not, and an inner join here would silently make every future account kind
// invisible rather than failing loudly.
//
// It takes an optional connection filter rather than being written per
// connection, so the same query answers across everything the household holds.
const visibleAccountsQuery = `
	select a.id, a.source, a.connection_id, coalesce(a.gateway_ref, ''),
	       a.gateway_uid_sealed, coalesce(a.key_id, ''),
	       coalesce(a.name, ''), coalesce(a.number_suffix, ''), coalesce(a.account_type, ''),
	       coalesce(a.holder_name, ''), a.currency, a.balance_minor, a.balance_read_at,
	       (o.member_id is not null) as owned,
	       case when o.member_id is not null then 'details' else g.level end as level,
	       c.id, c.bank_id, c.bank_name, coalesce(c.bank_logo_url, ''), c.connected_by,
	       c.consent_expires_at,
	       (c.expired_at is null and c.disconnected_at is null) as live
	from accounts a
	left join bank_connections c on c.id = a.connection_id
	left join account_owners o on o.account_id = a.id and o.member_id = $1
	left join account_grants g on g.account_id = a.id and g.member_id = $1
	where (c.id is null or c.disconnected_at is null)
	  and (o.member_id is not null or g.member_id is not null)
	  and ($2::uuid is null or a.connection_id = $2)
	order by coalesce(c.bank_name, ''), a.name`

// VisibleAccounts returns every account a member may see, at the level they may
// see it. Pass an empty connectionID for every bank.
func (db *DB) VisibleAccounts(ctx context.Context, memberID, connectionID string) ([]VisibleAccount, error) {
	rows, err := db.pool.Query(ctx, visibleAccountsQuery, memberID, nullString(connectionID))
	if err != nil {
		return nil, fmt.Errorf("reading visible accounts: %w", err)
	}
	defer rows.Close()

	var out []VisibleAccount
	for rows.Next() {
		var (
			v                                     VisibleAccount
			level                                 string
			conn                                  AccountConnection
			connID, bankID, bankName, connectedBy *string
			consentExpires                        *time.Time
			live                                  *bool
		)
		if err := rows.Scan(
			&v.ID, &v.Source, &v.ConnectionID, &v.GatewayRef,
			&v.GatewayUID.Ciphertext, &v.GatewayUID.KeyID,
			&v.Name, &v.NumberSuffix, &v.AccountType, &v.HolderName, &v.Currency,
			&v.BalanceMinor, &v.BalanceReadAt, &v.Owned, &level,
			&connID, &bankID, &bankName, &conn.BankLogoURL, &connectedBy,
			&consentExpires, &live,
		); err != nil {
			return nil, fmt.Errorf("reading visible accounts: %w", err)
		}
		v.Level = Level(level)
		if connID != nil {
			conn.ID = *connID
			conn.BankID = deref(bankID)
			conn.BankName = deref(bankName)
			conn.ConnectedBy = deref(connectedBy)
			if consentExpires != nil {
				conn.ConsentExpiresAt = *consentExpires
			}
			conn.Live = live != nil && *live
			v.Connection = &conn
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// AccountsForConnection lists every account on a connection regardless of who
// may see it. This is the chooser's read and the reader's own ownership is not
// what decides it, so the caller must have established that they own something
// here first.
func (db *DB) AccountsForConnection(ctx context.Context, connectionID string) ([]Account, error) {
	const query = `
		select id, source, connection_id, coalesce(gateway_ref, ''), gateway_uid_sealed, coalesce(key_id, ''),
		       coalesce(name, ''), coalesce(number_suffix, ''), coalesce(account_type, ''),
		       coalesce(holder_name, ''), currency, balance_minor, balance_read_at
		from accounts where connection_id = $1 order by name`

	rows, err := db.pool.Query(ctx, query, connectionID)
	if err != nil {
		return nil, fmt.Errorf("reading accounts: %w", err)
	}
	defer rows.Close()

	var out []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Source, &a.ConnectionID, &a.GatewayRef, &a.GatewayUID.Ciphertext,
			&a.GatewayUID.KeyID, &a.Name, &a.NumberSuffix, &a.AccountType, &a.HolderName,
			&a.Currency, &a.BalanceMinor, &a.BalanceReadAt); err != nil {
			return nil, fmt.Errorf("reading accounts: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ReadableAccounts lists the accounts a balance may be read for: those with at
// least one owner or one grant. An account nobody may see is never read, which
// is what makes disowning it meaningful rather than cosmetic (ADR 0019).
func (db *DB) ReadableAccounts(ctx context.Context, connectionID string) ([]Account, error) {
	const query = `
		select a.id, a.source, a.connection_id, coalesce(a.gateway_ref, ''), a.gateway_uid_sealed, coalesce(a.key_id, ''),
		       coalesce(a.name, ''), coalesce(a.number_suffix, ''), coalesce(a.account_type, ''),
		       coalesce(a.holder_name, ''), a.currency, a.balance_minor, a.balance_read_at
		from accounts a
		where ($1::uuid is null or a.connection_id = $1)
		  and (exists (select 1 from account_owners o where o.account_id = a.id)
		    or exists (select 1 from account_grants g where g.account_id = a.id))
		order by a.name`

	rows, err := db.pool.Query(ctx, query, nullString(connectionID))
	if err != nil {
		return nil, fmt.Errorf("reading readable accounts: %w", err)
	}
	defer rows.Close()

	var out []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Source, &a.ConnectionID, &a.GatewayRef, &a.GatewayUID.Ciphertext,
			&a.GatewayUID.KeyID, &a.Name, &a.NumberSuffix, &a.AccountType, &a.HolderName,
			&a.Currency, &a.BalanceMinor, &a.BalanceReadAt); err != nil {
			return nil, fmt.Errorf("reading readable accounts: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// RecordBalance stores a reading with the time it was taken. The two are
// written together because a balance without a read time is not a balance.
func (db *DB) RecordBalance(ctx context.Context, accountID string, minor int64, readAt time.Time) error {
	const update = `
		update accounts set balance_minor = $2, balance_read_at = $3 where id = $1`
	tag, err := db.pool.Exec(ctx, update, accountID, minor, readAt)
	if err != nil {
		return fmt.Errorf("recording a balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// AccountByID loads one account.
func (db *DB) AccountByID(ctx context.Context, id string) (Account, error) {
	const query = `
		select id, source, connection_id, coalesce(gateway_ref, ''), gateway_uid_sealed, coalesce(key_id, ''),
		       coalesce(name, ''), coalesce(number_suffix, ''), coalesce(account_type, ''),
		       coalesce(holder_name, ''), currency, balance_minor, balance_read_at
		from accounts where id = $1`

	var a Account
	err := db.pool.QueryRow(ctx, query, id).Scan(&a.ID, &a.Source, &a.ConnectionID, &a.GatewayRef,
		&a.GatewayUID.Ciphertext, &a.GatewayUID.KeyID, &a.Name, &a.NumberSuffix,
		&a.AccountType, &a.HolderName, &a.Currency, &a.BalanceMinor, &a.BalanceReadAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("reading an account: %w", err)
	}
	return a, nil
}

// ReplaceConnectionAccounts is what a restore writes. Accounts are matched on
// the gateway's cross-session hash, so owners and grants carry forward
// untouched; an account the bank no longer offers is deleted, taking its owners
// and grants with it by cascade; one newly offered arrives owned by the member
// who restored, with nobody granted.
func (db *DB) ReplaceConnectionAccounts(
	ctx context.Context, connectionID string, accounts []Account, newOwner string,
) ([]Account, []string, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("restoring accounts: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// An account's bank is the half of its identity that outlives the
	// connection, so it is read from the connection being restored rather than
	// carried in by the caller, who has no reason to know it.
	var bankID string
	if err := tx.QueryRow(ctx,
		`select bank_id from bank_connections where id = $1`, connectionID).Scan(&bankID); err != nil {
		return nil, nil, fmt.Errorf("reading the connection's bank: %w", err)
	}

	offered := make([]string, 0, len(accounts))
	for _, a := range accounts {
		offered = append(offered, a.GatewayRef)
	}

	// Withdrawn first, and named, because the member is told which account the
	// bank no longer offers.
	const withdraw = `
		delete from accounts
		where connection_id = $1 and gateway_ref <> all($2)
		returning coalesce(name, gateway_ref)`
	rows, err := tx.Query(ctx, withdraw, connectionID, offered)
	if err != nil {
		return nil, nil, fmt.Errorf("removing withdrawn accounts: %w", err)
	}
	var withdrawn []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, nil, fmt.Errorf("removing withdrawn accounts: %w", err)
		}
		withdrawn = append(withdrawn, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("removing withdrawn accounts: %w", err)
	}

	// Existing rows are updated in place, so their owners and grants survive.
	// A row that did not exist is created and owned by the restoring member.
	stored := make([]Account, 0, len(accounts))
	for _, a := range accounts {
		// Matched on the bank and the hash rather than on this connection, for
		// the reason insertAccounts is: that pair is the account's identity for
		// as long as it exists, and a row that outlived an earlier connection
		// at this bank already has owners.
		const exists = `select exists (
			select 1 from accounts where bank_id = $1 and gateway_ref = $2)`
		var known bool
		if err := tx.QueryRow(ctx, exists, bankID, a.GatewayRef).Scan(&known); err != nil {
			return nil, nil, fmt.Errorf("matching an account: %w", err)
		}

		owner := newOwner
		if known {
			// Already has owners; do not add one.
			owner = ""
		}
		written, err := insertAccounts(ctx, tx, connectionID, bankID, []Account{a}, owner)
		if err != nil {
			return nil, nil, err
		}
		stored = append(stored, written...)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("restoring accounts: %w", err)
	}
	return stored, withdrawn, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SetConnectionSecret stores the sealed gateway handle on a connection.
//
// It is a second step because the additional data a value is sealed against is
// the owning row's id, and that id does not exist until the row does (ADR
// 0018). The row is written with no secret and gains one immediately; the
// window holds a null, never plaintext.
func (db *DB) SetConnectionSecret(ctx context.Context, connectionID string, sealed Sealed) error {
	const update = `update bank_connections set gateway_ref_sealed = $2, key_id = $3 where id = $1`
	tag, err := db.pool.Exec(ctx, update, connectionID, nullBytes(sealed.Ciphertext), nullString(sealed.KeyID))
	if err != nil {
		return fmt.Errorf("storing a connection secret: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetAccountSecret stores the sealed per-session identifier on an account, for
// the same reason and with the same window.
func (db *DB) SetAccountSecret(ctx context.Context, accountID string, sealed Sealed) error {
	const update = `update accounts set gateway_uid_sealed = $2, key_id = $3 where id = $1`
	tag, err := db.pool.Exec(ctx, update, accountID, nullBytes(sealed.Ciphertext), nullString(sealed.KeyID))
	if err != nil {
		return fmt.Errorf("storing an account secret: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Members lists the household, so the chooser can offer a level per person.
// One instance serves one household, so there is no scoping argument.
func (db *DB) Members(ctx context.Context) ([]Member, error) {
	const query = `select id, email, first_name, last_name, created_at from members order by created_at`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	defer rows.Close()

	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Email, &m.FirstName, &m.LastName, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("listing members: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Grant is one member's level on one account.
type Grant struct {
	AccountID string
	MemberID  string
	Level     Level
}

// GrantsForConnection lists every grant on a connection's accounts, for the
// chooser. Hidden pairs are absent, because hidden is the absence of a row.
func (db *DB) GrantsForConnection(ctx context.Context, connectionID string) ([]Grant, error) {
	const query = `
		select g.account_id, g.member_id, g.level
		from account_grants g
		join accounts a on a.id = g.account_id
		where a.connection_id = $1
		order by g.account_id, g.member_id`

	rows, err := db.pool.Query(ctx, query, connectionID)
	if err != nil {
		return nil, fmt.Errorf("reading grants: %w", err)
	}
	defer rows.Close()

	var out []Grant
	for rows.Next() {
		var g Grant
		var level string
		if err := rows.Scan(&g.AccountID, &g.MemberID, &level); err != nil {
			return nil, fmt.Errorf("reading grants: %w", err)
		}
		g.Level = Level(level)
		out = append(out, g)
	}
	return out, rows.Err()
}
