## S1 · See what a changed page looked like when it was approved

**As a** person who designs and builds wimm's screens
**I want** to switch a changed page on the design canvas between how it looks now
and how it looked when it was last approved
**so that** I can see what moved without remembering the old page, and decide
whether to approve the new one.

### INVEST

- **Independent** — No. It needs `version-and-approve-page-stories` applied: the
  approve command and the "Changed since approval" status. Accepted: this is the
  missing half of that status, and it ships after it.
- **Negotiable** — Partly. Pictures taken on approval is a solution. Accepted:
  rebuilding the old version on demand takes minutes and depends on old commits
  still building.
- **Valuable** — Yes. After a change, a person compares the page with its
  approved look on one artboard instead of from memory.
- **Estimable** — Yes. The approve run already renders every page story
  headless; taking a picture there is the known part.
- **Small** — Yes. One picture set per story, one switch, one check.
- **Testable** — Yes. Approve a page, change it, and switch.

### Capabilities

`design-canvas/last-approved-look`

### Satisfied by

- `design-canvas/last-approved-look`: Requirement: Approving a page keeps a picture of the version approved
- `design-canvas/last-approved-look`: Requirement: Only the last approved version keeps pictures
- `design-canvas/last-approved-look`: Requirement: A changed artboard switches to its last approved look
- `design-canvas/last-approved-look`: Requirement: A picture that does not match its approval is refused

## S2 · Approve many pages in one run

**As a** person who designs and builds wimm's screens
**I want** to approve several page stories, or every one that needs a look, in
one run of the approve command
**so that** starting approvals on an existing app, or approving a batch after a
change that moved many pages, takes one run and one answer instead of one per
page.

### INVEST

- **Independent** — No. It extends the approve command this change already
  edits. Accepted: splitting it out would edit the same code twice.
- **Negotiable** — Yes. Named stories and a flag for what needs a look is one
  shape; a baseline-only flag was the smaller one.
- **Valuable** — Yes. 132 page stories need a first approval, and one approval
  per run takes about an hour of full test runs.
- **Estimable** — Yes. The run already renders every story; it pictures a list
  instead of one.
- **Small** — Yes. One command shape, one prompt, one record append.
- **Testable** — Yes. Approve three stories, and approve with `--needs`.

### Capabilities

`design-canvas/page-approvals`

### Satisfied by

- `design-canvas/page-approvals`: Requirement: A person approves the implemented version of a page story
- `design-canvas/page-approvals`: Requirement: An approval that cannot be true is refused
