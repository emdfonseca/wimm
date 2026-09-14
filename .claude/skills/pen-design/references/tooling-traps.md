## 9. Tooling traps

Failures of the tool, not of the design. Each one is silent: nothing errors, the
canvas looks plausible, and the defect is found by a person instead of a check.
Read this before the first `execute` call of a session.

**In this file:**

- 9.1. Silent property drops
- 9.2. Reading results
- 9.3. Visitor and query traps
- 9.4. Component and instance traps
- 9.5. Layout traps


### 9.1. Silent property drops

| Trap | What happens |
|---|---|
| `width`/`height` bound to a `$variable` | The property is **discarded**. The frame falls back to hugging its content, so a 40 px control renders at 17 px. Colour, padding, corner radius, stroke width and type size all bind normally — only the two size properties are affected. Use literals and keep the token authoritative in code. |
| Malformed hex in `SetVariables` | Accepted without complaint. Any contrast pass then *skips* the token it cannot parse, so a broken value looks like a clean run. Validate against `^#([0-9a-fA-F]{3}\|[0-9a-fA-F]{6}\|[0-9a-fA-F]{8})$` after every write. |
| Globals between `execute` calls | Assigning without `const`/`let` is documented to persist. It does not, reliably. Capture ids from the response mapping and paste them as literals into the next call. |

### 9.2. Reading results

**Anything read in the call that made the change is stale.** Bounds, `ctx.problems`
and `TakeScreenshot` all return the pre-layout frame. This produces both phantom
failures and, worse, clean passes over broken layout.

```text
1. execute  — mutate
2. execute  — verify (audit, bounds, screenshots)
3. fix, then back to 2
```

A screenshot taken in step 1 will show blank cards and missing rows that are
present and correct. Do not chase them.

### 9.3. Visitor and query traps

| Trap | What happens |
|---|---|
| `n.layout === "horizontal"` | Matches **nothing**. Default property values are omitted from serialized nodes, and horizontal is the frame default. Test `n.layout === undefined \|\| n.layout === "horizontal"`. A check written the obvious way scanned 0 of 397 rows and reported a clean pass on every run. |
| `Get` nested inside a `Get` visitor | Returns nothing. Collect candidate ids in the visitor, then read them in a second loop after the traversal finishes. |
| `Get(document, {depth: 1})` | Refused — a visitor-less whole-document read would dump everything. Use `Get((n,c)=>{if(c.depth!==0)return; …; c.skipChildren()})`. |

### 9.4. Component and instance traps

- **Moving a node inside a component origin voids every instance override keyed to
  it.** Silently: the instances fall back to the origin's defaults, which reads as
  a rendering glitch rather than an error. After restructuring an origin, re-apply
  overrides by resolved path (`Update("instanceId/childId", …)`) and look at every
  instance.
- **Instances do not pick up an origin's size retroactively.** Setting height on an
  origin after its instances exist leaves them hugging. Set it on each instance.
- **Deleting a block deletes any origin inside it.** Origins live in specimen
  cells. Move the origin out first, delete the instance standing in for it, then
  delete the block.
- **Origins inherit ancestor themes.** An origin inside a themed frame is pinned to
  that theme, along with its Components-panel tile. Instances still resolve from
  their own placement, so nothing looks wrong. Keep zone frames theme-free.
- **The app writes themes you did not author.** Selecting a frame and using the
  theme switcher writes a `theme` onto it. Re-check after any hand-editing session.
- **Dragging from the Components panel drops the instance at the document root**,
  not into the frame under the pointer. These accumulate unnoticed. The root should
  hold zone frames and nothing else.

### 9.5. Layout traps

- **A `fill_container` child inside a hugging parent is circular** and collapses
  both to zero. Give the parent a fixed size on that axis, or stop the child from
  filling. Underlines and rules are the usual casualties — use a stroke on the
  parent instead of a full-width child.
- **Text with an explicit `height` top-aligns its glyphs.** The box is centred, the
  text inside it is not, so it rides high by a pixel or two in an otherwise
  perfectly centred control. Set `textAlignVertical: "middle"`. Do not instead
  shrink the box to the natural line height — that number is font- and
  size-dependent and breaks the moment a type token changes.
- **Flex does not wrap.** A row wider than its parent spills silently and is
  invisible in a zoomed-out screenshot. Cap items per row explicitly.
- **Sub-pixel overflow is noise.** `fill_container` siblings divide odd widths into
  repeating decimals. Ignore overflow under 0.5 px or every split row reports.
