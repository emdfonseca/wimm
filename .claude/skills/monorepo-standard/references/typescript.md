# TypeScript / Node

## Workspace membership

```yaml
# pnpm-workspace.yaml
packages:
  - apps/*
  - packages/*
  - tools/*
```

pnpm resolves dependencies; internal packages are referenced as `"@repo/ui": "workspace:*"`. The `workspace:` protocol means a stale published copy can never shadow local source.

## Naming

Package name is `@repo/<directory>` — identical to the directory, so grep and import statements agree. Substitute your own scope for `@repo` and keep it consistent; a mixed scope is a permanent source of confusion about what is internal.

## tsconfig

One base at the root, extended everywhere:

```json
// tsconfig.base.json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "skipLibCheck": true,
    "declaration": true,
    "isolatedModules": true
  }
}
```

Packages extend it and add only what differs (`outDir`, `rootDir`, `lib`, JSX). A package overriding `strict` needs a comment explaining why, because it silently weakens every consumer that inlines its types.

## Internal packages: source or build?

Prefer publishing internal packages as **source** (`"exports": { ".": "./src/index.ts" }`) when every consumer is a bundler or a TS runtime. It removes a whole build step and its staleness failure mode.

Switch a package to a real build when it is consumed by plain Node, published externally, or when typechecking consumers gets slow. State which model a package uses in its `package.json` — the difference is invisible otherwise.

## Where Turborepo fits

Not needed on day one. The task layer is `just`, and `pnpm -r` plus native caching covers a small repo.

Add Turborepo when TS build times actually hurt — typically several packages with real build steps, or CI where remote caching pays off. Scope it to the TS subtree and keep `just` as the entry point (`just build web` → `pnpm turbo build --filter=@repo/web`), so the verb vocabulary is unchanged for everyone else and Go/Python/infra never need shim manifests.

What Turborepo will not do for you: orchestrate Go or Python tasks. Its graph comes from `package.json`. Dragging non-TS packages into it means writing fake manifests, which is the point at which Moon or Bazel becomes the honest choice instead.

## Outputs

Build artifacts go to the package's own `dist/`, never a shared root output directory — shared output directories make caching and cleaning ambiguous. `.gitignore` covers `dist/`, `.turbo/`, `node_modules/`.
