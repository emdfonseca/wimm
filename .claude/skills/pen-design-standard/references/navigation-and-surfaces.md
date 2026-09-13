## 6. Navigation and task surfaces

**In this file:**

- 6.1. Navigation depth model
- 6.2. Which navigation pattern represents which relationship?
- 6.3. Navigation should be visible in journey designs
- 6.4. Persistent sidebar vs drawer/side panel
- 6.5. Choosing between inline, modal, drawer, and full page
- 6.6. CRUD guidance: add, edit, update, delete
- 6.7. Route or overlay? Make it explicit
- 6.8. When an overlay is a local state vs a journey branch
- 6.9. Accessibility contract for navigation and overlays
- 6.10. What belongs in the design system vs the journey?


Navigation needs its own standard because it answers a different question from a user journey.

```text
NAVIGATION
Where am I in the product?

JOURNEY / TASK FLOW
What am I trying to accomplish, and what step am I on?

TASK SURFACE
Where should this particular action happen?
```

Do not use these concepts interchangeably.

### 6.1. Navigation depth model

Depth is the number of parent relationships below the documented product/workspace root. It is not a fixed label for an entity type, a count of URL segments, or a count of visible navigation controls.

```text
Workspace                         L0
└── Projects                      L1
    └── Project Alpha             L2
        ├── Overview              L3
        ├── Activity              L3
        └── Settings              L3
            └── Permissions       L4
```

If a genuine collection level such as Active Projects is inserted between Projects and Project Alpha, descendants move one level deeper. Omit levels that do not exist. Record route mappings separately; routing and information hierarchy may differ.

### 6.2. Which navigation pattern represents which relationship?

Use each pattern for one clear job.

| Pattern | Use it for | Do not use it for |
|---|---|---|
| **Primary navigation** | Major product sections | Steps in a task |
| **Secondary navigation** | Stable subsections inside a primary area | Temporary form states |
| **Breadcrumbs** | Showing hierarchical location and ancestors | Wizard/checkout progress |
| **Tabs** | Switching between peer views of the same context/entity | Representing parent/child hierarchy |
| **Stepper / progress indicator** | Ordered stages of a task | Product navigation |
| **Back action** | Returning to the previous/contextual location | Replacing clear hierarchy when users need direct access |
| **In-page anchors** | Navigating sections of one long document/page | Different product destinations |

A useful rule:

```text
Breadcrumb = where this thing lives
Tabs       = peer views of this thing
Stepper    = progress through a task
```

Do not use breadcrumbs and tabs to express the same dimension.

### 6.3. Navigation should be visible in journey designs

Every journey screen should make the user's current product context understandable.

For each important step, reviewers should be able to answer:

- Which primary section is active?
- Which secondary section or entity is active?
- Is the user still on the same route/context or have they navigated somewhere else?
- What happens when they use Back?
- Where do Cancel and Close return them?
- Does success return to the originating context, a detail page, or a new destination?

The journey overview should therefore include navigation context when relevant:

```text
J12 · EDIT TEAM MEMBER

Entry route: /team/members/:memberId
Navigation hierarchy: Team → Members → Member detail
Entry surface: Member detail page
Task surface: Edit drawer
Success return: Member detail, same route/context
Cancel return: Member detail, no changes
```

On the canvas, do **not** create a separate "navigation journey" merely to repeat the app shell. Use the shared shell/navigation components from the design-system library and show the correct active/current state in each journey frame.

### 6.4. Persistent sidebar vs drawer/side panel

Use consistent vocabulary:

```text
SIDEBAR
Persistent region of the application layout.
Often contains primary or secondary navigation.
It remains part of the page structure.

DRAWER / SIDE PANEL
Temporary or contextual surface that opens beside the current page.
Often used for viewing or editing an item while preserving surrounding context.
It may be modal or non-modal depending on behavior.
```

Do not call a temporary edit form a "sidebar" if the product also has a navigation sidebar.

### 6.5. Choosing between inline, modal, drawer, and full page

Choose the surface based on the **task**, not visual preference.

| Surface | Best when | Avoid when |
|---|---|---|
| **Inline edit** | One small value/action can be changed in place | The change has many fields, dependencies, or serious consequences |
| **Modal dialog** | The user must make a short, focused, blocking decision or complete a compact task | The form is long, needs comparison with the page, contains nested navigation, or may expand substantially |
| **Drawer / side panel** | A moderate task benefits from preserving visible page context, especially list/detail workflows | The task is a major destination, very complex, needs substantial horizontal space, or contains deep subflows |
| **Full page / route** | The task is complex, long-lived, multi-section, route-worthy, recoverable, deep-linkable, or likely to grow | The task is only a tiny local decision |

**Modality describes behavior; drawer describes placement.** These are independent dimensions:

```text
Modal
→ intentionally interrupts
→ foreground decision/task
→ background interaction is unavailable while modal
→ strong focus boundary

Drawer / side panel
→ preserves spatial relationship with current context
→ useful when the user benefits from seeing the list/detail behind or beside it
→ may be modal OR non-modal, but that behavior must be explicitly defined
```

### 6.6. CRUD guidance: add, edit, update, delete

`Update` is normally the save/commit action of an edit flow, not a separate surface category.

#### Add / Create

Prefer **inline** when creation is extremely small and naturally belongs in the surrounding UI.

Prefer a **modal** when:

- creation is compact;
- there is one focused group of inputs;
- no nested task is required;
- losing page interaction temporarily is acceptable.

