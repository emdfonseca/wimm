import { beforeEach, describe, expect, it, vi } from 'vitest';
import { LedgerDirection } from '@wimm/contracts/banking';
import { formatMoney } from '$lib/money';
import { filterHref } from './filters';

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

const CURRENT = '0f8c3d2a-1b2c-4d3e-8f90-a1b2c3d4e5f6';
const SAVINGS = '1a2b3c4d-5e6f-4a1b-9c2d-3e4f5a6b7c8d';

function at(iso: string) {
	return { seconds: BigInt(Math.floor(Date.parse(iso) / 1000)), nanos: 0 };
}

function transaction(overrides: Record<string, unknown> = {}) {
	return {
		id: 't1',
		accountId: CURRENT,
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
			filterAccounts: [
				{ accountId: CURRENT, name: 'Current account', bankName: 'Monzo' },
				{ accountId: SAVINGS, name: 'Savings', bankName: 'Monzo' }
			],
			months: ['2026-09', '2026-08', '2026-06'],
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
		days: { date: string; entries: { id: string; unusual?: boolean; transfer?: boolean }[] }[];
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
		filters: { account: string; q: string; month: string; direction: string };
		accounts: { value: string; label: string }[];
		months: { value: string; label: string }[];
		clearHref: string;
		pages: { key: string; href: string }[];
		narrow: { connectionId: string; bankName: string; widenHref: string }[];
		problems: { connectionId: string; bankName: string; kind: string }[];
		totals: { moneyIn: string; moneyOut: string }[];
		transfersLeftOut: number;
		notSettled: number;
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

		const data = await open(`?account=${CURRENT}`);

		const [request] = listTransactions.mock.calls[0]!;
		expect(request.accountId).toBe(CURRENT);
		// The account select names it, and clearing widens the same list.
		expect(data.filters.account).toBe(CURRENT);
		expect(data.accounts).toEqual([
			{ value: CURRENT, label: 'Current account · Monzo' },
			{ value: SAVINGS, label: 'Savings · Monzo' }
		]);
		expect(data.clearHref).toBe('/transactions');
		expect(data.accountId).toBe(CURRENT);
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

		const data = await open(`?account=${CURRENT}`);

		expect(data.syncOnArrival).toBe(true);
	});

	it('carries the account filter through paging', async () => {
		listTransactions.mockResolvedValue(ledger({ hasOlder: true }));

		const data = await open(`?account=${CURRENT}`);

		expect(data.oldestHref).toContain(`account=${CURRENT}`);
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
				failures: [{ connectionId: 'c2', bankName: 'Montepio', failure: 1, retryAfterSeconds: 0n }]
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

describe("a row that is half of a transfer between the member's own accounts", () => {
	it('carries the label, and never the unusual mark beside it', async () => {
		listTransactions.mockResolvedValue(
			ledger({
				transactions: [
					transaction({ id: 'out', bookingDate: at('2026-09-01T00:00:00Z'), ownTransfer: true }),
					transaction({ id: 'in', bookingDate: at('2026-09-01T00:00:00Z'), ownTransfer: true }),
					transaction({ id: 'coffee', bookingDate: at('2026-09-01T00:00:00Z'), unusual: true })
				]
			})
		);

		const data = await open('');

		expect(
			data.days[0]?.entries.map((e) => [e.id, e.transfer ?? false, e.unusual ?? false])
		).toEqual([
			['out', true, false],
			['in', true, false],
			['coffee', false, true]
		]);
	});
});

async function redirected(search: string): Promise<{ status: number; location: string }> {
	try {
		await open(search);
	} catch (thrown) {
		return thrown as { status: number; location: string };
	}
	throw new Error(`${search} was not redirected`);
}

describe('the filters', () => {
	it('sends every filter in the address to wimmd', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open(`?account=${CURRENT}&q=galp&month=2026-08&direction=out`);

		const [request] = listTransactions.mock.calls[0]!;
		expect(request).toMatchObject({
			accountId: CURRENT,
			search: 'galp',
			month: '2026-08',
			direction: LedgerDirection.OUT
		});
		expect(data.filters).toEqual({
			account: CURRENT,
			q: 'galp',
			month: '2026-08',
			direction: 'out'
		});
	});

	it('sends money in, and no direction when none is chosen', async () => {
		listTransactions.mockResolvedValue(ledger());
		await open('?direction=in');
		await open('');

		expect(listTransactions.mock.calls[0]![0].direction).toBe(LedgerDirection.IN);
		expect(listTransactions.mock.calls[1]![0].direction).toBe(LedgerDirection.UNSPECIFIED);
	});

	it.each([
		['an account that is not a uuid', '?account=a1&q=galp', '/transactions?q=galp'],
		['a blank search', '?q=%20%20&month=2026-08', '/transactions?month=2026-08'],
		[
			'a search over 100 characters',
			`?q=${'a'.repeat(101)}&direction=in`,
			'/transactions?direction=in'
		],
		['a month that does not exist', '?month=2026-13&q=galp', '/transactions?q=galp'],
		['a month written another way', '?month=August', '/transactions'],
		[
			'a direction that is not in or out',
			'?direction=sideways&month=2026-08',
			'/transactions?month=2026-08'
		],
		[
			'an empty value from a plain form',
			'?account=&q=galp&month=&direction=',
			'/transactions?q=galp'
		]
	])('drops %s, keeping the rest, without asking wimmd', async (_name, search, location) => {
		listTransactions.mockResolvedValue(ledger());

		const thrown = await redirected(search);

		expect(thrown).toMatchObject({ status: 303, location });
		expect(listTransactions).not.toHaveBeenCalled();
	});

	it('drops an account the member cannot narrow to, the same way whatever the reason', async () => {
		// A stranger's, a left-out and a nonexistent account are all absent from
		// the accounts wimmd offers, and the address must not tell them apart.
		const strangers = [
			'2b3c4d5e-6f7a-4b8c-9d0e-1f2a3b4c5d6e',
			'3c4d5e6f-7a8b-4c9d-8e1f-2a3b4c5d6e7f',
			'4d5e6f7a-8b9c-4d0e-9f2a-3b4c5d6e7f80'
		];
		listTransactions.mockResolvedValue(ledger({ transactions: [], totalCount: 0 }));

		const thrown = await Promise.all(strangers.map((id) => redirected(`?account=${id}&q=galp`)));

		for (const t of thrown) expect(t).toEqual(thrown[0]);
		expect(thrown[0]).toMatchObject({ status: 303, location: '/transactions?q=galp' });
	});

	it('carries every filter on the newest, the oldest and every page link', async () => {
		listTransactions.mockResolvedValue(
			ledger({
				hasOlder: true,
				pages: [
					{
						cursor: { bookingDate: at('2026-08-31T00:00:00Z'), transactionId: 't1' },
						newest: at('2026-08-31T00:00:00Z'),
						oldest: at('2026-08-18T00:00:00Z')
					},
					{
						cursor: { bookingDate: at('2026-08-17T00:00:00Z'), transactionId: 't9' },
						newest: at('2026-08-17T00:00:00Z'),
						oldest: at('2026-08-01T00:00:00Z')
					}
				]
			})
		);
		const filters = `account=${CURRENT}&q=galp&month=2026-08&direction=out`;

		const data = await open(`?${filters}&before=2026-08-20.t5`);

		for (const href of [data.newestHref, data.oldestHref, ...data.pages.map((p) => p.href)]) {
			const params = new URL(href!, 'http://localhost').searchParams;
			expect(params.get('account')).toBe(CURRENT);
			expect(params.get('q')).toBe('galp');
			expect(params.get('month')).toBe('2026-08');
			expect(params.get('direction')).toBe('out');
		}
	});

	it('names the months offered in UTC, and keeps a chosen month that holds nothing', async () => {
		listTransactions.mockResolvedValue(ledger({ months: ['2026-09', '2026-06'] }));

		const data = await open('?month=2026-08');

		expect(data.months).toEqual([
			{ value: '2026-09', label: 'September 2026' },
			{ value: '2026-08', label: 'August 2026' },
			{ value: '2026-06', label: 'June 2026' }
		]);
	});

	it('syncs a filtered newest page on arrival, as an unfiltered one', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open('?month=2026-08&q=galp');

		expect(data.syncOnArrival).toBe(true);
	});
});

