<script lang="ts">
	import Icon from './Icon.svelte';

	/**
	 * Origin `mRsB3`. A 36 square carrying a 16 glyph.
	 *
	 * In a shell footer or bar it is drawn with no fill and no border, the way
	 * the library's own Collapse control is: the chrome would compete with the
	 * account beside it.
	 *
	 * `label` is not optional. An icon with no accessible name is a button that
	 * announces itself as "button", which is the commonest way an icon-only
	 * control becomes unusable.
	 */
	interface Props {
		icon: 'log-out' | 'panel-left-close';
		label: string;
		/** Draws the component's own fill and border, as on a specimen tile. */
		chrome?: boolean;
		onclick?: () => void;
	}

	let { icon, label, chrome = false, onclick }: Props = $props();
</script>

<button class="icon-button" class:chrome type="button" aria-label={label} title={label} {onclick}>
	<Icon name={icon} size={16} />
</button>

<style>
	.icon-button {
		display: grid;
		place-items: center;
		flex: none;
		inline-size: 36px;
		block-size: 36px;
		padding: 0;
		border: 1px solid transparent;
		border-radius: var(--radius-control);
		background: none;
		color: var(--color-text-secondary);
		cursor: pointer;
	}

	.chrome {
		background: var(--color-bg-surface);
		border-color: var(--color-border-default);
		color: var(--color-text-primary);
	}

	.icon-button:hover {
		background: var(--color-bg-hover);
	}

	.icon-button:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}
</style>
