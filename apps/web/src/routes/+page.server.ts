import { redirect, type ServerLoad } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';

/**
 * The landing page, behind a session.
 *
 * Someone without one is sent to sign in rather than shown a dead end, and the
 * path they were trying to reach travels as a query parameter only so far as
 * the sign-in route; wimmd records it against the attempt and ignores anything
 * a client supplies later.
 */
export const load: ServerLoad = async ({ cookies, url }) => {
	try {
		const { member } = await call(cookies, (options) =>
			identity.getCurrentMember({}, options)
		);

		return {
			firstName: member?.firstName ?? '',
			lastName: member?.lastName ?? ''
		};
	} catch (caught) {
		if (ConnectError.from(caught).code === Code.Unauthenticated) {
			const next = encodeURIComponent(url.pathname + url.search);
			redirect(303, `/signin?expired&next=${next}`);
		}
		throw caught;
	}
};
