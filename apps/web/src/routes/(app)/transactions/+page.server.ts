import { redirect, type Cookies, type ServerLoad } from '@sveltejs/kit';
import type {
	FilterOption,
	LedgerDay,
	LedgerFigures,
	LedgerFilterValues,
	LedgerProblem,
	NarrowBank,
	ScrubberPage
} from '@wimm/ui';
import {
	Failure,
	LedgerDirection,
	TransactionStatus,
	type Ledger,
	type PageMarker,
	type NarrowConnection,
	type FilterAccount,
	type BankFailure,
	type LedgerTotal,
	type Transaction
} from '@wimm/contracts/banking';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';
import { formatMoney, isNegative } from '$lib/money';
import { filterParams, readFilters, without } from './filters';

/**
 * J07 · Transactions. The ledger, and the same page narrowed by its filters:
 * an account, a search, a month and a direction, each carried in the address.
 *
 * Paging is a route change rather than a control that mutates in place, so Back
 * works and a page is somewhere a member can return to. The cursor is the sort
 * key itself: `?before=<date>.<id>` reads away from today and `?after=` reads
 * back towards it.
 *
 * **This load never asks a bank.** A first fill is unbounded pages across every
 * account, and a load function blocks navigation until it returns — so syncing
 * here means clicking Transactions and watching Overview until every bank has
 * answered. The stored rows render immediately with the date they are synced
 * through, and the page runs the sync behind that arrival (design.md decision 1).
 *
 * A page a member paged to does not sync at all. A sync inserts at the newest
 * end, and asking the banks while somebody is reading August would move the
 * ground under them for no benefit.
 */
export const load: ServerLoad = async ({ cookies, depends, url }) => {
	// Named so the page can re-read the ledger alone when a sync lands.
	// invalidateAll() would re-run the layout's session check on every pass too,
	// which is a second round trip for an answer that has not changed.
	depends('wimm:ledger');

	// A filter wimm cannot use is dropped from the address before anything is
	// asked, so the address always says what is on screen.
	const { filters, dropped } = readFilters(url.searchParams);
	if (dropped.length > 0) redirect(303, without(url, dropped));

	const before = url.searchParams.get('before') ?? '';
	const after = url.searchParams.get('after') ?? '';
	const oldest = url.searchParams.get('oldest') === '1';
	const page = url.searchParams.get('page') ?? '';
	const query = { filters, before, after, oldest, page };

	const { ledger, noBank } = await read(cookies, query);

	// An account the member cannot narrow to — a stranger's, a left-out one, or
	// none at all — is dropped the same way in every case, so the address never
	// says which it was.
	if (
		ledger &&
		filters.account &&
		!ledger.filterAccounts.some((a) => a.accountId === filters.account)
	) {
		redirect(303, without(url, ['account']));
	}

	return {
		...present(ledger, query, noBank),
		accountId: filters.account || undefined,
		// Only the newest page brings itself up to date, and only when nothing
		// else named where to start reading — a jump is a member asking to be
		// somewhere specific, not an arrival at the newest page.
		syncOnArrival: !before && !after && !oldest && !page
	};
};

interface Query {
	filters: LedgerFilterValues;
	before: string;
	after: string;
	oldest: boolean;
	// A page named by the ledger's own page index — `<booking date>.<id>`,
	// the same shape `before`/`after` already take, because it is the same
	// kind of thing: a cursor, never a calendar bound.
	page: string;
}

/**
 * Banking is optional: an instance with no gateway configured answers
 * Unimplemented, and the screen shows its empty state rather than an error.
 */
