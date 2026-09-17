package banking_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/xuuid/wimm/apps/wimm/internal/banking"
)

// connection stands in for any struct that carries a sealed value and might be
// logged whole.
type connection struct {
	ID      string
	Bank    string
	Session banking.Sealed
}

// Every way Go offers to render a value, against a struct holding a sealed one.
// A single uncovered verb is the whole defect.
func TestASealedValueNeverRenders(t *testing.T) {
	k, err := banking.NewKeyring([]banking.Key{{ID: "k1", Material: bytes.Repeat([]byte{7}, banking.KeySize)}})
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	sealed, err := k.Seal([]byte("gateway-session-abc123"), rowA)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	c := connection{ID: rowA, Bank: "Montepio", Session: sealed}
	ct := sealed.Ciphertext()

	var logged bytes.Buffer
	slog.New(slog.NewJSONHandler(&logged, nil)).Info("connected", "connection", c, "session", sealed)

	asJSON, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Through any, so the verb is exercised rather than rewritten into a
	// String() call by a linter that cannot know that is the point.
	var field any = c.Session

	renderings := map[string]string{
		"%v":              fmt.Sprintf("%v", c),
		"%+v":             fmt.Sprintf("%+v", c),
		"%#v":             fmt.Sprintf("%#v", c),
		"%s on the field": fmt.Sprintf("%s", field),
		"%v on the field": fmt.Sprintf("%v", field),
		"%q on the field": fmt.Sprintf("%q", field),
		"Sprint":          fmt.Sprint(c),
		"json.Marshal":    string(asJSON),
		"slog":            logged.String(),
	}

	// Three encodings, because a leak that arrives base64'd is still a leak.
	forbidden := map[string]string{
		"raw bytes": string(ct),
		"hex":       hex.EncodeToString(ct),
		"base64":    base64.StdEncoding.EncodeToString(ct),
		// %#v renders bytes as a Go literal, which none of the encodings above
		// would match. These field names appear only if the struct was printed.
		"the ciphertext field": "ciphertext:",
		"the key id field":     "keyID:",
	}

	for name, out := range renderings {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(out, banking.Redaction) {
				t.Errorf("no redaction in %s output: %s", name, out)
			}
			for enc, needle := range forbidden {
				if len(needle) > 0 && strings.Contains(out, needle) {
					t.Errorf("%s output leaks the ciphertext as %s: %s", name, enc, out)
				}
			}
			if strings.Contains(out, "gateway-session-abc123") {
				t.Errorf("%s output leaks the plaintext: %s", name, out)
			}
		})
	}
}

// The key id is deliberately reachable through its accessor — an operator needs
// it — but it must not ride along in a rendering, where it would sit next to
// the row id that is the value's additional data.
func TestRenderingDoesNotCarryTheKeyID(t *testing.T) {
	s := banking.FromStored("key-2026-01", []byte("ciphertext"))
	for _, out := range []string{
		fmt.Sprintf("%v", s),
		fmt.Sprintf("%#v", s),
		s.GoString(),
	} {
		if strings.Contains(out, "key-2026-01") {
			t.Errorf("rendering carries the key id: %s", out)
		}
	}
	if s.KeyID() != "key-2026-01" {
		t.Error("the accessor should still return the key id")
	}
}

func TestASealedValueCannotBeDecodedFromJSON(t *testing.T) {
	var s banking.Sealed
	if err := json.Unmarshal([]byte(`"`+banking.Redaction+`"`), &s); err == nil {
		t.Fatal("decoded a sealed value from its own redaction")
	}

	var c connection
	if err := json.Unmarshal([]byte(`{"ID":"x","Bank":"y","Session":"[sealed]"}`), &c); err == nil {
		t.Fatal("decoded a struct carrying a sealed value")
	}
}

// The zero value renders too: a connection whose secrets were destroyed still
// gets logged, and must not print differently from one that has them.
func TestTheZeroValueRendersAsTheRedaction(t *testing.T) {
	var s banking.Sealed
	if got := fmt.Sprintf("%v", s); got != banking.Redaction {
		t.Errorf("zero Sealed rendered as %q, want %q", got, banking.Redaction)
	}
}
