## 1. Repository and toolchain

- [ ] 1.1 Add PostgreSQL to `devbox.json` and a `just db-up` / `just db-reset` pair that starts a local instance and applies migrations; verify `devbox run -- just db-up` leaves a reachable database and `just db-reset` returns it to empty.
- [ ] 1.2 Create the `apps/wimm` Go module with `cmd/wimmd` and `cmd/wimmctl` stubs and a `justfile` carrying the standard verbs; verify `just check apps/wimm` passes on the empty module.
- [ ] 1.3 Create the `apps/web` SvelteKit app with Svelte 5, importing `@wimm/ui` tokens, with a `justfile` carrying the standard verbs; verify `just check apps/web` passes and the dev server serves a page.
- [ ] 1.4 Proxy `/api` from the SvelteKit dev server to `wimmd` so the browser sees a single origin; verify a request to `/api` from the app's own origin reaches the service with no CORS preflight.

## 2. Decision record

- [ ] 2.1 Write `docs/decisions/0015-*.md` recording the enrolment-link trust model, the opaque server-side session, the relying-party identifier as a one-way door, the separated admin listener and the operator credential; verify `just adr-index-check` passes and the entry appears in `.claude/rules/decisions.md`.

## 3. Contracts

- [ ] 3.1 Define the public and admin Connect services as separate protobuf definitions in `packages/contracts` and wire `buf` into `just gen`; verify `just gen` is a no-op on a clean tree and `just check packages/contracts` passes.

## 4. Data

- [ ] 4.1 Write the goose migrations for members, enrolment links (hash only), credentials, ceremony challenges and sessions, owned by `wimmd`; verify an up/down round trip returns the schema to empty and `just check apps/wimm` passes.

## 5. Service

- [ ] 5.1 Load configuration for the relying-party identifier, expected origins, operator credential and link lifetime, and refuse to start when the admin listener is enabled without a credential; verify a test asserts the refusal and names the missing credential. Story S1.
- [ ] 5.2 Serve the admin service on its own listener, bound to loopback by default, authenticating the operator credential in constant time; verify a test asserts that a wrong and an absent credential are both refused and that nothing is written. Story S1.
- [ ] 5.3 Implement registering a member: reject a missing name, reject an already-registered email, mint a single-use link and return it with its expiry; verify tests cover all three and assert no link is issued on either refusal. Story S1.
- [ ] 5.4 Implement issuing a further link to a registered member, invalidating any outstanding link and leaving existing credentials working; verify tests assert the earlier link stops working, the credentials do not, and an unknown email is refused. Story S1.
- [ ] 5.5 Implement the enrolment ceremony: challenge stored as a row with a seconds-long lifetime deleted on use, discoverable credential and user verification required, and a registration rejected when the authenticator reports no discoverable credential; verify a test drives a rejection and asserts the member stays un-enrolled with the link still usable. Story S2.
- [ ] 5.6 Spend the link on first load and redirect to a path that does not carry its value, and ensure no log line records the enrolment path with the value in it; verify a test asserts the redirect and a log assertion covers the value never being emitted. Story S2.
- [ ] 5.7 Return one indistinguishable response for an expired, spent, replaced or never-issued link, in body, status and timing; verify a test asserts the four cases are identical. Story S2.
- [ ] 5.8 Create the session on successful enrolment and set the cookie `HttpOnly`, `SameSite=Lax`, and `Secure` when the origin is HTTPS, with lifetimes evaluated against database time; verify a test asserts the attributes and that a skewed process clock does not extend a session. Story S2.
- [ ] 5.9 Implement username-less sign-in, refusing a credential with no record, and sign-out that stops the session working; verify tests cover a successful sign-in, an unknown credential, and that a signed-out session identifier is rejected afterwards. Story S3.
- [ ] 5.10 Send an unauthenticated request to sign-in and return the member to the path they were trying to reach, holding that path server-side against the attempt; verify a test asserts the return path is honoured and that a path supplied by the client is ignored. Story S3.

## 6. Design system components

Each builds the origin in `product-ui.lib.pen` and the Svelte component in `packages/ui/src` together. All three precede section 7, which only composes instances.

- [ ] 6.1 Build the Svelte **Auth shell** in `packages/ui/src` matching origins `kJkV1` and `tqiBA` — centred single-card page, brand lockup at the card head, 480 at Wide and 342 at Compact, slots for feedback, heading, body, action and footnote; verify `just check packages/ui` passes and each slot is settable by prop. The pen origin already exists.
- [ ] 6.2 Build the Svelte **Signed-in landing** matching origins `IGbQe` and `pu6qZ` — header carrying the member's name, empty body. No navigation: there is nowhere to navigate yet; verify `just check packages/ui` passes.
- [ ] 6.3 Build the **Button · pending preset** wrapping a Button instance: non-interactive, own label, accessible status announced when it becomes active and when it resolves; verify `just check packages/ui` passes and the preset redraws none of Button's own chrome.
- [ ] 6.4 Build the **Notice · tone presets** for info and error, wrapping a Notice instance against `color-feedback-*`; verify `just check packages/ui` passes, including the lint that refuses a raw hex outside `tokens.css`.

## 7. Web app screens

- [ ] 7.1 Build `/enrol/<link>` with its Default, Passkey not saved and Prompt dismissed states, composing instances of 6.1–6.4; verify `just check apps/web` passes and each state is reachable by setting props. Story S2.
- [ ] 7.2 Build the unusable-link dead end with no path forward and the instruction to ask the operator; verify `just check apps/web` passes and the page offers no enrolment control. Story S2.
- [ ] 7.3 Build `/signin` with its Default, Prompt dismissed and Session expired states; verify `just check apps/web` passes and each state is reachable by setting props. Story S3.
- [ ] 7.4 Build the unrecognised-passkey dead end (J02.B): the member offered a credential this instance has no record of, nothing on the page recovers it, and the only way forward is an enrolment link from the operator; verify `just check apps/web` passes and the page offers no route back into the product. Story S3.
- [ ] 7.5 Build the signed-in landing: `Sidebar nav / in shell` with Overview current and every other destination absent, the member's name in the sidebar footer, and an empty body. No page header: it has no breadcrumbs, tabs or action to carry, and the sidebar already marks the location; verify `just check apps/web` passes. Stories S2, S3.
- [ ] 7.6 Implement focus and announcement as recorded in canvas.md — focus to the card heading on load, assertive error notices, polite info notices, neither moving focus; verify a test drives each state and asserts the focus target and the live-region politeness. Stories S2, S3.

## 8. End-to-end verification

- [ ] 8.1 Register a member with `wimmctl`, open the link, enrol a passkey from a phone over the cross-device flow, and confirm the landing names them; verify by doing it against a running instance, not by a unit test. Stories S1, S2.
- [ ] 8.2 Sign out, sign back in with no identifier typed, then expire the session and confirm the return-to-page promise holds; verify by doing it against a running instance. Story S3.
- [ ] 8.3 Re-export the zone 40 dark frames if any source frame changed during implementation, and confirm the canvas still matches what was built; verify `git status` shows the journey file saved and canvas.md's frame table still resolves.
