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

Every control instance carries its own hit area, so this measures instances and
exempts nothing. Two earlier versions of this check passed on every run while
scanning nothing useful, for two separate reasons, and both are worth stating.

**`resolveInstances` deletes the `ref` field.** A resolved node is a plain frame
whose id is `instanceId/originChildId`; it has no `ref` to match on. So a check
that filters `n.type === "ref"` and also passes `resolveInstances: true` matches
zero nodes. Without the flag it matches only top-level instances and never
descends into a row, which is where every checkbox in this library actually
lives. The fix is to collect control-instance ids in an unresolved pass, then
match the **last path segment** of each resolved node against that set.

**A large ancestor is not a licence.** The previous check walked up looking for an
ancestor over 24 px named `row`, `target`, `button` or `item`, and exempted the
mark if it found one. That encoded the belief the row-interaction contract
overturned: the row is the *merchant link's* target, not the checkbox's, so a row
being 44 px tall says nothing about whether the checkbox is reachable. It also
meant any container someone named "Row" silently exempted everything inside it.

**A replaced child never appears in the tree.** `Replace("instanceId/childId", …)`
stores the replacement inside the parent instance's `descendants` map, and a map
entry is not a child: an unresolved traversal walks straight past it, so the id
never reaches `isCtrlRef` and the resolved pass has nothing to match. The check
missed 19 control instances this way, including the icon button the Compact page
header swaps in for its labelled action — a control that exists on the canvas, is
measurable, and was invisible to the one check that measures controls. Collect
ids from the `descendants` maps as well as from the tree.

```js
const CONTROL=/^(button|icon button|checkbox|radio|switch|nav item|tab|pagination|segmented control|account trigger)/i;
const ctrl={}; Get(n=>{if(n.reusable&&CONTROL.test(n.name||""))ctrl[n.id]=n.name});
const isCtrlRef={};
const scan=n=>{ if(n.type==="ref"&&ctrl[n.ref])isCtrlRef[n.id]=ctrl[n.ref];
  if(n.descendants)for(const v of Object.values(n.descendants))
    if(v&&typeof v==="object"&&v.id){scan(v); if(v.children)v.children.forEach(scan)}};
ROOTS.forEach(r=>Get(r,n=>scan(n)));
const abs=c=>{let x=0,y=0,p=c; while(p){x+=p.bounds.x;y+=p.bounds.y;p=p.parentCtx} return[x,y]};
const T=[]; let seen=0,fails=0;
ROOTS.forEach(r=>Get(r,(n,c)=>{
  const seg=String(n.id).split("/").pop();
  if(!(isCtrlRef[seg]||(n.ref&&ctrl[n.ref]))||n.enabled===false)return;
  seen++;
  const w=Math.round(c.bounds.width),h=Math.round(c.bounds.height);
  const [x,y]=abs(c);
  T.push({nm:n.name,w,h,cx:x+c.bounds.width/2,cy:y+c.bounds.height/2});
  if(Math.min(w,h)<24){fails++;Print("UNDER 24",n.id,n.name,w+"×"+h)}
},{resolveInstances:true}));
let viol=0,minAll=Infinity;
for(let i=0;i<T.length;i++){ let best=Infinity,bj=-1;
  for(let j=0;j<T.length;j++){ if(i===j)continue;
    const d=Math.hypot(T[i].cx-T[j].cx,T[i].cy-T[j].cy); if(d<best){best=d;bj=j}}
  if(best<minAll)minAll=best;
  if(Math.min(T[i].w,T[i].h)<24&&best<24){viol++;
    Print("SPACING",T[i].nm,"nearest",T[bj].nm,Math.round(best)+"px")}}
Print("measured:",seen,"under 24:",fails,"spacing:",viol,"min centre distance:",Math.round(minAll));
```

It measures 292 control instances: 0 under 24, 0 spacing violations, and the
closest two target centres anywhere on the canvas are 37 px apart. Print `seen`
every run: a sudden drop is the check going blind, and that is the failure mode
this section exists for.

