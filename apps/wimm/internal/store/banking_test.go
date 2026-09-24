package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
	"github.com/emdfonseca/wimm/apps/wimm/internal/store/storetest"
)

func sealed(v string) store.Sealed {
	return store.Sealed{Ciphertext: []byte("sealed:" + v), KeyID: "k1"}
}

// connect records a live connection with the given accounts, owned by owner.
func connect(
	t *testing.T, ctx context.Context, db *store.DB, owner string, consent time.Duration, accounts ...store.Account,
) (store.BankConnection, []store.Account) {
	t.Helper()

	now, err := db.Now(ctx)
	if err != nil {
		t.Fatalf("reading database time: %v", err)
	}

	conn, stored, err := db.CreateBankConnection(ctx, store.BankConnection{
		Gateway:          "enablebanking",
		GatewayRef:       sealed("session"),
		BankID:           "PT:Montepio",
		BankName:         "Montepio",
		ConnectedBy:      owner,
		ConsentExpiresAt: now.T.Add(consent),
	}, accounts, owner)
	if err != nil {
		t.Fatalf("recording a connection: %v", err)
	}
	return conn, stored
}

func account(ref, name, suffix string) store.Account {
	return store.Account{
		GatewayRef: ref, GatewayUID: sealed(ref), Name: name,
		NumberSuffix: suffix, AccountType: "CACC", HolderName: "Ada Lovelace", Currency: "EUR",
	}
}

// The distinction 5.3 names: a return that was already exchanged is not the
// same answer as one that never existed, and a member reloading the page they
// landed on must not be told their connection is unknown.
func TestAConsumedPendingConnectionIsNotTheSameAsAnUnknownOne(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	m := member(t, ctx, db, "ada@example.com")

	hash := []byte("state-hash-one")
	if _, err := db.CreatePendingBankConnection(ctx, store.PendingBankConnection{
		Gateway: "enablebanking", GatewayRef: "auth-1", BankID: "PT:Montepio",
		BankName: "Montepio", RedirectURL: "https://localhost:8765/psd2/callback", StartedBy: m.ID,
	}, hash, time.Hour); err != nil {
		t.Fatalf("recording a pending connection: %v", err)
	}

	if _, err := db.ConsumePendingBankConnection(ctx, hash); err != nil {
		t.Fatalf("first consume: %v", err)
	}

	_, err := db.ConsumePendingBankConnection(ctx, hash)
	if !errors.Is(err, store.ErrPendingConnectionSpent) {
		t.Errorf("a replayed return gave %v, want ErrPendingConnectionSpent", err)
	}

	_, err = db.ConsumePendingBankConnection(ctx, []byte("never issued"))
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("an unknown return gave %v, want ErrNotFound", err)
	}
}

// Expiry is measured against database time, so a pending connection past its
// expiry cannot be consumed however the process clock is set.
func TestAnExpiredPendingConnectionCannotBeConsumed(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	m := member(t, ctx, db, "ada@example.com")

	hash := []byte("state-hash-expired")
	if _, err := db.CreatePendingBankConnection(ctx, store.PendingBankConnection{
		Gateway: "enablebanking", GatewayRef: "auth-1", BankID: "PT:Montepio",
		BankName: "Montepio", RedirectURL: "https://x", StartedBy: m.ID,
	}, hash, -time.Minute); err != nil {
		t.Fatalf("recording a pending connection: %v", err)
	}

	if _, err := db.ConsumePendingBankConnection(ctx, hash); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound for an expired row", err)
	}
}

// Every account the session returns is stored, because they are returned once.
// The connecting member owns them all and can then disown.
func TestEveryAccountIsStoredAndOwnedByTheConnectingMember(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour,
		account("hash-1", "Conta à Ordem", "0538"),
		account("hash-2", "Poupança", "5594"))

	if len(stored) != 2 {
		t.Fatalf("stored %d accounts, want 2", len(stored))
	}
	for _, a := range stored {
		owners, err := db.AccountOwners(ctx, a.ID)
		if err != nil {
			t.Fatalf("reading owners: %v", err)
		}
		if len(owners) != 1 || owners[0] != ada.ID {
			t.Errorf("%s owners = %v, want just the connecting member", a.Name, owners)
		}
	}
}

