# just: the task layer

Every command a human or CI runs is a justfile recipe. The payoff is that a newcomer can run `just test api` without knowing whether `api` is Go, TS, or Python — and that there is exactly one place to look for what a command does.

## The verb vocabulary

Every package implements the same verbs:

| Verb | Means |
|---|---|
| `build` | produce artifacts |
| `test` | run tests |
| `lint` | static analysis, no writes |
| `fmt` | format in place |
| `check` | lint + typecheck + test; what CI runs per package |
| `dev` | local run or watch loop |
| `clean` | remove build outputs |

A package that genuinely cannot support a verb should define it to fail with a clear message. Silence is worse: a missing `test` recipe and a broken one look identical from the root.

## Root justfile: delegation, not implementation

The root knows which packages exist and forwards to them. It does not know how any package builds.

```just
set shell := ["bash", "-uc"]

pkgs := "apps/web apps/api apps/ingest packages/ui packages/telemetry"

default:
    @just --list

# Run a verb in one directory: just run test apps/api
run verb dir:
    @just --justfile {{dir}}/justfile --working-directory {{dir}} {{verb}}

# Run a verb everywhere
all verb:
    @for d in {{pkgs}}; do echo "==> $d"; just run {{verb}} "$d" || exit 1; done

build dir="": (_fanout "build" dir)
test  dir="": (_fanout "test" dir)
lint  dir="": (_fanout "lint" dir)
check dir="": (_fanout "check" dir)

_fanout verb dir:
    @if [ -z "{{dir}}" ]; then just all {{verb}}; else just run {{verb}} {{dir}}; fi

ci: (all "check")
```

Resolving a bare name (`just test api` rather than `just test apps/api`) is a small convenience worth adding once the package list grows — keep the resolution in one helper recipe so it stays debuggable.

## Package justfile: the only place build logic lives

```just
# apps/api/justfile  (Go)
build:
    go build -o ../../dist/api ./cmd/api

test:
    go test ./...

lint:
    golangci-lint run

fmt:
    gofmt -w .

check: lint test

dev:
    go run ./cmd/api

clean:
    rm -rf ../../dist/api
```

```just
# apps/web/justfile  (TS)
build:
    pnpm build

test:
    pnpm vitest run

lint:
    pnpm eslint .

fmt:
    pnpm prettier --write .

check: lint
    pnpm tsc --noEmit
    just test

dev:
    pnpm dev

clean:
    rm -rf dist .turbo node_modules/.cache
```

## Conventions that keep justfiles readable

- Recipes stay short. More than roughly ten lines means it wants to be a script in `tools/` that the recipe calls — justfiles are an index of entry points, not a scripting language.
- Private helpers are prefixed `_`.
- Comments above a recipe show up in `just --list`, so write them for the person running `just` with no context.
- Never duplicate a language's own dependency graph in just. `pnpm` and `go build` already know what to rebuild; re-encoding that here creates a second, wrong answer.
