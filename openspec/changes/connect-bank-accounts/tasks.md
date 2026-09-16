## 1. Settle the gateway before anything is built on it

- [ ] 1.1 Ask Enable Banking whether the free Restricted Production tier, which covers accounts the application owner personally links, extends to a second household member's bank. Verify by getting an answer in writing. This decides whether wimm works for a household or only for one person, and no documentation answers it.
- [ ] 1.2 Confirm the target country's banks are covered and that a sandbox reachable from a developer machine accepts a localhost redirect. Verify by listing the sandbox banks with a throwaway script and naming at least one real bank the household would use.
- [ ] 1.3 Register the application, generate the RSA key pair, upload the public key, record the accepted redirect URL, and record `maximum_consent_validity` for the target banks. Verify by signing a JWT and calling the bank list successfully. No key material enters the repository.
- [ ] 1.4 Reconcile `docs/decisions/0018-reading-banks-through-a-gateway.md` with what 1.1 to 1.3 found and regenerate the index. Verify with `just adr-index` and that `.claude/rules/decisions.md` carries 0018. If 1.1 or 1.2 rules Enable Banking out, this is where the ADR changes gateway; the port does not.
- [ ] 1.5 Bump every pin in `devbox.json` as ADR 0015 requires, since this change touches it. Verify with `devbox run -- just ci`.

## 2. Sealing what reaches a bank

Built before the adapter, so no plaintext session identifier is ever written
even during development.

- [ ] 2.1 Implement `banking.Sealed`: AES-256-GCM with the key read at startup from the path in `WIMM_BANKING_ENCRYPTION_KEY`, the owning row's uuid as additional authenticated data, and a stored key id. Verify with tests asserting a ciphertext from one row fails to open against another, and that a wrong key fails closed.
- [ ] 2.2 Make a sealed value unprintable: `String`, `MarshalJSON` and `slog.LogValuer` all return a redaction, and plaintext exists only as a local for the length of one call. Verify with a test that formats a struct containing one every way Go offers and asserts the plaintext appears in none of them.
- [ ] 2.3 Refuse to start when a gateway is configured and either key file is missing, unreadable, or group or world readable. Verify with tests covering each case.
- [ ] 2.4 Implement key rotation by key id: open with the id the row carries, write with the current one. Verify with a test that opens a value written under a previous key.

## 3. The banking port

- [ ] 3.1 Define `apps/wimm/internal/banking` with the `Gateway` interface, the domain types, and money as `int64` minor units plus an ISO 4217 code. Verify the package compiles and `just check apps/wimm` passes with no implementation present.
- [ ] 3.2 Define the error taxonomy: bank unavailable, gateway unavailable, consent declined, consent expired, no accounts, rate limited with a retry-after. Verify with table-driven tests asserting each is distinguishable by the caller. Work this one test-first.
- [ ] 3.3 Add the import restriction so nothing outside `internal/banking/...` may import a gateway adapter, in the lint configuration that already bans `time.Now`. Verify by adding an offending import, seeing `just check apps/wimm` fail, and removing it.
- [ ] 3.4 Build `internal/banking/bankingtest` as an in-memory `Gateway` with scriptable outcomes for every error in 3.2, plus a bank that offers a new account on restore and one that withdraws an account. Verify the whole service's tests run against it with no network.

## 4. The Enable Banking adapter

- [ ] 4.1 Implement JWT request signing, reading the key from the path in `WIMM_ENABLEBANKING_PRIVATE_KEY`. Verify with a test asserting the signed header's claims and that the key never reaches a log.
- [ ] 4.2 Implement `Banks`, carrying each bank's `maximum_consent_validity` through to the domain type so the consent screen can state a real date. Verify against the sandbox with a build-tagged integration test that is not part of `just check`.
- [ ] 4.3 Implement `BeginConnection` requesting the `accounts` and `balances` scopes and never `transactions`, with `valid_until` set to that bank's maximum. Verify with a test asserting the request body carries no transactions scope.
- [ ] 4.4 Implement `CompleteConnection`, persisting everything the session returns because those details are returned once and cannot be read again. Verify with a test asserting every account in the response reaches the store, shared and unshared alike.
- [ ] 4.5 Implement `Balances` and `EndConnection`, sending PSU headers on every member-initiated read so the background rate limit does not apply. Verify with a test asserting the headers are present and an integration test against the sandbox.
- [ ] 4.6 Map every failure onto the taxonomy from 3.2, table-driven, including `EXPIRED_SESSION` 401 and `ASPSP_RATE_LIMIT_EXCEEDED` 429 with its retry-after. Verify with unit tests fed recorded responses.

## 5. Schema

- [ ] 5.1 Add the goose migration creating `bank_connections`, `pending_bank_connections` and `bank_accounts` as `design.md` sets out, with account identity on the gateway's cross-session hash and the session-scoped values sealed. Verify migrate up and down both run clean against `wimm_test`.
- [ ] 5.2 Add the `expires_at` index on `pending_bank_connections`, built `CONCURRENTLY` in a `NO TRANSACTION` migration as ADR 0017 established. Verify the migration runs and the sweep's plan uses it.
- [ ] 5.3 Add the store methods for connections, pending connections and accounts, comparing every lifetime against database time. Verify with store tests, including that a consumed pending connection is distinguishable from one that never existed.
- [ ] 5.4 Destroy sealed values on disconnection and on a connection reaching the end of its grant. Verify with a test asserting the row survives for audit and nothing openable remains on it.
- [ ] 5.5 Add abandoned pending connections as a fourth statement in the existing sweep, with its own count in the sweep's log line. Verify with a test asserting the other three statements still run when this one fails.

