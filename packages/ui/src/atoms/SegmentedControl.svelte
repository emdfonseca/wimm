<script lang="ts">
	/**
	 * Origin `e0JfP`. Height 32, three-pixel inset, `radius-control`, segments
	 * splitting the width evenly.
	 *
	 * One choice among three, as **one control**: a single tab stop, arrow keys
	 * between the options, and exactly one selection at all times. Three
	 * checkboxes would permit zero or two answers to a question that has one;
	 * three buttons would not announce themselves as a single control to a
	 * screen reader. That is why this exists rather than composing what the
	 * library already had.
	 *
	 * Implemented as a radiogroup with roving tabindex, which is the pattern
	 * that gives one tab stop and arrow-key movement without any of it being
	 * hand-rolled onto generic buttons.
	 */
	interface Option<T extends string> {
		value: T;
		label: string;
	}

	interface Props<T extends string> {
		options: Option<T>[];
		value: T;
		/** Names the group for assistive technology. Required: a bare group of
		 *  three words is not self-describing. */
		label: string;
		disabled?: boolean;
		onchange?: (value: T) => void;
	}

	let { options, value = $bindable(), label, disabled = false, onchange }: Props<string> = $props();

	let buttons: HTMLButtonElement[] = $state([]);

	function select(next: string) {
		if (disabled || next === value) return;
		value = next;
		onchange?.(next);
	}

	/**
	 * Arrow keys move the selection, which is what a radiogroup does — moving
	 * focus without selecting would leave the control showing one thing and
	 * about to do another.
	 */
	function onkeydown(event: KeyboardEvent) {
		if (disabled) return;

		const index = options.findIndex((o) => o.value === value);
		let next = index;

		switch (event.key) {
			case 'ArrowLeft':
			case 'ArrowUp':
				next = (index - 1 + options.length) % options.length;
				break;
			case 'ArrowRight':
			case 'ArrowDown':
				next = (index + 1) % options.length;
				break;
			case 'Home':
				next = 0;
				break;
			case 'End':
				next = options.length - 1;
				break;
			default:
				return;
		}

		const moved = options[next];
		if (!moved) return;

		event.preventDefault();
		select(moved.value);
		buttons[next]?.focus();
	}
</script>

<!-- tabindex is -1, not absent: the radios carry the roving tabindex and the
     group is never a Tab stop, but `radiogroup` is on the a11y linter's
     interactive-role list and -1 is the value that means "reachable in code
     only". -->
<div
	class="segmented"
	class:disabled
	role="radiogroup"
	aria-label={label}
	tabindex={-1}
	aria-disabled={disabled || undefined}
	{onkeydown}
>
	{#each options as option, index (option.value)}
		{@const selected = option.value === value}
		<button
			bind:this={buttons[index]}
			type="button"
			role="radio"
			class="segment"
			class:selected
			aria-checked={selected}
			tabindex={selected ? 0 : -1}
			{disabled}
			onclick={() => select(option.value)}
		>
			{option.label}
		</button>
	{/each}
</div>

<style>
	.segmented {
		display: inline-flex;
		align-items: center;
		gap: 3px;
		padding: 3px;
		block-size: 32px;
		background: var(--color-bg-subtle);
		border-radius: var(--radius-control);
	}

	/* The ring goes around the whole control, because the control is what has
	   focus — a ring around one segment would say the segment is the thing
	   being moved between. */
	.segmented:has(.segment:focus-visible) {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.segment {
		flex: 1 1 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		block-size: 100%;
		padding-inline: 12px;
		border: 1px solid transparent;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: 500;
		line-height: var(--type-line-tight);
		white-space: nowrap;
		cursor: pointer;
		transition:
			background-color var(--motion-duration-fast) var(--motion-ease-standard),
			color var(--motion-duration-fast) var(--motion-ease-standard);
	}

	.segment:focus-visible {
		outline: none;
	}

	.segment:hover:not(.selected):not(:disabled) {
		color: var(--color-text-primary);
	}

	.selected {
		background: var(--color-bg-surface);
		border-color: var(--color-control-border);
		color: var(--color-text-primary);
	}

	.disabled {
		background: var(--color-bg-disabled);
	}

	.segment:disabled {
		color: var(--color-text-disabled);
		cursor: not-allowed;
	}

	.disabled .selected {
		background: var(--color-bg-disabled);
		border-color: var(--color-border-disabled);
	}

	@media (any-pointer: coarse) {
		.segmented {
			block-size: 40px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.segment {
			transition: none;
		}
	}
</style>
