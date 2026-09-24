// Package bankingtest is an in-memory banking.Gateway.
//
// It is part of the port rather than a test helper (ADR 0018). It keeps
// `just check` off the network, and it is a second implementation written the
// same week as the first, which is the cheapest pressure test for whether the
// interface abstracts anything.
//
// Unlike a gateway adapter, this package is meant to be imported from outside
// internal/banking: anything that consumes a banking.Gateway tests against it.
package bankingtest

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

// Gateway is a banking.Gateway whose every outcome is scriptable. The zero
// value is not useful; build one with New.
type Gateway struct {
	mu sync.Mutex

	banks map[string]*bank

	// failures are outcomes scripted onto a method by name, consumed in order
	// so a test can say "fail once, then succeed".
	failures map[string][]error

	// handoffs are the authorisations begun and not yet completed.
	handoffs map[string]banking.PendingConnection

	// live are the connections currently granted, by gateway ref.
	live map[string]*session

	// Calls records every method invoked, in order, so a test can assert that
	// no balance was read for an account nobody may see.
	Calls []Call

	// issued is every session handle and per-session identifier this gateway
	// has ever handed out. A log can then be checked against all of them
	// rather than against whichever happen to be current — which is what task
	// 10.1 asks for.
	issued map[string]bool

	// LastState is the state value the most recent BeginConnection carried. A
	// real bank hands it back on the return, so a test completing the round
	// trip reads it from here rather than reaching into wimm's own hashing.
	LastState string

	// PageSize is how many transactions one read returns. Zero means all of
	// them, which is the common case; a test asserting that paging works to
	// exhaustion sets it small.
	PageSize int

	// Now is the fake's clock. It is a field rather than a call to time.Now
	// for two reasons: the process clock is banned from setting any lifetime
	// in wimm (ADR 0016), and a fake that reads the wall clock makes an expiry
	// assertion depend on how long the test took.
	Now time.Time

	nextRef int
}

// Call is one method invocation. Balance reads name their account, because
// "which accounts were read" is the assertion several requirements turn on.
type Call struct {
	Method     string
	BankID     string
	AccountRef string
}

type bank struct {
	info banking.Bank
	// accounts is what the bank exposes on the next completion.
	accounts []banking.Account
	// balances by account ref.
	balances map[string][]banking.Balance
	// transactions by account ref, newest first, as a real bank returns them.
	// The fake pages over this slice rather than holding pre-cut pages, so a
	// test scripts history and page size separately.
	transactions map[string][]banking.Transaction
}

type session struct {
	conn     banking.Connection
	bankID   string
	accounts []banking.Account
}

// Epoch is the fake's default clock, so an expiry is the same value on every
// run. Override Gateway.Now to move time.
var Epoch = time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)

// New returns a gateway with no banks. Add them with AddBank.
func New() *Gateway {
	return &Gateway{
		Now:      Epoch,
		issued:   map[string]bool{},
		banks:    map[string]*bank{},
		failures: map[string][]error{},
		handoffs: map[string]banking.PendingConnection{},
		live:     map[string]*session{},
	}
}

// Name identifies this gateway on stored connections.
func (g *Gateway) Name() string { return "bankingtest" }

// AddBank registers a connectable bank exposing accounts.
func (g *Gateway) AddBank(info banking.Bank, accounts ...banking.Account) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.banks[info.ID] = &bank{
		info: info, accounts: accounts,
		balances:     map[string][]banking.Balance{},
		transactions: map[string][]banking.Transaction{},
	}
}

// SetBalance scripts what an account reads as.
func (g *Gateway) SetBalance(bankID, accountRef string, balances ...banking.Balance) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if b, ok := g.banks[bankID]; ok {
		b.balances[accountRef] = balances
	}
}

// OffersOnRestore replaces what a bank will expose the next time a connection
// to it is completed. It is how the two restore cases are scripted: a bank that
// offers a new account, and one that withdraws an account it used to offer.
func (g *Gateway) OffersOnRestore(bankID string, accounts ...banking.Account) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if b, ok := g.banks[bankID]; ok {
		b.accounts = accounts
	}
}

