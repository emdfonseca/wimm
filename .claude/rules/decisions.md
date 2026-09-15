# Decisions

Generated from `docs/decisions/` by `just adr-index`. Never edit; write an ADR.

## 0001 · Stack and repo shape — Accepted

- Go for services and workers under `apps/`, importable code under `packages/`.
- SvelteKit with Svelte 5 for the UI, in `packages/ui`. Python only where a
  library forces it.
- Connect over protobuf between our own callers; REST with OpenAPI 3.1 only for
  third parties.
- Postgres with goose migrations owned by the service that owns the database.
- OpenTelemetry over OTLP for traces, metrics, and logs. The backend stays
  undecided; services never know it.
- devbox owns the toolchain, `just` owns every task entry point, and CI runs
  `devbox run -- just ci` and nothing else.
- Repo standards live as Claude skills in `.claude/skills/`; they reference only
  paths inside this repo.

## 0002 · Design system foundations — Accepted

**Typefaces.** Schibsted Grotesk for interface text, IBM Plex Mono for amounts,
account identifiers and table dates. Both are Google Fonts, self-hosted.
`type-family-display` resolves to the interface family — display is weight 600 at
−2% tracking, not a third font.

**Token naming.** Kebab-case (`color-bg-canvas`), mapping 1:1 to CSS custom
properties (`--color-bg-canvas`). This departs from the dotted vocabulary in the
pen-design standard's token reference; roles and structure are unchanged.

**Palette.** Pine & Signal. Deep pine green `#14402F` is the brand and primary
action; blue `#2563EB` is the highlight for links, selection and focus. Green and
red are reserved for money direction and feedback, never for branding, so an
amount's colour means exactly one thing.

**Theme axes.** Three, independent:

- `color` — light, dark.
- `device` — compact, medium, wide, **ultra**. A fourth regime beyond the
  standard's three, justified by one structural change at ≥1800 CSS px: the
  editing drawer stops being an overlay and becomes a persistent inspector pane.
- `density` — comfortable, compact. A third axis beyond the standard's two,
  justified by row height being the most consequential number in a ledger: at
  1440 × 900, roughly 684 px of table area shows 12 comfortable rows or 17
  compact ones.

## 0003 · Token pipeline provenance — Accepted

`design/tokens.json` is the contract CI owns. It carries, for every token, an
explicit `type` and `unit`; nothing downstream infers either.

`just check packages/ui` validates two separate things:

- **The export is a valid token document** — kebab-case names, known types,
  units legal for their type, values well-formed for their type, only declared
  axes, every branch of each axis present, no unknown fields, and no token
  carrying both a default and axis values.
- **The stylesheet agrees with the export** — regenerating is a no-op, every
  token reaches the CSS, and nothing else declares a custom property.

The validator is itself tested against inputs that must fail: a malformed hex, an
unknown type, an undeclared axis, a missing theme branch, and a non-kebab name.

**Library-to-export agreement is verified by a person, not by CI.** Re-exporting
is part of any change to variables, in the same commit, and the diff is the
review artifact.

## 0004 · Density availability follows pointer capability — Accepted

Layout regimes stay width-driven: the `device` axis describes available space,
which is what a width query measures.

The touch floor does not. `density = compact` reverts to comfortable under
`@media (any-pointer: coarse)`, and the control that sets density is hidden where
it does not apply rather than shown disabled — a greyed control invites someone to
work out how to enable it.

This supersedes the density consequence in ADR 0002. The three-axis decision
itself is unchanged.

## 0005 · Secondary navigation placement and content width — Accepted

**A subsection list is a card inside the page, at every pointer regime.** It sits
under the page header that names the area it belongs to. It is never a
full-height column beside the sidebar: the sections belong to the page, and a
column that starts above the page header claims they are a peer of the primary
nav. Width plays no part in the choice.

The form is chosen by the content and only by the content, which is what the
secondary-nav contract already said before the regime clause was added to it:

```text
a fixed handful of peer views   → tabs, in the page header, at every size
a longer or growing set         → a list card in the page, at every size
```

`layout-subnav-width` is 212 at Medium, Wide and Ultra, and 0 at Compact, where
Settings stacks.

**`layout-content-max` caps text measure only.** Prose, help and long-form
settings copy stay within 1200. Panels, tables and chart regions take the full
Main column. The token exists to protect line length; a form panel and a ledger
have no measure to protect.

## 0006 · A more vivid palette, and separating the two greens — Accepted

Three things move together, because moving any one of them alone is worse than
moving none.

