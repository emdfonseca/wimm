# 0009 · The canvas saves without a person

## Status

Accepted

## Context

ADR 0008 made the canvas a proposal artifact and recorded, as its sharpest
consequence, that a pencil edit is not on disk until a person saves it in
pen.dev. That was true of the pencil MCP alone: it mutates the document Pen.app
has open, the file keeps its mtime, and `git status` stays empty.

The failure is silent in an unusual way. Every reader agrees with the writer —
MCP reads, CLI reads and screenshots all resolve against the same unsaved
document — so nothing anywhere reports that the work is not persisted. It is
found when someone looks at `git status` and sees nothing, or worse, when they
do not.

`@pen.dev/cli` has an interactive shell that exposes the same tools plus `save()`,
which the MCP does not have, and it can attach to the already-running desktop
app. The shell reads from stdin, so the command can be piped.

## Decision

`bin/pen-save <file.pen>` pipes `save()` into the pinned CLI attached to the
running app, and `just pen-save <file.pen>` is the shape everything uses.

```text
1. execute   — mutate through the MCP
2. execute   — verify in a separate call
3. pen-save  — flush, and confirm git sees the file
```

The MCP stays the way edits are made. It is precise and deterministic: the
alternative the CLI offers is `pen --in … --out … --prompt …`, which runs a
second AI agent against the file and is neither. The CLI is used for the one
thing the MCP cannot do.

**The shim verifies rather than trusts.** It records the file's mtime, runs the
save, and fails if the file did not move — so "Is Pen.app running?" surfaces as
an error rather than as a canvas.md describing frames nobody can see.

**It runs through npm, not pnpm.** `@pen.dev/cli` imports `css-tree` without
declaring it, which resolves under npm's flat layout and fails under pnpm's
strict linking. This is the one place in the repo that reaches for npm, and the
reason is a defect in the package rather than a preference.

This supersedes the save consequence in ADR 0008. The canvas artifact itself is
unchanged.

## Consequences

- The workflow needs Pen.app running with the file open. That was already true
  of every MCP call; the shim now says so when it is not.
- The CLI is pinned to 0.3.7 while Pen.app updates on its own schedule. The two
  version independently, so an app update can outrun the pin — the symptom would
  be a connection or protocol error from `just pen-save`, not a silent failure.
- A save with no pending edits rewrites the same bytes and produces no diff, so
  saving too often is free. Saving too rarely loses work that every tool still
  claims is there.
- `stat` differs between the BSD one macOS ships and the GNU one devbox puts
  ahead of it. The shim handles both; anything else reading mtimes in this repo
  has to as well.
