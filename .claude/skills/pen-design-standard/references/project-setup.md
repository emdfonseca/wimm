## 4. Project setup

**In this file:**

- 4.1. The minimum output
- 4.2. Project scope and outcome
- 4.3. Actors, roles, permissions, and account states
- 4.4. Journeys and paths
- 4.5. Navigation and information architecture
- 4.6. Task-surface decisions: inline, modal, drawer, or page
- 4.7. Responsive strategy and viewport regimes
- 4.8. Theme and variable-axis decisions
- 4.9. Accessibility requirements
- 4.10. Design-system reuse and change scope
- 4.11. Content, data, localization, and edge conditions
- 4.12. Implementation, routing, and platform constraints
- 4.13. Ownership, review, and source of truth
- 4.14. Open decisions and assumptions
- 4.15. Where the Project Setup Record lives
- 4.16. Copy-paste Project Setup Record
- 4.17. Definition of Ready for detailed design


Before creating detailed journey frames, each project must complete a **Project Setup Record** to the readiness level in Section 4.17.

The purpose is not to predict every design decision. It is to make the decisions that determine **how the work must be organized** before the canvas grows.

Every category below must be reviewed. Record a decision, `N/A — reason`, or an open decision with question, assumption, owner, decision due date, and impact. Do not silently skip a category.

This prevents common problems such as:

```text
Designing every screen before agreeing what the journey is
Creating Desktop/Tablet/Mobile frames without agreed layout regimes
Using Device themes as if they were breakpoints
Discovering Dark mode requirements late
Mixing product navigation with task progress
Using modals/drawers inconsistently for the same class of task
Creating local components that duplicate the design system
Treating accessibility as a final QA pass
Discovering engineering/routing constraints after handoff
```

### 4.1. The minimum output

Before detailed design starts, the project should be able to state:

```text
1. What user/business problem is in scope?
2. Who are the relevant actors/roles?
3. What journeys are in scope?
4. What is the primary success outcome for each journey?
5. Which important failure/recovery paths must be designed?
6. Where does the work live in the product navigation hierarchy?
7. Which task surfaces are expected: inline, modal, drawer, full page?
8. Which responsive layout regimes must be designed explicitly?
9. Which theme axes/values are required, if any?
10. What accessibility baseline and project-specific requirements apply?
11. Which design-system library/components/tokens must be reused?
12. What new reusable system work might be required?
13. What implementation/routing/data constraints affect design?
14. Who approves the journey, design-system changes, and implementation readiness?
```

These answers become the project's **design contract**. If a decision changes later, update the record rather than relying on team memory.

---

### 4.2. Project scope and outcome

Answer:

- What problem are we solving?
- What is explicitly in scope?
- What is explicitly out of scope?
- What user outcome defines success?
- What business outcome defines success, if different?
- Is this a new capability, redesign, migration, experiment, or incremental change?
- Is there an existing implementation whose behavior must be preserved?
- Are there hard launch, platform, legal, or migration constraints?

**Why this matters to the design organization:** it determines whether we are creating a new journey, changing an existing one, or mainly updating the design system.

**pen.dev consequence:** decide which journey `.pen` file owns the work, whether an exploration file is needed first, and whether existing library assets should remain the source of truth.

---

### 4.3. Actors, roles, permissions, and account states

Answer:

- Who can enter this journey?
- Are there different roles such as owner, admin, member, guest, reviewer, or operator?
- Do roles see different actions, fields, navigation, or outcomes?
- Can users enter while signed out, partially onboarded, suspended, read-only, or offline?
- What permission-denied states are possible?
- Can the user request access or recover from insufficient permission?
- Does the same journey behave differently for new vs existing users?

Do not represent every permission combination as a separate full journey unless the behavior truly diverges.

**pen.dev consequence:** permission differences may become journey branches, local states, content overrides, or genuinely different compositions. Generic reusable permission UI belongs at the appropriate Atom/Molecule/Organism level in the design system; actual role logic belongs in the journey.

---

### 4.4. Journeys and paths

For each journey answer:

