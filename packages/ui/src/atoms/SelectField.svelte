<script lang="ts">
	import Icon from './Icon.svelte';

	/**
	 * A labelled native select in the control chrome. Native, because a list of
	 * a household's accounts or fourteen months is what the platform's own
	 * picker already does well on every pointer, and a hand-built listbox would
	 * be a second one to keep accessible.
	 */
	interface Option {
		value: string;
		label: string;
	}

	interface Props {
		label: string;
		options: Option[];
		value?: string;
		/** The field's name when it submits with a form. */
		name?: string;
		/** The label is read, not shown: the chosen option says what the
		 *  select is, as `All accounts` or `August 2026` does. */
		hideLabel?: boolean;
		onchange?: (value: string) => void;
	}

	let { label, options, value = '', name, hideLabel = false, onchange }: Props = $props();
</script>

<label class="field">
	<span class="label" class:hidden={hideLabel}>{label}</span>
	<span class="box">
		<select {name} {value} onchange={(event) => onchange?.(event.currentTarget.value)}>
			{#each options as option (option.value)}
				<option value={option.value} selected={option.value === value}>{option.label}</option>
			{/each}
		</select>
		<span class="chevron"><Icon name="chevron-down" size={14} /></span>
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

	.hidden {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	.box {
		position: relative;
		display: flex;
		align-items: center;
	}

	select {
		appearance: none;
		inline-size: 100%;
		block-size: var(--control-height-sm);
		padding-inline: 12px 32px;
		border: 1px solid var(--color-control-border);
		border-radius: var(--radius-control);
		background: var(--color-bg-surface);
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		text-overflow: ellipsis;
		cursor: pointer;
	}

	select:hover {
		border-color: var(--color-control-border-hover);
	}

	select:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.chevron {
		position: absolute;
		inset-inline-end: 10px;
		display: grid;
		color: var(--color-text-secondary);
		pointer-events: none;
	}

	@media (any-pointer: coarse) {
		select {
			block-size: var(--control-height-md);
		}
	}
</style>
