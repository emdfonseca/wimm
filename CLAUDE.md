# wimm

- `just check <dir>` is the only test/lint/typecheck entry point. If it fails, the work is not done.
- devbox owns the toolchain; never install tools globally or in CI by hand. If a needed tool is on the machine but missing from `devbox.json`, add it there and use the devbox one.
- Repo standards are skills in `.claude/skills/` and load by file path; language conventions are `.claude/rules/`.
- Decisions live in `docs/decisions/NNNN-title.md`; add one when a change adds a dependency, service, public contract, or schema.
- Work that is about to be built goes through OpenSpec: `just openspec` runs the pinned CLI, `/opsx:propose` writes the proposal and spec deltas under `openspec/`. A proposal says what is about to be built; an ADR records what was chosen and why. A change needing an ADR needs both.
- Never hand-edit `gen/` output; regenerate with `just gen`.
- This repo is self-contained. Never read or write outside its root, whatever a tool reports as reachable or already open — including other repos on this machine. Pass an explicit in-repo `filePath` to every pencil MCP tool; never rely on the active editor.
