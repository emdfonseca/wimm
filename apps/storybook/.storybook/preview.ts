import type { Preview } from '@storybook/sveltekit';

// The real token stylesheet, generated from design/tokens.json. Storybook must
// render against the contract, not a copy of it.
import '@wimm/ui/base.css';
import { DENSITY_KEY, THEME_KEY } from '@wimm/ui';

import ShellDecorator from './ShellDecorator.svelte';
import {
	MAX_HEIGHT,
	hiddenHeight,
	normaliseMarkup,
	sha256Hex,
	sizesFor
} from '../canvas/lib.js';
import { viewportOptions } from '../canvas/viewports.js';

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
		viewport: { options: viewportOptions }
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

	// Under vitest only (its mode is `test`), and only for page stories: what the story rendered, after
	// its play function, goes to the run file the versions are written from.
	async afterEach({ id, title, tags, canvasElement }) {
		if (import.meta.env.MODE !== 'test' || !title.startsWith('Pages/')) return;
		const { commands, page } = await import('vitest/browser');
		const run = commands as unknown as Record<string, (...args: unknown[]) => Promise<unknown>>;
		const digest = await sha256Hex(normaliseMarkup(canvasElement.outerHTML));
		await run.recordPageVersion!(id, digest);
		// The approve run pictures the stories being approved, from this same
		// render, so a picture can only show the version it is named for.
		if (!((await run.pictureStories!()) as string[]).includes(id)) return;
		const root = document.documentElement;
		const before = {
			width: innerWidth,
			height: innerHeight,
			theme: root.getAttribute('data-theme'),
			density: root.getAttribute('data-density')
		};
		root.setAttribute('data-theme', 'light');
		root.setAttribute('data-density', 'comfortable');
		const settled = async () => {
			await document.fonts.ready;
			for (let i = 0; i < 2; i++) await new Promise((done) => requestAnimationFrame(done));
		};
		for (const size of sizesFor({ tags }, Object.keys(viewportOptions))) {
			const { styles } = viewportOptions[size as keyof typeof viewportOptions];
			const width = parseInt(styles.width, 10);
			let height = parseInt(styles.height, 10);
			await run.pictureWindow!({ width, height });
			await page.viewport(width, height);
			await settled();
			// As tall as the artboard the canvas draws: grown until nothing sits
			// behind a scrollbar, as canvas.js measures it.
			for (let extra = hiddenHeight(document, window); extra > 1 && height < MAX_HEIGHT; ) {
				height = Math.min(MAX_HEIGHT, height + extra);
				await run.pictureWindow!({ width, height });
				await page.viewport(width, height);
				await settled();
				extra = hiddenHeight(document, window);
			}
			const png = await page.screenshot({ element: root, save: false });
			await run.keepPicture!(id, size, digest, png);
		}
		await run.pictureWindow!(null);
		await page.viewport(before.width, before.height);
		if (before.theme) root.setAttribute('data-theme', before.theme);
		if (before.density) root.setAttribute('data-density', before.density);
	},

	decorators: [
		// Innermost, so the shell wraps the screen that already has its callbacks.
		// Inert unless the design canvas asked to hear some: every test and every
		// ordinary visit passes through untouched.
		(story, context) => {
			const params = new URLSearchParams(location.search);
			if (!params.has('play')) return story();
			const asked = (params.get('play') ?? '').split(',').filter(Boolean);
			// Every callback the story wires is heard, mapped or not, so a screen
			// with nowhere to go can say so instead of going quiet.
			const names = new Set([
				...asked,
				...Object.keys(context.args).filter(
					(k) => /^on/.test(k) && typeof context.args[k] === 'function'
				)
			]);
			const args: Record<string, unknown> = { ...context.args };
			for (const name of names) {
				const original = context.args[name];
				args[name] = (...rest: unknown[]) => {
					const result = typeof original === 'function' ? original(...rest) : undefined;
					// A story's own play function dispatches synthetic events, which
					// never grant user activation. Only a person's click does.
					if (navigator.userActivation?.isActive) {
						window.parent.postMessage({ type: 'wimm-canvas-callback', name }, location.origin);
					}
					return result;
				};
			}
			return story({ args });
		},
		// A signed-in screen is seen inside the shell the real app puts around it.
		(story, context) =>
			context.parameters.shell
				? { Component: ShellDecorator, props: { current: String(context.parameters.shell) } }
				: story(),
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
