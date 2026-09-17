#!/usr/bin/env python3
"""The canvas check, fed inputs that must fail.

Checking that the current screens pass proves nothing about the check, because
the current screens match. These assert the opposite: these specific shapes are
refused, and refused for the stated reason.

Two of these are defects this check actually had. It read a component's
`descendants` overrides as unnamed and dropped every notice title and body, so
a whole set of designed failure messages went unasserted and was reinvented in
the screen. And its comment-stripping regex ran `.*` under DOTALL, so one `//`
comment swallowed the rest of the file.
"""

from __future__ import annotations

import importlib.util
import sys
import tempfile
from pathlib import Path

spec = importlib.util.spec_from_file_location(
    "check_canvas", Path(__file__).parent / "check-canvas.py"
)
assert spec and spec.loader
check_canvas = importlib.util.module_from_spec(spec)
spec.loader.exec_module(check_canvas)


FRAME = "J09.A / 01 · Example"
CONTRACT = {FRAME: ["Overview", "{bank} was not connected"]}
IMPLEMENTS = {"J09.A / 01 · Example": ["pages/Example.svelte"]}


def case(name: str, source: str, contract=None, implements=None, accepted=None):
    with tempfile.TemporaryDirectory() as raw:
        src = Path(raw) / "src"
        (src / "pages").mkdir(parents=True)
        (src / "pages" / "Example.svelte").write_text(source)
        problems, _ = check_canvas.check(
            CONTRACT if contract is None else contract,
            src,
            IMPLEMENTS if implements is None else implements,
            {} if accepted is None else accepted,
        )
        return name, problems


# A screen missing a sentence the canvas holds.
MISSING = """<main>
\t<h1>Overview</h1>
</main>
"""

# A screen that discusses the copy instead of rendering it. The explanation is
# where the canvas's own wording tends to appear, which is exactly how a screen
# could satisfy this check while showing something else.
IN_A_COMMENT = """<main>
\t<h1>Overview</h1>
\t<!-- Monzo was not connected, per the frame. -->
\t<p>Something else entirely.</p>
</main>
"""

IN_A_JSDOC = """<script lang="ts">
\t/**
\t * The frame says "{bank} was not connected".
\t */
\tlet { bank } = $props();
</script>
<main><h1>Overview</h1></main>
"""

# The regression for the DOTALL defect: a line comment in the script block must
# not hide the markup underneath it.
AFTER_A_LINE_COMMENT = """<script lang="ts">
\tlet { bank } = $props();
\t// Whatever this says, it is not the markup below.
\tconst x = 1;
</script>
<main>
\t<h1>Overview</h1>
\t<p>{bank} was not connected</p>
</main>
"""

# The interpolation the screen actually writes is not the canvas's placeholder,
# and the sentence around it is what must match.
A_DIFFERENT_VARIABLE = """<main>
\t<h1>Overview</h1>
\t<p>{outcomeBank} was not connected</p>
</main>
"""

# The same hole, a different sentence. Collapsing the interpolation must not
# collapse the difference.
A_DIFFERENT_SENTENCE = """<main>
\t<h1>Overview</h1>
\t<p>{outcomeBank} could not be added</p>
</main>
"""

VALID = """<main>
\t<h1>Overview</h1>
\t<p>{bank} was not connected</p>
</main>
"""


def main() -> int:
    must_fail = [
        case("a screen missing designed copy", MISSING),
        case("copy that only appears in a markup comment", IN_A_COMMENT),
        case("copy that only appears in a doc comment", IN_A_JSDOC),
        case("a sentence with the right hole and the wrong words", A_DIFFERENT_SENTENCE),
        case("a frame no screen is recorded as implementing", VALID, implements={}),
    ]

    must_pass = [
        case("markup below a line comment", AFTER_A_LINE_COMMENT),
        case("the screen's own variable in the canvas's hole", A_DIFFERENT_VARIABLE),
        case(
            "a difference recorded with a reason",
            MISSING,
            accepted={"{bank} was not connected": "the frame is being redrawn"},
        ),
        case("a screen that matches", VALID),
    ]

    failed = 0

    for name, problems in must_fail:
        if problems:
            print(f"  ok    refused  {name}")
        else:
            print(f"  FAIL  accepted {name}", file=sys.stderr)
            failed += 1

    for name, problems in must_pass:
        if problems:
            print(f"  FAIL  refused  {name}: {problems[0]}", file=sys.stderr)
            failed += 1
        else:
            print(f"  ok    accepted {name}")

    if failed:
        print(f"\n{failed} case(s) behaved wrongly", file=sys.stderr)
        return 1

    print(f"\n{len(must_fail)} divergences refused; {len(must_pass)} accepted")
    return 0


if __name__ == "__main__":
    sys.exit(main())
