import { json, type RequestHandler } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';
import { rememberHandoff } from '$lib/server/handoff';

/**
 * The web client's own REST endpoint for the actions Overview takes: refresh,
 * restore and disconnect.
 *
 * REST rather than Connect, per ADR 0001: the browser never speaks Connect, and
 * this surface is undocumented and unversioned because its only consumer ships
 * in the same deploy.
 *
 * Restore answers with a URL rather than redirecting: the caller is fetch, and
 * a 303 would be followed by the browser into the bank's page inside an XHR.
 */
export const POST: RequestHandler = async ({ request, cookies }) => {
	const body = (await request.json()) as {
		action?: string;
		connectionId?: string;
		/** Only for restore, and only so the return can name the bank. */
		bankName?: string;
		/** Only for refresh-transactions: the account the list is narrowed to. */
		accountId?: string;
	};

	try {
		switch (body.action) {
			// The balance read behind an arrival. The page has already rendered
			// what is stored, with the time each figure was read; this asks the
			// banks and the page re-reads itself when they answer.
			case 'read-balances':
				await call(cookies, (options) => banking.listAccounts({ skipRead: false }, options));
				return json({ ok: true });

			case 'refresh':
				await call(cookies, (options) => banking.refreshBalances({}, options));
				return json({ ok: true });

			// The sync behind an arrival. The page has already rendered what is
			// stored; this asks the banks and the page reloads its data when it
			// answers. Bounded by the per-account interval, so reloading does
			// not multiply the calls.
			case 'sync-transactions':
				await call(cookies, (options) =>
					banking.listTransactions({ accountId: body.accountId ?? '' }, options)
				);
				return json({ ok: true });

			// The ledger's own refresh, which is the member asking in as many
			// words and is not bound by that interval. Separate from balances
			// because they are different reads with different costs: a balance
			// is one number per account, and a sync is a round trip per account
			// with paging behind it.
			case 'refresh-transactions':
				await call(cookies, (options) =>
					banking.refreshTransactions({ accountId: body.accountId ?? '' }, options)
				);
				return json({ ok: true });

			case 'restore': {
				const { handoffUrl } = await call(cookies, (options) =>
					banking.restoreConnection({ connectionId: body.connectionId ?? '' }, options)
				);
				// Restoring and connecting come back through the same return and
				// produce the same response, so which one happened is only
				// knowable from here.
				rememberHandoff(cookies, { bankName: body.bankName ?? '', restoring: true });
				return json({ ok: true, handoffUrl });
			}

			case 'disconnect':
				await call(cookies, (options) =>
					banking.disconnectBank({ connectionId: body.connectionId ?? '' }, options)
				);
				return json({ ok: true });

			default:
				return json({ ok: false, reason: 'unknown-action' }, { status: 400 });
		}
	} catch (caught) {
		const problem = ConnectError.from(caught);

		if (problem.code === Code.Unauthenticated) {
			return json({ ok: false, reason: 'signed-out' }, { status: 401 });
		}
		// Only a member who owns one of a bank's accounts may disconnect it.
		if (problem.code === Code.PermissionDenied) {
			return json({ ok: false, reason: 'not-yours' }, { status: 403 });
		}
		// Everything else is the bank or the service behind it. The readings
		// already on screen stay where they are; the page says which bank.
		return json({ ok: false, reason: 'unavailable' }, { status: 503 });
	}
};
