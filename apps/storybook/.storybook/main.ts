import { resolve } from 'node:path';
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
	stories: ['../../../packages/ui/src/**/*.mdx', '../../../packages/ui/src/**/*.stories.svelte'],
	addons: [
		'@storybook/addon-svelte-csf',
		'@storybook/addon-a11y',
		'@storybook/addon-docs',
		'@storybook/addon-vitest'
	],
	// The design canvas is plain files served beside Storybook so its page shares
	// an origin with iframe.html, which is what lets it reach into an artboard.
	// The UI package's sources ride along at /canvas-ui for its stylesheet.
	staticDirs: [
		{ from: '../canvas', to: '/canvas' },
		{ from: '../../../packages/ui/src', to: '/canvas-ui' }
	],
	framework: { name: '@storybook/sveltekit', options: {} },
	// Every preview opens a socket to Storybook's server channel, and the server
	// has no error handler on it: one message over ws's 100 MiB limit ends the
	// dev server with "Max payload size exceeded". A story inside the canvas
	// reaches a window holding hundreds of frames, which is how a serialised
	// event gets that large. An artboard needs nothing from that channel, so
	// inside the canvas the socket never opens. Vite's own socket is untouched.
	previewHead: (head) => `${head}
<script>
	(() => {
		let inCanvas = false;
		try {
			inCanvas = parent !== window && parent.location.pathname.startsWith('/canvas/');
		} catch {}
		if (!inCanvas) return;
		window.WebSocket = new Proxy(window.WebSocket, {
			construct(target, args) {
				if (!String(args[0]).includes('/storybook-server-channel')) return new target(...args);
				return Object.assign(new EventTarget(), { readyState: 0, send() {}, close() {} });
			}
		});
	})();
</script>`,
	// The stories' fonts live in packages/ui, above this app's root, and Vite
	// serves nothing above it until the workspace root is allowed.
	viteFinal: (vite) => {
		vite.server ??= {};
		vite.server.fs ??= {};
		vite.server.fs.allow = [
			...(vite.server.fs.allow ?? []),
			resolve(import.meta.dirname, '../../..')
		];
		return vite;
	}
};

export default config;
