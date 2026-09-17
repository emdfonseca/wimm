package banking

import (
	"context"
	"errors"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// View is what a member sees: the accounts they may see, a total per currency
// of exactly those, and the banks that could not be brought up to date.
//
// Two members of one household get different Views of the same data, and each
// total agrees with the accounts beside it.
type View struct {
	Accounts []store.VisibleAccount
	Totals   []Total
	Failures []BankFailure
}

// Total is one currency's sum. Currencies are never added together: wimm holds
// no rates, and inventing one would invent the number a household trusts most.
type Total struct {
	Money        Money
	AccountCount int
}

// BankFailure names a bank that did not answer, so the member is told which.
type BankFailure struct {
	ConnectionID string
	BankName     string
	Err          error
	// RetryAfter is how long until it can be tried again, where the bank said.
	// Zero means "later", never "now".
	RetryAfter time.Duration
}

// Accounts returns what a member may see, reading balances first unless
// skipRead.
//
// Reading happens on arrival because a member sitting in front of the screen is
// exactly the case gateways do not throttle. A bank that fails leaves its
// previous readings exactly where they were, with their original times — that
// is the one thing that must not be lost.
func (s *Service) Accounts(ctx context.Context, memberID string, skipRead bool) (View, error) {
	var failures []BankFailure
	if !skipRead {
		failures = s.readAll(ctx)
	}

	visible, err := s.store.VisibleAccounts(ctx, memberID, "")
	if err != nil {
		return View{}, err
	}

	// Failures are reported only for banks this member can see something of.
	// Telling them a bank they have no account at is unreachable would leak
	// that the bank is connected at all.
	visibleConnections := make(map[string]bool, len(visible))
	for _, a := range visible {
		if a.Connection != nil {
			visibleConnections[a.Connection.ID] = true
		}
	}
	shown := make([]BankFailure, 0, len(failures))
	for _, f := range failures {
		if visibleConnections[f.ConnectionID] {
			shown = append(shown, f)
		}
	}

	return View{Accounts: visible, Totals: totals(visible), Failures: shown}, nil
}

// readAll reads every live connection's readable accounts, and returns what
// failed rather than stopping at the first one: a household with three banks
// and one outage still sees two banks' figures.
func (s *Service) readAll(ctx context.Context) []BankFailure {
	connections, err := s.store.BankConnections(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "listing connections to read", "error", err)
		return nil
	}

	var failures []BankFailure
	for _, conn := range connections {
		if !conn.Live() {
			continue
		}
		if err := s.readConnection(ctx, conn); err != nil {
			var limited *RateLimitError
			retry := time.Duration(0)
			if errors.As(err, &limited) {
				retry = limited.RetryAfter
			}

			// Access having run out is recorded, not just reported: the next
			// arrival must offer to restore rather than try again and fail.
			if errors.Is(err, ErrConsentExpired) {
				if markErr := s.store.MarkBankConnectionExpired(ctx, conn.ID); markErr != nil {
					s.log.ErrorContext(ctx, "recording an expired connection", "error", markErr)
				}
			}

			failures = append(failures, BankFailure{
				ConnectionID: conn.ID, BankName: conn.BankName, Err: err, RetryAfter: retry,
			})
		}
	}
	return failures
}

// readConnection reads every account on one connection that anybody may see.
//
// An account with no owner and no grant is never in this list, which is what
// makes disowning one meaningful rather than cosmetic: wimm knows it exists and
// does not know what is in it (ADR 0019).
func (s *Service) readConnection(ctx context.Context, conn store.BankConnection) error {
	accounts, err := s.store.ReadableAccounts(ctx, conn.ID)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return nil
	}

	gatewayRef, err := s.open(conn.GatewayRef, conn.ID)
	if err != nil {
		// A connection whose secret cannot be opened cannot be read, and the
		// member's experience of that is identical to access having run out.
		return ErrConsentExpired
	}
	live := Connection{GatewayRef: gatewayRef, ExpiresAt: conn.ConsentExpiresAt}

	var firstErr error
	for _, a := range accounts {
		uid, err := s.open(a.GatewayUID, a.ID)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		balances, err := s.gateway.Balances(ctx, live, Account{Ref: a.GatewayRef, GatewayUID: uid, Currency: a.Currency})
		if err != nil {
			// The first failure is what the member is told about; the rest of
			// the accounts are still attempted, because one account refusing
			// is not the whole bank refusing.
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		balance, ok := preferred(balances)
		if !ok {
			continue
		}
		if err := s.store.RecordBalance(ctx, a.ID, balance.Money.Minor, balance.ReadAt); err != nil {
			s.log.ErrorContext(ctx, "recording a balance", "error", err)
		}
	}
	return firstErr
}

