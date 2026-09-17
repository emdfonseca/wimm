## 1. Settle the gateway before anything is built on it

- [x] 1.1 Ask Enable Banking whether the free Restricted Production tier, which covers accounts the application owner personally links, extends to a second household member's bank. Verify by getting an answer in writing. This decides whether wimm works for a household or only for one person, and no documentation answers it.
- [x] 1.2 Confirm the target country's banks are covered and that a sandbox reachable from a developer machine accepts a localhost redirect. Verify by listing the sandbox banks with a throwaway script and naming at least one real bank the household would use.
- [x] 1.3 Register the application, generate the RSA key pair, upload the public key, record the accepted redirect URL, and record `maximum_consent_validity` for the target banks. Verify by signing a JWT and calling the bank list successfully. No key material enters the repository.
- [x] 1.4 Reconcile `docs/decisions/0018-reading-banks-through-a-gateway.md` with what 1.2 and 1.3 found, and write `docs/decisions/0019-accounts-have-owners-and-per-member-levels.md` superseding 0018's sharing half — its "member chooses what is shared", "household-visible once shared" and "`shared` defaulting to false". 0018's port, cross-session hash, sealing, money type and "the chooser is not a consent step" are unchanged. Regenerate with `just adr-index` and verify `.claude/rules/decisions.md` carries both.
- [x] 1.5 Bump every pin in `devbox.json` as ADR 0015 requires, since this change touches it. Verify with `devbox run -- just ci`.

## 2. Sealing what reaches a bank

Built before the adapter, so no plaintext session identifier is ever written
even during development.

- [x] 2.1 Implement `banking.Sealed`: AES-256-GCM with the key read at startup from the path in `WIMM_BANKING_ENCRYPTION_KEY`, the owning row's uuid as additional authenticated data, and a stored key id. Verify with tests asserting a ciphertext from one row fails to open against another, and that a wrong key fails closed.
- [x] 2.2 Make a sealed value unprintable: `String`, `MarshalJSON` and `slog.LogValuer` all return a redaction, and plaintext exists only as a local for the length of one call. Verify with a test that formats a struct containing one every way Go offers and asserts the plaintext appears in none of them.
- [x] 2.3 Refuse to start when a gateway is configured and either key file is missing, unreadable, or group or world readable. Verify with tests covering each case.
- [x] 2.4 Implement key rotation by key id: open with the id the row carries, write with the current one. Verify with a test that opens a value written under a previous key.

## 3. The banking port

- [x] 3.1 Define `apps/wimm/internal/banking` with the `Gateway` interface, the domain types, and money as `int64` minor units plus an ISO 4217 code. Verify the package compiles and `just check apps/wimm` passes with no implementation present.
- [x] 3.2 Define the error taxonomy: bank unavailable, gateway unavailable, consent declined, consent expired, no accounts, rate limited with a retry-after. Verify with table-driven tests asserting each is distinguishable by the caller. Work this one test-first.
- [x] 3.3 Add the import restriction so nothing outside `internal/banking/...` may import a gateway adapter, in the lint configuration that already bans `time.Now`. Verify by adding an offending import, seeing `just check apps/wimm` fail, and removing it.
- [x] 3.4 Build `internal/banking/bankingtest` as an in-memory `Gateway` with scriptable outcomes for every error in 3.2, plus a bank that offers a new account on restore and one that withdraws an account. Verify the whole service's tests run against it with no network.

## 4. The Enable Banking adapter

- [x] 4.1 Implement JWT request signing, reading the key from the path in `WIMM_ENABLEBANKING_PRIVATE_KEY`. Verify with a test asserting the signed header's claims and that the key never reaches a log.
- [x] 4.2 Implement `Banks`, carrying each bank's `maximum_consent_validity` through to the domain type so the consent screen can state a real date. Verify against the sandbox with a build-tagged integration test that is not part of `just check`.
- [x] 4.3 Implement `BeginConnection` requesting the `accounts` and `balances` scopes and never `transactions`, with `valid_until` set to that bank's maximum. Verify with a test asserting the request body carries no transactions scope.
- [x] 4.4 Implement `CompleteConnection`, persisting everything the session returns because those details are returned once and cannot be read again. Verify with a test asserting every account in the response reaches the store, owned by the connecting member and granted to nobody.
- [x] 4.5 Implement `Balances` and `EndConnection`, sending PSU headers on every member-initiated read so the background rate limit does not apply. Verify with a test asserting the headers are present and an integration test against the sandbox.
- [x] 4.6 Map every failure onto the taxonomy from 3.2, table-driven, including `EXPIRED_SESSION` 401 and `ASPSP_RATE_LIMIT_EXCEEDED` 429 with its retry-after. Verify with unit tests fed recorded responses.

