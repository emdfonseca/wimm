# 0005 · Secondary navigation placement and content width

## Status

Accepted

Supersedes the content-width consequence in ADR 0002 and the regime-conditional
promotion recorded on the canvas. The three-axis and four-regime decisions in
0002 are unchanged.

## Context

Settings has six subsections and growing. Two questions were entangled: what form
that list takes, and how much of a wide screen the content is allowed to use.

The library had answered both by regime — the list sat inside the page below
1800 CSS px and promoted to a full-height column above it, and every content
column was capped at `layout-content-max` 1200. Measuring the result showed both
answers were wrong, in opposite directions.

**The promotion was inverted against its own cost.** As a card inside the page
the list costs 268 px: a page gutter, its own 212, and a 24 gap. Flush against
the sidebar it costs 220. The cheaper composition was reserved for the regime
with the most room to spare, and the expensive one given to the regime with the
least.

**The cap made extra width buy nothing.** At Ultra the Main column is 1632 px and
the cap discarded 212 of it as gutter, so a 1920 screen showed 1160 px of content
against 836 at 1440 — 480 px more screen for 324 px more content.

## Decision

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

## Consequences

- Ultra content goes from 836 to 1260 px on the Settings shell, against 836 at
  Wide. The wider screen now buys content rather than gutter.
- One composition to draw and implement per area instead of two, and no
  breakpoint at which the navigation rearranges itself.
- `50 · TEMPLATES` draws Settings at Wide and Ultra with the same composition;
  the difference between them is content width and nothing else.
- The refused arrangement is drawn in `40 · ORGANISMS / Secondary nav form and
  placement`, with the arithmetic that rejects it.