// Fail scripts the next call to method to return err. Calls queue, so
// Fail("Balances", a); Fail("Balances", b) fails twice in that order and
// succeeds on the third call. Method names are the interface's own.
func (g *Gateway) Fail(method string, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.failures[method] = append(g.failures[method], err)
}

// next consumes a scripted failure for method, if one is queued.
func (g *Gateway) next(method string) error {
	queued := g.failures[method]
	if len(queued) == 0 {
		return nil
	}
	g.failures[method] = queued[1:]
	return queued[0]
}

func (g *Gateway) record(c Call) { g.Calls = append(g.Calls, c) }

// AccountsRead returns the refs of every account a balance was read for, in
// order. The assertion "no balance is ever read for an account with no owner
// and no grant" is made against this.
func (g *Gateway) AccountsRead() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	var refs []string
	for _, c := range g.Calls {
		if c.Method == "Balances" {
			refs = append(refs, c.AccountRef)
		}
	}
	return refs
}

// Banks lists what can be connected in a country.
func (g *Gateway) Banks(_ context.Context, country string) ([]banking.Bank, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.record(Call{Method: "Banks"})
	if err := g.next("Banks"); err != nil {
		return nil, err
	}

	var out []banking.Bank
	for _, b := range g.banks {
		if country == "" || b.info.Country == country {
			out = append(out, b.info)
		}
	}
	return out, nil
}

// BeginConnection returns somewhere to send the member.
func (g *Gateway) BeginConnection(_ context.Context, req banking.BeginRequest) (banking.Handoff, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.record(Call{Method: "BeginConnection", BankID: req.Bank.ID})
	if err := g.next("BeginConnection"); err != nil {
		return banking.Handoff{}, err
	}
	if _, ok := g.banks[req.Bank.ID]; !ok {
		return banking.Handoff{}, banking.ErrBankUnavailable
	}

	g.LastState = req.State
	g.nextRef++
	ref := fmt.Sprintf("handoff-%d", g.nextRef)
	g.handoffs[ref] = banking.PendingConnection{
		Bank:        req.Bank,
		GatewayRef:  ref,
		RedirectURL: req.RedirectURL,
		State:       req.State,
	}
	return banking.Handoff{URL: "https://bank.example/consent/" + ref, GatewayRef: ref}, nil
}

// CompleteConnection exchanges a return for live access and the accounts the
// bank exposes. Like the real thing, the accounts come back here and cannot be
// listed again afterwards.
func (g *Gateway) CompleteConnection(
	_ context.Context, pending banking.PendingConnection, cb banking.Callback,
) (banking.Connection, []banking.Account, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.record(Call{Method: "CompleteConnection", BankID: pending.Bank.ID})
	if err := g.next("CompleteConnection"); err != nil {
		return banking.Connection{}, nil, err
	}

	if cb.Error != "" {
		return banking.Connection{}, nil, banking.ErrConsentDeclined
	}
	if cb.State != pending.State {
		return banking.Connection{}, nil, fmt.Errorf("bankingtest: callback state does not match the pending row")
	}
	b, ok := g.banks[pending.Bank.ID]
	if !ok {
		return banking.Connection{}, nil, banking.ErrBankUnavailable
	}
	if len(b.accounts) == 0 {
		return banking.Connection{}, nil, banking.ErrNoAccounts
	}

	g.nextRef++
	ref := fmt.Sprintf("session-%d", g.nextRef)

	// Every restore issues fresh per-session identifiers while the
	// cross-session Ref stays put. That is the property account identity
	// depends on, so the fake has to reproduce it or nothing tests it.
	accounts := make([]banking.Account, len(b.accounts))
	for i, a := range b.accounts {
		a.GatewayUID = fmt.Sprintf("%s-uid-%d", ref, i)
		accounts[i] = a
	}

	conn := banking.Connection{
		GatewayRef: ref,
		ExpiresAt:  g.Now.Add(pending.Bank.MaxConsent),
	}
	g.issued[ref] = true
	for _, a := range accounts {
		g.issued[a.GatewayUID] = true
	}

	g.live[ref] = &session{conn: conn, bankID: pending.Bank.ID, accounts: accounts}
	delete(g.handoffs, pending.GatewayRef)

	return conn, accounts, nil
}

