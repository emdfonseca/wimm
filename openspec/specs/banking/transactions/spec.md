# banking/transactions Specification

## Purpose
Covers the household's ledger: which transactions a member may see, what the
list shows and in what order, how far back it reaches, how it is brought up to
date and what it says about its own freshness, how a transaction the bank has
not settled differs from one it has, and what the screen says before anything
has been read at all.

## Requirements

### Requirement: Transactions are seen by an account's owners
**Story**: S1
An account's transactions SHALL be visible to every member who owns that
account, and to no other member. A member who has been granted *balance* or
*details* on an account they do not own SHALL see none of its transactions and
SHALL NOT be told how many there are.

Seeing an account's balance and seeing what the account has been spent on are
different sentences, and a member who was shown the first has not agreed to the
second.

#### Scenario: An owner sees their account's transactions
- **WHEN** a member who owns an account opens Transactions
- **THEN** that account's transactions are among those listed

#### Scenario: Both owners of a joint account see its transactions
- **WHEN** an account is owned by two members
- **THEN** each of them sees its transactions in full

#### Scenario: A member granted the balance sees no transactions
- **WHEN** a member who does not own an account has been granted *balance* on
  it, and opens Transactions
- **THEN** none of that account's transactions appear, and nothing tells them
  that transactions were withheld

#### Scenario: A member granted the details sees no transactions either
- **WHEN** a member who does not own an account has been granted *details* on
  it, and opens Transactions
- **THEN** none of that account's transactions appear, because the details of
  an account and what has been spent from it are different things

#### Scenario: A transaction on a jointly owned account
- **WHEN** an account is owned by two members and a payment is made from it
- **THEN** that payment appears in both members' lists, identically, and wimm
  does not say which of them made it, because the bank does not say either

#### Scenario: Two members see different ledgers
- **WHEN** one member owns two accounts and a second member owns one of them
- **THEN** the first sees the transactions of both and the second sees the
  transactions of one, and neither is told what the other sees

### Requirement: The list of transactions
**Story**: S1
A member SHALL be able to see the transactions of every account they own in one
list, newest first, grouped by the day they happened. Each transaction SHALL
show who it was with, its amount carrying a sign, and which account it came
from. An amount's direction MUST NOT depend on colour alone.

#### Scenario: The transactions are listed newest first
- **WHEN** a member opens Transactions
- **THEN** the most recent transaction is at the top, and the transactions are
  grouped under the day they happened

#### Scenario: What a transaction shows
- **WHEN** a member looks at one transaction
- **THEN** they can see who it was with, how much it was, and which of their
  accounts it came from

#### Scenario: Money leaving and money arriving
- **WHEN** a member sees a payment out and a payment in
- **THEN** each amount carries a sign showing its direction, so the two can be
  told apart without relying on colour

#### Scenario: The bank gave no name for the other party
- **WHEN** a bank returns a transaction without naming who it was with
- **THEN** wimm shows what the bank did give and never a blank row or an
  invented name

### Requirement: The transactions of one account
**Story**: S1
A member SHALL be able to narrow the list to a single account they own, and to
widen it again. Narrowing SHALL name the account it is narrowed to and MUST NOT
present itself as a different screen.

#### Scenario: Narrowing to one account
- **WHEN** a member chooses to see only one account's transactions
- **THEN** only that account's transactions are listed, and the account it is
  narrowed to is named on screen

#### Scenario: Widening again
- **WHEN** a member removes the narrowing
- **THEN** every transaction they may see is listed again, without navigating
  anywhere

#### Scenario: An account with nothing in the period
- **WHEN** a member narrows to an account that has no transactions
- **THEN** they are told that account has none, rather than being shown an
  empty list with no explanation

### Requirement: A transaction the bank has not settled
**Story**: S1
A transaction the bank has not settled SHALL be shown alongside settled ones
and SHALL be marked as unsettled, because it is money that has moved as far as
the member is concerned. When the bank settles it, it SHALL appear once and not
twice.

