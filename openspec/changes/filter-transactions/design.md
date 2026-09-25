## Context

See proposal.md for why. What the approach has to fit:

- `apps/wimm/internal/store/transactions.go` reads the ledger three ways: `Ledger`
  (a keyset page on `(booking_date, id)`), `CountLedger` (the toolbar's number)
  and `LedgerPageIndex` (every real page, for the scrubber). All three scope by
  the `ownedAccounts` fragment, which joins `account_owners` and drops left-out
  accounts, and take an optional account id.
- `apps/wimm/internal/banking/ledger.go` assembles one `Ledger` from those reads
  plus `LedgerState` (last synced, reaches back to), the narrow banks, the
  owned-account labels and the patterns (unusual, own transfer), which are judged
  under All over the page's dates and never over the filtered rows.
- The route `apps/web/src/routes/(app)/transactions/+page.server.ts` reads
  `account`, `before`, `after`, `oldest` and `page` from the address, carries
  `account` on every pager link through `paramsFor`, and ignores `month`.
- `packages/ui/src/pages/TransactionsScreen.svelte` names a narrowed account in
  an inline `.chip` with a Show all accounts link. It has no search field, no
  select and no filter bar.
- A malformed `?account=` reaches `$2::uuid` in Postgres and fails; the route's
  catch-all then renders the no-bank state, `Connect a bank`. This change must
  not add more ways into that path.

## Language

- **filter**: one condition narrowing the ledger: the account, the search, the month or the direction. Never "facet", "query" or "view".
- **search**: the free text a member types, matched against the other party and the bank's line. Address parameter `q`. Never "keyword" or "query".
- **money in / money out**: the direction, from the sign of the amount. Address values `in` and `out`, proto `LEDGER_DIRECTION_IN` and `_OUT`. Never "credit", "debit", "income" or "expense".
- **month**: a calendar month of booking dates, `YYYY-MM` in the address and on the wire. Never "period". A month is not a **page**: a page is what the scrubber offers, and a month can hold several pages or share one.
- **Clear filters**: the one action that removes every filter. Never "reset".
- **nothing matches**: the state where filters are in force and no transaction matches them. Never "empty", which the spec already uses for a ledger with nothing read.

## Goals / Non-Goals

**Goals:**

- The four filters narrow exactly the rows `ownedAccounts` already scopes, and the count, the page index, the months offered and the page all agree because they share one predicate.
- Every filter lives in the address, so paging, Back, bookmarks and Overview's month link work with no client state.

**Non-Goals:**

- Ignoring accents. It needs `unaccent`, which is an extension and an ADR.
- Any index. A household ledger is a few thousand rows; a scan over one member's owned accounts is enough.
- Filtering by unusual, own transfer or not settled. The first two are derived on read over a window of dates, so a page of them cannot be sought by keyset.

## Decisions

### 1. One filter predicate, shared by four reads

`store.LedgerFilter` holds `AccountID`, `Search`, `Month` and `Direction`.
`LedgerQuery` embeds it, and `CountLedger`, `LedgerPageIndex` and a new
`LedgerMonths` take it in place of the bare account id. One SQL fragment,
`ledgerScope`, carries ownership and every filter at fixed parameter positions
(`$1` member, `$2` account, `$3` search pattern, `$4` first day of the month,
`$5` direction), each `null` when not in force. Every read appends its own
parameters from `$6`.

```text
account_id in (ownedAccounts)
and ($3::text is null or counterparty_name ilike $3 escape '\' or remittance ilike $3 escape '\')
and ($4::date is null or (booking_date >= $4 and booking_date < $4 + interval '1 month'))
and ($5::text is null or ($5 = 'in' and amount_minor > 0) or ($5 = 'out' and amount_minor < 0))
```

This is the reason `ownedAccounts` was a fragment: widening or narrowing the scope
is one edit in one place. Four copies of the predicate would let the count
disagree with the rows the first time one is edited.

*Alternative:* filter in Go after reading. Rejected: it breaks keyset paging, a
page would hold fewer than 50 rows, and the count would need every row read.

### 2. Search is `ilike` over two columns, with the pattern escaped in Go

Go trims the text, escapes `\`, `%` and `_`, and wraps it in `%…%`. Both
columns are compared under the built-in `und-x-icu` collation, because the
database's collation is `C`, whose case folding stops at ASCII: `café` would miss
`CAFÉ`. ICU ships with Postgres; it is not an extension. Searching
reads `counterparty_name` and `remittance`, which is what the bank wrote. The
display name is derived from those on read and is not a column, so a name wimm
invented for a blank transaction (`Card payment`) is not found; the spec says so.

A search longer than 100 characters is refused by wimmd with `InvalidArgument`
and dropped by the route before it gets there. The field's `maxlength` is 100.

*Alternative:* `pg_trgm` with a GIN index. Rejected for now: an extension, an
index and an ADR for a ledger that scans in milliseconds.

### 3. The contract

`ListTransactionsRequest` gains `string search = 7`, `string month = 8`
(`YYYY-MM`, empty for none) and `LedgerDirection direction = 9`, with
`LEDGER_DIRECTION_UNSPECIFIED = 0` meaning both. `Ledger` gains
`repeated FilterAccount filter_accounts = 13` (`account_id`, `name`,
`bank_name`) and `repeated string months = 14`, newest first.
`RefreshTransactionsRequest` is unchanged: the page re-reads through its own load
after a refresh, and the refresh's response is discarded.

The accounts come from wimmd, from the same `OwnedAccountLabels` read the rows
are labelled with. *Alternative:* build the picker in the route from
`ListAccounts`. Rejected: that list is what a member may *see*, including
accounts granted to them, and restating "owned and not left out" in TypeScript
is a second copy of the scope rule in a second language.

A month or direction wimmd cannot parse is `InvalidArgument`. The route never
sends one (decision 5), so this is defence, not the path a member takes.

### 4. The months offered are read with the month removed

`LedgerMonths` runs `ledgerScope` with `Month` cleared and returns the distinct
`date_trunc('month', booking_date)`, newest first. So the months follow the
account, the search and the direction, and every month offered holds a match.
The route adds the chosen month to the list when it is not there, so a month in
force is always shown as chosen.

### 5. The route owns the address, and drops what it cannot use

Parameters: `account`, `q`, `month`, `direction`, plus the existing `before`,
`after`, `oldest` and `page`. Before calling wimmd the route drops an `account`
that is not a uuid, a `q` that trims to nothing or runs past 100 characters, a
`month` not matching `^\d{4}-(0[1-9]|1[0-2])$`, and a `direction` other than `in`
or `out`. After the call, an `account` missing from `filter_accounts` is dropped
too. Anything dropped is a `303` to the same address without it, so the address
always says what is on screen and a stranger's account id reads exactly like a
nonexistent one.

`paramsFor` carries every filter on every pager and scrubber link. A filter
change drops `before`, `after`, `oldest` and `page`, which is what starts reading
again at the newest match.

### 6. Filters are a GET form; the screen emits them as one callback

The filter bar is a `<form method="get" action="/transactions">` whose fields
carry the address names, so Enter in the search field works as a plain
navigation. The screen takes `onfilter(next, { live })` and, when given, the bar
calls it instead of submitting: on a 300 ms pause in typing a search (`live`), on
Enter in search, on clearing the search, on a select changing, and on the
direction changing. The route turns `next` into an address and calls
`goto(url, { keepFocus: true, noScroll: true })`, so the control the member just
used keeps focus and the list is announced (canvas.md Contracts).

The search applies as the member types, because that is what a member expects of
a search box. Two costs follow, and each is bounded:

- **History.** The first live change of a search pushes an entry and every live
  change after it replaces that entry (`replaceState`), so Back goes to before
  the search rather than stepping through half-typed words. Any navigation the
  page did not make ends the run.
- **Reads.** Each pause is one load, four scans. The box reports only when the
  trimmed text changes, and a pause does not trigger the arrival sync
  (decision 7).

Letters typed while a live search is on its way are kept: the box only takes the
value from the address when it differs from what the box itself last sent, which
is how Back and Clear filters still reset it.

*Alternative:* apply the search on Enter only. Rejected by the member: a search
box that waits for Enter reads as broken.

### 7. A filtered view syncs exactly when an unfiltered one does

`syncOnArrival` keeps its condition: no `before`, `after`, `oldest` or `page`.
Filters do not change it. A filtered newest page is still the newest end, where a
sync inserts, and the sync already reads every owned account whatever `account`
says. This is the simplest option: no new condition, and a member who opens
August from Overview gets a late-booked August row the same as anyone.

A pause in typing a search is not an arrival and does not sync: the search's
first result already was one, and the sync is bounded by the per-account interval
anyway.

*Alternative:* never sync a filtered view. Rejected: it adds a condition, and a
member narrowed to one account would stop seeing new rows arrive, which today
they do.

### 8. The screen

- `.chip` and the "Showing one account" row go. The account select names the account it is narrowed to, which is what the spec asks; a chip beside it would say it twice.
- `filterAccount` and `showAllHref` give way to `filters` (the values in force), `accounts`, `months` and `clearHref`.
- A new empty reason, `no-match`, sits after `no-bank` and `owns-nothing` and before every other: a member with rows held and filters in force is told nothing matches, not that a bank is narrow or unread. The filter bar stays above it.
- The Refresh button shows whenever filters are in force, since the ledger behind them holds rows.
- The oldest page reads `Nothing older matches these filters.` when a search, month or direction is in force. With the account alone it keeps `Nothing older. This is as far back as the bank would go.`, which is still true.
- The freshness line keeps stating the whole ledger's reach. `LedgerState` stays scoped to the account only. It sits beside Refresh in the page header, not in the list, and there is no lede: the space above the ledger holds one filter row and nothing else.
- At compact the account, month and direction sit in a native `<dialog>` sheet behind a `Filters` button, so the list starts on the first screen. The dialog gives focus trapping, Escape and the backdrop without a hand-built modal.
- The totals share the count's line; what they leave out is a note behind an info button, because it is for checking the figures, not for reading every time.

### 9. What a filtered list adds up to

`Ledger` gains `repeated LedgerTotal totals = 15` (`currency`, `money_in`,
`money_out`, the second negative), `int32 transfers_left_out = 16` and
`int32 not_settled = 17`, empty unless a filter is in force.

wimmd reads every matching row with `LedgerRows` (`ledgerScope`, no page), then
reads the member's booked rows across every owned account over the matching
dates widened by `ownTransferWindowDays`, and pairs them with `OwnTransfers`. A
row is left out when it is half of a pair whose two accounts are both in the
ledger's scope: every account on the ledger, or the one it is narrowed to.
Unsettled rows are counted in `not_settled` and in no figure. This is the
month summary's rule (ADR 0026) with the ledger's scope in place of a group's,
so August here is August on Overview; the one place they can part is an owned
account in neither the household group nor the member's own (ADR 0024), which
Overview's All leaves out of its figures and the ledger lists.

The sum is taken in Go, not SQL, for the month summary's reason: SQL cannot
leave out a row whose partner it has not read.

*Alternative:* sum every listed row. Rejected by the member: August would read
differently here and on Overview whenever money moved between their accounts.

### 10. No ADR

None is needed: no dependency, service or migration, and the proto fields are
additive to a server-to-server contract whose only client ships in the same
deploy (ADR 0001), so they are not a public contract.

## Risks / Trade-offs

- [`ilike` folds case by the collation's ctype, and the database's is `C`] → The search is compared under `und-x-icu` (decision 2), and the store test searches `café` against `CAFÉ`. A Postgres built without ICU fails that test rather than shipping a narrower rule.
- [Four scans per page view instead of three, each with the filter predicate] → The wiring task times a filtered load on the household ledger and records it in the commit. An index needs an ADR and is not taken here.
- [A cursor from one filter used under another] → Cursors are sort keys, not offsets, so they still seek correctly; at worst the page starts somewhere the member did not expect. Changing a filter drops the cursor, so only a hand-edited address does this.
- [The filter bar costs height at compact, above a list that scrolls in place] → The look task judges it at 390 wide; canvas.md names it as the thing to look at.
- [The route's catch-all still renders any wimmd failure as "no bank"] → Out of scope; decision 5 keeps filters from reaching it.

## Migration Plan

Additive proto fields and no schema change. Deploy wimmd and web together, as
always. Rollback is a revert; an old address with `q` or `direction` is ignored
by the old route.
