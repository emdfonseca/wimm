## ADDED Requirements

### Requirement: An account always has an owner
**Story**: S1
Every account wimm holds SHALL have at least one owner at all times. wimm MUST
refuse any change that would leave an account with no owner, and MUST make that
refusal a property of the stored record rather than of the screen that happens
to be asking, so that no route, no later caller and no correction applied
directly to the data can produce an account nobody owns.

An owner SHALL be able to give up ownership of an account whenever another
member owns it, and SHALL be able to make any other member of the household an
owner. Where an owner wants an account out of wimm rather than in somebody
else's hands, that is a separate choice and is covered by *An account can be
left out of wimm* — giving up the last ownership is not the way to ask for it.

Accounts that already have no owner SHALL be brought back rather than left as
they are: each becomes owned by the member who connected its bank, and left out,
so no account stays in a state a member cannot reach.

#### Scenario: The last owner cannot step back
- **WHEN** the only owner of an account tries to stop owning it, and no other
  member owns it
- **THEN** the account stays theirs, they are told that an account has to belong
  to somebody, and they are offered leaving it out of wimm instead

#### Scenario: Stepping back once somebody else owns it
- **WHEN** an owner makes another member an owner of an account and then stops
  owning it themselves
- **THEN** the change is accepted, the other member owns it, and the first
  member no longer sees it

#### Scenario: Two owners, one steps back
- **WHEN** an account is owned by two members and one of them stops owning it
- **THEN** the account stays with the remaining owner and nothing is hidden from
  them

#### Scenario: Making somebody an owner does not remove the other owners
- **WHEN** a member makes a second member an owner of an account
- **THEN** both of them own it, and the first member's ownership is untouched

#### Scenario: An account that had no owner
- **WHEN** a member looks at a bank that had an account nobody owned
- **THEN** that account is listed again, owned by the member who connected the
  bank and marked as left out, with the way to bring it back

### Requirement: An account can be left out of wimm
**Story**: S3
A member who owns an account SHALL be able to leave it out of wimm, and SHALL be
able to bring it back afterwards. Leaving an account out MUST stop wimm reading
anything from it, MUST remove it from every list and every total belonging to
every member who is not an owner, and MUST leave no trace of it for them.

An owner SHALL still see the account, marked as left out, with the way to bring
it back — otherwise the choice is a door that locks behind them. They SHALL NOT
be shown a balance for it, because wimm is no longer reading one and the last
figure it read is not a figure about now.

Nothing already read SHALL be deleted. Leaving an account out and destroying
what wimm read from it are different decisions, and only the first is being
made. Bringing the account back SHALL restore it with the owners and the levels
it had, and the member doing so SHALL be told who will see it again before it
happens.

Leaving an account out is not the same as disconnecting its bank. Disconnecting
ends wimm's access and leaves what was read on screen for the members who could
see it; leaving an account out is a member saying this account is not in wimm,
so it goes from their household's view too.

#### Scenario: Leaving an account out
- **WHEN** an owner leaves an account out of wimm
- **THEN** no balance is read for it again, it disappears for every member who
  does not own it, and it counts towards nobody's total

#### Scenario: What the owner still sees
- **WHEN** an owner looks at an account they have left out
- **THEN** they see it marked as left out, with no balance, and a way to bring
  it back

#### Scenario: Nobody else is told it exists
- **WHEN** a member who was granted a level on an account looks at the household
  while that account is left out
- **THEN** it appears nowhere for them, not in the list, not in their total, and
  not as a count of accounts withheld

#### Scenario: Bringing an account back
- **WHEN** an owner brings a left-out account back
- **THEN** its balance is read again, it reappears with the owners and the
  levels it had before, and it counts towards the totals of everyone who may see
  it

#### Scenario: Being told who will see it again
- **WHEN** an owner is about to bring back an account that another member had
  been granted a level on
- **THEN** they are told which members will see it again, and at what level,
  before the account comes back

