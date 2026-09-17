import { describe, expect, it, vi, beforeEach } from 'vitest';
import { Code, ConnectError } from '@connectrpc/connect';

/**
 * What Overview does with what wimmd gives it.
 *
 * The property that matters most: **a reading is never lost**. A bank that
 * refuses, a bank that does not answer, and a bank whose access has run out all
 * leave the figures already on screen exactly where they were, with their own
 * read times — and each routes to a different thing for the member to do.
 */

const listAccounts = vi.fn();

vi.mock('$lib/server/banking', () => ({
	banking: { listAccounts: (...args: unknown[]) => listAccounts(...args) }
}));

vi.mock('$lib/server/call', () => ({
	call: <T>(_cookies: unknown, invoke: (o: unknown) => Promise<T>) =>
		invoke({ headers: {}, onHeader: () => {} })
}));

const cookies = { get: () => 'session', delete: () => {}, set: () => {} };

async function overview() {
	const { load } = await import('./+page.server');
	return (await load!({ cookies, url: new URL('http://localhost/') } as never)) as {
		accounts: {
			name: string;
			balance?: string;
			readAt?: string;
			numberSuffix?: string;
			stale: boolean;
			notUpdating: boolean;
			negative: boolean;
		}[];
		totals: { total: string; currency: string }[];
		problems: { bankName: string; kind: string; retryAfter?: string }[];
	};
}

function account(overrides: Record<string, unknown> = {}) {
	return {
		id: 'a1',
		name: 'Conta à Ordem',
		numberSuffix: '0538',
		owned: true,
		connection: { id: 'c1', bankName: 'Montepio', live: true },
		balance: {
			money: { minor: 420_010n, currency: 'EUR' },
			readAt: { seconds: BigInt(Math.floor(Date.now() / 1000) - 120) },
			stale: false
		},
		...overrides
	};
}

beforeEach(() => {
	vi.resetAllMocks();
});

describe('Overview', () => {
	it('formats balances from minor units and says when they were read', async () => {
		listAccounts.mockResolvedValue({
			accounts: [account()],
			totals: [{ total: { minor: 420_010n, currency: 'EUR' } }],
			failures: []
		});

		const data = await overview();

		expect(data.accounts[0]!.balance).toMatch(/4[.,]200[.,]10/);
		expect(data.accounts[0]!.readAt).toBe('2 minutes ago');
		expect(data.totals[0]!.total).toMatch(/4[.,]200[.,]10/);
	});

	// Reading happens on arrival: a member in front of the screen is exactly
	// the case gateways do not throttle.
	it('reads on arrival rather than showing what was stored', async () => {
		listAccounts.mockResolvedValue({ accounts: [], totals: [], failures: [] });

		await overview();

		expect(listAccounts).toHaveBeenCalledWith({ skipRead: false }, expect.anything());
	});

	// The rate-limited path: the member is told when, and the balance stays.
	it('keeps the readings and reports when a bank can be tried again', async () => {
		listAccounts.mockResolvedValue({
			accounts: [account({ balance: { ...account().balance, stale: true } })],
			totals: [{ total: { minor: 420_010n, currency: 'EUR' } }],
			failures: [
				{
					connectionId: 'c1',
					bankName: 'Montepio',
					failure: 6, // RATE_LIMITED
					retryAfterSeconds: 21_600n
				}
			]
		});

		const data = await overview();

		expect(data.problems[0]!.kind).toBe('rate-limited');
		// A clock time, as the frame draws it — "possible after 14:20" rather
		// than a duration the member has to add to whenever they loaded the
		// page. Matched by shape, because the format follows the locale.
		expect(data.problems[0]!.retryAfter).toMatch(/\d{1,2}[:.]\d{2}/);
		// The figure is still there.
		expect(data.accounts[0]!.balance).toMatch(/4[.,]200[.,]10/);
		expect(data.accounts[0]!.stale).toBe(true);
	});

	// A refusal with no retry-after gets no time at all rather than a zero
	// rendered as "now".
	it('gives no retry time when the bank did not say', async () => {
		listAccounts.mockResolvedValue({
			accounts: [account()],
			totals: [],
			failures: [
				{ connectionId: 'c1', bankName: 'Montepio', failure: 6, retryAfterSeconds: 0n }
			]
		});

		const data = await overview();

		expect(data.problems[0]!.retryAfter).toBeUndefined();
	});

	// The partial path: one bank fails and the other's figures are untouched.
	it('one bank failing leaves the other bank alone', async () => {
		listAccounts.mockResolvedValue({
			accounts: [
				account(),
				account({
					id: 'a2',
					name: 'Conta ActivoBank',
					connection: { id: 'c2', bankName: 'ActivoBank', live: true },
					balance: { ...account().balance, stale: true }
				})
			],
			totals: [{ total: { minor: 840_020n, currency: 'EUR' } }],
			failures: [
				{ connectionId: 'c2', bankName: 'ActivoBank', failure: 1, retryAfterSeconds: 0n }
			]
		});

		const data = await overview();

		expect(data.problems).toHaveLength(1);
		expect(data.problems[0]!.bankName).toBe('ActivoBank');
		expect(data.problems[0]!.kind).toBe('unreachable');
		// Both accounts still carry a figure.
		expect(data.accounts.every((a) => a.balance !== undefined)).toBe(true);
	});

	// Access running out is not a failure to retry: it routes to restoring.
	it('access having run out is its own kind', async () => {
		listAccounts.mockResolvedValue({
			accounts: [account({ connection: { id: 'c1', bankName: 'ActivoBank', live: false } })],
			totals: [],
			failures: [
				{ connectionId: 'c1', bankName: 'ActivoBank', failure: 4, retryAfterSeconds: 0n }
			]
		});

		const data = await overview();

		expect(data.problems[0]!.kind).toBe('access-ended');
		expect(data.accounts[0]!.notUpdating).toBe(true);
	});

	// An account seen at balance level carries no identifier, because the
	// server sent none. An empty string would render as "•••• ".
	it('an absent number suffix stays absent', async () => {
		listAccounts.mockResolvedValue({
			accounts: [account({ numberSuffix: '' })],
			totals: [],
			failures: []
		});

		const data = await overview();

		expect(data.accounts[0]!.numberSuffix).toBeUndefined();
	});

	it('carries the sign of an overdrawn account', async () => {
		listAccounts.mockResolvedValue({
			accounts: [
				account({ balance: { ...account().balance, money: { minor: -31_240n, currency: 'EUR' } } })
			],
			totals: [],
			failures: []
		});

		const data = await overview();

		expect(data.accounts[0]!.negative).toBe(true);
		expect(data.accounts[0]!.balance).toMatch(/-|−/);
	});

	// Banking is optional. An instance with no gateway answers Unimplemented,
	// and the member sees the empty state — which is what lets this deploy
	// before anyone holds gateway credentials.
	it('shows nothing rather than failing when no gateway is configured', async () => {
		listAccounts.mockRejectedValue(new ConnectError('no gateway', Code.Unimplemented));

		const data = await overview();

		expect(data.accounts).toEqual([]);
		expect(data.totals).toEqual([]);
		expect(data.problems).toEqual([]);
	});

	// A bank that never answers must not take the page down with it. Who the
	// member is comes from the (app) layout, so this load failing would leave
	// the shell rendered around nothing — it returns an empty view instead.
	it('a failing banking call does not throw', async () => {
		listAccounts.mockRejectedValue(new ConnectError('timeout', Code.DeadlineExceeded));

		await expect(overview()).resolves.toMatchObject({
			accounts: [],
			totals: [],
			problems: [],
			banks: []
		});
	});
});

