import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';
import { sveltekit } from '@sveltejs/kit/vite';
import { playwright } from '@vitest/browser-playwright';
import { appendFileSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { defineConfig } from 'vitest/config';
import { PICTURES_DIR, RUN_FILE } from './canvas/lib.js';

const runFile = resolve(import.meta.dirname, RUN_FILE);
const picturesDir = resolve(import.meta.dirname, PICTURES_DIR);

/** Room around the frame, so the runner's fit-to-window scale stays at 1. */
const WINDOW_MARGIN = 100;
let outerViewport: { width: number; height: number } | undefined;

/**
 * The run file the page versions are written from: emptied when a run starts,
 * given one line per page story by the `recordPageVersion` command, and closed
 * with whether the run as a whole passed.
 */
const runRecorder = {
	onTestRunStart() {
		mkdirSync(dirname(runFile), { recursive: true });
		writeFileSync(runFile, '');
		rmSync(picturesDir, { recursive: true, force: true });
	},
	onTestRunEnd(_modules: unknown, errors: readonly unknown[], reason: string) {
		const run = reason === 'passed' && errors.length === 0 ? 'passed' : 'failed';
		appendFileSync(runFile, `${JSON.stringify({ run })}\n`);
	}
};

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
		reporters: ['default', runRecorder],
		projects: [
			// Pure functions for the design canvas page, which has no DOM of its
			// own worth a browser: node is enough and far faster.
			{
				test: {
					name: 'unit',
					environment: 'node',
					include: ['canvas/**/*.test.js', 'scripts/**/*.test.js'],
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
						instances: [{ browser: 'chromium' }],
						commands: {
							recordPageVersion(_context, id: string, digest: string) {
								appendFileSync(runFile, `${JSON.stringify({ id, digest })}\n`);
							},
							// The runner scales its frame down to fit the window, and a
							// picture of a scaled frame is a thumbnail. Growing the window
							// past the frame keeps the scale at 1; `null` puts it back.
							async pictureWindow(context, size: { width: number; height: number } | null) {
								if (!size) {
									if (outerViewport) await context.page.setViewportSize(outerViewport);
									outerViewport = undefined;
									return;
								}
								outerViewport ??= context.page.viewportSize() ?? undefined;
								await context.page.setViewportSize({
									width: size.width + WINDOW_MARGIN,
									height: size.height + WINDOW_MARGIN
								});
							},
							// The stories `just approve` is approving, or none.
							pictureStories() {
								return (process.env.WIMM_PICTURE_STORIES ?? '').split(',').filter(Boolean);
							},
							// Named by size alone: the fingerprint is only known once
							// versions.mjs has added the shared styles to `digest`, so
							// approvals.mjs names the picture when it keeps it.
							keepPicture(_context, story: string, size: string, digest: string, png: string) {
								const file = resolve(picturesDir, story, `${size}.png`);
								mkdirSync(dirname(file), { recursive: true });
								writeFileSync(file, Buffer.from(png, 'base64'));
								appendFileSync(
									resolve(picturesDir, 'pictures.jsonl'),
									`${JSON.stringify({ story, size, digest })}\n`
								);
							}
						}
					}
				}
			}
		]
	}
});