#### Scenario: An unsettled transaction is marked
- **WHEN** a member's bank returns a transaction it has not settled
- **THEN** it is listed with the others and marked as not yet settled

#### Scenario: A transaction settles
- **WHEN** a transaction that was unsettled is settled by the bank
- **THEN** it appears once in the list, no longer marked, and the household is
  not shown both versions of it

#### Scenario: An unsettled transaction the bank drops
- **WHEN** a bank stops returning a transaction it never settled
- **THEN** it stops appearing, because it never happened

### Requirement: The list says how current it is
**Story**: S1
The list SHALL show when it was last brought up to date from the banks, and a
member SHALL be able to ask for it to be brought up to date at any time. wimm
MUST NOT present the list as current when it has not been re-read.

Where a bank or the service reaching it refuses because wimm has asked too
often, the member SHALL be told when it can next be tried, and the transactions
already on screen SHALL remain with their original time.

#### Scenario: The list says when it was last updated
- **WHEN** a member opens Transactions
- **THEN** they can see when the list was last brought up to date from the
  banks

#### Scenario: Transactions appear before the banks have answered
- **WHEN** a member opens Transactions and wimm has not yet finished asking the
  banks
- **THEN** the transactions already held are shown straight away with the time
  they were last brought up to date, and the list updates when the banks answer

#### Scenario: Bringing the list up to date
- **WHEN** a member asks for the list to be brought up to date
- **THEN** new transactions appear and the list shows a new time

#### Scenario: Refusing to be asked again so soon
- **WHEN** a member asks to bring the list up to date and the bank or the
  service reaching it refuses because it has been asked too often
- **THEN** the member is told when it can next be tried, and the transactions
  on screen stay where they are with their original time

#### Scenario: One bank answers and another does not
- **WHEN** the list is brought up to date and one bank cannot be reached
- **THEN** the transactions that were re-read are current, the ones that could
  not be re-read are still shown, and the member is told which bank did not
  answer

### Requirement: How far back the transactions go
**Story**: S1
When wimm first reads an account's transactions it SHALL ask that bank for the
earliest it will give and take what it returns. wimm SHALL tell the member the
date the list actually reaches back to, and MUST NOT state a period before it
has read one, because no bank says in advance how much history it will hand
over.

#### Scenario: The list says how far back it goes
- **WHEN** a member has opened Transactions for an account read at least once
- **THEN** they can see the date the list reaches back to

#### Scenario: A bank that gives very little history
- **WHEN** a bank returns only a few weeks of transactions
- **THEN** the list says it reaches back to that date, rather than implying
  months are missing or that more will arrive

#### Scenario: Nothing is promised before the first read
- **WHEN** a member has just widened a bank's access and nothing has been read
  yet
- **THEN** they are told the transactions are being fetched, and are not given
  a period that wimm cannot yet know

### Requirement: Two transactions that look identical are two transactions
**Story**: S1
Where a member makes the same payment twice, both SHALL be listed. Where wimm
reads the same transaction more than once, it SHALL be listed once. wimm MUST
NOT collapse two real transactions into one, because a household that paid
twice needs to see that it paid twice.

#### Scenario: The same payment made twice in one day
- **WHEN** a member pays the same amount to the same place twice on the same
  day
- **THEN** both payments are listed

#### Scenario: The same transaction read twice
- **WHEN** wimm reads a bank twice and the same transaction is returned both
  times
- **THEN** it appears once in the list

### Requirement: Access running out does not reset the ledger
**Story**: S1
When a bank's access runs out, the transactions already read SHALL still be
listed, and the list SHALL say that this bank is no longer being read. When a
member restores the access, reading SHALL carry on from where it stopped rather
than starting the history again.

At a bank whose access lasts a single day this is the difference between a
ledger and a screen that empties every night.

#### Scenario: A bank whose access has run out
- **WHEN** a member opens Transactions and one bank's access has run out
- **THEN** that bank's transactions are still listed, and the member is told
  which bank is no longer being read and how to restore it

