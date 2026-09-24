## MODIFIED Requirements

### Requirement: A typical month and an average month
**Story**: S1
Overview SHALL state a typical month's net, which is the middle one of the
newest six full months held, and the average month's net over those same
months beside it. Where six or fewer full months are held, both SHALL be drawn
from every full month held. Overview SHALL say how many full months they are
drawn from, and where more than six are held SHALL say they are the last six.
Neither SHALL be stated from fewer than three full months; Overview SHALL say
instead that they appear once three full months are held. The month so far and
any partly held month SHALL be in neither. The months shown SHALL NOT change:
a full month older than the last six stays in the months, in their chart, and
in what they say about their span.

Overview SHALL say in a sentence whether more comes in than goes out in a
typical month, or more goes out than comes in, with the amount. That sentence
SHALL be drawn from the typical month and never from the average, because one
exceptional month moves an average and barely moves the middle month.

#### Scenario: Living within what comes in
- **WHEN** a member holds twelve full months, and the typical month of the
  newest six is €189.40 more in than out
- **THEN** Overview says that in a typical month €189.40 more comes in than
  goes out, and that this is drawn from the last 6 full months

#### Scenario: Six or fewer full months held
- **WHEN** a member's ledger holds four full months
- **THEN** the typical and average month are drawn from all four, and Overview
  says they are drawn from 4 full months

#### Scenario: An exceptional month older than the last six
- **WHEN** a member holds twelve full months, and the eighth newest held a
  €4,000 repair
- **THEN** neither the typical nor the average month moves because of it, and
  that month is still shown among the months with its repair

#### Scenario: More goes out than comes in
- **WHEN** the typical month is €212.40 more out than in
- **THEN** Overview says plainly that more goes out than comes in, and by how
  much in a typical month

#### Scenario: The two figures disagree
- **WHEN** one of the last six months held a large repair, so the average month
  is below zero and the typical month is above it
- **THEN** both figures are shown, and the sentence follows the typical month

#### Scenario: Two full months held
- **WHEN** a member's ledger holds two full months
- **THEN** the months are shown with no typical or average month, and Overview
  says they appear once three full months are held

#### Scenario: The month so far is a bad one
- **WHEN** the month so far is well below zero on the 5th, before the salary
- **THEN** the typical and average month do not move, because the month so far
  is in neither
