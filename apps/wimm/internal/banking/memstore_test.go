package banking_test

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// memStore is banking.Store in memory, so the service's own rules are tested
// without a database deciding any of them. The visibility rules are
// reimplemented here rather than imported, which is the point: if the SQL and
// this disagree, one of them is wrong and the store tests cover the SQL.
type memStore struct {
	now time.Time

	pending     map[string]store.PendingBankConnection
	consumed    map[string]bool
	connections map[string]*store.BankConnection
	accounts    map[string]*store.Account
	owners      map[string]map[string]bool // account -> member
	grants      map[string]map[string]store.Level

	scopes        map[string]store.ConnectionScope
	transactions  map[string][]store.Transaction
	syncedThrough map[string]*time.Time
	syncedAt      map[string]*time.Time

	// expired records that a connection was marked as having run out.
	expired bool

	seq int
}

func newMemStore() *memStore {
	return &memStore{
		now:         time.Date(2026, time.September, 17, 9, 0, 0, 0, time.UTC),
		pending:     map[string]store.PendingBankConnection{},
		consumed:    map[string]bool{},
		connections: map[string]*store.BankConnection{},
		accounts:    map[string]*store.Account{},
		owners:      map[string]map[string]bool{},
		grants:      map[string]map[string]store.Level{},

		scopes:        map[string]store.ConnectionScope{},
		transactions:  map[string][]store.Transaction{},
		syncedThrough: map[string]*time.Time{},
		syncedAt:      map[string]*time.Time{},
	}
}

func (m *memStore) id(prefix string) string {
	m.seq++
	return fmt.Sprintf("%s-%d", prefix, m.seq)
}

func (m *memStore) Now(context.Context) (store.Time, error) { return store.Time{T: m.now}, nil }

func (m *memStore) CreatePendingBankConnection(
	_ context.Context, p store.PendingBankConnection, stateHash []byte, lifetime time.Duration,
) (store.PendingBankConnection, error) {
	p.ID = m.id("pending")
	p.ExpiresAt = m.now.Add(lifetime)
	key := string(stateHash)
	m.pending[key] = p
	return p, nil
}

func (m *memStore) ConsumePendingBankConnection(_ context.Context, stateHash []byte) (store.PendingBankConnection, error) {
	key := string(stateHash)
	p, ok := m.pending[key]
	if !ok {
		return store.PendingBankConnection{}, store.ErrNotFound
	}
	if m.consumed[key] {
		return store.PendingBankConnection{}, store.ErrPendingConnectionSpent
	}
	m.consumed[key] = true
	return p, nil
}

func (m *memStore) CreateBankConnection(
	_ context.Context, c store.BankConnection, accounts []store.Account, owner string,
) (store.BankConnection, []store.Account, error) {
	c.ID = m.id("conn")
	c.CreatedAt = m.now
	m.connections[c.ID] = &c

	stored := make([]store.Account, 0, len(accounts))
	for _, a := range accounts {
		a.ID = m.id("acct")
		id := c.ID
		a.ConnectionID = &id
		m.accounts[a.ID] = &a
		if owner != "" {
			m.own(a.ID, owner)
		}
		stored = append(stored, a)
	}
	return c, stored, nil
}

func (m *memStore) own(accountID, memberID string) {
	if m.owners[accountID] == nil {
		m.owners[accountID] = map[string]bool{}
	}
	m.owners[accountID][memberID] = true
	delete(m.grants[accountID], memberID)
}

func (m *memStore) BankConnectionByID(_ context.Context, id string) (store.BankConnection, error) {
	c, ok := m.connections[id]
	if !ok {
		return store.BankConnection{}, store.ErrNotFound
	}
	return *c, nil
}

func (m *memStore) BankConnections(context.Context) ([]store.BankConnection, error) {
	out := make([]store.BankConnection, 0, len(m.connections))
	for _, c := range m.connections {
		if c.DisconnectedAt == nil {
			out = append(out, *c)
		}
	}
	slices.SortFunc(out, func(a, b store.BankConnection) int { return cmpString(a.ID, b.ID) })
	return out, nil
}

