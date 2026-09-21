## MODIFIED Requirements

### Requirement: The month so far, against the same days of last month
**Story**: S2
Overview SHALL show, per currency, the money that came in, the money that
went out, and the difference between them, for the current calendar month up
to today. It SHALL be drawn only from the booked transactions of accounts the
member owns. Currencies MUST NOT be combined or converted.

Each figure SHALL be set against the same figure for the same days of the
previous month — the 1st to the 20th against the 1st to the 20th — because a
part-month set against a whole month always looks like an improvement. wimm
SHALL show that comparison only where every contributing account's ledger
reaches back to the start of the previous month, and SHALL otherwise show
this month's figures alone rather than a comparison against a month it only
partly holds.

Where a contributing account's ledger begins after the 1st of the current
month, the summary SHALL say which date it counts from. A member who owns no
account, or whose owned accounts have no booked transaction this month, SHALL
see no month summary at all.

Money moved between the member's own accounts SHALL be left out of all three
figures, in this month and in the month it is set against, as
`banking/own-transfers` defines. A transfer whose other half falls just outside
the days counted SHALL still be left out. The summary SHALL say that such
transfers are left out where wimm holds both accounts, and that others are
counted, since a member who sends money to an account wimm does not hold will
otherwise look for it.

#### Scenario: A month under way
- **WHEN** a member whose ledger reaches back three months opens Overview on
  the 20th
- **THEN** they see money in, money out and the difference for the 1st to the
  20th, each beside the same figure for the 1st to the 20th of last month

#### Scenario: Spending is up
- **WHEN** more has gone out this month so far than in the same days of last
  month
- **THEN** the member is shown by how much, in words or a sign, never by
  colour alone

#### Scenario: Last month is not all there
- **WHEN** a contributing account's ledger begins part-way through last month
- **THEN** this month's figures are shown with no comparison beside them

#### Scenario: The ledger starts inside this month
- **WHEN** a member's only ledger begins on the 19th
- **THEN** the summary says it counts from the 19th, and is not presented as
  the whole month so far

#### Scenario: The 31st against a shorter month
- **WHEN** today is the 31st and last month had 30 days
- **THEN** this month so far is set against the whole of last month

#### Scenario: Two currencies
- **WHEN** a member has booked transactions this month in two currencies
- **THEN** each currency has its own summary, and no figure combines them

#### Scenario: Pending payments are not counted
- **WHEN** a payment has not been settled by the bank
- **THEN** it is in neither figure, and joins them once it is booked

#### Scenario: Transfers between own accounts
- **WHEN** a member who moved €500.00 to their own savings account this month
  reads the month summary
- **THEN** the €500.00 is in neither money in nor money out, and the summary
  tells them that money moved between their accounts is left out where wimm
  holds both, and that other transfers are counted

#### Scenario: A transfer across the end of the month
- **WHEN** money left one of a member's accounts on the 31st and arrived in
  another on the 1st
- **THEN** it is in neither month's figures

#### Scenario: A member who owns nothing
- **WHEN** a member owns no account, whatever they have been granted
- **THEN** Overview shows no month summary, rather than one of zeros

#### Scenario: Looking at the month in full
- **WHEN** a member wants to see what made up a figure
- **THEN** the summary offers a way to open Transactions at this month

### Requirement: Month by month, as far back as the ledger is whole
**Story**: S2
Overview SHALL show, per currency, money in, money out and net for the month so
far and for up to the 12 calendar months before it, newest first, drawn only
from the booked transactions of accounts the member owns, within the scope in
force. Currencies MUST NOT be combined or converted. The months SHALL be
readable as a trend at a glance and as figures, and the direction of a month's
net SHALL be carried by its sign or by words, never by colour alone.

A month SHALL be called full only when it has ended and the oldest contributing
ledger reaches back to its first day. A younger account beside it SHALL NOT
hold the months back: it simply adds nothing before its own ledger begins, and
Overview SHALL say which accounts the earlier months do not include and from
when. The month so far, and the month the oldest ledger begins inside, SHALL be
shown as what they are, the first as so far and the second with the date it is
held from, because a part-month read as a whole one always looks like a good
month. A month before every ledger begins SHALL not be shown at all.

Money moved between the member's own accounts SHALL be left out of every
month's figures where both accounts are inside the scope in force, as
`banking/own-transfers` defines, and the section SHALL say so in the words the
month summary uses. Under Household and under Yours it SHALL also say that
money moved to or from the member's other accounts is counted, because that is
the one case where a transfer the member can see labelled is still in the
figures.

