<script lang="ts">
	import type { Snippet } from 'svelte';
	import SidebarNav from '../organisms/SidebarNav.svelte';
	import BottomNav from '../organisms/BottomNav.svelte';
	import Brand from '../atoms/Brand.svelte';
	import IconButton from '../atoms/IconButton.svelte';
	import { destinations as allDestinations, type Destination } from '../destinations.js';

	/**
	 * Where a member arrives once they are in. Origins `IGbQe` (Wide and Ultra)
	 * and `pu6qZ` (Compact).
	 *
	 * There is no page header. `App header` is a stack of breadcrumbs, a title
	 * row and tabs; here all three are empty but the title, which would render
	 * one word at 32 in a box built for three rows, repeating what the sidebar
	 * already marks as current. A header with nothing to carry is not a smaller
	 * header, it is no header. It returns when the page gains breadcrumbs, tabs
	 * or an action.
	 *
	 * At Compact there is no sidebar: a 390 viewport has no room for a 264
	 * column. The brand and the signed-in member do not vanish with it — they
	 * move into a top bar, because a screen showing neither does not tell a
	 * member which app they are in or who they are signed in as. That bar is
	 * not a page header: it carries no title, no breadcrumbs and no tabs.
	 *
	 * Navigation at Compact is a bottom bar rather than a hamburger, and it is
	 * a sibling below the scrolling region rather than an overlay — so nothing
	 * is obscured and the body needs no bottom padding to clear it.
	 */
	interface Props {
		memberName: string;
		memberInitials?: string;
		destinations?: Destination[];
		onsignout?: () => void;
		children?: Snippet;
	}

	let {
		memberName,
		memberInitials,
		// Overview and Transactions. Both exist; anything else the canvas draws
		// is disabled, and a disabled node is absent rather than greyed.
		//
		// Nothing is current by default. Only the route knows which page this
		// is, so a caller that forgets to say gets no marker rather than a
		// wrong one — a sidebar that always says Overview is worse than one
		// that says nothing.
		destinations = allDestinations(),
		onsignout,
		children
	}: Props = $props();

	const initials = $derived(
		memberInitials ??
			memberName
				.split(/\s+/)
				.filter(Boolean)
				.slice(0, 2)
				.map((part) => part[0]?.toUpperCase() ?? '')
				.join('')
	);
</script>

<div class="shell">
	<header class="bar">
		<Brand />
		<span class="bar-spacer"></span>
		<span class="avatar" aria-hidden="true">{initials}</span>
		{#if onsignout}
			<IconButton icon="log-out" label="Sign out" onclick={onsignout} />
		{/if}
	</header>

	<div class="sidebar">
		<SidebarNav {destinations} {memberName} {onsignout} />
	</div>
	<main class="body">
		{#if children}{@render children()}{/if}
	</main>

	<div class="bottom">
		<BottomNav {destinations} />
	</div>
</div>

<style>
	/* svh, not dvh: dvh recomputes as a mobile URL bar shows and hides, which
	   re-runs this layout mid-scroll and is what made scrolling judder. The
	   small viewport height is the one that does not move. */
	.shell {
		display: flex;
		flex-direction: column;
		min-block-size: 100svh;
		background: var(--color-bg-canvas);
		font-family: var(--type-family-body);
		color: var(--color-text-primary);
	}

	/* The Compact bar. 56 is layout-header-height at this regime; the library
	   holds it as a literal because pen discards a size bound to a variable. */
	.bar {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		block-size: 56px;
		padding-inline: var(--space-4);
		background: var(--color-bg-surface);
		border-block-end: 1px solid var(--color-border-default);
	}

	.bar-spacer {
		flex: 1;
	}

	.avatar {
		display: grid;
		place-items: center;
		inline-size: 30px;
		block-size: 30px;
		border-radius: var(--radius-full);
		background: var(--color-bg-inverse);
		color: var(--color-text-inverse);
		font-size: 11px;
		font-weight: 600;
	}

	.sidebar {
		display: none;
	}

	/* The sidebar is chrome, not content. Sticky and exactly one viewport tall,
	   so it stays put while the main column scrolls — as a flex sibling it
	   stretched to the tallest child instead, growing with a long ledger and
	   scrolling away with it. */

	/* Below the body in the same column, so it is a sibling rather than an
	   overlay. It never hides on scroll. */
	.bottom {
		position: sticky;
		inset-block-end: 0;
	}

	/* Slot `xNVFG`: fill, vertical, padding $space-8, centred on both axes.
	   The centring is inert for a screen that fills the slot — which every
	   screen of content does, absorbing the slack with a trailing spacer the
	   way the frames draw it. It is what centres a short screen, such as the
	   empty state the library's own specimen shows.

	   The padding was --space-4 and the slot says --space-8, which is why every
	   screen sat closer to the chrome than it was drawn. */
	.body {
		display: flex;
		flex: 1;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		padding: var(--space-8);
	}

	@media (min-width: 1024px) {
		.shell {
			flex-direction: row;
		}

		/* The sidebar carries the brand and the account at this regime, so the
		   bar would be a second copy of both. */
		.bar {
			display: none;
		}

		.sidebar {
			display: flex;
			position: sticky;
			inset-block-start: 0;
			align-self: start;
			block-size: 100svh;
		}

		/* The sidebar carries the same destinations at this regime. */
		.bottom {
			display: none;
		}

		.body {
			padding: var(--space-8);
		}
	}
</style>
