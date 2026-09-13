---
name: pen-design-standard
description: Organizing standard for pen.dev product design - where journeys, paths, states, responsive compositions, Light/Dark themes, accessibility contracts, and reusable Atomic Design assets belong in .pen and .lib.pen files, plus the project setup, readiness gates, and review checklists that surround them. Use this whenever work touches a .pen or .lib.pen file or the pencil MCP tools; whenever someone asks to design, lay out, or restructure a screen, flow, journey, user goal, error or empty state, modal, drawer, app shell, navigation, design system, component library, design token, or theme; whenever a design is being prepared for handoff, review, or implementation; and whenever a designer or engineer is starting design work and needs to know what must be decided first. Reach for it even when the request sounds like a single screen ("mock up the checkout page", "add a dark mode version", "make this responsive") - the point of the standard is that those requests usually should not become new frames.
---

# pen.dev product design organization

**Core rule: journeys own product flow, the design system owns reusable rules, accessibility applies to both.**

Canvas disorder almost always comes from one mistake — representing every kind of information by duplicating screens. The standard exists to separate them, so keep this shape in mind before adding anything to a canvas:

```text
Canvas        = journey × path × meaningful responsive composition
Themes        = variable-driven context on a frame
Local states  = examples beside the relevant step
System rules  = shared .lib.pen library (foundations → atoms → molecules → organisms → templates)
Accessibility = requirements applied across all of the above
```

Never build `journey × path × breakpoint × theme × state × component variant` as frames. That matrix is the failure mode this whole standard prevents.

## Start here, every time

Three questions decide almost everything. Answer them before opening the editor:

1. **Is this reusable mechanics, or product flow?** Reusable → `.lib.pen` design system. Actual routes, content, decisions, and states → journey `.pen` file. The library defines *how a mechanism behaves*; the journey defines *when and why it is used*.
2. **Does the Project Setup Record exist and cover this work?** If not, that gap is the real task — see "Driving the human" below. Designing on top of invisible assumptions is what makes canvases unreviewable later.
3. **Is this a new frame at all?** Most requests are a theme value, an instance override, slot content, or a local state beside an existing step. Add a frame only for a genuinely different composition.

## Where to read next

SKILL.md carries the rules you need constantly. Everything else lives in `references/` — read the file when its situation comes up, not preemptively.

Those reference files are generated from `docs/pen-dev-product-design-organization-standard.md`, which is the canonical standard. Edit the standard and run `python3 tools/split-standard.py`; edits made directly to `references/` are overwritten.

| Read this | When |
|---|---|
| `references/project-setup.md` | Starting a project or feature; filling or auditing the Project Setup Record; checking Definition of Ready. Long — it has the full category list and the copy-paste record. |
| `references/pen-mechanics.md` | Using variables, theme axes, frames, flex/Hug/Fill, components, slots, or converting a file to a library. Read before assuming a pen.dev capability exists. |
| `references/journeys-and-canvas.md` | Laying out a journey file, naming frames, deciding local state vs branch, classifying paths. |
| `references/navigation-and-surfaces.md` | Choosing inline / modal / drawer / full page; navigation patterns; CRUD flows; route-backed vs ephemeral; overlay accessibility contracts. |
| `references/design-system-library.md` | Library structure, Atomic Design classification, responsive App Shells, foundations and token vocabulary, component contracts, slots. |
| `references/accessibility-and-responsive.md` | Any accessibility question; the measurable WCAG 2.2 AA baseline with exact numbers; theme policy; responsive regimes and boundaries. |
| `references/ownership-and-repo.md` | Where files live in the repo; who owns a decision; shared screens across journeys. |
| `references/verification-gates.md` | Before claiming READY FOR BUILD or IMPLEMENTED. Full checklists. |
| `references/lifecycle.md` | Status labels, promoting local work into the library, archiving, library release and migration. |
| `references/design-to-code.md` | Token and component naming across design and code; accessibility handoff. |
| `references/worked-example.md` | A complete journey (create account) laid out end to end — useful as a model to imitate. |
| `references/conventions.md` | Requirement language (must/should/may), applicability, authority for exceptions. |
| `references/appendices.md` | Cheat sheet translating standard vocabulary into pen.dev mechanisms; source list. |

## Rules that are constantly needed

### Themes are context, not copies

A frame resolves **one value from each theme axis at a time**. Setting `Color = Dark` re-renders that same frame; children inherit unless they override.

- `Color` (Light/Dark) and `Device` (Mobile/Tablet/Desktop) are **independent axes**. Never create combined values like `Mobile Dark`.
- **`Device` is not a breakpoint.** Resizing a frame from 1440 to 390 changes nothing about which `Device` value resolves. It only selects variable values (gutters, type sizes). Structure comes from explicit compositions.
- Side-by-side Light/Dark belongs only in the QA zone, on representative screens — never as a rule applied to every journey screen.
- Components reference semantic variables (`color.bg.surface`), never raw hex, and never duplicate a component per theme.

### Responsive means two mechanisms

Structure (explicit compositions + flex/Hug/Fill + annotated wrapping and overflow) and tokens (optional `Device` axis). Shell regimes are **Compact / Medium / Wide**; `Device` values stay **Mobile / Tablet / Desktop**, mapping Compact→Mobile, Medium→Tablet, Wide→Desktop. If a regime adds no structural difference, do not draw it — record the shared-shell mapping in a note instead.

