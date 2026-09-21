import { beforeEach, describe, expect, it, vi } from 'vitest';

/**
 * What the ledger route does with each query shape, and what it asks wimmd for.
 *
 * The property that matters most: **the load never asks a bank.** A load
 * function blocks navigation until it returns, and a first fill is unbounded
 * pages across every account — so syncing here means clicking Transactions and
 * watching Overview until every bank has answered. The rows already held render
 * first and the page syncs behind that arrival.
 *
 * The second: **only the newest page syncs at all.** A sync inserts at the
 * newest end, so asking the banks while a member is reading August would move
 * the ground under them for nothing.
 */

const listTransactions = vi.fn();
const listAccounts = vi.fn();

vi.mock('$lib/server/banking', () => ({
	banking: {
		listTransactions: (...args: unknown[]) => listTransactions(...args),
		listAccounts: (...args: unknown[]) => listAccounts(...args)
	}
}));

vi.mock('$lib/server/call', () => ({
	call: <T>(_cookies: unknown, invoke: (o: unknown) => Promise<T>) =>
		invoke({ headers: {}, onHeader: () => {} })
}));

const cookies = { get: () => 'session', delete: () => {}, set: () => {} };

function at(iso: string) {
	return { seconds: BigInt(Math.floor(Date.parse(iso) / 1000)), nanos: 0 };
}

function transaction(overrides: Record<string, unknown> = {}) {
	return {
		id: 't1',
		accountId: 'a1',
		accountName: 'Current account',
		bankName: 'Monzo',
		status: 1,
		amount: { minor: -4218n, currency: 'EUR' },
		bookingDate: at('2026-09-17T00:00:00Z'),
		counterpartyName: 'Pingo Doce',
		remittance: '',
		...overrides
	};
}

function ledger(overrides: Record<string, unknown> = {}) {
	return {
		ledger: {
			transactions: [transaction()],
			totalCount: 382,
			hasOlder: false,
			hasNewer: false,
			narrowConnections: [],
			failures: [],
			ownsNoAccount: false,
			syncedAt: at('2026-09-17T09:14:00Z'),
			reachesBackTo: at('2026-06-04T00:00:00Z'),
			newestOnPage: at('2026-09-17T00:00:00Z'),
			oldestOnPage: at('2026-09-15T00:00:00Z'),
			...overrides
		}
	};
}

async function open(search = '') {
	const { load } = await import('./+page.server');
	return (await load!({
		cookies,
		depends: () => {},
		url: new URL(`http://localhost/transactions${search}`)
	} as never)) as {
		days: { date: string; entries: { id: string; unusual?: boolean }[] }[];
		count: number;
		accountId?: string;
		freshness?: string;
		span?: string;
		oldestHref?: string;
		newestHref?: string;
		atOldest: boolean;
		syncOnArrival: boolean;
		noBank: boolean;
		ownsNothing: boolean;
		filterAccount?: string;
		narrow: { connectionId: string; bankName: string; widenHref: string }[];
		problems: { connectionId: string; bankName: string; kind: string }[];
	};
}

beforeEach(() => {
	vi.resetAllMocks();
	// A household with a bank, unless a test says otherwise.
	listAccounts.mockResolvedValue({ accounts: [{ id: 'a1' }] });
});