- What user goal names the journey?
- What are the entry points?
- What is the primary success path?
- What is the success outcome/destination?
- What alternative success paths matter?
- Which failures are local states?
- Which failures require a recovery branch?
- Which exits matter: Cancel, Close, Back, timeout, abandonment?
- Can the user resume later?
- Are drafts/autosave required?
- Are there irreversible steps?
- Are there handoffs to another user, device, service, or channel?

Classify each important state:

```text
LOCAL STATE
User remains at the same conceptual step.
Example: invalid email, loading, retryable server error.

JOURNEY BRANCH
The user enters additional decisions or recovery steps.
Example: expired verification → request new code → return to verification.
```

**pen.dev consequence:** this determines the horizontal journey structure, where branches sit, and which states stay beside a screen instead of becoming separate files.

---

### 4.5. Navigation and information architecture

Answer before drawing screens:

- Where does the journey live in the product hierarchy?
- What is the entry route/context?
- What is the expected navigation depth?
- Which primary navigation item is active?
- Is there stable secondary navigation?
- Are breadcrumbs needed because the user is moving through hierarchy?
- Are tabs needed because the user switches between peer views?
- Is a stepper needed because the user is progressing through an ordered task?
- What should browser/system Back do?
- Where should Save, Cancel, Close, and success return the user?
- Which destinations need stable/deep-linkable URLs?
- Which states are allowed to be ephemeral?

Write the hierarchy explicitly when it is more than one obvious level:

```text
Workspace
└── Team
    └── Members
        └── Member detail
            └── Permissions
```

**pen.dev consequence:** journey frames must show the correct shell/current state using imported navigation components; the hierarchy itself is documented in the journey, not encoded as a theme.

---

### 4.6. Task-surface decisions: inline, modal, drawer, or page

For Add/Create/Edit/Delete and other local tasks, answer:

- Does the user need to preserve visible context from the current page?
- Is the task atomic or multi-section?
- Is interruption intentional?
- Does the task need substantial horizontal space?
- Does it contain nested tasks or navigation?
- Must it be deep-linkable/bookmarkable?
- Should refresh restore it?
- Does browser Back close the surface or navigate away?
- Can the user safely dismiss it without saving?
- Does the task need draft/autosave behavior?
- Is deletion simple confirmation or a multi-decision destructive flow?

Record the selected pattern, for example:

```text
Add member      → Drawer / modal behavior / ephemeral
Edit member     → Drawer / modal behavior / ephemeral
Permissions     → Full page / route-backed
Delete member   → Destructive confirmation dialog
Rename member   → Inline edit
```

If a drawer is used, explicitly answer **modal or non-modal**.

**pen.dev consequence:** the journey uses the reusable Molecule/Organism surface from the library but documents the actual route, open/close, focus, save, cancel, and recovery behavior locally.

---

### 4.7. Responsive strategy and viewport regimes

Do not start by asking “which devices should we draw?” Ask “where does the product structure change?”

Answer:

- What is the minimum product layout width, and how will the web implementation satisfy the reflow baseline in Section 5.1.5?
- What is the maximum/content-max behavior on wide screens?
- At which widths does navigation materially change?
- Which canonical App Shell regime applies at each meaningful width (for example Compact, Medium, Wide)?
- Does the design system already provide those shell compositions, or must the project introduce a new reusable shell regime?
- At which widths do columns stack or disappear?
- At which widths do task surfaces change behavior?
- Does a desktop drawer become a mobile full-screen/page treatment?
- Which layouts need explicit design compositions?
- Can Tablet inherit Desktop behavior with fluid layout, or does it genuinely differ?
- Which intermediate widths must be tested even if they are not kept as permanent canvas frames?
- Are there orientation requirements?
- Are data tables/charts/editors especially sensitive to width?

Record the project's explicit design regimes, for example:

```text
Compact / Mobile   → representative frame: 390 px
Wide / Desktop     → representative frame: 1440 px
Tablet             → no separate canvas row; uses Wide structure at widths ≥ 768 CSS px
                     and Compact structure below 768 CSS px
```

or:

```text
Mobile  → 390 px representative frame
Tablet  → 768 px representative frame
Desktop → 1440 px representative frame
```

