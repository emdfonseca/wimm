## Why

The bank picker offers every ASPSP Enable Banking returns for the country —
several dozen entries, most of which cannot actually be connected on wimm's
current Enable Banking plan, and whose logos come in every aspect ratio, which
today get cropped by a fixed square box. A member sees a long list of banks
that mostly fail at the hand-off, rendered with logos that clip and misalign
the names beside them.

## What Changes

- The bank list a member picks from is filtered to a wimm-configured set of
  banks, rather than every ASPSP the gateway returns for the country. Default
  set: Montepio, Revolut, PayPal, Activo Bank.
- The bank row's logo renders whole, at whatever aspect ratio the bank's mark
  is, without cropping — and every row's name stays aligned to the same column
  regardless of the logo's shape.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `banking/bank-connections`: "Choosing a bank to connect" now describes a
  wimm-curated list rather than the gateway's full country list, and adds the
  logo-legibility guarantee (no cropping, names aligned) as part of what a
  member sees when picking a bank.

## Impact

- `apps/wimm/internal/banking/enablebanking/client.go`: `Banks(country)` gains
  a configured allowlist filter.
- `apps/wimm/internal/config/config.go`: new operator config for the allowlist,
  alongside the existing `WIMM_ENABLEBANKING_*` settings.
- `packages/ui/src/molecules/BankRow.svelte`: logo box changes from a
  cropping fixed square to a non-cropping fit, with name alignment held
  constant across logo shapes.
- No new dependency, service, contract, or schema — filtering an existing
  gateway port method and a presentational fix. No ADR required.
