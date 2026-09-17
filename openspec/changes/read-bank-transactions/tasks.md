## 1. The decision record

- [x] 1.1 Write `docs/decisions/0021-*.md` covering the four things this change decides that outlive it: the consent scope widens and every existing connection needs a member to act, the ledger is keyset-paged because a sync inserts at the newest end, pending is a replaceable set while booked is append-only, and an account now outlives its connection. Verify `just adr-index` regenerates `.claude/rules/decisions.md` with 0021 in it.

## 2. Schema

- [x] 2.1 Add `00005_transactions.sql`: the `transactions` table with its unique index on `(account_id, dedup_key, occurrence)` for booked rows and its `(account_id, booking_date desc, id desc)` read index, `bank_connections.scope` defaulting to `balances`, and `accounts.bank_id`, `transactions_synced_through` and `transactions_synced_at`. Verify `just check apps/wimm` passes and the migration applies and rolls back cleanly.
- [x] 2.2 Backfill `accounts.bank_id` from each account's connection in the same migration, then constrain it for gateway-sourced accounts. Verify a test fixture with existing accounts comes out with every `bank_id` set and the constraint refuses a gateway account without one.
- [x] 2.3 Add the unique index on `(bank_id, gateway_ref)` where `gateway_ref` is not null. Verify a test asserts two accounts at one bank with the same gateway ref are refused, which is what makes re-attach have nothing to disambiguate.

## 3. Disconnection stops deleting

- [x] 3.1 Replace the `delete from accounts where connection_id = $1` in `DisconnectBankConnection` with nulling each account's `gateway_uid_sealed` and `key_id`. Verify a test disconnects a bank and asserts the accounts still exist, hold no sealed value, and are absent from `VisibleAccounts`.
- [x] 3.2 Re-attach on reconnection: completing a connection at a bank matches each returned account on `(bank_id, gateway_ref)` and re-points the existing row rather than inserting. Verify a test connects, disconnects, reconnects and asserts one account row with its owners, grants and transactions intact.

## 4. The gateway port

- [x] 4.1 Add `Transactions(ctx, conn, account, TransactionsRequest) (TransactionsPage, error)` to `Gateway` with its request and page types, and the `Transaction` struct carrying signed minor units. Verify `just check apps/wimm` compiles every implementation.
- [x] 4.2 Implement it in `bankingtest` with controllable pages, a cursor, and injectable failures. Verify its own tests cover paging to exhaustion and a mid-page error.
- [x] 4.3 Implement it in `enablebanking`: a zero `From` maps to `strategy=longest`, a set `From` to `strategy=default` with `date_from`, `continuation_key` drives paging, and `BOOK`/`PEND`/`OTHR` map to booked, pending and booked. Verify table tests over recorded response bodies, including a response with no `entry_reference`.
- [x] 4.4 Retry `WRONG_TRANSACTIONS_PERIOD` once with `strategy=longest` inside the adapter and surface only a second failure. Verify a test asserts one retry, and that the taxonomy gains no member a person could be shown.
- [x] 4.5 Widen the consent in `BeginConnection` to ask for transactions alongside balances, and write `scope` on the connection. Verify a test asserts the request body carries both and that an existing row still reads as `balances`.

## 5. Storing a ledger

- [x] 5.1 Compute the dedup key: the bank's `entry_reference` where given, otherwise a digest over booking date, amount, currency, counterparty and remittance. Verify unit tests pin both branches and that the digest is stable across runs.
- [x] 5.2 Assign `occurrence` by counting how many rows with one key a sync returned against how many are stored, so the same payment made twice is two rows and the same transaction read twice is one. Verify a test feeds overlapping syncs containing a genuine duplicate pair and asserts the counts.
- [x] 5.3 Write booked rows append-only and replace an account's pending rows wholesale each sync. Verify a test turns a pending row into a booked row with a different reference, amount and date, and asserts one row remains.
- [x] 5.4 Read a page by seeking on `(booking_date, id)` in both directions, returning the span of dates and whether an older or newer page exists. Verify tests page to both ends, assert nothing is repeated or skipped, and assert a page is unchanged after rows are inserted at the newest end.
- [x] 5.5 Scope every read to accounts the member owns, joining `account_owners` only. Verify a test gives a member `balance` and `details` grants on accounts they do not own and asserts they read zero transactions and no count.
- [x] 5.6 Record `transactions_synced_through` and `transactions_synced_at` when a sync finishes, including a partial fill. Verify a test asserts a capped first fill records how far it actually reached and that the next sync continues rather than restarting.

## 6. Ownership

