import { describe, expect, it, vi, beforeEach } from 'vitest';
import { Code, ConnectError } from '@connectrpc/connect';
import { isRedirect } from '@sveltejs/kit';

/**
 * The banking routes' own behaviour: who is let in, and what is left in the
 * URL afterwards.
 *
 * The return route is the one that matters most. The value a bank sends back
 * is a bearer credential while it is live, so it must be exchanged server side
 * and redirected away from — the shape ADR 0016 established for enrolment
 * links and ADR 0018 reuses. A test that only checked "the member ends up on
 * the chooser" would pass with the value still in the address.
 */

const listBanks = vi.fn();
const beginConnection = vi.fn();
const completeConnection = vi.fn();
const listConnectionAccounts = vi.fn();
const setAccountOwners = vi.fn();
const setAccountLevel = vi.fn();
const setAccountLeftOut = vi.fn();
const setAccountName = vi.fn();
const getCurrentMember = vi.fn();

vi.mock('$lib/server/banking', () => ({
	banking: {
		listBanks: (...args: unknown[]) => listBanks(...args),
		beginConnection: (...args: unknown[]) => beginConnection(...args),
		completeConnection: (...args: unknown[]) => completeConnection(...args),
		listConnectionAccounts: (...args: unknown[]) => listConnectionAccounts(...args),
		setAccountOwners: (...args: unknown[]) => setAccountOwners(...args),
		setAccountLevel: (...args: unknown[]) => setAccountLevel(...args),
		setAccountLeftOut: (...args: unknown[]) => setAccountLeftOut(...args),
		setAccountName: (...args: unknown[]) => setAccountName(...args)
	}
}));

vi.mock('$lib/server/identity', () => ({
	identity: { getCurrentMember: (...args: unknown[]) => getCurrentMember(...args) },
	SESSION_COOKIE: 'wimm_session'
}));

vi.mock('$lib/server/call', () => ({
	call: <T>(_cookies: unknown, invoke: (o: unknown) => Promise<T>) =>
		invoke({ headers: {}, onHeader: () => {} })
}));

const cookies = { get: () => undefined, delete: () => {}, set: () => {} };

/**
 * Runs a load or handler and returns the redirect it threw, or null.
 *
 * The callback returns `unknown` rather than a promise: SvelteKit's handlers
 * are typed `MaybePromise`, and awaiting covers both.
 */
async function redirectFrom(run: () => unknown) {
	try {
		const result = await run();
		return { location: null as string | null, result };
	} catch (caught) {
		if (!isRedirect(caught)) throw caught;
		return { location: caught.location as string, result: null };
	}
}

beforeEach(() => {
	vi.resetAllMocks();
});

describe('/connect', () => {
	// The session gate is the (app) layout's and has already run by the time
	// this load is called, so an auth failure here is just another reason the
	// list is unavailable — see session-gate.test.ts for the redirect itself.
	it('does not try to guard the session itself', async () => {
		listBanks.mockRejectedValue(new ConnectError('nope', Code.Unauthenticated));

		const { load } = await import('./+page.server');
		const data = (await load!({
			cookies,
			url: new URL('http://localhost/connect')
		} as never)) as { state: string };

		expect(data.state).toBe('unavailable');
	});

	// A list that cannot be loaded is a state of the page, not an error page:
	// nothing has been connected and the member can try again.
	it('renders its own unavailable state rather than failing', async () => {
		listBanks.mockRejectedValue(new ConnectError('down', Code.Unavailable));

		const { load } = await import('./+page.server');
		const data = (await load!({
			cookies,
			url: new URL('http://localhost/connect')
		} as never)) as { state: string; banks: unknown[] };

		expect(data.state).toBe('unavailable');
		expect(data.banks).toEqual([]);
	});

	it('lists the banks it was given', async () => {
		listBanks.mockResolvedValue({
			banks: [{ id: 'PT:Montepio', name: 'Montepio', logoUrl: '' }]
		});

		const { load } = await import('./+page.server');
		const data = (await load!({
			cookies,
			url: new URL('http://localhost/connect')
		} as never)) as { banks: { id: string; logoUrl?: string }[] };

		expect(data.banks).toHaveLength(1);
		// An empty logo is absent, not an empty string a component would render
		// as a broken image.
		expect(data.banks[0]!.logoUrl).toBeUndefined();
	});
});

