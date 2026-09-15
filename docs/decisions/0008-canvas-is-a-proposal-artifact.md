# 0008 · The canvas is a proposal artifact

## Status

Accepted

## Context

ADR 0007 adopted OpenSpec for change proposals. Its default `spec-driven`
workflow is proposal → specs → design → tasks, where "design" means a technical
design document. Nothing in it reaches the pen.dev canvas.

In this project the canvas is where a screen is decided. Writing tasks for a
screen that has not been drawn inverts the order: the implementation invents the
layout, and the canvas becomes a record of what was built rather than a decision
about what to build.

## Decision

A project schema, `openspec/schemas/wimm/`, adds a fifth artifact between design
and tasks. `openspec/config.yaml` sets `schema: wimm`, so every new change is
stamped with it.

```text
proposal → specs → design → canvas → tasks
                             │
                             └── tasks is blocked until canvas exists
```

**`canvas` draws the journey and records what was drawn.** The drawing goes in a
journey `.pen` under `packages/ui/design/`; `product-ui.lib.pen` keeps holding
reusable mechanics only. `canvas.md` records the journey ID, the frame IDs, the
task surfaces, the library components instanced, the components the library is
missing, the states drawn, and the behavioural contracts the canvas cannot
execute.

**It is conditional.** A change with no user-facing surface — a worker, a
migration, a build script — records a deliberate skip. The condition is stated
in the artifact's own instruction, which is the mechanism OpenSpec already uses
for a conditional artifact.

**Components come before the screen that uses them.** Each entry under Components
missing becomes one task covering both the origin in `product-ui.lib.pen` and the
Svelte component in `packages/ui/src`. The screen composes instances and cannot
share a task with the components it instances.

## Consequences

- **A pencil edit is not on disk until a person saves it.** The MCP mutates the
  document open in the app; the file keeps its mtime and git reports nothing.
  `canvas.md` carries a Saved to disk field for exactly this reason, and the
  archive guidance refuses to archive a change whose frames were never
  committed. This is the sharpest edge in the workflow and it is silent.
- **The `.pen` does not diff.** It is a 1.5 MB encrypted blob, so no reviewer can
  see a design change by reading the commit. `canvas.md` is the review artifact;
  that is why frame IDs and contracts are recorded in full rather than
  summarised.
- The schema is a copy of the built-in `spec-driven` one with an artifact added,
  so it does not track upstream improvements to the other four instructions.
  Bumping the CLI pin in `bin/openspec` means diffing the packaged schema
  against this one.
- A change that is genuinely backend-only now carries one extra decision — skip
  or draw. That is cheaper than the alternative, which is a UI change reaching
  tasks with no design.
- `openspec templates` reports the default schema unless given `--schema wimm`;
  it does not read `config.yaml`. Every other command does. This affects only
  that diagnostic.