// The outcome of a hand-off reaches the screen. Every one of these redirected
// to Overview and was ignored, so PayPal granting access with no accounts
// looked identical to nothing happening.
describe('the outcome of a hand-off', () => {
	/**
	 * An outcome is only read where there was a hand-off, so the cookie the
	 * outgoing hand-off wrote is part of the setup. Without it the parameter
	 * is ignored, which is what stops a reload re-announcing it.
	 */
	async function outcomeFor(query: string, handoff: string | null = '{"bankName":"Monzo"}') {
		listAccounts.mockResolvedValue({ accounts: [], totals: [], failures: [] });
		const { load } = await import('./+page.server');
		const data = (await load!({
			cookies: { ...cookies, get: () => handoff ?? undefined },
			url: new URL(`http://localhost/${query}`)
		} as never)) as { outcome?: string; outcomeBank?: string };
		return data.outcome;
	}

	it('carries each outcome this app produces', async () => {
		expect(await outcomeFor('?outcome=no-accounts')).toBe('no-accounts');
		expect(await outcomeFor('?outcome=declined')).toBe('declined');
		expect(await outcomeFor('?outcome=bank-unavailable')).toBe('bank-unavailable');
		expect(await outcomeFor('?outcome=already-connected')).toBe('already-connected');
	});

	it('is absent when there was no hand-off', async () => {
		expect(await outcomeFor('')).toBeUndefined();
	});

	// The outcome lives in the URL and the URL survives a reload. Without the
	// hand-off record being spent, a member who reloaded Overview was told
	// again that a bank exposed no accounts, long after connecting others.
	it('is absent when the hand-off has already been spent', async () => {
		expect(await outcomeFor('?outcome=no-accounts', null)).toBeUndefined();
	});

	// A restore and a first connection come back through the same return and
	// produce the same response, so only the outgoing record tells them apart.
	it('is restored when the hand-off was a restore', async () => {
		expect(await outcomeFor('?outcome=connected', '{"bankName":"Monzo","restoring":true}')).toBe(
			'restored'
		);
		expect(await outcomeFor('?outcome=connected')).toBe('connected');
	});

	// A query parameter is not a message: anything else is ignored rather than
	// rendered.
	it('ignores anything it did not produce', async () => {
		expect(await outcomeFor('?outcome=your-account-was-deleted')).toBeUndefined();
		expect(await outcomeFor('?outcome=%3Cscript%3E')).toBeUndefined();
	});
});
