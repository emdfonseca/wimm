# pen.dev Product Design Organization Standard

> **Purpose:** Define how we organize product design in pen.dev so multiple user journeys, success and failure paths, responsive layouts, light/dark themes, accessibility requirements, reusable Atomic Design assets, and design-to-code work stay understandable as the product grows.
>
> **Audience:** Product designers, design-system designers, engineers, product managers, and reviewers working with `.pen` files.
>
> **Accessibility baseline:** WCAG 2.2 Level AA for web experiences. Record additional platform, contractual, and legal requirements in project setup; they may require separate assessment.
>
> **Core rule:** **Journeys own product flow. The design system owns reusable rules. Accessibility applies to both.**

---

**Review date:** 2026-09-13. **Revision:** Technical and consistency review. **Adoption status:** Proposed standard; organizational approval is separate from this review.

This is a reference standard for organizing design work, with setup and review procedures. It covers design artifacts, reusable assets, interaction contracts, and implementation verification. It does not prescribe a frontend framework or replace a product specification, security review, or full accessibility evaluation.

**Requirement language:** “must” and “do not” express requirements of this team standard; “should” and “prefer” express recommendations that permit a documented reason to depart; “may” and “optional” identify choices. These are team rules unless explicitly attributed to an external specification. Examples illustrate the rules and do not independently create requirements.

**Applicability:** Project setup selects supported platforms, layout regimes, and presentation modes. References to Light/Dark, Mobile/Tablet/Desktop, or other optional capabilities apply only when adopted. Mark an inapplicable checklist item `N/A` with a reason. A narrower representative canvas does not reduce the implementation's accessibility obligations.

**Authority:** The detailed clauses define the standard; examples and checklists apply them. Record approved team-convention exceptions with scope, rationale, owner, impact, and review date. An internal exception does not waive an external requirement.

**Contents**

