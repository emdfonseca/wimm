## 1. Ground

- [ ] 1.1 Commit `apps/storybook/canvas/` as it stands, on its own, before any
      other task. It is untracked today and one script has already cut tests out
      of it with nothing to restore from. A person makes this commit. Verify:
      `git ls-files apps/storybook/canvas | wc -l` is not zero and `git status
      --short apps/storybook/canvas` is empty.
- [ ] 1.2 Record the decision with `/adr`: page stories are versioned by what
      they render, and only a person approves one. It covers design.md decisions
      1, 2, 5, 6, 7 and 8; what the approvals record proves (who committed an
      approval) and what it cannot (that a person looked); that signed commits
      are the stronger evidence and are not required; what the append-only check
      does not see (a rewrite pushed to `main`, a record and history edited
      together); and the relation to ADR 0023: Seen says what a change looked at,
      an approval is the standing record against a version. Verify: `just
      adr-index` regenerates `.claude/rules/decisions.md` and `just
      adr-index-check` passes.
- [ ] 1.3 Add one line to `CLAUDE.md`: an agent never runs `just approve` and
      never writes to `apps/storybook/canvas/approvals.jsonl`; approvals are a
      person's, made in their own terminal and committed on their own. Verify:
      the line is there and names both the command and the file.

## 2. The version

- [ ] 2.1 Test-first (`/tdd`), add `normaliseMarkup(html)` to
      `apps/storybook/canvas/lib.js` to design.md decision 3. Verify: unit tests
      in `lib.test.js` for each rule, one failing test written and seen failing
      before each: hydration comments gone; whitespace collapsed; two renders
      whose generated ids differ normalise the same while `for` still points at
      its `id`; attribute order ignored; a changed word differs; a changed scoped
      class differs; inline `style` kept.
- [ ] 2.2 Set `compilerOptions.cssHash` in `apps/storybook/svelte.config.js` to
      hash the component's CSS. Verify: change one declaration in a component's
      `<style>`, render, and the scoped class in the markup differs; revert and
      it returns; `just check apps/storybook` passes with all play functions
      green, since none may depend on a class name.
- [ ] 2.3 Grep `packages/ui/src` outside stories and tests for `Date.now`, `new
      Date()`, `Math.random`, `matchMedia` and `innerWidth`. Give any page story
      that reaches one a fixed input. Verify: the findings and what was done
      about each are listed in this task's commit message; `TransactionsScreen`
      stories all pass `today`.
- [ ] 2.4 Record a digest per page story in the headless run: an `afterEach` in
      `.storybook/preview.ts` that does nothing outside vitest and nothing for a
      story not under `Pages/`, and a browser command in `vitest.config.ts`
      appending `{ id, digest }` to `node_modules/.cache/wimm-canvas/run.jsonl`,
      emptied at the start of each run. Verify: after `just test apps/storybook`
      the run file holds exactly one line per page story in Storybook's index
      (112 today) and none for a molecule.
- [ ] 2.5 Test-first, write `apps/storybook/scripts/versions.mjs`, which writes
      and has no compare mode, to design.md decisions 4 and 5, its file shapes exported
      once from `lib.js`. Verify, each with a committed fixture under
      `apps/storybook/scripts/fixtures/` and each asserting **both** a non-zero
      exit **and** that `versions.json` is byte-identical afterwards: a run file
      missing a page story; a run file with a story twice; a digest that is not
      64 hex; a run marked failed. And the behaviours: an unchanged run writes
      nothing, leaves the file's bytes and mtime alone and prints
      "versions.json unchanged"; a changed run names the stories that moved and
      prints "versions.json changed, commit it"; a new fingerprint keeps the replaced version only when
      the approvals record approves it; a story gone from the index is dropped
      without approvals and kept `gone` with them; `firstSeen` never moves for an
      unchanged fingerprint.
- [ ] 2.6 Wire it: `gen` and `check` in `apps/storybook/justfile` both run
      `versions.mjs` after the test run, chained with `&&`. Regenerate and commit
      `versions.json`. Verify: `just check apps/storybook` twice in a row leaves
      `git status` clean the second time; rewording one sentence in
      `Overview.svelte` and running `just check apps/storybook` **passes**, names
      the Overview stories as moved, and leaves a diff in `versions.json`
      touching only those stories; breaking one page story's play function makes
      the check fail and leaves `versions.json` byte-identical.
- [ ] 2.7 Prove determinism and cost. Verify: two regenerations on one tree are
      identical; one under `TZ=Pacific/Auckland` is identical; one on another
      date is identical, by `faketime` if devbox offers it and otherwise by
      running the same commit the next day, with which one recorded here; `just
      check apps/storybook` timed three times before and after this group, both
      medians written here, and the difference under two seconds.

## 3. The approvals record

- [ ] 3.1 Test-first, add `approvalProblems(lines, versions)` to `lib.js`:
      unparseable line, unknown or missing field, fingerprint not 64 hex, time
      not ISO, story not in the versions file, fingerprint not in that story's
      versions, one person approving one version twice. Verify: one unit test
      per refusal asserting the message names the line number and the reason, and
      one passing case with two people on one version.
