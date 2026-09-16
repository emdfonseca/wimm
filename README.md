# wimm

A household ledger you run yourself. One instance serves one household; the
person who runs it is the **operator**, and everyone who uses it is a **member**.

There is no signup page. Members do not create accounts: the operator registers
each person and hands them a link, and the link is how they create a passkey.
That is deliberate — in a household product the operator already knows who lives
there, so vouching is a better answer than an email round trip.

## Run it

Everything is one command. It starts Postgres, the service, the web app and
Storybook together:

```bash
devbox run -- just up
```

```text
http://localhost:9466    the app
http://localhost:9469    Storybook, the design system in isolation
127.0.0.1:9467           wimmd, the API the web app talks to
127.0.0.1:9468           wimmd's operator listener, loopback only
```

Postgres listens on a Unix socket under `.devbox/`, not a port, so it cannot
collide with a system-wide install or another checkout of this repo.

```bash
devbox run -- just down     # stop everything
devbox run -- just ps       # what is running
devbox run -- just db-psql  # a shell on the database
```

First run is slow: devbox fetches the toolchain and `initdb` creates the
cluster. After that it is a few seconds.

## Bring a member in

Two steps, both yours. Registering creates the person; enrolling creates their
passkey.

```bash
cd apps/wimm
go run ./cmd/wimmctl register ada@example.com Ada Lovelace
```

That prints a link and the moment it stops working:

```text
Registered Ada Lovelace <ada@example.com>.

  http://localhost:9466/enrol/Lrdk1cOpFfb_vsw-rW92elot5BBoJ0V3krZ7JZQgbGQ

It stops working at 12:00 on 17 September 2026.
Send it over a channel you trust: whoever opens it becomes this member.
```

Send it however you already talk to them. **Whoever opens that link becomes that
member**, so the channel is the security boundary — it is a bearer credential,
not an invitation that later checks who accepted it.

They open it, confirm with their fingerprint, face or device PIN, and they are
in. No password is chosen and nothing is typed.

## When a member needs a new link

Same command, for someone already registered. Use it for a new device, or for
one they have lost:

```bash
go run ./cmd/wimmctl link ada@example.com
```

Issuing a new link **stops any earlier link working** and **leaves their existing
passkeys working**. So a lost phone is: issue a link, they enrol on the new
device. Their old passkey keeps working until there is a way to remove it —
that is the next change, not this one.

## What the operator can and cannot do

```text
can     register a person, issue them a link
cannot  see or recover a link after it is printed — only its hash is stored
cannot  remove a passkey or revoke a session yet
cannot  change a member's name or address yet
```

## Things that will bite you

**Links are single use and last a day.** Single use means one passkey, not one
page load: if their device refuses to save a passkey, or they close the prompt,
the same link still works so they can try elsewhere. It closes the moment a
passkey is actually saved.

**A member needs a device that saves passkeys.** wimm requires a discoverable
credential, which is what lets them sign in later without typing anything. An old
security key with no room to store one is refused, at enrolment, with the link
still usable.

**The relying-party identifier is a one-way door.** `WIMM_RP_ID` defaults to
`localhost`. Changing it after anyone has enrolled invalidates every passkey ever
created, with no migration. Pick the broadest domain the instance will ever use
before the first member enrols against a deployment.

**A phone can be the authenticator, but not the browser.** Over the cross-device
(QR) flow the ceremony runs in your laptop's browser and the phone just confirms,
which works on `localhost`. A phone browsing _to_ this instance needs a hostname
and a certificate.

## Configuration

`wimmd` reads its configuration from the environment and refuses to start if the
operator listener is reachable with no credential set. Inside `devbox run` these
all have working defaults.

```text
WIMM_DATABASE_URL           Postgres connection string
WIMM_RP_ID                  WebAuthn relying party, default localhost
WIMM_ORIGINS                origins a ceremony may come from, comma separated
WIMM_BASE_URL               the origin enrolment links are built against
WIMM_PUBLIC_ADDR            the listener the web app calls
WIMM_OPERATOR_ADDR          the listener wimmctl calls, loopback by default
WIMM_OPERATOR_ENABLED       false turns the operator surface off entirely
WIMM_OPERATOR_CREDENTIAL    what wimmctl presents; minted per checkout into
                            .devbox on first use, never committed
WIMM_ENROLMENT_LINK_LIFETIME  default 24h
WIMM_SESSION_LIFETIME         default 14 days
```

## Working on it

```bash
devbox run -- just ci             # everything CI runs, and nothing else
devbox run -- just check apps/web # one package
devbox run -- just gen            # regenerate contracts and tokens
```

`just check <dir>` is the only test, lint and typecheck entry point. If it fails,
the work is not done.

Tests use their own database, `wimm_test`, on the same cluster. The suite
truncates every table it finds, so pointing it at the development database would
delete whatever you were working with.

Decisions live in `docs/decisions/`. Read them before proposing something they
already settle — particularly ADR 0016, which covers why enrolment works this
way.