## 5. Schema

- [x] 5.1 Add the goose migration creating `bank_connections`, `pending_bank_connections`, `bank_accounts`, `bank_account_owners` and `bank_account_grants` as `design.md` sets out, with account identity on the gateway's cross-session hash, the session-scoped values sealed, `level` a `text` column with a `CHECK` constraint allowing `balance` and `details` only — not a Postgres `enum` type, whose fourth value would need `ALTER TYPE … ADD VALUE` outside a transaction — no `connection_id` on a grant, and both member references cascading. Verify migrate up and down both run clean against `wimm_test`, and that a member delete removes their owner and grant rows.
- [x] 5.2 Add the `expires_at` index on `pending_bank_connections`, built `CONCURRENTLY` in a `NO TRANSACTION` migration as ADR 0017 established. Verify the migration runs and the sweep's plan uses it.
- [x] 5.3 Add the store methods for connections, pending connections and accounts, comparing every lifetime against database time. Verify with store tests, including that a consumed pending connection is distinguishable from one that never existed.
- [x] 5.4 Destroy sealed values on disconnection and on a connection reaching the end of its grant. Verify with a test asserting the row survives for audit and nothing openable remains on it.
- [x] 5.5 Add abandoned pending connections as a fourth statement in the existing sweep, with its own count in the sweep's log line. Verify with a test asserting the other three statements still run when this one fails.
- [x] 5.6 Enforce that one member never holds both an owner row and a grant row on one account, refusing the pair in the store rather than resolving it. Verify with a test asserting the write is refused and that neither row is left behind.

## 6. The service surface

- [x] 6.1 Add `packages/contracts/proto/wimm/banking/v1/banking.proto` with protovalidate constraints, no gateway vocabulary, and no sealed value in any message. The account message carries its owners and the requesting member's level, and a level enum whose zero value is the hidden one so an unset field cannot mean "visible". Verify with `just gen` and `just check packages/contracts`.
- [x] 6.2 Implement the handlers: list banks, begin a connection, complete a connection, list a connection's accounts for choosing, set an account's owners, set a member's level on an account, read balances, restore, disconnect. Setting owners or levels is refused for a member who does not own the account. Verify with handler tests driven by `bankingtest`, covering every scenario in `specs/banking/bank-connections`.
- [x] 6.3 Implement per-member visibility, computed for a set of accounts rather than for a connection so the same code can later answer across banks: an owner sees their account in full, a granted member sees exactly their level and no field above it, an ungranted member is returned nothing at all, and `connected_by` is returned for display. Redaction happens in the handler, so a field the member may not see is never serialised. Verify with a test where one member owns three accounts, grants a second member `balance` on one and `details` on another, and both members read — asserting the second member's response carries no number suffix, type or holder name on the `balance` account.
- [x] 6.4 Implement reading on arrival and on request, with no read for an account that has no owner and no grant, and the partial path where one bank fails leaving other readings untouched. Verify with tests asserting no reading is ever lost, that an account disowned in the chooser is never read at all, and that a member's totals cover exactly the accounts they may see.
- [x] 6.5 Implement restore: match accounts by cross-session hash, carry owners and grants forward, give a newly offered account to the restoring member with no grants, and drop one the bank no longer offers along with its owners and grants. Verify with tests for all three cases using `bankingtest`.

## 7. Design system components

Built before any screen that instances them, per ADR 0008. Each task is the
origin in `product-ui.lib.pen`, the Svelte component in `packages/ui/src`, and
its story beside it.

