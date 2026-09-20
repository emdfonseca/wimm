# Surface and navigation decisions

Which component to reach for, and what each one must never be used for. The
components live in `packages/ui/src/organisms`; this is the rule that picks
between them.

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

## Feedback surfaces — ask these in order

Stop at the first yes. The order runs **from most specific to most general**, not
by severity — because the general questions swallow the specific ones. "Is this a
persisting condition?" is true of invalid input, of a Pending record, *and* of a
stale account, so it has to be asked last or it captures all three.

Question 4 covers row states needing **nothing**; question 5 covers conditions
needing **action** at any scope, *including a single row*. A row that needs a
decision is neither a badge nor nothing.

| | question | → | why it sits here |
|---|---|---|---|
| 1 | Must the user decide before anything can continue? | **Dialog** | Blocking, modal, focus trapped. If it is really "we would like them to decide", it is not a dialog. |
| 2 | Is it about what the user just typed? | **Field error** | In the Form field message row, naming the fix. Asked early precisely because invalid input is *also* a persisting condition — ask question 5 first and every validation message becomes a banner. |
| 3 | Did a money operation fail? | **Page banner** | Persistent, carries the retry, survives refresh. Ahead of question 4 because a failed transfer is *also* the state of one row, and a badge cannot carry a retry. |
| 4 | Is it the state of one row, needing **no** action? | **Badge** | Pending, Cleared, Failed. "Needing no action" is the whole of it: the moment a row's condition asks for a decision, question 5 takes it. |
| 5 | Is it a condition that persists and needs acting on? | **Notice, scoped to what it affects** | Scope decides placement, and **one row is a scope**: a suspected duplicate gets a notice on that row. Section → inline alert. Page or account → page banner. Above the shell → global banner. Stays until the condition clears, not until dismissed. |
| 6 | Did something succeed, needing no decision? | **Toast** | Auto-dismisses. Only when the same fact is visible elsewhere, so missing it costs nothing. |
| — | none of the above | **show nothing** | Most interfaces are noisy because every event was assumed to deserve a surface. A save already visible in the row it changed needs no toast. |

### Worked examples

Most of these are the ones people get wrong, and several resolve to no surface.

| situation | surface | why |
|---|---|---|
| Recategorised a transaction | **Toast** | Q6. Succeeded, no decision, and the row already shows it — the toast is a courtesy, not the record. Carries Undo because reversing is cheap. |
| Transfer to Savings failed | **Page banner** | Q3. The worst possible toast in this product: it disappears and leaves someone believing the money moved. |
| Typed `84.2O` in an amount | **Field error** | Q2. The person is looking at the field; a banner sends them hunting for which one. |
| A transaction is pending at the bank | **Badge** | Q4. The state of one row, needing nothing. A banner per pending row buries the ledger it describes. |
| Amex has not synced in six days | **Inline alert on that account** | Q5, scoped. It affects one account, so it sits on that account — not atop a page listing five. |
| Groceries is over budget | **Inline alert in that budget** | Q5. The condition belongs to one budget and travels with it. |
| Import finished, 12 uncategorised | **Inline alert on the list** | Q5, not Q6. It succeeded but left work, and work needs a surface that stays. |
| Deleting a reconciled transaction | **Dialog** | Q1. Consequential under SC 3.3.4. Names merchant, amount and date — "Are you sure?" asks about nothing. |
| Card on file expires this month | **Global banner** | Q5 at account scope. Not about whichever page you are on. |
| Saved an edit in the drawer | **nothing** | The panel shows the saved value. A toast here trains people to ignore toasts. |
| Data is still loading | **nothing** | A skeleton is the answer. Feedback is for what happened, not what has not finished. |
| One transaction looks like a duplicate | **Inline alert on the row** | Q5 at row scope. It needs a decision, so Q4 does not take it — a badge carries no action at all. The action is **Review**, singular, which opens the comparison; Keep and Merge are decided there, with both records visible. An earlier version of these rules had this case falling through to showing nothing, and a later one put Keep and Merge in the row. |
| A filter returned no rows | **nothing — empty state** | Not feedback. The list owns this state, and it belongs in the list. |