#### Scenario: Nothing read is thrown away
- **WHEN** an account is left out and later brought back
- **THEN** everything wimm had already read from it is still there, rather than
  the account starting again as though it were new

#### Scenario: A left-out account at a bank that is reconnected
- **WHEN** a bank is disconnected and reconnected while one of its accounts is
  left out
- **THEN** that account comes back still left out, rather than quietly starting
  to be read again

### Requirement: An account can be given the household's own name
**Story**: S5
An owner SHALL be able to give an account a name of their own choosing, and that
name SHALL be what every member who may see the account is shown. It SHALL
survive the bank being reconnected, restored or read again, because a bank
renaming its products MUST NOT rename a household's accounts.

The bank's own name for the account SHALL still be available to anyone who may
see the account, so the household's name adds to what the bank said rather than
replacing the record of it. Clearing the household's name SHALL return the
account to being shown by the bank's name.

#### Scenario: Naming an account
- **WHEN** an owner gives an account a name
- **THEN** that name is what they and every member who may see the account are
  shown, wherever the account appears

#### Scenario: The name survives a reconnect
- **WHEN** a bank is disconnected and connected again, or its access is restored
- **THEN** each account keeps the name the household gave it

#### Scenario: The bank's name is still there
- **WHEN** a member looks at an account the household has renamed
- **THEN** they can still see what the bank calls it

#### Scenario: Telling apart two accounts the bank names identically
- **WHEN** a bank offers two accounts with the same name and an owner names one
  of them
- **THEN** the two are distinguishable by name alone, without wimm inventing
  anything or revealing any part of an account number

#### Scenario: Clearing the name
- **WHEN** an owner clears the name they gave an account
- **THEN** it is shown by the bank's name again

#### Scenario: A member who was granted a level sees the household's name
- **WHEN** a member who does not own an account has been granted *balance* on it
  and an owner has named it
- **THEN** they see the household's name, because it is the name the household
  uses and it reveals nothing the bank's name did not

## MODIFIED Requirements

### Requirement: Seeing the accounts a member may see
**Story**: S5
The screen a member lands on after signing in SHALL list every account that
member may see, each showing its name, the bank it belongs to and its balance,
and showing enough of its number to tell two accounts at the same bank apart
wherever that member owns it or has been granted *details*.

An account's name is the one the household gave it where one has been given, and
the bank's otherwise.

#### Scenario: The accounts are listed on arrival
- **WHEN** a member who may see at least one account signs in
- **THEN** those accounts are listed on the screen they land on, without them
  navigating anywhere

#### Scenario: Telling two accounts at the same bank apart
- **WHEN** a member owns two accounts at the same bank, or has been granted
  *details* on them
- **THEN** each is shown with enough of its account number to tell which is
  which, and never with the number in full

#### Scenario: Two accounts at one bank seen only as balances
- **WHEN** a member has been granted *balance* on two accounts at the same bank
  that the bank names identically and the household has named neither
- **THEN** both are listed with that name and their own balances, and wimm
  neither shows any part of their numbers nor invents anything to tell them
  apart

#### Scenario: An account the bank gave no name
- **WHEN** a bank makes an account available without a product name
- **THEN** wimm shows it by what the bank did give, and never as a blank row or
  an invented name

#### Scenario: An account holding no money
- **WHEN** an account's balance is zero
- **THEN** it is listed with a zero balance, not hidden

#### Scenario: An account that is overdrawn
- **WHEN** an account's balance is below zero
- **THEN** the balance is shown as negative, carrying a sign, so its direction
  does not depend on colour alone

#### Scenario: An account the owner has left out
- **WHEN** an owner looks at the list and one of their accounts is left out
- **THEN** it is listed among their accounts, marked as left out and without a
  balance, rather than vanishing from the screen

