import { redirect, type ServerLoad } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { RestoreReason } from '@wimm/contracts/banking';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';
import { rememberHandoff } from '$lib/server/handoff';

/**
 * J08.A / 01 · Include a bank's transactions.
 *
 * The third reason a connection goes back to its bank. The other two are access
 * having run out and access having been cut off early; this one is a connection
 * that is live and narrow, which is neither.
 *
 * It is the restore flow with a different sentence in front of it: the same
 * hand-off, no picker, no account chooser, and every account keeps the owners
 * and the levels it had. Nothing here disconnects anything first.
 */
export const load: ServerLoad = async ({ cookies, params, url }) => {
	const connectionId = params.connection ?? '';

	let banks: { connectionId: string; bankName: string; accessEndsOn: string }[];
	try {
		banks = await narrowBanks(cookies);
	} catch (caught) {
		if (ConnectError.from(caught).code === Code.Unauthenticated) {
			redirect(303, `/signin?next=${encodeURIComponent(url.pathname)}`);
		}
		redirect(303, '/transactions');
	}

	const bank = banks.find((candidate) => candidate.connectionId === connectionId);
	// A bank that is already wide has nothing to widen, and a member who
	// arrives here from a stale link is sent to the list rather than shown a
	// screen offering something that has already happened.
	if (!bank) redirect(303, '/transactions');

	return { connectionId, bankName: bank.bankName, accessEndsOn: bank.accessEndsOn };
};

/**
 * The banks this member owns an account at whose consent does not cover
 * transactions, with the date each one's access runs to.
 *
 * The ledger names them; the accounts listing carries the date. Neither alone
 * is enough, and asking for both is two reads on a screen a member reaches
 * rarely.
 */
async function narrowBanks(cookies: Parameters<typeof call>[0]) {
	const [ledger, accounts] = await Promise.all([
		call(cookies, (options) => banking.listTransactions({ skipSync: true }, options)),
		call(cookies, (options) => banking.listAccounts({ skipRead: true }, options))
	]);

	const endsOn = new Map<string, string>();
	for (const account of accounts.accounts) {
		const connection = account.connection;
		if (!connection?.consentExpiresAt) continue;
		endsOn.set(
			connection.id,
			new Date(Number(connection.consentExpiresAt.seconds) * 1000).toLocaleDateString(undefined, {
				day: 'numeric',
				month: 'long',
				year: 'numeric'
			})
		);
	}

	return (ledger.ledger?.narrowConnections ?? []).map((connection) => ({
		connectionId: connection.connectionId,
		bankName: connection.bankName,
		accessEndsOn: endsOn.get(connection.connectionId) ?? ''
	}));
}

export const actions = {
	/** Sends the member back to their bank to confirm once more. */
	default: async ({ cookies, params, request }) => {
		const connectionId = params.connection ?? '';

		const { handoffUrl } = await call(cookies, (options) =>
			banking.restoreConnection({ connectionId, reason: RestoreReason.WIDEN_SCOPE }, options)
		);

		// The return produces the same response whether it widened, restored or
		// connected, so which one happened is only knowable from here.
		const submitted = await request.formData();
		const bankName = submitted.get('bankName');
		rememberHandoff(cookies, {
			bankName: typeof bankName === 'string' ? bankName : '',
			restoring: true
		});

		redirect(303, handoffUrl);
	}
};
