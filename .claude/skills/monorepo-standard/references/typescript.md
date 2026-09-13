# TypeScript / Node

## Workspace

```yaml
# pnpm-workspace.yaml
packages:
  - apps/*
  - packages/*
```

Internal dependencies are `"@repo/<dir>": "workspace:*"`; the package name equals the directory.

Every `tsconfig.json` extends the root `tsconfig.base.json` (see `assets/tsconfig.base.json`) and adds only what differs (`outDir`, `rootDir`, `lib`, JSX). Overriding `strict` needs a comment saying why.

## Source or build

Default: internal packages export source, `"exports": { ".": "./src/index.ts" }`. Consumers are bundlers or TS runtimes, so there is no build step to go stale.

Switch to a real build only when a package is consumed by plain Node or published externally. State which model a package uses in its `package.json`.

Build outputs go to the package's own `dist/`, never a shared root directory. `.gitignore` covers `dist/` and `node_modules/`.
