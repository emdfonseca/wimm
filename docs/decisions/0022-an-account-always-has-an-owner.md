# 0022 · An account always has an owner, and being out of wimm is its own fact

## Status

Accepted

Supersedes the orphan half of ADR 0019: "No balance is read for an account with
no owner and no grant", "an account that reaches the unowned-and-ungranted state
is never read again", and the end state that text calls going dark and staying
dark. Everything else in 0019 — ownership as a set, the three levels, hidden as
the absence of a row, the refusal of an owner holding a grant, any owner being
able to change owners and levels, the connecting member owning what they
connected, per-member totals, and restoring carrying owners and grants forward —
is unchanged.

## Context

0019 gave one mechanism two jobs. The owner set says who sees an account in
full, and the *absence* of that set — together with the absence of every grant —
was also how a member said the account should not be in wimm at all.

It is a tidy model and it is wrong in use. A member unticks one checkbox in the
chooser, with no confirmation, and four things happen at once:

```text
the account leaves Overview
its transactions become unreadable, by everyone, including its owners
no balance is ever read for it again
the bank card it sat under disappears, if it was that bank's last visible
  account — and the link into the chooser goes with it
```

The last one is the defect that makes the rest unrecoverable. The link is
offered only where the member owns something at that bank, so orphaning an
account can remove the only route to the screen that could undo it.

0019 saw this coming and patched it in Go: `requireOwner` and
`requireOwnerOnConnection` each let the connection's `connected_by` member act
on an account nobody owns, with a comment saying that without it "disowning your
last account locks you out of the only screen that could undo it. A dead end
reachable by one click." That is an escape hatch around a model, and it is
narrow in exactly the ways an escape hatch is: it works for one member, only
where a connection still exists, and only for somebody who knows the URL.

The tell was there at the time. A state whose recovery needs a special case in
the authorisation check is not an end state a product should be able to reach.

## Decision

**Every account has at least one owner, and the database is what says so.** A
deferred constraint trigger on `account_owners` and on `accounts` refuses any
transaction that would commit an account with no owner row.

It is deferred because `SetAccountOwners` deletes every owner row and inserts
the new set inside one transaction, so an immediate trigger would fire in the
gap and refuse a legal change. It checks the account still exists, because
deleting an account — or a connection, which cascades to its accounts — cascades
to its owner rows, and a trigger that did not look would refuse every
disconnection.

*Alternative: a check in Go.* Rejected, and the evidence is in this repository.
The other rule 0019 states about these tables — that a member never holds both
an owner row and a grant row — is guarded in Go on one write path and quietly
resolved by deleting the grant on the other. Two call sites, two behaviours, one
rule. This rule has no route that may bend it, so it belongs where no route can.

*Alternative: `accounts.owner_id not null`.* Rejected: it would undo the
many-to-many a joint account needs, which is the case 0019 exists for.

**"Not in wimm" becomes its own fact, and it is reversible.** An account is
**left out** or it is not — `accounts.left_out_at`, null or a time. Leaving an
account out stops every read of it, removes it from every list and every total
belonging to every member who is not an owner, and leaves them no trace of it.
Its owners keep a row saying it is left out, with no balance on it and a way to
bring it back.

This is ADR 0018's privacy affordance kept, with the boundary moved for the
second time. 0018 drew it at *unshared*; 0019 moved it to *unowned and
ungranted*; it is now *left out*. The guarantee itself has not moved once: wimm
knows an account exists and does not know what is in it.

**A left-out account's transactions stop being listed, for everyone.** This is
deliberately unlike a bank being disconnected, where what was already read stays
on screen for the members who could see it. Disconnecting ends wimm's access to
a bank; leaving an account out is a member saying that account is not in wimm,
and a ledger still listing it would contradict them. Nothing is deleted either
way, and bringing the account back brings its rows back with it.

**Grants survive being left out, and bringing an account back says who will see
it.** Dropping the grants would make a reversible act destructive and would make
the member re-grant from memory. Keeping them means an account can come back and
re-expose itself to somebody who is not in the room, so the member bringing it
back is told which members will see it again and at what level, before it
happens. This is the only confirmation in the flow, and it is there because the
consequence lands on somebody else.

**The escape hatches come out.** `requireOwner` and `requireOwnerOnConnection`
lose their orphan branches, and the tests that pinned them become tests of the
refusal. A branch that grants authority in a state that can no longer occur is
not dead code; it is authority waiting for the invariant to break.

**Existing orphans are brought back rather than left.** Each becomes owned by
the member who connected its bank, and left out. The migration asserts no
ownerless account remains and fails if one does, because a backfill that half
worked is worse than one that did not run.

## Consequences

The chooser's "mine" checkbox becomes an owner set naming every member of the
household. That is a bigger control, and it is what the specification has
required since `connect-bank-accounts`: "Handing an account to the member it
belongs to" was a scenario nobody could perform, because the only control ever
built could add the current member and nobody else — and ticking it replaced the
whole owner set, so it silently removed a joint account's other owner.

Leaving out is offered where disowning used to be, which is the same place a
member was already reaching for. The difference is that it says what it does,
asks before it does it, and can be undone.

**Two facts now have to be read where one was.** "May this member see this
account" is still the owner-or-grant join. "May wimm read this account" was that
same join and is now `left_out_at is null` — simpler than what it replaces,
because an account always has an owner and the two-part existence test collapses.

**A member deletion, if one is ever added, will be refused by this trigger**
where that member is an account's last owner. That is the right failure and it
is asserted by a test rather than left to be met in production. Whoever adds
member deletion has to decide what happens to their accounts first, which is a
decision that should never have been possible to skip.

**The chooser is still scoped to one bank.** 0019's note that nothing shows what
one member can see of the household, and that a sharing surface under Settings
is the shape, is unchanged and still not built. This change adds Settings as a
screen, which is where that surface will go; it does not put it there.
