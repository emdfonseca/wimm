## Journey

- **J07 · See where the money went** — the Transactions destination and every
  state it can be in. Covers S1 and the visible half of S3.
- **J08 · Include a bank's transactions** — widening a connection made before
  wimm could read transactions. Covers S2.
- **J05 · Disconnect a bank** — existing journey, one copy change, because
  disconnecting no longer destroys what was read. Covers the rest of S3.

J01 to J06 were taken by the access and banking journeys. J07 and J08 are new
and are not reused.

## File

- `apps/web/design/03-transactions.pen` — J07. New file, a new product area.
- `apps/web/design/02-banking.pen` — J08 and the J05 change. Widening and
  disconnecting are connection journeys and belong beside J03, J05 and J06.

Neither is `packages/ui/design/product-ui.lib.pen`. `just pen-manifest verify`
reports `nothing renamed or removed, 0 node(s) added`, so the library is
untouched by this change.

## Saved to disk

Saved. `just pen-exec` writes on a clean run and restores the file on a failed
one, which it did twice during this work.

```text
apps/web/design/03-transactions.pen   2026-09-18T08:03   untracked, new
apps/web/design/02-banking.pen        2026-09-18T08:05   modified
```

`git status` shows both.

## Frames

| Frame | ID | Zone |
| --- | --- | --- |
| J07.A / 01 · Transactions / Wide / As it opens | `cKLtG` | 10 · PRIMARY SUCCESS |
| J07.A / 02 · Transactions / Wide / One account | `W77e8` | 10 · PRIMARY SUCCESS |
| J07.A / 03 · Transactions / Wide / An older page | `UzsQX` | 10 · PRIMARY SUCCESS |
| J07.A / 04 · Transactions / Wide / The oldest page | `rL1mC` | 10 · PRIMARY SUCCESS |
| J07.A / 01 · Transactions / Compact / As it opens | `i42KQd` | 10 · PRIMARY SUCCESS |
| J07.A / 02 · Transactions / Compact / One account | `fz6lh` | 10 · PRIMARY SUCCESS |
| J07.B / 01 · Transactions / Wide / No bank connected | `MXbPS` | 20 · ALTERNATIVE |
| J07.B / 02 · Transactions / Wide / Bank not sending transactions | `QNiY0` | 20 · ALTERNATIVE |
| J07.B / 03 · Transactions / Wide / Owns no account | `FRlQH` | 20 · ALTERNATIVE |
| J07.B / 04 · Transactions / Wide / Reading for the first time | `G1Fwh` | 20 · ALTERNATIVE |
| J07.C / 01 · Transactions / Wide / Refresh refused | `dawYE` | 20 · ALTERNATIVE |
| J07.C / 02 · Transactions / Wide / One bank did not answer | `YNN3l` | 20 · ALTERNATIVE |
| J07.D / 01 · Transactions / Wide / Access has run out | `m4swG` | 20 · ALTERNATIVE |
| J07.D / 02 · Transactions / Wide / Bank disconnected | `QPXP7` | 20 · ALTERNATIVE |
| J07.A / 01 · Transactions / Wide / As it opens · dark | `Sonbq` | 40 · THEME + A11Y QA |
| J08.A / 01 · What wimm will see / Wide / Widening | `R4QXc0` | 20 · ALTERNATIVE (02-banking) |
| J08.A / 02 · Transactions / Wide / Bank now included | `NwGCZ` | 20 · ALTERNATIVE (02-banking) |

Changed rather than added: `zfsws`, the Disconnect dialog in J05.A, whose
message now says the transactions already read are kept.

## Surfaces

```text
Transactions            full page, route-backed, /transactions
Paging                  the same page with a cursor, /transactions?before=<c>
                        and ?after=<c>. Route-backed, so a page is a place a
                        member can return to
The account filter      the same page with a query, /transactions?account=<id>
                        a chip names the account; removing it widens the list
                        no breadcrumb: a filter is not containment
Widening a connection   full page, route-backed, then out to the bank
                        re-enters the existing hand-off, skips the picker and
                        the account chooser
Narrow-scope state      inline alert on that bank, never a page banner
Refresh refused         inline alert above the list, list unchanged beneath
Disconnect confirmation unchanged surface, dialog, copy only
```

