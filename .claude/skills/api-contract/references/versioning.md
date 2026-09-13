# Versioning

## Additive within a major

Inside `v1` you may: add a field (new number), add an RPC, add an enum value, add an optional query parameter, relax validation.

You may not: remove or rename anything, change a type, reuse a field number, change what an existing value means, tighten validation, change default behaviour.

`buf breaking --against '.git#branch=main'` enforces this for protobuf in CI. OpenAPI has no mechanical check; review does it.

## Breaking means a new major package

`proto/<org>/<domain>/v2/` is a copy-and-modify of `v1`, served alongside it while `v1` has callers. `v1` gets `[deprecated = true]` with a comment naming `v2` and the sunset date. REST: `/v2/` prefix, `Deprecation` and `Sunset` headers on `v1` responses.
