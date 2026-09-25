<script lang="ts">
	import Icon from './Icon.svelte';

	/**
	 * A labelled search box with a button that empties it. The search applies
	 * as the member types, once they pause, and at once on Enter or clearing.
	 * Each application is reported with whether it came from typing, so the
	 * caller can keep half-typed words out of the history.
	 *
	 * The browser's own clear control is hidden, because a second one would be
	 * a second way to do one thing that reports nothing.
	 */
	interface Props {
		label: string;
		/** The search in force. */
		value?: string;
		placeholder?: string;
		/** The field's name when it submits with a form. */
		name?: string;
		/** The label is read, not shown: the placeholder and the glyph say
		 *  what the box is where space is short. */
		hideLabel?: boolean;
		/** The text, trimmed, whenever it changes what is searched. `live` is
		 *  true when it came from a pause in typing. Clearing sends `''`. */
		onsearch?: (text: string, how: { live: boolean }) => void;
	}

	let { label, value = '', placeholder, name, hideLabel = false, onsearch }: Props = $props();

	const MAX_LENGTH = 100;
	/** Long enough to skip the letters of a word, short enough to feel live. */
	const PAUSE_MS = 300;

	// What the member typed, until the search in force changes from outside —
	// Back, or Clear filters. A search this box sent coming back as the new
	// value is not a change from outside, and must not overwrite letters typed
	// while it was on its way.
	let draft = $state<string | null>(null);
	let sent = '';
	const text = $derived(draft ?? value);

	$effect(() => {
		const inForce = value.trim();
		if (inForce !== sent) {
			draft = null;
			sent = inForce;
		}
	});

	let input = $state<HTMLInputElement | null>(null);
	let timer: ReturnType<typeof setTimeout> | undefined;

	$effect(() => () => clearTimeout(timer));

	function report(live: boolean) {
		clearTimeout(timer);
		const trimmed = text.trim();
		if (trimmed === sent) return;
		sent = trimmed;
		onsearch?.(trimmed, { live });
	}

	function oninput(event: Event & { currentTarget: HTMLInputElement }) {
		draft = event.currentTarget.value;
		clearTimeout(timer);
		timer = setTimeout(() => report(true), PAUSE_MS);
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key !== 'Enter' || !onsearch) return;
		event.preventDefault();
		report(false);
	}

	function clear() {
		draft = '';
		input?.focus();
		report(false);
	}
</script>

<label class="field">
	<span class="label" class:hidden={hideLabel}>{label}</span>
	<span class="box">
		<span class="glyph"><Icon name="search" size={14} /></span>
		<input
			bind:this={input}
			value={text}
			type="search"
			{name}
			{placeholder}
			maxlength={MAX_LENGTH}
			autocomplete="off"
			{oninput}
			{onkeydown}
		/>
		{#if text !== ''}
			<button
				type="button"
				class="clear"
				aria-label="Clear search"
				title="Clear search"
				onclick={clear}
			>
				<Icon name="x" size={14} />
			</button>
		{/if}
	</span>
</label>

<style>
	.field {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-inline-size: 0;
	}

	.label {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	.box {
		position: relative;
		display: flex;
		align-items: center;
	}

	input {
		inline-size: 100%;
		block-size: var(--control-height-sm);
		padding-inline: 32px 36px;
		border: 1px solid var(--color-control-border);
		border-radius: var(--radius-control);
		background: var(--color-bg-surface);
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.glyph {
		position: absolute;
		inset-inline-start: 11px;
		display: grid;
		color: var(--color-text-secondary);
		pointer-events: none;
	}

	.hidden {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	input::placeholder {
		color: var(--color-text-secondary);
	}

	input::-webkit-search-cancel-button {
		appearance: none;
	}

	input:hover {
		border-color: var(--color-control-border-hover);
	}

	input:focus-visible,
	.clear:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.clear {
		position: absolute;
		inset-inline-end: 4px;
		display: grid;
		place-items: center;
		inline-size: 28px;
		block-size: 28px;
		padding: 0;
		border: 0;
		border-radius: var(--radius-sm);
		background: none;
		color: var(--color-text-secondary);
		cursor: pointer;
	}

	.clear:hover {
		background: var(--color-bg-hover);
		color: var(--color-text-primary);
	}

	@media (any-pointer: coarse) {
		input {
			block-size: var(--control-height-md);
		}
	}
</style>
