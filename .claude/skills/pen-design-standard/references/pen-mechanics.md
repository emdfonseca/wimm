## 2. pen.dev capabilities and limits

**In this file:**

- 2.1. `.pen` document
- 2.2. `.lib.pen` design library
- 2.3. Frames
- 2.4. Flex layout, Hug, and Fill
- 2.5. Variables
- 2.6. Theme values and theme axes
- 2.7. Creating and applying theme axes
- 2.8. Theme context and side-by-side comparison
- 2.9. Critical limitation: `Device` does not make a design automatically responsive
- 2.10. Components, instances, and slots
- 2.11. Atomic Design terms are taxonomy, not pen.dev object types
- 2.12. Design representations and executable behavior


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

