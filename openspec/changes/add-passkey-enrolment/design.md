## Context

Nothing runs yet. There is no Go module, no database, no web app and no Connect
contract — only the design system in `packages/ui`. See proposal.md · Why for
the motivation. This design therefore has to settle more than identity: it picks
the shape every later change inherits.

Constraints that shape it, from ADR 0001: Go services and the deployable web app
under `apps/`, importable code under `packages/`, Connect over protobuf between
our own callers, Postgres with goose migrations owned by the service that owns
the database, devbox for the toolchain and `just check <dir>` as the only gate.

## Language

The nouns below are binding on the specs, the interface copy and the Go
identifiers. Where the code needs a different word it is a signal the model is
wrong, not that the code needs a synonym.

- **Operator** — the person who runs this instance and has shell access to it.
  Not a role a member can hold, not a row in the database.
- **Member** — a person registered to use this instance. The specs say
  *household member*; `Member` in code, never `User` or `Account`.
- **Enrolment link** — the one-time URL that lets a member create a passkey.
  Never *invite*, never *token* in anything a person reads.
- **Passkey** — what the member has. `Credential` in code where the WebAuthn
  vocabulary makes it clearer, never in interface copy.
- **Session** — the thing that makes a browser signed in.

## Goals / Non-Goals

**Goals:**

- A shape that works unchanged with more than one instance of the service
  behind a load balancer, without sticky routing.
- Credential loss is an operator action, not an engineering one.
- Configuration, not code, carries everything that differs between a laptop and
  a deployment.

**Non-Goals:**

- Authorisation. Every member is equal; there are no roles and nothing to
  authorise against yet.
- Multi-tenancy. One instance serves one household.
- Operator credential rotation, audit trails, rate limiting.

## Decisions

### One Go module, two binaries

`apps/wimm` holds `cmd/wimmd` (the service) and `cmd/wimmctl` (the operator
CLI). The CLI calls the service over Connect rather than reaching the database,
so there is exactly one place where an enrolment link can be minted.

A single module makes the operator contract a compile-time agreement between the
two binaries with no version skew, and gives `just check apps/wimm` one target.
*Alternative:* separate modules, which only pays off if the CLI ships on its own
schedule — it does not. *Alternative:* the CLI writing to Postgres directly,
rejected because the invariants around link issuance would then live in two
places.

### Public and operator surfaces are separate services on separate listeners

Two Connect service definitions in `packages/contracts`, served on two ports.
The operator listener binds to loopback by default and requires a bearer
credential from configuration, compared in constant time. `wimmd` refuses to
start if that listener is enabled without a credential configured.

Separate definitions rather than one service with a guarded method: the
distinction is *which port answers at all*, and a method-level check is one
forgotten annotation away from being public. *Alternative:* mTLS, which is the
right answer once there is more than one operator and is disproportionate for
one.

### The enrolment link is a bearer credential, stored hashed

A high-entropy random value, 256 bits from a cryptographic source, rendered in
the URL path. Only its hash is stored, so a database copy does not yield working
links. Single-use, and issuing a new link to a member invalidates any
outstanding one, keeping at most one live bearer credential per member.

Default lifetime 24 hours, configurable. Shorter is safer and worse: the
operator sends the link over chat and the member may not open it until morning,
and every expiry is another operator round-trip. Single-use and
replacement-on-issue are doing most of the security work here; the lifetime is
the weakest of the three controls.

*Alternative:* a signed payload carrying the member's identity with no stored
row. Rejected — single-use and invalidation both need server-side state, so the
row exists either way, and a self-contained token that cannot be revoked is the
opposite of what this design wants.

**The value appears in the URL**, so it reaches browser history, and any
referrer or log that records paths. Mitigated by exchanging it server-side on
first load and redirecting to a path that does not contain it, so what lingers
in history is spent. Logs must not record the enrolment path with its value.

### WebAuthn parameters are policy, and the relying-party identifier is a one-way door

Relying-party identifier and the list of expected origins are configuration,
read at startup. Changing the identifier invalidates every passkey ever
registered against it, with no migration path, so it must never be a value
compiled in and adjusted later.

Registration requires a discoverable credential and user verification, and the
result is rejected unless the authenticator confirms it created a discoverable
one. This is what makes sign-in username-less: without it a member can enrol a
credential they cannot later be offered, and the only recovery is another
operator link.

*Alternative:* accepting non-discoverable credentials and asking for an email
address at sign-in. Rejected — it trades a failure the system can detect at
enrolment for a worse experience on every future sign-in.

### All ceremony state lives in Postgres