describe('the query shapes', () => {
	it('opens on the newest page without asking a bank, and says to sync after', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open();

		const [request] = listTransactions.mock.calls[0]!;
		expect(request.cursor).toBeUndefined();
		// The whole point: navigation is not held open while banks answer.
		expect(request.skipSync).toBe(true);
		expect(request.accountId).toBe('');
		expect(data.count).toBe(382);
		expect(data.syncOnArrival).toBe(true);
	});

	it('narrows to one account without becoming a different screen', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open('?account=a1');

		const [request] = listTransactions.mock.calls[0]!;
		expect(request.accountId).toBe('a1');
		// The account is named on screen, and removing the narrowing widens the
		// same list rather than navigating anywhere.
		expect(data.filterAccount).toBe('Current account · Monzo');
		expect(data.accountId).toBe('a1');
	});

	it('reads away from today on ?before, and never syncs', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open('?before=2026-08-04.t9');

		const [request] = listTransactions.mock.calls[0]!;
		expect(request.older).toBe(true);
		expect(request.cursor?.transactionId).toBe('t9');
		expect(request.skipSync).toBe(true);
		// Not even afterwards: new transactions do not belong on this page.
		expect(data.syncOnArrival).toBe(false);
	});

	it('reads back towards today on ?after, and never syncs', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open('?after=2026-08-04.t9');

		const [request] = listTransactions.mock.calls[0]!;
		expect(request.older).toBe(false);
		expect(request.cursor?.transactionId).toBe('t9');
		expect(request.skipSync).toBe(true);
		expect(data.syncOnArrival).toBe(false);
	});

	it('narrowing to one account still syncs on arrival', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open('?account=a1');

		expect(data.syncOnArrival).toBe(true);
	});

	it('carries the account filter through paging', async () => {
		listTransactions.mockResolvedValue(ledger({ hasOlder: true }));

		const data = await open('?account=a1');

		expect(data.oldestHref).toContain('account=a1');
		expect(data.oldestHref).toContain('oldest=1');
	});
});

describe('the sync behind the arrival', () => {
	/**
	 * The page syncs after it has rendered and re-reads this load when the
	 * banks answer. It re-reads *this* load and not the whole tree: the layout's
	 * session check has not changed, and re-running it is a second round trip
	 * per sync.
	 *
	 * The loop this pins down was real. The sync used to run from an `$effect`
	 * that read the page data it then invalidated, so every answer re-triggered
	 * the request that produced it. It fires once per navigation now.
	 */
	it('names the dependency a landed sync re-reads', async () => {
		listTransactions.mockResolvedValue(ledger());
		const depends = vi.fn();

		const { load } = await import('./+page.server');
		await load!({
			cookies,
			depends,
			url: new URL('http://localhost/transactions')
		} as never);

		expect(depends).toHaveBeenCalledWith('wimm:ledger');
	});
});

describe('a cursor wimm did not issue', () => {
	/**
	 * Scoping is wimmd's, not this route's: a cursor is a position in a list
	 * that is already filtered to the accounts the member owns, so one naming
	 * somebody else's transaction seeks past nothing rather than reaching it.
	 */
	it('returns nothing for an account the member does not own', async () => {
		listTransactions.mockResolvedValue(
			ledger({ transactions: [], totalCount: 0, newestOnPage: undefined, oldestOnPage: undefined })
		);

		const data = await open('?before=2026-08-04.someone-elses-transaction');

		expect(data.days).toEqual([]);
		expect(data.count).toBe(0);
	});

	it('treats a malformed cursor as the newest page', async () => {
		listTransactions.mockResolvedValue(ledger());

		await open('?before=not-a-cursor');

		const [request] = listTransactions.mock.calls[0]!;
		expect(request.cursor).toBeUndefined();
	});
});

describe('what the list says about itself', () => {
	it('states when it was updated and how far back it reaches', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open();

		expect(data.freshness).toContain('Updated at');
		// The date's shape follows the reader's locale, as every other date in
		// this app does. What is asserted is that it is the date the ledger
		// actually reached, not a period wimm promised.
		expect(data.freshness).toMatch(/Reaching back to .*June.*2026/);
	});

	it('promises no period before the first read', async () => {
		listTransactions.mockResolvedValue(
			ledger({
				transactions: [],
				syncedAt: undefined,
				reachesBackTo: undefined,
				newestOnPage: undefined,
				oldestOnPage: undefined
			})
		);

		const data = await open();

		expect(data.freshness).toBeUndefined();
		expect(data.atOldest).toBe(false);
	});

	it('groups by the day the bank named, not the reader zone', async () => {
		listTransactions.mockResolvedValue(
			ledger({
				transactions: [
					transaction({ id: 't1', bookingDate: at('2026-09-17T00:00:00Z') }),
					transaction({ id: 't2', bookingDate: at('2026-09-16T00:00:00Z') })
				]
			})
		);

		const data = await open();

		expect(data.days).toHaveLength(2);
		expect(data.days[0]!.date).toBe('2026-09-17');
		expect(data.days[1]!.date).toBe('2026-09-16');
	});
});

