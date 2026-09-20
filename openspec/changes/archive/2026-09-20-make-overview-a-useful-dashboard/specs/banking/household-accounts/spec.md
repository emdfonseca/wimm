## REMOVED Requirements

### Requirement: The total is the member's own
**Reason**: Replaced by "Household money and a member's own money are separate
figures". One figure summing everything a member may see answered neither of
the questions a household asks, what is ours and what is mine, and it added a
partner's personal account to a member's own money the moment it was shared.
**Migration**: None for members. Per-currency totals, left-out accounts
counting for nothing, figures that agree with the accounts listed under them,
and a figure that never announces what it omits are all restated by the
replacement.

## ADDED Requirements

### Requirement: Household money and a member's own money are separate figures
**Story**: S5
Wherever wimm shows a total, it SHALL show two, and never one figure that mixes
them.

**Household money** is the sum of the accounts that every member of the
household either owns or has been given *details* on. It is the same figure
for every member, and it reveals nothing, because every member already sees
each of those accounts in full. A household of one member has no household
money: there is nobody to hold it with.

**A member's own money** is the sum of the accounts that member owns which are
not household money.

The two are disjoint: an account counts in one of them or in neither, never in
both. An account a member may see that is neither — one shared with them but
not with everyone, or shared with everyone at *balance* only — SHALL be listed
with its balance and SHALL count in no figure.

Each figure SHALL agree with the accounts listed under it. Where accounts are
held in more than one currency, each figure SHALL be shown per currency and
currencies MUST NOT be added together. A figure over no accounts SHALL NOT be
shown, rather than shown as zero. An account that is left out SHALL count in
no figure, including its owners'. No figure SHALL say what it omits.

#### Scenario: A joint account is household money
- **WHEN** both members of a household own an account
- **THEN** its balance is in household money for both of them, and in neither
  one's own money

#### Scenario: An account shared in full with everyone
- **WHEN** a member owns an account and has given every other member *details*
  on it
- **THEN** it counts as household money for every member, its owner included

#### Scenario: A personal account is the member's own
- **WHEN** a member owns an account nobody else has been given anything on
- **THEN** its balance is in that member's own money, and no other member sees
  any figure change because of it

#### Scenario: Every member sees the same household money
- **WHEN** two members of one household look at household money
- **THEN** they see the same figure

#### Scenario: A household of one
- **WHEN** the only member of a household owns three accounts
- **THEN** all three are their own money, and no household money figure is
  shown

#### Scenario: Shared at balance only
- **WHEN** a member has been given *balance* on an account they do not own
- **THEN** the account is listed for them with its balance, and it is in
  neither of their figures

#### Scenario: Shared with everyone, but only the balance
- **WHEN** an owner has given every other member *balance* on an account, and
  nobody *details*
- **THEN** the account is the owner's own money, and for everyone else it is
  listed and counted in no figure

#### Scenario: Shared in full with one member of three
- **WHEN** a household has three members and an owner gives *details* on an
  account to one of them only
- **THEN** the account is not household money for anybody

#### Scenario: Somebody joins the household
- **WHEN** a new member joins and has been given nothing yet
- **THEN** accounts stop being household money until that member owns them or
  is given *details* on them, and each returns to its owners' own money in the
  meantime

#### Scenario: The figures agree with the accounts
- **WHEN** a member adds up the balances listed under household money, and
  those listed as their own
- **THEN** each sum equals the figure shown for it, including where balances
  are negative

#### Scenario: Nothing in a group
- **WHEN** no account is household money for a member
- **THEN** no household money figure is shown, rather than one of zero

#### Scenario: Accounts in more than one currency
- **WHEN** a member's own accounts are held in two currencies
- **THEN** their own money is shown once per currency, and no combined figure
  is shown

#### Scenario: A left-out account is in no figure
- **WHEN** an owner leaves an account out
- **THEN** whichever figure it was in drops by its balance, and the accounts
  listed still add up to the figures shown

#### Scenario: A figure does not say what it leaves out
- **WHEN** a member has been granted nothing on one of a bank's accounts
- **THEN** no figure they see mentions it, because a figure that announces
  what it omits reveals the omission
