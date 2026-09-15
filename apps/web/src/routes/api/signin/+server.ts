import { json, type RequestHandler } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';

/** Finishes the sign-in ceremony and relays back the session cookie. */
export const POST: RequestHandler = async ({ request, cookies }) => {
	const { ceremonyId, credentialJson } = await request.json();

	try {
		const result = await call(cookies, (options) =>
			identity.finishSignIn({ ceremonyId, credentialJson }, options)
		);
		// The path came from the server's own record of the attempt, never from
		// the client, so there is nothing to tamper with here either.
		return json({ returnPath: result.returnPath });
	} catch (caught) {
		if (ConnectError.from(caught).code === Code.PermissionDenied) {
			return json({ reason: 'not-recognised' }, { status: 403 });
		}
		return json({ reason: 'failed' }, { status: 500 });
	}
};