async function read(
	cookies: Cookies,
	query: Query
): Promise<{ ledger: Ledger | undefined; noBank: boolean }> {
	try {
		// Cursor, oldest and page are mutually exclusive ways to say where to
		// start reading, and a cursor wins if somehow more than one arrived —
		// the contract's own stated precedence.
		const cursor = parseCursor(query.before || query.after);
		const pageStart = parseCursor(query.page);
		const [response, accounts] = await Promise.all([
			call(cookies, (options) =>
				banking.listTransactions(
					{
						accountId: query.filters.account,
						search: query.filters.q,
						month: query.filters.month,
						direction: directionOf(query.filters.direction),
						cursor,
						older: query.before !== '',
						oldest: !cursor && query.oldest,
						pageStart: !cursor && !query.oldest ? pageStart : undefined,
						// Stored rows only. The sync is the page's own, after it
						// has rendered.
						skipSync: true
					},
					options
				)
			),
			// The one thing the ledger cannot say. "No bank connected" and
			// "you own no account" are both an empty list, and the screen tells
			// them apart because they are different things for a member to do
			// something about. Reads nothing from the banks.
			call(cookies, (options) => banking.listAccounts({ skipRead: true }, options))
		]);
		return { ledger: response.ledger, noBank: accounts.accounts.length === 0 };
	} catch {
		return { ledger: undefined, noBank: true };
	}
}

function directionOf(direction: LedgerFilterValues['direction']): LedgerDirection {
	if (direction === 'in') return LedgerDirection.IN;
	if (direction === 'out') return LedgerDirection.OUT;
	return LedgerDirection.UNSPECIFIED;
}

/**
 * A cursor is `<booking date>.<transaction id>` — the sort key, rendered.
 *
 * It is opaque to the member and is not trusted here: a cursor naming an
 * account they do not own simply seeks past nothing, because the read is scoped
 * to ownership in wimmd rather than by this string.
 */
function parseCursor(raw: string) {
	if (!raw) return undefined;

	const separator = raw.indexOf('.');
	if (separator <= 0) return undefined;

	const date = Date.parse(raw.slice(0, separator));
	const transactionId = raw.slice(separator + 1);
	if (Number.isNaN(date) || !transactionId) return undefined;

	return { bookingDate: { seconds: BigInt(Math.floor(date / 1000)), nanos: 0 }, transactionId };
}

function present(ledger: Ledger | undefined, query: Query, noBank: boolean) {
	const transactions = ledger?.transactions ?? [];

	// Every link that moves through the list carries every filter.
	const paramsFor = (extra: Record<string, string>) => {
		const params = filterParams(query.filters);
		for (const [key, value] of Object.entries(extra)) params.set(key, value);
		return params;
	};

	return {
		days: group(transactions),
		count: ledger?.totalCount ?? 0,
		freshness: freshnessOf(ledger),
		span: spanOf(ledger),
		// Only said on a page that has reached the end, and only once something
		// has been read: before the first read wimm knows no date to reach to.
		atOldest: Boolean(
			ledger && !ledger.hasOlder && transactions.length > 0 && ledger.reachesBackTo
		),
		filters: query.filters,
		accounts: accountsFrom(ledger),
		months: monthsFrom(ledger, query.filters.month),
		clearHref: '/transactions',
		totals: totalsFrom(ledger),
		transfersLeftOut: ledger?.transfersLeftOut ?? 0,
		notSettled: ledger?.notSettled ?? 0,
		narrow: narrowFrom(ledger),
		problems: problemsFrom(ledger),
		// A member who can see nothing at all is in a household with no bank; one
		// who can see accounts but owns none is told that instead.
		noBank,
		ownsNothing: !noBank && (ledger?.ownsNoAccount ?? false),
		pages: pagesFrom(ledger, paramsFor),
		currentPage: currentPageKey(ledger),
		// Absent exactly where the page is already there: the newest page has
		// no cursor, no oldest flag and no page, and the oldest page is the
		// one place hasOlder says there is nothing further.
		newestHref:
			query.before || query.after || query.oldest || query.page
				? `/transactions?${paramsFor({})}`
				: undefined,
		oldestHref: ledger?.hasOlder ? `/transactions?${paramsFor({ oldest: '1' })}` : undefined
	};
}

/** `<booking date>.<id>`, the cursor a page's own newest row encodes to. */
function cursorKey(cursor: PageMarker['cursor']): string {
	const seconds = cursor?.bookingDate?.seconds;
	return seconds === undefined ? '' : `${isoDate(seconds)}.${cursor?.transactionId ?? ''}`;
}

