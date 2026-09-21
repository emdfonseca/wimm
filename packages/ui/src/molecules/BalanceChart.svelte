<script lang="ts" module>
	export interface BalancePoint {
		/** Short: `3 Aug`. */
		date: string;
		/** What the line is drawn from. */
		value: number;
		/** Already formatted, carrying its currency. */
		amount: string;
		/** Against the day before: `€805.79 less than 2 Aug`, `No change from
		 *  8 Aug`. Absent on the first day, which has no day before. */
		change?: string;
		/** What moved the day, largest first. */
		movers?: BalanceMover[];
		/** `and 3 smaller`, when rows were left out of `movers`. */
		smaller?: string;
		/** Nothing happened this day. */
		empty?: boolean;
	}

	export interface BalanceMover {
		name: string;
		/** Already formatted, carrying its sign. */
		amount: string;
		negative?: boolean;
		/** The one payment rule marked it. */
		unusual?: boolean;
		/** It is one half of a movement between two accounts the member owns.
		 *  Takes the tag from `unusual`: a transfer is never also unusual. */
		transfer?: boolean;
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

	const point = $derived(marked === null ? undefined : points[marked]);

	function unusualLabel(mover: BalanceMover): string {
		return mover.negative === false ? 'Unusual income' : 'Unusual';
	}

	// One mover carries one tag. A transfer wins it, so the popover and the
	// announcement cannot disagree about which a row is.
	function moverTag(mover: BalanceMover): string | undefined {
		if (mover.transfer) return 'Between your accounts';
		return mover.unusual ? unusualLabel(mover) : undefined;
	}

	// The popover is aria-hidden; this is the same words as one string, in the
	// order date, balance, change, movers, smaller.
	const readout = $derived.by(() => {
		if (!point) return '';
		const parts = [`${point.date}.`, `${point.amount}.`];
		if (point.change) parts.push(`${point.change}.`);
		for (const mover of point.movers ?? []) {
			const tag = moverTag(mover);
			parts.push(`${mover.name} ${mover.amount}${tag ? `, ${tag}` : ''}.`);
		}
		if (point.smaller)
			parts.push(`${point.smaller.charAt(0).toUpperCase()}${point.smaller.slice(1)}.`);
		if (point.empty) parts.push('No transactions this day.');
		return parts.join(' ');
	});

	let plot: HTMLElement | undefined = $state();
	const flip = $derived(marked !== null && last > 0 && marked / last > 0.5);
	// The popover takes the room on its side of the marker and no more, so on a
	// phone it is narrowed rather than run off the plot.
	const room = $derived(
		marked === null || last === 0 ? 100 : (flip ? marked / last : 1 - marked / last) * 100
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
		else if (event.key === 'Escape') marked = null;
		else return;
		event.preventDefault();
	}

	function mark(event: PointerEvent) {
		const box = (event.currentTarget as HTMLElement).getBoundingClientRect();
		if (box.width === 0) return;
		marked = clamp(Math.round(((event.clientX - box.left) / box.width) * last));
	}

	// A mouse leaving clears the day; a tap keeps it until the next tap.
	function onpointerleave(event: PointerEvent) {
		if (event.pointerType === 'mouse') marked = null;
	}

	function outside(event: PointerEvent) {
		if (marked !== null && plot && !plot.contains(event.target as Node)) marked = null;
	}
</script>

<svelte:window onpointerdown={outside} />

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
				bind:this={plot}
				role="img"
				aria-label={summary}
				tabindex="0"
				{onkeydown}
				onpointermove={mark}
				onpointerdown={mark}
				{onpointerleave}
			>
				<svg viewBox="0 0 {WIDTH} {HEIGHT}" preserveAspectRatio="none" aria-hidden="true">
					<path class="area" d={geometry.area} />
					<path class="line" d={geometry.line} />
				</svg>
				{#if marked !== null}
					<span class="marker" style:left="{(marked / last) * 100}%" aria-hidden="true"></span>
				{/if}
				{#if point && marked !== null}
					<div
						class="popover"
						class:flip
						style:left="{(marked / last) * 100}%"
						style:max-inline-size="min(260px, calc({room}% - 14px))"
						aria-hidden="true"
					>
						<span class="pop-date">{point.date}</span>
						<span class="pop-balance">{point.amount}</span>
						{#if point.change}
							<span class="pop-change">{point.change}</span>
						{/if}
						{#each point.movers ?? [] as mover (mover.name + mover.amount)}
							<span class="mover">
								<span class="mover-name">{mover.name}</span>
								<span class="mover-amount">{mover.amount}</span>
								{#if moverTag(mover)}
									<span class="mover-tag">{moverTag(mover)}</span>
								{/if}
							</span>
						{/each}
						{#if point.smaller}
							<span class="pop-change">{point.smaller}</span>
						{/if}
						{#if point.empty}
							<span class="pop-change">No transactions this day.</span>
						{/if}
					</div>
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
		position: relative;
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

	/* The announcement. The popover shows the same words; this carries them
	   to a screen reader and takes no room. */
	.readout {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	.popover {
		position: absolute;
		inset-block-start: 8px;
		display: flex;
		flex-direction: column;
		gap: 2px;
		inline-size: max-content;
		min-inline-size: 120px;
		max-inline-size: 260px;
		margin-inline-start: 10px;
		padding: 8px 10px;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border-strong);
		border-radius: var(--radius-md);
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 12px;
		pointer-events: none;
	}

	.popover.flip {
		transform: translateX(calc(-100% - 20px));
	}

	.pop-date {
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 11px;
	}

	.pop-balance {
		font-family: var(--type-family-mono);
		font-size: 13px;
		font-weight: 600;
	}

	.pop-change {
		color: var(--color-text-secondary);
		font-size: 11px;
	}

	.mover {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		column-gap: 8px;
	}

	.mover-name {
		flex: 1 1 5rem;
		min-inline-size: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.mover-amount {
		flex: none;
		font-family: var(--type-family-mono);
	}

	.mover-tag {
		flex: none;
		color: var(--color-text-secondary);
		font-size: 11px;
	}
</style>
