## Skipped

No canvas. This change alters nothing a person sees or operates: no screen, no
flow, no task surface, no state, and no component's appearance or behaviour.
The rows it removes were already unreachable, so no drawn state changes — a
frame invented for it would be a drawing nobody builds.

The condition is the one the artifact's own instruction states: a change with no
user-facing surface records a deliberate skip rather than an empty canvas.md.

`product-ui.lib.pen` and `apps/web/design/01-access.pen` are untouched, so
`just pen-manifest verify` should report nothing added and nothing changed for
the duration of this change.
