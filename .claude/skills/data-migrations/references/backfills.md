# Backfills

A backfill is a job, not a migration.

- **Batched** by primary-key range (or a stable cursor), each batch its own transaction, thousands of rows, not millions.
- **Idempotent**: the `WHERE` clause selects only rows that still need the change (`WHERE amount_cents IS NULL`), so rerunning from the start is safe.
- **Resumable**: last processed key recorded in a `backfills` table (`name, cursor, updated_at`).
- **Throttled**: configurable sleep between batches.
- **Observable**: rows processed and remaining as metrics; a span per batch.
- **Verifiable**: a query returning the count of unmigrated rows; done means zero.

Lives in `apps/<service>/internal/backfill/` and runs as a subcommand (`<service> backfill <name>`), triggered by the deploy pipeline or by hand, never at service start.
