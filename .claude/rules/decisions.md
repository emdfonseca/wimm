# Decisions

Generated from `docs/decisions/` by `just adr-index`. Never edit; write an ADR.

## 0001 · Stack and repo shape — Accepted

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
