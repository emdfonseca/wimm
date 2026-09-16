# Root justfile — delegation only. Build logic lives in package justfiles.
set shell := ["bash", "-uc"]

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

# What CI runs. The database is started first: store tests refuse to run
# without one rather than skipping.
ci: adr-index-check pen-exec-test db-up (all "check")

# The whole stack under process-compose: postgres, wimmd, the web app, Storybook.
up *args:
    @devbox services up {{args}}

# Stop it.
down:
    @devbox services stop || true
    @bin/db down

# What is running.
ps:
    @devbox services ls

# Local PostgreSQL: start the cluster, create the database, migrate to head.
db-up:
    @bin/db up

# Stop the local cluster.
db-down:
    @bin/db down

# Drop the database, recreate it, migrate to head. Empty of data, schema at head.
db-reset:
    @bin/db reset

# Connection string for the local database.
db-url:
    @bin/db url

# A psql shell on the local database.
db-psql:
    @bin/db psql

# OpenSpec CLI. The pin lives in bin/openspec, which devbox puts on PATH.
openspec *args:
    @bin/openspec {{args}}

# Flush a .pen document from the running Pen.app to disk.
pen-save file:
    @bin/pen-save {{file}}

# Bind a .pen library into a journey file. Refs then read <alias>:<nodeId>.
pen-import journey alias library:
    @bin/pen-import {{journey}} {{alias}} {{library}}

# Run an execute snippet against a .pen file, headless. Snippet on stdin.
pen-exec file snippet="":
    @bin/pen-exec {{file}} {{snippet}}

# Issue a locally-trusted certificate for the dev server. Enable Banking
# requires an https redirect URI even on localhost.
dev-cert *args:
    @bin/dev-cert {{args}}

# Assert pen-exec leaves a file byte-identical when a snippet fails (ADR 0011).
# Runs against a stub rather than the real CLI: it is the shim's own failure
# handling under test, and a test needing the network would stop being run.
pen-exec-test:
    @bin/test-pen-exec

# Write or verify the library manifest. `verify` fails if a node that already
# existed was renamed, retyped, reparented or deleted.
pen-manifest mode="verify":
    @bin/pen-manifest {{mode}}

# Assert the pen library's variables match design/tokens.json.
pen-verify-tokens library="":
    @bin/pen-verify-tokens {{library}}
