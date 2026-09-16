import { sveltekit } from '@sveltejs/kit/vite';
import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { defineConfig } from 'vite';

/**
 * No API proxy. The browser talks only to this app; this app's server talks to
 * wimmd over Connect (ADR 0001). One origin, and nothing to configure for
 * WebAuthn or for the session cookie.
 */

/**
 * TLS when a certificate has been issued, plain HTTP when it has not.
 *
 * Enable Banking requires an https redirect URI even on localhost, so
 * connecting a bank needs the dev server to serve TLS — `just dev-cert` issues
 * one with mkcert. Conditional rather than required, because a checkout that
 * has not issued one should still run: identity, Storybook and every test work
 * over http, and only the bank hand-off does not.
 *
 * Read here rather than passed as paths so a missing file is a start-up
 * decision rather than a listener that binds and then fails a handshake.
 */
const certDir = resolve(import.meta.dirname, '../../.devbox/certs');
const cert = resolve(certDir, 'localhost.pem');
const key = resolve(certDir, 'localhost-key.pem');

const https =
	existsSync(cert) && existsSync(key)
		? { cert: readFileSync(cert), key: readFileSync(key) }
		: undefined;

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		https,
		// The design system's font files live in packages/ui, outside this app's
		// root, and the dev server refuses to serve outside it by default. A
		// production build emits them as assets and never reaches this.
		fs: { allow: [resolve(import.meta.dirname, '../..')] }
	}
});
