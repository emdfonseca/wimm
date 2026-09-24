package banking_test

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emdfonseca/wimm/apps/wimm/internal/banking"
)

func writeKeyFile(t *testing.T, contents string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sealing.key")
	if err := os.WriteFile(path, []byte(contents), mode); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}
	// WriteFile is subject to umask; the mode is the point of several of these
	// tests, so set it explicitly.
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	return path
}

func encodedKey(fill byte) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{fill}, banking.KeySize))
}

func TestLoadKeyringReadsAFileOwnerOnly(t *testing.T) {
	path := writeKeyFile(t, "# the current key seals; the rest only open\n2026-01 "+encodedKey(1)+"\n", 0o600)

	k, err := banking.LoadKeyring(path)
	if err != nil {
		t.Fatalf("LoadKeyring: %v", err)
	}
	if k.CurrentKeyID() != "2026-01" {
		t.Errorf("CurrentKeyID = %q, want 2026-01", k.CurrentKeyID())
	}
}

// The refusals 2.3 exists for. Each is a separate way of getting the same thing
// wrong, and a check that covers three of the four is the one that lets the
// fourth through.
func TestLoadKeyringRefuses(t *testing.T) {
	good := "2026-01 " + encodedKey(1) + "\n"

	for _, tc := range []struct {
		name string
		path func(t *testing.T) string
		want string
	}{
		{
			"a file that does not exist",
			func(t *testing.T) string { return filepath.Join(t.TempDir(), "absent.key") },
			"does not exist",
		},
		{
			"a directory",
			func(t *testing.T) string { return t.TempDir() },
			"is a directory",
		},
		{
			"a group-readable file",
			func(t *testing.T) string { return writeKeyFile(t, good, 0o640) },
			"readable beyond its owner",
		},
		{
			"a world-readable file",
			func(t *testing.T) string { return writeKeyFile(t, good, 0o604) },
			"readable beyond its owner",
		},
		{
			"a world-writable file",
			func(t *testing.T) string { return writeKeyFile(t, good, 0o602) },
			"readable beyond its owner",
		},
		{
			"a file this process cannot read",
			func(t *testing.T) string { return writeKeyFile(t, good, 0o200) },
			"cannot be opened",
		},
		{
			"an empty path",
			func(t *testing.T) string { return "" },
			"no path given",
		},
		{
			"a file with no keys in it",
			func(t *testing.T) string { return writeKeyFile(t, "# nothing but a comment\n\n", 0o600) },
			"contains no keys",
		},
		{
			"a key that is not base64",
			func(t *testing.T) string { return writeKeyFile(t, "2026-01 not-base64!!\n", 0o600) },
			"not valid base64",
		},
		{
			"a key of the wrong length",
			func(t *testing.T) string {
				return writeKeyFile(t, "2026-01 "+base64.StdEncoding.EncodeToString([]byte("short"))+"\n", 0o600)
			},
			"needs 32 bytes",
		},
		{
			"a line with no key on it",
			func(t *testing.T) string { return writeKeyFile(t, "2026-01\n", 0o600) },
			"want \"<key-id> <base64 key>\"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.path(t)
			_, err := banking.LoadKeyring(path)
			if err == nil {
				t.Fatal("accepted")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not say %q", err, tc.want)
			}
		})
	}
}

// A refusal must never quote what it refused.
func TestARefusalNeverCarriesKeyMaterial(t *testing.T) {
	material := encodedKey(9)
	path := writeKeyFile(t, "2026-01 "+material+"\n", 0o644)

	_, err := banking.LoadKeyring(path)
	if err == nil {
		t.Fatal("accepted a world-readable key file")
	}
	if strings.Contains(err.Error(), material) {
		t.Errorf("the refusal quotes the key material: %v", err)
	}
}

func TestRequireSecretFileAcceptsOwnerOnly(t *testing.T) {
	for _, mode := range []os.FileMode{0o600, 0o400} {
		path := writeKeyFile(t, "x", mode)
		if err := banking.RequireSecretFile(path); err != nil {
			t.Errorf("mode %04o refused: %v", mode, err)
		}
	}
}
