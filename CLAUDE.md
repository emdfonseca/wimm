# wimm

Polyglot monorepo template: Go services and workers under `apps/`, shared code under `packages/`, SvelteKit UI in `packages/ui`, Python only where a library forces it.

- `just check <dir>` is the only test/lint/typecheck entry point. If it fails, the work is not done.
- devbox owns the toolchain; never install tools globally or in CI by hand. If a needed tool is on the machine but missing from `devbox.json`, add it there and use the devbox one.
- Repo standards are skills in `.claude/skills/` and load by file path; language conventions are `.claude/rules/`.
- Decisions live in `docs/decisions/NNNN-title.md`; add one when a change adds a dependency, service, public contract, or schema.
- Never hand-edit `gen/` output; regenerate with `just gen`.
