# Token contract

```text
design/tokens.json   the source of values, edited by hand
src/tokens.css       generated from tokens.json, never hand-edited
```

## Changing a token

Edit `design/tokens.json`, keeping each token's explicit `type` and `unit`, then
`just gen packages/ui`, in the same commit. Bare `just gen` regenerates
everything, including the decision index.

A token with axis values carries every branch of each axis it names, and no
default beside them.

## What `just check packages/ui` enforces

**`tokens.json` is a valid token document**

- Names are kebab-case; types are known; units are legal for their type and
  explicit on every token — nothing downstream infers a unit.
- Values are well-formed for their type: a colour is a real hex, a duration is a
  non-negative number, a cubic-bezier parses.
- Only declared axes appear, every branch of each axis is present, and no token
  carries both a default and axis values.

**Motion respects the preference**

Under `prefers-reduced-motion: reduce` every duration token resolves to `0ms`.
The easing tokens are left as they are: a curve with no time to run is harmless,
and keeping them means a component never has to branch. Nothing is removed and no
state becomes unreachable — a drawer still opens, it simply is open.

**The stylesheet agrees with `tokens.json`**

- Regenerating `tokens.css` is a no-op.
- Every token appears in the stylesheet, and nothing else declares a
  custom property.
- No raw hex anywhere in `src/` except `tokens.css`.

The validator is tested against inputs that must fail — a malformed hex, an
unknown type, an undeclared axis, a missing theme branch, a non-kebab name. A
check that has never rejected anything is not evidence.

## How the axes reach the browser

| Axis | In CSS |
|---|---|
| `color` | `prefers-color-scheme`, overridden by `[data-theme]` in both directions |
| `device` | automatic: `min-width` at 768 / 1200 / 1800, mobile-first from compact |
| `density` | `[data-density="compact"]` on any subtree |

`device` resolves from the viewport. A story is seen at a regime by being given
that regime's width on the design canvas, never by a label.

`density = compact` reverts to comfortable under `@media (any-pointer: coarse)`,
not below a width. Touch is an input capability: a large tablet is a wide viewport
with coarse input and a narrow desktop window is the reverse, so keying the floor
to a breakpoint gets both backwards. Layout regimes stay width-driven — that is
what they describe — and the target floor does not. The control that sets density
is hidden where it does not apply, rather than shown disabled.
