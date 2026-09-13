# Postgres safety

The point of this file is that the *safe form* of most operations exists and is only slightly more typing than the dangerous one.

## Always

```sql
-- top of every migration
SET lock_timeout = '5s';
SET statement_timeout = '30s';   -- raise only with a comment saying why
```

`lock_timeout` makes a migration that would queue behind a long transaction fail fast instead of blocking every query behind it. Failing fast is recoverable; a lock queue on the invoices table is an outage.

## Operation → safe form

| Operation | Dangerous form | Safe form |
|---|---|---|
| Add index | `CREATE INDEX` (blocks writes for the duration) | `CREATE INDEX CONCURRENTLY` in a `-- +goose NO TRANSACTION` migration; check for `INVALID` index on failure and drop/retry |
| Add `NOT NULL` | `ALTER COLUMN SET NOT NULL` (full scan under lock) | Add `CHECK (col IS NOT NULL) NOT VALID`; `VALIDATE CONSTRAINT` (scan without exclusive lock); then `SET NOT NULL` (PG12+ uses the validated check, metadata-only) |
| Add foreign key | `ADD CONSTRAINT ... FOREIGN KEY` (locks both tables, scans) | `ADD CONSTRAINT ... NOT VALID`; later `VALIDATE CONSTRAINT` |
| Add unique constraint | `ADD CONSTRAINT ... UNIQUE` (builds index under lock) | `CREATE UNIQUE INDEX CONCURRENTLY`; `ADD CONSTRAINT ... UNIQUE USING INDEX` |
| Add column with default | volatile default (`now()`, `gen_random_uuid()`) rewrites the table | Constant default is metadata-only on PG11+; volatile → add nullable, backfill, then set default for new rows |
| Change column type | `ALTER COLUMN TYPE` (rewrite under lock) | New column + expand/contract; exceptions: widening `varchar(n)` and some numeric widenings are metadata-only |
| Rename | `RENAME` is fast but breaks running code | Expand/contract, or rename plus a temporary view with the old name during the window |
| Drop column | Fast, but irreversible and breaks old code | Only as the final contract step; consider a `deprecated_` rename one release earlier so a stray reader fails loudly |
| Backfill rows | `UPDATE table SET ...` (one transaction, one lock, one undo log) | Batched by primary-key range in a job; see `references/backfills.md` |
| Drop index | Takes a lock briefly | `DROP INDEX CONCURRENTLY` in `NO TRANSACTION` |

## Know the size

Before merging a migration on a table, know its row count and the time the operation took on a staging copy of production size. Record it in the PR. The number does not need to be precise; it needs to exist, so nobody discovers it in production.

## Extensions and enums

`CREATE EXTENSION` needs superuser or a pre-provisioned extension; do it in infra provisioning, not a migration. `ALTER TYPE ... ADD VALUE` cannot run inside a transaction on older versions and cannot be rolled back — prefer a lookup table or a text column with a check constraint when the value set changes.
