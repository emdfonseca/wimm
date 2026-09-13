#!/usr/bin/env python3
"""PreToolUse hook for Bash: block history-destroying git commands.

Exit 2 blocks the call and returns the message to Claude. The user can still
run the command themselves. --force-with-lease is allowed on purpose.
"""
import json
import re
import sys

BLOCKED = [
    (r"\bgit\s+push\b(?!.*--force-with-lease).*(\s--force\b|\s-f\b|\s--force\s)", "force push"),
    (r"\bgit\s+push\b.*\s--delete\b", "remote branch delete"),
    (r"\bgit\s+reset\s+--hard\b", "reset --hard"),
    (r"\bgit\s+clean\b.*\s-[a-zA-Z]*f", "clean -f"),
    (r"\bgit\s+branch\s+(-D|--delete\s+--force)\b", "branch -D"),
    (r"\bgit\s+stash\s+(drop|clear)\b", "stash drop/clear"),
    (r"\bgit\s+(checkout|restore)\s+(--\s+)?\.\s*$", "discard all working tree changes"),
]

try:
    payload = json.load(sys.stdin)
except Exception:
    sys.exit(0)

cmd = (payload.get("tool_input") or {}).get("command", "")
for pattern, label in BLOCKED:
    if re.search(pattern, cmd):
        print(
            f"Blocked ({label}): '{cmd.strip()}'. Irreversible git operation; "
            "show the command to the user and let them run it.",
            file=sys.stderr,
        )
        sys.exit(2)
sys.exit(0)
