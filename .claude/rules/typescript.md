---
paths:
  - "**/*.ts"
  - "**/*.svelte"
---

# TypeScript / Svelte conventions

Decisions only; idiomatic Svelte 5 and TypeScript apply as documented upstream.

- **Svelte 5 runes** — `$state`, `$derived`, `$props`, `$effect`, snippets. No legacy `$:` reactivity, no `export let`, no `<slot>`.
- **Strictness** — every `tsconfig.json` extends `tsconfig.base.json`. Loosening `strict` or `noUncheckedIndexedAccess` needs a comment saying why.
- **Imports** — internal packages via `@repo/<pkg>` entry points only. No deep imports into another package's `src/`; the `exports` map is the public surface.
- **Screens vs routes** — presentational screen components take data as props and emit intent as callbacks; `+page.svelte` / `+page.server.ts` stay thin and do the wiring. No `$app/state` reads inside `packages/ui` (see `storybook`).
- **Styling** — semantic CSS custom properties from the tokens stylesheet (`var(--color-bg-surface)`), never raw values or primitive ramps in components.
- **Tooling** — ESLint + Prettier, config inherited from the repo root. `no-console` is on; server-side logs go through the OTel setup in `hooks.server.ts`.
- **Tests** — Vitest beside the file for pure logic. Component behaviour is asserted in Storybook play functions, not unit tests. Types from `@repo/contracts` are the only request/response shapes.
- **Generated code** — `*_pb.ts`, `*_connect.ts` are never edited; change the proto and run `just gen`.
