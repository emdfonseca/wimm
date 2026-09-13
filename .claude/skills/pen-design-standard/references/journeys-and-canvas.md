## 7. Journeys and canvas organization

**In this file:**

- 7.1. What counts as a journey?
- 7.2. When to use a separate journey file
- 7.3. Standard journey file structure
- 7.4. Journey overview
- 7.5. Canvas grammar: left → right means progress
- 7.6. Primary success path
- 7.7. Failure, recovery, alternative success, and exit paths
- 7.8. Local state or separate branch?
- 7.9. Important screen states
- 7.10. Theme context and comparison frames
- 7.11. Journey naming convention


### 7.1. What counts as a journey?

A journey is a **user goal**, not a page or route.

```text
✅ Create account · Recover password · Invite teammate · Export report
❌ Settings page · Modal screens · Wide screens · Forms
```

---

### 7.2. When to use a separate journey file

One file per journey family or product area; closely related journeys share a file.

```text
01-onboarding.pen
02-authentication.pen
05-checkout.pen
06-account.pen
```

Split a file when it has many screens or branches, becomes hard to scan zoomed out, or evolves on a different cadence from its neighbors. Never one file per error state.

---

### 7.3. Standard journey file structure

```text
┌──────────────────────────────────────────────────────────────┐
│ 00 · JOURNEY OVERVIEW                                        │
│ Goal / actor / entry / outcome / key constraints             │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│ 10 · PRIMARY SUCCESS                                         │
│ Main flow, with responsive compositions                      │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│ 20 · ALTERNATIVE / FAILURE / RECOVERY PATHS                  │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│ 30 · STATE / EDGE-CASE INDEX                                 │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│ 40 · THEME + ACCESSIBILITY QA                                │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│ 90 · TEMPORARY NOTES                                         │
└──────────────────────────────────────────────────────────────┘
```

**Local-state placement:** Keep canonical local-state frames beside their parent step. Zone 30 indexes those frames by journey/path/step and may contain large stress cases that cannot fit beside the step. Do not maintain a second editable copy of the same state. Zone 40 contains explicitly labeled QA comparisons with a source-frame reference and reviewed revision.

If setup is local to this file, place the `00 · PROJECT SETUP` frame before the journey overview and group both in the `00` documentation zone. In a multi-journey file, give each journey its own overview and path groups; file-level documentation indexes them.

---

### 7.4. Journey overview

```text
J05 · CHECKOUT

Goal: Complete a purchase
Primary actor: Signed-in customer
Entry point: Cart
Entry route/context: Commerce → Cart
Navigation hierarchy: Product → Commerce → Cart (L0 → L1 → L2)
Task context: Checkout
Primary task surface: Full-page checkout flow
Primary success outcome: Order confirmed
Success destination: Order confirmation / order detail
```

Add prerequisites, permissions, business constraints, analytics events, or related routes only when they carry information.

---

### 7.5. Canvas grammar: left → right means progress

Left → right is progression through a path on the canvas. This is a reading convention; it does not dictate text direction or layout in a localized product.

```text
[01 Cart] → [02 Delivery] → [03 Payment] → [04 Review] → [05 Confirmation]
```

#### Top → bottom means responsive composition

```text
                  01 Cart      02 Delivery      03 Payment      04 Review      05 Confirmation

WIDE              [ frame ]    [ frame ]        [ frame ]       [ frame ]      [ frame ]

MEDIUM            [ frame ]    [ frame ]        [ frame ]       [ frame ]      [ frame ]

COMPACT           [ frame ]    [ frame ]        [ frame ]       [ frame ]      [ frame ]
```

If Medium adds no structural information, do not draw it. Add a note:

```text
Medium follows the Wide shell. Apply Device=Medium only if Medium tokens exist;
otherwise use the adopted shared token context. See the mapping in project setup.
```

---

### 7.6. Primary success path

```text
J05.A · CHECKOUT · PRIMARY SUCCESS

01 Cart → 02 Delivery → 03 Payment → 04 Review → 05 Confirmation
```

The primary line shows normal success only; validation errors sit beside steps, not on this line.

---

### 7.7. Failure, recovery, alternative success, and exit paths

Meaningful changes to the flow become branches. Record one primary path category and optional cause tags; do not duplicate a branch to give it several classifications.

