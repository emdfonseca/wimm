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

## No component origin may inherit a theme

An origin inside a themed frame is pinned to that theme. Instances still resolve
from wherever they are placed, so the canvas looks correct and nothing fails —
but the origin itself, and therefore its Components-panel tile, is frozen. It is
invisible until someone wonders why half the panel is one theme.

```js
let bad=0;
ROOTS.forEach(r=>Get(r,(n,c)=>{
  if(!n.reusable)return;
  const chain=[]; let p=c;
  while(p){ if(p.node.theme&&Object.keys(p.node.theme).length)
              chain.push(p.node.name||p.node.id); p=p.parentCtx}
  if(n.theme&&Object.keys(n.theme).length)chain.push("SELF");
  if(chain.length){bad++;Print("PINNED",n.name,chain.join(" <- "))}}));
Print("origins pinned to a theme:",bad);
```

Zone frames must not carry a theme. Only frames whose *purpose* is a fixed context
may: the QA light/dark pairs, the token swatch chips, the device and density
demos, and the regime templates in `50 · TEMPLATES`, each of which is named for
the regime it renders. Clear one with `Update(id,{theme:{}})`.

A template named for a regime and *not* carrying that regime's theme is the worse
failure, and it is silent: both Wide shells here resolved `device = compact` for
months, drawing a 16 px gutter and a 24 px page title under the name "Wide 1440".
Assert the resolved value, never the frame's width:

```js
Print(Get(BODY_ID,{depth:0,resolveVariables:true}).padding);
```

A theme can appear on a frame without being authored in a script — selecting a
frame and using the app's theme switcher writes one. Re-run this after any session
of hand-editing on the canvas.

## Order

Nothing enforces order, so everything drifts into build order — the sequence
things happened to be made in, which is meaningless to a reader. Check it
explicitly; it is the cheapest thing here to get wrong and the most visible.

**Zones** are numbered, so this is mechanical:

```js
const z=[]; Get((n,c)=>{if(c.depth!==0)return;
  z.push({name:n.name||"",x:c.bounds.x}); c.skipChildren()});
z.sort((a,b)=>a.x-b.x);
const nums=z.map(v=>parseInt((v.name.match(/^(\d+)/)||[0,999])[1],10));
const sorted=nums.every((v,i)=>i===0||nums[i-1]<=v);
z.forEach(v=>Print(v.name)); Print("zones in numeric order:",sorted);
```

Reflow with `ids.sort()` by that prefix, then lay out left to right.

**Blocks inside a zone** have no numeric key, so the order is a judgement and
has to be re-stated deliberately after any move:

```js
const order=(zone,ids)=>{ids.forEach((id,i)=>Move(id,zone,i))};
```

Group by family, not by when it was built — inputs together, navigation
together, feedback together, domain components together. A component's
configuration and state blocks sit immediately after it.

Sorting also surfaces duplicates that are invisible in build order: putting
`Layout and motion` next to a later `Motion` section made it obvious that motion
had two homes, which is how a token ends up with two values.

## Stray nodes at the document root

Dragging a component out of the Components panel drops the instance at the
**document root**, not into whatever zone is under the pointer. Seven accumulated
here unnoticed. While zones were stacked vertically the strays sat in the gaps and
looked deliberate; laying the zones out in a row put them straight on top of two
of them.

The root holds zone frames and nothing else:

```js
const boxes=[];
Get((n,c)=>{if(c.depth!==0)return;
  boxes.push({id:n.id,name:n.name||"(unnamed)",type:n.type});c.skipChildren()});
let stray=0;
boxes.forEach(b=>{ if(b.type!=="frame"||!/^\d\d · /.test(b.name)){
  stray++; Print("STRAY",b.id,b.name,b.type)}});
Print("stray root nodes:",stray);
```

Before deleting one, check it is a plain instance — no own overrides, no
descendant overrides. A stray carrying overrides is somebody's work in progress,
not litter:

```js
const n=Get(id,{depth:0});
const own=Object.keys(n).filter(k=>!["id","type","ref","name","x","y","width","height"].includes(k));
Print(id,"own:",own,"descendants:",Object.keys(n.descendants||{}));
```

## Repeated-row height

A size bound to a variable is discarded and the property left absent, so no scan
of the source finds it and `resolveVariables` cannot see it either. Check the
rendered result against intent instead, wherever something repeats:

