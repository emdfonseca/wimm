# banking/overview Specification

## Purpose
Covers the screen a member lands on after signing in: the total they already
have, a recent slice of their own transactions, and how their balance has
been trending — each scoped to what that member may see, and never account
management.

## Requirements

### Requirement: Overview shows a recent slice of the member's own transactions
**Story**: S3
Overview SHALL show the member's most recent transactions, drawn only from
accounts they own — transactions stay owner-only, per
`banking/transactions`. A member who owns no account, or whose owned accounts
have no transactions read yet, SHALL see no recent-transactions section at
all, rather than one stating there is nothing, because a section that could
state either emptiness or absence-of-access would tell an owner-less member
something they are not owed.

Each transaction SHALL show enough to identify it without opening
Transactions: its date, its name as `banking/transactions` defines it, its
amount, and which account it belongs to. Overview SHALL offer a way to reach
the full ledger.

#### Scenario: Recent activity at a glance
- **WHEN** a member who owns at least one account with transactions opens
  Overview
- **THEN** they see their most recent transactions, each showing its date,
  its name, its amount, and its account

#### Scenario: A payment is named, not quoted
- **WHEN** one of the recent transactions is a card payment whose bank line
  carries a transaction word and a reference number
- **THEN** it is listed under the merchant's name, without the word or the
  number

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

### Requirement: Before any bank is connected, Overview says so
**Story**: S4
A household that has connected no bank, or a member who may see no account,
SHALL be shown on Overview what the screen is for and a way to reach Accounts
to connect one — not a total of zero, nor any empty section.

#### Scenario: The first member arrives
- **WHEN** a member of a household with no connected banks signs in
- **THEN** Overview explains what connecting a bank does and offers a way to
  reach Accounts, showing no figure, no list of accounts, no month summary, no
  top spending, no recent-transactions section, and no balance chart

#### Scenario: A member who may see no account
- **WHEN** a bank is connected but a member has been granted nothing on any
  of its accounts
- **THEN** they see Overview's before-any-bank explanation, not a total of
  zero

### Requirement: The month so far, against the same days of last month
**Story**: S1
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
see no month summary at all. The summary SHALL say that money moved between
the member's own accounts is counted, since wimm cannot tell it apart.

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
- **WHEN** a member reads the month summary
- **THEN** it tells them that money moved between their own accounts is
  counted as money out of one and into the other

#### Scenario: A member who owns nothing
- **WHEN** a member owns no account, whatever they have been granted
- **THEN** Overview shows no month summary, rather than one of zeros

#### Scenario: Looking at the month in full
- **WHEN** a member wants to see what made up a figure
- **THEN** the summary offers a way to open Transactions at this month

### Requirement: Where the most money went this month
**Story**: S1
Overview SHALL show, per currency and for the same days the month summary
covers, the five merchants the most money went to, each with its total and
the number of payments, and the five largest single payments, each with its
date, its name and its account. Both SHALL be drawn only from the booked
transactions of accounts the member owns, SHALL count money going out only,
and SHALL group by the name `banking/transactions` defines.

A member who owns no account, or whose owned accounts have no booked payment
this month, SHALL see no such section.

#### Scenario: Several payments to one merchant
- **WHEN** a member has paid one merchant four times this month
- **THEN** that merchant appears once, with the four payments' total and the
  count of four

#### Scenario: The ranking is by money, not by visits
- **WHEN** one merchant was paid once for a large amount and another ten times
  for small ones totalling less
- **THEN** the first ranks above the second

#### Scenario: The largest single payments
- **WHEN** a member looks for what the biggest payments were
- **THEN** they see the five largest this month, each with its date, its name
  and the account it left

#### Scenario: Fewer than five
- **WHEN** a member has paid three merchants this month
- **THEN** three are listed, with nothing standing in for the missing two

#### Scenario: Money coming in is not spending
- **WHEN** a member has received a salary and a refund this month
- **THEN** neither appears among the merchants or the largest payments

#### Scenario: Nothing spent yet
- **WHEN** no payment has been booked on any owned account this month
- **THEN** the section is absent, rather than present and empty

### Requirement: Overview lists the accounts behind the figures
**Story**: S4
Overview SHALL list every account the member may see, each with its name, its
bank, its balance and when that balance was read, grouped under what it counts
towards: household money, the member's own, or shared with them and counted in
neither. Each figure SHALL agree with the accounts in its group, as
`banking/household-accounts` requires. A group with no account in it SHALL NOT
be shown.
An account SHALL be shown at the level the member holds on it and no further.
An account left out of wimm SHALL NOT be listed, as it counts towards no
total.

The list is for reading. Each account SHALL lead to the Accounts screen, and
the list SHALL carry no control that changes anything.

#### Scenario: Seeing where each figure sits
- **WHEN** a member with a joint account and two personal ones opens Overview
- **THEN** the joint account is listed under household money and the other two
  as theirs, and each group's balances add up to the figure shown for it

#### Scenario: An account shared with them
- **WHEN** a member has been given *balance* on a partner's savings account
- **THEN** it is listed apart from the other two groups with its balance, and
  neither figure includes it

#### Scenario: A balance read some time ago
- **WHEN** an account's balance was last read yesterday
- **THEN** the account says when it was read, and is not presented as current

#### Scenario: An account granted at balance
- **WHEN** a member has been granted *balance* on an account they do not own
- **THEN** it is listed with its bank, name and balance, and nothing that only
  *details* or ownership would show

#### Scenario: An account left out
- **WHEN** a member owns an account they have left out of wimm
- **THEN** it does not appear in Overview's list, and is still found on
  Accounts

