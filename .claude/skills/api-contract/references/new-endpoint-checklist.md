# Adding an endpoint

In order. The items at the end are the ones that get skipped once the happy path works, and they are the ones that page someone later.

## Contract

- [ ] Protocol decided by caller: Connect (ours) or REST (third parties). Not "REST because faster".
- [ ] `.proto` or OpenAPI change written first; naming follows `references/connect.md` / `references/rest-public.md`.
- [ ] Request validation expressed in the contract (`buf validate` constraints / OpenAPI schema), not only in code.
- [ ] `buf lint` and `buf breaking` clean, or this is explicitly a new-major PR.
- [ ] Contract comment states: deadline (typical/max), idempotency behaviour, auth requirement, pagination if a list.

## Generation

- [ ] `just gen` run; generated code committed; nothing under `gen/` hand-edited.
- [ ] TypeScript / Python clients regenerated in the same PR if the app consumes them.

## Handler

- [ ] Implements the generated interface; converts, calls the domain, converts back — no business logic.
- [ ] Every failure mapped to a code in the single boundary mapping; unmapped → `internal` with a generic message.
- [ ] Field violations returned as `BadRequest` details, not concatenated strings.
- [ ] `context.Context` passed to every downstream call; server deadline enforced by the interceptor chain.
- [ ] Idempotency key honoured if the operation can be retried.
- [ ] Lists: `page_size` clamped, `page_token` opaque and validated, stable sort.

## Cross-cutting

- [ ] Auth interceptor covers it; permission check is inside the domain call, not the handler.
- [ ] Telemetry: span name and attributes per the `observability` skill; the RED metrics come from the interceptor for free — confirm they appear.
- [ ] A test per error code the contract can return, not only the success case.
- [ ] Public REST only: OpenAPI examples added, changelog entry written, second reviewer from outside the team.