A member who owns no account, or whose ledger holds no full month, SHALL see no
month-by-month section, rather than one bar standing alone.

#### Scenario: A year of months
- **WHEN** a member whose ledger reaches back two years opens Overview on 20
  September
- **THEN** they see September so far and the twelve months from September last
  year to August, each with its money in, money out and net

#### Scenario: A shorter ledger
- **WHEN** a member's ledger begins on 12 April
- **THEN** they see April marked as held from 12 April, May to August as full
  months, and September so far, and nothing before April

#### Scenario: A month that ended below zero
- **WHEN** more went out than came in during March
- **THEN** March's net reads with a minus sign, and its place in the trend is
  below the line the other months stand on

#### Scenario: Reading one month
- **WHEN** a member points at a month, by pointer, tap or keyboard
- **THEN** they are given that month's name, money in, money out and net as
  figures

#### Scenario: A second bank connected last month
- **WHEN** a member connects a bank whose ledger begins on 19 September, beside
  an account whose ledger reaches back to March
- **THEN** the months since April stay full, and Overview says the months
  before 19 September do not include that bank

#### Scenario: No full month yet
- **WHEN** a member's only ledger begins this month
- **THEN** Overview shows no month-by-month section

#### Scenario: A member who owns nothing
- **WHEN** a member owns no account, whatever they have been granted
- **THEN** Overview shows no month-by-month section

#### Scenario: Two currencies
- **WHEN** a member has full months in two currencies
- **THEN** each currency has its own months, and no figure combines them

#### Scenario: Pending payments are not counted
- **WHEN** a payment has not been settled by the bank
- **THEN** it is in no month's figures until it is booked

#### Scenario: Transfers are left out, and the section says so
- **WHEN** a member reads the months under All
- **THEN** the section says that money moved between their accounts is left out
  where wimm holds both

#### Scenario: Under a narrower scope
- **WHEN** a member reads the months under Yours
- **THEN** the section also says that money moved to or from their other
  accounts is counted

## ADDED Requirements

### Requirement: A month says how many transfers it left out
**Story**: S2
A month that left out one or more transfers between the member's own accounts
SHALL say so where its figures are read: how many transfers, and the money they
moved, so a member adding the month up from Transactions can account for the
difference. A month that left out none SHALL say nothing about transfers,
rather than report a count of zero. The count SHALL follow the scope in force,
and a transfer whose two rows fall in different months SHALL be counted in the
month its money left.

#### Scenario: A month with two transfers
- **WHEN** a member reads a month in which they twice moved money to their own
  savings account, €500.00 and €900.00
- **THEN** the month says 2 transfers between their accounts are left out,
  €1,400.00

#### Scenario: One transfer
- **WHEN** the month held one
- **THEN** it says 1 transfer, never 1 transfers

#### Scenario: A month with none
- **WHEN** a month held no such transfer
- **THEN** it says nothing about transfers

#### Scenario: Under a scope that holds only one side
- **WHEN** a member reads, under Yours, a month whose only transfer went to the
  joint account
- **THEN** the month says nothing about transfers left out, because that one
  was counted

### Requirement: A transfer between own accounts is labelled where Overview lists it
**Story**: S1
Wherever Overview lists a single transaction that is one half of a transfer
between the member's own accounts, as `banking/own-transfers` defines, it SHALL
label it as between their accounts, in words and never by colour alone: among
what moved a day on the balance chart, among largest payments where the scope
in force still counts it, and in recent transactions. The label
SHALL be shown whatever the scope in force, because it is a fact about the row
and not about what is being counted, and such a row SHALL never also be
labelled unusual.

A day on the balance chart SHALL still list a transfer among what moved it,
because where the two halves fall on different days, or one account is not part
of the chart, the line really did move.

#### Scenario: The day the money left
- **WHEN** a member points at a day on which €2,000.00 left for their own
  savings account
- **THEN** the row is listed with its amount and labelled as between their
  accounts, and not as unusual

#### Scenario: Among largest payments
- **WHEN** a member reading under Yours paid €800.00 into the joint account,
  and it is among the month's five largest payments
- **THEN** it is listed there, labelled as between their accounts

#### Scenario: In recent transactions
- **WHEN** one of the five most recent transactions is half of a transfer
- **THEN** it is listed as before, and labelled as between their accounts

#### Scenario: A transfer that is counted under this scope
- **WHEN** a member reading under Yours points at the day they paid into the
  joint account
- **THEN** the row is still labelled as between their accounts
