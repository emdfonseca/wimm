import { redirect, type ServerLoad } from '@sveltejs/kit';
import { Code, ConnectError } from '@connectrpc/connect';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';
import { rememberHandoff } from '$lib/server/handoff';

/**
 * J03.A / 02 · What wimm will see, then the hand-off.
 *
 * The date shown is the one wimmd got from this bank, not a wimm policy: it is
 * 90 days at some banks and 1 day at others, and a member who is not shown the
 * real date reads a daily prompt as a defect.
 *
 * Beginning the connection is a form action rather than part of the load, so
 * arriving at this page does not create a pending row — a member who opens it
 * and leaves should cost nothing.
 */
export const load: ServerLoad = async ({ cookies, params, url }) => {
	const bankId = params.bank ?? '';

	// The listing is the only thing that can fail here. Redirects are decided
	// afterwards, because redirect() throws and calling it inside the try would
	// have its own control flow caught below and remapped to a failure.
	let banks: { id: string; name: string; maxConsentSeconds: bigint }[];
	try {
		({ banks } = await call(cookies, (options) => banking.listBanks({ country: 'PT' }, options)));
	} catch (caught) {
		if (ConnectError.from(caught).code === Code.Unauthenticated) {
			redirect(303, `/signin?next=${encodeURIComponent(url.pathname)}`);
		}
		redirect(303, '/connect?unavailable');
	}

	const bank = banks.find((candidate) => candidate.id === bankId);
	if (!bank) redirect(303, '/connect');

	return {
		bankId: bank.id,
		bankName: bank.name,
		maxConsentSeconds: Number(bank.maxConsentSeconds)
	};
};

export const actions = {
	/**
	 * Begins the connection and sends the member to their bank.
	 *
	 * wimmd mints the state value and stores only its hash, so nothing here
	 * holds anything replayable.
	 */
	default: async ({ cookies, params, request }) => {
		const bankId = params.bank ?? '';

		const { handoffUrl } = await call(cookies, (options) =>
			banking.beginConnection({ bankId }, options)
		);

		// The name travels from the page that already loaded it rather than
		// through a second listing: the bank's return carries no name, and
		// Overview names the bank in every notice it draws.
		const submitted = await request.formData();
		const bankName = submitted.get('bankName');
		rememberHandoff(cookies, {
			bankName: typeof bankName === 'string' ? bankName : bankId,
			restoring: false
		});

		redirect(303, handoffUrl);
	}
};
