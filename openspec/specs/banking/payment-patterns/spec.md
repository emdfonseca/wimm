# banking/payment-patterns Specification

## Purpose
States once what wimm calls a recurring payment and what it calls an unusual
payment, so that Overview, the balance chart and Transactions mark the same
payments for the same reasons and a member can check any mark against the
published rule.

## Requirements

### Requirement: What wimm calls a recurring payment
**Story**: S3
wimm SHALL call a payment recurring when money has gone out to the same
merchant, by the name `banking/transactions` defines, at about the same amount,
at a steady weekly, monthly or yearly interval, at least three times for a
weekly or monthly payment and at least twice for a yearly one. A yearly payment
seen only twice SHALL be called likely yearly until a third confirms it,
because two payments a year apart are a strong sign and not yet a pattern, and
few banks hand over the two years a third needs. It SHALL be recognised only from booked transactions of accounts the member owns, within
the scope `banking/overview` defines, and per currency.

About the same amount SHALL mean within a stated tolerance, so a subscription
whose price rose slightly stays one recurring payment and a supermarket visited
weekly for different amounts is not one. A steady interval SHALL allow for
weekends and for months of different lengths. One merchant MAY hold more than
one recurring payment where the amounts are clearly different.

A recurring payment SHALL carry the amount last paid, its cadence, and the date
the next one is expected. It SHALL stop being called recurring once the next
one is overdue by more than the slack its cadence allows, because a cancelled
subscription still listed is a commitment the member no longer has. Money
coming in is never a recurring payment.

#### Scenario: A subscription paid every month
- **WHEN** a member has paid Netflix €12.99 on the 14th of each of the last
  four months
- **THEN** Netflix is a recurring payment, monthly, €12.99, next expected on
  the 14th of next month

#### Scenario: The price went up a little
- **WHEN** a monthly payment of €12.99 becomes €13.99
- **THEN** it stays one recurring payment and shows €13.99

#### Scenario: The same merchant for different amounts every week
- **WHEN** a member shops at one supermarket most weeks for amounts between €30
  and €140
- **THEN** the supermarket is not a recurring payment

#### Scenario: Only twice so far
- **WHEN** a member has paid a new gym the same amount two months running
- **THEN** it is not yet a recurring payment, and becomes one with the third

#### Scenario: An insurance premium paid twice
- **WHEN** a member has paid one insurer €386.00 in March of last year and
  again in March of this year
- **THEN** it is a recurring payment, shown as likely yearly, next expected in
  March of next year

#### Scenario: The third year confirms it
- **WHEN** that premium is paid a third March running
- **THEN** it is shown as yearly

#### Scenario: The day moves around a weekend
- **WHEN** a monthly payment lands on the 1st, the 3rd and the 2nd of three
  months
- **THEN** it is a recurring payment, monthly

#### Scenario: Two subscriptions with one merchant
- **WHEN** a member pays one merchant €4.99 and €17.99 each month
- **THEN** they are two recurring payments

#### Scenario: A subscription that was cancelled
- **WHEN** a monthly payment was last seen two months ago
- **THEN** it is no longer a recurring payment

#### Scenario: One that is a few days late
- **WHEN** a monthly payment was expected two days ago and has not been booked
- **THEN** it is still a recurring payment, shown with the date it was expected

#### Scenario: A salary
- **WHEN** the same amount arrives from one employer every month
- **THEN** it is not a recurring payment, because money coming in is not one

#### Scenario: A payment not yet settled
- **WHEN** the third payment in a run is still pending at the bank
- **THEN** it does not count towards the three until it is booked

### Requirement: What wimm calls an unusual payment
**Story**: S4
wimm SHALL use one rule, everywhere, to call a payment unusual: the payment is
far above what is usual for its baseline, measured in a way one earlier large
payment cannot distort. It SHALL apply to money going out and to money coming
in, each judged only against its own direction. For money going out the
baseline SHALL be the member's own booked payments to that merchant where there
are at least five, and otherwise every booked payment going out in the 90 days
up to that payment. For money coming in it SHALL be what that payer has paid
before where there are at least five such payments, and otherwise every booked
amount coming in over the same 90 days. One coming in SHALL be labelled unusual
income, so the label and the sign both say which way it went.

A first payment to a merchant SHALL also be unusual when it alone is more than
a stated share of a typical month's money out, because a garage paid once has
no history to be far above and the general run of payments is too wide to catch
it. Such a payment SHALL say that it is a first payment to that merchant,
rather than what is usual. The rule SHALL work on
the ratio between amounts rather than the difference, so €400 against a usual
€40 and €4,000 against a usual €400 are judged alike.

Where every payment in a baseline is the same amount, a payment SHALL be
unusual when it is a stated multiple of that amount. A payment below a stated
share of a typical month's money in its own direction SHALL never be unusual,
however far above its baseline it is, and where no full month is held nothing
SHALL be called unusual, because wimm has no typical month to measure a share
of. Only booked transactions are ever unusual.

The rule MUST NOT be a percentile. The top twentieth of payments is unusual by
construction in every month, including a month where nothing unusual happened.

An unusual payment SHALL carry what is typical for its baseline, or that it is
a first payment to its merchant, so a member can see what it was measured
against. It is never hidden, removed or excluded
from a list because it is unusual.

#### Scenario: Far above what that merchant usually costs
- **WHEN** a member who has filled up at one station a dozen times for about
  €60 pays it €640
- **THEN** that payment is unusual, and says it is usually about €60

#### Scenario: A first payment that is a large part of a month
- **WHEN** a member pays a garage €1,650 having never paid it before, and about
  €2,100 goes out in a typical month
- **THEN** that payment is unusual, and says it is a first payment to that
  garage

#### Scenario: A first payment of an ordinary size
- **WHEN** a member pays a restaurant €48 having never paid it before
- **THEN** that payment is not unusual

#### Scenario: A merchant paid a few times
- **WHEN** a member pays a merchant they have paid three times before €900, and
  their payments over the 90 days before it are mostly between €10 and €60
- **THEN** that payment is unusual, measured against their payments in general,
  and says what a usual payment is

#### Scenario: Large, and what that merchant always costs
- **WHEN** a member pays €820 rent to the same landlord for the sixth month
- **THEN** the payment is not unusual, because €820 is what that merchant costs

#### Scenario: Far above usual, and still small
- **WHEN** a member who buys a €1.20 coffee most days buys €9 of coffee
- **THEN** the payment is not unusual, because it is too small a share of a
  month to matter

#### Scenario: A fixed price that tripled
- **WHEN** every payment to a merchant has been exactly €12.99 and one is
  €38.97
- **THEN** that payment is unusual, if it clears the share of a month that
  matters

#### Scenario: A quiet month
- **WHEN** a month holds no payment far above its baseline
- **THEN** nothing in it is unusual, however its largest payment ranks

#### Scenario: Too little history
- **WHEN** a member's ledger holds no full calendar month
- **THEN** no payment is called unusual

#### Scenario: A bonus
- **WHEN** an employer who has paid €2,450 each month for a year pays €9,804
- **THEN** that payment is unusual income, and says it is usually about €2,450

#### Scenario: The salary itself
- **WHEN** the month's €2,450 salary arrives, far larger than anything going
  out
- **THEN** it is not unusual, because money coming in is judged against money
  coming in
