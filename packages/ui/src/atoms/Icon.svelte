<script lang="ts">
	/**
	 * The lucide glyphs the canvas names, at the sizes it draws them.
	 *
	 * Inlined rather than pulled from a package: nine glyphs across this whole
	 * change, against a dependency whose tree-shaking is a build concern to
	 * verify. Each `d` is lucide's own, unchanged.
	 */
	type Name =
		| 'circle-help'
		| 'shield-check'
		| 'smartphone'
		| 'layout-dashboard'
		| 'user-round-check'
		| 'panel-left-close'
		| 'log-out'
		| 'arrow-left-right'
		| 'settings'
		| 'chevron-left'
		| 'chevron-right';

	interface Props {
		name: Name;
		size?: number;
	}

	let { name, size = 16 }: Props = $props();

	// settings is the one glyph lucide draws with a circle rather than only
	// paths, so it is kept separately instead of forcing a circle into a path.
	const circles: Partial<Record<Name, { cx: number; cy: number; r: number }>> = {
		settings: { cx: 12, cy: 12, r: 3 }
	};

	const paths: Record<Name, string[]> = {
		'circle-help': [
			'M12 17h.01',
			'M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3',
			'M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20'
		],
		'shield-check': [
			'M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z',
			'm9 12 2 2 4-4'
		],
		smartphone: [
			'M5 4a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v16a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2z',
			'M12 18h.01'
		],
		'layout-dashboard': ['M4 4h7v7H4z', 'M13 4h7v5h-7z', 'M13 13h7v7h-7z', 'M4 15h7v5H4z'],
		'user-round-check': [
			'M2 21a8 8 0 0 1 13.292-6',
			'M13 8a4 4 0 1 0-8 0 4 4 0 0 0 8 0',
			'm16 19 2 2 4-4'
		],
		'log-out': ['m16 17 5-5-5-5', 'M21 12H9', 'M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4'],
		'arrow-left-right': ['M8 3 4 7l4 4', 'M4 7h16', 'm16 21 4-4-4-4', 'M20 17H4'],
		'panel-left-close': [
			'M3 5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z',
			'M9 3v18',
			'm16 15-3-3 3-3'
		],
		settings: [
			'M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z'
		],
		'chevron-left': ['m15 18-6-6 6-6'],
		'chevron-right': ['m9 18 6-6-6-6']
	};

	const circle = $derived(circles[name]);
</script>

<svg
	viewBox="0 0 24 24"
	width={size}
	height={size}
	fill="none"
	stroke="currentColor"
	stroke-width="2"
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
	focusable="false"
>
	{#each paths[name] as d (d)}
		<path {d} />
	{/each}
	{#if circle}
		<circle cx={circle.cx} cy={circle.cy} r={circle.r} />
	{/if}
</svg>