// Balances reads one account.
func (g *Gateway) Balances(
	_ context.Context, conn banking.Connection, account banking.Account,
) ([]banking.Balance, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.record(Call{Method: "Balances", AccountRef: account.Ref})
	if err := g.next("Balances"); err != nil {
		return nil, err
	}

	s, ok := g.live[conn.GatewayRef]
	if !ok {
		return nil, banking.ErrConsentExpired
	}
	b := g.banks[s.bankID]
	balances, ok := b.balances[account.Ref]
	if !ok {
		return nil, banking.ErrBankUnavailable
	}
	return balances, nil
}

// SetTransactions scripts an account's history, newest first. PageSize decides
// how much of it one read returns.
func (g *Gateway) SetTransactions(bankID, accountRef string, txs ...banking.Transaction) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if b, ok := g.banks[bankID]; ok {
		b.transactions[accountRef] = txs
	}
}

// Transactions reads one page.
//
// The cursor is an offset rendered as a string, which is exactly as opaque to
// the caller as Enable Banking's continuation key: the point of the fake is
// that nothing above the adapter can learn anything from it.
//
// A zero From returns everything the bank holds, as strategy=longest does. A
// set From returns only what was booked on or after it, which is what makes an
// incremental sync's overlap window testable.
func (g *Gateway) Transactions(
	_ context.Context, conn banking.Connection, account banking.Account, req banking.TransactionsRequest,
) (banking.TransactionsPage, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.record(Call{Method: "Transactions", AccountRef: account.Ref})
	if err := g.next("Transactions"); err != nil {
		return banking.TransactionsPage{}, err
	}

	s, ok := g.live[conn.GatewayRef]
	if !ok {
		return banking.TransactionsPage{}, banking.ErrConsentExpired
	}

	held := g.banks[s.bankID].transactions[account.Ref]
	matching := make([]banking.Transaction, 0, len(held))
	for _, tx := range held {
		if req.From.IsZero() || !tx.BookingDate.Before(req.From) {
			matching = append(matching, tx)
		}
	}

	start := 0
	if req.Cursor != "" {
		n, err := strconv.Atoi(req.Cursor)
		if err != nil || n < 0 || n > len(matching) {
			return banking.TransactionsPage{}, fmt.Errorf("bankingtest: %q is not a cursor this gateway issued", req.Cursor)
		}
		start = n
	}

	size := g.PageSize
	if size <= 0 {
		size = len(matching)
	}
	end := min(start+size, len(matching))

	page := banking.TransactionsPage{Transactions: matching[start:end]}
	if end < len(matching) {
		page.NextCursor = strconv.Itoa(end)
	}
	return page, nil
}

// AccountsSynced returns the refs of every account transactions were read for,
// in order. "No sync runs without a member" and "a second arrival inside the
// interval fetches nothing" are both asserted against this.
func (g *Gateway) AccountsSynced() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	var refs []string
	for _, c := range g.Calls {
		if c.Method == "Transactions" {
			refs = append(refs, c.AccountRef)
		}
	}
	return refs
}

// EndConnection tells the bank wimm is done.
func (g *Gateway) EndConnection(_ context.Context, conn banking.Connection) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.record(Call{Method: "EndConnection"})
	if err := g.next("EndConnection"); err != nil {
		return err
	}
	delete(g.live, conn.GatewayRef)
	return nil
}

// Expire cuts a connection off the way a bank does when consent runs out early,
// so the EXPIRED_SESSION path can be exercised without waiting.
func (g *Gateway) Expire(conn banking.Connection) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.live, conn.GatewayRef)
}

// IssuedSecrets returns every value this gateway has handed out that could
// reach a bank. Nothing in wimm should ever be able to print one.
func (g *Gateway) IssuedSecrets() []string {
	g.mu.Lock()
	defer g.mu.Unlock()

	out := make([]string, 0, len(g.issued))
	for secret := range g.issued {
		if secret != "" {
			out = append(out, secret)
		}
	}
	sort.Strings(out)
	return out
}

// Compile-time proof this is a banking.Gateway. If the port grows a method,
// this is what fails first.
var _ banking.Gateway = (*Gateway)(nil)
