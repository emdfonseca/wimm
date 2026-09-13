#!/usr/bin/env python3
"""Stop hook: `just check` every touched package before the turn can end.

Exit 2 blocks the stop and hands stderr back to Claude. Silent no-op when
`just` is absent (pre-devbox), when nothing is touched, or when this hook
already blocked once this turn.
"""
import json
import shutil
import subprocess
import sys

try:
    payload = json.load(sys.stdin)
except Exception:
    sys.exit(0)

if payload.get("stop_hook_active"):
    sys.exit(0)
if shutil.which("just") is None:
    sys.exit(0)

changed = subprocess.run(
    ["git", "status", "--porcelain", "--untracked-files=all"],
    capture_output=True, text=True,
).stdout

pkgs = set()
for line in changed.splitlines():
    path = line[3:].split(" -> ")[-1].strip().strip('"')
    parts = path.split("/")
    if len(parts) > 2 and parts[0] in ("apps", "packages"):
        pkgs.add(f"{parts[0]}/{parts[1]}")

failures = []
for pkg in sorted(pkgs):
    r = subprocess.run(["just", "check", pkg], capture_output=True, text=True)
    if r.returncode != 0:
        tail = (r.stdout + r.stderr).strip().splitlines()[-15:]
        failures.append(f"{pkg}: just check failed\n" + "\n".join(tail))

if failures:
    print(
        "`just check` failed. Fix it before ending the turn:\n\n"
        + "\n\n".join(failures),
        file=sys.stderr,
    )
    sys.exit(2)
sys.exit(0)
