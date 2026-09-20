# app/navigation-feedback Specification

## Purpose
Covers what wimm tells a member between their asking for a screen and that
screen arriving. Every signed-in navigation waits on the server and some wait on
a bank, so this capability owns the difference between "this is slow" and "my
click did not land".

## Requirements

### Requirement: A navigation in progress is visible
**Story**: S8
Whenever a member has asked for a screen and wimm has not yet shown it, the
product SHALL show that something is happening, in one place, on every screen,
however that navigation was started — a link, the primary navigation, a form or
the browser's own back button.

The indicator SHALL be in the same place every time and SHALL NOT move the page
under the member. The screen they are leaving stays legible and stays where it
is; the indicator is added to it, not a replacement for it.

It SHALL disappear when the new screen arrives, and SHALL also disappear when
the navigation fails or is abandoned, so it never outlives the thing it is
describing.

#### Scenario: A slow screen says it is coming
- **WHEN** a member opens a screen whose data takes a noticeable moment to
  arrive
- **THEN** they can see that wimm has taken their click and is working, before
  the new screen appears

#### Scenario: The same indicator wherever they started
- **WHEN** a member navigates by a link, by the primary navigation, by
  submitting a form, or by pressing back
- **THEN** the same indicator appears in the same place each time

#### Scenario: It goes when the screen arrives
- **WHEN** the new screen is shown
- **THEN** the indicator is gone

#### Scenario: A navigation that fails
- **WHEN** a navigation cannot be completed
- **THEN** the indicator is gone and the member is told what happened, rather
  than being left with something that says work is still going on

#### Scenario: A navigation the member abandons
- **WHEN** a member starts one navigation and immediately starts another
- **THEN** one indicator describes the one that is still running, and it clears
  when that one finishes

#### Scenario: The page does not jump
- **WHEN** the indicator appears
- **THEN** nothing already on screen moves, so a member mid-click does not click
  something else

### Requirement: A navigation in progress is announced
**Story**: S8
A member who cannot see the indicator SHALL still be told. When a navigation
starts, wimm SHALL announce that the screen is loading, and when it finishes,
SHALL announce the screen that has arrived — a member using a screen reader
otherwise hears nothing at all between pressing a link and finding themselves
somewhere new.

The announcement MUST NOT interrupt whatever is being read at the time.

#### Scenario: Told that a screen is coming
- **WHEN** a member using a screen reader follows a link to a screen that takes
  a moment
- **THEN** they are told that it is loading

#### Scenario: Told which screen arrived
- **WHEN** the new screen is shown
- **THEN** they are told which screen they are now on

#### Scenario: The announcement waits its turn
- **WHEN** an announcement is made while something else is being read out
- **THEN** it waits rather than cutting across it

### Requirement: Feedback is not a flicker
**Story**: S8
A navigation that completes quickly MUST NOT produce a visible flash. wimm SHALL
wait a short moment before showing the indicator, and once it is showing SHALL
keep it long enough to be read, so that a fast screen looks instant and a slow
one looks like work.

Where a member has asked for reduced motion, the indicator SHALL still appear
and SHALL NOT animate.

#### Scenario: A fast screen shows nothing
- **WHEN** a screen arrives almost immediately
- **THEN** no indicator appears at all

#### Scenario: An indicator that appeared stays readable
- **WHEN** the indicator has appeared and the screen then arrives very shortly
  afterwards
- **THEN** the indicator was on screen long enough to be seen rather than
  flashing

#### Scenario: Reduced motion
- **WHEN** a member has asked their system for reduced motion
- **THEN** they still see that a navigation is in progress, without anything
  moving or pulsing