// The visibility contract, end to end through the query that decides it.
func TestWhatEachMemberMaySee(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour,
		account("hash-1", "Joint", "0538"),
		account("hash-2", "Personal", "5594"),
		account("hash-3", "Savings", "7712"))
	joint, personal := stored[0], stored[2]
	_ = personal

	if err := db.SetAccountLevel(ctx, joint.ID, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatalf("granting: %v", err)
	}

	adaSees, err := db.VisibleAccounts(ctx, ada.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(adaSees) != 3 {
		t.Errorf("the owner sees %d accounts, want all 3", len(adaSees))
	}
	for _, a := range adaSees {
		if !a.Owned || a.Level != store.LevelDetails {
			t.Errorf("%s: owned=%v level=%q, want an owner to see it in full", a.Name, a.Owned, a.Level)
		}
	}

	graceSees, err := db.VisibleAccounts(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(graceSees) != 1 {
		t.Fatalf("the granted member sees %d accounts, want 1", len(graceSees))
	}
	if graceSees[0].ID != joint.ID {
		t.Errorf("sees %q, want the joint account", graceSees[0].Name)
	}
	if graceSees[0].Owned {
		t.Error("a granted member is reported as an owner")
	}
	if graceSees[0].Level != store.LevelBalance {
		t.Errorf("level = %q, want balance", graceSees[0].Level)
	}
}

// A member with nothing gets no rows at all — not a redacted row, and not a
// count of what was withheld.
func TestAMemberWithNoOwnershipAndNoGrantSeesNothing(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Personal", "0538"))

	seen, err := db.VisibleAccounts(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(seen) != 0 {
		t.Errorf("sees %d accounts, want none", len(seen))
	}
}

// Ownership and a grant on one account are mutually exclusive, refused rather
// than resolved.
func TestAnOwnerCannotAlsoHoldAGrant(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	err := db.SetAccountLevel(ctx, stored[0].ID, ada.ID, store.LevelBalance, ada.ID)
	if !errors.Is(err, store.ErrOwnerHoldsAGrant) {
		t.Errorf("got %v, want ErrOwnerHoldsAGrant", err)
	}
}

// Handing an account over removes the grant the new owner held, so the two
// never coexist even by that route.
func TestMakingAGranteeAnOwnerClearsTheirGrant(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))
	id := stored[0].ID

	if err := db.SetAccountLevel(ctx, id, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatalf("granting: %v", err)
	}
	if err := db.SetAccountOwners(ctx, id, []string{ada.ID, grace.ID}); err != nil {
		t.Fatalf("setting owners: %v", err)
	}

	seen, err := db.VisibleAccounts(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(seen) != 1 || !seen[0].Owned || seen[0].Level != store.LevelDetails {
		t.Errorf("after being made an owner: %+v", seen)
	}
}

// An account always has an owner (ADR 0022): disowning it down to zero is
// refused rather than producing the unreadable orphan ADR 0019 used to allow.
// Leaving an account out, once that exists, is how this case is now handled.
func TestDisowningAnAccountToZeroOwnersIsRefused(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	conn, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour,
		account("hash-1", "Joint", "0538"),
		account("hash-2", "Personal", "5594"))

	if err := db.SetAccountOwners(ctx, stored[1].ID, nil); !errors.Is(err, store.ErrAccountWouldHaveNoOwner) {
		t.Fatalf("disowning an account's only owner gave %v, want ErrAccountWouldHaveNoOwner", err)
	}

	readable, err := db.ReadableAccounts(ctx, conn.ID)
	if err != nil {
		t.Fatalf("ReadableAccounts: %v", err)
	}
	if len(readable) != 2 {
		t.Errorf("readable = %d accounts after a refused disowning, want both untouched", len(readable))
	}
}

