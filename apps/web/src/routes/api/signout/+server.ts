import { json, type RequestHandler } from '@sveltejs/kit';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';

/** Ends the session on this browser. wimmd revokes the row; the cookie it
 *  clears is relayed back, so going back in the browser shows nothing. */
export const POST: RequestHandler = async ({ cookies }) => {
	await call(cookies, (options) => identity.signOut({}, options));
	return json({ ok: true });
};