1. [Scope and organizing model](#1-scope-and-organizing-model)
2. [pen.dev capabilities and limits](#2-pendev-capabilities-and-limits)
3. [Repository and ownership model](#3-repository-and-ownership-model)
4. [Project setup](#4-project-setup)
5. [Cross-cutting standards](#5-cross-cutting-standards)
6. [Navigation and task surfaces](#6-navigation-and-task-surfaces)
7. [Journeys and canvas organization](#7-journeys-and-canvas-organization)
8. [Design-system library](#8-design-system-library)
9. [Lifecycle and change management](#9-lifecycle-and-change-management)
10. [Verification and readiness gates](#10-verification-and-readiness-gates)
11. [Design-to-code contract](#11-design-to-code-contract)
12. [Worked example: create account](#12-worked-example-create-account)

[Appendix A: pen.dev translation](#appendix-a--pendev-translation-cheat-sheet) · [Appendix B: sources](#appendix-b--references)

---

## 1. Scope and organizing model

A product design contains several different kinds of information. The biggest source of canvas disorder is representing all of them by duplicating screens.

We separate them instead.

| Concern | What it means | Where we represent it |
|---|---|---|
| **Journey** | A user goal, such as creating an account | Journey `.pen` file or major journey area |
| **Path** | Primary success, alternative success, failure, recovery, exit | Branches inside the journey |
| **Step** | A stage in the journey | Frames arranged left → right |
| **Navigation hierarchy** | Where the user is in the product | App shell + primary/secondary navigation + breadcrumbs/tabs where appropriate |
| **Task surface** | Where a local action happens without confusing it with product navigation | Inline edit, modal dialog, drawer/side panel, or full page |
| **Responsive structure** | Layout changes between mobile/tablet/desktop | Explicit frames/compositions + pen.dev flex layout |
| **Responsive token values** | Systematic spacing/type/size changes by layout regime | Optional `Device` theme axis + variables |
| **Light/Dark** | Appearance mode | `Color` theme axis + semantic variables |
| **Local UI state** | Loading, validation error, empty, disabled, etc. | Beside the relevant journey step |
| **Reusable UI hierarchy** | Atoms, molecules, organisms, and templates | Shared `.lib.pen` design library |
| **Accessibility** | Requirements that make all of the above usable | Cross-cutting rules in Foundations, reusable Atomic Design layers, Pages/Journeys, implementation, and QA |

The target architecture is therefore:

```text
PROJECT SETUP RECORD
Scope → actors → journeys → navigation → surfaces → responsive → themes → accessibility → implementation

JOURNEY FILES
User goal → paths → steps → responsive compositions → contextual states

DESIGN SYSTEM LIBRARY
Foundations → atoms → molecules → organisms → templates

CROSS-CUTTING STANDARDS
Accessibility + navigation rules + naming + responsive rules + theme rules + implementation contracts
```

### The rule that prevents most duplication

Do **not** build this matrix on the canvas:

```text
journey × path × breakpoint × theme × state × component variant
```

Instead:

```text
Canvas        = journey × path × meaningful responsive composition
Themes        = variable-driven context
Local states  = examples beside the relevant step
System rules  = shared design library across foundations/atoms/molecules/organisms/templates
Accessibility = requirements applied across all layers
```

---

## 2. pen.dev capabilities and limits

Before discussing our organization, these are the pen.dev mechanisms the standard relies on.

### 2.1. `.pen` document

A normal `.pen` file is a design document on an infinite canvas. We use normal `.pen` files for **product journeys, explorations, and prototypes**.

Examples:

```text
01-onboarding.pen
02-authentication.pen
checkout.pen
account.pen
```

### 2.2. `.lib.pen` design library

A design library is a `.pen` document turned into a reusable library. pen.dev gives design-library files the `.lib.pen` suffix.

A library can provide reusable:

- components;
- variables;
- themes attached to those variables.

We use `.lib.pen` for the **design system**, not for journey-specific product flows.

Example:

```text
product-ui.lib.pen
```

Journey documents import this library and consume its assets.

Create it through **Libraries → Turn this file into a library**, then import the saved library into each consuming document. The documentation says this conversion cannot be undone in the editor; retain version history before converting. Renaming a suffix alone is not the documented creation workflow.

**Update contract:** Save the source library, then reopen consuming documents to load changes. Existing instance overrides take precedence. Treat library adoption as a reviewed update, including representative consumer checks; it is not a promise of live refresh across open documents. See [pen.dev Design Libraries](https://docs.pen.dev/core-concepts/design-libraries).

### 2.3. Frames

Frames are the main containers we use for:

- screens;
- layout regions;
- journey groups;
- theme context;
- flex layout.

A frame can contain other frames and components.

### 2.4. Flex layout, Hug, and Fill

pen.dev frames can use horizontal or vertical flex layout. In the UI, select a frame and use flex layout to define direction, gap, padding, and alignment.

For resizing behavior:

- **Hug Width / Hug Height** makes a flex frame size around its children;
- **Fill Width / Fill Height** makes a child use available space in its parent flex layout;
- fixed dimensions are appropriate when a viewport or structural region has an intentional size.

This is the main mechanism for **fluid behavior inside a responsive composition**. Avoid a circular sizing dependency: a parent cannot reliably Hug a dimension if every child Fills that same dimension. Give the layout a fixed or content-sized basis. Clipping hides overflow; it does not specify runtime scrolling. See [pen.dev Interface](https://docs.pen.dev/core-concepts/pencil-interface).

The documented layout model is flexbox-style, not full CSS Flexbox/Grid. The current published layout schema does not expose CSS `flex-wrap`; annotate intended wrapping and implement/test it in code rather than assuming native canvas support. See [the .pen format](https://docs.pen.dev/for-developers/the-pen-format).

### 2.5. Variables

Variables in pen.dev work like design tokens or CSS custom properties.

They can represent values such as:

```text
color.bg.surface
color.text.primary
layout.sectionGap
radius.control
layout.pageGutter
type.size.pageTitle
```

Instead of hardcoding the property on every object, the object references the variable. Variables have Color, Number, String, or Boolean types; bind them only to supported properties of the matching type. Same-type aliases can express semantic roles. See [pen.dev Variables](https://docs.pen.dev/core-concepts/variables).

### 2.6. Theme values and theme axes

This document uses the term **theme axis** because that is how pen.dev models independent theme dimensions.

A theme axis is **not a canvas axis** and it is **not a CSS media query**.

It is a named dimension in pen.dev's Variables system.

For example:

```text
Theme axis: Color
Values: Light, Dark
```

and independently:

```text
Theme axis: Device
Values: Mobile, Tablet, Desktop
```

A variable can define values conditional on an axis value or a combination of axis values. Axis names and values in this standard are team conventions, not built-in device detection.

Example:

```text
color.bg.canvas
  Color=Light  → light surface value
  Color=Dark   → dark surface value

layout.pageGutter
  Device=Mobile  → 16
  Device=Tablet  → 24
  Device=Desktop → 32
```

### 2.7. Creating and applying theme axes

In the **Variables** panel:

1. Define the variables you need.
2. Add another value/column to an existing theme axis for values such as `Light` and `Dark`.
3. Add a **separate theme axis** when you need an independent dimension such as `Device`.
4. Give each variable the appropriate value for the relevant theme value.

Then on the canvas:

1. Select a frame.
2. In the properties panel, use **Theme → Add theme**.
3. Choose the axis and value, for example `Color = Dark` or `Device = Mobile`.
4. Child objects inherit that theme selection unless they explicitly override the same axis.

Conceptually:

```text
Mobile checkout screen
Theme:
  Color  = Light
  Device = Mobile

└── child components inherit both values
    ├── colors resolve through Color=Light
    └── device-aware numbers resolve through Device=Mobile
```

#### Defaults and resolution

The first value of each axis is the document default. Removing a local theme selection restores inheritance. Populate every supported context, and inspect parent and instance overrides when a value resolves unexpectedly. Keep component origins free of fixed `Color`/`Device` selections unless their contract intentionally requires one. See [pen.dev Variables](https://docs.pen.dev/core-concepts/variables).

For file-level tooling, variable value lists use the **last matching entry**; a more specific condition does not automatically beat a later broad condition. Put an unconditional fallback first, broad conditions next, and intended combination overrides last. Avoid overlapping ambiguous conditions and alias cycles. See [the .pen format](https://docs.pen.dev/for-developers/the-pen-format).

### 2.8. Theme context and side-by-side comparison

The canvas does **not** automatically create one visible copy for every theme value.

A frame resolves **one value from each axis at a time**.

For example, this frame can simultaneously use:

```text
Checkout / Mobile
Theme:
  Color  = Dark
  Device = Mobile
```

That means two independent axes are active on the same frame, but only one value from each axis is selected for that frame.

If you change the frame to:

```text
Color = Light
```

that **same frame** re-renders with the Light values. Its children inherit the new selection unless they override the `Color` axis.

Think of this as a context selector, not a canvas dimension:

```text
Variables define what is possible:

Color  → Light | Dark
Device → Mobile | Tablet | Desktop

A frame selects its current context:

Frame A → Color=Light + Device=Desktop
Frame B → Color=Dark  + Device=Desktop
Frame C → Color=Light + Device=Mobile
```

#### Can Light and Dark be visible at the same time?

Yes, but **only when you intentionally place separate frames/instances on the canvas and give them different theme selections**.

For example, a focused QA comparison may show:

```text
THEME QA

                  LIGHT                     DARK
Checkout form     [same composition]        [same composition]
Error state       [same composition]        [same composition]
Dialog            [same component]          [same component]
```

The Light frame has `Color=Light`; the Dark frame has `Color=Dark`.

This is useful for QA. It should **not** become a rule that duplicates every journey screen in every theme.

#### What does not happen automatically

pen.dev does not automatically produce this matrix:

```text
Desktop Light | Desktop Dark
Tablet Light  | Tablet Dark
Mobile Light  | Mobile Dark
```

We create only the responsive compositions that communicate meaningful structural differences. We toggle themes on those frames during design/review, and create side-by-side theme copies only for representative QA cases.

---

### 2.9. Critical limitation: `Device` does not make a design automatically responsive

`Device` is a **team-defined theme axis** used to resolve variable values. It has no automatic relationship to the width of a frame.

Changing a frame from 1440 px to 390 px does **not** automatically change:

```text
Device=Desktop → Device=Mobile
```

There are no implicit media-query semantics in this convention.

Therefore responsive design in pen.dev uses **two separate mechanisms**:

```text
1. STRUCTURAL RESPONSIVENESS
   Explicit Desktop / Tablet / Mobile compositions
   + flex layout, Hug, Fill, alignment
   + annotated wrapping/overflow requirements for implementation

2. RESPONSIVE TOKENS
   Optional Device theme axis
   + variables for systematic values such as gutter or type size
```

When we show a Mobile frame, we explicitly assign `Device = Mobile` to that frame if it consumes Device-aware variables.

### 2.10. Components, instances, and slots

Reusable components have an origin and instances. Instances inherit changes except where a property is overridden; imported-library changes follow the save/reopen workflow in Section 2.2.

Slots provide designated areas in a component where instance-specific content can be inserted. In the editor, **Make slot** applies to an empty frame inside a component origin. Suggested slot components guide selection; they do not restrict allowed content. See [pen.dev Slots](https://docs.pen.dev/core-concepts/slots).

Inherited instance layers cannot be freely reparented in the editor. Use slots for supported composition. Detaching disconnects the outer instance, but nested instances and variable references remain linked. See [pen.dev Components](https://docs.pen.dev/core-concepts/components).

Use these for flexible system components such as:

```text
Card
├── Header
├── Content slot
└── Actions slot
```

### 2.11. Atomic Design terms are taxonomy, not pen.dev object types

This standard uses the Atomic Design vocabulary **Atom → Molecule → Organism → Template → Page** to describe the scope of reusable design. These are not native pen.dev object types.

In pen.dev:

- an **Atom** can be implemented as a reusable component;
- a **Molecule** can be a component composed of Atoms;
- an **Organism** can be a larger component/composition built from Atoms and Molecules;
- a **Template** can be a reusable page-level component/composition with slots;
- a **Page** is usually a product-specific screen Frame in a journey `.pen` file.

The important distinction is:

```text
Atomic Design term = level of composition and reuse
pen.dev component  = implementation mechanism
```

Do not assume every pen.dev component is a Molecule, and do not create separate technical mechanisms for each Atomic Design level.

---

### 2.12. Design representations and executable behavior

Canvas arrows, state frames, accessibility annotations, and the `prototypes/` folder are organizational representations. They do not establish routing, focus management, live announcements, or interactive state transitions by themselves. Record the preview/runtime used for any executable prototype and which behaviors it exercises. Verify product behavior in the target implementation.

---

## 3. Repository and ownership model

The design organization has two distinct homes:

1. **Reusable design-system work** lives with the shared design-system package.
2. **Product journey work** lives with the deployable app it describes.

For a pnpm workspace, the recommended shape is:

```text
my-product/
│
├── apps/
│   └── web/
│       ├── src/
│       └── design/
│           ├── journeys/
│           │   ├── 01-onboarding.pen
│           │   ├── 02-authentication.pen
│           │   ├── 03-dashboard.pen
│           │   ├── 04-search.pen
│           │   ├── 05-checkout.pen
│           │   └── 06-account.pen
│           ├── explorations/
│           │   ├── navigation-exploration.pen
│           │   └── checkout-redesign.pen
│           ├── prototypes/
│           │   └── optional-prototype.pen
│           └── archive/
│               └── deprecated-work/
│
└── packages/
    └── design-system/
        ├── src/
        └── design/
            └── product-ui.lib.pen
```

This is the **canonical organizational model** used throughout this document. A small product may temporarily keep several closely related journeys in one `.pen` file, but that is a scaling choice—not a different naming model. Do not create one design file per route or page.

### What belongs where?

**Journey files answer:**

> How does this user accomplish this product goal, including success, failure, recovery, states, and responsive behavior?

**The design-system library answers:**

> What reusable Foundations, Atoms, Molecules, Organisms, Templates, interaction contracts, accessibility rules, and tokens should every journey use consistently?

Do not mix those responsibilities into one giant file.

The `apps/` and `packages/` names are team conventions. A pnpm workspace requires a root `pnpm-workspace.yaml`; actual packages need manifests and declared dependencies. The tree above omits ordinary code/build files. A design folder is not automatically a package. See [pnpm Workspaces](https://pnpm.io/workspaces).

Keep the reusable library beside its shared implementation. Create a separate tokens package only when tokens have independent consumers, a build pipeline, or ownership needs. Do not create packages per journey or force code folders to mirror Atomic Design labels. Repository boundaries follow deployability and code ownership; journey files follow goals and product areas.

For non-monorepos, preserve these ownership relationships within the existing repository structure instead of creating unnecessary packages. A journey file can cover a product area such as projects or billing; a family filename need not itself be phrased as a user goal. The journeys inside it must identify concrete goals. Splitting criteria are in Section 7.2.

---

### 3.1. Design, token, and product authority

> **The design system owns Foundations plus reusable Atoms, Molecules, Organisms, and Templates. Journey files own realized Pages, product-specific composition, content, and flow.**

| Item | Owner |
|---|---|
| Semantic color definitions | Design system |
| Button appearance and generic accessibility contract | Design system |
| Button used to submit checkout | Checkout journey |
| Generic error banner | Design system |
| Exact payment failure copy and recovery route | Checkout journey |
| Dialog shell and focus contract | Design system |
| Cancel-subscription confirmation content | Account journey |
| Mobile navigation structure | Design system if shared |
| Which navigation item is active | Journey/screen |
| Generic form validation treatment | Design system |
| Which field fails and what happens next | Journey |

---

#### 3.1.1. Authority, revisions, and shared screens

Record separate authorities for product decisions, reusable visual/interaction specifications, tokens, and executable behavior. “Source of truth” does not mean both design and code may overwrite each other independently. Choose one authoritative token source and a reviewed synchronization direction; generated counterparts are projections of that source.

When a Page participates in several journeys, name one owner and canonical frame for each shared composition/state. Other journeys reference that frame or use a clearly labeled contextual/QA instance. Link the canonical location, purpose of the copy, and reviewed revision. Promote generic mechanics, not an entire business Page merely to avoid a reference. Cross-app flows record one coordinating journey owner and links to app-owned portions and handoff contracts.

When a library, token, or shared Page changes, review its registered consumers, update affected journey references, and reconcile the code mapping. Local overrides must document intent; overrides to accessibility-critical properties need the same review as the underlying contract.

---

## 4. Project setup

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

## 5. Cross-cutting standards

### 5.1. Accessibility

Accessibility is **not one folder inside the design system** and it is **not a theme axis**.

It applies across the product:

```text
ACCESSIBILITY REQUIREMENTS
          ↓
┌──────────────────────────────┐
│ Foundations                  │
│ Atoms / Molecules            │
│ Organisms / Templates        │
│ Journey flows                │
│ Content                      │
│ Responsive behavior          │
│ Light/Dark themes            │
│ Engineering semantics        │
│ QA                           │
└──────────────────────────────┘
```

Use **WCAG 2.2 Level AA** as the web baseline. AA includes applicable Level A and AA criteria and the conformance requirements for full pages and complete processes. Project-specific obligations may need additional assessment. See [WCAG 2.2 conformance](https://www.w3.org/TR/WCAG22/#conformance).

#### 5.1.1. Accessibility in foundations

The design system should establish accessible constraints for:

- semantic foreground/background color combinations;
- focus indication;
- typography and text legibility;
- scalable spacing and layout;
- target sizing;
- disabled-state treatment;
- motion and reduced-motion behavior where motion is part of the product;
- error, warning, success, and informational feedback;
- states that do not rely on color alone.

#### 5.1.2. Accessibility in reusable Atomic Design layers

Atoms, Molecules, Organisms, and Templates should define an **accessibility contract** appropriate to their scope, not only appearance.

Examples:

**Button**

```text
Visual states:
Default / Hover / Focus / Pressed / Disabled

Implementation contract:
Semantic button when it performs an action
Visible focus treatment
Accessible name required
Disabled behavior defined
```

**Form field**

```text
Label relationship defined
Hint/help relationship defined
Required treatment defined
Error message treatment defined
Error cannot rely only on red color
Focus behavior defined
```

**Dialog**

```text
Initial focus expectation
Focus containment expectation
Escape/close behavior
Return-focus expectation
Accessible title/name expectation
```

pen.dev can communicate the visual and interaction specification, but the design artifact alone does not guarantee correct HTML/native semantics or assistive-technology behavior. Those requirements must also be implemented and tested in code.

#### 5.1.3. Accessibility in journeys

An Atom, Molecule, Organism, or Template can be accessible in isolation while the full user journey is still inaccessible.

For every important journey ask:

```text
Can the goal be completed using keyboard-only interaction?
Is focus order logical through the entire flow?
When a new step opens, where should focus go?
When validation fails, is the problem identified and recoverable?
Are status changes communicated without requiring visual detection?
Does the journey remain usable at supported zoom/text scaling?
Do responsive versions preserve access to the same capabilities?
Does Dark mode preserve sufficient contrast and state recognition?
```

#### 5.1.4. Accessibility is not an `Accessibility` theme axis by default

Do not model general accessibility as:

```text
Accessibility = On / Off
```

Accessibility is required in the default experience.

A theme axis may make sense when the product intentionally supports a distinct presentation mode selected by a user or system policy, for example:

```text
Contrast
├── Standard
└── High Contrast
```

Even then, the standard mode must still meet the accessibility baseline. A product high-contrast theme is not evidence of operating-system forced-colors support; record and verify that implementation behavior separately when applicable.

---

#### 5.1.5. Measurable web accessibility baseline

The following is a design and implementation review minimum, not an exhaustive WCAG checklist. Measure the rendered implementation as well as the design specification. `CSS px` means CSS pixels, not hardware pixels or screenshot pixels.

| Concern | Required check and relevant distinction |
|---|---|
| Text contrast | At least 4.5:1 for ordinary text and 3:1 for large text: at least 18 pt (24 CSS px), or 14 pt bold (about 18.67 CSS px), with equivalent sizing for CJK fonts. Do not round a failing ratio up. Inactive controls, incidental text, and logotypes have defined exceptions. [SC 1.4.3](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html) |
| Non-text contrast | Visual information needed to identify controls, their states, and meaningful graphics needs 3:1 against adjacent colors, subject to the criterion's exceptions. This is not a blanket contrast requirement for every decorative border. [SC 1.4.11](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html) |
| Text resizing | Text can grow to 200% without loss of content or functionality, except captions and images of text. A fixed-height control must not clip enlarged text. [SC 1.4.4](https://www.w3.org/WAI/WCAG22/Understanding/resize-text.html) |
| Reflow | Vertically scrolling content must work at 320 CSS px width without two-dimensional scrolling; horizontally scrolling content has a corresponding 256 CSS px height condition. Content requiring two-dimensional layout has a limited exception. A 1280 CSS px viewport at 400% zoom is a useful web test. A 390 px design frame alone is insufficient evidence. [SC 1.4.10](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html) |
| Text spacing | In supporting markup, tolerate user overrides of line height to 1.5 times font size, paragraph spacing to 2 times font size, letter spacing to 0.12 em, and word spacing to 0.16 em without losing content or functionality. These are tolerance tests, not mandatory default typography. Account for the criterion's language/script exceptions. [SC 1.4.12](https://www.w3.org/WAI/WCAG22/Understanding/text-spacing.html) |
| Focus | Keyboard focus must be visible. Under AA SC 2.4.11, author-created content must not entirely obscure the focused component; keeping it fully visible is the team preference. Focus Appearance (2.4.13) and Focus Not Obscured (Enhanced, 2.4.12) are AAA, not AA. [Focus guidance](https://www.w3.org/WAI/WCAG22/Understanding/focus-not-obscured-minimum.html) |
| Pointer targets | SC 2.5.8 AA requires at least 24 × 24 CSS px or a qualifying spacing, equivalent-control, inline, user-agent, or essential exception. For the spacing exception, 24 CSS px diameter circles centered on undersized targets must not intersect another target or another such circle. A 44 × 44 target is a useful stronger design choice; do not label it the AA minimum. [SC 2.5.8](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html) |
| Dragging | Provide a single-pointer alternative that does not require dragging unless an applicable exception applies. Keyboard support alone does not satisfy this pointer requirement. [SC 2.5.7](https://www.w3.org/WAI/WCAG22/Understanding/dragging-movements.html) |
| Authentication | Avoid requiring memory, transcription, or puzzle-solving without an allowed alternative, assistance mechanism, or exception. Support password managers and paste; make verification-code entry compatible with paste/autofill where supported. [SC 3.3.8](https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum.html) |
| Repeated input and help | Reuse or offer previously supplied information within the same process, subject to applicable exceptions. Keep repeated help mechanisms in consistent relative order. [SC 3.3.7](https://www.w3.org/WAI/WCAG22/Understanding/redundant-entry.html), [SC 3.2.6](https://www.w3.org/WAI/WCAG22/Understanding/consistent-help.html) |
| Consequential submissions | For covered legal, financial, and user-data changes, provide reversal, input checking with correction, or review/confirmation before final submission as appropriate. A confirmation dialog is one possible method. [SC 3.3.4](https://www.w3.org/WAI/WCAG22/Understanding/error-prevention-legal-financial-data.html) |

Disabled-state styling should remain understandable under this team standard, even where WCAG contrast exceptions apply. Define the actual disabled behavior; a dimmed appearance alone is not an implementation contract.

Reduced-motion support is a team expectation when motion is used. SC 2.3.3 on disabling interaction-triggered animation is AAA; do not present every reduced-motion recommendation as AA. Other applicable timing, flashing, and moving-content requirements remain part of the baseline. See [Animation from Interactions](https://www.w3.org/WAI/WCAG22/Understanding/animation-from-interactions.html).

#### 5.1.6. Coverage beyond visual components

Maintain a criterion-level applicability record using the [W3C WCAG Quick Reference](https://www.w3.org/WAI/WCAG22/quickref/). Include non-text alternatives; captions/audio description where applicable; meaningful reading order and relationships; orientation and input purpose; keyboard operation and traps; timing and moving/flashing content; page titles and bypass navigation; pointer gestures/cancellation and label-in-name; language; predictable navigation; instructions and error suggestions; and programmatic names, roles, values, and status messages.

For native applications, document the applicable platform semantics, assistive technologies, and assessment method separately. Canvas review and this web checklist do not establish native conformance. For all platforms, representative QA helps discover problems; it does not certify untested pages or processes.

---

### 5.2. Theme policy

Use themes for **systematic contextual values**, not for duplicating complete screens.

Recommended independent axis:

```text
Color
├── Light
└── Dark
```

Do not create combined values such as:

```text
Mobile Light
Mobile Dark
Desktop Light
Desktop Dark
```

If responsive token differences are useful, add a separate axis:

```text
Device
├── Mobile
├── Tablet
└── Desktop
```

Independence matters because a frame can then resolve both:

```text
Color=Dark + Device=Mobile
```

without creating a special `Mobile Dark` mode.

---

### 5.3. Responsive policy

Responsive design is based on **layout behavior**, not a catalog of devices.

Avoid designing every popular width:

```text
320 / 360 / 375 / 390 / 412 / 430 / 768 / 820 / 1024 / 1280 / 1440 / 1920
```

Instead define layout regimes based on when the product actually changes structure.

Example only:

```text
Mobile   width < 768 CSS px
Tablet   768 ≤ width < 1200 CSS px
Desktop  width ≥ 1200 CSS px
```

The exact values belong to the product and implementation. Use non-overlapping ranges, including the exact boundary and fractional widths. The example above means `width < 768`, `768 ≤ width < 1200`, and `width ≥ 1200`, measured in CSS pixels for web.

#### 5.3.1. Canonical responsive vocabulary

Use **Compact / Medium / Wide** for structural shell regimes. This standard retains **Mobile / Tablet / Desktop** as the exact `Device` axis values and short responsive row labels, with this explicit mapping:

| Shell regime | Device value / row label | Representative width in the three-regime example |
|---|---|---|
| Compact | Mobile | 390 px |
| Medium | Tablet | 768 px |
| Wide | Desktop | 1440 px |

These labels do not identify hardware or input capabilities. A tablet can use the Wide shell. Adopt only meaningful regimes and token contexts; a two-regime product can omit Medium/Tablet. If Tablet tokens share Wide structure, document that mapping rather than creating a duplicate shell. Do not mix `Device=Compact` and `Device=Mobile` in the same system. The journey and library clauses apply this vocabulary.

#### 5.3.2. Responsive behavior contract

For each structural change, record the controlling measurement (viewport or containing region), exact ranges, shell mapping, content order, visibility alternatives, overflow/scroll regions, and task-surface transformations. Record representative frame width **and height**, content maxima, sticky regions, and virtual-keyboard behavior where relevant. Avoid guessing touch, hover, or keyboard capability from width alone.

Implementation QA covers just below, at, and above each boundary, intermediate and extreme widths, short viewports, content growth, supported orientations, and the accessibility conditions in Section 5.1.5. A component may respond to its container independently of the app shell; annotate that requirement and implement it explicitly. In a static canvas, describe intended overflow and wrapping rather than using clipping to conceal missing behavior.

Within each composition, prefer flex behavior, Hug, Fill, and tokenized spacing over manual positioning.

---

## 6. Navigation and task surfaces

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

## 7. Journeys and canvas organization

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

## 8. Design-system library

### 8.1. Library structure, taxonomy, and responsive shells

Use a dedicated library:

```text
packages/design-system/design/product-ui.lib.pen
```

Its responsibility is to define the reusable product language.

This includes the reusable **mechanics** of navigation and task surfaces—such as app shell, breadcrumbs, tabs, modal dialogs, drawers, and confirmation patterns—but **not** the product's actual route hierarchy or journey-specific form content.

It should **not** become a catalog of complete business journeys.

Recommended organization:

```text
00 · README / PRINCIPLES

10 · FOUNDATIONS
     ├── Color / semantic tokens
     ├── Typography
     ├── Spacing
     ├── Sizing
     ├── Radius
     ├── Borders
     ├── Elevation
     ├── Motion guidance
     ├── Iconography
     └── Accessibility foundations

20 · ATOMS
     ├── Button
     ├── Icon Button
     ├── Input Control
     ├── Checkbox
     ├── Radio
     ├── Switch
     ├── Badge
     ├── Avatar
     ├── Divider
     └── Spinner / Progress primitive

30 · MOLECULES
     ├── Form Field
     ├── Search Field
     ├── Select Field
     ├── Button Group
     ├── Breadcrumbs
     ├── Tabs
     ├── Pagination
     ├── Toast
     └── Compact Card / summary unit

40 · ORGANISMS
     ├── Primary Navigation
     ├── Secondary Navigation
     ├── Page Header
     ├── Table / Data Region
     ├── Filter Bar / Filter Panel
     ├── Modal Dialog
     ├── Drawer / Side Panel
     ├── Empty State
     ├── Error State
     ├── Destructive Confirmation
     ├── Create / Edit Form
     └── Authentication Form

50 · TEMPLATES
     ├── App Shell / Compact
     ├── App Shell / Medium (when genuinely distinct)
     ├── App Shell / Wide
     ├── App Shell / Reference Examples
     ├── List Page Scaffold
     ├── Detail Page Scaffold
     ├── Settings Page Scaffold
     └── Dashboard Scaffold

90 · QA / STRESS TESTS

99 · DEPRECATED
```

The examples above are a **classification standard, not a purity test**. A complex Card might be a Molecule in one system and an Organism in another. Classify by composition, responsibility, and reuse scope, then stay consistent. Do not move an element between levels merely because its implementation became more complex.

#### 8.1.1. Atomic Design taxonomy used in this standard

Atomic Design provides a useful hierarchy for reusable UI, but Foundations and Journeys sit outside the core atomic ladder.

| Layer | Definition in this standard | Typical examples | Lives where? |
|---|---|---|---|
| **Foundations** | Tokens, constraints, and principles used by every layer | color, type, spacing, focus, motion | `.lib.pen` design system |
| **Atom** | Smallest reusable UI unit that still has a meaningful contract | Button, Checkbox, Input Control, Badge | `.lib.pen` design system |
| **Molecule** | Small functional composition of Atoms with one focused purpose | Form Field, Search Field, Breadcrumbs | `.lib.pen` design system |
| **Organism** | Larger reusable section or interaction composed from Atoms/Molecules | Navigation, Table region, Drawer, Form | `.lib.pen` design system |
| **Template** | Reusable page-level structural scaffold with regions/slots but no journey-specific business content | App Shell, List/Detail/Settings scaffold | `.lib.pen` design system |
| **Page** | A realized screen: a Template/Organisms populated with actual product content, state, and navigation context | Checkout review, Project detail | journey `.pen` file |
| **Journey** | A goal over time that sequences Pages, overlays, states, failures, and recovery | Create account, Invite teammate | journey `.pen` file |

Two rules matter most:

1. **Pages are not shared design-system assets by default.** They are realized product screens and belong in journeys.
2. **Journeys are not another Atomic Design level.** They describe time, branching, and user goals across Pages and states.

#### 8.1.2. Responsive App Shells are a design-system responsibility

The design system should provide a **canonical App Shell composition for every meaningful layout regime used across the product**. This is different from merely providing navigation components. The shell defines how persistent product structure responds when available layout space changes.

Prefer regime names that describe layout behavior rather than physical devices:

```text
Compact
Medium
Wide
```

A project may map these to representative review frames such as:

```text
Compact → 390 px
Medium  → 768 px
Wide    → 1440 px
```

but the regime represents **available layout space and structural behavior**, not a promise that a particular device always uses that shell.

##### What each shell should define

A responsive App Shell should make these decisions explicit:

- primary navigation placement and collapsed/expanded behavior;
- header structure and available actions;
- content container width and page gutters;
- secondary navigation treatment;
- breadcrumbs or location context;
- placement of persistent contextual panels;
- default page-header behavior;
- how temporary task surfaces such as drawers adapt;
- where global/system feedback appears;
- minimum expectations for keyboard order, focus movement, landmarks, and accessible navigation naming.

Example:

| Concern | Wide | Medium | Compact |
|---|---|---|---|
| Primary navigation | Persistent sidebar | Compact/collapsed navigation | Hidden behind navigation control |
| Header | Full | Reduced | Compact mobile-style header |
| Content | Max-width or multi-column where appropriate | Fluid/reduced columns | Single-column/fluid |
| Page gutter | Large token | Medium token | Small token |
| Secondary navigation | Persistent/tabs | Tabs or compact treatment | Scrollable tabs/dropdown/appropriate compact pattern |
| Context panel | May remain side-by-side | Conditional | Usually drawer or full-page treatment |
| Edit drawer | Side panel | Side panel when space allows | Often full-screen sheet/page |

These are illustrative system defaults to adopt or adjust in project setup, not immutable rules. A product area can justify a different composition, but that exception should be deliberate and documented rather than invented independently inside a journey.

##### Structural shell vs responsive token context

Do not confuse the shell with a `Device` theme axis.

```text
App Shell / Compact
```

is an **actual structural composition**. It changes placement, visibility, hierarchy, or interaction model.

```text
Device = Mobile       # maps to the Compact shell in this example
```

is only a **variable context**. It can resolve values such as:

```text
layout.pageGutter
layout.sectionGap
type.size.pageTitle
control.height.md
```

but it does not create or resize the Compact shell.

The intended relationship is:

```text
Available width / layout regime
        ↓
Canonical App Shell composition
        ↓
Optional Device-aware token values
        ↓
Journey content placed inside the shell
```

##### What the design system owns vs what the project owns

The **design system owns**:

- the canonical Compact/Medium/Wide shell structures;
- reusable navigation/header/breadcrumb/secondary-navigation mechanics;
- shell slots and layout contracts;
- default responsive transformations;
- accessibility contracts for shell/navigation behavior;
- representative reference examples showing intended usage.

The **project/journey owns**:

- the actual routes and destinations;
- labels and information architecture;
- which navigation item is active;
- which shell regime a given frame is demonstrating;
- justified exceptions to the default shell;
- the actual content and task behavior inside the shell.

##### In pen.dev

Keep the reusable shell compositions in the shared `.lib.pen` design-system library. Journey Pages should consume those shell Templates and supporting Organisms rather than redraw the product chrome. For responsive review, create explicitly sized journey frames for the meaningful regimes and place the corresponding shell composition inside each one.

For example:

```text
Desktop journey frame
width = 1440
shell = App Shell / Wide
Device = Desktop       # only if Device-aware tokens are used

Mobile journey frame
width = 390
shell = App Shell / Compact
Device = Mobile        # only if Device-aware tokens are used
```

If the same shell structure works across two regimes through fluid layout alone, do **not** create a fake extra shell variant. Create a separate shell only when structure or interaction genuinely changes.

---

### 8.2. Foundations and token vocabulary

Foundations define values and constraints used throughout the system. **Foundations sit below and across the Atomic Design hierarchy; they are not Atoms.** Atoms, Molecules, Organisms, and Templates should consume foundation variables rather than redefining them locally.

#### Color

Prefer semantic tokens:

```text
color.bg.canvas
color.bg.surface
color.bg.elevated

color.text.primary
color.text.secondary
color.text.disabled
color.text.inverse

color.border.default
color.border.subtle
focus.ring.color

color.action.primary
color.action.onPrimary
color.action.destructive
color.action.onDestructive

color.feedback.success
color.feedback.warning
color.feedback.error
color.feedback.info
```

Avoid product components depending directly on raw hex values when a semantic token exists.

#### Spacing

```text
space.0
space.1
space.2
space.3
space.4
space.6
space.8
space.12
```

#### Layout

```text
layout.pageGutter
layout.contentMax
layout.sidebarWidth
layout.headerHeight
layout.sectionGap
```

#### Typography

```text
type.family.body
type.family.display

type.size.bodySm
type.size.bodyMd
type.size.bodyLg

type.size.headingSm
type.size.headingMd
type.size.headingLg
type.size.pageTitle
```

#### Radius and control sizing

```text
radius.sm
radius.md
radius.lg
radius.full
radius.control

control.height.sm
control.height.md
control.height.lg
```

#### Accessibility foundations

Accessibility belongs in foundations because foundations establish constraints used by every component.

Examples:

```text
focus.ring.color
focus.ring.width
focus.ring.offset

control.target.min
motion.duration.*
motion.reduced.*      where applicable
```

The exact token set should match engineering and the product rather than creating tokens solely for documentation aesthetics.

---

### 8.3. Reusable component contracts

In pen.dev, **Atoms, Molecules, Organisms, and Templates can all be implemented as reusable components/compositions**. The word `component` describes the implementation mechanism; the Atomic Design term describes the element's scope.

Examples:

```text
Atom      → Button component
Molecule  → Form Field component composed from Label + Input + Message
Organism  → Filter Panel component composed from fields/actions
Template  → App Shell composition with navigation/header/content slots
```

Reusable design elements specify UI behavior and structure; the target implementation makes that behavior executable.

A Button might conceptually expose:

```text
Visual emphasis / variant (not an ARIA role)
├── Primary
├── Secondary
├── Ghost
└── Destructive

Size
├── Small
├── Medium
└── Large

Interaction state specification
├── Default
├── Hover
├── Focus
├── Pressed
└── Disabled
```

These are conceptual API dimensions, not a claim that pen.dev provides automatic variant sets or interactive state transitions. Represent them with supported components, instances, overrides, slots, and annotated state examples; document the corresponding code API. Distinguish keyboard focus, momentary press, and persistent selected/toggled state.

Theme should normally come from variables, not separate component copies.

Avoid:

```text
Button / Primary / Light
Button / Primary / Dark
Button / Secondary / Light
Button / Secondary / Dark
```

For icon-only controls, specify an accessible name independently of whether a tooltip is shown. Each state example identifies the trigger and expected behavior; it does not merely name a visual style. A reusable field also specifies label, hint, error, required, read-only, and disabled relationships. Generic accessibility contracts are in Section 5.1.2; overlay and navigation contracts are in Section 6.9.

---

### 8.4. Slots and flexible composition

Use slots when a reusable component has a stable shell but variable child content; this is commonly a Molecule, Organism, or Template.

Example:

```text
Card
├── Heading region
├── Content slot
└── Actions slot
```

This is generally better than creating many almost-identical Card components for every content combination.

Document each slot's purpose, allowed product content by team convention, empty behavior, sizing/overflow, and accessibility responsibility. pen.dev suggestions are advisory (Section 2.10); enforce any runtime restrictions in the code API when needed.

---

## 9. Lifecycle and change management

### 9.1. Artifact status

Names should make the canvas understandable without opening every frame.

Use deterministic labels rather than names such as `Final`, `Latest`, or `Frame 423`.

Use this status vocabulary for reviewed artifacts:

```text
EXPLORING
IN REVIEW
READY FOR BUILD
IMPLEMENTED
DEPRECATED
```

---

#### 9.1.1. Status meaning and evidence

| Status | Meaning / entry evidence |
|---|---|
| EXPLORING | Alternatives or incomplete contracts; not a build specification. |
| IN REVIEW | A named revision is ready for product, design-system, accessibility, and engineering review as applicable. |
| READY FOR BUILD | Section 10.1 passes or records justified N/A items; implementation-blocking decisions are resolved; accountable reviewers and the approved design/library revision are recorded. |
| IMPLEMENTED | The linked code/build implements the approved scope and passes Section 10.2; material deviations are reconciled. It does not automatically mean deployed to users. |
| DEPRECATED | A replacement or retirement rationale, affected consumers, owner, and removal condition are documented. |

Record status, owner, revision, review date, decision/evidence links, and open issues. Status applies to the named artifact or scoped journey, not automatically to every frame in a file. A material design change returns affected work to IN REVIEW; adding an exploratory alternative does not change the approved baseline. Only the accountable owners identified in setup approve readiness and convention exceptions.

---

### 9.2. Reuse and local exceptions

Avoid local copies such as:

```text
checkout.pen
├── local Button
├── local Input
├── local Dialog
└── local Toast
```

when the shared library already provides them.

During exploration, local UI is acceptable. Before a journey becomes implementation-ready:

1. replace duplicated shared controls with imported library instances;
2. promote a genuinely reusable solution into the library; or
3. retain a named, product-specific local composition with an owner and documented rationale, while reusing shared controls inside it.

---

### 9.3. Promotion and variant decisions

Promote a local solution when it is reusable enough to become an Atom, Molecule, Organism, Template, or Foundation rule. Promote when it is:

- repeated across journeys;
- likely to be reused;
- behaviorally consistent across contexts;
- important to standardize;
- risky or expensive to implement differently.

Useful test:

> If this changes later, should all usages update together?

If yes, it is a strong design-system candidate. Then classify it at the **lowest Atomic Design level that fully describes its responsibility**; do not inflate a Molecule into an Organism merely because it has many visual variants.

Before adding a component or frame, ask whether the difference is a variable value, an allowed content override, slot content, a local state, or a structural/behavioral change. Use a new reusable composition when existing mechanisms cannot express the contract clearly. Do not multiply components by every color, device, emphasis, and interaction-state combination.

---

### 9.4. Exploration and archive

Archive superseded approved artifacts with their revision, replacement reference, owner, and reason. Keep stable references resolvable or update their consumers. Delete disposable exploration only when it carries no required decision history; version history should retain approved baselines. Temporary canvas notes are not the permanent decision log.

Use:

```text
apps/<app>/design/explorations/
```

for unconstrained concept work.

Once a direction is accepted:

1. move/refine the accepted solution into the appropriate journey;
2. replace temporary UI with design-system instances;
3. connect semantic variables;
4. define responsive behavior;
5. add important error/recovery paths;
6. perform theme and accessibility review;
7. archive or delete obsolete alternatives.

Do not leave multiple unlabeled “final” versions beside implementation-ready screens.

---

### 9.5. Library release and migration

1. Name the library owner, proposed revision, affected assets, and consumer list. Record changed defaults, overrides, slots, themes, and tokens.
2. Review the source change and perform the Section 10.3 checks before consumer adoption. Preserve stable component and variable identities where practical.
3. Save the source library and reopen representative consuming documents (Section 2.2). Inspect both inherited properties and intentional overrides.
4. Record design/library and code/token revisions together. Coordinate incompatible changes with affected applications; publish replacement and migration instructions before removal.
5. Mark deprecated assets with replacement, owner, and removal condition. Retain them until affected consumers migrate or explicitly retire their usage.
6. Verify references after moving/renaming files. Preserve required fonts and image assets. Copy/paste from an ordinary `.pen` file does not preserve a cross-document source link; resolve variable conflicts deliberately. See [pen.dev Design Libraries](https://docs.pen.dev/core-concepts/design-libraries).
7. Commit reviewed design and code changes with reproducible evidence. Treat `.pen` JSON merges as structural changes: reopen and inspect the result. Record a known-good revision and coordinated rollback procedure.

This release record is a team workflow. A library filename, file conversion, or pnpm package version does not by itself guarantee compatible design imports.

---

## 10. Verification and readiness gates

### 10.1. Ready for build

#### Project alignment

- [ ] The Project Setup Record exists and reflects the current scope.
- [ ] Open decisions record question, assumption, owner, decision due date, and impact; none block the proposed implementation scope.
- [ ] Journey screens match the agreed navigation hierarchy and task-surface model.
- [ ] Responsive frames match the agreed layout regimes rather than an arbitrary device catalog.
- [ ] Theme axes/values used by the journey match the project setup decision.
- [ ] Design-system reuse/new-system-work decisions are still valid.
- [ ] Implementation constraints that materially affect the experience are reflected in the design.

#### Flow

- [ ] User goal is stated.
- [ ] Entry point is clear.
- [ ] Primary success path is complete.
- [ ] Success outcome is explicit.
- [ ] Important alternative success paths are shown.
- [ ] Important failure paths are shown.
- [ ] Recovery behavior is shown where relevant.
- [ ] Cancel/back/exit behavior is clear where relevant.

#### States

- [ ] Loading/submitting behavior is defined where material.
- [ ] Validation errors are defined.
- [ ] System/server errors are covered where material.
- [ ] Empty states are covered where material.
- [ ] Permission/authentication states are covered where material.
- [ ] Destructive actions have intentional confirmation/recovery behavior.

#### Responsive

- [ ] Each adopted structural regime has a composition or an explicit shared-structure mapping.
- [ ] Frame widths/heights and exact structural boundaries are recorded.
- [ ] All adopted Device values map to the intended shell/token behavior.
- [ ] Structural breakpoint changes are visible.
- [ ] Flex/Hug/Fill behavior is used intentionally.
- [ ] Device theme values are explicitly assigned where Device-aware variables are used.

#### Themes

- [ ] Semantic variables are used instead of journey-local theme colors.
- [ ] Representative screens were checked in every adopted Color mode.
- [ ] Focus, error, disabled, overlay, border, and elevation states remain understandable in every adopted mode.

#### Navigation and task surfaces

- [ ] Current product location and active navigation state are clear.
- [ ] Breadcrumbs, tabs, steppers, and Back each represent the correct relationship.
- [ ] Save, Cancel, Close, and Back destinations are explicit where relevant.
- [ ] Temporary forms are intentionally classified as inline, modal, drawer/side panel, or full page.
- [ ] Modal/drawer states are identified as ephemeral or route-backed when that distinction matters.
- [ ] Persistent sidebar navigation is not confused with a temporary drawer/side panel.
- [ ] Destructive actions use an appropriate confirmation/recovery model.

#### Accessibility

- [ ] Critical path can be completed without pointer-only interaction assumptions.
- [ ] Focus order and important focus transitions are specified/reviewed.
- [ ] Validation and error recovery do not rely only on color.
- [ ] Status/success/error changes that require assistive-technology announcement are identified for implementation.
- [ ] The measurable requirements in Section 5.1.5 and criterion-level applicability record in Section 5.1.6 are reflected in the design and implementation acceptance criteria.
- [ ] Responsive variants preserve access to essential actions/content.
- [ ] Motion-sensitive behavior has an appropriate reduced-motion approach where relevant.
- [ ] Every adopted presentation mode preserves the applicable accessibility baseline.

#### System integrity

- [ ] Shared controls use imported design-system instances.
- [ ] No accidental detached/local copies remain.
- [ ] Reusable new solutions were considered for promotion.
- [ ] Instance overrides are intentional.

---

### 10.2. Implementation verification and definition of done

`READY FOR BUILD` approves a specification; `IMPLEMENTED` requires evidence from working code. For the named scope, record:

- [ ] Approved design revision, library revision, token source, code commit/build, and review owners are linked.
- [ ] Required paths, local states, permission boundaries, and transitions work with realistic data and failures.
- [ ] Routing, deep links, direct entry, refresh, Back, cancellation, unsaved work, and resume behavior match the contract.
- [ ] Repeated submissions, delayed responses, unknown outcomes, and concurrent edits have defined handling where relevant.
- [ ] Every adopted shell and presentation mode has been verified; boundary widths, text growth, overflow, and Section 5.1.5 checks pass.
- [ ] Automated accessibility checks are supplemented by manual keyboard, focus, and supported browser/assistive-technology testing. Record versions and results.
- [ ] Applicable WCAG criteria are evaluated for the complete process, including third-party steps; defects and limitations are recorded with owners.
- [ ] Design/code differences are resolved or explicitly accepted within the team's authority; applicable external failures remain failures.
- [ ] The canonical design and documentation reflect implemented behavior. Deployment/release status is recorded separately.

Use a compact coverage record: `journey/path/step → scenario → required contexts → expected outcome → evidence → owner → result`. Mark results Pass, Fail, Not tested, or N/A with reason. Sampling controls canvas size; it does not turn Not tested into Pass. Select representative cases to cover every changed reusable asset, unique layout/interaction, and risky cross-axis combination, and expand coverage when failures or material differences warrant it.

---

### 10.3. Design-system release checklist

Before releasing changes to Foundations, Atoms, Molecules, Organisms, or Templates:

- [ ] The consumer impact is explicit, including additive changes and changes to existing defaults.
- [ ] Every adopted Color value resolves correctly.
- [ ] Relevant Device values resolve correctly.
- [ ] Relevant Atoms, Molecules, Organisms, and Templates were checked at narrow and wide widths where relevant.
- [ ] Focus, disabled, error, selected, and other important states remain coherent.
- [ ] Accessibility contract is still valid.
- [ ] Long text/localization stress cases were considered.
- [ ] Text growth and wrapping were considered.
- [ ] Consumer compatibility is verified, or an intentional breaking change has a coordinated migration under Section 9.5.
- [ ] Code token/reusable-asset impact was considered.
- [ ] The Atomic Design classification is still appropriate and does not duplicate a neighboring layer.
- [ ] Deprecated patterns are clearly marked.

---

### 10.4. Library stress-test fixtures

The library should contain dedicated stress-test frames across representative Atoms, Molecules, Organisms, and Templates.

Recommended coverage:

```text
All adopted Color values
All adopted Device token contexts
Long labels
Maximum content
Empty content
Error state
Focus state
Disabled state
Nested components
Localization stress
Text growth
High-content-density examples
```

For theme testing, set the relevant frame's theme explicitly rather than recoloring instances manually.

---

## 11. Design-to-code contract

### 11.1. Vocabulary and token mapping

Whenever practical, the same semantic concept should have recognizably related names in pen.dev and code.

```text
pen.dev                    code
------------------------------------------------
color.bg.surface     ↔     --color-bg-surface
color.text.primary  ↔     --color-text-primary
layout.pageGutter   ↔     --layout-page-gutter
radius.control      ↔     --radius-control
```

Avoid design calling a value `Grey 3` while engineering understands it as `surfaceElevatedMuted`.

Use the same conceptual Atomic Design vocabulary in design and engineering documentation:

```text
Design taxonomy          Implementation concept
--------------------------------------------------------------
Foundation          ↔    tokens / styles / shared constraints
Atom                ↔    small reusable control
Molecule            ↔    focused composition of controls
Organism            ↔    reusable section / complex interaction
Template            ↔    page-level scaffold / shell
Page                ↔    route/view state in the product app
Journey             ↔    cross-view user flow and behavior
```

The repository does not have to mirror these labels mechanically, but the same asset should not be called an `Organism` in design and a `primitive` in engineering without a deliberate reason.

---

#### 11.1.1. Token and component mapping contract

For each shared token or asset, record the design identifier, code identifier, type, unit, meaning, supported contexts, default, authority, owner, and deprecation/replacement if applicable. Keep aliases acyclic and type-compatible. Distinguish primitive scales from semantic roles and optional component-specific aliases; consumers should depend on semantic roles where available.

```text
Design token: layout.pageGutter
Type / design unit: Number / canvas px
Code token: --layout-page-gutter
Code unit: rem; conversion uses the documented project base
Contexts: Mobile=16, Tablet=24, Desktop=32 design px (example)
Authority / sync direction: recorded in project setup
```

The dotted-name to CSS-name mapping is an explicit convention, not an automatic or lossless conversion. Document units and scale conversion for each token class; not all numbers are pixels. In code, specify how user/system color preference is resolved and how viewport/container changes select responsive values.

pen.dev documents agent-assisted design/code updates and HTML export. Neither workflow guarantees continuous synchronization or preservation of application semantics. Review generated/imported changes against the existing code API and the approved interaction contract. See [pen.dev Design ↔ Code](https://docs.pen.dev/design-and-code/design-to-code).

---

### 11.2. Accessibility handoff

Visual parity is not enough.

Handoff/implementation should preserve:

- semantic element roles;
- labels and accessible names;
- keyboard behavior;
- focus management;
- status/error announcements;
- reduced-motion behavior;
- responsive reading/focus order;
- contrast and non-color cues.

A screenshot-perfect implementation can still fail the accessibility contract.

---

## 12. Worked example: create account

### File

Account creation belongs to the onboarding/authentication journey area defined by the same organization used throughout this document:

```text
apps/web/design/journeys/01-onboarding.pen
```

If account creation and authentication are intentionally grouped because the product is still small, use an explicit grouped filename such as:

```text
apps/web/design/journeys/onboarding-and-authentication.pen
```

Do not introduce a generic `product.pen` file solely for this example.

### Imported library

```text
packages/design-system/design/product-ui.lib.pen
```

### Example setup and scope

This example adopts Light/Dark, all three Device values, and three structurally distinct shells: Compact below 768 CSS px, Medium from 768 to below 1200 CSS px, and Wide from 1200 CSS px. Representative frames are 390, 768, and 1440 px wide; choose and record realistic heights. If Medium adds no distinct structure in the actual product, omit that row and document its shared-shell mapping.

Actor: a visitor creating their own account. Account creation and Sign in (`J02`) are distinct journeys even if they share a file. Invitation entry is in scope only with an explicit invitation-validation branch. Copy uses synthetic data; never place credentials or real verification codes in the design file.

### Journey overview

```text
J01 · CREATE ACCOUNT

Goal:
Create a usable account and enter the product.

Entry points:
Marketing site / invitation / direct sign-up route

Primary success outcome:
User reaches initial product home.
```

### Primary success path

```text
J01.A · PRIMARY SUCCESS

                01 Welcome   02 Details   03 Verify   04 Preferences   05 Home

DESKTOP         [ frame ]    [ frame ]    [ frame ]   [ frame ]        [ frame ]

TABLET          [ frame ]    [ frame ]    [ frame ]   [ frame ]        [ frame ]

MOBILE          [ frame ]    [ frame ]    [ frame ]   [ frame ]        [ frame ]
```

For Mobile frames that use responsive tokens:

```text
Theme:
  Device = Mobile
```

For Tablet and Desktop frames, assign `Device = Tablet` and `Device = Desktop` respectively, and use the mapped Medium and Wide shells.

Main-flow screens can use the product-default color mode:

```text
Theme:
  Color = Light
```

### Failure / recovery paths

```text
J01.B · EXISTING-ACCOUNT HANDLING

02 Details
   ↓
Server detects an existing account (internal condition)
   ↓
Public response follows the reviewed account-disclosure policy
   ↓
Offer Sign in / account recovery without confirming registration status
   OR
Continue through an approved private-channel recovery flow


J01.C · VERIFICATION CODE EXPIRED

03 Verify
   ↓
Expired code
   ↓
Request new code
   ↓
Verification sent
   ↓
Return to Verify


J01.D · VERIFICATION SERVICE ERROR

03 Verify
   ↓
Service unavailable
   ↓
Retry
   OR
Exit and return later
```

Public responses must not automatically reveal that an email address is registered. Review response content, timing, and recovery behavior with security; the server's internal branch can differ from what the interface discloses. See [OWASP Authentication: error messages](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-and-error-messages).

### Additional required scenarios for this example

| Scenario | Representation and contract |
|---|---|
| Invalid verification code | Local state at Step 03; preserve entered data appropriately and explain retry. |
| Resend throttled | Local state; explain when retry is available and announce meaningful status without a noisy countdown. |
| Resend or verification times out | Unknown-result handling; check status before duplicating the operation where supported. |
| Invitation invalid, expired, or mismatched | Branch from invitation entry; define request-new-invitation and ordinary sign-up eligibility. |
| User cancels or returns later | Exit/resume branch; define persistence, authentication state, return route, and expired drafts. |
| Verification succeeds on another device | Cross-device handoff; define how the original device discovers completion and what the next step is. |
| Preferences skipped | Alternative success path if optional; required legal agreements must remain distinct from optional preferences. |

Each branch records its rejoin or terminal outcome. These are example-specific acceptance scenarios, not mandatory account architecture for every product.

### Local states near Step 02

```text
02 · DETAILS / MOBILE

[ Default ]
[ Validation error ]
[ Submitting ]
[ Server error ]
```

These are local because the user remains at the same step.

### Theme QA

```text
40 · THEME + ACCESSIBILITY QA

                      LIGHT             DARK
Details form          [ frame ]         [ frame ]
Validation error      [ frame ]         [ frame ]
Verification          [ frame ]         [ frame ]
Success               [ frame ]         [ frame ]
```

Set the right-hand frames to:

```text
Theme → Color = Dark
```

The reusable elements should update through imported semantic variables.

### Accessibility expectations for this journey

```text
Keyboard:
Entire sign-up flow is completable without pointer-only controls.

Authentication:
Support password managers, paste, and accessible verification entry; apply Section 5.1.5.

Validation:
Submitting invalid data identifies each error in text and defines sensible focus behavior.

Verification:
Expired/sent/success statuses require an implementation announcement strategy.

Responsive:
Mobile retains every action required to complete the account.

Themes:
Error, focus, disabled, and success states remain identifiable in Light and Dark.
```

---

## Appendix A — pen.dev translation cheat sheet

| If this document says… | In pen.dev, that means… |
|---|---|
| Journey file | Normal `.pen` document |
| Design-system library | Converted through Libraries and saved as `.lib.pen`; save source and reopen consumers to adopt updates |
| Foundation | Shared variables, constraints, and principles; not an Atom |
| Atom | Small reusable component such as Button, Checkbox, or Input Control |
| Molecule | Reusable component composed from Atoms, such as Form Field or Search Field |
| Organism | Larger reusable component/composition such as Navigation, Drawer, Filter Panel, or Form |
| Template | Reusable page-level composition/scaffold, often implemented with components + slots |
| Page | Product-specific screen Frame in a journey `.pen` file populated with real content/state |
| Journey | Ordered/branching collection of Pages, overlays, states, and outcomes in a journey `.pen` file |
| Screen | Usually a Frame |
| Responsive row | Group of explicitly sized screen Frames on the canvas |
| Responsive App Shell | Canonical reusable shell composition from the `.lib.pen` library for a meaningful layout regime such as Compact/Medium/Wide |
| Fluid layout | Frame flex layout + Hug/Fill/fixed sizing decisions |
| Variable/token | Variable created in the Variables panel |
| `Color` axis | A theme axis we create with values such as Light and Dark |
| `Device` axis | A separate theme axis we create with Mobile/Tablet/Desktop values |
| Theme axis on canvas | A selectable context on a Frame; one value per axis is active on that Frame at a time |
| Side-by-side Light/Dark QA | Two intentionally duplicated representative Frames/instances with different `Color` selections |
| Set theme on screen | Select Frame → properties → Theme → Add theme → choose axis/value |
| Theme inheritance | Child objects inherit parent frame's axis selection unless overridden |
| Light/Dark component behavior | Component properties reference Color-aware semantic variables |
| Responsive token behavior | Component properties reference Device-aware variables; exact value/shell mapping is in Section 5.3.1 |
| Structural mobile difference | Explicit mobile composition/component structure, not merely Device variables |
| Shared reusable asset | Atom/Molecule/Organism/Template implemented as a component/composition in the library and consumed as instances |
| Flexible reusable region | Empty frame marked as a slot in a component origin; instance content is customizable and suggestions are advisory |
| Persistent sidebar | Part of the app/page layout, typically navigation; represented with reusable shell/navigation components |
| Drawer / side panel | Temporary/contextual task surface; usually a reusable Organism; define modal vs non-modal behavior |
| Modal dialog | Short blocking task/decision with modal focus behavior |
| Route-backed task | A journey state with a stable destination/deep-link behavior, even if visually presented as an overlay |
| Navigation depth | Parent-relationship count from the documented root; route mapping is recorded separately |
| Accessibility | Requirements/annotations/contracts across system and journeys; not a pen.dev theme mechanism by itself |

---

## Appendix B — References

The technical review checked the following primary sources on **2026-09-13**. pen.dev capabilities can change; record the tested app/extension version in project setup and recheck relevant documentation after upgrades. The standard combines product-design conventions with documented capabilities and external guidance. Not every recommendation in this document is dictated by these sources; where the document makes an organizational recommendation, it is a team convention built on top of the referenced capabilities.

### pen.dev capabilities

- [pen.dev Variables and theme dimensions](https://docs.pen.dev/core-concepts/variables)
- [pen.dev Design Libraries](https://docs.pen.dev/core-concepts/design-libraries)
- [pen.dev Components](https://docs.pen.dev/core-concepts/components)
- [pen.dev Slots](https://docs.pen.dev/core-concepts/slots)
- [pen.dev Interface / Flex layout](https://docs.pen.dev/core-concepts/pencil-interface)
- [pen.dev `.pen` format](https://docs.pen.dev/for-developers/the-pen-format)
- [pen.dev Design ↔ Code](https://docs.pen.dev/design-and-code/design-to-code)

Additional capability details are cited where used in Section 2. Canvas/file-format support does not by itself establish runtime behavior.

### Design-system methodology

- [Brad Frost, *Atomic Design — Atomic Design Methodology*](https://atomicdesign.bradfrost.com/chapter-2/)

### Accessibility

- [W3C, Web Content Accessibility Guidelines (WCAG) 2.2](https://www.w3.org/TR/WCAG22/)
- [W3C, ARIA Authoring Practices Guide (APG)](https://www.w3.org/WAI/ARIA/apg/)

APG is practical, non-normative guidance for interaction patterns, including keyboard and focus behavior. WCAG remains the conformance baseline.

### Security-sensitive example

- [OWASP Authentication Cheat Sheet — Authentication and Error Messages](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-and-error-messages)

The individual W3C Understanding pages cited in Section 5.1.5 explain the relevant success criteria and exceptions. They and APG are informative guidance; the WCAG Recommendation is normative.

### Repository / workspace model

- [pnpm Workspaces](https://pnpm.io/workspaces)

---

**Document principle:** If a first-time reader cannot answer **“what must be decided before work starts, which file does this belong in, which pen.dev mechanism represents it, and who owns the rule?”**, this standard should be clarified rather than relying on team memory.
