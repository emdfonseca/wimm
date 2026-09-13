# Python

## Tooling

`uv` handles environments, resolution, and locking; devbox pins `uv` itself. No `pip install` outside uv, no global virtualenvs — that path ends in a machine that works for reasons nobody can reproduce.

## Workspace

```toml
# pyproject.toml at the repo root
[tool.uv.workspace]
members = ["apps/example-worker", "packages/example-lib"]
```

Each member has its own `pyproject.toml`. Internal dependencies use the workspace source:

```toml
[tool.uv.sources]
pipeline = { workspace = true }
```

`uv.lock` is committed at the root — one lockfile for the workspace keeps versions consistent across members.

## Layout

```text
apps/ingest/
├── pyproject.toml
├── justfile
├── src/ingest/
│   ├── __init__.py      # the public surface
│   └── _internal/       # private by convention
└── tests/
```

Use the src layout — it prevents tests from accidentally importing the working directory instead of the installed package, which is a genuinely confusing class of bug. Distribution name is `<org>-<dir>`, import name is the directory with underscores.

## Tasks

```just
build:
    uv build

test:
    uv run pytest

lint:
    uv run ruff check .

fmt:
    uv run ruff format .

check: lint
    uv run mypy src
    just test
```

`ruff` covers formatting and linting; `mypy` runs in `check`, not `lint`, so `lint` stays fast.
