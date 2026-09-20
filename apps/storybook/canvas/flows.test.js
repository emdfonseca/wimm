import { resolve } from 'node:path';
import { buildIndex } from 'storybook/internal/core-server';
import { beforeAll, expect, it } from 'vitest';
import { flows, routes, unplaced } from './flows.js';
import { flowFrom, kindProblems, placementProblems } from './lib.js';

// The real story index, built the way Storybook builds it. Checking the flows
// against ids taken from the flows themselves would pass whatever they named.
/** @type {Awaited<ReturnType<typeof buildIndex>>} */
let index;
beforeAll(async () => {
	index = await buildIndex({ configDir: resolve(import.meta.dirname, '../.storybook') });
}, 60_000);

it('every declared flow names stories that exist', () => {
	for (const name of Object.keys(flows)) {
		expect(() => flowFrom(flows, name, index)).not.toThrow();
	}
});

it('every page story that shows something is on a flow or listed as unplaced', () => {
	const pages = Object.values(index.entries).filter(
		(e) => e.type === 'story' && e.title.startsWith('Pages/')
	);
	expect(placementProblems(flows, unplaced, pages)).toEqual([]);
});

it('every page story says what kind it is', () => {
	const pages = Object.values(index.entries).filter(
		(e) => e.type === 'story' && e.title.startsWith('Pages/')
	);
	expect(kindProblems(pages)).toEqual([]);
});

it('every route stands for a story that exists', () => {
	expect(Object.keys(routes)).toEqual(['/', '/accounts', '/transactions', '/settings']);
	for (const id of Object.values(routes)) {
		expect(index.entries[id]?.type, id).toBe('story');
	}
});
