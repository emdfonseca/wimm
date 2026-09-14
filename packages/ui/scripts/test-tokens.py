#!/usr/bin/env python3
"""Regression tests for the token pipeline.

Every case here is a defect that reached review: a malformed value the checker
accepted, an axis branch the generator dropped in silence, a build that
overwrote a good stylesheet with a bad one. Checking that the current document
passes says nothing about any of them, because the current document is valid —
these assert that bad documents are *rejected*.

A check that has never failed is indistinguishable from one that always passes.
"""
import json, pathlib, shutil, subprocess, sys, tempfile

PKG = pathlib.Path(__file__).resolve().parent.parent
BASE = json.loads((PKG / "design/tokens.json").read_text())


def _sepia(d):
    d["themes"]["color"] = ["light", "dark", "sepia"]
    for spec in d["tokens"].values():
        if "color" in spec:
            spec["color"]["sepia"] = spec["color"]["light"]


def _device_xl(d):
    d["themes"]["device"] = ["compact", "medium", "wide", "ultra", "xl"]
    for spec in d["tokens"].values():
        if "device" in spec:
            spec["device"]["xl"] = spec["device"]["ultra"]


def _drop_comfortable(d):
    d["themes"]["density"] = ["compact"]
    for spec in d["tokens"].values():
        if "density" in spec:
            spec["density"].pop("comfortable", None)


# name -> (mutation, must appear in the failure output)
CASES = {
    "malformed hex":        (lambda d: d["tokens"]["color-accent"]["color"].__setitem__("dark", "#GGGGGG"), "not a valid color"),
    "unknown type":         (lambda d: d["tokens"]["color-accent"].__setitem__("type", "banana"), "unknown type"),
    "non-kebab name":       (lambda d: d["tokens"].__setitem__("Color_Accent", {"type": "color", "unit": None, "default": "#000000"}), "not kebab-case"),
    "missing unit":         (lambda d: d["tokens"]["color-accent"].pop("unit"), "missing explicit unit"),
    "unit wrong for type":  (lambda d: d["tokens"]["color-accent"].__setitem__("unit", "px"), "invalid for type"),
    "dimension Infinity":   (lambda d: d["tokens"].__setitem__("space-bad", {"type": "dimension", "unit": "px", "default": float("inf")}), "not a valid dimension"),
    "dimension NaN":        (lambda d: d["tokens"].__setitem__("space-nan", {"type": "dimension", "unit": "px", "default": float("nan")}), "not a valid dimension"),
    "negative duration":    (lambda d: d["tokens"]["motion-duration-base"].__setitem__("default", -5), "not a valid duration"),
    "bezier of dots":       (lambda d: d["tokens"]["motion-ease-enter"].__setitem__("default", "cubic-bezier(., ., ., .)"), "not a valid cubicBezier"),
    "bezier x out of range": (lambda d: d["tokens"]["motion-ease-enter"].__setitem__("default", "cubic-bezier(5, 0, 9, 1)"), "not a valid cubicBezier"),
    "undeclared axis":      (lambda d: d["tokens"]["color-accent"].__setitem__("colour", {"light": "#000000"}), "unknown axis"),
    "missing theme branch": (lambda d: d["tokens"]["color-accent"]["color"].pop("dark"), "expected"),
    "default and axes":     (lambda d: d["tokens"]["color-accent"].__setitem__("default", "#000000"), "pick one"),
    "extra axis value":     (_sepia, "cannot render"),
    "extra device value":   (_device_xl, "cannot render"),
    "missing axis value":   (_drop_comfortable, "is missing"),
}


def run(script, pkg):
    return subprocess.run([sys.executable, str(pkg / "scripts" / script)],
                          capture_output=True, text=True)


def main():
    failures = []
    for label, (mutate, expect) in CASES.items():
        with tempfile.TemporaryDirectory() as tmp:
            pkg = pathlib.Path(tmp) / "ui"
            shutil.copytree(PKG, pkg, ignore=shutil.ignore_patterns("__pycache__"))
            good = (pkg / "src/tokens.css").read_text()
            doc = json.loads(json.dumps(BASE))
            mutate(doc)
            (pkg / "design/tokens.json").write_text(json.dumps(doc, indent=2))

            gen = run("gen-tokens.py", pkg)
            after = (pkg / "src/tokens.css").read_text()
            chk = run("check-tokens.py", pkg)
            out = gen.stderr + chk.stderr

            if gen.returncode == 0:
                failures.append(f"{label}: generator wrote output from an invalid document")
            elif after != good:
                failures.append(f"{label}: generator modified tokens.css before failing")
            if chk.returncode == 0:
                failures.append(f"{label}: checker accepted an invalid document")
            elif expect not in out:
                failures.append(f"{label}: rejected, but not for the expected reason "
                                f"(wanted {expect!r})")
            status = "ok" if not any(label in f for f in failures) else "FAIL"
            print(f"  {status:4}  rejected, css untouched  {label}")

    if failures:
        print()
        for f in failures:
            print("FAIL " + f, file=sys.stderr)
        return 1
    print(f"\n{len(CASES)} invalid documents rejected; none reached tokens.css")
    return 0


if __name__ == "__main__":
    sys.exit(main())
