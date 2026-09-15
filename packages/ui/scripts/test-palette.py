#!/usr/bin/env python3
"""Feed check-palette.py inputs that must fail, and assert it fails on them.

Running the checker against the current document proves nothing about the
checker: the current document is valid. The only thing that establishes a check
works is an input it must reject, kept in the tree so it keeps being asserted.

The fixture is the dark palette exactly as it shipped before ADR 0014. It carried
three defects nothing caught at the time: a gradient ending on the page
background, a primary button three times louder than light's, and a notice
indistinguishable from the card under it.
"""
import importlib.util
import io
import json
import pathlib
import sys
import tempfile
from contextlib import redirect_stderr

HERE = pathlib.Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("check_palette", HERE / "check-palette.py")
check_palette = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(check_palette)

CURRENT = HERE.parent / "design" / "tokens.json"
FIXTURE = HERE / "fixtures" / "dark-before-0014.json"


def run(path: pathlib.Path) -> tuple[int, str]:
    captured = io.StringIO()
    with redirect_stderr(captured):
        code = check_palette.main(path)
    return code, captured.getvalue()


def dark(name: str) -> str:
    """The current dark value of a token.

    Constructed failures derive their values from the document rather than
    hard-coding them. A fixture that hard-codes "the surface colour" stops
    reproducing its defect the moment the surface moves, and then asserts nothing
    while still reporting a pass.
    """
    document = json.loads(CURRENT.read_text())
    tokens = document.get("tokens", document)
    return tokens[name]["color"]["dark"]


def mutate(**dark: str) -> pathlib.Path:
    document = json.loads(CURRENT.read_text())
    tokens = document.get("tokens", document)
    for name, value in dark.items():
        tokens[name.replace("_", "-")]["color"]["dark"] = value
    handle = tempfile.NamedTemporaryFile("w", suffix=".json", delete=False)
    handle.write(json.dumps(document))
    handle.close()
    return pathlib.Path(handle.name)


CASES = [
    ("the palette as it shipped before ADR 0014", FIXTURE,
     ["gradient must not end", "must not dominate"]),
    ("a gradient ending on the page background", mutate(gradient_brand_to="#07120D"),
     ["gradient must not end"]),
    ("a primary button brighter than the ceiling", mutate(color_action_primary="#5FE7B8"),
     ["must not dominate"]),
    ("an unreadable primary button label", mutate(color_action_primary="#0A1410"),
     ["primary button label"]),
    ("a card that matches the page and has no border",
     mutate(color_bg_surface=dark("color-bg-canvas"), color_border_default=dark("color-bg-canvas")),
     ["distinguishable from the page"]),
    ("a selected row that is invisible",
     mutate(color_accent_subtle=dark("color-bg-surface"), color_accent=dark("color-bg-surface")),
     ["selected row"]),
]


def main() -> int:
    failures = 0

    code, _ = run(CURRENT)
    if code != 0:
        print("FAIL  the current document should pass and did not", file=sys.stderr)
        failures += 1
    else:
        print("  ok    current document passes")

    for name, path, expected in CASES:
        code, output = run(path)
        if code == 0:
            print(f"FAIL  not rejected: {name}", file=sys.stderr)
            failures += 1
            continue
        missing = [phrase for phrase in expected if phrase not in output]
        if missing:
            print(f"FAIL  rejected for the wrong reason: {name} (missing {missing})", file=sys.stderr)
            failures += 1
            continue
        print(f"  ok    rejected  {name}")

    if failures:
        print(f"\n{failures} palette check(s) did not behave", file=sys.stderr)
        return 1
    print(f"\n{len(CASES)} invalid palettes rejected; the current one passes")
    return 0


if __name__ == "__main__":
    sys.exit(main())