**Validate it before trusting it.** Shrink one instance in a specimen matrix and
one inside a row instance, confirm both are reported, then restore them. The
second is the one that matters — the row-nested case is what every previous
version missed. Both branches have fired here: a row-nested checkbox forced to
16 × 16 reported `UNDER 24`, and two 20 px icon buttons placed 2 px apart —
centres 22 px — reported `SPACING` in both directions. The fixtures were deleted
afterwards, and restoring the checkbox needed an explicit `width: 24, height: 24`
rather than a removed key, for the reason in `library-conventions.md`.

SC 2.5.8 AA is 24 × 24 or a qualifying spacing exception. `control-target-min` is
24 and binds always; `density-row-min-touch` is 44 and binds only where the
pointer is coarse, which is also where compact density is refused. Do not report
44 as the AA minimum.

### The spacing exception

Where several targets share a row, 24 px circles centred on each must not
intersect. The check above measures this across the **whole canvas**, not one
container: it is a distance between two targets, and two targets in different
blocks are as capable of interfering as two in one row.

That needs absolute coordinates. `ctx.bounds` is relative to the parent, so
comparing `y` between two rows reports 0 px of separation and means nothing;
`abs()` sums the chain to the document. An earlier version scoped this to the
table region and took the row pitch from `density-row-height` to work around the
relative bounds — a correct measurement of one container, offered as a statement
about the library.

Read the result as a whole: **no target on this canvas is under 24 px, so the
exception is not relied on anywhere.** That is the strong form of the claim, and
it is what the check reports — not that undersized targets are adequately spaced,
but that there are none. The nearest pair of target centres is 37 px, which also
says no future 20 px mark could be dropped between two existing controls without
the check noticing.

## Reflow at 320 CSS px

SC 1.4.10 AA asks that content reflow to 320 CSS px without horizontal scrolling
and without losing anything. The Compact regime's representative frame is 390,
which is not the floor and hides the failures — both defects found here were
invisible at 390 and obvious at 320. Draw the floor, with the longest realistic
strings, and measure it.

Measure every descendant's left and right edge against its parent. Vertical
overflow is not a finding: the page scrolls that way by design.

```js
let over=0;
Get(FRAME,(n,c)=>{
  if(n.enabled===false)return void c.skipChildren();
  const b=c.bounds,p=c.parentCtx&&c.parentCtx.bounds; if(!p)return;
  const m=Math.max(-b.x, b.x+b.width-p.width);
  if(m>0.5){over++;Print("OVERFLOW",n.id,n.name,"w",Math.round(b.width),
    "parent",Math.round(p.width),"by",m.toFixed(1))}},{resolveInstances:true});
Print("horizontal overflow nodes:",over);
```

**Skip disabled nodes, and skip their subtrees.** A node with `enabled: false` is
not rendered but still reports bounds, so it overflows on paper and not on screen.
Three of them — the tab counts the Compact header switches off — were reported as
overflow here and are not. The same filter belongs on the `ctx.problems` pass
below, for the same reason.

Measured: `Transactions / 320` 77 rendered descendants and `Transactions / 320 ·
editing` 39, both with 0 crossing either edge and a worst overhang of 0.00 px.
There is no dark twin of either: reflow is geometry, the same widths resolve in
both colour themes, and a twin would assert the same measurement twice.

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

Every text node against the nearest filled ancestor, resolved through whatever
theme the node sits in. Six things have gone wrong here, each producing a clean
run over a check measuring the wrong thing. They are listed because every one of
them looked correct when it was written.

**Alpha is not the last two characters.** `/00$/` as a transparency test drops
opaque `#000000` — and any colour whose blue channel ends in `00`. Parse the
length: only `#RRGGBBAA` and `#RGBA` carry alpha.

**A translucent colour must be composited, not measured.** `#00000080` read as
pure black scores 21:1 against white. Over white it is mid-grey and scores 4.00.
Composite every translucent layer onto what is behind it, the foreground included.

**Gradients must be resolved at all.** A gradient fill is not a hex string, so a
naive lookup walks past it to the solid fill above — on the balance card, the page
canvas — inventing three failures at 1.08, 1.25 and 1.40 for white and mint on
deep pine. This regression has shipped twice.

