## 1. Configuration

- [x] 1.1 In `apps/wimm/internal/config/config.go`, add
  `ConnectableBanks []string` to `Config`, populated from
  `WIMM_BANKING_CONNECTABLE_BANKS` (comma-separated, each entry trimmed,
  blanks dropped), falling back to a `DefaultConnectableBanks` constant of
  `Caixa Económica Montepio Geral`, `Revolut`, `PayPal`, `Activo Bank` when the
  env var is unset. Verify with a `config_test.go` case asserting the default
  applies when unset and a custom list is parsed when set.

## 2. Filtering the picker

- [x] 2.1 In `apps/wimm/internal/banking/service.go`, filter `Service.Banks`'
  result to banks whose `Name` case-insensitively matches an entry in
  `Config.ConnectableBanks`, leaving `BeginConnection` and
  `RestoreConnection`'s own `gateway.Banks` calls unfiltered. Verify with a
  `service_test.go` case using `bankingtest` seeded with banks both inside and
  outside the configured set, asserting `Banks` returns only the configured
  ones while `RestoreConnection` still resolves a bank no longer in that set.
- [x] 2.2 Verify `just check apps/wimm` passes.

## 3. The bank row's logo

- [x] 3.1 In `packages/ui/src/molecules/BankRow.svelte`, change `.mark`'s
  `object-fit` from `cover` to `contain` and add
  `background: var(--color-bg-subtle)` so a non-square logo letterboxes
  instead of appearing to float on nothing. Verify in Storybook (`just check
  packages/ui` runs its checks) with a story case adding a wide, non-square
  logo alongside the existing square one, confirming neither is cropped and
  both names start at the same horizontal position.
- [x] 3.2 Verify `just check packages/ui` passes.

## 4. Whole-repo check

- [x] 4.1 Run `just check` across every touched directory and
  `devbox run -- just ci`. Verify both pass.