These representative widths are **design/review frames**, not a declaration that the UI only works at those exact widths.

**pen.dev consequence:** create a separate screen frame for each meaningful responsive composition. Set its width explicitly. Use flex/Hug/Fill inside it. Do not expect a theme axis to resize the frame.

---

### 4.8. Theme and variable-axis decisions

For every potential theme dimension answer:

- Is this a real independent product/system dimension?
- Does it change systematic variable values across many components?
- Is it user-selectable, system-selected, design-only context, or implementation context?
- Do we need to view values side-by-side for QA, or is toggling enough?
- Does the dimension affect appearance only, values only, or actual structure?

#### Color

Answer:

- Is Light required?
- Is Dark required?
- Does the product follow OS/system preference, user preference, or product setting?
- Is High Contrast a real distinct product mode?
- Are any surfaces intentionally fixed to one color treatment?

Typical decision:

```text
Color axis
├── Light
└── Dark
```

#### Device

Only create a `Device` theme axis if the project/system benefits from **systematic responsive token values**.

Ask:

- Do spacing, typography, control sizes, or other reusable numeric tokens change consistently by layout regime?
- Are those differences broad enough to justify central variables rather than local values?
- Do engineering tokens have a corresponding concept?

Possible decision:

```text
Device axis
├── Mobile
├── Tablet
└── Desktop
```

Important:

```text
Device axis ≠ frame width
Device axis ≠ breakpoint engine
Device axis ≠ media query
```

A Mobile journey frame may therefore be:

```text
Frame width = 390
Theme:
  Color  = Light
  Device = Mobile
```

**pen.dev consequence:** theme axes are created in Variables and values are assigned to frames through Theme. Separate frames are still required to visualize meaningful responsive compositions.

---

### 4.9. Accessibility requirements

Accessibility is part of project setup, not only final QA.

Answer:

- Confirm WCAG 2.2 AA for web and record any additional platform, contractual, or legal requirements.
- Are there contractual/legal/platform accessibility requirements beyond the baseline?
- How will all functionality support keyboard operation, apart from applicable WCAG exceptions?
- What important focus transitions exist?
- Which dynamic statuses need assistive-technology announcements?
- What text zoom/reflow expectations must be supported?
- Are there touch-target constraints?
- Is motion used, and is reduced-motion behavior needed?
- Are charts/data visualizations present?
- Are drag interactions present, and what non-drag alternative exists?
- Are complex tables, trees, grids, editors, or custom controls involved?
- How will every adopted Color mode satisfy the baseline?
- Are there known high-contrast or forced-colors requirements?

**pen.dev consequence:** Foundations and reusable Atoms/Molecules/Organisms/Templates supply accessibility constraints, while journey frames/annotations show the context-specific focus, order, error, status, and recovery expectations engineering must implement.

---

### 4.10. Design-system reuse and change scope

Before creating local UI, answer:

- Which `.lib.pen` library does this project consume?
- Which existing Atoms, Molecules, Organisms, or Templates cover the work?
- Are any existing reusable elements missing required states or behaviors?
- Does the project require a genuinely new reusable Atom, Molecule, Organism, or Template?
- Is a proposed difference truly reusable, or journey-specific?
- Does a token need to change globally?
- Does a new semantic token need to be introduced?
- Are there deprecated components that must not be used?
- Who owns approval of library changes?
- Will changing a library component unintentionally affect existing journeys?

Use this classification:

```text
Reuse unchanged          → consume library instance
Reuse with content       → instance override / slot content
New indivisible control  → Atom candidate
Small reusable group     → Molecule candidate
Reusable section/pattern → Organism candidate
Reusable page scaffold   → Template candidate
Product-specific screen  → Page in the journey, not the shared library
Journey-only behavior    → keep in journey
Global visual rule       → foundation variable/token change
```

**pen.dev consequence:** journey files should import the existing library rather than recreating controls locally. Shared changes are made at the library origin and reviewed for global impact.

---

### 4.11. Content, data, localization, and edge conditions

Answer:

