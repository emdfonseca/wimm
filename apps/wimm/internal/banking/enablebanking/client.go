package enablebanking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// DefaultBaseURL is the production API. Overridden in tests and against the
// sandbox.
const DefaultBaseURL = "https://api.enablebanking.com"

// maxErrorBody caps what is read from a failed response before mapping it. A
// gateway having a bad day can return a very large HTML error page, and none of
// it is worth holding.
const maxErrorBody = 8 << 10

// Client is the Enable Banking implementation of banking.Gateway.
type Client struct {
	baseURL     string
	http        *http.Client
	signer      *signer
	redirectURL string

	// now is the wall clock, needed for the JWT's iat and for nothing else.
	// It is a field so tests mint deterministic tokens.
	now func() time.Time
}

// Options configures a Client. Every field is required except BaseURL, HTTP
// and Now.
type Options struct {
	ApplicationID  string
	PrivateKeyPath string
	RedirectURL    string
	BaseURL        string
	HTTP           *http.Client
	Now            func() time.Time
}

// New builds a Client, reading and validating the signing key.
func New(opts Options) (*Client, error) {
	s, err := newSigner(opts.ApplicationID, opts.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	if opts.RedirectURL == "" {
		return nil, errors.New("enablebanking: no redirect url")
	}

	c := &Client{
		baseURL:     strings.TrimRight(cmpOr(opts.BaseURL, DefaultBaseURL), "/"),
		http:        opts.HTTP,
		signer:      s,
		redirectURL: opts.RedirectURL,
		now:         opts.Now,
	}
	if c.http == nil {
		// A bank that never answers must not hold a request open forever: a
		// member is waiting on the other end of this.
		c.http = &http.Client{Timeout: 30 * time.Second}
	}
	if c.now == nil {
		c.now = time.Now
	}
	return c, nil
}

func cmpOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// Name is stored on every connection this client opens.
func (c *Client) Name() string { return "enablebanking" }

// request is one signed call. psuPresent adds the headers that tell the gateway
// a member is sitting in front of the screen, which is what exempts the call
// from the background-fetch rate limit (ADR 0018).
func (c *Client) request(
	ctx context.Context, method, path string, body, out any, psuPresent bool,
) error {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("enablebanking: encoding the request: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return fmt.Errorf("enablebanking: building the request: %w", err)
	}

	token, err := c.signer.token(c.now())
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if psuPresent {
		// The values matter less than their presence: they assert this call
		// was made because a member asked, not on a timer.
		req.Header.Set("PSU-IP-Address", "127.0.0.1")
		req.Header.Set("PSU-User-Agent", "wimm")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		// A transport failure is the gateway being unreachable, not a bank
		// refusing: the request never got far enough to find out.
		return fmt.Errorf("enablebanking: %s %s: %w", method, path, errors.Join(err, banking.ErrGatewayUnavailable))
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 300 {
		return c.mapFailure(resp, method, path)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("enablebanking: %s %s: decoding the response: %w", method, path, err)
	}
	return nil
}

// apiError is the gateway's error envelope. Only the code is used: the message
// is the gateway's prose and is never shown to a member.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// mapFailure turns a response into wimm's taxonomy. Status alone is not enough
// — 401 is both "our credentials are wrong" and "this bank session has expired"
// — so the body's code decides wherever it is present (ADR 0018).
func (c *Client) mapFailure(resp *http.Response, method, path string) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))

	var envelope apiError
	_ = json.Unmarshal(body, &envelope)
	code := strings.ToUpper(cmpOr(envelope.Code, envelope.Error))

	kind := classify(resp.StatusCode, code, retryAfter(resp.Header))

	// The gateway's own message is deliberately absent: it is prose in another
	// vocabulary, and everything the caller routes on is in kind.
	return fmt.Errorf("enablebanking: %s %s: %d %s: %w", method, path, resp.StatusCode, code, kind)
}

