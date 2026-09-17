import { redirect, type ServerLoad } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { identity, SESSION_COOKIE } from '$lib/server/identity';
import { call } from '$lib/server/call';

/**
 * Everything behind a session, and everything inside the app shell.
 *
 * The session check lives here rather than in each page because a screen added
 * later cannot forget a layout. The same reasoning applies to the shell itself
 * in `+layout.svelte`: `/connect` shipped without it, flush to the left edge
 * with no nav, precisely because wrapping was each page's job.
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
		// that sent no session cookie is arriving for the first time.
		if (cookies.get(SESSION_COOKIE)) {
			query.push('expired');
			cookies.delete(SESSION_COOKIE, { path: '/' });
		}

		const intended = url.pathname + url.search;
		if (intended !== '/') query.push(`next=${encodeURIComponent(intended)}`);

		redirect(303, query.length ? `/signin?${query.join('&')}` : '/signin');
	}
};
