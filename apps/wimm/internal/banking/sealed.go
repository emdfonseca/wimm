// Package banking is wimm's own vocabulary for reading accounts at a bank.
// Nothing here names a gateway: adapters live under internal/banking/<gateway>
// and a lint rule refuses an import of one from anywhere else (ADR 0018).
package banking

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

// Redaction is what a sealed value renders as, everywhere Go can be asked to
// print one. It is a constant so a test can assert on it rather than on a
// string literal repeated in three places.
const Redaction = "[sealed]"

// ErrSealed is returned by every accessor that would otherwise hand back
// something readable by accident.
var ErrSealed = errors.New("banking: a sealed value cannot be read except by opening it")

// Sealed is a value that could reach a bank — a gateway session identifier, an
// account's per-session uid — held encrypted with the key that opens it named
// beside it.
//
// Its fields are unexported on purpose. A Sealed can only be built by sealing
// plaintext or by reading a row back, and its contents can only be reached by
// opening it against the same additional data it was sealed with, so a
// ciphertext lifted from one row cannot be replayed into another (ADR 0018).
type Sealed struct {
	keyID      string
	ciphertext []byte
}

// FromStored rebuilds a Sealed from the two columns a row carries. It performs
// no cryptography: an unopenable value is a value that fails at Open, not one
// that fails to load, because a row with a key id nothing recognises still has
// to be selectable, deletable and countable.
func FromStored(keyID string, ciphertext []byte) Sealed {
	return Sealed{keyID: keyID, ciphertext: ciphertext}
}

// KeyID names the key this value was sealed with. It is stored beside the
// ciphertext so a key can be rotated without a flag day.
func (s Sealed) KeyID() string { return s.keyID }

// Ciphertext is the bytes a row stores. Safe to log in the sense that it
// reveals nothing, and pointless to log for the same reason.
func (s Sealed) Ciphertext() []byte { return s.ciphertext }

// IsZero reports whether this is the absence of a sealed value rather than a
// sealed empty one. A connection whose access has ended keeps its row and
// loses its secrets, so "there is nothing here" is a state the model has.
func (s Sealed) IsZero() bool { return s.keyID == "" && len(s.ciphertext) == 0 }

// Key is one sealing key: an identifier that reaches the database and 32 bytes
// that never do.
type Key struct {
	ID       string
	Material []byte
}

// KeySize is AES-256: 32 bytes, and exactly 32. A shorter key is a
// configuration error rather than something to stretch.
const KeySize = 32

// Keyring holds every key that can open a stored value and the one key new
// values are sealed with. Rotation is adding a key at the front and keeping
// the old one until nothing references it.
type Keyring struct {
	currentID string
	aeads     map[string]cipher.AEAD
}

// NewKeyring builds a keyring from keys in priority order: the first seals,
// and all of them open. It reports every problem it finds rather than the
// first, so one restart is enough to see them all.
func NewKeyring(keys []Key) (*Keyring, error) {
	if len(keys) == 0 {
		return nil, errors.New("banking: a keyring needs at least one key")
	}

	var problems []error
	seen := make(map[string]bool, len(keys))
	aeads := make(map[string]cipher.AEAD, len(keys))

	for i, k := range keys {
		switch {
		case k.ID == "":
			problems = append(problems, fmt.Errorf("key %d: an empty key id cannot be stored beside a value", i))
			continue
		case seen[k.ID]:
			problems = append(problems, fmt.Errorf("key %q: declared twice, so which one opens a value is ambiguous", k.ID))
			continue
		case len(k.Material) != KeySize:
			// The length is named and the material is not.
			problems = append(problems, fmt.Errorf("key %q: needs %d bytes, got %d", k.ID, KeySize, len(k.Material)))
			continue
		}
		seen[k.ID] = true

		block, err := aes.NewCipher(k.Material)
		if err != nil {
			problems = append(problems, fmt.Errorf("key %q: %w", k.ID, err))
			continue
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			problems = append(problems, fmt.Errorf("key %q: %w", k.ID, err))
			continue
		}
		aeads[k.ID] = aead
	}

	if len(problems) > 0 {
		return nil, errors.Join(problems...)
	}
	return &Keyring{currentID: keys[0].ID, aeads: aeads}, nil
}

// CurrentKeyID is the key new values are sealed with.
func (k *Keyring) CurrentKeyID() string { return k.currentID }

// Seal encrypts plaintext against owner, which is the uuid of the row the value
// will live on. The owner is authenticated but not encrypted: it is already in
// the row, and binding to it is what stops a ciphertext being moved between
// rows.
func (k *Keyring) Seal(plaintext []byte, owner string) (Sealed, error) {
	if owner == "" {
		return Sealed{}, errors.New("banking: sealing needs the owning row's id, so a value cannot be replayed into another row")
	}
	aead, ok := k.aeads[k.currentID]
	if !ok {
		return Sealed{}, fmt.Errorf("banking: the current key %q is not in the keyring", k.currentID)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return Sealed{}, fmt.Errorf("banking: generating a nonce: %w", err)
	}

	// The nonce is prepended rather than stored in its own column: it is
	// meaningless on its own and always travels with the ciphertext.
	sealed := aead.Seal(nonce, nonce, plaintext, []byte(owner))
	return Sealed{keyID: k.currentID, ciphertext: sealed}, nil
}

// Open decrypts a stored value against the row it belongs to. Every failure —
// an unknown key, the wrong key, the wrong owner, a truncated or altered
// ciphertext — returns an error and no bytes. There is no partial result and
// no way to tell the failures apart from the outside, because the caller can
// act on none of the distinctions.
func (k *Keyring) Open(s Sealed, owner string) ([]byte, error) {
	if owner == "" {
		return nil, errors.New("banking: opening needs the owning row's id")
	}
	if s.IsZero() {
		return nil, errors.New("banking: there is no sealed value here to open")
	}

	aead, ok := k.aeads[s.keyID]
	if !ok {
		// The key id is named because an operator has to know which key to
		// restore. It is an identifier, not material.
		return nil, fmt.Errorf("banking: no key %q in the keyring", s.keyID)
	}
	if len(s.ciphertext) < aead.NonceSize() {
		return nil, errors.New("banking: sealed value is too short to contain a nonce")
	}

	nonce, body := s.ciphertext[:aead.NonceSize()], s.ciphertext[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, body, []byte(owner))
	if err != nil {
		return nil, errors.New("banking: sealed value did not open")
	}
	return plaintext, nil
}
