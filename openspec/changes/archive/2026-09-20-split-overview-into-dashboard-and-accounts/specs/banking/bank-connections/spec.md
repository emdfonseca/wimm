## MODIFIED Requirements

### Requirement: Access that has run out
**Story**: S2
A bank's grant SHALL be treated as expired both when the date it set is reached
and when the bank rejects wimm's access before then. An expired connection SHALL
say so where a member is looking at the balances it stopped updating, and SHALL
offer the way to restore it. wimm MUST NOT present a balance from an expired
connection as current, and MUST NOT remove it either.

#### Scenario: A connection reaches the end of its access
- **WHEN** a member opens Accounts and a connection's access has run out
- **THEN** that bank's accounts say they have stopped updating, keep showing
  their last readings and the date each was taken, and offer a way to restore
  access

#### Scenario: The bank rejects access before the date
- **WHEN** the bank refuses wimm's access before the date it originally set
- **THEN** the connection is treated exactly as if it had reached its end date,
  with the same wording and the same way to restore it

#### Scenario: A household with one working bank and one that has stopped
- **WHEN** a household has one expired connection and one working one
- **THEN** the member is told which bank has stopped updating and that the total
  includes readings taken from it, rather than a single figure presented as
  current with nothing said

### Requirement: Restoring access to a bank
**Story**: S2
A member SHALL be able to restore an expired connection by confirming again at
the bank, without disconnecting first and without choosing the bank again.
Every account SHALL keep the owners and the levels it had before, even though
the bank issues new identifiers each time.

#### Scenario: Restoring access
- **WHEN** a member chooses to restore an expired connection and confirms at
  their bank
- **THEN** every account keeps the owners and the levels it had before, their
  balances are read again, and Accounts stops saying the bank has stopped
  updating

#### Scenario: The bank now offers an account it did not before
- **WHEN** a bank makes an account available on restoring that it did not make
  available before
- **THEN** it belongs to the member who restored the connection, no other member
  has any level on it, and that member is told there is something new to choose

#### Scenario: The bank no longer offers an account that members could see
- **WHEN** an account members could see is not among those the bank makes
  available on restoring
- **THEN** it stops appearing on Accounts for everyone who could see it, its
  owners and levels go with it, and the member restoring is told which account
  the bank no longer offers

#### Scenario: Restoring is refused at the bank
- **WHEN** a member declines at the bank while restoring
- **THEN** the connection stays exactly as it was, still expired, still showing
  its last readings, and still offering to restore