// A left-out account is read by nobody and appears for its owners and for
// nobody else (ADR 0022): a grant survives being left out, but the grantee
// half of visibleAccountsQuery's join is gated on left_out_at, and the owner
// half is not.
func TestALeftOutAccountIsReadableByNobodyAndVisibleOnlyToItsOwner(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	conn, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Personal", "0538"))
	acc := stored[0]

	if err := db.SetAccountLevel(ctx, acc.ID, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatalf("granting: %v", err)
	}
	if err := db.SetAccountLeftOut(ctx, acc.ID, true); err != nil {
		t.Fatalf("SetAccountLeftOut: %v", err)
	}

	readable, err := db.ReadableAccounts(ctx, conn.ID)
	if err != nil {
		t.Fatalf("ReadableAccounts: %v", err)
	}
	if len(readable) != 0 {
		t.Errorf("a left-out account is readable by %d accounts' worth, want none", len(readable))
	}

	adaSees, err := db.VisibleAccounts(ctx, ada.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts(owner): %v", err)
	}
	if len(adaSees) != 1 || adaSees[0].LeftOutAt == nil {
		t.Errorf("the owner does not see the left-out account it owns: %+v", adaSees)
	}

	graceSees, err := db.VisibleAccounts(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts(grantee): %v", err)
	}
	if len(graceSees) != 0 {
		t.Errorf("a grantee sees %d accounts of a left-out one, want none", len(graceSees))
	}

	// Bringing it back restores both.
	if err := db.SetAccountLeftOut(ctx, acc.ID, false); err != nil {
		t.Fatalf("bringing it back: %v", err)
	}
	readable, err = db.ReadableAccounts(ctx, conn.ID)
	if err != nil {
		t.Fatalf("ReadableAccounts after bringing back: %v", err)
	}
	if len(readable) != 1 {
		t.Errorf("readable = %d accounts after bringing one back, want 1", len(readable))
	}
	graceSees, err = db.VisibleAccounts(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts(grantee) after bringing back: %v", err)
	}
	if len(graceSees) != 1 {
		t.Errorf("the grantee sees %d accounts after it came back, want 1", len(graceSees))
	}
}

// A household name survives a reconnect: the gateway's upsert never names
// household_name in its `do update set` list, which is the entire mechanism
// (ADR 0022).
func TestAHouseholdNameSurvivesAReconnect(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	conn, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Conta a Ordem", "0538"))
	acc := stored[0]

	if err := db.SetAccountName(ctx, acc.ID, "Rent"); err != nil {
		t.Fatalf("SetAccountName: %v", err)
	}

	// The gateway's own upsert, exactly as a restore runs it, with its own name
	// for the account unchanged.
	if _, _, err := db.ReplaceConnectionAccounts(ctx, conn.ID,
		[]store.Account{account("hash-1", "Conta a Ordem", "0538")}, ada.ID); err != nil {
		t.Fatalf("ReplaceConnectionAccounts: %v", err)
	}

	got, err := db.AccountByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("AccountByID: %v", err)
	}
	if got.HouseholdName != "Rent" {
		t.Errorf("HouseholdName = %q after a reconnect, want it to survive as %q", got.HouseholdName, "Rent")
	}
	if got.Name != "Conta a Ordem" {
		t.Errorf("Name = %q, want the bank's own name kept beside it", got.Name)
	}

	// Clearing it reverts the screen to the bank's own name.
	if err := db.SetAccountName(ctx, acc.ID, ""); err != nil {
		t.Fatalf("clearing the household name: %v", err)
	}
	got, err = db.AccountByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("AccountByID: %v", err)
	}
	if got.HouseholdName != "" {
		t.Errorf("HouseholdName = %q after clearing, want empty", got.HouseholdName)
	}
}

// A balance and its read time are written together: a balance without one is
// not a balance in this system, and the schema refuses the pair being split.
func TestABalanceIsStoredWithItsReadTime(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	now, err := db.Now(ctx)
	if err != nil {
		t.Fatalf("reading database time: %v", err)
	}
	if err := db.RecordBalance(ctx, stored[0].ID, 420_010, now.T); err != nil {
		t.Fatalf("RecordBalance: %v", err)
	}

	a, err := db.AccountByID(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("AccountByID: %v", err)
	}
	if a.BalanceMinor == nil || *a.BalanceMinor != 420_010 {
		t.Errorf("BalanceMinor = %v, want 420010", a.BalanceMinor)
	}
	if a.BalanceReadAt == nil {
		t.Error("a balance was stored with no read time")
	}
}

