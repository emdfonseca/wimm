# devbox: the toolchain layer

devbox answers exactly one question: **which tools, at which versions, exist in this repo.** Local shells and CI runners both resolve tools through it, which is what makes "works on my machine" a non-statement.

## What goes in devbox.json

Every binary any task needs, pinned:

```json
{
  "packages": [
    "go@1.23",
    "nodejs@22",
    "pnpm@9",
    "uv@0.5",
    "terraform@1.9",
    "just@1.36",
    "golangci-lint@1.61",
    "protobuf@28"
  ],
  "shell": {
    "init_hook": ["export GOFLAGS=-mod=readonly"],
    "scripts": {}
  }
}
```

Pin to a real version, never a floating alias. `devbox.lock` is committed — the pin plus the lock is what makes two machines agree.

## What does not go in devbox.json

**Task logic.** `shell.scripts` stays empty, or holds at most a bootstrap entry. The reason is concrete: devbox scripts cannot take arguments, cannot declare dependencies on each other, and cannot be listed with their documentation. Splitting tasks between `devbox.json` and the justfile means every future reader has to check two places to find where a command lives, and they will check the wrong one first.

The one defensible exception is a `setup` script that installs language-level dependencies after entering the shell — and even that is better as `just setup`.

## CI parity

CI installs devbox and then runs tasks through it:

```yaml
- uses: jetify-com/devbox-install-action@v0.12.0
- run: devbox run -- just ci
```

That single line is the contract. If CI installs a tool devbox does not declare, or runs a command no justfile recipe defines, the repo now has two sources of truth and the CI one will drift.

## Adding a tool

1. Add it to `packages` with a pinned version.
2. Run `devbox install` and commit `devbox.lock`.
3. Wire it into a justfile recipe — a tool nobody can invoke through `just` is a tool nobody will discover.
4. If it generates code or artifacts, say where those go and whether they are committed.

## Language-level dependencies are not devbox's job

devbox provides `go`, `pnpm`, and `uv`. It does not resolve your Go modules, npm packages, or Python requirements — those belong to `go.work`/`go.mod`, the pnpm lockfile, and `uv.lock`. Keeping that line clean is what lets each language's native caching work properly.
