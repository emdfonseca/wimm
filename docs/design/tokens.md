# Token contract

```text
design/product-ui.lib.pen   the source of truth for values
design/tokens.json          an export of it — the contract CI can read
src/tokens.css              generated from tokens.json, never hand-edited
```

## Why there is an export in the middle

CI cannot read a `.pen` file: it is encrypted and only reachable through the
pencil MCP, which needs the desktop app running. So the library cannot be the
thing `just check` compares against.

`tokens.json` is that stand-in. Drift between **tokens.json and tokens.css** is
mechanical and fully checked. Drift between **the library and tokens.json** is
caught by a person re-exporting, and by nothing else. Treat re-export as part of
any change to variables, in the same commit.

## Re-exporting after changing variables

In an `execute` call against the library:

```js
const g=GetVariables();
Print("THEMES",JSON.stringify(g.themes));
const rows=[];
for(const [k,d] of Object.entries(g.variables)){
  const out={n:k,t:d.type};
  if(Array.isArray(d.value)){for(const e of d.value){
    const key=e.theme?Object.entries(e.theme).map(([a,b])=>a+"="+b).join(","):"_";
    out[key]=e.value}} else out._=d.value;
  rows.push(out)}
rows.sort((a,b)=>a.n<b.n?-1:1);
for(const r of rows)Print(JSON.stringify(r));
```

Fold the output into `design/tokens.json` — keeping each token's explicit `type`
and `unit` — then `just gen packages/ui`. Bare `just gen` regenerates everything,
including the decision index.

Unconditional fallback entries (`"_"`) exist so a value list always has a match
before its conditional entries — file-level tooling takes the **last** matching
entry, so a fallback goes first. They are dropped on export; only the axis values
reach the JSON.

## What `just check packages/ui` enforces

**The export is a valid token document**

- Names are kebab-case; types are known; units are legal for their type and
  explicit on every token — nothing downstream infers a unit.
- Values are well-formed for their type: a colour is a real hex, a duration is a
  non-negative number, a cubic-bezier parses.
- Only declared axes appear, every branch of each axis is present, and no token
  carries both a default and axis values.

**The stylesheet agrees with the export**

- Regenerating `tokens.css` is a no-op.
- Every exported token appears in the stylesheet, and nothing else declares a
  custom property.
- No raw hex anywhere in `src/` except `tokens.css`.

The validator is tested against inputs that must fail — a malformed hex, an
unknown type, an undeclared axis, a missing theme branch, a non-kebab name. A
check that has never rejected anything is not evidence.

## How the axes reach the browser

| Axis | On the canvas | In CSS |
|---|---|---|
| `color` | set per frame | `prefers-color-scheme`, overridden by `[data-theme]` in both directions |
| `device` | set per frame, manually | **automatic** — `min-width` at 768 / 1200 / 1800, mobile-first from compact |
| `density` | set per frame | `[data-density="compact"]` on any subtree |

`device` is the one that behaves differently in the two places. On the canvas it
is a label with no relationship to frame width — a 390 px frame does not become
compact by being 390 px wide. In CSS it resolves from the viewport. That is the
intended asymmetry: the canvas shows chosen compositions, the browser adapts.

`density = compact` reverts to comfortable under `@media (any-pointer: coarse)`,
not below a width. Touch is an input capability: a large tablet is a wide viewport
with coarse input and a narrow desktop window is the reverse, so keying the floor
to a breakpoint gets both backwards. Layout regimes stay width-driven — that is
what they describe — and the target floor does not. The control that sets density
is hidden where it does not apply, rather than shown disabled.
