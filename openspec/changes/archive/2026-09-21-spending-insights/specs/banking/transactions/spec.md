## ADDED Requirements

### Requirement: An unusual payment is marked in the ledger
**Story**: S4
Transactions SHALL mark a booked transaction that `banking/payment-patterns`
calls unusual with the word Unusual, or Unusual income where the money came in,
by the same rule Overview uses, judged across
every account the member owns. The mark SHALL be words, never colour alone. It
changes nothing else about the row: its place, its amount and its name are as
they would be without it. Transactions MUST NOT offer to hide unusual payments.

A payment older than the months Overview shows SHALL carry no mark, because
wimm does not judge what it has no surrounding months for.

#### Scenario: Finding the repair in the ledger
- **WHEN** a member pages to March, where a €1,650.00 garage bill was unusual
- **THEN** that row carries the word Unusual, and the rows around it do not

#### Scenario: A bonus in the ledger
- **WHEN** a member pages to May, where a €9,804.00 bonus was unusual income
- **THEN** that row carries the words Unusual income and a plus sign

#### Scenario: Arriving from Overview
- **WHEN** a member follows an unusual payment from a month on Overview
- **THEN** they land on a page of Transactions with that payment on it, marked
  Unusual

#### Scenario: Overview and Transactions agree
- **WHEN** Overview counts two unusual payments in March under All
- **THEN** exactly those two rows are marked in March's transactions

#### Scenario: A payment not yet settled
- **WHEN** a large payment is still pending at the bank
- **THEN** it is marked as not settled and not as unusual

#### Scenario: Two years ago
- **WHEN** a member pages back past the months Overview shows
- **THEN** no row there is marked Unusual
