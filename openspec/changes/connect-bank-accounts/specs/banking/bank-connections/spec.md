## Purpose

Covers how a household member grants wimm read access to a bank, chooses which
of that bank's accounts the household sees, restores access when the bank's
grant runs out, and takes the bank away again. The grant is made at the bank,
not in wimm, so this capability owns the hand-off out of the product, the
return, everything that can fail on the member part-way through, and the
protection of anything wimm keeps that could reach the bank afterwards.

## ADDED Requirements

### Requirement: Choosing a bank to connect
**Story**: S1
A member SHALL be able to start connecting a bank by choosing it from a list of
banks wimm can reach. The list SHALL be searchable by name, because a member
knows the name of their bank and not its position in an alphabetical list.

#### Scenario: Picking a bank from the list
- **WHEN** a signed-in member starts connecting a bank
- **THEN** they are shown a list of banks they can connect, each identified by
  a name and logo they would recognise from their own banking app

#### Scenario: Finding a bank by typing its name
- **WHEN** a member types part of their bank's name
- **THEN** the list narrows to matching banks

#### Scenario: The bank is not in the list
- **WHEN** a member's search matches no bank
- **THEN** they are told plainly that wimm cannot connect that bank yet, rather
  than being shown an empty list with no explanation

#### Scenario: The list cannot be loaded
- **WHEN** wimm cannot reach the service that supplies the list of banks
- **THEN** the member is told that connecting a bank is unavailable right now
  and invited to try again, and no partly-made connection is left behind

### Requirement: Consenting at the bank
**Story**: S1
Consent SHALL be given at the member's own bank and never inside wimm. wimm
MUST NOT ask a member for their banking credentials, MUST ask the bank only for
account details and balances, and MUST tell the member what it will be able to
read before sending them there. wimm SHALL also tell the member how long the
bank's access will last, using the limit that bank actually sets.

#### Scenario: Handing off to the bank
- **WHEN** a member confirms the bank they picked
- **THEN** they are told that wimm will be able to read that bank's account
  names and balances and not its transactions, and are then taken to the bank's
  own site to confirm

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

### Requirement: Returning from the bank
**Story**: S1
When a member returns from their bank having granted access, wimm SHALL take
them straight to choosing which accounts the household sees. The one-time value
the bank returns with SHALL be exchanged immediately and MUST NOT remain in the
address a member can bookmark, share or find in their history.

#### Scenario: Access granted
- **WHEN** a member confirms at their bank and is returned to wimm
- **THEN** they are shown the accounts that bank has made available, and asked
  which of them the household should see

#### Scenario: The return address does not keep the one-time value
- **WHEN** a member has finished returning from their bank
- **THEN** the address shown in their browser no longer contains the value the
  bank sent them back with

#### Scenario: The same return is used twice
- **WHEN** a member goes back to the return address, reloads it, or opens it a
  second time on another device
- **THEN** nothing is connected a second time, no account is duplicated, and
  they are shown the household's accounts as they already stand

#### Scenario: Returning long after consenting
- **WHEN** a member returns to wimm so long after starting that the attempt is
  no longer valid
- **THEN** they are told the attempt has expired and offered the chance to
  start again, and no bank is connected

### Requirement: Choosing which accounts the household sees
**Story**: S1
After access is granted, a member SHALL choose which of the bank's accounts the
household sees. This SHALL be presented as a choice about what the household
sees and MUST NOT be presented as granting or withholding access to the bank,
because the bank has already granted it. Nothing SHALL be shared until the
member chooses it, and the choice SHALL remain changeable afterwards.

#### Scenario: Nothing is shared by default
- **WHEN** a member is shown the accounts a bank has made available
- **THEN** none of them is marked as shared yet, and the member is told that
  everyone in the household will see whichever ones they choose

#### Scenario: Sharing everything is one action
- **WHEN** a member wants the household to see every account at that bank
- **THEN** a single action selects all of them

#### Scenario: Finishing with nothing chosen
- **WHEN** a member has chosen no account
- **THEN** they cannot finish, and they are told that a bank with nothing shared
  would show the household nothing

#### Scenario: Only chosen accounts appear
- **WHEN** a member finishes having chosen some of the bank's accounts
- **THEN** exactly those accounts appear on Overview for every member of the
  household, and the ones not chosen appear for nobody, including the member who
  connected the bank

#### Scenario: Changing the choice later
- **WHEN** a member reopens the choice for a connected bank
- **THEN** they see the same accounts with their current selection, and can
  share one that was not shared or stop sharing one that was

