## Skipped

No canvas. The library's Bank row origin (`ui:hHYs5` in `product-ui.lib.pen`)
instances the Avatar origin (`ui:Z2oYAy`), and Avatar defines only an initials
text child — no image variant exists anywhere in the design record. The row's
`logoUrl` image path is a Svelte-only addition the component's own code
comment already flags as invented ahead of the library: "The library has an
Avatar origin but no Svelte component yet, so the mark is inline here." There
is therefore no drawn appearance for a bank logo image to match or diverge
from; fixing its `object-fit` is an implementation-only correction with no
design counterpart to update.

Curating which banks populate the list is a data change to an existing list,
not a change to how the list or its rows are drawn.

The condition is the one the artifact's own instruction states: create a
canvas only when the change alters a screen, flow, task surface, state, or a
component's drawn appearance or behaviour. None of those apply here.

`product-ui.lib.pen` and every journey `.pen` are untouched, so
`just pen-manifest verify` should report nothing added and nothing changed
for the duration of this change.
