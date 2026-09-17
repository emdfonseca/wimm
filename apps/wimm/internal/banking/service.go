package banking

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// Store is the part of the database this service uses. Declared here, where it
// is consumed, and sized to what is actually called.
type Store interface {
	CreatePendingBankConnection(ctx context.Context, p store.PendingBankConnection, stateHash []byte, lifetime time.Duration) (store.PendingBankConnection, error)
	ConsumePendingBankConnection(ctx context.Context, stateHash []byte) (store.PendingBankConnection, error)

	CreateBankConnection(ctx context.Context, c store.BankConnection, accounts []store.Account, owner string) (store.BankConnection, []store.Account, error)
	BankConnectionByID(ctx context.Context, id string) (store.BankConnection, error)
	BankConnections(ctx context.Context) ([]store.BankConnection, error)
	MarkBankConnectionExpired(ctx context.Context, id string) error
	DisconnectBankConnection(ctx context.Context, id string) error
	ReplaceConnectionAccounts(ctx context.Context, connectionID string, accounts []store.Account, newOwner string) ([]store.Account, []string, error)

	// Now is database time. Every lifetime in wimm is measured against it
	// rather than a process clock, so a skewed instance cannot grant itself a
	// longer consent than the bank agreed to (ADR 0016).
	Now(ctx context.Context) (store.Time, error)

	VisibleAccounts(ctx context.Context, memberID, connectionID string) ([]store.VisibleAccount, error)
	AccountsForConnection(ctx context.Context, connectionID string) ([]store.Account, error)
	ReadableAccounts(ctx context.Context, connectionID string) ([]store.Account, error)
	AccountByID(ctx context.Context, id string) (store.Account, error)
	AccountOwners(ctx context.Context, accountID string) ([]string, error)
	MemberOwnsAccount(ctx context.Context, accountID, memberID string) (bool, error)
	SetAccountOwners(ctx context.Context, accountID string, memberIDs []string) error
	SetAccountLevel(ctx context.Context, accountID, memberID string, level store.Level, grantedBy string) error
	RecordBalance(ctx context.Context, accountID string, minor int64, readAt time.Time) error

	// Secrets are stored in a second step: a value is sealed against the
	// owning row's id, and that id does not exist until the row does.
	Members(ctx context.Context) ([]store.Member, error)
	GrantsForConnection(ctx context.Context, connectionID string) ([]store.Grant, error)

	SetConnectionSecret(ctx context.Context, connectionID string, sealed store.Sealed) error
	SetAccountSecret(ctx context.Context, accountID string, sealed store.Sealed) error

	// The ledger. Reads are scoped to ownership and syncs to live connections
	// whose consent covers transactions; neither scope is the caller's to
	// decide.
	Ledger(ctx context.Context, q store.LedgerQuery) (store.LedgerPage, error)
	CountLedger(ctx context.Context, memberID, accountID string) (int, error)
	LedgerState(ctx context.Context, memberID, accountID string) (store.LedgerState, error)
	SyncableAccounts(ctx context.Context, memberID string) ([]store.SyncableAccount, error)
	WriteAccountTransactions(ctx context.Context, accountID string, txs []store.Transaction, syncedThrough time.Time) (store.SyncResult, error)
	NarrowConnections(ctx context.Context, memberID string) ([]store.NarrowConnection, error)
	OwnedAccountLabels(ctx context.Context, memberID string) (map[string]store.AccountLabel, error)
	SetConnectionScope(ctx context.Context, connectionID string, scope store.ConnectionScope) error
	BankConnectionScope(ctx context.Context, connectionID string) (store.ConnectionScope, error)
}

// LedgerOptions bounds the ledger. Every one of them is configuration with a
// default a household instance never sets, and each is refused at startup when
// non-positive: a sync bounded by zero pages reads nothing and reports success.
type LedgerOptions struct {
	// Overlap is how far back before the synced-through date an incremental
	// sync re-reads, because a bank can book a transaction with a booking date
	// earlier than the day wimm last synced.
	Overlap time.Duration
	// SyncInterval is the least time between two syncs of one account on
	// arrival. Refresh is the member asking in as many words and is not bound
	// by it.
	SyncInterval time.Duration
	// MaxPages bounds one account's sync, because a first fill at a bank that
	// keeps years of history is otherwise unbounded.
	MaxPages int
	// PageSize is how many transactions one screen of the ledger holds.
	PageSize int
}

