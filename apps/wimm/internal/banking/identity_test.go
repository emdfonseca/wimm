package banking_test

import (
	"testing"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

func coffee() banking.Transaction {
	return banking.Transaction{
		Status:           banking.StatusBooked,
		Amount:           banking.Money{Minor: -250, Currency: "EUR"},
		BookingDate:      time.Date(2026, time.March, 4, 0, 0, 0, 0, time.UTC),
		CounterpartyName: "Padaria Ribeiro",
		Remittance:       "Café",
	}
}

// The bank's own reference is the key where it gives one: it is the field the
// standard defines for this, and it survives an amendment that a digest would
// not.
func TestTheBanksEntryReferenceIsTheKey(t *testing.T) {
	tx := coffee()
	tx.Ref = "ref-a"

	if got := banking.DedupKey(tx); got != "ref-a" {
		t.Errorf("DedupKey = %q, want the bank's entry reference", got)
	}

	// Even when everything else about the transaction changes.
	tx.Amount.Minor = -999
	tx.CounterpartyName = "Somewhere else"
	if got := banking.DedupKey(tx); got != "ref-a" {
		t.Errorf("DedupKey = %q after the transaction was amended, want the reference unchanged", got)
	}
}

// Where the bank gives none — the field is optional in the schema — the digest
// stands in, and it is the same value every time the same transaction is read.
func TestTheDigestIsStableAcrossReads(t *testing.T) {
	first := banking.DedupKey(coffee())
	second := banking.DedupKey(coffee())

	if first == "" {
		t.Fatal("a transaction with no entry reference produced no key")
	}
	if first != second {
		t.Errorf("the same transaction digested to %q then %q", first, second)
	}
	if len(first) != 64 {
		t.Errorf("the digest is %d characters, want a 64-character sha256", len(first))
	}

	// Parsed in another location, the same day is the same key: a digest that
	// moved with the reader's timezone would insert the transaction again.
	elsewhere := coffee()
	lisbon := time.FixedZone("WEST", 3600)
	elsewhere.BookingDate = time.Date(2026, time.March, 4, 0, 0, 0, 0, lisbon)
	if got := banking.DedupKey(elsewhere); got != first {
		t.Errorf("the same day in another location digested to %q, want %q", got, first)
	}
}

// Each field that identifies a transaction moves the key, and the separator is
// why two fields cannot be rearranged into the same digest.
func TestEveryIdentifyingFieldChangesTheDigest(t *testing.T) {
	base := banking.DedupKey(coffee())

	for _, tc := range []struct {
		name  string
		apply func(*banking.Transaction)
	}{
		{"a different day", func(tx *banking.Transaction) { tx.BookingDate = tx.BookingDate.AddDate(0, 0, 1) }},
		{"a different amount", func(tx *banking.Transaction) { tx.Amount.Minor = -260 }},
		{"a different currency", func(tx *banking.Transaction) { tx.Amount.Currency = "GBP" }},
		{"a different counterparty", func(tx *banking.Transaction) { tx.CounterpartyName = "Padaria Rib" }},
		{"a different description", func(tx *banking.Transaction) { tx.Remittance = "Chá" }},
		{
			"the counterparty and description rearranged",
			func(tx *banking.Transaction) {
				tx.CounterpartyName = "Padaria RibeiroCafé"
				tx.Remittance = ""
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := coffee()
			tc.apply(&tx)
			if got := banking.DedupKey(tx); got == base {
				t.Error("digested to the same key as the original")
			}
		})
	}
}

// Two identical coffees digest to one key, which is exactly right: they are one
// key and two transactions, and what tells them apart is the occurrence index
// the store assigns.
func TestTwoIdenticalTransactionsShareOneKey(t *testing.T) {
	first, second := coffee(), coffee()
	if banking.DedupKey(first) != banking.DedupKey(second) {
		t.Error("two identical transactions produced different keys")
	}
}

// The ledger orders and groups by one date, so the fallback order is what
// decides where a transaction the bank dated only partially is filed.
func TestTheEffectiveDateFallsBackThroughTheThreeDates(t *testing.T) {
	booking := time.Date(2026, time.March, 4, 0, 0, 0, 0, time.UTC)
	value := time.Date(2026, time.March, 5, 0, 0, 0, 0, time.UTC)
	transaction := time.Date(2026, time.March, 6, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name string
		tx   banking.Transaction
		want time.Time
		ok   bool
	}{
		{"all three", banking.Transaction{BookingDate: booking, ValueDate: value, TransactionDate: transaction}, booking, true},
		{"no booking date", banking.Transaction{ValueDate: value, TransactionDate: transaction}, value, true},
		{"only the transaction date", banking.Transaction{TransactionDate: transaction}, transaction, true},
		{"none at all", banking.Transaction{}, time.Time{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := banking.EffectiveDate(tc.tx)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && !got.Equal(tc.want) {
				t.Errorf("EffectiveDate = %s, want %s", got, tc.want)
			}
		})
	}
}
