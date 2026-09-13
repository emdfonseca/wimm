# 0001 · Stack and repo shape

## Status

Accepted

## Context

wimm is a single self-contained monorepo. It needs one toolchain, one task entry
point, and one place for each kind of code, decided before the first service
exists so nothing is retrofitted later.

## Decision

- Go for services and workers under `apps/`, importable code under `packages/`.
- SvelteKit with Svelte 5 for the UI, in `packages/ui`. Python only where a
  library forces it.
- Connect over protobuf between our own callers; REST with OpenAPI 3.1 only for
  third parties.
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
  never hand-copied.
- The toolchain is reproducible from `devbox.json` alone; nothing is installed
  globally.
