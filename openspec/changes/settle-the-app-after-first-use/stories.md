## S1 · An account is never left with nobody

**As a** member setting up who sees the accounts a bank returned
**I want** wimm to refuse to leave an account with no owner at all, rather than
letting one checkbox do it silently
**so that** I cannot lose an account, its balance and its whole ledger by
misreading a control — and so that nobody has to know a URL to get it back

### INVEST

- **Independent** — Yes. The constraint and the backfill stand on their own: an
  account that has an owner today keeps one, and the orphans that already exist
  come back. S3 is what gives the member the thing they were reaching for when
  they orphaned it, but S1 is safe to ship without it.
- **Negotiable** — Yes. It states that zero owners is not a state wimm holds. It
  does not say where the guard lives, what happens to the control that used to
  orphan, or what the member is told instead.
- **Valuable** — Yes, and it is the only finding here that loses data a person
  cannot get back without help. Today one click takes an account off Overview,
  makes its transactions unreadable by every member, and can remove the only link
  to the screen that could undo it.
- **Estimable** — Yes. The unknown is priced rather than discovered: the
  constraint has to survive a member being deleted and an account being deleted,
  which is what makes it a deferred constraint trigger rather than a check.
- **Small** — Yes. One migration, one refusal, one backfill.
- **Testable** — Yes, and at the layer that matters: the refusal is asserted
  against the database directly, not only through the handler that happens to
  call it today.

### Capabilities

- `banking/household-accounts`
- `banking/bank-connections`

### Satisfied by

- `banking/household-accounts`: Requirement: An account always has an owner
- `banking/bank-connections`: Requirement: Choosing who owns each account and who sees it

## S2 · Hand an account to the person whose it is

**As a** member who connected the household's bank and now holds an account that
is really my partner's
**I want** to make them an owner of it and step back from it myself
**so that** the account is theirs in wimm the way it is theirs in life, and I
stop seeing an account I have no business seeing

### INVEST

- **Independent** — Yes, given S1. It needs the owner set to exist as an editable
  thing, which is the same control S1 makes safe.
- **Negotiable** — Yes. It states that ownership can be given and given up. Where
  the control sits and whether a member is told they were made an owner are open.
- **Valuable** — Yes, and it closes a hole rather than adding a feature: the spec
  has required this since `connect-bank-accounts` shipped — "Handing an account
  to the member it belongs to" — and no control for it was ever built. The
  scenario passes review and cannot be performed.
- **Estimable** — Yes.
- **Small** — Yes. One control, and the write path already exists.
- **Testable** — Yes. Two members, one account, and the first member stops seeing
  it.

### Capabilities

- `banking/bank-connections`
- `banking/household-accounts`

### Satisfied by

- `banking/bank-connections`: Requirement: Choosing who owns each account and who sees it
- `banking/household-accounts`: Requirement: An account always has an owner

## S3 · Keep an account out of wimm, and bring it back

**As a** member whose personal account sits at the same bank as the household's
joint one
**I want** to leave that account out of wimm entirely, and to bring it back later
if I change my mind
**so that** connecting the bank we share does not put my own balance into the
product, and deciding that is not a one-way door

### INVEST

- **Independent** — Yes once S1 exists, and it is what makes S1 kind rather than
  merely safe: refusing to orphan an account without offering this would take
  away the only way a member had to keep an account private.
- **Negotiable** — Yes. It states that a member can take an account out of wimm
  and put it back. Whether what was already read is deleted, and what an owner
  sees of a left-out account, are decided in the spec.
- **Valuable** — Yes. It is the affordance `ADR 0018` exists for, kept, and made
  reversible for the first time.
- **Estimable** — Yes.
- **Small** — Yes.
- **Testable** — Yes. Leave an account out, refresh, and no balance is read for
  it; bring it back, refresh, and one is.

### Capabilities

- `banking/household-accounts`
- `banking/bank-connections`
- `banking/transactions`

### Satisfied by

- `banking/household-accounts`: Requirement: An account can be left out of wimm
- `banking/bank-connections`: Requirement: Choosing who owns each account and who sees it
- `banking/transactions`: Requirement: An account left out stops filling the ledger

## S4 · Get to the part of the ledger I mean

