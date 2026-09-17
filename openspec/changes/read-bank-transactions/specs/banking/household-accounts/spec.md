## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Accounts leave with their connection
**Story**: S3
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
- **THEN** members see the same explanation and invitation they saw before any
  bank was connected

#### Scenario: The account comes back as itself
- **WHEN** a member reconnects a bank they had disconnected
- **THEN** each account it offers again appears once, with the owners and the
  levels it had before, rather than as a new account nobody owns
