<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';

	/**
	 * Origin `W2gOKx`. Height 36, 18 of horizontal padding, `radius-control`,
	 * label at `type-size-body-md` weight 500.
	 *
	 * The library's contract: pointer regimes resolve 30 / 36 / 44, and Compact
	 * resolves 36 / 44 / 52, because a 36 px control is a comfortable mouse
	 * target and a poor thumb target. Every size clears the 24 px AA target
	 * floor of SC 2.5.8.
	 *
	 * Focus is additive: the ring is drawn outside the control, which keeps its
	 * own fill underneath in every state including disabled.
	 */
	interface Props extends HTMLButtonAttributes {
		variant?: 'primary' | 'secondary' | 'ghost' | 'destructive';
		size?: 'sm' | 'md' | 'lg';
		/** Takes the full width of its container, as it does inside an auth card. */
		block?: boolean;
		children: Snippet;
	}

	let {
		variant = 'primary',
		size = 'md',
		block = false,
		type = 'button',
		children,
		...rest
	}: Props = $props();
</script>

<button class="button {variant} {size}" class:block {type} {...rest}>
	{@render children()}
</button>

<style>
	.button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding-inline: 18px;
		padding-block: 0;
		border: 1px solid transparent;
		border-radius: var(--radius-control);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		font-weight: 500;
		line-height: var(--type-line-tight);
		cursor: pointer;
		transition:
			background-color var(--motion-duration-fast) var(--motion-ease-standard),
			border-color var(--motion-duration-fast) var(--motion-ease-standard);
	}

	.sm {
		block-size: 30px;
	}
	.md {
		block-size: 36px;
	}
	.lg {
		block-size: 44px;
	}

	/* A 36 px control is a comfortable mouse target and a poor thumb target, so
	   the whole ladder moves up where the pointer is coarse. */
	@media (any-pointer: coarse) {
		.sm {
			block-size: 36px;
		}
		.md {
			block-size: 44px;
		}
		.lg {
			block-size: 52px;
		}
	}

	.block {
		inline-size: 100%;
	}

	.primary {
		background: var(--color-action-primary);
		color: var(--color-action-on-primary);
	}
	.primary:hover:not(:disabled) {
		background: var(--color-action-primary-hover);
	}
	.primary:active:not(:disabled) {
		background: var(--color-action-primary-active);
	}

	.secondary {
		background: var(--color-action-secondary);
		border-color: var(--color-control-border);
		color: var(--color-text-primary);
	}
	.secondary:hover:not(:disabled) {
		background: var(--color-action-secondary-hover);
		border-color: var(--color-control-border-hover);
	}
	.secondary:active:not(:disabled) {
		background: var(--color-action-secondary-active);
		border-color: var(--color-control-border-hover);
	}

	.ghost {
		background: transparent;
		color: var(--color-text-primary);
	}
	.ghost:hover:not(:disabled) {
		background: var(--color-bg-hover);
	}
	.ghost:active:not(:disabled) {
		background: var(--color-bg-active);
	}

	.destructive {
		background: var(--color-action-destructive);
		color: var(--color-action-on-destructive);
	}
	.destructive:hover:not(:disabled) {
		background: var(--color-action-destructive-hover);
	}
	.destructive:active:not(:disabled) {
		background: var(--color-action-destructive-active);
	}

	.button:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	/* Disabled is the leftmost state in the precedence chain, so it owns the
	   fill and the border whatever the variant was. */
	.button:disabled,
	.button[aria-disabled='true'] {
		background: var(--color-bg-disabled);
		border-color: var(--color-border-disabled);
		color: var(--color-text-disabled);
		cursor: not-allowed;
	}

	@media (prefers-reduced-motion: reduce) {
		.button {
			transition: none;
		}
	}
</style>
