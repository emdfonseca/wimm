# Canvas audit

A verification pass for `packages/ui/design/product-ui.lib.pen`.

**Run it as its own `execute` call, never appended to the call that made the
change.** Layout has not settled while a mutating call is still running: bounds
read in that same call are stale, `ctx.problems` reports phantom clipping, and
`TakeScreenshot` returns the pre-layout frame. Every false alarm and every missed
defect in this library so far came from reading results in the mutating call.

## What it catches

| Check | Why it exists |
|---|---|
| Overflow and clipping | Flex layout has no wrap. A row wider than its parent silently spills instead of wrapping, and the overflow is invisible in a zoomed-out screenshot. |
| Dropped size bindings | This pen build discards `width`/`height` when given a `$variable`, with no warning. The frame falls back to hugging its content, so a 40 px control renders at 17 px. |
| Hugging controls | Catches the same failure from the other side: any control-shaped node shorter than the 24 px AA target floor. |
| Nested radius | A focus ring's radius must equal the control's radius plus the ring's padding, or the corners visibly disagree. |
| Text contrast | Every text node against its nearest filled ancestor, at the WCAG 2.2 AA threshold for its size and weight. Resolves through whatever theme the node actually sits in. |
| Text sizing | `width` set without `textGrowth` is ignored, so the text silently refuses to wrap. |
| Sibling alignment | Cells in a row that stack a mark over a caption drift apart when the marks differ in height. The captions stop sharing a baseline and the row reads as broken even though nothing overflows. |
| Invalid colour values | `SetVariables` accepts a malformed hex without complaint, and the token then resolves to nothing at render time. Validate every colour value against `^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`. |
| Root overlap | Zone frames hug their content, so adding a section to one grows it downward into the next. `FindEmptySpace` only knows the sizes at the moment it runs. |

Sub-pixel overflow (under 0.5 px) is ignored: `fill_container` siblings divide odd
widths into repeating decimals and would otherwise report every split row.
All-lowercase kebab names are skipped by the control-height check: those are token
documentation rows such as `control-target-min`, not controls.

## The pass

The sibling-alignment check compares, for every horizontal row, the `y` of the
caption in each specimen cell. More than 1 px of spread is a finding. The fix is a
fixed-height slot around the mark so every caption starts at the same offset.

**Two traps that made an earlier version of this check silently pass everything:**

1. **`layout` is omitted when it holds its default.** A default frame is horizontal
   and simply has no `layout` key, so `n.layout === "horizontal"` matches nothing.
   Test `n.layout === undefined || n.layout === "horizontal"`. Before this fix the
   check found 0 candidate rows out of 397 and reported a clean pass every time.
2. **`Get` nested inside a `Get` visitor returns nothing.** Collect candidate ids
   in the visitor, then read their children in a second loop after the traversal
   has finished.

Scope it to the specimen-cell shape — a vertical cell with exactly two children
whose second is text. Without that, any two-column prose layout is flagged for the
crime of having columns of unequal length.

Validate the check itself by building a deliberately broken row, confirming it is
flagged, and deleting it. A check that cannot fail is indistinguishable from a
check that passes.

```js
const isRow=n=>n.type==="frame"&&(n.layout===undefined||n.layout==="horizontal");
const kidsOf=id=>Get(id,(k,kc)=>kc.depth===1?{n:k,b:kc.bounds}:undefined);
const rows=[];
ROOTS.forEach(root=>Get(root,n=>{if(isRow(n))rows.push({id:n.id,name:n.name})}));
for(const r of rows){
  const kids=kidsOf(r.id); if(kids.length<2) continue;
  const st=kids.filter(k=>{
    if(k.n.type!=="frame"||k.n.layout!=="vertical")return false;
    const ch=kidsOf(k.n.id); return ch.length===2&&ch[1].n.type==="text"});
  if(st.length<2) continue;
  const l=st.map(s=>{const ch=kidsOf(s.n.id); return s.b.y+ch[1].b.y});
  const sp=Math.max(...l)-Math.min(...l);
  if(sp>1)Print("FLAGGED",r.id,r.name,"caption spread",Math.round(sp)+"px")}
```

Define `ROOTS` as the zone frames to check, then read the output. Silence is the
pass condition. Fix what it prints, then run it again — the fix call cannot
report on itself.

## Target size

Run against a frame themed `device: "compact"` — that is where the floor binds.

