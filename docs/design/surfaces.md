# Surface and navigation decisions

Which component to reach for, and what each one must never be used for. The
components live in `40 · ORGANISMS`; this is the rule that picks between them.
Mirrored as a canvas section in `00 · README`.

## Navigation

```text
Breadcrumb = where this thing lives
Tabs       = peer views of this thing
Stepper    = progress through a task
```

| Pattern | Use it for | Never for |
|---|---|---|
| Primary nav | Top-level areas: Accounts, Transactions, Budgets, Reports, Settings | Steps in a task, or a filter |
| Secondary nav | Stable subsections inside one area | Temporary form state |
| Breadcrumbs | Hierarchical location: Accounts → Current account → Statement | Progress through a task |
| Tabs | Peer views of one entity: an account's Activity / Statements / Details | Parent-and-child hierarchy |
| Stepper | Ordered stages of one task, such as connecting an account | Product navigation |
| Back | Returning to the contextual origin of an overlay or detail | Replacing hierarchy people need direct access to |

Never use two of these to express the same dimension. Navigation depth counts
parent relationships from the documented root, not URL segments.

## Task surfaces

Modality is behaviour; drawer is placement. State both. For every overlay also
state **route-backed** or **ephemeral** — it decides Back, refresh, analytics and
focus return.

| Surface | Best when | Avoid when |
|---|---|---|
| Inline edit | One small value changes in place — recategorise, rename | The change has dependencies or real consequences |
| Modal dialog | A short blocking decision or compact task | The form is long, needs the page visible, or will grow |
| Drawer | A moderate task that benefits from keeping the list visible. Non-modal and route-backed in this product | The task is a destination of its own, or needs width |
| Full page | Complex, resumable, deep-linkable, likely to grow: connect an account, split a transaction, reconcile | A tiny local decision |
| Confirmation dialog | Anything destructive or money-bearing; names the record and the amount | Routine edits — confirming everything trains people to dismiss |

## Feedback surfaces

The question is always: **does this need action, and does it survive a refresh?**
Money operations answer yes far more often than most products.

| Surface | Use it for | Never for |
|---|---|---|
| Toast | A completed action needing no decision. Auto-dismisses, and the same fact is visible elsewhere | A failed money operation |
| Inline alert | A condition affecting one section, shown where the condition is | Global conditions |
| Page banner | A condition affecting the page or account; persistent, carries the fix | Transient confirmations |
| Global banner | Account-wide: payment method expired, service degraded | Anything page-specific |
| Dialog | The system needs a decision before it can continue | Information needing no action |
| Badge | The state of one row: Pending, Cleared, Failed | Anything needing an action — that is an alert |
| Field error | Invalid input, in the Form field message row, naming the fix | System failures unrelated to the input |

## Rulings specific to money

**A failed money operation never gets a toast.** It gets a persistent surface
carrying the retry. A transient message can be missed, and the user is left
believing money moved.

**Moving, deleting or reclassifying money is a consequential submission**
(SC 3.3.4): review-and-confirm naming the record and the amount, or a reversal
with a stated window.

**Optimistic UI is allowed for categorisation and renaming, never for a transfer.**
The test is whether being wrong for two seconds is recoverable by the user alone.

**Direction is a control, not a character.** Money in or out is chosen with an
explicit control, never by typing a leading minus. A typed sign is easy to omit,
easy to double, and indistinguishable from a hyphen or en-dash pasted out of a
statement. The field holds an unsigned magnitude; the direction control owns the
sign and the ledger renders it. This removes the class of error where income is
recorded as an expense and nobody notices until month end.

**The currency symbol is an affix, not content.** Never selectable, never
deletable, never part of the value the field returns. Digits are mono and
right-aligned so the decimal sits where it will sit in the ledger. The currency
code appears as a trailing affix only when the account is multi-currency.

**An amount carries no direction colour while being typed.** Green and red are for
reading money, not entering it.
