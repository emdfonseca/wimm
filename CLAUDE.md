# wimm

- `just check <dir>` is the only test/lint/typecheck entry point. If it fails, the work is not done.
- devbox owns the toolchain; never install tools globally or in CI by hand. If a needed tool is on the machine but missing from `devbox.json`, add it there and use the devbox one.
- Repo standards are skills in `.claude/skills/` and load by file path; language conventions are `.claude/rules/`.
- Decisions live in `docs/decisions/NNNN-title.md`; add one when a change adds a dependency, service, public contract, or schema.
- Work that is about to be built goes through OpenSpec: `just openspec` runs the pinned CLI, `/opsx:propose` writes the proposal, stories, spec deltas and canvas record under `openspec/`. Stories name a real user and the value; the specs' scenarios are the acceptance criteria. Every requirement carries `**Story**: S<n>`. A proposal says what is about to be built; an ADR records what was chosen and why. A change needing an ADR needs both.
- Anything with a user-facing surface is drawn on the canvas before tasks are written, and the components it needs are built before the screen that instances them. See ADR 0008.
- Never hand-edit `gen/` output; regenerate with `just gen`.
- This repo is self-contained. Never read or write outside its root, whatever a tool reports as reachable or already open — including other repos on this machine. Draw on the canvas with `just pen-exec <file.pen>`: the pencil MCP ignores its `filePath` and edits whatever Pen.app has open.
