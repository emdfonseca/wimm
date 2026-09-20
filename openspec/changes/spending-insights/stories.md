## S1 · See whether more goes out than comes in

**As a** household member who owns the accounts the salary arrives in and the
bills leave from
**I want** to see money in, money out and net for each of the past months, with
what a typical month and an average month look like
**so that** I know whether we are living within what comes in, rather than
judging it from the one month on screen

### INVEST

- **Independent** — yes. It reads the stored ledger and needs nothing else in
  this change. S4 changes how it can be read, not whether it works.
- **Negotiable** — yes. It asks for the months and the two figures, not for
  bars or a list.
- **Valuable** — yes. Today the only history is the balance chart, which shows
  where the balance went and not whether a month paid for itself.
- **Estimable** — yes. One grouped sum over rows already stored, and a median.
- **Small** — yes.
- **Testable** — yes, against a ledger with known rows in known months,
  including a ledger that starts part-way through the window.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: Month by month, as far back as the ledger
  is whole
- `banking/overview`: Requirement: A typical month and an average month

## S2 · See which merchants are behind a rise

**As a** household member who has just seen that a month cost more than usual
**I want** to be told which merchants took more than they usually do that month
**so that** I know where to look for something to cut, without opening
Transactions and adding it up

### INVEST

- **Independent** — no. It hangs off S1's months. Accepted: a list of merchants
  that rose has no meaning without the month it rose in.
- **Negotiable** — yes. It states the comparison, not how many merchants or
  where they sit.
- **Valuable** — yes. It is the step from "we spent more" to "on what".
- **Estimable** — yes. Per-month totals on the merchant name that already
  exists, against each merchant's own median month.
- **Small** — yes.
- **Testable** — yes.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: The merchants behind a month's rise

## S3 · See what leaves every month without being asked

**As a** household member who pays the subscriptions and standing bills
**I want** a list of the payments that repeat, with how much, how often and
when the next one is due, including a yearly one wimm has only seen twice
**so that** I can see what I am committed to before the month starts, and spot
one I forgot I was still paying for

### INVEST

- **Independent** — yes.
- **Negotiable** — yes. It asks for the repeating payments, not for a rule. The
  rule is `banking/payment-patterns`'s and can be tuned without touching this.
- **Valuable** — yes. A forgotten subscription is the cheapest saving there is,
  and nothing in wimm surfaces one today.
- **Estimable** — yes. The detector is a pure function over rows and is written
  test-first.
- **Small** — yes.
- **Testable** — yes. Each cadence, the amount tolerance, a payment that
  stopped, one that is late and a yearly one seen twice are each a scenario.

### Capabilities

- `banking/payment-patterns`
- `banking/overview`

### Satisfied by

- `banking/payment-patterns`: Requirement: What wimm calls a recurring payment
- `banking/overview`: Requirement: Recurring payments are listed with when the
  next is expected

## S4 · Keep one large payment from bending the picture

**As a** household member reading the months after a car repair landed in one
of them
**I want** the payments that are far outside what is usual, going out or coming
in, to be marked, counted against their month, and optionally set aside from
the typical and average month
**so that** I can tell a month that went wrong from a month that held one
exceptional bill, and a good month from one that held a bonus, without losing
sight of either

### INVEST

- **Independent** — no. The view without unusual payments reads S1's figures.
  The marking itself stands alone and ships in Transactions regardless.
- **Negotiable** — yes. It states that the payment is marked and never hidden.
  The rule is one requirement in `banking/payment-patterns`.
- **Valuable** — yes. An average over twelve months with one repair in it says
  the household overspends when it does not.
- **Estimable** — yes. One published rule, written test-first.
- **Small** — no. It touches the months, the ledger and the balance chart.
  Accepted: one rule marked in one place and not another is the two-answers
  defect this change is trying not to ship.
- **Testable** — yes. The rule's cases are table tests, and what the member
  sees is a scenario per surface.

### Capabilities

- `banking/payment-patterns`
- `banking/overview`
- `banking/transactions`

### Satisfied by

- `banking/payment-patterns`: Requirement: What wimm calls an unusual payment
- `banking/overview`: Requirement: A month says how many unusual payments it
  held, and opens to list them
- `banking/overview`: Requirement: The months can be read without unusual
  payments
- `banking/transactions`: Requirement: An unusual payment is marked in the
  ledger

## S5 · Look at household money and my own separately

**As a** household member who owns a joint account with a partner and also has
accounts of their own
**I want** to switch everything Overview draws from transactions between the
household's accounts, my own, and both
**so that** I can answer "are we overspending" and "am I overspending" as two
questions, the way the two money figures already separate them

### INVEST

- **Independent** — yes. It applies to the month summary, top spending and the
  balance chart that exist before this change, as well as to S1 to S4 and S6.
- **Negotiable** — yes, within one fixed line: every scope is accounts the
  member owns (ADR 0021).
- **Valuable** — yes. Rent from a joint account and a member's own hobby
  spending are one figure today.
- **Estimable** — yes. The groups exist (ADR 0024); the scope is a filter on
  the account set each query already takes.
- **Small** — yes.
- **Testable** — yes. Each scope, the hidden options, a household of one and
  the narrower-than-the-figure case are scenarios.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: One scope for everything drawn from
  transactions

## S6 · Ask the balance chart what happened on a day

**As a** household member who has spotted a drop on the balance chart
**I want** to point at that day, with a mouse, a finger or the keyboard, and be
told the balance, the change from the day before and the transactions that
explain it
**so that** I do not have to note the date, open Transactions and find the day
myself

### INVEST

- **Independent** — yes. The unusual label on a row comes from S4's rule and is
  simply absent without it.
- **Negotiable** — yes. It asks for what explains the day, not for a fixed
  number of rows.
- **Valuable** — yes. The chart raises the question and today cannot answer it.
- **Estimable** — yes. The rows are already in hand where the chart is built.
- **Small** — yes.
- **Testable** — yes. The pointer, the tap and the keys are each asserted, and
  so is the announcement.

### Capabilities

- `banking/overview`

### Satisfied by

- `banking/overview`: Requirement: Pointing at a day says what moved it