### Local state or branch?

Beside the step when the user stays on the same conceptual step, one screen is affected, and recovery is immediate — validation errors, loading, retryable failures. A branch when navigation changes, several steps appear, business rules alter the path, or a different end state becomes reachable. Being visually inside a drawer does not make something too small to deserve a branch.

### Naming

Fully qualified: `J05.A / 03 · Payment / Mobile / Validation error`. Short names are fine inside an unambiguous group, but use the full identifier in links, annotations, and exports. `.A` is always primary success. Journey IDs are stable and never reused. Files are lowercase kebab-case, one per journey family or product area — never one per route or page. Do not append `Light`/`Dark` to frame names; the frame's theme value already says it.

### Journey file zoning

```text
00 · JOURNEY OVERVIEW        goal / actor / entry / outcome / constraints
10 · PRIMARY SUCCESS         main flow, responsive compositions
20 · ALTERNATIVE / FAILURE / RECOVERY PATHS
30 · STATE / EDGE-CASE INDEX index of states; never a second editable copy
40 · THEME + ACCESSIBILITY QA labelled comparisons, each citing its source frame
90 · TEMPORARY NOTES
```

Left → right is progress through a path. Top → bottom is responsive regime. Steps never go wherever there is empty canvas.

### Task surfaces

Inline for one small value. Modal for a short blocking decision or compact task. Drawer when a moderate task benefits from visible page context — and always state modal or non-modal. Full page when the task is complex, deep-linkable, recoverable, or likely to grow. Destructive actions get a confirmation dialog naming what is affected, not a drawer just because editing used one. For every overlay, say whether it is **route-backed** or **ephemeral**; that decision drives Back, refresh, analytics, and focus behavior.

### A Page is one design artifact and two code artifacts

Pages stay out of the shared library — they are product screens, so they live in journey files. That is a statement about *design* ownership, not a claim that a Page has no implementation counterpart. In code it becomes two:

```text
presentational screen component   data in as properties, intent out as callbacks
route / view module               fetching, session, permissions, navigation
```

The design describes the first. Keeping it free of data wiring is what lets each designed state — loading, validation error, empty, permission denied — be rendered and reviewed by setting properties. When a state can only be produced by standing up routing or a data layer, the states in the design have quietly become unverifiable. `references/design-to-code.md` §11.1.2 has the detail; if the implementation uses Storybook, the `storybook-svelte` skill covers how those states are storied.

### Accessibility is not a phase or an axis

WCAG 2.2 AA is the web baseline and applies to the default experience — never model it as `Accessibility = On/Off`. Component-level accessibility does not make a journey accessible: check keyboard completion, focus order and transitions, error identification and recovery, status announcement, zoom and reflow, and whether responsive variants still expose every essential action. When you need exact thresholds (contrast ratios, target sizes, reflow widths, text spacing), read `references/accessibility-and-responsive.md` rather than recalling numbers — the criteria have exceptions that matter and are easy to state wrongly.

## Working with .pen files

`.pen` files are encrypted. Use the pencil MCP tools only — never Read or Grep them. Call `get_editor_state(include_schema: true)` before any other pencil tool if the schema is not already in context.

If the pencil MCP's own `get_guidelines` conflicts with this standard, pen.dev's documentation wins on **mechanics** (what the tool can actually do) and this standard wins on **organization** (how we choose to arrange work). Say out loud when the two disagree rather than silently picking one.

The canvas represents intent; it does not execute it. Arrows, state frames, and annotations establish no routing, focus management, or announcements. Anything behavioral has to be written as a contract for implementation and verified in code.

## Driving the human

The standard is a team process, and processes fail quietly. When a decision is missing, the useful move is to name it and keep working — not to stop, and not to invent an answer and bury it in thirty frames.

**When setup is missing or thin:** say which categories are unanswered and why they change the work, offer `assets/project-setup-record.md` as the starting template, and proceed on stated assumptions for anything that does not depend on the answer. Checking the full category list in `references/project-setup.md` takes a minute and saves rework.

**When a decision is genuinely open,** record it in this form rather than encoding a guess:

```text
OPEN DECISION
Question:
Current assumption:
Owner:
Decision needed by:
What changes if the assumption is wrong:
```

**Checkpoint at the gates,** because they are where silent drift becomes expensive:

- *Before detailed design* — Definition of Ready (`references/project-setup.md`). Small unknowns can stay open; invisible assumptions cannot.
- *Before READY FOR BUILD* — run `references/verification-gates.md` §10.1 and report honestly. Mark items `N/A` with a reason rather than skipping them, and never mark "not tested" as passing.
- *Before IMPLEMENTED* — §10.2 wants evidence from working code, not an approved specification.
- *Before changing anything shared* — §10.3, plus the consumer review in `references/lifecycle.md`. A library edit reaches every journey that imports it.

**When you see the anti-patterns, name them.** Duplicated shared controls that should be library instances, a screen copied per theme, a local component recreating something the library already has, breadcrumbs and tabs expressing the same dimension, a temporary edit panel called a "sidebar", a journey file per route. These are cheap to fix early and expensive later, so flag them as you notice them rather than at review time.

**Docs are state, not history.** Update the Project Setup Record and journey overview to say what is true now; the commit history carries how it got that way.