function pagesFrom(
	ledger: Ledger | undefined,
	paramsFor: (extra: Record<string, string>) => URLSearchParams
): ScrubberPage[] {
	return (ledger?.pages ?? []).map((p: PageMarker) => ({
		key: cursorKey(p.cursor),
		label: spanLabel(p.newest?.seconds, p.oldest?.seconds),
		href: `/transactions?${paramsFor({ page: cursorKey(p.cursor) })}`
	}));
}

/** Which of the ledger's own pages this response's page actually is, so the
 *  scrubber can say so without guessing from the query string — the page a
 *  cursor names and the page a member is shown are the same fact stated
 *  twice, and only one of them should have to know how a cursor is spelled. */
function currentPageKey(ledger: Ledger | undefined): string | undefined {
	const newest = ledger?.newestOnPage?.seconds;
	if (newest === undefined) return undefined;
	const match = (ledger?.pages ?? []).find((p) => p.newest?.seconds === newest);
	return match ? cursorKey(match.cursor) : undefined;
}

function accountsFrom(ledger: Ledger | undefined): FilterOption[] {
	return (ledger?.filterAccounts ?? []).map((a: FilterAccount) => ({
		value: a.accountId,
		label: a.bankName ? `${a.name} · ${a.bankName}` : a.name
	}));
}

/**
 * What the filtered list adds up to, one entry per currency, each written with
 * its sign. wimmd sends none unless a filter is in force.
 */
function totalsFrom(ledger: Ledger | undefined): LedgerFigures[] {
	return (ledger?.totals ?? []).map((total: LedgerTotal) => ({
		moneyIn: formatMoney(total.moneyIn, { signed: true }) ?? '',
		moneyOut: formatMoney(total.moneyOut, { signed: true }) ?? ''
	}));
}

/**
 * The months offered, newest first, named as `August 2026` in UTC like every
 * booking day. A chosen month is always offered, even where the other filters
 * leave nothing in it, so the member can see it and remove it.
 */
function monthsFrom(ledger: Ledger | undefined, chosen: string): FilterOption[] {
	const months = new Set(ledger?.months ?? []);
	if (chosen) months.add(chosen);
	return [...months]
		.sort((a, b) => b.localeCompare(a))
		.map((value) => {
			const [year, month] = value.split('-').map(Number);
			return {
				value,
				label: new Date(Date.UTC(year!, month! - 1, 1)).toLocaleDateString('en-GB', {
					month: 'long',
					year: 'numeric',
					timeZone: 'UTC'
				})
			};
		});
}

/**
 * Grouping is presentation over a flat ordered list, so a page simply starts
 * mid-day and the heading is repeated. Paging by whole days would make page
 * size unpredictable and a single busy day unbounded.
 */
function group(transactions: Transaction[]): LedgerDay[] {
	const days: LedgerDay[] = [];

	for (const transaction of transactions) {
		const seconds = transaction.bookingDate?.seconds;
		const date = seconds === undefined ? '' : isoDate(seconds);
		const last = days[days.length - 1];

		const entry = {
			id: transaction.id,
			description: transaction.displayName,
			// The bank's line, as the bank wrote it; the screen shows it only
			// where it differs from the name.
			banksLine: transaction.remittance || undefined,
			account: transaction.accountName,
			amount: formatMoney(transaction.amount, { signed: true }) ?? '',
			date: shortDate(seconds),
			negative: isNegative(transaction.amount),
			unsettled: transaction.status === TransactionStatus.PENDING,
			unusual: transaction.unusual,
			// The label takes the status slot from unusual; under All every
			// pair is inside the scope, so wimmd never sends both (ADR 0026).
			transfer: transaction.ownTransfer
		};

		if (last && last.date === date) last.entries.push(entry);
		else days.push({ date, entries: [entry] });
	}

	return days;
}

function narrowFrom(ledger: Ledger | undefined): NarrowBank[] {
	return (ledger?.narrowConnections ?? []).map((connection: NarrowConnection) => ({
		connectionId: connection.connectionId,
		bankName: connection.bankName,
		widenHref: `/connect/widen/${connection.connectionId}`
	}));
}

