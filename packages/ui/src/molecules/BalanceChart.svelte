<script lang="ts" module>
	export interface BalancePoint {
		/** Short: `3 Aug`. */
		date: string;
		/** What the line is drawn from. */
		value: number;
		/** Already formatted, carrying its currency. */
		amount: string;
	}
</script>

<script lang="ts">
	/**
	 * A balance over time as one line, one point a day. The line is data; the
	 * box, the stroke, the labels and where they sit are fixed. The scale is
	 * the low and high padded 8% each side and never forced to zero, since a
	 * baseline at zero is what flattened the bars this replaces.
	 *
	 * This formats nothing: every amount and date arrives already written.
	 */
	interface Props {
		/** `Balance · EUR`. */
		title: string;
		/** `22 Jun to 20 Sep`. Absent when there is no chart to span. */
		span?: string;
		/** Oldest first. */
		points?: BalancePoint[];
		/** The top and bottom of the scale column, already formatted. */
		high?: string;
		low?: string;
		/** The chart in words: the span, both ends, the extremes and when they
		 *  happened. It is the chart's accessible name. */
		summary?: string;
		/** Said under the dates when the chart leaves an account out. */
		coverage?: string;
		/** Less than a week of history anywhere: no line, and this says why. */
		shortHistory?: boolean;
		/** The index of the day pointed at, or null. */
		marked?: number | null;
	}

	let {
		title,
		span,
		points = [],
		high,
		low,
		summary,
		coverage,
		shortHistory = false,
		marked = $bindable(null)
	}: Props = $props();

	const WIDTH = 1000;
	const HEIGHT = 200;

	const last = $derived(points.length - 1);

	const geometry = $derived.by(() => {
		if (points.length < 2) return null;
		const values = points.map((p) => p.value);
		const min = Math.min(...values);
		const max = Math.max(...values);
		const pad = (max - min) * 0.08 || 1;
		const lo = min - pad;
		const hi = max + pad;
		const xs = points.map((_, i) => (i / last) * WIDTH);
		const ys = points.map((p) => HEIGHT - ((p.value - lo) / (hi - lo)) * HEIGHT);
		const line = xs
			.map((x, i) => `${i === 0 ? 'M' : 'L'}${x.toFixed(2)} ${ys[i]!.toFixed(2)}`)
			.join(' ');
		return { line, area: `${line} L${WIDTH} ${HEIGHT} L0 ${HEIGHT} Z` };
	});

	const readout = $derived(
		marked === null ? '' : `${points[marked]?.date} · ${points[marked]?.amount}`
	);

	function clamp(index: number): number {
		return Math.min(last, Math.max(0, index));
	}

	function onkeydown(event: KeyboardEvent) {
		const from = marked ?? last;
		if (event.key === 'ArrowLeft') marked = clamp(from - 1);
		else if (event.key === 'ArrowRight') marked = clamp(from + 1);
		else if (event.key === 'Home') marked = 0;
		else if (event.key === 'End') marked = last;
		else return;
		event.preventDefault();
	}

	function onpointermove(event: PointerEvent) {
		const box = (event.currentTarget as HTMLElement).getBoundingClientRect();
		if (box.width === 0) return;
		marked = clamp(Math.round(((event.clientX - box.left) / box.width) * last));
	}
</script>

<section class="chart" aria-label={title}>
	<header class="header">
		<h2 class="title">{title}</h2>
		{#if span && !shortHistory}
			<span class="span">{span}</span>
		{/if}
	</header>

	{#if shortHistory || !geometry}
		<p class="note">A balance chart appears once there is a week of history.</p>
	{:else}
		<div class="plot-row">
			<div class="scale">
				<span>{high}</span>
				<span>{low}</span>
			</div>
			<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div
				class="plot"
				role="img"
				aria-label={summary}
				tabindex="0"
				{onkeydown}
				{onpointermove}
				onpointerleave={() => (marked = null)}
			>
				<svg viewBox="0 0 {WIDTH} {HEIGHT}" preserveAspectRatio="none" aria-hidden="true">
					<path class="area" d={geometry.area} />
					<path class="line" d={geometry.line} />
				</svg>
				{#if marked !== null}
					<span class="marker" style:left="{(marked / last) * 100}%" aria-hidden="true"></span>
				{/if}
			</div>
		</div>

		<div class="dates">
			<span>{points[0]?.date}</span>
			<span>{points[last]?.date}</span>
		</div>
	{/if}

	<p class="readout" aria-live="polite">{readout}</p>

	{#if coverage}
		<p class="coverage">{coverage}</p>
	{/if}
</section>

<style>
	.chart {
		display: flex;
		flex-direction: column;
		gap: 12px;
		inline-size: 100%;
		padding: 16px;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-md);
	}

	.header {
		display: flex;
		align-items: baseline;
		gap: 12px;
		min-block-size: 21px;
	}

	.title {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		font-weight: 700;
	}

	.span {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	.note,
	.coverage {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	.plot-row {
		display: flex;
		gap: 8px;
		block-size: 200px;
	}

	.scale {
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		flex: none;
		inline-size: 72px;
		text-align: end;
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 11px;
	}

	.plot {
		position: relative;
		flex: 1 1 auto;
		min-inline-size: 0;
		border-block-end: 1px solid var(--color-border-subtle);
		touch-action: none;
	}

	.plot:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
	}

	svg {
		display: block;
		inline-size: 100%;
		block-size: 100%;
	}

	.area {
		fill: var(--color-chart-1);
		fill-opacity: 0.12;
	}

	.line {
		fill: none;
		stroke: var(--color-chart-1);
		stroke-width: 2px;
		vector-effect: non-scaling-stroke;
		stroke-linejoin: round;
	}

	.marker {
		position: absolute;
		inset-block: 0;
		inline-size: 1px;
		background: var(--color-border-strong);
		pointer-events: none;
	}

	.dates {
		display: flex;
		justify-content: space-between;
		padding-inline-start: 80px;
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 11px;
		/* A fixed line, so the card's height does not follow the font's metrics. */
		line-height: 14px;
	}

	.readout {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-mono);
		font-size: 12px;
	}

	/* Stays in the tree so a change is announced, and gives back the gap it
	   would otherwise add while nothing is pointed at. */
	.readout:empty {
		margin-block-start: -12px;
	}
</style>