func (m *memStore) MarkBankConnectionExpired(_ context.Context, id string) error {
	c, ok := m.connections[id]
	if !ok {
		return store.ErrNotFound
	}
	at := m.now
	c.ExpiredAt = &at
	c.GatewayRef = store.Sealed{}
	m.expired = true
	return nil
}

func (m *memStore) DisconnectBankConnection(_ context.Context, id string) error {
	c, ok := m.connections[id]
	if !ok {
		return store.ErrNotFound
	}
	at := m.now
	c.DisconnectedAt = &at
	c.GatewayRef = store.Sealed{}
	for aid, a := range m.accounts {
		if a.ConnectionID != nil && *a.ConnectionID == id {
			delete(m.accounts, aid)
		}
	}
	return nil
}

func (m *memStore) ReplaceConnectionAccounts(
	_ context.Context, connectionID string, accounts []store.Account, newOwner string,
) ([]store.Account, []string, error) {
	offered := map[string]bool{}
	for _, a := range accounts {
		offered[a.GatewayRef] = true
	}

	var withdrawn []string
	for id, a := range m.accounts {
		if a.ConnectionID != nil && *a.ConnectionID == connectionID && !offered[a.GatewayRef] {
			withdrawn = append(withdrawn, a.Name)
			delete(m.accounts, id)
			delete(m.owners, id)
			delete(m.grants, id)
		}
	}

	stored := make([]store.Account, 0, len(accounts))
	for _, incoming := range accounts {
		existing := m.byGatewayRef(connectionID, incoming.GatewayRef)
		if existing != nil {
			existing.GatewayUID = incoming.GatewayUID
			existing.Name = incoming.Name
			stored = append(stored, *existing)
			continue
		}
		incoming.ID = m.id("acct")
		id := connectionID
		incoming.ConnectionID = &id
		m.accounts[incoming.ID] = &incoming
		m.own(incoming.ID, newOwner)
		stored = append(stored, incoming)
	}
	return stored, withdrawn, nil
}

func (m *memStore) byGatewayRef(connectionID, ref string) *store.Account {
	for _, a := range m.accounts {
		if a.ConnectionID != nil && *a.ConnectionID == connectionID && a.GatewayRef == ref {
			return a
		}
	}
	return nil
}

func (m *memStore) VisibleAccounts(_ context.Context, memberID, connectionID string) ([]store.VisibleAccount, error) {
	var out []store.VisibleAccount
	for _, a := range m.accounts {
		if connectionID != "" && (a.ConnectionID == nil || *a.ConnectionID != connectionID) {
			continue
		}
		if a.ConnectionID != nil {
			if c, ok := m.connections[*a.ConnectionID]; ok && c.DisconnectedAt != nil {
				continue
			}
		}

		owned := m.owners[a.ID][memberID]
		level, granted := m.grants[a.ID][memberID]
		if !owned && !granted {
			continue
		}
		if owned {
			level = store.LevelDetails
		}

		v := store.VisibleAccount{Account: *a, Owned: owned, Level: level}
		if a.ConnectionID != nil {
			if c, ok := m.connections[*a.ConnectionID]; ok {
				v.Connection = &store.AccountConnection{
					ID: c.ID, BankID: c.BankID, BankName: c.BankName,
					ConnectedBy: c.ConnectedBy, ConsentExpiresAt: c.ConsentExpiresAt, Live: c.Live(),
				}
			}
		}
		out = append(out, v)
	}
	slices.SortFunc(out, func(a, b store.VisibleAccount) int { return cmpString(a.ID, b.ID) })
	return out, nil
}

func (m *memStore) AccountsForConnection(_ context.Context, connectionID string) ([]store.Account, error) {
	var out []store.Account
	for _, a := range m.accounts {
		if a.ConnectionID != nil && *a.ConnectionID == connectionID {
			out = append(out, *a)
		}
	}
	slices.SortFunc(out, func(a, b store.Account) int { return cmpString(a.ID, b.ID) })
	return out, nil
}

