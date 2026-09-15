import type { Cookies } from '@sveltejs/kit';
import { forward, relay } from './cookies';

/**
 * One call to wimmd, with the browser's cookies carried there and wimmd's
 * carried back.
 *
 * wimmd keeps ownership of the session cookie and its attributes, because it
 * is the thing that can revoke a session. The app relays rather than mints:
 * two places setting one cookie is two lifetimes, and only one of them is real.
 */
export async function call<T>(
	cookies: Cookies,
	invoke: (options: {
		headers: Record<string, string>;
		onHeader: (headers: Headers) => void;
	}) => Promise<T>
): Promise<T> {
	let received: Headers | null = null;

	try {
		return await invoke({
			headers: forward(cookies),
			onHeader: (headers) => {
				received = headers;
			}
		});
	} finally {
		if (received) relay(received, cookies);
	}
}
