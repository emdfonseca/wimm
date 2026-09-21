## ADDED Requirements

### Requirement: A transfer between own accounts is labelled in the ledger
**Story**: S1
Transactions SHALL label each of the two rows of a transfer between the
member's own accounts, as `banking/own-transfers` defines, as between their
accounts. The label SHALL be words and never colour alone, SHALL sit where a
row's status is shown, and SHALL be worked out across every account the member
owns, whatever scope Overview is reading in, so a row's label does not depend
on a control on another screen. A row so labelled SHALL never also be labelled
unusual.

Nothing about being a transfer SHALL remove a row from the list, change its
amount or its sign, or change how the list is ordered, paged or counted. A row
not yet settled SHALL read as not settled and nothing else.

#### Scenario: Both halves are labelled
- **WHEN** a member who moved €500.00 from their current account to their
  savings account reads Transactions
- **THEN** the row leaving the current account and the row arriving in savings
  both read as between their accounts

#### Scenario: Reading one account
- **WHEN** the member reads the savings account alone
- **THEN** the arriving row still reads as between their accounts, though its
  other half is not in the list

#### Scenario: A large transfer
- **WHEN** the transfer is far above anything the member usually pays
- **THEN** its rows read as between their accounts and not as unusual

#### Scenario: One that could not be paired
- **WHEN** a member sent €1,000.00 to an account wimm does not hold
- **THEN** the row carries no such label

#### Scenario: The other owner of a joint account
- **WHEN** a member reads a joint account into which their partner paid from an
  account only the partner owns
- **THEN** the arriving row carries no such label

#### Scenario: The count does not change
- **WHEN** a member's ledger holds forty transfers
- **THEN** the list holds every row it held before, and says the same number of
  transactions
