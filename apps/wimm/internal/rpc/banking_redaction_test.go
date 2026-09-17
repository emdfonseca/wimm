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
