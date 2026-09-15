---
name: pen-design
description: "pen.dev design organization: journey .pen vs .lib.pen ownership, theme axes vs frames, local state vs branch, frame naming, task surfaces. Use whenever work touches a .pen or .lib.pen file or the pencil MCP tools, or the ask is to design, mock up, lay out, or restructure a screen, flow, state, modal, drawer, app shell, design system, token, or theme - even 'add dark mode' or 'make it responsive'."
---

# pen.dev product design organization

**Core rule: journeys own product flow, the design system owns reusable rules, accessibility applies to both.**

```text
Canvas        = journey × path × meaningful responsive composition
Themes        = variable-driven context on a frame
Local states  = examples beside the relevant step
System rules  = shared .lib.pen library (foundations → atoms → molecules → organisms → templates)
Accessibility = requirements applied across all of the above
```

Never build `journey × path × regime × theme × state × component variant` as frames.

## Start here, every time

1. **Reusable mechanics, or product flow?** Reusable → `.lib.pen`. Routes, content, decisions, and states → journey `.pen`.
2. **Does the Project Setup Record cover this work?** If not, fill the gaps in `assets/project-setup-record.md` and proceed on stated assumptions.
3. **Is this a new frame at all?** Most requests are a theme value, an instance override, slot content, or a local state beside an existing step.

## Where to read next

| Read this | When |
|---|---|
| `references/project-setup.md` | Starting a project or feature; the decision list and example records. |
| `references/pen-mechanics.md` | Variables, theme axes, frames, flex/Hug/Fill, components, slots, library conversion; vocabulary → pen.dev cheat sheet. |
| `references/journeys-and-canvas.md` | Journey file zoning, canvas grammar, local state vs branch, naming. |
| `references/navigation-and-surfaces.md` | Navigation pattern table; inline / modal / drawer / full page; route-backed vs ephemeral; overlay accessibility contracts. |
| `references/design-system-library.md` | Library zoning, layer table, token names. |
| `references/accessibility-and-responsive.md` | WCAG 2.2 AA measurable baseline; regime widths. |
| `references/tooling-traps.md` | **Before the first `execute` call.** Silent property drops, stale reads, visitor and instance traps. |
| `references/library-method.md` | Building or extending the design-system library: component tables, specimen cells, origins, presets, theme QA. |
| `references/verification-gates.md` | Readiness checklist before READY FOR BUILD. |
| `assets/project-setup-record.md` | Copy-paste Project Setup Record. |

## Rules that are constantly needed

### Themes are context, not copies

A frame resolves **one value from each theme axis at a time**. Setting `Color = Dark` re-renders that same frame; children inherit unless they override.

- `Color` (Light/Dark) and `Device` (Compact/Medium/Wide) are **independent axes**. Never create combined values like `Compact Dark`.
- **`Device` is not a breakpoint.** Resizing a frame from 1440 to 390 changes nothing about which `Device` value resolves. It only selects variable values (gutters, type sizes). Structure comes from explicit compositions.
- Side-by-side Light/Dark belongs only in the QA zone, on representative screens — never as a rule applied to every journey screen.
- Components reference semantic variables (`color.bg.surface`), never raw hex, and never duplicate a component per theme.

### Responsive means two mechanisms

Structure (explicit compositions + flex/Hug/Fill + annotated wrapping and overflow) and tokens (optional `Device` axis). Regimes are **Compact / Medium / Wide** for frames, shells, and `Device` values alike. If a regime adds no structural difference, do not draw it — record the shared-shell mapping in a note instead.

### Local state or branch?

Beside the step when the user stays on the same conceptual step, one screen is affected, and recovery is immediate — validation errors, loading, retryable failures. A branch when navigation changes, several steps appear, business rules alter the path, or a different end state becomes reachable. Being visually inside a drawer does not make something too small to deserve a branch.

### Naming

Fully qualified: `J05.A / 03 · Payment / Compact / Validation error`. Short names are fine inside an unambiguous group, but use the full identifier in links, annotations, and exports. `.A` is always primary success. Journey IDs are stable and never reused. Files are lowercase kebab-case, one per journey family or product area — never one per route or page. Do not append `Light`/`Dark` to frame names; the frame's theme value already says it.

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

Pages live in journey files, not the library. In code a Page is a presentational screen component (data in as props, intent out as callbacks) plus a thin route module; the design describes only the first, so every designed state is reachable by setting props (`storybook` covers how those states are storied).

