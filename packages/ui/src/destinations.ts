/** One entry in the sidebar. Without an href it is shown but unreachable. */
export interface Destination {
	label: string;
	href?: string;
	current?: boolean;
}
