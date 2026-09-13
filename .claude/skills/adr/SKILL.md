---
name: adr
description: Record an architecture decision as docs/decisions/NNNN-kebab-title.md. Use when the user says adr, "record this decision", or when /ship flags a missing ADR for a new dependency, service, public contract, or schema.
argument-hint: "[title]"
allowed-tools: Read Write Glob Bash
---

# ADR

Title: $ARGUMENTS

1. Gate. An ADR records a choice among alternatives: a dependency, a service, a public contract, or a schema. A convention with nothing to weigh (naming, layout, style) belongs in `.claude/rules/<lang>.md` or the owning skill. If the title fails the gate, say so and stop.
2. Path: `docs/decisions/NNNN-kebab-title.md`. NNNN = highest existing number + 1, zero-padded to 4; `0001` if the dir is empty.
3. Copy `assets/template.md`. Sections, in order: Status, Context, Decision, Consequences. Nothing else. The first line is `# NNNN · Title`.
4. Status is one of `Proposed`, `Accepted`, `Superseded by NNNN`.
5. ≤ 1 page. State present truth; no history narrative, no "previously", no dates.
6. An accepted ADR is never edited. Changing the decision → new ADR, and set the old one's Status to `Superseded by NNNN`.
7. Run `just adr-index`. It rewrites `.claude/rules/decisions.md`, which is generated and never edited; `just ci` fails if it is stale.
8. Report the path and stop. `/commit` if the user asks.
