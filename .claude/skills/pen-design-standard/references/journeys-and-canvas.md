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

A journey is a **user goal**, not simply a page or route.

Good journey names:

```text
Create account
Recover password
Complete first purchase
Upgrade subscription
Invite teammate
Create project
Export report
Cancel subscription
```

Less useful journey names:

```text
Settings page
Modal screens
Desktop screens
Forms
New UI
```

---

### 7.2. When to use a separate journey file

For a small product, closely related journeys can share a `.pen` file.

For a medium or large product, use one file per major journey family or product area.

Examples:

```text
01-onboarding.pen
02-authentication.pen
05-checkout.pen
06-account.pen
```

Create a separate file when one or more are true:

- the journey contains many screens or branches;
- different teams own the area;
- the file becomes difficult to scan at zoomed-out level;
- the journey has substantial independent review/handoff work;
- the area evolves on a different cadence from neighboring journeys.

Do **not** create separate journey files for every small error state.

---

### 7.3. Standard journey file structure

Every journey file should follow approximately the same top-to-bottom zoning:

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
│ Move obsolete work to the app design/archive/ directory.     │
└──────────────────────────────────────────────────────────────┘
```

---

**Local-state placement:** Keep canonical local-state frames beside their parent step. Zone 30 indexes those frames by journey/path/step and may contain large stress cases that cannot fit beside the step. Do not maintain a second editable copy of the same state. Zone 40 contains explicitly labeled QA comparisons with a source-frame reference and reviewed revision.

If setup is local to this file, place the Section 4.15 setup frame before the journey overview and group both in the `00` documentation zone. In a multi-journey file, give each journey its own overview and path groups; file-level documentation indexes them.

---

### 7.4. Journey overview

Start the file with a compact overview.

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

For complex journeys add only the useful context:

```text
Prerequisites
Permissions / roles
Business constraints
Important analytics events
Related feature/route
Owner / team
```

---

### 7.5. Canvas grammar: left → right means progress

Use **left → right** for progression through a path on the documentation canvas. This canvas-reading convention does not dictate text direction, navigation order, or layout in a localized product. Specify right-to-left behavior independently.

```text
[01 Cart] → [02 Delivery] → [03 Payment] → [04 Review] → [05 Confirmation]
```

Do not place steps wherever empty canvas exists.

#### Top → bottom means responsive composition

Within a path, organize meaningful layout regimes vertically:

```text
                  01 Cart      02 Delivery      03 Payment      04 Review      05 Confirmation

DESKTOP           [ frame ]    [ frame ]        [ frame ]       [ frame ]      [ frame ]

TABLET            [ frame ]    [ frame ]        [ frame ]       [ frame ]      [ frame ]

MOBILE            [ frame ]    [ frame ]        [ frame ]       [ frame ]      [ frame ]
```

If Tablet adds no useful design information, do not duplicate it merely to complete the matrix. Add a note such as:

```text
Tablet follows the Wide shell. Apply Device=Tablet only if Tablet tokens exist;
otherwise use the adopted shared token context. See the mapping in project setup.
```

---

### 7.6. Primary success path

Every journey has one clearly identified primary success path.

```text
J05.A · CHECKOUT · PRIMARY SUCCESS

