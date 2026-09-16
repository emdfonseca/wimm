## 8. Design-system library

**In this file:**

- 8.1. Library structure and taxonomy
- 8.2. Foundations and token vocabulary


### 8.1. Library structure and taxonomy

```text
packages/ui/design/product-ui.lib.pen
```

The library is where reusable mechanics belong (app shell, breadcrumbs, tabs, dialogs, drawers, confirmation patterns), never the product's route hierarchy or journey-specific content.

```text
00 · README

10 · FOUNDATIONS     variables: color, type, space, layout, radius, control, focus, motion

20 · ATOMS

30 · MOLECULES

40 · ORGANISMS

50 · TEMPLATES       complete screens only: App Shell / Compact · App Shell / Medium
                     (only if structurally distinct) · App Shell / Wide, and the
                     auth and landing shells. Nothing that is a region rather
                     than a screen - a content scaffold is an organism.

90 · QA / STRESS TESTS
```

Components are added when a journey needs them. Classify by composition, responsibility, and reuse scope, then stay consistent; do not move an element between levels because its implementation grew.

**The test for 50 is "is this a whole screen".** Not "is it page-level", not "does it have slots" - a scaffold drawn at 1176 is the Main column of a 1440 shell, so it is a region and belongs in 40. Everything in 50 is a frame someone could screenshot and call a screen.

**A frame in 50 carries the zone's chrome**, so the whole set reads as one board:
`fill $color-bg-canvas`, `cornerRadius $radius-lg`, `stroke $color-border-default`,
`strokeWidth 1`, `strokeAlignment inner`, `clip true`, and a `theme` binding whose
`device` matches its width - compact 390, medium 1024, wide 1440, ultra 1920. The
theme binding is not decoration: without it the frame resolves every device-axis
token at the wrong regime, and a sidebar comes out the wrong width inside a
correct-looking frame.

| Layer | Definition | Typical examples | Lives where? |
|---|---|---|---|
| **Foundations** | Tokens, constraints, and principles used by every layer | color, type, spacing, focus, motion | `.lib.pen` |
| **Atom** | Smallest reusable UI unit with a meaningful contract | Button, Checkbox, Input Control, Badge | `.lib.pen` |
| **Molecule** | Small composition of Atoms with one focused purpose | Form Field, Search Field, Breadcrumbs | `.lib.pen` |
| **Organism** | Larger reusable section or interaction, including a scaffold that fills one region of a screen rather than being one | Navigation, Table region, Drawer, Form, List/Detail scaffold | `.lib.pen` |
| **Template** | A **complete screen** at real size: the whole frame a person would see, with regions/slots and no business content | App Shell, Auth shell, Signed-in landing | `.lib.pen` |
| **Page** | A Template/Organisms populated with real product content, state, and navigation context | Checkout review, Project detail | journey `.pen` |
| **Journey** | A goal over time sequencing Pages, overlays, states, failures, and recovery | Create account, Invite teammate | journey `.pen` |

An App Shell is a structural composition; `Device = Compact` is only a variable context and does not create or resize the shell. One shell that adapts through fluid layout is one Template, not three.

Theme comes from variables, never component copies:

```text
❌ Button / Primary / Light
❌ Button / Primary / Dark
```

---

### 8.2. Foundations and token vocabulary

Foundations sit below and across the Atomic Design hierarchy; they are not Atoms. Every layer consumes foundation variables rather than redefining them.

#### Color

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

color.action.primary
color.action.onPrimary
color.action.destructive
color.action.onDestructive

color.feedback.success
color.feedback.warning
color.feedback.error
color.feedback.info
```

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

```text
focus.ring.color
focus.ring.width
focus.ring.offset

control.target.min
motion.duration.*
motion.reduced.*
```

The names above are direction, not a ceiling or an inventory: a project adds what it needs and omits what it doesn't. Values live in the project's `tokens.css` (`storybook/references/tokens.md`); do not add tokens solely for documentation.

---
