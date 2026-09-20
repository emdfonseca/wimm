import { fail, isRedirect, redirect, type ServerLoad, type Actions } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { Level } from '@wimm/contracts/banking';
import { banking } from '$lib/server/banking';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';
import { formatMoney } from '$lib/money';

/**
 * J03.A / 04 · Choose accounts, reachable again later.
 *
 * Route-backed and resumable on purpose: by the time a member gets here the
 * bank has already granted access, so closing the tab mid-choice must not cost
 * the connection. They come back to this same address and carry on.
 *
 * A member who owns none of these accounts may look and change nothing — the
 * handler refuses either way, and the screen says so rather than presenting
 * controls that fail.
 */
export const load: ServerLoad = async ({ cookies, params, url }) => {
	// A path parameter, not a query: the chooser is for one connection, and a
	// missing one is then a 404 rather than an empty id handed to the service
	// — which is exactly how every successful hand-off failed here once.
	const connectionId = params.connection ?? '';
	const bankId = params.bank ?? '';

	let response;
	try {
		response = await call(cookies, (options) =>
			banking.listConnectionAccounts({ connectionId }, options)
		);
	} catch (caught) {
		const problem = ConnectError.from(caught);
		if (problem.code === Code.Unauthenticated) {
			redirect(303, `/signin?next=${encodeURIComponent(url.pathname + url.search)}`);
		}
		// Not theirs to change. The screen is still shown, read-only, because a
		// member may legitimately hold a level on one of these.
		if (problem.code === Code.PermissionDenied) {
			return {
				bankId,
				bankName: bankId,
				accounts: [],
				members: [],
				owners: {},
				levels: {},
				currentMemberId: '',
				readOnly: true
			};
		}
		redirect(303, '/accounts');
	}

	// Who is asking. The chooser needs it to claim an account for the member
	// looking at the screen, and the browser cannot be trusted to supply it.
	const { member } = await call(cookies, (options) => identity.getCurrentMember({}, options));

	const levels: Record<string, Record<string, string>> = {};
	for (const grant of response.grants) {
		levels[grant.accountId] ??= {};
		levels[grant.accountId]![grant.memberId] = levelName(grant.level);
	}

	const owners: Record<string, Record<string, boolean>> = {};
	for (const account of response.accounts) {
		owners[account.id] = {};
		for (const owner of account.owners) {
			if (owner.id !== member?.id) owners[account.id]![owner.id] = true;
		}
	}

	return {
		bankId,
		bankName: response.accounts[0]?.connection?.bankName ?? bankId,
		accounts: response.accounts.map((account) => ({
			id: account.id,
			name: account.householdName || account.name,
			householdName: account.householdName || undefined,
			bankAccountName: account.name,
			bankName: account.connection?.bankName ?? '',
			meta: account.numberSuffix
				? `${account.connection?.bankName ?? ''} · •••• ${account.numberSuffix}`
				: (account.connection?.bankName ?? ''),
			balance: formatMoney(account.balance?.money),
			owned: account.owned,
			leftOut: Boolean(account.leftOutAt),
			newlyOffered: false
		})),
		// Everyone except the member looking at the screen: a level control for
		// yourself is a control for a decision ownership already made.
		members: response.members
			.filter((other) => other.id !== member?.id)
			.map((other) => ({ id: other.id, name: other.displayName })),
		owners,
		levels,
		currentMemberId: member?.id ?? '',
		readOnly: false
	};
};

export const actions: Actions = {
	/** Ownership. An empty list means the account is nobody's and is never read. */
	owners: async ({ cookies, request }) => {
		const form = await request.formData();
		const accountId = String(form.get('accountId') ?? '');
		const memberIds = form.getAll('memberIds').map(String);

		return refusalOrNull(() =>
			call(cookies, (options) => banking.setAccountOwners({ accountId, memberIds }, options))
		);
	},

	/** One member's level on one account. Hidden removes the grant. */
	level: async ({ cookies, request }) => {
		const form = await request.formData();
		const accountId = String(form.get('accountId') ?? '');
		const memberId = String(form.get('memberId') ?? '');

		return refusalOrNull(() =>
			call(cookies, (options) =>
				banking.setAccountLevel(
					{ accountId, memberId, level: levelValue(String(form.get('level') ?? '')) },
					options
				)
			)
		);
	},

	/** Leave an account out, or bring it back. Both are the confirmed action —
	 *  the dialog is the screen's, not a second round trip through the
	 *  server. */
	leaveOut: async ({ cookies, request }) => {
		const form = await request.formData();
		const accountId = String(form.get('accountId') ?? '');

		return refusalOrNull(() =>
			call(cookies, (options) => banking.setAccountLeftOut({ accountId, leftOut: true }, options))
		);
	},

	bringBack: async ({ cookies, request }) => {
		const form = await request.formData();
		const accountId = String(form.get('accountId') ?? '');

		return refusalOrNull(() =>
			call(cookies, (options) => banking.setAccountLeftOut({ accountId, leftOut: false }, options))
		);
	},

	/** The household's own name for an account. Empty clears it back to the
	 *  bank's own name. */
	rename: async ({ cookies, request }) => {
		const form = await request.formData();
		const accountId = String(form.get('accountId') ?? '');
		const householdName = String(form.get('householdName') ?? '');

		return refusalOrNull(() =>
			call(cookies, (options) => banking.setAccountName({ accountId, householdName }, options))
		);
	}
};

/**
 * A refusal is a form failure rather than an error page: the member is looking
 * at a screen full of other controls that still work.
 */
async function refusalOrNull(run: () => Promise<unknown>) {
	try {
		await run();
		return { ok: true };
	} catch (caught) {
		if (isRedirect(caught)) throw caught;

		const problem = ConnectError.from(caught);
		if (problem.code === Code.PermissionDenied) {
			return fail(403, { reason: 'not-yours' as const });
		}
		// The database refused: this account would be left with no owner.
		// Leaving it out is the way to the same end, and the screen names it
		// inline, beside the row (ADR 0022).
		if (problem.code === Code.FailedPrecondition) {
			return fail(409, { reason: 'last-owner' as const });
		}
		if (problem.code === Code.Unauthenticated) redirect(303, '/signin');
		return fail(500, { reason: 'failed' as const });
	}
}

function levelName(level: Level): string {
	// UNSPECIFIED and HIDDEN both mean no grant, so both render as hidden. The
	// failure mode has to be accidentally hidden (ADR 0019).
	if (level === Level.BALANCE) return 'balance';
	if (level === Level.DETAILS) return 'details';
	return 'hidden';
}

function levelValue(name: string): Level {
	if (name === 'balance') return Level.BALANCE;
	if (name === 'details') return Level.DETAILS;
	return Level.HIDDEN;
}
