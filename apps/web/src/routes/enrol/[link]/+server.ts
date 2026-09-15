import { error, redirect, type RequestHandler } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { identity } from '$lib/server/identity';
import { call } from '$lib/server/call';

/**
 * The enrolment link's only job: exchange its value for a ticket, then send the
 * member to a path that does not contain it.
 *
 * The redirect is what stops the value travelling. Afterwards browser history
 * holds `/enrol`, and the ticket lives in an HttpOnly cookie no script can read.
 *
 * The link is not closed here. Closing it on load would make a dismissed prompt,
 * or a device that cannot save a discoverable credential, cost another round
 * trip through the operator. It closes when a passkey is actually saved
 * (ADR 0016).
 */
export const GET: RequestHandler = async ({ params, cookies }) => {
	const linkValue = params.link ?? '';

	try {
		await call(cookies, (options) => identity.redeemEnrolmentLink({ linkValue }, options));
	} catch (caught) {
		const failure = ConnectError.from(caught);
		// Expired, spent, replaced and never issued arrive as one code and render
		// one page. Nothing here tells them apart.
		if (failure.code === Code.PermissionDenied) {
			redirect(303, '/enrol/unusable');
		}
		error(500, 'Something went wrong');
	}

	redirect(303, '/enrol');
};
