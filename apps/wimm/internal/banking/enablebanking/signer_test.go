package enablebanking

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

const applicationID = "16560d8b-2dc5-4b4d-ac41-8266f62e719b"

// issuedAt is fixed so a token is the same bytes on every run.
var issuedAt = time.Date(2026, time.September, 17, 9, 0, 0, 0, time.UTC)

func generateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	// 2048 keeps the suite fast; the algorithm under test is the same.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating a key: %v", err)
	}
	return key
}

func writeKey(t *testing.T, key *rsa.PrivateKey, pkcs8 bool, mode os.FileMode) string {
	t.Helper()
	var block *pem.Block
	if pkcs8 {
		der, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			t.Fatalf("marshalling PKCS#8: %v", err)
		}
		block = &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	} else {
		block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	}

	path := filepath.Join(t.TempDir(), "signing.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(block), mode); err != nil {
		t.Fatalf("writing the key: %v", err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	return path
}

// The claims the gateway requires, asserted exactly. Two of them read
// backwards — the issuer is the gateway and the application id is in the
// header's kid — so a test that only checks "a token was produced" would not
// catch them being swapped.
func TestTheSignedTokenCarriesTheClaimsTheGatewayRequires(t *testing.T) {
	key := generateKey(t)
	s, err := newSigner(applicationID, writeKey(t, key, false, 0o600))
	if err != nil {
		t.Fatalf("newSigner: %v", err)
	}

	token, err := s.token(issuedAt)
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d segments, want 3", len(parts))
	}

	var header map[string]string
	decodeSegment(t, parts[0], &header)
	if header["alg"] != "RS256" {
		t.Errorf(`alg = %q, want "RS256"`, header["alg"])
	}
	if header["typ"] != "JWT" {
		t.Errorf(`typ = %q, want "JWT"`, header["typ"])
	}
	if header["kid"] != applicationID {
		t.Errorf("kid = %q, want the application id", header["kid"])
	}

	var claims map[string]any
	decodeSegment(t, parts[1], &claims)
	if claims["iss"] != "enablebanking.com" {
		t.Errorf(`iss = %v, want "enablebanking.com" (the gateway, not us)`, claims["iss"])
	}
	if claims["aud"] != "api.enablebanking.com" {
		t.Errorf(`aud = %v`, claims["aud"])
	}
	if got, want := int64(claims["iat"].(float64)), issuedAt.Unix(); got != want {
		t.Errorf("iat = %d, want %d", got, want)
	}
	if got, want := int64(claims["exp"].(float64)), issuedAt.Add(time.Hour).Unix(); got != want {
		t.Errorf("exp = %d, want iat + 3600", got)
	}
}

// The signature has to verify under the public half, or the gateway rejects
// every request and the failure looks like bad credentials.
func TestTheSignatureVerifiesUnderThePublicKey(t *testing.T) {
	key := generateKey(t)
	s, err := newSigner(applicationID, writeKey(t, key, false, 0o600))
	if err != nil {
		t.Fatalf("newSigner: %v", err)
	}
	token, err := s.token(issuedAt)
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	parts := strings.Split(token, ".")
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("the signature is not base64url: %v", err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))

	if err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, digest[:], signature); err != nil {
		t.Errorf("the signature does not verify: %v", err)
	}
}

// Segments are unpadded base64url. A padded segment is rejected by strict
// parsers and is the classic hand-rolled-JWT defect.
func TestSegmentsAreUnpaddedBase64URL(t *testing.T) {
	s, err := newSigner(applicationID, writeKey(t, generateKey(t), false, 0o600))
	if err != nil {
		t.Fatalf("newSigner: %v", err)
	}
	token, err := s.token(issuedAt)
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	if strings.Contains(token, "=") {
		t.Error("the token carries base64 padding")
	}
	if strings.ContainsAny(token, "+/") {
		t.Error("the token uses standard base64 rather than base64url")
	}
}

func TestBothPEMEncodingsAreAccepted(t *testing.T) {
	key := generateKey(t)
	for _, tc := range []struct {
		name  string
		pkcs8 bool
	}{{"PKCS#1", false}, {"PKCS#8", true}} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newSigner(applicationID, writeKey(t, key, tc.pkcs8, 0o600)); err != nil {
				t.Errorf("refused: %v", err)
			}
		})
	}
}

func TestNewSignerRefuses(t *testing.T) {
	key := generateKey(t)
	for _, tc := range []struct {
		name string
		id   string
		path func(t *testing.T) string
		want string
	}{
		{"no application id", "", func(t *testing.T) string { return writeKey(t, key, false, 0o600) }, "no application id"},
		{"a missing key", applicationID, func(t *testing.T) string { return filepath.Join(t.TempDir(), "absent") }, "does not exist"},
		{"a group-readable key", applicationID, func(t *testing.T) string { return writeKey(t, key, false, 0o640) }, "readable beyond its owner"},
		{"a world-readable key", applicationID, func(t *testing.T) string { return writeKey(t, key, false, 0o644) }, "readable beyond its owner"},
		{
			"a file that is not PEM",
			applicationID,
			func(t *testing.T) string {
				p := filepath.Join(t.TempDir(), "k.pem")
				if err := os.WriteFile(p, []byte("not pem at all"), 0o600); err != nil {
					t.Fatal(err)
				}
				return p
			},
			"is not PEM",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newSigner(tc.id, tc.path(t))
			if err == nil {
				t.Fatal("accepted")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not say %q", err, tc.want)
			}
		})
	}
}

// The key must not reach a log, an error or any rendering of the client that
// holds it.
func TestTheKeyNeverReachesALog(t *testing.T) {
	key := generateKey(t)
	path := writeKey(t, key, false, 0o600)
	s, err := newSigner(applicationID, path)
	if err != nil {
		t.Fatalf("newSigner: %v", err)
	}

	// A fragment of the real key material, base64 as it appears in the PEM.
	material := base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(key))[:64]

	var logged strings.Builder
	handler := slog.NewJSONHandler(&logged, nil)
	slog.New(handler).Info("signing", "signer", s)

	// Through any, so the verb is exercised rather than rewritten into a
	// String() call by a linter that cannot know that is the point.
	var value any = s

	for name, out := range map[string]string{
		"%v":     fmt.Sprintf("%v", value),
		"%s":     fmt.Sprintf("%s", value),
		"%+v":    fmt.Sprintf("%+v", value),
		"a log":  logged.String(),
		"String": s.String(),
	} {
		if strings.Contains(out, material) {
			t.Errorf("%s leaks the key: %s", name, out)
		}
		if strings.Contains(out, "PRIVATE KEY") {
			t.Errorf("%s leaks the PEM: %s", name, out)
		}
	}
	if s.String() != banking.Redaction {
		t.Errorf("String() = %q, want the redaction", s.String())
	}
}

// A refusal names the path so an operator can fix it, and never the contents.
func TestARefusalNamesThePathAndNotTheKey(t *testing.T) {
	key := generateKey(t)
	path := writeKey(t, key, false, 0o644)

	_, err := newSigner(applicationID, path)
	if err == nil {
		t.Fatal("accepted a world-readable key")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("the refusal does not name the path: %v", err)
	}
	if strings.Contains(err.Error(), "PRIVATE KEY") {
		t.Errorf("the refusal quotes the key: %v", err)
	}
}

func decodeSegment(t *testing.T, segment string, into any) {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		t.Fatalf("segment is not base64url: %v", err)
	}
	if err := json.Unmarshal(raw, into); err != nil {
		t.Fatalf("segment is not JSON: %v", err)
	}
}