describe('/connect/return', () => {
	it('redirects to a path that does not carry the one-time value', async () => {
		completeConnection.mockResolvedValue({
			connection: { id: 'c1', bankId: 'PT:Montepio' }
		});

		const { GET } = await import('./return/+server');
		const { location } = await redirectFrom(() =>
			GET!({
				cookies,
				url: new URL('http://localhost/connect/return?state=secret-state&code=secret-code')
			} as never)
		);

		expect(location).toBe('/?outcome=connected');
		// The assertion that matters: neither value survives into the address.
		expect(location).not.toContain('secret-state');
		expect(location).not.toContain('secret-code');
		expect(location).not.toContain('state=');
		expect(location).not.toContain('code=');
	});

	// Straight to Overview, with no stop in between.
	//
	// ADR 0018 put a chooser here because a `shared` flag meant connecting a
	// bank to share a joint account exposed every personal account at it. Under
	// ADR 0019 accounts arrive owned by the connecting member with nobody else
	// granted anything, so the question that screen asked has a correct answer
	// already filled in.
	it('lands on Overview rather than a sharing step', async () => {
		completeConnection.mockResolvedValue({
			connection: { id: 'conn-42', bankId: 'PT:Montepio' }
		});

		const { GET } = await import('./return/+server');
		const { location } = await redirectFrom(() =>
			GET!({
				cookies,
				url: new URL('http://localhost/connect/return?state=s&code=c')
			} as never)
		);

		expect(location).toBe('/?outcome=connected');
		expect(location).not.toContain('/accounts');
	});

	it('passes the callback through to wimmd rather than interpreting it', async () => {
		completeConnection.mockResolvedValue({ connection: { bankId: 'PT:Montepio' } });

		const { GET } = await import('./return/+server');
		await redirectFrom(() =>
			GET!({
				cookies,
				url: new URL('http://localhost/connect/return?state=s&code=c')
			} as never)
		);

		expect(completeConnection).toHaveBeenCalledWith(
			{ state: 's', code: 'c', error: '' },
			expect.anything()
		);
	});

	// A replayed return finds the row already consumed. wimmd answers
	// PermissionDenied for a refusal and NotFound for one that never existed
	// or expired, and the two land in different places.
	it('sends a declined consent home, and nothing is connected', async () => {
		completeConnection.mockRejectedValue(new ConnectError('declined', Code.PermissionDenied));

		const { GET } = await import('./return/+server');
		const { location } = await redirectFrom(() =>
			GET!({ cookies, url: new URL('http://localhost/connect/return?state=s&error=denied') } as never)
		);

		expect(location).toBe('/?outcome=declined');
	});

	it('sends an expired or unknown return back to the picker', async () => {
		completeConnection.mockRejectedValue(new ConnectError('gone', Code.NotFound));

		const { GET } = await import('./return/+server');
		const { location } = await redirectFrom(() =>
			GET!({ cookies, url: new URL('http://localhost/connect/return?state=s&code=c') } as never)
		);

		expect(location).toBe('/connect?return-expired');
	});

	// A return with no state at all is not a connection attempt: nothing is
	// called, because there is nothing to exchange.
	it('refuses a return with no state without calling wimmd', async () => {
		const { GET } = await import('./return/+server');
		const { location } = await redirectFrom(() =>
			GET!({ cookies, url: new URL('http://localhost/connect/return?code=c') } as never)
		);

		expect(location).toBe('/connect?returned-without-state');
		expect(completeConnection).not.toHaveBeenCalled();
	});

	// Gateways disagree on the parameter's name; both mean the member said no.
	it('recognises a refusal under either spelling', async () => {
		completeConnection.mockResolvedValue({ connection: { bankId: 'b' } });

		const { GET } = await import('./return/+server');
		await redirectFrom(() =>
			GET!({
				cookies,
				url: new URL('http://localhost/connect/return?state=s&error_code=access_denied')
			} as never)
		);

		expect(completeConnection).toHaveBeenCalledWith(
			expect.objectContaining({ error: 'access_denied' }),
			expect.anything()
		);
	});
});

