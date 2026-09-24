package identity

import (
	"encoding/json"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/emdfonseca/wimm/apps/wimm/internal/store"
)

// member adapts a registered person to the shape the WebAuthn library expects.
//
// The user handle is the member's identifier, so an assertion that carries a
// handle names a member directly.
type member struct {
	row         store.Member
	credentials []webauthn.Credential
}

func newWebAuthnUser(m store.Member, rows []store.Credential) (*member, error) {
	creds := make([]webauthn.Credential, 0, len(rows))
	for _, r := range rows {
		c, err := decodeCredential(r)
		if err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return &member{row: m, credentials: creds}, nil
}

func (m *member) WebAuthnID() []byte                         { return []byte(m.row.ID) }
func (m *member) WebAuthnName() string                       { return m.row.Email }
func (m *member) WebAuthnDisplayName() string                { return m.row.FirstName + " " + m.row.LastName }
func (m *member) WebAuthnCredentials() []webauthn.Credential { return m.credentials }

func decodeCredential(r store.Credential) (webauthn.Credential, error) {
	var c webauthn.Credential
	if err := json.Unmarshal(r.Data, &c); err != nil {
		return webauthn.Credential{}, fmt.Errorf("reading a stored passkey: %w", err)
	}
	return c, nil
}
