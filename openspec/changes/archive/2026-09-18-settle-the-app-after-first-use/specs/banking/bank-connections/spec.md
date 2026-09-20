## ADDED Requirements

### Requirement: The steps of connecting a bank read as one flow
**Story**: S6
Connecting a bank SHALL read as one flow with a beginning and an end, not as a
run of unrelated pages. At every step before the hand-off the member SHALL be
able to see which step they are on and how many there are, SHALL find the way
forward in the same place with the same emphasis each time, and SHALL be able to
go back a step without losing what they have already chosen.

The step that commits the member — the hand-off to the bank — SHALL say what
happens next in its own label, rather than being a generic word for continuing,
because it is the point at which the member leaves wimm.

Restoring access and widening what a bank shares SHALL follow the same pattern.
They are shorter routes through the same flow, and a member who has connected a
bank once MUST NOT have to learn a second arrangement to do it again.

#### Scenario: Knowing where you are
- **WHEN** a member is on any step of connecting a bank before the hand-off
- **THEN** they can see which step it is and how many steps there are

#### Scenario: The actions are in the same place
- **WHEN** a member moves from one step to the next
- **THEN** the step's actions sit in the same place and take the same form, and
  where a step has a way forward it is the most prominent control on it

#### Scenario: A step whose list is the way forward
- **WHEN** a member is on the step that lists the banks
- **THEN** choosing a bank is what moves them on, and the step still carries its
  way out in the place every other step puts it

#### Scenario: Going back
- **WHEN** a member goes back from a step to the one before it
- **THEN** they return to it with the bank they had already chosen still chosen,
  and nothing has been connected

#### Scenario: The step that sends them to the bank says so
- **WHEN** a member reaches the step that hands them to their bank
- **THEN** the control that does it names the bank it will send them to, rather
  than saying only that it continues

#### Scenario: Restoring uses the same arrangement
- **WHEN** a member restores a bank's access, or widens what it shares
- **THEN** the steps carry the same arrangement of actions as connecting a bank
  for the first time

#### Scenario: Leaving the flow
- **WHEN** a member decides not to connect a bank after all
- **THEN** every step before the hand-off offers a way out that is clearly the
  lesser action, and taking it leaves the household exactly as it was

## MODIFIED Requirements

### Requirement: Choosing who owns each account and who sees it
**Story**: S2
A member SHALL be able to choose, for each account the bank made available, who
owns it and what each other member may see of it. This SHALL be presented as a
choice about what the household sees and MUST NOT be presented as granting or
withholding access to the bank, because the bank has already granted it.

The choice SHALL be reachable at any time from the household's accounts, and
SHALL NOT be a step a member passes through to finish connecting a bank. It is
about who sees an account, which is a different question from whether wimm can
read it, asked on a different day.

The member who connected the bank SHALL own every account it returned when the
choice opens. Ownership SHALL be shown as the set of members who own the
account, and an owner SHALL be able to add any member of the household to that
set and to remove themselves from it. Adding an owner MUST NOT remove any
existing one. No other member SHALL be granted any level until an owner grants
it. Every one of these choices SHALL remain changeable afterwards by any owner
of the account.

An owner SHALL NOT be able to leave an account with no owner at all. Where they
want the account out of wimm, the choice offers leaving it out, which is
reversible and is covered by the household's accounts. The way back into this
choice SHALL remain available to every owner of any account at that bank, so a
member cannot put an account beyond their own reach.

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

#### Scenario: Both members own it
- **WHEN** a member makes another member an owner of an account and stays an
  owner themselves
- **THEN** both are shown as owners and both see it in full

#### Scenario: An account nobody owns and nobody is granted
- **WHEN** the only owner of an account tries to remove themselves as its owner,
  with no level granted to anybody
- **THEN** the change is refused, because an account always belongs to somebody.
  They are offered leaving it out of wimm instead, which stops every read of it
  and can be undone

#### Scenario: Leaving an account out from the choice
- **WHEN** an owner leaves one of a bank's accounts out of wimm from this choice
- **THEN** it stays theirs, no balance is read for it, nobody else sees it, and
  it is shown to them as left out with a way to bring it back

#### Scenario: The way back stays open
- **WHEN** an owner has left every one of a bank's accounts out of wimm
- **THEN** they can still reach this choice from the household's accounts and
  bring any of them back

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
