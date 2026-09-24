import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/**
 * A minimal SvelteKit app whose only job is to be the thing
 * `@storybook/sveltekit` builds against. It serves no routes of its own.
 */
export default {
	preprocess: vitePreprocess(),
	// In the headless run the scoped class is a hash of the component's own CSS
	// rather than of its file name, so it moves exactly when the component's
	// styles do, which lets the markup alone carry a page's version. Dev keeps
	// the file-name hash: a class that changes on every style edit costs hot
	// reload, and nothing is versioned there.
	compilerOptions: process.env.VITEST ? { cssHash: ({ css, hash }) => `svelte-${hash(css)}` } : {},
	kit: { adapter: adapter() }
};
