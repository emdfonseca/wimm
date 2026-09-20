## Why

The design canvas shows every screen of wimm, and nothing on it says whether a
person has looked at what is drawn and agreed to it. ADR 0023 made the rendered
story the design of record and a look the gate before wiring, and the only trace
that look leaves is prose under Seen in a change's `canvas.md`, which archives
with the change and says nothing about the page as it renders today. A page can
change after it was looked at and the canvas shows it exactly as before.

## What Changes

- **Every page story has a version**: a fingerprint of what the story renders,
  computed in the headless Storybook run and written to a generated, committed
  versions file. A refactor that changes nothing a person can see keeps the
  version; a change to the markup, to a component's styles or to the tokens
  makes a new one.
- **A person approves a version** with `just approve <story-id> [note]`, which
  appends one record to a committed approvals file: the story, the version, who,
  when, and the note. It refuses a story that does not exist, a story with no
  implemented version, and a version that person already approved.
- **The canvas shows both versions of every page story**, the one a person
  approved and the one implemented now, wherever a status appears: one version
  when they are the same, both when they differ, the implemented one alone when
  nothing was ever approved. On each artboard; in a history that marks which
  entry is implemented now and which is the latest approved; in a sidebar filter
  for the stories that need a look; and rolled up onto each flow, which counts
  the pages implemented at a version other than the one approved.
- **A visible change never fails `just check`.** The check regenerates the
  versions file and says when it left a change to commit. What does fail it: an
  approvals file that is malformed, that names a story the versions file has
  never held, that names a version the story never had, or that was changed other
  than by appending; and a run too broken to write versions from.
- **Only a person approves.** `CLAUDE.md` gains the rule that an agent never
  runs `just approve`, the command refuses to run where no person is at the
  keyboard, and an ADR records the record, the rule, what the file proves and
  what it cannot, and how this formalises the look of ADR 0023.
- `kind-behaviour` stories are versioned and need no approval: they assert
  something and show nothing a state does not.

## Capabilities

### New Capabilities

- `design-canvas/page-approvals`: the version of a page story, a person's
  approval of one, the status the canvas shows for it, its history, and what
  `just check` refuses about the two files.

### Modified Capabilities

None. No existing capability covers the design canvas; the two earlier canvas
changes recorded no spec because they added a viewer and no rule. This one adds
a rule that outlives the change, which is what a spec is for.

## Impact

- `apps/storybook/canvas/`: `versions.json` (generated) and `approvals.jsonl`
  (appended by `just approve`), both served beside the canvas page; `lib.js`
  gains the fingerprint normaliser, the approvals validator and the status
  function; `canvas.js` and `canvas.css` gain the badge, the history panel, the
  filter and the flow roll-up. The folder is untracked today and is committed
  before anything else.
- `apps/storybook/.storybook/preview.ts`, `vitest.config.ts`, `svelte.config.js`
  and `justfile`: the headless run records a fingerprint per page story, and
  both `gen` and `check` write the versions file from it, refusing to write from
  a run that failed or is incomplete.
- Root `justfile`: `approve`. `CLAUDE.md`: the rule. `README.md`: how to read a
  badge and how to approve. `docs/decisions/`: one new ADR, and 0023's index
  entry gains the relation.
- Out of scope: showing what the approved version looked like. That needs an
  image stored with each approval and is a possible later change.
- No new dependency. No change to `packages/ui` components, to `apps/web` or to
  `apps/wimm`. Page stories are read, not edited.