// ErrNotOwner refuses a member changing an account that is not theirs. Any
// owner may; the member who connected the bank has no standing power once
// ownership is assigned (ADR 0019).
var ErrNotOwner = errors.New("banking: only an owner of this account may change it")

// ErrBankNotFound is a bank id the gateway does not offer.
var ErrBankNotFound = errors.New("banking: no such bank")

// handoffLifetime is how long a member has to finish at their bank. Long
// enough to find a card reader and a phone; short enough that an abandoned
// authorisation does not sit around.
const handoffLifetime = 30 * time.Minute

// Service is the banking domain. It holds the gateway behind the port, so
// nothing above it names one.
type Service struct {
	store   Store
	gateway Gateway
	keys    *Keyring
	log     *slog.Logger

	redirectURL string
	// staleAfter is how old a reading may be before it is shown as stale. It
	// never hides a reading.
	staleAfter time.Duration

	overlap      time.Duration
	syncInterval time.Duration
	maxPages     int
	pageSize     int
}

// NewService wires the domain.
func NewService(
	s Store, g Gateway, keys *Keyring, log *slog.Logger,
	redirectURL string, staleAfter time.Duration, ledger LedgerOptions,
) *Service {
	return &Service{
		store: s, gateway: g, keys: keys, log: log,
		redirectURL: redirectURL, staleAfter: staleAfter,
		overlap:      ledger.Overlap,
		syncInterval: ledger.SyncInterval,
		maxPages:     ledger.MaxPages,
		pageSize:     ledger.PageSize,
	}
}

// Banks lists what can be connected.
func (s *Service) Banks(ctx context.Context, country string) ([]Bank, error) {
	return s.gateway.Banks(ctx, country)
}

// ConsentHandoff is where to send a member and when their access would end.
type ConsentHandoff struct {
	URL              string
	ConsentExpiresAt time.Time
}

// BeginConnection records a pending connection and returns where to send the
// member. The consent date is the bank's own maximum, which is why the screen
// can state a real date rather than a wimm policy.
func (s *Service) BeginConnection(ctx context.Context, memberID, bankID string) (ConsentHandoff, error) {
	banks, err := s.gateway.Banks(ctx, "")
	if err != nil {
		return ConsentHandoff{}, err
	}
	bank, ok := findBank(banks, bankID)
	if !ok {
		return ConsentHandoff{}, fmt.Errorf("%q: %w", bankID, ErrBankNotFound)
	}
	return s.beginFor(ctx, memberID, bank, nil)
}

// RestoreConnection re-enters the hand-off for a connection whose access has
// run out, skipping the picker: the bank is already known.
func (s *Service) RestoreConnection(ctx context.Context, memberID, connectionID string) (ConsentHandoff, error) {
	conn, err := s.store.BankConnectionByID(ctx, connectionID)
	if err != nil {
		return ConsentHandoff{}, err
	}

	banks, err := s.gateway.Banks(ctx, "")
	if err != nil {
		return ConsentHandoff{}, err
	}
	bank, ok := findBank(banks, conn.BankID)
	if !ok {
		return ConsentHandoff{}, fmt.Errorf("%q: %w", conn.BankID, ErrBankNotFound)
	}
	return s.beginFor(ctx, memberID, bank, &conn.ID)
}

func (s *Service) beginFor(ctx context.Context, memberID string, bank Bank, restores *string) (ConsentHandoff, error) {
	state, hash, err := mintState()
	if err != nil {
		return ConsentHandoff{}, err
	}

	// The bank's maximum, asked for in full: a shorter consent buys nothing and
	// costs the member another trip sooner. Measured from database time, not
	// this process's clock.
	now, err := s.store.Now(ctx)
	if err != nil {
		return ConsentHandoff{}, err
	}
	validUntil := now.T.Add(bank.MaxConsent)

	handoff, err := s.gateway.BeginConnection(ctx, BeginRequest{
		Bank: bank, RedirectURL: s.redirectURL, ValidUntil: validUntil, State: state,
	})
	if err != nil {
		return ConsentHandoff{}, err
	}

	if _, err := s.store.CreatePendingBankConnection(ctx, store.PendingBankConnection{
		Gateway: s.gateway.Name(), GatewayRef: handoff.GatewayRef,
		BankID: bank.ID, BankName: bank.Name, RedirectURL: s.redirectURL,
		Restores: restores, StartedBy: memberID,
	}, hash, handoffLifetime); err != nil {
		return ConsentHandoff{}, err
	}

	return ConsentHandoff{URL: handoff.URL, ConsentExpiresAt: validUntil}, nil
}

