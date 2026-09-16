# Decisions

Generated from `docs/decisions/` by `just adr-index`. Never edit; write an ADR.

## 0001 · Stack and repo shape — Accepted

- Go for services and workers under `apps/`, importable code under `packages/`.
- SvelteKit with Svelte 5 for the UI. The app is deployable, so it lives under
  `apps/` like any other service; `packages/ui` is the design system it imports —
  the token contract and the Svelte components screens instance. Python only
  where a library forces it.
- **Three API surfaces, and each has exactly one job.**

  ```text
  our own services, server to server   Connect over protobuf
  our own web client, browser to app   REST over HTTP and JSON, in SvelteKit
  third parties and webhooks           REST with OpenAPI 3.1, RFC 9457 errors
  ```

  **The browser speaks REST, never Connect.** SvelteKit server routes are the
  web client's API: load functions and form actions for anything the framework
  already models, and JSON endpoints under `/api` for what it does not, such as
  handing an authenticator's response back. They hold the session cookie and
  call Connect from the server.

  Shipping a protobuf runtime to the browser costs 15-25 kB gzipped and buys
  little: the payloads that are actually error-prone, such as a WebAuthn
  credential, are JSON shapes the browser defines, so the schema would describe
  an envelope around an opaque string. It also cannot express a redirect, which
  any link-exchange flow needs. Server to server, the generated contract keeps
  its whole value: a Go CLI and a Go service cannot skew.

  **The web client's REST is not documented and not versioned**, which is what
  separates it from the third-party surface. Its only consumer ships in the same
  deploy, so a change is one commit rather than a negotiation, and `PageData`
  flowing from a load function into its page is stronger typing than a schema
  would give. The moment a second consumer appears it stops being this and
  becomes the third row, OpenAPI and all.
- Postgres with goose migrations owned by the service that owns the database.
- OpenTelemetry over OTLP for traces, metrics, and logs. The backend stays
  undecided; services never know it.
- devbox owns the toolchain, `just` owns every task entry point, and CI runs
  `devbox run -- just ci` and nothing else.
- Repo standards live as Claude skills in `.claude/skills/`; they reference only
  paths inside this repo.

## 0002 · Design system foundations — Accepted

**Typefaces.** Schibsted Grotesk for interface text, IBM Plex Mono for amounts,
account identifiers and table dates. Both are Google Fonts, self-hosted.
`type-family-display` resolves to the interface family — display is weight 600 at
−2% tracking, not a third font.

**Token naming.** Kebab-case (`color-bg-canvas`), mapping 1:1 to CSS custom
properties (`--color-bg-canvas`). This departs from the dotted vocabulary in the
pen-design standard's token reference; roles and structure are unchanged.

**Palette.** Pine & Signal. Deep pine green `#14402F` is the brand and primary
action; blue `#2563EB` is the highlight for links, selection and focus. Green and
red are reserved for money direction and feedback, never for branding, so an
amount's colour means exactly one thing.

**Theme axes.** Three, independent:

- `color` — light, dark.
- `device` — compact, medium, wide, **ultra**. A fourth regime beyond the
  standard's three, justified by one structural change at ≥1800 CSS px: the
  editing drawer stops being an overlay and becomes a persistent inspector pane.
- `density` — comfortable, compact. A third axis beyond the standard's two,
  justified by row height being the most consequential number in a ledger: at
  1440 × 900, roughly 684 px of table area shows 12 comfortable rows or 17
  compact ones.

## 0003 · Token pipeline provenance — Accepted

`design/tokens.json` is the contract CI owns. It carries, for every token, an
explicit `type` and `unit`; nothing downstream infers either.

`just check packages/ui` validates two separate things:

- **The export is a valid token document** — kebab-case names, known types,
  units legal for their type, values well-formed for their type, only declared
  axes, every branch of each axis present, no unknown fields, and no token
  carrying both a default and axis values.
- **The stylesheet agrees with the export** — regenerating is a no-op, every
  token reaches the CSS, and nothing else declares a custom property.

