#!/usr/bin/env python3
"""The copy the canvas holds, extracted so a screen can be checked against it.

Every screen in this design system was drawn before it was built, and nothing
tied the built screen to its drawing. check-geometry.py asserts four
measurements on atoms and templates; no screen was checked against any frame.
So the screens diverged — "Your money" for a heading the canvas writes as
"Overview", a flat account list where the canvas nests accounts under their
bank — and nothing failed.

This writes what the frames say. check-canvas.py asserts the screens say it.

Only *static* copy is extracted. A frame's fixture data — bank names, amounts,
account numbers, member names — is drawn as placeholder content and must not
become an assertion: the screen renders whatever the household actually has.
What is left is the labels and the sentences, which is exactly the part a
screen is supposed to reproduce.
"""

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
JOURNEYS = sorted((ROOT / "apps/web/design").glob("*.pen"))
LIBRARY = Path(__file__).resolve().parents[1] / "design/product-ui.lib.pen"
OUT = Path(__file__).resolve().parents[1] / "src/canvas-contract.json"

# A drawn screen state, named "J03.A / 04 · Choose accounts / Wide / …".
#
# The step number after the slash is what distinguishes a screen from a journey
# header ("J04 · Keep the household's balances current") or a stage note
# ("J05.A · Disconnect a bank · PRIMARY SUCCESS"), both of which carry prose
# about the flow rather than copy on a screen.
SCREEN = re.compile(r"^J\d+\.\w+ / \d+")

# Nodes whose text is fixture data: a bank's name, an account's name, an
# amount, a member. The screen renders these from real rows, so asserting the
# canvas's placeholders would assert the fixture rather than the design.
#
# Classified by the node's name rather than by guessing from its text. Guessing
# flagged "Montepio" and "Conta à Ordem" as copy a screen must contain, which
# would have made the check unusable and then ignored.
DATA_NODES = {
    "Name", "Meta", "Balance", "Bank", "Who", "Reading", "Amount",
    "Value", "Mark", "Initials", "Badge",
}

# The fixtures a frame draws, and what each one stands for. Substitution runs
# before DATA_TEXT, so a sentence is kept for its shape rather than dropped for
# the figure sitting in it.
#
# Copy like "Monzo was not connected" is a template with a bank interpolated,
# not fixture data. Excluding it let a whole set of designed failure messages
# go unasserted, and they were promptly reinvented in the screen with different
# wording — a clock time or a date in the sentence hid two more of them.
#
# Spelled-out "one" is not a count here: it is a common word in this copy
# ("one of these"), and substituting it would compare a sentence that does not
# exist.
FIXTURES = (
    (re.compile(r"Monzo"), "{bank}"),
    (re.compile(r"Ana Reis"), "{member}"),
    (re.compile(r"\b\d{1,2}:\d{2}\b"), "{time}"),
    (re.compile(
        r"\b\d{1,2} (?:January|February|March|April|May|June|July|August"
        r"|September|October|November|December)\b"
    ), "{date}"),
    (re.compile(
        r"\b(?:two|three|four|five|six|seven|eight|nine|ten)\b"
    ), "{count}"),
)

# Whatever figure is left after that is fixture data — an amount, an account
# number, a balance — and the sentence around it is not a template.
DATA_TEXT = re.compile(r"\d")

# A plate is documentation about a frame, not copy inside it.
PLATE = ("Reached when:", "WIDE", "COMPACT", "ULTRA", "MEDIUM")


def library_names(library=None):
    """Every library node id to its name.

    A journey overrides a component's innards through a `descendants` map keyed
    by library id, and those entries carry no name of their own. Without this
    lookup they are all unnamed, so every notice title and body — the designed
    copy for each failure state — was dropped from the contract, went
    unasserted, and was then reinvented in the screen with different wording.
    """
    names = {}

    def walk(node):
        if isinstance(node, dict):
            if node.get("id") and node.get("name"):
                names[node["id"]] = node["name"]
            for value in node.values():
                walk(value)
        elif isinstance(node, list):
            for item in node:
                walk(item)

    walk(json.loads((library or LIBRARY).read_text()))
    return names


NAMES = library_names()


def named(key, node, names=None):
    """The name of a node, resolving a `descendants` key through the library."""
    if isinstance(node, dict) and isinstance(node.get("name"), str):
        return node["name"]
    if isinstance(key, str):
        # "ui:Z5ONBF" or "ui:vaLOy/ui:rv5n0" — the last segment is the node.
        return (NAMES if names is None else names).get(
            key.split("/")[-1].removeprefix("ui:")
        )
    return None


def static_copy(node, frame=None, out=None, names=None):
    """Collect the static text of every drawn screen, in document order."""
    if out is None:
        out = {}

    if isinstance(node, dict):
        name = node.get("name")
        if isinstance(name, str) and SCREEN.match(name) and "plate" not in name:
            frame = name
            out.setdefault(frame, [])

        for key, value in node.items():
            resolved = named(key, node, names) or name
            readable = isinstance(resolved, str) and resolved not in DATA_NODES

            if key == "content" and isinstance(value, str) and frame and readable:
                text = " ".join(value.split())
                for fixture, placeholder in FIXTURES:
                    text = fixture.sub(placeholder, text)
                if text and not DATA_TEXT.search(text) and not text.startswith(PLATE):
                    out[frame].append(text)
            elif key == "descendants" and isinstance(value, dict):
                # Each entry is keyed by the library id it overrides, which is
                # what `named` resolves.
                for override_key, override in value.items():
                    static_copy(
                        {"name": named(override_key, override, names), **override},
                        frame,
                        out,
                        names,
                    )
            else:
                static_copy(value, frame, out, names)

    elif isinstance(node, list):
        for item in node:
            static_copy(item, frame, out, names)

    return out


def main() -> int:
    if not JOURNEYS:
        print("no journey .pen files found", file=sys.stderr)
        return 1

    contract = {}
    for journey in JOURNEYS:
        # A .pen is pretty-printed JSON, read directly rather than through a
        # round trip (CLAUDE.md).
        found = static_copy(json.loads(journey.read_text()))
        for frame, lines in found.items():
            if lines:
                contract[frame] = lines

    if not contract:
        print("no drawn screens found in the journeys", file=sys.stderr)
        return 1

    OUT.write_text(json.dumps(contract, indent=2, ensure_ascii=False, sort_keys=True) + "\n")
    print(f"wrote {OUT.relative_to(ROOT)} ({len(contract)} frames)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
