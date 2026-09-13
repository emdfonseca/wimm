# Versioning and compatibility

## Additive within a major

Inside `v1` you may: add a field (new number), add an RPC, add an enum value (clients must treat unknown values as unspecified), add an optional query parameter, relax validation. You may not: remove or rename anything, change a type, reuse a field number, change what an existing value means, tighten validation, change default behaviour.

`buf breaking --against '.git#branch=main'` catches the protobuf cases mechanically. OpenAPI has no equivalent enforcement in the box, which is one more reason the public REST surface is reviewed by a second person.

## Breaking means a new major package

```text
proto/org/billing/v1/   ← unchanged, still served
proto/org/billing/v2/   ← the new shape
```

Both are served for the migration window. `v2` is a copy-and-modify of `v1`, not a rewrite, so the diff shows exactly what changed. The service implements both for as long as `v1` has callers, and `v1` gets `[deprecated = true]` on the service with a comment naming `v2` and the sunset date.

For REST the path prefix changes (`/v1/` → `/v2/`) and the same window applies, with `Deprecation` and `Sunset` headers on `v1` responses.

## Deprecation is a process, not a flag

1. Mark deprecated in the contract with the replacement and date.
2. Log usage of the deprecated surface with the caller identity, so you know who has not moved.
3. Communicate — internal callers via the contract PR, external ones via the changelog and headers.
4. Remove only after the sunset date has passed and usage is zero (or explicitly accepted).

Skipping step 2 is how a sunset becomes an incident.

## What is not versioned

Internal implementation, storage, and the shape of things behind the contract. Versioning is a promise to callers about the contract; keep it there, and keep the contract narrow enough that the promise is cheap to keep.