```js
let fails=0;
Get(COMPACT_ROOT,(n,c)=>{
  if(n.type!=="ref"&&!/row target/i.test(n.name||""))return;
  const w=Math.round(c.bounds.width),h=Math.round(c.bounds.height);
  if(Math.min(w,h)>=24)return;
  let p=c.parentCtx,covered=false;
  while(p){ if(Math.min(p.bounds.width,p.bounds.height)>=24 &&
                /row|target|button|item/i.test(p.node.name||"")){covered=true;break}
            p=p.parentCtx}
  if(!covered){fails++;Print("UNDER 24",n.id,n.name,w+"×"+h)}});
Print("uncovered targets:",fails);
```

A mark smaller than 24 px is only a defect when **no interactive ancestor covers
it**. A 20 px checkbox inside a 44 px row passes, and that is the intended design —
the mark is never the target. Without the ancestry walk this check reports every
checkbox, radio and switch in the library as a failure and gets ignored.

SC 2.5.8 AA is 24 × 24 or a qualifying spacing exception. 44 px is a stronger
design choice, not the AA minimum; do not report it as such.

## Component legibility on an unknown background

A component is dropped onto surfaces the library does not control — including the
Components panel, which renders each origin with **no page behind it**. Anything
that borrows the page for contrast is illegible there, and the panel is how people
browse a library.

The dangerous shape is a container that draws an **edge but no surface**: it looks
like it provides a background and does not, so its content contrast is whatever
happens to be underneath.

```js
const hex=c=>!c?null:typeof c==="string"?c:Array.isArray(c)?hex(c[0]):(c.color||null);
const origins=[]; ROOTS.forEach(r=>Get(r,n=>{if(n.reusable)origins.push(n.id)}));
let fails=0;
for(const id of origins){
  const n=Get(id,{depth:0,resolveVariables:true});
  const f=hex(n.fill), s=hex(n.stroke);
  const noFill=!f||/00$/.test(f), hasEdge=!!s&&!/00$/.test(s);
  if(hasEdge&&noFill){fails++;Print("EDGE WITHOUT SURFACE",n.name)}}
Print("failures:",fails);
```

Text-only components (an amount, a form field) legitimately have no surface and
inherit the one they are placed on — they are exempt, not broken.

A corollary for choosing origins: **the origin is what the panel shows**, so it
must be the representative variant. Icon button originally originated as its ghost
variant — transparent fill, near-black glyph — and rendered as an empty box in the
panel. It now originates as the outlined variant, with ghost as an override.

## Colour value validity

```js
const v=GetVariables().variables; let bad=0;
for(const [k,d] of Object.entries(v)){
  if(d.type!=="color") continue;
  const vals=Array.isArray(d.value)?d.value.map(x=>x.value):[d.value];
  for(const val of vals)
    if(typeof val==="string"&&val[0]==="#"&&
       !/^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/.test(val)){
      bad++; Print("INVALID HEX",k,JSON.stringify(val))}}
Print("invalid colour values:",bad);
```

The contrast pass silently skips a token it cannot parse, so a malformed hex looks
like a clean run. Run this whenever tokens are written.

## Root overlap

The pass above walks *inside* each root and never compares one root against
another. Zone frames hug their content: adding a section to `10 · FOUNDATIONS`
grows it downward into `20 · ATOMS`, and neither frame reports a problem because
neither one contains the other. Run this alongside the pass above.

```js
const boxes=[];
Get((n,c)=>{if(c.depth!==0)return;
  boxes.push({name:n.name,x:c.bounds.x,y:c.bounds.y,w:c.bounds.width,h:c.bounds.height});
  c.skipChildren()});
boxes.sort((a,b)=>a.y-b.y);
boxes.forEach(b=>Print(b.name,"| y",Math.round(b.y),"→",Math.round(b.y+b.h)));
let o=0;
for(let i=0;i<boxes.length;i++)for(let j=i+1;j<boxes.length;j++){
  const a=boxes[i],b=boxes[j];
  const ox=Math.min(a.x+a.w,b.x+b.w)-Math.max(a.x,b.x);
  const oy=Math.min(a.y+a.h,b.y+b.h)-Math.max(a.y,b.y);
  if(ox>0.5&&oy>0.5){o++;Print("OVERLAP",a.name,"×",b.name,"|",Math.round(oy),"px")}}
Print("overlaps:",o);
```

`Get(document, {depth:1})` is refused — a visitor-less whole-document read would
dump everything. Use the visitor with `c.depth!==0` and `skipChildren()` as above.

Re-run it after any call that adds content to an existing zone, not just after
creating one. Keep at least 120 px between zones so a later addition has room.

## Order of work

1. One `execute` call that mutates.
2. One `execute` call that runs both passes.
3. Fix, then go back to step 2.

Screenshots belong in step 2 as well, for the same reason: taken in the mutating
call they capture the pre-layout frame, which is how blank white specimen cards
and phantom clipping warnings both got through.
