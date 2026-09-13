# Expand / contract

Any change that is not purely additive ships as a sequence of releases, each of which leaves the system working for both the previous and the next code version. The pattern is the same whether it is a rename, a type change, a move between tables, or a split.

## The sequence

```text
Release 1  EXPAND    migration: add the new shape alongside the old (nullable, no constraints yet)
                     code: unchanged

Release 2  DUAL-WRITE code: write both old and new shapes; still read old
                     (no migration)

Release 3  BACKFILL  job: copy old → new for existing rows, batched, idempotent
                     verify: counts and spot checks match

Release 4  READ NEW  code: read from the new shape; still write both
                     migration (optional): add constraints now that data is complete — NOT VALID, then VALIDATE

Release 5  STOP OLD  code: stop writing the old shape
                     (no migration)

Release 6  CONTRACT  migration: drop the old shape
                     only once release 5 has been stable long enough that rolling it back is off the table
```

Not every step needs its own deploy — releases 2 and 3 can ship together if the backfill is triggered after deploy — but each *transition* has to be safe on its own, because a deploy is not atomic and a rollback might land on any of them.

## Worked example: rename `invoices.total` → `invoices.amount_cents`

1. **Expand** — `ALTER TABLE invoices ADD COLUMN amount_cents bigint;` (nullable). Code unchanged.
2. **Dual-write** — code writes both `total` and `amount_cents`; reads `total`.
3. **Backfill** — job: `UPDATE invoices SET amount_cents = total * 100 WHERE amount_cents IS NULL AND id BETWEEN $1 AND $2`, batched by ID range, run until no rows remain. Verify `COUNT(*) WHERE amount_cents IS NULL` is 0.
4. **Read new** — code reads `amount_cents`; writes both. Migration: `ALTER TABLE invoices ADD CONSTRAINT amount_cents_not_null CHECK (amount_cents IS NOT NULL) NOT VALID;` then, separately, `VALIDATE CONSTRAINT` (or, on PG12+, `ALTER COLUMN SET NOT NULL` after the validated check makes it a metadata-only change).
5. **Stop old** — code stops writing `total`.
6. **Contract** — `ALTER TABLE invoices DROP COLUMN total;` one or more releases later.

Six steps for a rename is the honest cost. The alternative is one step and a production outage the first time the table is larger than the laptop.

## Which changes need this

| Change | Additive? | Sequence |
|---|---|---|
| Add nullable column | yes | one migration |
| Add table | yes | one migration |
| Add index | yes, if `CONCURRENTLY` | one migration, `NO TRANSACTION` |
| Add column with `NOT NULL` | no | expand nullable → backfill → constrain |
| Add column with default on a large table | mostly (PG11+ constant defaults are metadata-only) | one migration; volatile defaults need the backfill path |
| Rename column / table | no | full sequence, or use a view during transition |
| Change column type | no | full sequence via a new column |
| Drop column / table | no | stop use → wait → drop; drop is the *last* step, never the first |
| Add foreign key | no | `NOT VALID` → `VALIDATE` |
| Add unique constraint | no | `CREATE UNIQUE INDEX CONCURRENTLY` → `ADD CONSTRAINT ... USING INDEX` |

## Tracking

Each PR in the sequence says which step it is and links the others: `expand (1/6) — see #123 for the plan`. A half-finished expand/contract that nobody remembers is a column named `total` that everyone is afraid to drop.
