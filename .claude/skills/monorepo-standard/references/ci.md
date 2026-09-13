# CI

## The one rule

CI runs `devbox run -- just <verb>` and nothing else. Every step must correspond to a recipe a developer can run locally with the same result. A CI step that exists only in YAML is a command nobody can reproduce or debug, and it will drift from the repo it is supposed to guard.

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }        # affected detection needs history
      - uses: jetify-com/devbox-install-action@v0.12.0
      - run: devbox run -- just ci
```

## Affected-package detection

Running everything on every commit stops being viable surprisingly early. Derive the changed set from git and fan out only to those directories:

```just
# changed directories under apps/ and packages/, versus a base ref
_affected base="origin/main":
    @git diff --name-only {{base}}...HEAD \
      | grep -E '^(apps|packages)/' \
      | cut -d/ -f1,2 | sort -u

ci-affected base="origin/main":
    @for d in $(just _affected {{base}}); do just run check "$d" || exit 1; done
```

Two honest caveats. A change to a package must also test its dependents — either walk the dependency graph or, more simply, treat a `packages/*` change as triggering everything that depends on it. And a change to shared roots (`devbox.json`, the root justfile, workspace manifests, CI config) invalidates the whole thing: run the full suite for those.

Start with the full run, add affected detection when it hurts. Wrong affected logic silently skips tests, which is worse than a slow pipeline.

## Caching

Cache per language, keyed on the relevant lockfile: Go build and module cache, the pnpm store, and the uv cache. Do not build a caching layer in `just` — every one of those tools already does this correctly, and a second cache is a second thing to invalidate wrongly.

## Required checks

- `just ci` (or `ci-affected`) passes.
- Generated code is current: run `just gen`, then fail if `git diff --exit-code` shows changes.
- Lockfiles are current and committed (`--frozen-lockfile` equivalents per tool).
- Boundary linters pass — they live in `just lint`, so this comes for free.
