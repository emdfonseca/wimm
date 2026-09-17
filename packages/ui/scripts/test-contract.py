#!/usr/bin/env python3
"""The contract generator, fed frames whose copy it must and must not extract.

This is the half of canvas fidelity that fails silently. A check that refuses
too much is loud; a generator that extracts too little just asserts less, and
the screens drift in the gap. Both of these are defects it had:

- A component instance overrides its innards through a `descendants` map keyed
  by library id, and those entries carry no name. Reading them as unnamed
  dropped every notice title and body — a whole set of designed failure
  messages went unasserted and was promptly reinvented in the screen.
- Any sentence holding a figure was dropped whole, so a message naming a clock
  time or a date was never compared either.

And the thing it must keep doing: leaving fixture data out. Guessing from the
text flagged "Montepio" and "Conta à Ordem" as copy a screen must contain,
which would have made the check unusable and then ignored.
"""

from __future__ import annotations

import importlib.util
import json
import sys
import tempfile
from pathlib import Path

spec = importlib.util.spec_from_file_location(
    "gen_canvas_contract", Path(__file__).parent / "gen-canvas-contract.py"
)
assert spec and spec.loader
gen = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gen)


# A library whose ids carry the names a journey's overrides do not.
LIBRARY = {
    "nodes": [
        {"id": "Z5ONBF", "name": "Title"},
        {"id": "NAIMe", "name": "Message"},
        {"id": "egfCw", "name": "Name"},
        {"id": "kELqu", "name": "Meta"},
        {"id": "JEiqn", "name": "Balance"},
        {"id": "WIXtr", "name": "Title"},
        {"id": "HViAc", "name": "Body"},
        {"id": "A4SKK", "name": "Action"},
        {"id": "iYHDG", "name": "Label"},
    ]
}


def journey(overrides: dict) -> dict:
    """One drawn screen whose component instance overrides its innards."""
    return {
        "nodes": [
            {
                "name": "J09.A / 01 · Example / A state",
                "children": [{"ref": "ui:WFSg5", "descendants": overrides}],
            }
        ]
    }


def extract(overrides: dict) -> list[str]:
    with tempfile.TemporaryDirectory() as raw:
        library = Path(raw) / "product-ui.lib.pen"
        library.write_text(json.dumps(LIBRARY))
        names = gen.library_names(library)
        found = gen.static_copy(journey(overrides), names=names)
        return next(iter(found.values()), [])


def main() -> int:
    checks = [
        (
            "a notice's title and body, which are id-keyed overrides",
            {"ui:Z5ONBF": {"content": "Monzo was not connected"},
             "ui:NAIMe": {"content": "Access was not granted at Monzo."}},
            ["{bank} was not connected", "Access was not granted at {bank}."],
        ),
        (
            "a sentence holding a clock time",
            {"ui:NAIMe": {"content": "The next refresh is possible after 14:20."}},
            ["The next refresh is possible after {time}."],
        ),
        (
            "a sentence holding a date",
            {"ui:NAIMe": {"content": "Access runs until 11 September."}},
            ["Access runs until {date}."],
        ),
        (
            "a sentence holding a spelled-out count",
            {"ui:NAIMe": {"content": "Its three accounts are no longer shown."}},
            ["Its {count} accounts are no longer shown."],
        ),
        (
            "a member's name, which is a fixture and not copy",
            {"ui:NAIMe": {"content": "Ana Reis granted the access."}},
            ["{member} granted the access."],
        ),
        (
            "nothing from an account's own name, bank or amount",
            {"ui:egfCw": {"content": "Conta à Ordem"},
             "ui:kELqu": {"content": "Montepio · ••••4021"},
             "ui:JEiqn": {"content": "€14,208.55"}},
            [],
        ),
        (
            "nothing from an amount sitting in an unnamed override",
            {"ui:unknown": {"content": "€6,111.02"}},
            [],
        ),
        # The ledger's frames carry both of their reasons this way: a notice
        # explaining why a bank is not contributing, and an empty state saying
        # why there is nothing to show. Both are id-keyed overrides with no
        # name of their own, which is exactly what the generator once dropped —
        # and dropping them here would leave every reason a member is given
        # unasserted, in the one screen whose whole job is to give reasons.
        (
            "an empty state's title and body on a ledger frame",
            {"ui:WIXtr": {"content": "Nothing read from Monzo"},
             "ui:HViAc": {"content": "There is nothing to show until the bank "
                                     "starts sending transactions."}},
            ["Nothing read from {bank}",
             "There is nothing to show until the bank starts sending transactions."],
        ),
        (
            "a notice action's label, nested a level deeper",
            {"ui:A4SKK/ui:iYHDG": {"content": "Widen at Monzo"}},
            ["Widen at {bank}"],
        ),
    ]

    failed = 0
    for name, overrides, wanted in checks:
        got = extract(overrides)
        if got == wanted:
            print(f"  ok    {name}")
        else:
            print(f"  FAIL  {name}\n        wanted {wanted}\n        got    {got}", file=sys.stderr)
            failed += 1

    if failed:
        print(f"\n{failed} case(s) extracted the wrong copy", file=sys.stderr)
        return 1

    print(f"\n{len(checks)} extraction case(s) hold")
    return 0


if __name__ == "__main__":
    sys.exit(main())
