# Root justfile — delegation only. Build logic lives in package justfiles.
set shell := ["bash", "-uc"]

openspec_version := "1.13.0"

# Every apps/* and packages/* directory that has a justfile
pkgs := `ls -d apps/*/justfile packages/*/justfile 2>/dev/null | sed 's|/justfile$||' | tr '\n' ' '`

# List available recipes
default:
    @just --list

# Run a verb in one directory: just run test apps/api
run verb dir:
    @just --justfile {{dir}}/justfile --working-directory {{dir}} {{verb}}

# Run a verb in every registered package
all verb:
    @for d in {{pkgs}}; do echo "==> $d"; just run {{verb}} "$d" || exit 1; done

build dir="": (_fanout "build" dir)
test  dir="": (_fanout "test"  dir)
lint  dir="": (_fanout "lint"  dir)
fmt   dir="": (_fanout "fmt"   dir)
check dir="": (_fanout "check" dir)
clean dir="": (_fanout "clean" dir)

_fanout verb dir:
    @if [ -z "{{dir}}" ]; then just all {{verb}}; else just run {{verb}} {{dir}}; fi

# Regenerate generated artifacts. No directory regenerates everything.
gen dir="":
    @set -e; if [ -z "{{dir}}" ]; then just adr-index && just all gen; else just run gen {{dir}}; fi

# Rebuild .claude/rules/decisions.md from docs/decisions/
adr-index:
    @python3 .claude/skills/adr/assets/adr-index.py

# Fail if the decision index is stale
adr-index-check:
    @python3 .claude/skills/adr/assets/adr-index.py --check

# What CI runs
ci: adr-index-check (all "check")

# OpenSpec CLI, version-pinned. `just openspec list`, `just openspec validate`.
openspec *args:
    @pnpm dlx @fission-ai/openspec@{{openspec_version}} {{args}}
