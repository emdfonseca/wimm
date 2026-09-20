## Purpose

Covers the screen a member lands on after signing in: the total they already
have, a recent slice of their own transactions, and how their balance has
been trending — each scoped to what that member may see, and never account
management.

## ADDED Requirements

### Requirement: Overview shows the total and nothing about accounts or connections
**Story**: S1
Overview SHALL show the household total exactly as `banking/household-accounts`
defines it. Overview SHALL NOT list individual accounts, name a bank, or offer
any connection action — connecting, restoring, disconnecting, or choosing who
sees an account. Where a member wants any of that, Overview SHALL offer a way
to reach the Accounts screen.

#### Scenario: The total is there, the accounts are not
- **WHEN** a member with two connected banks opens Overview
- **THEN** they see their total, and no per-bank card, account row, or
  connection control anywhere on the screen

#### Scenario: Getting to account management from Overview
- **WHEN** a member on Overview wants to connect, restore, or disconnect a
  bank
- **THEN** Overview offers a way to reach the Accounts screen, and performs
  none of those actions itself

### Requirement: Overview shows a recent slice of the member's own transactions
**Story**: S1
Overview SHALL show the member's most recent transactions, drawn only from
accounts they own — transactions stay owner-only, per
`banking/transactions`. A member who owns no account, or whose owned accounts
have no transactions read yet, SHALL see no recent-transactions section at
all, rather than one stating there is nothing, because a section that could
state either emptiness or absence-of-access would tell an owner-less member
something they are not owed.

Each transaction SHALL show enough to identify it without opening
Transactions: its date, its counterparty or description, its amount, and
which account it belongs to. Overview SHALL offer a way to reach the full
ledger.

#### Scenario: Recent activity at a glance
- **WHEN** a member who owns at least one account with transactions opens
  Overview
- **THEN** they see their most recent transactions, each showing its date,
  description, amount, and account

#### Scenario: A member who owns nothing sees no section
- **WHEN** a member owns no account, however many they have been granted
  *balance* or *details* on
- **THEN** Overview shows no recent-transactions section, rather than one
  saying there are no transactions

#### Scenario: An owned account with no transactions yet
- **WHEN** a member owns an account whose transactions have not been read yet
- **THEN** that account contributes nothing to the recent-transactions
  section, and the section is absent if no owned account has any

#### Scenario: Reaching the full ledger
- **WHEN** a member looks at the recent-transactions section
- **THEN** they can open Transactions to see more than what is shown here

### Requirement: The trend follows the ledger, and only as far as the ledger reaches
**Story**: S1
Overview SHALL show a balance trend, per currency, derived from the member's
owned accounts' booked transactions — the same rows `banking/transactions`
already stores, never a separate reading kept for this purpose. The trend
SHALL cover only the span each contributing account's ledger actually
reaches back to — its own earliest booked transaction, never
`transactions_synced_through`: that column is a forward completeness
watermark, and an exhausted sync sets it to today however far back the
history it exhausted actually reaches. wimm MUST NOT draw a trend line
across a gap it has no transactions for, and MUST NOT imply a longer history
than it holds.

An account whose connection has never been granted transaction access SHALL
contribute to the total shown on Overview but SHALL NOT contribute to the
trend, and Overview SHALL say plainly that the trend does not cover every
account rather than drawing a line as if it did.

Where a member's accounts are held in more than one currency, wimm SHALL show
one trend per currency and MUST NOT convert or combine them, matching how the
total is shown.

#### Scenario: A trend from what has been read
- **WHEN** a member owns an account whose transactions have synced 90 days
  back
- **THEN** the trend covers that span, and no further

#### Scenario: An account with balances only
- **WHEN** a member owns one account with transaction access and one with
  balance-only access
- **THEN** both contribute to the total, only the first contributes to the
  trend, and wimm says the trend does not cover every account

#### Scenario: Two currencies, two trends
- **WHEN** a member's owned accounts are held in two currencies
- **THEN** a separate trend is shown for each currency, and none is combined
  or converted into the other

#### Scenario: No transaction history anywhere
- **WHEN** none of a member's owned accounts has any transaction access
- **THEN** no trend is shown, rather than a flat or invented one

### Requirement: Before any bank is connected, Overview says so
**Story**: S1
A household that has connected no bank, or a member who may see no account,
SHALL be shown on Overview what the screen is for and a way to reach Accounts
to connect one — not a total of zero, an empty recent-transactions section,
or an empty trend.

#### Scenario: The first member arrives
- **WHEN** a member of a household with no connected banks signs in
- **THEN** Overview explains what connecting a bank does and offers a way to
  reach Accounts, showing no total, no recent-transactions section, and no
  trend

#### Scenario: A member who may see no account
- **WHEN** a bank is connected but a member has been granted nothing on any
  of its accounts
- **THEN** they see Overview's before-any-bank explanation, not a total of
  zero