// ReadableAccounts is the guarantee in miniature: an account with no owner and
// no grant is not in this list, so nothing can read it.
func (m *memStore) ReadableAccounts(_ context.Context, connectionID string) ([]store.Account, error) {
	var out []store.Account
	for _, a := range m.accounts {
		if connectionID != "" && (a.ConnectionID == nil || *a.ConnectionID != connectionID) {
			continue
		}
		if len(m.owners[a.ID]) == 0 && len(m.grants[a.ID]) == 0 {
			continue
		}
		out = append(out, *a)
	}
	slices.SortFunc(out, func(a, b store.Account) int { return cmpString(a.ID, b.ID) })
	return out, nil
}

func (m *memStore) AccountByID(_ context.Context, id string) (store.Account, error) {
	a, ok := m.accounts[id]
	if !ok {
		return store.Account{}, store.ErrNotFound
	}
	return *a, nil
}

func (m *memStore) AccountOwners(_ context.Context, accountID string) ([]string, error) {
	var out []string
	for member := range m.owners[accountID] {
		out = append(out, member)
	}
	slices.Sort(out)
	return out, nil
}

func (m *memStore) MemberOwnsAccount(_ context.Context, accountID, memberID string) (bool, error) {
	return m.owners[accountID][memberID], nil
}

func (m *memStore) SetAccountOwners(_ context.Context, accountID string, memberIDs []string) error {
	delete(m.owners, accountID)
	for _, id := range memberIDs {
		m.own(accountID, id)
	}
	return nil
}

func (m *memStore) SetAccountLevel(_ context.Context, accountID, memberID string, level store.Level, _ string) error {
	if level == "" {
		delete(m.grants[accountID], memberID)
		return nil
	}
	if m.owners[accountID][memberID] {
		return store.ErrOwnerHoldsAGrant
	}
	if m.grants[accountID] == nil {
		m.grants[accountID] = map[string]store.Level{}
	}
	m.grants[accountID][memberID] = level
	return nil
}

func (m *memStore) RecordBalance(_ context.Context, accountID string, minor int64, readAt time.Time) error {
	a, ok := m.accounts[accountID]
	if !ok {
		return store.ErrNotFound
	}
	a.BalanceMinor = &minor
	a.BalanceReadAt = &readAt
	return nil
}

