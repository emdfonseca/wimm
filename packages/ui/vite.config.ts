import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

// Only for svelte-check and editor tooling. The design system is not built:
// it exports source, and apps/storybook is what renders it.
export default defineConfig({
	plugins: [svelte()]
});
