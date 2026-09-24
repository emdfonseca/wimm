## Context

See proposal.md for why. This builds on `version-and-approve-page-stories`
(ADR 0027), which must be applied first.

- `just approve` runs `approvals.mjs --approve`. It regenerates with `just gen`
  in `apps/storybook`, which runs every page story headless in Chromium through
  `@vitest/browser-playwright` and records each story's fingerprint from an
  `afterEach` in `.storybook/preview.ts` via the custom command
  `recordPageVersion` (`vitest.config.ts`). It then refuses or asks for `yes`
  and appends one line to `canvas/approvals.jsonl`.
- The canvas draws an artboard per story and size (`viewports.js`: compact
  390×844, medium 834×1112, wide 1440×900, ultra 1920×1080), in the theme and
  density from the address, defaulting to light and comfortable. `artboard()` in
  `canvas.js` attaches the badge and the History button.
- `.storybook/main.ts` serves `apps/storybook/canvas/` at `/canvas/`, so a file
  under `canvas/approved/` is fetchable beside the page.
- No image is tracked in the repository today, and there is no LFS.

## Language

- **picture**: a PNG of one page story at one size, light and comfortable, of
  its last approved version. Never "screenshot", "snapshot" or "baseline" in
  copy or identifiers; "snapshot" already means something else in test tooling.
- **approved look**: what an artboard shows while switched to its picture. The
  control is named for it. Never "diff" or "compare": nothing is compared.
- **last approved version**: the fingerprint of the newest line in the
  approvals record for a story, as ADR 0027 defines it. The words "Latest
  approved" in the history panel name the same thing.

## Goals / Non-Goals

**Goals:**
- A picture always shows exactly the version its file name says.
- Nothing about pictures is written unless an approval is.

**Non-Goals:**
- Pictures of any version but the last approved one.
- Dark theme or compact density pictures. The person chose light and
  comfortable only.
- Showing what changed. The picture and the page are shown in turn, not
  overlaid or subtracted.
- Pictures for approvals recorded before this change.

## Decisions

**1. The picture is taken in the approve run's own render.** The regenerate step
already renders every page story headless. `approvals.mjs` sets
`WIMM_PICTURE_STORIES=<story ids>` for that run; the `afterEach` in `preview.ts`,
for that story only and after recording its markup digest, sets the light theme
and comfortable density, then for each size the story is drawn at sets the
viewport with `page.viewport`, waits for fonts and two animation frames, grows
the height until nothing sits behind a scrollbar (`hiddenHeight`, shared with
`canvas.js`, so the picture is the artboard's height), and takes the document
root with `page.screenshot({ element, save: false })`. The command
`pictureWindow` grows the runner's own window past the frame first, because the
runner scales a frame larger than its window down. The command `keepPicture`
writes `<story>/<size>.png` and a line naming story, size and markup digest to
`node_modules/.cache/wimm-canvas/pictures/`, which every run empties when it
starts. The fingerprint is the digest plus the shared styles, known only once
`versions.mjs` has run, so `approvals.mjs` names each picture when it keeps it.

*Alternative:* a separate Playwright script against a static Storybook build
after the prompt. Rejected: it renders a second time, possibly on a different
build, so the picture could show a version other than the one approved, which
is the one thing it must never do.

**2. A picture's file name carries its fingerprint.**
`canvas/approved/<story id>/<fingerprint>-<size>.png`, with the full 64
characters. The canvas builds the address from the last approved fingerprint it
already computes for the badge, so no index file is needed, and the check can
tell a stale picture from its name alone.

*Alternative:* `<story id>/<size>.png` plus a manifest. Rejected: two files that
must agree, where one name cannot disagree with itself.

**3. Pictures move only after the approval line is written.** After `yes`,
`approvals.mjs` confirms every cached picture names the fingerprint being
approved and every size the story is drawn at is present, copies them to
`canvas/approved/<story id>.next/`, appends the approval line, then replaces
`canvas/approved/<story id>/` with the `.next` folder. A refusal or a `no` never
reaches this step. A failure before the append leaves `.next` behind, which the
check refuses by name, so it is noticed rather than committed.

**4. The check lives in `approvals.mjs --check`.** It already validates the
approvals record. It gains: every file under `canvas/approved/` is
`<story>/<fingerprint>-<size>.png` where the story has an approval, the
fingerprint is its last approved one, and the size is in `viewports.js`; no
`.next` folder exists. A story with no pictures passes. Each rule gets a
fixture that must be refused, per `.claude/rules/generated-artifacts.md`.

**5. The switch is a toggle button on the artboard's label.** Beside History,
reading `Approved look`, with `aria-pressed`. Pressed, the artboard's iframe is
hidden, not removed, so switching back does not re-render, and an `<img>` at the
size's width shows the picture at its natural height. The badge gains the line
`Showing approved 77b0d4a`. The control is drawn only where the status is
changed, the theme light, the density comfortable, and a request for the
picture at that size succeeded; otherwise, where it is changed but the request
failed, the badge gains `No picture of the approved version`. The request is a
`HEAD` made for changed artboards only, so an unchanged canvas makes none.

Pictures resolve against the same base as the data files, so a fixture pair
under `canvas/fixtures/<name>/` can carry its own `approved/` folder and the
states are looked at without approving anything.

**7. One run approves many stories.** `just approve` takes story ids, then an
optional note: leading arguments shaped like a story id are stories, and the
first that is not starts the note. `--needs` instead takes every page story whose
status was changed or never when the command started, from the versions file as
it stood before regenerating, which is what the canvas showed. The regenerate
step runs once with `WIMM_PICTURE_STORIES` naming every story, comma-separated.
Every refusal is checked for every story before the prompt; any one refuses the
run, naming each. Pictures for all are staged, the lines appended in one write,
then every folder swapped in.

*Alternative:* a baseline-only `--all`. Rejected by the person: a global change
moves every page, and approving after it is the same batch.

**6. PNG, uncompressed beyond what Chromium writes.** No image dependency is
added. The first approval measures the files, and the ADR records the sizes.

## Risks / Trade-offs

- [Binary files grow the repository, and every re-approval adds a set to
  history even though the working tree keeps one] → Only the last version is
  kept on disk; the ADR states the measured size per approval, and a later
  change can move pictures to LFS if the total becomes a problem.
- [A size drawn by JavaScript breakpoints may render differently after
  `page.viewport` than on first render] → It is still the approved version at
  that size, which is what the picture claims. The look task compares a picture
  with the live artboard at every size.
- [The approve run gets slower by four element captures] → One story only; the
  task times it.
- [A person in dark or compact finds no control] → Stated in the state plan;
  switching the canvas to light and comfortable brings it back.

## Migration Plan

None. Approvals already recorded keep working and read "No picture of the
approved version" when changed. Rollback is deleting `canvas/approved/` and
reverting the code.
