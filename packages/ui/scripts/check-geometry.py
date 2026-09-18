#!/usr/bin/env python3
"""The measurements the canvas records, asserted against the stylesheets.

These four numbers were all wrong at once and none of them looked wrong: the
declarations said 480 and 560, and the browser laid out 546 and 688, because
nothing set box-sizing. A number that is right in the source and wrong on screen
is the kind of defect that gets "fixed" by changing the number.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path


def check(src: Path) -> list[str]:
    problems: list[str] = []

    base = src / "base.css"
    if not base.is_file():
        return [f"{base.name} is missing: it carries the border-box reset"]

    text = base.read_text(encoding="utf-8")
    if not re.search(r"\*\s*,\s*\*::before\s*,\s*\*::after\s*\{[^}]*box-sizing:\s*border-box", text, re.S):
        problems.append(
            "base.css does not set box-sizing: border-box on every element. "
            "The canvas measures whole frames; without this the browser adds "
            "padding on top of every width the design system declares."
        )

    # Each is (file, selector, property, expected) straight from canvas.md.
    expected = [
        ("templates/AuthShell.svelte", ".card", "max-inline-size", "480px"),
        ("templates/AuthShell.svelte", ".panel", "flex", "0 0 560px"),
        # `layout-sidebar-width`: 264 at Wide, 288 at Ultra (`NEIet`) — one
        # token rather than a literal, because the two regimes disagree on the
        # number and a literal can only ever be right for one of them.
        ("organisms/SidebarNav.svelte", ".sidebar", "inline-size", "var(--layout-sidebar-width)"),
        # The rail (`WMlvF`): 72 at Medium, the same token.
        ("templates/SignedInLanding.svelte", ".rail", "inline-size", "var(--layout-sidebar-width)"),
        # Slot `xNVFG`'s child is always a `Page`, which owns its own gutter —
        # for its content, not its header. The slot itself had this same
        # padding too, which inset the header the frames draw edge to edge.
        ("templates/Page.svelte", ".body", "padding", "var(--layout-page-gutter)"),
    ]

    # A screen of content fills its slot and packs to the top, which is how
    # every frame's Page is drawn: height fill_container with a trailing
    # spacer. The slot centres on both axes, so a screen that does not fill
    # floats into the middle of the page — which is what happened to all four
    # of these at once, and what no measurement above would have caught.
    for screen in sorted((src / "pages").glob("*.svelte")):
        if screen.name.endswith(".stories.svelte"):
            continue
        body = screen.read_text(encoding="utf-8")
        if ".screen {" not in body:
            continue
        rule = body[body.index(".screen {"):]
        rule = rule[: rule.index("}")]
        if "flex: 1" not in rule:
            problems.append(
                f"pages/{screen.name}: .screen does not declare flex: 1, so the "
                "shell's slot will centre it instead of the frame's own Page "
                "filling and packing content to the top."
            )

    # The token itself, at the three regimes that give it a value: 72 at
    # Medium (the rail, `WMlvF`), 264 at Wide, 288 at Ultra (`NEIet`). A screen
    # can declare `var(--layout-sidebar-width)` correctly and still be wrong if
    # the token's own per-regime value drifts — which is what the check above
    # cannot see.
    tokens_css = src / "tokens.css"
    if not tokens_css.is_file():
        problems.append("tokens.css is missing")
    else:
        css = tokens_css.read_text(encoding="utf-8")
        for min_width, want_px, device in ((768, 72, "medium"), (1200, 264, "wide"), (1800, 288, "ultra")):
            block = re.search(
                rf"@media \(min-width:\s*{min_width}px\)\s*\{{(.*?\n\}}\n)", css, re.S
            )
            if not block or f"--layout-sidebar-width: {want_px}px" not in block.group(1):
                problems.append(
                    f"tokens.css: --layout-sidebar-width is not {want_px}px at "
                    f"{device} (min-width: {min_width}px), which is what the "
                    "canvas measures there."
                )

    for filename, selector, prop, want in expected:
        path = src / filename
        if not path.is_file():
            problems.append(f"{filename} is missing")
            continue
        body = path.read_text(encoding="utf-8")
        if f"{prop}: {want}" not in body:
            problems.append(
                f"{filename}: {selector} no longer declares {prop}: {want}, "
                "which is what the canvas measures."
            )

    return problems


def main() -> int:
    src = Path(sys.argv[1] if len(sys.argv) > 1 else "src").resolve()
    problems = check(src)
    if problems:
        print("geometry does not match the canvas:", file=sys.stderr)
        for problem in problems:
            print(f"  {problem}", file=sys.stderr)
        return 1
    print("geometry matches the canvas (border-box, card 480, panel 560, sidebar/rail tokenised)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
