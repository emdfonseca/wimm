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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// Enable Banking authenticates every request with a short-lived RS256 JWT.
//
// The claim values are fixed by the gateway and two of them read backwards: the
// issuer is the gateway's own domain rather than the caller, and the
// application id travels in the JOSE header's kid rather than in a claim.
// They are constants here so a future reader does not "correct" them.
const (
	tokenIssuer   = "enablebanking.com"
	tokenAudience = "api.enablebanking.com"

	// TokenLifetime is the maximum the gateway accepts. Asking for less buys
	// nothing: the token never leaves this process except on the request it
	// authenticates.
	TokenLifetime = time.Hour
)

// signer mints request tokens. It only ever signs — wimm never receives or
// verifies a JWT from anybody — which is why this is forty lines of stdlib
// rather than a dependency. Verification is the half of JWT that carries the
// algorithm-confusion and "alg: none" failure modes, and wimm does not do it.
type signer struct {
	applicationID string
	key           *rsa.PrivateKey
}

// newSigner reads the signing key from a path, refusing a file anyone but its
// owner can read. The key never appears in configuration as a value, only as a
// path (ADR 0018).
func newSigner(applicationID, keyPath string) (*signer, error) {
	if applicationID == "" {
		return nil, errors.New("enablebanking: no application id")
	}
	if err := banking.RequireSecretFile(keyPath); err != nil {
		return nil, fmt.Errorf("enablebanking signing key: %w", err)
	}

	encoded, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("enablebanking signing key: %s cannot be read: %w", keyPath, err)
	}
	key, err := parseRSAPrivateKey(encoded)
	if err != nil {
		// The path, never the contents.
		return nil, fmt.Errorf("enablebanking signing key %s: %w", keyPath, err)
	}
	return &signer{applicationID: applicationID, key: key}, nil
}

// parseRSAPrivateKey accepts either PEM encoding a control panel might hand
// out. Neither the key nor any part of it reaches the error.
func parseRSAPrivateKey(encoded []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(encoded)
	if block == nil {
		return nil, errors.New("is not PEM")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("is not an RSA private key in PKCS#1 or PKCS#8")
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("is a %T, not an RSA private key", parsed)
	}
	return key, nil
}

// token mints a bearer token valid from issuedAt.
//
// issuedAt is a parameter rather than a clock read so this is a pure function:
// the same inputs give the same token, and a test can assert the claims without
// racing a wall clock.
func (s *signer) token(issuedAt time.Time) (string, error) {
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": s.applicationID,
	}
	iat := issuedAt.Unix()
	claims := map[string]any{
		"iss": tokenIssuer,
		"aud": tokenAudience,
		"iat": iat,
		"exp": issuedAt.Add(TokenLifetime).Unix(),
	}

	encodedHeader, err := encodeSegment(header)
	if err != nil {
		return "", fmt.Errorf("enablebanking: encoding the token header: %w", err)
	}
	encodedClaims, err := encodeSegment(claims)
	if err != nil {
		return "", fmt.Errorf("enablebanking: encoding the token claims: %w", err)
	}

	signingInput := encodedHeader + "." + encodedClaims
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("enablebanking: signing the token: %w", err)
	}

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func encodeSegment(v any) (string, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	// Raw URL encoding: JWT segments are unpadded base64url.
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

// String keeps a signer out of any log line that formats the client holding it.
// The key is in here.
func (s *signer) String() string { return banking.Redaction }