**As a** member looking for something that happened in April
**I want** to jump to the newest page, to the oldest, or to a month, rather than
pressing Older until I arrive
**so that** finding a transaction from months back takes one action instead of
twenty, and I still always know which dates I am looking at

### INVEST

- **Independent** — Yes. It changes how the existing list is navigated and needs
  nothing else in this change.
- **Negotiable** — Yes. It states three ways to move. Whether the months are a
  list, a strip or a menu, and how far the aggregate looks back, are open.
- **Valuable** — Yes. At fifty rows a page, a household with two accounts reaches
  April in about a dozen clicks, each one a server round-trip with no indicator
  behind it.
- **Estimable** — Yes. Newest is the cursor already being absent; oldest is the
  same seek reversed; the months are one aggregate over the index the list
  already uses.
- **Small** — Yes.
- **Testable** — Yes, and against the property that matters: a jump lands on a
  page that is contiguous with the ones either side of it, with nothing repeated
  and nothing skipped.

### Capabilities

- `banking/transactions`

### Satisfied by

- `banking/transactions`: Requirement: Reading a ledger longer than one page
- `banking/transactions`: Requirement: Going straight to a month

## S5 · Call an account what we call it

**As a** member whose bank names three accounts "Conta à Ordem"
**I want** to give an account the name the household actually uses for it
**so that** Overview and the ledger read like our money rather than like the
bank's product list, and the name survives the next time the bank is reconnected

### INVEST

- **Independent** — Yes.
- **Negotiable** — Yes. It states that the household's name wins where one is
  set. Whether the bank's name is still shown, and who may set it, are decided in
  the spec.
- **Valuable** — Yes. The spec already admits the failure it fixes: "Two accounts
  at one bank seen only as balances … wimm neither shows any part of their
  numbers nor invents anything to tell them apart". A name the household sets is
  the household telling them apart itself.
- **Estimable** — Yes. The one real constraint is known: the account upsert sets
  `name` from the gateway on every connect, restore and reconnect, so the
  household's name cannot live in that column.
- **Small** — Yes.
- **Testable** — Yes. Rename, restore the connection, and the name is still
  there.

### Capabilities

- `banking/household-accounts`

### Satisfied by

- `banking/household-accounts`: Requirement: An account can be given the household's own name
- `banking/household-accounts`: Requirement: Seeing the accounts a member may see

## S6 · Connect a bank without losing the thread

**As a** member connecting a bank for the first time
**I want** the four steps to read as four steps of one thing, with the way
forward in the same place each time
**so that** I know where I am, what the next click does and how to go back,
before I am sent to my bank

### INVEST

- **Independent** — Yes as copy and placement; it does not depend on the layout
  pass, though both touch the same screens and should land together to avoid
  redrawing twice.
- **Negotiable** — Yes. It states one flow with a consistent way forward. Whether
  the steps are numbered, and what the step copy says, are decided on the canvas.
- **Valuable** — Yes. The four steps place their actions four different ways
  today: none at all on step one, a row that stacks on narrow screens on step
  two, a column on step four, and a row that never stacks on the widen variant.
  A member reaches the irreversible step — the hand-off to the bank — having been
  given no consistent signal about which control commits them.
- **Estimable** — Yes.
- **Small** — **No.** Four screens and a shared action pattern. Kept whole
  anyway: fixing one step's placement in isolation makes the inconsistency worse,
  not better, because the inconsistency *is* the finding.
- **Testable** — Yes, against the frames, which is what `ADR 0020` makes the
  standard.

### Capabilities

- `banking/bank-connections`
- `app/screen-layout`

### Satisfied by

- `banking/bank-connections`: Requirement: The steps of connecting a bank read as one flow
- `banking/bank-connections`: Requirement: Choosing a bank to connect
- `banking/bank-connections`: Requirement: Consenting at the bank

## S7 · Every window is a window wimm fits

**As a** member who uses wimm on a phone, on a tablet in landscape and on a
desktop
**I want** each of those to be a size the app was designed for, not a size it
falls between
**so that** the product does not look broken at the width I happen to use, and
the space a large screen gives is used rather than left empty

### INVEST

- **Independent** — Yes, and everything else that touches a screen depends on it,
  which is why it goes first.
