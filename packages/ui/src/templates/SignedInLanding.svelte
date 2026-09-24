<script lang="ts">
	import type { Snippet } from 'svelte';
	import SidebarNav from '../organisms/SidebarNav.svelte';
	import BottomNav from '../organisms/BottomNav.svelte';
	import Brand from '../atoms/Brand.svelte';
	import Icon from '../atoms/Icon.svelte';
	import IconButton from '../atoms/IconButton.svelte';
	import { destinations as allDestinations, type Destination } from '../destinations.js';

	/**
	 * Where a member arrives once they are in. Origins `IGbQe` (Wide),
	 * `pu6qZ` (Compact), `WMlvF` (Medium, the rail) and `NEIet` (Ultra).
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
	 *
	 * At Medium the sidebar is a 72px rail: icon-only, the same destinations
	 * the labelled sidebar reaches, because labels do not fit beside a content
	 * column at this width and a bottom bar would waste a row of it. At Wide
	 * and Ultra it is the labelled sidebar, `layout-sidebar-width` giving it
	 * 264 or 288 — the one number that differs between the two.
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

	const shown = $derived(destinations.filter((d) => d.href));
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

	<nav class="rail" aria-label="Sections">
		<div class="rail-brand"><Brand markOnly /></div>
		<ul class="rail-items">
			{#each shown as destination (destination.label)}
				<li>
					<a
						class="rail-item"
						class:current={destination.current}
						href={destination.href}
						aria-current={destination.current ? 'page' : undefined}
						aria-label={destination.label}
					>
						<Icon name={destination.icon ?? 'layout-dashboard'} size={20} />
					</a>
				</li>
			{/each}
		</ul>
		<span class="rail-spacer"></span>
		{#if onsignout}
			<IconButton icon="log-out" label="Sign out" onclick={onsignout} />
		{/if}
	</nav>

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
		block-size: 100svh;
		overflow: hidden;
		background: var(--color-bg-canvas);
		font-family: var(--type-family-body);
		color: var(--color-text-primary);
	}

	/* The Compact bar, `layout-header-height` at this regime. */
	.bar {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		block-size: var(--layout-header-height);
		padding-inline: var(--layout-page-gutter);
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

	.rail,
	.sidebar {
		display: none;
	}

	/* The sidebar and the rail are chrome, not content. Sticky and exactly one
	   viewport tall, so either stays put while the main column scrolls — as a
	   flex sibling it stretched to the tallest child instead, growing with a
	   long ledger and scrolling away with it. */

	/* Below the body in the same column, so it is a sibling rather than an
	   overlay. It never hides on scroll. */
	.bottom {
		position: sticky;
		inset-block-end: 0;
	}

	/* Slot `xNVFG`: fill, vertical, centred on both axes. The centring is
	   inert for a screen that fills the slot — which every screen of content
	   does, absorbing the slack with a trailing spacer the way the frames draw
	   it. It is what centres a short screen, such as the empty state the
	   library's own specimen shows.

	   No padding here: the slot's child is always a `Page`, whose own header
	   spans edge to edge of the main column and owns `layout-page-gutter`
	   itself, for its content only. Padding on this wrapper insets the header
	   too, on every regime, which is not what any frame draws. */
	.body {
		display: flex;
		flex: 1;
		flex-direction: column;
		/* `safe`: a page taller than this box, centred, overflows above the
		   scroll origin too, where no scrolling can reach its top. */
		justify-content: safe center;
		align-items: safe center;
		min-block-size: 0;
		overflow-y: auto;
		/* The containing block for anything absolutely positioned in a page —
		   every visually hidden label is. Without it they are placed against the
		   document, deep in a long page, and the whole document scrolls. */
		position: relative;
	}

	/* Medium: the rail. Labels do not fit beside a content column at this
	   width and a bottom bar would waste a row of it (`WMlvF`). */
	@media (min-width: 768px) and (max-width: 1199px) {
		.shell {
			flex-direction: row;
		}
		.bar {
			display: none;
		}
		.rail {
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 2px;
			position: sticky;
			inset-block-start: 0;
			align-self: start;
			block-size: 100svh;
			inline-size: var(--layout-sidebar-width);
			padding-block: 12px;
			background: var(--color-bg-surface);
			border-inline-end: 1px solid var(--color-border-default);
		}
		.bottom {
			display: none;
		}
	}

	/* Wide and Ultra: the labelled sidebar. `layout-sidebar-width` is the one
	   number that differs between them — 264 or 288 (`IGbQe`, `NEIet`). */
	@media (min-width: 1200px) {
		.shell {
			flex-direction: row;
		}
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
		.bottom {
			display: none;
		}
	}

	.rail-brand {
		display: grid;
		place-items: center;
		block-size: 36px;
		margin-block-end: 8px;
	}

	.rail-items {
		display: flex;
		flex-direction: column;
		gap: 2px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.rail-item {
		display: grid;
		place-items: center;
		inline-size: 44px;
		block-size: 36px;
		border-radius: var(--radius-sm);
		color: var(--color-text-secondary);
		text-decoration: none;
	}

	.rail-item:hover {
		background: var(--color-bg-hover);
	}

	.rail-item:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.rail-item.current {
		background: var(--color-accent-subtle);
		color: var(--color-accent);
	}

	.rail-spacer {
		flex: 1;
	}
</style>
