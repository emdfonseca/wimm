package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// secretBytes is 256 bits: enough that guessing an enrolment link, a ticket or
// a session identifier is not a strategy.
const secretBytes = 32

// mintSecret returns a value to hand out and the hash to store. The value is
// never stored and never read back — a copy of the database yields nothing
// that works (ADR 0016).
func mintSecret() (value string, hash []byte, err error) {
	raw := make([]byte, secretBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("minting a secret: %w", err)
	}
	value = base64.RawURLEncoding.EncodeToString(raw)
	return value, hashSecret(value), nil
}

// hashSecret is the one-way function the stored column holds. SHA-256 rather
// than a password hash: the input is 256 bits of uniform randomness, so there
// is no dictionary to slow an attacker down against, and the lookup is on the
// hot path of every request that carries a session.
func hashSecret(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}
