#!/usr/bin/env bash
# PreToolUse (Bash): before a git commit, every apps/* and packages/* directory
# must carry a justfile, otherwise it is invisible to `just check` and CI.
set -euo pipefail

cmd=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("tool_input",{}).get("command",""))')
case "$cmd" in
  *"git commit"*) ;;
  *) exit 0 ;;
esac

root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
missing=()
for d in "$root"/apps/*/ "$root"/packages/*/; do
  [ -d "$d" ] || continue
  [ -f "$d/justfile" ] || missing+=("${d#"$root"/}")
done

if [ "${#missing[@]}" -gt 0 ]; then
  echo "Commit blocked: these directories have no justfile, so 'just check' and CI skip them:" >&2
  printf '  %s\n' "${missing[@]}" >&2
  echo "Add a justfile implementing build/test/lint/fmt/check/dev/clean (see monorepo-standard)." >&2
  exit 2
fi
exit 0