The validator is itself tested against inputs that must fail: a malformed hex, an
unknown type, an undeclared axis, a missing theme branch, and a non-kebab name.

**Library-to-export agreement is verified by a person, not by CI.** Re-exporting
is part of any change to variables, in the same commit, and the diff is the
review artifact.

## 0004 · Density availability follows pointer capability — Accepted

Layout regimes stay width-driven: the `device` axis describes available space,
which is what a width query measures.

The touch floor does not. `density = compact` reverts to comfortable under
`@media (any-pointer: coarse)`, and the control that sets density is hidden where
it does not apply rather than shown disabled — a greyed control invites someone to
work out how to enable it.

This supersedes the density consequence in ADR 0002. The three-axis decision
itself is unchanged.

## 0005 · Secondary navigation placement and content width — Accepted

**A subsection list is a card inside the page, at every pointer regime.** It sits
under the page header that names the area it belongs to. It is never a
full-height column beside the sidebar: the sections belong to the page, and a
column that starts above the page header claims they are a peer of the primary
nav. Width plays no part in the choice.

The form is chosen by the content and only by the content, which is what the
secondary-nav contract already said before the regime clause was added to it:

```text
a fixed handful of peer views   → tabs, in the page header, at every size
a longer or growing set         → a list card in the page, at every size
```

`layout-subnav-width` is 212 at Medium, Wide and Ultra, and 0 at Compact, where
Settings stacks.

**`layout-content-max` caps text measure only.** Prose, help and long-form
settings copy stay within 1200. Panels, tables and chart regions take the full
Main column. The token exists to protect line length; a form panel and a ledger
have no measure to protect.

## 0006 · A more vivid palette, and separating the two greens — Accepted

Three things move together, because moving any one of them alone is worse than
moving none.

```text
brand    #14402F / #9FE0C0  →  #0C7A57 / #5FE7B8
accent   #2563EB / #6699FF  →  #1F5FF0 / #7AA8FF
income   #067647 / #3DD68C  →  #127A33 / #4ADB6E
```

`#0C7A57` is the lightest the brand can be: white on it measures 5.33, and one
step lighter (`#0E8A62`) drops the button label to 4.35 and fails. Income moves
off the brand's hue rather than its lightness — 139 and 135 against the brand's
161 and 159, a gap of 22° and 24° where it was 2° and 1°. Every threshold holds:
white on brand 5.33, brand on canvas 4.95, accent as text 5.31 and 4.92, income
text 5.44 and 9.64, and the dark equivalents higher.

Sixteen tokens move, because the brand, accent and income families each carry
derivatives — hover and active states, the chart ramp's first two slots, the
focus ring, the info and success feedback pairs, and the hero gradient.

**The fill-only / text-only rule stays.** It is tempting to retire it now that the
greens differ by hue, and that would be wrong: a 22° separation inside the green
band is close to invisible under deuteranopia, which is the condition the rule was
protecting against in the first place. Hue separation helps typical vision and
does nothing for the case that mattered. What actually carries direction is the
sign: every amount shows `+` or `−`, so no reading of the ledger depends on
telling two greens apart.

Two tokens are added rather than adjusted. `color-text-on-brand` and
`color-text-on-brand-secondary` replace two colours that were hard-coded onto the
balance card, and which the new gradient broke — the card was carrying its own
palette and nothing said so until the contrast pass failed on it.

The hero gradient does not follow the brand all the way. `#0E6E4F` rather than
`#12805C`, because at the lighter value no secondary tone clears 4.5 against it:
the lightest candidate reached 4.44. The gradient is still far more saturated
than it was — 78% against 51% — at about the same lightness.

## 0007 · OpenSpec for change proposals — Accepted

OpenSpec holds change proposals and living specs. ADRs keep their role: a
decision record says what was chosen and why, a proposal says what is about to
be built. They are not substitutes.

