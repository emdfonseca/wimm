import { redirect, type ServerLoad } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { identity, SESSION_COOKIE } from '$lib/server/identity';
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
		const { member } = await call(cookies, (options) => identity.getCurrentMember({}, options));

		return {
			firstName: member?.firstName ?? '',
			lastName: member?.lastName ?? ''
		};
	} catch (caught) {
		if (ConnectError.from(caught).code !== Code.Unauthenticated) throw caught;

		const query: string[] = [];

		// "You were signed out" is only true if they were signed in. A browser
		// that sent no session cookie is arriving for the first time, and
		// telling it a session ended is telling it something that never
		// happened. The two are separate scenarios in
		// identity/passkey-sign-in, and separate states on the canvas.
		if (cookies.get(SESSION_COOKIE)) {
			query.push('expired');
			// And clear it, so the notice is shown once rather than on every
			// visit from now on. The session it names is already gone.
			cookies.delete(SESSION_COOKIE, { path: '/' });
		}

		// The landing is where sign-in goes anyway, so carrying it as a return
		// path says nothing and puts a redirect target in the URL for no reason.
		const intended = url.pathname + url.search;
		if (intended !== '/') query.push(`next=${encodeURIComponent(intended)}`);

		redirect(303, query.length ? `/signin?${query.join('&')}` : '/signin');
	}
};
