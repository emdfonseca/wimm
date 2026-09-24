import { userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it } from 'vitest';
import {
	approvedLook,
	badge,
	closeHistory,
	filterControl,
	historyButton,
	idTag,
	openHistory,
	shortcutsInert
} from './approvals-ui.js';
import { historyOf, statusOf } from './lib.js';

const fp = (c) => c.repeat(64);
const entry = {
	implemented: fp('a'),
	versions: [
		{ fingerprint: fp('a'), firstSeen: '2026-09-20' },
		{ fingerprint: fp('b'), firstSeen: '2026-09-12' }
	]
};
const emanuel = (fingerprint) => ({
	story: 'pages-overview--populated',
	fingerprint,
	name: 'Emanuel Fonseca',
	email: 'e@x',
	at: '2026-09-20T18:04:11Z',
	note: ''
});
const story = { id: 'pages-overview--populated', title: 'Pages/Overview', name: 'Populated' };
const history = historyOf(entry, [], 'state');

afterEach(() => {
	closeHistory();
	document.body.replaceChildren();
});

/** A button on the page and its panel, opened. */
function mountHistory() {
	const bar = document.createElement('div');
	const button = historyButton(document, story);
	bar.append(button);
	document.body.append(bar);
	const copied = [];
	const said = [];
	const open = (b = button, s = story) =>
		openHistory(document, {
			button: b,
			story: s,
			history,
			copy: async (text) => {
				copied.push(text);
				return true;
			},
			say: (m) => said.push(m)
		});
	return { button, open, copied, said };
}

describe('the history button and panel', () => {
	it('is a button reached with Tab, named for its story', async () => {
		const { button } = mountHistory();
		expect(button.tagName).toBe('BUTTON');
		expect(button.getAttribute('aria-label')).toBe('History of Populated');
		await userEvent.tab();
		expect(document.activeElement).toBe(button);
	});

	it('moves focus to the heading on open', () => {
		const { open } = mountHistory();
		const dialog = open();
		const heading = dialog.querySelector('h2');
		expect(heading?.textContent).toBe('Overview · Populated');
		expect(document.activeElement).toBe(heading);
	});

	it('closes with Escape and returns focus to the button', async () => {
		const { button, open } = mountHistory();
		open();
		await userEvent.keyboard('{Escape}');
		expect(document.querySelector('dialog.history')).toBeNull();
		expect(document.activeElement).toBe(button);
	});

	it('closes with Close and returns focus to the button', async () => {
		const { button, open } = mountHistory();
		const dialog = open();
		/** @type {HTMLButtonElement} */ (dialog.querySelector('button.close')).click();
		expect(document.querySelector('dialog.history')).toBeNull();
		expect(document.activeElement).toBe(button);
	});

	it('is one panel at a time', () => {
		const { open } = mountHistory();
		const other = historyButton(document, { name: 'Empty' });
		document.body.append(other);
		open();
		open(other, { ...story, id: 'pages-overview--empty', name: 'Empty' });
		expect(document.querySelectorAll('dialog.history')).toHaveLength(1);
		expect(document.querySelector('dialog.history h2')?.textContent).toBe('Overview · Empty');
	});

	it('stays inside the window when its button is past the right edge', () => {
		const { button, open } = mountHistory();
		const style = document.createElement('style');
		style.textContent =
			'dialog.history { position: fixed; margin: 0; inset-inline-end: auto; inline-size: 340px }';
		document.body.append(style);
		button.style.position = 'fixed';
		button.style.left = `${innerWidth + 200}px`;
		const box = open().getBoundingClientRect();
		expect(box.left).toBeGreaterThanOrEqual(0);
		expect(box.right).toBeLessThanOrEqual(innerWidth);
	});

	it('leaves the canvas shortcuts inert while focus is inside', () => {
		const { open } = mountHistory();
		const dialog = open();
		const inside = new KeyboardEvent('keydown', { key: 'm', bubbles: true });
		dialog.querySelector('h2')?.dispatchEvent(inside);
		expect(shortcutsInert(inside)).toBe(true);
		const outside = new KeyboardEvent('keydown', { key: 'm', bubbles: true });
		document.body.dispatchEvent(outside);
		expect(shortcutsInert(outside)).toBe(false);
	});

	it('puts exactly the command on the clipboard and says so', async () => {
		const { open, copied, said } = mountHistory();
		const dialog = open();
		const copy = [...dialog.querySelectorAll('button')].find((b) => b.textContent === 'Copy');
		copy?.click();
		await Promise.resolve();
		await Promise.resolve();
		expect(copied).toEqual(['just approve pages-overview--populated']);
		expect(said).toEqual(['Copied.']);
	});

	it('says nobody has approved a page nobody approved', () => {
		const { open } = mountHistory();
		expect(open().textContent).toContain('Nobody has approved this page yet.');
	});
});

