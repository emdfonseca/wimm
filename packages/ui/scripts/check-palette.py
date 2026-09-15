#!/usr/bin/env python3
"""Assert the palette's relationships, in every theme.

check-tokens.py asks whether the document is well formed and whether the
stylesheet agrees with it. Neither question can see a button that is legible and
still three times louder than anything around it, a gradient that ends on the
page background, or a card that differs from the canvas by a ratio of 1.10. Those
are relationships between values, and until this file existed nothing checked
them: the dark theme shipped with gradient-brand-to identical to color-bg-canvas
and no run ever failed.

Two kinds of assertion, and the second is the one that was missing:

  contrast    a foreground must be readable on a background (WCAG 2.2 AA)
  separation  two adjacent surfaces must be distinguishable from each other,
              and a fill must not be so bright it dominates the page
"""
import json
import pathlib
import sys

TOKENS = pathlib.Path(__file__).resolve().parent.parent / "design" / "tokens.json"

# foreground, background, minimum ratio, what it is
CONTRAST = [
    ("color-text-primary", "color-bg-surface", 4.5, "heading on a card"),
    ("color-text-secondary", "color-bg-surface", 4.5, "body on a card"),
    ("color-text-primary", "color-bg-canvas", 4.5, "heading on the page"),
    ("color-text-secondary", "color-bg-canvas", 4.5, "body on the page"),
    ("color-text-placeholder", "color-bg-surface", 4.5, "placeholder on a card"),
    ("color-action-on-primary", "color-action-primary", 4.5, "primary button label"),
    ("color-action-on-destructive", "color-action-destructive", 4.5, "destructive button label"),
    ("color-accent", "color-bg-surface", 4.5, "link on a card"),
    ("color-accent", "color-accent-subtle", 4.5, "selected item label"),
    ("color-accent", "color-feedback-info-bg", 4.5, "info notice text"),
    ("color-feedback-error", "color-feedback-error-bg", 4.5, "error notice text"),
    ("color-feedback-success", "color-feedback-success-bg", 4.5, "success notice text"),
    ("color-feedback-warning", "color-feedback-warning-bg", 4.5, "warning notice text"),
    ("color-text-on-brand", "gradient-brand-from", 4.5, "headline on the brand panel"),
    ("color-text-on-brand-secondary", "gradient-brand-from", 4.5, "panel body, gradient top"),
    ("color-text-on-brand-secondary", "gradient-brand-to", 4.5, "panel body, gradient bottom"),
    ("color-amount-positive", "color-bg-surface", 4.5, "income amount"),
    ("color-amount-negative", "color-bg-surface", 4.5, "expense amount"),
]

# a, b, minimum ratio, what it is. Where a pair appears as a list of candidates,
# it is enough that ONE of them clears the floor: light separates surfaces with a
# border and dark separates them with luminance, and a rule that demands both
# fails a theme for doing it the other way round.
SEPARATION = [
    ([("color-bg-surface", "color-bg-canvas"), ("color-border-default", "color-bg-canvas")],
     1.20, "a card must be distinguishable from the page, by its fill or its border"),
    ([("color-bg-subtle", "color-bg-surface")],
     1.06, "a subtle region must differ from the card it sits on"),
    ([("color-accent-subtle", "color-bg-surface"), ("color-accent", "color-bg-surface")],
     1.15, "a selected row must be distinguishable, by its fill or its label"),
    ([("color-feedback-error-bg", "color-bg-surface")],
     1.06, "an error notice must differ from the card it sits on"),
    ([("color-feedback-success-bg", "color-bg-surface")],
     1.06, "a success notice must differ from the card it sits on"),
    ([("color-feedback-warning-bg", "color-bg-surface")],
     1.06, "a warning notice must differ from the card it sits on"),
    ([("gradient-brand-to", "color-bg-canvas")],
     1.20, "the brand gradient must not end on the page background"),
]

# fill, reference background, maximum ratio, what it is
# A fill brighter than this stops reading as an action and starts reading as a light source.
LOUDNESS = [
    ("color-action-primary", "color-bg-canvas", 9.5, "the primary button must not dominate the page"),
]


def _linear(channel: float) -> float:
    channel /= 255
    return channel / 12.92 if channel <= 0.04045 else ((channel + 0.055) / 1.055) ** 2.4


def luminance(value: str) -> float:
    value = value.lstrip("#")[:6]
    r, g, b = (int(value[i:i + 2], 16) for i in (0, 2, 4))
    return 0.2126 * _linear(r) + 0.7152 * _linear(g) + 0.0722 * _linear(b)


def ratio(a: str, b: str) -> float:
    la, lb = luminance(a), luminance(b)
    return (max(la, lb) + 0.05) / (min(la, lb) + 0.05)


def resolve(tokens: dict, name: str, theme: str) -> str | None:
    token = tokens.get(name)
    if token is None or token.get("type") != "color":
        return None
    colour = token.get("color")
    if isinstance(colour, dict):
        return colour.get(theme)
    return colour


def main(path: pathlib.Path = TOKENS) -> int:
    document = json.loads(path.read_text())
    tokens = document.get("tokens", document)
    themes = sorted({
        key
        for token in tokens.values()
        if token.get("type") == "color" and isinstance(token.get("color"), dict)
        for key in token["color"]
    })
    if not themes:
        print("check-palette: no colour themes declared", file=sys.stderr)
        return 1

    failures = []
    for theme in themes:
        for fg, bg, floor, what in CONTRAST:
            a, b = resolve(tokens, fg, theme), resolve(tokens, bg, theme)
            if a is None or b is None:
                failures.append(f"{theme}: {what}: {fg} or {bg} is not a colour token")
                continue
            measured = ratio(a, b)
            if measured < floor:
                failures.append(
                    f"{theme}: {what}: {fg} on {bg} is {measured:.2f}, needs {floor} ({a} on {b})"
                )
        for candidates, floor, what in SEPARATION:
            best, shown = 0.0, []
            for one, two in candidates:
                a, b = resolve(tokens, one, theme), resolve(tokens, two, theme)
                if a is None or b is None:
                    failures.append(f"{theme}: {what}: {one} or {two} is not a colour token")
                    best = float("inf")
                    break
                measured = ratio(a, b)
                shown.append(f"{one} against {two} {measured:.2f}")
                best = max(best, measured)
            if best < floor:
                failures.append(f"{theme}: {what}: best is {best:.2f}, needs {floor} ({'; '.join(shown)})")
        for fill, bg, ceiling, what in LOUDNESS:
            a, b = resolve(tokens, fill, theme), resolve(tokens, bg, theme)
            if a is None or b is None:
                failures.append(f"{theme}: {what}: {fill} or {bg} is not a colour token")
                continue
            measured = ratio(a, b)
            if measured > ceiling:
                failures.append(
                    f"{theme}: {what}: {fill} against {bg} is {measured:.2f}, at most {ceiling} ({a}, {b})"
                )

    if failures:
        for failure in failures:
            print(f"check-palette: {failure}", file=sys.stderr)
        print(f"\n{len(failures)} palette relationship(s) failed", file=sys.stderr)
        return 1

    checks = len(themes) * (len(CONTRAST) + len(SEPARATION) + len(LOUDNESS))
    print(f"palette ok: {checks} relationships across {', '.join(themes)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
