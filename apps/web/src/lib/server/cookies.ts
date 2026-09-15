import type { Cookies } from '@sveltejs/kit';

/**
 * wimmd owns the session, so it owns the cookie. The app's job is to carry it
 * in both directions without interpreting it: forward what the browser sent,
 * and relay back what wimmd set.
 *
 * Minting a second cookie here would put the session's lifetime in two places,
 * and only one of them can be revoked.
 */

/** Everything the browser sent, as a Cookie header for wimmd. */
export function forward(cookies: Cookies): Record<string, string> {
	const jar = cookies
		.getAll()
		.map(({ name, value }) => `${name}=${value}`)
		.join('; ');
	return jar ? { cookie: jar } : {};
}

/** Whatever wimmd set, applied to this response. */
export function relay(headers: Headers, cookies: Cookies): void {
	// getSetCookie keeps multiple Set-Cookie headers separate; reading the header
	// as a string joins them on a comma, which is unparseable once a cookie
	// carries an Expires date.
	const setCookie =
		typeof headers.getSetCookie === 'function'
			? headers.getSetCookie()
			: [headers.get('set-cookie')].filter((v): v is string => typeof v === 'string');

	for (const raw of setCookie) {
		const parsed = parse(raw);
		if (!parsed) continue;

		if (parsed.maxAge !== undefined && parsed.maxAge <= 0) {
			cookies.delete(parsed.name, { path: parsed.path });
			continue;
		}

		cookies.set(parsed.name, parsed.value, {
			path: parsed.path,
			httpOnly: parsed.httpOnly,
			secure: parsed.secure,
			sameSite: parsed.sameSite,
			expires: parsed.expires
		});
	}
}

interface Parsed {
	name: string;
	value: string;
	path: string;
	httpOnly: boolean;
	secure: boolean;
	sameSite: 'lax' | 'strict' | 'none';
	expires?: Date;
	maxAge?: number;
}

function parse(raw: string): Parsed | null {
	const [pair, ...rest] = raw.split(';');
	if (!pair) return null;

	const eq = pair.indexOf('=');
	if (eq < 0) return null;

	const attributes = new Map(
		rest.map((part) => {
			const [key, ...value] = part.trim().split('=');
			return [(key ?? '').toLowerCase(), value.join('=')];
		})
	);

	const sameSite = (attributes.get('samesite') ?? 'lax').toLowerCase();
	const maxAge = attributes.has('max-age') ? Number(attributes.get('max-age')) : undefined;
	const expiresAt = attributes.get('expires');

	return {
		name: pair.slice(0, eq).trim(),
		value: pair.slice(eq + 1),
		path: attributes.get('path') || '/',
		httpOnly: attributes.has('httponly'),
		secure: attributes.has('secure'),
		sameSite: sameSite === 'strict' ? 'strict' : sameSite === 'none' ? 'none' : 'lax',
		expires: expiresAt ? new Date(expiresAt) : undefined,
		maxAge: Number.isFinite(maxAge) ? maxAge : undefined
	};
}
