## MODIFIED Requirements

### Requirement: A person approves the implemented version of a page story
**Story**: S2
A person SHALL be able to approve the implemented version of one or more page
stories with one command, naming the stories and optionally a note, or asking
for every page story that needs approval. The command SHALL regenerate the
versions once, list what it is about to approve, and ask once. The approval
SHALL record, for each story, the story, the version approved, the name and
email the person commits under, the time, and the note, as new lines at the end
of the approvals record, and SHALL change nothing else in it. The command SHALL
say what it recorded.

#### Scenario: Approving a page
- **WHEN** a person runs the approve command for the Overview's populated story
- **THEN** the command prints the story, the version it approved and their name,
  the approvals record is one line longer, and the canvas shows that story as
  approved by them once reloaded

#### Scenario: Approving with a note
- **WHEN** a person approves a story with the note "after the merchant rows were
  tightened"
- **THEN** that note appears beside the approval in the story's history

#### Scenario: Approving several pages
- **WHEN** a person runs the approve command naming three page stories and
  answers yes once
- **THEN** the command lists the three with their versions before asking, the
  approvals record is three lines longer, and it says it approved three page
  stories

#### Scenario: Approving everything that needs a look
- **WHEN** a person runs the approve command with `--needs`
- **THEN** it approves every page story that was changed since approval or
  never approved when the command started, and no story that only asserts
  something

#### Scenario: Nothing needs a look
- **WHEN** a person runs the approve command with `--needs` and every page story
  is approved
- **THEN** the command says nothing needs approval and records nothing

#### Scenario: The page changed since the person looked
- **WHEN** a person runs the approve command and the story now renders a version
  other than the one the versions file held
- **THEN** the command brings the versions file up to date, refuses, names the
  new implemented version, asks the person to look again, and the approvals
  record is not modified

### Requirement: An approval that cannot be true is refused
**Story**: S2
The approve command SHALL refuse, recording nothing and leaving the approvals
record unmodified, when a story does not exist, when a story has no current
version, when the same person has already approved that version, when a story
only asserts something and needs no approval, and when the command is given both
named stories and `--needs`. Each refusal SHALL say which of these it is. When
any story in a run is refused, the whole run SHALL be refused, naming every
story refused.

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

#### Scenario: One bad story in a run of several
- **WHEN** a person names three stories and one of them does not exist
- **THEN** the command fails naming that story, and none of the three is
  recorded
