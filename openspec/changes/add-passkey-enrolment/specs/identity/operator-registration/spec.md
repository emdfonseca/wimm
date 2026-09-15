## Purpose

Household members do not sign themselves up. The operator who runs the instance
registers each person and is handed an enrolment link to pass on, which is also
how a member who has lost their device is brought back.

## ADDED Requirements

### Requirement: Registering a household member
**Story**: S1
The operator SHALL be able to register a person by giving their email address,
first name and last name, and SHALL be given back a single enrolment link to
share with that person. The system SHALL NOT verify that the email address
belongs to them; it records what the operator asserts.

#### Scenario: Registering someone new
- **WHEN** the operator registers a person with an email address nobody is
  registered under, together with a first and last name
- **THEN** that person is registered under the name the operator gave
- **AND** the operator is shown one enrolment link to pass on

#### Scenario: The operator is told how long the link lasts
- **WHEN** the operator is given an enrolment link
- **THEN** they are also told when it stops working, so they can say so when
  they send it

#### Scenario: A name is missing
- **WHEN** the operator registers a person without a first or last name
- **THEN** the registration is refused and the operator is told which part is
  missing
- **AND** nobody is registered and no link is issued

### Requirement: Refusing a registration that names an already-registered email
**Story**: S1
The system SHALL refuse to register a person under an email address that is
already registered, and SHALL leave the existing person untouched.

#### Scenario: The email is already registered
- **WHEN** the operator registers a person using an email address that is
  already registered
- **THEN** the registration is refused and the operator is told the address is
  already in use
- **AND** no second person is created, no link is issued, and the existing
  person keeps their name and their passkeys

### Requirement: Issuing a further enrolment link to a registered member
**Story**: S1
The operator SHALL be able to issue a further enrolment link to someone who is
already registered, so that they can add another device or replace one they have
lost. Issuing a new link SHALL stop any enrolment link previously issued to that
person from working, and SHALL leave the passkeys they have already enrolled
working.

#### Scenario: Replacing a lost device
- **WHEN** the operator issues a further enrolment link for someone already
  registered
- **THEN** the operator is given a new link to pass on
- **AND** the person can still sign in with any passkey they enrolled before

#### Scenario: The earlier link stops working
- **WHEN** a person is issued a new enrolment link while an earlier one is still
  unused
- **THEN** opening the earlier link no longer allows enrolment

#### Scenario: The person is not registered
- **WHEN** the operator asks for an enrolment link for an email address nobody
  is registered under
- **THEN** the request is refused and the operator is told that nobody is
  registered under it

### Requirement: Refusing an unauthenticated operator
**Story**: S1
Registering a person and issuing an enrolment link SHALL require the operator
credential configured for the instance. The instance SHALL refuse to start with
its operator surface reachable but no operator credential configured, rather
than starting with that surface open.

#### Scenario: The operator credential is wrong or absent
- **WHEN** someone attempts to register a person without the configured
  operator credential, or with the wrong one
- **THEN** the attempt is refused
- **AND** nobody is registered and no link is issued

#### Scenario: The instance is started with no operator credential configured
- **WHEN** the instance is started with its operator surface reachable and no
  operator credential configured
- **THEN** it refuses to start and reports that the credential is missing
