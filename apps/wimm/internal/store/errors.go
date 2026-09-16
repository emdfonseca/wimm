package store

import "errors"

// ErrNotFound is returned when a row a caller named does not exist. Callers
// branch on it; nothing here knows about Connect codes or HTTP status.
var ErrNotFound = errors.New("not found")

// ErrEmailAlreadyRegistered is returned when a member already holds the
// address. This is the one failure the operator is told about plainly: they
// are trusted and have to act on it.
var ErrEmailAlreadyRegistered = errors.New("email already registered")