#### Scenario: Going to an account
- **WHEN** a member chooses an account in the list
- **THEN** they are taken to the Accounts screen

### Requirement: The balance chart follows the ledger, and only as far as the ledger reaches
**Story**: S2
Overview SHALL show how the member's balance has moved, per currency, as a
line with one point for each day, derived from the member's owned accounts'
booked transactions — the same rows `banking/transactions` already stores,
never a separate reading kept for this purpose. The chart SHALL cover at most
the last 90 days, and SHALL end on today's combined balance of the accounts it
covers, which are the accounts the member owns, household or their own.

A member SHALL be able to read the chart, not only look at it: the dates it
starts and ends on, its highest and lowest balance, and the balance on any
single day they point at or move to with the keyboard. What the chart shows
SHALL also be stated in words, for a member who cannot see it.

The chart SHALL cover only the span each contributing account's ledger
actually reaches back to — its own earliest booked transaction. wimm MUST NOT
draw a line across a gap it has no transactions for, and MUST NOT imply a
longer history than it holds. One account with a short history MUST NOT
shorten the chart for the rest: where an account's ledger reaches back fewer
than 30 days and another's reaches further, the shorter one SHALL be left out
of the chart. Where no account's ledger reaches a week, wimm SHALL draw no
chart and SHALL say that one appears once there is a week of history.

An account left out of the chart — because its connection has never been
granted transaction access, or because its history is too short — SHALL still
count in whichever figure it belongs to, and Overview SHALL say plainly that the chart does
not cover every account rather than drawing a line as if it did.

Where a member's accounts are held in more than one currency, wimm SHALL show
one chart per currency and MUST NOT convert or combine them, matching how the
total is shown.

#### Scenario: A chart from what has been read
- **WHEN** a member owns an account whose transactions reach 90 days back
- **THEN** the chart covers those 90 days with a point for each, and ends on
  today's balance

#### Scenario: Reading a day off the chart
- **WHEN** a member points at a place on the chart, or moves along it with
  the keyboard
- **THEN** they are shown that day's date and the balance on it

#### Scenario: The ends and the extremes are labelled
- **WHEN** a member looks at the chart without touching it
- **THEN** they can read the date it starts, the date it ends, and the
  highest and lowest balance it reached

#### Scenario: A member who cannot see the chart
- **WHEN** a member reads Overview with a screen reader
- **THEN** they are told the span the chart covers, the balance at each end,
  and the highest and lowest balance and when each happened

#### Scenario: One account connected yesterday
- **WHEN** a member owns one account with three months of transactions and a
  second whose transactions reach back one day
- **THEN** the chart covers the three months from the first account alone,
  and wimm says the chart does not cover every account

#### Scenario: Less than a week of history anywhere
- **WHEN** no account the member owns has transactions reaching back a week
- **THEN** no chart is drawn, and the member is told one appears once there
  is a week of history

#### Scenario: An account with balances only
- **WHEN** a member owns one account with transaction access and one with
  balance-only access
- **THEN** both count in the figures above, only the first contributes to the
  chart, and wimm says the chart does not cover every account

#### Scenario: Two currencies, two charts
- **WHEN** a member's owned accounts are held in two currencies, both in use
- **THEN** a separate chart is shown for each currency, and none is combined
  or converted into the other

#### Scenario: No transaction history anywhere
- **WHEN** none of a member's owned accounts has any transaction access
- **THEN** no chart is shown, rather than a flat or invented one

### Requirement: Overview shows the total and performs no connection action
**Story**: S5
Overview SHALL show household money and the member's own money as
`banking/household-accounts` defines them, as two figures side by side, with
one exception: a currency whose figures are zero and whose accounts have had
no movement in the span the balance chart covers SHALL get no figure, no month
summary and no balance chart on Overview. Its accounts
SHALL still appear in Overview's list of accounts, so nothing the member may
see is hidden from them.

Overview SHALL NOT offer any connection action: connecting, restoring,
disconnecting, leaving an account out, or choosing who sees an account. Where
a member wants any of that, Overview SHALL offer a way to reach the Accounts
screen.

At every width wider than a phone, Overview SHALL use the full width of the
content area, and MUST NOT stack narrow sections down one edge beside empty
space.

#### Scenario: The total is there, the connection controls are not
- **WHEN** a member with two connected banks opens Overview
- **THEN** they see household money and their own money, and no control
  anywhere on the screen that connects, restores, disconnects or changes who
  sees anything

#### Scenario: Ours and mine, side by side
- **WHEN** a member who owns a personal account and co-owns a joint one opens
  Overview
- **THEN** the joint account's balance is shown as household money, the
  personal one as their own, and nowhere are the two added together

#### Scenario: Getting to account management from Overview
- **WHEN** a member on Overview wants to connect, restore, or disconnect a
  bank
- **THEN** Overview offers a way to reach the Accounts screen, and performs
  none of those actions itself

#### Scenario: A currency holding nothing
- **WHEN** a member may see one account in a second currency, its balance is
  zero, and nothing has moved on it
- **THEN** Overview shows no figure, no month summary and no balance chart for
  that currency, and the account still appears in the list of accounts with
  its zero balance

#### Scenario: A currency at zero that has been used
- **WHEN** a currency's figures are zero today and its accounts have had
  transactions within the span the balance chart covers
- **THEN** that currency is shown like any other, because a balance that
  reached zero is something the member came to see

#### Scenario: The screen uses the space it has
- **WHEN** a member opens Overview on a laptop or a wider screen
- **THEN** the sections are laid out across the width of the content area,
  and the balance chart spans all of it