describe('/connect/[bank]', () => {
	it('carries the bank’s own consent maximum through to the page', async () => {
		listBanks.mockResolvedValue({
			banks: [{ id: 'PT:ActivoBank', name: 'ActivoBank', maxConsentSeconds: 86400n }]
		});

		const { load } = await import('./[bank]/+page.server');
		const data = (await load!({
			cookies,
			params: { bank: 'PT:ActivoBank' },
			url: new URL('http://localhost/connect/PT:ActivoBank')
		} as never)) as { maxConsentSeconds: number; bankName: string };

		// One day, not the 90 most banks grant. The screen states the real date
		// because of this number.
		expect(data.maxConsentSeconds).toBe(86400);
		expect(data.bankName).toBe('ActivoBank');
	});

	it('sends a member to the picker when the bank is not offered', async () => {
		listBanks.mockResolvedValue({ banks: [] });

		const { load } = await import('./[bank]/+page.server');
		const { location } = await redirectFrom(() =>
			load!({
				cookies,
				params: { bank: 'PT:Nope' },
				url: new URL('http://localhost/connect/PT:Nope')
			} as never)
		);

		expect(location).toBe('/connect');
	});

	// Arriving at the explainer must not create a pending row: a member who
	// opens it and leaves should cost nothing.
	it('does not begin a connection on load', async () => {
		listBanks.mockResolvedValue({
			banks: [{ id: 'b', name: 'Bank', maxConsentSeconds: 100n }]
		});

		const { load } = await import('./[bank]/+page.server');
		await load!({
			cookies,
			params: { bank: 'b' },
			url: new URL('http://localhost/connect/b')
		} as never);

		expect(beginConnection).not.toHaveBeenCalled();
	});

	it('begins the connection and hands off only when the form is submitted', async () => {
		beginConnection.mockResolvedValue({ handoffUrl: 'https://bank.example/consent/xyz' });

		const { actions } = await import('./[bank]/+page.server');
		const submitted = new FormData();
		submitted.set('bankName', 'Monzo');

		const { location } = await redirectFrom(() =>
			actions.default({
				cookies,
				params: { bank: 'b' },
				// The name the page already loaded, so Overview can name the
				// bank in a notice on the way back.
				request: new Request('http://localhost/connect/b', {
					method: 'POST',
					body: submitted
				})
			} as never)
		);

		expect(beginConnection).toHaveBeenCalledWith({ bankId: 'b' }, expect.anything());
		expect(location).toBe('https://bank.example/consent/xyz');
	});
});

