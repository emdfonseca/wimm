## 1. Ground

- [x] 1.1 Confirm `version-and-approve-page-stories` is applied and archived, or
      at least committed, before any other task: this change edits its files.
      Verify: `git status --short apps/storybook` is empty.
- [x] 1.2 Record the decision with `/adr` as ADR 0029, amending 0027: an
      approval also keeps pictures of the version approved, the last approved
      version only, light and comfortable only, named by fingerprint, taken in
      the approve run's own render (design.md decisions 1 to 4 and 6), and
      that pictures are the first binary files in the repository. Leave the
      size per approval to be filled by task 2.4. Verify: `just adr-index`
      regenerates `.claude/rules/decisions.md` and `just adr-index-check`
      passes.
- [x] 1.3 Extend the `CLAUDE.md` approvals line to name
      `apps/storybook/canvas/approved/` beside `approvals.jsonl`. Verify: the
      line names both.

## 2. Keeping the picture

- [x] 2.1 Test-first (`/tdd`), add to `lib.js` `pictureName(story, fingerprint,
      size)` and `parsePictureName(path)`, and `pictureProblems(paths,
      approvals, sizes)` returning one line per file in the words of canvas.md
      (stale, no approved story, unknown size, `.next` left behind). Verify:
      one failing test seen before each rule in `lib.test.js`, and a story
      with no pictures produces no problem.
- [x] 2.2 `keepPicture` command in `vitest.config.ts` writing the PNG and a
      line of story, size and fingerprint under
      `node_modules/.cache/wimm-canvas/pictures/`; in `preview.ts`'s
      `afterEach`, when `WIMM_PICTURE_STORY` names the story, light and
      comfortable, each size the story is drawn at via `page.viewport`, fonts
      and two frames awaited, element captured at full height. Verify: running
      `WIMM_PICTURE_STORY=pages-overview--populated pnpm test` in
      `apps/storybook` writes four PNGs whose fingerprint line matches
      `versions.json`, and a run without the variable writes none.
- [x] 2.3 Test-first, extend `approvals.mjs --approve` per design.md decision 3:
      empty the cache before regenerating, set the variable, after `yes`
      confirm every picture names the approved fingerprint and every size is
      present, stage to `<story>.next/`, append, swap; print the words of
      canvas.md. Verify in `scripts/approve.test.js`: approving writes the
      pictures and the line; approving a newer version leaves only the new
      pictures; a second approver of one version leaves one picture per size;
      `no` and every refusal leave `canvas/approved/` byte-identical; a missing
      size refuses and records nothing.
- [x] 2.5 Test-first, `approve` takes several stories or `--needs` per
      design.md decision 7: one regenerate with `WIMM_PICTURE_STORIES`, every
      refusal checked for every story before the prompt, one question, one
      append; `justfile` passes every argument through. Verify in
      `scripts/approve.test.js`: three named stories record three lines and
      three picture folders; `--needs` takes changed and never, skips approved
      and behaviour; nothing needing approval records nothing; one bad story
      refuses the run and names it; named stories with `--needs` refuses.
- [ ] 2.4 Measure. Approve one real story in your own terminal (a person, not
      an agent) and write the four files' sizes and the approve run's added
      time into ADR 0029. Verify: the ADR states both numbers.

## 3. The check

- [x] 3.1 Test-first, `approvals.mjs --check` lists `canvas/approved/` and fails
      with `pictureProblems`. Verify: a committed fixture per rule under
      `scripts/fixtures/` is refused with its line, and the real tree passes.

## 4. The switch

- [x] 4.1 Test-first, add to `lib.js` `offersApprovedLook(status, view)` (changed,
      light, comfortable) and `pictureLine(state, fingerprint)` returning the
      two badge lines. Verify: unit tests for each row of canvas.md's table
      held by a unit test.
- [x] 4.2 Add the `changed-pictured` fixture: `changed` plus
      `approved/pages-overview--populated/<fingerprint>-compact.png` and
      `-wide.png`, small real PNGs. Verify: `?data=changed-pictured` loads with
      no console error.
- [x] 4.3 `approvals-ui.js`: the toggle and picture functions; `canvas.js`:
      `HEAD` the picture for changed artboards only when offered, draw the
      toggle or the no-picture line, hide and restore the iframe; `canvas.css`:
      the toggle styled like History. Verify: every contract in canvas.md
      asserted in `approvals.browser.test.js`, and a canvas with no changed
      artboard makes no picture request.

## 5. Look

- [ ] 5.1 Open `changed-pictured`, `changed`, `approved` and `never` on a screen
      view and a flow view, light and dark, at compact, medium, wide and ultra
      (`?sizes=compact,medium,wide,ultra`), switch every offered toggle both
      ways, and compare one real picture from task 2.4 with its live artboard
      at every size. Change what is wrong and write what was seen into
      canvas.md under Seen. Verify: Seen names every fixture at every size.

## 6. Close

- [x] 6.1 Update `README.md`'s approvals section: what a picture is, that only
      the last approved version keeps one, light and comfortable only, and the
      commit instruction the command prints. Verify: `just check apps/storybook`
      and `just check packages/ui` pass.
