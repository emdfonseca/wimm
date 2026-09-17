## S1 · See where the money went

**As a** member who owns an account at a bank wimm reads
**I want** the transactions behind its balance — newest first, grouped by day,
with the ones my bank has not settled yet marked as unsettled, and the whole
list telling me when it was last brought up to date
**so that** I can answer why we have less than we did on Friday without signing
in at the bank, and without trusting a figure whose age I cannot see

### INVEST

- **Independent** — Yes, for a bank connected after this change: the consent it
  asks for already covers transactions, so the ledger fills on its own. S2 is
  what rescues the banks connected before it, and S1 does not wait on it.
- **Negotiable** — Yes. It states a ledger, an order, an unsettled marker and a
  visible age. How far back history reaches, when the sync runs, what is stored
  against what is re-read, and how a duplicate is surfaced are all open.
- **Valuable** — Yes, and it is the half of the product's own question that has
  been missing. wimm can say the household has €11,693.55 and cannot say that
  €400 of it left on Tuesday. Every later capability — categories, budgets,
  reports — is a reading of this list, and none of them can be built over a
  balance.
- **Estimable** — Yes, with one unknown priced rather than discovered: banks
  disagree about whether a transaction carries a stable identifier, and a
  pending entry becomes a booked one under a different identifier, amount and
  date. Treating pending as a set that is rewritten each sync and booked as a
  list that only grows removes that problem instead of solving it, which is
  what makes the story sizeable.
- **Small** — **No.** It is a new destination, the primary navigation to reach
  it, the first stored ledger in the product, a first fill and an incremental
  one, and a freshness promise weaker than Overview's that has to be said out
  loud. Kept whole anyway: a screen that renders without a sync behind it shows
  nothing, and a sync with no screen is invisible. Neither half is usable.
- **Testable** — Yes, against `bankingtest` and the gateway's sandbox banks,
  which is what keeps `just check` off the network.

### Capabilities

- `banking/transactions`

### Satisfied by

- `banking/transactions`: Requirement: Transactions are seen by an account's owners
- `banking/transactions`: Requirement: The list of transactions
- `banking/transactions`: Requirement: The transactions of one account
- `banking/transactions`: Requirement: A transaction the bank has not settled
- `banking/transactions`: Requirement: The list says how current it is
- `banking/transactions`: Requirement: How far back the transactions go
- `banking/transactions`: Requirement: Two transactions that look identical are two transactions
- `banking/transactions`: Requirement: Reading a ledger longer than one page
- `banking/household-accounts`: Requirement: Who owns an account is the household's answer, not the bank's
- `banking/transactions`: Requirement: Access running out does not reset the ledger
- `banking/transactions`: Requirement: Before any transaction has been read

## S2 · Turn transactions on for a bank I already connected

**As a** member who connected a bank while wimm could only read balances
**I want** to widen what I granted by confirming once more at that bank, without
disconnecting it first and without finding it in the list again
**so that** the bank I set up last month starts filling the ledger today, rather
than on whatever date its access happens to run out

### INVEST

- **Independent** — **No.** Widening a consent with nowhere for the
  transactions to land delivers nothing; it needs S1's ledger to be worth
  doing. Accepted, because the alternative is telling every household already
  using wimm to wait up to ninety days for a grant to lapse before the feature
  they can see on screen starts working.
- **Negotiable** — Yes. It states that the member confirms again at their bank
  and keeps their accounts. That this re-enters the existing restore flow at
  the hand-off, and what the state is called where a connection is live but
  narrow, are open.
- **Valuable** — Yes. Without it the feature arrives for new households and is
  invisible to existing ones until their consent dies of old age, which is the
  worst possible first impression of a release.
- **Estimable** — Yes. The flow exists and already skips the picker and carries
  owners forward; this is a third reason to enter it and a state that is
  neither working nor expired.
- **Small** — Yes. One state on a connection, one banner that offers the way
  out of it, and one reason threaded through a route that already exists.
- **Testable** — Yes: a connection written with the old scope shows the state,
  the flow returns with the wider one, and the ledger fills.

### Capabilities

- `banking/bank-connections`

### Satisfied by

- `banking/bank-connections`: Requirement: A bank connected before wimm could read transactions
- `banking/bank-connections`: Requirement: Consenting at the bank

## S3 · Take a bank away without burning the record

**As a** member who wants wimm to stop reading a bank — because the account is
closed, or because I have changed my mind about it being here at all
**I want** disconnecting to end the access and leave what was already read
behind, and reconnecting later to find the same accounts rather than a second
copy of them
**so that** stopping wimm reading my bank is not the same decision as destroying
months of records, which no bank will hand back past ninety days

### INVEST

- **Independent** — **No.** The half that matters needs S1 to have produced a
  history worth keeping; today disconnection deletes accounts whose only
  content is a balance, and losing that is not a loss. The account-survives
  half is observable on its own. Accepted, because both halves are the same
  mechanism and splitting them means writing the delete twice.
- **Negotiable** — Yes. It states that access ends, the record stays and
  reconnecting re-attaches. Whether a disconnected account is still reachable
  anywhere, and what the confirmation says now that nothing is destroyed, are
  open.
- **Valuable** — Yes, and asymmetrically so: the cost of getting this wrong is
  unrecoverable by anyone, including us. A bank returns at most ninety days of
  history, so a delete here is permanent in a way almost nothing else in the
  product is.
- **Estimable** — Yes.
- **Small** — Yes, borderline. Stopping the delete and destroying the sealed
  identifier in its place is small; matching accounts on the gateway's
  cross-session hash *across* connections rather than within one is most of the
  work, and is the same rule restoring already uses one level down.
- **Testable** — Yes, and the test is the point: disconnect a bank, reconnect
  it, and assert one account rather than two with the ledger still attached.

### Capabilities

- `banking/household-accounts`
- `banking/bank-connections`
- `banking/transactions`

### Satisfied by

- `banking/household-accounts`: Requirement: Accounts leave with their connection
- `banking/bank-connections`: Requirement: Disconnecting a bank
- `banking/transactions`: Requirement: Transactions outlive the bank being taken away
