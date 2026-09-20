## Journey

J12 · See where things stand (new). Plus a rename-only pass across the
existing J03–J10 frames in `apps/web/design/02-banking.pen` that were named
"Overview" and now carry that screen's real name, Accounts.

## File

`apps/web/design/05-overview.pen` (new — Overview is its own product area,
distinct from bank connections and from the ledger). The rename pass touched
`apps/web/design/02-banking.pen` in place. Library addition (Trend
sparkline) in `packages/ui/design/product-ui.lib.pen`.

## Saved to disk

All three saved by `bin/pen-exec`, headless, git-visible:
- `apps/web/design/05-overview.pen` — new file, mtime 2026-09-19 17:54:44
- `apps/web/design/02-banking.pen` — modified, mtime 2026-09-19 17:47:50
- `packages/ui/design/product-ui.lib.pen` — modified, mtime 2026-09-19 17:55:36

`just pen-manifest verify` passes: nothing renamed or removed in the
library, 18 nodes added (Trend sparkline + its 12 bars + label + caption,
plus the Bottom nav's 4th tab and its nav-item ref).

## Frames

| Frame | ID | Zone |
| --- | --- | --- |
| J12.A / 01 · Overview / Wide / As it opens | ymTip | 10 · PRIMARY SUCCESS |
| J12.A / 01 · Overview / Compact / As it opens | pm6Li | 10 · PRIMARY SUCCESS |
| J12.A / 01 · Overview / Medium / As it opens | GLXgY | 10 · PRIMARY SUCCESS |
| J12.A / 01 · Overview / Ultra / As it opens | Jn6Me | 10 · PRIMARY SUCCESS |
| J12.B / 01 · Overview / Wide / Before any bank is connected | BpJN6 | 10 · PRIMARY SUCCESS |
| J12.B / 01 · Overview / Medium / Before any bank is connected | J2EbOB | 10 · PRIMARY SUCCESS |
| J12.B / 01 · Overview / Ultra / Before any bank is connected | QM51C | 10 · PRIMARY SUCCESS |
| J12.C / 01 · Overview / Wide / Member owns no account | TYCSM | 10 · PRIMARY SUCCESS |
| J12.D / 01 · Overview / Wide / An account with balances only | XGDJb | 10 · PRIMARY SUCCESS |

J12.C's shell (Total tile only, no Recent transactions card, no Trend card)
also stands for the "no transaction history anywhere" scenario: both leave a
member with no owned account that has synced transactions, and the two are
visually identical — the Total tile is the only thing either state can show,
so a second frame with the same pixels would carry no new information. The
distinction (owns nothing vs. owns accounts with no transaction access) is
in the data behind the frame, not on the canvas — verified by both
scenarios' text in `banking/overview`, not by a second screenshot.

Plus 22 frames in `02-banking.pen` renamed from "· Overview ·" to
"· Accounts ·" (J03.A/01 ×4 regimes, J03.B/01, J03.C/01, J03.D/01, J04.B/01,
J05.A/01, J05.A/02, J06.A/01, J06.A/03, J09.A/05, J10.A/02, each with its
"plate" twin where one exists) — metadata rename only, see Contracts below
for what still needs a content pass.

## Surfaces

Overview is a full page, route-backed (`/`), not a task surface — no modal,
drawer, or inline decision on it. Nothing new here; Accounts' existing
surfaces (connect/restore/disconnect dialogs) are unchanged by the rename.

## Components used

- `ui:IGbQe` Signed-in landing (Wide shell), `ui:pu6qZ` Signed-in landing /
  compact (Compact shell), `ui:WMlvF` Signed-in landing / rail (Medium
  shell), `ui:NEIet` Signed-in landing / ultra (Ultra shell) — existing,
  reused as-is.
- `ui:oweC0` Metric tile — existing (already instanced once before, in
  J09.A/05's Household total; never built as a Svelte component, see
  Components missing).
- `ui:VE2z9` Ledger row — existing library node, used at Wide/Medium/Ultra
  (column layout, no clipping at those widths). At Compact its columns clip
  the Amount at 390px (`ui:VE2z9` was sized for the Wide table and has no
  compact variant of its own), so Overview's Compact frame (`pm6Li`) does not
  instance it for the recent-transactions rows; it composes a two-line
  stack — Mark, then Description+Amount on one line and Account+Date on the
  next — matching `LedgerRow.svelte`'s existing `compact` prop exactly. This
  is a one-off frame-level composition, not a new library node.
- `ui:w5ZouR` Empty state (via `IGbQe`'s default slot content, same pattern
  J03.A/01 already used for Accounts' own empty state).
- `ui:W2gOKx` Button, styled as secondary (fill `$color-action-secondary`,
  stroke `$color-control-border`) for "See all", matching the existing
  secondary-button pattern already used for "Older"/"Back" elsewhere in
  these journeys.
- `ui:rEOk1` Nav item, instanced as a 4th Bottom nav tab (icon `wallet`,
  matching the Sidebar's pre-existing Accounts item, which used `wallet`
  already — not `landmark`, which is reserved for the empty-state glyph).

## Components missing

**Trend sparkline** — added to `packages/ui/design/product-ui.lib.pen` this
session (id in the library file; instanced in these frames as
`ui:K3w7Lx`). A row of 12 bars, a label, and a caption — bars rather than a
line because pen.dev's layout system positions flex children, not arbitrary
points (the pen-dev skill's own "Graphs" note: "Line charts cannot be easily
built because the layout system cannot position individual points"), and
because design.md already chose a bar sparkline over a line for the same
reason on the code side. Needs: the origin (done), a Svelte component under
`packages/ui/src/molecules/` or `atoms/`, and its `.stories.svelte`.
Searched for an existing owner first: neither Metric tile's Delta (a single
number, not a series) nor any chart/graph component in the get_app_state
component list could carry a series of values — nothing else in the library
draws more than one data point.

**Metric tile, as code.** Not new to the library (`ui:oweC0` already
existed and was already instanced once, in an existing J09 frame), but it
has no Svelte component or story anywhere in `packages/ui/src` — the real
`AccountsOverview.svelte` hand-rolls its own `.total`/`.total-label`/
`.total-value` markup instead of instancing a shared tile. This predates
this change, but Overview's new total now depends on it existing for real,
so building it stops being optional. Searched: `AccountRow.svelte`,
`BankRow.svelte` — neither is a metric display, both are list rows.

## States drawn

- Overview, populated (total + recent transactions + trend), at all four
  regimes — Wide (`ymTip`), Compact (`pm6Li`), Medium (`GLXgY`), Ultra
  (`Jn6Me`).
- Overview, before any bank connected, at Wide (`BpJN6`), Medium (`J2EbOB`)
  and Ultra (`QM51C`). Not drawn at Compact — `pu6qZ`'s own default empty
  state (`IGbQe`'s `w5ZouR` slot) already covers this shape at that width,
  the same as every other empty state in these journeys; nothing about this
  one needs Compact-specific structure.
- Member owns no account: total only, no recent-transactions card, no trend
  card — Wide (`TYCSM`). Also stands for "no transaction history anywhere"
  (see Frames).
- An account with balance-only access: total including it, recent
  transactions from the transaction-access account only, and a trend card
  carrying a coverage note ("Doesn't cover every account — Nordea shares
  balances only, not transaction history") rather than a line drawn as if
  every account fed it — Wide (`XGDJb`).

The Sidebar nav, Bottom nav, Rail and Ultra sidebar all show Accounts as a
4th destination now (`wallet` icon) — `ui:q7dY7c`/`fwKnX` already carried an
Accounts item structurally; it was only ever hidden by the `IGbQe`/`WMlvF`/
`NEIet` shells' own descendant overrides, now flipped to visible in all
three. `ui:b3MTx` (Bottom nav) had no Accounts tab at all; one was added,
ordered to match the Sidebar (Overview, Accounts, Transactions, Settings).

The 22 renamed Accounts frames' internal content now matches the rename:
every in-app page title reads "Accounts" (6 frames carried one; the other 16
are pure empty/notice states with no page header to fix), the 8 "plate"
labels read "Accounts" too, and the 5 frames that still showed a Household
total (`ui:J09.A/05`, `ui:J10.A/02`, and the Disconnect confirmation, Access
has run out and Access restored plates) have that Totals section disabled —
Accounts shows no total, per `banking/household-accounts`'s "Before any bank
is connected, Accounts says so" requirement extending to every state, not
only the empty one.

## Contracts for implementation

- Overview's Body slot replaces `IGbQe`'s / `pu6qZ`'s default centered empty
  state with a vertical stack: Page header (Title only, no Refresh button —
  refresh stays on Accounts per design.md), Household total (one `Metric
  tile` per currency), Recent transactions (card: header with "Recent
  transactions" + secondary "See all" button, then up to N rows), Trend
  (card: one Trend sparkline per currency the member's owned accounts
  cover).
- Before-any-bank empty state reuses `IGbQe`'s existing default `Empty
  state` descendant overrides exactly as J03.A/01 (Accounts) already does:
  icon `landmark`, title, message, one enabled button. Button label "Go to
  Accounts", navigating to Accounts rather than performing a connection
  action — Overview never connects a bank itself (`banking/overview`:
  "Overview shows the total and nothing about accounts or connections").
  Icon on the button is disabled (no leading icon), unlike Accounts' own
  empty state which uses a leading `plus`.
- Every number on this canvas (bar heights, "90 days to 19 Sep", the three
  fixture transactions) is placeholder fixture data, not a measured
  contract — no pixel thresholds are asserted here beyond what the
  screenshots show at their captured regimes.
