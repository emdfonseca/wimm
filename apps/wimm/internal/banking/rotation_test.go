package banking_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// Rotation is a line added at the top of the key file. The old key stays until
// nothing references it, and the point of the key id on every row is that both
// can be true at once — no flag day, no migration, no window where a value is
// unopenable (ADR 0018).
func TestAValueSealedUnderAPreviousKeyStillOpens(t *testing.T) {
	old := banking.Key{ID: "2025-07", Material: testKey(1)}
	newer := banking.Key{ID: "2026-01", Material: testKey(2)}

	before := keyring(t, old)
	sealed, err := before.Seal([]byte("session-from-last-year"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// Rotation: the new key first, the old one kept.
	after := keyring(t, newer, old)

	got, err := after.Open(sealed, rowA)
	if err != nil {
		t.Fatalf("a value sealed under the previous key did not open: %v", err)
	}
	if string(got) != "session-from-last-year" {
		t.Errorf("Open = %q", got)
	}
}

// Writing uses the current key, so a value rewritten after rotation carries the
// new id and the old key becomes retirable.
func TestNewValuesUseTheCurrentKey(t *testing.T) {
	k := keyring(t,
		banking.Key{ID: "2026-01", Material: testKey(2)},
		banking.Key{ID: "2025-07", Material: testKey(1)},
	)
	if k.CurrentKeyID() != "2026-01" {
		t.Fatalf("CurrentKeyID = %q, want the first key", k.CurrentKeyID())
	}

	sealed, err := k.Seal([]byte("fresh"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if sealed.KeyID() != "2026-01" {
		t.Errorf("sealed under %q, want the current key", sealed.KeyID())
	}
}

// Resealing is what empties the old key out: open with whatever the row names,
// seal with the current one. Proving the round trip here is what makes retiring
// a key a decision rather than a gamble.
func TestResealingMovesAValueOntoTheCurrentKey(t *testing.T) {
	old := banking.Key{ID: "2025-07", Material: testKey(1)}
	newer := banking.Key{ID: "2026-01", Material: testKey(2)}

	before := keyring(t, old)
	sealed, err := before.Seal([]byte("secret"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	after := keyring(t, newer, old)
	plaintext, err := after.Open(sealed, rowA)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	resealed, err := after.Seal(plaintext, rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if resealed.KeyID() != "2026-01" {
		t.Errorf("resealed under %q, want 2026-01", resealed.KeyID())
	}

	// And once resealed, the old key is no longer needed to read it.
	onlyNew := keyring(t, newer)
	if _, err := onlyNew.Open(resealed, rowA); err != nil {
		t.Errorf("a resealed value needs the retired key: %v", err)
	}
	// While the original still does, which is why retiring is a decision.
	if _, err := onlyNew.Open(sealed, rowA); err == nil {
		t.Error("a value sealed under the retired key opened without it")
	}
}

// Rotation must not weaken the binding: a value moved onto the new key is still
// bound to its row.
func TestRotationKeepsTheRowBinding(t *testing.T) {
	k := keyring(t,
		banking.Key{ID: "2026-01", Material: testKey(2)},
		banking.Key{ID: "2025-07", Material: testKey(1)},
	)
	sealed, err := k.Seal([]byte("secret"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if _, err := k.Open(sealed, rowB); err == nil {
		t.Error("opened against the wrong row after rotation")
	}
}

// The key file is the rotation interface, so the ordering it expresses is part
// of the contract: first line seals.
func TestTheFirstKeyInTheFileSeals(t *testing.T) {
	contents := strings.Join([]string{
		"# newest first",
		"2026-01 " + encodedKey(2),
		"2025-07 " + encodedKey(1),
		"",
	}, "\n")
	path := writeKeyFile(t, contents, 0o600)

	k, err := banking.LoadKeyring(path)
	if err != nil {
		t.Fatalf("LoadKeyring: %v", err)
	}
	if k.CurrentKeyID() != "2026-01" {
		t.Errorf("CurrentKeyID = %q, want 2026-01", k.CurrentKeyID())
	}

	// And the second key still opens what it sealed.
	older := keyring(t, banking.Key{ID: "2025-07", Material: bytes.Repeat([]byte{1}, banking.KeySize)})
	sealed, err := older.Seal([]byte("old"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if _, err := k.Open(sealed, rowA); err != nil {
		t.Errorf("the loaded keyring cannot open a value from its second key: %v", err)
	}
}
