# Finish the engineering-standards setup

## Context

Eight project skills, one reviewer agent, and the enforcement-hook assets now exist in `.claude/`. Two things remain from the "great engineering piece without token cost" discussion: the global command prompts that mis-trigger and load large bodies, and the path-scoped language rules that replace the "Go conventions in CLAUDE.md" idea. Conventions were decided in this session: stdlib `testing`, explicit constructor wiring, `net/http` 1.22 `ServeMux` for public REST, ESLint + Prettier for TS. Nothing here needs new skills.

## 1. Prune global command prompts (user-level, reversible)

Move — do not delete — from `~/.claude/commands/` to `~/.claude/commands-disabled/`:

```text
adaptive.md          581 lines  43 KB   no description; orchestration prompt
agentflow.md         312 lines  31 KB   no description
controlflow.md       225 lines   9 KB   no description
devops.md            253 lines   7 KB   no description
context_pipeline.md  180 lines   6 KB   placeholder description
taskengine.md         42 lines   2 KB   no description (duplicate of taskflow)
taskflow.md           42 lines   2 KB   no description
refactor.md          100 lines   4 KB   generic; bundled `simplify` / `code-review` cover it
pr-draft.md           87 lines   6 KB   thin `gh pr create` wrapper
pr-review.md          45 lines   2 KB   bundled `code-review` covers it
```

`~/.claude/commands/` ends up empty. Honest framing of the saving: the always-on cost is one short listing entry each; the real cost is an accidental trigger loading a 43 KB prompt, which a file with no description invites. Moving keeps rollback to a `mv`.

## 2. Path-scoped rules (project-level)

Create `.claude/rules/`, three files, each ≤ 60 lines, `paths:` frontmatter so they load only when a matching file is read or edited. Content is decisions, not tutorials — anything the model already knows (Effective Go, idiomatic Svelte) stays out.

**`go.md`** — `paths: ["**/*.go"]`
- Logging: `log/slog` through `packages/telemetry`; no `fmt.Println` / `log.Printf` (forbidigo).
- Errors: stdlib; `fmt.Errorf("…: %w", err)`; sentinel `ErrX` for conditions, typed errors when data rides along; mapped to codes only at the handler boundary (`api-contract`).
- Wiring: explicit constructors in `cmd/<app>/main.go`; no DI framework; no package-level mutable state.
- HTTP: Connect handlers on `net/http`; public REST on `net/http` 1.22 `ServeMux` method patterns; no third-party router.
- Layout: `cmd/` + `internal/`; no `pkg/`; `internal/` by default, promote on second consumer.
- `context.Context` first parameter on anything that does I/O; propagate, never `context.Background()` below `main`.
- Tests: stdlib `testing`, table-driven, `_test` external package, `-race` in `just test`.
- Lint: root `golangci.yml` inherited; `depguard` encodes boundary rules; `GOFLAGS=-mod=readonly`.

**`typescript.md`** — `paths: ["**/*.ts", "**/*.svelte"]`
- Svelte 5 runes; no legacy `$:` / `export let`.
- `tsconfig.base.json` inherited; `strict` never loosened without a comment.
- Imports via `@repo/<pkg>` entry points only; no deep `src/` imports.
- Screens: data in as props, intent out as callbacks; route modules stay thin (`storybook-svelte` taxonomy).
- ESLint + Prettier; `no-console` (logs go through the server-side OTel setup).
- Vitest beside the file; component behaviour in play functions, not unit tests.

**`python.md`** — `paths: ["**/*.py"]`
- `uv` only; src layout; distribution `<org>-<dir>`, import `<dir>` with underscores.
- Type hints on every public function; `mypy` in `just check`.
- `ruff` format + lint; no `print` (structured logging via OTel).
- `pytest`, `parametrize`, fixtures in `tests/`; Hypothesis for parsers/arithmetic.

## 3. Consistency touch (project skills)

- `.claude/skills/api-contract/references/rest-public.md`: one sentence naming `net/http` 1.22 `ServeMux` as the REST router, so the rule and the skill agree.

No other skill edits: `tdd/references/go.md` already assumes stdlib; `observability/references/logging.md` already names forbidigo.

## Files

```text
~/.claude/commands-disabled/            (new; all 10 files moved in)
.claude/rules/go.md                     (new)
.claude/rules/typescript.md             (new)
.claude/rules/python.md                 (new)
.claude/skills/api-contract/references/rest-public.md   (one line)
```

## Verification

1. `ls ~/.claude/commands` is empty; `ls ~/.claude/commands-disabled` shows the ten.
2. Frontmatter check on the three rules files: YAML parses, `paths` is a list, each file ≤ 60 lines.
3. `git status` clean after one commit: `feat(rules): add path-scoped Go, TypeScript, Python conventions`.
4. In a fresh session, `/context` after opening any `.go` file should list `rules/go.md` as loaded; after a `.svelte` file, `rules/typescript.md`.
