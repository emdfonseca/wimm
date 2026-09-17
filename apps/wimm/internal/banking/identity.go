package banking

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// DedupKey identifies a booked transaction within one account.
//
// The bank's own entry reference where it gives one, because that is the field
// the standard defines for exactly this. Where it does not — the field is
// optional in the schema and some banks omit it — a digest over the values that
// identify the transaction stands in.
//
// The gateway's transaction_id is deliberately not used. There is no evidence
// it is stable across sessions, and the same reasoning that keys an account on
// its cross-session hash rather than its per-session uid applies here (ADR
// 0021).
//
// The key is not unique on its own: two identical coffees on the same day at
// the same shop produce one key and are two real transactions. What tells them
// apart is the occurrence index the store assigns.
func DedupKey(t Transaction) string {
	if ref := strings.TrimSpace(t.Ref); ref != "" {
		return ref
	}

	// A unit separator between fields, so a counterparty ending in the text a
	// remittance starts with cannot produce the same digest as the pair the
	// other way round.
	const sep = "\x1f"
	digest := sha256.Sum256([]byte(strings.Join([]string{
		dateKey(t.BookingDate),
		strconv.FormatInt(t.Amount.Minor, 10),
		strings.ToUpper(strings.TrimSpace(t.Amount.Currency)),
		strings.TrimSpace(t.CounterpartyName),
		strings.TrimSpace(t.Remittance),
	}, sep)))
	return hex.EncodeToString(digest[:])
}

// dateKey renders the calendar day, taken as it stands rather than converted.
//
// A booking date is a day the bank named, not an instant: converting it to UTC
// moves it backwards for anything parsed east of it, so the same transaction
// read twice would digest to two keys and be stored twice. A zero time is the
// empty string rather than year one.
func dateKey(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateOnly)
}

// EffectiveDate is the day a transaction is filed under: the day the bank
// booked it, falling back to the value date and then the date of the
// transaction itself.
//
// The ledger orders and groups by this one date, so a transaction that carries
// none of the three cannot be placed in it at all. Inventing today would file
// it on the wrong day, and the day is the thing a member scans by.
func EffectiveDate(t Transaction) (time.Time, bool) {
	for _, d := range []time.Time{t.BookingDate, t.ValueDate, t.TransactionDate} {
		if !d.IsZero() {
			// The calendar day as the bank named it, for the reason dateKey
			// does not convert either.
			return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC), true
		}
	}
	return time.Time{}, false
}
