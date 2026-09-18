<script lang="ts" module>
	export type Density = 'comfortable' | 'compact';

	/** The key the pre-paint script in app.html reads. Shared, so they cannot
	 *  drift — the same reasoning `ThemeToggle`'s `THEME_KEY` already applies. */
	export const DENSITY_KEY = 'wimm-density';

	/** Applies a choice to the document and remembers it. */
	export function applyDensity(density: Density) {
		document.documentElement.setAttribute('data-density', density);
		try {
			localStorage.setItem(DENSITY_KEY, density);
		} catch {
			// A browser with storage blocked still gets the density it asked
			// for; it just will not be remembered. That is worth less than
			// failing.
		}
	}

	/** What is stored, or "comfortable" when nothing is — the default the
	 *  token set itself states. */
	export function storedDensity(): Density {
		try {
			const stored = localStorage.getItem(DENSITY_KEY);
			if (stored === 'comfortable' || stored === 'compact') return stored;
		} catch {
			// Unreadable storage is the same as none.
		}
		return 'comfortable';
	}
</script>

<script lang="ts">
	import SegmentedControl from '../atoms/SegmentedControl.svelte';

	/**
	 * Origin `Yt1rq`. A preset of Segmented control that sets how tall a row is.
	 * Two options and no third: comfortable is the default and compact is what
	 * a member reaches for when they are reconciling a month and want to see it
	 * without scrolling.
	 *
	 * Absent where a pointer is coarse, not disabled (ADR 0004): compact rows
	 * are smaller than a finger reliably hits, so on a touch device there is
	 * nothing here to work out how to enable. `display: none` under `(any-
	 * pointer: coarse)` is how, rather than a prop the caller has to remember to
	 * pass — a member who chose compact on a laptop keeps that choice, and the
	 * phone simply does not offer to change it.
	 *
	 * A member's choice has to be in force before the first paint, by the same
	 * inline script in `app.html` that already applies the stored theme — a
	 * component runs after first paint, and the product would visibly
	 * rearrange.
	 */
	interface Props {
		/** Omitted, it reads what is stored on mount. */
		value?: Density;
		/** Names the group for assistive technology. Settings labels the row
		 *  "Rows". */
		label?: string;
		onchange?: (density: Density) => void;
	}

	let { value = $bindable('comfortable'), label = 'Row density', onchange }: Props = $props();

	// Read on mount rather than at module scope: this component renders on the
	// server too, where there is no localStorage and no document.
	$effect(() => {
		value = storedDensity();
	});

	const options = [
		{ value: 'comfortable' as const, label: 'Comfortable' },
		{ value: 'compact' as const, label: 'Compact' }
	];

	function choose(next: string) {
		value = next as Density;
		applyDensity(value);
		onchange?.(value);
	}
</script>

<div class="density-control">
	<SegmentedControl {options} {value} {label} onchange={choose} />
</div>

<style>
	@media (any-pointer: coarse) {
		.density-control {
			display: none;
		}
	}
</style>