- What content is real vs placeholder?
- What are realistic shortest and longest labels/values?
- Are names, filenames, URLs, IDs, numbers, currencies, dates, or addresses unusually long?
- Is localization required now or likely soon?
- Which languages materially increase text length?
- Are right-to-left languages in scope?
- What are zero/empty/min/max data cases?
- Are there large collections, pagination, filtering, or search?
- Are uploads involved? What file states/failures matter?
- Are slow network/offline states relevant?
- What happens when data changes while the user is editing?
- How are truncation and full-content access handled with keyboard, touch, and assistive technology?
- Which locale, time-zone, date/number/currency, pluralization, and bidirectional-content rules apply?
- What sorting, filtering, selection, and pagination state persists after navigation or mutation?

**pen.dev consequence:** include representative stress cases beside the relevant component/journey rather than creating a separate generic “edge cases” product file.

---

### 4.12. Implementation, routing, and platform constraints

Design should not discover these constraints at handoff.

Answer with engineering/product:

- What platforms are in scope: web, desktop, mobile web, native?
- What routing model exists?
- Which states can have stable URLs?
- How are modal/drawer routes handled today?
- Which responsive breakpoints, layout tokens, or layout utilities already exist in code?
- Which design tokens already exist in code?
- Is Dark mode already implemented? How is preference stored/resolved?
- Are there existing shell/navigation constraints?
- Are there server-driven states or permissions that affect available actions?
- Are optimistic updates used?
- What validation happens client-side vs server-side?
- What loading, latency, timeout, or retry behaviors are realistic?
- Are there browser/platform limitations?
- Are analytics events required at particular steps/outcomes?
- Are security/privacy requirements relevant to content shown in UI?

Do not blindly copy existing implementation limitations into new design, but make deliberate deviations visible rather than accidental.

**pen.dev consequence:** annotate meaningful implementation contracts in the journey overview or relevant steps and keep token/component naming aligned with code where practical.

---

### 4.13. Ownership, review, and source of truth

Answer:

- Who owns the product decision?
- Who owns the journey design?
- Who owns design-system review?
- Who owns accessibility review?
- Who owns engineering implementation decisions?
- Where is the source-of-truth journey file?
- Where is the source-of-truth design-system library?
- What status vocabulary will be used?
- What does `READY FOR BUILD` mean for this project?
- Who can approve deviations from the design system?
- Who closes unresolved decisions?

**pen.dev consequence:** filenames, status labels, and review areas should make ownership/readiness understandable without relying on chat history.

---

### 4.14. Open decisions and assumptions

Do not block all design work because every question is unresolved. Instead make uncertainty explicit.

Use:

```text
OPEN DECISION
Question:
Current assumption:
Owner:
Decision needed by:
What changes if the assumption is wrong:
```

Examples:

```text
Question: Does Tablet need its own composition?
Assumption: No; Wide structure remains valid at widths ≥ 768 CSS px.
Owner: Named design owner + frontend owner
Decision needed by: Before responsive review
Impact if wrong: Add a distinct composition only to affected steps.
```

```text
Question: Should Edit Member be route-backed?
Assumption: Ephemeral modal drawer.
Owner: Named product owner + frontend owner
Decision needed by: Before route/overlay implementation
Impact if wrong: Back/refresh/deep-link behavior and journey branching change.
```

This is better than silently encoding an assumption into dozens of screens.

---

### 4.15. Where the Project Setup Record lives

Treat the Project Setup Record as a **living source of truth**, not a kickoff questionnaire that disappears after the first meeting.

For a project centered on one main journey file, place a compact frame at the top-left of the canvas:

```text
00 · PROJECT SETUP
Scope / actors / journeys / navigation / responsive / themes / accessibility / open decisions
```

Keep it visually separate from the actual journey flow so nobody mistakes documentation for a product screen.

For a larger initiative spanning several journey files, keep the full Project Setup Record in project-level documentation (for example a repository Markdown file) and put a short reference/summary in each journey file. Do not duplicate the entire record across every `.pen` file; duplicated setup information will drift.

The distinction is:

