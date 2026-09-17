package banking

import (
	"context"
	"time"
)

// Gateway is wimm's whole vocabulary for reaching a bank. It is deliberately
// narrow and deliberately not shaped like any one provider: the method set was
// validated against a second gateway on paper before it was written, and the
// two findings from that exercise are why CompleteConnection takes both the
// pending row and the callback, and why it returns accounts (ADR 0018).
//
// Every implementation lives under internal/banking/<name>. A lint rule refuses
// an import of one from anywhere else, so the only way to reach a bank from the
// rest of wimm is through this interface.
type Gateway interface {
	// Name is the value stored on every connection, so a later gateway is new
	// connections rather than a reinterpretation of old rows.
	Name() string

	// Banks lists what can be connected in a country, each carrying the
	// longest consent that bank will grant.
	Banks(ctx context.Context, country string) ([]Bank, error)

	// BeginConnection asks the gateway for somewhere to send the member. It
	// returns the URL to send them to and whatever the gateway needs handed
	// back at completion.
	BeginConnection(ctx context.Context, req BeginRequest) (Handoff, error)

	// CompleteConnection exchanges a member's return for live access.
	//
	// It takes both the pending row and the callback because gateways disagree
	// about where the truth is: Enable Banking's is in the callback, and
	// GoCardless's is in the row.
	//
	// It returns accounts because the details are shown once and there is no
	// endpoint that lists them again. A gateway that needs a second round-trip
	// makes it here rather than leaking the cheaper gateway's shape upward.
	CompleteConnection(ctx context.Context, pending PendingConnection, cb Callback) (Connection, []Account, error)

	// Balances reads one account. PSU headers indicating the member is present
	// are the adapter's business, not the caller's.
	Balances(ctx context.Context, conn Connection, account Account) ([]Balance, error)

	// EndConnection tells the bank wimm is done. A gateway that cannot be told
	// is not a failure to disconnect: wimm forgets either way.
	EndConnection(ctx context.Context, conn Connection) error
}

// Bank is one connectable institution.
type Bank struct {
	// ID is the gateway's identifier for this bank, stored on the connection.
	ID string
	// Name is what a member recognises.
	Name string
	// Country is the ISO 3166-1 alpha-2 code.
	Country string
	// LogoURL may be empty; the UI falls back to a mark rather than a gap.
	LogoURL string
	// MaxConsent is the longest this bank will grant. It is the bank's limit
	// and not a wimm policy, which is why the member is shown the resulting
	// date before the hand-off. It varies by an order of magnitude between
	// banks in one country — 90 days at some, 1 day at others.
	MaxConsent time.Duration
}

// BeginRequest is what wimm knows before the member leaves for their bank.
type BeginRequest struct {
	Bank Bank
	// RedirectURL must match one registered with the gateway.
	RedirectURL string
	// ValidUntil is the consent expiry wimm is asking for, already capped to
	// the bank's own maximum by the caller.
	ValidUntil time.Time
	// State is wimm's own opaque value, returned in the callback and matched
	// against the pending row.
	State string
}

// Handoff is where to send the member, plus what completion will need.
type Handoff struct {
	// URL is the bank's consent screen.
	URL string
	// GatewayRef is the gateway's handle for this in-flight authorisation. It
	// is sealed before it is stored.
	GatewayRef string
}

// PendingConnection is the row written before the member left. It is the port's
// type rather than the store's: a gateway needs exactly these fields back, and
// nothing about how they were persisted.
type PendingConnection struct {
	Bank       Bank
	GatewayRef string
	// RedirectURL is repeated at completion because some gateways verify it
	// matches the one the authorisation began with.
	RedirectURL string
	State       string
}

// Callback is what the member's return carried. The raw values are exchanged
// server side and never reach an address a member can bookmark (ADR 0016's
// shape, reused).
type Callback struct {
	// Code is the authorisation code, where the gateway issues one.
	Code string
	// State is the value wimm sent, for matching against the pending row.
	State string
	// Error is the gateway's own refusal code, where the member declined or
	// the bank refused. An adapter maps it onto this package's taxonomy.
	Error string
}

// Connection is live read access to one bank.
type Connection struct {
	// GatewayRef is the gateway's session handle. Sealed at rest: it reads
	// this household's accounts until the grant runs out, which makes it a
	// live credential rather than an identifier.
	GatewayRef string
	// ExpiresAt is when access ends. There is no renewal in open banking:
	// restoring is the whole flow again.
	ExpiresAt time.Time
}

// Account is one account at a connected bank.
type Account struct {
	// Ref is the gateway's cross-session identity — Enable Banking's
	// identification_hash. It survives a restore, which the per-session uid
	// does not, so it is what wimm keys on (ADR 0018).
	Ref string
	// GatewayUID is the per-session identifier the gateway wants back on a
	// read. It is rewritten on every restore and sealed at rest.
	GatewayUID string
	// Name is what the bank calls the account. It may be empty, and wimm shows
	// what the bank did give rather than inventing one.
	Name string
	// NumberSuffix is the last few characters of the account number, which is
	// all that telling two accounts at one bank apart requires. The full
	// number is never stored.
	NumberSuffix string
	// Type is the bank's product type, where it gives one.
	Type string
	// HolderName is the name the bank has on the account, where it gives one.
	HolderName string
	// Currency is the ISO 4217 code this account is denominated in.
	Currency string
}

// Balance is a signed amount with the time it was read. A balance without a
// read time is not a balance in this system: wimm never presents a figure as
// current when it has not been re-read.
type Balance struct {
	Money  Money
	ReadAt time.Time
	// Kind is the gateway's name for which balance this is — available,
	// booked, and so on — kept so the UI can prefer one without the adapter
	// deciding for it.
	Kind string
}

// Money is a signed amount in minor units with its currency. Never a float,
// never a decimal string, and never a bare number: a number without its
// currency is the one that eventually gets added to another.
type Money struct {
	// Minor is the amount in the currency's smallest unit — cents for EUR.
	Minor int64
	// Currency is the ISO 4217 alphabetic code.
	Currency string
}

// SameCurrency reports whether two amounts can be added at all. wimm holds no
// exchange rates, so mixed currencies are never summed: inventing a rate would
// invent the number a household trusts most.
func (m Money) SameCurrency(other Money) bool { return m.Currency == other.Currency }

// Add returns the sum, and false if the currencies differ. The caller totals
// per currency rather than getting a plausible wrong answer.
func (m Money) Add(other Money) (Money, bool) {
	if !m.SameCurrency(other) {
		return Money{}, false
	}
	return Money{Minor: m.Minor + other.Minor, Currency: m.Currency}, true
}

// IsZero reports an account holding nothing, which is listed rather than
// hidden.
func (m Money) IsZero() bool { return m.Minor == 0 }

// Negative reports an overdrawn account. Direction is carried by the sign
// because colour alone must never carry it.
func (m Money) Negative() bool { return m.Minor < 0 }