// classify is the whole status-to-taxonomy mapping, in one place and table
// driven so 4.6's cases are read rather than traced.
func classify(status int, code string, retry time.Duration) error {
	switch code {
	case "EXPIRED_SESSION", "SESSION_EXPIRED", "INVALID_SESSION":
		return banking.ErrConsentExpired
	case "ASPSP_RATE_LIMIT_EXCEEDED", "RATE_LIMIT_EXCEEDED", "TOO_MANY_REQUESTS":
		return banking.RateLimited(retry)
	case "ACCESS_DENIED", "CONSENT_DENIED", "USER_CANCELLED", "PSU_CANCELLED":
		return banking.ErrConsentDeclined
	case "ASPSP_ERROR", "ASPSP_UNAVAILABLE", "ASPSP_CONNECTION_ERROR":
		return banking.ErrBankUnavailable
	case "NO_ACCOUNTS", "ACCOUNTS_NOT_FOUND":
		return banking.ErrNoAccounts
	}

	switch {
	case status == http.StatusTooManyRequests:
		return banking.RateLimited(retry)
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		// With no code to narrow it, 401 at this layer means the grant is no
		// longer good. Treating it as a wimm misconfiguration would tell a
		// member nothing they can act on.
		return banking.ErrConsentExpired
	case status == http.StatusNotFound:
		return banking.ErrConsentExpired
	case status >= 500:
		// The gateway proxies the bank, so its 5xx is usually the bank's.
		return banking.ErrBankUnavailable
	default:
		return banking.ErrGatewayUnavailable
	}
}

// retryAfter reads the header in either form the RFC allows. A malformed value
// is no value: the member is told "later" rather than a wrong time.
func retryAfter(h http.Header) time.Duration {
	raw := strings.TrimSpace(h.Get("Retry-After"))
	if raw == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 0
}

// Banks lists the institutions connectable in a country.
func (c *Client) Banks(ctx context.Context, country string) ([]banking.Bank, error) {
	var response struct {
		ASPSPs []struct {
			Name     string   `json:"name"`
			Country  string   `json:"country"`
			Logo     string   `json:"logo"`
			PSUTypes []string `json:"psu_types"`
			// Seconds. Carried through to the domain type so the consent
			// screen can state a real date: it is 90 days at some banks in one
			// country and 1 day at others.
			MaximumConsentValidity int64 `json:"maximum_consent_validity"`
		} `json:"aspsps"`
	}

	path := "/aspsps"
	if country != "" {
		path += "?country=" + url.QueryEscape(strings.ToUpper(country))
	}
	if err := c.request(ctx, http.MethodGet, path, nil, &response, false); err != nil {
		return nil, err
	}

	banks := make([]banking.Bank, 0, len(response.ASPSPs))
	for _, a := range response.ASPSPs {
		if !offersPersonalAccounts(a.PSUTypes) {
			continue
		}
		banks = append(banks, banking.Bank{
			ID:         bankID(a.Name, a.Country),
			Name:       a.Name,
			Country:    a.Country,
			LogoURL:    a.Logo,
			MaxConsent: time.Duration(a.MaximumConsentValidity) * time.Second,
		})
	}
	return banks, nil
}

// This gateway identifies a bank by name and country rather than by an id, so
// wimm's Bank.ID is the pair. It is stored on connections, so the encoding has
// to be stable and reversible.
func bankID(name, country string) string { return country + ":" + name }

func splitBankID(id string) (name, country string) {
	country, name, _ = strings.Cut(id, ":")
	return name, country
}

// A bank offering only business accounts is not connectable by a household, and
// listing it produces a hand-off that fails at the bank.
func offersPersonalAccounts(types []string) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if strings.EqualFold(t, "personal") {
			return true
		}
	}
	return false
}

// BeginConnection asks for somewhere to send the member.
func (c *Client) BeginConnection(ctx context.Context, req banking.BeginRequest) (banking.Handoff, error) {
	name, country := splitBankID(req.Bank.ID)

	body := map[string]any{
		"access": map[string]any{
			// Both, and only these. Transactions are a separate scope at the
			// bank, so a connection granted before this asked for balances
			// alone and no reading of it produces transactions — widening is a
			// member confirming again (ADR 0021).
			"balances":     true,
			"transactions": true,
			"valid_until":  req.ValidUntil.UTC().Format(time.RFC3339),
		},
		"aspsp":        map[string]string{"name": name, "country": country},
		"state":        req.State,
		"redirect_url": cmpOr(req.RedirectURL, c.redirectURL),
		"psu_type":     "personal",
	}

	var response struct {
		URL             string `json:"url"`
		AuthorizationID string `json:"authorization_id"`
	}
	if err := c.request(ctx, http.MethodPost, "/auth", body, &response, true); err != nil {
		return banking.Handoff{}, err
	}
	if response.URL == "" {
		return banking.Handoff{}, fmt.Errorf("enablebanking: the gateway returned no url to send the member to: %w", banking.ErrGatewayUnavailable)
	}
	return banking.Handoff{URL: response.URL, GatewayRef: response.AuthorizationID}, nil
}

