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

	// members is the household; nil means Ada and Grace.
	members []store.Member

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

// ReadableAccounts is the guarantee in miniature: a left-out account is not in
// this list, so nothing can read it (ADR 0022). An account always has an
// owner, so that is the only test left.
func (m *memStore) ReadableAccounts(_ context.Context, connectionID string) ([]store.Account, error) {
	var out []store.Account
	for _, a := range m.accounts {
		if connectionID != "" && (a.ConnectionID == nil || *a.ConnectionID != connectionID) {
			continue
		}
		if a.LeftOutAt != nil {
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

// SetAccountOwners refuses emptying an account's owners, the same as the
// database's deferred constraint trigger does for real (ADR 0022): an account
// always has at least one.
func (m *memStore) SetAccountOwners(_ context.Context, accountID string, memberIDs []string) error {
	if len(memberIDs) == 0 {
		return fmt.Errorf("account %s would be left with no owner", accountID)
	}
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

func (m *memStore) SetAccountLeftOut(_ context.Context, accountID string, leftOut bool) error {
	a, ok := m.accounts[accountID]
	if !ok {
		return store.ErrNotFound
	}
	if leftOut {
		at := m.now
		a.LeftOutAt = &at
	} else {
		a.LeftOutAt = nil
	}
	return nil
}

func (m *memStore) SetAccountName(_ context.Context, accountID, householdName string) error {
	a, ok := m.accounts[accountID]
	if !ok {
		return store.ErrNotFound
	}
	a.HouseholdName = householdName
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
	if m.members != nil {
		return m.members, nil
	}
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

func (m *memStore) OwnersForConnection(_ context.Context, connectionID string) ([]store.Owner, error) {
	var out []store.Owner
	for accountID, byMember := range m.owners {
		a, ok := m.accounts[accountID]
		if !ok || a.ConnectionID == nil || *a.ConnectionID != connectionID {
			continue
		}
		for memberID := range byMember {
			out = append(out, store.Owner{AccountID: accountID, MemberID: memberID})
		}
	}
	slices.SortFunc(out, func(a, b store.Owner) int {
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
	for id, a := range m.accounts {
		if accountID != "" && id != accountID {
			continue
		}
		if a.LeftOutAt != nil {
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

// LedgerPageIndex sorts this member's owned, not-left-out transactions the
// same way Ledger does and buckets them into pages of pageSize, newest first
// — the same independent reimplementation Ledger itself is, so a disagreement
// with the SQL is a defect in one of them rather than something this masks.
func (m *memStore) LedgerPageIndex(
	_ context.Context, memberID, accountID string, pageSize int,
) ([]store.PageMarker, error) {
	var all []store.Transaction
	for _, id := range m.ownedBy(memberID, accountID) {
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

	var out []store.PageMarker
	for start := 0; start < len(all); start += pageSize {
		end := min(start+pageSize, len(all))
		newest := all[start]
		oldest := all[end-1]
		out = append(out, store.PageMarker{
			Cursor: store.Cursor{BookingDate: newest.BookingDate, ID: newest.ID},
			Newest: newest.BookingDate,
			Oldest: oldest.BookingDate,
		})
	}
	return out, nil
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

// OwnedAccountsForTrend mirrors the SQL's own filter: owned, not left out,
// and not at a disconnected connection — expired is fine, the same as
// totals().
func (m *memStore) OwnedAccountsForTrend(_ context.Context, memberID string, accountIDs []string) ([]store.TrendAccount, error) {
	var out []store.TrendAccount
	for _, id := range m.ownedBy(memberID, "") {
		if !slices.Contains(accountIDs, id) {
			continue
		}
		a := m.accounts[id]
		var scope store.ConnectionScope
		if a.ConnectionID != nil {
			c, ok := m.connections[*a.ConnectionID]
			if !ok || c.DisconnectedAt != nil {
				continue
			}
			scope = m.scopeOf(c.ID)
		}
		out = append(out, store.TrendAccount{
			ID: id, Currency: a.Currency, BalanceMinor: a.BalanceMinor,
			Oldest: m.oldestBooked(id), Scope: scope,
		})
	}
	return out, nil
}

// oldestBooked is the earliest booked transaction an account holds, nil if
// it holds none — mirrors the real SQL's min(booking_date), never the
// synced-through watermark (which an exhausted sync sets to today, however
// far back the history it exhausted actually reaches).
func (m *memStore) oldestBooked(accountID string) *time.Time {
	var oldest *time.Time
	for _, t := range m.transactions[accountID] {
		if t.Status != store.StatusBooked {
			continue
		}
		if oldest == nil || t.BookingDate.Before(*oldest) {
			d := t.BookingDate
			oldest = &d
		}
	}
	return oldest
}

func (m *memStore) TransactionsSince(_ context.Context, accountID string, since time.Time) ([]store.Transaction, error) {
	var out []store.Transaction
	for _, t := range m.transactions[accountID] {
		if t.Status == store.StatusBooked && !t.BookingDate.Before(since) {
			out = append(out, t)
		}
	}
	slices.SortFunc(out, func(a, b store.Transaction) int {
		switch {
		case a.BookingDate.Before(b.BookingDate):
			return -1
		case a.BookingDate.After(b.BookingDate):
			return 1
		default:
			return 0
		}
	})
	return out, nil
}

// FullAccessCounts mirrors the SQL: per account, the members who own it or
// hold a details grant on it, each counted once.
func (m *memStore) FullAccessCounts(context.Context) (map[string]int, error) {
	out := map[string]int{}
	for id := range m.accounts {
		full := map[string]bool{}
		for member := range m.owners[id] {
			full[member] = true
		}
		for member, level := range m.grants[id] {
			if level == store.LevelDetails {
				full[member] = true
			}
		}
		if len(full) > 0 {
			out[id] = len(full)
		}
	}
	return out, nil
}

func (m *memStore) ownedNotDisconnected(memberID string) []string {
	var out []string
	for _, id := range m.ownedBy(memberID, "") {
		if a := m.accounts[id]; a.ConnectionID != nil {
			if c, ok := m.connections[*a.ConnectionID]; !ok || c.DisconnectedAt != nil {
				continue
			}
		}
		out = append(out, id)
	}
	return out
}

func (m *memStore) OwnedWindowSums(_ context.Context, memberID string, accountIDs []string, from, to time.Time) ([]store.WindowSum, error) {
	byCurrency := map[string]*store.WindowSum{}
	for _, id := range m.scopedOwned(memberID, accountIDs) {
		for _, t := range m.transactions[id] {
			if t.Status != store.StatusBooked || t.BookingDate.Before(from) || !t.BookingDate.Before(to) {
				continue
			}
			w, ok := byCurrency[t.Currency]
			if !ok {
				w = &store.WindowSum{Currency: t.Currency}
				byCurrency[t.Currency] = w
			}
			if t.AmountMinor > 0 {
				w.InMinor += t.AmountMinor
			} else {
				w.OutMinor -= t.AmountMinor
			}
			w.Rows++
		}
	}
	var out []store.WindowSum
	for _, w := range byCurrency {
		out = append(out, *w)
	}
	slices.SortFunc(out, func(a, b store.WindowSum) int { return cmpString(a.Currency, b.Currency) })
	return out, nil
}

// scopedOwned is the accounts a member owns, not disconnected, among ids: the
// SQL's own filter, so a scope can only narrow.
func (m *memStore) scopedOwned(memberID string, accountIDs []string) []string {
	var out []string
	for _, id := range m.ownedNotDisconnected(memberID) {
		if slices.Contains(accountIDs, id) {
			out = append(out, id)
		}
	}
	return out
}

func (m *memStore) OwnedBooked(_ context.Context, memberID string, accountIDs []string, from, to time.Time) ([]store.Transaction, error) {
	var out []store.Transaction
	for _, id := range m.scopedOwned(memberID, accountIDs) {
		for _, t := range m.transactions[id] {
			if t.Status == store.StatusBooked && !t.BookingDate.Before(from) && t.BookingDate.Before(to) {
				out = append(out, t)
			}
		}
	}
	slices.SortFunc(out, func(a, b store.Transaction) int {
		if c := b.BookingDate.Compare(a.BookingDate); c != 0 {
			return c
		}
		return cmpString(b.ID, a.ID)
	})
	return out, nil
}

func (m *memStore) OwnedOutgoing(ctx context.Context, memberID string, accountIDs []string, from, to time.Time) ([]store.Transaction, error) {
	booked, _ := m.OwnedBooked(ctx, memberID, accountIDs, from, to)
	var out []store.Transaction
	for _, t := range booked {
		if t.AmountMinor < 0 {
			out = append(out, t)
		}
	}
	return out, nil
}

func (m *memStore) OwnedMonthlySums(ctx context.Context, memberID string, accountIDs []string, from, to time.Time) ([]store.MonthSum, error) {
	booked, _ := m.OwnedBooked(ctx, memberID, accountIDs, from, to)
	type key struct {
		currency string
		month    time.Time
	}
	sums := map[key]*store.MonthSum{}
	for _, t := range booked {
		k := key{t.Currency, time.Date(t.BookingDate.Year(), t.BookingDate.Month(), 1, 0, 0, 0, 0, time.UTC)}
		sum, ok := sums[k]
		if !ok {
			sum = &store.MonthSum{Currency: k.currency, Month: k.month}
			sums[k] = sum
		}
		if t.AmountMinor > 0 {
			sum.InMinor += t.AmountMinor
		} else {
			sum.OutMinor -= t.AmountMinor
		}
	}
	var out []store.MonthSum
	for _, sum := range sums {
		out = append(out, *sum)
	}
	slices.SortFunc(out, func(a, b store.MonthSum) int {
		if c := cmpString(a.Currency, b.Currency); c != 0 {
			return c
		}
		return a.Month.Compare(b.Month)
	})
	return out, nil
}
