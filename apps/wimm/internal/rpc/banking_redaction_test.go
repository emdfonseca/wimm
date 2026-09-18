package rpc

import (
	"testing"
	"time"

	bankingv1 "github.com/xuuid/wimm/packages/contracts/gen/go/wimm/banking/v1"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// Redaction happens on the way out, not on the way in to the browser. These
// assert the shape of what is serialised, because "the client does not show it"
// is not the same promise as "the server does not send it".

// toProtoAccountForTest renders without a name lookup: these assert which
// fields reach the wire, not who connected the bank.
func toProtoAccountForTest(a store.VisibleAccount) *bankingv1.Account {
	return toProtoAccount(a, names{})
}

func visible(level store.Level, owned bool) store.VisibleAccount {
	minor := int64(420_010)
	at := time.Date(2026, time.September, 17, 9, 0, 0, 0, time.UTC)
	return store.VisibleAccount{
		Account: store.Account{
			ID: "acct-1", Source: store.SourceGateway, Name: "Conta à Ordem",
			NumberSuffix: "0538", AccountType: "CACC", HolderName: "Ada Lovelace",
			Currency: "EUR", BalanceMinor: &minor, BalanceReadAt: &at,
		},
		Owned: owned, Level: level,
		Connection: &store.AccountConnection{
			ID: "conn-1", BankID: "PT:Montepio", BankName: "Montepio", Live: true,
		},
	}
}

// At balance, the identifying fields are absent — not empty strings that a
// client might render, and not present-but-blank.
func TestBalanceLevelSendsNoIdentifyingFields(t *testing.T) {
	out := toProtoAccountForTest(visible(store.LevelBalance, false))

	if out.GetName() != "Conta à Ordem" {
		t.Errorf("Name = %q, want the account's name", out.GetName())
	}
	if out.GetBalance().GetMoney().GetMinor() != 420_010 {
		t.Errorf("balance = %d", out.GetBalance().GetMoney().GetMinor())
	}
	if out.GetConnection().GetBankName() != "Montepio" {
		t.Error("the bank should still be named at balance level")
	}

	for name, got := range map[string]string{
		"number_suffix": out.GetNumberSuffix(),
		"account_type":  out.GetAccountType(),
		"holder_name":   out.GetHolderName(),
	} {
		if got != "" {
			t.Errorf("%s reached the wire at balance level: %q", name, got)
		}
	}
}

func TestDetailsLevelSendsTheIdentifyingFields(t *testing.T) {
	out := toProtoAccountForTest(visible(store.LevelDetails, false))

	if out.GetNumberSuffix() != "0538" {
		t.Errorf("NumberSuffix = %q", out.GetNumberSuffix())
	}
	if out.GetAccountType() != "CACC" {
		t.Errorf("AccountType = %q", out.GetAccountType())
	}
	if out.GetHolderName() != "Ada Lovelace" {
		t.Errorf("HolderName = %q", out.GetHolderName())
	}
}

// An owner always sees their own account in full, whatever level is recorded.
func TestAnOwnerSeesEverything(t *testing.T) {
	out := toProtoAccountForTest(visible(store.LevelDetails, true))
	if !out.GetOwned() {
		t.Error("Owned did not reach the wire")
	}
	if out.GetNumberSuffix() == "" {
		t.Error("an owner was redacted")
	}
}

// The belt-and-braces case: a level that is neither balance nor details sends
// nothing about the account at all. A member in this state never gets a row in
// the first place, so reaching here means something upstream went wrong — and
// the safe outcome is an empty account, not a full one.
func TestAnUnknownLevelSendsNothing(t *testing.T) {
	out := toProtoAccountForTest(visible(store.Level("something-else"), false))

	if out.GetName() != "" {
		t.Errorf("Name reached the wire at an unrecognised level: %q", out.GetName())
	}
	if out.GetBalance() != nil {
		t.Error("a balance reached the wire at an unrecognised level")
	}
	if out.GetNumberSuffix() != "" || out.GetHolderName() != "" {
		t.Error("identifying fields reached the wire at an unrecognised level")
	}
}

// An unset level on the way in means hidden, never "leave unchanged" and never
// a refusal: the failure mode has to be accidentally hidden.
func TestAnUnsetLevelMeansHidden(t *testing.T) {
	if got := fromProtoLevel(0); got != "" {
		t.Errorf("UNSPECIFIED mapped to %q, want no grant", got)
	}
}

// A left-out account omits its balance even for its owner (ADR 0022): the
// query still returns it, so this is asserted at the one place redaction
// happens — the field is absent from the response, not present as a zero
// value that a client might render as "€0.00, just read".
func TestALeftOutAccountSendsNoBalance(t *testing.T) {
	a := visible(store.LevelDetails, true)
	leftOutAt := time.Date(2026, time.September, 18, 8, 0, 0, 0, time.UTC)
	a.LeftOutAt = &leftOutAt

	out := toProtoAccountForTest(a)

	if out.GetBalance() != nil {
		t.Error("a left-out account's balance reached the wire")
	}
	if !out.GetLeftOutAt().IsValid() {
		t.Error("LeftOutAt did not reach the wire")
	}
	// Everything else an owner sees stays: leaving an account out is a fact
	// about it, not a demotion in what its owner may see of it.
	if out.GetName() == "" {
		t.Error("a left-out account's name was redacted from its owner")
	}
	if out.GetNumberSuffix() == "" {
		t.Error("a left-out account's details were redacted from its owner")
	}
}

// The household's own name reaches the wire beside the bank's, never in place
// of it: household_name is the reader's job (ADR 0022), so both are sent.
func TestHouseholdNameIsSerialisedBesideTheBanksOwn(t *testing.T) {
	a := visible(store.LevelBalance, false)
	a.HouseholdName = "Rent"

	out := toProtoAccountForTest(a)

	if out.GetHouseholdName() != "Rent" {
		t.Errorf("HouseholdName = %q, want %q", out.GetHouseholdName(), "Rent")
	}
	if out.GetName() != "Conta à Ordem" {
		t.Errorf("Name = %q, want the bank's own name kept beside it", out.GetName())
	}
}

// Bringing an account back is clearing the fact, not restating the balance:
// once left_out_at is nil again the query's own balance columns are what
// decide whether one renders.
func TestAnAccountBroughtBackSendsItsBalanceAgain(t *testing.T) {
	out := toProtoAccountForTest(visible(store.LevelDetails, true))

	if out.GetLeftOutAt() != nil {
		t.Error("LeftOutAt reached the wire for an account that was never left out")
	}
	if out.GetBalance() == nil {
		t.Error("an account in wimm sent no balance")
	}
}
