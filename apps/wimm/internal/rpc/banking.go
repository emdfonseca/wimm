package rpc

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	bankingv1 "github.com/emdfonseca/wimm/packages/contracts/gen/go/wimm/banking/v1"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
	"github.com/emdfonseca/wimm/apps/wimm/internal/identity"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// BankingServer answers the banking surface the browser reaches.
//
// Every method needs a session: there is no unauthenticated banking call, and
// the member the session names is the one whose view is returned.
type BankingServer struct {
	banking  *banking.Service
	identity *identity.Service
}

// NewBankingServer wires the handler to the domain.
func NewBankingServer(b *banking.Service, i *identity.Service) *BankingServer {
	return &BankingServer{banking: b, identity: i}
}

// member resolves the session cookie to the member making the call.
func (s *BankingServer) member(ctx context.Context, header http.Header) (store.Member, error) {
	return s.identity.MemberForSession(ctx, cookieValue(header, SessionCookie))
}

// ListBanks lists what can be connected in a country.
func (s *BankingServer) ListBanks(
	ctx context.Context, req *connect.Request[bankingv1.ListBanksRequest],
) (*connect.Response[bankingv1.ListBanksResponse], error) {
	if _, err := s.member(ctx, req.Header()); err != nil {
		return nil, toConnectError(err)
	}

	banks, err := s.banking.Banks(ctx, req.Msg.GetCountry())
	if err != nil {
		return nil, toConnectError(err)
	}

	out := make([]*bankingv1.Bank, 0, len(banks))
	for _, b := range banks {
		out = append(out, &bankingv1.Bank{
			Id: b.ID, Name: b.Name, Country: b.Country, LogoUrl: b.LogoURL,
			MaxConsentSeconds: int64(b.MaxConsent.Seconds()),
		})
	}
	return connect.NewResponse(&bankingv1.ListBanksResponse{Banks: out}), nil
}

// BeginConnection returns where to send the member and when access would end.
func (s *BankingServer) BeginConnection(
	ctx context.Context, req *connect.Request[bankingv1.BeginConnectionRequest],
) (*connect.Response[bankingv1.BeginConnectionResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	handoff, err := s.banking.BeginConnection(ctx, m.ID, req.Msg.GetBankId())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.BeginConnectionResponse{
		HandoffUrl:       handoff.URL,
		ConsentExpiresAt: timestamppb.New(handoff.ConsentExpiresAt),
	}), nil
}

// CompleteConnection exchanges the member's return for live access.
func (s *BankingServer) CompleteConnection(
	ctx context.Context, req *connect.Request[bankingv1.CompleteConnectionRequest],
) (*connect.Response[bankingv1.CompleteConnectionResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	done, err := s.banking.CompleteConnection(ctx, m.ID, banking.Callback{
		State: req.Msg.GetState(), Code: req.Msg.GetCode(), Error: req.Msg.GetError(),
	})
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&bankingv1.CompleteConnectionResponse{
		Connection:            toProtoConnection(done.Accounts, s.names(ctx)),
		Accounts:              toProtoAccounts(done.Accounts, s.names(ctx)),
		WithdrawnAccountNames: done.WithdrawnNames,
	}), nil
}

