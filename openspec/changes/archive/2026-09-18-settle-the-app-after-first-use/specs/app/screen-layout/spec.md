## Purpose
Covers how every screen in wimm uses the window it is given: the size regimes
the product recognises, what changes between them, what form the navigation
takes in each, and the rule that a screen is finished when it matches the frame
drawn for that regime rather than when it merely fits.

## ADDED Requirements

### Requirement: One set of size regimes, meaning the same thing everywhere
**Story**: S7
wimm SHALL recognise four window sizes and no others, and every screen and every
piece of chrome SHALL change at the same widths:

```text
compact   below 768                a phone, or a narrow window
medium    768 and above            a tablet, or a half-screen window
wide      1200 and above           a laptop or desktop window
ultra     1800 and above           a large desktop display
```

A screen MUST NOT introduce a boundary of its own. Where a part of a screen needs
to rearrange sooner than its regime does, that is the regime being wrong for that
screen, and it is settled by redrawing the frame rather than by adding a width
nothing else knows about.

Every regime SHALL be a regime something was designed for. A regime that no frame
has been drawn at is not a supported size, and MUST NOT be reached by leaving a
screen to fall between two others.

#### Scenario: The same width does the same thing on every screen
- **WHEN** a member resizes the window past the point where one screen changes
  its arrangement
- **THEN** every other screen and the navigation change at that same width, so
  the product never shows one regime's navigation beside another's content

#### Scenario: A window between two regimes
- **WHEN** a member uses wimm in a window 1000 wide
- **THEN** they see the medium arrangement complete — its navigation, its
  spacing and its content column together — rather than one regime's chrome
  around another's page

#### Scenario: Every regime is designed
- **WHEN** a member opens any screen at any of the four sizes
- **THEN** what they see is an arrangement that was drawn for that size, not one
  inherited from a smaller or larger window by accident

### Requirement: Every screen fits the window it is given
**Story**: S7
Every screen SHALL use the space its window offers: the margins, the gaps
between sections and the width of the page's content all follow the regime.
A screen MUST NOT stretch a line of text across a large display, and MUST NOT
keep a small window's margins on a large one.

Long-form text SHALL be held to a readable measure. Panels, tables and lists
SHALL take the width available to them, because a table has no measure to
protect and a cramped ledger is harder to read than a wide one.

Nothing SHALL overflow horizontally or be cut off at any of the four sizes, down
to a window 320 wide.

#### Scenario: A large display is used
- **WHEN** a member opens Overview on a large desktop display
- **THEN** the accounts and the total use the width available rather than
  sitting in a narrow column with empty space beside them

#### Scenario: Prose keeps its measure
- **WHEN** a member reads an explanation or a long setting description on a
  large display
- **THEN** the lines stay short enough to read comfortably rather than running
  the full width of the window

#### Scenario: Margins follow the window
- **WHEN** a member moves from a phone to a desktop
- **THEN** the space around the page grows with the window rather than staying
  at the phone's margin

#### Scenario: Nothing spills
- **WHEN** a member views any screen at any width from 320 upwards
- **THEN** nothing is cut off, nothing overlaps and the page never scrolls
  sideways

#### Scenario: The ledger at a narrow width
- **WHEN** a member opens Transactions on a phone
- **THEN** each transaction is readable as a stacked row rather than as a table
  squeezed into the width, and the arrangement is the one drawn for that size

### Requirement: Navigation takes the form the window allows
**Story**: S7
The primary navigation SHALL take the form each regime allows, and each form
SHALL reach every destination the others do:

```text
compact   along the bottom, within reach of a thumb
medium    a narrow rail beside the content
wide      a full sidebar, with labels
ultra     a full sidebar, with the extra width given to the page
```

Whichever form is showing, exactly one SHALL be present. A member MUST NOT see
two navigations at once, and MUST NOT see none.

Where a screen has subsections of its own, they SHALL be shown as a card within
the page under its header at every regime, never as a second full-height column
beside the primary navigation.

#### Scenario: One navigation at a time
- **WHEN** a member resizes the window across any regime boundary
- **THEN** the navigation changes form once, and at no width are two
  navigations visible together or none visible at all

#### Scenario: Every form reaches everywhere
- **WHEN** a member uses the navigation at any of the four sizes
- **THEN** they can reach every destination the product has, and the one they
  are on is marked as current

#### Scenario: The rail says what it is
- **WHEN** a member uses the narrow rail at medium
- **THEN** each destination is identifiable without its label being visible, and
  its name is available to anyone using a screen reader or hovering it

#### Scenario: Subsections belong to their page
- **WHEN** a member opens a screen that has subsections
- **THEN** those subsections appear inside the page beneath its heading, at
  every size, rather than as a column starting above it

### Requirement: A screen matches the frame drawn for its regime
**Story**: S7
A screen SHALL match the frame drawn for it at the regime it is being shown at:
the same content in the same order, the same words, and the same arrangement. A
difference between the two SHALL be settled rather than left — either the screen
is corrected or the frame is redrawn — and the decision SHALL be recorded with
its reason.

A regime with no frame is not a size wimm supports, so drawing comes before
building, and a screen MUST NOT be changed for a regime whose frame does not yet
say what it should be.

#### Scenario: The words on the screen are the words on the frame
- **WHEN** a screen is built or changed
- **THEN** the headings, labels, help text and messages it shows are the ones
  the frame for that screen carries

#### Scenario: A difference that was decided
- **WHEN** a screen deliberately differs from its frame
- **THEN** the difference is recorded with the reason for it, and the frame and
  the screen agree about which of them is right

#### Scenario: A screen with no frame
- **WHEN** a screen would be built for a regime nothing has been drawn at
- **THEN** it is drawn first, because the frame is what says what the screen is
