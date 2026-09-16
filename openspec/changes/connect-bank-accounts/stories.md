## S1 · Connect my bank, and share only what I mean to

**As a** member of the household who banks somewhere wimm does not yet know about
**I want** to hand wimm access to my bank by confirming it at the bank itself,
and then choose which of those accounts the rest of the household sees
**so that** the joint account arrives complete and correct without me copying it
off a statement, and my personal savings does not arrive with it

### INVEST

- **Independent** — **No.** S1 produces accounts and S2 is what shows them;
  shipping S1 alone lands a member on a confirmation with nowhere to go.
  Accepted, because the other split is worse: a screen that lists accounts
  before anything can create one is a screen with only an empty state.
- **Negotiable** — Yes. It says the member confirms at their bank and then
  chooses what to share. Which gateway carries that, and what the picker and
  the chooser look like, are open.
- **Valuable** — Yes, and the second half is not decoration. Some banks let a
  member narrow accounts in their own consent screen and some hand over
  everything, so without the chooser, connecting a bank to share one account
  shares every account at that bank to everyone in the house.
- **Estimable** — Yes, with one unknown priced in: the gateway's sandbox has to
  be reachable before the consent round-trip can be exercised end to end.
- **Small** — **No.** It is a picker, a hand-off out of the product, a return,
  a chooser and a failure surface. Kept whole anyway: the round-trip is the
  story, and no half of it is usable. Splitting the chooser off would ship the
  exposure this story exists to prevent and then close it in a later change.
- **Testable** — Yes, against the gateway's sandbox banks and against the
  in-memory fake, which is what keeps `just check` off the network.

### Capabilities

- `banking/bank-connections`

### Satisfied by

- `banking/bank-connections`: Requirement: Choosing a bank to connect
- `banking/bank-connections`: Requirement: Consenting at the bank
- `banking/bank-connections`: Requirement: Returning from the bank
- `banking/bank-connections`: Requirement: Choosing which accounts the household sees
- `banking/bank-connections`: Requirement: A connection that cannot be completed
- `banking/bank-connections`: Requirement: What reaches a bank is unreadable at rest

---

## S2 · See what the household has, in one place

**As a** member of a household whose money sits in several banks
**I want** every shared account, with a balance I can tell the age of, and a
total, on the screen I land on
**so that** I can answer "where is my money" in one look instead of signing in
to three banks

### INVEST

- **Independent** — **No.** It has nothing to show until S1 has run. It is the
  dependent half of the pair described in S1.
- **Negotiable** — Yes. It states that the household sees its shared accounts
  and a total, not how they are grouped, sorted or laid out.
- **Valuable** — Yes, and this is where the change's value lands. S1 is plumbing
  a member endures once; S2 is what they come back for.
- **Estimable** — Yes.
- **Small** — Yes. One screen, one list, one total.
- **Testable** — Yes. A member signed in to a household with shared accounts
  sees them; a member of a household with none sees the empty state and the way
  to fix it.

### Capabilities

- `banking/household-accounts`

### Satisfied by

- `banking/household-accounts`: Requirement: Shared accounts belong to the household
- `banking/household-accounts`: Requirement: Seeing the household's accounts
- `banking/household-accounts`: Requirement: A balance is a reading, not a live figure
- `banking/household-accounts`: Requirement: Bringing balances up to date
- `banking/household-accounts`: Requirement: The household total
- `banking/household-accounts`: Requirement: Before any bank is connected

---

## S3 · Take a bank back out

**As a** member who connected a bank they no longer want wimm to read — the
wrong one, an account they have closed, or simply a decision reversed
**I want** to disconnect it and see its accounts leave
**so that** granting wimm access to my bank is a reversible decision rather than
a one-way door

### INVEST

- **Independent** — Yes, given a connection exists. It can ship after S1 and S2
  without changing either.
- **Negotiable** — Yes. It says access ends and the accounts go; whether the
  gateway is told, and what happens to anything already fetched, are open.
- **Valuable** — Yes. It is also the honest half of asking for bank access at
  all: a member weighing the decision needs to know it is reversible.
- **Estimable** — Yes.
- **Small** — Yes.
- **Testable** — Yes. Disconnect, and the accounts are gone from Overview and
  from the household total.

### Capabilities

- `banking/bank-connections`
- `banking/household-accounts`

### Satisfied by

- `banking/bank-connections`: Requirement: Disconnecting a bank
- `banking/household-accounts`: Requirement: Accounts leave with their connection

---

## S4 · Get a bank working again when its access runs out

**As a** member whose bank connection has reached the end of the access the bank
granted
**I want** to be told plainly that it has stopped updating, and to restore it
without disconnecting and starting over
**so that** a ledger I have been relying on for six months does not quietly
become a page of stale numbers I have no way to fix

### INVEST

- **Independent** — No. It needs a connection to have expired, so it follows S1.
- **Negotiable** — Yes. It says access is restored and what was shared stays
  shared; whether the member re-picks the bank, and where the prompt sits, are
  open.
- **Valuable** — Yes, and this is not a rare path. Most banks cap consent at
  180 days, so every connection reaches this state. Open banking has no refresh
  mechanism: renewal is the whole authorisation flow again, which S1 already
  builds.
- **Estimable** — Yes. It re-enters the flow S1 draws, from a different entry
  point.
- **Small** — Yes, given S1. It is an entry point, a prompt and the rule that
  previously shared accounts stay shared.
- **Testable** — Yes. Expire a connection, restore it, and confirm the same
  accounts are shared and reading again.

### Capabilities

- `banking/bank-connections`
- `banking/household-accounts`

### Satisfied by

- `banking/bank-connections`: Requirement: Access that has run out
- `banking/bank-connections`: Requirement: Restoring access to a bank
