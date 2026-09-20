## S1 · See where things stand, without reading past account admin

**As a** member of a household using wimm to track shared money
**I want** Overview to show my total, my recent activity, and the trend, and
nothing about managing bank connections
**so that** I can tell where things stand in the time it takes to open the app

### INVEST

- **Independent** — Yes: this is the dashboard content itself, shippable once
  the total (already computed) has a recent-transactions section and a trend
  beside it.
- **Negotiable** — Yes: it states what Overview must show, not the chart
  library or component names.
- **Valuable** — Yes: today a member sees bank-management chrome before any
  figure that answers "where do we stand"; this puts the answer first.
- **Estimable** — Yes: the total already exists; recent transactions and the
  trend are both reads of data already stored.
- **Small** — No, honestly: this is "the new Overview" as a whole, which is
  three sections (total, recent, trend) plus the two negative-space
  requirements (nothing about accounts, nothing about connections). Splitting
  those three into separate stories would let one ship without the others,
  producing a dashboard that is still half bank-management — accepted as one
  story because the value is the combination, not any one section alone.
- **Testable** — Yes: each section's presence and content is a scenario.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: Overview shows the total and nothing about
  accounts or connections
- `banking/overview`: Requirement: Overview shows a recent slice of the
  member's own transactions
- `banking/overview`: Requirement: The trend follows the ledger, and only as
  far as the ledger reaches
- `banking/overview`: Requirement: Before any bank is connected, Overview
  says so

## S2 · Manage banks in one place, away from the screen I check daily

**As a** member connecting, restoring, or disconnecting a bank
**I want** a dedicated Accounts screen for that work
**so that** managing a connection never crowds out the dashboard I open every
day, and I still have one clear place to do it

### INVEST

- **Independent** — No, deliberately: this is the other half of the Overview
  split. Shipping S1 without this would leave the account list and every
  connection action with no screen to live on at all — accepted, both stories
  land in the same change for exactly that reason.
- **Negotiable** — Yes: it names the destination and what it carries, not the
  route path or component layout.
- **Valuable** — Yes: a member can still connect, restore, disconnect, and
  choose who sees an account — on a screen that does not double as the daily
  check-in.
- **Estimable** — Yes: the account list and every connection action already
  exist; this relocates them and adds the navigation entry and the arrival
  balance read they carry with them.
- **Small** — Yes: it is a relocation of existing, already-specified
  behaviour plus one new "reachable from navigation" requirement, not new
  behaviour.
- **Testable** — Yes: reachability, the account list, and every connection
  action are each a scenario.

### Capabilities

- `banking/household-accounts`
- `banking/bank-connections`

### Satisfied by

- `banking/household-accounts`: Requirement: Accounts is reachable from the
  primary navigation
- `banking/household-accounts`: Requirement: Seeing the accounts a member may
  see
- `banking/household-accounts`: Requirement: A balance is a reading, not a
  live figure
- `banking/household-accounts`: Requirement: Before any bank is connected
- `banking/household-accounts`: Requirement: Accounts leave with their
  connection
- `banking/bank-connections`: Requirement: Access that has run out
- `banking/bank-connections`: Requirement: Restoring access to a bank
