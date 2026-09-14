## 10. Building a library

How components are organised and verified. Every rule here replaced a round of
rework; none of them are stylistic.

**In this file:**

- 10.1. One table per component
- 10.2. The specimen cell
- 10.3. Choosing the origin
- 10.4. Presets, not copies
- 10.5. Both themes, or neither
- 10.6. Judging colour differences
- 10.7. Verify, then report


### 10.1. One table per component

**Variants are rows. Interaction states are columns.** The `rest` column *is* the
variant showcase.

```text
<Component>
├── contract        what it guarantees, and what it refuses
└── matrix          rows = variants · columns = rest · hover · active · focus · disabled
                    the origin lives in the first row's rest cell

<Component> size and configuration    ← only if such axes exist
```

A separate "variants" row beside a state matrix is duplication, not documentation:
every cell already exists in the matrix's `rest` and `disabled` columns. It is also
two places to update, which is the drift the single-block rule exists to prevent.

Keep a dimension out of the matrix only when it is genuinely orthogonal. Size does
not interact with state, so three sizes are one small row rather than tripling the
table. A leading icon is a configuration every variant can carry, not a variant.

**Draw the crossing rather than describing it.** Prose verdicts gravitate to the
obvious ("a disabled control has no hover") and skip what is actually undecided —
what a state does to an affix, an icon, a badge. Only a drawn cell settles that.
Reserve prose for combinations that genuinely cannot be drawn, plus a precedence
line for when several apply at once:

```text
disabled  →  read-only  →  error  →  focus  →  hover  →  rest
```

The leftmost applicable state owns the fill and border. Focus stacks rather than
replaces, so a control disabled via `aria-disabled` still shows its ring.

### 10.2. The specimen cell

Alignment should be structural, not achieved per row:

```text
Cell (vertical, fill_container)
├── Slot (fixed height, alignItems end)   ← the mark, whatever its size
└── Caption (fill_container, fixed-width text growth)
```

- The slot's height is fixed so every caption in a row starts at the same offset.
  Marks of differing heights are the normal case.
- **The caption's alignment matches the mark's.** A left caption under a
  right-aligned value is the commonest way a row reads as broken while every
  element is individually correct.
- A mark that cannot share the row's scale gets its own row.
- A cell narrow enough to wrap its content is itself a finding. Cap a state matrix
  at three or four columns rather than squeezing six.

### 10.3. Choosing the origin

A library is browsed through the Components panel, which renders each origin alone
on its own chrome — no canvas, no parent, no theme.

- **Originate the representative variant.** Whichever variant holds the origin is
  what everyone sees when picking the component.
- **A container that draws an edge must draw a surface.** Border without fill looks
  like it provides a background and does not, so content contrast becomes whatever
  happens to be underneath. A ghost variant as origin renders as an empty box.
- **And the inverse: a surface must draw its own edge.** A fill that reads as a
  card against the editor's dark chrome disappears against its light chrome, where
  the fill and the background are the same value. The editor's appearance is not
  part of the design and will not be there later. Check the containers in both
  editor themes, not only the components inside them.
- Text-only components legitimately have no surface and inherit the one they are
  placed on. They are the exception, not a defect.

### 10.4. Presets, not copies

There are no variant properties: a component is one object tree and all variation
is per-instance overrides. So a configuration people should be able to pick from
the panel **has to be a component**.

Build it as a thin preset — a component containing an *instance* of the atom with
overrides baked in. It never redraws the control, so there is still one border, one
height and one focus ring to maintain.

**The test: if the new thing redraws an atom's chrome, extend the atom instead.**
A "molecule" that reproduces a control's border, height, focus ring and full state
set is that control with a configuration, and it will drift.

Name presets so they sort beside the atom they wrap.

### 10.5. Both themes, or neither

A component verified only in the default theme is unverified. The opposite theme
exposes failures the first cannot: several near-identical values reading as one,
shadows vanishing against the canvas, a "recessed" tint that recedes in one theme
and advances in the other. Contrast maths passes throughout, because the tokens are
fine and the *rendering* is not.

