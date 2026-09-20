## ADDED Requirements

### Requirement: A transaction is named by who it was with
**Story**: S3
Wherever wimm lists a transaction, it SHALL name it by the merchant or person
it was with, not by the bank's statement line. Where the bank supplies only a
statement line, wimm SHALL remove the bank's own bookkeeping from it: the
word for the kind of transaction at its start, reference numbers at its end,
and any trailer about the country or the original amount. Two payments to one
merchant that differ only in that bookkeeping SHALL carry the same name.

A name MUST NOT be empty. Where removing the bookkeeping would leave nothing,
the line SHALL be shown as the bank wrote it, and where the bank wrote
nothing, the transaction SHALL be named for what it is.

wimm MUST NOT discard what the bank wrote. On Transactions, where the name
differs from the bank's line, the bank's line SHALL be shown with the
transaction, because a name wimm derived has to be checkable against the
statement it came from.

#### Scenario: A card payment with a reference number
- **WHEN** the bank's line for a payment is
  `COMPRA WWW.AMAZON NM4HU1VZ4 230002268264350`
- **THEN** the transaction is listed as `Amazon`

#### Scenario: Two payments to one merchant
- **WHEN** a member has paid the same merchant twice and the bank's two lines
  differ only in their reference numbers
- **THEN** both transactions carry the same name

#### Scenario: A trailer about where and how much
- **WHEN** the bank's line ends with the country and the original amount of a
  payment made abroad
- **THEN** the name stops before that trailer

#### Scenario: The bank names the other party
- **WHEN** the bank supplies the name of the person or company on the other
  side of a transaction
- **THEN** the transaction is named by it

#### Scenario: Nothing would be left
- **WHEN** a bank's line consists only of a transaction word and a number
- **THEN** the transaction is listed under the line as the bank wrote it,
  never under an empty name

#### Scenario: The bank wrote nothing
- **WHEN** a transaction arrives with no other party and no statement line
- **THEN** it is listed as a card payment

#### Scenario: Checking a name against the statement
- **WHEN** a member looks at a transaction on Transactions whose name differs
  from the bank's line
- **THEN** the bank's line is shown with it, exactly as the bank wrote it

#### Scenario: The same name everywhere
- **WHEN** a member sees a transaction on Overview and then on Transactions
- **THEN** it carries the same name in both places