### What each surface refuses

| surface | never |
|---|---|
| Toast | A failed money operation. Anything the user must act on. Anything whose only copy is the toast — if missing it loses information, it is the wrong surface. |
| Inline alert | Conditions not local to that section. Transient confirmations. A per-row state that needs **no** action — that is a badge. A per-row state needing a decision *is* this, at row scope. |
| Page banner | Success messages. Anything unactionable from this page. Stacking — two page banners means neither is read. |
| Global banner | Anything page-specific. More than one at a time: a stack means the product is in a state nobody designed. |
| Dialog | Information needing no action. Anything that could be inline. Chaining — a dialog opening another is a flow that wanted a page. |
| Badge | Anything carrying an action. The moment a badge needs a button it is an alert. |
| Notice, any placement | More than one action. The molecule has a single action slot, so a message offering a choice is offering the wrong thing: give it one action that opens where the choice belongs. |
| Field error | System failures unrelated to the input. Import or sync problems are not the typist's fault and belong in a banner. |

## Navigation depth — when a resource appears at more than one level

Transactions are a top-level destination. They are also reachable from an
account, a category, and a budget. Left alone that produces four routes to one
list, four empty states, four sets of filters, and breadcrumbs that disagree
about where you are.

**Every resource has exactly one home.** Everywhere else it appears, it is that
same page filtered — not a second copy with its own hierarchy.

> **The test:** can this resource be listed and acted on without naming a parent?
> **Yes** → top-level destination; every other route is a filtered view.
> **No** → genuinely nested; it lives under its parent, with breadcrumbs saying so.

| reached from | verdict | route | what the shell shows |
|---|---|---|---|
| Transactions, from the sidebar | home | `/transactions` | Sidebar item current. No breadcrumb — it is top level. |
| An account's transactions | filtered view | `/transactions?account=current` | Sidebar shows **Transactions** current, not Accounts. A removable filter chip names the account. No breadcrumb. |
| A category's transactions | filtered view | `/transactions?category=groceries` | Same page, same rule. Arriving from Budgets does not change which sidebar item is current. |
| A statement | nested | `/accounts/current/statements/2026-03` | Fails the test — a statement has no meaning without its account. Breadcrumb: Accounts › Current account › March 2026. |
| A budget period | nested | `/budgets/groceries/2026-03` | A period belongs to its budget. Breadcrumb names both. |
| Categories, from Settings | home, one level down | `/settings/categories` | Settings is the destination, Categories its subsection. Secondary nav marks it; the sidebar stays on Settings. |

### What follows

**A filtered view never gets a breadcrumb.** Breadcrumbs express containment and a
filter is not containment. `Accounts › Current account › Transactions` would claim
the list lives under the account — the thing the rule denies. The filter chip
carries that context instead, and removing it widens the list rather than
navigating anywhere.

**The sidebar marks where you are, not how you arrived.** Landing on filtered
transactions from a budget still marks Transactions current. Marking Budgets would
make the sidebar a history rather than a location.

**Depth counts from the documented root, not URL segments.**
`/transactions?account=current` is depth 1, not depth 3.

**Nesting must be earned by the test, not chosen for convenience.** If a resource
can be listed on its own, nesting it costs a duplicate page.

## Rulings specific to money

**A failed money operation never gets a toast.** It gets a persistent surface
carrying the retry. A transient message can be missed, and the user is left
believing money moved.

**Moving, deleting or reclassifying money is a consequential submission**
(SC 3.3.4): review-and-confirm naming the record and the amount, or a reversal
with a stated window. Merging two transactions is reclassifying money, so it
cannot be a single click from a row with neither record visible — the row's
notice offers Review, and the merge is confirmed against both records.

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