// CompleteConnection exchanges the member's return for live access. The
// accounts come back here and cannot be listed again, which is why every one of
// them is stored.
func (c *Client) CompleteConnection(
	ctx context.Context, pending banking.PendingConnection, cb banking.Callback,
) (banking.Connection, []banking.Account, error) {
	if cb.Error != "" {
		return banking.Connection{}, nil, fmt.Errorf("enablebanking: %s: %w", cb.Error, banking.ErrConsentDeclined)
	}
	if cb.State != pending.State {
		return banking.Connection{}, nil, fmt.Errorf(
			"enablebanking: the callback state does not match the pending connection: %w", banking.ErrGatewayUnavailable)
	}

	var response struct {
		SessionID string `json:"session_id"`
		Accounts  []struct {
			UID                string                `json:"uid"`
			IdentificationHash string                `json:"identification_hash"`
			Name               string                `json:"name"`
			Product            string                `json:"product"`
			Currency           string                `json:"currency"`
			CashAccountType    string                `json:"cash_account_type"`
			AccountID          accountIdentification `json:"account_id"`
		} `json:"accounts"`
		Access struct {
			ValidUntil string `json:"valid_until"`
		} `json:"access"`
	}

	body := map[string]string{"code": cb.Code}
	if err := c.request(ctx, http.MethodPost, "/sessions", body, &response, true); err != nil {
		return banking.Connection{}, nil, err
	}
	if len(response.Accounts) == 0 {
		return banking.Connection{}, nil, fmt.Errorf("enablebanking: %w", banking.ErrNoAccounts)
	}

	accounts := make([]banking.Account, 0, len(response.Accounts))
	for _, a := range response.Accounts {
		accounts = append(accounts, banking.Account{
			// The cross-session hash, never the uid: the uid is reissued on
			// every authorisation and keying on it would lose every sharing
			// choice the first time a member restored.
			Ref:          a.IdentificationHash,
			GatewayUID:   a.UID,
			Name:         cmpOr(a.Product, a.Name),
			NumberSuffix: suffix(a.AccountID.identifier()),
			Type:         a.CashAccountType,
			HolderName:   a.Name,
			Currency:     strings.ToUpper(a.Currency),
		})
	}

	conn := banking.Connection{GatewayRef: response.SessionID}
	if parsed, err := time.Parse(time.RFC3339, response.Access.ValidUntil); err == nil {
		conn.ExpiresAt = parsed
	}
	return conn, accounts, nil
}

// accountIdentification is how a bank names an account.
//
// `other` is an **object**, not a string: an account with no IBAN — a card
// account, which is how Montepio exposes one of these — is identified by a
// generic identification instead. Declaring it as a string decoded every
// IBAN-bearing account fine and failed the whole connection on the first card
// account, with the shape mismatch buried in an Internal error.
//
// The fields not read are still named, because a reader comparing this against
// a live response needs to see that they were looked at and skipped.
type accountIdentification struct {
	IBAN  string `json:"iban"`
	Other struct {
		Identification string `json:"identification"`
		SchemeName     string `json:"scheme_name"`
		Issuer         string `json:"issuer"`
	} `json:"other"`
}

// identifier is the number to take a suffix from: the IBAN where there is one,
// and the generic identification otherwise.
func (a accountIdentification) identifier() string {
	return cmpOr(a.IBAN, a.Other.Identification)
}

// suffixLength is enough to tell two accounts at one bank apart and no more.
// The full number is never stored: an IBAN at rest is a liability with no use
// here (ADR 0018).
const suffixLength = 4

func suffix(number string) string {
	number = strings.TrimSpace(number)
	if len(number) <= suffixLength {
		return number
	}
	return number[len(number)-suffixLength:]
}

