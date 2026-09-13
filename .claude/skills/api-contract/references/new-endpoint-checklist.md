# Adding an endpoint

## Contract

- [ ] Protocol decided by caller: Connect (ours) or REST (third parties).
- [ ] `.proto` or OpenAPI change written first; naming per `SKILL.md` rule 5.
- [ ] Request validation expressed in the contract (`buf validate` / OpenAPI schema).
- [ ] `buf lint` and `buf breaking` clean, or this is explicitly a new-major PR.
- [ ] Contract comment states: deadline (typical/max), idempotency behaviour, auth requirement, pagination if a list.

## Generation

- [ ] `just gen` run; generated code committed; nothing under `gen/` hand-edited.
- [ ] TypeScript / Python clients regenerated in the same PR if an app consumes them.

## Handler

- [ ] Implements the generated interface; converts, calls the domain, converts back — no business logic.
- [ ] Every failure mapped in the single boundary mapping; unmapped → `internal` with a generic message.
- [ ] Field violations returned as `BadRequest` details.
- [ ] `context.Context` passed to every downstream call; deadline enforced by the interceptor chain.
- [ ] Idempotency key honoured if the operation can be retried.
- [ ] Lists: `page_size` clamped, `page_token` opaque and validated, stable sort.

## Cross-cutting

- [ ] Auth interceptor covers it; permission check is inside the domain call.
- [ ] Telemetry: span name and attributes per the `observability` skill; confirm RED metrics appear.
- [ ] A test per error code the contract can return, not only the success case.
- [ ] Public REST only: OpenAPI examples added in the same PR.
