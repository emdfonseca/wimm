---
paths:
  - "**/*.py"
---

# Python conventions

Decisions only; PEP 8 and idiomatic modern Python apply as written.

- **Tooling** — `uv` for everything: environments, resolution, locking, running. No bare `pip`, no global virtualenvs. Workspace members declared in the root `pyproject.toml`.
- **Layout** — src layout (`src/<package>/`), tests in `tests/`. Distribution name `<org>-<dir>`, import name `<dir>` with underscores. Public surface exported from `__init__.py`; underscore-prefixed modules are private.
- **Types** — hints on every public function and method; `mypy` (strict on new packages) runs in `just check`.
- **Format / lint** — `ruff format` and `ruff check`, config at the repo root. No `print`; structured logging through the OTel setup.
- **Tests** — `pytest`, `parametrize` for tables, small composable fixtures, in-memory fakes over mocks.
- **Errors** — raise specific exception classes defined in the package; never bare `except:`; no exceptions for control flow across package boundaries.
