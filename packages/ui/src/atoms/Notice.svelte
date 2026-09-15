<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Origin `WFSg5`. A 3 px left edge rather than a full border, a
	 * `feedback-*-bg` fill, and a 16 px icon beside a title and a message.
	 *
	 * Notice carries no tone of its own: pen has no variant-property mechanism,
	 * so every tone in the library is an instance override. In Svelte that
	 * becomes a prop, and the tonal presets are what fix the token quadruple, so
	 * a feedback colour is chosen in one place.
	 *
	 * `live` is the half the canvas could not draw. An error the member has to
	 * act on interrupts; "nothing happened, try again" waits. Neither moves
	 * focus: the member is returning from a native prompt and a focus jump loses
	 * their place.
	 */
	interface Props {
		title?: string;
		live?: 'assertive' | 'polite' | 'off';
		children?: Snippet;
	}

	let { title, live = 'polite', children }: Props = $props();
</script>

<div
	class="notice"
	role={live === 'assertive' ? 'alert' : 'status'}
	aria-live={live === 'off' ? undefined : live}
>
	<span class="icon" aria-hidden="true">
		<svg viewBox="0 0 16 16" width="16" height="16" focusable="false">
			<circle cx="8" cy="8" r="7" fill="none" stroke="currentColor" stroke-width="1.5" />
			<path d="M8 4.5v4.2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
			<circle cx="8" cy="11.4" r="0.9" fill="currentColor" />
		</svg>
	</span>

	<div class="body">
		{#if title}<p class="title">{title}</p>{/if}
		{#if children}<p class="message">{@render children()}</p>{/if}
	</div>
</div>

<style>
	.notice {
		display: flex;
		align-items: start;
		gap: 11px;
		inline-size: 100%;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		/* The tone supplies all three. Notice has no colour of its own. */
		background: var(--notice-bg, var(--color-feedback-warning-bg));
		color: var(--notice-accent, var(--color-feedback-warning));
		/* A left edge, not a border: the library draws `strokeWidth {left: 3}`. */
		border-inline-start: 3px solid currentColor;
		font-family: var(--type-family-body);
	}

	.icon {
		flex: none;
		display: grid;
		place-items: center;
		block-size: 16px;
	}

	.body {
		display: flex;
		flex-direction: column;
		gap: 3px;
		min-inline-size: 0;
	}

	.title {
		margin: 0;
		color: var(--color-text-primary);
		font-size: 13px;
		font-weight: 600;
	}

	.message {
		margin: 0;
		color: var(--color-text-secondary);
		font-size: 12px;
		line-height: 1.45;
	}
</style>
