## S1 · See how the month is going

**As a** household member who owns the accounts the bills leave from
**I want** to see what came in, what went out and where the most went this
month, against the same point last month
**so that** I know whether we are spending more than usual while there is
still month left to do something about it

### INVEST

- **Independent** — yes. It reads the stored ledger and needs nothing else in
  this change, though it reads better with S3's names.
- **Negotiable** — yes. It states the comparison, not the tiles or the
  ranking's length.
- **Valuable** — yes. Today the only way to answer it is to add up
  Transactions by hand.
- **Estimable** — yes. Two date windows and two groupings over rows already
  stored.
- **Small** — no. It carries both the summary and top spending. They are one
  question asked twice, "how much" and "on what", and either alone leaves the
  member going to Transactions for the other half.
- **Testable** — yes, against a ledger with known rows in both months.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: The month so far, against the same days of
  last month
- `banking/overview`: Requirement: Where the most money went this month

## S2 · See where the balance has been

**As a** household member who owns at least one account with history
**I want** to see my balance over the past months as a line I can read dates
and amounts off
**so that** I can tell a normal dip before payday from a balance that is
actually falling

### INVEST

- **Independent** — yes.
- **Negotiable** — yes. It asks for a readable history, not a chart type.
- **Valuable** — yes. The bars that exist show neither when nor how much, and
  for a one-day ledger they show nothing true at all.
- **Estimable** — yes. The backward walk exists; the points and the drawing
  change.
- **Small** — yes.
- **Testable** — yes. The short-history, no-history and empty-currency cases
  are each a scenario.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: The balance chart follows the ledger, and
  only as far as the ledger reaches

## S3 · Recognise a payment by who it was with

**As a** household member reading a list of my own payments
**I want** each one named by the merchant rather than by the bank's
statement line
**so that** I can recognise a payment at a glance and see several payments to
one merchant as the same merchant

### INVEST

- **Independent** — yes. It improves both lists on its own.
- **Negotiable** — yes. Which noise is removed is a rule set that can grow.
- **Valuable** — yes. `COMPRA ... 230002268264350` is the bank's bookkeeping,
  not a name.
- **Estimable** — yes, as a fixed first rule set. The long tail of bank
  formats is not estimable and is not promised.
- **Small** — yes.
- **Testable** — yes, as a table of raw lines and expected names.

### Capabilities

- `banking/transactions`
- `banking/overview`

### Satisfied by

- `banking/transactions`: Requirement: A transaction is named by who it was
  with
- `banking/overview`: Requirement: Overview shows a recent slice of the
  member's own transactions

## S4 · See every account's balance from the landing screen

**As a** household member with more than one account
**I want** the landing screen to show each account I may see beside the
total, using the whole screen and leaving out what holds nothing
**so that** I can see which account the total is sitting in without opening
Accounts, and am not shown a row of zeros for a currency I do not use

### INVEST

- **Independent** — yes.
- **Negotiable** — yes.
- **Valuable** — yes. A total of 5,588 across a current account and a credit
  card means different things depending on the split.
- **Estimable** — yes. The accounts and their balances are already loaded for
  the total.
- **Small** — yes.
- **Testable** — yes.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: Overview lists the accounts behind the
  figures
- `banking/overview`: Requirement: Before any bank is connected, Overview
  says so

## S5 · Tell our money from mine

**As a** household member who shares a joint account and keeps a personal one
**I want** what the household holds together and what is mine shown as two
figures
**so that** I can answer "how much do we have" and "how much do I have"
without subtracting one account from a total in my head

### INVEST

- **Independent** — yes. It needs only the accounts and levels already stored.
- **Negotiable** — partly. The two figures are the need. Which accounts count
  as the household's is a rule, and it was settled with the person asking:
  everyone owns it or sees it in full.
- **Valuable** — yes. One figure mixing a partner's shared savings into "my"
  total answers neither question.
- **Estimable** — yes. A classification over rows already loaded.
- **Small** — yes.
- **Testable** — yes. Every case is a scenario, including the household of
  three and the member who joins.

### Capabilities

- `banking/household-accounts`
- `banking/overview`

### Satisfied by

- `banking/household-accounts`: Requirement: Household money and a member's
  own money are separate figures
- `banking/overview`: Requirement: Overview shows the total and performs no
  connection action