### Accessibility is not a phase or an axis

WCAG 2.2 AA is the web baseline and applies to the default experience — never model it as `Accessibility = On/Off`. Component-level accessibility does not make a journey accessible: check keyboard completion, focus order and transitions, error identification and recovery, status announcement, zoom and reflow, and whether responsive variants still expose every essential action. When you need exact thresholds (contrast ratios, target sizes, reflow widths, text spacing), read `references/accessibility-and-responsive.md` rather than recalling numbers — the criteria have exceptions that matter and are easy to state wrongly.

### Verify in a separate call, and before reporting

Anything read in the call that made the change is stale — bounds, `ctx.problems`
and screenshots all return the pre-layout frame. Mutate, then verify in a second
call, then fix. Run the audit before showing the work rather than after a reviewer
finds something; a defect a person had to catch is one the check should have.

A check that has never fired is not evidence. Build the defect it is meant to
catch, confirm it fires, delete the fixture.

### Variants are rows, states are columns

One table per component, with the `rest` column serving as the variant showcase.
A separate variants row beside a state matrix is the same specimens twice and two
places to update. Draw the crossing rather than writing prose verdicts about it —
prose gravitates to the obvious and skips what is actually undecided.

### A contract that states a number needs a measurement

*Clears 3:1*, *never reflows*, *one level of depth* — prose like this is read as
verified, and a false claim in a contract is worse than no claim, because it
tells the next person the check was already done. Measure every number and every
absolute when you write it, and again when the thing it describes changes.

### Anything that redraws an atom's chrome is that atom

A component reproducing a control's border, height, focus ring and states is a
configuration of it, not a new component. Build a preset that wraps an instance.

## Working with .pen files

A `.pen` file is pretty-printed JSON. Use the pencil MCP tools to read and edit
the design — a whole-document Read is thousands of lines and tells you less than
one `Get` visitor — but the top-level keys (`version`, `imports`, `themes`,
`variables`) are ordinary JSON and are edited as such. Call
`get_editor_state(include_schema: true)` before any other pencil tool if the
schema is not already in context.

**A journey imports the library; it never copies it.** `imports` maps an alias to
a relative path, and everything in the library is then addressed through it:

```text
imports     { "ui": "product-ui.lib.pen" }      ← set by `just pen-import`
components  ref: "ui:W2gOKx"
variables   "$ui:color-bg-canvas"
```

The colon is load-bearing. A bare `W2gOKx` is a non-existent node and `ui/W2gOKx`
is rejected outright — `ref` may not contain a slash. Imported components resolve
their own tokens against the library's variables, but a node **you** draw in the
journey file must alias-qualify every token or it silently falls back to black.

No MCP or CLI call can write `imports`: `execute` has no document-level mutator
beyond `SetVariables`, and `Update(document, …)` reports `Node 'document' not
found`. `just pen-import <journey.pen> <alias> <library.pen>` writes it, and
creates the journey file if it does not exist.

**Draw headless; the MCP cannot target a file.** `just pen-exec <file.pen>` pipes
an execute snippet into `pen interactive --in X --out X`, which really opens X,
resolves its imports, saves, and fails without writing if the snippet errored.
That is how journeys are drawn.

`mcp__pencil__execute` ignores its `filePath` and acts on whatever Pen.app has
open — silently, so an edit aimed at a journey lands in the library. Use it only
for the document a person is actually looking at. When you do, the edit stays in
the app's memory until `just pen-save <file.pen>` flushes it; the file keeps its
mtime and git reports nothing until then.

Never headless-write a file the app is holding: the app's copy is stale and its
next save wins.

If the pencil MCP's own `get_guidelines` conflicts with this standard, pen.dev's documentation wins on **mechanics** (what the tool can actually do) and this standard wins on **organization** (how we choose to arrange work). Say out loud when the two disagree rather than silently picking one.

The canvas represents intent; it does not execute it. Arrows, state frames, and annotations establish no routing, focus management, or announcements. Anything behavioral has to be written as a contract for implementation and verified in code.

## Open decisions and gates

When a decision is missing, name it and keep working on a stated assumption; before claiming READY FOR BUILD, run `references/verification-gates.md`.

```text
OPEN DECISION
Question:
Current assumption:
What changes if the assumption is wrong:
```
