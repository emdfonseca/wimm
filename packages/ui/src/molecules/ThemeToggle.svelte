<script lang="ts" module>
	/**
	 * Three states, not two, because that is what the stylesheet already has:
	 * `tokens.css` resolves light and dark from `prefers-color-scheme` by
	 * default and lets `[data-theme]` on the root override it in either
	 * direction (ADR 0002's colour axis).
	 *
	 * So "system" is a real choice and the honest default — not the absence of
	 * one — and a two-way switch would make following the operating system
	 * unreachable once anybody had touched it.
	 */
	export type Theme = 'system' | 'light' | 'dark';

	/** The key the pre-paint script in app.html reads. Shared, so they cannot drift. */
	export const THEME_KEY = 'wimm-theme';

	/**
	 * Applies a choice to the document and remembers it.
	 *
	 * "system" removes the attribute rather than writing a value, because the
	 * absence of `data-theme` is what lets the media query decide.
	 */
	export function applyTheme(theme: Theme) {
		const root = document.documentElement;
		if (theme === 'system') {
			root.removeAttribute('data-theme');
		} else {
			root.setAttribute('data-theme', theme);
		}

		try {
			localStorage.setItem(THEME_KEY, theme);
		} catch {
			// A browser with storage blocked still gets the theme it asked for;
			// it just will not be remembered. That is worth less than failing.
		}
	}

	/** What is stored, or "system" when nothing is. */
	export function storedTheme(): Theme {
		try {
			const stored = localStorage.getItem(THEME_KEY);
			if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
		} catch {
			// Unreadable storage is the same as none.
		}
		return 'system';
	}
</script>

<script lang="ts">
	import SegmentedControl from '../atoms/SegmentedControl.svelte';

	interface Props {
		/** Omitted, it reads what is stored on mount. */
		value?: Theme;
		/** Names the group for assistive technology. Settings labels the row
		 *  "Theme"; anywhere else "Appearance" is the clearer name for the
		 *  same choice. */
		label?: string;
		onchange?: (theme: Theme) => void;
	}

	let { value = $bindable('system'), label = 'Appearance', onchange }: Props = $props();

	// Read on mount rather than at module scope: this component renders on the
	// server too, where there is no localStorage and no document.
	$effect(() => {
		value = storedTheme();
	});

	const options = [
		{ value: 'system' as const, label: 'System' },
		{ value: 'light' as const, label: 'Light' },
		{ value: 'dark' as const, label: 'Dark' }
	];

	function choose(next: string) {
		value = next as Theme;
		applyTheme(value);
		onchange?.(value);
	}
</script>

<SegmentedControl {options} {value} {label} onchange={choose} />
