## Purpose

Lets a household member who has already enrolled a passkey get back into wimm on
their own, without typing an identifier and without asking the operator for
anything.

## Requirements

### Requirement: Signing in with an enrolled passkey
**Story**: S3
A person with an enrolled passkey SHALL be able to sign in using it without
typing an email address, a username or any other identifier. An abandoned
attempt SHALL leave them able to try again.

#### Scenario: Coming back after closing the browser
- **WHEN** a person who enrolled earlier opens the app with no active session
  and starts signing in
- **THEN** their browser offers them the passkey they enrolled
- **AND** confirming with fingerprint, face or device PIN takes them into the
  app with their own name shown

#### Scenario: The person dismisses the prompt
- **WHEN** a person cancels the passkey prompt while signing in
- **THEN** they are left on the sign-in page, told that nothing happened, and
  able to try again

### Requirement: Sending an unauthenticated visitor to sign in
**Story**: S3
Someone without a valid session SHALL be sent to sign in rather than shown the
product or a dead end. After signing in they SHALL arrive at the page they were
trying to reach.

#### Scenario: The session has expired mid-use
- **WHEN** a person's session expires and they then try to do something in the
  app
- **THEN** they are asked to sign in
- **AND** signing in returns them to what they were looking at, not to a
  generic starting page

#### Scenario: Opening the app for the first time on a new browser
- **WHEN** a person opens the app with no session
- **THEN** they are shown the sign-in page

### Requirement: Signing out
**Story**: S3
A signed-in person SHALL be able to sign out, after which that session SHALL no
longer give access on that browser.

#### Scenario: Signing out on a shared computer
- **WHEN** a signed-in person signs out
- **THEN** they are returned to the sign-in page
- **AND** going back in the browser or reopening the app does not show their
  data without signing in again

### Requirement: Refusing a passkey that is not enrolled
**Story**: S3
A passkey this instance has no record of SHALL NOT sign anybody in, and the
person SHALL be told what to do instead.

#### Scenario: A passkey this instance does not know
- **WHEN** someone tries to sign in with a passkey for this site that this
  instance has no record of
- **THEN** they are not signed in
- **AND** they are told to ask the operator for an enrolment link
