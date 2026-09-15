# 0016 · Passkey identity and enrolment links

## Status

Accepted

## Context

wimm has no notion of a person. It is self-hosted for one household, so there is
no signup funnel to build and nobody to verify an email address against: the
operator who runs the instance already knows who lives there. What the system
needs is a way for that vouching to become a credential a browser can present.

Passkeys remove the password and the reset flow that always follows it. They
bring their own constraints — a relying-party identifier that cannot change, a
challenge per ceremony, and a distinction between credentials the browser can
offer unprompted and credentials it cannot.

## Decision

**Registration and enrolment are separate operations.** Registering creates the
member; enrolling creates a credential for them. The first enrolment link is
issued at registration, and the same mechanism issues a further one later, which
is what makes adding a device and replacing a lost one fall out of the design
rather than arrive as a feature.

**The enrolment link is a bearer credential, stored hashed.** 256 bits from a
cryptographic source, rendered in the URL path, with only its hash in the
database. Single-use; issuing a new link to a member invalidates any outstanding
one, so a member has at most one live link. Default lifetime 24 hours,
configurable.

Single-use and replacement-on-issue carry most of the security; the lifetime is
the weakest of the three. A shorter one is safer and worse — the operator sends
the link over chat and the member opens it in the morning, and every expiry is
another round-trip through a person.

The value reaches browser history and anything that records paths, so it is
exchanged server-side on first load and replaced by a redirect to a path that
does not contain it. Logs never record the enrolment path with its value.

*Alternative:* a signed self-contained token with no stored row. Rejected —
single-use and invalidation both need server-side state, so the row exists
anyway, and a token that cannot be revoked is the opposite of what this wants.

**Sessions are opaque identifiers backed by a row.** `HttpOnly`, `SameSite=Lax`,
`Secure` wherever the origin is HTTPS, naming a row that carries the member, an
absolute expiry and a last-seen time. Every lifetime is evaluated against
database time, so a skewed clock on one instance cannot extend a session.

*Alternative:* a signed self-contained cookie, which avoids a read per request
and cannot be revoked before it expires. In a product whose entire recovery story
is that the operator can cut a session off, that trade is the wrong way round.
The read is a primary-key lookup on a table with one row per browser.

**Enrolment signs the member in.** The ceremony just performed user verification;
a second one seconds later re-proves the same fact and costs a prompt. The
session's authority derives from the link, and anyone holding the link could sign
in immediately afterwards regardless, so nothing is conceded.

**The relying-party identifier is configuration and a one-way door.** It and the
list of expected origins are read at startup. Changing the identifier invalidates
every passkey ever registered against it with no migration path, so it must never
be a value compiled in and adjusted later. Choose the broadest domain the product
will ever need: a parent domain covers subdomains later, a subdomain locks the
product to it forever.

**Enrolment requires a discoverable credential.** `residentKey` and user
verification are both required, and a registration is rejected unless the
authenticator confirms it created a discoverable one. This is what makes sign-in
username-less, and what stops a member enrolling a credential they can never be
offered — whose only recovery would be another operator link.

*Alternative:* accept non-discoverable credentials and ask for an email address
at sign-in. Rejected: it trades a failure the system detects once, at enrolment,
with the link still usable, for a worse experience on every future sign-in.

**All ceremony state lives in Postgres.** A challenge is a row with a lifetime in
seconds, deleted when consumed. The obvious alternative, an in-memory map, works
on one instance and fails intermittently on two — a defect found in production
rather than in tests. With challenges, links and sessions all in the database,
`wimmd` holds no request-spanning state and needs no sticky routing.

**The operator surface is a separate service on a separate listener.** Two Connect
service definitions, two ports. The operator listener binds to loopback by default
and requires a bearer credential from configuration, compared in constant time.
`wimmd` refuses to start if that listener is enabled with no credential set.

Separate definitions rather than one service with a guarded method: the
distinction is which port answers at all, and a method-level check is one
forgotten annotation away from being public. *Alternative:* mTLS, which is right
once there is more than one operator and disproportionate for one.

**Failures reveal nothing, except to the operator.** Expired, spent, replaced and
never-issued links produce one identical response in body, status and timing. A
sign-in with an unknown passkey does not say whether the passkey, the member or
neither is the problem. Registering an already-registered email *does* say so
plainly: the operator is trusted and has to act on it.

**Dependencies.** A Go WebAuthn library, a Postgres driver, and Connect. Each is
new to the repo and each is load-bearing: the ceremony, the storage and the
contract respectively.

## Consequences

- A leaked enrolment link is account takeover for its lifetime. Mitigated by
  single-use, hashing, reissue-invalidation and immediate spending, and the
  residual risk is accepted because the operator chose the delivery channel.
- Nothing is revocable yet beyond a session expiring on its own. The schema
  carries what revocation needs, so the next change is behaviour rather than a
  migration.
- Older security keys with limited resident-key storage cannot enrol. Accepted:
  platform authenticators and passkey managers are the realistic case, and the
  refusal is explicit with the link still usable.
- Postgres is a prerequisite for running anything locally. `just db-up` owns it.
- The production hostname must be settled before the first member enrols against
  a deployment, because the relying-party identifier cannot change afterwards.
- One operator credential, no rotation and no audit trail. Changing it is a
  configuration edit and a restart; nothing persists a reference to it.
- Development runs on `http://localhost`, which is a secure context, and the web
  app proxies `/api` so the browser sees one origin. A phone can be the
  authenticator over the cross-device flow; a phone browsing *to* a dev instance
  needs a hostname and a certificate, which is a deployment concern.
