#!/usr/bin/env python3
"""The story check, fed inputs that must fail.

Checking that the current tree passes proves nothing about the check, because
the current tree is valid. These assert the opposite: these specific shapes are
refused, and refused for the stated reason.
"""

from __future__ import annotations

import sys
import tempfile
from pathlib import Path

import importlib.util

spec = importlib.util.spec_from_file_location(
    "check_stories", Path(__file__).parent / "check-stories.py"
)
assert spec and spec.loader
check_stories = importlib.util.module_from_spec(spec)
spec.loader.exec_module(check_stories)


COMPONENT = "<script lang=\"ts\">let { x } = $props();</script>\n<p>{x}</p>\n"


def case(name: str, build) -> tuple[str, list[str]]:
    with tempfile.TemporaryDirectory() as raw:
        src = Path(raw) / "src"
        (src / "atoms").mkdir(parents=True)
        build(src)
        return name, check_stories.check(src)


def a_component_with_no_story(src: Path) -> None:
    (src / "atoms" / "Widget.svelte").write_text(COMPONENT)


def a_story_with_the_wrong_title(src: Path) -> None:
    (src / "atoms" / "Widget.svelte").write_text(COMPONENT)
    (src / "atoms" / "Widget.stories.svelte").write_text(
        "<script module>defineMeta({ title: 'Molecules/Widget' });</script>\n"
    )


def a_component_outside_every_layer(src: Path) -> None:
    (src / "Loose.svelte").write_text(COMPONENT)


PAGE_META = "<script module>defineMeta({ title: 'Pages/Ledger' });</script>\n"


def a_page_story_with_no_play_function(src: Path) -> None:
    (src / "pages").mkdir()
    (src / "pages" / "Ledger.svelte").write_text(COMPONENT)
    (src / "pages" / "Ledger.stories.svelte").write_text(
        PAGE_META
        + '<Story name="Default" play={async () => {}} />\n'
        + '<Story name="Loading" args={{ state: \'loading\' }} />\n'
    )


def a_valid_tree(src: Path) -> None:
    (src / "atoms" / "Widget.svelte").write_text(COMPONENT)
    (src / "atoms" / "Widget.stories.svelte").write_text(
        "<script module>defineMeta({ title: 'Atoms/Widget' });</script>\n"
    )


def main() -> int:
    must_fail = [
        ("a component with no story", a_component_with_no_story),
        ("a story with the wrong title", a_story_with_the_wrong_title),
        ("a component outside every layer", a_component_outside_every_layer),
        ("a page story with no play function", a_page_story_with_no_play_function),
    ]

    # The refusal has to say which story, or the repair is a search.
    reasons = {
        "a page story with no play function": (
            "pages/Ledger.stories.svelte",
            "'Loading'",
            "no play function",
        ),
    }

    failures: list[str] = []

    for name, build in must_fail:
        _, problems = case(name, build)
        if not problems:
            failures.append(f"accepted {name}")
        elif missing := [w for w in reasons.get(name, ()) if w not in problems[0]]:
            failures.append(f"refused {name} without saying {missing}: {problems[0]}")
        elif len(problems) > 1:
            failures.append(f"refused more than {name}: {problems}")
        else:
            print(f"refused {name}: {problems[0]}")

    _, problems = case("a valid tree", a_valid_tree)
    if problems:
        failures.append(f"refused a valid tree: {problems}")
    else:
        print("accepted a valid tree")

    if failures:
        for failure in failures:
            print(f"  {failure}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
