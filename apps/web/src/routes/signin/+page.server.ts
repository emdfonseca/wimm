import type { ServerLoad } from '@sveltejs/kit';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';

/**
 * `/signin`. The ceremony begins server side so the challenge is a row before
 * the page renders, and so the path the member was heading to is recorded
 * against the attempt rather than carried in the URL.
 */
export const load: ServerLoad = async ({ cookies, url }) => {
	const intendedPath = url.searchParams.get('next') ?? '';
	const expired = url.searchParams.has('expired');

	const ceremony = await call(cookies, (options) =>
		identity.beginSignIn({ intendedPath }, options)
	);

	return {
		ceremonyId: ceremony.ceremonyId,
		requestOptionsJson: ceremony.requestOptionsJson,
		expired
	};
};
