#!/usr/bin/env python3
"""The breakpoint check is only worth having if it still fails.

Feeds it a file carrying the exact defect that shipped — a 1024 breakpoint
beside the declared 768/1200/1800 family — and asserts it is refused.
"""
from __future__ import annotations

import importlib.util
import pathlib
import sys
import tempfile

spec = importlib.util.spec_from_file_location(
    "check_breakpoints", pathlib.Path(__file__).parent / "check-breakpoints.py"
)
assert spec and spec.loader
cb = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cb)


def main() -> int:
    with tempfile.TemporaryDirectory() as tmp:
        root = pathlib.Path(tmp)
        (root / "AuthShell.svelte").write_text(
            "<style>\n@media (min-width: 1024px) {\n.shell { flex-direction: row; }\n}\n</style>\n"
        )
        problems = cb.check(root, pathlib.Path("/nonexistent"))
        if not problems:
            print("the breakpoint check accepted a 1024px media query", file=sys.stderr)
            return 1
        if not any("1024" in p for p in problems):
            print(f"the breakpoint check refused for the wrong reason: {problems}", file=sys.stderr)
            return 1

    with tempfile.TemporaryDirectory() as tmp:
        root = pathlib.Path(tmp)
        (root / "Fine.svelte").write_text(
            "<style>\n@media (min-width: 768px) { .a {} }\n"
            "@media (min-width: 1200px) { .b {} }\n"
            "@media (min-width: 1800px) { .c {} }\n"
            "@media (max-width: 767px) { .d {} }\n"
            "</style>\n"
        )
        problems = cb.check(root, pathlib.Path("/nonexistent"))
        if problems:
            print(f"the breakpoint check refused the declared family: {problems}", file=sys.stderr)
            return 1

    print("the breakpoint check refuses a 1024px media query and accepts the declared family")
    return 0


if __name__ == "__main__":
    sys.exit(main())
