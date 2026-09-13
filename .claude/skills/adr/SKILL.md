---
name: adr
description: Record an architecture decision as docs/decisions/NNNN-kebab-title.md. Use when the user says adr, "record this decision", or when /ship flags a missing ADR for a new dependency, service, public contract, or schema.
argument-hint: "[title]"
allowed-tools: Read Write Glob
---

# ADR

Title: $ARGUMENTS

1. Path: `docs/decisions/NNNN-kebab-title.md`. NNNN = highest existing number + 1, zero-padded to 4; `0001` if the dir is empty.
2. Copy `assets/template.md`. Sections, in order: Status, Context, Decision, Consequences. Nothing else.
3. Status is one of `Proposed`, `Accepted`, `Superseded by NNNN`.
4. ≤ 1 page. State present truth; no history narrative, no "previously", no dates.
5. An accepted ADR is never edited. Changing the decision → new ADR, and set the old one's Status to `Superseded by NNNN`.
6. Report the path and stop. `/commit` if the user asks.
