package identity

import "errors"

// The conditions a caller branches on. Domain code never knows about Connect
// codes or HTTP status; the handler package maps these once.
var (
	// ErrEmailAlreadyRegistered is told to the operator plainly: they are
	// trusted and have to act on it.
	ErrEmailAlreadyRegistered = errors.New("that email address is already registered")

	// ErrMemberNotFound is told to the operator, for the same reason.
	ErrMemberNotFound = errors.New("nobody is registered under that email address")

	// ErrNameMissing names the part that is missing.
	ErrNameMissing = errors.New("a first and last name are both required")

	// ErrEnrolmentLinkUnusable is the one answer for expired, spent, replaced
	// and never-issued. Nothing distinguishes them, in body, status or timing.
	ErrEnrolmentLinkUnusable = errors.New("that enrolment link cannot be used")

	// ErrNotDiscoverable is the refusal of a passkey the member could never be
	// offered at sign-in.
	ErrNotDiscoverable = errors.New("that device did not save a passkey you can sign in with")

	// ErrPasskeyNotRecognised is returned for a credential this instance has
	// no record of. It does not say whether the passkey, the member or neither
	// is the problem.
	ErrPasskeyNotRecognised = errors.New("that passkey is not enrolled here")

	// ErrNotSignedIn means there is no live session on this browser.
	ErrNotSignedIn = errors.New("not signed in")
)