// Completed is what a member's return produced.
type Completed struct {
	Connection store.BankConnection
	Accounts   []store.VisibleAccount
	// WithdrawnNames is set on a restore: accounts the bank no longer offers.
	WithdrawnNames []string
}

// CompleteConnection exchanges a return for live access.
//
// It takes the callback and finds the pending row itself, because the row is
// what says whether this is a first connection or a restore — and a caller that
// had to know already would be doing the service's job.
func (s *Service) CompleteConnection(ctx context.Context, memberID string, cb Callback) (Completed, error) {
	pending, err := s.store.ConsumePendingBankConnection(ctx, hashState(cb.State))
	if err != nil {
		return Completed{}, err
	}

	conn, accounts, err := s.gateway.CompleteConnection(ctx, PendingConnection{
		Bank:        Bank{ID: pending.BankID, Name: pending.BankName},
		GatewayRef:  pending.GatewayRef,
		RedirectURL: pending.RedirectURL,
		State:       cb.State,
	}, cb)
	if err != nil {
		return Completed{}, err
	}

	if pending.Restores != nil {
		return s.completeRestore(ctx, memberID, *pending.Restores, conn, accounts)
	}
	return s.completeFirst(ctx, memberID, pending, conn, accounts)
}

func (s *Service) completeFirst(
	ctx context.Context, memberID string, pending store.PendingBankConnection,
	conn Connection, accounts []Account,
) (Completed, error) {
	// Written with no secrets, then sealed against the ids the write produced.
	stored, storedAccounts, err := s.store.CreateBankConnection(ctx, store.BankConnection{
		Gateway: s.gateway.Name(),
		BankID:  pending.BankID, BankName: pending.BankName,
		ConnectedBy: memberID, ConsentExpiresAt: conn.ExpiresAt,
	}, s.toStoreAccounts(accounts), memberID)
	if err != nil {
		return Completed{}, err
	}

	if err := s.sealInto(ctx, stored.ID, conn.GatewayRef, storedAccounts, accounts); err != nil {
		return Completed{}, err
	}

	// What the member just granted at their bank. A connection made after this
	// change asks for both, so a household connecting its first bank never sees
	// the narrow state (ADR 0021).
	if err := s.store.SetConnectionScope(ctx, stored.ID, store.ScopeBalancesAndTransactions); err != nil {
		return Completed{}, err
	}

	visible, err := s.store.VisibleAccounts(ctx, memberID, stored.ID)
	if err != nil {
		return Completed{}, err
	}
	return Completed{Connection: stored, Accounts: visible}, nil
}

