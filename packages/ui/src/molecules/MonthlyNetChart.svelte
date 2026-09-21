<script lang="ts" module>
	export interface NetMonth {
		/** `March 2026`. What the readout and the month's row are called. */
		label: string;
		/** `Mar`. Under the bar. */
		short: string;
		/** What the bar is drawn from, in the currency's unit. */
		value: number;
		/** Already formatted, carrying its currency and its sign. */
		net: string;
		/** So far, or held from part way through: drawn lighter, since it is
		 *  not a whole month set beside whole months. */
		partial?: boolean;
	}
</script>

<script lang="ts">
	/**
	 * One bar per month from a zero rule, above it or below it, with a dashed
	 * rule at the typical month. A month's net is a signed quantity whose
	 * meaning is which side of zero it is on, so the scale always includes zero
	 * and the bars are drawn from it.
	 *
	 * This formats nothing: every amount, month and sentence arrives written.
	 */
	interface Props {
		/** `Month by month · EUR`. */
		title: string;
		/** `September 2025 to September 2026`. */
		span?: string;
		/** Oldest first. */
		months: NetMonth[];
		/** Where the typical month sits, in the same unit as `value`. Absent
		 *  until three full months are held. */
		typical?: number;
		/** The top and bottom of the scale column, already formatted. */
		high: string;
		low: string;
		/** The chart in words. It is the chart's accessible name. */
		summary: string;
		/** The index of the month pointed at, or null. */
		marked?: number | null;
	}

	let {
		title,
		span,
		months,
		typical,
		high,
		low,
		summary,
		marked = $bindable(null)
	}: Props = $props();

	const WIDTH = 1000;
	const HEIGHT = 200;
	const BAR = 0.6;

	const last = $derived(months.length - 1);
	const slot = $derived(WIDTH / Math.max(1, months.length));

	const scale = $derived.by(() => {
		const values = [0, ...months.map((m) => m.value)];
		const min = Math.min(...values);
		const max = Math.max(...values);
		const pad = (max - min) * 0.08 || 1;
		const lo = min - pad;
		const hi = max + pad;
		const y = (value: number) => HEIGHT - ((value - lo) / (hi - lo)) * HEIGHT;
		return { y, zero: y(0) };
	});

	const bars = $derived(
		months.map((m, i) => {
			const top = Math.min(scale.y(m.value), scale.zero);
			const height = Math.abs(scale.y(m.value) - scale.zero);
			return { x: i * slot + (slot * (1 - BAR)) / 2, y: top, width: slot * BAR, height };
		})
	);

	const readout = $derived(
		marked === null ? '' : `${months[marked]?.label} · ${months[marked]?.net}`
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

	// Pointing at a month selects it and it stays selected: the month's row
	// beside the chart is open because of it.
	function point(event: PointerEvent) {
		const box = (event.currentTarget as HTMLElement).getBoundingClientRect();
		if (box.width === 0) return;
		marked = clamp(Math.floor(((event.clientX - box.left) / box.width) * months.length));
	}
</script>

<section class="chart" class:dense={months.length > 8} aria-label={title} style:--months={months.length}>
	<header class="header">
		<h2 class="title">{title}</h2>
		{#if span}
			<span class="span">{span}</span>
		{/if}
	</header>

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
			onpointermove={point}
			onpointerdown={point}
		>
			{#if marked !== null}
				<span
					class="band"
					style:left="{(marked / months.length) * 100}%"
					style:inline-size="{100 / months.length}%"
					aria-hidden="true"
				></span>
			{/if}
			<svg viewBox="0 0 {WIDTH} {HEIGHT}" preserveAspectRatio="none" aria-hidden="true">
				<line class="zero" x1="0" x2={WIDTH} y1={scale.zero} y2={scale.zero} />
				{#each months as month, i (month.label)}
					<rect
						class="bar"
						class:partial={month.partial}
						data-month={month.label}
						x={bars[i]!.x}
						y={bars[i]!.y}
						width={bars[i]!.width}
						height={bars[i]!.height}
					/>
				{/each}
				{#if typical !== undefined}
					<line class="typical" x1="0" x2={WIDTH} y1={scale.y(typical)} y2={scale.y(typical)} />
				{/if}
			</svg>
		</div>
	</div>

	<div class="labels">
		{#each months as month (month.label)}
			<span>{month.short}</span>
		{/each}
	</div>

	<p class="readout" aria-live="polite">{readout}</p>
</section>

<style>
	.chart {
		/* A few months are not stretched across the card into slabs. */
		--slot-max: 120px;
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
		flex-wrap: wrap;
		align-items: baseline;
		column-gap: 12px;
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
		max-inline-size: calc(var(--months) * var(--slot-max));
		touch-action: none;
	}

	.plot:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
	}

	svg {
		position: relative;
		display: block;
		inline-size: 100%;
		block-size: 100%;
	}

	.band {
		position: absolute;
		inset-block: 0;
		background: var(--color-bg-subtle);
		pointer-events: none;
	}

	.bar {
		fill: var(--color-chart-1);
	}

	.bar.partial {
		fill-opacity: 0.4;
	}

	.zero {
		stroke: var(--color-border-strong);
		stroke-width: 1px;
		vector-effect: non-scaling-stroke;
	}

	.typical {
		stroke: var(--color-text-secondary);
		stroke-width: 1px;
		stroke-dasharray: 4 4;
		vector-effect: non-scaling-stroke;
	}

	.labels {
		display: flex;
		max-inline-size: calc(80px + var(--months) * var(--slot-max));
		padding-inline-start: 80px;
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 11px;
		line-height: 14px;
	}

	.labels span {
		flex: 1 1 0;
		min-inline-size: 0;
		text-align: center;
	}

	.readout {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-mono);
		font-size: 12px;
	}

	.readout:empty {
		margin-block-start: -12px;
	}

	/* Compact leaves the plot about 260 px: thirteen three-letter labels run
	   together there, so every other one is left to the readout. */
	@media (max-width: 767px) {
		.scale {
			inline-size: 56px;
		}

		.labels {
			max-inline-size: calc(64px + var(--months) * var(--slot-max));
			padding-inline-start: 64px;
		}

		.dense .labels span:nth-child(even) {
			visibility: hidden;
		}
	}
</style>