```text
03 Payment
    │
    ├── Card declined
    │      ↓
    │   Choose another card / retry
    │      ↓
    │   Return to Payment
    │
    ├── Authentication required
    │      ↓
    │   Bank verification
    │      ↓
    │   Success → Review
    │      OR
    │   Failure → Recovery
    │
    └── Payment service unavailable
           ↓
        Retry / exit
```

Controlled labels for primary category and cause tags:

- **PRIMARY SUCCESS** — normal intended completion;
- **ALTERNATIVE SUCCESS** — another valid completion route;
- **FAILURE** — the goal cannot be completed;
- **RECOVERY** — steps that let the user recover;
- **CANCEL / EXIT** — intentional abandonment;
- **PERMISSION / ACCESS** — authentication or authorization blocks progress;
- **SYSTEM INTERRUPTION** — network/service/timeout/dependency issue.

---

### 7.8. Local state or separate branch?

#### Keep it beside the screen when:

- the user stays on the same journey step;
- it affects only one screen;
- the recovery action is obvious and immediate;
- no meaningful navigation/business-path change occurs.

```text
03 · Payment / Compact

[ Default ]   [ Validation error ]   [ Submitting ]   [ Server error ]
```

#### Create a branch when:

- navigation changes;
- several additional steps appear;
- recovery is non-trivial;
- business rules alter the path;
- the user can reach a different end state;
- the scenario requires explicit product review.

For every documented transition, identify the trigger, precondition, resulting state, data preserved or changed, destination/rejoin step, and focus/status behavior. A timeout leaves the operation's result unknown; do not label it a confirmed failure without evidence.

---

### 7.9. Important screen states

```text
PRIMARY        [ Default ]
TRANSIENT      [ Loading ] [ Submitting ]
RESULT         [ Success ] [ Error ]
CONTENT        [ Empty ] [ Partial ] [ Long / max content ]
ACCESS/SYSTEM  [ Offline ] [ Permission denied ] [ Session expired ] [ Service unavailable ]
```

Draw the states that change behavior, accessibility, content, layout, or implementation — not every permutation.

---

### 7.10. Theme context and comparison frames

Use the product-default Color value in the main flow. Set the frame's context explicitly when needed; do not manually recolor descendants (`pen-mechanics.md` §2.7).

Zone 40 contains selected comparisons for every adopted presentation mode, focusing on forms, feedback, overlays, focus, and data graphics. Label each QA frame with its source frame and reviewed revision. Side-by-side copies are test fixtures, not separate product specifications.

---

### 7.11. Journey naming convention

Journey IDs are stable when a journey moves or is renamed; retired IDs are never reused. Filenames identify journey families/product areas and may have optional ordering prefixes; the prefix need not equal a journey ID.

Use lowercase kebab-case filenames. Use sentence case for human-readable names; uppercase zone/status labels are visual signage. Theme identifiers retain their exact declared casing. A fully qualified frame name is `J05.A / 03 · Payment / Compact / Validation error`; short names below are acceptable inside an unambiguous journey/path group. Include the full identifier in review links, annotations, and exports. These human identifiers are distinct from pen.dev object IDs.

Allocate `.A` to primary success and stable additional path IDs to meaningful branches. Keep a step identifier stable; record display order separately if inserting a step would otherwise force renumbering. Responsive and local-state representations share the conceptual step ID. Themes may appear in QA labels for comparison, but not as a suffix on main-flow names.

#### Journey IDs

```text
J01 · Create account
J02 · Sign in
J05 · Checkout
J12 · Edit team member
```

#### Branches

```text
J05.A · Checkout · Primary success
J05.B · Checkout · Card declined
J05.C · Checkout · Payment timeout
J05.E · Checkout · Item becomes unavailable
```

#### Steps

```text
01 · Cart
02 · Delivery
03 · Payment
04 · Review
05 · Confirmation
```

#### Responsive frames

```text
03 · Payment / Wide
03 · Payment / Medium
03 · Payment / Compact
```

#### Local states

```text
03 · Payment / Compact / Default
03 · Payment / Compact / Validation error
03 · Payment / Compact / Submitting
```

Do not append `Light` or `Dark` to frame names when the frame's `Color` theme value already expresses it.

---
