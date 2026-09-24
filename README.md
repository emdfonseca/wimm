# wimm

A household ledger you run yourself. One instance serves one household; the
person who runs it is the **operator**, and everyone who uses it is a **member**.

There is no signup page. Members do not create accounts: the operator registers
each person and hands them a link, and the link is how they create a passkey.
That is deliberate — in a household product the operator already knows who lives
there, so vouching is a better answer than an email round trip.

Members connect their own banks through an open-banking gateway and decide,
per account, who else in the household sees the balance and the details.
Overview then shows each person the money they are allowed to see, and a
typical month drawn from the last six.

## Why I built it

Every money app I tried assumes one person per login, so a household ends up
sharing one password or keeping two sets of books. I wanted the opposite: one
instance for the house, one passkey per person, and each account's visibility
decided by whoever owns it rather than by whoever set the software up. It is
also where I try out a way of working I believe in — the code is the design of
record, every screen state is a story, and a screen counts as done when a
person has looked at it and approved that exact render.

## Stack

```text
apps/wimm            Go 1.27: wimmd (the service) and wimmctl (the operator CLI)
                     Postgres 18 via pgx, goose migrations, go-webauthn, Connect RPC
apps/web             SvelteKit 2 / Svelte 5; server routes hold the session and speak Connect
apps/storybook       Storybook 10, Vitest browser mode on Chromium, the design canvas
packages/ui          the design system: tokens, components, the pure screens
packages/contracts   protobuf schemas; buf generates Go and TypeScript into gen/
```

devbox pins the toolchain, `just` owns every task, process-compose runs the
stack. CI is `devbox run -- just ci` and nothing else
(`.github/workflows/ci.yml`).

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

Registering creates the person; enrolling creates their passkey. Both are the
operator's, from `apps/wimm`:

```bash
go run ./cmd/wimmctl register ada@example.com Ada Lovelace   # prints an enrolment link
go run ./cmd/wimmctl link ada@example.com                    # a fresh link, for a new or lost device
```

Whoever opens the link becomes that member, so the channel you send it over is
the security boundary. Links are single use and last a day. The rest — what the
operator can and cannot do, bank consent expiry, the relying-party identifier
being a one-way door, every `WIMM_*` variable — is in
[docs/operating.md](docs/operating.md).

## Connect a bank

Optional and off by default; `wimmd` runs without a gateway configured. With
one, a member picks a bank at `/connect`, confirms at the bank (wimm never sees
the password), and then chooses who in the household sees each account and at
which level: `nothing`, `balance` or `details`. An account nobody is allowed to
see is never read. Two members see different screens and different totals, and
neither is told what the other sees.

## Working on it

```bash
devbox run -- just ci             # everything CI runs, and nothing else
devbox run -- just check apps/web # one package: lint, typecheck, test
devbox run -- just gen            # regenerate contracts, tokens and the decision index
```

`just check <dir>` is the only test, lint and typecheck entry point. Go tests
run with `-race` against their own database, `wimm_test`, on the same local
cluster; the suite truncates every table it finds. Stories are the UI tests:
their play functions run in a real Chromium under Vitest.

The design canvas, page versions and approvals — including the pictures kept
under `apps/storybook/canvas/approved/` — are described in
[apps/storybook/README.md](apps/storybook/README.md).

## Architecture

Three API surfaces with one job each (ADR 0001): the web app's server routes
call `wimmd` over Connect; the browser speaks only to SvelteKit; a third-party
REST surface is reserved and not built. Identity is passkeys with
operator-issued enrolment links (ADR 0016). Banks are read through a gateway
adapter, currently Enable Banking (ADR 0018); accounts always have an owner and
per-member levels (ADRs 0019, 0022); transactions are stored in a ledger the
consent fills (ADR 0021).

```text
docs/decisions/      29 ADRs, the reasons behind anything that looks odd
docs/design/         tokens, surfaces, the project setup record
openspec/            specs and change proposals (ADR 0007)
.claude/skills/      engineering standards: API contract, migrations, Storybook, TDD
```

Start with ADR 0016 (enrolment), 0023 (code is the design of record) and
0027–0029 (versioned, approved page stories).

## Status

Pre-release, unreleased. It runs locally under devbox; there is no deployment
story yet. The operator cannot remove a passkey or revoke a session, and the
bank gateway's free tier reaches one person's accounts only.

## License

MIT, see [LICENSE](LICENSE).