The CLI is pinned in `bin/openspec`, a shim that execs `pnpm dlx` at an exact
version. devbox prepends `$DEVBOX_PROJECT_ROOT/bin` to `PATH`, so inside the
project shell a bare `openspec` is the pinned one.

The pin has to live on `PATH` rather than only in a just recipe. OpenSpec
generates its own instruction files — six `opsx` commands and six `openspec-*`
skills — and they invoke a bare `openspec` around a hundred times. A recipe
would leave every one of those calls resolving to whatever the machine happens
to have installed globally.

`just openspec <command>` delegates to the same shim, so the version is stated
once. There is no root `package.json`: a single dev CLI does not justify
introducing an npm manifest to a repo whose one Node package has no
dependencies.

Claude Code is the only configured tool integration.

## 0008 · The canvas is a proposal artifact — Accepted

A project schema, `openspec/schemas/wimm/`, adds a fifth artifact between design
and tasks. `openspec/config.yaml` sets `schema: wimm`, so every new change is
stamped with it.

```text
proposal → specs → design → canvas → tasks
                             │
                             └── tasks is blocked until canvas exists
```

**`canvas` draws the journey and records what was drawn.** The drawing goes in a
journey `.pen` under `apps/web/design/`, because a journey is the app's flow and
the app owns it; `packages/ui/design/product-ui.lib.pen` keeps holding reusable
mechanics only. `canvas.md` records the journey ID, the frame IDs, the
task surfaces, the library components instanced, the components the library is
missing, the states drawn, and the behavioural contracts the canvas cannot
execute.

**It is conditional.** A change with no user-facing surface — a worker, a
migration, a build script — records a deliberate skip. The condition is stated
in the artifact's own instruction, which is the mechanism OpenSpec already uses
for a conditional artifact.

**Components come before the screen that uses them.** Each entry under Components
missing becomes one task covering both the origin in `product-ui.lib.pen` and the
Svelte component in `packages/ui/src`. The screen composes instances and cannot
share a task with the components it instances.

## 0009 · The canvas saves without a person — Accepted

`bin/pen-save <file.pen>` pipes `save()` into the pinned CLI attached to the
running app, and `just pen-save <file.pen>` is the shape everything uses.

```text
1. execute   — mutate through the MCP
2. execute   — verify in a separate call
3. pen-save  — flush, and confirm git sees the file
```

The MCP stays the way edits are made. It is precise and deterministic: the
alternative the CLI offers is `pen --in … --out … --prompt …`, which runs a
second AI agent against the file and is neither. The CLI is used for the one
thing the MCP cannot do.

**The shim verifies rather than trusts.** It records the file's mtime, runs the
save, and fails if the file did not move — so "Is Pen.app running?" surfaces as
an error rather than as a canvas.md describing frames nobody can see.

**It runs through npm, not pnpm.** `@pen.dev/cli` imports `css-tree` without
declaring it, which resolves under npm's flat layout and fails under pnpm's
strict linking. This is the one place in the repo that reaches for npm, and the
reason is a defect in the package rather than a preference.

This supersedes the save consequence in ADR 0008. The canvas artifact itself is
unchanged.

## 0010 · Journeys import the library — Accepted

A journey file imports the library. `just pen-import <journey.pen> <alias>
<library.pen>` writes the `imports` key and creates the journey file if it does
not exist.

```text
imports     { "ui": "product-ui.lib.pen" }
components  ref: "ui:W2gOKx"
variables   "$ui:color-bg-canvas"
```

The alias qualifies everything. A bare id is a non-existent node, and a slash is
rejected — `ref` may not contain one.

**It writes JSON directly, because nothing else can.** `execute` has no
document-level mutator beyond `SetVariables`, and `Update(document, …)` reports
`Node 'document' not found`. A `.pen` is pretty-printed JSON, so the key is three
lines; the command exists to make the edit checked and idempotent rather than
hand-made. It refuses an alias already bound elsewhere, a library outside the
repository, and a non-kebab alias. The library does not sit beside the journey:
journeys live under `apps/web/design/` and the library under
`packages/ui/design/`, so the written path traverses upward, which pen resolves.