The narrow-scope alert is scoped rather than page-level for the reason ADR 0018
already recorded about ActivoBank: a household where one bank is always amber
and two are current is a steady state, and a page banner would cry wolf daily.

## Components used

| Component | Origin | Used for |
| --- | --- | --- |
| Signed-in landing | `IGbQe` | every wide screen |
| Signed-in landing / compact | `pu6qZ` | both compact screens |
| Sidebar nav / in shell | `q7dY7c` via the template | primary navigation |
| Nav item | `rEOk1` via the template | Overview and Transactions |
| Button | `W2gOKx` | Refresh, Continue to Monzo, notice actions |
| Notice | `WFSg5` | every alert, at three severities |
| Empty state | `w5ZouR` | all four empty reasons |
| Badge | `iyeM0` | the unsettled marker |
| Avatar | `Z2oYAy` | the row's mark |
| Dialog | `kEly8` via J05.A | the disconnect confirmation |

**The shell disables what does not exist.** `Signed-in landing` switches off
every destination but Overview in its own overrides, which is why a one-screen
product has looked like a product with no navigation. Each screen here turns
`l4tOM` back on and moves the current marker from Overview onto it. Adding a
destination is a change to that template's overrides plus a `destinations`
entry in code, not a new component.

## Components missing

**`BottomNav`** — persistent primary navigation for compact, sitting below the
content rather than over it.

*What was opened and read, searching by the data rather than the shape:*

- `App bar / compact` (`EpRPP`), which the signed-in shell uses — brand,
  account and sign out, and no navigation of any kind.
- `App header` (`uC4nV`) and `App header/compact` (`CowGt`) — read because the
  canvas audit mentions a nav trigger. It exists, disabled everywhere and
  enabled in the compact header, and nothing in the library draws what it
  opens. It is an inherited default rather than a decision, and design.md
  decision 12 rejects it.
- `Sidebar nav` (`fwKnX`) and `Sidebar nav/rail` (`UwIRb`) — the same job on a
  vertical axis. The rail is a collapsed sidebar, still a column, and neither
  survives being laid on its side: the item is horizontal, 36 px tall and
  `fill_container` wide.
- `Secondary nav` (`L0hfW`) and `Tab` (`CtsNw`) — read because either could
  have been argued into this role. Both are spoken for by `surfaces.md` as
  views inside one area, and using them for top-level areas would be the
  category error that table exists to prevent.

**The items are `Nav item` instances**, laid out vertically and centred, so
there is still one component that owns what a navigation item looks like.
**Three deliverables**: the origin in `product-ui.lib.pen` beside `Sidebar
nav`, `packages/ui/src/organisms/BottomNav.svelte`, and `BottomNav.stories.svelte`.

**`SeekPager`** — a footer that moves a list one page older or one page newer
and states the span of dates on screen.

*What was opened and read:*

- `Pagination` (`HlRZX`) — the system's existing answer to "how does a list get
  longer than a screen". It is built for offsets: a Range text reading "41 to
  60 of 1,284" and a Page text reading "3 / 65". This ledger seeks on
  `(booking_date, id)` rather than counting, so neither number exists without a
  second query that would be stale before it rendered. Bending it would mean
  overriding one text, switching another off, and re-pointing both chevrons,
  which is a different component wearing this one's name.
- `Table region` (`MFU8U`) — read because it is what places `Pagination` today,
  in a footer. Confirms the footer position and nothing else; see below for why
  the region itself is not instanced.

The two are not interchangeable and the difference is visible: this pager says
"4 August to 31 July 2026" and offers Newer and Older, and on the last page it
says there is nothing older rather than greying a chevron. **Three
deliverables**: the origin beside `Pagination`, `packages/ui/src/molecules/
SeekPager.svelte`, and `SeekPager.stories.svelte`.

**`LedgerRow`** — a transaction row carrying the account it came from and
whether the bank has settled it, with no categorisation and no bulk selection.