- [x] 6.1 Assert in tests that nothing infers, suggests or defaults an owner from `holder_name`, including where it matches a member's name exactly. Verify the test fails if such inference is added.
- [x] 6.2 Expose an account's owners so a joint account reads as one both members own, without editing `holder_name`. Verify a test asserts the bank's holder name is returned verbatim beside two owners.

## 7. Syncing

- [x] 7.1 Sync one account: a first fill when `transactions_synced_through` is null, otherwise an incremental read from that date less the overlap window. Verify tests cover both branches and that the overlap re-reads without duplicating.
- [x] 7.2 Bound a sync by a configured maximum pages per account and a configured minimum interval between syncs of one account, both refused at startup when non-positive. Verify tests assert a second arrival inside the interval performs no fetch.
- [x] 7.3 Trigger syncs from a member's arrival and from Refresh, with PSU-present headers, never from a ticker. Verify a test asserts no sync path exists that runs without a member.
- [x] 7.4 Map sync failures onto the existing taxonomy and leave stored rows and their time untouched on every one of them. Verify a test asserts the rows and `transactions_synced_at` are unchanged after each failure kind.

## 8. Contracts and routes

- [x] 8.1 Add the Connect methods for reading a page and for triggering a sync, and regenerate with `just gen`. Verify `just check` passes and nothing under `gen/` is hand-edited.
- [x] 8.2 Add SvelteKit routes for `/transactions`, `?account=<id>`, `?before=<cursor>` and `?after=<cursor>`, holding the session and calling Connect from the server. Verify route tests cover each query shape and that a cursor for an account the member does not own returns nothing.
- [x] 8.3 Thread a third reason through the existing restore route so a live, narrow connection can be widened without disconnecting or picking the bank again. Verify a test asserts owners and grants survive and that the connection's scope changes.

## 9. Components, before the screens that use them

- [x] 9.1 Build `LedgerRow`: the origin in `product-ui.lib.pen` beside `Transaction row`, `packages/ui/src/molecules/LedgerRow.svelte`, and `LedgerRow.stories.svelte`. It carries a mark, a description, the account and its bank, an unsettled badge, a date and a signed amount, and has no category column and no selection control. Verify `just check packages/ui` passes and `just pen-manifest verify` reports only additions.
- [x] 9.2 Build `SeekPager`: the origin beside `Pagination`, `packages/ui/src/molecules/SeekPager.svelte`, and `SeekPager.stories.svelte`. It states a span of dates and offers Newer and Older, and on the oldest page says there is nothing older rather than greying a control. Verify `just check packages/ui` passes and stories cover the newest page, a middle page and the oldest page.
- [x] 9.3 Build `BottomNav`: the origin in `product-ui.lib.pen` beside `Sidebar nav`, `packages/ui/src/organisms/BottomNav.svelte`, and `BottomNav.stories.svelte`. Items are `Nav item` instances laid out vertically and centred; the current one carries an accent top edge, an accent icon, an accent label at 600 and `aria-current="page"`. Verify `just check packages/ui` passes, a story covers each item being current, and the targets measure at least 44 px.
- [x] 9.4 Add Transactions to the shell's destinations, in `SignedInLanding`'s default and in the drawn template's overrides, and mark the current one from the route. Verify a story shows both destinations with Transactions current and `just check packages/ui` passes.

## 10. The screens

- [x] 10.1 Build the Transactions page component composing those instances, with props for every drawn state, plus stories for each. Verify the stories cover all eleven states in canvas.md's index and `just check packages/ui` passes.
- [x] 10.2 Build the compact composition from the same component, with `BottomNav` as a sibling below the scrolling region rather than an overlay, so no bottom padding is needed to clear it. Verify a story at 390 renders the stacked row and the bar, that the last row is fully visible, and `just check packages/ui` passes.
- [x] 10.3 Build the widening explainer screen for J08.A / 01. Verify its copy matches the frame and `just check packages/ui` passes.
- [x] 10.4 Change the disconnect dialog's message to say the transactions already read are kept. Verify `just check packages/ui` passes and the copy matches `zfsws`.

## 11. The canvas checks

- [x] 11.1 Map all seventeen new frames to the files that render them in `check-canvas.py`, and regenerate `canvas-contract.json`. Verify `just check packages/ui` passes with no unmapped frame, and that removing one mapping makes it fail.
- [x] 11.2 Add a fixture asserting the generator still extracts notice and empty-state copy from `descendants` overrides on these frames. Verify `test-contract.py` fails when the resolution is removed.

## 12. Verification

- [x] 12.1 Run `just check apps/wimm && just check apps/web && just check packages/ui` and fix everything it reports.
- [x] 12.2 Connect a bank against the gateway sandbox, confirm a first fill arrives and pages, then disconnect and reconnect and confirm one account with its history. Verify by reading the screens, not only the tests.
