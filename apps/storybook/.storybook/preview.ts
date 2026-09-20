import type { Preview } from '@storybook/sveltekit';

// The real token stylesheet, generated from design/tokens.json. Storybook must
// render against the contract, not a copy of it.
import '@wimm/ui/base.css';
import { DENSITY_KEY, THEME_KEY } from '@wimm/ui';

/**
 * Theme, device and density are globals, not stories. A frame resolves one
 * value from each axis at a time; multiplying stories by theme × viewport is
 * exactly what the design standard forbids.
 */
const preview: Preview = {
	parameters: {
		layout: 'centered',
		// A failing contrast or missing label fails the run rather than
		// decorating a panel nobody opens.
		a11y: { test: 'error' },
		controls: { expanded: true },
		viewport: {
			options: {
				compact: { name: 'Compact', styles: { width: '390px', height: '844px' } },
				medium: { name: 'Medium', styles: { width: '834px', height: '1112px' } },
				wide: { name: 'Wide', styles: { width: '1440px', height: '900px' } },
				ultra: { name: 'Ultra', styles: { width: '1920px', height: '1080px' } }
			}
		}
	},

	initialGlobals: {
		theme: 'light',
		density: 'comfortable',
		viewport: { value: 'wide' }
	},

	globalTypes: {
		theme: {
			description: 'Colour axis',
			toolbar: {
				title: 'Theme',
				icon: 'paintbrush',
				items: [
					{ value: 'light', title: 'Light' },
					{ value: 'dark', title: 'Dark' }
				],
				dynamicTitle: true
			}
		},
		density: {
			description: 'Density axis',
			toolbar: {
				title: 'Density',
				icon: 'component',
				items: [
					{ value: 'comfortable', title: 'Comfortable' },
					{ value: 'compact', title: 'Compact' }
				],
				dynamicTitle: true
			}
		}
	},

	// Every story file shares one origin, so a choice one story stores is still
	// there when the next mounts. The theme and density controls read what is
	// stored on mount: with "compact" left behind, clicking Compact changes
	// nothing and the story asserting its onchange fails, in whichever run
	// happens to order the files that way.
	beforeEach() {
		localStorage.removeItem(THEME_KEY);
		localStorage.removeItem(DENSITY_KEY);
	},

	decorators: [
		(story, context) => {
			// The tokens stylesheet keys off these two attributes on the root, so
			// the globals are applied where the real page applies them.
			const root = document.documentElement;
			root.setAttribute('data-theme', String(context.globals.theme ?? 'light'));
			root.setAttribute('data-density', String(context.globals.density ?? 'comfortable'));
			document.body.style.background = 'var(--color-bg-canvas)';
			return story();
		}
	]
};

export default preview;
