# banking/own-transfers Specification

## Purpose
States once what wimm calls a transfer between a member's own accounts, and
when such a transfer is left out of what is counted, so that the month summary,
the months, top spending, the payment patterns, the balance chart and
Transactions treat the same rows the same way and a member can check any of
them against the published rule.

## Requirements

### Requirement: What wimm calls a transfer between a member's own accounts
**Story**: S1
wimm SHALL call two booked transactions a transfer between a member's own
accounts when one goes out of an account the member owns and the other comes in
to a different account the member owns, in the same currency, for exactly the
same amount, with booking dates no more than a stated number of days apart. It
SHALL be recognised only from booked transactions of accounts the member owns;
an account the member merely holds a grant on is never either side.

Each transaction SHALL be one half of at most one transfer. Where a transaction
could pair with more than one other, the one closest in date SHALL be chosen;
where two are equally close, one whose text names the other account, by its
name or by the end of its number, SHALL be chosen over one that does not; and
where they still cannot be told apart, none of them SHALL be paired. A pairing
wimm is not sure of is worse than none, because it removes a real payment from
the figures.

A transaction wimm cannot pair SHALL stay an ordinary payment or an ordinary
amount coming in, however much it looks like a transfer: money sent to an
account wimm does not hold has left the accounts wimm can see. A transaction
not yet settled by the bank is never half of a transfer.

What is a transfer SHALL be worked out for the member who is looking, from the
accounts that member owns. It MUST NOT tell a member anything about an account
they do not own: a row coming in to a joint account from a partner's own
account is, for the member who does not own that account, an ordinary amount
coming in.

#### Scenario: Moving money to savings
- **WHEN** €500.00 leaves a member's current account on 3 June and €500.00
  arrives in their savings account on 4 June
- **THEN** the two rows are one transfer between their own accounts

#### Scenario: Both sides on the same day
- **WHEN** the money leaves and arrives on the same day
- **THEN** the two rows are one transfer

#### Scenario: Too far apart
- **WHEN** €500.00 leaves one of a member's accounts and €500.00 arrives in
  another eleven days later
- **THEN** they are two ordinary transactions

#### Scenario: Not the same amount
- **WHEN** €500.00 leaves one account and €499.00 arrives in another the next
  day
- **THEN** they are two ordinary transactions

#### Scenario: Two currencies
- **WHEN** €500.00 leaves one account and £500.00 arrives in another
- **THEN** they are two ordinary transactions

#### Scenario: The same amount twice in a week
- **WHEN** €200.00 leaves a current account on the 1st and again on the 3rd,
  and €200.00 arrives in savings on the 1st and again on the 3rd
- **THEN** they are two transfers, each row paired with the one on its own day

#### Scenario: One arrival, two possible sources
- **WHEN** €200.00 leaves two different accounts of the member on the same day
  and €200.00 arrives in a third that day, and the arriving row names neither
- **THEN** none of the three is called a transfer

#### Scenario: The text names the other account
- **WHEN** the same happens and the arriving row's text ends with the last
  digits of one of the two accounts
- **THEN** the arriving row is paired with the row from that account, and the
  other stays an ordinary payment

#### Scenario: Sent to an account wimm does not hold
- **WHEN** a member sends €1,000.00 to a savings account at a bank they have
  not connected
- **THEN** it is an ordinary payment, because no row arrives anywhere wimm can
  see

#### Scenario: A refund that happens to match
- **WHEN** a member pays a shop €50.00 from one account and the shop refunds
  €50.00 to the same account two days later
- **THEN** they are two ordinary transactions, because a transfer is between
  two different accounts

#### Scenario: Not yet settled
- **WHEN** the money has left one account and the arriving row is still pending
- **THEN** neither is called a transfer until both are booked

#### Scenario: A partner's own account
- **WHEN** a partner moves €800.00 from an account only they own to the joint
  account both own
- **THEN** the partner sees a transfer between their own accounts, and the
  other owner sees €800.00 coming in to the joint account and nothing more

#### Scenario: An account held only by a grant
- **WHEN** a member holds details on an account they do not own, and money
  moves from it to an account they do own
- **THEN** the member sees an ordinary amount coming in

### Requirement: A transfer inside the scope is left out of what is counted
**Story**: S2
Everything wimm draws from transactions SHALL leave out a transfer between a
member's own accounts when both of its rows are on accounts inside the scope in
force (`banking/overview`): money in, money out and net in the month summary
and in every month, the typical and the average month, the merchants behind a
month's rise, top merchants and largest payments. Such a row SHALL never be
called an unusual payment or a recurring payment, and SHALL not count towards
what is usual for anything else.

A transfer with one row inside the scope and one outside it SHALL be counted as
what it is for that scope: money that left the accounts counted, or money that
arrived in them. So under Yours, money sent to a household account is money
out; under Household, the same money is money in; under All, it is neither.
Every owner of a household account SHALL therefore read the same Household
figures, whichever of them paid in.

Leaving a transfer out SHALL change no balance, no account total and no row in
any list: the balance chart's line, Household money, Your money, the accounts
list, recent transactions and Transactions show what they showed. It is a way
of counting, and nothing is hidden by it.

#### Scenario: A month with a transfer to savings
- **WHEN** a member whose salary was €2,450.00 and whose payments came to
  €1,900.00 also moved €500.00 to their own savings account, and reads that
  month under All
- **THEN** the month reads €2,450.00 in, €1,900.00 out and +€550.00, not
  €2,950.00 in and €2,400.00 out

#### Scenario: A growing balance no longer reads as overspending
- **WHEN** a member who moves most of each salary to a savings account wimm
  holds reads their typical month under Yours
- **THEN** the typical month is what came in less what they paid, and is above
  zero

#### Scenario: Paying into the joint account, under Yours
- **WHEN** a member who sends €800.00 a month from their own account to the
  joint account reads the month under Yours
- **THEN** the €800.00 is money out, because it left the accounts Yours counts

#### Scenario: Paying into the joint account, under Household
- **WHEN** that member reads the month under Household
- **THEN** the €800.00 is money in

#### Scenario: Paying into the joint account, under All
- **WHEN** that member reads the month under All
- **THEN** the €800.00 is in neither figure

#### Scenario: Both owners read the same household month
- **WHEN** two members own a joint account and one of them pays €800.00 into
  it from an account only they own
- **THEN** both read the same money in, money out and net under Household

#### Scenario: A large transfer is not an unusual payment
- **WHEN** a member moves €6,000.00 between two of their own accounts, far
  above anything they usually pay
- **THEN** neither row is called unusual, in any scope that holds both

#### Scenario: A standing transfer is not a recurring payment
- **WHEN** a member moves €300.00 to their own savings account on the 1st of
  every month, and reads Overview under All
- **THEN** it is not listed among recurring payments

#### Scenario: A standing transfer that crosses the scope
- **WHEN** a member sends €800.00 to the joint account on the 1st of every
  month, and reads Overview under Yours
- **THEN** it is listed among recurring payments, because under Yours it is
  money that leaves every month

#### Scenario: Not among the merchants
- **WHEN** a month's largest single movement was a transfer between two
  accounts inside the scope
- **THEN** it is in neither top merchants nor largest payments

#### Scenario: The balance does not move
- **WHEN** a member reads the balance chart for a month that held a transfer
- **THEN** the line is where it was before transfers were recognised

#### Scenario: One that could not be paired
- **WHEN** a member sends €1,000.00 to an account wimm does not hold
- **THEN** it is money out in every scope that holds the account it left
