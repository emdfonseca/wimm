## S1 · Know which rows are my own money moving

**As a** household member who owns several accounts and moves money between
them
**I want** wimm to recognise a transfer between my own accounts, and to show me
which rows it took for one
**so that** I can check it against what I know I did, and trust a figure that
leaves them out

### INVEST

- **Independent** — yes. The rule and the label stand without S2: a member can
  see the pairs before any figure changes.
- **Negotiable** — yes. It asks that transfers are recognised and shown, not
  how they are matched.
- **Valuable** — yes. Today nothing tells a transfer from a payment, in the
  ledger or anywhere else.
- **Estimable** — yes. A pure function over rows already read, and one label in
  a slot that exists.
- **Small** — yes.
- **Testable** — yes, against a ledger with known pairs, near misses and ties.

### Capabilities

- `banking/own-transfers`
- `banking/transactions`
- `banking/overview`

### Satisfied by

- `banking/own-transfers`: Requirement: What wimm calls a transfer between a
  member's own accounts
- `banking/transactions`: Requirement: A transfer between own accounts is
  labelled in the ledger
- `banking/overview`: Requirement: A transfer between own accounts is labelled
  where Overview lists it

## S2 · Read months that are not inflated by my own transfers

**As a** household member whose balance grows most months
**I want** the money I move between my own accounts left out of money in, money
out, the typical month and the average month
**so that** the months say what I spent and what I was paid, and a growing
balance does not read as more going out than coming in

### INVEST

- **Independent** — no. It needs S1's rule. Accepted: the figures are the point
  of the change, and S1 alone ships a label nobody asked for by itself.
- **Negotiable** — yes. It states which figures must stop counting transfers,
  not where in the service that happens.
- **Valuable** — yes. It is the difference between a negative typical month and
  a true one.
- **Estimable** — yes, once the read-only run in tasks.md has shown how many
  pairs the household's ledger holds.
- **Small** — no. It touches the month summary, the months, top spending, and
  both detectors. Accepted: leaving transfers out of some of them would put two
  different Septembers on one screen, which is the fault `spending-insights`
  already refused once for the scope.
- **Testable** — yes. A month with a known transfer has known figures with and
  without it, under each scope.

### Capabilities

- `banking/own-transfers`
- `banking/overview`

### Satisfied by

- `banking/own-transfers`: Requirement: A transfer inside the scope is left out
  of what is counted
- `banking/overview`: Requirement: The month so far, against the same days of
  last month
- `banking/overview`: Requirement: Month by month, as far back as the ledger is
  whole
- `banking/overview`: Requirement: A month says how many transfers it left out

## S3 · See the household's months the same way my partner does

**As a** household member who shares a joint account with a partner and pays
into it from an account of my own
**I want** what I pay into the joint account to count as money coming in to the
household, for both of us
**so that** we read the same Household figures, and my paying in is not erased
for me and counted for them

### INVEST

- **Independent** — no. It is S2's rule seen from the scope. Accepted: it is
  written separately because it is the case most likely to be got wrong, and it
  has a different person checking it.
- **Negotiable** — yes.
- **Valuable** — yes. Two owners reading different Household figures would make
  the scope useless for the conversation it exists for.
- **Estimable** — yes.
- **Small** — yes.
- **Testable** — yes. Both owners' Household months over one ledger are
  compared figure by figure.

### Capabilities

- `banking/own-transfers`

### Satisfied by

- `banking/own-transfers`: Requirement: A transfer inside the scope is left out
  of what is counted
