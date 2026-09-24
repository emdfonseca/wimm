# banking/overview Specification

## Purpose
Covers the screen a member lands on after signing in: the total they already
have, a recent slice of their own transactions, and how their balance has
been trending, their months with a typical and an average month, their
recurring and unusual payments, and the one scope that decides which owned
accounts those count — each scoped to what that member may see, and never
account management.

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

### Requirement: A typical month and an average month
**Story**: S1
Overview SHALL state a typical month's net, which is the middle one of the
newest six full months held, and the average month's net over those same
months beside it. Where six or fewer full months are held, both SHALL be drawn
from every full month held. Overview SHALL say how many full months they are
drawn from, and where more than six are held SHALL say they are the last six.
Neither SHALL be stated from fewer than three full months; Overview SHALL say
instead that they appear once three full months are held. The month so far and
any partly held month SHALL be in neither. The months shown SHALL NOT change:
a full month older than the last six stays in the months, in their chart, and
in what they say about their span.

Overview SHALL say in a sentence whether more comes in than goes out in a
typical month, or more goes out than comes in, with the amount. That sentence
SHALL be drawn from the typical month and never from the average, because one
exceptional month moves an average and barely moves the middle month.

#### Scenario: Living within what comes in
- **WHEN** a member holds twelve full months, and the typical month of the
  newest six is €189.40 more in than out
- **THEN** Overview says that in a typical month €189.40 more comes in than
  goes out, and that this is drawn from the last 6 full months

#### Scenario: Six or fewer full months held
- **WHEN** a member's ledger holds four full months
- **THEN** the typical and average month are drawn from all four, and Overview
  says they are drawn from 4 full months

#### Scenario: An exceptional month older than the last six
- **WHEN** a member holds twelve full months, and the eighth newest held a
  €4,000 repair
- **THEN** neither the typical nor the average month moves because of it, and
  that month is still shown among the months with its repair

#### Scenario: More goes out than comes in
- **WHEN** the typical month is €212.40 more out than in
- **THEN** Overview says plainly that more goes out than comes in, and by how
  much in a typical month

#### Scenario: The two figures disagree
- **WHEN** one of the last six months held a large repair, so the average month
  is below zero and the typical month is above it
- **THEN** both figures are shown, and the sentence follows the typical month

#### Scenario: Two full months held
- **WHEN** a member's ledger holds two full months
- **THEN** the months are shown with no typical or average month, and Overview
  says they appear once three full months are held

#### Scenario: The month so far is a bad one
- **WHEN** the month so far is well below zero on the 5th, before the salary
- **THEN** the typical and average month do not move, because the month so far
  is in neither

### Requirement: The merchants behind a month's rise
**Story**: S2
For the month a member is reading, Overview SHALL name up to five merchants
that took more that month than they usually do, each with the month's total and
how much more than usual that is, largest rise first. Usual SHALL be that
merchant's own middle month across the full months shown, counting a month
with no payment to them as nothing. Merchants SHALL be grouped by the name
`banking/transactions` defines, from booked money going out of accounts the
member owns, within the scope in force.

A merchant not usually paid at all SHALL say so rather than show a rise
against nothing. The month so far SHALL list no such merchants, because a
part-month set against whole months understates everything. Where fewer than
three full months are held there is no usual, and no such list.

#### Scenario: Fuel cost more in August
- **WHEN** a member reads August, where €246.80 went to Galp against a usual
  €164.40
- **THEN** Galp is listed with €246.80 and €82.40 more than usual

#### Scenario: A merchant they do not usually pay
- **WHEN** a member reads a month holding €120.00 to a merchant paid in no
  other month
- **THEN** that merchant is listed with €120.00 and as not usually paid

#### Scenario: Nothing rose
- **WHEN** a member reads a month where no merchant took more than usual
- **THEN** Overview says nothing rose that month, rather than showing an empty
  list

#### Scenario: Reading the month so far
- **WHEN** a member reads the month so far
- **THEN** no merchants are listed as having risen

#### Scenario: Several payments to one merchant
- **WHEN** a merchant was paid four times in the month
- **THEN** it appears once, with the four payments' total

### Requirement: Recurring payments are listed with when the next is expected
**Story**: S3
Overview SHALL list the member's recurring payments, as
`banking/payment-patterns` defines them, within the scope in force: each with
the merchant's name, the amount last paid, its cadence in words, the account
it leaves, and the date the next one is expected, soonest first. One that is
late SHALL say when it was expected rather than give a date in the past as if
it were ahead.

A member with no recurring payment SHALL see no such section. The list MUST NOT
offer to cancel, pause or change anything: wimm reads banks and instructs none.

#### Scenario: What is due and when
- **WHEN** a member with a monthly rent, two monthly subscriptions and a yearly
  insurance opens Overview
- **THEN** they see all four, each with its amount, its cadence, its account
  and its next expected date, the soonest first

