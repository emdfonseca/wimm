package banking

import (
	"errors"
	"log/slog"
)

// A sealed value renders as the redaction everywhere Go can be asked to print
// one. The ciphertext reveals nothing on its own, so this is not what keeps the
// secret — the key outside the database is. What it prevents is the ciphertext
// and its key id travelling into a log, a bug report or an error string, where
// they sit next to the row id that is the additional data and turn two harmless
// facts into one useful one.
//
// Every interface below is one fmt, slog or encoding/json reaches for. Missing
// any single one of them means the value prints fine in four places and leaks
// in the fifth, which is the failure this is written to make impossible.
var (
	_ interface{ String() string }   = Sealed{}
	_ interface{ GoString() string } = Sealed{}
	_ slog.LogValuer                 = Sealed{}
)

// String satisfies fmt.Stringer, which covers %v, %s and any implicit
// stringification.
func (s Sealed) String() string { return Redaction }

// GoString satisfies fmt.GoStringer, which covers %#v. Without it, %#v prints
// the struct's unexported fields whatever String does.
func (s Sealed) GoString() string { return Redaction }

// LogValue satisfies slog.LogValuer, so a sealed value attached to a log record
// is resolved to the redaction rather than to its fields.
func (s Sealed) LogValue() slog.Value { return slog.StringValue(Redaction) }

// MarshalJSON renders the redaction rather than the value. A sealed value has
// no business in a JSON document — no Connect message carries one (ADR 0018) —
// and this exists so that a struct which accidentally reaches an encoder
// produces something inert instead of something interesting.
func (s Sealed) MarshalJSON() ([]byte, error) {
	return []byte(`"` + Redaction + `"`), nil
}

// UnmarshalJSON always fails. The only thing that could arrive here is the
// redaction this type emits, and quietly decoding that into an empty Sealed
// would turn a value nobody can read into a value nobody notices is missing.
func (s *Sealed) UnmarshalJSON([]byte) error {
	return errors.New("banking: a sealed value cannot be decoded from JSON; read it from its row with FromStored")
}