*What was opened and read, searching by the data rather than the shape:*

- `Transaction row` (`PviJn`) and `Transaction row/compact` (`wUywT`) — the
  system's existing answer to "where does a transaction live". Both carry a
  category column (`qCqxX` holding a `Category chip`), which this change does
  not have, and neither carries the account or a settled marker, which it
  needs. The selection control is already switched off in the origin, so that
  part is not the mismatch.
- `Table region` (`MFU8U`) — read for the same reason, since it owns the row's
  container. Its toolbar is a count plus a search field and a filter button,
  and its header row is select-all, Merchant, Category, Date, Amount. Search
  and categories are both out of scope, so instancing it would draw two
  controls this change does not build.
- `List scaffold` (`Eelod`) and `Detail scaffold` (`oNyVa`) — read to check
  whether a list already existed that was not table-shaped. Both are layout
  scaffolds with slots, and neither holds a row.
- `Bank row` (`hHYs5`) and `Account choice row` (`QU4bL`) — read because they
  are the other two row molecules and both carry a bank. Both describe an
  account, not something that happened to one.

So the rows on these frames are drawn in the journey file, and building the
component is a task. **Three deliverables**: the origin in
`product-ui.lib.pen` beside `Transaction row`, `packages/ui/src/molecules/
LedgerRow.svelte`, and `LedgerRow.stories.svelte` beside it.

**The library's word is "Merchant" and these frames say "Description".** A
ledger that carries a salary and a transfer between accounts has no merchant
on most of its rows. Recorded here rather than silently forked: the new
component uses Description, and `Transaction row` keeps Merchant until
something makes it wrong.

## States drawn

Every scenario in `banking/transactions` that a person can see:

```text
drawn   owners see their transactions          J07.A / 01
drawn   narrowed to one account                J07.A / 02
drawn   an older page, a day split over two     J07.A / 03
drawn   the oldest page, nothing older          J07.A / 04
drawn   unsettled transaction                  J07.A / 01, row 2
drawn   no bank connected                      J07.B / 01
drawn   bank connected before transactions     J07.B / 02
drawn   member owns no account                 J07.B / 03
drawn   reading for the first time             J07.B / 04
drawn   refresh refused, rows kept             J07.C / 01
drawn   one bank did not answer                J07.C / 02
drawn   access has run out, history kept       J07.D / 01
drawn   disconnected, history kept             J07.D / 02
drawn   widening a narrow connection           J08.A / 01, J08.A / 02
drawn   dark                                   J07.A / 01 · dark
```

**Deliberately not drawn.**

- *A transaction on a jointly owned account.* Both owners see the identical
  frame, because that is the requirement: nothing on the row says a second
  person is looking at it. A second drawing of J07.A / 01 with a different name
  on the sidebar would assert nothing.
- *Everything fits on one page.* The footer is simply absent. It is the same
  frame as J07.A / 01 minus one element.
- *A member granted balance or details sees no transactions.* The scenario's
  outcome is that nothing appears and nothing says anything was withheld, so
  the frame would be identical to J07.B / 03. Drawing an absence twice teaches
  nobody anything.
- *The same payment made twice in one day, and the same transaction read
  twice.* Both are assertions about what the list contains, not about what it
  looks like. They are spec scenarios and tests, and a frame showing two rows
  proves nothing a fixture could not fake.
- *Compact for the eleven alternative states.* The wide and compact primary
  screens establish the two compositions; the notices and empty states reflow
  by filling rather than by changing structure. Drawing eleven more phone
  frames would be eleven copies of a decision already made.
- *Medium and Ultra.* Neither adds a structural difference to this screen.
  Ultra's inspector pane has nothing to inspect, because nothing here is
  editable, so the regime widens the list and no more. Recorded as a note in
  zone 90 rather than drawn.

## Contracts for implementation

Behaviour the canvas states but cannot execute. Every number below was measured
on the frame named, not recalled.

```text
row height, wide          56 px, uniform across all five rows on cKLtG
row height, compact       64 px, uniform across all five rows on i42KQd
```

