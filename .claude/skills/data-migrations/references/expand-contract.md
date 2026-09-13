# Expand / contract

## Example: rename `invoices.total` → `invoices.amount_cents`

1. **Expand** — `ALTER TABLE invoices ADD COLUMN amount_cents bigint;` (nullable). Code unchanged.
2. **Dual-write** — code writes both `total` and `amount_cents`; reads `total`. No migration.
3. **Backfill** — job: `UPDATE invoices SET amount_cents = total * 100 WHERE amount_cents IS NULL AND id BETWEEN $1 AND $2`, batched by ID range. Verify `COUNT(*) WHERE amount_cents IS NULL` is 0.
4. **Read new** — code reads `amount_cents`; writes both. Migration: `ALTER TABLE invoices ADD CONSTRAINT amount_cents_not_null CHECK (amount_cents IS NOT NULL) NOT VALID;` then, separately, `VALIDATE CONSTRAINT`, then `ALTER COLUMN SET NOT NULL` (metadata-only once the check is validated).
5. **Stop old** — code stops writing `total`. No migration.
6. **Contract** — `ALTER TABLE invoices DROP COLUMN total;` only once release 5 has been stable long enough that rolling it back is off the table.

Steps 2 and 3 may ship together if the backfill is triggered after deploy. Each transition must be safe on its own; a rollback may land on any of them.

## Which changes need this

| Change | Additive? | Sequence |
|---|---|---|
| Add nullable column | yes | one migration |
| Add table | yes | one migration |
| Add index | yes, if `CONCURRENTLY` | one migration, `NO TRANSACTION` |
| Add column with `NOT NULL` | no | expand nullable → backfill → constrain |
| Add column with default | constant default: yes (metadata-only) | volatile default → backfill path |
| Rename column / table | no | full sequence |
| Change column type | no | full sequence via a new column |
| Drop column / table | no | stop use → wait → drop; drop is the last step |
| Add foreign key | no | `NOT VALID` → `VALIDATE` |
| Add unique constraint | no | `CREATE UNIQUE INDEX CONCURRENTLY` → `ADD CONSTRAINT ... USING INDEX` |

Each PR in the sequence names its step and links the plan: `expand (1/6) — see #123`.
