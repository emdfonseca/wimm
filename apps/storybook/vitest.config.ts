import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';
import { sveltekit } from '@sveltejs/kit/vite';
import { playwright } from '@vitest/browser-playwright';
import { resolve } from 'node:path';
import { defineConfig } from 'vitest/config';

/**
 * Stories run as tests. The play functions are the component assertions
 * (.claude/rules/typescript.md), so they belong to `just check` rather than to
 * a tab someone remembers to open.
 *
 * A real browser, not jsdom: focus, live regions and `:focus-visible` are the
 * contracts being asserted, and a DOM emulation is exactly where those stop
 * being true.
 */
export default defineConfig({
	// pnpm links every dependency into a store outside this app, and the stories
	// themselves live in packages/ui. Both are outside Vite's default served
	// root, so the browser cannot fetch them until the workspace root is allowed.
	server: { fs: { allow: [resolve(import.meta.dirname, '../..')] } },
	test: {
		projects: [
			// Pure functions for the design canvas page, which has no DOM of its
			// own worth a browser: node is enough and far faster.
			{
				test: {
					name: 'unit',
					environment: 'node',
					include: ['canvas/**/*.test.js'],
					exclude: ['canvas/**/*.browser.test.js']
				}
			},
			// Pointing reads what Svelte's dev build leaves on real elements, so
			// its one test needs a real browser and a compiled component.
			{
				extends: true,
				plugins: [sveltekit()],
				test: {
					name: 'browser',
					include: ['canvas/**/*.browser.test.js'],
					browser: {
						enabled: true,
						headless: true,
						provider: playwright(),
						instances: [{ browser: 'chromium' }]
					}
				}
			},
			{
				extends: true,
				plugins: [
					// `@storybook/sveltekit` contributes overrides only — the `$app/*`
					// mocks and the CSF plugins. It never adds vite-plugin-svelte,
					// because `storybook dev` and `storybook build` inherit it from the
					// app's own vite.config.ts. Vitest loads this file instead, so
					// without it every .svelte file reaches vite:import-analysis
					// untransformed and each story fails at import.
					sveltekit(),
					storybookTest({
						configDir: resolve(import.meta.dirname, '.storybook'),
						// Docs are a Storybook surface, not a test one, and pulling them
						// in makes Vite pre-bundle a React tree that pnpm never links
						// into this app.
						disableAddonDocs: true
					})
				],
				test: {
					name: 'storybook',
					browser: {
						enabled: true,
						headless: true,
						provider: playwright(),
						instances: [{ browser: 'chromium' }]
					}
				}
			}
		]
	}
});
