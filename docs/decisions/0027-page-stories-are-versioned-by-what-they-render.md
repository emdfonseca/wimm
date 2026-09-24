# 0027 · Page stories are versioned by what they render, and only a person approves one

## Status

Accepted

## Context

ADR 0023 made a page's rendered story the design of record and a look the gate
before wiring. The only trace a look leaves is prose under Seen in a change's
`canvas.md`, which archives with the change and says nothing about the page as it
renders today. A page can move after it was looked at and the design canvas
shows it as before.

Options weighed for the version: computed styles of every element (multiplies
the run by elements and sizes, differs between browsers); a hash of the source
files a story imports (every refactor reads as a change); pixel comparison
(ADR 0020's rejection stands: a tolerance loose enough for two text engines
misses what matters). For the record: a server or database (the canvas is static
and has no server), or a file beside `tokens.json` (needs a second static mount,
and `design/` holds what CI owns).

## Decision

A page story's version is a fingerprint: sha256 of the normalised markup of the
story's root, taken after its play function, plus one digest of `tokens.css`,
`base.css`, `fonts.css` and the font files. In the headless run the Storybook app hashes each component's CSS into its scoped
class (`compilerOptions.cssHash`), so a style change moves the markup; dev keeps
the file-name hash and its hot reload. One version covers a story at every size.

The headless run records a digest per page story. `apps/storybook/scripts/
versions.mjs` validates the run, then writes `apps/storybook/canvas/versions.json`:
the implemented version, and each earlier version only if an approval names it.
`just gen` and `just check` both write it; a check never fails because the file
was out of date, so a visible change never fails CI.

An approval is one line of `apps/storybook/canvas/approvals.jsonl`: story,
fingerprint, name, email, time, note. `just approve <story> [note]` is the only
writer. It refuses inside an agent session (`CLAUDECODE`,
`CLAUDE_CODE_ENTRYPOINT`) and without a terminal, regenerates first and refuses
if the implemented version moved, asks for `yes`, and appends. `just check`
validates every line and requires the file at `HEAD` (and at the merge base with
`origin/main`, where reachable) to be a byte prefix of the file on disk.
`CLAUDE.md` forbids an agent from running the command or writing the file.

Seen in `canvas.md` still says what a change looked at. An approval is the
standing record against a version.

## Consequences

The record proves who committed an approval. It does not prove that a person
looked. Signed commits are stronger evidence and are available, not required.

The append-only check does not see a rewrite pushed straight to `main`, or a
record and its history edited together. The agent guard is a convention backed by
a check on shape, not proof.

A check run can leave a diff in `versions.json` to commit; that diff is the list
of pages whose look moved. Until it is committed the canvas can read Approved
over a page that has since moved. A global change marks every approvable story
changed at once, which is correct. Two people approving on two branches both
append; the merge keeps both lines.

A renamed story is a new story nobody has approved. Versions of a removed story
survive only when approved. No new dependency.
