## 4. Project setup

**In this file:**

- 4.1. The decision list
- 4.2. Task-surface record
- 4.3. Responsive regimes record
- 4.4. Theme axis record
- 4.5. Where the Project Setup Record lives


### 4.1. The decision list

Before detailed journey frames, the Project Setup Record (`assets/project-setup-record.md`) answers:

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
```

Each item gets a decision, `N/A — reason`, or an open decision (question, assumption, impact if wrong). When a decision changes, update the record.

---

### 4.2. Task-surface record

```text
Add member      → Drawer / modal behavior / ephemeral
Edit member     → Drawer / modal behavior / ephemeral
Permissions     → Full page / route-backed
Delete member   → Destructive confirmation dialog
Rename member   → Inline edit
```

A drawer always states modal or non-modal. The journey uses the library's surface component and records route, open/close, focus, save, cancel, and recovery behavior locally.

---

### 4.3. Responsive regimes record

Ask "where does the product structure change?", not "which devices should we draw?".

```text
Compact → representative frame 390 px
Wide    → representative frame 1440 px
Medium  → no separate canvas row; uses Wide structure at widths ≥ 768 CSS px
          and Compact structure below 768 CSS px
```

or:

```text
Compact → 390 px representative frame
Medium  → 768 px representative frame
Wide    → 1440 px representative frame
```

One screen frame per meaningful composition, width set explicitly, flex/Hug/Fill inside it. A theme axis does not resize a frame.

---

### 4.4. Theme axis record

```text
Color axis
├── Light
└── Dark
```

Create a `Device` axis only when spacing, typography, or control sizes change systematically by regime:

```text
Device axis
├── Compact
├── Medium
└── Wide
```

```text
Device axis ≠ frame width
Device axis ≠ breakpoint engine
Device axis ≠ media query
```

A Compact journey frame is therefore:

```text
Frame width = 390
Theme:
  Color  = Light
  Device = Compact
```

---

### 4.5. Where the Project Setup Record lives

For one journey file, a compact `00 · PROJECT SETUP` frame at the top-left of the canvas, visually separate from the flow. For an initiative spanning several journey files, the full record lives in a repository Markdown file and each journey file carries a short summary; never duplicate the whole record per file.

```text
PROJECT SETUP RECORD   project-level: regimes, axes, navigation model, accessibility baseline, library source, routing constraints
JOURNEY OVERVIEW       journey-level: goal, actor, entry, success outcome, branches, return destination
```

---
