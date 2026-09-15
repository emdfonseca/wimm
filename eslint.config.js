import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import ts from 'typescript-eslint';
import globals from 'globals';
import svelteConfig from './apps/web/svelte.config.js';

export default ts.config(
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs.recommended,
	{
		languageOptions: { globals: { ...globals.browser, ...globals.node } },
		rules: {
			// Server-side logs go through the OTel setup in hooks.server.ts.
			'no-console': 'error'
		}
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts'],
		languageOptions: { parserOptions: { parser: ts.parser, svelteConfig } }
	},
	{
		ignores: [
			'**/.svelte-kit/',
			'**/dist/',
			'**/build/',
			'**/node_modules/',
			'**/gen/',
			'.devbox/',
			'.venv/',
			'.pgdata/',
			'.claude/',
			'.claude-home/',
			'.artifacts/',
			'bin/'
		]
	}
);
