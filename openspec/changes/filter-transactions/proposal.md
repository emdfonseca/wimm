## Why

A member looking for one payment on Transactions can only scroll or page through
every row they own. Narrowing to one account works only by typing `?account=` into
the address bar, a month is a place to start reading that never cuts the list
down, and Overview's month links (`/transactions?month=YYYY-MM`) land on the
newest page because the route ignores the parameter. The ledger now holds more
than a year of rows across several accounts, which is too many to find anything
by reading.

## What Changes

- **Account.** An on-screen picker narrows the list to one account or widens it
  to all. It lists the accounts the ledger already covers: owned by the member
  and not left out. One account at a time, never several.
- **Search.** Free text matched against who the transaction was with and the
  bank's line, ignoring case. Accents are not ignored.
- **Month.** Choosing a month shows only that month's transactions. This
  **replaces** the rule that a month is a place to start reading rather than a
  filter, and Overview's existing month links now open that month filtered.
- **Direction.** Money in, money out, or both, from the sign of the amount.
- Filters combine: a row is listed only when it matches every filter in force.
- Filters live in the address, so every way of paging carries them, Back works,
  and a filtered view can be returned to.
- The number of transactions and the page scrubber describe the filtered list,
  not the whole ledger.
- A filter matching nothing says so and offers to clear the filters. It never
  reads as having no bank connected.
- Filtering never widens what a member may see: only accounts they own and have
  not left out, as today.
- Out of scope: filtering by unusual, by between your accounts or by not
  settled (the first two are worked out on read and would break keyset paging),
  categories (none exist), and choosing several accounts at once.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `banking/transactions`: the account narrowing gains an on-screen picker; a
  month becomes a filter; search and direction are added; filters combine, live
  in the address, drive the count and the scrubber, and have their own empty
  state.

## Impact

- `packages/contracts/proto/wimm/banking/v1/banking.proto`:
  `ListTransactionsRequest` gains search, month and direction; `Ledger` gains
  the accounts and months a member can filter by.
  Regenerated with `just gen`. Server to server only, both ends ship together.
- `apps/wimm/internal/store/transactions.go`: `Ledger`, `CountLedger` and
  `LedgerPageIndex` take the filters, and a new read lists the months that hold
  matching rows. No migration, no index, no extension.
- `apps/wimm/internal/banking/ledger.go`, `apps/wimm/internal/rpc/banking.go`:
  carry the filters through.
- `apps/web/src/routes/(app)/transactions/`: reads the filters from the
  address, carries them on every pager link, applies a change as a navigation.
- `packages/ui`: new filter components, `TransactionsScreen.svelte` and its
  state stories.
- `apps/storybook/canvas/flows.js`: the new states join the Read transactions
  flow.
- No new dependency, service or schema.
