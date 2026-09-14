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
