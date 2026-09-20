## Purpose
Covers the place a member sets how wimm looks and how much of it fits on the
screen, how those choices are remembered and applied, and the rule that a choice
which cannot take effect where a member is sitting is not offered to them there.

## ADDED Requirements

### Requirement: A place to set how wimm looks
**Story**: S9
wimm SHALL have a Settings screen, reachable from the primary navigation at
every size, that holds the choices a member makes about the product itself. The
choices about how wimm looks SHALL live there and nowhere else — a preference
that appears underneath the content of every page is not a setting, it is
clutter on every screen.

Settings SHALL be organised as named subsections listed within the page, so
further subsections can be added without rearranging it.

Every choice SHALL take effect immediately, with no separate saving step, and
SHALL say what it does in the words a member would use.

#### Scenario: Reaching Settings
- **WHEN** a member uses the primary navigation at any size
- **THEN** Settings is one of the destinations, and opening it shows its
  subsections listed within the page

#### Scenario: Appearance lives in Settings
- **WHEN** a member wants to change between the light and dark appearance
- **THEN** they find that choice in Settings, and it is not repeated at the
  bottom of other screens

#### Scenario: A choice takes effect at once
- **WHEN** a member changes a choice in Settings
- **THEN** the product changes to match immediately, with nothing to confirm or
  save

#### Scenario: The choice that is in force is shown
- **WHEN** a member opens Settings
- **THEN** each choice shows which option is currently in force

### Requirement: Choosing how much fits on the screen
**Story**: S9
A member SHALL be able to choose between a comfortable and a compact arrangement
of rows, and the choice SHALL apply everywhere rows are listed, most of all in
the ledger, where row height decides how much of a month a member can see at
once.

The choice SHALL be that member's own and SHALL be remembered between visits.

It SHALL be in force by the time the first screen is drawn, so a member never
sees the product rearrange itself after it has appeared.

#### Scenario: More rows on the screen
- **WHEN** a member chooses the compact arrangement
- **THEN** more transactions fit in the same space, and every list in the
  product follows

#### Scenario: Remembered next time
- **WHEN** a member who chose compact comes back later
- **THEN** the product is still compact

#### Scenario: No rearranging after the screen appears
- **WHEN** a member who chose compact opens wimm
- **THEN** the first thing they see is already compact, rather than appearing
  comfortable and then changing

#### Scenario: One member's choice is their own
- **WHEN** one member chooses compact
- **THEN** another member of the same household is unaffected

### Requirement: A setting that cannot apply is not offered
**Story**: S9
Where a choice cannot take effect for the member in front of it, wimm SHALL NOT
show it. A control shown but disabled invites somebody to work out how to enable
it, and there is nothing for them to work out.

The compact arrangement is the case this exists for: it makes rows smaller than
a finger can reliably hit, so on a touch device the product stays comfortable
and the choice is absent rather than greyed out.

#### Scenario: Not offered on a touch device
- **WHEN** a member opens Settings on a device they are touching rather than
  pointing at
- **THEN** the choice of arrangement is not shown at all, and the product is
  comfortable

#### Scenario: Offered where it applies
- **WHEN** a member opens Settings on a device with a mouse or a trackpad
- **THEN** the choice of arrangement is shown and works

#### Scenario: A choice made earlier, on a device it cannot apply to
- **WHEN** a member who chose compact on a laptop opens wimm on a phone
- **THEN** the phone shows the comfortable arrangement, and their choice is
  still in force when they return to the laptop
