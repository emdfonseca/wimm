## ADDED Requirements

### Requirement: Searching the ledger
**Story**: S2
A member SHALL be able to search their ledger with free text. A transaction SHALL
match when the text appears anywhere within who the bank says it was with, or
anywhere within the bank's line, ignoring the difference between upper and lower
case. Accents SHALL be matched as written. Spaces at either end of the text SHALL
be ignored, and a search of nothing but spaces SHALL be no search at all. Every
character SHALL match itself, including characters a pattern would otherwise
treat as a wildcard. The search SHALL apply as the member types, once they
pause, without their having to confirm it, and going Back from a search SHALL
return to the list as it was before the search rather than to each word
typed on the way.

Search reads what the bank wrote. A name wimm gives a transaction the bank wrote
nothing for is not searched, because it is not the bank's.

#### Scenario: Finding a merchant
- **WHEN** a member searches for `galp`
- **THEN** only the transactions whose other party or bank's line contains
  `GALP`, `Galp` or `galp` are listed

#### Scenario: Found by the bank's line
- **WHEN** a member searches for part of a reference number that appears only in
  one transaction's bank's line
- **THEN** that transaction is listed

#### Scenario: Accents count
- **WHEN** a member searches for `cafe` and the bank wrote `Café Central`
- **THEN** that transaction is not listed, and searching for `café` lists it

#### Scenario: A percent sign is a percent sign
- **WHEN** a member searches for `50%`
- **THEN** only transactions whose other party or bank's line contains `50%` are
  listed, not every one containing `50`

#### Scenario: Searching as you type
- **WHEN** a member types `galp` and pauses, without pressing Enter
- **THEN** only the transactions matching `galp` are listed, and the search
  field still holds `galp` with the cursor in it

#### Scenario: Back from a typed search
- **WHEN** a member types `plumber` letter by letter, then goes Back
- **THEN** the list is as it was before they started typing

#### Scenario: Clearing the search
- **WHEN** a member empties the search
- **THEN** every transaction matching the other filters in force is listed
  again, from the newest

#### Scenario: A name wimm gave
- **WHEN** a transaction the bank sent with no other party and no line is listed
  as a card payment, and a member searches for `card payment`
- **THEN** that transaction is not listed, because nothing the bank wrote
  contains those words

### Requirement: Money in or money out
**Story**: S4
A member SHALL be able to show only money in, only money out, or both. Money in
is a transaction whose amount carries a plus sign, and money out one whose amount
carries a minus sign. Both SHALL be shown until the member chooses otherwise. A
transaction the bank has not settled SHALL be judged by its sign like any other.

Choosing a direction SHALL NOT change any row's label: a row that is half of a
transfer between the member's own accounts, or unusual, reads the same as it
does with both directions shown.

#### Scenario: Only money in
- **WHEN** a member chooses money in
- **THEN** every transaction listed carries a plus sign, and the number of
  transactions says how many came in

#### Scenario: Only money out
- **WHEN** a member chooses money out
- **THEN** every transaction listed carries a minus sign

#### Scenario: Both again
- **WHEN** a member who chose money out chooses both
- **THEN** money in and money out are listed together again, from the newest

#### Scenario: One half of a transfer
- **WHEN** a member who moved €500.00 from current to savings chooses money in
- **THEN** the row arriving in savings is listed and still reads as between
  their accounts, and the row leaving current is not listed

#### Scenario: A payment not yet settled
- **WHEN** a card payment the bank has not settled is on the ledger and the
  member chooses money out
- **THEN** it is listed, marked as not yet settled

### Requirement: Filters combine and stay in the address
**Story**: S1
The account, the search, the month and the direction SHALL combine: a
transaction is listed only when it matches every filter in force. The filters
in force SHALL be part of the page's address, so that every way of moving
through the ledger keeps them: older, newer, straight to the newest, straight to
the oldest, and any page offered for jumping. Back SHALL return to the filters
that were in force before, and opening the same address again SHALL show the
same filtered list.

Changing a filter SHALL start reading again at the newest transactions that
match. The number of transactions and the pages offered for jumping SHALL
describe the filtered list, never the whole ledger. A member SHALL be able to
remove every filter in one action.

A filtered list is the same screen as the whole ledger: it states how current
the ledger is, how far back it reaches, and which banks are not answering,
exactly as the whole ledger does, and it can be brought up to date the same way.

Filtering MUST NOT widen what a member may see. Every filter narrows the
transactions of the accounts the member owns and has not left out, and nothing
else is ever searched, counted or offered.

A filter value in the address that wimm cannot use SHALL be dropped from the
address, and the list shown without it.

#### Scenario: Two filters at once
- **WHEN** a member chooses their current account and searches for `galp`
- **THEN** only Galp transactions from the current account are listed