// Restoring: owners and grants carry forward on matched accounts, a newly
// offered one arrives owned by the restorer with nobody granted, and a
// withdrawn one goes and is named.
func TestRestoringCarriesOwnersAndGrantsForward(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	conn, stored := connect(t, ctx, db, ada.ID, 24*time.Hour,
		account("hash-1", "Joint", "0538"),
		account("hash-2", "Withdrawn", "5594"))

	if err := db.SetAccountLevel(ctx, stored[0].ID, grace.ID, store.LevelDetails, ada.ID); err != nil {
		t.Fatalf("granting: %v", err)
	}

	// The bank comes back offering the joint account with a new per-session
	// uid, a new account, and no longer the second one.
	restored, withdrawn, err := db.ReplaceConnectionAccounts(ctx, conn.ID, []store.Account{
		account("hash-1", "Joint", "0538"),
		account("hash-3", "Newly offered", "7712"),
	}, ada.ID)
	if err != nil {
		t.Fatalf("ReplaceConnectionAccounts: %v", err)
	}
	if len(restored) != 2 {
		t.Fatalf("restored %d accounts, want 2", len(restored))
	}
	if len(withdrawn) != 1 || withdrawn[0] != "Withdrawn" {
		t.Errorf("withdrawn = %v, want the account the bank no longer offers", withdrawn)
	}

	// The grant survived the restore.
	graceSees, err := db.VisibleAccounts(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(graceSees) != 1 || graceSees[0].Level != store.LevelDetails {
		t.Errorf("grace sees %+v, want the joint account still at details", graceSees)
	}

	// The newly offered account is the restorer's, and nobody else's.
	adaSees, err := db.VisibleAccounts(ctx, ada.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(adaSees) != 2 {
		t.Errorf("ada sees %d accounts, want 2", len(adaSees))
	}
}

// Disconnecting keeps the connection row for the audit question and leaves
// nothing openable on it.
//
// The accounts stay too, and this is the assertion that says so: ending wimm's
// access and destroying what that access read are different decisions, and only
// the first is made here (ADR 0021). What a member experiences is unchanged —
// the accounts stop being listed — and that is checked through the one query
// that decides it rather than by counting rows.
func TestDisconnectingKeepsTheRowAndDestroysTheSecrets(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	conn, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	if err := db.DisconnectBankConnection(ctx, conn.ID); err != nil {
		t.Fatalf("DisconnectBankConnection: %v", err)
	}

	after, err := db.BankConnectionByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("the connection row did not survive: %v", err)
	}
	if after.DisconnectedAt == nil {
		t.Error("disconnected_at was not set")
	}
	if len(after.GatewayRef.Ciphertext) != 0 || after.GatewayRef.KeyID != "" {
		t.Error("a sealed value survived disconnection")
	}

	if got := count(t, ctx, db, "accounts"); got != 1 {
		t.Errorf("%d accounts remain after disconnecting, want the 1 that was read", got)
	}
	kept, err := db.AccountByID(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("the account did not survive disconnection: %v", err)
	}
	if len(kept.GatewayUID.Ciphertext) != 0 || kept.GatewayUID.KeyID != "" {
		t.Error("a sealed per-session identifier survived on an account after disconnection")
	}

	visible, err := db.VisibleAccounts(ctx, ada.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(visible) != 0 {
		t.Errorf("a disconnected bank's accounts are still shown: %+v", visible)
	}
}

// Reconnecting the same bank finds the account that outlived the disconnection
// and re-points it, rather than putting a second copy beside it. This was free
// while disconnection deleted; it is what the identity index buys now.
func TestReconnectingReattachesTheAccountWithEverythingOnIt(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	first, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))
	if err := db.SetAccountLevel(ctx, stored[0].ID, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatalf("granting a level: %v", err)
	}
	if _, err := db.Pool().Exec(ctx, `
		insert into transactions (account_id, status, dedup_key, amount_minor, currency, booking_date)
		values ($1, 'booked', 'ref-1', -1250, 'EUR', current_date)`, stored[0].ID); err != nil {
		t.Fatalf("storing a transaction: %v", err)
	}

	if err := db.DisconnectBankConnection(ctx, first.ID); err != nil {
		t.Fatalf("DisconnectBankConnection: %v", err)
	}

	second, again := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	if got := count(t, ctx, db, "accounts"); got != 1 {
		t.Fatalf("%d account rows after reconnecting, want 1", got)
	}
	if again[0].ID != stored[0].ID {
		t.Errorf("reconnecting produced account %s, want the one that was there: %s", again[0].ID, stored[0].ID)
	}
	if again[0].ConnectionID == nil || *again[0].ConnectionID != second.ID {
		t.Error("the account was not re-pointed at the new connection")
	}

	owners, err := db.AccountOwners(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("reading owners: %v", err)
	}
	if len(owners) != 1 || owners[0] != ada.ID {
		t.Errorf("owners = %v, want the one owner it had", owners)
	}

	grants, err := db.GrantsForConnection(ctx, second.ID)
	if err != nil {
		t.Fatalf("reading grants: %v", err)
	}
	if len(grants) != 1 || grants[0].MemberID != grace.ID || grants[0].Level != store.LevelBalance {
		t.Errorf("grants = %+v, want the one grant it had", grants)
	}

	if got := count(t, ctx, db, "transactions"); got != 1 {
		t.Errorf("%d transactions after reconnecting, want the 1 that was read", got)
	}
}

