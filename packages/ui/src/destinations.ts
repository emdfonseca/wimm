/** One entry in the primary navigation. Without an href it is shown but
 *  unreachable, which is how the canvas draws a destination that does not
 *  exist yet — disabled, and a disabled node in pen is absent rather than
 *  greyed. */
export interface Destination {
	label: string;
	href?: string;
	current?: boolean;
	/** The glyph beside the label. The sidebar and the bottom bar draw the
	 *  same one, because they are the same destination in two places. */
	icon?: 'layout-dashboard' | 'arrow-left-right' | 'settings';
}

/**
 * Everywhere a member can go. One list, so the sidebar at Wide and the bottom
 * bar at Compact cannot disagree about what the product has.
 *
 * `current` is set from the route by the caller, because only the route knows.
 */
export function destinations(current?: string): Destination[] {
	return [
		{ label: 'Overview', href: '/', icon: 'layout-dashboard', current: current === '/' },
		{
			label: 'Transactions',
			href: '/transactions',
			icon: 'arrow-left-right',
			current: current === '/transactions'
		},
		{
			label: 'Settings',
			href: '/settings',
			icon: 'settings',
			current: current === '/settings'
		}
	];
}
