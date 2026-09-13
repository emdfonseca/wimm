<!-- Generated from docs/pen-dev-product-design-organization-standard.md
     by tools/split-standard.py. Edit the standard, not this file. -->
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
Page                ↔    presentational screen component + the route that wires it
Journey             ↔    cross-view user flow and behavior
```

The repository does not have to mirror these labels mechanically, but the same asset should not be called an `Organism` in design and a `primitive` in engineering without a deliberate reason.

#### 11.1.2. A Page maps to two code artifacts

A Page is one design artifact and two implementation ones, and keeping them separate is what makes a designed screen reviewable after it is built:

```text
Presentational screen component   receives data as properties, emits intent as callbacks;
                                  owns everything the Page design specifies
Route / view module               fetches data, holds session and permission context,
                                  performs navigation; owns nothing visual
```

The design artifact describes the presentational half. Its states — loading, validation error, empty, permission denied — are reachable by changing properties, which is what allows each one to be rendered, reviewed, and tested in isolation against the design. When a state can only be produced by standing up routing or a data layer, the split has not been made, and the design's states become unverifiable in practice.

Record which component implements a Page in the journey's implementation notes, the same way tokens and reusable assets are mapped. The route module needs no design counterpart; it implements the journey's routing, entry, and return contracts rather than any single frame.

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
