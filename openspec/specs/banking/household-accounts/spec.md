# banking/household-accounts Specification

## Purpose
Covers what the household sees on Accounts once a bank is connected: which
accounts are owned by whom and seen by whom at what level, how fresh their
balances are and how they are brought up to date, what Accounts says before
any bank has been connected, and the total those accounts add up to — shown
on Overview.

## Requirements

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

### Requirement: Who owns an account is the household's answer, not the bank's
**Story**: S1
An account owned by more than one member is a joint account, and it is joint
because two members own it in wimm. wimm SHALL treat ownership as the
household's own record and MUST NOT derive it from the name the bank has on the
account, MUST NOT add an owner because a name appears to match a member, and
MUST NOT contradict or overwrite the holder name the bank gave.

The two can disagree, and the disagreement is not an error. An account held
jointly at the bank may have one owner in wimm until somebody says otherwise,
and an account held by one person at the bank may be owned by both members
because that is how the household treats it.

#### Scenario: An account both members own
- **WHEN** two members own one account
- **THEN** each of them sees it in full, including every transaction read from
  it, and neither sees any mark saying the other is looking at the same rows

#### Scenario: The bank's holder name is left alone
- **WHEN** an account's holder name at the bank names one person and both
  members own it in wimm
- **THEN** wimm shows the holder name exactly as the bank gave it, and shows
  both owners, without presenting either as a correction of the other

#### Scenario: Ownership is never guessed
- **WHEN** a bank returns an account whose holder name matches a member of the
  household
- **THEN** no owner is added because of that name, and ownership stays with the
  member who connected the bank until somebody changes it

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

### Requirement: Household money and a member's own money are separate figures
**Story**: S5
Wherever wimm shows a total, it SHALL show two, and never one figure that mixes
them.

**Household money** is the sum of the accounts that every member of the
household either owns or has been given *details* on. It is the same figure
for every member, and it reveals nothing, because every member already sees
each of those accounts in full. A household of one member has no household
money: there is nobody to hold it with.

**A member's own money** is the sum of the accounts that member owns which are
not household money.

The two are disjoint: an account counts in one of them or in neither, never in
both. An account a member may see that is neither — one shared with them but
not with everyone, or shared with everyone at *balance* only — SHALL be listed
with its balance and SHALL count in no figure.

Each figure SHALL agree with the accounts listed under it. Where accounts are
held in more than one currency, each figure SHALL be shown per currency and
currencies MUST NOT be added together. A figure over no accounts SHALL NOT be
shown, rather than shown as zero. An account that is left out SHALL count in
no figure, including its owners'. No figure SHALL say what it omits.

#### Scenario: A joint account is household money
- **WHEN** both members of a household own an account
- **THEN** its balance is in household money for both of them, and in neither
  one's own money

#### Scenario: An account shared in full with everyone
- **WHEN** a member owns an account and has given every other member *details*
  on it
- **THEN** it counts as household money for every member, its owner included

#### Scenario: A personal account is the member's own
- **WHEN** a member owns an account nobody else has been given anything on
- **THEN** its balance is in that member's own money, and no other member sees
  any figure change because of it

#### Scenario: Every member sees the same household money
- **WHEN** two members of one household look at household money
- **THEN** they see the same figure

#### Scenario: A household of one
- **WHEN** the only member of a household owns three accounts
- **THEN** all three are their own money, and no household money figure is
  shown

#### Scenario: Shared at balance only
- **WHEN** a member has been given *balance* on an account they do not own
- **THEN** the account is listed for them with its balance, and it is in
  neither of their figures

#### Scenario: Shared with everyone, but only the balance
- **WHEN** an owner has given every other member *balance* on an account, and
  nobody *details*
- **THEN** the account is the owner's own money, and for everyone else it is
  listed and counted in no figure

#### Scenario: Shared in full with one member of three
- **WHEN** a household has three members and an owner gives *details* on an
  account to one of them only
- **THEN** the account is not household money for anybody

#### Scenario: Somebody joins the household
- **WHEN** a new member joins and has been given nothing yet
- **THEN** accounts stop being household money until that member owns them or
  is given *details* on them, and each returns to its owners' own money in the
  meantime

#### Scenario: The figures agree with the accounts
- **WHEN** a member adds up the balances listed under household money, and
  those listed as their own
- **THEN** each sum equals the figure shown for it, including where balances
  are negative

#### Scenario: Nothing in a group
- **WHEN** no account is household money for a member
- **THEN** no household money figure is shown, rather than one of zero

#### Scenario: Accounts in more than one currency
- **WHEN** a member's own accounts are held in two currencies
- **THEN** their own money is shown once per currency, and no combined figure
  is shown

#### Scenario: A left-out account is in no figure
- **WHEN** an owner leaves an account out
- **THEN** whichever figure it was in drops by its balance, and the accounts
  listed still add up to the figures shown

#### Scenario: A figure does not say what it leaves out
- **WHEN** a member has been granted nothing on one of a bank's accounts
- **THEN** no figure they see mentions it, because a figure that announces
  what it omits reveals the omission