// Balances reads one account, always as a member-present call.
func (c *Client) Balances(
	ctx context.Context, conn banking.Connection, account banking.Account,
) ([]banking.Balance, error) {
	var response struct {
		Balances []struct {
			Name          string `json:"name"`
			BalanceType   string `json:"balance_type"`
			BalanceAmount struct {
				Amount   string `json:"amount"`
				Currency string `json:"currency"`
			} `json:"balance_amount"`
			ReferenceDate      string `json:"reference_date"`
			LastChangeDateTime string `json:"last_change_date_time"`
		} `json:"balances"`
	}

	path := "/accounts/" + url.PathEscape(account.GatewayUID) + "/balances"
	if err := c.request(ctx, http.MethodGet, path, nil, &response, true); err != nil {
		return nil, err
	}

	readAt := c.now().UTC()
	balances := make([]banking.Balance, 0, len(response.Balances))
	for _, b := range response.Balances {
		money, err := parseAmount(b.BalanceAmount.Amount, cmpOr(b.BalanceAmount.Currency, account.Currency))
		if err != nil {
			// An amount wimm cannot represent exactly is not shown as an
			// approximation. The account keeps its previous reading.
			return nil, fmt.Errorf("enablebanking: balance for account %s: %w", account.Ref, err)
		}
		balances = append(balances, banking.Balance{
			Money: money,
			// The time wimm read it, not a date the bank attached: the
			// contract shown to a member is "when did wimm ask".
			ReadAt: readAt,
			Kind:   cmpOr(b.BalanceType, b.Name),
		})
	}
	if len(balances) == 0 {
		return nil, fmt.Errorf("enablebanking: the bank returned no balance for account %s: %w", account.Ref, banking.ErrBankUnavailable)
	}
	return balances, nil
}