describe('the badge', () => {
	it('reads its status as text with no stylesheet applied', () => {
		const el = badge(document, statusOf(entry, [], 'state'));
		document.body.append(el);
		expect(el.textContent).toContain('Never approved');
		expect(el.textContent).toContain('Implemented aaaaaaa · 20 Sep 2026');
		expect(el.querySelector('.mark')?.getAttribute('aria-hidden')).toBe('true');
	});

	it('shows both versions when they differ, and one when they are the same', () => {
		const changed = badge(document, statusOf(entry, [emanuel(fp('b'))], 'state'));
		expect(changed.textContent).toContain('Changed since approval');
		expect(changed.textContent).toContain('Approved bbbbbbb');
		expect(changed.textContent).toContain('Implemented aaaaaaa');
		const same = badge(document, statusOf(entry, [emanuel(fp('a'))], 'state'));
		expect(same.textContent).not.toContain('Implemented');
	});

	it('draws nothing when there is nothing to say', () => {
		expect(badge(document, statusOf(undefined, [], 'state'))).toBeNull();
	});
});

describe('a story id', () => {
	it('shows the id and copies exactly it, saying so', async () => {
		const copied = [];
		const said = [];
		const el = idTag(document, 'pages-overview--populated', {
			copy: async (text) => {
				copied.push(text);
				return true;
			},
			say: (m) => said.push(m)
		});
		document.body.append(el);
		expect(el.querySelector('code')?.textContent).toBe('pages-overview--populated');
		const button = /** @type {HTMLButtonElement} */ (el.querySelector('button'));
		expect(button.getAttribute('aria-label')).toBe('Copy id pages-overview--populated');
		button.click();
		await Promise.resolve();
		await Promise.resolve();
		expect(copied).toEqual(['pages-overview--populated']);
		expect(said).toEqual(['Copied.']);
	});
});

describe('the Needs approval filter', () => {
	it('carries its count in a polite live region', () => {
		const control = filterControl(document, { checked: false, onChange() {} });
		document.body.append(control.el);
		control.update(23, 23, '');
		const live = control.el.querySelector('[aria-live]');
		expect(live?.getAttribute('aria-live')).toBe('polite');
		expect(control.el.querySelector('label')?.textContent).toBe('Needs approval (23)');
	});

	it('says everything is approved when the filter is on and nothing needs a look', () => {
		const control = filterControl(document, { checked: true, onChange() {} });
		document.body.append(control.el);
		control.update(0, 0, '');
		expect(control.el.querySelector('.empty')?.textContent).toBe(
			'Every page is approved at its implemented version.'
		);
	});
});

