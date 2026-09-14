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

Fold the output into `design/tokens.json`, then `just gen packages/ui`.

Unconditional fallback entries (`"_"`) exist so a value list always has a match
before its conditional entries — file-level tooling takes the **last** matching
entry, so a fallback goes first. They are dropped on export; only the axis values
reach the JSON.

## What `just check packages/ui` enforces

- Regenerating `tokens.css` from `tokens.json` is a no-op.
- Every exported token appears in the stylesheet.
- Every custom property in the stylesheet is an exported token — nothing
  hand-added.
- Every themed token supplies all values of the axes it uses, so a missing dark
  value cannot ship as a silent fallback to light.
- No raw hex anywhere in `src/` except `tokens.css`.

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

`density = compact` is reverted to comfortable below 768 px in CSS, because a
40 px row is a poor thumb target. The control that sets it is hidden on touch
rather than shown disabled.
