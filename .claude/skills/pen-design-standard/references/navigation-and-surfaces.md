## 6. Navigation and task surfaces

**In this file:**

- 6.2. Which navigation pattern represents which relationship?
- 6.4. Persistent sidebar vs drawer/side panel
- 6.5. Choosing between inline, modal, drawer, and full page
- 6.7. Route or overlay? Make it explicit
- 6.9. Accessibility contract for navigation and overlays


### 6.2. Which navigation pattern represents which relationship?

| Pattern | Use it for | Do not use it for |
|---|---|---|
| **Primary navigation** | Major product sections | Steps in a task |
| **Secondary navigation** | Stable subsections inside a primary area | Temporary form states |
| **Breadcrumbs** | Showing hierarchical location and ancestors | Wizard/checkout progress |
| **Tabs** | Switching between peer views of the same context/entity | Representing parent/child hierarchy |
| **Stepper / progress indicator** | Ordered stages of a task | Product navigation |
| **Back action** | Returning to the previous/contextual location | Replacing clear hierarchy when users need direct access |
| **In-page anchors** | Navigating sections of one long document/page | Different product destinations |

```text
Breadcrumb = where this thing lives
Tabs       = peer views of this thing
Stepper    = progress through a task
```

Do not use breadcrumbs and tabs to express the same dimension. Navigation depth is the count of parent relationships below the documented root (`Workspace L0 → Projects L1 → Project Alpha L2`), not URL segments.

### 6.4. Persistent sidebar vs drawer/side panel

```text
SIDEBAR             Persistent region of the application layout; usually navigation.
DRAWER / SIDE PANEL Temporary surface beside the current page; modal or non-modal, stated explicitly.
```

Do not call a temporary edit form a "sidebar" if the product also has a navigation sidebar.

### 6.5. Choosing between inline, modal, drawer, and full page

| Surface | Best when | Avoid when |
|---|---|---|
| **Inline edit** | One small value/action can be changed in place | The change has many fields, dependencies, or serious consequences |
| **Modal dialog** | The user must make a short, focused, blocking decision or complete a compact task | The form is long, needs comparison with the page, contains nested navigation, or may expand substantially |
| **Drawer / side panel** | A moderate task benefits from preserving visible page context, especially list/detail workflows | The task is a major destination, very complex, needs substantial horizontal space, or contains deep subflows |
| **Full page / route** | The task is complex, long-lived, multi-section, route-worthy, recoverable, deep-linkable, or likely to grow | The task is only a tiny local decision |

**Modality describes behavior; drawer describes placement.** A drawer may be modal or non-modal; say which. Delete uses a destructive confirmation dialog that names the affected object, or an explicit Undo model with a defined restoration window; a deletion with dependent decisions (ownership transfer, retention) gets its own flow.

### 6.7. Route or overlay? Make it explicit

For every modal or drawer in a journey, define whether it is:

```text
ROUTE-BACKED   The state has a stable route/deep link and can be restored directly.
EPHEMERAL      The state exists only within the current page/task context.
```

This affects Back behavior, refresh/recovery, analytics, implementation, and accessibility. Annotate it on the canvas:

```text
03 · Member detail
    ↓ Open Edit
03a · Edit member / Drawer / Ephemeral
    ↓ Save
03 · Member detail / Updated

03 · Member detail
    ↓ Open Permissions
04 · Permissions / Full page / Route-backed
```

### 6.9. Accessibility contract for navigation and overlays

**Navigation**

- current/active destination is programmatically identifiable;
- keyboard order follows a predictable structure;
- navigation labels remain meaningful without relying on icons alone;
- breadcrumbs represent hierarchy and expose the current item correctly;
- actual tab widgets use tab/panel semantics, arrow-key navigation, and an explicit automatic/manual activation model;
- links to separate routes keep link semantics even when styled like tabs; styling does not justify `role="tab"`. See [APG Tabs](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/).

**Modal dialog**

- focus moves into the dialog when it opens;
- focus is contained while the dialog is modal;
- the dialog has an accessible name;
- Close/Cancel behavior is explicit;
- Escape closes; unsaved work gets a safe cancellation or discard-confirmation path;
- background content is inert while modal; `aria-modal="true"` alone does not implement modality;
- focus returns to the invoker, or to the next logical destination if the invoker is gone.

Initial focus: a heading for long structured content; the least-destructive action for irreversible decisions. See [APG Modal Dialog](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/).

**Drawer / side panel**

- a modal drawer follows the modal dialog contract;
- a non-modal panel keeps background interaction and keyboard movement coherent;
- the panel has an accessible name and predictable open/close behavior.

**Destructive confirmation**

- name the affected object/scope;
- distinguish destructive and cancel actions clearly;
- do not rely on red color alone;
- explain irreversible consequences before commitment.

---
