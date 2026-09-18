import { describe, expect, it } from 'vitest';
import { navigationAnnouncement } from './navigation-announcement.js';

describe('navigationAnnouncement', () => {
	it('says Loading while a navigation is in flight', () => {
		expect(navigationAnnouncement(true, '/(app)/transactions')).toBe('Loading');
	});

	it('names the destination once it lands', () => {
		expect(navigationAnnouncement(false, '/(app)/transactions')).toBe('Transactions');
	});

	it('names Overview for the signed-in landing route', () => {
		expect(navigationAnnouncement(false, '/(app)')).toBe('Overview');
	});

	it('says nothing for a route with no name in the map', () => {
		expect(navigationAnnouncement(false, '/(app)/unknown-route')).toBe('');
	});

	it('says nothing with no route id and no navigation', () => {
		expect(navigationAnnouncement(false, null)).toBe('');
	});
});
