# design-canvas/page-approvals Specification

## Purpose

Lets a person see, on the design canvas, which version of each page story a
person approved, which pages changed since, and how each page got to where it
is, with the record kept in the repository and only ever added to.

## Requirements

### Requirement: Every page story has a version
**Story**: S1
Every page story SHALL have exactly one implemented version: the version of what
it renders now, shown to a person as a short identifier and the date that version
was first seen. One version covers the story at every size the canvas draws it.
A story that only asserts something is versioned like any other.

#### Scenario: A page story on the canvas
- **WHEN** a person opens a screen on the design canvas
- **THEN** every artboard of a story names the same implemented version, written
  as seven characters and the date first seen, such as "a3f9c21 · 20 Sep 2026"

#### Scenario: A new page story
- **WHEN** a page story is added and the versions are regenerated
- **THEN** the story has a version whose first-seen date is that day, and every
  other story's version and date are unchanged

#### Scenario: A page story that is removed
- **WHEN** a page story that was once approved is deleted and the versions are
  regenerated
- **THEN** it is drawn nowhere on the canvas, its approvals stay in the record,
  and the check still passes

### Requirement: A version changes when what a person sees can change, and only then
**Story**: S1
A story's version SHALL change when the markup it renders changes, when the
styles of a component it renders change, or when the design tokens, base styles
or fonts change. It SHALL NOT change when code is rearranged with no effect on
any of those, when the same tree is rendered again, or when it is rendered on
another day.

#### Scenario: A refactor that changes nothing visible
- **WHEN** a component is split into two files that render the same markup with
  the same styles, and the versions are regenerated
- **THEN** no story's version changes and the versions file is not modified

#### Scenario: A word on a screen changes
- **WHEN** one sentence on the Overview screen is reworded and the versions are
  regenerated
- **THEN** every Overview story showing that sentence has a new version, and no
  other screen's stories do

#### Scenario: A component's spacing changes
- **WHEN** the padding in one component's own styles changes and the versions
  are regenerated
- **THEN** every story rendering that component has a new version

#### Scenario: A token moves
- **WHEN** a colour token changes value and the versions are regenerated
- **THEN** every page story has a new version

#### Scenario: Rendering twice
- **WHEN** the versions are regenerated twice in a row with nothing changed
  between
- **THEN** the second run reports nothing to write and the file is not modified

### Requirement: A person approves the implemented version of a page story
**Story**: S2
A person SHALL be able to approve the implemented version of one page story with one
command, giving the story and optionally a note. The approval SHALL record the
story, the version approved, the name and email the person commits under, the
time, and the note, as one new line at the end of the approvals record, and
SHALL change nothing else in it. The command SHALL say what it recorded.

#### Scenario: Approving a page
- **WHEN** a person runs the approve command for the Overview's populated story
- **THEN** the command prints the story, the version it approved and their name,
  the approvals record is one line longer, and the canvas shows that story as
  approved by them once reloaded

#### Scenario: Approving with a note
- **WHEN** a person approves a story with the note "after the merchant rows were
  tightened"
- **THEN** that note appears beside the approval in the story's history

#### Scenario: The page changed since the person looked
- **WHEN** a person runs the approve command and the story now renders a version
  other than the one the versions file held
- **THEN** the command brings the versions file up to date, refuses, names the
  new implemented version, asks the person to look again, and the approvals
  record is not modified

### Requirement: An approval that cannot be true is refused
**Story**: S2
The approve command SHALL refuse, recording nothing and leaving the approvals
record unmodified, when the story does not exist, when the story has no current
version, when the same person has already approved that version, and when the
story only asserts something and needs no approval. Each refusal SHALL say which
of these it is.

#### Scenario: A story that does not exist
- **WHEN** a person runs the approve command with a story id that matches no
  page story
- **THEN** the command fails, names the id, suggests the closest ids that do
  exist, and the approvals record is not modified

#### Scenario: A story with no version yet
- **WHEN** a person runs the approve command for a story added since the
  versions were last regenerated
- **THEN** the command fails saying the story has no version yet and how to
  generate one, and the approvals record is not modified

#### Scenario: Approving the same version twice
- **WHEN** a person who has approved a story's implemented version runs the approve
  command for it again
- **THEN** the command fails saying they approved that version already and
  when, and the approvals record is not modified

#### Scenario: A story that only asserts something
- **WHEN** a person runs the approve command for a story of the behaviour kind
- **THEN** the command fails saying that kind of story needs no approval, and
  the approvals record is not modified

