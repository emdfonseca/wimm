---
name: data-migrations
description: The procedure for changing a Postgres schema safely in this repo - goose SQL migrations embedded in the owning service, expand/contract across separate releases, batched idempotent backfills, lock-aware DDL, and forward-only recovery in production. Use this whenever someone creates, edits, or reviews a migration file; adds, renames, drops, or retypes a column, table, index, or constraint; needs to backfill data; asks whether a schema change is safe to deploy, how to roll it back, or in what order to ship it; or sets up goose in a new service. Reach for it even for "just add a column" - NOT NULL, defaults, and indexes are exactly where a one-line change locks a table.
paths:
  - "**/migrations/**"
  - "**/*.sql"
---

# Data migrations

**Schema changes ship in the order expand → migrate code → backfill → contract, each step its own release, and nothing in production ever runs a down migration.** The database is the one component that cannot be redeployed from git, so every change to it is designed to be survivable at each intermediate state.

The failure this prevents is not the dramatic one. It is `ALTER TABLE ... ADD COLUMN ... NOT NULL` on a large table at 2pm, the lock queue behind it, and the "quick rollback" that drops a column customers already wrote to.

## Start here, every time

1. **Is it additive?** Adding a nullable column, a table, an index `CONCURRENTLY`, or a `NOT VALID` constraint: safe alone. Anything else — rename, retype, drop, `NOT NULL`, default on a huge table — is a multi-release sequence. Read `references/expand-contract.md` before writing it.
2. **What does the running code see mid-way?** The old release and the new release both run during a deploy. Both must work against the schema at every step. If that is not true, split the step.
3. **Is there data to move?** Then there is a backfill, and it is a separate, batched, resumable job — not a `UPDATE` inside the migration.

## Where to read next

| Read this | When |
|---|---|
| `references/expand-contract.md` | The full procedure for any non-additive change, with the release-by-release sequence and what code ships when. |
| `references/postgres-safety.md` | Which DDL locks what, and the safe form of each operation — `CONCURRENTLY`, `NOT VALID`, `lock_timeout`, defaults. |
| `references/goose.md` | File layout, naming, annotations, embedding, applying in deploy; `goose fix` before merge. |
| `references/backfills.md` | Writing a backfill that can be stopped, resumed, and rerun. |
| `references/recovery.md` | What "rollback" actually means for the database, and how to prepare for it before deploying. |
| `references/review-checklist.md` | Reviewing a migration PR. |

## Rules that are constantly needed

### Migrations live with the service that owns the schema

```text
apps/billing/migrations/
├── 20260913120000_add_invoice_currency.sql
└── ...
```

Embedded via `embed.FS`, applied by `just migrate` locally and by an explicit migrate step in the deploy pipeline — **never** by the service at startup in production. A service that migrates on boot turns every scale-out and every crash-loop into a schema change attempt, and it makes the migration's failure indistinguishable from the service's.

One database, one owning service. Two services sharing tables is a service boundary problem the migration tooling cannot fix.

### Timestamps in development, sequential before merge

`goose create <name> sql` produces timestamp-versioned files, which is what avoids collisions across branches. Before a PR merges, `goose fix` renumbers to sequential so the applied order is unambiguous and out-of-order application stays *disabled*. A migration that applied out of order is one whose tested sequence was not the production sequence.

### Every migration has a Down, and Down is never the production plan

The `-- +goose Down` block is required, tested locally (`just migrate-redo`), and exists so developers can iterate. In production, recovery is forward: a new migration that undoes what needs undoing, deployed like any other. A down migration on a table that has received writes is data loss with a friendly name. `references/recovery.md`.

### Locks are the actual risk

Postgres takes `ACCESS EXCLUSIVE` for most `ALTER TABLE`, and that lock queues behind any running query and blocks everything after it. So: `SET lock_timeout = '5s'` at the top of every migration, `CREATE INDEX CONCURRENTLY` (with `-- +goose NO TRANSACTION`), constraints added `NOT VALID` then `VALIDATE` separately, and no `UPDATE` of existing rows inside DDL. The table in `references/postgres-safety.md` says which form is safe for each operation.

### Schema changes and data changes are different artifacts

A migration changes structure. A backfill changes rows, in batches, idempotently, with progress recorded, run as a job. Mixing them makes the migration unbounded in time and unrepeatable on failure.

## Checkpoints

- **Before writing** a non-additive change — the expand/contract sequence is written down as a list of releases, each with the migration and the code change it carries. If you cannot write that list, the change is not understood yet.
- **Before merging** — `goose fix` applied, Down tested, `lock_timeout` set, the review checklist walked, and the PR states which release step this is (`expand 1/3`).
- **Before deploying** — the migration was applied to a staging database with production-like size, and its duration is known. "It was instant locally" is what everyone says about the one that locks production.
- **Before contracting** — the code that stops reading the old shape has been in production long enough that a rollback of *that* code is no longer plausible. Only then is the drop safe.
