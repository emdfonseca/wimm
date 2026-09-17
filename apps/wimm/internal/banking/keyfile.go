package banking

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// SecretFileMask is the permission bits a file holding key material may not
// have. Anything readable, writable or executable by group or other fails:
// the key that opens every bank credential in the database is not a file with
// a relaxed mode and a note in the runbook.
const SecretFileMask fs.FileMode = 0o077

// RequireSecretFile reports why a path is unfit to hold key material, or nil.
// It is the check both key files go through — the sealing key here and the
// gateway's signing key in the adapter — so neither can be the one that was
// forgotten.
func RequireSecretFile(path string) error {
	if path == "" {
		return errors.New("no path given")
	}
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("%s does not exist", path)
	case err != nil:
		return fmt.Errorf("%s cannot be read: %w", path, err)
	case info.IsDir():
		return fmt.Errorf("%s is a directory", path)
	}

	if mode := info.Mode().Perm(); mode&SecretFileMask != 0 {
		return fmt.Errorf(
			"%s is readable beyond its owner (mode %04o); run: chmod 600 %s",
			path, mode, path)
	}

	// Stat says nothing about whether this process can actually open it.
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s cannot be opened: %w", path, err)
	}
	return f.Close()
}

// LoadKeyring reads sealing keys from a file, refusing one that anybody but its
// owner can read.
//
// The file is lines of "<key-id> <base64 of 32 bytes>". Blank lines and lines
// beginning with # are ignored. The first key seals new values; every key opens
// old ones, which is what makes rotation a line added at the top rather than a
// migration.
func LoadKeyring(path string) (*Keyring, error) {
	if err := RequireSecretFile(path); err != nil {
		return nil, fmt.Errorf("banking sealing key: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("banking sealing key: %s cannot be opened: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	keys, err := parseKeys(f)
	if err != nil {
		// The path is named and the contents never are.
		return nil, fmt.Errorf("banking sealing key %s: %w", path, err)
	}
	return NewKeyring(keys)
}

func parseKeys(r io.Reader) ([]Key, error) {
	var (
		keys     []Key
		problems []error
	)

	scanner := bufio.NewScanner(r)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		id, encoded, ok := strings.Cut(text, " ")
		if !ok {
			problems = append(problems, fmt.Errorf("line %d: want \"<key-id> <base64 key>\"", line))
			continue
		}
		material, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
		if err != nil {
			// Never the value, decoded or otherwise.
			problems = append(problems, fmt.Errorf("line %d: the key is not valid base64", line))
			continue
		}
		keys = append(keys, Key{ID: strings.TrimSpace(id), Material: material})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(problems) > 0 {
		return nil, errors.Join(problems...)
	}
	if len(keys) == 0 {
		return nil, errors.New("contains no keys")
	}
	return keys, nil
}
