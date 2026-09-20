## Skipped

No canvas. The change adds no screen, flow, state or component to wimm. What
it adds is a developer's page whose purpose is to show the product's real
screens, and drawing that page in a `.pen` would be drawing the viewer rather
than anything it views. `product-ui.lib.pen` and every journey `.pen` are
untouched, so `just pen-manifest verify` should report nothing added and
nothing changed by this change.
