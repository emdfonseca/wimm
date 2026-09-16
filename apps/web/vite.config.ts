import { sveltekit } from '@sveltejs/kit/vite';
import { resolve } from 'node:path';
import { defineConfig } from 'vite';

/**
 * No API proxy. The browser talks only to this app; this app's server talks to
 * wimmd over Connect (ADR 0001). One origin, and nothing to configure for
 * WebAuthn or for the session cookie.
 */
export default defineConfig({
	plugins: [sveltekit()],
	server: {
		// The design system's font files live in packages/ui, outside this app's
		// root, and the dev server refuses to serve outside it by default. A
		// production build emits them as assets and never reaches this.
		fs: { allow: [resolve(import.meta.dirname, '../..')] }
	}
});
