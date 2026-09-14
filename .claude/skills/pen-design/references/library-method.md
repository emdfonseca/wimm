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

### 10.7. Verify, then report

Run the audit before showing the work, not after it comes back. Findings a reviewer
has to catch are findings the check should have.

**Validate the check itself.** Build the defect it is supposed to catch, confirm it
fires, then delete the fixture. A check that has never fired is indistinguishable
from a check that passes, and one here reported clean on every run while scanning
zero rows.

Audit coverage has to include the document root. A pass that walks each zone
subtree is structurally blind to anything sitting between them.