```js
const rows=Get(ROWS_CONTAINER,(k,kc)=>kc.depth===1?Math.round(kc.bounds.height):undefined);
if(new Set(rows).size!==1||rows[0]!==EXPECTED)
  Print("row height",JSON.stringify(rows),"expected",EXPECTED);
```

Rows hugging to 28 px instead of 56 read as a slightly tight table, not a defect.

## Text contrast

The table above promises this check and the file shipped without the code for it,
which meant it was never run document-wide. Every text node is measured against
the nearest filled ancestor, resolved through whatever theme the node sits in.

```js
const hx=c=>{if(!c)return null;if(typeof c==="string")return /^#/.test(c)?c:null;
  if(Array.isArray(c))return hx(c[0]);if(c.type==="color")return hx(c.color);return null};
const rgb=h=>{let s=h.slice(1);if(s.length===3)s=s.split("").map(x=>x+x).join("");
  return[parseInt(s.slice(0,2),16),parseInt(s.slice(2,4),16),parseInt(s.slice(4,6),16)]};
const lum=h=>{const[r,g,b]=rgb(h);const f=v=>{v/=255;
  return v<=0.03928?v/12.92:Math.pow((v+0.055)/1.055,2.4)};
  return 0.2126*f(r)+0.7152*f(g)+0.0722*f(b)};
const ratio=(a,b)=>{const l1=lum(a),l2=lum(b);
  return(Math.max(l1,l2)+0.05)/(Math.min(l1,l2)+0.05)};
let seen=0,fail=0;
ROOTS.forEach(R=>Get(R,(n,c)=>{
  if(n.type!=="text"||n.enabled===false)return;
  const fg=hx(n.fill); if(!fg)return;
  let p=c.parentCtx,bg=null;
  while(p){const b=hx(p.node.fill); if(b&&!/00$/.test(b)){bg=b;break} p=p.parentCtx}
  if(!bg)return;
  seen++;
  const sz=n.fontSize||14, bold=/600|700|bold/.test(String(n.fontWeight||""));
  const need=(sz>=24||(sz>=18.66&&bold))?3:4.5;
  const r=ratio(fg,bg);
  if(r<need){fail++;Print("CONTRAST",n.id,n.name,r.toFixed(2),"<",need,fg,"on",bg)}
},{resolveVariables:true,resolveInstances:true}));
Print("measured:",seen,"failures:",fail);
```

`color-text-disabled` on a disabled specimen is the expected failure and is exempt
under SC 1.4.3. Everything else is a defect. Run it with `resolveInstances: true`
or it measures the origins and skips every instance override.

## `ctx.problems` reports nodes that fit

`partially clipped` fires on nodes whose bounds lie entirely inside their parent —
eight of them here, in three different zones. Chasing them wastes a session.
Before treating one as real, measure the overflow on each side:

```js
const b=c.bounds,p=c.parentCtx.bounds;
Print(n.name,JSON.stringify({l:-b.x,t:-b.y,r:b.x+b.width-p.width,bo:b.y+b.height-p.height}));
```

All four numbers at or below zero means the node fits and the flag is noise. Only a
positive number on some side is a finding.

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
boxes.sort((a,b)=>a.x-b.x);
boxes.forEach(b=>Print(b.name,"| x",Math.round(b.x),"→",Math.round(b.x+b.w)));
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
creating one.

**Zones sit side by side, not stacked.** A zone grows downward as content is
added, so stacking them vertically means every addition pushes the rest of the
library down and the canvas becomes a ribbon — 1200 × 20,000 at one point here,
a 1:17 aspect ratio that makes zoom-to-fit useless. In a row the same content is
8160 × 6317, close enough to square to see at once, and a zone growing taller
disturbs nothing beside it.

Reflow with a fixed 160 px gutter, all zones top-aligned at y = 0:

```js
const ORDER=[/* zone ids, in numeric order */];
const w={},h={};
Get((n,c)=>{if(c.depth!==0)return;w[n.id]=c.bounds.width;h[n.id]=c.bounds.height;c.skipChildren()});
let x=0; for(const id of ORDER){Update(id,{x,y:0}); x+=w[id]+160}
```

## Order of work

1. One `execute` call that mutates.
2. One `execute` call that runs both passes.
3. Fix, then go back to step 2.

Screenshots belong in step 2 as well, for the same reason: taken in the mutating
call they capture the pre-layout frame, which is how blank white specimen cards
and phantom clipping warnings both got through.
