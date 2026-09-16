// Package authenticatortest is a software authenticator: enough of one to
// drive a real WebAuthn ceremony end to end in a test.
//
// It exists because the interesting behaviour here — a refused registration
// leaving the link usable, a passkey this instance has never seen, a signature
// counter moving — only happens against a ceremony that actually verifies. A
// fake that returns a canned credential would assert the mapping and nothing
// the library does.
package authenticatortest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// Authenticator holds one key pair and answers ceremonies with it.
type Authenticator struct {
	key          *ecdsa.PrivateKey
	credentialID []byte
	signCount    uint32

	// Discoverable is what the credProps extension reports back. Setting it
	// false is how a test drives the refusal of a passkey the member could
	// never be offered at sign-in.
	Discoverable bool
	// ReportCredProps false omits the extension output entirely, which is what
	// a client that does not implement it does.
	ReportCredProps bool
}

// New returns an authenticator that saves discoverable credentials.
func New(t *testing.T) *Authenticator {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating a key: %v", err)
	}
	id := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		t.Fatalf("generating a credential id: %v", err)
	}

	return &Authenticator{
		key:             key,
		credentialID:    id,
		Discoverable:    true,
		ReportCredProps: true,
	}
}

// CredentialID is the identifier this authenticator presents.
func (a *Authenticator) CredentialID() []byte { return a.credentialID }

type creationOptions struct {
	RelyingParty struct {
		ID string `json:"id"`
	} `json:"rp"`
	User struct {
		ID string `json:"id"`
	} `json:"user"`
	Challenge string `json:"challenge"`
}

type requestOptions struct {
	Challenge string `json:"challenge"`
	RPID      string `json:"rpId"`
}

// Register answers a registration ceremony, returning the
// RegistrationResponseJSON the browser would send.
func (a *Authenticator) Register(t *testing.T, optionsJSON, origin string) string {
	t.Helper()

	var opts creationOptions
	if err := json.Unmarshal([]byte(optionsJSON), &opts); err != nil {
		t.Fatalf("reading creation options: %v", err)
	}

	clientData := a.clientData(t, "webauthn.create", opts.Challenge, origin)
	authData := a.authenticatorData(t, opts.RelyingParty.ID, true)

	attestation, err := cbor.Marshal(map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	})
	if err != nil {
		t.Fatalf("encoding the attestation object: %v", err)
	}

	response := map[string]any{
		"id":    b64(a.credentialID),
		"rawId": b64(a.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"attestationObject": b64(attestation),
			"transports":        []string{"internal"},
		},
		"clientExtensionResults": a.extensionResults(),
	}
	return marshal(t, response)
}

// Authenticate answers an authentication ceremony. userHandle is the member
// identifier the credential was created against.
func (a *Authenticator) Authenticate(t *testing.T, optionsJSON, origin, userHandle string) string {
	t.Helper()

	var opts requestOptions
	if err := json.Unmarshal([]byte(optionsJSON), &opts); err != nil {
		t.Fatalf("reading request options: %v", err)
	}

	a.signCount++
	clientData := a.clientData(t, "webauthn.get", opts.Challenge, origin)
	authData := a.authenticatorData(t, opts.RPID, false)

	hash := sha256.Sum256(clientData)
	signed := append(append([]byte{}, authData...), hash[:]...)
	digest := sha256.Sum256(signed)

	signature, err := ecdsa.SignASN1(rand.Reader, a.key, digest[:])
	if err != nil {
		t.Fatalf("signing the assertion: %v", err)
	}

	response := map[string]any{
		"id":    b64(a.credentialID),
		"rawId": b64(a.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"authenticatorData": b64(authData),
			"signature":         b64(signature),
			"userHandle":        b64([]byte(userHandle)),
		},
		"clientExtensionResults": map[string]any{},
	}
	return marshal(t, response)
}

func (a *Authenticator) extensionResults() map[string]any {
	if !a.ReportCredProps {
		return map[string]any{}
	}
	return map[string]any{"credProps": map[string]any{"rk": a.Discoverable}}
}

func (a *Authenticator) clientData(t *testing.T, ceremonyType, challenge, origin string) []byte {
	t.Helper()
	return []byte(marshal(t, map[string]any{
		"type":        ceremonyType,
		"challenge":   challenge,
		"origin":      origin,
		"crossOrigin": false,
	}))
}

// authenticatorData is the flags, counter and — for a registration — the
// attested credential data the library parses the public key out of.
func (a *Authenticator) authenticatorData(t *testing.T, rpID string, attested bool) []byte {
	t.Helper()

	rpHash := sha256.Sum256([]byte(rpID))
	out := append([]byte{}, rpHash[:]...)

	// UP | UV | BE | BS, and AT when attested credential data follows.
	flags := byte(0x01 | 0x04 | 0x08 | 0x10)
	if attested {
		flags |= 0x40
	}
	out = append(out, flags)

	counter := make([]byte, 4)
	binary.BigEndian.PutUint32(counter, a.signCount)
	out = append(out, counter...)

	if !attested {
		return out
	}

	out = append(out, make([]byte, 16)...) // AAGUID: all zeroes

	idLen := make([]byte, 2)
	binary.BigEndian.PutUint16(idLen, uint16(len(a.credentialID)))
	out = append(out, idLen...)
	out = append(out, a.credentialID...)

	return append(out, a.coseKey(t)...)
}

// coseKey is the ES256 public key in the COSE_Key form WebAuthn expects.
func (a *Authenticator) coseKey(t *testing.T) []byte {
	t.Helper()

	key := map[int]any{
		1:  2,            // kty: EC2
		3:  -7,           // alg: ES256
		-1: 1,            // crv: P-256
		-2: pad(a.key.X), // x
		-3: pad(a.key.Y), // y
	}
	encoded, err := cbor.Marshal(key)
	if err != nil {
		t.Fatalf("encoding the COSE key: %v", err)
	}
	return encoded
}

// pad left-pads a coordinate to the 32 bytes P-256 requires; big.Int drops
// leading zeroes and a short coordinate is rejected.
func pad(n *big.Int) []byte {
	b := n.Bytes()
	if len(b) >= 32 {
		return b
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func marshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("encoding: %v", err)
	}
	return string(b)
}
