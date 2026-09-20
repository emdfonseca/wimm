<script lang="ts">
	/**
	 * Origin `ui:K3w7Lx`, Trend sparkline. A row of bars, not a line — pen.dev's
	 * layout system positions flex children, not arbitrary points, and a
	 * dependency-free line chart needs a coordinate system this doesn't have
	 * (design.md). Height is proportional to value; there is no axis.
	 */
	interface Props {
		label?: string;
		/** One point per period, oldest first. Never negative — a balance trend,
		 *  not a change. */
		values: number[];
		/** Already formatted, e.g. "90 days to 19 Sep". */
		caption: string;
	}

	let { label = 'Trend', values, caption }: Props = $props();

	const max = $derived(Math.max(1, ...values));
</script>

<div class="sparkline">
	<span class="label">{label}</span>
	<div class="bars">
		{#each values as value, i (i)}
			<span class="bar" style:height="{Math.max(2, (value / max) * 48)}px"></span>
		{/each}
	</div>
	<span class="caption">{caption}</span>
</div>

<style>
	.sparkline {
		display: flex;
		flex-direction: column;
		gap: 8px;
		inline-size: 100%;
	}

	.label {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 12px;
		font-weight: 500;
	}

	.bars {
		display: flex;
		align-items: flex-end;
		gap: 3px;
		block-size: 48px;
		inline-size: 100%;
	}

	.bar {
		flex: 1 1 auto;
		min-inline-size: 0;
		background: var(--color-chart-1);
		border-radius: 2px;
	}

	.caption {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
	}
</style>