describe('/connect/[bank]/accounts', () => {
	// Route-backed and resumable on purpose: by the time a member reaches the
	// chooser the bank has already granted access, so closing the tab must not
	// cost the connection.
	it('a member who left mid-choice comes back to the same choice', async () => {
		listConnectionAccounts.mockResolvedValue({
			accounts: [
				{
					id: 'a1',
					name: 'Conta à Ordem',
					numberSuffix: '0538',
					owned: true,
					owners: [{ id: 'ada', displayName: 'Ada' }],
					connection: { bankName: 'Montepio' },
					balance: { money: { minor: 420010n, currency: 'EUR' } }
				}
			],
			members: [
				{ id: 'ada', displayName: 'Ada' },
				{ id: 'grace', displayName: 'Grace' }
			],
			grants: [{ accountId: 'a1', memberId: 'grace', level: 2 }]
		});
		getCurrentMember.mockResolvedValue({ member: { id: 'ada' } });

		const { load } = await import('./[bank]/accounts/[connection]/+page.server');
		const data = (await load!({
			cookies,
			params: { bank: 'PT:Montepio', connection: 'c1' },
			url: new URL('http://localhost/connect/PT:Montepio/accounts/c1')
		} as never)) as {
			accounts: { owned: boolean; balance?: string }[];
			levels: Record<string, Record<string, string>>;
			members: { id: string }[];
			readOnly: boolean;
		};

		// The choice as it stands, not as it started.
		expect(data.levels).toEqual({ a1: { grace: 'balance' } });
		expect(data.accounts[0]!.owned).toBe(true);
		expect(data.readOnly).toBe(false);
	});

	// A level control for yourself is a control for a decision ownership has
	// already made.
	it('offers levels for everyone except the member looking at the screen', async () => {
		listConnectionAccounts.mockResolvedValue({
			accounts: [],
			members: [
				{ id: 'ada', displayName: 'Ada' },
				{ id: 'grace', displayName: 'Grace' }
			],
			grants: []
		});
		getCurrentMember.mockResolvedValue({ member: { id: 'ada' } });

		const { load } = await import('./[bank]/accounts/[connection]/+page.server');
		const data = (await load!({
			cookies,
			params: { bank: 'b', connection: 'c1' },
			url: new URL('http://localhost/connect/b/accounts/c1')
		} as never)) as { members: { id: string }[]; currentMemberId: string };

		expect(data.members.map((m) => m.id)).toEqual(['grace']);
		expect(data.currentMemberId).toBe('ada');
	});

	// UNSPECIFIED and HIDDEN both mean no grant: the failure mode has to be
	// accidentally hidden, never accidentally visible.
	it('treats an unset level as hidden', async () => {
		listConnectionAccounts.mockResolvedValue({
			accounts: [],
			members: [{ id: 'grace', displayName: 'Grace' }],
			grants: [{ accountId: 'a1', memberId: 'grace', level: 0 }]
		});
		getCurrentMember.mockResolvedValue({ member: { id: 'ada' } });

		const { load } = await import('./[bank]/accounts/[connection]/+page.server');
		const data = (await load!({
			cookies,
			params: { bank: 'b', connection: 'c1' },
			url: new URL('http://localhost/connect/b/accounts/c1')
		} as never)) as { levels: Record<string, Record<string, string>> };

		expect(data.levels.a1!.grace).toBe('hidden');
	});

	// A member who owns none of it may look and change nothing. The screen is
	// still shown, because they may hold a level on one of these.
	it('shows a read-only chooser to a member who owns none of it', async () => {
		listConnectionAccounts.mockRejectedValue(new ConnectError('not yours', Code.PermissionDenied));

		const { load } = await import('./[bank]/accounts/[connection]/+page.server');
		const data = (await load!({
			cookies,
			params: { bank: 'b', connection: 'c1' },
			url: new URL('http://localhost/connect/b/accounts/c1')
		} as never)) as { readOnly: boolean };

		expect(data.readOnly).toBe(true);
	});

	it('sends a member who is not signed in to sign in', async () => {
		listConnectionAccounts.mockRejectedValue(new ConnectError('nope', Code.Unauthenticated));

		const { load } = await import('./[bank]/accounts/[connection]/+page.server');
		const { location } = await redirectFrom(() =>
			load!({
				cookies,
				params: { bank: 'b' },
				url: new URL('http://localhost/connect/b/accounts?connection=c1')
			} as never)
		);

		expect(location).toContain('/signin?next=');
	});

	// A refusal is a form failure, not an error page: the member is looking at
	// a screen full of other controls that still work.
	it('reports a refused change as a form failure', async () => {
		setAccountLevel.mockRejectedValue(new ConnectError('not yours', Code.PermissionDenied));

		const { actions } = await import('./[bank]/accounts/[connection]/+page.server');
		const form = new FormData();
		form.set('accountId', 'a1');
		form.set('memberId', 'grace');
		form.set('level', 'balance');

		const result = (await actions.level!({
			cookies,
			request: { formData: async () => form }
		} as never)) as { status: number; data: { reason: string } };

		expect(result.status).toBe(403);
		expect(result.data.reason).toBe('not-yours');
	});

	it('sets a level by name', async () => {
		setAccountLevel.mockResolvedValue({});

		const { actions } = await import('./[bank]/accounts/[connection]/+page.server');
		const form = new FormData();
		form.set('accountId', 'a1');
		form.set('memberId', 'grace');
		form.set('level', 'details');

		await actions.level!({ cookies, request: { formData: async () => form } } as never);

		expect(setAccountLevel).toHaveBeenCalledWith(
			// Level.DETAILS is 3 on the wire.
			{ accountId: 'a1', memberId: 'grace', level: 3 },
			expect.anything()
		);
	});

	// Disowning sends an empty list: the account becomes nobody's, and nothing
	// reads it.
	it('disowning sends no owners at all', async () => {
		setAccountOwners.mockResolvedValue({});

		const { actions } = await import('./[bank]/accounts/[connection]/+page.server');
		const form = new FormData();
		form.set('accountId', 'a1');

		await actions.owners!({ cookies, request: { formData: async () => form } } as never);

		expect(setAccountOwners).toHaveBeenCalledWith(
			{ accountId: 'a1', memberIds: [] },
			expect.anything()
		);
	});

	// The last owner cannot step back (ADR 0022): the database's refusal
	// surfaces as FailedPrecondition, and the route reports it as a 409 with a
	// reason the screen recognises, distinct from a plain failure or a
	// permission refusal.
	it('reports the last-owner refusal distinctly from a permission refusal', async () => {
		setAccountOwners.mockRejectedValue(
			new ConnectError('would have no owner', Code.FailedPrecondition)
		);

		const { actions } = await import('./[bank]/accounts/[connection]/+page.server');
		const form = new FormData();
		form.set('accountId', 'a1');

		const result = (await actions.owners!({
			cookies,
			request: { formData: async () => form }
		} as never)) as { status: number; data: { reason: string } };

		expect(result.status).toBe(409);
		expect(result.data.reason).toBe('last-owner');
	});

	it('leaves an account out', async () => {
		setAccountLeftOut.mockResolvedValue({});

		const { actions } = await import('./[bank]/accounts/[connection]/+page.server');
		const form = new FormData();
		form.set('accountId', 'a1');

		await actions.leaveOut!({ cookies, request: { formData: async () => form } } as never);

		expect(setAccountLeftOut).toHaveBeenCalledWith(
			{ accountId: 'a1', leftOut: true },
			expect.anything()
		);
	});

	it('brings an account back', async () => {
		setAccountLeftOut.mockResolvedValue({});

		const { actions } = await import('./[bank]/accounts/[connection]/+page.server');
		const form = new FormData();
		form.set('accountId', 'a1');

		await actions.bringBack!({ cookies, request: { formData: async () => form } } as never);

		expect(setAccountLeftOut).toHaveBeenCalledWith(
			{ accountId: 'a1', leftOut: false },
			expect.anything()
		);
	});

	it('renames an account', async () => {
		setAccountName.mockResolvedValue({});

		const { actions } = await import('./[bank]/accounts/[connection]/+page.server');
		const form = new FormData();
		form.set('accountId', 'a1');
		form.set('householdName', 'Rent');

		await actions.rename!({ cookies, request: { formData: async () => form } } as never);

		expect(setAccountName).toHaveBeenCalledWith(
			{ accountId: 'a1', householdName: 'Rent' },
			expect.anything()
		);
	});
});