```text
brand    #14402F / #9FE0C0  →  #0C7A57 / #5FE7B8
accent   #2563EB / #6699FF  →  #1F5FF0 / #7AA8FF
income   #067647 / #3DD68C  →  #127A33 / #4ADB6E
```

`#0C7A57` is the lightest the brand can be: white on it measures 5.33, and one
step lighter (`#0E8A62`) drops the button label to 4.35 and fails. Income moves
off the brand's hue rather than its lightness — 139 and 135 against the brand's
161 and 159, a gap of 22° and 24° where it was 2° and 1°. Every threshold holds:
white on brand 5.33, brand on canvas 4.95, accent as text 5.31 and 4.92, income
text 5.44 and 9.64, and the dark equivalents higher.

Sixteen tokens move, because the brand, accent and income families each carry
derivatives — hover and active states, the chart ramp's first two slots, the
focus ring, the info and success feedback pairs, and the hero gradient.

**The fill-only / text-only rule stays.** It is tempting to retire it now that the
greens differ by hue, and that would be wrong: a 22° separation inside the green
band is close to invisible under deuteranopia, which is the condition the rule was
protecting against in the first place. Hue separation helps typical vision and
does nothing for the case that mattered. What actually carries direction is the
sign: every amount shows `+` or `−`, so no reading of the ledger depends on
telling two greens apart.

Two tokens are added rather than adjusted. `color-text-on-brand` and
`color-text-on-brand-secondary` replace two colours that were hard-coded onto the
balance card, and which the new gradient broke — the card was carrying its own
palette and nothing said so until the contrast pass failed on it.

The hero gradient does not follow the brand all the way. `#0E6E4F` rather than
`#12805C`, because at the lighter value no secondary tone clears 4.5 against it:
the lightest candidate reached 4.44. The gradient is still far more saturated
than it was — 78% against 51% — at about the same lightness.

## 0007 · OpenSpec for change proposals — Accepted

OpenSpec holds change proposals and living specs. ADRs keep their role: a
decision record says what was chosen and why, a proposal says what is about to
be built. They are not substitutes.

The CLI is pinned in `bin/openspec`, a shim that execs `pnpm dlx` at an exact
version. devbox prepends `$DEVBOX_PROJECT_ROOT/bin` to `PATH`, so inside the
project shell a bare `openspec` is the pinned one.

The pin has to live on `PATH` rather than only in a just recipe. OpenSpec
generates its own instruction files — six `opsx` commands and six `openspec-*`
skills — and they invoke a bare `openspec` around a hundred times. A recipe
would leave every one of those calls resolving to whatever the machine happens
to have installed globally.

`just openspec <command>` delegates to the same shim, so the version is stated
once. There is no root `package.json`: a single dev CLI does not justify
introducing an npm manifest to a repo whose one Node package has no
dependencies.

Claude Code is the only configured tool integration.

## 0008 · The canvas is a proposal artifact — Accepted

A project schema, `openspec/schemas/wimm/`, adds a fifth artifact between design
and tasks. `openspec/config.yaml` sets `schema: wimm`, so every new change is
stamped with it.

```text
proposal → specs → design → canvas → tasks
                             │
                             └── tasks is blocked until canvas exists
```

**`canvas` draws the journey and records what was drawn.** The drawing goes in a
journey `.pen` under `packages/ui/design/`; `product-ui.lib.pen` keeps holding
reusable mechanics only. `canvas.md` records the journey ID, the frame IDs, the
task surfaces, the library components instanced, the components the library is
missing, the states drawn, and the behavioural contracts the canvas cannot
execute.

**It is conditional.** A change with no user-facing surface — a worker, a
migration, a build script — records a deliberate skip. The condition is stated
in the artifact's own instruction, which is the mechanism OpenSpec already uses
for a conditional artifact.

**Components come before the screen that uses them.** Each entry under Components
missing becomes one task covering both the origin in `product-ui.lib.pen` and the
Svelte component in `packages/ui/src`. The screen composes instances and cannot
share a task with the components it instances.

## 0009 · The canvas saves without a person — Accepted

`bin/pen-save <file.pen>` pipes `save()` into the pinned CLI attached to the
running app, and `just pen-save <file.pen>` is the shape everything uses.

```text
1. execute   — mutate through the MCP
2. execute   — verify in a separate call
3. pen-save  — flush, and confirm git sees the file
```

The MCP stays the way edits are made. It is precise and deterministic: the
alternative the CLI offers is `pen --in … --out … --prompt …`, which runs a
second AI agent against the file and is neither. The CLI is used for the one
thing the MCP cannot do.

