# 0020 · A screen is finished when it matches its frame

## Status

Accepted; superseded by 0023

A screen and its state stories are the design, so there is no frame to match.
The reasoning against image comparison stands.

## Context

ADR 0008 made the canvas a proposal artifact: a user-facing surface is drawn
before tasks are written, and `canvas.md` records what was drawn. It said
nothing about what happens afterwards, and nothing checked.

So this happened. The four components in `connect-bank-accounts` were built
from their origins in `product-ui.lib.pen` — geometry, tokens, states, all
traceable. Every *screen* was composed from its task description instead, and
the frames were never opened. The results diverged:

```text
frame                                    built
Overview                                 Your money
Refresh balances                         Refresh
Household total, a labelled tile at 300  a bare figure
accounts nested under their bank         one flat list
Connected by … · access ends …           an invented "Banks" block
No banks connected                       Nothing here yet
```

Every check passed throughout. `check-geometry.py` asserts four measurements —
card 480, panel 560, sidebar 264, border-box — on atoms and templates. No
screen was compared to any frame, so the gap was filled by whatever seemed
reasonable at the time, repeatedly, across a whole change.

## Decision

**A UI implementation is not finished until it matches its pen frame.** Not
"captures the intent", not "close enough" — matches. The frame is the design of
record and the screen is its implementation, in that order.

So writing a screen starts by reading its frame, not its task description:

```bash
jq -r '.. | objects | select(.name != null and (.name | startswith("J"))) | .name' \
  apps/web/design/*.pen | sort -u
```

A `.pen` is pretty-printed JSON. Read the node tree directly and take from it
the hierarchy and its order, every `layout`, `gap`, `padding`, `width`,
`height`, `fontSize` and `fontWeight`, every piece of copy, and which library
component each `ref` instances. A structure the frame does not have is not a
detail to decide later; it is a divergence.

**`just check` enforces the part a script can hold.**

```text
gen-canvas-contract.py   each drawn frame's static copy -> canvas-contract.json
check-canvas.py          the screens implementing a frame contain that copy
test-contract.py         the generator extracts copy and leaves fixtures
test-canvas.py           the check refuses the differences it must refuse
check-geometry.py        the measurements the canvas records
```

`check-canvas.py` maps every drawn frame to the files that render it, and a
frame with no mapping fails: a drawn screen nobody built is a gap rather than a
silence. Comments are stripped before matching, because otherwise a screen
satisfies the check by explaining the copy instead of rendering it — which the
first version of this check allowed.

**The generator is tested harder than the check, because it fails quietly.** A
check that refuses too much is loud and gets fixed within the hour. A generator
that extracts too little just asserts less, nothing fails, and the screens
drift in the gap it left. Both of its first two defects were that:

```text
a component instance overrides its innards through a `descendants` map keyed
by library id, and those entries carry no name — so every notice title and
body was read as unnamed and dropped, and the designed failure messages were
reinvented in the screen with different wording

any sentence holding a figure was dropped whole, which took the two messages
naming a clock time and a date with it
```

The first cost 32 of the 59 strings now asserted. Overrides are resolved
through the library's own id-to-name map, and a fixture inside a sentence is
substituted for a placeholder — `{bank}`, `{member}`, `{time}`, `{date}`,
`{count}` — rather than disqualifying the sentence. The check collapses those
holes on both sides, so what is compared is the shape of the sentence and the
words around the hole.

**Fixture data is excluded by node name, never by guessing from text.** A frame
draws `Monzo` and `€11,693.55`; a screen renders what the household has.
Guessing flagged `Montepio` and `Conta à Ordem` as copy a screen must contain,
which would have made the check unusable within a day. Copy lives in titles,
ledes, labels, helpers and notices; `Name`, `Meta`, `Balance`, `Bank` and `Who`
hold data.

**The checks are a floor and not the standard.** They hold copy and a handful of
numbers. They cannot see that accounts were drawn under their bank, or that a
total is a tile rather than a figure. Passing them is not evidence of fidelity.

**A deliberate difference is recorded with its reason** — in `ACCEPTED` for
copy, in the change's `canvas.md` otherwise, with the frame redrawn to agree.
Both directions are legitimate; an undecided difference is not.

*Alternative:* image comparison, exporting each frame to PNG and diffing it
against a screenshot of the built screen. That is what "pixel perfect" means
literally and it is the obvious next step. It is not this decision because a
diff across two renderers needs a tolerance, and a tolerance loose enough to
pass two different text engines is loose enough to miss the divergences listed
above. Copy and structure caught all of them.

## Consequences

This ADR is the only place the rule lives. `.claude/rules/decisions.md` is
generated from `docs/decisions/` and carries each decision's body, so an ADR is
already a loaded rule — a separate rule file and a paragraph in
`openspec/config.yaml` were the first draft of this and were removed. Three
copies of a rule are three things to keep in agreement, and the one that drifts
is the one somebody reads.

`canvas-contract.json` is generated and never edited, like `tokens.css`. A
change to a frame regenerates it and fails the check until the screen follows —
which is the intended direction: the canvas moves first.

The mapping in `check-canvas.py` has to be extended when a screen is added.
That is the point at which somebody states which frames it implements, and it
is the cheapest moment to notice that a frame has none.