01 Cart → 02 Delivery → 03 Payment → 04 Review → 05 Confirmation
```

Keep the primary flow clean. It answers:

> What happens when the intended journey succeeds normally?

Do not interrupt this line with every possible validation error.

---

### 7.7. Failure, recovery, alternative success, and exit paths

Meaningful changes to the flow become branches. Record one primary path category and optional cause tags. Categories can describe outcome or purpose; permission and system conditions may be cause tags on a recovery path. Do not infer category from an identifier or duplicate a branch just to give it several classifications.

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

Use the following controlled labels for primary category and optional cause tags:

- **PRIMARY SUCCESS** — normal intended completion;
- **ALTERNATIVE SUCCESS** — another valid completion route;
- **FAILURE** — the goal cannot be completed;
- **RECOVERY** — steps that let the user recover;
- **CANCEL / EXIT** — intentional abandonment;
- **PERMISSION / ACCESS** — authentication or authorization blocks progress;
- **SYSTEM INTERRUPTION** — network/service/timeout/dependency issue.

Example path names (category is a separate metadata field):

```text
J05.A · Checkout · Primary success
J05.B · Checkout · Card declined
J05.C · Checkout · Payment timeout
J05.D · Checkout · User cancels authentication
J05.E · Checkout · Item becomes unavailable
```

---

### 7.8. Local state or separate branch?

This decision keeps journey files from becoming either incomplete or enormous.

#### Keep it beside the screen when:

- the user stays on the same journey step;
- it affects only one screen;
- the recovery action is obvious and immediate;
- no meaningful navigation/business-path change occurs.

Example:

```text
03 · Payment / Mobile

[ Default ]   [ Validation error ]   [ Submitting ]   [ Server error ]
```

#### Create a branch when:

- navigation changes;
- several additional steps appear;
- recovery is non-trivial;
- business rules alter the path;
- the user can reach a different end state;
- the scenario requires explicit product review.

---

For every documented transition, identify the trigger, precondition, resulting state, data preserved or changed, destination/rejoin step, and focus/status behavior. A timeout can leave the operation's result unknown; do not label it a confirmed failure without evidence. Define how status is checked and when retry is safe with engineering.

---

### 7.9. Important screen states

Capture states that materially affect behavior, accessibility, content, layout, or implementation.

Typical categories:

```text
PRIMARY
[ Default ]

TRANSIENT
[ Loading ] [ Submitting ]

RESULT
[ Success ] [ Error ]

CONTENT
[ Empty ] [ Partial ] [ Long / max content ]

ACCESS / SYSTEM
[ Offline ] [ Permission denied ] [ Session expired ] [ Service unavailable ]
```

Do not draw every theoretical permutation.

---

### 7.10. Theme context and comparison frames

Use the product-default Color value in the main flow. Set the frame's context explicitly when needed; do not manually recolor descendants. Configure themes using Section 2.7 and the responsive mapping in Section 5.3.1.

Zone 40 contains selected comparisons for every adopted presentation mode, focusing on forms, feedback, overlays, focus, and data graphics. Label each QA frame with its source frame and reviewed revision. Side-by-side copies are intentional test fixtures; they are not separate product specifications. Check the implementation scope using Section 10.2.

---

### 7.11. Journey naming convention

Maintain an app-scoped journey registry containing ID, goal, owning file, owner, status, and related journeys. IDs remain stable when a journey moves or is renamed; do not reuse retired IDs. Filenames identify journey families/product areas and may have optional ordering prefixes. Their prefixes need not equal a journey ID.

Use lowercase kebab-case filenames. Use sentence case for human-readable names; uppercase zone/status labels are visual signage. Theme identifiers retain their exact declared casing. A fully qualified frame name is `J05.A / 03 · Payment / Mobile / Validation error`; short names below are acceptable inside an unambiguous journey/path group. Include the full identifier in review links, annotations, and exports. These human identifiers are distinct from pen.dev object IDs.

Allocate `.A` to primary success and stable additional path IDs to meaningful branches. Keep a step identifier stable; record display order separately if inserting a step would otherwise force renumbering. Responsive and local-state representations share the conceptual step ID. Optional themes may appear in QA labels for comparison, but not as a mandatory suffix on main-flow names.

#### Journey IDs

```text
J01 · Create account
J02 · Sign in
J03 · View dashboard
J04 · Search
J05 · Checkout
J06 · Manage account
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
03 · Payment / Desktop
03 · Payment / Tablet
03 · Payment / Mobile
```

#### Local states

```text
03 · Payment / Mobile / Default
03 · Payment / Mobile / Validation error
03 · Payment / Mobile / Submitting
```

Do not append `Light` or `Dark` to every frame name when the frame's `Color` theme value already expresses it.

---