#### Scenario: Restoring carries on
- **WHEN** a member restores a bank's access
- **THEN** the transactions already read are still there, anything that
  happened while the access was out arrives, and nothing is listed twice

#### Scenario: A bank whose access runs out every day
- **WHEN** a member restores a connection whose bank grants access for one day
- **THEN** the history is not read again from the beginning, and the list keeps
  reaching back to the date it already reached

### Requirement: Going straight to a month
**Story**: S4
A member SHALL be able to go straight to a month of the ledger. wimm SHALL offer
the months that actually hold transactions that member can see, newest first,
and MUST NOT offer a month with nothing in it — an empty month is a control that
does nothing, and the reason the list is built from what is there rather than
from a calendar.

Choosing a month SHALL land the member on that month's newest transactions, and
paging from there SHALL continue through the ledger normally in both directions,
because a month is a place to start reading rather than a filter that cuts the
list down.

The months offered SHALL be the months of the transactions on screen: where a
member is looking at one account, the months are that account's.

#### Scenario: Jumping to a month
- **WHEN** a member chooses April from the months offered
- **THEN** they land on the newest transactions in April, and the dates on
  screen say so

#### Scenario: Paging on from a month
- **WHEN** a member has jumped to a month and asks for older transactions
- **THEN** the next page continues from where that page stopped, running into
  the month before it when the month runs out, with nothing repeated and nothing
  skipped

#### Scenario: Only months that hold something
- **WHEN** a household read nothing in March and several things in February
- **THEN** March is not offered and February is

#### Scenario: The months follow the account being looked at
- **WHEN** a member narrows the ledger to one account
- **THEN** the months offered are the months that account has transactions in

#### Scenario: A month holding more than one page
- **WHEN** a member jumps to a month with more transactions than fit on a page
- **THEN** they land on that month's newest page and can page through the rest
  of it

#### Scenario: Not enough ledger to need it
- **WHEN** every transaction a member can see falls in one month
- **THEN** no months are offered, because there is nowhere else to go

#### Scenario: A member who owns no account
- **WHEN** a member owns no account
- **THEN** no months are offered, as no transactions are

### Requirement: An account left out stops filling the ledger
**Story**: S3
While an account is left out of wimm, no transaction SHALL be read for it and
none of its transactions SHALL be listed, for any member including its owners.
Its transactions SHALL NOT be deleted, and bringing the account back SHALL bring
them back with it.

This is deliberately unlike a bank being disconnected, where what was already
read stays on screen. Disconnecting ends wimm's access to a bank; leaving an
account out is a member saying that account is not in wimm, and a ledger that
still listed it would contradict them.

#### Scenario: The rows go
- **WHEN** an owner leaves an account out
- **THEN** its transactions stop being listed under Transactions, for them and
  for everybody

#### Scenario: Nothing new arrives
- **WHEN** an account is left out
- **THEN** no transaction is read for it, however often the ledger is brought up
  to date

#### Scenario: The rows come back
- **WHEN** an owner brings a left-out account back
- **THEN** its transactions are listed again, carrying on from where they
  stopped rather than starting empty

#### Scenario: The ledger's count and dates follow
- **WHEN** an account is left out
- **THEN** the number of transactions shown and the span of dates on screen
  describe what is actually listed, without counting what is no longer there

#### Scenario: Every account a member owns is left out
- **WHEN** a member owns accounts and has left all of them out
- **THEN** they are told the ledger is empty because their accounts are left
  out, and how to bring one back, rather than being shown a bare empty list

### Requirement: Reading a ledger longer than one page
**Story**: S4
A ledger SHALL be read a page at a time, oldest-ward and newest-ward, and the
member SHALL always be told where in time they are rather than which page they
are on. A page's position MUST NOT shift because transactions arrived while the
member was reading, because the whole list moves when the newest end grows.

The member SHALL also be able to go straight to the newest page and straight to
the oldest, without walking there. Both are a seek like any other and neither
requires wimm to know how many pages there are.

