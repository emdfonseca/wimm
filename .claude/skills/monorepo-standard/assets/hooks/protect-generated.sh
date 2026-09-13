#!/usr/bin/env bash
# PreToolUse (Edit|Write): refuse edits to generated files.
# Exit 2 blocks the tool call; stderr is shown to Claude as the reason.
set -euo pipefail

path=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("tool_input",{}).get("file_path",""))')
[ -n "$path" ] || exit 0

# Keep in sync with .gitattributes linguist-generated entries.
generated=(
  '*/packages/contracts/gen/*'
  '*.pb.go'
  '*.connect.go'
  '*_pb.ts'
  '*_connect.ts'
  '*/sqlc/*'
  '*.gen.go'
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