## 0011 · Draw headless, not through the MCP — Accepted

Journeys are drawn headless. `just pen-exec <file.pen>` pipes an execute snippet
into `pen interactive --in X --out X`, which opens X, resolves its imports, saves,
and exits.

```text
just pen-import   create the journey and bind the library
just pen-exec     draw, and verify in a second call
Export + read     look at the result
```

The MCP keeps one job: the document a person is actually looking at. `just
pen-save` keeps one job with it — flushing that document, which stays in the
app's memory until something saves it.

**`pen-exec` fails closed.** A snippet that errors exits non-zero and leaves the
file byte-identical; both are asserted, because a tool that writes garbage and
then exits non-zero has still corrupted the artifact.

**Headless saves only because it is told to.** `pen interactive` does not save on
exit. A pipe without `save()` reports every id it created and writes none of
them — the failure looks exactly like success.

This supersedes the workflow in ADR 0009. The reason that ADR exists — an edit
sitting unsaved in the app with nothing reporting it — is unchanged and still
applies to MCP work.

## 0012 · Stories carry the user, specs carry the contract — Accepted

A `stories` artifact sits between proposal and specs. `specs` requires it.

```text
proposal → stories → specs → design → canvas → tasks
```

The two levels split by audience, and the split is the point:

```text
story criterion   what a person can do, checked by using the product
spec scenario     normative system behaviour, checked by a test
```

A criterion naming a function, table, endpoint or component is not a criterion —
it is an implementation note, and it belongs in design or tasks.

**Stories name a real user by role.** "As a developer" and "as the user" are the
tell that there is no actor, and without an actor there is no value to state.

**INVEST is recorded, not performed.** Six lines, each answered. A story that
fails one is not automatically wrong: name the letter, say why shipping it anyway
is right. An honest "Small — no, this covers three screens, and splitting them
would ship a half-usable ledger" is worth more than six unconsidered ticks.

**Every requirement traces to a story.** If one does not, either the story is
missing or the requirement is.

**It is conditional**, like design and canvas. Tooling, refactors and internal
migrations record a one-line skip rather than inventing a user.

## 0013 · Acceptance criteria are scenarios — Accepted

The spec's scenarios are the acceptance criteria. There is no second list.

The line between the two artifacts is **lifetime**, not audience:

```text
openspec/specs/<cap>/spec.md   living. Accumulates across changes.
stories.md                     change-scoped. Archives with the change.
```

A requirement is a persistent behaviour contract; a story is a unit of work.
INVEST asks whether the work item is Independent, Estimable, Small — questions
with no meaning about a permanent capability.

So `stories.md` keeps only what a spec cannot hold: the actor, the value, the
INVEST answer, and `Satisfied by`. The user-testable bar moves onto the scenarios
themselves — a scenario for a user-facing capability must be observable by a
person, and one naming a function, table, endpoint or component is an
implementation note.

**The link is a metadata line.** Every requirement carries `**Story**: S<n>`
under its header. `**Key**: value` is the one pattern the parser recognises: it is
excluded from the requirement body when other text is present, and survives
archive verbatim. Free prose would not.

## 0014 · A neutral dark palette, and a check that can see it — Accepted

**Dark is a cool near-black, and its surfaces step evenly.** The ramp is built in
L\*, not in contrast ratio, at a fixed slate hue, with a step of five between
each surface and the next:

```text
                      L*
bg-canvas    #0A0B0D     3.0
bg-surface   #16181C     8.2      chrome: the sidebar, an auth card
bg-subtle    #1F2228    13.2
bg-elevated  #282C34    17.9      content that sits on the chrome
bg-hover     #323741    23.0
bg-active    #3D424E    27.9
```

Only the areas carrying meaning are saturated: the brand panel's gradient, the
primary action, the selected row, links, amounts and charts. The ground is
slightly blue rather than neutral grey, because a neutral one reads as the
absence of a colour decision rather than as a colour; the cast sits far below the
chroma of anything it carries.