The challenge issued for a registration or an authentication is a row with a
lifetime measured in seconds, deleted when consumed. The obvious alternative —
an in-memory map — works perfectly on one instance and fails intermittently on
two, which is the kind of defect that is found in production rather than in
tests. With challenges, links and sessions all in the database, `wimmd` holds no
request-spanning state and needs no sticky routing.

Every lifetime is evaluated against database time, not process time, so a
skewed clock on one instance cannot extend a link or a session.

### Sessions are opaque identifiers backed by a row

An opaque random identifier in an `HttpOnly`, `SameSite=Lax` cookie, marked
`Secure` wherever the origin is HTTPS, naming a session row that carries the
member, an absolute expiry and a last-seen time. The identifier is regenerated
when a session is created so that enrolment does not continue a pre-existing
one.

*Alternative:* a signed self-contained cookie, which avoids a database read per
request and cannot be revoked before it expires. In a product whose entire
recovery story is "the operator can cut this off", that trade is the wrong way
round. The read is a primary-key lookup on a table with one row per browser.

### Enrolment signs the member in

The registration ceremony has just performed user verification. A second
ceremony seconds later re-proves the same fact and costs a prompt. The session's
authority derives from the enrolment link rather than from the passkey — but
anyone holding the link could sign in immediately afterwards anyway, so nothing
is conceded.

### After sign-in, return to the intended page

The path a member was trying to reach is remembered server-side against the
sign-in attempt rather than carried in the URL, so there is nothing to tamper
with and no open redirect to defend against. Anything not recognised as an
in-app path falls back to the landing page.

### Failures that reveal nothing

Expired, spent, replaced and never-issued enrolment links produce one indistinguishable
response. A sign-in with an unknown passkey does not say whether the passkey, the
member or neither is the problem. Registering an already-registered email *does*
tell the operator plainly — the operator is trusted and needs to act on it.

### Development runs on localhost without TLS

`http://localhost` is a secure context, so WebAuthn and `Secure` cookies both
work. The web app proxies API calls so that the browser sees a single origin,
which keeps the relying-party origin check, the cookie and CORS trivial.

A phone can still be used as the authenticator over the cross-device (QR) flow,
because the browser doing the ceremony is the laptop's. A phone browsing *to*
the dev instance is out of reach until there is a hostname and a certificate;
that is a deployment concern, not a local-development one.

*Alternative:* locally trusted certificates via mkcert. Rejected for now — it
buys nothing on localhost and does not solve the phone-as-browser case either,
which needs a hostname regardless.

## Risks / Trade-offs

- **A leaked enrolment link is account takeover.** → Single-use, 24-hour
  default, invalidated by reissue, stored hashed, and spent immediately on first
  load. Residual risk accepted: the operator chose the channel, and this is
  stated in the proposal rather than assumed away.
- **There is no revocation until the next change.** A member whose device is
  lost has a live session and a working passkey until the session expires. →
  Session lifetimes are bounded rather than indefinite, and the schema carries
  what revocation will need so the next change is behaviour, not migration.
- **Requiring discoverable credentials excludes some authenticators**, notably
  older security keys with limited resident-key storage. → Accepted: platform
  authenticators and passkey managers are the realistic case, and the failure is
  explicit at enrolment with the link still usable.
- **The relying-party identifier cannot change after the first member enrols.**
  → Choose the broadest domain the product will ever need; a parent domain
  covers subdomains later, a subdomain locks the product to it forever. Recorded
  in the ADR so it is not rediscovered.
- **One operator credential, no rotation.** → Changing it is a configuration
  edit and a restart; nothing persists a reference to it.
- **Postgres becomes a prerequisite for running anything locally.** → Accepted;
  it was always going to be, and the alternative is a second storage path that
  exists only in development.
- **Safari may not honour `http://localhost` for WebAuthn as Chrome does.** →
  Unverified. If it does not, locally trusted certificates are a small, isolated
  addition; nothing in this design depends on their absence.

## Migration Plan

Nothing to migrate — this is the first schema. Deployment is goose migrations
followed by `wimmd`, then the operator registering themselves the same way as
anyone else; there is no bootstrap user and no seeded account, because operator
authority comes from configuration rather than from a row.

Rollback is dropping the database, which is only true for this change and stops
being true the moment real data exists.

## Open Questions

- The production hostname and its certificate. Deferrable because it is
  configuration, but it must be settled before the first member enrols against a
  deployment, since the relying-party identifier cannot change afterwards.
- Whether enrolment links are eventually delivered by email from the service or
  continue to be passed on by the operator. The link is issued the same way
  either case, so this changes nothing here.
