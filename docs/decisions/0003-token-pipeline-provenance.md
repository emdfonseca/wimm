# 0003 · Token pipeline provenance

## Status

Accepted; library agreement superseded by 0023

`design/tokens.json` is the source of token values and nothing exports it, so
there is no library for it to agree with. The validation of the token document
and of the stylesheet is unchanged.

## Context

ADR 0002 records the token contract and states that drift between the pen.dev
library and `tokens.json` is caught by a committed token-name snapshot. That
overstates what the check does, and the gap is not theoretical: a token was
reconciled by hand and had diverged again within the same working session, with
`just ci` green throughout. Names matched; one value pair did not.

CI cannot read a `.pen` file — it is encrypted and reachable only through the
pencil MCP with the desktop app running. So the library cannot be the artifact CI
compares against, and any claim that CI verifies design-to-code agreement is
false.

## Decision

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

## Consequences

- A green `just ci` means the export is internally valid and the CSS matches it.
  It does not mean the CSS matches the library. Do not read it as design-to-code
  verification.
- Changing a variable without re-exporting produces a silent divergence that no
  automated check in this repository will catch. The review of a design commit
  must include the token diff.
- Adding a token type or unit requires updating the validator's allowed sets,
  which is the intended friction.
- Should the library ever become machine-readable outside the desktop app, this
  decision is superseded: the export step disappears and CI compares directly.