- [ ] 3.2 Test-first, write `apps/storybook/scripts/approvals.mjs --check`:
      `approvalProblems`, then the append-only comparison of design.md decision 7
      against `HEAD` and, where it can be reached, the merge base with
      `origin/main`. Verify, in temporary git repositories built by the test: an
      edited committed line fails naming it; a removed line fails; a line
      inserted in the middle fails; an appended line passes; a file never
      committed passes and prints that the comparison was skipped and why; no git
      at all passes the same way.
- [ ] 3.3 Add an empty `apps/storybook/canvas/approvals.jsonl` and run
      `approvals.mjs --check` from `check` in `apps/storybook/justfile`. Verify:
      `just check apps/storybook` passes; hand-appending a line for a story that
      does not exist makes it fail naming the line.
- [ ] 3.4 Test-first, `approvals.mjs --approve <story> [note]` and the root
      recipe `approve story *note`, to design.md decision 8, with the words of
      canvas.md's Command line section. The regenerate step is injected and the
      tests stub it. Verify, each asserting the refusal
      **and** that `approvals.jsonl` is byte-identical afterwards: `CLAUDECODE`
      set; stdin not a terminal; unknown story, with the closest ids offered; a
      story with no version; the regenerate step moving the story's implemented
      version, which refuses and names the new one; the same person and version
      twice; a `kind-behaviour` story; no git identity; answering `no`. And with
      a pseudo-terminal answering `yes`: exactly one line is appended, nothing
      before it moves, and the recorded name, email and fingerprint are right.
      The agent implementing this runs the tests and never the recipe against
      the real record.
- [ ] 3.5 Confirm a changed page never blocks. Verify: with one story approved in
      a fixture checkout, change its screen and run `just check apps/storybook`
      with no separate regenerate: it passes, and `statusOf` for the story gives
      `changed` with the approved version and a different implemented one.

## 4. The canvas

- [ ] 4.1 Test-first, add to `lib.js`: `statusOf` returning `{ status,
      implemented, approved }` so both versions travel together, `historyOf` with
      its `Implemented now` and `Latest approved` marks, `needsApproval`,
      `flowStatus` counting approved, changed and never, and the word functions `versionWords`, `badgeWords`,
      `filterWords`, `flowWords`. Write the fixture pairs canvas.md names under
      `apps/storybook/canvas/fixtures/`. Verify: a unit test for every row of
      canvas.md's state table and every string under Words that a function
      returns, including the three badge cases (one version when approved and
      implemented match, both when they differ, implemented alone when never
      approved), a history entry carrying both marks, a flow reading `7 of 12
      approved · 3 changed since approval · 2 never approved` and one with no
      changed part, two approvers, three or more, a behaviour story being
      exempt, and a flow counting a story once when it is both a step and a
      branch target elsewhere.
- [ ] 4.2 `approvals-ui.js`: functions returning the badge, the version line, the
      history button and the history panel, doing nothing on import. Verify:
      `apps/storybook/canvas/approvals.browser.test.js` holds every contract in
      canvas.md: focus to the heading on open, Escape and Close return focus to
      the button, one panel at a time, shortcuts inert while focus is inside, the
      status readable with the stylesheet removed, the polite live count, and
      `Copy` putting exactly the command on the clipboard.
- [ ] 4.3 Wire `canvas.js` and `canvas.css`: fetch `./versions.json` and
      `./approvals.jsonl` beside `index.json`, or the fixture pair under
      `?data=<name>`; draw the version and badge on each artboard's label and
      the history button; the `Needs approval` filter under the find box, kept as
      `?needs=1`, working with find; the mark beside each listed story; the count
      on each flow in the sidebar and on its heading. Colours from the feedback
      tokens only. Verify: `just check apps/storybook` passes; with
      `?data=missing` every artboard draws, no badge shows and the status bar
      reads the unreadable message; confirm in the running dev server that
      editing `approvals.jsonl` on disk is served on the next reload without a
      restart.

## 5. The look

- [ ] 5.1 Open the canvas with each fixture pair (`?data=approved`,
      `approved-by-two`, `changed`, `never`, `never-some`, `mixed`, `history`,
      `history-long`, `missing`) on a screen view and on a flow view, in light and dark, with the
      sidebar open and closed, and at a narrow and a wide window. Change what is
      wrong, then write what was seen and what changed into canvas.md under Seen.
      The four regimes are the artboards' own sizes, which this change does not
      alter, so Seen names the fixtures and the two window widths instead and
      says so. Verify: Seen names every fixture in both views.

## 6. Close

- [ ] 6.1 Update `README.md`: what a version is and when it moves, how to read
      the badge's two versions, approved and implemented, that `just check
      apps/storybook` regenerates `versions.json` and can leave a diff to commit,
      `just approve <story-id> [note]` in your own terminal and committed on its
      own, that an agent never runs it, that two people's approvals on two
      branches merge by keeping both lines, and `?data=<name>` for looking at a
      state. Verify: every command written there was run once as written.
- [ ] 6.2 `just check apps/storybook` and `just check packages/ui` pass, and `git
      status` shows `approvals.jsonl` empty and unmodified: no approval was
      recorded by this change.
