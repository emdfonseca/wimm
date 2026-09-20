## ADDED Requirements

### Requirement: A screen is designed at its regime before it is built into the app
**Story**: S7
A screen SHALL be designed as itself: shown on example data, in every state a
member can reach, at each of the four regimes, and looked at there before it is
connected to anything real. What was looked at and what ships SHALL be the same
screen, so there is no second copy for the product to drift from.

The words a screen shows in each state SHALL be held to: a heading, label, help
text or message that changes without anyone deciding it should is a defect, and
it SHALL be caught before the change ships.

A state or a regime nobody has looked at is not one wimm supports, so a screen
MUST NOT ship a state, or be changed for a regime, that was never seen on
example data first.

#### Scenario: The words on the screen are the words that were decided
- **WHEN** a screen is built or changed
- **THEN** the headings, labels, help text and messages it shows in each state
  are the ones that were looked at and agreed for that state

#### Scenario: Words that changed by accident
- **WHEN** a change alters a heading, label or message nobody set out to alter
- **THEN** the change is refused until the wording is either restored or
  deliberately accepted

#### Scenario: A state nobody looked at
- **WHEN** a screen can reach a state, such as empty, loading or failed, that
  was never shown on example data
- **THEN** that state is shown and looked at, at every regime, before the
  screen ships with it

## MODIFIED Requirements

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
screen, and it is settled by designing the screen again at that regime rather
than by adding a width nothing else knows about.

Every regime SHALL be a regime something was designed for. A regime a screen has
never been looked at in is not a supported size for it, and MUST NOT be reached
by leaving a screen to fall between two others.

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
- **THEN** what they see is an arrangement that was designed for that size, not
  one inherited from a smaller or larger window by accident

## REMOVED Requirements

### Requirement: A screen matches the frame drawn for its regime
**Reason**: There is no frame any more. A screen and its state stories are the design, so there is no second copy for a screen to match.
**Migration**: "A screen is designed at its regime before it is built into the app" carries the same promise to a member: the words were decided, every state and size was looked at, and an accidental change is refused.
