<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Origin `w5ZouR`. Says why a region is empty rather than leaving it blank.
	 *
	 * The library's fill is `bg-subtle`; both landing shells override it to
	 * `bg-elevated` with a border, because content sits one step above the
	 * chrome (ADR 0014). That override is the `elevated` prop.
	 */
	interface Props {
		title: string;
		icon?: Snippet;
		children?: Snippet;
		action?: Snippet;
		elevated?: boolean;
	}

	let { title, icon, children, action, elevated = false }: Props = $props();
</script>

<div class="empty" class:elevated>
	{#if icon}<span class="icon" aria-hidden="true">{@render icon()}</span>{/if}
	<p class="title">{title}</p>
	{#if children}<p class="message">{@render children()}</p>{/if}
	{#if action}<div class="action">{@render action()}</div>{/if}
</div>

<style>
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 7px;
		inline-size: 100%;
		padding: 30px 22px;
		border: 1px solid transparent;
		border-radius: var(--radius-lg);
		background: var(--color-bg-subtle);
		font-family: var(--type-family-body);
	}

	.elevated {
		background: var(--color-bg-elevated);
		border-color: var(--color-border-default);
	}

	.icon {
		display: grid;
		place-items: center;
		inline-size: 24px;
		block-size: 24px;
		color: var(--color-text-secondary);
	}

	.title {
		margin: 0;
		color: var(--color-text-primary);
		font-size: 14px;
		font-weight: 600;
		text-align: center;
	}

	.message {
		margin: 0;
		color: var(--color-text-secondary);
		font-size: 12px;
		line-height: 1.45;
		text-align: center;
	}

	.action {
		margin-block-start: var(--space-1);
	}
</style>
