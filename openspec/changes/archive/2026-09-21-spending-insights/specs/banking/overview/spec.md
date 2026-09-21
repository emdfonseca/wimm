## ADDED Requirements

### Requirement: Month by month, as far back as the ledger is whole
**Story**: S1
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
month. A month before every ledger begins SHALL not be shown at all. The section SHALL say that money moved between the
member's own accounts is counted, as the month summary does.

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

### Requirement: A typical month and an average month
**Story**: S1
Overview SHALL state a typical month's net, which is the middle one of the full
months shown, and the average month's net beside it, and SHALL say how many
full months they are drawn from. Neither SHALL be stated from fewer than three
full months; Overview SHALL say instead that they appear once three full months
are held. The month so far and any partly held month SHALL be in neither.

Overview SHALL say in a sentence whether more comes in than goes out in a
typical month, or more goes out than comes in, with the amount. That sentence
SHALL be drawn from the typical month and never from the average, because one
exceptional month moves an average and barely moves the middle month.

#### Scenario: Living within what comes in
- **WHEN** the typical month over twelve full months is €189.40 more in than
  out
- **THEN** Overview says that in a typical month €189.40 more comes in than
  goes out, and that this is drawn from 12 full months

#### Scenario: More goes out than comes in
- **WHEN** the typical month is €212.40 more out than in
- **THEN** Overview says plainly that more goes out than comes in, and by how
  much in a typical month

#### Scenario: The two figures disagree
- **WHEN** one month held a large repair, so the average month is below zero
  and the typical month is above it
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
