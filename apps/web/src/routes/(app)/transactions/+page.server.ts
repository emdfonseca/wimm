import type { Cookies, ServerLoad } from '@sveltejs/kit';
import type { LedgerDay, LedgerProblem, NarrowBank } from '@wimm/ui';
import {
	Failure,
	TransactionStatus,
	type Ledger,
	type NarrowConnection,
	type BankFailure,
	type Transaction
} from '@wimm/contracts/banking';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';
import { formatMoney, isNegative } from '$lib/money';

/**
 * J07 · Transactions. The ledger, and the same page narrowed to one account.
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

	const accountId = url.searchParams.get('account') ?? '';
	const before = url.searchParams.get('before') ?? '';
	const after = url.searchParams.get('after') ?? '';

	const ledger = await read(cookies, { accountId, before, after });

	return {
		...ledger,
		accountId: accountId || undefined,
		// Only the newest page brings itself up to date. A cursor means the
		// member is reading somewhere that new transactions do not belong.
		syncOnArrival: !before && !after
	};
};

interface Query {
	accountId: string;
	before: string;
	after: string;
}

/**
 * Banking is optional: an instance with no gateway configured answers
 * Unimplemented, and the screen shows its empty state rather than an error.
 */
async function read(cookies: Cookies, query: Query) {
	try {
		const cursor = parseCursor(query.before || query.after);
		const [response, accounts] = await Promise.all([
			call(cookies, (options) =>
				banking.listTransactions(
					{
						accountId: query.accountId,
						cursor,
						older: query.before !== '',
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
		return present(response.ledger, query, accounts.accounts.length === 0);
	} catch {
		return present(undefined, query, true);
	}
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

function cursorOf(transaction: Transaction | undefined): string {
	const seconds = transaction?.bookingDate?.seconds;
	if (!transaction || seconds === undefined) return '';
	return `${isoDate(seconds)}.${transaction.id}`;
}

function present(ledger: Ledger | undefined, query: Query, noBank: boolean) {
	const transactions = ledger?.transactions ?? [];
	const filter = query.accountId ? transactions[0]?.accountName : undefined;

	const href = (key: 'before' | 'after', cursor: string) => {
		const params = new URLSearchParams();
		if (query.accountId) params.set('account', query.accountId);
		params.set(key, cursor);
		return `/transactions?${params}`;
	};

	const oldest = cursorOf(transactions[transactions.length - 1]);
	const newest = cursorOf(transactions[0]);

	return {
		days: group(transactions),
		count: ledger?.totalCount ?? 0,
		freshness: freshnessOf(ledger),
		span: spanOf(ledger),
		olderHref: ledger?.hasOlder && oldest ? href('before', oldest) : undefined,
		newerHref: ledger?.hasNewer && newest ? href('after', newest) : undefined,
		// Only said on a page that has reached the end, and only once something
		// has been read: before the first read wimm knows no date to reach to.
		atOldest: Boolean(ledger && !ledger.hasOlder && transactions.length > 0 && ledger.reachesBackTo),
		filterAccount: query.accountId
			? `${filter ?? 'this account'}${bankOf(transactions[0])}`
			: undefined,
		narrow: narrowFrom(ledger),
		problems: problemsFrom(ledger),
		// A member who can see nothing at all is in a household with no bank; one
		// who can see accounts but owns none is told that instead.
		noBank,
		ownsNothing: !noBank && (ledger?.ownsNoAccount ?? false)
	};
}

function bankOf(transaction: Transaction | undefined): string {
	return transaction?.bankName ? ` · ${transaction.bankName}` : '';
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
			description: transaction.counterpartyName || transaction.remittance || 'Card payment',
			account: `${transaction.accountName}${bankOf(transaction)}`,
			amount: formatMoney(transaction.amount) ?? '',
			date: shortDate(seconds),
			negative: isNegative(transaction.amount),
			unsettled: transaction.status === TransactionStatus.PENDING
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
	const newest = ledger?.newestOnPage?.seconds;
	const oldest = ledger?.oldestOnPage?.seconds;
	if (newest === undefined || oldest === undefined) return undefined;

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
	return new Date(Number(seconds) * 1000).toLocaleDateString(undefined, {
		day: 'numeric',
		month: 'short',
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
