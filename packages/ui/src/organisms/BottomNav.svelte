<script lang="ts">
	import Icon from '../atoms/Icon.svelte';
	import type { Destination } from '../destinations.js';

	/**
	 * Mirrors `Bottom nav` (`b3MTx`). Primary navigation where there is no
	 * column for a sidebar.
	 *
	 * Not a hamburger. ADR 0005 chose the form of navigation from the content
	 * and said width plays no part in it: a fixed handful of top-level areas is
	 * persistent at every size, a sidebar where there is a column and a bar
	 * where there is not. Overview answers how much and Transactions answers
	 * what happened, and a member moves between them constantly.
	 *
	 * It is a sibling below the scrolling region rather than an overlay, so
	 * nothing is obscured and the list needs no bottom padding to clear it. It
	 * never hides on scroll: navigation that comes and goes re-introduces the
	 * hamburger's problem intermittently, which is worse than having it always.
	 *
	 * Current is marked three ways and only one of them is colour — an accent
	 * top edge, an accent icon and an accent label at 600 — plus
	 * `aria-current="page"`.
	 */
	interface Props {
		destinations: Destination[];
	}

	let { destinations }: Props = $props();

	const shown = $derived(destinations.filter((d) => d.href));
</script>

<nav class="bottom-nav" aria-label="Sections">
	{#each shown as destination (destination.label)}
		<a
			class="tab"
			class:current={destination.current}
			href={destination.href}
			aria-current={destination.current ? 'page' : undefined}
		>
			<span class="edge" aria-hidden="true"></span>
			<span class="icon"><Icon name={destination.icon ?? 'layout-dashboard'} size={20} /></span>
			<span class="label">{destination.label}</span>
		</a>
	{/each}
</nav>

<style>
	.bottom-nav {
		display: flex;
		align-items: stretch;
		inline-size: 100%;
		block-size: 60px;
		background: var(--color-bg-surface);
		border-block-start: 1px solid var(--color-border-default);
		font-family: var(--type-family-body);
	}

	.tab {
		display: flex;
		flex: 1 1 0;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 3px;
		/* The target is the whole tab: at 390 that is 195 x 60, well clear of
		   any target-size floor. */
		min-block-size: 44px;
		padding-inline: 4px;
		color: var(--color-text-secondary);
		text-decoration: none;
	}

	.tab:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: calc(var(--focus-ring-offset) * -1);
	}

	/* Always present so marking current changes a colour rather than the
	   layout: an edge that appears would shift the label by two pixels. */
	.edge {
		inline-size: 100%;
		block-size: 2px;
		margin-block-start: -2px;
		background: transparent;
	}

	.icon {
		display: grid;
		place-items: center;
		color: inherit;
	}

	.label {
		font-size: 11px;
		font-weight: normal;
		text-align: center;
	}

	.current {
		color: var(--color-accent);
	}

	.current .edge {
		background: var(--color-accent);
	}

	.current .label {
		font-weight: 600;
	}
</style>