wimm MUST NOT offer page numbers or a count of pages. A seek knows neither
without a second count that would be stale before it rendered, and the span of
dates is the thing a person scanning backwards is actually looking for.

Where a day's transactions do not fit on one page, the day SHALL be named again
at the top of the next, so a row is never shown under no date.

#### Scenario: Reading further back
- **WHEN** a member reaches the end of a page and asks for older transactions
- **THEN** the next page continues from where the last one stopped, with
  nothing repeated and nothing skipped

#### Scenario: Coming back towards today
- **WHEN** a member who has paged backwards asks for newer transactions
- **THEN** they return through the same transactions in the same order

#### Scenario: Straight back to the newest
- **WHEN** a member deep in the ledger asks for the newest transactions
- **THEN** they land on the newest page in one action, and it is the same page
  they would have reached by paging newer-ward the whole way

#### Scenario: Straight to the oldest
- **WHEN** a member asks for the oldest transactions wimm holds
- **THEN** they land on the oldest page in one action, told there is nothing
  older, and paging newer-ward from it works normally

#### Scenario: Already at the end being jumped to
- **WHEN** a member on the newest page asks for the newest page
- **THEN** nothing moves and nothing is offered that would do nothing

#### Scenario: Where the member is, is a date
- **WHEN** a member is reading any page
- **THEN** they are told the span of dates they are looking at, and are not
  asked to think in page numbers

#### Scenario: The span is shown even on a single page
- **WHEN** every transaction a member can see fits on one page
- **THEN** they are still told the span of dates they are looking at, even
  though no way to page is offered

#### Scenario: New transactions arrive while a member is reading
- **WHEN** transactions are read from a bank while the member is on a page
  other than the newest
- **THEN** the page they are reading shows the same transactions it did
  before, and none of them moves to another page

#### Scenario: A day that does not fit on one page
- **WHEN** a day's transactions run past the end of a page
- **THEN** the next page names that day again above the rest of them

#### Scenario: The oldest page
- **WHEN** a member reaches the oldest transaction wimm holds
- **THEN** they are told there is nothing older, and how far back the list
  reaches

#### Scenario: Everything fits on one page
- **WHEN** a member has fewer transactions than one page holds
- **THEN** no way to page is offered

### Requirement: Transactions outlive the bank being taken away
**Story**: S3
When a bank is disconnected, the transactions already read from it SHALL remain
visible to the members who could see them, even though the accounts stop
appearing among the household's accounts and stop counting towards any total.
No bank returns history indefinitely, so ending wimm's access MUST NOT be the
same act as destroying the record.

#### Scenario: Disconnecting keeps the record
- **WHEN** a member disconnects a bank
- **THEN** its accounts no longer appear among the household's accounts, and
  the transactions already read from them are still listed under Transactions

#### Scenario: Reconnecting continues the same record
- **WHEN** a member reconnects a bank they had disconnected
- **THEN** each account carries on with the transactions it already had, rather
  than appearing a second time with a new empty list

#### Scenario: Nothing new arrives while a bank is away
- **WHEN** a bank has been disconnected
- **THEN** no new transaction appears for its accounts, because wimm is no
  longer reading them

### Requirement: Before any transaction has been read
**Story**: S1
A member with nothing to show SHALL be told why, and told the thing they can do
about it. The reasons differ and MUST NOT be collapsed into one empty list.

#### Scenario: No bank connected at all
- **WHEN** a member of a household with no connected bank opens Transactions
- **THEN** they are shown what the screen is for and a way to connect a bank

#### Scenario: A bank that wimm may not read transactions from
- **WHEN** a member's only connected bank was connected before wimm could read
  transactions
- **THEN** they are told that this bank's transactions are not being read yet,
  which bank it is, and how to change that

#### Scenario: A member who owns no account
- **WHEN** a member owns no account, whatever they have been granted on others
- **THEN** they are told that they see transactions for accounts that are
  theirs, rather than being shown an empty list

#### Scenario: An account read, with nothing in it
- **WHEN** an account has been read and the bank returned no transactions
- **THEN** the member is told the account has no transactions, and when it was
  last checked

