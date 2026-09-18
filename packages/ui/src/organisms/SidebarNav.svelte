<script lang="ts">
	import Brand from '../atoms/Brand.svelte';
	import Icon from '../atoms/Icon.svelte';
	import IconButton from '../atoms/IconButton.svelte';
	import type { Destination } from '../destinations.js';

	/**
	 * `Sidebar nav / in shell`, origin `q7dY7c`: a preset wrapping `fwKnX` that
	 * trades its card chrome for the single edge a shell part needs — a right
	 * border and no corner radius.
	 *
	 * **Destinations that do not exist are not rendered.** The canvas disables
	 * them, and a disabled node in pen is absent, not greyed. Showing six
	 * greyed rows would draw a product nobody agreed to build; the canvas
	 * instruction says so directly.
	 */
	interface Props {
		destinations: Destination[];
		/** The signed-in member, in the footer. */
		memberName: string;
		memberInitials?: string;
		onsignout?: () => void;
	}

	let { destinations, memberName, memberInitials, onsignout }: Props = $props();

	const shown = $derived(destinations.filter((d) => d.href));

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

<nav class="sidebar" aria-label="Sections">
	<div class="brand"><Brand /></div>

	<ul class="items">
		{#each shown as destination (destination.label)}
			<li>
				<a
					class="item"
					class:current={destination.current}
					href={destination.href}
					aria-current={destination.current ? 'page' : undefined}
				>
					<span class="item-icon">
						<Icon name={destination.icon ?? 'layout-dashboard'} size={16} />
					</span>
					<span class="item-label">{destination.label}</span>
				</a>
			</li>
		{/each}
	</ul>

	<div class="footer">
		<span class="avatar" aria-hidden="true">{initials}</span>
		<span class="identity">
			<span class="name">{memberName}</span>
			<span class="meta">Signed in</span>
		</span>
		{#if onsignout}
			<IconButton icon="log-out" label="Sign out" onclick={onsignout} />
		{/if}
	</div>
</nav>

<style>
	.sidebar {
		display: flex;
		flex-direction: column;
		gap: 2px;
		inline-size: var(--layout-sidebar-width);
		/* Fills whatever the shell gives it, which at Wide is one viewport. */
		block-size: 100%;
		padding: 12px;
		background: var(--color-bg-surface);
		/* The preset's whole purpose: one edge, no radius. */
		border-inline-end: 1px solid var(--color-border-default);
		border-radius: 0;
		font-family: var(--type-family-body);
	}

	.brand {
		display: flex;
		align-items: center;
		block-size: 44px;
		padding-inline: 10px;
	}

	.items {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 1;
		min-block-size: 0;
		/* Scrolls within the column rather than pushing the footer off the
		   bottom, for the day this list is longer than a short screen. */
		overflow-y: auto;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.item {
		display: flex;
		align-items: center;
		gap: 10px;
		block-size: 36px;
		padding-inline: 10px;
		border-radius: var(--radius-sm);
		color: var(--color-text-primary);
		font-size: 13px;
		text-decoration: none;
	}

	.item-icon {
		display: grid;
		place-items: center;
		color: var(--color-text-secondary);
	}

	.item:hover {
		background: var(--color-bg-hover);
	}

	.item:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	/* The pattern the library already uses to mark where you are. A page whose
	   nav item is not marked is a page that does not say where you are. */
	.current {
		background: var(--color-accent-subtle);
		color: var(--color-accent);
		font-weight: 600;
	}

	.current .item-icon {
		color: var(--color-accent);
	}

	.footer {
		display: flex;
		align-items: center;
		gap: 10px;
		block-size: 48px;
		padding-inline: 8px;
	}

	.avatar {
		flex: none;
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

	.identity {
		display: flex;
		flex-direction: column;
		gap: 1px;
		flex: 1;
		min-inline-size: 0;
	}

	.name {
		color: var(--color-text-primary);
		font-size: 12px;
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.meta {
		color: var(--color-text-secondary);
		font-size: 10px;
	}
</style>