// sealInto seals the connection handle and every account identifier against
// the rows that now hold them.
func (s *Service) sealInto(
	ctx context.Context, connectionID, gatewayRef string, stored []store.Account, fromGateway []Account,
) error {
	sealedRef, err := s.seal(gatewayRef, connectionID)
	if err != nil {
		return err
	}
	if err := s.store.SetConnectionSecret(ctx, connectionID, sealedRef); err != nil {
		return err
	}

	uids := make(map[string]string, len(fromGateway))
	for _, a := range fromGateway {
		uids[a.Ref] = a.GatewayUID
	}
	for _, a := range stored {
		uid, ok := uids[a.GatewayRef]
		if !ok || uid == "" {
			continue
		}
		sealed, err := s.seal(uid, a.ID)
		if err != nil {
			return err
		}
		if err := s.store.SetAccountSecret(ctx, a.ID, sealed); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) completeRestore(
	ctx context.Context, memberID, connectionID string, conn Connection, accounts []Account,
) (Completed, error) {
	stored, withdrawn, err := s.store.ReplaceConnectionAccounts(
		ctx, connectionID, s.toStoreAccounts(accounts), memberID)
	if err != nil {
		return Completed{}, err
	}

	// A restore issues fresh identifiers for everything, so every secret is
	// resealed — against the same rows, which is why the owners and levels on
	// them are untouched.
	if err := s.sealInto(ctx, connectionID, conn.GatewayRef, stored, accounts); err != nil {
		return Completed{}, err
	}

	// A restore is also how a narrow connection is widened: the member has just
	// confirmed at the bank against a request that asked for both, so the row
	// now says so. Widening and restoring are one journey with one reason each.
	if err := s.store.SetConnectionScope(ctx, connectionID, store.ScopeBalancesAndTransactions); err != nil {
		return Completed{}, err
	}

	restored, err := s.store.BankConnectionByID(ctx, connectionID)
	if err != nil {
		return Completed{}, err
	}
	visible, err := s.store.VisibleAccounts(ctx, memberID, connectionID)
	if err != nil {
		return Completed{}, err
	}
	return Completed{Connection: restored, Accounts: visible, WithdrawnNames: withdrawn}, nil
}

// toStoreAccounts converts what the gateway returned into rows. Secrets are
// absent here and sealed once the rows exist.
func (s *Service) toStoreAccounts(accounts []Account) []store.Account {
	out := make([]store.Account, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, store.Account{
			Source: store.SourceGateway, GatewayRef: a.Ref,
			Name: a.Name, NumberSuffix: a.NumberSuffix, AccountType: a.Type,
			HolderName: a.HolderName, Currency: a.Currency,
		})
	}
	return out
}

func (s *Service) seal(plaintext, owner string) (store.Sealed, error) {
	if plaintext == "" {
		return store.Sealed{}, nil
	}
	sealed, err := s.keys.Seal([]byte(plaintext), owner)
	if err != nil {
		return store.Sealed{}, err
	}
	return store.Sealed{Ciphertext: sealed.Ciphertext(), KeyID: sealed.KeyID()}, nil
}

func (s *Service) open(sealed store.Sealed, owner string) (string, error) {
	if len(sealed.Ciphertext) == 0 {
		return "", errors.New("banking: nothing sealed here to open")
	}
	plaintext, err := s.keys.Open(FromStored(sealed.KeyID, sealed.Ciphertext), owner)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// SetOwners changes who an account belongs to. Only an owner may.
func (s *Service) SetOwners(ctx context.Context, memberID, accountID string, owners []string) error {
	if err := s.requireOwner(ctx, accountID, memberID); err != nil {
		return err
	}
	return s.store.SetAccountOwners(ctx, accountID, owners)
}

// SetLevel changes what one member may see of one account. Only an owner may.
// An empty level removes the grant, which is what hidden is.
func (s *Service) SetLevel(ctx context.Context, memberID, accountID, subjectID string, level store.Level) error {
	if err := s.requireOwner(ctx, accountID, memberID); err != nil {
		return err
	}
	return s.store.SetAccountLevel(ctx, accountID, subjectID, level, memberID)
}

// requireOwner allows any owner of the account, and the member who connected
// the bank when the account has no owner at all.
//
// The second clause is what makes disowning reversible. Without it an account
// with no owners can never gain one: the check asks for an owner, and there is
// none to be.
func (s *Service) requireOwner(ctx context.Context, accountID, memberID string) error {
	owners, err := s.store.AccountOwners(ctx, accountID)
	if err != nil {
		return err
	}
	for _, owner := range owners {
		if owner == memberID {
			return nil
		}
	}
	if len(owners) > 0 {
		return ErrNotOwner
	}

	account, err := s.store.AccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account.ConnectionID == nil {
		return ErrNotOwner
	}
	conn, err := s.store.BankConnectionByID(ctx, *account.ConnectionID)
	if err != nil {
		return err
	}
	if conn.ConnectedBy != memberID {
		return ErrNotOwner
	}
	return nil
}

func findBank(banks []Bank, id string) (Bank, bool) {
	for _, b := range banks {
		if b.ID == id {
			return b, true
		}
	}
	return Bank{}, false
}

// mintState returns the value handed to the gateway and the hash stored. The
// value is a bearer credential while it is live, so only its hash is kept —
// the same shape enrolment links use (ADR 0016).
func mintState() (value string, hash []byte, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("banking: minting a state value: %w", err)
	}
	value = base64.RawURLEncoding.EncodeToString(raw)
	return value, hashState(value), nil
}

func hashState(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}
