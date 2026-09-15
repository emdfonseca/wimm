## Purpose

Turns an enrolment link into a passkey and a signed-in session, so that a person
the operator has registered goes from a link in a chat window to being inside the
product without choosing a password or proving an email address.

## ADDED Requirements

### Requirement: Enrolling a passkey from a valid enrolment link
**Story**: S2
Opening a usable enrolment link SHALL show the person the name they were
registered under and offer to create a passkey. On success the system SHALL
record the passkey against them, SHALL sign them in without asking them to
confirm a second time, and SHALL stop the link from working again.

#### Scenario: Enrolling on a phone or laptop
- **WHEN** a registered person opens their enrolment link and confirms with
  their fingerprint, face or device PIN
- **THEN** their passkey is saved
- **AND** they are taken straight into the app, signed in, with their own name
  shown — they are not asked to sign in again

#### Scenario: The person sees who the link is for before committing
- **WHEN** a registered person opens their enrolment link
- **THEN** they are shown the name the operator registered them under, before
  they are asked to create anything

#### Scenario: The link cannot be used a second time
- **WHEN** someone opens an enrolment link that has already been used to enrol a
  passkey
- **THEN** they cannot enrol, and are told to ask the operator for a new link

### Requirement: Requiring a passkey the member can sign in with later
**Story**: S2
The system SHALL only accept a passkey that the person will be able to sign in
with without typing an identifier, and SHALL refuse one their device will not
save. A refused or abandoned attempt SHALL leave the person un-enrolled with
their link still usable, so they can try again on another device.

#### Scenario: The device will not save the passkey
- **WHEN** a person confirms the prompt but their device or passkey manager
  cannot save a passkey that works without typing an identifier
- **THEN** enrolment is refused and they are told to try a device or passkey
  manager that can save it
- **AND** their link still works, so the next attempt can succeed

#### Scenario: The person dismisses the prompt
- **WHEN** a person closes or cancels the passkey prompt
- **THEN** nothing is saved and they are left on the page able to try again
- **AND** their link still works

### Requirement: Refusing an enrolment link that cannot be used
**Story**: S2
An enrolment link that has expired, has already been used, has been replaced, or
was never issued SHALL show the same dead end: no way to proceed, and an
instruction to ask the operator for a new link. The response SHALL NOT reveal
whether the link ever existed.

#### Scenario: The link has expired
- **WHEN** a person opens an enrolment link after it has stopped working
- **THEN** they are told it is no longer usable and to ask the operator for a
  new one
- **AND** they are offered no way to enrol

#### Scenario: The link was mistyped or never issued
- **WHEN** someone opens a link that was never issued
- **THEN** they see the same message as for an expired link, which does not tell
  them whether any link like it ever existed
