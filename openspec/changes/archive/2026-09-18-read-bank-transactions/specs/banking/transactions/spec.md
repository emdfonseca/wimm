## Purpose
Covers the household's ledger: which transactions a member may see, what the
list shows and in what order, how far back it reaches, how it is brought up to
date and what it says about its own freshness, how a transaction the bank has
not settled differs from one it has, and what the screen says before anything
has been read at all.

## ADDED Requirements

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

### Requirement: Reading a ledger longer than one page
**Story**: S1
A ledger SHALL be read a page at a time, oldest-ward and newest-ward, and the
member SHALL always be told where in time they are rather than which page they
are on. A page's position MUST NOT shift because transactions arrived while the
member was reading, because the whole list moves when the newest end grows.

Where a day's transactions do not fit on one page, the day SHALL be named again
at the top of the next, so a row is never shown under no date.

#### Scenario: Reading further back
- **WHEN** a member reaches the end of a page and asks for older transactions
- **THEN** the next page continues from where the last one stopped, with
  nothing repeated and nothing skipped

#### Scenario: Coming back towards today
- **WHEN** a member who has paged backwards asks for newer transactions
- **THEN** they return through the same transactions in the same order

#### Scenario: Where the member is, is a date
- **WHEN** a member is reading any page
- **THEN** they are told the span of dates they are looking at, and are not
  asked to think in page numbers

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
