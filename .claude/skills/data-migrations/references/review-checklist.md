# Reviewing a migration

## Safety

- [ ] `SET lock_timeout = '5s';` and `SET statement_timeout = '30s';` at the top (a longer `statement_timeout` needs a comment saying why).
- [ ] Each DDL statement uses the safe form from `references/postgres-safety.md` (`CONCURRENTLY`, `NOT VALID`, no volatile defaults on large tables, no in-migration `UPDATE`).
- [ ] `-- +goose NO TRANSACTION` present exactly when a statement requires it, and the file then contains only statements that tolerate partial application.
- [ ] Row count of the affected table and expected duration are stated in the PR.

## Sequence

- [ ] Additive-only, or the PR names its expand/contract step and links the plan.
- [ ] The currently deployed code works against the schema after this migration.
- [ ] No drop, rename, or retype unless this is the final contract step and the stop-old release has been in production.
- [ ] Any data movement is a backfill job, not in the migration.

## Hygiene

- [ ] Versioned sequentially (`goose fix` run); one logical change per file; description names the change.
- [ ] Down block present and `migrate-redo` passes.
- [ ] Lives under the owning service; no cross-service table access introduced.
- [ ] Application code in the same PR works whether the migration is applied first or last.

## Before production

Recovery is forward only: a migration that production code has written through is corrected by a new migration, never by `goose down`.

- [ ] Applied to a staging database of production size; duration recorded.
- [ ] `lock_timeout` set, so the failure mode on contention is "migration fails, nothing changed".
- [ ] The previous application release runs correctly against the post-migration schema.
- [ ] The forward-fix migration for the most likely failure is sketched in the PR.
- [ ] A recent backup or point-in-time recovery window exists.