// An account is identified by its bank and the gateway's cross-session hash,
// across connections rather than within one. Making a second copy
// unrepresentable is what leaves re-attach with nothing to disambiguate: a
// missed match would present a member with two of their own accounts, one
// holding the history.
func TestOneBankNeverHoldsTwoCopiesOfOneAccount(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	// A second connection row at the same bank, and an account inserted
	// directly against it: the upsert would have re-attached, so the index is
	// asked the question on its own.
	var secondID string
	if err := db.Pool().QueryRow(ctx, `
		insert into bank_connections (gateway, bank_id, bank_name, connected_by, consent_expires_at)
		values ('enablebanking', $1, 'Montepio', $2, now() + interval '90 days')
		returning id`, "PT:Montepio", ada.ID).Scan(&secondID); err != nil {
		t.Fatalf("recording a second connection: %v", err)
	}

	_, err := db.Pool().Exec(ctx, `
		insert into accounts (source, connection_id, bank_id, gateway_ref, name, currency)
		values ('gateway', $1, $2, 'hash-1', 'Joint again', 'EUR')`, secondID, "PT:Montepio")
	if err == nil {
		t.Error("a second copy of one account at one bank was accepted")
	}
}

// A gateway account whose bank is unknown cannot be re-attached after a
// disconnection, because its bank is half of its identity.
func TestAGatewayAccountMustCarryItsBank(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	var connID string
	if err := db.Pool().QueryRow(ctx, `
		insert into bank_connections (gateway, bank_id, bank_name, connected_by, consent_expires_at)
		values ('enablebanking', 'PT:Montepio', 'Montepio', $1, now() + interval '90 days')
		returning id`, ada.ID).Scan(&connID); err != nil {
		t.Fatalf("recording a connection: %v", err)
	}

	if _, err := db.Pool().Exec(ctx, `
		insert into accounts (source, connection_id, gateway_ref, name, currency)
		values ('gateway', $1, 'hash-9', 'Bankless', 'EUR')`, connID); err == nil {
		t.Error("a gateway account with no bank was accepted")
	}
}

