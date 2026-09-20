<script lang="ts" module>
	/**
	 * Origin `ui:oweC0`, Metric tile. A labelled figure — household money on
	 * Overview, or any other single number a screen wants to lead with.
	 */
	export interface Delta {
		/** Already formatted, in words or carrying its own sign: `€150.00 more`.
		 *  Direction is never carried by colour, so the words are the signal. */
		change: string;
		/** e.g. `than 1 to 20 Aug`. */
		period: string;
		direction: 'up' | 'down';
	}
</script>

<script lang="ts">
	import Icon from '../atoms/Icon.svelte';

	interface Props {
		label: string;
		/** Already formatted, carrying its currency. */
		value: string;
		delta?: Delta;
		/** Makes the whole tile a link. */
		href?: string;
	}

	let { label, value, delta, href }: Props = $props();
</script>

<svelte:element this={href ? 'a' : 'div'} class="tile" class:link={!!href} {href}>
	<span class="label">{label}</span>
	<span class="value">{value}</span>
	{#if delta}
		<span class="delta">
			<Icon name={delta.direction === 'up' ? 'arrow-up-right' : 'arrow-down-right'} size={13} />
			<span class="change">{delta.change}</span>
			<span class="period">{delta.period}</span>
		</span>
	{/if}
</svelte:element>

<style>
	.tile {
		display: flex;
		flex-direction: column;
		gap: 5px;
		inline-size: 100%;
		block-size: 100%;
		padding: 16px;
		background: var(--color-bg-surface);
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-lg);
	}

	.link {
		color: inherit;
		text-decoration: none;
	}

	.link:hover {
		background: var(--color-bg-elevated);
	}

	.link:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
	}

	.label {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 12px;
		font-weight: 500;
	}

	.value {
		color: var(--color-text-primary);
		font-family: var(--type-family-mono);
		font-size: var(--type-size-heading-lg);
		line-height: 1.2;
	}

	.delta {
		display: flex;
		align-items: center;
		gap: 5px;
		/* Green and red mean money direction only (ADR 0002): "€182.40 more"
		   is not a direction of money. */
		color: var(--color-text-secondary);
	}

	.change {
		font-family: var(--type-family-mono);
		font-size: 12px;
	}

	.period {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
	}
</style>
