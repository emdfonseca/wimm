## 1. Carry the copy check before removing it

- [x] 1.1 `/tdd`: `check-stories.py` refuses a story under `pages/` that has
      no play function, naming the file and the story. Verify:
      `test-stories.py` feeds it a page story without one and asserts the
      refusal and its reason, and a page story with one passes.
- [x] 1.2 Add play functions to the eight page stories that lack one
      (`AccountsScreen` 1, `ChooseAccountsScreen` 3, `ChooseBankScreen` 1,
      `SettingsScreen` 1, `TransactionsScreen` 1, `WidenConsentScreen` 1),
      each asserting the words of its state. Verify: `just check packages/ui`
      and `just check apps/storybook` pass.
- [x] 1.3 Read `src/canvas-contract.json` against the page stories and list,
      per screen, every notice title, notice body and failure message no play
      function asserts. Add the assertion or record why not in this change's
      design.md under a Coverage heading. Verify: the list is in design.md
      with no entry left undecided, and `just check apps/storybook` passes.

## 2. The schema and the project instructions

- [x] 2.1 Rewrite the `canvas` artifact in `openspec/schemas/wimm/schema.yaml`
      and `templates/canvas.md` as a state plan: state stories to write, the
      fixture and the words of each, the flow they join, components used from
      `packages/ui/src`, components missing as two deliverables with the
      search-by-data rule, contracts, and a Seen section filled in by the look
      task. Keep the copy rules, draw-what-exists-now, the layer rule and the
      conditional skip. Reword the schema description and the stories line
      that mention the canvas. Verify: no `pen`, `frame`, `journey` or
      `library` remains in the artifact, and `just openspec instructions
      canvas` on a scratch change prints the new text.
- [x] 2.2 Rewrite the `tasks` instruction's UI order: components, then the
      presentational screen with its state stories, then a look task at all
      four regimes on the design canvas that writes Seen into canvas.md, then
      wiring, none sharing a task. A new token is an edit to
      `design/tokens.json` and `just gen`. Verify: `just openspec instructions
      tasks` prints it, and `just openspec validate --all --strict` passes.
- [x] 2.3 `openspec/config.yaml`: replace the pen paragraph in `context` and
      the apply and archive guidance about frames and the journey `.pen` with
      the design-of-record equivalents. `CLAUDE.md`: replace the canvas line,
      drop the two `.pen` lines, and keep the self-contained-repo rule without
      its pen half. Verify: `grep -n -i "pen" CLAUDE.md openspec/config.yaml`
      finds only "OpenSpec" and "open".
- [x] 2.4 Docs: delete `docs/design/canvas-audit.md` and
      `library-conventions.md`; reword `tokens.md` so `design/tokens.json` is
      the source of values; remove zone names from `surfaces.md`; read
      `project-setup-record.md`, remove its pen sections, and delete it if
      nothing else is left. Reword the pen mentions in
      `packages/ui/justfile`, `packages/ui/package.json`, `gen-tokens.py`,
      `check-tokens.py` and `check-geometry.py`. Verify: `grep -rn -i
      "\.pen\|pen\.dev\|lib\.pen" docs/design packages/ui --include="*.md"
      --include="*.py" --include=justfile --include=package.json` is empty,
      and `just check packages/ui` passes.
- [x] 2.5 `.claude/skills/storybook/`: reword the references to the
      `.lib.pen` and the Journey row in `SKILL.md:17,30` and the six reference
      files. Needs a write outside the sandbox; ask before running. Verify:
      `grep -rn -i "pen" .claude/skills/storybook` finds nothing about
      pen.dev.

## 3. Retire the frame checks

- [x] 3.1 Delete `check-canvas.py`, `gen-canvas-contract.py`,
      `test-canvas.py`, `test-contract.py` and `src/canvas-contract.json`, and
      remove them from `gen` and `test` in `packages/ui/justfile`. Verify:
      `just gen` and `just check packages/ui` pass, and `git grep -n
      "canvas-contract\|check-canvas\|gen-canvas"` finds nothing outside
      `openspec/changes/archive` and `docs/decisions`.

## 4. Delete the pen files and tooling

Each deletion below is destructive. Show the `git rm` command and wait for a
go-ahead before running it.

- [x] 4.1 Record the commit that last holds the pen files, for the ADR.
      Then `git rm` the five `apps/web/design/*.pen`,
      `packages/ui/design/product-ui.lib.pen` and `library-manifest.tsv`.
      Verify: `packages/ui/design/` holds `tokens.json` only, and `just check
      packages/ui` passes.
- [x] 4.2 `git rm` `bin/pen-exec`, `pen-import`, `pen-manifest`, `pen-save`,
      `pen-verify-tokens` and `test-pen-exec`; remove their six recipes from
      the root `justfile` and `pen-exec-test` from `ci`; remove the pencil
      server from `.mcp.json`. Verify: `just --list` shows no `pen-` recipe
      and `just ci` passes.
- [x] 4.3 Delete `.claude/skills/pen-design/`. Needs a write outside the
      sandbox; ask before running. Verify: the directory is gone and nothing
      under `.claude/` or `openspec/schemas/` names it.

## 5. The decision record

- [x] 5.1 `/adr`: 0023, code is the design of record. It supersedes 0020,
      0009, 0010 and 0011 whole, the drawing in 0008 and the library clause of
      0003; states that copy coverage is reduced to what play functions
      assert and that the play-function rule cannot see what is asserted; and
      names the commit from 4.1. Set the Status line of each superseded ADR
      following 0018's form. Verify: `just adr-index` regenerates
      `.claude/rules/decisions.md` and `just adr-index-check` passes.
- [x] 5.2 `openspec/specs/app/screen-layout/spec.md`: reword the Purpose so it
      no longer says a screen is finished when it matches the frame drawn. The
      requirements change through this change's delta at archive. Verify:
      `just openspec validate --all --strict` passes.

## 6. The change already in flight

- [x] 6.1 `openspec-update-change` on `make-overview-a-useful-dashboard`:
      canvas.md becomes a state plan under the new template; tasks 2.1, 2.5,
      2.7, 3.1 and 3.5 lose their library, generator and frame-map work; the
      proposal's pen mentions go. Verify: `grep -rn -i "pen\|frame\|IMPLEMENTS"
      openspec/changes/make-overview-a-useful-dashboard` finds nothing about
      pen.dev, and `just openspec validate make-overview-a-useful-dashboard
      --strict` passes.

## 7. Verification

- [x] 7.1 `just ci` passes. `git grep -n -i "pen-exec\|\.pen\b\|pen\.dev"`
      finds nothing outside `openspec/changes/archive` and `docs/decisions`.
      README has no pen mention.