Prefer a **drawer/side panel** when:

- the user benefits from retaining list/detail context;
- the form is moderate rather than tiny;
- the user may need to see information on the underlying page; if they must interact with it, use a non-modal panel or another appropriate surface;
- opening and closing several records is a common workflow.

Prefer a **full page** when:

- the form has multiple sections or steps;
- the task needs its own route or deep link;
- autosave/drafts/recovery matter;
- uploads, previews, complex editors, permissions, or dependent subflows are involved;
- the task is likely to grow.

#### Edit

Use **inline edit** for atomic values such as a name, status, or small property when the consequences are obvious.

Use a **drawer/side panel** for contextual record editing when staying anchored to the list/detail is valuable.

Use a **full page** for complex settings or entity editing with multiple sections, dependencies, or nested tasks.

A **modal** can work for compact edits, but should not become the default container for every form.

#### Delete

For a straightforward irreversible or high-impact destructive action, prefer a **destructive confirmation dialog** close to the point of action. A low-impact reversible action may use an explicit Undo model; define the restoration window and failure behavior. Do not add confirmation solely because the action is named Delete.

The confirmation must explain what is being deleted and, when meaningful, whether the action is reversible.

If deletion involves complex consequences—for example ownership transfer, dependent resources, retention choices, or multiple required decisions—use a larger dedicated flow rather than forcing everything into a small confirmation dialog.

Do not use a drawer merely because the entity was edited in a drawer. The surface should match the deletion task itself.

#### Bulk actions

Keep selection and bulk controls in the list context. Use a confirmation dialog for destructive bulk actions when confirmation is warranted, and clearly state scope/consequences.

### 6.7. Route or overlay? Make it explicit

For every modal or drawer in a journey, define whether it is:

```text
ROUTE-BACKED
The state has a stable route/deep link and can be restored directly.

OR

EPHEMERAL
The state exists only within the current page/task context.
```

This affects Back behavior, refresh/recovery, analytics, implementation, and accessibility.

The canvas should annotate this when it is not obvious:

```text
03 · Member detail
    ↓ Open Edit
03a · Edit member / Drawer / Ephemeral
    ↓ Save
03 · Member detail / Updated
```

or:

```text
03 · Member detail
    ↓ Open Permissions
04 · Permissions / Full page / Route-backed
```

### 6.8. When an overlay is a local state vs a journey branch

Keep a modal/drawer beside the parent step when it is a short local interaction:

```text
02 · Members list
├── Add member / Drawer
├── Delete member / Confirmation dialog
└── Filter / Popover
```

Create a distinct journey branch when the overlay contains meaningful subflow logic, for example:

```text
Invite member
→ choose role
→ SSO restriction discovered
→ request elevated permission
→ invitation succeeds/fails
```

The fact that the UI is visually inside a drawer does **not** mean it is too small to deserve journey documentation.

### 6.9. Accessibility contract for navigation and overlays

Navigation and task surfaces have behavioral requirements, not just visual ones.

At minimum define:

**Navigation**

- current/active destination is programmatically identifiable;
- keyboard order follows a predictable structure;
- navigation labels remain meaningful without relying on icons alone;
- breadcrumbs represent hierarchy and expose the current item correctly;
- actual tab widgets use tab/panel semantics, arrow-key navigation, and an explicit automatic/manual activation model; automatic activation is suitable when switching is effectively immediate;
- links to separate routes keep link/navigation semantics even when styled like tabs; styling does not by itself justify `role="tab"`. See [APG Tabs](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/).

**Modal dialog**

- focus moves into the dialog when it opens;
- focus is contained while the dialog is modal;
- the dialog has an accessible name;
- Close/Cancel behavior is explicit;
- Escape closes the standard modal pattern; define a safe cancellation or discard-confirmation path for unsaved work rather than silently losing it;
- background content is inert while modal; setting `aria-modal="true"` alone does not implement modality;
- focus normally returns to the invoker, or to the next logical destination if the invoker is gone or the completed workflow requires it.

Choose initial focus according to content and task: a heading/static introduction can be appropriate for long structured content; a least-destructive action can suit irreversible decisions. A full-screen visual treatment does not determine whether a surface is a dialog or a routed page. See [APG Modal Dialog](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/).

**Drawer / side panel**

First decide whether it is **modal or non-modal**.

- A modal drawer follows the same focus-boundary expectations as a modal dialog.
- A non-modal contextual panel should not pretend to be modal; background interaction and keyboard movement must remain coherent.
- The panel needs a clear accessible name and predictable open/close behavior.

**Destructive confirmation**

- name the affected object/scope;
- distinguish destructive and cancel actions clearly;
- do not rely on red color alone;
- explain irreversible consequences before commitment.

### 6.10. What belongs in the design system vs the journey?

The **design system** owns reusable navigation/surface building blocks and their behavior, classified at the appropriate Atomic Design level:

```text
TEMPLATES
App Shell / responsive page scaffolds

ORGANISMS
Primary Navigation
Secondary Navigation
Page Header
Modal Dialog
Drawer / Side Panel
Destructive Confirmation
Reusable Form layouts

MOLECULES
Breadcrumbs
Tabs
Stepper
Popover
```

The **journey/product design** owns the actual information architecture and decisions:

```text
Which destinations exist
Which item is active
How deep the route is
Whether a task is inline/modal/drawer/page
What happens on Save/Cancel/Back
Where success and failure lead
What the actual form contains
```

The library defines **how the mechanism behaves**. The journey defines **when and why it is used**.

---

