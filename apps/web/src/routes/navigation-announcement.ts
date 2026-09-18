/**
 * The one route-to-name map the app needs (design.md decision: "navigation
 * feedback"). `NavigationProgress` draws the bar; this is the half it cannot
 * draw — the announcement says the screen reader landed on Transactions, not
 * merely that it landed somewhere.
 */
export const routeNames: Record<string, string> = {
	'/(app)': 'Overview',
	'/(app)/transactions': 'Transactions',
	'/(app)/settings': 'Settings',
	'/(app)/connect': 'Connect a bank',
	'/(app)/connect/[bank]': 'Choose a bank',
	'/(app)/connect/[bank]/accounts/[connection]': 'Choose accounts',
	'/(app)/connect/widen/[connection]': 'Include transactions',
	'/signin': 'Sign in',
	'/enrol': 'Enrol a passkey',
	'/enrol/unusable': 'That link no longer works'
};

/**
 * What the polite live region says: "Loading" the moment a navigation is
 * taken, the destination's name once it lands. A route with no name in the
 * map, or no navigation and no known route, says nothing — silence over a
 * guess.
 */
export function navigationAnnouncement(navigatingTo: boolean, routeId: string | null): string {
	if (navigatingTo) return 'Loading';
	if (!routeId) return '';
	return routeNames[routeId] ?? '';
}
