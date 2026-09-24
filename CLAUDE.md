# wimm

- `just check <dir>` is the only test/lint/typecheck entry point. If it fails, the work is not done.
- devbox owns the toolchain; never install tools globally or in CI by hand. If a needed tool is on the machine but missing from `devbox.json`, add it there and use the devbox one.
- Repo standards are skills in `.claude/skills/` and load by file path; language conventions are `.claude/rules/`.
- Decisions live in `docs/decisions/NNNN-title.md`; add one when a change adds a dependency, service, public contract, or schema.
- Work that is about to be built goes through OpenSpec: `just openspec` runs the pinned CLI, `/opsx:propose` writes the proposal, stories, spec deltas and state plan under `openspec/`. Stories name a real user and the value; the specs' scenarios are the acceptance criteria. Every requirement carries `**Story**: S<n>`. A proposal says what is about to be built; an ADR records what was chosen and why. A change needing an ADR needs both.
- Once a change's scope is confirmed, carry it through every artifact to tasks without stopping to ask whether to continue. Ask only when an answer would change what gets built.
- Code is the design of record (ADR 0023). Anything with a user-facing surface gets a state plan in `canvas.md` before tasks are written: the state stories, the words each shows, and what was seen. Components are built before the screen that uses them, the presentational screen and its state stories before a look at all four regimes on the design canvas (`just canvas`), and that look before any wiring.
- Never hand-edit `gen/` output; regenerate with `just gen`.
- An agent never runs `just approve` and never writes to `apps/storybook/canvas/approvals.jsonl`; approvals are a person's, made in their own terminal and committed on their own.
- This repo is self-contained. Never read or write outside its root, whatever a tool reports as reachable or already open — including other repos on this machine.
