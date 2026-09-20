## Context

`Service.Banks` (`apps/wimm/internal/banking/service.go:136`) returns whatever
`gateway.Banks(ctx, country)` returns, unfiltered — currently every ASPSP the
Enable Banking adapter's `GET /aspsps` call returns for the country, minus
business-only institutions (`offersPersonalAccounts`,
`enablebanking/client.go:271`). `BeginConnection` and `RestoreConnection` also
call `gateway.Banks` directly, to resolve a stored `bankID` back to a `Bank`
(`service.go:150,169`).

`BankRow.svelte` renders each entry's logo in a fixed 32×32 box with
`object-fit: cover`, which crops any logo whose aspect ratio isn't square. The
box is already a fixed-width flex item ahead of the name (`gap-12`, then the
name at `flex: 1 1 auto`), so the name's start position is already independent
of the image's own dimensions — the crop, not misalignment, is the actual
defect. See proposal.md - Why.

## Language

- **Connectable banks** — the wimm-configured set of bank names the picker
  offers. Never "the bank list" alone, which the codebase already uses loosely
  for the gateway's raw ASPSP list; this change gives the curated subset its
  own name so the two are never confused in code or copy.

## Goals / Non-Goals

**Goals:**
- Filter the picker to a wimm-operator-configured set of bank names.
- Stop cropping bank logos of any aspect ratio.

**Non-Goals:**
- Any way to disable filtering and show the gateway's full raw list — the
  operator sets it to whatever list they want, including a long one, but
  there is no "unfiltered" switch.
- Country-aware allowlist entries. wimm today only ever queries one country;
  matching is by name alone.
- Changing how `BeginConnection` or `RestoreConnection` resolve a stored
  `bankID` — see Decisions.

## Decisions

**Filter in `Service.Banks`, not in the Enable Banking adapter.** The curated
set is a wimm policy — which banks wimm currently wants to offer — not
something Enable Banking's API expresses or constrains. ADR 0018 draws the
gateway boundary at "nothing outside one package knows Enable Banking's name";
putting the filter in the adapter would put a wimm business decision behind
that boundary, and would need re-implementing for a second gateway. Filtering
in `Service.Banks` after `gateway.Banks` returns applies uniformly regardless
of which adapter is configured, including `bankingtest`.

**`BeginConnection` and `RestoreConnection` keep resolving any bank the
gateway knows, not just the curated set.** They already call `gateway.Banks`
directly rather than `Service.Banks`, and that stays unfiltered. Restoring a
connection to a bank an operator later drops from the curated list must keep
working — the member already has a live connection there, and orphaning it
because the allowlist shrank would be a regression the curation wasn't meant
to cause. `Service.Banks` is only ever used to render the picker.

**Configuration is a plain name list, matched case-insensitively, exact
match.** `WIMM_BANKING_CONNECTABLE_BANKS`, comma-separated, following the
existing `WIMM_ENABLEBANKING_*` pattern in `config.go`. Default value is the
four banks in scope: `Caixa Económica Montepio Geral`, `Revolut`, `PayPal`,
`Activo Bank` — the exact strings Enable Banking's `/aspsps` returns as
`name`, which is also what the picker already displays, so an operator
copying a name they see in the picker gets a working entry. Matching ignores
case because there is no other reason for an entry to fail a copy-paste.

*Alternative considered:* matching by the gateway's bank ID
(`country:name`, `enablebanking/client.go:264`) instead of the bare name.
Rejected — the ID is gateway-adapter shape crossing the port boundary into
operator configuration, which is exactly what `Service.Banks` filtering above
avoids for the request path.

**Logo fix: `object-fit: contain`, no background on the image itself, box
widened to 120×48, row grown to 64.** The image's own `object-fit` changes
from `cover` to `contain`; the box carries no fill of its own, so a logo with
a transparent background stays transparent rather than sitting on a visible
letterbox. The initials fallback keeps its own background
(`var(--color-bg-subtle)`) — it is a chip standing in for a mark, not an
image, and needs one to read as an avatar. The box itself moves from the
library origin's square 32×32 to a wide 120×48. Most banks'
marks are wordmarks, and a square box letterboxed them down to a sliver
top-to-bottom; a wide box holds a wordmark nearly edge to edge while still
fitting a square or tall mark without clipping. The row grows from 56 to 64
to hold it; `BankRow` is never shown beside an `AccountRow`, so decoupling the
two heights costs nothing. The name's start position moves with the box,
uniformly across every row.

## Risks / Trade-offs

- An operator who mistypes an allowlist entry silently gets a shorter list
  with no error — there's no ASPSP catalog check at startup. Mitigated by
  defaulting to names copied verbatim from what the picker already shows.
- A very wide or very tall logo, once no longer cropped, may render small
  inside the 120×48 box. Accepted: the spec's guarantee is "whole and
  legible", not "fills the box," and a small correct logo beats a large
  wrong one.
