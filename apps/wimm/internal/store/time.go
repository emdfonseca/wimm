package store

import "time"

// Time is a timestamp read from the database. The named type exists so that a
// value read from Postgres cannot be confused at a call site with one taken
// from the process clock.
type Time struct {
	T time.Time
}
