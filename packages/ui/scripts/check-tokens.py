#!/usr/bin/env python3
"""Validate design/tokens.json and its agreement with src/tokens.css.

Two different jobs, and only the second was ever done here before:

  1. Is the export itself a valid token document? Names, types, units, value
     shapes, known axes, complete axis coverage, no unknown fields.
  2. Does the generated stylesheet still match it?

Checking only (2) means a token typed "banana", a colour of #GGGGGG, or an axis
nobody defined all pass, because the generator faithfully renders nonsense and
the comparison then agrees with itself.

What this still cannot check: whether tokens.json matches the pen.dev library.
CI cannot read a .pen file. That link is re-export plus review — see
docs/design/tokens.md.
"""
import json, pathlib, re, subprocess, sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from tokens_schema import validate  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
fails = []

doc = json.loads((ROOT / "design/tokens.json").read_text())
themes = doc.get("themes", {})
tokens = doc["tokens"]
fails += validate(doc)

css_path = ROOT / "src/tokens.css"
if not css_path.exists():
    fails.append("src/tokens.css is missing — run `just gen packages/ui`")
else:
    css = css_path.read_text()
    gen = subprocess.run([sys.executable, str(ROOT / "scripts/gen-tokens.py"), "--check"],
                         capture_output=True, text=True)
    if gen.returncode != 0:
        fails.append(gen.stderr.strip() or "tokens.css is stale")
    missing = [k for k in tokens if f"--{k}:" not in css]
    if missing:
        fails.append("missing from tokens.css: " + ", ".join(sorted(missing)))
    declared = {m.group(1) for m in re.finditer(r"^\s*--([a-z0-9-]+):", css, re.M)}
    extra = declared - set(tokens)
    if extra:
        fails.append("in tokens.css but not exported: " + ", ".join(sorted(extra)))

if fails:
    for f in fails:
        print("FAIL " + f, file=sys.stderr)
    sys.exit(1)
print(f"tokens.json valid and tokens.css matches: {len(tokens)} tokens, "
      f"axes {', '.join(sorted(themes))}")