#### Scenario: One is late
- **WHEN** a monthly payment was expected two days ago and has not been booked
- **THEN** it is listed as having been expected on that date

#### Scenario: A yearly payment seen twice
- **WHEN** an insurance premium has been paid in two Marches running
- **THEN** it is listed as likely yearly, with next March as its expected date

#### Scenario: Nothing recurs
- **WHEN** a member has no recurring payment
- **THEN** Overview shows no recurring payments section

#### Scenario: A member who owns nothing
- **WHEN** a member owns no account, whatever they have been granted
- **THEN** Overview shows no recurring payments section

#### Scenario: Two currencies
- **WHEN** a member has recurring payments in two currencies
- **THEN** each is listed in its own currency, and no figure adds them up

### Requirement: A month says how many unusual payments it held, and lists them when selected
**Story**: S4
Each month shown SHALL say how many unusual payments it held, as
`banking/payment-patterns` defines them, and SHALL give its net both as it was
and without them. A month that held none SHALL say nothing about unusual
payments, rather than report a count of zero.

The count and the net without them SHALL cover unusual payments in both
directions. Selecting a month SHALL list them, money going out and money coming
in together, told apart by the sign on the amount and by the label unusual
income on one coming in. Each SHALL show its date, its name, its signed amount
and what is typical for the baseline it was measured against, or that it is a
first payment to that merchant, and each SHALL lead to that payment in
Transactions. They
SHALL follow the scope in force. Nothing about a payment being unusual removes
it from any other list on Overview.

#### Scenario: A month with a repair in it
- **WHEN** March held a €1,650.00 garage bill and a €210.00 fuel bill, both
  unusual, and ended at −€1,480.30
- **THEN** March reads 2 unusual, −€1,480.30, and +€379.70 without them

#### Scenario: Selecting that month
- **WHEN** the member selects March
- **THEN** they see both payments, each with its date, its name and its amount,
  the fuel bill with "usually about €60" and the garage bill as a first payment
  to that garage

#### Scenario: A month with a bonus in it
- **WHEN** May held a €9,804.00 bonus that is unusual income and ended at
  +€10,029.00
- **THEN** May reads 1 unusual, +€10,029.00, and +€225.00 without it, and
  selecting it lists the bonus with a plus sign and the label Unusual income

#### Scenario: Going to the payment
- **WHEN** the member chooses one of those payments
- **THEN** they land in Transactions with that payment on screen

#### Scenario: A month with none
- **WHEN** a month held no unusual payment
- **THEN** it shows its net once and no mention of unusual payments

#### Scenario: Still in the other lists
- **WHEN** an unusual payment is among the month's five largest
- **THEN** it is still listed under Largest payments

### Requirement: The months can be read without unusual payments
**Story**: S4
Where any month shown held an unusual payment, Overview SHALL offer one view
that reads the months, the typical month, the average month and the sentence
drawn from the typical month without unusual payments, and SHALL say while it
is on that they are set aside and how many. It SHALL set aside unusual payments
in both directions, so a bonus flatters the typical month no more than a repair
harms it. It SHALL be a way of reading the
figures and nothing else: every unusual payment stays listed and labelled in
its month and in Transactions, and the month summary, top spending and balance
chart do not change. All payments SHALL be the view Overview opens in.

Where no month shown held an unusual payment, the view SHALL not be offered.

#### Scenario: Setting the repair aside
- **WHEN** a member whose average month is −€37.80 because of one repair
  chooses to read without unusual payments
- **THEN** the average month reads +€157.20, March no longer stands below the
  line, and Overview says 3 unusual payments are set aside

#### Scenario: Setting a bonus aside
- **WHEN** one month held a bonus that is unusual income and the member reads
  without unusual payments
- **THEN** that month's net and the average month are read without the bonus

#### Scenario: The payments are still there
- **WHEN** the view without unusual payments is on and the member selects March
- **THEN** March still lists its unusual payments

#### Scenario: Coming back
- **WHEN** a member opens Overview again
- **THEN** it reads with all payments

#### Scenario: Nothing to set aside
- **WHEN** no month shown held an unusual payment
- **THEN** the view is not offered

### Requirement: One scope for everything drawn from transactions
**Story**: S5
Overview SHALL offer one control, with the choices Household, Yours and All,
that decides which accounts everything it draws from transactions counts: the
month summary, top spending, the balance chart and what moved a day, the
months and their typical and average, the merchants behind a rise, recurring
payments and unusual payments. It MUST NOT be offered per section. It SHALL not
change the two money figures, the accounts list or recent transactions.

Every scope SHALL count only accounts the member owns, since an account's
transactions are its owners' alone (`banking/transactions`). Household SHALL be
the household-money accounts the member owns; Yours SHALL be the accounts they
own that are not household money; All SHALL be both, and SHALL be what Overview
opens in. An account the member merely holds a grant on SHALL be in no scope.