### Requirement: Only a person approves
**Story**: S2
An approval SHALL be recorded only by a person at a keyboard. The approve
command SHALL refuse to run inside an agent's session or where there is no
terminal to answer from, and SHALL ask the person to confirm the story and
version before recording. The repository's standing instructions SHALL forbid an
agent from running it or from writing to the approvals record by any other
means. The decision record SHALL state what the approvals record proves, who
committed an approval, and what it cannot, that a person looked.

#### Scenario: An agent runs the command
- **WHEN** the approve command is run from an agent's session
- **THEN** it fails saying approvals are recorded by a person in their own
  terminal, and the approvals record is not modified

#### Scenario: No terminal to answer from
- **WHEN** the approve command is run with its input piped from a file
- **THEN** it fails without asking anything, and the approvals record is not
  modified

#### Scenario: A person changes their mind
- **WHEN** a person runs the approve command and answers the confirmation with
  anything other than yes
- **THEN** nothing is recorded and the command says so

### Requirement: More than one person can approve a version
**Story**: S3
Each person's approval of a version SHALL be its own record. A story whose
implemented version two people approved SHALL show both.

#### Scenario: A second household member approves
- **WHEN** one person has approved a story's implemented version and a second person
  approves the same version from their own checkout
- **THEN** the canvas shows the story as approved by both, by name, and the
  history lists two approvals against that version

#### Scenario: The page changes after both approved
- **WHEN** the story's version then changes
- **THEN** the story shows as changed since approval, and its history still
  names both people against the earlier version

### Requirement: Each artboard says where its page stands
**Story**: S1
Each artboard of a page story that needs approval SHALL show the version that is
approved and the version that is implemented, wherever a status appears. When
they are the same version it SHALL show that one version once, as approved, with
who approved it and when. When they differ it SHALL show both: the latest
approved version with who and when, and the implemented version with the date it
was first seen, and SHALL say the page changed since approval. When no version
was ever approved it SHALL say so and show the implemented version alone. A story
that only asserts something SHALL show its implemented version and no approval
status. The status SHALL be readable without colour alone.

#### Scenario: The approved version is the implemented one
- **WHEN** a story's implemented version has an approval
- **THEN** its artboards read "Approved a3f9c21 · 20 Sep 2026 by Emanuel", and
  name no second version

#### Scenario: The approved version and the implemented one differ
- **WHEN** a story was approved at an earlier version and not at its implemented
  one
- **THEN** its artboards read "Changed since approval", then "Approved 77b0d4a ·
  12 Sep 2026 by Emanuel", then "Implemented a3f9c21 · 20 Sep 2026"

#### Scenario: A page nobody approved
- **WHEN** a story has no approval at any version
- **THEN** its artboards read "Never approved", then "Implemented a3f9c21 · 20
  Sep 2026"

#### Scenario: A story that only asserts something
- **WHEN** a person opens a screen's Behaviour row
- **THEN** those artboards read "Implemented a3f9c21 · 20 Sep 2026" and show no
  approval status

#### Scenario: The canvas before any versions exist
- **WHEN** the canvas is opened and no versions file can be read
- **THEN** the artboards draw as they did before, no status is shown, and the
  status bar says versions could not be read and names the command that writes
  them

### Requirement: The stories that need a look can be found
**Story**: S1
The sidebar SHALL offer one filter that leaves only the stories whose status is
changed since approval or never approved, with a count, and SHALL work together
with finding by name.

#### Scenario: Filtering to what needs a look
- **WHEN** a person turns on "Needs approval"
- **THEN** the sidebar lists only stories that changed since approval or were
  never approved, each marked with which, and the filter shows how many there
  are

#### Scenario: Nothing needs a look
- **WHEN** every story that needs approval is approved at its implemented version
  and the filter is turned on
- **THEN** the sidebar says "Every page is approved at its implemented version."

#### Scenario: Filtering and finding together
- **WHEN** the filter is on and a person types "overview" in the find box
- **THEN** only Overview stories that need approval are listed

### Requirement: A flow says where its pages stand
**Story**: S1
Each flow SHALL show how many of its steps and branches are approved at their
implemented version, how many are implemented at a version other than the one
approved, and how many were never approved, and SHALL read as approved only when
all of them are approved at their implemented version.

#### Scenario: A flow with pages still to approve
- **WHEN** of a flow's twelve steps and branches seven are approved at their
  implemented version, three changed since approval and two were never approved
