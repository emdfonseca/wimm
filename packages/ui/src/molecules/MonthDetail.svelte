<script lang="ts" module>
	import type { ComponentProps } from 'svelte';
	import LedgerRow from './LedgerRow.svelte';

	export interface MonthFigure {
		/** `Money in`, `Money out`, `Net`, `Without unusual payments`. */
		label: string;
		/** Already formatted, carrying its currency and its sign. */
		value: string;
	}

	export interface MonthRiser {
		label: string;
		/** `€246.80 · €82.40 more than usual`, or `… · not usually paid`. */
		value: string;
		/** 0 to 1 of the track, against the month's largest rise. */
		proportion: number;
	}

	export type MonthPayment = ComponentProps<typeof LedgerRow>;
</script>

<script lang="ts">
	import BudgetMeter from './BudgetMeter.svelte';

	/**
	 * The month the table has selected, in full: its figures, the merchants
	 * that rose and its unusual payments. Which month it shows is the screen's
	 * to decide. Nothing in it is operable except the payments' links.
	 *
	 * This formats nothing: every figure, date and sentence arrives written.
	 */
	interface Props {
		/** `March 2026`. */
		month: string;
		/** The element the table's rows name in `aria-controls`. */
		id?: string;
		compact?: boolean;
		/** What the panel covers, when it is not the whole month: `1 to 20 Sep`. */
		span?: string;
		figures?: MonthFigure[];
		/** The merchants that rose. */
		risers?: MonthRiser[];
		/** Said in the risers' place when the month has none to name. */
		nothingRose?: string;
		/** Unusual payments, newest first. */
		payments?: MonthPayment[];
		/** Said under the figures: `A month under way is not set against whole
		 *  months.` */
		note?: string;
		/** Said under the figures when the month's figures left transfers
		 *  between the member's own accounts out: `2 transfers between your
		 *  accounts left out · €1,400.00`. Written by the load, singular and
		 *  all. Absent, not empty, when nothing was left out. */
		leftOut?: string;
	}

	let {
		month,
		id,
		compact = false,
		span,
		figures = [],
		risers = [],
		nothingRose,
		payments = [],
		note,
		leftOut
	}: Props = $props();

</script>

<section class="panel" {id} aria-label={month}>
	<h3 class="title">{month}</h3>
	{#if span}
		<p class="span">{span}</p>
	{/if}
	{#if figures.length}
		<dl class="figures">
			{#each figures as figure (figure.label)}
				<div class="figure">
					<dt>{figure.label}</dt>
					<dd>{figure.value}</dd>
				</div>
			{/each}
		</dl>
	{/if}
	{#if note}
		<p class="note">{note}</p>
	{/if}
	{#if leftOut}
		<p class="note">{leftOut}</p>
	{/if}
	{#if risers.length}
		<section class="section">
			<h3>More than usual</h3>
			<div class="risers">
				{#each risers as riser (riser.label)}
					<BudgetMeter {...riser} />
				{/each}
			</div>
		</section>
	{:else if nothingRose}
		<p class="note">{nothingRose}</p>
	{/if}
	{#if payments.length}
		<section class="section">
			<h3>Unusual payments</h3>
			<div class="payments">
				{#each payments as payment (payment.href ?? payment.description + payment.date)}
					<LedgerRow {...payment} {compact} />
				{/each}
			</div>
		</section>
	{/if}
</section>

<style>
	.panel {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 16px var(--density-cell-padding-x);
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-md);
	}

	.title {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 14px;
		font-weight: 700;
	}

	.span,
	.note {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 12px;
	}

	.figures {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 32px;
		margin: 0;
	}

	.figure {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	dt {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
	}

	dd {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-mono);
		font-size: 13px;
	}

	.section {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	h3 {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 12px;
		font-weight: 700;
	}

	.risers {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* A ledger row carries its own inline padding; the panel already has it, so
	   the avatar would otherwise sit inset from the text above. */
	.payments {
		display: flex;
		flex-direction: column;
		margin-inline: calc(-1 * var(--density-cell-padding-x));
		--ledger-account-basis: auto;
	}
</style>