Build the crossing in one theme, then QA it in the other before calling the
component done, and expect the second pass to overturn decisions the first made.

A QA zone of themed pairs doubles as the fixture the contrast audit needs:
variables resolve per frame theme, so the non-default values can only be measured
where a frame actually renders them.

### 10.6. Judging colour differences

Contrast ratios measure **text legibility**. They are the wrong instrument for two
large adjacent surfaces, which separate perfectly at ratios far below any text
threshold — zebra rows work at about 1.05:1. Rejecting a surface tint because it
measures 1.2:1 is a misreading.

What matters for surfaces is the **direction** of the step, not its size. "Recede"
means moving toward the canvas, which is darker in a light theme and darker still
in a dark one. A single token named for an appearance ("subtle", "raised") will be
correct in one theme and backwards in the other.

**Name tokens for their role, not their look.** `disabled`, `affix` and `elevated`
survive a theme flip; `subtle` does not.

### 10.7. Re-sort after every move

Canvas order defaults to the order things were built in, which carries no meaning
for a reader and drifts further with every addition. Zones, the blocks inside
them, and the sections inside those all need their order re-stated deliberately —
moving one block is enough to leave a zone incoherent.

Zones carrying a numeric prefix can be checked mechanically: read them left to
right and assert the prefixes ascend. Blocks have no key, so group them by family
and keep a component's configuration and state blocks immediately after it.

The reason to bother is not tidiness. Sorting puts related things side by side,
and adjacency is what exposes duplication: two sections covering the same idea are
invisible when separated by ten unrelated blocks and obvious the moment they land
next to each other.

### 10.8. Measure composition

A library is only a system if the larger pieces are built from the smaller ones.
Drawn-by-hand components look right on the day and drift apart by the week: three
copies of one row ended up at 36, 38 and 34 px here, with three different
treatments of "current", before anyone noticed.

Count it rather than eyeballing it:

```js
const names={}; ROOTS.forEach(r=>Get(r,n=>{if(n.reusable)names[n.id]=n.name}));
Get(ZONE,n=>{ if(!n.reusable)return;
  let nodes=0,refs=0,used={};
  Get(n.id,x=>{nodes++; if(x.type==="ref"){refs++; used[names[x.ref]||x.ref]=1}});
  Print(n.name, nodes, refs, Math.round(refs/nodes*100)+"%", Object.keys(used).join(", "))});
```

Read the result by layer:

- **A leaf molecule at 0% is correct** — it is made of primitives, and that is what
  a leaf is.
- **A large component at 0% is a defect.** It is redrawing things the library
  already has.
- **The reuse list is the interesting column.** A component that reuses nothing
  while a near-identical component exists is a missing shared molecule, not a
  coincidence.

Three questions the numbers answer that intuition does not:

1. **Is this a molecule wearing an organism's name?** A small composition with one
   purpose is a molecule wherever it was filed.
2. **Do several components share a shape?** Four surfaces differing only in
   placement and tone are one molecule with four presets.
3. **Is the same row implemented more than once?** Menu rows, sidebar rows and
   sub-nav rows are one component. So are a list row and a table row.

Run it whenever a zone is finished, and before claiming a layer is done.

### 10.9. A contract that states a number needs a measurement

Component contracts accumulate claims: *clears 3:1 unaided*, *never reflows*,
*one level of depth*, *the row is the target*. Prose like this is read as
verified. Much of it is not, and a false claim in a contract is worse than no
claim — it tells the next person the check has already been done.

Any sentence containing a number, or an absolute like *never* or *always*, is a
testable assertion. Measure it when you write it, and measure it again whenever
the thing it describes changes. Three examples from one library, all shipped and
all wrong: a border contract asserting 3:1 while measuring 1.99, a field
promising no reflow while varying 14 px between states, and a token described as
recessed that advanced in the opposite theme.

If a claim cannot be measured, it is decoration — cut it or rewrite it as
something that can be.

### 10.10. A decision recorded twice will disagree

The same rule written on the canvas, in a conventions document, and in a decision
record drifts the moment one is revised. In one library this produced four
contradictions: an affix treatment, a density rule, a decision sequence, and a
reflow promise — each correct in one place and stale in another, with nothing
failing.

