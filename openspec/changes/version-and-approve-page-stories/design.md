## Context

See proposal.md for why. What shapes the approach:

- The design canvas is plain files under `apps/storybook/canvas/`, served beside
  Storybook by one `staticDirs` entry (`.storybook/main.ts`), so it shares an
  origin with `iframe.html`. It has no build step, no server of its own, and
  cannot write a file. Its logic is pure functions in `lib.js` with node unit
  tests (`lib.test.js`, `flows.test.js`) and one browser test
  (`pointing.browser.test.js`).
- `just check apps/storybook` already renders every story once, headless, in the
  `storybook` vitest project (307 stories in about 20 s), and `flows.test.js`
  already builds Storybook's real index from node in about 1.2 s.
- Page stories carry one `kind-*` tag; `lib.js` exports `KINDS`, `kindOf` and
  `kindProblems`. Flows and their branches are in `flows.js`.
- Svelte's scoped class is `svelte-<hash of the file name>` by default
  (`svelte/src/compiler/validate-options.js`). So rendered markup does **not**
  move when a component's own `<style>` changes. This was checked, and it
  decides the first decision below.
- The folder `apps/storybook/canvas/` is untracked. Nothing here is safe to
  build until it is committed.
- `.claude/rules/generated-artifacts.md` binds the versions file: validate before
  writing, one source of truth for paired checks, refusal tests that assert the
  file did not move, name the guarantee precisely, one command shape that runs.

## Language

- **implemented version** — the version a story renders now. `implemented` in the
  versions file, "Implemented" on the canvas. Never "current", "live" or "latest".
- **approved version** — the newest version of a story that holds an approval.
  "Approved" on the canvas. It is the implemented version or an earlier one.
- **version** — the fingerprint of what one page story renders. Shown as its
  first seven characters and the date first seen. Never "revision", "hash",
  "snapshot" or "build".
- **fingerprint** — the full digest a version is the short form of. Used in code
  and in the versions file, never on the canvas.
- **approval** — one person's record that they agree to one version of one story.
  Never "sign-off", "review" or "LGTM".
- **approvals record** — `approvals.jsonl`. Never "log" or "database".
- **versions file** — `versions.json`.
- **status** — one of `approved`, `changed`, `never`, `exempt`. On the canvas:
  "Approved", "Changed since approval", "Never approved", and nothing for exempt.
  Never "stale", "dirty", "pending" or "outdated".
- **history** — a story's implemented version and its approved earlier ones.
- **needs approval** — status `changed` or `never`.
- **regenerate** — what `just gen apps/storybook` does to the versions file.

## Goals / Non-Goals

**Goals:**

- A version that moves when the look can move and not otherwise, identical on
  two runs and on two days.
- An approvals record whose every line is one person's act, that only grows, and
  that `just check` can hold to that.
- A canvas that shows both versions, approved and implemented, everywhere a
  status appears, from a versions file every check run brings up to date.
- A visible change never fails `just check`.
- No server, no database, no new dependency.

**Non-Goals:**

- Proving a person looked. The record proves who committed an approval.
- Approving from the canvas. A static page cannot write; it offers the command
  to copy.
- A version per size, per theme or per density (decision 2).
- Pixel comparison. ADR 0020's rejection of it stands: a tolerance loose enough
  for two text engines is loose enough to miss what matters.
- Keeping every version a page ever had (decision 5).
- Showing what the approved version looked like. It needs an image stored with
  each approval; a possible later change.
- Blocking anything on approval status, in `just check` or in CI.
- Retiring or transferring approvals when a story is renamed. A renamed story is
  a new story that nobody has approved, which is the honest reading.

## Decisions

### 1. A fingerprint is normalised markup, content-hashed component styles, and a digest of the global styles

`fingerprint = sha256(normalised markup of the story root + "\n" + global digest)`.

- **Markup** is `outerHTML` of the story's root element, taken after the play
  function has finished, because that is the state the canvas draws.