function problemsFrom(ledger: Ledger | undefined): LedgerProblem[] {
	return (ledger?.failures ?? []).map((failure: BankFailure) => ({
		connectionId: failure.connectionId,
		bankName: failure.bankName,
		kind: kindOf(failure.failure),
		retryAfter:
			failure.retryAfterSeconds > 0n
				? new Date(Date.now() + Number(failure.retryAfterSeconds) * 1000).toLocaleTimeString(
						undefined,
						{ hour: '2-digit', minute: '2-digit' }
					)
				: undefined,
		lastReadAt: ledger?.syncedAt ? clockTime(ledger.syncedAt.seconds) : undefined,
		restoreHref: `/transactions?restore=${failure.connectionId}`
	}));
}

/**
 * Access having run out routes to restoring; everything else is transient and
 * routes to "did not answer".
 */
function kindOf(failure: Failure): LedgerProblem['kind'] {
	if (failure === Failure.CONSENT_EXPIRED) return 'access-ended';
	if (failure === Failure.RATE_LIMITED) return 'rate-limited';
	return 'unreachable';
}

/**
 * What the list says about its own currency. Absent before the first read: no
 * bank says in advance how much history it will hand over, so wimm states the
 * date it reached rather than promising a period.
 */
function freshnessOf(ledger: Ledger | undefined): string | undefined {
	if (!ledger?.syncedAt) return undefined;

	const updated = `Updated at ${clockTime(ledger.syncedAt.seconds)}.`;
	if (!ledger.reachesBackTo) return updated;

	return `${updated} Reaching back to ${longDate(ledger.reachesBackTo.seconds)}.`;
}

/** Where in time the member is. Never a page number. */
function spanOf(ledger: Ledger | undefined): string | undefined {
	return spanLabel(ledger?.newestOnPage?.seconds, ledger?.oldestOnPage?.seconds);
}

/** Where in time a span of dates sits. Shared by the toolbar's own span and
 *  every page the scrubber offers, because a page is described exactly the
 *  way the page a member is already on is. */
function spanLabel(newest: bigint | undefined, oldest: bigint | undefined): string {
	if (newest === undefined || oldest === undefined) return '';

	if (newest === oldest) return longDate(newest);

	const from = new Date(Number(newest) * 1000);
	const to = new Date(Number(oldest) * 1000);
	// The year is said once when both ends share it: "4 August to 31 July 2026"
	// is the span the frame draws.
	const options: Intl.DateTimeFormatOptions =
		from.getUTCFullYear() === to.getUTCFullYear()
			? { day: 'numeric', month: 'long', timeZone: 'UTC' }
			: { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC' };

	return `${from.toLocaleDateString(undefined, options)} to ${longDate(oldest)}`;
}

/**
 * Timestamp seconds as a calendar day, in UTC.
 *
 * The server sends a day the bank named, not an instant, so converting it to
 * the reader's zone would move a transaction booked at midnight onto the
 * previous day — and the day is what the ledger groups by.
 */
function isoDate(seconds: bigint): string {
	return new Date(Number(seconds) * 1000).toISOString().slice(0, 10);
}

function shortDate(seconds: bigint | undefined): string {
	if (seconds === undefined) return '';
	// The year is never omitted: the day group heading above a row states it,
	// but the row itself is what a member actually reads while scanning down
	// a list that can span years, and "Sep 15" alone is a different day in
	// nearly every one of them.
	return new Date(Number(seconds) * 1000).toLocaleDateString(undefined, {
		day: 'numeric',
		month: 'short',
		year: '2-digit',
		timeZone: 'UTC'
	});
}

function longDate(seconds: bigint): string {
	return new Date(Number(seconds) * 1000).toLocaleDateString(undefined, {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
		timeZone: 'UTC'
	});
}

function clockTime(seconds: bigint): string {
	return new Date(Number(seconds) * 1000).toLocaleTimeString(undefined, {
		hour: '2-digit',
		minute: '2-digit'
	});
}