**A gradient is not its endpoints.** Scoring the stops and taking the worst is a
false pass whenever the text's luminance lies *between* them: interpolation is
continuous, so somewhere along the fill contrast passes through exactly 1:1.
`#777777` on white-to-black scores 4.48 at the worst endpoint and 1.00 in the
middle.

**Never truncate the stop set.** Capping it at six to bound the compositing work
silently discarded everything after the sixth — and `slice(0, 6)` keeps the
*first* six, so a gradient ending in white loses precisely the stop that matters.
A seven-stop pine-to-white gradient then reported a comfortable pass for white
text that has a 1:1 region in it. When the paint is too complex to evaluate,
**say so and stop**; do not evaluate part of it and report a number. A number is
read as an answer.

**The foreground must be paired with the background it sits on.** Compositing
translucent text against one arbitrary member of the stop set makes the verdict
depend on the order the stops happen to be listed in: `#ffffff58` on
`#000000`→`#0a0a0a` scored 2.86 forward and 3.00 reversed, which straddles the
large-text threshold. Same paint, different answer. Composite the foreground
against *each* candidate background and score that pair, and sample along a
continuous fill rather than only at its stops, since both the text and the
backdrop vary together across it.

**Discrete candidates are not a sweep.** The interior test is only valid for a
continuous fill. Alternative backgrounds produced by compositing translucent
layers are discrete possibilities that never blend into one another, so applying
the range test to them reports 1:1 between two backgrounds that never touch.
Track whether any contributing layer was a gradient, and only sweep then.

```js
const paint=c=>{if(!c)return{cols:[],cont:false};
  if(typeof c==="string")return{cols:/^#/.test(c)?[c]:[],cont:false};
  if(Array.isArray(c)){const r={cols:[],cont:false};
    c.forEach(x=>{const p=paint(x);r.cols=r.cols.concat(p.cols);r.cont=r.cont||p.cont});return r}
  if(c.enabled===false)return{cols:[],cont:false};
  if(c.type==="color")return paint(c.color);
  if(c.type==="gradient"||c.type==="mesh_gradient"){
    const cols=(c.colors||[]).flatMap(s=>paint(s&&s.color!==undefined?s.color:s).cols);
    return{cols:cols,cont:cols.length>1}}
  return{cols:[],cont:false}};
const parse=h=>{let s=h.slice(1);
  if(s.length===3||s.length===4)s=s.split("").map(x=>x+x).join("");
  const a=s.length===8?parseInt(s.slice(6,8),16)/255:1;
  return[parseInt(s.slice(0,2),16),parseInt(s.slice(2,4),16),parseInt(s.slice(4,6),16),a]};
const over=(f0,b0)=>{const f=parse(f0),b=parse(b0),a=f[3];
  const m=i=>f[i]*a+b[i]*(1-a);
  return"#"+[m(0),m(1),m(2)].map(v=>Math.round(v).toString(16).padStart(2,"0")).join("")};
const lum=h=>{const[r,g,b]=parse(h);const f=v=>{v/=255;
  return v<=0.03928?v/12.92:Math.pow((v+0.055)/1.055,2.4)};
  return 0.2126*f(r)+0.7152*f(g)+0.0722*f(b)};
const ratio=(a,b)=>{const l1=lum(a),l2=lum(b);
  return(Math.max(l1,l2)+0.05)/(Math.min(l1,l2)+0.05)};
const hx=(r,g,b)=>"#"+[r,g,b].map(v=>Math.round(v).toString(16).padStart(2,"0")).join("");
const lerp=(a0,b0,t)=>{const a=parse(a0),b=parse(b0);
  return hx(a[0]+(b[0]-a[0])*t,a[1]+(b[1]-a[1])*t,a[2]+(b[2]-a[2])*t)};
const CANVAS="#F4F7F5", LIMIT=64, N=8;
let seen=0,fail=0,manual=0;
ROOTS.forEach(R=>Get(R,(n,c)=>{
  if(n.type!=="text"||n.enabled===false)return;
  const fp=paint(n.fill), fs=fp.cols.filter(h=>parse(h)[3]>0); if(!fs.length)return;
  const chain=[]; let p=c.parentCtx;
  while(p){const lp=paint(p.node.fill); const cols=lp.cols.filter(h=>parse(h)[3]>0);
    if(cols.length)chain.push({cols:cols,cont:lp.cont}); p=p.parentCtx}
  if(!chain.length)return;
  seen++;
  let set=[CANVAS], cont=false, bail=false;
  for(let i=chain.length-1;i>=0;i--){const L=chain[i];
    cont=cont||L.cont;
    if(L.cols.every(h=>parse(h)[3]===1)){set=L.cols}
    else{ if(L.cols.length*set.length>LIMIT){bail=true;break}
      const out=[];L.cols.forEach(s=>set.forEach(b=>out.push(over(s,b))));set=[...new Set(out)]}}
  if(bail){manual++;Print("MANUAL REVIEW",n.id,n.name,"paint too complex to evaluate");return}
  const raw=fs[0], opaque=parse(raw)[3]===1;
  const at=b=>ratio(opaque?raw:over(raw,b), b);
  const sz=n.fontSize||14, bold=/600|700|bold/.test(String(n.fontWeight||""));
  const need=(sz>=24||(sz>=18.66&&bold))?3:4.5;
  let w;
  if(!cont||set.length<2)w=Math.min(...set.map(at));
  else{ w=Infinity;
    for(let i=0;i<set.length-1;i++)for(let k=0;k<=N;k++)w=Math.min(w,at(lerp(set[i],set[i+1],k/N)));
    if(opaque){const lf=lum(raw),ls=set.map(lum);
      if(lf>=Math.min(...ls)&&lf<=Math.max(...ls))w=1}}
  if(w<need){fail++;Print("CONTRAST",n.id,n.name,w.toFixed(2),"<",need,cont?"[sweep]":"",raw,"on",JSON.stringify(set.slice(0,4)))}
},{resolveVariables:true,resolveInstances:true}));
Print("measured:",seen,"failures:",fail,"manual:",manual);
```

