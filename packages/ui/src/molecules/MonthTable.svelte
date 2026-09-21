<script lang="ts" module>
	export interface MonthTableRow {
		key: string;
		/** `March 2026`. */
		label: string;
		/** `So far`, `Held from 12 Apr`. Absent for a full month. */
		state?: string;
		/** Already formatted. Hidden at Compact. */
		moneyIn?: string;
		moneyOut?: string;
		/** Already formatted, carrying its sign. */
		net: string;
		/** `2 unusual · +€379.70 without them`. Absent when none. */
		unusualLine?: string;
	}
</script>

<script lang="ts">
	/**
	 * Every month, always visible, one button row each. Choosing a row selects
	 * that month and the detail panel it names shows it; nothing opens or
	 * closes. Which month is selected is the screen's, so `selected` is
	 * bindable and shared with the chart's marker.
	 *
	 * This formats nothing: every figure and sentence arrives written.
	 */
	interface Props {
		/** What the list is called to a screen reader. */
		label: string;
		/** Newest first. */
		rows: MonthTableRow[];
		selected?: string | null;
		/** The id of the detail panel the rows control. */
		controls?: string;
		compact?: boolean;
	}

	let { label, rows, selected = $bindable(null), controls, compact = false }: Props = $props();
</script>

<div class="table" class:compact role="group" aria-label={label}>
	<div class="header" aria-hidden="true">
		<span class="name">Month</span>
		<span class="cell wide">In</span>
		<span class="cell wide">Out</span>
		<span class="cell">Net</span>
	</div>
	{#each rows as row (row.key)}
		<button
			type="button"
			class="row"
			class:selected={row.key === selected}
			aria-pressed={row.key === selected}
			aria-controls={controls}
			onclick={() => (selected = row.key)}
		>
			<span class="name">
				<span class="line">
					<span class="month">{row.label}</span>
					{#if row.state}<span class="state">{row.state}</span>{/if}
				</span>
				{#if row.unusualLine}<span class="unusual">{row.unusualLine}</span>{/if}
			</span>
			{#if row.moneyIn}<span class="cell wide in"><span class="sr">Money in: </span>{row.moneyIn}</span>{/if}
			{#if row.moneyOut}<span class="cell wide out"><span class="sr">Money out: </span>{row.moneyOut}</span>{/if}
			<span class="cell net"><span class="sr">Net: </span>{row.net}</span>
		</button>
	{/each}
</div>

<style>
	.table {
		display: flex;
		flex-direction: column;
		inline-size: 100%;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	.header,
	.row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 112px 112px 112px;
		align-items: center;
		column-gap: 12px;
		padding-inline: var(--density-cell-padding-x);
	}

	.header {
		block-size: 32px;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
		border-block-end: 1px solid var(--color-border-subtle);
	}

	.row {
		inline-size: 100%;
		block-size: 56px;
		border: 0;
		border-block-end: 1px solid var(--color-border-subtle);
		background: transparent;
		color: inherit;
		font: inherit;
		text-align: start;
		cursor: pointer;
	}

	.compact .row {
		block-size: 64px;
	}

	.row:last-child {
		border-block-end: 0;
	}

	.row:hover {
		background: var(--color-bg-hover);
	}

	.row.selected {
		background: var(--color-bg-active);
	}

	.row:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: calc(-1 * var(--focus-ring-width));
	}

	.sr {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	.name {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-inline-size: 0;
	}

	.header .name {
		display: block;
	}

	.line {
		display: flex;
		align-items: baseline;
		gap: 8px;
	}

	.month {
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 13px;
		font-weight: 500;
	}

	.state,
	.unusual {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
	}

	.cell {
		text-align: end;
		color: var(--color-text-primary);
		font-family: var(--type-family-mono);
		font-size: 13px;
	}

	.header .cell {
		font-family: var(--type-family-body);
		font-size: 11px;
		color: var(--color-text-secondary);
	}

	/* Money in and out need room a phone does not have: the net and the unusual
	   line are what a month is read for there. */
	@media (max-width: 767px) {
		.header,
		.row {
			grid-template-columns: minmax(0, 1fr) auto;
		}

		.wide {
			display: none;
		}
	}
</style>
