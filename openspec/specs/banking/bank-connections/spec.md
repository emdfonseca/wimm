# banking/bank-connections Specification

## Purpose
Covers how a household member grants wimm read access to a bank, chooses which
of that bank's accounts the household sees, restores access when the bank's
grant runs out, and takes the bank away again. The grant is made at the bank,
not in wimm, so this capability owns the hand-off out of the product, the
return, everything that can fail on the member part-way through, and the
protection of anything wimm keeps that could reach the bank afterwards.

## Requirements

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
them straight to the household's accounts, where that bank's accounts now
appear. The one-time value the bank returns with SHALL be exchanged immediately
and MUST NOT remain in the address a member can bookmark, share or find in
their history.

Returning SHALL NOT ask the member anything. Every account the bank returned is
already theirs and already invisible to everyone else, so there is no question
at this point whose answer is not already filled in.

#### Scenario: Access granted
- **WHEN** a member confirms at their bank and is returned to wimm
- **THEN** that bank's accounts appear among the household's accounts, already
  theirs, and they are told that nobody else in the household sees them yet

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

### Requirement: Choosing who owns each account and who sees it
**Story**: S1
A member SHALL be able to choose, for each account the bank made available, who
owns it and what each other member may see of it. This SHALL be presented as a
choice about what the household sees and MUST NOT be presented as granting or
withholding access to the bank, because the bank has already granted it.

The choice SHALL be reachable at any time from the household's accounts, and
SHALL NOT be a step a member passes through to finish connecting a bank. It is
about who sees an account, which is a different question from whether wimm can
read it, asked on a different day.

The member who connected the bank SHALL own every account it returned when the
choice opens, and SHALL be able to disown any of them and to add another member
as an owner. No other member SHALL be granted any level until an owner grants it.
Every one of these choices SHALL remain changeable afterwards by any owner of
the account.

#### Scenario: The connecting member owns what they connected
- **WHEN** a member has connected a bank and opens the choice for it
- **THEN** every one of them is already theirs, and they are told that nobody
  else sees any of it yet

#### Scenario: Nothing is granted to anyone by default
- **WHEN** a member is shown the accounts a bank has made available
- **THEN** no other member has any level on any of them, and the member is told
  that each other member sees only what they are given here

#### Scenario: Handing an account to the member it belongs to
- **WHEN** a member makes another member an owner of an account and removes
  themselves
- **THEN** the other member sees that account in full, the first member no
  longer sees it at all, and the first member is not offered it again as
  something to grant

#### Scenario: Giving one member the balance and another nothing
- **WHEN** a member grants *balance* on an account to a second member and
  leaves a third member with nothing
- **THEN** the second member sees the bank, the name and the amount, and the
  third member sees no trace of the account

#### Scenario: Giving one member everything at one bank is one action
- **WHEN** a member wants another member to see every account at that bank at
  the same level
- **THEN** a single action sets that level on all of them for that member

#### Scenario: Finishing having granted nothing
- **WHEN** a member finishes having granted no level to anybody
- **THEN** they can finish, because the accounts are theirs and they can see
  them; the household simply sees nothing of this bank yet

#### Scenario: An account nobody owns and nobody is granted
- **WHEN** a member disowns an account and grants no member any level on it
- **THEN** it appears for nobody, no balance is ever read for it, and it is not
  counted in anyone's total

#### Scenario: Changing the choice later
- **WHEN** an owner reopens the choice for a connected bank
- **THEN** they see the accounts they own with their current owners and levels,
  and can change any of them

#### Scenario: A member who owns nothing at a bank cannot change it
- **WHEN** a member who owns none of a bank's accounts opens that bank's choice
- **THEN** they can change nothing, and they see only the accounts they have
  been granted

#### Scenario: A bank that offers one account
- **WHEN** a bank makes exactly one account available
- **THEN** the member is still asked, because who sees it is still a decision,
  and leaving it theirs alone is one action

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
Every account SHALL keep the owners and the levels it had before, even though
the bank issues new identifiers each time.

#### Scenario: Restoring access
- **WHEN** a member chooses to restore an expired connection and confirms at
  their bank
- **THEN** every account keeps the owners and the levels it had before, their
  balances are read again, and Overview stops saying the bank has stopped
  updating

#### Scenario: The bank now offers an account it did not before
- **WHEN** a bank makes an account available on restoring that it did not make
  available before
- **THEN** it belongs to the member who restored the connection, no other member
  has any level on it, and that member is told there is something new to choose

#### Scenario: The bank no longer offers an account that members could see
- **WHEN** an account members could see is not among those the bank makes
  available on restoring
- **THEN** it stops appearing on Overview for everyone who could see it, its
  owners and levels go with it, and the member restoring is told which account
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
