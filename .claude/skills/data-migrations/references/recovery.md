# Recovery

"Rollback" for application code means redeploying the previous version. For the database it means something different, and pretending otherwise is how data is lost.

## Forward only in production

A migration that has been applied to production and that any code has written through is not rolled back with `goose down`. It is corrected with a *new* migration, reviewed and deployed like any other. Reasons:

- Down blocks are tested against empty local databases, not against production data volume or shape.
- Dropping the column the new code wrote to destroys the writes; there is no undo for that.
- A forward migration is visible in history; a down leaves the version table lying about what happened.

The Down block's job is developer iteration and the `migrate-redo` test. Its existence is required; its use in production is not a plan.

## Design for the rollback that will happen

Because expand/contract keeps both code versions working at every step, rolling back the *application* is always safe — the previous release still finds the schema it expects. That is the recovery path: revert the deploy, leave the schema, fix forward. It works only if the sequence was followed; the moment a release both changes the schema and requires it, application rollback is off the table.

## Before every production migration

- [ ] Applied to a staging database of production size; duration recorded.
- [ ] `lock_timeout` set; the failure mode on timeout is "migration fails, nothing changed".
- [ ] The previous application release runs correctly against the post-migration schema (this is the expand/contract guarantee — confirm it, do not assume it).
- [ ] The forward-fix migration for the most likely failure is sketched, even if only as a sentence in the PR.
- [ ] A recent backup or point-in-time recovery window exists and someone has checked that restore actually works this quarter.

## When it goes wrong anyway

1. Stop the bleeding: if the migration is still running and blocking, cancel it (`pg_cancel_backend`); `lock_timeout` should have already done this.
2. Establish state: `goose status`; what applied, what did not; is the transaction rolled back or was it `NO TRANSACTION` and partially done (an `INVALID` index, for instance)?
3. Restore service: roll back the application if it is the new code that is failing; the schema stays.
4. Fix forward: write the correcting migration, test it on a copy of the current production schema, deploy.
5. Write it down: what happened, in the PR or the incident record, so the checklist above grows by one line if it needs to.
