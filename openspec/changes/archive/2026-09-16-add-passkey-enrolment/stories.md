## S1 · Register a household member and get a link to send them

**As a** household operator — the person who runs this wimm instance for the
people who live with them, with shell access to the machine it runs on
**I want** to register someone by their email and name from the command line and
be handed a link for them
**so that** I can bring a household member in through whatever channel we
already talk on, without standing up a signup page or an email server first.

### INVEST

- **Independent** — No. The link this produces goes nowhere until S2 exists.
  Shipping it alone would mean an operator holding a URL that 404s. Accepted:
  splitting the arc by actor is what makes each story testable by one person at
  one surface, and the alternative is a single story spanning a CLI and two web
  pages.
- **Negotiable** — Yes. It says the operator registers someone and receives a
  link; whether that is one command or two, and what the URL looks like, is open.
- **Valuable** — Yes. Nobody can be known to wimm at all today, and this is the
  only way anybody becomes known.
- **Estimable** — Yes. One command, one row, one signed URL.
- **Small** — Yes. One actor, one surface, one outcome.
- **Testable** — Yes. The operator runs a command and either gets a link or a
  refusal they can read.

### Capabilities

- `identity/operator-registration`

### Satisfied by

- `identity/operator-registration`: Requirement: Registering a household member
- `identity/operator-registration`: Requirement: Refusing a registration that
  names an already-registered email
- `identity/operator-registration`: Requirement: Issuing a further enrolment link
  to a registered member
- `identity/operator-registration`: Requirement: Refusing an unauthenticated
  operator

## S2 · Enrol a passkey from the link and land inside

**As a** newly registered household member who has never used wimm and has been
sent a link by the operator
**I want** to open the link, confirm with my fingerprint or face, and find myself
inside the app
**so that** I have an account I can get back into, without choosing a password or
being asked to prove an email address the operator already vouched for.

### INVEST

- **Independent** — No. It needs a link, and S1 is what produces one. Accepted
  for the same reason as S1.
- **Negotiable** — Yes. The value is "I am in and I can come back"; requiring a
  discoverable credential is how that is met, not what the member asked for.
- **Valuable** — Yes. This is the moment a person exists in the product.
- **Estimable** — Yes. One page, one WebAuthn ceremony, one session.
- **Small** — Borderline. It carries the enrolment page, the ceremony, the
  session, and every way a link can fail. Accepted: a link that works only on the
  happy path is not shippable, and the failure states share the page.
- **Testable** — Yes. A person opens the link on a real device and either ends up
  signed in or reads why not.

### Capabilities

- `identity/passkey-enrolment`

### Satisfied by

- `identity/passkey-enrolment`: Requirement: Enrolling a passkey from a valid
  enrolment link
- `identity/passkey-enrolment`: Requirement: Requiring a passkey the member can
  sign in with later
- `identity/passkey-enrolment`: Requirement: Refusing an enrolment link that
  cannot be used

## S3 · Get back in without asking anyone

**As a** household member who enrolled a passkey on an earlier visit and has
since closed the browser, rebooted, or been signed out
**I want** to sign in with the passkey I already have, without typing my email
or anything else
**so that** coming back is my own business rather than a message to the operator
asking for another link.

### INVEST

- **Independent** — No. There is nothing to sign in with until S2 has run.
  Accepted: the dependency is ordering, not design — once a credential exists,
  sign-in stands entirely on its own.
- **Negotiable** — Yes. "Without typing anything" states the need; that it falls
  out of discoverable credentials is the mechanism.
- **Valuable** — Yes, and it is the story that makes S2's promise true. Without
  it the first expired session costs an operator round-trip, and every household
  member is one cookie expiry away from being locked out.
- **Estimable** — Yes. One page, one ceremony, one session.
- **Small** — Yes.
- **Testable** — Yes, and sign-out is what makes it so: without it, proving this
  story means waiting for a session to expire.

### Capabilities

- `identity/passkey-sign-in`

### Satisfied by

- `identity/passkey-sign-in`: Requirement: Signing in with an enrolled passkey
- `identity/passkey-sign-in`: Requirement: Sending an unauthenticated visitor to
  sign in
- `identity/passkey-sign-in`: Requirement: Signing out
- `identity/passkey-sign-in`: Requirement: Refusing a passkey that is not enrolled
