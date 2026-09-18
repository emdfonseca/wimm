## ADDED Requirements

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

## MODIFIED Requirements

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
