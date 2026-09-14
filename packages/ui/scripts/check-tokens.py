#!/usr/bin/env python3
"""Verify the generated stylesheet still matches the token export.

CI cannot read the .pen file, so drift between the library and tokens.json is
caught by review, not here. What this catches is drift between tokens.json and
tokens.css — the part that is mechanical and therefore worth automating.
"""
import json, pathlib, re, subprocess, sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
failures = []

doc = json.loads((ROOT / "design/tokens.json").read_text())
tokens = doc["tokens"]
css_path = ROOT / "src/tokens.css"

if not css_path.exists():
    print("src/tokens.css is missing — run `just gen packages/ui`", file=sys.stderr)
    sys.exit(1)
css = css_path.read_text()

# 1. regenerating must be a no-op
gen = subprocess.run([sys.executable, str(ROOT / "scripts/gen-tokens.py"), "--check"],
                     capture_output=True, text=True)
if gen.returncode != 0:
    failures.append(gen.stderr.strip() or "tokens.css is stale")

# 2. every exported token reaches the stylesheet
missing = [k for k in tokens if f"--{k}:" not in css]
if missing:
    failures.append("missing from tokens.css: " + ", ".join(sorted(missing)))

# 3. no custom property in the stylesheet that the export does not define
declared = {m.group(1) for m in re.finditer(r"^\s*--([a-z0-9-]+):", css, re.M)}
extra = declared - set(tokens)
if extra:
    failures.append("in tokens.css but not exported: " + ", ".join(sorted(extra)))

# 4. every themed token supplies all of its axis values
for name, spec in tokens.items():
    for axis, expected in doc["themes"].items():
        if axis in spec and set(spec[axis]) != set(expected):
            failures.append(f"{name}: {axis} has {sorted(spec[axis])}, expected {sorted(expected)}")

if failures:
    for f in failures:
        print("FAIL " + f, file=sys.stderr)
    sys.exit(1)

print(f"tokens.css matches tokens.json: {len(tokens)} tokens, {len(declared)} declared")
