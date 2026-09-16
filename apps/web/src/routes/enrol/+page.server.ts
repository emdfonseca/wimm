import { redirect, type ServerLoad } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';

/**
 * `/enrol` — reached by the redirect from `/enrol/<link>`, so the value has
 * already left the URL and the ticket is in an HttpOnly cookie.
 *
 * Someone arriving with no ticket has no enrolment in flight and sees the same
 * dead end as an unusable link.
 */
export const load: ServerLoad = async ({ cookies }) => {
	try {
		const ceremony = await call(cookies, (options) =>
			// The ticket travels in the cookie the exchange set; the field is for
			// callers that hold it directly.
			identity.beginEnrolment({ enrolmentTicket: '' }, options)
		);

		return {
			ceremonyId: ceremony.ceremonyId,
			creationOptionsJson: ceremony.creationOptionsJson,
			firstName: ceremony.member?.firstName ?? '',
			lastName: ceremony.member?.lastName ?? ''
		};
	} catch (caught) {
		if (ConnectError.from(caught).code === Code.PermissionDenied) {
			redirect(303, '/enrol/unusable');
		}
		throw caught;
	}
};
