# Library conventions

How `packages/ui/design/product-ui.lib.pen` is built. These exist because each one
was broken first; they are the rules that stop it happening again.

## One table per component

**Variants are rows. Interaction states are columns.** The `rest` column *is* the
variant showcase, so nothing is drawn twice.

```text
<Component>
├── name
├── contract          what the component guarantees, what it refuses to do
└── matrix            rows = variants · columns = rest · hover · active · focus · disabled
                      the component origin lives in the first row's rest cell

<Component> size and configuration     ← only if these dimensions exist
└── the orthogonal axes: size, leading icon, content shape
```

A separate "variants" row alongside a state matrix is duplication, not
documentation: every cell in it already exists in the matrix's `rest` and
`disabled` columns. It was worth about 17 redundant specimens here, and worse, it
was two places to update — the exact drift the one-block rule is supposed to
prevent.

Keep a dimension out of the matrix only when it is genuinely orthogonal. Size does
not interact with state, so three sizes stay one small row instead of multiplying
the table by three. A leading icon is a configuration every variant can carry, not
a variant of its own.

**Moving a block that contains a component origin deletes the origin with it.**
Origins live inside specimen cells. Move the origin into its new home *first*,
delete the duplicate instance that was standing in for it, and only then delete the
old block.

## The specimen cell

Every specimen is the same two-part structure, so alignment is enforced by the
structure rather than arrived at per row:

```text
Cell (vertical, width fill_container)
├── Slot (fixed height, alignItems end)   ← the mark, whatever its size
└── Caption (width fill_container, textGrowth fixed-width)
```

- The slot's height is fixed so every caption in the row starts at the same offset.
  Marks of different heights are the normal case, not the exception.
- Cells in a row are `fill_container`, so the row divides evenly.
- **The caption's alignment matches the mark's.** Right-aligned amounts get
  right-aligned captions. A left caption under a right value is the single most
  common way a specimen row reads as broken while every element is technically
  correct.
- A mark that cannot share the row's scale gets its own row. A 48 px hero amount
  does not belong in a grid with 13 px ones.

## Text inside a control never wraps

A control's value or label is a single line. `textGrowth: "fixed-width"` on a value
inside a fixed-height control will wrap the moment the cell narrows, and the control
does not grow to match — the second line is simply outside it. Inputs use a fixed
line height plus `clip: true` on the field, which is also what the real control does
(it clips and scrolls). Button labels stay `auto`.

A specimen cell narrow enough to wrap is itself a finding: a 150 px input is not
something anyone ships. Cap a state matrix at three columns per row rather than
squeezing six.

## States that need a matrix

| Component | Columns |
|---|---|
| Button, per variant | rest · hover · active · focus · disabled |
| Input | rest · hover · focus, then error · disabled · read-only — three per row |
| Selection controls, per mark and checked state | rest · hover · focus · disabled |

Focus is always the ring drawn outside the control, additive to whatever fill is
underneath, with radius = the control's radius + the ring's padding. Disabled is
exempt from the contrast minimum but not from being understandable, and it stays
focusable. Read-only is not disabled: the value stays selectable and copyable,
which matters for an account number.

## Text inside a fixed-height box

A text node given an explicit `height` is centred as a *box*, but its glyphs sit at
the top of that box, because `textAlignVertical` defaults to `top`. The result is
text that rides high by a pixel or two inside an otherwise perfectly centred
control — visible, and hard to name.

Set `textAlignVertical: "middle"`. Do not instead shrink the box to the natural
line height: that number is font- and size-dependent, so it silently breaks the
moment a type token changes.

## Expressive treatment

Ornament near a number costs trust. Amounts are flat, crisp and unstyled: no
gradient text, no glass, no count-up, no shimmer once data has arrived. Everything
expressive goes where there is no figure to undermine.

| Treatment | Where it is allowed | Where it is a bug |
|---|---|---|
| Gradient | One hero surface — the balance card. Chart area fills. | On an amount, a button, a table, or behind any dense data. |
| Elevation | Menus, popovers, dialogs, drawers. Always with a border as well, since shadow vanishes on a dark canvas. | Cards in a list. A table. Anything at rest on the page. |
| Motion | Answering a question the user already has: where did this come from, did it save, what is loading, what changed. | Anything animating to be pleasant. Amounts never count up — a number still moving is a number you cannot read. |
| Texture | Empty states, onboarding, the marketing surface. | Behind a ledger. |

The best answer to "make it feel alive" in a money product is that **the data is
the decoration**: sparklines, category ribbons, a budget meter filling. It carries
colour and movement while being the actual content, so it never reads as an
attempt to dress up a screen.

Under `prefers-reduced-motion` every duration collapses to 0 and the shimmer holds
at its mid tone. Nothing is removed and no state becomes unreachable — a drawer
still opens, it simply is open. Anything that communicates only through movement
communicates nothing to the person who turned movement off.

## Build the atom before the molecule

A molecule that needs a control's border, height, focus ring and full state set is
not a molecule — it is the atom with a configuration. Amount field was built as a
separate component and immediately became a second control to keep in sync with
Input for every state change, which is the same duplication the one-table rule
removes. It was deleted and rebuilt as affix cells on the Input atom.

The test: **if the new thing re-draws the atom's chrome, extend the atom.**

