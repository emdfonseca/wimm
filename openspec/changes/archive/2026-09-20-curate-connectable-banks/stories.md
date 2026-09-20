## S1 · Pick from banks that actually connect

**As a** household member connecting a bank
**I want** the list to only offer banks wimm can actually connect on its
current plan
**so that** I don't pick a bank, go through the hand-off, and land back with
nothing connected

### INVEST

- **Independent** — Ships without S2; the list can be curated with the
  cropped logos unchanged.
- **Negotiable** — States the need (a working list), not the mechanism (an
  env var allowlist is one way to satisfy it).
- **Valuable** — A member no longer wastes a hand-off to the bank on a
  connection that was never going to work.
- **Estimable** — One filter point in an existing gateway call; well
  understood.
- **Small** — One list, one filter.
- **Testable** — The requirement's scenarios cover an offered bank
  succeeding and an unoffered one being absent from the list.

### Capabilities

- `banking/bank-connections`

### Satisfied by

- `banking/bank-connections`: Requirement: Choosing a bank to connect

## S2 · Recognise a bank by its logo, whatever its shape

**As a** household member connecting a bank
**I want** each bank's logo to render whole and every bank's name to line up
in the same place
**so that** I can tell banks apart at a glance instead of reading past
clipped marks and drifting text

### INVEST

- **Independent** — Ships without S1; applies to whatever list is shown.
- **Negotiable** — States the need (a legible, aligned row), not the
  mechanism (`object-fit`, a fixed name column, or something else).
- **Valuable** — A member can trust the logo they see is the bank's actual
  mark, not a crop of it.
- **Estimable** — One component's layout; well understood.
- **Small** — One row, one fix.
- **Testable** — The requirement's scenario covers logos of differing aspect
  ratios in the same list, checked for both no cropping and aligned names.

### Capabilities

- `banking/bank-connections`

### Satisfied by

- `banking/bank-connections`: Requirement: A bank's logo renders whole,
  whatever its shape