Keep the reasoning in one place and have the others point at it. When a decision
changes, search for its terms before claiming it is done — and the canvas needs
searching too, which nothing in the editor does for you:

```js
ROOTS.forEach(r=>Get(r,n=>{
  if(n.type!=="text")return;
  const t=Get(n.id,{depth:0}).content;
  if(typeof t==="string"&&/\b(TERM|OLD_TOKEN|OLD_RULE)\b/i.test(t))
    Print(n.id,n.name,JSON.stringify(t.slice(0,120)))}));
```

Run it for the old token name, the old number, and the phrase that stated the old
rule. Four contradictions survived in one library because the canvas was revised
and the documents were not, or the reverse — each correct somewhere, and nothing
failing anywhere.

### 10.11. Decision sequences run specific to general

A decision table people read top to bottom is an ordered sequence whether or not
it was designed as one. General questions swallow specific ones: *is this a
persisting condition?* is true of invalid input, of a per-row status, and of a
system-wide outage, so asking it early captures all three and the specific
answers are never reached.

Order from most specific to most general, and state in each row why it sits where
it does — that reasoning is what stops the next person reordering it.

Then walk the sequence backwards looking for cases that reach the end and match
nothing. A gap between *needs no action* and *affects more than one row* leaves
single items needing a decision with nowhere to go, and nothing in the table says
so.

### 10.12. A container is not a target

"The whole row is the target", "the card is clickable", "the tile opens the
record" — convenient for a pointer, and impossible for everything else the
moment the container holds a control of its own. A control inside a control is
invalid markup and unreachable by keyboard, so the container cannot be the
button.

The resolution is always the same shape:

- **One named control inside is the primary action** — a link or button with an
  accessible name, in the tab order. That is the real route in.
- **The container's click handler is pointer convenience**, doing the same thing,
  and never the only way to reach it.
- **Every other interactive thing inside keeps its own hit area**, its own label,
  and its own place in the tab order.

The correction this forces is easy to miss: a small mark inside such a container
is *not* covered by the container's target size. A 20 px checkbox in a 44 px row
is only acceptable if the row is the checkbox's target — and it is not, it is the
primary control's. The mark needs its own padded hit area, and the spacing
exception then governs whether the several targets in one row interfere.

**Check every interaction claim against each input modality.** A sentence that is
true for a mouse and false for a keyboard is not a partial truth; it is a defect
with a comfortable half.

### 10.13. Reserving space needs an overflow rule

Reserving height to prevent reflow — a message row, a caption, a status line —
is only half a decision. The other half is what happens when content exceeds the
reservation, and leaving it unstated just relocates the bug.

There are two honest answers and one dishonest one. **Grow** and accept the
reflow in the exceptional case. **Constrain the content** so it cannot exceed the
reservation, and check that. **Clip** is the dishonest one, and for an error
message it is a conformance failure: an error that cannot be read has not been
identified.

Say which one applies, in the contract, next to the reservation.

### 10.14. Findings cluster around one wrong belief

A review returns a list, and the list invites fixing items. Read it for the
premise underneath instead — several findings usually share one, and patching
them individually leaves the belief intact to generate more.

One belief, *"the whole row is a single target"*, produced three separate
findings across two review rounds: a checkbox exempted from target sizing that
was not exempt, a rule forbidding row-level alerts that contradicted a worked
example, and unresolved competition between selecting, opening and acting on a
row. Fixed one at a time they would have stayed inconsistent with each other.
Fixed at the premise, all three resolve and so do the ones nobody had noticed
yet.

The tell is a finding that feels like an edge case in something you already
decided. It is usually the decision surfacing.

### 10.15. Verify, then report

Run the audit before showing the work, not after it comes back. Findings a reviewer
has to catch are findings the check should have.

**Validate the check itself.** Build the defect it is supposed to catch, confirm it
fires, then delete the fixture. A check that has never fired is indistinguishable
from a check that passes, and one here reported clean on every run while scanning
zero rows.

Audit coverage has to include the document root. A pass that walks each zone
subtree is structurally blind to anything sitting between them.
