package rpc

import (
	"testing"
	"time"

	bankingv1 "github.com/xuuid/wimm/packages/contracts/gen/go/wimm/banking/v1"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

func TestAListedTransactionCarriesTheMerchantNameAndTheBanksLine(t *testing.T) {
	const line = "COMPRA WWW.AMAZON NM4HU1VZ4 230002268264350"
	out := toProtoTransaction(store.Transaction{
		ID: "tx-1", AccountID: "acct-1", Status: store.StatusBooked,
		AmountMinor: -6499, Currency: "EUR", Remittance: line,
		BookingDate: time.Date(2026, time.September, 9, 0, 0, 0, 0, time.UTC),
	}, store.AccountLabel{Name: "Conta à Ordem", BankName: "Montepio"})

	if out.GetDisplayName() != "Amazon" {
		t.Errorf("DisplayName = %q, want Amazon", out.GetDisplayName())
	}
	if out.GetRemittance() != line {
		t.Errorf("Remittance = %q, want the bank's line untouched", out.GetRemittance())
	}
	if out.GetBankName() != "Montepio" || out.GetAccountName() != "Conta à Ordem" {
		t.Errorf("account label lost: %q at %q", out.GetAccountName(), out.GetBankName())
	}
}

func TestAMonthWithNoComparisonSendsNoPriorFigures(t *testing.T) {
	out := toProtoMonthSummary(banking.MonthSummary{Months: []banking.CurrencyMonth{{
		Currency: "EUR", In: banking.Money{Minor: 5, Currency: "EUR"},
		Out: banking.Money{Minor: 3, Currency: "EUR"}, Net: banking.Money{Minor: 2, Currency: "EUR"},
	}}})

	m := out.GetMonths()[0]
	if m.PriorIn != nil || m.PriorOut != nil || m.PriorNet != nil || m.CountedFrom != nil {
		t.Errorf("unset figures reached the wire: %+v", m)
	}
}

func TestEveryGroupHasItsOwnWireValue(t *testing.T) {
	seen := map[bankingv1.AccountGroup]banking.Group{}
	for _, g := range []banking.Group{banking.GroupNone, banking.GroupHousehold, banking.GroupOwn, banking.GroupShared} {
		w := toProtoGroup(g)
		if prev, dup := seen[w]; dup {
			t.Errorf("groups %v and %v share wire value %v", prev, g, w)
		}
		seen[w] = g
	}
}
