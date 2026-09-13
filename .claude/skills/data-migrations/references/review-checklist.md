# Reviewing a migration

## Safety

- [ ] `SET lock_timeout` and `statement_timeout` at the top.
- [ ] Each DDL statement uses the safe form from `references/postgres-safety.md` (`CONCURRENTLY`, `NOT VALID`, no volatile defaults on large tables, no in-migration `UPDATE`).
- [ ] `-- +goose NO TRANSACTION` present exactly when a statement requires it, and the file then contains only statements that tolerate partial application.
- [ ] Row count of the affected table and expected duration are stated in the PR.

## Sequence

- [ ] Additive-only, or the PR names its expand/contract step and links the plan.
- [ ] The currently deployed code works against the schema after this migration (expand/contract guarantee).
- [ ] No drop, rename, or retype unless this is the final contract step and the stop-old release has been in production.
- [ ] Any data movement is a backfill job, not in the migration.

## Hygiene

- [ ] Versioned sequentially (`goose fix` run); one logical change per file; description names the change.
- [ ] Down block present and `migrate-redo` passes.
- [ ] Lives under the owning service; no cross-service table access introduced.
- [ ] Application code in the same PR still works if the migration is applied first — and if it is applied last. Deploy order is not guaranteed unless the pipeline enforces it; say which it is.