## Moving a node inside a component origin voids its overrides

Every instance override keyed to a descendant stops resolving when that descendant
is reparented, and it fails silently — the instances simply fall back to the
origin's defaults, which looks like a rendering glitch rather than an error. Six
inputs reverted to the placeholder text this way.

After restructuring an origin, re-apply the overrides by resolved path
(`Update("instanceId/childId", {...})`) and screenshot every instance. The audit
does not catch this: nothing overflows and nothing is misaligned, the content is
just wrong.

## States and configurations are not orthogonal — draw the crossing

A matrix per axis implies every combination is valid and answers nothing about
what a state does to a configuration. The useful artifact is the crossing itself,
drawn: configurations as rows, states as columns.

Writing prose verdicts instead is the failure mode. "Hover and disabled are
incompatible" is obvious and costs a table row; "what happens to the currency
affix in the error state" is not obvious, and only a drawn cell settles it.

Where a combination genuinely cannot be drawn, one line is enough:

```text
disabled  →  read-only  →  error  →  focus  →  hover  →  rest
```

The leftmost applicable state owns the fill and border. Focus stacks rather than
replaces, which is why a control disabled via `aria-disabled` still shows a ring.
Message rows are independent: an error message renders whenever an error exists,
even under a disabled control.

### What the Input crossing settled

- **Error leaves the chrome alone.** The border and the message carry the error;
  the affix keeps its neutral treatment. A red currency cell claims the currency
  is what is wrong.
- **Disabled dims the chrome too.** Symbol, icon and value drop together. An affix
  that stays crisp beside a dimmed value reads as still-editable.
- **The affix carries no fill and no divider.** The drawn crossing first produced
  a tinted affix cell with a divider, which dark theme then killed: field fill,
  affix fill and canvas are three near-blacks, and in read-only — no outer border —
  the cell vanished. Every fill and divider candidate measured 1.2:1 to 2.6:1, so
  none was worth keeping, while the symbol alone is 7.4:1 dark and 6.5:1 light
  against the field's own fill. The affix is a glyph in a padded cell.

The last one is worth noting as a sequence: the light-theme crossing invented a
rule, and the dark-theme QA overturned it. Neither pass alone was sufficient.

## "Recede" is a different direction in each theme

A disabled control recedes by moving **toward the canvas**. In light that means
getting darker; in dark it means getting *darker still*. Reusing one "subtle"
token for both gets it backwards in one of them.

That is exactly what happened here. `color-bg-subtle` served as the disabled fill:
in light it sits below the white surface and recedes correctly, but in dark it is
**lighter** than the surface, so a disabled field advanced toward the viewer. A
disabled secondary button was worse — `color-action-secondary` and
`color-bg-subtle` resolve to the same value in dark, so rest and disabled were
pixel-identical apart from the label.

Two fixes, and the second is the one that carries it:

- `color-bg-disabled` — always toward the canvas, in both themes.
- `color-border-disabled` — near-invisible by design. In a palette compressed at
  the dark end, fills cannot separate states (every candidate measured 1.1–1.2:1).
  **The edge does the work**: a resting control has a visible border at ~2:1
  against its own fill, a disabled one has none. "Has an edge / has no edge" reads
  at a glance where "slightly different grey" does not.

Generalisation: any token whose name describes an *appearance* ("subtle",
"raised") rather than a *role* ("disabled", "elevated") will eventually be wrong
in one theme. Name tokens for what they are for.

## The origin is what the panel shows

A library is browsed through the Components panel, which renders each origin alone
on its own chrome — no canvas, no parent, no theme context from the document. Two
consequences:

- **Originate the representative variant.** Whichever variant holds the origin is
  the one everyone sees when picking a component. Icon button originated as ghost
  (transparent fill, near-black glyph) and appeared as an empty box; it now
  originates as outlined, with ghost one override away.
- **A container that draws an edge must draw a surface.** Border without fill looks
  like it provides a background and does not, so content contrast depends on
  whatever is underneath. Text-only components are the exception — they inherit
  their surface by design.

## Both themes, or neither

A component verified only in light theme is unverified. Light hides the failures
dark exposes — three near-blacks that read as one, shadows that vanish against the
canvas — and the contrast maths passes throughout, because the tokens are fine and
the *rendering* is not.

Build the light-theme crossing first, then QA it in dark before the component is
called done. Expect the dark pass to overturn decisions the light pass made.

## Validate the checks

A check that has never fired is not evidence of anything. Before trusting a new
one, build the defect it is supposed to catch, confirm it is flagged, and delete
the fixture. The alignment check in `canvas-audit.md` reported a clean pass on
every run for its first three revisions while scanning zero rows.

## Order of work

0. Paste the audit from `canvas-audit.md` verbatim. Retyping it inline drops
   branches — the gradient-aware background lookup got dropped once and produced
   three false contrast failures on the hero card.
1. Mutate in one `execute` call.
2. Run `canvas-audit.md` in the next call — both passes.
3. Fix, return to 2.
4. Only then report.

Reading bounds, `ctx.problems` or a screenshot inside the mutating call returns
pre-layout values. Every defect that reached review in this library reached it
because a check ran in the same call as the change, or ran only after someone
else had already spotted the problem.