// Consent running out destroys the secrets too, and is measured in SQL.
func TestConsentRunningOutExpiresTheConnectionAndDestroysTheSecrets(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")

	// ActivoBank's one day, already past.
	conn, _ := connect(t, ctx, db, ada.ID, -time.Minute, account("hash-1", "Joint", "0538"))

	n, err := db.ExpireBankConnectionsPastConsent(ctx)
	if err != nil {
		t.Fatalf("ExpireBankConnectionsPastConsent: %v", err)
	}
	if n != 1 {
		t.Errorf("expired %d connections, want 1", n)
	}

	after, err := db.BankConnectionByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("BankConnectionByID: %v", err)
	}
	if after.ExpiredAt == nil {
		t.Error("expired_at was not set")
	}
	if len(after.GatewayRef.Ciphertext) != 0 {
		t.Error("a sealed value survived the grant running out")
	}
	if after.Live() {
		t.Error("an expired connection reports itself live")
	}
}

// A connection whose consent is still good is left alone.
func TestALiveConnectionIsNotExpired(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	conn, _ := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	if _, err := db.ExpireBankConnectionsPastConsent(ctx); err != nil {
		t.Fatalf("ExpireBankConnectionsPastConsent: %v", err)
	}
	after, err := db.BankConnectionByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("BankConnectionByID: %v", err)
	}
	if !after.Live() {
		t.Error("a live connection was expired")
	}
}

func TestOnlyAnOwnerMayChangeAnAccount(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	owns, err := db.MemberOwnsAccount(ctx, stored[0].ID, ada.ID)
	if err != nil || !owns {
		t.Errorf("MemberOwnsAccount(owner) = %v, %v", owns, err)
	}
	owns, err = db.MemberOwnsAccount(ctx, stored[0].ID, grace.ID)
	if err != nil || owns {
		t.Errorf("MemberOwnsAccount(non-owner) = %v, %v", owns, err)
	}
}

// Removing a member takes their visibility with them, in one statement.
func TestRemovingAMemberRemovesTheirOwnershipAndGrants(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")
	_, stored := connect(t, ctx, db, ada.ID, 90*24*time.Hour, account("hash-1", "Joint", "0538"))

	if err := db.SetAccountLevel(ctx, stored[0].ID, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatalf("granting: %v", err)
	}
	if _, err := db.Pool().Exec(ctx, `delete from members where id = $1`, grace.ID); err != nil {
		t.Fatalf("removing a member: %v", err)
	}

	if got := count(t, ctx, db, "account_grants"); got != 0 {
		t.Errorf("%d grants remain after the member was removed", got)
	}
}

// The fourth sweep statement. Abandoned authorisations are the bulk of this
// table: most members who reach a bank's consent screen and stop leave one.
func TestAbandonedBankConnectionsAreSweptAndLiveOnesStay(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	m := member(t, ctx, db, "ada@example.com")

	pending := func(state string, lifetime time.Duration) {
		t.Helper()
		if _, err := db.CreatePendingBankConnection(ctx, store.PendingBankConnection{
			Gateway: "enablebanking", GatewayRef: "auth-" + state, BankID: "PT:Montepio",
			BankName: "Montepio", RedirectURL: "https://x", StartedBy: m.ID,
		}, []byte(state), lifetime); err != nil {
			t.Fatalf("recording %s: %v", state, err)
		}
	}

	pending("abandoned-1", -time.Hour)
	pending("abandoned-2", -time.Minute)
	pending("still-live", time.Hour)

	n, err := db.DeleteAbandonedBankConnections(ctx)
	if err != nil {
		t.Fatalf("DeleteAbandonedBankConnections: %v", err)
	}
	if n != 2 {
		t.Errorf("removed %d, want the 2 that expired", n)
	}
	if got := count(t, ctx, db, "pending_bank_connections"); got != 1 {
		t.Errorf("%d rows remain, want the one still live", got)
	}
}