// preferred picks which balance to show when a bank returns several. Closing
// booked and available are both common; the first is taken when nothing is
// recognised rather than showing none.
func preferred(balances []Balance) (Balance, bool) {
	if len(balances) == 0 {
		return Balance{}, false
	}
	for _, want := range []string{"CLAV", "available", "XPCD", "CLBD", "closingBooked"} {
		for _, b := range balances {
			if b.Kind == want {
				return b, true
			}
		}
	}
	return balances[0], true
}

// totals sums per currency, over exactly the accounts the member may see.
func totals(accounts []store.VisibleAccount) []Total {
	order := make([]string, 0, 4)
	sums := map[string]*Total{}

	for _, a := range accounts {
		if a.BalanceMinor == nil {
			continue
		}
		t, ok := sums[a.Currency]
		if !ok {
			t = &Total{Money: Money{Currency: a.Currency}}
			sums[a.Currency] = t
			order = append(order, a.Currency)
		}
		t.Money.Minor += *a.BalanceMinor
		t.AccountCount++
	}

	out := make([]Total, 0, len(order))
	for _, c := range order {
		out = append(out, *sums[c])
	}
	return out
}

// Disconnect removes a bank. Only an owner of one of its accounts may.
//
// The gateway is told, and a gateway that cannot be told does not stop the
// disconnection: wimm forgets either way, and a member must be able to remove a
// bank that is already gone.
func (s *Service) Disconnect(ctx context.Context, memberID, connectionID string) error {
	if err := s.requireOwnerOnConnection(ctx, memberID, connectionID); err != nil {
		return err
	}

	conn, err := s.store.BankConnectionByID(ctx, connectionID)
	if err != nil {
		return err
	}

	if gatewayRef, err := s.open(conn.GatewayRef, conn.ID); err == nil {
		if err := s.gateway.EndConnection(ctx, Connection{GatewayRef: gatewayRef}); err != nil {
			s.log.WarnContext(ctx, "the bank could not be told about a disconnection", "error", err)
		}
	}

	return s.store.DisconnectBankConnection(ctx, connectionID)
}

// ConnectionAccounts is the chooser's read: every account on a connection, with
// who owns each. Only a member who owns at least one of them may see it.
func (s *Service) ConnectionAccounts(ctx context.Context, memberID, connectionID string) ([]store.Account, error) {
	if err := s.requireOwnerOnConnection(ctx, memberID, connectionID); err != nil {
		return nil, err
	}
	return s.store.AccountsForConnection(ctx, connectionID)
}

// requireOwnerOnConnection allows a member who owns any account on the
// connection. Ownership of one account at a bank is what makes that bank
// theirs to manage.
//
// It also allows the member who connected the bank whenever an account on it
// has no owner at all. Without that, disowning your last account locks you out
// of the only screen that could undo it: nobody owns it, so nobody may claim
// it, so it can never be owned again. A dead end reachable by one click.
//
// This is not the standing power ADR 0019 withholds from the connecting
// member. They cannot touch an account somebody else owns; they can only pick
// up one that nobody does.
func (s *Service) requireOwnerOnConnection(ctx context.Context, memberID, connectionID string) error {
	accounts, err := s.store.AccountsForConnection(ctx, connectionID)
	if err != nil {
		return err
	}

	orphaned := false
	for _, a := range accounts {
		owners, err := s.store.AccountOwners(ctx, a.ID)
		if err != nil {
			return err
		}
		for _, owner := range owners {
			if owner == memberID {
				return nil
			}
		}
		if len(owners) == 0 {
			orphaned = true
		}
	}

	if orphaned {
		conn, err := s.store.BankConnectionByID(ctx, connectionID)
		if err != nil {
			return err
		}
		if conn.ConnectedBy == memberID {
			return nil
		}
	}
	return ErrNotOwner
}

// AccountsOnConnection is what the chooser shows: the accounts on one
// connection as the calling member sees them. Only a member who owns at least
// one of them may look.
func (s *Service) AccountsOnConnection(ctx context.Context, memberID, connectionID string) ([]store.VisibleAccount, error) {
	if err := s.requireOwnerOnConnection(ctx, memberID, connectionID); err != nil {
		return nil, err
	}
	return s.store.VisibleAccounts(ctx, memberID, connectionID)
}

// GrantsOnConnection is what each member currently holds on that connection's
// accounts, so the chooser opens showing the choice as it stands.
func (s *Service) GrantsOnConnection(ctx context.Context, memberID, connectionID string) ([]store.Grant, error) {
	if err := s.requireOwnerOnConnection(ctx, memberID, connectionID); err != nil {
		return nil, err
	}
	return s.store.GrantsForConnection(ctx, connectionID)
}

// HouseholdMembers lists who a level can be granted to.
func (s *Service) HouseholdMembers(ctx context.Context) ([]store.Member, error) {
	return s.store.Members(ctx)
}
