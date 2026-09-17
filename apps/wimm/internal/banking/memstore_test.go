package banking_test

import (
	"context"
	"fmt"
	"slices"
	"time"

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
