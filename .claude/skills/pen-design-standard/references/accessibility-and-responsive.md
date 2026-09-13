<!-- Generated from docs/pen-dev-product-design-organization-standard.md
     by tools/split-standard.py. Edit the standard, not this file. -->
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