Both are literals on the frame rather than bindings: a `width` or `height`
bound to a variable is silently discarded by the canvas, so `density-row-height`
stays authoritative in code and the drawing states the comfortable value.

- **The list renders before the sync finishes.** Stored rows and their
  synced-through time paint first; the sync runs behind the arrival and updates
  in place. The canvas can only show the settled result, so the loading pass is
  a contract: no spinner over the list, no layout shift when it lands, and the
  time in the toolbar is the only thing that changes.
- **The updated list must be announced.** A member who asked to refresh gets no
  visual cue if nothing changed. A polite live region saying how many
  transactions arrived, or that nothing did, is required; the canvas cannot
  draw it.
- **Focus on removing the filter chip** moves to the list heading, not to the
  top of the document, so a keyboard user is not sent back through the sidebar.
- **The chip's remove control is a button, not the chip.** The chip names the
  account and is not itself interactive.
- **Order within a day is not stated by the canvas.** The frames show a day
  group with rows under it; which order they take inside one day is a spec and
  code question.
- **The unsettled badge carries text, not colour alone.** It reads "Not
  settled" beside a dot, and both themes keep the label.
- **Amounts carry a sign in every state.** Direction never depends on the
  red/green pair, per ADR 0006.
- **Paging is a route change, and focus follows it.** Older and Newer are
  links to a cursor, not buttons that mutate in place, so Back works. On
  arrival focus moves to the list heading rather than staying on a control that
  may no longer exist, and the new span of dates is announced.
- **A day repeated at the top of a page is the same heading, not a new one.**
  The frames show `4 August 2026` opening J07.A / 03 with rows from that day
  continuing. Nothing marks it as a continuation; if that reads as a second
  4 August, the fix is copy on the heading and the frame is redrawn to match.
- **The bottom bar is a sibling, not an overlay.** It sits below the scrolling
  region in the same column, so nothing is obscured and the list needs no
  bottom padding to clear it. Measured on `i42KQd`: the bar is 60 px and each
  target is 195 x 60, well clear of any target-size floor.
- **The bar never hides on scroll.** Navigation that comes and goes
  re-introduces the problem the hamburger has, intermittently, which is worse
  than having it all the time.
- **Current is marked three ways and only one of them is colour.** An accent
  top edge, an accent icon and an accent label at 600, plus `aria-current="page"`.
- **Neither owner of a joint account is marked on a row.** The rows a member
  sees on a jointly owned account are indistinguishable from their own, because
  the bank does not say which cardholder acted and wimm does not guess.

## Notes

**The canvas warns about the shell's own disabled nav nodes.** Every
`pen-exec` touching a `Signed-in landing` instance reports `fill_container`
sizing outside a flexbox for `u5RU0`, `e9JS0V`, `A0RiHq`, `zVJvb`, `Nc5C5`,
`B8M5f` and `N6xsc`. Those are the destinations the template itself switches
off. They are disabled, so they do not render, and the warning predates this
change.

**`partially clipped` fired eight times on the first primary frame and none of
it was real.** Measured per side, every reported node was either inside a
disabled subtree or reporting a coordinate-space artifact from
`resolveInstances`. The export confirms nothing is clipped. This is the failure
mode `docs/design/canvas-audit.md` already documents; filtering `enabled ===
false` on the node alone is not enough, because the disable sits on an
ancestor.

**"Joint savings" is a bank product name, not a wimm concept.** An earlier pass
drew the account column as `Joint account · Montepio`, which reads as wimm
asserting joint ownership in a place that only ever shows what the bank called
the account. Ownership is `account_owners` and appears where accounts are
described, never in a ledger row.

**Monzo is the interpolated bank.** Copy that names a bank uses Monzo, matching
02-banking, so `gen-canvas-contract.py` reads those sentences as templates
rather than dropping them as fixture text. Montepio is the household's second
bank and appears only in data nodes.

**`check-canvas.py` will fail until the screens exist.** It maps every drawn
frame to the files that render it and treats an unmapped frame as a gap. The
fifteen frames above are unmapped today; adding the mapping is part of the
screen tasks, not a separate one.
