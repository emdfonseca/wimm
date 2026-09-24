package banking_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

// testKey returns 32 deterministic bytes. Deterministic because a test that
// fails should fail the same way twice.
func testKey(fill byte) []byte { return bytes.Repeat([]byte{fill}, banking.KeySize) }

func keyring(t *testing.T, keys ...banking.Key) *banking.Keyring {
	t.Helper()
	k, err := banking.NewKeyring(keys)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	return k
}

const (
	rowA = "0f8c3d2a-1b4e-4c7a-9d51-6e2f8a0b3c4d"
	rowB = "7a1e5c90-2d3f-4b8a-8e6c-1f0d9b7a2e53"
)

func TestSealAndOpenRoundTrips(t *testing.T) {
	k := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})
	want := []byte("session-identifier-from-the-gateway")

	sealed, err := k.Seal(want, rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if sealed.KeyID() != "k1" {
		t.Errorf("KeyID = %q, want k1", sealed.KeyID())
	}
	if bytes.Contains(sealed.Ciphertext(), want) {
		t.Error("the plaintext is present in the ciphertext")
	}

	got, err := k.Open(sealed, rowA)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("Open = %q, want %q", got, want)
	}
}

// The property ADR 0018 exists to get: a ciphertext lifted out of one row and
// dropped into another does not open.
func TestCiphertextDoesNotMoveBetweenRows(t *testing.T) {
	k := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})

	sealed, err := k.Seal([]byte("belongs to row A"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if _, err := k.Open(sealed, rowB); err == nil {
		t.Fatal("a value sealed for one row opened against another")
	}
}

func TestWrongKeyFailsClosed(t *testing.T) {
	sealing := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})
	// Same key id, different material: the id is not a secret and an attacker
	// controlling a keyring could name their key anything.
	opening := keyring(t, banking.Key{ID: "k1", Material: testKey(9)})

	sealed, err := sealing.Seal([]byte("secret"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	got, err := opening.Open(sealed, rowA)
	if err == nil {
		t.Fatal("a value opened under the wrong key")
	}
	if got != nil {
		t.Errorf("bytes returned alongside an error: %q", got)
	}
}

func TestUnknownKeyIDIsNamedButNotOpened(t *testing.T) {
	k := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})

	_, err := k.Open(banking.FromStored("retired", []byte("whatever")), rowA)
	if err == nil {
		t.Fatal("a value with an unknown key id opened")
	}
	if !strings.Contains(err.Error(), "retired") {
		t.Errorf("error does not name the missing key, so an operator cannot act on it: %v", err)
	}
}

func TestTamperedCiphertextDoesNotOpen(t *testing.T) {
	k := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})
	sealed, err := k.Seal([]byte("secret"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	for _, tc := range []struct {
		name string
		ct   []byte
	}{
		{"a flipped bit in the body", flip(sealed.Ciphertext(), len(sealed.Ciphertext())-1)},
		{"a flipped bit in the nonce", flip(sealed.Ciphertext(), 0)},
		{"truncated", sealed.Ciphertext()[:len(sealed.Ciphertext())-1]},
		{"empty", []byte{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := k.Open(banking.FromStored("k1", tc.ct), rowA); err == nil {
				t.Error("opened")
			}
		})
	}
}

func flip(b []byte, i int) []byte {
	out := bytes.Clone(b)
	out[i] ^= 0x01
	return out
}

// Sealing twice must not produce the same bytes, or equal balances at equal
// banks would be visible as equal ciphertexts in a dump.
func TestSealingIsNotDeterministic(t *testing.T) {
	k := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})

	first, err := k.Seal([]byte("same"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	second, err := k.Seal([]byte("same"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if bytes.Equal(first.Ciphertext(), second.Ciphertext()) {
		t.Error("sealing the same plaintext twice produced identical ciphertext")
	}
}

func TestSealingRefusesAnOwnerlessValue(t *testing.T) {
	k := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})
	if _, err := k.Seal([]byte("secret"), ""); err == nil {
		t.Fatal("sealed with no owning row, so nothing binds the value to its row")
	}
}

func TestZeroSealedIsTheAbsenceOfAValue(t *testing.T) {
	var s banking.Sealed
	if !s.IsZero() {
		t.Error("the zero Sealed does not report itself as absent")
	}
	k := keyring(t, banking.Key{ID: "k1", Material: testKey(1)})
	if _, err := k.Open(s, rowA); err == nil {
		t.Error("opening an absent value succeeded")
	}
}

func TestNewKeyringRejects(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []banking.Key
		want string
	}{
		{"no keys at all", nil, "at least one key"},
		{"an empty key id", []banking.Key{{ID: "", Material: testKey(1)}}, "empty key id"},
		{"a short key", []banking.Key{{ID: "k1", Material: []byte("too short")}}, "needs 32 bytes"},
		{
			"the same id twice",
			[]banking.Key{{ID: "k1", Material: testKey(1)}, {ID: "k1", Material: testKey(2)}},
			"declared twice",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := banking.NewKeyring(tc.keys)
			if err == nil {
				t.Fatal("accepted")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not explain the problem (want %q)", err, tc.want)
			}
		})
	}
}

// Every problem, not only the first: one restart should be enough to see them
// all, the way config.Load already behaves.
func TestNewKeyringReportsEveryProblem(t *testing.T) {
	_, err := banking.NewKeyring([]banking.Key{
		{ID: "", Material: testKey(1)},
		{ID: "k2", Material: []byte("short")},
	})
	if err == nil {
		t.Fatal("accepted")
	}
	if n := len(errors.Join(err).(interface{ Unwrap() []error }).Unwrap()); n == 0 {
		t.Fatal("no joined errors")
	}
	for _, want := range []string{"empty key id", "needs 32 bytes"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q: %v", want, err)
		}
	}
}
