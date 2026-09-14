#!/usr/bin/env python3
"""The token document schema, shared by the generator and the checker.

It lives here rather than in either of them because both need it and only one
can own it: when the generator did not validate, `just build` happily rendered
a malformed colour into tokens.css and exited 0, overwriting a valid stylesheet.

SUPPORTED_AXES is the single declaration of what the generator can render —
every axis and every value within it. Names alone are not enough: a `color` axis
that gained a `sepia` value would satisfy an axis-name check, satisfy the
completeness check because every token supplied all three branches, and then be
dropped without a word.
"""
import math, re

NAME = re.compile(r"^[a-z][a-z0-9]*(-[a-z0-9]+)*$")
HEX = re.compile(r"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$")
_NUM = r"-?(?:\d+\.?\d*|\.\d+)"
BEZIER = re.compile(rf"^cubic-bezier\(\s*({_NUM})\s*,\s*({_NUM})\s*,\s*({_NUM})\s*,\s*({_NUM})\s*\)$")

TOKEN_FIELDS = {"type", "unit", "default"}
TYPE_UNITS = {
    "color": {None},
    "dimension": {"px", "rem", "%"},
    "duration": {"ms", "s"},
    "number": {None},
    "fontFamily": {None},
    "cubicBezier": {None},
}
SUPPORTED_AXES = {
    "color": ("light", "dark"),
    "device": ("compact", "medium", "wide", "ultra"),
    "density": ("comfortable", "compact"),
}


def _finite(v):
    """A real number. json.loads accepts Infinity and NaN by default, and a bare
    isinstance check lets them through to be rendered as `infpx`."""
    return isinstance(v, (int, float)) and not isinstance(v, bool) and math.isfinite(v)


def value_ok(kind, v):
    if kind == "color":
        return isinstance(v, str) and bool(HEX.match(v))
    if kind in ("dimension", "number"):
        return _finite(v)
    if kind == "duration":
        return _finite(v) and v >= 0
    if kind == "fontFamily":
        return isinstance(v, str) and bool(v.strip())
    if kind == "cubicBezier":
        m = BEZIER.match(v) if isinstance(v, str) else None
        if not m:
            return False
        x1, y1, x2, y2 = (float(g) for g in m.groups())
        # CSS constrains the control-point x values to [0, 1]; y is unbounded.
        return 0.0 <= x1 <= 1.0 and 0.0 <= x2 <= 1.0
    return False


def validate(doc):
    """Return a list of problems. Empty means the document is renderable."""
    fails = []
    themes = doc.get("themes", {})
    tokens = doc.get("tokens", {})
    if not tokens:
        fails.append("document declares no tokens")

    supported = {k: set(v) for k, v in SUPPORTED_AXES.items()}
    for axis, values in themes.items():
        if not NAME.match(axis) or not values:
            fails.append(f"theme axis {axis!r} is not a valid axis name with values")
            continue
        if axis not in supported:
            fails.append(f"theme axis {axis!r} is declared but the generator cannot "
                         f"render it — its values would be dropped silently. "
                         f"Supported axes: {sorted(supported)}")
            continue
        extra = set(values) - supported[axis]
        if extra:
            fails.append(f"theme axis {axis!r} declares {sorted(extra)}, which the "
                         f"generator cannot render — those branches would be dropped "
                         f"silently. It emits {sorted(supported[axis])}.")
        absent = supported[axis] - set(values)
        if absent:
            fails.append(f"theme axis {axis!r} is missing {sorted(absent)}, which the "
                         f"generator expects to emit.")

    for name, spec in tokens.items():
        if not NAME.match(name):
            fails.append(f"{name}: not kebab-case")
        kind = spec.get("type")
        if kind not in TYPE_UNITS:
            fails.append(f"{name}: unknown type {kind!r}")
            continue
        if "unit" not in spec:
            fails.append(f"{name}: missing explicit unit (use null where unitless)")
        elif spec["unit"] not in TYPE_UNITS[kind]:
            fails.append(f"{name}: unit {spec['unit']!r} invalid for type {kind}")

        axes = set(spec) - TOKEN_FIELDS
        unknown = axes - set(themes)
        if unknown:
            fails.append(f"{name}: unknown axis {sorted(unknown)}")
        if not axes and "default" not in spec:
            fails.append(f"{name}: has neither a default nor any axis")
        if axes and "default" in spec:
            fails.append(f"{name}: has both a default and axis values — pick one")

        for axis in axes & set(themes):
            got, want = set(spec[axis]), set(themes[axis])
            if got != want:
                fails.append(f"{name}: {axis} has {sorted(got)}, expected {sorted(want)}")
            for branch, v in spec[axis].items():
                if not value_ok(kind, v):
                    fails.append(f"{name}.{axis}.{branch}: {v!r} is not a valid {kind}")
        if "default" in spec and not value_ok(kind, spec["default"]):
            fails.append(f"{name}.default: {spec['default']!r} is not a valid {kind}")
    return fails
