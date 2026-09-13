#!/usr/bin/env python3
"""Regenerate the pen-design-standard skill references from the canonical standard.

The standard in docs/ is the single source of truth; the skill's references/ are a
split of it for progressive disclosure. Run this after editing the standard.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "docs" / "pen-dev-product-design-organization-standard.md"
DST = ROOT / ".claude" / "skills" / "pen-design-standard" / "references"

# Section number (or "appendix") -> reference filename.
SECTIONS = {
    "1": "conventions.md",  # also carries the document front matter
    "2": "pen-mechanics.md",
    "3": "ownership-and-repo.md",
    "4": "project-setup.md",
    "5": "accessibility-and-responsive.md",
    "6": "navigation-and-surfaces.md",
    "7": "journeys-and-canvas.md",
    "8": "design-system-library.md",
    "9": "lifecycle.md",
    "10": "verification-gates.md",
    "11": "design-to-code.md",
    "12": "worked-example.md",
    "appendix": "appendices.md",
}

HEADER = (
    "<!-- Generated from docs/pen-dev-product-design-organization-standard.md\n"
    "     by tools/split-standard.py. Edit the standard, not this file. -->\n"
)

TOC_THRESHOLD = 300


def section_key(heading: str) -> str | None:
    """Map a '## ...' heading to its SECTIONS key."""
    if heading.startswith("## Appendix"):
        return "appendix"
    m = re.match(r"^## (\d+)\. ", heading)
    return m.group(1) if m else None


def add_toc(body: str) -> str:
    """Give long references a table of contents so the reader can jump."""
    lines = body.split("\n")
    if len(lines) < TOC_THRESHOLD:
        return body
    entries = [l.lstrip("# ").strip() for l in lines if l.startswith("### ")]
    if not entries:
        return body
    top = next(n for n, l in enumerate(lines) if l.startswith("## "))
    toc = ["", "**In this file:**", ""] + [f"- {e}" for e in entries] + [""]
    lines[top + 1 : top + 1] = toc
    return "\n".join(lines)


def main() -> int:
    text = SRC.read_text()
    lines = text.split("\n")

    # Where each mapped section starts. Everything before the first one is front matter.
    starts: list[tuple[int, str]] = []
    for n, line in enumerate(lines):
        if line.startswith("## "):
            key = section_key(line)
            if key and (not starts or starts[-1][1] != key):
                starts.append((n, key))

    missing = set(SECTIONS) - {k for _, k in starts}
    if missing:
        print(f"error: no heading found for sections {sorted(missing)}", file=sys.stderr)
        return 1

    front = "\n".join(lines[: starts[0][0]])
    # The document's own table of contents is navigation for the single-file
    # reading; inside a split reference it points at anchors that do not exist.
    front = re.sub(r"\*\*Contents\*\*\n\n(?:.*\n)*?\[Appendix A[^\n]*\n\n", "", front)

    DST.mkdir(parents=True, exist_ok=True)
    for existing in DST.glob("*.md"):
        existing.unlink()

    bounds = [(start, key, starts[i + 1][0] if i + 1 < len(starts) else len(lines))
              for i, (start, key) in enumerate(starts)]

    for start, key, end in bounds:
        body = "\n".join(lines[start:end])
        if key == "1":
            body = front + body
        out = DST / SECTIONS[key]
        out.write_text(HEADER + add_toc(body))
        print(f"{out.relative_to(ROOT)}  ({len(body.splitlines())} lines)")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
