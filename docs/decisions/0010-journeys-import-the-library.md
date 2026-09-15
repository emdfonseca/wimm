# 0010 · Journeys import the library

## Status

Accepted

## Context

ADR 0008 put journey flow in a journey `.pen` and left reusable mechanics in
`product-ui.lib.pen`, but never said how a journey reaches a component. pen.dev's
own component guide says components cannot be referenced across files and must be
copied — which, if true, would make the split untenable: 42 components duplicated
per journey, drifting from the moment one changes.

It is not true of the file format. `Document` carries an `imports` map — an alias
to a relative path — and the app resolves components and variables through it.
What is true is that nothing in the automation surface can write that key.

## Decision

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

## Consequences

- **A `.pen` is not encrypted.** It is pretty-printed JSON, and the pen-design
  skill said otherwise. The reason to read it through the MCP is that a
  whole-document read is thousands of lines and a `Get` visitor is not — not
  secrecy.
- **Imported components resolve their own tokens; a node you draw does not.** A
  library Button renders `#0C7A57` inside a journey file that declares no
  variables at all. A frame drawn in that same file with `$color-bg-canvas` falls
  back to `#000000` silently — it must be `$ui:color-bg-canvas`. This asymmetry
  is the likeliest source of a wrong-looking journey.
- The library is loaded whole: a `Get` visitor in a journey file walks the
  library's nodes too, prefixed. Traversals that assume they only see the
  journey's own nodes will need to filter on the prefix.
- `imports` is relative to the journey file, so journey files live beside the
  library. Moving either breaks the binding, and nothing reports it until a ref
  fails to resolve.
- Imports are read-only from the journey side. Changing a component still means
  editing the library and re-exporting the token contract as ADR 0003 requires.