#### Scenario: The count follows the filters
- **WHEN** a member whose ledger holds 384 transactions searches for something
  12 of them match
- **THEN** the screen says 12 transactions

#### Scenario: Paging keeps the filters
- **WHEN** a member has chosen money out and pages older, then jumps to the
  oldest page
- **THEN** every page they reach lists money out only, and the filter is still
  shown as chosen

#### Scenario: The pages offered follow the filters
- **WHEN** a member has a search in force and jumps to a page offered by the
  page scrubber
- **THEN** they land on that page of the filtered list, holding the dates the
  scrubber named

#### Scenario: Changing a filter starts at the newest
- **WHEN** a member deep in the ledger changes the search
- **THEN** they are shown the newest transactions that match the new search

#### Scenario: Back undoes a filter
- **WHEN** a member chooses an account and then goes Back
- **THEN** the list is as it was before the account was chosen

#### Scenario: Returning to a filtered view
- **WHEN** a member opens the address of a filtered list they saved
- **THEN** the same filters are in force and shown as chosen

#### Scenario: Clearing every filter
- **WHEN** a member with an account, a search and a month in force clears the
  filters
- **THEN** every transaction they may see is listed from the newest, in one
  action, without leaving Transactions

#### Scenario: Nothing is found that the member may not see
- **WHEN** a member searches for a merchant that appears only on an account
  they were granted *details* on and do not own
- **THEN** nothing is listed for it, and nothing tells them that a match was
  withheld

#### Scenario: A left-out account is not searched
- **WHEN** a member searches for a merchant that appears only on an account
  they have left out
- **THEN** nothing is listed for it

#### Scenario: The oldest page of a filtered list
- **WHEN** a member with a filter in force reaches the oldest matching
  transaction
- **THEN** they are told nothing older matches the filters, and are not told
  that the bank would go no further back

#### Scenario: A filtered list says how current it is
- **WHEN** a member has a filter in force
- **THEN** the screen still says when the ledger was last brought up to date
  and how far back it reaches, and still names any bank that did not answer

#### Scenario: Bringing a filtered list up to date
- **WHEN** a member with a filter in force asks for the list to be brought up
  to date
- **THEN** the filters stay in force, and new transactions that match them
  appear

#### Scenario: Marks do not depend on the filters
- **WHEN** a member searches for the garage that took an unusual €1,650.00 in
  March
- **THEN** that row still reads Unusual, as it does in the whole ledger

#### Scenario: A filter the address gets wrong
- **WHEN** an address names a month that does not exist, such as `2026-13`, or
  a direction other than money in or money out
- **THEN** that filter is dropped from the address and the list is shown
  without it, with any other filter still in force

### Requirement: What a filtered list adds up to
**Story**: S5
While a filter is in force and transactions match it, wimm SHALL show the money
in and the money out across every matching transaction, not only the page on
screen, once per currency and never summed across currencies. Where the
direction is money in, only money in SHALL be shown, and where it is money out,
only money out.

The figures SHALL count the way Overview counts a month: only transactions the
bank has settled, and a transfer between two of the member's accounts left out
when both of those accounts are in the list's scope. When the list is narrowed
to one account, a transfer to another of the member's accounts is money that
left it and SHALL be counted. wimm SHALL say how many transfers it left out and
how many unsettled transactions it did not count, so the figures can be checked
against the rows.

No figures SHALL be shown for the whole ledger with no filter in force.

#### Scenario: A search adds up
- **WHEN** a member searches for `galp`
- **THEN** they are shown the money in and the money out across every Galp
  transaction they may see, however many pages those fill

#### Scenario: The same month as Overview
- **WHEN** a member follows August from Overview's month summary
- **THEN** the money in and money out shown for August on Transactions are the
  figures Overview gave August

#### Scenario: A transfer between their accounts
- **WHEN** a member who moved €500.00 from current to savings in August opens
  August
- **THEN** the €500.00 is in neither figure, and they are told one transfer
  between their accounts was left out

#### Scenario: One account, and a transfer out of it
- **WHEN** the same member narrows August to their current account
- **THEN** the €500.00 is counted as money out

#### Scenario: A payment not yet settled
- **WHEN** a matching card payment has not been settled by the bank
- **THEN** it is listed and not counted, and they are told one payment not yet
  settled was not counted

#### Scenario: Two currencies
- **WHEN** a month holds euro and dollar transactions
- **THEN** the figures are shown once for euros and once for dollars, and never
  added together

#### Scenario: One direction
- **WHEN** a member chooses money in
- **THEN** only money in is shown

#### Scenario: No filter
- **WHEN** no filter is in force
- **THEN** no figures are shown

### Requirement: When nothing matches the filters
**Story**: S2
When the filters in force match no transaction, wimm SHALL say that nothing
matches and offer one action that clears every filter. The filters SHALL stay on
screen, as chosen, so the member can change one instead. This state MUST NOT
read as having no bank connected, as owning no account, or as nothing having
been read, because each of those asks the member to do something different.

