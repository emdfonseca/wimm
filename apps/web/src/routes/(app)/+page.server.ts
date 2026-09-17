import type { Cookies, ServerLoad } from '@sveltejs/kit';
import type { Connection } from '@wimm/contracts/banking';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';
import { takeHandoff } from '$lib/server/handoff';
import { formatMoney, isNegative } from '$lib/money';
import { Failure } from '@wimm/contracts/banking';

/**
 * The landing page, behind a session.
 *
 * Someone without one is sent to sign in rather than shown a dead end, and the
 * path they were trying to reach travels as a query parameter only so far as
 * the sign-in route; wimmd records it against the attempt and ignores anything
 * a client supplies later.
 */
/**
 * The banks behind the accounts a member can see.
 *
 * `manageable` is whether they own at least one account on it — the same rule
 * the chooser enforces — so a control that would be refused is not offered.
 */
function banksFrom(accounts: { connection?: Connection; owned: boolean }[]) {
	const byConnection = new Map<string, ReturnType<typeof bankOf>>();

	for (const account of accounts) {
		const connection = account.connection;
		if (!connection) continue;

		const existing = byConnection.get(connection.id);
		if (existing) {
			existing.accountCount += 1;
			existing.manageable ||= account.owned;
			continue;
		}
		byConnection.set(connection.id, bankOf(connection, account.owned));
	}

	return [...byConnection.values()];
}

function bankOf(connection: Connection, owned: boolean) {
	return {
		connectionId: connection.id,
		bankId: connection.bankId,
		bankName: connection.bankName,
		// A uuid is not an answer to "who connected this bank".
		connectedBy: connection.connectedBy?.displayName || 'someone else',
		accessEndsOn: connection.consentExpiresAt
			? new Date(Number(connection.consentExpiresAt.seconds) * 1000).toLocaleDateString(undefined, {
					day: 'numeric',
					month: 'long',
					year: 'numeric'
				})
			: '',
		live: connection.live,
		accountCount: 1,
		manageable: owned
	};
}

/**
 * Overview's own data. Who the member is comes from the layout; this load is
 * the accounts they may see.
 */
export const load: ServerLoad = async ({ cookies, url }) => {
	const view = await accounts(cookies);

	// What the last hand-off produced, if the member has just come back from
	// one. Every one of these redirected here and said nothing, which is how a
	// bank that granted access and exposed no accounts looked like nothing
	// happening at all.
	//
	// Which bank it was about, and whether there was a hand-off at all. Spent
	// on read: the outcome lives in the URL, so without this a member who
	// reloads Overview — or returns to it from history — is told again that a
	// bank exposed no accounts, weeks after connecting three others.
	const handoff = takeHandoff(cookies);

	const outcome = handoff ? outcomeOf(url.searchParams.get('outcome')) : undefined;

	return {
		...view,
		// A restore and a first connection return through the same route and
		// produce the same response; only the outgoing record tells them
		// apart, and the frames draw them as different notices.
		outcome: outcome === 'connected' && handoff?.restoring ? 'restored' : outcome,
		outcomeBank: handoff?.bankName || undefined,
		// The notice for a restored bank names the date access now runs to,
		// which is the bank's own limit and not a wimm policy.
		outcomeAccessEndsOn: view.banks?.find((bank) => bank.bankName === handoff?.bankName)
			?.accessEndsOn
	};
};

/** Only outcomes this app produces. A query parameter is not a message. */
function outcomeOf(raw: string | null): Outcome | undefined {
	switch (raw) {
		case 'connected':
		case 'declined':
		case 'no-accounts':
		case 'bank-unavailable':
		case 'already-connected':
			return raw;
		default:
			return undefined;
	}
}

export type Outcome =
	| 'connected'
	// Not produced by a query parameter: the return says "connected" and the
	// outgoing record says it was a restore.
	| 'restored'
	| 'disconnected'
	| 'declined'
	| 'no-accounts'
	| 'bank-unavailable'
	| 'already-connected';

/**
 * The banking half of the landing page.
 *
 * Banking is optional: an instance with no gateway configured answers
 * Unimplemented, and the screen shows its empty state rather than an error —
 * which is what lets this deploy before anyone holds gateway credentials.
 */
async function accounts(cookies: Cookies) {
	try {
		const view = await call(cookies, (options) => banking.listAccounts({ skipRead: false }, options));

		return {
			accounts: view.accounts.map((account) => ({
				id: account.id,
				// Which bank's card this row belongs under. Without it every
				// account fell through the "no bank behind it" branch and the
				// grouping the frame draws never rendered at all.
				connectionId: account.connection?.id,
				name: account.name,
				bank: account.connection?.bankName ?? '',
				// Absent at balance level, because the server sent none.
				numberSuffix: account.numberSuffix || undefined,
				balance: formatMoney(account.balance?.money),
				readAt: account.balance?.readAt ? relative(account.balance.readAt.seconds) : undefined,
				stale: account.balance?.stale ?? false,
				notUpdating: account.connection ? !account.connection.live : false,
				negative: isNegative(account.balance?.money)
			})),
			totals: view.totals.map((total) => ({
				total: formatMoney(total.total) ?? '',
				currency: total.total?.currency ?? ''
			})),
			// One entry per connection the member can see something of. Derived
			// from the accounts rather than fetched separately: a bank they can
			// see no account of must not appear, because its existence is
			// itself something they were not told.
			banks: banksFrom(view.accounts),
			problems: view.failures.map((failure) => ({
				connectionId: failure.connectionId,
				bankName: failure.bankName,
				kind: problemKind(failure.failure),
				// The clock time the frame draws, not a duration: "possible
				// after 14:20" is something a member can act on without doing
				// arithmetic against when they happened to load the page.
				retryAfter:
					failure.retryAfterSeconds > 0n
						? new Date(Date.now() + Number(failure.retryAfterSeconds) * 1000)
								.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
						: undefined
			}))
		};
	} catch {
		// No gateway, or the service behind it is down. Either way the member
		// sees the accounts screen with nothing on it rather than a failure.
		return { accounts: [], totals: [], problems: [], banks: [] };
	}
}

/**
 * Access having run out routes to restoring; everything else is transient and
 * routes to "could not be updated".
 */
function problemKind(failure: Failure): 'unreachable' | 'rate-limited' | 'access-ended' {
	if (failure === Failure.CONSENT_EXPIRED) return 'access-ended';
	if (failure === Failure.RATE_LIMITED) return 'rate-limited';
	return 'unreachable';
}

/**
 * A read time as a phrase, because "when did wimm ask" is the contract.
 *
 * Takes the protobuf timestamp's seconds directly rather than pulling in
 * `@bufbuild/protobuf` for one conversion: the web app does not otherwise
 * depend on the protobuf runtime, and keeping it that way is the point of
 * ADR 0001's rule that the browser never speaks Connect.
 */
function relative(epochSeconds: bigint): string {
	const seconds = Math.round(Date.now() / 1000 - Number(epochSeconds));
	if (seconds < 60) return 'just now';
	const minutes = Math.round(seconds / 60);
	if (minutes < 60) return `${minutes} ${minutes === 1 ? 'minute' : 'minutes'} ago`;
	const hours = Math.round(minutes / 60);
	if (hours < 24) return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
	const days = Math.round(hours / 24);
	return `${days} ${days === 1 ? 'day' : 'days'} ago`;
}
