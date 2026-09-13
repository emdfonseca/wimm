#!/usr/bin/env bash
# PreToolUse (Edit|Write): refuse edits to generated files.
#
# Install: copy this file to .claude/hooks/ and merge settings.hooks.json into
# .claude/settings.json. It is invoked via `bash`, so no chmod is needed.
#
# Protocol: the tool call arrives as JSON on stdin (tool_name, tool_input.file_path).
# Exit 2 blocks the call and shows stderr to Claude as the reason; exit 0 allows it.
# Keep the message specific: name the source to edit and the command to run.
#
# The boundary linter (depguard, dependency-cruiser, import-linter) stays in
# `just lint`, not in a hook: it is slow and fires on transiently broken trees.
#
# Extend the pattern list when a new generator lands (sqlc, openapi-typescript).
# Keep it in sync with .gitattributes linguist-generated entries.
set -euo pipefail

path=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("tool_input",{}).get("file_path",""))')
[ -n "$path" ] || exit 0

generated=(
  '*/packages/contracts/gen/*'
  '*.pb.go'
  '*.connect.go'
  '*_pb.ts'
  '*_connect.ts'
  '*/sqlc/*'
  '*.gen.go'
  '*/.claude/rules/decisions.md'
)

for g in "${generated[@]}"; do
  # shellcheck disable=SC2254
  case "$path" in
    $g)
      echo "Refusing to edit generated file: $path" >&2
      echo "Change the source (proto / OpenAPI / SQL queries) and run 'just gen'." >&2
      exit 2
      ;;
  esac
done
exit 0
