## MODIFIED Requirements

### Requirement: Choosing a bank to connect
**Story**: S1
A member SHALL be able to start connecting a bank by choosing it from a list of
banks wimm can reach. That list SHALL be the set wimm has configured as
connectable, not the gateway's full list of every institution in the country —
a bank absent from wimm's configured set is treated the same as a bank the
gateway does not offer at all. The list SHALL be searchable by name, because a
member knows the name of their bank and not its position in an alphabetical
list.

#### Scenario: Picking a bank from the list
- **WHEN** a signed-in member starts connecting a bank
- **THEN** they are shown a list of the banks wimm has configured as
  connectable, each identified by a name and logo they would recognise from
  their own banking app

#### Scenario: Finding a bank by typing its name
- **WHEN** a member types part of their bank's name
- **THEN** the list narrows to matching banks

#### Scenario: The bank is not in the list
- **WHEN** a member's search matches no bank, whether because the gateway does
  not offer it or because wimm has not configured it as connectable
- **THEN** they are told plainly that wimm cannot connect that bank yet, rather
  than being shown an empty list with no explanation

#### Scenario: The list cannot be loaded
- **WHEN** wimm cannot reach the service that supplies the list of banks
- **THEN** the member is told that connecting a bank is unavailable right now
  and invited to try again, and no partly-made connection is left behind

## ADDED Requirements

### Requirement: A bank's logo renders whole, whatever its shape
**Story**: S2
Every bank's logo in the list SHALL render in full, without cropping any part
of it, regardless of the logo's own width-to-height proportions. Every row's
bank name SHALL start at the same horizontal position as every other row's,
whether that row's logo is wide, narrow, or square.

#### Scenario: A wide logo next to a square one
- **WHEN** the list shows one bank whose logo is a wide wordmark and another
  whose logo is roughly square
- **THEN** both logos are shown whole, with none of either one cut off, and
  both banks' names begin at the same horizontal position

#### Scenario: A tall or narrow logo
- **WHEN** a bank's logo is taller than it is wide
- **THEN** it is shown whole rather than cropped to a square, and its bank's
  name still begins at the same horizontal position as every other row's
