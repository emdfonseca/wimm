## ADDED Requirements

### Requirement: A bank connected before wimm could read transactions
**Story**: S2
A connection whose consent covers balances but not transactions SHALL be shown
as such wherever a member is looking for that bank's transactions, and SHALL
offer the way to widen it. It MUST NOT be presented as broken, expired or
failing: it is working, and it is reading everything it was ever granted.

Widening SHALL be done by confirming once more at the bank, without
disconnecting first and without choosing the bank from the list again. Every
account SHALL keep the owners and the levels it had.

#### Scenario: A bank that cannot be read for transactions says so
- **WHEN** a member looks for the transactions of an account at a bank
  connected before wimm could read them
- **THEN** they are told that this bank's transactions are not being read yet
  and that it can be changed, and the bank is not described as broken or out of
  date

#### Scenario: Widening what a bank was granted
- **WHEN** a member chooses to widen a connection and confirms at their bank
- **THEN** every account keeps the owners and the levels it had, that bank's
  transactions begin to be read, and the member is not asked to pick the bank
  or choose accounts again

#### Scenario: Declining while widening
- **WHEN** a member declines at the bank while widening a connection
- **THEN** the connection is left exactly as it was, still working for
  balances, and still offering to be widened

#### Scenario: Balances are unaffected while a bank stays narrow
- **WHEN** a member never widens a connection
- **THEN** its balances keep being read and shown as before, and only its
  transactions are absent

#### Scenario: A bank connected after transactions could be read
- **WHEN** a member connects a bank for the first time
- **THEN** its transactions are read from the start and the member is never
  shown anything about widening it

## MODIFIED Requirements

### Requirement: Consenting at the bank
**Story**: S1
Consent SHALL be given at the member's own bank and never inside wimm. wimm
MUST NOT ask a member for their banking credentials, MUST ask the bank only for
account details, balances and transactions, and MUST tell the member what it
will be able to read before sending them there. wimm SHALL also tell the member
how long the bank's access will last, using the limit that bank actually sets.

#### Scenario: Handing off to the bank
- **WHEN** a member confirms the bank they picked
- **THEN** they are told that wimm will be able to read that bank's account
  names, balances and transactions, and are then taken to the bank's own site
  to confirm

#### Scenario: How long access will last
- **WHEN** a member is about to be sent to their bank
- **THEN** they are told the date the bank's access will run out, taken from
  that bank's own limit rather than from a figure wimm has chosen

#### Scenario: wimm never asks for banking credentials
- **WHEN** a member connects a bank
- **THEN** at no point in wimm are they asked for their bank username,
  password, card number or any other banking credential

#### Scenario: The member abandons the hand-off
- **WHEN** a member is taken to their bank and closes the window without
  finishing
- **THEN** no bank is connected, nothing appears in the household's accounts,
  and starting again from the list works normally

### Requirement: Disconnecting a bank
**Story**: S3
A member SHALL be able to disconnect any bank connected to their household.
Disconnecting SHALL be confirmed first, naming the bank and how many accounts
will go, because it is money-bearing. wimm SHALL end its own access even when
it cannot tell the bank.

Disconnecting ends wimm's access and MUST NOT destroy what that access already
read. The confirmation SHALL say so, because a member deciding whether to
disconnect is weighing exactly that.

#### Scenario: Confirming before disconnecting
- **WHEN** a member chooses to disconnect a bank
- **THEN** they are asked to confirm, and the question names the bank, how many
  of the household's accounts will disappear, and that the transactions already
  read are kept

#### Scenario: Disconnecting
- **WHEN** a member confirms the disconnection
- **THEN** that bank's accounts are no longer shown to anyone in the household

#### Scenario: Changing their mind
- **WHEN** a member dismisses the confirmation without confirming
- **THEN** nothing is disconnected and the accounts remain

#### Scenario: The bank cannot be told
- **WHEN** a member disconnects a bank and wimm cannot reach that bank to
  withdraw its access
- **THEN** the bank is still disconnected in wimm and its accounts still go,
  and the member is told that access may need to be withdrawn at the bank as
  well

#### Scenario: Reconnecting afterwards
- **WHEN** a member disconnects a bank and later connects the same bank again
- **THEN** it connects normally, its accounts appear once and not twice, and
  each one carries the owners, the levels and the transactions it had before
