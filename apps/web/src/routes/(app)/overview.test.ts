import { describe, expect, it, vi } from 'vitest';

const listAccounts = vi.fn();
const listTransactions = vi.fn();
const getBalanceTrend = vi.fn();
const getMonthSummary = vi.fn();

vi.mock('$lib/server/banking', () => ({
	banking: {
		listAccounts: (...args: unknown[]) => listAccounts(...args),
		listTransactions: (...args: unknown[]) => listTransactions(...args),
		getBalanceTrend: (...args: unknown[]) => getBalanceTrend(...args),
		getMonthSummary: (...args: unknown[]) => getMonthSummary(...args)
	}
}));

vi.mock('$lib/server/call', () => ({
	call: <T>(_cookies: unknown, invoke: (o: unknown) => Promise<T>) =>
		invoke({ headers: {}, onHeader: () => {} })
}));

const cookies = { get: () => 'session', delete: () => {}, set: () => {} };
const money = (minor: bigint, currency: string) => ({ minor, currency });
const total = (minor: bigint, currency: string) => ({ total: money(minor, currency), accountCount: 1 });

async function overview() {
	const { load } = await import('./+page.server');
	return (await load!({ cookies, depends: () => {}, url: new URL('http://localhost/') } as never)) as {
		currencies: { currency: string; household?: string; own?: string; month?: unknown; chart?: unknown }[];
		recentTransactions: { description: string; amount: string }[];
	};
}

describe('Overview load', () => {
	it('gives a currency with zero figures, no trend and no summary no tile', async () => {
		listAccounts.mockResolvedValue({
			accounts: [],
			householdTotals: [total(669_800n, 'EUR')],
			ownTotals: [total(0n, 'USD')],
			failures: []
		});
		listTransactions.mockResolvedValue({ ledger: { transactions: [] } });
		getBalanceTrend.mockResolvedValue({ trends: [], partialCoverage: false });
		getMonthSummary.mockResolvedValue({ months: [] });

		const view = await overview();

		expect(view.currencies.map((c) => c.currency)).toEqual(['EUR']);
	});

	it('keeps a zero currency that has a month summary, and reads the merchant name', async () => {
		listAccounts.mockResolvedValue({
			accounts: [],
			householdTotals: [total(669_800n, 'EUR')],
			ownTotals: [total(0n, 'GBP')],
			failures: []
		});
		listTransactions.mockResolvedValue({
			ledger: {
				transactions: [
					{
						id: 't1',
						displayName: 'Pingo Doce',
						counterpartyName: 'PINGO DOCE 123456789',
						accountName: 'Current account',
						amount: money(-4_218n, 'EUR'),
						bookingDate: { seconds: 1_789_000_000n }
					}
				]
			}
		});
		getBalanceTrend.mockResolvedValue({ trends: [], partialCoverage: false });
		getMonthSummary.mockResolvedValue({
			months: [
				{
					currency: 'GBP',
					in: money(10n, 'GBP'),
					out: money(0n, 'GBP'),
					net: money(10n, 'GBP'),
					monthStart: { seconds: 1_788_220_800n },
					topMerchants: [],
					largestPayments: []
				}
			]
		});

		const view = await overview();

		expect(view.currencies.map((c) => c.currency).sort()).toEqual(['EUR', 'GBP']);
		expect(view.recentTransactions[0]?.description).toBe('Pingo Doce');
	});
});
