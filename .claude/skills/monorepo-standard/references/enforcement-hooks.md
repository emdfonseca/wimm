# Enforcement hooks

Conventions in a skill are advice; Claude follows them most of the time. Two of the rules in this standard are cheap to check mechanically and expensive to violate, so they get Claude Code hooks — deterministic scripts that run around tool calls and can block them. Install these into the monorepo's `.claude/settings.json` when the repo is scaffolded; they are inert here because there is nothing to check yet.

| Rule | Hook | Effect |
|---|---|---|
| Never hand-edit generated code | `PreToolUse` on `Edit\|Write` | Blocks the edit; tells Claude to change the source and run `just gen` |
| Every `apps/*` and `packages/*` dir has a `justfile` | `PreToolUse` on `Bash` matching `git commit` | Blocks the commit and names the directories missing one |

The boundary linter (`dependency-cruiser`, `depguard`, `import-linter`) deliberately stays in `just lint` and CI rather than a `PostToolUse` hook: running it after every single edit is slow, and a check that fires mid-refactor on a transiently broken tree teaches people to ignore it.

## Install

Copy `assets/hooks/` to `.claude/hooks/` in the monorepo and merge `assets/hooks/settings.hooks.json` into `.claude/settings.json`. Scripts must be executable. Hooks run inside the same sandbox as the Bash tool, so they need no extra permissions to read the tree.

## How blocking works

A `PreToolUse` hook receives the tool call as JSON on stdin (`tool_name`, `tool_input.file_path` or `tool_input.command`). Exiting `2` blocks the call and feeds stderr back to Claude as the reason; exiting `0` lets it through. Keep the message specific — "generated file, edit `packages/contracts/proto/…` and run `just gen`" gets the right next action; "denied" gets a retry.

## Adjusting the generated-file globs

`protect-generated.sh` lists the patterns. Extend it when a new generator lands (`sqlc`, `openapi-typescript`), and keep the list in one place — the hook, the `.gitattributes` `linguist-generated` entries, and any `eslint`/`golangci` excludes should agree.