// ListConnectionAccounts is the chooser's read.
func (s *BankingServer) ListConnectionAccounts(
	ctx context.Context, req *connect.Request[bankingv1.ListConnectionAccountsRequest],
) (*connect.Response[bankingv1.ListConnectionAccountsResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	// The member's own view of the connection, which is what they may see.
	visible, err := s.banking.AccountsOnConnection(ctx, m.ID, req.Msg.GetConnectionId())
	if err != nil {
		return nil, toConnectError(err)
	}

	members, err := s.banking.HouseholdMembers(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	grants, err := s.banking.GrantsOnConnection(ctx, m.ID, req.Msg.GetConnectionId())
	if err != nil {
		return nil, toConnectError(err)
	}
	owners, err := s.banking.OwnersOnConnection(ctx, m.ID, req.Msg.GetConnectionId())
	if err != nil {
		return nil, toConnectError(err)
	}

	protoMembers := make([]*bankingv1.Member, 0, len(members))
	for _, other := range members {
		protoMembers = append(protoMembers, &bankingv1.Member{Id: other.ID, DisplayName: other.FirstName})
	}
	protoGrants := make([]*bankingv1.AccountGrant, 0, len(grants))
	for _, g := range grants {
		protoGrants = append(protoGrants, &bankingv1.AccountGrant{
			AccountId: g.AccountID, MemberId: g.MemberID, Level: toProtoLevel(g.Level),
		})
	}

	who := s.names(ctx)
	protoAccounts := toProtoAccounts(visible, who)
	// toProtoAccount has no owners list to draw from — VisibleAccount only
	// carries whether the calling member owns the row, not who else does — so
	// the full set is matched in by account id here instead.
	byAccount := make(map[string][]*bankingv1.Member, len(protoAccounts))
	for _, o := range owners {
		if member := who.member(o.MemberID); member != nil {
			byAccount[o.AccountID] = append(byAccount[o.AccountID], member)
		}
	}
	for _, a := range protoAccounts {
		a.Owners = byAccount[a.Id]
	}

	return connect.NewResponse(&bankingv1.ListConnectionAccountsResponse{
		Accounts: protoAccounts, Members: protoMembers, Grants: protoGrants,
	}), nil
}

// SetAccountOwners changes who an account belongs to.
func (s *BankingServer) SetAccountOwners(
	ctx context.Context, req *connect.Request[bankingv1.SetAccountOwnersRequest],
) (*connect.Response[bankingv1.SetAccountOwnersResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := s.banking.SetOwners(ctx, m.ID, req.Msg.GetAccountId(), req.Msg.GetMemberIds()); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.SetAccountOwnersResponse{
		Account: s.accountAfterChange(ctx, m.ID, req.Msg.GetAccountId()),
	}), nil
}

// SetAccountLevel changes what one member may see of one account.
func (s *BankingServer) SetAccountLevel(
	ctx context.Context, req *connect.Request[bankingv1.SetAccountLevelRequest],
) (*connect.Response[bankingv1.SetAccountLevelResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	if err := s.banking.SetLevel(ctx, m.ID, req.Msg.GetAccountId(), req.Msg.GetMemberId(),
		fromProtoLevel(req.Msg.GetLevel())); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.SetAccountLevelResponse{
		Account: s.accountAfterChange(ctx, m.ID, req.Msg.GetAccountId()),
	}), nil
}

// SetAccountLeftOut leaves an account out of wimm, or brings it back.
func (s *BankingServer) SetAccountLeftOut(
	ctx context.Context, req *connect.Request[bankingv1.SetAccountLeftOutRequest],
) (*connect.Response[bankingv1.SetAccountLeftOutResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := s.banking.SetLeftOut(ctx, m.ID, req.Msg.GetAccountId(), req.Msg.GetLeftOut()); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.SetAccountLeftOutResponse{
		Account: s.accountAfterChange(ctx, m.ID, req.Msg.GetAccountId()),
	}), nil
}

// SetAccountName sets the household's own name for an account, or clears it
// back to the bank's own name.
func (s *BankingServer) SetAccountName(
	ctx context.Context, req *connect.Request[bankingv1.SetAccountNameRequest],
) (*connect.Response[bankingv1.SetAccountNameResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := s.banking.SetName(ctx, m.ID, req.Msg.GetAccountId(), req.Msg.GetHouseholdName()); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.SetAccountNameResponse{
		Account: s.accountAfterChange(ctx, m.ID, req.Msg.GetAccountId()),
	}), nil
}

// accountAfterChange returns the account as the caller now sees it, or nil if
// they can no longer see it at all — which is what disowning their own account
// does, and is a correct answer rather than an error.
func (s *BankingServer) accountAfterChange(ctx context.Context, memberID, accountID string) *bankingv1.Account {
	view, err := s.banking.Accounts(ctx, memberID, true)
	if err != nil {
		return nil
	}
	who := s.names(ctx)
	for _, a := range view.Accounts {
		if a.ID == accountID {
			return toProtoAccount(a, who)
		}
	}
	return nil
}

// names resolves the household once per call. An empty map is a degraded
// answer, not a failed one: an id with no name still identifies a member, and
// failing a whole read because a name could not be looked up would be worse.
func (s *BankingServer) names(ctx context.Context) names {
	members, err := s.banking.HouseholdMembers(ctx)
	if err != nil {
		return names{}
	}
	out := make(names, len(members))
	for _, m := range members {
		out[m.ID] = m.FirstName
	}
	return out
}

// ListAccounts is the screen a member lands on.
func (s *BankingServer) ListAccounts(
	ctx context.Context, req *connect.Request[bankingv1.ListAccountsRequest],
) (*connect.Response[bankingv1.ListAccountsResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	view, err := s.banking.Accounts(ctx, m.ID, req.Msg.GetSkipRead())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.ListAccountsResponse{
		Accounts:        toProtoGroupedAccounts(view, s.names(ctx)),
		HouseholdTotals: toProtoTotals(view.HouseholdTotals),
		OwnTotals:       toProtoTotals(view.OwnTotals),
		Failures:        toProtoFailures(view.Failures),
	}), nil
}

// RefreshBalances reads again because the member asked.
func (s *BankingServer) RefreshBalances(
	ctx context.Context, req *connect.Request[bankingv1.RefreshBalancesRequest],
) (*connect.Response[bankingv1.RefreshBalancesResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	view, err := s.banking.Accounts(ctx, m.ID, false)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.RefreshBalancesResponse{
		Accounts:        toProtoGroupedAccounts(view, s.names(ctx)),
		HouseholdTotals: toProtoTotals(view.HouseholdTotals),
		OwnTotals:       toProtoTotals(view.OwnTotals),
		Failures:        toProtoFailures(view.Failures),
	}), nil
}

// RestoreConnection re-enters the hand-off for access that has run out.
func (s *BankingServer) RestoreConnection(
	ctx context.Context, req *connect.Request[bankingv1.RestoreConnectionRequest],
) (*connect.Response[bankingv1.RestoreConnectionResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	// The reason is what the screen said before sending the member; the flow
	// itself is identical, which is the point of threading it rather than
	// building a second one.
	handoff, err := s.banking.RestoreConnection(ctx, m.ID, req.Msg.GetConnectionId())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.RestoreConnectionResponse{
		HandoffUrl:       handoff.URL,
		ConsentExpiresAt: timestamppb.New(handoff.ConsentExpiresAt),
	}), nil
}

// DisconnectBank removes a bank.
func (s *BankingServer) DisconnectBank(
	ctx context.Context, req *connect.Request[bankingv1.DisconnectBankRequest],
) (*connect.Response[bankingv1.DisconnectBankResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := s.banking.Disconnect(ctx, m.ID, req.Msg.GetConnectionId()); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.DisconnectBankResponse{}), nil
}

// ListTransactions returns one page of the ledger, having first brought the
// member's accounts up to date.
func (s *BankingServer) ListTransactions(
	ctx context.Context, req *connect.Request[bankingv1.ListTransactionsRequest],
) (*connect.Response[bankingv1.ListTransactionsResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	ledger, err := s.banking.Transactions(ctx, banking.LedgerRequest{
		MemberID:  m.ID,
		AccountID: req.Msg.GetAccountId(),
		Cursor:    fromProtoCursor(req.Msg.GetCursor()),
		Older:     req.Msg.GetOlder(),
		Oldest:    req.Msg.GetOldest(),
		PageStart: fromProtoPageStart(req.Msg.GetPageStart()),
		SkipSync:  req.Msg.GetSkipSync(),
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.ListTransactionsResponse{Ledger: toProtoLedger(ledger)}), nil
}

// RefreshTransactions asks the banks again because the member asked.
func (s *BankingServer) RefreshTransactions(
	ctx context.Context, req *connect.Request[bankingv1.RefreshTransactionsRequest],
) (*connect.Response[bankingv1.RefreshTransactionsResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	// A refresh always returns the newest page. A member asking for the list to
	// be brought up to date is asking to see what just arrived, and what just
	// arrived is at the newest end.
	ledger, err := s.banking.Transactions(ctx, banking.LedgerRequest{
		MemberID: m.ID, AccountID: req.Msg.GetAccountId(), Refresh: true,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.RefreshTransactionsResponse{Ledger: toProtoLedger(ledger)}), nil
}

// GetBalanceTrend returns the calling member's balance trend, per currency.
// Reads no bank.
func (s *BankingServer) GetBalanceTrend(
	ctx context.Context, req *connect.Request[bankingv1.GetBalanceTrendRequest],
) (*connect.Response[bankingv1.GetBalanceTrendResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	trend, err := s.banking.BalanceTrend(ctx, m.ID, fromProtoScope(req.Msg.GetScope()))
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&bankingv1.GetBalanceTrendResponse{
		Trends:          toProtoTrends(trend.Currencies),
		PartialCoverage: trend.PartialCoverage,
	}), nil
}

// GetMonthSummary returns the calling member's month so far per currency, and
// where the most money went. Reads no bank.
func (s *BankingServer) GetMonthSummary(
	ctx context.Context, req *connect.Request[bankingv1.GetMonthSummaryRequest],
) (*connect.Response[bankingv1.GetMonthSummaryResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	summary, err := s.banking.MonthSummary(ctx, m.ID, fromProtoScope(req.Msg.GetScope()))
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(toProtoMonthSummary(summary)), nil
}

// GetMonthHistory returns the calling member's months, recurring payments and
// unusual payments within a scope. Reads no bank.
func (s *BankingServer) GetMonthHistory(
	ctx context.Context, req *connect.Request[bankingv1.GetMonthHistoryRequest],
) (*connect.Response[bankingv1.GetMonthHistoryResponse], error) {
	m, err := s.member(ctx, req.Header())
	if err != nil {
		return nil, toConnectError(err)
	}

	history, err := s.banking.MonthHistory(ctx, m.ID, fromProtoScope(req.Msg.GetScope()))
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(toProtoHistory(history)), nil
}

func toProtoHistory(h banking.History) *bankingv1.GetMonthHistoryResponse {
	out := &bankingv1.GetMonthHistoryResponse{
		Scope:               toProtoScope(h.Scope),
		HouseholdCounted:    h.Counted,
		HouseholdNotCounted: h.NotCounted,
	}
	for _, s := range h.Available {
		out.Available = append(out.Available, toProtoScope(s))
	}
	for _, c := range h.Currencies {
		// With a scope control on offer a scope's sections stay in place, and say
		// why they are empty, rather than vanishing.
		if len(h.Available) == 0 && c.FullMonths == 0 && len(c.Recurring) == 0 {
			continue
		}
		out.Histories = append(out.Histories, toProtoCurrencyHistory(c, h.Labels))
	}
	return out
}

func toProtoCurrencyHistory(c banking.CurrencyHistory, labels map[string]store.AccountLabel) *bankingv1.CurrencyHistory {
	out := &bankingv1.CurrencyHistory{
		Currency:        c.Currency,
		FullMonths:      int32(c.FullMonths),
		TypicalMonths:   int32(c.TypicalMonths),
		TypicalNet:      toProtoMoneyPtr(c.TypicalNet),
		AverageNet:      toProtoMoneyPtr(c.AverageNet),
		TypicalNetUsual: toProtoMoneyPtr(c.TypicalNetUsual),
		AverageNetUsual: toProtoMoneyPtr(c.AverageNetUsual),
		UnusualCount:    int32(c.UnusualCount),
	}
	for _, m := range c.Months {
		month := &bankingv1.HistoryMonth{
			MonthStart: timestamppb.New(m.Start),
			In:         toProtoMoney(m.In),
			Out:        toProtoMoney(m.Out),
			Net:        toProtoMoney(m.Net),
			NetUsual:   toProtoMoneyPtr(m.NetUsual),
			SoFar:      m.SoFar,

			TransfersLeftOut: int32(m.TransfersLeftOut),
		}
		if m.TransfersLeftOut > 0 {
			month.TransfersTotal = toProtoMoney(m.TransfersTotal)
		}
		if m.HeldFrom != nil {
			month.HeldFrom = timestamppb.New(*m.HeldFrom)
		}
		for _, r := range m.Risers {
			month.Risers = append(month.Risers, &bankingv1.MerchantRise{
				Name: r.Name, Total: toProtoMoney(r.Total), Usual: toProtoMoney(r.Usual), Payments: int32(r.Payments),
			})
		}
		for _, u := range m.Unusual {
			row := toProtoTransaction(u.Transaction, labels[u.Transaction.AccountID])
			row.Unusual = true
			month.Unusual = append(month.Unusual, &bankingv1.UnusualPayment{
				Transaction: row, Typical: toProtoMoneyPtr(u.Typical), FirstPayment: u.FirstPayment,
			})
		}
		out.Months = append(out.Months, month)
	}
	for _, l := range c.LateLedgers {
		out.LateLedgers = append(out.LateLedgers, &bankingv1.LateLedger{
			AccountName: l.AccountName, From: timestamppb.New(l.From),
		})
	}
	for _, r := range c.Recurring {
		out.Recurring = append(out.Recurring, &bankingv1.RecurringPayment{
			Name:      r.Name,
			Amount:    &bankingv1.Money{Minor: r.Amount, Currency: r.Currency},
			Cadence:   toProtoCadence(r.Cadence),
			Likely:    r.Likely,
			Expected:  timestamppb.New(r.Expected),
			Late:      r.Late,
			AccountId: r.AccountID,
		})
	}
	return out
}

func toProtoCadence(c banking.Cadence) bankingv1.Cadence {
	switch c {
	case banking.CadenceWeekly:
		return bankingv1.Cadence_CADENCE_WEEKLY
	case banking.CadenceMonthly:
		return bankingv1.Cadence_CADENCE_MONTHLY
	case banking.CadenceYearly:
		return bankingv1.Cadence_CADENCE_YEARLY
	}
	return bankingv1.Cadence_CADENCE_UNSPECIFIED
}

func fromProtoScope(s bankingv1.InsightScope) banking.Scope {
	switch s {
	case bankingv1.InsightScope_INSIGHT_SCOPE_HOUSEHOLD:
		return banking.ScopeHousehold
	case bankingv1.InsightScope_INSIGHT_SCOPE_OWN:
		return banking.ScopeOwn
	}
	return banking.ScopeAll
}

func toProtoScope(s banking.Scope) bankingv1.InsightScope {
	switch s {
	case banking.ScopeHousehold:
		return bankingv1.InsightScope_INSIGHT_SCOPE_HOUSEHOLD
	case banking.ScopeOwn:
		return bankingv1.InsightScope_INSIGHT_SCOPE_OWN
	}
	return bankingv1.InsightScope_INSIGHT_SCOPE_ALL
}

func toProtoMoney(m banking.Money) *bankingv1.Money {
	return &bankingv1.Money{Minor: m.Minor, Currency: m.Currency}
}

func toProtoMoneyPtr(m *banking.Money) *bankingv1.Money {
	if m == nil {
		return nil
	}
	return toProtoMoney(*m)
}

func toProtoMonthSummary(summary banking.MonthSummary) *bankingv1.GetMonthSummaryResponse {
	out := &bankingv1.GetMonthSummaryResponse{Months: make([]*bankingv1.CurrencyMonth, 0, len(summary.Months))}
	for _, c := range summary.Months {
		month := &bankingv1.CurrencyMonth{
			Currency:   c.Currency,
			In:         toProtoMoney(c.In),
			Out:        toProtoMoney(c.Out),
			Net:        toProtoMoney(c.Net),
			PriorIn:    toProtoMoneyPtr(c.PriorIn),
			PriorOut:   toProtoMoneyPtr(c.PriorOut),
			PriorNet:   toProtoMoneyPtr(c.PriorNet),
			MonthStart: timestamppb.New(c.MonthStart),
		}
		if c.CountedFrom != nil {
			month.CountedFrom = timestamppb.New(*c.CountedFrom)
		}
		for _, m := range c.TopMerchants {
			month.TopMerchants = append(month.TopMerchants, &bankingv1.MerchantTotal{
				Name: m.Name, Total: toProtoMoney(m.Total), Payments: int32(m.Payments),
			})
		}
		for _, t := range c.LargestPayments {
			month.LargestPayments = append(month.LargestPayments, toProtoTransaction(t, summary.Labels[t.AccountID]))
		}
		out.Months = append(out.Months, month)
	}
	return out
}

func fromProtoCursor(c *bankingv1.LedgerCursor) store.Cursor {
	if c == nil || c.GetTransactionId() == "" {
		return store.Cursor{}
	}
	return store.Cursor{BookingDate: c.GetBookingDate().AsTime(), ID: c.GetTransactionId()}
}

func fromProtoPageStart(c *bankingv1.LedgerCursor) *store.Cursor {
	if c == nil || c.GetTransactionId() == "" {
		return nil
	}
	cursor := store.Cursor{BookingDate: c.GetBookingDate().AsTime(), ID: c.GetTransactionId()}
	return &cursor
}

func toProtoCursor(c store.Cursor) *bankingv1.LedgerCursor {
	return &bankingv1.LedgerCursor{
		BookingDate:   timestamppb.New(c.BookingDate),
		TransactionId: c.ID,
	}
}

// toProtoPages renders the ledger's page index, newest first, the way
// LedgerPageIndex already ordered it.
func toProtoPages(pages []store.PageMarker) []*bankingv1.PageMarker {
	out := make([]*bankingv1.PageMarker, 0, len(pages))
	for _, p := range pages {
		out = append(out, &bankingv1.PageMarker{
			Cursor: toProtoCursor(p.Cursor),
			Newest: timestamppb.New(p.Newest),
			Oldest: timestamppb.New(p.Oldest),
		})
	}
	return out
}

// toProtoLedger renders one page. The account's name and its bank travel on
// every row, because the list is every account the member owns and a row has to
// say which one it came from.
func toProtoLedger(l banking.Ledger) *bankingv1.Ledger {
	out := &bankingv1.Ledger{
		TotalCount:        int32(l.Count),
		HasOlder:          l.Page.HasOlder,
		HasNewer:          l.Page.HasNewer,
		OwnsNoAccount:     l.OwnsNothing,
		Failures:          toProtoFailures(l.Failures),
		NarrowConnections: toProtoNarrow(l.Narrow),
		Pages:             toProtoPages(l.Pages),
	}

	for _, t := range l.Page.Transactions {
		row := toProtoTransaction(t, l.Accounts[t.AccountID])
		row.OwnTransfer = l.Patterns.IsTransfer(t.ID)
		_, row.Unusual = l.Patterns.MarkFor(t.ID)
		out.Transactions = append(out.Transactions, row)
	}

	// A span with no transactions in it is no span. Left absent rather than
	// rendered as the zero time, which would date an empty page to year one.
	if len(l.Page.Transactions) > 0 {
		out.OldestOnPage = timestamppb.New(l.Page.Oldest)
		out.NewestOnPage = timestamppb.New(l.Page.Newest)
	}
	if l.SyncedAt != nil {
		out.SyncedAt = timestamppb.New(*l.SyncedAt)
	}
	if l.ReachesBack != nil {
		out.ReachesBackTo = timestamppb.New(*l.ReachesBack)
	}
	return out
}

// toProtoTransaction carries both the derived name and the bank's line: a name
// wimm derived has to stay checkable against the statement it came from.
func toProtoTransaction(t store.Transaction, label store.AccountLabel) *bankingv1.Transaction {
	return &bankingv1.Transaction{
		Id:               t.ID,
		AccountId:        t.AccountID,
		AccountName:      label.Name,
		BankName:         label.BankName,
		Status:           toProtoTransactionStatus(t.Status),
		Amount:           &bankingv1.Money{Minor: t.AmountMinor, Currency: t.Currency},
		BookingDate:      timestamppb.New(t.BookingDate),
		CounterpartyName: t.CounterpartyName,
		Remittance:       t.Remittance,
		DisplayName:      banking.MerchantName(t.CounterpartyName, t.Remittance),
	}
}

func toProtoNarrow(narrow []store.NarrowConnection) []*bankingv1.NarrowConnection {
	out := make([]*bankingv1.NarrowConnection, 0, len(narrow))
	for _, n := range narrow {
		out = append(out, &bankingv1.NarrowConnection{ConnectionId: n.ConnectionID, BankName: n.BankName})
	}
	return out
}

func toProtoTransactionStatus(s store.TransactionStatus) bankingv1.TransactionStatus {
	if s == store.StatusPending {
		return bankingv1.TransactionStatus_TRANSACTION_STATUS_PENDING
	}
	return bankingv1.TransactionStatus_TRANSACTION_STATUS_BOOKED
}

// names maps member ids to something a person reads. Without it connected_by
// carries a uuid, and "connected by 0f8c3d2a-…" is not the answer to "who
// connected this bank".
type names map[string]string

func (n names) member(id string) *bankingv1.Member {
	if id == "" {
		return nil
	}
	return &bankingv1.Member{Id: id, DisplayName: n[id]}
}

// toProtoAccount renders one account at the level the member holds.
//
// **This is where redaction happens.** A field above the member's level is
// never serialised, so there is no client-side step that could be skipped and
// nothing above the level ever reaches the browser to be filtered there.
func toProtoAccount(a store.VisibleAccount, who names) *bankingv1.Account {
	out := &bankingv1.Account{
		Id:     a.ID,
		Source: toProtoSource(a.Source),
		Level:  toProtoLevel(a.Level),
		Owned:  a.Owned,
	}

	// Below balance, nothing about the account is rendered at all. A member
	// with no level never gets a row in the first place, so this is the
	// belt-and-braces case rather than the normal one.
	if a.Level != store.LevelBalance && a.Level != store.LevelDetails {
		return out
	}

	out.Name = a.Name
	out.HouseholdName = a.HouseholdName
	if a.LeftOutAt != nil {
		out.LeftOutAt = timestamppb.New(*a.LeftOutAt)
	}
	// Left out omits the balance here, in the handler, rather than in the
	// query: visibleAccountsQuery still returns it to an owner, and this is the
	// one place "what may this member see" is decided (ADR 0022).
	if a.LeftOutAt == nil && a.BalanceMinor != nil && a.BalanceReadAt != nil {
		out.Balance = &bankingv1.Balance{
			Money:  &bankingv1.Money{Minor: *a.BalanceMinor, Currency: a.Currency},
			ReadAt: timestamppb.New(*a.BalanceReadAt),
		}
	}
	if a.Connection != nil {
		out.Connection = &bankingv1.Connection{
			Id: a.Connection.ID, BankId: a.Connection.BankID,
			BankName: a.Connection.BankName, BankLogoUrl: a.Connection.BankLogoURL,
			ConnectedBy:      who.member(a.Connection.ConnectedBy),
			ConsentExpiresAt: timestamppb.New(a.Connection.ConsentExpiresAt),
			Live:             a.Connection.Live,
		}
	}

	// Identifying details only at details. This is the difference between "we
	// have €4,200 between us" and handing over an account number.
	if a.Level == store.LevelDetails {
		out.NumberSuffix = a.NumberSuffix
		out.AccountType = a.AccountType
		out.HolderName = a.HolderName
	}
	return out
}

func toProtoAccounts(accounts []store.VisibleAccount, who names) []*bankingv1.Account {
	out := make([]*bankingv1.Account, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, toProtoAccount(a, who))
	}
	return out
}

// toProtoConnection takes the connection from whichever account carries one.
func toProtoConnection(accounts []store.VisibleAccount, who names) *bankingv1.Connection {
	for _, a := range accounts {
		if a.Connection != nil {
			return toProtoAccount(a, who).GetConnection()
		}
	}
	return nil
}

// toProtoGroupedAccounts renders a member's view with each account's group,
// which wimmd decides and no client recomputes.
func toProtoGroupedAccounts(view banking.View, who names) []*bankingv1.Account {
	out := toProtoAccounts(view.Accounts, who)
	for _, a := range out {
		a.Group = toProtoGroup(view.Groups[a.GetId()])
	}
	return out
}

func toProtoGroup(g banking.Group) bankingv1.AccountGroup {
	switch g {
	case banking.GroupHousehold:
		return bankingv1.AccountGroup_ACCOUNT_GROUP_HOUSEHOLD
	case banking.GroupOwn:
		return bankingv1.AccountGroup_ACCOUNT_GROUP_OWN
	case banking.GroupShared:
		return bankingv1.AccountGroup_ACCOUNT_GROUP_SHARED
	default:
		return bankingv1.AccountGroup_ACCOUNT_GROUP_UNSPECIFIED
	}
}

func toProtoTotals(totals []banking.Total) []*bankingv1.CurrencyTotal {
	out := make([]*bankingv1.CurrencyTotal, 0, len(totals))
	for _, t := range totals {
		out = append(out, &bankingv1.CurrencyTotal{
			Total:        &bankingv1.Money{Minor: t.Money.Minor, Currency: t.Money.Currency},
			AccountCount: int32(t.AccountCount),
		})
	}
	return out
}

func toProtoTrends(trends []banking.CurrencyTrend) []*bankingv1.CurrencyTrend {
	out := make([]*bankingv1.CurrencyTrend, 0, len(trends))
	for _, t := range trends {
		points := make([]*bankingv1.TrendPoint, 0, len(t.Points))
		for _, p := range t.Points {
			point := &bankingv1.TrendPoint{
				Date:    timestamppb.New(p.Date),
				Balance: &bankingv1.Money{Minor: p.Money.Minor, Currency: p.Money.Currency},
				Smaller: int32(p.Smaller),
			}
			for _, m := range p.Movers {
				point.Movers = append(point.Movers, &bankingv1.DayMover{
					DisplayName: m.Name,
					Amount:      &bankingv1.Money{Minor: m.Amount, Currency: p.Money.Currency},
					Unusual:     m.Unusual,
					OwnTransfer: m.OwnTransfer,
				})
			}
			points = append(points, point)
		}
		out = append(out, &bankingv1.CurrencyTrend{Currency: t.Currency, Points: points, ShortHistory: t.ShortHistory})
	}
	return out
}

func toProtoFailures(failures []banking.BankFailure) []*bankingv1.BankFailure {
	out := make([]*bankingv1.BankFailure, 0, len(failures))
	for _, f := range failures {
		out = append(out, &bankingv1.BankFailure{
			ConnectionId:      f.ConnectionID,
			BankName:          f.BankName,
			Failure:           toProtoFailure(f.Err),
			RetryAfterSeconds: int64(f.RetryAfter.Seconds()),
		})
	}
	return out
}

// toProtoFailure maps wimm's taxonomy onto the wire. No gateway status code
// and no gateway prose reaches the browser.
func toProtoFailure(err error) bankingv1.Failure {
	switch {
	case errors.Is(err, banking.ErrBankUnavailable):
		return bankingv1.Failure_FAILURE_BANK_UNAVAILABLE
	case errors.Is(err, banking.ErrGatewayUnavailable):
		return bankingv1.Failure_FAILURE_SERVICE_UNAVAILABLE
	case errors.Is(err, banking.ErrConsentDeclined):
		return bankingv1.Failure_FAILURE_CONSENT_DECLINED
	case errors.Is(err, banking.ErrConsentExpired):
		return bankingv1.Failure_FAILURE_CONSENT_EXPIRED
	case errors.Is(err, banking.ErrNoAccounts):
		return bankingv1.Failure_FAILURE_NO_ACCOUNTS
	case errors.Is(err, banking.ErrRateLimited):
		return bankingv1.Failure_FAILURE_RATE_LIMITED
	default:
		return bankingv1.Failure_FAILURE_UNSPECIFIED
	}
}

func toProtoLevel(l store.Level) bankingv1.Level {
	switch l {
	case store.LevelBalance:
		return bankingv1.Level_LEVEL_BALANCE
	case store.LevelDetails:
		return bankingv1.Level_LEVEL_DETAILS
	default:
		return bankingv1.Level_LEVEL_HIDDEN
	}
}

// fromProtoLevel maps the wire onto the domain. UNSPECIFIED and HIDDEN both
// mean no grant: an unset field must end in the member seeing nothing, never in
// a refusal and never in "leave unchanged".
func fromProtoLevel(l bankingv1.Level) store.Level {
	switch l {
	case bankingv1.Level_LEVEL_BALANCE:
		return store.LevelBalance
	case bankingv1.Level_LEVEL_DETAILS:
		return store.LevelDetails
	default:
		return ""
	}
}

func toProtoSource(s store.Source) bankingv1.AccountSource {
	switch s {
	case store.SourceGateway:
		return bankingv1.AccountSource_ACCOUNT_SOURCE_GATEWAY
	case store.SourceManual:
		return bankingv1.AccountSource_ACCOUNT_SOURCE_MANUAL
	default:
		return bankingv1.AccountSource_ACCOUNT_SOURCE_UNSPECIFIED
	}
}