### Requirement: A balance is a reading, not a live figure
**Story**: S3
Every balance SHALL be shown together with when it was read from the bank. wimm
MUST NOT present a balance as current when it has not been re-read, because a
stale number about money that is presented as live is worse than no number.

No balance SHALL be read for an account that is left out, and none SHALL be
shown for one.

#### Scenario: Balances are read when a member arrives
- **WHEN** a member opens the screen listing the accounts they may see
- **THEN** the balances are read from the banks as part of that arrival, so the
  figures a member sees are the ones their banks hold at that moment

#### Scenario: A balance says when it was read
- **WHEN** a member looks at an account's balance
- **THEN** they can see when that balance was read from the bank

#### Scenario: A reading that could not be taken
- **WHEN** a balance could not be read on arrival
- **THEN** the previous reading is shown with the time it was taken and the
  account says it could not be updated, rather than showing a blank or a figure
  presented as current

#### Scenario: A left-out account is not read
- **WHEN** a member arrives and one of the accounts they own is left out
- **THEN** no balance is read for it and none is shown, including the one that
  was read before it was left out

### Requirement: Bringing balances up to date
**Story**: S3
A member SHALL be able to ask wimm to re-read the balances of the accounts they
may see at any time. Where a bank or the service reaching it refuses because wimm has asked too
often, the member SHALL be told when it can be tried again and the balances
already on screen SHALL remain, with their original read times.

A refresh SHALL NOT read an account that is left out.

#### Scenario: Refreshing
- **WHEN** a member asks for the balances they can see to be brought up to date
- **THEN** the balances are re-read and shown with a new read time

#### Scenario: Refusing to be asked again so soon
- **WHEN** a member asks to refresh and the bank or the service reaching it
  refuses because it has been asked too often
- **THEN** the member is told when it can next be tried, and the balances
  already shown stay where they are, still showing when they were read

#### Scenario: One bank fails while others succeed
- **WHEN** a refresh succeeds for some connected banks and fails for another
- **THEN** the balances that were re-read show their new read time, the ones
  that could not be re-read keep their old one, and the member is told which
  bank did not answer

#### Scenario: Nothing to refresh
- **WHEN** a household with no connected banks looks at its accounts
- **THEN** no way to refresh is offered

#### Scenario: Refreshing with an account left out
- **WHEN** a member refreshes and one of the accounts they own is left out
- **THEN** every other account is re-read and that one is not, and the refresh
  does not report it as having failed

### Requirement: The total is the member's own
**Story**: S3
The accounts a member may see SHALL be shown with a total of exactly those, so
two members of one household MAY see different totals and each total MUST agree
with the accounts shown beneath it. Where those accounts are held in more than
one currency, wimm SHALL show a total per currency and MUST NOT add different
currencies together, because this change converts nothing and a converted figure
would be invented.

An account that is left out SHALL count towards no total, including its owners'.

#### Scenario: A total across accounts
- **WHEN** every account a member may see is in one currency
- **THEN** a single total is shown alongside them

#### Scenario: Two members, two totals
- **WHEN** one member owns two accounts and has granted a second member
  *balance* on only one of them
- **THEN** the first member's total covers both accounts, the second member's
  total covers one, and neither is told that the other's differs

#### Scenario: Accounts in more than one currency
- **WHEN** the accounts a member may see are held in more than one currency
- **THEN** a separate total is shown for each currency, and no combined figure
  is shown

#### Scenario: The total agrees with the accounts
- **WHEN** a member reads the total and adds up the listed balances themselves
- **THEN** the two agree, including where balances are negative

#### Scenario: The total counts only what that member may see
- **WHEN** a member has been granted nothing on one of a bank's accounts
- **THEN** that account's balance is absent from their total, and the total is
  not marked as partial, because a total that announces what it omits reveals
  the omission

#### Scenario: A left-out account is outside its owner's total
- **WHEN** an owner leaves an account out
- **THEN** their own total drops by that account's balance, and the listed
  accounts still add up to the total shown