- **Negotiable** — No, partly. The regimes and their boundaries are not
  negotiable here: `ADR 0002` declares four and `ADR 0004` sets what density does
  at a coarse pointer. What each regime looks like is entirely open, and is
  decided on the canvas.
- **Valuable** — Yes. Three breakpoint families are live at once — the tokens
  flip at 768, 1200 and 1800, both shells flip at 1024, and four leaf components
  flip at 599 — so between 768 and 1024 the tokens say one thing and the shell
  renders another. Medium has never been drawn at all, and the busiest screen's
  compact layout is written and never switched on.
- **Estimable** — Yes, once the frames exist. Not before: this is the story that
  cannot be sized from a task description, which is what `ADR 0020` says about
  every screen.
- **Small** — **No.** Every screen, at four regimes, drawn before it is built.
  Splitting it per screen was considered and rejected: the defect is that the
  regimes disagree *across* screens, so a per-screen fix cannot find it.
- **Testable** — Partly, and the limit is stated rather than papered over.
  `check-canvas.py` holds copy and `check-geometry.py` holds measurements; that a
  screen is laid out well at Medium is a judgement a person makes against the
  frame. The checks are a floor.

### Capabilities

- `app/screen-layout`

### Satisfied by

- `app/screen-layout`: Requirement: One set of size regimes, meaning the same thing everywhere
- `app/screen-layout`: Requirement: Every screen fits the window it is given
- `app/screen-layout`: Requirement: Navigation takes the form the window allows

## S8 · Know wimm heard me

**As a** member clicking through to Transactions while wimm waits on a bank
**I want** the app to show me that it has taken my click and is working
**so that** I do not click again, or conclude the app is broken, when the only
thing wrong is that it is slow

### INVEST

- **Independent** — Yes.
- **Negotiable** — Yes. It states that a pending navigation is visible and
  announced. What it looks like is decided on the canvas.
- **Valuable** — Yes. Every signed-in navigation is a blocking server round-trip,
  two of the connect steps call an open-banking gateway, and there is no
  indicator of any kind in the app: `navigating` appears nowhere, there is no
  progress bar and no skeleton component exists. A slow click and a click that
  did not land look identical.
- **Estimable** — Yes.
- **Small** — Yes. One component and one place in the root layout.
- **Testable** — Yes, and it must be tested as an announcement and not only as a
  drawing: a bar nobody sees is not feedback.

### Capabilities

- `app/navigation-feedback`

### Satisfied by

- `app/navigation-feedback`: Requirement: A navigation in progress is visible
- `app/navigation-feedback`: Requirement: A navigation in progress is announced

## S9 · Fit more of the ledger on the screen

**As a** member reconciling a month of transactions on a laptop
**I want** to set wimm to a tighter row height, and to find that setting in a
place that also holds how wimm looks
**so that** I can see seventeen rows instead of twelve without leaning in, and
so appearance stops being a control at the bottom of every page

### INVEST

- **Independent** — Yes, given S7: the Settings screen is a screen and needs the
  layout it is drawn into.
- **Negotiable** — Yes on what Settings holds and how it is laid out. Not on the
  touch rule: `ADR 0004` says the control is hidden where a pointer is coarse,
  not shown disabled.
- **Valuable** — Yes, twice over. The density axis has been fully specified and
  entirely unreachable since it was written — `data-density` is set by no line of
  code, so every screen has always been comfortable and the four files that read
  density tokens have never seen a second value. And appearance currently renders
  *after* the page content in the signed-in layout, which is not a place a
  preference belongs.
- **Estimable** — Yes.
- **Small** — **No.** It is a new screen, a new route, the first subsection list
  in the product, and a preference that reaches the document root before first
  paint. Kept whole anyway: a density control with nowhere to live is not
  shippable, and a Settings screen holding one moved toggle is not worth a route.
- **Testable** — Yes, including the part that is easy to skip: at a coarse
  pointer the control is absent, not disabled.

### Capabilities

- `app/settings`
- `app/screen-layout`

### Satisfied by

- `app/settings`: Requirement: A place to set how wimm looks
- `app/settings`: Requirement: Choosing how much fits on the screen
- `app/settings`: Requirement: A setting that cannot apply is not offered
