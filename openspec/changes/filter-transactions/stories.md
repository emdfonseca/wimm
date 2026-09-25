## S1 · Read one of my accounts without typing an address

**As a** household member reading their ledger across several accounts they own
**I want** to choose one account on the Transactions screen, or all of them, and keep paging through what I chose
**so that** I can read one account's statement without the others' rows in between, which today only works if I know to type `?account=` into the address bar

### INVEST

- **Independent** — Yes. The picker drives a narrowing the ledger already supports.
- **Negotiable** — Yes. It asks for a way to choose an account on screen, not a control shape.
- **Valuable** — Yes. Narrowing exists and nobody can reach it.
- **Estimable** — Yes. The read already takes an account; the screen gains one control.
- **Small** — Partly. It also carries the rule that every filter stays in the address and survives paging, which S2 to S4 lean on. Accepted: the account is the first filter built, so that rule lands with it.
- **Testable** — Yes. A member chooses an account, pages, and sees only that account's rows with it still named.

### Capabilities

`banking/transactions`

### Satisfied by

- `banking/transactions`: Requirement: The transactions of one account
- `banking/transactions`: Requirement: Filters combine and stay in the address

## S2 · Find a payment by who it was with

**As a** household member reading their ledger who remembers a payment by the shop or the person, not the date
**I want** to type part of a name or the bank's line and see only the transactions that match
**so that** I can answer "when did we last pay the plumber, and how much" without reading a year of rows

### INVEST

- **Independent** — Yes. Search works with or without an account chosen.
- **Negotiable** — Partly. Matching ignores case and not accents. Accepted: ignoring accents needs a Postgres extension and an ADR, and a household typing `continente` still finds `CONTINENTE`.
- **Valuable** — Yes. Finding one payment is the most common reason to open the ledger.
- **Estimable** — Yes. A text match over rows the member owns, no index.
- **Small** — Yes, with its empty state, since a search is the filter most likely to match nothing.
- **Testable** — Yes. A search for a merchant lists only its rows, and a search for nothing that exists says so.

### Capabilities

`banking/transactions`

### Satisfied by

- `banking/transactions`: Requirement: Searching the ledger
- `banking/transactions`: Requirement: When nothing matches the filters

## S3 · See one month's transactions

**As a** household member reading their ledger who has just looked at a month on Overview
**I want** choosing a month to show that month's transactions and nothing else
**so that** the rows I read are the ones behind the month's money in and money out, and I know where the month ends

### INVEST

- **Independent** — Yes. The month control exists; this changes what choosing one does.
- **Negotiable** — No. It replaces a rule the spec states, that a month is a place to start reading. Accepted: the member asked for a filter, and Overview already links to a month expecting one.
- **Valuable** — Yes. Overview's month links land on the newest page today, which is not the month they name.
- **Estimable** — Yes.
- **Small** — Yes.
- **Testable** — Yes. Following August from Overview lists August's rows only, and the last page of August says nothing older matches.

### Capabilities

`banking/transactions`

### Satisfied by

- `banking/transactions`: Requirement: Going straight to a month

## S5 · See what the list I narrowed adds up to

**As a** household member reading their ledger who has searched for a merchant or opened a month
**I want** to see how much came in and how much went out across everything the filters match
**so that** "how much did we spend at Galp this year" is one search, not a sum done by hand over pages of rows

### INVEST

- **Independent** — Partly. It sums whatever S1 to S4 narrow to. Accepted: it adds nothing to a list that is not filtered.
- **Negotiable** — Yes. It asks for the figures, not a layout.
- **Valuable** — Yes. Asked for by the member while searching.
- **Estimable** — Yes. The month summary already sums settled rows and leaves out transfers between the member's accounts; this is the same rule over the filtered rows.
- **Small** — Yes.
- **Testable** — Yes. Opening August from Overview shows the money in and money out Overview gives August.

### Capabilities

`banking/transactions`

### Satisfied by

- `banking/transactions`: Requirement: What a filtered list adds up to

## S4 · Read only money in, or only money out

**As a** household member reading their ledger who wants to check what came in, or what went out
**I want** to show only money in or only money out
**so that** checking that every salary and refund arrived does not mean reading past every card payment

### INVEST

- **Independent** — Yes.
- **Negotiable** — Yes. It asks for a direction, not a control.
- **Valuable** — Yes, and least of the four. The member ranked it last.
- **Estimable** — Yes. The sign of the amount.
- **Small** — Yes.
- **Testable** — Yes. Money in lists no row with a minus sign.

### Capabilities

`banking/transactions`

### Satisfied by

- `banking/transactions`: Requirement: Money in or money out
