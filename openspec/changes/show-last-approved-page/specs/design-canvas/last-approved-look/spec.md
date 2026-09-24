## Purpose

Keeps a picture of each page story's last approved version and lets a person
switch a changed artboard on the design canvas between that picture and the page
as implemented, so what moved since approval can be seen rather than remembered.

## ADDED Requirements

### Requirement: Approving a page keeps a picture of the version approved
**Story**: S1
When a person approves a page story, the approve command SHALL keep a picture of
that story at every size the canvas draws it at, in the light theme at
comfortable density, showing the whole story at its full height. The pictures
SHALL be taken from the same render that settled the implemented version being
approved. The command SHALL say the pictures were kept and that they are
committed with the approval. When the command records no approval, for any
reason, it SHALL keep no picture and change no picture already kept.

#### Scenario: Approving a page
- **WHEN** a person approves the Overview's populated story and answers yes
- **THEN** the command says it recorded the approval and kept a picture at
  compact, medium, wide and ultra, and names the files to commit with the
  approvals record

#### Scenario: A story drawn at one size only
- **WHEN** a person approves a story the canvas draws only at compact
- **THEN** one picture is kept, at compact

#### Scenario: Declining
- **WHEN** a person runs the approve command and answers anything but yes
- **THEN** no picture is written and the pictures already kept for that story
  are unchanged

#### Scenario: A refused approval
- **WHEN** the approve command refuses, for example because the page changed
  since the person looked
- **THEN** no picture is written and the pictures already kept for that story
  are unchanged

### Requirement: Only the last approved version keeps pictures
**Story**: S1
A page story SHALL keep pictures of its last approved version only. Approving a
newer version SHALL replace the story's earlier pictures, leaving none of them.
A second person approving the version that already has pictures SHALL replace
them with pictures of the same version.

#### Scenario: Approving a newer version
- **WHEN** a story has pictures of version 77b0d4a and a person approves version
  a3f9c21
- **THEN** the story has pictures of a3f9c21 only

#### Scenario: A second person approves
- **WHEN** Grace approves the version Emanuel already approved
- **THEN** the story still has one picture per size, of that version

### Requirement: A changed artboard switches to its last approved look
**Story**: S1
An artboard that reads "Changed since approval", viewed in the light theme at
comfortable density, SHALL offer a control that switches it between the page as
implemented and the picture of the last approved version at that artboard's
size, and back. While the picture shows, the artboard SHALL say it shows the
approved version and name it. Where no picture was kept for the last approved
version at that size, the artboard SHALL say so and offer no control. An
artboard that is approved, never approved, exempt, or viewed in the dark theme
or at compact density SHALL offer no control.

#### Scenario: Switching to the approved look
- **WHEN** a person activates "Approved look" on a changed artboard at wide
- **THEN** the artboard shows the picture taken at wide when 77b0d4a was
  approved, and reads "Showing approved 77b0d4a"

#### Scenario: Switching back
- **WHEN** a person activates "Approved look" again on that artboard
- **THEN** the artboard shows the page as implemented and the "Showing approved"
  line is gone

#### Scenario: Approved before pictures were kept
- **WHEN** a changed story's last approval was recorded with no picture
- **THEN** its artboards read "No picture of the approved version" and offer no
  control

#### Scenario: Viewed in dark
- **WHEN** a person opens a changed story's screen with the dark theme
- **THEN** its artboards offer no control

#### Scenario: A page that has not changed
- **WHEN** a story's implemented version is its last approved version
- **THEN** its artboards offer no control

### Requirement: A picture that does not match its approval is refused
**Story**: S1
`just check apps/storybook` SHALL fail, naming each file, when a kept picture is
not of its story's last approved version, when it belongs to no story in the
approvals record, or when it is at a size the canvas does not have. A story whose
last approval has no picture SHALL NOT fail the check.

#### Scenario: A picture left from an earlier version
- **WHEN** a story's pictures are of 77b0d4a and its last approval is of a3f9c21
- **THEN** the check fails naming those pictures

#### Scenario: An approval recorded with no picture
- **WHEN** a story's last approval predates this change and has no picture
- **THEN** the check passes
