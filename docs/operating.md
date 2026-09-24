# Operating an instance

What the operator does, what a member does, and the environment `wimmd`
reads. The quickstart is in the root README.

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

## Connect a bank

A member connects their own bank and then decides who in the household sees
each account it exposes.

```text
1. Choose a bank            /connect
2. See what will be shared  /connect/<bank> — states the real date access ends
3. Confirm at the bank      off wimm entirely; wimm never sees the password
4. Choose who sees what     /connect/<bank>/accounts
5. Overview                 the accounts, their balances, and a total
```

**Every account the bank exposes is stored, and starts as yours.** The details
come back once and no endpoint lists them again, so wimm keeps them all. You
then release the ones that are not yours and give each other member a level:

```text
nothing   they do not know the account exists — the default for everyone
balance   the bank, the account's name, and the balance with its read time
details   the above, plus the number's last digits, the type, the holder name
```

An account with no owner and no level given to anyone is **never read**: wimm
knows it exists and does not know what is in it. Any owner can change owners
and levels later; whoever connected the bank gets no standing power over it.

Two people in one household see different screens and different totals, and
neither is told what the other sees.

## When access runs out

Open banking has no renewal. Consent lasts until a date **the bank** sets, and
restoring is the whole flow again — which is why Overview offers it rather than
just reporting a failure.

```text
Revolut, Montepio    90 days
ActivoBank            1 day
```

One day is not a variation on ninety: with ActivoBank connected, reconnecting
is a daily routine and "has stopped updating" is that connection's normal
resting state rather than an incident. Restoring keeps every owner and level,
matched on an identifier that survives a new authorisation. An account the bank
newly offers arrives as yours with nobody else granted; one it no longer offers
goes, and you are told which.

A bank that simply does not answer is different: the balances already on screen
stay exactly where they are, with their original read times, and the page says
which bank could not be reached.

## What the operator can and cannot do

```text
can     register a person, issue them a link
cannot  see or recover a link after it is printed — only its hash is stored
cannot  remove a passkey or revoke a session yet
cannot  change a member's name or address yet
cannot  connect a bank on anyone's behalf, or see any balance
cannot  change who owns an account or who sees it
```

Banking has no operator surface at all. A member connects their own bank by
confirming at it, and only an owner of an account changes who sees it — there
is nothing here for an operator to do, and so no way for them to do it.

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

**Banking is optional, and off by default.** With `WIMM_BANKING_GATEWAY` unset,
`wimmd` starts normally and Overview shows its empty state. Set it and every
credential it needs must be present and every key file must be readable by its
owner alone, or the process refuses to start rather than running half
configured.

**The free tier of the bank gateway covers one person's accounts.** Enable
Banking's terms reach only accounts the Control Panel user links as themselves,
so one member connects the banks they can authenticate at — joint accounts
included — and wimm's owners and levels decide what the rest of the household
sees of them. Lifting that needs a contract and a company.

**Connecting a bank needs https, even on localhost.** Enable Banking refuses a
plain-http redirect URI. `just dev-cert` issues a locally-trusted certificate
with mkcert and the dev server picks it up; without one the app still runs over
http and everything except the bank hand-off works.

Serving https changes the origin, which WebAuthn validates. `WIMM_BASE_URL` and
`WIMM_ORIGINS` are derived from whether that certificate exists rather than set
by hand, so they cannot disagree with what the dev server is actually doing —
they did, once, and every passkey ceremony would have been rejected for it.

Passkeys already enrolled survive the switch: the relying-party identifier
stays `localhost`, and that — not the scheme — is the value that cannot be
changed after anyone enrols.

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

Banking, all unset by default. Naming a gateway requires the rest of them.

Put them in `local.env`, which the devbox shell sources before anything else
and which is gitignored:

```bash
cp local.env.example local.env   # then fill it in
```

It is sourced rather than declared with devbox's `env_from`, because `env_from`
refuses to start when the file is absent and a clean checkout has none.

```text
WIMM_BANKING_GATEWAY              which adapter; unset disables connecting
WIMM_BANKING_ENCRYPTION_KEY       path to the file holding the sealing keys
WIMM_ENABLEBANKING_APPLICATION_ID the registered application
WIMM_ENABLEBANKING_PRIVATE_KEY    path to the request signing key
WIMM_ENABLEBANKING_REDIRECT_URL   must match one registered with the gateway
WIMM_BALANCE_STALE_AFTER          default 24h
```

**Both key variables are paths, never the material.** A key in an environment
variable reaches every child process and every crash report. Both files must be
readable by their owner alone — mode `600` — and `wimmd` refuses to start
otherwise.

The sealing key file is lines of `<key-id> <base64 of 32 bytes>`. The first line
seals new values; the rest only open old ones, so rotating is a line added at
the top and the old key kept until nothing references it:

```text
# newest first
2026-01 aGVyZSBiZSAzMiBieXRlcyBvZiBrZXkgbWF0ZXJpYWw=
2025-07 c29tZSBvdGhlciAzMiBieXRlcyBvZiBrZXkgbWF0ZXJpYWw=
```
