# 0011 · Draw headless, not through the MCP

## Status

Accepted; superseded by 0023

Nothing is drawn, headless or otherwise.

## Context

ADR 0009 built the canvas workflow around the pencil MCP, with `just pen-save` to
flush the app's document afterwards. That rests on the MCP editing the file it is
given.

It does not. `mcp__pencil__execute` ignores its `filePath` argument and acts on
whatever document Pen.app has open. Pointing it at an empty journey file and
asking for the root children returns the library's seven zones and its 120
variables. Nothing reports the mismatch.

So an MCP call meant to draw a journey edits `product-ui.lib.pen` instead — the
file every screen instances. The only thing that prevented it here is that a
`ui:` ref cannot resolve inside the library itself, so the block failed and rolled
back.

## Decision

Journeys are drawn headless. `just pen-exec <file.pen>` pipes an execute snippet
into `pen interactive --in X --out X`, which opens X, resolves its imports, saves,
and exits.

```text
just pen-import   create the journey and bind the library
just pen-exec     draw, and verify in a second call
Export + read     look at the result
```

The MCP keeps one job: the document a person is actually looking at. `just
pen-save` keeps one job with it — flushing that document, which stays in the
app's memory until something saves it.

**`pen-exec` fails closed, by restoring rather than by abstaining.** A snippet
that errors exits non-zero and leaves the file byte-identical; both are
asserted, because a tool that writes garbage and then exits non-zero has still
corrupted the artifact.

The shim has to *make* that true rather than report it. `pen interactive`
applies operations one at a time and the `save()` piped after them writes
whatever succeeded before the error, so a snippet failing part way through
leaves the file partly rewritten. So the shim snapshots the file first and
restores it when the output carries an error — and `bin/test-pen-exec` feeds it
a stub that writes and then fails, asserting both that the run failed and that
the file did not move. Reverting the restore turns three of its five cases red.

**Headless saves only because it is told to.** `pen interactive` does not save on
exit. A pipe without `save()` reports every id it created and writes none of
them — the failure looks exactly like success.

This supersedes the workflow in ADR 0009. The reason that ADR exists — an edit
sitting unsaved in the app with nothing reporting it — is unchanged and still
applies to MCP work.

## Consequences

- Visual verification changes shape. Headless returns no screenshot, so a frame
  is exported to PNG and the file is read. This works and costs one extra step.
- The app and headless must not hold the same file. The app's copy is stale the
  moment headless writes, and its next save wins silently. Headless owns files
  nobody has open.
- `Get` visitors in an importing file walk the library's nodes too, prefixed.
  Anything counting or auditing journey nodes filters on the prefix first.
- The pen-design skill's instruction to pass an explicit `filePath` was worse
  than useless: it read as a safeguard while doing nothing. It is now a named
  trap.
