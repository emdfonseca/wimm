## 8. Design-system library

**In this file:**

- 8.1. Library structure, taxonomy, and responsive shells
- 8.2. Foundations and token vocabulary
- 8.3. Reusable component contracts
- 8.4. Slots and flexible composition


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