describe('the approved look on a changed artboard', () => {
	// The picture the changed-pictured fixture keeps of the approved version.
	const approved = '20f7800a2e231414a46bcf96107cace1bd0481bd5e390dffc7d40a71c270a654';
	const base = '/canvas/fixtures/changed-pictured';
	const light = { theme: 'light', density: 'comfortable' };
	const changed = statusOf(
		{
			implemented: fp('a'),
			versions: [
				{ fingerprint: fp('a'), firstSeen: '2026-09-20' },
				{ fingerprint: approved, firstSeen: '2026-09-12' }
			]
		},
		[emanuel(approved)],
		'state'
	);

	/**
	 * One artboard as canvas.js draws it: a label, a badge and a frame holding
	 * the story's iframe.
	 * @param {Partial<Parameters<typeof approvedLook>[1]>} [options]
	 */
	async function mountBoard(options = {}) {
		const label = document.createElement('div');
		const mark = badge(document, options.result ?? changed);
		const frame = document.createElement('div');
		const iframe = document.createElement('iframe');
		iframe.srcdoc = '<p>The page as implemented</p>';
		frame.append(iframe);
		document.body.append(label, ...(mark ? [mark] : []), frame);
		await new Promise((done) => iframe.addEventListener('load', done, { once: true }));
		const asked = [];
		const button = await approvedLook(document, {
			story,
			size: 'wide',
			width: 1440,
			result: changed,
			view: light,
			base,
			label,
			badge: mark,
			frame,
			iframe,
			exists: async (url) => {
				asked.push(url);
				return (await fetch(url, { method: 'HEAD' })).ok;
			},
			...options
		});
		return { label, mark, frame, iframe, button, asked };
	}

	it('is a toggle button, not pressed, reached with Tab, named for the story and size', async () => {
		const { button } = await mountBoard();
		expect(button?.tagName).toBe('BUTTON');
		expect(button?.textContent).toBe('Approved look');
		expect(button?.getAttribute('aria-pressed')).toBe('false');
		expect(button?.getAttribute('aria-label')).toBe('Approved look of Populated at wide');
		await userEvent.tab();
		expect(document.activeElement).toBe(button);
	});

	it('shows the picture in place of the page, says which version, and keeps focus', async () => {
		const { button, frame, iframe, mark } = await mountBoard();
		button?.focus();
		await userEvent.keyboard('{Enter}');
		expect(button?.getAttribute('aria-pressed')).toBe('true');
		expect(iframe.isConnected).toBe(true);
		expect(iframe.hidden).toBe(true);
		const img = frame.querySelector('img');
		expect(img?.alt).toBe('Overview · Populated at wide, approved 20f7800');
		await expect.poll(() => img?.naturalWidth).toBe(1440);
		expect(mark?.textContent).toContain('Showing approved 20f7800');
		expect(document.activeElement).toBe(button);
	});

	it('switches back to the same page, not a reloaded one', async () => {
		const { button, frame, iframe, mark } = await mountBoard();
		const page = iframe.contentDocument;
		const src = iframe.srcdoc;
		button?.click();
		button?.click();
		expect(button?.getAttribute('aria-pressed')).toBe('false');
		expect(iframe.hidden).toBe(false);
		expect(frame.querySelector('img')).toBeNull();
		expect(mark?.textContent).not.toContain('Showing approved');
		expect(iframe.contentDocument).toBe(page);
		expect(iframe.srcdoc).toBe(src);
	});

	it('says there is no picture, and offers nothing, when none was kept', async () => {
		const { button, label, mark } = await mountBoard({ base: '/canvas/fixtures/changed' });
		expect(button).toBeNull();
		expect(label.querySelector('button')).toBeNull();
		expect(mark?.textContent).toContain('No picture of the approved version');
	});

	it.each([
		['dark', { theme: 'dark', density: 'comfortable' }],
		['compact density', { theme: 'light', density: 'compact' }]
	])('offers nothing and asks for nothing viewed %s', async (_name, view) => {
		const { button, mark, asked } = await mountBoard({ view });
		expect(button).toBeNull();
		expect(mark?.textContent).not.toContain('No picture');
		expect(asked).toEqual([]);
	});

	it('asks for no picture on an artboard that has not changed', async () => {
		const same = statusOf(
			{ implemented: approved, versions: [{ fingerprint: approved, firstSeen: '2026-09-12' }] },
			[emanuel(approved)],
			'state'
		);
		const { button, asked } = await mountBoard({ result: same });
		expect(button).toBeNull();
		expect(asked).toEqual([]);
	});
});

