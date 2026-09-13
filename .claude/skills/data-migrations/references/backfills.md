# Backfills

A backfill changes existing rows to match a new shape. It is a job, not a migration, because it has properties a migration cannot have: it runs for a long time, it can be paused, it can fail halfway and resume, and it must be safe to run twice.

## Properties

- **Batched** by primary-key range (or a stable cursor), each batch its own transaction, small enough that the lock and the undo are trivial — thousands of rows, not millions.
- **Idempotent**: the `WHERE` clause selects only rows that still need the change (`WHERE amount_cents IS NULL`), so rerunning from the start is safe.
- **Resumable**: progress (last processed key) recorded somewhere durable — a `backfills` table with `name, cursor, updated_at` — so a restart continues instead of rescanning.
- **Throttled**: a configurable sleep between batches, so it can be slowed under load rather than killed.
- **Observable**: rows processed, rows remaining, batch duration as metrics; one log line per N batches; a span per batch.
- **Verifiable**: a query that returns the count of unmigrated rows, run before, during, and after; done means zero.

## Shape

```go
for {
	n, last, err := backfillBatch(ctx, db, cursor, batchSize)   // UPDATE ... WHERE id > $cursor AND amount_cents IS NULL ORDER BY id LIMIT $n
	if err != nil { return err }                                   // resume from stored cursor next run
	if n == 0 { break }
	saveCursor(ctx, db, "invoices_amount_cents", last)
	metrics.rowsProcessed.Add(ctx, int64(n))
	time.Sleep(throttle)
}
```

Lives in `apps/<service>/internal/backfill/` and runs as a subcommand (`billing backfill invoices-amount-cents`), triggered by the deploy pipeline or by hand, never at service start.

## Dual-write first

A backfill only produces a consistent result if new writes are already landing in the new shape. Otherwise rows written during the backfill are missed. That is why dual-write ships before the backfill in the expand/contract sequence, and why the verification query is run *after* the backfill reports done.

## Large tables

Above tens of millions of rows, expect hours. That is fine; the job is designed for it. What is not fine is a backfill that holds a transaction across the whole run, or one that nobody can tell the progress of. If it will take days, say so in the PR and plan the release cadence around it.
