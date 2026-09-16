## Purpose

Covers what the household sees once a bank is connected: which accounts are
shared and with whom, the balances and total on the screen a member lands on,
how fresh those balances are and how they are brought up to date, and what that
screen says before any bank has been connected at all.

## ADDED Requirements

### Requirement: Shared accounts belong to the household
**Story**: S2
An account a member has chosen to share SHALL be visible to every member of the
household, regardless of which member connected it. An account that has not been
shared SHALL be visible to nobody. wimm SHALL record which member connected each
bank and SHALL show that to the household, because someone has to know whose
consent is holding a connection open and whose will have to renew it.

#### Scenario: Another member sees a newly shared account
- **WHEN** one member connects a bank and shares two of its accounts, and a
  different member of the same household signs in
- **THEN** the second member sees those two accounts and their balances

#### Scenario: An unshared account is visible to nobody
- **WHEN** a member connects a bank and does not share one of its accounts
- **THEN** that account appears nowhere in wimm for any member, including the
  member who connected the bank

#### Scenario: Who connected a bank is visible
- **WHEN** any member looks at a connected bank
- **THEN** they can see which member connected it and when its access ends

### Requirement: Seeing the household's accounts
**Story**: S2
The screen a member lands on after signing in SHALL list every shared account,
each showing its name, the bank it belongs to, enough of its number to tell two
accounts at the same bank apart, and its balance.

#### Scenario: The accounts are listed on arrival
- **WHEN** a member of a household with shared accounts signs in
- **THEN** the accounts are listed on the screen they land on, without them
  navigating anywhere

#### Scenario: Telling two accounts at the same bank apart
- **WHEN** a household shares two accounts at the same bank
- **THEN** each is shown with enough of its account number to tell which is
  which, and never with the number in full

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

### Requirement: A balance is a reading, not a live figure
**Story**: S2
Every balance SHALL be shown together with when it was read from the bank. wimm
MUST NOT present a balance as current when it has not been re-read, because a
stale number about money that is presented as live is worse than no number.

#### Scenario: Balances are read when a member arrives
- **WHEN** a member opens the screen listing the household's accounts
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

### Requirement: Bringing balances up to date
**Story**: S2
A member SHALL be able to ask wimm to re-read the household's balances at any
time. Where a bank or the service reaching it refuses because wimm has asked too
often, the member SHALL be told when it can be tried again and the balances
already on screen SHALL remain, with their original read times.

#### Scenario: Refreshing
- **WHEN** a member asks for the household's balances to be brought up to date
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

### Requirement: The household total
**Story**: S2
The household's shared accounts SHALL be shown with a total. Where accounts are
held in more than one currency, wimm SHALL show a total per currency and MUST
NOT add different currencies together, because this change converts nothing and
a converted figure would be invented.

#### Scenario: A total across accounts
- **WHEN** a household's shared accounts are all in one currency
- **THEN** a single total is shown alongside them

#### Scenario: Accounts in more than one currency
- **WHEN** a household shares accounts in more than one currency
- **THEN** a separate total is shown for each currency, and no combined figure
  is shown

#### Scenario: The total agrees with the accounts
- **WHEN** a member reads the total and adds up the listed balances themselves
- **THEN** the two agree, including where balances are negative

#### Scenario: The total counts only what is shared
- **WHEN** a member has chosen not to share one of a bank's accounts
- **THEN** that account's balance is absent from the total

### Requirement: Before any bank is connected
**Story**: S2
A household that has connected no bank SHALL be shown what the screen is for
and the way to fill it, not an empty list.

#### Scenario: The first member arrives
- **WHEN** a member of a household with no connected banks signs in
- **THEN** they are shown an explanation of what connecting a bank does and a
  way to start connecting one

#### Scenario: No total over nothing
- **WHEN** a household has no shared accounts
- **THEN** no total is shown, rather than a total of zero

### Requirement: Accounts leave with their connection
**Story**: S3
When a bank is disconnected its accounts SHALL stop being shown to every member
of the household, and SHALL stop counting towards the total.

#### Scenario: The accounts go
- **WHEN** a bank is disconnected
- **THEN** its accounts are no longer listed for any member of the household

#### Scenario: The total follows
- **WHEN** a bank is disconnected
- **THEN** the household total no longer includes its accounts

#### Scenario: Disconnecting the only bank
- **WHEN** a household disconnects the only bank it had connected
- **THEN** members see the same explanation and invitation they saw before any
  bank was connected