- **Component styles ride in the markup.** `apps/storybook/svelte.config.js` sets
  `compilerOptions.cssHash` to hash the component's CSS rather than its file
  name, so a scoped class changes exactly when that component's styles do.
  Without this a padding change is invisible to the fingerprint, which was the
  one way the proposed "markup plus tokens" would have approved pages that then
  moved. It applies only to the Storybook app; `apps/web` compiles as before.
- **Global digest** is one sha256 over `packages/ui/src/tokens.css`, `base.css`,
  `fonts.css` and the bytes of every file under `packages/ui/src/fonts/`, in
  path order, computed once per run in node.

*Alternative: computed styles of every element.* Rejected: it multiplies the run
by the number of elements and by every size, and differs between browsers.
*Alternative: hash the source files a story imports.* Rejected: every refactor
would read as a change, which is the opposite of the requirement.

### 2. One version per story, not one per size

With decision 1 in place a size adds nothing: the markup is the same at every width
unless a component branches on width in script, and every media query is inside
styles the fingerprint already carries. Rendering each story at two sizes would
double the headless run to buy that one case. So a version covers the story at
every size, a person approves after looking at the sizes the canvas draws, and
each artboard of a story shows the same status.

The one case this misses, script that branches on width, is covered where it
exists by its own story (`Compact` stories set their viewport), and a task greps
`packages/ui/src` for `matchMedia` and `innerWidth` and records what it finds.

### 3. Normalisation, and proving it

Stripped or rewritten before hashing, by one pure function in `lib.js`:

- Svelte's hydration and block comments (`<!---->`, `<!--[-->`, `<!--]-->`,
  `<!--[!-->`).
- Runs of whitespace between tags, collapsed to nothing; runs inside text,
  collapsed to one space.
- Generated ids: any `id`, `for`, `aria-labelledby`, `aria-describedby`,
  `aria-controls` and `href="#..."` value is replaced by its order of first
  appearance (`id-1`, `id-2`), so the wiring is kept and the counter is not.
- Attributes Storybook or the test harness adds to the root.
- Attribute order, sorted by name.

Kept: scoped classes (decision 1), inline `style`, every word.

Time is not normalised away; it is pinned. `TransactionsScreen` defaults `today`
to the clock, and its stories already pass `today`. A task greps `packages/ui/src`
for `Date.now`, `new Date()` and `Math.random` outside stories and tests, and any
page story that reaches one is given a fixed input. `NavigationProgress` uses
`Date.now` for timing only and renders nothing from it.

Determinism is asserted, not assumed: a task regenerates twice and diffs, and
regenerates once under `TZ=Pacific/Auckland` with the system date moved a day by
`faketime` if devbox has it, otherwise by running the same tree the next day and
recording that in the task.

### 4. The headless run records; `gen` and `check` both write

- `preview.ts` gains an `afterEach` that, only under vitest and only for
  `Pages/` stories, normalises the root's markup and hands
  `{ id, markup digest }` to a vitest browser **command** declared in
  `vitest.config.ts` (`test.browser.commands`). The command runs in node and
  appends to a run file under `node_modules/.cache/wimm-canvas/`. In the
  Storybook UI and on the canvas the hook does nothing.
- `apps/storybook/scripts/versions.mjs` reads the run file, builds the real index
  with `buildIndex` (as `flows.test.js` does), adds the global digest, and
  writes. It validates before it writes: the run exited clean, every page story
  in the index has exactly one record, every digest is 64 hex characters.
  Anything else, it names every problem, exits non-zero and touches nothing.
- `gen` and `check` in `apps/storybook/justfile` both run it after the test run,
  chained with `&&` so a failed run cannot be followed by a write. There is no
  compare mode: **a check run never fails because the file was out of date, it
  brings it up to date.** So a check run can leave a diff in `versions.json` to
  commit. It prints the stories whose version moved and "versions.json changed,
  commit it", or "versions.json unchanged". Regenerating with nothing visible
  changed is a no-op, and a test holds that.
- Cost is one `outerHTML` and one digest per page story, about 112 of them. A
  task measures `just check apps/storybook` before and after and records both;
  more than two seconds added is a finding to fix, not to accept.

