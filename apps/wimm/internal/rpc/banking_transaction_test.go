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

func TestAListedRowCarriesItsUnusualMark(t *testing.T) {
	when := time.Date(2026, time.March, 12, 0, 0, 0, 0, time.UTC)
	out := toProtoLedger(banking.Ledger{
		Page: store.LedgerPage{Transactions: []store.Transaction{
			{ID: "garage", AmountMinor: -165_000, Currency: "EUR", BookingDate: when},
			{ID: "coffee", AmountMinor: -120, Currency: "EUR", BookingDate: when},
		}},
		Patterns: banking.Patterns{
			Unusual: map[string]banking.UnusualMark{"garage": {TransactionID: "garage", FirstPayment: true}},
		},
	})
	if !out.Transactions[0].Unusual || out.Transactions[1].Unusual {
		t.Errorf("unusual = %v and %v, want the garage only", out.Transactions[0].Unusual, out.Transactions[1].Unusual)
	}
}

func TestAScopeOnOfferKeepsACurrencyWithNoFullMonthAndNoRecurringPayment(t *testing.T) {
	h := banking.History{
		Available:  []banking.Scope{banking.ScopeHousehold, banking.ScopeOwn, banking.ScopeAll},
		Currencies: []banking.CurrencyHistory{{Currency: "EUR"}},
	}
	if got := len(toProtoHistory(h).Histories); got != 1 {
		t.Errorf("got %d histories, want the currency kept so the screen can say why it is empty", got)
	}
	h.Available = nil
	if got := len(toProtoHistory(h).Histories); got != 0 {
		t.Errorf("got %d histories, want none where no control is on offer", got)
	}
}

func TestAListedRowSaysWhenItIsHalfOfATransfer(t *testing.T) {
	when := time.Date(2026, time.September, 10, 0, 0, 0, 0, time.UTC)
	out := toProtoLedger(banking.Ledger{
		Page: store.LedgerPage{Transactions: []store.Transaction{
			{ID: "out", AmountMinor: -50_000, Currency: "EUR", BookingDate: when},
			{ID: "in", AmountMinor: 50_000, Currency: "EUR", BookingDate: when},
			{ID: "coffee", AmountMinor: -120, Currency: "EUR", BookingDate: when},
		}},
		Patterns: banking.Patterns{
			// The coffee is marked unusual; the out row is marked both, and
			// the label takes the slot.
			Unusual: map[string]banking.UnusualMark{
				"coffee": {TransactionID: "coffee"},
				"out":    {TransactionID: "out"},
			},
			Transfers: map[string]string{"out": "in", "in": "out"},
		},
	})

	if !out.Transactions[0].GetOwnTransfer() || !out.Transactions[1].GetOwnTransfer() {
		t.Error("both halves of the pair must carry own_transfer")
	}
	if out.Transactions[2].GetOwnTransfer() {
		t.Error("the coffee is half of nothing")
	}
	if out.Transactions[0].GetUnusual() {
		t.Error("the out row is unusual as well as a transfer, and the label takes the slot")
	}
	if !out.Transactions[2].GetUnusual() {
		t.Error("the coffee lost its unusual mark")
	}
}

func TestAMonthSaysHowManyTransfersItLeftOut(t *testing.T) {
	out := toProtoCurrencyHistory(banking.CurrencyHistory{
		Currency: "EUR",
		Months: []banking.HistoryMonth{
			{
				Start:            time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
				In:               banking.Money{Minor: 245_000, Currency: "EUR"},
				Out:              banking.Money{Minor: 190_000, Currency: "EUR"},
				Net:              banking.Money{Minor: 55_000, Currency: "EUR"},
				TransfersLeftOut: 2,
				TransfersTotal:   banking.Money{Minor: 140_000, Currency: "EUR"},
			},
			{
				Start: time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
				In:    banking.Money{Minor: 1, Currency: "EUR"},
				Out:   banking.Money{Minor: 1, Currency: "EUR"},
				Net:   banking.Money{Currency: "EUR"},
			},
		},
	}, nil)

	sep, aug := out.GetMonths()[0], out.GetMonths()[1]
	if sep.GetTransfersLeftOut() != 2 {
		t.Errorf("September left out %d, want 2", sep.GetTransfersLeftOut())
	}
	if sep.GetTransfersTotal().GetMinor() != 140_000 || sep.GetTransfersTotal().GetCurrency() != "EUR" {
		t.Errorf("September total = %+v, want 140000 EUR", sep.GetTransfersTotal())
	}
	if aug.GetTransfersLeftOut() != 0 || aug.GetTransfersTotal() != nil {
		t.Errorf("August = %d left out, total %+v, want nothing",
			aug.GetTransfersLeftOut(), aug.GetTransfersTotal())
	}
}

func TestAMoverSaysWhenItIsHalfOfATransfer(t *testing.T) {
	m := banking.DayMover{TransactionID: "out", Name: "Transfer", Amount: -50_000, OwnTransfer: true}
	got := &bankingv1.DayMover{
		DisplayName: m.Name,
		Amount:      &bankingv1.Money{Minor: m.Amount, Currency: "EUR"},
		Unusual:     m.Unusual,
		OwnTransfer: m.OwnTransfer,
	}
	if !got.GetOwnTransfer() || got.GetUnusual() {
		t.Errorf("mover = own_transfer %v unusual %v, want true and false",
			got.GetOwnTransfer(), got.GetUnusual())
	}
}

func TestACurrencySaysHowManyMonthsItsTypicalMonthIsDrawnFrom(t *testing.T) {
	typical := banking.Money{Minor: 163_350, Currency: "EUR"}
	out := toProtoCurrencyHistory(banking.CurrencyHistory{
		Currency: "EUR", FullMonths: 12, TypicalMonths: 6,
		TypicalNet: &typical, AverageNet: &typical,
	}, nil)

	if out.GetFullMonths() != 12 || out.GetTypicalMonths() != 6 {
		t.Errorf("full %d, typical %d, want 12 and 6", out.GetFullMonths(), out.GetTypicalMonths())
	}
}
