## 1. Canvas follow-ups

Recorded as gaps in canvas.md rather than rushed there; each is a
prerequisite for the frames or screens it blocks, so do these before the
work that depends on them.

- [x] 1.1 Open `ui:q7dY7c` (Sidebar nav / in shell) and `ui:b3MTx` (Bottom
      nav) in `product-ui.lib.pen`, determine whether either already
      supports more than the 2 items every drawn frame currently renders,
      and add a 4th item, Accounts, with a new icon (landmark or
      equivalent). Verify by exporting a frame using each and confirming 4
      items render, all 4 labelled.
- [x] 1.2 Resolve the Compact recent-transactions row: either fix `ui:VE2z9`
      (Ledger row)'s sizing so its Amount column does not clip at 390px, or
      switch Overview's compact frame to `ui:wUywT` (Transaction
      row/compact), which the library already has for exactly this width.
      Verify with an exported screenshot at 390px showing no clipped text.
- [x] 1.3 Draw Overview's Medium and Ultra regimes (both `As it opens` and
      `Before any bank is connected`), and the three states canvas.md
      recorded as not drawn: a member who owns no account (no
      recent-transactions section), an account with balance-only access
      contributing to the total but not the trend, and no transaction
      history anywhere (no trend at all). Verify each against its scenario
      in `banking/overview`.
- [x] 1.4 Pass over the 22 renamed Accounts frames in `02-banking.pen`: fix
      each frame's internal page-title text from "Overview" to "Accounts"
      (e.g. node `ffDzD` on J09.A/05), and remove the Household total tile
      from frames that still show one, since Accounts no longer shows a
      total. Verify by exporting each and confirming no "Overview" text and
      no total tile remain.

## 2. Components (packages/ui)

- [x] 2.1 Build `MetricTile.svelte` (atoms or molecules, per
      `.claude/skills/storybook`) matching `ui:oweC0`'s fields — label,
      value, optional delta — with a `.stories.svelte` covering positive
      delta, negative delta, and no delta. Verify with `just check
      packages/ui`.
- [x] 2.2 Build `TrendSparkline.svelte` matching `ui:K3w7Lx` — a row of bars
      plus a caption, no line, no charting dependency — with a
      `.stories.svelte` covering an empty series and a populated one.
      Verify with `just check packages/ui`.
- [x] 2.3 Whichever component task 1.2 resolved on (`LedgerRow` fix, or a
      new compact row component with its own story), build or fix it here.
      Verify with `just check packages/ui`.
- [x] 2.4 Add a 4th `Destination['icon']` value to
      `packages/ui/src/destinations.ts` and its glyph path data to
      `Icon.svelte`, matching how the existing nine were added. Add
      `Accounts` to the `destinations()` list. Verify `just check
      packages/ui` and a Storybook look at `SidebarNav`/`BottomNav` with 4
      items.

## 3. Backend (apps/wimm)

- [x] 3.1 Add a wimmd RPC method (or extend `ListAccounts`) that returns,
      per currency, trend points computed by walking each owned account's
      booked transactions backward from `balance_minor`, bounded by
      `transactions_synced_through`, per design.md's "trend is computed in
      wimmd" decision. Verify with a new `banking` package test asserting
      the walk against a fixture with a known bound.
- [x] 3.2 Confirm `ListTransactions` already serves what Overview's recent
      slice needs (household-wide, owner-only, first page, small size) with
      no new endpoint; add a test if the first-page/small-size path is not
      already covered. Verify with `just check apps/wimm`.
- [x] 3.3 Regenerate the Connect contract (`just gen`) for whatever 3.1
      added, and update `apps/wimm/internal/rpc/banking.go`'s conversion
      functions. Verify with `just check apps/wimm`.

## 4. Web routes

- [x] 4.1 Create `apps/web/src/routes/(app)/accounts/+page.server.ts` and
      `+page.svelte`, moving the account-list/bank-card logic and the
      connect/restore/disconnect/who-sees-this wiring out of
      `(app)/+page.server.ts`/`+page.svelte` largely unchanged (design.md:
      "the account-list and bank-card markup moves as a block"). Verify
      with the existing `overview.test.ts`'s assertions ported to a new
      `accounts.test.ts`, passing under `just check apps/web`.
- [x] 4.2 Rewrite `(app)/+page.server.ts` and `+page.svelte` as the Overview
      dashboard: total (via `banking.listAccounts`), recent transactions
      (via `banking.listTransactions`), trend (via 3.1's RPC), composed with
      `MetricTile`, the row component from 2.3, and `TrendSparkline`. No
      account list, no connection actions. Verify with a new
      `overview.test.ts` covering the scenarios in `banking/overview`,
      passing under `just check apps/web`.
- [x] 4.3 Wire the before-any-bank-connected and owns-no-account empty
      states on Overview, matching canvas.md's Contracts section (icon,
      copy, "Go to Accounts" button, no connection action performed by
      Overview itself). Verify against `banking/overview`'s two empty-state
      scenarios.

## 5. Verification

- [x] 5.1 Run `just check apps/web`, `just check apps/wimm`, and `just check
      packages/ui`; all three pass.
- [ ] 5.2 Run the app (`/run` or the project's own dev-server skill) and
      walk both screens at Compact and Wide: total/recent/trend on
      Overview, account list and every connection action on Accounts,
      confirming no screen shows the other's content. Not done in this
      session: `wimmd`/the web dev server started cleanly (confirmed by log
      and a reachability check) but this sandbox kills backgrounded
      long-running processes, and passkey sign-in cannot be automated
      headlessly. Storybook's 250 interaction tests and the canvas export
      screenshots (canvas.md) are the evidence in their place; still needs a
      human pass, or an interactive session, through the actual signed-in
      app.