**Contrast ratio is the wrong instrument down here.** Two adjacent surfaces at
L\* 8 and L\* 13 measure 1.07 against each other, and two at L\* 3 and L\* 8
measure 1.31 — the second pair looks no further apart than the first, because
ratio compresses towards black while perception does not. An earlier version of
this ramp was tuned by ratio and came out visibly muddy in its middle, with
perceptual steps of 4.9, 3.5, 2.4, 2.4 and 4.1 where it should have been even.

**Content sits one step above chrome.** A card on `bg-surface` inside a shell
also on `bg-surface` gives the eye two equal planes and no hierarchy. Page
content uses `bg-elevated`.

Hover, active, secondary, disabled, control and feedback-background derivatives
follow their families. `color-chart-1` follows the brand, so dark carries one
green rather than two.

**The brand gradient stays green.** It is the one branded surface in the product
and the only large area of colour on a dark screen, which is what makes it read
as branding rather than as decoration.

**The inverted button stays, and it is not a preference.** It is tempting to give
dark the same treatment as light — white label on a deep green fill — and it
cannot be done. A green dark enough to carry white text at 4.5 tops out around
`#12805C`, which measures 3.88 against the dark canvas: dimmer than light's own
button at 4.95, on a page where dimmer means invisible. In dark the fill has to be
light, and a light fill demands a dark label. The remedy for a button that shouts
is a quieter mint, not a return to white on green.

**A fill has a ceiling as well as a floor.** `color-action-primary` must clear 3.0
against the canvas so it reads as an action, and must not exceed 9.5, beyond which
it stops reading as a control and starts reading as a light source. This is the
first threshold in the system with a maximum, and it is the one that would have
caught what shipped.

**Separation is satisfied by fill or by border, not by both.** A card in light is
`#FFFFFF` on `#F4F7F5` — a ratio of 1.08 — and is perfectly legible because its
border does the work. The same rule applied to dark, where luminance does the
work, would fail light for doing it the other way round. Each separation rule
names the candidates that could carry it and requires that one of them does.

**`packages/ui/scripts/check-palette.py` asserts all of this, in every theme the
document declares**, and runs inside `just check packages/ui`: eighteen contrast
pairs, seven separation rules and one loudness ceiling, 52 relationships across
light and dark.

**Its fixture is the palette that shipped.**
`scripts/fixtures/dark-before-0014.json` holds the dark values as they were, and
`scripts/test-palette.py` asserts they are rejected, alongside five constructed
failures. Checking that the current document passes proves nothing about the
checker, because the current document is valid.

This supersedes the dark half of ADR 0006's palette. The light values, the
fill-only / text-only rule, and the reasoning about hue separation under
deuteranopia are unchanged.

## 0015 · The toolchain tracks the latest stable release — Accepted

Every package in `devbox.json` is pinned to the highest **stable** major
available from the devbox search index, and every one of them carries a major
pin — `postgresql@18`, never `postgresql` and never `postgresql@18.6`.

Stable excludes anything carrying `rc`, `beta`, `alpha` or a `-pre` suffix.
Python is the case that makes the word do work: `3.15.0rc1` is published and
`3.14` is the pin.

Bumping is part of any change that touches `devbox.json`. A change adding a
package checks every other pin in the same edit, so drift is caught by the work
already in flight rather than by a scheduled sweep nobody runs.

## 0016 · Passkey identity and enrolment links — Accepted

**Registration and enrolment are separate operations.** Registering creates the
member; enrolling creates a credential for them. The first enrolment link is
issued at registration, and the same mechanism issues a further one later, which
is what makes adding a device and replacing a lost one fall out of the design
rather than arrive as a feature.

**The enrolment link is a bearer credential, stored hashed.** 256 bits from a
cryptographic source, rendered in the URL path, with only its hash in the
database. Single-use; issuing a new link to a member invalidates any outstanding
one, so a member has at most one live link. Default lifetime 24 hours,
configurable.