Where the account is the only filter in force, wimm SHALL say that the account
has no transactions.

#### Scenario: A search that finds nothing
- **WHEN** a member searches for something no transaction they may see contains
- **THEN** they are told nothing matches the filters, the search they typed is
  still in the search field, and they are offered a way to clear the filters

#### Scenario: Clearing from nothing
- **WHEN** a member told nothing matches clears the filters
- **THEN** every transaction they may see is listed from the newest

#### Scenario: Not mistaken for an empty ledger
- **WHEN** a member whose ledger holds transactions is told nothing matches
- **THEN** nothing on screen offers to connect a bank or says that nothing has
  been read, and the ledger can still be brought up to date

#### Scenario: An account with nothing in it
- **WHEN** a member chooses an account the bank returned no transactions for,
  and no other filter is in force
- **THEN** they are told that account has no transactions, and offered a way
  back to every account

## MODIFIED Requirements

### Requirement: The transactions of one account
**Story**: S1
A member SHALL be able to narrow the list to a single account they own, and to
widen it again, from a control on the Transactions screen. The control SHALL
offer every account the member owns and has not left out, and no other account,
and SHALL narrow to one account at a time. Narrowing SHALL name the account it is
narrowed to and MUST NOT present itself as a different screen.

An address naming an account the member may not narrow to SHALL show every
account they may see, and MUST NOT say whether that account exists.

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

#### Scenario: The accounts offered
- **WHEN** a member owns a current account and a joint account, has left out a
  savings account, and was granted *details* on their partner's account
- **THEN** the current account and the joint account are offered, and neither
  the savings account nor the partner's account is

#### Scenario: Only one account at a time
- **WHEN** a member has narrowed to one account and chooses another
- **THEN** the list shows the second account alone, not both

#### Scenario: An account named in the address that is not theirs
- **WHEN** a member opens an address naming an account they do not own, have
  left out, or that does not exist
- **THEN** they see every account they may see, and the screen reads the same
  in all three cases

### Requirement: Going straight to a month
**Story**: S3
A member SHALL be able to choose a month and see only that month's
transactions. wimm SHALL offer the months that actually hold transactions
matching the member's other filters, newest first, and MUST NOT offer a month
with nothing in it: an empty month is a control that does nothing, and the
reason the list is built from what is there rather than from a calendar. A month
in force SHALL always be shown as chosen, even where the other filters leave
nothing in it, so the member can see it and remove it.

Choosing a month SHALL list only the transactions booked in that month, newest
first, read a page at a time like the rest of the ledger. Paging SHALL stay
within the month: its oldest page is the oldest transaction in it. Removing the
month SHALL widen the list again without leaving Transactions.

A link to a month from elsewhere in wimm SHALL open Transactions with that month
chosen.

The months offered SHALL follow the other filters: where a member is looking at
one account, the months are that account's, and where they have searched, the
months are those holding a match.

#### Scenario: Jumping to a month
- **WHEN** a member chooses April from the months offered
- **THEN** only April's transactions are listed, newest first, the number of
  transactions says how many April holds, and the dates on screen say so

#### Scenario: Paging on from a month
- **WHEN** a member has chosen a month and reaches the oldest transaction in it
- **THEN** they are told nothing older matches, and the month before is not
  listed

#### Scenario: A month holding more than one page
- **WHEN** a member chooses a month with more transactions than fit on a page
- **THEN** they land on that month's newest page and can page through the rest
  of it, with nothing repeated and nothing skipped

#### Scenario: Removing the month
- **WHEN** a member removes the month they chose
- **THEN** every transaction matching the other filters is listed again, from
  the newest

#### Scenario: Arriving from Overview
- **WHEN** a member follows August from Overview's month summary
- **THEN** Transactions opens with August chosen and lists only August's
  transactions

#### Scenario: Only months that hold something
- **WHEN** a household read nothing in March and several things in February
- **THEN** March is not offered and February is

#### Scenario: The months follow the account being looked at
- **WHEN** a member narrows the ledger to one account
- **THEN** the months offered are the months that account has transactions in

#### Scenario: The months follow the search
- **WHEN** a member searches for a plumber paid only in March and June
- **THEN** only March and June are offered

#### Scenario: A chosen month with nothing left in it
- **WHEN** a member has chosen September and then chooses an account with no
  transactions in September
- **THEN** September is still shown as chosen, and they are told nothing
  matches the filters

#### Scenario: Not enough ledger to need it
- **WHEN** every transaction a member can see falls in one month, and no month
  is chosen
- **THEN** no months are offered, because there is nowhere else to go

#### Scenario: A member who owns no account
- **WHEN** a member owns no account
- **THEN** no months are offered, as no transactions are