*Alternative: a second, dedicated render pass.* Rejected: it renders everything
twice to avoid six lines in a hook.

### 5. The versions file holds the implemented version and the approved ones

```json
{
  "pages-overview--populated": {
    "implemented": "a3f9c21e...64 hex",
    "versions": [
      { "fingerprint": "a3f9c21e...", "firstSeen": "2026-09-20" },
      { "fingerprint": "77b0d4aa...", "firstSeen": "2026-09-12" }
    ]
  }
}
```

Keys sorted; `versions` newest first. On regenerate, for each story:

- same fingerprint as `implemented`: nothing changes, including `firstSeen`. This is
  what makes regeneration a no-op.
- a new fingerprint: it becomes `implemented` with today's date. The version it
  replaces is kept only if the approvals record holds an approval for it;
  otherwise it is dropped. Twenty edits in an afternoon leave one entry, and git
  keeps the rest for anyone who wants them.
- a story gone from the index: dropped if it has no approval, kept with
  `"gone": true` and no `implemented` if it has, so the approvals record never names
  a story the versions file has never held.

`firstSeen` is a date that is evidence, which is the exception the repo's own
rule for dates allows. It is written only when a fingerprint is new, so it never
makes a no-op run write.

The generator therefore reads the approvals record, and the validator reads the
versions file: one direction each, from one exported description of both shapes
in `lib.js`, which is the paired-check rule.

### 6. The approvals record is JSON Lines, beside the canvas

`apps/storybook/canvas/approvals.jsonl`, one record per line:

```json
{"story":"pages-overview--populated","fingerprint":"a3f9c21e...","name":"Emanuel Fonseca","email":"...","at":"2026-09-20T18:04:11Z","note":"after the merchant rows were tightened"}
```

One line per approval makes an append a one-line diff and makes "changed other
than by appending" a prefix comparison. It lives in the canvas folder because
that folder is already served at `/canvas/`, so the page fetches
`./versions.json` and `./approvals.jsonl` with no second `staticDirs` entry. A
task confirms Storybook's dev server serves a static file as it is on disk now
rather than as it was at start.

*Alternative: `design/approvals.jsonl` beside `tokens.json`.* Rejected for now:
it needs a second static mount for one file, and `design/` holds the contract CI
owns, which an approvals record is not. Moving it later is one path in two places.

The email is recorded because a name alone is not an identity. It is already in
every commit, so the record publishes nothing git does not.

### 7. Append-only is checked against the committed copy, and says what it cannot see

`scripts/approvals.mjs --check`, run by `just check apps/storybook`:

1. Content: every line parses, has exactly the known fields, a 64-hex
   fingerprint, an ISO time; the story is in the versions file; the fingerprint
   is in that story's `versions`; no person approves one version twice.
2. Append-only: the copy at `HEAD` (`git show HEAD:<path>`) must be a byte prefix
   of the file on disk. Where `origin/main` exists and `HEAD` is not on it, the
   copy at `git merge-base HEAD origin/main` must be a prefix of the copy at
   `HEAD` too, which is what catches a rewriting commit in a branch.
3. Where the file is not in `HEAD`, there is no git, or the clone is too shallow
   to reach the merge base, step 2 is skipped and the check prints that it was
   skipped and why. It does not fail, because a fresh clone is not a defect.

What this does not cover, stated in the ADR: a rewrite pushed straight to `main`,
and anyone who edits the record and the git history together.

### 8. `just approve` is the only writer, and it fails closed without a person

Root recipe `approve story *note`, running `scripts/approvals.mjs --approve`:

- refuses when `CLAUDECODE` or `CLAUDE_CODE_ENTRYPOINT` is set, or when stdin is
  not a terminal. An agent's shell has no terminal, so the common case fails
  closed rather than relying on the agent having read a rule.
- regenerates first, by the same chain `gen` runs. If that changes the story's
  implemented version it refuses, names the new version and asks the person to
  look again, so an approval is never recorded against a version that is not
  what renders and not what they looked at. The regenerate step is injected, so
  the tests stub it and do not pay for a headless run each.
