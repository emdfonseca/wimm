# Versioning and release

## Internal packages carry no version

Anything consumed only inside the repo is referenced by workspace protocol (`workspace:*`, `go.work`, `uv` workspace source) and never gets a version number. Versioning internal packages creates a bump-and-publish ritual that buys nothing, because there is only ever one consumer set and it is in the same commit.

## Published packages

Only packages with external consumers get versions, and each one says so explicitly in its manifest (`"private": false` for npm; a tagged Go module path; a PyPI distribution).

- **TypeScript:** Changesets. A contributor adds a changeset in the PR; release automation bumps and publishes.
- **Go:** module tags of the form `packages/telemetry/v1.2.3`. Consumers outside the repo resolve that tag; the tag path must match the module path exactly or nothing resolves.
- **Python:** version in `pyproject.toml`, published with `uv build` + `uv publish`.

## Apps

Apps are deployed, not versioned. Their identity is the commit and the built artifact digest. A semantic version on a deployable is a second source of truth about what is running, and it will disagree with reality eventually.

## Commits

Conventional Commits, with the package as the scope: `feat(billing): add proration`. The scope is the directory name, which keeps it greppable and lets tooling attribute changes to packages without a mapping table.

## Changelogs

Generated from commits and tags. Never hand-maintained: a hand-written changelog is history duplicated out of git, and it goes stale in exactly the way git cannot.
