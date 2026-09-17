import type { ServerLoad } from '@sveltejs/kit';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';

/**
 * J03.A / 01 · Choose a bank.
 *
 * Behind a session, like every banking route: there is no unauthenticated
 * banking call, and the member the session names is whose view this is.
 *
 * A bank list that cannot be loaded is a state of this page rather than an
 * error page. Nothing has been connected, and the member can try again.
 */
export const load: ServerLoad = async ({ cookies, url }) => {
	const country = url.searchParams.get('country') ?? 'PT';

	try {
		const { banks } = await call(cookies, (options) =>
			banking.listBanks({ country }, options)
		);

		return {
			state: 'default' as const,
			banks: banks.map((bank) => ({
				id: bank.id,
				name: bank.name,
				logoUrl: bank.logoUrl || undefined
			}))
		};
	} catch {
		// The session gate is the (app) layout's, and it has already run. Every
		// failure that reaches here is the same thing to a member: wimm could
		// not list banks, and nothing was connected.
		return { state: 'unavailable' as const, banks: [] };
	}
};