// transactionsResponse is one page as the gateway returns it. The fields not
// read are still named, because a reader comparing this against a live response
// needs to see that they were looked at and skipped.
type transactionsResponse struct {
	Transactions []struct {
		EntryReference    string `json:"entry_reference"`
		TransactionID     string `json:"transaction_id"`
		TransactionAmount struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"transaction_amount"`
		CreditDebitIndicator  string   `json:"credit_debit_indicator"`
		Status                string   `json:"status"`
		BookingDate           string   `json:"booking_date"`
		ValueDate             string   `json:"value_date"`
		TransactionDate       string   `json:"transaction_date"`
		Creditor              party    `json:"creditor"`
		Debtor                party    `json:"debtor"`
		RemittanceInformation []string `json:"remittance_information"`
		BankTransactionCode   struct {
			Description string `json:"description"`
		} `json:"bank_transaction_code"`
	} `json:"transactions"`
	ContinuationKey string `json:"continuation_key"`
}

type party struct {
	Name string `json:"name"`
}

// Transactions reads one page of one account's transactions, always as a
// member-present call.
//
// The two strategies are the reason this is one method and not two. A zero From
// means "as far back as this bank goes", which is strategy=longest: the gateway
// finds the earliest transaction available and fetches forward from it, and
// date_to is ignored. A set From is strategy=default with a date_from, which is
// what the changelog recommends for updating a feed already fetched — and which
// answers WRONG_TRANSACTIONS_PERIOD when the window is unavailable.
func (c *Client) Transactions(
	ctx context.Context, conn banking.Connection, account banking.Account, req banking.TransactionsRequest,
) (banking.TransactionsPage, error) {
	page, err := c.transactionsPage(ctx, account, req)
	if err == nil || req.From.IsZero() || !wrongTransactionsPeriod(err) {
		return page, err
	}

	// Retried once, here, with the widest strategy the gateway has — and only
	// from a dated request, because retrying longest with longest asks the same
	// question twice. A member cannot act on "the window you asked for is
	// unavailable", so it never becomes one of their failures and the taxonomy
	// does not grow; a second refusal surfaces mapped like any other.
	return c.transactionsPage(ctx, account, banking.TransactionsRequest{Cursor: req.Cursor})
}

func (c *Client) transactionsPage(
	ctx context.Context, account banking.Account, req banking.TransactionsRequest,
) (banking.TransactionsPage, error) {
	query := url.Values{}
	if req.From.IsZero() {
		query.Set("strategy", "longest")
	} else {
		query.Set("strategy", "default")
		query.Set("date_from", req.From.UTC().Format(time.DateOnly))
	}
	if req.Cursor != "" {
		query.Set("continuation_key", req.Cursor)
	}

	path := "/accounts/" + url.PathEscape(account.GatewayUID) + "/transactions?" + query.Encode()

	var response transactionsResponse
	if err := c.request(ctx, http.MethodGet, path, nil, &response, true); err != nil {
		return banking.TransactionsPage{}, err
	}

	out := make([]banking.Transaction, 0, len(response.Transactions))
	for _, t := range response.Transactions {
		money, err := parseAmount(t.TransactionAmount.Amount, cmpOr(t.TransactionAmount.Currency, account.Currency))
		if err != nil {
			// An amount wimm cannot represent exactly is not stored as an
			// approximation, for the reason a balance is not.
			return banking.TransactionsPage{}, fmt.Errorf(
				"enablebanking: transaction on account %s: %w", account.Ref, err)
		}
		if debit(t.CreditDebitIndicator) {
			money.Minor = -abs(money.Minor)
		} else {
			money.Minor = abs(money.Minor)
		}

		out = append(out, banking.Transaction{
			Ref:    t.EntryReference,
			Status: transactionStatus(t.Status),
			Amount: money,
			// Booking is what the ledger orders and groups by; the other two
			// are kept because they are what a member sometimes means by
			// "when". A date the bank omitted stays zero rather than becoming
			// today.
			BookingDate:     parseDate(t.BookingDate),
			ValueDate:       parseDate(t.ValueDate),
			TransactionDate: parseDate(t.TransactionDate),
			// The other party is whichever end wimm's account is not, and the
			// indicator is what says which. Where the bank names neither, the
			// field stays empty and the screen shows what it did give.
			CounterpartyName: counterparty(t.CreditDebitIndicator, t.Creditor.Name, t.Debtor.Name),
			Remittance: cmpOr(
				strings.TrimSpace(strings.Join(t.RemittanceInformation, " ")),
				t.BankTransactionCode.Description),
		})
	}

	return banking.TransactionsPage{Transactions: out, NextCursor: response.ContinuationKey}, nil
}

// wrongTransactionsPeriod reports the one gateway code this adapter handles
// itself. It is matched on the error's own text because mapFailure puts the
// code there and maps the taxonomy member separately: a code with no taxonomy
// meaning would otherwise be indistinguishable from any other 4xx.
func wrongTransactionsPeriod(err error) bool {
	return strings.Contains(err.Error(), "WRONG_TRANSACTIONS_PERIOD")
}

// BOOK is settled, PEND is not — PEND, not PDNG. OTHR is filed as booked,
// because a transaction wimm cannot classify has already moved money and hiding
// it is worse than filing it (ADR 0021).
func transactionStatus(s string) banking.TransactionStatus {
	if strings.EqualFold(strings.TrimSpace(s), "PEND") {
		return banking.StatusPending
	}
	return banking.StatusBooked
}

func debit(indicator string) bool {
	return strings.EqualFold(strings.TrimSpace(indicator), "DBIT")
}

func counterparty(indicator, creditor, debtor string) string {
	if debit(indicator) {
		return cmpOr(creditor, debtor)
	}
	return cmpOr(debtor, creditor)
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// parseDate reads a gateway date. A date the bank omitted or malformed stays
// the zero time: inventing today would put a transaction on the wrong day, and
// the day is what the ledger groups by.
func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if d, err := time.Parse(time.DateOnly, s); err == nil {
		return d
	}
	if d, err := time.Parse(time.RFC3339, s); err == nil {
		return d.UTC().Truncate(24 * time.Hour)
	}
	return time.Time{}
}

// EndConnection tells the bank wimm is done.
func (c *Client) EndConnection(ctx context.Context, conn banking.Connection) error {
	if conn.GatewayRef == "" {
		return nil
	}
	err := c.request(ctx, http.MethodDelete, "/sessions/"+url.PathEscape(conn.GatewayRef), nil, nil, true)

	// A session the gateway has already forgotten is a disconnection that has
	// already happened. Reporting it as a failure would leave a member unable
	// to remove a bank that is, as far as anyone can tell, already gone.
	if errors.Is(err, banking.ErrConsentExpired) {
		return nil
	}
	return err
}

var _ banking.Gateway = (*Client)(nil)
