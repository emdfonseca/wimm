import { beforeEach, describe, expect, it, vi } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { RestoreReason } from '@wimm/contracts/banking';

/**
 * Widening a connection made before wimm could read transactions.
 *
 * It is the restore flow with a third reason, not a second flow: the same
 * hand-off, no picker, no account chooser, and nothing is disconnected first.
 * That every account keeps its owners and its levels is asserted where it is
 * decided — `TestWideningKeepsOwnersAndLevelsAndChangesTheScope` in
 * `apps/wimm/internal/banking` — because this route never touches either.
 */

const listTransactions = vi.fn();
const listAccounts = vi.fn();
const restoreConnection = vi.fn();
const rememberHandoff = vi.fn();

vi.mock('$lib/server/banking', () => ({
	banking: {
		listTransactions: (...args: unknown[]) => listTransactions(...args),
		listAccounts: (...args: unknown[]) => listAccounts(...args),
		restoreConnection: (...args: unknown[]) => restoreConnection(...args)
	}
}));

vi.mock('$lib/server/call', () => ({
	call: <T>(_cookies: unknown, invoke: (o: unknown) => Promise<T>) =>
		invoke({ headers: {}, onHeader: () => {} })
}));

vi.mock('$lib/server/handoff', () => ({
	rememberHandoff: (...args: unknown[]) => rememberHandoff(...args)
}));

const cookies = { get: () => 'session', delete: () => {}, set: () => {} };

function narrow(connections: { connectionId: string; bankName: string }[]) {
	listTransactions.mockResolvedValue({ ledger: { narrowConnections: connections } });
	listAccounts.mockResolvedValue({
		accounts: connections.map((connection) => ({
			connection: {
				id: connection.connectionId,
				bankName: connection.bankName,
				consentExpiresAt: { seconds: BigInt(Math.floor(Date.parse('2026-12-17') / 1000)) }
			}
		}))
	});
}

async function open(connection: string) {
	const { load } = await import('./[connection]/+page.server');
	return (await load!({
		cookies,
		params: { connection },
		url: new URL(`http://localhost/connect/widen/${connection}`)
	} as never)) as { connectionId: string; bankName: string; accessEndsOn: string };
}

async function submit(connection: string) {
	const { actions } = await import('./[connection]/+page.server');
	const form = new FormData();
	form.set('bankName', 'Monzo');
	return actions.default({
		cookies,
		params: { connection },
		request: { formData: async () => form }
	} as never);
}

beforeEach(() => {
	vi.resetAllMocks();
});

describe('the explainer', () => {
	it('names the bank and the date its access runs to', async () => {
		narrow([{ connectionId: 'c1', bankName: 'Monzo' }]);

		const data = await open('c1');

		expect(data.bankName).toBe('Monzo');
		expect(data.accessEndsOn).toMatch(/2026/);
	});

	it('sends a bank that is already wide back to the list', async () => {
		narrow([]);

		await expect(open('c1')).rejects.toSatisfy(isRedirect);
	});
});

describe('the hand-off', () => {
	it('re-enters the restore flow with the widening reason', async () => {
		restoreConnection.mockResolvedValue({ handoffUrl: 'https://bank.example/consent' });

		await expect(submit('c1')).rejects.toSatisfy(isRedirect);

		const [request] = restoreConnection.mock.calls[0]!;
		expect(request.connectionId).toBe('c1');
		expect(request.reason).toBe(RestoreReason.WIDEN_SCOPE);
	});

	it('disconnects nothing and picks no bank', async () => {
		restoreConnection.mockResolvedValue({ handoffUrl: 'https://bank.example/consent' });

		await expect(submit('c1')).rejects.toSatisfy(isRedirect);

		// The only call is the one that hands the member back to their bank.
		expect(restoreConnection).toHaveBeenCalledTimes(1);
		expect(rememberHandoff).toHaveBeenCalledWith(cookies, {
			bankName: 'Monzo',
			restoring: true
		});
	});
});