func cmpString(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func (m *memStore) SetConnectionSecret(_ context.Context, connectionID string, sealed store.Sealed) error {
	c, ok := m.connections[connectionID]
	if !ok {
		return store.ErrNotFound
	}
	c.GatewayRef = sealed
	return nil
}

func (m *memStore) SetAccountSecret(_ context.Context, accountID string, sealed store.Sealed) error {
	a, ok := m.accounts[accountID]
	if !ok {
		return store.ErrNotFound
	}
	a.GatewayUID = sealed
	return nil
}

func (m *memStore) Members(context.Context) ([]store.Member, error) {
	return []store.Member{
		{ID: ada, FirstName: "Ada", LastName: "Lovelace"},
		{ID: grace, FirstName: "Grace", LastName: "Hopper"},
	}, nil
}

func (m *memStore) GrantsForConnection(_ context.Context, connectionID string) ([]store.Grant, error) {
	var out []store.Grant
	for accountID, byMember := range m.grants {
		a, ok := m.accounts[accountID]
		if !ok || a.ConnectionID == nil || *a.ConnectionID != connectionID {
			continue
		}
		for memberID, level := range byMember {
			out = append(out, store.Grant{AccountID: accountID, MemberID: memberID, Level: level})
		}
	}
	slices.SortFunc(out, func(a, b store.Grant) int {
		if c := cmpString(a.AccountID, b.AccountID); c != 0 {
			return c
		}
		return cmpString(a.MemberID, b.MemberID)
	})
	return out, nil
}

// everyCiphertext returns every sealed value currently on a row.
func (m *memStore) everyCiphertext() [][]byte {
	var out [][]byte
	for _, c := range m.connections {
		out = append(out, c.GatewayRef.Ciphertext)
	}
	for _, a := range m.accounts {
		out = append(out, a.GatewayUID.Ciphertext)
	}
	return out
}

// The ledger, in memory. The scoping rule is reimplemented here rather than
// imported, for the reason the visibility rules above are: if this and the SQL
// disagree, one of them is wrong and the store tests cover the SQL.

func (m *memStore) scopeOf(connectionID string) store.ConnectionScope {
	if s, ok := m.scopes[connectionID]; ok {
		return s
	}
	// Every row that exists was granted balances and nothing else, which is
	// what the column's default states.
	return store.ScopeBalances
}

func (m *memStore) SetConnectionScope(_ context.Context, connectionID string, scope store.ConnectionScope) error {
	if _, ok := m.connections[connectionID]; !ok {
		return store.ErrNotFound
	}
	m.scopes[connectionID] = scope
	return nil
}

func (m *memStore) BankConnectionScope(_ context.Context, connectionID string) (store.ConnectionScope, error) {
	if _, ok := m.connections[connectionID]; !ok {
		return "", store.ErrNotFound
	}
	return m.scopeOf(connectionID), nil
}

// ownedBy is the ledger's whole scope: accounts this member owns, and no
// account they merely hold a level on.
func (m *memStore) ownedBy(memberID, accountID string) []string {
	var out []string
	for id := range m.accounts {
		if accountID != "" && id != accountID {
			continue
		}
		if m.owners[id][memberID] {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}

func (m *memStore) SyncableAccounts(_ context.Context, memberID string) ([]store.SyncableAccount, error) {
	var out []store.SyncableAccount
	for _, id := range m.ownedBy(memberID, "") {
		a := m.accounts[id]
		if a.ConnectionID == nil {
			continue
		}
		c, ok := m.connections[*a.ConnectionID]
		if !ok || c.DisconnectedAt != nil || c.ExpiredAt != nil {
			continue
		}
		out = append(out, store.SyncableAccount{
			Account:          *a,
			BankID:           c.BankID,
			SyncedThrough:    m.syncedThrough[id],
			SyncedAt:         m.syncedAt[id],
			Connection:       c.ID,
			BankName:         c.BankName,
			ConsentExpiresAt: c.ConsentExpiresAt,
			Scope:            m.scopeOf(c.ID),
		})
	}
	return out, nil
}

func (m *memStore) NarrowConnections(_ context.Context, memberID string) ([]store.NarrowConnection, error) {
	seen := map[string]bool{}
	var out []store.NarrowConnection
	for _, id := range m.ownedBy(memberID, "") {
		a := m.accounts[id]
		if a.ConnectionID == nil {
			continue
		}
		c, ok := m.connections[*a.ConnectionID]
		if !ok || c.DisconnectedAt != nil || seen[c.ID] {
			continue
		}
		if m.scopeOf(c.ID).ReadsTransactions() {
			continue
		}
		seen[c.ID] = true
		out = append(out, store.NarrowConnection{ConnectionID: c.ID, BankName: c.BankName})
	}
	slices.SortFunc(out, func(a, b store.NarrowConnection) int { return cmpString(a.BankName, b.BankName) })
	return out, nil
}

func (m *memStore) WriteAccountTransactions(
	_ context.Context, accountID string, txs []store.Transaction, syncedThrough time.Time,
) (store.SyncResult, error) {
	// Pending is a replaceable set; booked is append-only. The ordinal within
	// what this sync returned is the occurrence, which is what makes the same
	// payment made twice two rows and the same transaction read twice one.
	kept := make([]store.Transaction, 0, len(m.transactions[accountID]))
	for _, t := range m.transactions[accountID] {
		if t.Status == store.StatusBooked {
			kept = append(kept, t)
		}
	}

	var result store.SyncResult
	seen := map[string]int{}
	for _, t := range txs {
		if t.BookingDate.IsZero() {
			continue
		}
		t.AccountID = accountID
		seen[string(t.Status)+"\x1f"+t.DedupKey]++
		t.Occurrence = seen[string(t.Status)+"\x1f"+t.DedupKey]

		if t.Status == store.StatusPending {
			t.ID = m.id("txn")
			kept = append(kept, t)
			result.PendingWritten++
			continue
		}

		replaced := false
		for i, existing := range kept {
			if existing.Status == store.StatusBooked &&
				existing.DedupKey == t.DedupKey && existing.Occurrence == t.Occurrence {
				t.ID = existing.ID
				t.FirstSeenAt = existing.FirstSeenAt
				t.LastSeenAt = m.now
				kept[i] = t
				replaced = true
				result.BookedUpdated++
				break
			}
		}
		if !replaced {
			t.ID = m.id("txn")
			t.FirstSeenAt, t.LastSeenAt = m.now, m.now
			kept = append(kept, t)
			result.BookedInserted++
		}
	}

	slices.SortFunc(kept, func(a, b store.Transaction) int {
		if !a.BookingDate.Equal(b.BookingDate) {
			if a.BookingDate.After(b.BookingDate) {
				return -1
			}
			return 1
		}
		return cmpString(b.ID, a.ID)
	})
	m.transactions[accountID] = kept

	through := syncedThrough
	at := m.now
	m.syncedThrough[accountID] = &through
	m.syncedAt[accountID] = &at
	return result, nil
}

func (m *memStore) Ledger(_ context.Context, q store.LedgerQuery) (store.LedgerPage, error) {
	var all []store.Transaction
	for _, id := range m.ownedBy(q.MemberID, q.AccountID) {
		all = append(all, m.transactions[id]...)
	}
	slices.SortFunc(all, func(a, b store.Transaction) int {
		if !a.BookingDate.Equal(b.BookingDate) {
			if a.BookingDate.After(b.BookingDate) {
				return -1
			}
			return 1
		}
		return cmpString(b.ID, a.ID)
	})

	page := store.LedgerPage{Transactions: all}
	if len(all) > q.Limit {
		page.Transactions = all[:q.Limit]
		page.HasOlder = true
	}
	if len(page.Transactions) > 0 {
		page.Newest = page.Transactions[0].BookingDate
		page.Oldest = page.Transactions[len(page.Transactions)-1].BookingDate
	}
	return page, nil
}

func (m *memStore) CountLedger(_ context.Context, memberID, accountID string) (int, error) {
	n := 0
	for _, id := range m.ownedBy(memberID, accountID) {
		n += len(m.transactions[id])
	}
	return n, nil
}

func (m *memStore) LedgerState(_ context.Context, memberID, accountID string) (store.LedgerState, error) {
	owned := m.ownedBy(memberID, accountID)
	s := store.LedgerState{OwnedAccounts: len(owned)}
	for _, id := range owned {
		if at := m.syncedAt[id]; at != nil && (s.SyncedAt == nil || at.After(*s.SyncedAt)) {
			s.SyncedAt = at
		}
		for _, t := range m.transactions[id] {
			if s.ReachesBack == nil || t.BookingDate.Before(*s.ReachesBack) {
				d := t.BookingDate
				s.ReachesBack = &d
			}
		}
	}
	return s, nil
}

// testLedgerOptions are the bounds the service's tests run under. The page cap
// is small so that a fill stopping at it is reachable, and the interval is long
// so that "a second arrival inside the interval fetches nothing" is a fact
// about the rule rather than about how fast the test ran.
func testLedgerOptions() banking.LedgerOptions {
	return banking.LedgerOptions{
		Overlap:      7 * 24 * time.Hour,
		SyncInterval: time.Hour,
		MaxPages:     3,
		PageSize:     50,
	}
}

func (m *memStore) OwnedAccountLabels(_ context.Context, memberID string) (map[string]store.AccountLabel, error) {
	out := map[string]store.AccountLabel{}
	for _, id := range m.ownedBy(memberID, "") {
		a := m.accounts[id]
		label := store.AccountLabel{Name: a.Name}
		if a.ConnectionID != nil {
			if c, ok := m.connections[*a.ConnectionID]; ok {
				label.BankName = c.BankName
			}
		}
		out[id] = label
	}
	return out, nil
}
