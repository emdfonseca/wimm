import type { StorybookConfig } from '@storybook/sveltekit';

/**
 * Storybook is its own app because a static build is deployable, and because
 * `@storybook/sveltekit` wants a real SvelteKit project to build against.
 *
 * It globs `packages/ui` only. Reaching into `apps/web` would be an apps → apps
 * dependency, which the monorepo standard bans and whose stated remedy is to
 * extract a package — so the pure screens live in `packages/ui/src/pages/` and
 * `apps/web` holds nothing but route wiring.
 */
const config: StorybookConfig = {
	stories: [
		'../../../packages/ui/src/**/*.mdx',
		'../../../packages/ui/src/**/*.stories.svelte'
	],
	addons: [
		'@storybook/addon-svelte-csf',
		'@storybook/addon-a11y',
		'@storybook/addon-docs',
		'@storybook/addon-vitest'
	],
	framework: { name: '@storybook/sveltekit', options: {} }
};

export default config;
