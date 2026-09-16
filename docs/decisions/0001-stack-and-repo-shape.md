# 0001 · Stack and repo shape

## Status

Accepted

## Context

wimm is a single self-contained monorepo. It needs one toolchain, one task entry
point, and one place for each kind of code, decided before the first service
exists so nothing is retrofitted later.

## Decision

- Go for services and workers under `apps/`, importable code under `packages/`.
- SvelteKit with Svelte 5 for the UI. The app is deployable, so it lives under
  `apps/` like any other service; `packages/ui` is the design system it imports —
  the token contract and the Svelte components screens instance. Python only
  where a library forces it.
- **Three API surfaces, and each has exactly one job.**

  ```text
  our own services, server to server   Connect over protobuf
  our own web client, browser to app   REST over HTTP and JSON, in SvelteKit
  third parties and webhooks           REST with OpenAPI 3.1, RFC 9457 errors
  ```

  **The browser speaks REST, never Connect.** SvelteKit server routes are the
  web client's API: load functions and form actions for anything the framework
  already models, and JSON endpoints under `/api` for what it does not, such as
  handing an authenticator's response back. They hold the session cookie and
  call Connect from the server.

  Shipping a protobuf runtime to the browser costs 15-25 kB gzipped and buys
  little: the payloads that are actually error-prone, such as a WebAuthn
  credential, are JSON shapes the browser defines, so the schema would describe
  an envelope around an opaque string. It also cannot express a redirect, which
  any link-exchange flow needs. Server to server, the generated contract keeps
  its whole value: a Go CLI and a Go service cannot skew.

  **The web client's REST is not documented and not versioned**, which is what
  separates it from the third-party surface. Its only consumer ships in the same
  deploy, so a change is one commit rather than a negotiation, and `PageData`
  flowing from a load function into its page is stronger typing than a schema
  would give. The moment a second consumer appears it stops being this and
  becomes the third row, OpenAPI and all.
- Postgres with goose migrations owned by the service that owns the database.
- OpenTelemetry over OTLP for traces, metrics, and logs. The backend stays
  undecided; services never know it.
- devbox owns the toolchain, `just` owns every task entry point, and CI runs
  `devbox run -- just ci` and nothing else.
- Repo standards live as Claude skills in `.claude/skills/`; they reference only
  paths inside this repo.

## Consequences

- Adding a language or a task runner is a new ADR, not a local choice.
- Every package carries a `justfile` with the standard verbs, so `just check`
  works the same from a terminal, from CI, and from the Stop hook.
- Cross-language types come from generated contracts in `packages/contracts`,
  never hand-copied. The generated TypeScript client is imported by a SvelteKit
  server route, never by a component, so it stays out of the browser bundle.
- The toolchain is reproducible from `devbox.json` alone; nothing is installed
  globally.
