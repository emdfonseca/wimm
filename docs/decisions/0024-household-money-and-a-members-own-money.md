# 0024 · Household money and a member's own money

## Status

Accepted

Supersedes the totals paragraph of ADR 0019, "Totals are per member". Owners,
levels and every other rule in 0019 are unchanged.

## Context

One total over everything a member may see answers neither question a household
asks, what is ours and what is mine, and it adds a partner's personal account to
a member's own money the moment it is shared at any level. ADR 0019 made totals
per member because a figure shared by every member could reveal an account
somebody may not see. That holds for a figure over accounts with mixed access. It
does not hold for accounts every member already sees in full.

Options: one total per member as now; a flag an owner sets on an account to mark
it household money; a figure derived from the owners and grants that already
exist.

## Decision

Wherever wimm shows a total it shows two, per currency, and never one that mixes
them. Household money is the sum of accounts that every member owns or holds a
details grant on, with two or more members in the household. Own money is the sum
of accounts the calling member owns that are not household money. The two are
disjoint. Any other account the member may see is listed and counted in neither;
a left-out account is counted nowhere.

The group of each account is decided once, in `wimmd`, from one query counting the
members who own or hold a details grant on it against the household's size, and
is sent as `Account.group`. `ListAccountsResponse` and `RefreshBalancesResponse`
carry `household_totals` and `own_totals` in place of `totals`. No client
classifies an account or sums a figure.

A flag was rejected: it restates what the grants already say and drifts the first
time a level changes.

## Consequences

Household money is the same for every member and leaks nothing, because every
member already sees each of those accounts in full. A member joining, or a grant
dropping below details, moves accounts out of household money into their owners'
own money until access is restored; no balance leaves the screen. A household of
one has no household money. Clients must read the two fields and the group; the
old `totals` field is reserved.