A choice with no account in it SHALL not be offered, All SHALL be offered only
alongside both others, and where one choice would remain the control SHALL be
absent. Where Household counts fewer accounts than the household money figure
does, Overview SHALL name the accounts it counts and the ones it does not, and
say why. The choice SHALL survive a reload and be shareable as an address.

#### Scenario: Are we overspending
- **WHEN** a member who owns a joint account and two of their own chooses
  Household
- **THEN** the month summary, top spending, the balance chart, the months,
  recurring payments and unusual payments all count the joint account only

#### Scenario: Am I overspending
- **WHEN** that member chooses Yours
- **THEN** the same sections count their two own accounts only

#### Scenario: As it opens
- **WHEN** that member opens Overview
- **THEN** All is chosen and every section counts all three accounts, as the
  month summary did before the control existed

#### Scenario: Household money they do not own
- **WHEN** household money includes a savings account the member holds details
  on and does not own, and they choose Household
- **THEN** Overview says Household counts the joint account, and that the
  savings account is household money whose transactions are its owners' to see

#### Scenario: A scope with no full month
- **WHEN** the scope control is on offer and a member chooses a scope whose
  accounts hold no full calendar month
- **THEN** Month by month is still shown, saying no full month is held for
  these accounts yet, and no typical or average month

#### Scenario: A scope with no recurring payment
- **WHEN** the scope control is on offer and a member chooses a scope whose
  accounts hold no recurring payment
- **THEN** Recurring payments is still shown, saying there are none in these
  accounts

#### Scenario: A scope with nothing to set aside
- **WHEN** a member chooses a scope holding a full month and no unusual payment
- **THEN** the view control is still shown, and says there are no unusual
  payments to set aside

#### Scenario: A household of one
- **WHEN** the member is the only member of their household
- **THEN** no scope control is shown

#### Scenario: Nothing of their own
- **WHEN** every account a member owns is household money
- **THEN** no scope control is shown, because Household and All would count the
  same accounts

#### Scenario: The figures do not move
- **WHEN** a member changes the scope
- **THEN** Household money, Your money, the accounts list and recent
  transactions stay as they were

#### Scenario: Reloading
- **WHEN** a member chooses Yours and reloads the page
- **THEN** Yours is still chosen

#### Scenario: A scope that no longer applies
- **WHEN** a member follows an address naming Household after the joint account
  was handed to somebody else
- **THEN** Overview opens in All

### Requirement: Pointing at a day says what moved it
**Story**: S6
Pointing at a day on the balance chart SHALL show, beside that day, its date,
its balance, the change from the day before in words, and the booked
transactions that explain the day: the largest by size, in either direction,
until they account for most of what moved that day, never more than five, with
the rest as one line saying how many smaller ones there were. Each SHALL show
its merchant name and its signed amount, and one that
`banking/payment-patterns` calls unusual SHALL be labelled Unusual, or Unusual
income where it came in. They SHALL
be drawn only from the accounts the chart is drawn from, within the scope in
force.

It SHALL open from a pointer, from a tap on a touch screen, and from the
keyboard by the same keys that move along the chart, and what it shows SHALL be
announced to assistive technology. It MUST NOT be reachable by hover alone. A
day with no booked transaction SHALL say so. The first day of the chart SHALL
give no change, since there is no day before it on the chart.

#### Scenario: What happened on the day it dropped
- **WHEN** a member points at 3 August, when €640.00 went to a DIY store, €92.10
  to a petrol station, and three small payments were made
- **THEN** they see 3 August, its balance, €805.79 less than 2 August, the two
  larger payments by name with their amounts, and "and 3 smaller"

#### Scenario: One of them is unusual
- **WHEN** the €640.00 payment is unusual
- **THEN** it carries the label Unusual

#### Scenario: A bonus arrives
- **WHEN** a member points at the day an unusual bonus arrived
- **THEN** it is listed with a plus sign and the label Unusual income

#### Scenario: A day when nothing happened
- **WHEN** a member points at a day with no booked transaction
- **THEN** they see the date, the balance, and "No transactions this day."

#### Scenario: Money arriving
- **WHEN** a member points at the day the salary arrived
- **THEN** the salary is listed with a plus sign, and the change reads as more
  than the day before

#### Scenario: By keyboard
- **WHEN** a member focuses the chart and presses the left arrow, Home or End
- **THEN** the same information opens for the day the marker lands on, and is
  announced

#### Scenario: By touch
- **WHEN** a member taps a day on a phone
- **THEN** the same information opens for that day and stays until they tap
  elsewhere

#### Scenario: A busy day
- **WHEN** a day holds twelve payments of similar size
- **THEN** five are listed, followed by "and 7 smaller"

#### Scenario: An account the chart leaves out
- **WHEN** an account is left out of the chart because its history is short
- **THEN** its transactions are not listed for any day, since the change shown
  does not include them

#### Scenario: Under another scope
- **WHEN** the scope is Household
- **THEN** a day lists only what moved on the household accounts the member
  owns

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
