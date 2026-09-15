package rpc

import (
	"net/http"
	"strings"
	"time"
)

// Cookie names. Both are HttpOnly: nothing in the page has any use for either
// value, and script that can read a session identifier can take it.
const (
	SessionCookie   = "wimm_session"
	EnrolmentCookie = "wimm_enrolment"
)

// cookiePolicy decides the attributes a cookie is written with.
type cookiePolicy struct {
	// secure marks cookies Secure. It follows the origin: http://localhost is
	// a secure context for WebAuthn but not for a Secure cookie, and a Secure
	// cookie there would simply never be stored.
	secure bool
}

func policyForOrigins(origins []string) cookiePolicy {
	for _, o := range origins {
		if !strings.HasPrefix(o, "https://") {
			return cookiePolicy{secure: false}
		}
	}
	return cookiePolicy{secure: len(origins) > 0}
}

func (p cookiePolicy) set(name, value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
		// Nothing in this product is reached from another site, and a session
		// that travels cross-site is a session that can be used cross-site.
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   p.secure,
		Expires:  expires,
	}
}

func (p cookiePolicy) clear(name string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   p.secure,
		MaxAge:   -1,
	}
}

// cookieValue reads one cookie out of a request's headers. Connect hands the
// handler headers rather than an *http.Request, so the parse is done here.
func cookieValue(header http.Header, name string) string {
	for _, c := range (&http.Request{Header: header}).Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}