#### Scenario: A bank that offers one account
- **WHEN** a bank makes exactly one account available
- **THEN** the member is still asked, because sharing it with the household is
  still a decision, and confirming it is one action

### Requirement: A connection that cannot be completed
**Story**: S1
Every way the round-trip can fail SHALL leave the household exactly as it was
before the member started, and SHALL tell the member which of the two things
happened: their bank refused, or wimm could not complete its side.

#### Scenario: The member declines at their bank
- **WHEN** a member reaches their bank and declines to grant access
- **THEN** they are returned to wimm and told that the bank was not connected
  because access was not granted, with the option to try again

#### Scenario: The bank is unavailable
- **WHEN** the bank or the service wimm reaches it through fails while the
  connection is being completed
- **THEN** the member is told the bank could not be reached and invited to try
  again later, and no half-made connection appears in the household

#### Scenario: Access is granted but exposes no accounts
- **WHEN** a member grants access and their bank offers no accounts wimm can
  read
- **THEN** the member is told that no accounts came back from that bank, and
  the connection is not kept

#### Scenario: A failed attempt leaves nothing behind
- **WHEN** any attempt to connect a bank fails for any reason
- **THEN** the household's accounts and total are unchanged, and a later
  successful attempt at the same bank connects normally

### Requirement: Access that has run out
**Story**: S4
A bank's grant SHALL be treated as expired both when the date it set is reached
and when the bank rejects wimm's access before then. An expired connection SHALL
say so where a member is looking at the balances it stopped updating, and SHALL
offer the way to restore it. wimm MUST NOT present a balance from an expired
connection as current, and MUST NOT remove it either.

#### Scenario: A connection reaches the end of its access
- **WHEN** a member opens Overview and a connection's access has run out
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
**Story**: S4
A member SHALL be able to restore an expired connection by confirming again at
the bank, without disconnecting first and without choosing the bank again.
Accounts the household was already sharing SHALL still be shared afterwards,
even though the bank issues new identifiers each time.

#### Scenario: Restoring access
- **WHEN** a member chooses to restore an expired connection and confirms at
  their bank
- **THEN** the same accounts are shared as before, their balances are read
  again, and Overview stops saying the bank has stopped updating

#### Scenario: The bank now offers an account it did not before
- **WHEN** a bank makes an account available on restoring that it did not make
  available before
- **THEN** it is not shared, and the member is told there is something new to
  choose

#### Scenario: The bank no longer offers an account that was shared
- **WHEN** an account the household was sharing is not among those the bank
  makes available on restoring
- **THEN** it stops appearing on Overview, and the member is told which account
  the bank no longer offers

#### Scenario: Restoring is refused at the bank
- **WHEN** a member declines at the bank while restoring
- **THEN** the connection stays exactly as it was, still expired, still showing
  its last readings, and still offering to restore

### Requirement: Disconnecting a bank
**Story**: S3
A member SHALL be able to disconnect any bank connected to their household.
Disconnecting SHALL be confirmed first, naming the bank and how many accounts
will go, because it is destructive and money-bearing. wimm SHALL end its own
access even when it cannot tell the bank.

#### Scenario: Confirming before disconnecting
- **WHEN** a member chooses to disconnect a bank
- **THEN** they are asked to confirm, and the question names the bank and how
  many of the household's accounts will disappear

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
- **THEN** it connects normally and its accounts appear once, not twice

### Requirement: What reaches a bank is unreadable at rest
**Story**: S1
wimm stores a value that lets it read a household's bank until the grant runs
out. That value, and anything else wimm keeps that could reach a bank, SHALL be
unreadable to anyone holding the stored data alone, and SHALL never appear in
logs, error messages, diagnostics or anything wimm shows a member. wimm MUST NOT
store a member's banking credentials at all, because it never receives them.

#### Scenario: The stored data alone opens nothing
- **WHEN** someone reads wimm's stored data directly, including a complete copy
  of it
- **THEN** nothing in it can be used to reach any bank without the separate
  secret that wimm is configured with, which is not kept alongside it

#### Scenario: Nothing that reaches a bank is ever logged
- **WHEN** a connection is made, used, restored, or fails in any way
- **THEN** no log line, error message or diagnostic contains a value that could
  be used to reach the bank

#### Scenario: Losing access ends the access
- **WHEN** a connection is disconnected or its access has run out
- **THEN** what wimm kept in order to reach that bank is destroyed rather than
  left behind

#### Scenario: A member never sees it
- **WHEN** a member looks at a connected bank anywhere in wimm
- **THEN** they are shown the bank, who connected it and when access ends, and
  never a value that could be used to reach the bank
