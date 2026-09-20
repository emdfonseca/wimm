The two earlier design canvas changes skipped this artifact: they added a viewer,
and nobody is served by a viewer existing. This one is different in one respect
that earns stories. An approval is an act only a person may perform, so the
change has an actor whose judgement is the whole point, and a second person who
is asked for theirs. Neither is "a developer" wanting a mechanism.

## S1 · Know whether a page has been looked at and agreed to

**As a** person who designs and builds wimm's screens
**I want** each page on the design canvas to show the version a person approved
beside the version that is implemented, or say that nobody approved one
**so that** I stop re-checking pages nobody touched and stop trusting pages that
quietly moved after someone agreed to them.

### INVEST

- **Independent** — No. It needs S2's approvals to show anything but "Never
  approved". Accepted: versions with no way to approve one is half a feature,
  and they ship together.
- **Negotiable** — Yes. It states the three things to know, not where they are
  drawn.
- **Valuable** — Yes. After a token change the canvas names the pages whose
  look moved, which today means opening all of them.
- **Estimable** — Yes, once design.md settles how a fingerprint is made
  deterministic. That was the unknown and it is decided there.
- **Small** — No. It covers the version, the badge, the filter and the flow
  roll-up. Accepted: a version nobody can see is not a deliverable, and the
  filter and roll-up are the same status drawn in two more places.
- **Testable** — Yes. Each status has a scenario a person can reach by changing
  a page and looking.

### Capabilities

`design-canvas/page-approvals`

### Satisfied by

- `design-canvas/page-approvals`: Requirement: Every page story has a version
- `design-canvas/page-approvals`: Requirement: A version changes when what a person sees can change, and only then
- `design-canvas/page-approvals`: Requirement: Each artboard says where its page stands
- `design-canvas/page-approvals`: Requirement: The stories that need a look can be found
- `design-canvas/page-approvals`: Requirement: A flow says where its pages stand
- `design-canvas/page-approvals`: Requirement: A changed page never fails the check
- `design-canvas/page-approvals`: Requirement: The versions file is what renders

## S2 · Approve the version I looked at

**As a** person who designs and builds wimm's screens
**I want** to approve the version of a page I have just looked at, with a note if
I have one, in one command
**so that** my look is on record against exactly what I saw, under my name, and
not as a sentence in a change that gets archived.

### INVEST

- **Independent** — Yes against S3 and S4. It needs S1's version to exist.
- **Negotiable** — Partly. "One command" is a solution. Accepted: the canvas is
  a static page that cannot write a file, so the alternative is a server, and
  the proposal rules one out.
- **Valuable** — Yes. An approval names the version; Seen in `canvas.md` never
  could.
- **Estimable** — Yes.
- **Small** — Yes.
- **Testable** — Yes. The command's refusals are each a scenario.

### Capabilities

`design-canvas/page-approvals`

### Satisfied by

- `design-canvas/page-approvals`: Requirement: A person approves the implemented version of a page story
- `design-canvas/page-approvals`: Requirement: An approval that cannot be true is refused
- `design-canvas/page-approvals`: Requirement: Only a person approves

## S3 · Be asked for my approval and give it

**As a** second member of the household, asked whether a screen that shows our
money reads right to me
**I want** to give my own approval of the same version, separately from the
person who built it
**so that** a page can show that both of us agreed to it, and which of us has
not yet.

### INVEST

- **Independent** — Yes. One approver works without it.
- **Negotiable** — Yes.
- **Valuable** — Yes. The household-money and sharing screens are the ones where
  the builder is the wrong only judge.
- **Estimable** — Yes.
- **Small** — Yes. It is S2 with a second name and the badge listing names.
- **Testable** — Partly. The scenarios prove two names are recorded and shown.
  Nothing proves the second person looked, which is the limit the ADR states.

### Capabilities

`design-canvas/page-approvals`

### Satisfied by

- `design-canvas/page-approvals`: Requirement: More than one person can approve a version
- `design-canvas/page-approvals`: Requirement: Only a person approves

## S4 · See how a page got here

**As a** person who designs and builds wimm's screens
**I want** the history of a page: each version a person approved, who approved it
and when, which of them is the latest approved, and which version is implemented now
**so that** when a page reads "Changed since approval" I can tell what it was
last agreed as, and by whom, before deciding whether to look again or ask.

### INVEST

- **Independent** — Yes against S3. It needs S1 and S2.
- **Negotiable** — Yes.
- **Valuable** — Yes.
- **Estimable** — Yes.
- **Small** — Yes.
- **Testable** — Yes.

### Capabilities

`design-canvas/page-approvals`

### Satisfied by

- `design-canvas/page-approvals`: Requirement: A page story's history can be read
- `design-canvas/page-approvals`: Requirement: The record of approvals only grows