describe('banks that are not contributing', () => {
	it('offers a narrow bank the way to widen it', async () => {
		listTransactions.mockResolvedValue(
			ledger({ narrowConnections: [{ connectionId: 'c1', bankName: 'Monzo' }] })
		);

		const data = await open();

		expect(data.narrow).toEqual([
			{ connectionId: 'c1', bankName: 'Monzo', widenHref: '/connect/widen/c1' }
		]);
	});

	it('keeps the rows when a bank did not answer', async () => {
		listTransactions.mockResolvedValue(
			ledger({
				failures: [
					{ connectionId: 'c2', bankName: 'Montepio', failure: 1, retryAfterSeconds: 0n }
				]
			})
		);

		const data = await open();

		expect(data.count).toBe(382);
		expect(data.problems[0]!.bankName).toBe('Montepio');
		expect(data.problems[0]!.kind).toBe('unreachable');
	});
});

describe('having nothing to show', () => {
	/**
	 * The reasons differ and must not be collapsed into one empty list: a
	 * household with no bank is told to connect one, and a member who owns no
	 * account is told to ask for one.
	 */
	it('tells a household with no bank to connect one', async () => {
		listTransactions.mockResolvedValue(
			ledger({ transactions: [], totalCount: 0, ownsNoAccount: true, syncedAt: undefined })
		);
		listAccounts.mockResolvedValue({ accounts: [] });

		const data = await open();

		expect(data.noBank).toBe(true);
		expect(data.ownsNothing).toBe(false);
	});

	it('tells a member who owns no account of a bank that is connected', async () => {
		listTransactions.mockResolvedValue(
			ledger({ transactions: [], totalCount: 0, ownsNoAccount: true, syncedAt: undefined })
		);
		listAccounts.mockResolvedValue({ accounts: [{ id: 'a1' }] });

		const data = await open();

		expect(data.noBank).toBe(false);
		expect(data.ownsNothing).toBe(true);
	});

	it('does not claim there is no bank when nothing has been read yet', async () => {
		listTransactions.mockResolvedValue(
			ledger({ transactions: [], totalCount: 0, syncedAt: undefined, reachesBackTo: undefined })
		);

		const data = await open();

		expect(data.noBank).toBe(false);
		expect(data.ownsNothing).toBe(false);
	});
});

describe('a gateway wimm has not been given', () => {
	it('shows the empty screen rather than a failure', async () => {
		listTransactions.mockRejectedValue(new Error('unimplemented'));

		const data = await open();

		expect(data.days).toEqual([]);
		expect(data.count).toBe(0);
	});
});

describe('arriving from an unusual payment on Overview', () => {
	it('asks for the page that starts at that payment and passes its mark on', async () => {
		listTransactions.mockResolvedValue(
			ledger({
				transactions: [
					transaction({ id: 'garage', bookingDate: at('2026-03-12T00:00:00Z'), unusual: true }),
					transaction({ id: 'coffee', bookingDate: at('2026-03-12T00:00:00Z'), unusual: false })
				]
			})
		);

		const data = await open('?page=2026-03-12.garage');

		const [request] = listTransactions.mock.calls[0]!;
		expect(request.pageStart.transactionId).toBe('garage');
		expect(data.days[0]?.entries.map((e) => [e.id, e.unusual])).toEqual([
			['garage', true],
			['coffee', false]
		]);
	});
});
