---
name: data-migrations
description: How a Postgres schema changes safely here - goose SQL migrations in the owning service, expand/contract across releases, batched idempotent backfills, lock-aware DDL, forward-only recovery. Use whenever someone creates, edits, or reviews a migration; adds, renames, drops, or retypes a column, table, index, or constraint; needs a backfill; or asks if a schema change is safe. Even "add a column".
paths:
  - "**/migrations/**"
---

# Data migrations

Schema changes ship expand → migrate code → backfill → contract, each step its own release. Production never runs a down migration.

## Rules

1. Migrations live in `apps/<service>/migrations/`, embedded with `embed.FS`, applied by `just migrate` locally and an explicit deploy step in production. Never at startup. One database, one owning service.
2. `goose create <name> sql` gives timestamp versions; `goose fix` renumbers before merge. Out-of-order stays off.
3. Every migration has a `-- +goose Down`, tested with `just migrate-redo`. Production recovery is a new forward migration.
4. Top of every migration: `SET lock_timeout = '5s';` and `SET statement_timeout = '30s';`. `CREATE INDEX CONCURRENTLY` in a `-- +goose NO TRANSACTION` file; constraints `NOT VALID` then `VALIDATE`; no `UPDATE` inside DDL.
5. A migration changes structure. A backfill changes rows, batched and idempotent, as a job.
6. Additive changes (nullable column, table, `CONCURRENTLY` index, `NOT VALID` constraint) ship alone. Everything else follows `references/expand-contract.md`; the PR names its step (`expand 1/6`).

## References

| Read | When |
|---|---|
| `references/expand-contract.md` | Any non-additive change |
| `references/postgres-safety.md` | The safe form of each DDL operation |
| `references/goose.md` | File layout, annotations, embedding, `just` verbs |
| `references/backfills.md` | Batching and idempotency for row changes |
| `references/review-checklist.md` | Reviewing a migration PR; before production |