```text
PROJECT SETUP RECORD
Project-level decisions that shape all design work.
Examples: supported responsive regimes, Color/Device axes, navigation model, accessibility baseline, library source, routing constraints.

JOURNEY OVERVIEW
Journey-specific context.
Examples: user goal, actor, entry point, success outcome, branches, return destination.
```

If a project-level decision changes, update the Project Setup Record first and then update affected journey files/components.

---

### 4.16. Copy-paste Project Setup Record

Create this in the agreed project source of truth and surface a compact version in pen.dev when it helps reviewers.

```text
PROJECT SETUP — <project / feature>

STATUS
Owner:
Design owner:
Engineering owner:
Design-system reviewer:
Accessibility reviewer:
Target milestone:
pen.dev app/extension version:
Reviewed design revision:
Implementation/code revision:
Status (Section 9.1):

SCOPE
Problem:
In scope:
Out of scope:
Primary user outcome:
Business outcome:

ACTORS / PERMISSIONS
Primary actor:
Other roles:
Permission differences:
Authentication/account-state constraints:

JOURNEYS
Primary journey:
Entry point(s):
Primary success outcome:
Alternative success paths:
Important failure/recovery paths:
Exit/cancel/resume behavior:

NAVIGATION
Entry route/context:
Navigation hierarchy/depth:
Active primary/secondary navigation:
Breadcrumbs/tabs/stepper requirements:
Back behavior:
Success destination:
Cancel/close destination:

TASK SURFACES
Create/Add:
Edit:
Delete:
Other overlays:
Route-backed vs ephemeral decisions:
Modal vs non-modal drawer decisions:

RESPONSIVE
Minimum product layout width / reflow verification:
Maximum/content-max behavior:
Explicit canvas compositions:
Representative frame widths:
Structural change points:
Intermediate widths to test:

THEMES / VARIABLES
Color axis required: Yes / No
Color values:
Device axis required: Yes / No
Device values:
Other real theme axes:
Representative side-by-side theme QA needed:

ACCESSIBILITY
Baseline:
Project-specific requirements:
Important focus transitions:
Supported browser / assistive-technology test matrix:
Dynamic announcements:
Zoom/reflow expectations:
Motion/reduced-motion:
Complex widgets/interactions:

DESIGN SYSTEM
Library path + revision:
Token authority + sync direction:
Atoms/Molecules/Organisms/Templates to reuse:
Potential new shared Atomic Design assets:
Potential token changes:
Deprecated patterns to avoid:

CONTENT / DATA
Localization:
Long-content stress cases:
Empty/zero/max cases:
Network/offline/error considerations:

IMPLEMENTATION
Platforms:
Routing constraints:
Existing code breakpoints/layout rules:
Existing code tokens/theme behavior:
Validation/data constraints:
Analytics requirements:
Security/privacy constraints:

OPEN DECISIONS
1.
2.
3.

DEFINITION OF READY
[ ] Every setup category has a decision, a justified N/A, or a tracked open decision
[ ] Unknowns have question + assumption + owner + decision due date + impact
[ ] Primary journey and success outcome agreed
[ ] Navigation context agreed
[ ] Responsive compositions agreed
[ ] Theme axes/values agreed
[ ] Accessibility baseline agreed
[ ] Design-system source identified
[ ] Major route/overlay decisions agreed
[ ] Engineering constraints reviewed
```

---

### 4.17. Definition of Ready for detailed design

Detailed journey design can begin when:

- the primary user goal and scope are understood;
- the main actors/permission differences are known;
- the primary journey and success outcome are agreed;
- major failure/recovery paths are identified, even if not fully designed;
- navigation context and return destinations are understood;
- major task-surface decisions are made or explicitly open;
- responsive layout regimes are defined well enough to know which frames must exist;
- theme requirements are defined well enough to know which axes/values must exist;
- the accessibility baseline is explicit;
- the design-system source of truth is known;
- meaningful implementation constraints have been reviewed with engineering;
- unresolved decisions record question, assumption, owner, decision due date, and impact.

This is a **readiness gate, not a waterfall gate**. Small unknowns can remain open. What we avoid is building a large canvas on top of invisible assumptions.

---

