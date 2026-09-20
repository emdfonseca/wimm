# 0019 · Accounts have owners, and each member sees a level

## Status

Accepted; totals superseded by 0024

Supersedes the sharing half of ADR 0018: "Everything the bank returns is stored,
and the member chooses what is shared", "Accounts are household-visible once
shared", and `shared` defaulting to false. Everything else in 0018 — the gateway
port, account identity on the cross-session hash, sealing, money as `int64`
minor units, and the chooser not being a consent step — is unchanged.

## Context

0018 gave an account one boolean. Shared meant every member of the household saw
it; unshared meant nobody did, including the member who connected the bank. One
instance serves one household, so there was no scoping rule beyond the flag.

That is the right model for a household that shares everything, and the wrong
one for the household wimm is actually for. Two people run a joint account and
each keep their own. The partner should see the joint balance without seeing the
number on a personal account, and "we have €4,200 between us" and "here is my
account number, my account type and the name on it" are not the same sentence.
A boolean forces them to be.

The free tier makes this sharper rather than softer. Enable Banking's terms
reach only accounts the Control Panel user links as themselves, so **one member
connects every bank** and everything the household sees arrives through that one
member's access. Under a boolean, that member's only choices are to expose a
personal account to everyone or to hide it from themselves.

## Decision

**An account has owners, and an owner sees it in full.** Ownership is
many-to-many: a joint account is owned by both partners, which is the case the
model exists for. Ownership is never a level and never a degree — an owner sees
everything wimm holds about their account, whatever anyone else has been given.

**Every other member gets a level, per account.** Three, and no others:

```text
hidden    the account leaves no trace this member can see
balance   the bank, the account's name, the balance and its read time
details   balance, plus the number suffix, the account type, the holder name
```

The split is not "some fields versus more fields": it is between what a
household needs to answer "where is our money" and what identifies an account to
a person holding it.

**Change 2's transactions do not ride on `details`; they get a fourth level.**
The top level is named for what it grants — the account's identifying details —
rather than "all", precisely so that it cannot silently widen. A member who
granted it before transactions existed consented to seeing an account number,
not to seeing someone's spending, and those are different sentences. Adding the
fourth level is one `ALTER TABLE`, which is what the text-plus-check choice
above was for.

**Hidden is the absence of a row.** `bank_account_grants` holds only `balance`
and `details`, so the common read — what may this member see — is a join
returning what exists rather than a filter over what does not. A member removed
from the household loses their visibility and their ownership by cascade, in one
statement, with nothing to remember to clean up.

**A member never holds both an owner row and a grant row on one account.**
Ownership outranks every level, so a grant beside it is a second answer to a
settled question. The store refuses the pair rather than resolving it: a row
that is ignored is a row that will one day be believed.

**Any owner may change owners and levels; the connecting member has no standing
power.** Whoever holds the consent is recorded in `connected_by` because someone
has to restore it, and that is all it confers. An account handed to its real
owner is fully theirs, including the right to hand it on.

**The connecting member owns every account a connection returns, and may disown
it.** They linked the bank as themselves, so every account it returned is one
they can already see by logging in there — showing it to them reveals nothing,
and showing it with its balance is what makes the choice informed rather than a
list of names.

*Alternative:* arrive with no owner, so nothing is read until a member claims it.
Rejected: the member would assign levels to bare names with no figures, and the
balance is the one fact that tells a current account from a mortgage.

**No balance is read for an account with no owner and no grant.** This is 0018's
guarantee kept, with the boundary moved from "unshared" to "unowned and
ungranted".

**Every account a bank returns is read once, before anyone can disown it.** The
return from the bank goes to Overview, where the connecting member owns
everything that bank returned, so everything is readable and everything is
read. Disowning is a later, separate action and stops every read after it.

This is stated rather than designed away. Closing it would mean showing no
figures until a member had finished choosing, which leaves the member who
connects one bank and finishes staring at an empty screen — and the account
being read is one that member can already see by logging in at their own bank,
so the read reveals nothing to them that they did not already have.

An account that reaches the unowned-and-ungranted state is never read again,
which is the case the rule exists for: an account handed to its real owner, or
disowned by the member who connected it, goes dark and stays dark.

**Totals are per member.** Two members of one household can land on the same
screen and correctly see different numbers. A total never announces what it
omits, because a total that says "and three more you cannot see" reveals the
omission it exists to respect.

**Restoring carries owners and grants forward**, matched on the gateway's
cross-session hash. An account newly offered belongs to the member who restored
it with nobody granted; one the bank no longer offers goes with its owners and
grants.

## Consequences

The schema gains two tables and loses a column. `accounts` carries no `shared`
flag and no owner column — both were single answers to questions that turn out
to have one answer per member.

The table is `accounts`, not `bank_accounts`, and its `connection_id` is
nullable. Ownership and levels attach to an account, not to a bank: an account
kind with no gateway behind it is owned, shared and totalled through exactly
these tables. Tying an account's existence to a gateway connection would make
every later account kind a schema change, and this is the moment where that
costs nothing.

Redaction moves into the handler. A field a member may not see is never
serialised, rather than being fetched and dropped by the client, so there is one
place where "what may this member see" is decided and it is the place that
answers the request.

The chooser stops being a checkbox list. It carries an ownership control and one
three-way control per other member of the household, which is why the design
system needs a segmented control it did not have, and why the chooser's frames
had to be redrawn. At Compact the level controls stack under the account.

Finishing the chooser having granted nothing is now a legitimate outcome, where
0018 blocked it. The accounts are the connecting member's and they can see them;
the household simply sees nothing of that bank yet.

**What this does not add is a place to manage it all.** The chooser is scoped to
one bank, so answering "what can my partner see?" means opening every bank in
turn. Nothing sets sharing for a member across banks, and nothing shows what one
member can see of the household — which is the question a person actually has.
A sharing surface under Settings — ADR 0005's "longer or growing set → a list
card in the page" — is the shape, and it is deliberately not in this change.

Three choices are made so that it stays additive rather than becoming a
migration. `level` is `text` with a check constraint rather than a Postgres
`enum` type, because a fourth level on an enum type needs `ALTER TYPE … ADD
VALUE`, which cannot run in a transaction. A grant is keyed on the account and
the member alone, with no `connection_id`, so every grant for one member across
every bank is a single-table query. And visibility is computed for a set of
accounts rather than for a connection, so the question can span banks without
the signature changing.

None of this is speculative generality: each is the cheaper of two options
today, and each is the one that does not have to be undone.