describe('what a filtered list adds up to', () => {
	it('writes each currency signed, and passes on what the figures leave out', async () => {
		listTransactions.mockResolvedValue(
			ledger({
				totals: [
					{
						currency: 'EUR',
						moneyIn: { minor: 221499n, currency: 'EUR' },
						moneyOut: { minor: -120455n, currency: 'EUR' }
					},
					{
						currency: 'USD',
						moneyIn: { minor: 0n, currency: 'USD' },
						moneyOut: { minor: -4200n, currency: 'USD' }
					}
				],
				transfersLeftOut: 1,
				notSettled: 2
			})
		);

		const data = await open('?month=2026-08');

		expect(data.totals).toEqual([
			{
				moneyIn: formatMoney({ minor: 221499n, currency: 'EUR' }, { signed: true }),
				moneyOut: formatMoney({ minor: -120455n, currency: 'EUR' }, { signed: true })
			},
			{
				moneyIn: formatMoney({ minor: 0n, currency: 'USD' }, { signed: true }),
				moneyOut: formatMoney({ minor: -4200n, currency: 'USD' }, { signed: true })
			}
		]);
		expect(data.totals[0]!.moneyIn).toMatch(/^\+/);
		expect(data.totals[0]!.moneyOut).toMatch(/^[-−]/);
		expect(data.transfersLeftOut).toBe(1);
		expect(data.notSettled).toBe(2);
	});

	it('has no figures when wimmd sends none', async () => {
		listTransactions.mockResolvedValue(ledger());

		const data = await open();

		expect(data.totals).toEqual([]);
		expect(data.transfersLeftOut).toBe(0);
		expect(data.notSettled).toBe(0);
	});
});

describe('the address a filter change produces', () => {
	it('starts at the newest match: every cursor is dropped and the filters kept', () => {
		const url = filterHref({ account: CURRENT, q: 'galp', month: '', direction: 'in' });

		expect(url).toBe(`/transactions?account=${CURRENT}&q=galp&direction=in`);
	});

	it('is the whole ledger when nothing is in force', () => {
		expect(filterHref({ account: '', q: '', month: '', direction: '' })).toBe('/transactions');
	});
});