**The shim verifies rather than trusts.** It records the file's mtime, runs the
save, and fails if the file did not move — so "Is Pen.app running?" surfaces as
an error rather than as a canvas.md describing frames nobody can see.

**It runs through npm, not pnpm.** `@pen.dev/cli` imports `css-tree` without
declaring it, which resolves under npm's flat layout and fails under pnpm's
strict linking. This is the one place in the repo that reaches for npm, and the
reason is a defect in the package rather than a preference.

This supersedes the save consequence in ADR 0008. The canvas artifact itself is
unchanged.

## 0010 · Journeys import the library — Accepted

A journey file imports the library. `just pen-import <journey.pen> <alias>
<library.pen>` writes the `imports` key and creates the journey file if it does
not exist.

```text
imports     { "ui": "product-ui.lib.pen" }
components  ref: "ui:W2gOKx"
variables   "$ui:color-bg-canvas"
```

The alias qualifies everything. A bare id is a non-existent node, and a slash is
rejected — `ref` may not contain one.

**It writes JSON directly, because nothing else can.** `execute` has no
document-level mutator beyond `SetVariables`, and `Update(document, …)` reports
`Node 'document' not found`. A `.pen` is pretty-printed JSON, so the key is three
lines; the command exists to make the edit checked and idempotent rather than
hand-made. It refuses an alias already bound elsewhere, a library outside the
journey's directory, and a non-kebab alias.

## 0011 · Draw headless, not through the MCP — Accepted

Journeys are drawn headless. `just pen-exec <file.pen>` pipes an execute snippet
into `pen interactive --in X --out X`, which opens X, resolves its imports, saves,
and exits.

```text
just pen-import   create the journey and bind the library
just pen-exec     draw, and verify in a second call
Export + read     look at the result
```

The MCP keeps one job: the document a person is actually looking at. `just
pen-save` keeps one job with it — flushing that document, which stays in the
app's memory until something saves it.

**`pen-exec` fails closed.** A snippet that errors exits non-zero and leaves the
file byte-identical; both are asserted, because a tool that writes garbage and
then exits non-zero has still corrupted the artifact.

**Headless saves only because it is told to.** `pen interactive` does not save on
exit. A pipe without `save()` reports every id it created and writes none of
them — the failure looks exactly like success.

This supersedes the workflow in ADR 0009. The reason that ADR exists — an edit
sitting unsaved in the app with nothing reporting it — is unchanged and still
applies to MCP work.

## 0012 · Stories carry the user, specs carry the contract — Accepted

A `stories` artifact sits between proposal and specs. `specs` requires it.

```text
proposal → stories → specs → design → canvas → tasks
```

The two levels split by audience, and the split is the point:

```text
story criterion   what a person can do, checked by using the product
spec scenario     normative system behaviour, checked by a test
```

A criterion naming a function, table, endpoint or component is not a criterion —
it is an implementation note, and it belongs in design or tasks.

**Stories name a real user by role.** "As a developer" and "as the user" are the
tell that there is no actor, and without an actor there is no value to state.

**INVEST is recorded, not performed.** Six lines, each answered. A story that
fails one is not automatically wrong: name the letter, say why shipping it anyway
is right. An honest "Small — no, this covers three screens, and splitting them
would ship a half-usable ledger" is worth more than six unconsidered ticks.

**Every requirement traces to a story.** If one does not, either the story is
missing or the requirement is.

**It is conditional**, like design and canvas. Tooling, refactors and internal
migrations record a one-line skip rather than inventing a user.

## 0013 · Acceptance criteria are scenarios — Accepted

The spec's scenarios are the acceptance criteria. There is no second list.

The line between the two artifacts is **lifetime**, not audience:

```text
openspec/specs/<cap>/spec.md   living. Accumulates across changes.
stories.md                     change-scoped. Archives with the change.
```

A requirement is a persistent behaviour contract; a story is a unit of work.
INVEST asks whether the work item is Independent, Estimable, Small — questions
with no meaning about a permanent capability.

So `stories.md` keeps only what a spec cannot hold: the actor, the value, the
INVEST answer, and `Satisfied by`. The user-testable bar moves onto the scenarios
themselves — a scenario for a user-facing capability must be observable by a
person, and one naming a function, table, endpoint or component is an
implementation note.

**The link is a metadata line.** Every requirement carries `**Story**: S<n>`
under its header. `**Key**: value` is the one pattern the parser recognises: it is
excluded from the requirement body when other text is present, and survives
archive verbatim. Free prose would not.