Single-use and replacement-on-issue carry most of the security; the lifetime is
the weakest of the three. A shorter one is safer and worse — the operator sends
the link over chat and the member opens it in the morning, and every expiry is
another round-trip through a person.

The value reaches browser history and anything that records paths, so it is
exchanged server-side on first load and replaced by a redirect to a path that
does not contain it. Logs never record the enrolment path with its value.

*Alternative:* a signed self-contained token with no stored row. Rejected —
single-use and invalidation both need server-side state, so the row exists
anyway, and a token that cannot be revoked is the opposite of what this wants.

**Sessions are opaque identifiers backed by a row.** `HttpOnly`, `SameSite=Lax`,
`Secure` wherever the origin is HTTPS, naming a row that carries the member, an
absolute expiry and a last-seen time. Every lifetime is evaluated against
database time, so a skewed clock on one instance cannot extend a session.

*Alternative:* a signed self-contained cookie, which avoids a read per request
and cannot be revoked before it expires. In a product whose entire recovery story
is that the operator can cut a session off, that trade is the wrong way round.
The read is a primary-key lookup on a table with one row per browser.

**Enrolment signs the member in.** The ceremony just performed user verification;
a second one seconds later re-proves the same fact and costs a prompt. The
session's authority derives from the link, and anyone holding the link could sign
in immediately afterwards regardless, so nothing is conceded.

**The relying-party identifier is configuration and a one-way door.** It and the
list of expected origins are read at startup. Changing the identifier invalidates
every passkey ever registered against it with no migration path, so it must never
be a value compiled in and adjusted later. Choose the broadest domain the product
will ever need: a parent domain covers subdomains later, a subdomain locks the
product to it forever.

**Enrolment requires a discoverable credential.** `residentKey` and user
verification are both required, and a registration is rejected unless the
authenticator confirms it created a discoverable one. This is what makes sign-in
username-less, and what stops a member enrolling a credential they can never be
offered — whose only recovery would be another operator link.

*Alternative:* accept non-discoverable credentials and ask for an email address
at sign-in. Rejected: it trades a failure the system detects once, at enrolment,
with the link still usable, for a worse experience on every future sign-in.

**All ceremony state lives in Postgres.** A challenge is a row with a lifetime in
seconds, deleted when consumed. The obvious alternative, an in-memory map, works
on one instance and fails intermittently on two — a defect found in production
rather than in tests. With challenges, links and sessions all in the database,
`wimmd` holds no request-spanning state and needs no sticky routing.

**The operator surface is a separate service on a separate listener.** Two Connect
service definitions, two ports. The operator listener binds to loopback by default
and requires a bearer credential from configuration, compared in constant time.
`wimmd` refuses to start if that listener is enabled with no credential set.

Separate definitions rather than one service with a guarded method: the
distinction is which port answers at all, and a method-level check is one
forgotten annotation away from being public. *Alternative:* mTLS, which is right
once there is more than one operator and disproportionate for one.

**Failures reveal nothing, except to the operator.** Expired, spent, replaced and
never-issued links produce one identical response in body, status and timing. A
sign-in with an unknown passkey does not say whether the passkey, the member or
neither is the problem. Registering an already-registered email *does* say so
plainly: the operator is trusted and has to act on it.

**Dependencies.** A Go WebAuthn library, a Postgres driver, and Connect. Each is
new to the repo and each is load-bearing: the ceremony, the storage and the
contract respectively.

## 0017 · Sweeping expired identity rows — Accepted

`wimmd` sweeps on a ticker from one goroutine started beside the listeners,
sharing their signal context. One pass is three independent statements, one per
table, each reporting its own count and its own error; a failure in one does not
skip the others, because no invariant spans them. Each delete runs in bounded
batches of `store.DeleteBatchSize`, repeating until a batch comes back short, so
the lock is a function of the batch rather than of the backlog. Every comparison
is `now()` in SQL, enforced by `forbidigo` banning `time.Now` outside tests; the
interval between sweeps is a Go ticker, which is not a lifetime.

