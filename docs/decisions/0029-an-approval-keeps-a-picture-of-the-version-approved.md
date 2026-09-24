# 0029 · An approval keeps a picture of the version approved

## Status

Accepted

## Context

ADR 0027 records an approval as a fingerprint of a page story's markup. The
canvas can say a page changed since approval, but a fingerprint cannot be drawn,
so seeing what moved means remembering the old page.

Options weighed: rebuilding the approved version on demand (minutes per look, and
it depends on an old commit still building); a separate Playwright run against a
static build after the prompt (a second render, possibly of another build, so the
picture could show a version other than the one approved); a picture per size
under a fixed name plus a manifest (two files that must agree); Git LFS (a
service and a setup step for one folder of small files).

## Decision

`just approve` keeps a PNG of each story it approves at every size the canvas
draws it at, light theme and comfortable density only, the whole page grown to
the height the canvas gives its artboard. The pictures come from the approve
run's own render: `approvals.mjs` sets `WIMM_PICTURE_STORIES`, and the
`afterEach` in `.storybook/preview.ts` captures each named story after recording
its markup digest, through a `keepPicture` command in `vitest.config.ts` writing
to `node_modules/.cache/wimm-canvas/pictures/`, emptied at the start of every
run. The fingerprint is that digest plus the shared styles, so `approvals.mjs`
names each picture when it keeps it.

One run approves many stories: `just approve <story>... [note]`, or
`just approve --needs [note]` for every page story that was changed or never
approved when the command started. It regenerates once, lists the stories, asks
once, and refuses the whole run when any story is refused. This amends 0027's
one story per approval.

A picture is `apps/storybook/canvas/approved/<story>/<fingerprint>-<size>.png`,
the full 64-character fingerprint. Only the last approved version keeps pictures;
approving replaces the story's folder. Before asking, the command confirms every
size has a picture of the markup the run recorded. After `yes`, it copies them
to `<story>.next/`, appends the approval line, then swaps the folder in. A
`no` or a refusal writes no picture.

`approvals.mjs --check`, inside `just check apps/storybook`, refuses a picture not
of its story's last approved version, one belonging to no approved story, one at
a size `viewports.js` lacks, and any `.next` folder left behind. A story with no
picture passes. No image dependency: the PNG is what Chromium writes.

## Consequences

The canvas builds a picture's address from the fingerprint it already shows, so
no index exists and a stale picture is visible from its name alone.

These are the first binary files in the repository. The working tree holds one
set per approved story, and every re-approval adds a set to history. The first
real approval measures one set and the time it adds to the approve run, and this
record states both.

Approvals recorded before this decision have no picture; their changed artboards
say so until someone approves again. Dark theme and compact density have no
picture. `CLAUDE.md` extends the agent rule to `canvas/approved/`.
