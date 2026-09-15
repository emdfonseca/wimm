import { json, type RequestHandler } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';

/**
 * Finishes the enrolment ceremony. The browser cannot call wimmd itself, so
 * this is the app's own endpoint: it takes the authenticator's response, hands
 * it on, and relays back the session cookie wimmd sets.
 */
export const POST: RequestHandler = async ({ request, cookies }) => {
	const { ceremonyId, credentialJson } = await request.json();

	try {
		const result = await call(cookies, (options) =>
			identity.finishEnrolment({ ceremonyId, credentialJson }, options)
		);
		return json({ member: result.member });
	} catch (caught) {
		const failure = ConnectError.from(caught);
		// The link is untouched, so the member can try again on another device.
		if (failure.code === Code.FailedPrecondition) {
			return json({ reason: 'not-saved' }, { status: 422 });
		}
		if (failure.code === Code.PermissionDenied) {
			return json({ reason: 'unusable' }, { status: 403 });
		}
		return json({ reason: 'failed' }, { status: 500 });
	}
};