It measures 1985 text nodes across the seven zones: 30 failures, all
`color-text-disabled` on disabled specimens and exempt under SC 1.4.3, and 0
requiring manual review. Print `manual` every run — a non-zero count is paint the
check declined to score, which is a result, not a pass. Print `sweeps` too: it is
3 here, the balance card's labels, and a sudden 0 means gradient resolution has
broken again.

**Run it in two calls.** `execute` is killed at 60 s and the whole document with
`resolveVariables` and `resolveInstances` does not finish inside it. Split the
zones — `00/10/20` then `30/40/50/90` — and add the two totals. Sampling a
continuous fill at `N=8` per stop pair is what keeps each half inside the limit;
33 samples was enough to blow it.

**Validate it on values, not on the document.** The canvas passing proves nothing
about the check, because the canvas is valid. Assert these directly:

| case | correct answer | what it catches |
|---|---|---|
| `#000000` through the alpha filter | kept | six-digit hex read as having alpha |
| `#00000080` over `#FFFFFF` | 4.00, not 21.00 | alpha ignored instead of composited |
| `#777777` on `#FFFFFF`→`#000000` | 1.00, not 4.48 | endpoints scored, interior missed |
| `#FFFFFF` on a 7-stop gradient ending `#FFFFFF` | 1.00, not a pass | stop set truncated |
| `#777777` against two *discrete* candidates | 4.48, not 1.00 | sweep test applied to non-sweep |
| `#ffffff58` on `#000000`→`#0a0a0a`, then reversed | same both ways | foreground paired with one arbitrary stop |
| `#FFFFFF` on `#1B5340`→`#0C2B20` | 8.91 | the balance card still passes |

Run it with `resolveInstances: true` or it measures the origins and skips every
instance override.

## `ctx.problems` reports nodes that fit

`partially clipped` fires on nodes whose bounds lie entirely inside their parent —
eight of them here, in three different zones. It also fires on disabled nodes,
which have bounds and no rendering: the App header's own nav trigger is 44 px in a
36 px title row and is switched off everywhere except Compact, where the row grows
to fit it. Filter `enabled === false` and skip its subtree before measuring.

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
