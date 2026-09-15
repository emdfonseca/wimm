## Why

wimm has a design system and no product. Nothing runs, nothing stores anything,
and there is no way for a person to be known to the system at all. Before any
ledger behaviour can be built there has to be an answer to "who is this", and in
a self-hosted household product that answer starts with the operator vouching
for someone rather than with self-service signup.

This change builds that answer end to end: the operator registers a household
member from the command line, hands them a link through whatever channel they
already use, and the member enrols a passkey and is in. It is the smallest slice
that puts a real person behind a real session, and it is the slice every later
change depends on.

## What Changes

- **Registration and enrolment are separate.** Registering a person creates the
  person. Enrolling creates a credential for them. The first enrolment link is
  issued at registration, but the same mechanism issues a second link later —
  which is what makes adding a device, and recovering a lost one, free rather
  than a new feature.
- **The operator CLI registers a person and prints an enrolment URL.** Email,
  first name and last name are operator assertions; nothing verifies them.
- **The enrolment link is a bearer credential.** Whoever holds it becomes that
  person. It is single-use, short-lived, and stored hashed. Delivery is manual
  for now; an email sender later consumes the same primitive without touching
  enrolment.
- **Enrolment requires a discoverable credential.** `residentKey` and user
  verification are both required, and a registration whose authenticator reports
  it did not create a discoverable credential is rejected. This is what makes
  username-less sign-in possible and what stops someone enrolling a credential
  they can never sign in with.
- **Enrolment signs the person in.** The WebAuthn ceremony has just verified
  them; a second ceremony seconds later proves nothing new.
- **Returning members sign in with no username.** One control, the browser
  offers the passkey, done. No email field and no lookup, because the credential
  is discoverable.
- **Sessions are server-side and revocable.** An opaque identifier in an
  HttpOnly cookie, backed by a row. Revocation is the entire recovery story for
  a lost device, and a session that cannot be killed before it expires forecloses
  it.
- **The service holds no ceremony state in memory.** WebAuthn challenges live in
  Postgres with the rest, expire in seconds, and are deleted on use, so a second
  instance needs no sticky routing.
- **The relying-party identifier and expected origins are configuration.**
  Changing the RP identifier invalidates every credential ever registered, with
  no migration, so it is a value the deployment sets rather than a value the code
  knows. Development runs on `localhost` over plain HTTP, which is a secure
  context; the web app proxies the API so the browser sees one origin.
- **The admin surface is separated from the public one** and authenticated with
  an operator token from configuration. The service refuses to start if the admin
  listener is enabled without one.

Not in this change: revoking sessions or removing a credential, email delivery,
roles or authorisation beyond "is signed in", and the production hostname and
its certificate.

## Capabilities

### New Capabilities

- `identity/operator-registration`: the operator registers a household member
  and is given an enrolment link to share, and can issue a further link to an
  already-registered member.
- `identity/passkey-enrolment`: a person follows an enrolment link, creates a
  passkey, and is signed in — including what they see when the link is expired,
  already used, or unknown.
- `identity/passkey-sign-in`: a returning member signs in with their passkey
  without typing anything, and an expired session sends them there rather than
  to a dead end.

### Modified Capabilities

None. This is the first capability in the project.

## Impact

**New code.** `apps/wimm`, one Go module with two entry points: `cmd/wimmd`, the
service, and `cmd/wimmctl`, the operator CLI. A single module makes the admin
contract a compile-time agreement between them with no version skew.
`apps/web`, the SvelteKit app, carrying the enrolment page, the sign-in page and
the signed-in landing. `packages/ui` gains the components those pages instance,
built before the pages that use them.

**Contracts.** The first Connect service definition, generated into
`packages/contracts`. Public and admin services are separate definitions because
they are separately exposed.

**Data.** The first Postgres database and the first goose migrations, owned by
`wimmd`: people, enrolment links, credentials, ceremony challenges and sessions.

**Dependencies.** A Go WebAuthn library, a Postgres driver, and Connect — each
new to the repo, and together the reason this change needs an architecture
decision recorded alongside it covering the enrolment-link trust model, the
session design, the relying-party identifier as a one-way door, and the admin
listener.

**Development setup.** Postgres becomes something a contributor needs running,
and the web app proxies the API so that WebAuthn and the session cookie see a
single origin.
