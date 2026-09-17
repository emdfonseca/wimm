import type { Cookies } from '@sveltejs/kit';

/**
 * Which bank a member was sent to, and whether they were restoring it.
 *
 * The bank's return carries `state` and `code` and nothing else, and a return
 * that failed has no response to read a name out of — so every notice on
 * Overview could only say "That bank", while the frames name the bank in every
 * one of them. This records the name on the way out instead.
 *
 * It also records the one thing no response can say: `CompleteConnection`
 * returns the same shape whether it made a connection or renewed one, so
 * "{bank} is updating again" is not derivable on the way back.
 *
 * It carries no authority. What a return is allowed to do is decided by wimmd
 * from the hash of `state`, and a member who rewrites this cookie changes the
 * name in a sentence they are reading to themselves.
 */
const COOKIE = 'wimm_handoff';

/** Long enough for a bank's own screens, short enough not to outlive them. */
const LIFETIME_SECONDS = 20 * 60;

export interface Handoff {
	bankName: string;
	restoring: boolean;
}

export function rememberHandoff(cookies: Cookies, handoff: Handoff): void {
	cookies.set(COOKIE, JSON.stringify(handoff), {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		secure: import.meta.env.PROD,
		maxAge: LIFETIME_SECONDS
	});
}

/**
 * The hand-off, spent. Reading it clears it: the notice is shown once, and a
 * member who reloads Overview an hour later is not told again about a bank
 * they have since been using.
 */
export function takeHandoff(cookies: Cookies): Handoff | undefined {
	const raw = cookies.get(COOKIE);
	if (!raw) return undefined;

	cookies.delete(COOKIE, { path: '/' });

	try {
		const parsed = JSON.parse(raw) as Partial<Handoff>;
		if (typeof parsed.bankName !== 'string') return undefined;
		// An empty name is still a hand-off: whether it was a restore is the
		// other half of this, and the screen has a fallback for the name.
		return { bankName: parsed.bankName, restoring: parsed.restoring === true };
	} catch {
		// A cookie this app did not write, or one truncated in transit.
		return undefined;
	}
}
