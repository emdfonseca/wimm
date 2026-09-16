import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/**
 * A minimal SvelteKit app whose only job is to be the thing
 * `@storybook/sveltekit` builds against. It serves no routes of its own.
 */
export default {
	preprocess: vitePreprocess(),
	kit: { adapter: adapter() }
};
