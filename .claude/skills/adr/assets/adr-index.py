#!/usr/bin/env python3
"""Generate .claude/rules/decisions.md from docs/decisions/*.md.

A rules file without `paths:` frontmatter is loaded at every session start,
so this puts each ADR's Decision section permanently in context.

Usage: adr-index.py [--check]
  --check   exit 1 if the generated file is missing or stale (CI gate)
"""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
SRC = ROOT / "docs" / "decisions"
OUT = ROOT / ".claude" / "rules" / "decisions.md"

HEADER = (
    "# Decisions\n\n"
    "Generated from `docs/decisions/` by `just adr-index`. Never edit; write an ADR.\n\n"
)


def section(path: Path, text: str, name: str) -> str:
    m = re.search(rf"^## {name}\s*\n(.*?)(?=^## |\Z)", text, re.S | re.M)
    if not m:
        sys.exit(f"{path}: missing '## {name}' section")
    return m.group(1).strip()


def render(path: Path) -> str:
    text = path.read_text()
    m = re.match(r"# (\d{4})\s*[·.]\s*(.+)", text)
    if not m:
        sys.exit(f"{path}: first line must be '# NNNN · Title'")
    num, title = m.groups()
    status = section(path, text, "Status").splitlines()[0]
    decision = section(path, text, "Decision")
    return f"## {num} · {title.strip()} — {status}\n\n{decision}\n"


def main() -> None:
    files = sorted(SRC.glob("[0-9][0-9][0-9][0-9]-*.md"))
    body = HEADER + "\n".join(render(p) for p in files)
    if "--check" in sys.argv:
        if not OUT.exists() or OUT.read_text() != body:
            sys.exit(f"{OUT.relative_to(ROOT)} is stale; run 'just adr-index'")
        return
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(body)
    print(f"wrote {OUT.relative_to(ROOT)} ({len(files)} decisions)")


if __name__ == "__main__":
    main()
