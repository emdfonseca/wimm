## ADDED Requirements

### Requirement: Accounts is reachable from the primary navigation
**Story**: S2
wimm SHALL have an Accounts screen, reachable from the primary navigation at
every size, that holds the account list and every connection action —
connecting, restoring, disconnecting, and choosing who sees an account.

#### Scenario: Reaching Accounts
- **WHEN** a member uses the primary navigation at any size
- **THEN** Accounts is one of the destinations, and opening it shows the
  household's accounts

### Requirement: Before any bank is connected, Accounts says so
**Story**: S2
A household that has connected no bank SHALL be shown, on Accounts, what the
screen is for and the way to fill it, not an empty list. Accounts SHALL NOT
show a total — the total moved to Overview, and `banking/overview` covers
what a member sees there before any bank is connected.

#### Scenario: The first member arrives
- **WHEN** a member of a household with no connected banks opens Accounts
- **THEN** they are shown an explanation of what connecting a bank does and a
  way to start connecting one, and no total

## REMOVED Requirements

### Requirement: Before any bank is connected
**Reason**: Split by screen. Accounts has its own version (see ADDED:
"Before any bank is connected, Accounts says so"), and Overview has its own
(see `banking/overview`: "Before any bank is connected, Overview says so"),
because the two screens no longer show the same things before a bank exists
— Accounts shows no total, and Overview shows no account list.
**Migration**: No data migration. A screen reading this requirement's old
text follows the split above instead.

A household that has connected no bank SHALL be shown what the screen is for
and the way to fill it, not an empty list.

#### Scenario: The first member arrives
- **WHEN** a member of a household with no connected banks signs in
- **THEN** they are shown an explanation of what connecting a bank does and a
  way to start connecting one

#### Scenario: No total over nothing
- **WHEN** a member may see no account
- **THEN** no total is shown, rather than a total of zero

## MODIFIED Requirements

### Requirement: Seeing the accounts a member may see
**Story**: S2
The Accounts screen SHALL list every account that member may see, each
showing its name, the bank it belongs to and its balance, and showing enough
of its number to tell two accounts at the same bank apart wherever that
member owns it or has been granted *details*.

An account's name is the one the household gave it where one has been given, and
the bank's otherwise.

#### Scenario: The accounts are listed on arrival
- **WHEN** a member who may see at least one account opens Accounts
- **THEN** those accounts are listed, without them navigating anywhere else

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
**Story**: S1
Every balance SHALL be shown together with when it was read from the bank. wimm
MUST NOT present a balance as current when it has not been re-read, because a
stale number about money that is presented as live is worse than no number.

No balance SHALL be read for an account that is left out, and none SHALL be
shown for one.

#### Scenario: Balances are read when a member arrives
- **WHEN** a member signs in
- **THEN** the balances behind their Overview total are read from the banks
  as part of that arrival, so the figures a member sees are the ones their
  banks hold at that moment

#### Scenario: A balance says when it was read
- **WHEN** a member looks at an account's balance, on Accounts
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

### Requirement: Accounts leave with their connection
**Story**: S2
When a bank is disconnected its accounts SHALL stop being shown to every member
of the household, and SHALL stop counting towards the total.

They SHALL NOT be erased. What wimm already read from an account is kept, and
reconnecting the same bank SHALL find the same account rather than create a
second one beside it. Ending wimm's access to a bank and destroying the record
of what it read are different decisions, and only the first is being made here.

#### Scenario: The accounts go
- **WHEN** a bank is disconnected
- **THEN** its accounts are no longer listed for any member of the household

#### Scenario: The total follows
- **WHEN** a bank is disconnected
- **THEN** no member's total includes its accounts any longer

#### Scenario: Disconnecting the only bank
- **WHEN** a household disconnects the only bank it had connected
- **THEN** Accounts shows the same explanation and invitation it showed
  before any bank was connected, and Overview shows its own before-any-bank
  explanation too

#### Scenario: The account comes back as itself
- **WHEN** a member reconnects a bank they had disconnected
- **THEN** each account it offers again appears once, with the owners and the
  levels it had before, rather than as a new account nobody owns