- prints the story's title and name, the short version and its first-seen date,
  and asks `Approve this version? yes/no`. Anything but `yes` records nothing.
- reads `git config user.name` and `user.email`; refuses if either is empty.
- builds the whole new file in memory, validates it as step 1 of decision 7
  would, then appends. It never rewrites.

This is a guard, not proof. A person can set the variables aside and an agent
could write the line by hand, which is why `CLAUDE.md` forbids both and why the
check in decision 7 only proves shape. Signed commits are the stronger evidence
and are named in the ADR as available, not required.

A person working inside an agent session runs it in their own terminal. The `!`
prefix runs in the agent's shell and is refused, deliberately.

### 9. Status is one pure function

`statusOf(entry, approvals, kind)` in `lib.js` returns `{ status, implemented,
approved }`. `implemented` is the implemented version with its first-seen date,
always. `approved` is the approved version with its approvals, or nothing.
`status` is `exempt` for `kind-behaviour`, `approved` when the two are one
version, `changed` when they differ, `never` when there is no approved version.
Every place a status is drawn takes both versions from here, so none can show
one without the other. The badge, the filter, the sidebar marks and `flowStatus(flow, ...)`
all call it, so a page cannot be approved in one place and changed in another.

`flowStatus` counts the flow's steps and branch targets that need approval, once
each, and how many are `approved`, `changed` and `never`.

### 10. The canvas reads both files once, and says when it cannot

`main()` fetches both beside `index.json`. If either is missing or does not
parse, the canvas draws as it does today with no badges, and the status bar says
"Versions could not be read. Run just gen apps/storybook." A canvas that cannot
know does not guess "Never approved".

The history panel is one `<dialog>` owned by the canvas page, opened
non-modally from a button on the artboard's label, closed with Escape, focus
returned to that button. The label is outside the iframe and outside the cover,
so it takes a click without Alt.

### 11. The ADR and the rule

One ADR, next free number, titled for the decision: page stories are versioned by
what they render, and only a person approves one. It records decisions 1, 2, 5,
6, 7 and 8, the honest limit, and that an approval is what ADR 0023's look leaves
behind: Seen in `canvas.md` still says what was looked at and changed during a
change; the approval is the standing record against the version. It supersedes
nothing.

`CLAUDE.md` gains one line: an agent never runs `just approve` and never writes
to `apps/storybook/canvas/approvals.jsonl`; approvals are a person's, made in
their own terminal and committed on their own.

## Risks / Trade-offs

- [A check run can leave a diff to commit, and a commit made without running the
  check carries a versions file that is behind] → Accepted, by decision: a visible
  change never fails the check. Until the next check run the canvas shows the
  older implemented version, which can read "Approved" over a page that moved.
  The repo's stop hook runs the check on every touched package, CI's run prints
  the stories that moved, and the diff of `versions.json` is the list of pages
  whose look changed, a review artifact the repo lacks.
- [Content-hashed scoped classes mark a page changed when a style rule changes
  that does not apply to it] → Over-reporting is the safe direction. It costs a
  look, never a wrong "Approved".
- [A global change marks all 82 approvable stories changed at once] → Correct,
  and the filter plus the flow roll-up are how a person works through them.
- [Two people approving on two branches both append to the record's end] → A
  merge conflict git resolves by keeping both lines; the check then passes. The
  README says so.
- [The normaliser misses a source of noise and versions flap] → The run-twice
  task and the other-day task catch it before anything is approved, and a flap
  after that shows as a `versions.json` diff with no source change, which is
  loud.
- [The agent guard is a convention wearing a check] → Said plainly in the ADR.

## Migration Plan

1. Commit `apps/storybook/canvas/` as it stands.
2. Land the generator with an empty approvals record; regenerate; commit
   `versions.json`. Every page reads "Never approved", which is true.
3. Approvals accumulate as people look. Nothing else depends on them.

Rollback is deleting the two files and the hook; no other package reads them.

## Open Questions

- Whether `faketime` is in the devbox index. If not, the other-day determinism
  proof is a recorded manual step. It changes no spec and no task's outcome.