An expired session is removed at once. A revoked one is kept for
`WIMM_REVOKED_SESSION_RETENTION` after `revoked_at`, because revocation is the
whole recovery story for a lost device and a row deleted on revocation cannot
answer "was this actually cut off". Both that window and `WIMM_SWEEP_INTERVAL`
are configuration with defaults a household instance never sets, refused at
startup when non-positive like every other lifetime.

Each sweep logs one line carrying a count per kind, including when every count
is zero, and nothing else: no row identifier, and nothing that is stored hashed.

Sessions and enrolment tickets gain indexes on the columns the sweep scans —
`expires_at` on both, and a partial index on `sessions.revoked_at` for the
branch that reads it — built `CONCURRENTLY` in a `NO TRANSACTION` migration.
`ceremony_challenges` already had its own.

## 0018 · Reading banks through a gateway — Accepted

**Enable Banking is the first gateway, and nothing outside one package knows its
name.** `apps/wimm/internal/banking` holds a `Gateway` port in wimm's own
vocabulary; `internal/banking/enablebanking` implements it; `internal/banking/
bankingtest` implements it in memory. An import restriction in the lint
configuration that already bans `time.Now` refuses any import of an adapter from
outside `internal/banking/...`. Coupling a check can catch is the only kind
still absent in six months.

**The port is validated against a second gateway on paper.** Writing a second
adapter to prove the abstraction is disproportionate; asserting it without
checking is how wrapper-shaped interfaces get written.

```text
port method            Enable Banking            GoCardless Bank Account Data
---------------------  ------------------------  ----------------------------
Banks(country)         GET /aspsps?country=      GET /institutions/?country=
                       maximum_consent_validity  max_historical_days
BeginConnection        POST /auth -> url +       agreement + requisition
                       authorization_id          -> link + id
CompleteConnection     POST /sessions {code}     GET /requisitions/{id}
                       -> session_id + accounts  -> account ids, then fetch
                       (details returned once)
Balances(account)      GET /accounts/{uid}/      GET /accounts/{id}/balances/
                       balances
EndConnection          DELETE /sessions/{id}     DELETE /requisitions/{id}
Transactions (ch. 2)   GET .../transactions      GET .../transactions/
                       continuation_key          date window
```

Two findings came out of the exercise and both changed the signature.
`CompleteConnection` takes **both** the row written at begin and the callback,
because Enable Banking's truth is in the callback and GoCardless's is in the row.
It **returns accounts**, even though GoCardless needs a second round-trip to
produce them; the extra fetch lives in the adapter rather than leaking the
cheaper gateway's shape upward.

**Redirect-based gateways only, stated rather than designed around.** Enable
Banking, GoCardless, TrueLayer, Yapily and Tink all redirect. Plaid's Link step
is client-side and does not fit `BeginConnection -> URL`. Naming the constraint
beats a variant nothing exercises.

**Account identity is the gateway's cross-session hash, never its per-session
id.** Enable Banking issues a new `uid` for every account on every
authorisation; `identification_hash` is what matches an account across them. A
schema keyed on `uid` would lose every sharing choice the moment a member
restored a connection, which at 180 days is certain rather than possible.

**Everything the bank returns is stored, and the member chooses what is shared.**
`POST /sessions` returns account details **once**, and no endpoint lists them
again. So an earlier plan — store only the shared accounts, re-read the list if
the member changes their mind — is not implementable: changing your mind would
mean going back to your bank. Every account the session returns is stored with
`shared` defaulting to false, and **no balance is ever read for an unshared
account**, so wimm knows an account exists and does not know what is in it.

**The account chooser is not a consent step.** Some banks let the member narrow
accounts in their own consent screen; others hand over everything, and wimm
cannot know which happened or narrow what was granted. By the time the chooser
appears, access exists either way. It therefore asks one question — which of
these should the household see — and never implies it restricts the bank.
Nothing is preselected, because the whole reason the step exists is that
clicking through should not share a personal account; sharing everything is one
action, so the fast case stays fast.

