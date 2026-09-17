## Purpose

Covers what the household sees once a bank is connected: which accounts are
owned by whom and seen by whom at what level, the balances and total on the
screen a member lands on,
how fresh those balances are and how they are brought up to date, and what that
screen says before any bank has been connected at all.

## ADDED Requirements

### Requirement: An account is seen by its owners, and by whoever they grant
**Story**: S2
An account SHALL be visible in full to every member who owns it, and an account
MAY have more than one owner. Every other member SHALL see it only at the level
an owner has granted them, and SHALL see nothing of it by default. There are
three levels and no others:

- **hidden** — the account leaves no trace anywhere that member can see. This is
  the absence of a grant, and it is what every member who is not an owner starts
  with.
- **balance** — the bank, the account's name, and its balance with its read time.
- **details** — everything *balance* shows, plus enough of the account number to
  tell two accounts at one bank apart, the account type, and the holder name.

wimm SHALL record which member connected each bank and SHALL show that to the
household, because someone has to know whose consent is holding a connection
open and whose will have to renew it.

#### Scenario: An owner sees their own account in full
- **WHEN** a member who owns an account looks at it
- **THEN** they see its balance and its identifying details, whatever levels
  they have granted anyone else

#### Scenario: A joint account has two owners
- **WHEN** an account is owned by two members and a third member has been
  granted nothing
- **THEN** both owners see it in full and the third member sees no trace of it

#### Scenario: A member granted the balance sees the amount and not the number
- **WHEN** a member who does not own an account has been granted *balance* on it
- **THEN** they see the bank, the account's name and its balance, and they are
  shown no part of its account number, its type or its holder name

#### Scenario: A member granted the details sees the identifiers too
- **WHEN** a member who does not own an account has been granted *details* on it
- **THEN** they additionally see enough of its account number to tell it from
  another account at the same bank, its type, and its holder name

#### Scenario: A member granted nothing sees no trace
- **WHEN** a member has been granted no level on an account they do not own
- **THEN** that account appears nowhere for them: not in the list, not in any
  total, and not as a count of accounts withheld

#### Scenario: Two members see the same household differently
- **WHEN** one member owns three accounts and has granted a second member
  *balance* on one of them
- **THEN** the first member sees three accounts and the second sees one, and
  neither is told what the other sees

#### Scenario: Who connected a bank is visible
- **WHEN** any member looks at a bank whose accounts they can see
- **THEN** they can see which member connected it and when its access ends

### Requirement: Seeing the accounts a member may see
**Story**: S2
The screen a member lands on after signing in SHALL list every account that
member may see, each showing its name, the bank it belongs to and its balance,
and showing enough of its number to tell two accounts at the same bank apart
wherever that member owns it or has been granted *details*.

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
  that the bank names identically
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

### Requirement: A balance is a reading, not a live figure
**Story**: S2
Every balance SHALL be shown together with when it was read from the bank. wimm
MUST NOT present a balance as current when it has not been re-read, because a
stale number about money that is presented as live is worse than no number.

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

### Requirement: Bringing balances up to date
**Story**: S2
A member SHALL be able to ask wimm to re-read the balances of the accounts they
may see at any time. Where a bank or the service reaching it refuses because wimm has asked too
often, the member SHALL be told when it can be tried again and the balances
already on screen SHALL remain, with their original read times.

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

### Requirement: The total is the member's own
**Story**: S2
The accounts a member may see SHALL be shown with a total of exactly those, so
two members of one household MAY see different totals and each total MUST agree
with the accounts shown beneath it. Where those accounts are held in more than
one currency, wimm SHALL show a total per currency and MUST NOT add different
currencies together, because this change converts nothing and a converted figure
would be invented.

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

### Requirement: Before any bank is connected
**Story**: S2
A household that has connected no bank SHALL be shown what the screen is for
and the way to fill it, not an empty list.

#### Scenario: The first member arrives
- **WHEN** a member of a household with no connected banks signs in
- **THEN** they are shown an explanation of what connecting a bank does and a
  way to start connecting one

#### Scenario: No total over nothing
- **WHEN** a member may see no account
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
- **THEN** no member's total includes its accounts any longer

#### Scenario: Disconnecting the only bank
- **WHEN** a household disconnects the only bank it had connected
- **THEN** members see the same explanation and invitation they saw before any
  bank was connected
