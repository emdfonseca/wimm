import { mount, unmount } from 'svelte';
import { expect, it } from 'vitest';
import Overview from '../../../packages/ui/src/pages/Overview.svelte';
import { nearestLocation } from './lib.js';

// The one test in a real browser: pointing rests on `__svelte_meta`, a dev
// internal of Svelte and not an API, so this is the first thing to fail if an
// upgrade moves it.
it('reports the Overview source for an element Overview.svelte renders', () => {
	const target = document.createElement('div');
	document.body.append(target);
	const screen = mount(Overview, {
		target,
		props: {
			hasAccounts: true,
			currencies: [{ currency: 'EUR', household: '€1.00' }],
			recentTransactions: [
				{
					id: 't1',
					date: '17 Sep',
					description: 'Coffee',
					amount: '−€1.00',
					account: 'Current account · Monzo',
					negative: true
				}
			]
		}
	});

	const title = [...target.querySelectorAll('h2')].find(
		(el) => el.textContent === 'Recent transactions'
	);
	const found = nearestLocation(title);

	expect(found?.path).toMatch(/packages\/ui\/src\/pages\/Overview\.svelte$/);
	expect(found?.line).toBeGreaterThan(0);

	unmount(screen);
	target.remove();
});