Without this step, connecting a bank to share a joint account also shares every
personal account at that bank with everyone in the house. That is the failure
this decision exists to prevent.

**Balances are read when a member arrives, and again on request.** An earlier
draft made reading manual-only on the belief that balance reads are tightly
rate-limited. They are not in the case that matters: the roughly four-a-day ASPSP
cap applies to background fetching, and does not apply when PSU headers indicate
the member is present. There is no background reading at all in this change. A
429 carrying `ASPSP_RATE_LIMIT_EXCEEDED` is still possible and is handled by
leaving the previous readings on screen with their original times, which is the
one thing that must not be lost.

**Restoring is the whole flow again, because open banking has no renewal.**
Consent is valid until a date wimm sets, capped by that bank's
`maximum_consent_validity` — 180 days at most banks. A session can also expire
early, surfacing as `EXPIRED_SESSION` with 401, and both are treated identically
because the member's experience of "my bank stopped updating" is the same. wimm
requests the bank's maximum and shows the member the resulting date before the
hand-off, because it is that bank's limit and not a wimm policy.

Restoring re-enters the flow at the hand-off, skipping the picker. On return,
accounts are matched by hash so sharing choices carry forward; an account newly
offered arrives unshared, and one no longer offered goes with the member told
which.

This is in the first change rather than a later one. Every connection reaches
this state, so shipping without it means shipping a screen that tells a member
their bank stopped working and offers nothing.

**Nothing stored opens a bank on its own.** wimm never receives a member's
banking credentials and never holds the bank's OAuth tokens — the gateway holds
those and refreshes them internally. What wimm does hold is the gateway
`session_id`, which reads that household's accounts until the grant runs out,
and the per-account identifiers used with it. Those are live credentials and a
column is not where they belong.

They are sealed with AES-256-GCM before reaching the database. The key is read at
startup from a **path**, never from an environment variable holding the
material. The owning row's uuid is the additional authenticated data, so a
ciphertext lifted from one row cannot be replayed into another, and a key id is
stored beside each value so a key can be rotated without a flag day.

*Alternative:* `pgcrypto`, or full-disk encryption. Both rejected for the same
reason: the threat is a copy of the data — a dump, a backup, a replica, a stolen
volume — and in each of those the key travels with the rows, or the SQL carrying
it reaches the query log. Encrypting in Go keeps the key out of the database
entirely.

A sealed value has a type that cannot be printed: `String`, `MarshalJSON` and
`slog.LogValuer` all return a redaction, and plaintext exists only as a local
inside the adapter for the length of one call. Sealed values are destroyed when
access ends — on disconnection and on a grant running out — while the connection
row survives for the audit question. The application signing key is the master
credential and is treated the same way, with startup refused if its file is
group or world readable.

**Failures are wimm's taxonomy**: bank unavailable, gateway unavailable, consent
declined, consent expired, no accounts, rate limited with a retry-after.
`docs/design/surfaces.md` routes a page banner, an inline alert and a field error
off what kind of failure it is, and no two gateways agree on status codes.

**`bankingtest` is part of the port, not a test helper.** It keeps `just check`
off the network, and it is a second implementation written the same week as the
first, which is the cheapest pressure test for whether the interface abstracts
anything.

**Money is `int64` minor units plus an ISO 4217 code.** Never a float, never the
gateway's decimal string, never a bare number without its currency. Totals are
per currency and mixed currencies are never summed: wimm holds no rates, and
inventing one would invent the number a household trusts most.

**Accounts are household-visible once shared.** One instance serves one
household, so there is no scoping rule to implement beyond `shared` itself.
`connected_by` records whose consent is holding a connection open and who will
have to restore it.

**The full account number is never stored.** Only enough trailing characters to
tell two accounts at one bank apart, which is all the behaviour requires.
