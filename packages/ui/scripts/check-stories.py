#!/usr/bin/env python3
"""Every component carries a story.

A component with no story cannot be seen in isolation, in either theme, at any
viewport, and its behavioural contract - focus, live regions, disabled states -
is asserted nowhere. `.claude/rules/typescript.md` puts component behaviour in
Storybook play functions rather than unit tests, which only works if the story
exists.

Reading a rule is not a guarantee. This is.
"""

from __future__ import annotations

import sys
from pathlib import Path

# Atomic design, mirroring the pen library's own zoning. A template and a page
# are components; they get stories like anything else.
LAYERS = ("atoms", "molecules", "organisms", "templates", "pages")

TITLES = {
    "atoms": "Atoms",
    "molecules": "Molecules",
    "organisms": "Organisms",
    "templates": "Templates",
    "pages": "Pages",
}


def components(src: Path) -> list[Path]:
    found: list[Path] = []
    for layer in LAYERS:
        directory = src / layer
        if not directory.is_dir():
            continue
        found.extend(
            path
            for path in sorted(directory.rglob("*.svelte"))
            if not path.name.endswith(".stories.svelte")
        )
    return found


def check(src: Path) -> list[str]:
    problems: list[str] = []

    for component in components(src):
        story = component.parent / (component.stem + ".stories.svelte")
        rel = component.relative_to(src.parent)

        layer = next((p for p in component.parts if p in TITLES), None)
        expected = f"{TITLES[layer]}/{component.stem}" if layer else None

        if not story.is_file():
            problems.append(
                f"{rel}: no story. Create {story.relative_to(src.parent)} "
                f"with title '{expected}'."
            )
            continue

        text = story.read_text(encoding="utf-8")

        if expected and expected not in text:
            problems.append(
                f"{story.relative_to(src.parent)}: title is not '{expected}'. "
                "The sidebar follows the design taxonomy, not the file path."
            )

    # A component outside every layer has no place in the taxonomy, so nothing
    # above would have checked it.
    stray = [
        path
        for path in sorted(src.glob("*.svelte"))
        if not path.name.endswith(".stories.svelte")
    ]
    problems.extend(
        f"{path.relative_to(src.parent)}: not in a layer. Move it under "
        f"src/{{{', '.join(LAYERS)}}}/."
        for path in stray
    )

    return problems


def main() -> int:
    src = Path(sys.argv[1] if len(sys.argv) > 1 else "src").resolve()
    if not src.is_dir():
        print(f"no such directory: {src}", file=sys.stderr)
        return 2

    problems = check(src)
    if problems:
        print("components without a story:", file=sys.stderr)
        for problem in problems:
            print(f"  {problem}", file=sys.stderr)
        return 1

    total = len(components(src))
    print(f"every component has a story ({total} checked)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
