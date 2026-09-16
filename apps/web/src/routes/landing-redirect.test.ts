import { describe, expect, it, vi, beforeEach } from 'vitest';
import { Code, ConnectError } from '@connectrpc/connect';
import { isRedirect } from '@sveltejs/kit';

/**
 * Where an unauthenticated visitor is sent, and what they are told when they
 * get there.
 *
 * This shipped wrong: every first visit was told "You were signed out", because
 * the redirect added `expired` whenever the call came back unauthenticated
 * rather than whenever a session had actually ended.
 * `identity/passkey-sign-in` has those as two scenarios and the canvas draws
 * them as two states.
 */

const getCurrentMember = vi.fn();

vi.mock('$lib/server/identity', () => ({
	identity: {
		getCurrentMember: (...args: unknown[]) => getCurrentMember(...args)
	},
	SESSION_COOKIE: 'wimm_session'
}));

// The real one carries cookies to wimmd and back; here it just invokes.
vi.mock('$lib/server/call', () => ({
	call: <T>(_cookies: unknown, invoke: (o: unknown) => Promise<T>) =>
		invoke({ headers: {}, onHeader: () => {} })
}));

const { load } = await import('./+page.server');

function visit(path: string, jar: Record<string, string> = {}) {
	const deleted: string[] = [];
	const event = {
		url: new URL(`http://localhost${path}`),
		cookies: {
			get: (name: string) => jar[name],
			delete: (name: string) => deleted.push(name)
		}
	};
	// The load signature carries far more than this test supplies; it reads
	// only url and cookies.
	return { run: () => (load as (e: unknown) => unknown)(event), deleted };
}

async function redirectFrom(path: string, jar?: Record<string, string>) {
	const { run, deleted } = visit(path, jar);
	try {
		await run();
	} catch (thrown) {
		if (!isRedirect(thrown)) throw thrown;
		return { status: thrown.status, location: thrown.location, deleted };
	}
	throw new Error('expected a redirect, got a page');
}

beforeEach(() => {
	getCurrentMember.mockReset();
	getCurrentMember.mockRejectedValue(new ConnectError('no session', Code.Unauthenticated));
});

describe('a visitor without a session', () => {
	it('is sent to sign in with nothing claimed about a session ending', async () => {
		const { status, location } = await redirectFrom('/');
		expect(status).toBe(303);
		expect(location).toBe('/signin');
	});

	it('does not carry the landing as a return path', async () => {
		const { location } = await redirectFrom('/');
		expect(location).not.toContain('next');
	});

	it('carries the path it was actually trying to reach', async () => {
		const { location } = await redirectFrom('/accounts?filter=open');
		expect(location).toBe(`/signin?next=${encodeURIComponent('/accounts?filter=open')}`);
		expect(location).not.toContain('expired');
	});
});

describe('a visitor whose session has ended', () => {
	it('is told so, because this time there was a session', async () => {
		const { location } = await redirectFrom('/', { wimm_session: 'stale' });
		expect(location).toBe('/signin?expired');
	});

	it('has the dead cookie cleared, so it is said once and not forever', async () => {
		const { deleted } = await redirectFrom('/', { wimm_session: 'stale' });
		expect(deleted).toContain('wimm_session');
	});

	it('still gets the return path when it was going somewhere', async () => {
		const { location } = await redirectFrom('/accounts', { wimm_session: 'stale' });
		expect(location).toBe(`/signin?expired&next=${encodeURIComponent('/accounts')}`);
	});
});

describe('a visitor with a live session', () => {
	it('is shown the page, named', async () => {
		getCurrentMember.mockResolvedValue({ member: { firstName: 'Ada', lastName: 'Lovelace' } });

		const { run } = visit('/');
		await expect(run()).resolves.toEqual({ firstName: 'Ada', lastName: 'Lovelace' });
	});
});

describe('anything that is not an authentication problem', () => {
	it('is not turned into a redirect', async () => {
		getCurrentMember.mockRejectedValue(new ConnectError('database is down', Code.Internal));

		const { run } = visit('/');
		await expect(run()).rejects.toThrow('database is down');
	});
});