- **THEN** the flow's entry in the sidebar and its heading on the canvas read "7
  of 12 approved · 3 changed since approval · 2 never approved"

#### Scenario: Nothing changed, some never approved
- **WHEN** ten of a flow's twelve are approved at their implemented version and
  two were never approved
- **THEN** the flow reads "10 of 12 approved · 2 never approved"

#### Scenario: A fully approved flow
- **WHEN** every step and branch of a flow is approved at its implemented version
- **THEN** the flow reads "Approved", and one later change to any of its pages
  returns it to a count naming one page changed since approval

### Requirement: A page story's history can be read
**Story**: S4
A person SHALL be able to open, from any artboard, the history of its story:
the implemented version first, then each earlier version a person approved,
newest first, each with its first-seen date and every approval against it with
who, when and the note. The history SHALL mark which version is implemented now
and which is the latest approved, on the same entry when they are one version. A
version nobody approved that has since been replaced is not kept.

#### Scenario: A page approved twice over its life
- **WHEN** a person opens the history of a story approved at one version, changed,
  approved again, then changed once more
- **THEN** they see three versions: the first marked "Implemented now, not
  approved", the second marked "Latest approved", and the third unmarked, each
  approved one with who approved it and when

#### Scenario: The implemented version is the latest approved
- **WHEN** a person opens the history of a story approved at its implemented
  version
- **THEN** the first entry is marked both "Implemented now" and "Latest approved"

#### Scenario: A page with no history
- **WHEN** a person opens the history of a story never approved
- **THEN** they see its implemented version marked "Implemented now, not
  approved" and "Nobody has approved this page yet.", with the command that
  approves it, ready to copy

#### Scenario: A long history
- **WHEN** a story has more approved versions than fit in the panel
- **THEN** the panel scrolls, the implemented version stays first, and the panel
  can be closed with Escape, returning focus to the artboard's history control

### Requirement: The record of approvals only grows
**Story**: S4
The check SHALL fail when the approvals record is malformed, names a story the
versions file has never held, names a version that story never had, or differs
from its last committed copy other than by lines added at its end. Each failure
SHALL name the line and the reason.

#### Scenario: A line is edited
- **WHEN** someone changes the name in an approval already committed and runs the
  check
- **THEN** the check fails naming that line and saying committed approvals are
  never changed

#### Scenario: A line is removed
- **WHEN** someone deletes a committed approval and runs the check
- **THEN** the check fails saying a committed approval is missing

#### Scenario: An approval for a version the story never had
- **WHEN** a line is added by hand naming a version the story never had
- **THEN** the check fails naming the story and the version

#### Scenario: A line that is not a record
- **WHEN** the approvals record holds a line that is not a complete approval
- **THEN** the check fails naming the line and what it lacks

#### Scenario: No earlier copy to compare against
- **WHEN** the check runs where the approvals record has never been committed
- **THEN** the content checks still run, the comparison is skipped, and the
  check says it was skipped and why

### Requirement: A changed page never fails the check
**Story**: S1
The check SHALL pass whatever the approval status of any page story, and
whatever a page now renders. A page changed since approval, or never approved,
is something the canvas shows and never something the check refuses.

#### Scenario: A token moves and nothing is re-approved
- **WHEN** a token changes and the check is run
- **THEN** the check passes, and the canvas then shows every page story as
  changed since approval, each with its approved version and its new implemented
  one

#### Scenario: A sentence is reworded
- **WHEN** one sentence on a screen is reworded and the check is run
- **THEN** the check passes

### Requirement: The versions file is what renders
**Story**: S1
Running the check SHALL bring the versions file up to date with what renders,
and SHALL never fail because the file was out of date. When it changes the file
it SHALL name the stories whose version moved and say there is a change to
commit. When nothing a person can see changed it SHALL leave the file
unmodified. It SHALL refuse to write, leaving the file unmodified and failing,
when the run that feeds it failed or did not cover every page story.

#### Scenario: A page changed since the versions were written
- **WHEN** a sentence on a screen is reworded and the check is run
- **THEN** the check passes, names the stories that have a new version, and says
  the versions file changed and is to be committed

#### Scenario: Nothing visible changed
- **WHEN** the check is run twice in a row
- **THEN** the second run leaves the versions file unmodified and says nothing
  changed

#### Scenario: A story failed in the run
- **WHEN** the check is run while one page story's play function fails
- **THEN** the check fails for that story, nothing is written, and the versions
  file is not modified

