---
name: monorepo
paths:
  - "justfile"
  - "**/justfile"
  - "devbox.json"
  - "pnpm-workspace.yaml"
  - "go.work"
  - ".github/workflows/**"
description: Layout standard for the polyglot monorepo - apps/ and packages/, devbox owns the toolchain, just owns every entry point. Use whenever someone adds an app, service, package, or module; asks where a file or shared code lives; changes devbox.json, a justfile, pnpm-workspace.yaml, go.work, or CI workflows; or shares code between languages. Even for "add a script" or "where do I put this helper".
---

# Monorepo standard

Deployable → `apps/<name>`. Importable → `packages/<name>`. Root holds only `devbox.json`, `justfile`, workspace manifests, CI. No `tools/`, no `infra/`. One directory = one language.

## Rules

1. `devbox.json` owns the toolchain; `just` owns every entry point. CI runs `devbox run -- just ci` only. No task logic in devbox scripts, no versions in CI YAML.
2. Every `apps/*` and `packages/*` dir has a `justfile` with `build test lint fmt check dev clean`; unsupported verbs fail loudly. The root justfile discovers packages by it.
3. Dependencies: `apps → packages`, `packages → packages` (acyclic). Never `packages → apps` or `apps → apps`; extract a package.
4. Banned directory names: `lib`, `utils`, `common`, `misc`, `helpers`. Name the domain or the capability.
5. Cross-language sharing is a generated contract in `packages/contracts` (`just gen`), never copied types. Generated files are never edited (`hooks/protect-generated.sh`).
6. Kebab-case domain nouns; manifests match the directory: `@repo/<dir>`, `<module-prefix>/<path>`, `<org>-<dir>`.

## References

| Read | When |
|---|---|
| `references/go.md` | `go.work`, module paths |
| `references/typescript.md` | pnpm, source vs build |
| `references/python.md` | uv workspace, src layout |
| `references/ci.md` | The CI job |
| `references/new-package-checklist.md` | Adding a package |
| `assets/` | `devbox.json` (pins checked 2026-09-13), justfiles, tsconfig, `hooks/` |
