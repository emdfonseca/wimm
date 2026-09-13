# goose

## Layout and naming

```text
apps/billing/
├── migrations/
│   ├── 00001_create_invoices.sql
│   ├── 00002_add_invoice_currency.sql
│   └── 20260913120000_add_amount_cents.sql   ← in-flight; renumbered by goose fix before merge
└── internal/db/migrate.go                    ← embeds and applies
```

Names are `<version>_<snake_case_description>.sql`; the description says what changes, not why (`add_invoice_currency`, not `fix_bug_123`). One logical change per file.

## File format

```sql
-- +goose Up
SET lock_timeout = '5s';
ALTER TABLE invoices ADD COLUMN currency text;

-- +goose Down
ALTER TABLE invoices DROP COLUMN currency;
```

For statements that cannot run in a transaction:

```sql
-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY idx_invoices_customer ON invoices (customer_id);

-- +goose Down
DROP INDEX CONCURRENTLY idx_invoices_customer;
```

Multi-statement functions or triggers with semicolons inside need `-- +goose StatementBegin` / `-- +goose StatementEnd` around them.

## Embedding and applying

```go
//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations, /* no out-of-order */)
	if err != nil { return err }
	_, err = provider.Up(ctx)
	return err
}
```

Exposed as a `migrate` subcommand of the service binary (`billing migrate up|status`), so the deploy pipeline runs the same code developers do. The service's normal start path does not call it.

## Just verbs

```just
migrate *args:
    go run ./cmd/billing migrate {{args}}

migrate-redo:
    go run ./cmd/billing migrate down && go run ./cmd/billing migrate up

migrate-new name:
    goose -dir migrations create {{name}} sql
```

`migrate-redo` is how the Down block gets tested — on a local database, every time.

## Before merge

1. `goose -dir migrations fix` — renumber timestamp versions to sequential.
2. `just migrate-redo` passes.
3. `goose status` against staging shows the expected pending set and nothing surprising.

Out-of-order application stays off. If two branches both add migration `00007`, the second to merge renumbers; that is the whole cost, and it is smaller than the cost of a production sequence that was never tested.
