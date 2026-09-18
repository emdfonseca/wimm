#!/usr/bin/env python3
"""One breakpoint family, everywhere (ADR 0002's declared device axis).

768 / 1200 / 1800 is the only family this product has. A `min-width` or
`max-width` at any other number is a second, undeclared breakpoint family
living beside the real one — which is exactly how 1024 and 599 got in.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ALLOWED = {767, 768, 1199, 1200, 1799, 1800}

PATTERN = re.compile(r"\(\s*(min-width|max-width)\s*:\s*(\d+)px\s*\)")


def check(*roots: Path) -> list[str]:
    problems: list[str] = []
    for root in roots:
        if not root.is_dir():
            continue
        for path in sorted(root.rglob("*.svelte")):
            text = path.read_text(encoding="utf-8")
            for match in PATTERN.finditer(text):
                value = int(match.group(2))
                if value not in ALLOWED:
                    line = text.count("\n", 0, match.start()) + 1
                    problems.append(
                        f"{path}:{line}: {match.group(1)}: {value}px is not one of "
                        f"{sorted(ALLOWED)} — 768 / 1200 / 1800 is the one breakpoint "
                        "family this product has."
                    )
    return problems


def main() -> int:
    ui_src = Path(sys.argv[1] if len(sys.argv) > 1 else "src").resolve()
    web_src = Path(sys.argv[2]) if len(sys.argv) > 2 else ui_src.parent.parent.parent / "apps/web/src"
    problems = check(ui_src, web_src.resolve())
    if problems:
        print("a breakpoint outside the declared family:", file=sys.stderr)
        for problem in problems:
            print(f"  {problem}", file=sys.stderr)
        return 1
    print("every @media (min-width/max-width) matches the declared breakpoint family")
    return 0


if __name__ == "__main__":
    sys.exit(main())
