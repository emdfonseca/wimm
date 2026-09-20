#!/usr/bin/env python3
"""Every component carries a story, and every page story a play function.

A component with no story cannot be seen in isolation, in either theme, at any
viewport, and its behavioural contract - focus, live regions, disabled states -
is asserted nowhere. `.claude/rules/typescript.md` puts component behaviour in
Storybook play functions rather than unit tests, which only works if the story
exists.

A page's stories are its design of record, and the words of each state are
held by that story's play function. A page story without one holds nothing, so
it is refused. What the play function asserts is not visible from here: that
stays a review matter.

Reading a rule is not a guarantee. This is.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

# Atomic design. A template and a page are components; they get stories like
# anything else.
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


STORY_TAG = re.compile(r"<Story(?=[\s/>])")
STORY_NAME = re.compile(r"""\bname=(?:"([^"]*)"|'([^']*)')""")


def story_tags(text: str) -> list[str]:
    """The opening tag of each <Story>, attributes included.

    A play function is an attribute holding braces and arrows, so the tag ends
    at the first `>` outside every brace and quote, not at the first `>`.
    """
    tags: list[str] = []
    for match in STORY_TAG.finditer(text):
        depth = 0
        quote = ""
        for index in range(match.end(), len(text)):
            char = text[index]
            if quote:
                if char == quote:
                    quote = ""
            elif depth == 0 and char in "\"'":
                quote = char
            elif char == "{":
                depth += 1
            elif char == "}":
                depth -= 1
            elif char == ">" and depth == 0:
                tags.append(text[match.start() : index + 1])
                break
    return tags


def stories_without_play(text: str) -> list[str]:
    missing: list[str] = []
    for tag in story_tags(text):
        if re.search(r"\bplay=", tag):
            continue
        name = STORY_NAME.search(tag)
        missing.append((name.group(1) or name.group(2)) if name else "(unnamed)")
    return missing


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

        if layer == "pages":
            problems.extend(
                f"{story.relative_to(src.parent)}: story '{name}' has no play "
                "function. Assert the words of its state."
                for name in stories_without_play(text)
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
        print("stories missing or incomplete:", file=sys.stderr)
        for problem in problems:
            print(f"  {problem}", file=sys.stderr)
        return 1

    total = len(components(src))
    print(f"every component has a story ({total} checked)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