// The backlog on an instance running for months is larger than one batch, so
// the delete repeats until a batch comes up short — the same property the
// other three statements have.
func TestAbandonedBankConnectionsAreDeletedPastTheFirstBatch(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	m := member(t, ctx, db, "ada@example.com")

	const expired = store.DeleteBatchSize + 5
	if _, err := db.Pool().Exec(ctx, `
		insert into pending_bank_connections
			(state_hash, gateway, gateway_ref, bank_id, bank_name, redirect_url, started_by, expires_at)
		select decode(md5(g::text), 'hex'), 'enablebanking', 'ref-' || g, 'PT:Montepio', 'Montepio',
		       'https://x', $1, now() - make_interval(secs => g)
		from generate_series(1, $2) g`, m.ID, expired); err != nil {
		t.Fatalf("seeding: %v", err)
	}

	n, err := db.DeleteAbandonedBankConnections(ctx)
	if err != nil {
		t.Fatalf("DeleteAbandonedBankConnections: %v", err)
	}
	if n != expired {
		t.Errorf("removed %d, want %d", n, expired)
	}
}

// The decoupling, asserted rather than assumed: an account exists on its own
// terms. Nothing in this change creates one this way, which is exactly why it
// is worth a test — the constraint that would forbid it would otherwise be
// discovered by whoever adds the first non-gateway account type.
func TestAnAccountCanExistWithNoGatewayBehindIt(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()
	ada := member(t, ctx, db, "ada@example.com")
	grace := member(t, ctx, db, "grace@example.com")

	// An account always has an owner (ADR 0022), so its creation and its first
	// owner row are written in one transaction: the deferred trigger checks at
	// commit, and a bare autocommitted insert would never reach one.
	tx, err := db.Pool().Begin(ctx)
	if err != nil {
		t.Fatalf("beginning a transaction: %v", err)
	}
	var id string
	if err := tx.QueryRow(ctx, `
		insert into accounts (source, name, currency) values ('manual', 'Cash tin', 'EUR')
		returning id`).Scan(&id); err != nil {
		t.Fatalf("an account with no connection was refused: %v", err)
	}
	if _, err := tx.Exec(ctx,
		`insert into account_owners (account_id, member_id) values ($1, $2)`, id, ada.ID); err != nil {
		t.Fatalf("owning a sourceless account: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("committing a sourceless account with its owner: %v", err)
	}

	// It is shared through exactly the same table as any other account.
	if err := db.SetAccountLevel(ctx, id, grace.ID, store.LevelBalance, ada.ID); err != nil {
		t.Fatalf("granting on a sourceless account: %v", err)
	}

	// And it reaches both members through the one visibility query, with no
	// connection attached.
	adaSees, err := db.VisibleAccounts(ctx, ada.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(adaSees) != 1 {
		t.Fatalf("the owner sees %d accounts, want 1", len(adaSees))
	}
	if adaSees[0].Connection != nil {
		t.Error("an account with no gateway reported a connection")
	}
	if adaSees[0].Source != store.SourceManual {
		t.Errorf("Source = %q, want manual", adaSees[0].Source)
	}
	if !adaSees[0].Owned || adaSees[0].Level != store.LevelDetails {
		t.Error("an owner does not see their own sourceless account in full")
	}

	graceSees, err := db.VisibleAccounts(ctx, grace.ID, "")
	if err != nil {
		t.Fatalf("VisibleAccounts: %v", err)
	}
	if len(graceSees) != 1 || graceSees[0].Level != store.LevelBalance {
		t.Errorf("the granted member sees %+v, want it at balance", graceSees)
	}
}

// A gateway account must carry its connection and reference, and a sourceless
// one must carry neither. Without this the source column and the columns it
// describes could disagree, and the disagreement would be found by a reader.
func TestSourceAndTheGatewayColumnsCannotDisagree(t *testing.T) {
	db := storetest.New(t)
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		sql  string
	}{
		{
			"a gateway account with no connection",
			`insert into accounts (source, name, currency) values ('gateway', 'Orphan', 'EUR')`,
		},
		{
			"a manual account carrying a gateway reference",
			`insert into accounts (source, gateway_ref, name, currency)
			 values ('manual', 'hash-1', 'Confused', 'EUR')`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := db.Pool().Exec(ctx, tc.sql); err == nil {
				t.Error("accepted")
			}
		})
	}
}
