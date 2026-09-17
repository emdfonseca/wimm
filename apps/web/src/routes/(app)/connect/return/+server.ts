import { isRedirect, redirect, type RequestHandler } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';

/**
 * The bank's return. Its only job is to exchange the one-time value and send
 * the member to a path that does not contain it.
 *
 * The redirect is what stops the value travelling. Afterwards browser history
 * holds `/connect/<bank>/accounts`, and nothing replayable is left in an
 * address a member can bookmark or share — the same shape `/enrol/[link]` uses
 * (ADR 0016, reused by ADR 0018).
 *
 * A replay finds the pending row already consumed. wimmd answers
 * PermissionDenied for a refusal and NotFound for a value that never existed
 * or expired, and those land in different places: "this was already used" and
 * "this never existed" are different answers and only one of them is true.
 */
export const GET: RequestHandler = async ({ url, cookies }) => {
	const state = url.searchParams.get('state') ?? '';
	const code = url.searchParams.get('code') ?? '';
	// Gateways disagree on the name; both mean the member said no.
	const failure = url.searchParams.get('error') ?? url.searchParams.get('error_code') ?? '';

	// Not a connection attempt at all: there is nothing to exchange, so nothing
	// is called.
	if (!state) redirect(303, '/connect?returned-without-state');

	try {
		await call(cookies, (options) =>
			banking.completeConnection({ state, code, error: failure }, options)
		);
	} catch (caught) {
		// redirect() throws. Calling it inside this try would have its own
		// control flow caught here and remapped to a failure, which is exactly
		// the defect the route tests found — so every redirect is outside.
		if (isRedirect(caught)) throw caught;

		const problem = ConnectError.from(caught);

		if (problem.code === Code.Unauthenticated) redirect(303, '/signin');

		// Declined at the bank. Nothing was connected, and a connection they
		// were restoring is exactly as it was.
		if (problem.code === Code.PermissionDenied) redirect(303, '/?outcome=declined');

		if (problem.code === Code.NotFound) redirect(303, '/connect?return-expired');

		// A return that was already exchanged. The connection it made exists;
		// saying "this never happened" would be false.
		if (problem.code === Code.AlreadyExists) redirect(303, '/?outcome=already-connected');

		// Access granted that exposes nothing, and consent that has already
		// run out. Both are FailedPrecondition and both need saying: a bank
		// that grants access and offers no accounts otherwise looks like
		// nothing happened at all.
		if (problem.code === Code.FailedPrecondition) {
			redirect(303, '/?outcome=no-accounts');
		}

		redirect(303, '/?outcome=bank-unavailable');
	}

	// Overview, not the chooser.
	//
	// ADR 0018 put a chooser here because a `shared` flag meant connecting a
	// bank to share a joint account exposed every personal account at it. Under
	// ADR 0019 that cannot happen: accounts arrive owned by the connecting
	// member and nobody else sees anything until granted. So the question the
	// chooser asked at this point has a correct answer already filled in, and
	// asking it is a step that protects nothing.
	//
	// Who sees what is a later, separate action, reached from Overview.
	redirect(303, '/?outcome=connected');
};