## 6. The service surface

- [ ] 6.1 Add `packages/contracts/proto/wimm/banking/v1/banking.proto` with protovalidate constraints, no gateway vocabulary, and no sealed value in any message. Verify with `just gen` and `just check packages/contracts`.
- [ ] 6.2 Implement the handlers: list banks, begin a connection, complete a connection, list a connection's accounts for choosing, set what is shared, read balances, restore, disconnect. Verify with handler tests driven by `bankingtest`, covering every scenario in `specs/banking/bank-connections`.
- [ ] 6.3 Implement household visibility: every signed-in member sees every shared account, an unshared account is returned to nobody, and `connected_by` is returned for display. Verify with a test where one member connects and shares two of three, and a different member reads.
- [ ] 6.4 Implement reading on arrival and on request, with no read for an unshared account, and the partial path where one bank fails leaving other readings untouched. Verify with tests asserting no reading is ever lost and no unshared account is ever read.
- [ ] 6.5 Implement restore: match accounts by cross-session hash, carry `shared` forward, surface a newly offered account as unshared, and drop one the bank no longer offers. Verify with tests for all three cases using `bankingtest`.

## 7. Design system components

Built before any screen that instances them, per ADR 0008. Each task is the
origin in `product-ui.lib.pen`, the Svelte component in `packages/ui/src`, and
its story beside it.

- [ ] 7.1 Add the `Bank row` molecule with `BankRow.svelte` and its story covering rest, hover, focus and pressed. Verify with `just check packages/ui` and `just pen-manifest verify`.
- [ ] 7.2 Add the `Account choice row` molecule with `AccountChoiceRow.svelte` and its story covering unchecked, checked, newly offered with its badge, and disabled. Verify with `just check packages/ui`.
- [ ] 7.3 Extend `Account selector` additively with a `Reading` text and a `Badge` slot, both disabled by default, and mirror it in `AccountRow.svelte` with a story covering fresh, stale and no-longer-updating readings and an overdrawn balance carrying its sign. Verify that every existing instance renders unchanged, with `just pen-manifest write` accepting the deliberate change and the diff reviewed.
- [ ] 7.4 Add the `color-scrim` token to `design/tokens.json`, re-export the library and run `just gen`. Verify `just gen` is a no-op afterwards and `just check packages/ui` passes. Nothing hand-edits `tokens.css`.

## 8. Screens

- [ ] 8.1 Build `ChooseBankScreen` in `packages/ui/src/pages` with its story, covering default, narrowed, no match and list unavailable. Verify with `just check packages/ui`.
- [ ] 8.2 Build `ConsentExplainerScreen` with its story, stating the scope and the real date access will end. Verify with `just check packages/ui`.
- [ ] 8.3 Build `ChooseAccountsScreen` with its story, covering nothing chosen with Finish disabled, some chosen, all chosen, reopened with an existing selection, and restoring with a newly offered account. Verify with `just check packages/ui`.
- [ ] 8.4 Build `AccountsOverview` with its story, covering empty, connected, two currencies, refresh refused, one bank not answering, and access run out. Verify with `just check packages/ui`.
- [ ] 8.5 Add the disconnect confirmation using the existing `Dialog` over a scrim, naming the bank and the account count. Verify with a story asserting the copy carries both and that the buttons read as the design says rather than as the component's defaults.

## 9. Routes

- [ ] 9.1 Add `/connect` and `/connect/[bank]` SvelteKit routes calling Connect from the server. Verify with route tests covering a member who is not signed in.
- [ ] 9.2 Add the consent return route, exchanging the one-time value server side and redirecting to the chooser. Verify with a test asserting the redirect target contains no callback value and that a replayed return connects nothing twice.
- [ ] 9.3 Add `/connect/[bank]/accounts` for choosing and for reopening the choice later. Verify with a test asserting a member who closes the tab mid-choice can return to it, since access is already granted by then.
- [ ] 9.4 Wire reading on arrival, refresh, restore and disconnect from Overview. Verify with tests covering the rate-limited response, the partial failure, and a slow bank not blocking the screen.
- [ ] 9.5 Implement the behavioural contracts in `canvas.md`: focus on return, on failure and on dialog dismissal; live-region announcements for refresh, for search results and for the chooser's running count; and keyboard completion of every step. Verify with tests asserting focus target and announcement text, since the canvas cannot execute any of it.

## 10. Close out

- [ ] 10.1 Confirm by inspection that no log line, error message or Connect message emitted anywhere in this change carries a value that could reach a bank. Verify by running the full suite with logging at its most verbose and grepping the output for the sandbox's session identifier.
- [ ] 10.2 Update `README.md`: how a member connects a bank and chooses what to share, what happens when access runs out, what the operator can and cannot do, and the new configuration including both key files. Verify by following it on a clean checkout against the sandbox.
- [ ] 10.3 Run `devbox run -- just ci` and confirm it is green.
