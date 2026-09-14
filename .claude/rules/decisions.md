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