- [x] 7.1 Add the `Bank row` molecule with `BankRow.svelte` and its story covering rest, hover, focus and pressed. Verify with `just check packages/ui` and `just pen-manifest verify`.
- [x] 7.2 Add the `Segmented control` atom with `SegmentedControl.svelte` and its story covering each of three options selected, focus, and disabled. One control, one tab stop, arrow keys between options. Verify with `just check packages/ui` and `just pen-manifest verify`.
- [x] 7.3 Add the `Account choice row` molecule with `AccountChoiceRow.svelte` and its story covering owned, disowned, a second member at each of the three levels, newly offered with its badge, and Compact with the level controls stacked. It instances `Segmented control` from 7.2 and carries a balance. Verify with `just check packages/ui`.
- [x] 7.4 Extend `Account selector` additively with a `Reading` text and a `Badge` slot, both disabled by default, and mirror it in `AccountRow.svelte` with a story covering fresh, stale and no-longer-updating readings and an overdrawn balance carrying its sign. Verify that every existing instance renders unchanged, with `just pen-manifest write` accepting the deliberate change and the diff reviewed.
- [x] 7.5 Add the `color-scrim` token to `design/tokens.json`, re-export the library and run `just gen`. Verify `just gen` is a no-op afterwards and `just check packages/ui` passes. Nothing hand-edits `tokens.css`.
- [x] 7.6 Redraw the five chooser frames marked † in `canvas.md` against the ownership-and-levels row, at Wide and Compact, and update `canvas.md`'s frame table and States drawn to match what is on the canvas. Draw with `just pen-exec` and verify in a second call, per ADR 0011.

## 8. Screens

- [x] 8.1 Build `ChooseBankScreen` in `packages/ui/src/pages` with its story, covering default, narrowed, no match and list unavailable. Verify with `just check packages/ui`.
- [x] 8.2 Build `ConsentExplainerScreen` with its story, stating the scope and the real date access will end. Verify with `just check packages/ui`.
- [x] 8.3 Build `ChooseAccountsScreen` with its story, covering as-it-opens with every row owned and every other member at nothing, one row disowned, a second member at `balance` on one row and `details` on another, reopened with existing owners and levels, a member who owns none of it and can change nothing, and restoring with a newly offered account. Finish is enabled throughout. Verify with `just check packages/ui`.
- [x] 8.4 Build `AccountsOverview` with its story, covering empty, connected, an account seen at `balance` with no identifiers rendered beside one seen in full, two currencies, refresh refused, one bank not answering, and access run out. Verify with `just check packages/ui`.
- [x] 8.5 Add the disconnect confirmation using the existing `Dialog` over a scrim, naming the bank and the account count. Verify with a story asserting the copy carries both and that the buttons read as the design says rather than as the component's defaults.

## 9. Routes

- [x] 9.1 Add `/connect` and `/connect/[bank]` SvelteKit routes calling Connect from the server. Verify with route tests covering a member who is not signed in.
- [x] 9.2 Add the consent return route, exchanging the one-time value server side and redirecting to the chooser. Verify with a test asserting the redirect target contains no callback value and that a replayed return connects nothing twice.
- [x] 9.3 Add `/connect/[bank]/accounts` for choosing and for reopening the choice later. Verify with a test asserting a member who closes the tab mid-choice can return to it, since access is already granted by then, and one asserting a member who owns none of the bank's accounts can change nothing there.
- [x] 9.4 Wire reading on arrival, refresh, restore and disconnect from Overview. Verify with tests covering the rate-limited response, the partial failure, and a slow bank not blocking the screen.
- [x] 9.5 Implement the behavioural contracts in `canvas.md`: focus on return, on failure and on dialog dismissal; live-region announcements for refresh, for search results and for a level change naming the account and its new level; and keyboard completion of every step, including moving between the three levels with arrow keys inside one control. Verify with tests asserting focus target and announcement text, since the canvas cannot execute any of it.

## 10. Close out

- [x] 10.1 Confirm by inspection that no log line, error message or Connect message emitted anywhere in this change carries a value that could reach a bank. Verify by running the full suite with logging at its most verbose and grepping the output for the sandbox's session identifier.
- [x] 10.2 Update `README.md`: how a member connects a bank, who owns the accounts it returns, and how the three levels decide what each other member sees, what happens when access runs out, what the operator can and cannot do, and the new configuration including both key files. Verify by following it on a clean checkout against the sandbox.
- [x] 10.3 Run `devbox run -- just ci` and confirm it is green.