### Requirement: A transaction is named by who it was with
**Story**: S3
Wherever wimm lists a transaction, it SHALL name it by the merchant or person
it was with, not by the bank's statement line. Where the bank supplies only a
statement line, wimm SHALL remove the bank's own bookkeeping from it: the
word for the kind of transaction at its start, reference numbers at its end,
and any trailer about the country or the original amount. Two payments to one
merchant that differ only in that bookkeeping SHALL carry the same name.

A name MUST NOT be empty. Where removing the bookkeeping would leave nothing,
the line SHALL be shown as the bank wrote it, and where the bank wrote
nothing, the transaction SHALL be named for what it is.

wimm MUST NOT discard what the bank wrote. On Transactions, where the name
differs from the bank's line, the bank's line SHALL be shown with the
transaction, because a name wimm derived has to be checkable against the
statement it came from.

#### Scenario: A card payment with a reference number
- **WHEN** the bank's line for a payment is
  `COMPRA WWW.AMAZON NM4HU1VZ4 230002268264350`
- **THEN** the transaction is listed as `Amazon`

#### Scenario: Two payments to one merchant
- **WHEN** a member has paid the same merchant twice and the bank's two lines
  differ only in their reference numbers
- **THEN** both transactions carry the same name

#### Scenario: A trailer about where and how much
- **WHEN** the bank's line ends with the country and the original amount of a
  payment made abroad
- **THEN** the name stops before that trailer

#### Scenario: The bank names the other party
- **WHEN** the bank supplies the name of the person or company on the other
  side of a transaction
- **THEN** the transaction is named by it

#### Scenario: Nothing would be left
- **WHEN** a bank's line consists only of a transaction word and a number
- **THEN** the transaction is listed under the line as the bank wrote it,
  never under an empty name

#### Scenario: The bank wrote nothing
- **WHEN** a transaction arrives with no other party and no statement line
- **THEN** it is listed as a card payment

#### Scenario: Checking a name against the statement
- **WHEN** a member looks at a transaction on Transactions whose name differs
  from the bank's line
- **THEN** the bank's line is shown with it, exactly as the bank wrote it

#### Scenario: The same name everywhere
- **WHEN** a member sees a transaction on Overview and then on Transactions
- **THEN** it carries the same name in both places

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

### Requirement: A transfer between own accounts is labelled in the ledger
**Story**: S1
Transactions SHALL label each of the two rows of a transfer between the
member's own accounts, as `banking/own-transfers` defines, as between their
accounts. The label SHALL be words and never colour alone, SHALL sit where a
row's status is shown, and SHALL be worked out across every account the member
owns, whatever scope Overview is reading in, so a row's label does not depend
on a control on another screen. A row so labelled SHALL never also be labelled
unusual.

Nothing about being a transfer SHALL remove a row from the list, change its
amount or its sign, or change how the list is ordered, paged or counted. A row
not yet settled SHALL read as not settled and nothing else.

#### Scenario: Both halves are labelled
- **WHEN** a member who moved €500.00 from their current account to their
  savings account reads Transactions
- **THEN** the row leaving the current account and the row arriving in savings
  both read as between their accounts

#### Scenario: Reading one account
- **WHEN** the member reads the savings account alone
- **THEN** the arriving row still reads as between their accounts, though its
  other half is not in the list

#### Scenario: A large transfer
- **WHEN** the transfer is far above anything the member usually pays
- **THEN** its rows read as between their accounts and not as unusual

#### Scenario: One that could not be paired
- **WHEN** a member sent €1,000.00 to an account wimm does not hold
- **THEN** the row carries no such label

#### Scenario: The other owner of a joint account
- **WHEN** a member reads a joint account into which their partner paid from an
  account only the partner owns
- **THEN** the arriving row carries no such label

#### Scenario: The count does not change
- **WHEN** a member's ledger holds forty transfers
- **THEN** the list holds every row it held before, and says the same number of
  transactions
