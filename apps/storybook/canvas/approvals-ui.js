// The DOM for approvals on the design canvas: the badge on an artboard's label,
// the history button and panel, the sidebar filter, and a story's copyable id.
// Each function takes a document and data and returns elements; nothing runs on
// import, so a browser test can mount them. canvas.js does the wiring.
import { badgeWords, filterWords, historyWords } from './lib.js';

const SHAPES = /** @type {Record<string, string>} */ ({ approved: '✓', changed: '●', never: '○' });

/**
 * Where a page story stands, with its versions, as text first: the shape beside
 * it is decoration, so the status reads with the stylesheet removed.
 * @param {Document} doc
 * @param {ReturnType<typeof import('./lib.js').statusOf>} result
 * @returns {HTMLElement | null} null when there is nothing to say
 */
export function badge(doc, result) {
	const lines = badgeWords(result);
	if (lines.length === 0) return null;
	const el = doc.createElement('div');
	el.className = 'approval';
	el.dataset.status = result.status;
	const shape = SHAPES[result.status];
	if (shape) {
		const mark = doc.createElement('span');
		mark.className = 'mark';
		mark.setAttribute('aria-hidden', 'true');
		mark.textContent = shape;
		el.append(mark);
	}
	const text = doc.createElement('div');
	for (const line of lines) {
		const row = doc.createElement('div');
		row.textContent = line;
		text.append(row);
	}
	el.append(text);
	return el;
}

/**
 * A story's id, in monospace, with a button that copies it: what to paste when
 * asking about this story by name.
 * @param {Document} doc
 * @param {string} id
 * @param {{ copy: (text: string) => Promise<boolean>, say: (message: string) => void }} options
 */
export function idTag(doc, id, { copy, say }) {
	const el = doc.createElement('span');
	el.className = 'id-tag';
	const code = doc.createElement('code');
	code.textContent = id;
	const button = doc.createElement('button');
	button.type = 'button';
	button.className = 'id-copy';
	button.textContent = '⧉';
	button.setAttribute('aria-label', `Copy id ${id}`);
	button.addEventListener('click', async () => {
		say((await copy(id)) ? 'Copied.' : 'Could not copy.');
	});
	el.append(code, button);
	return el;
}

/**
 * @param {Document} doc
 * @param {{ name: string }} story
 */
export function historyButton(doc, story) {
	const button = doc.createElement('button');
	button.type = 'button';
	button.className = 'history-button';
	button.textContent = 'History';
	button.setAttribute('aria-label', `History of ${story.name}`);
	return button;
}

/** @type {{ dialog: HTMLDialogElement, button: HTMLElement } | null} */
let open = null;

/** Closes the panel that is open, if any, and hands focus back to its button. */
export function closeHistory() {
	if (!open) return;
	const { dialog, button } = open;
	open = null;
	dialog.remove();
	button.focus();
}

/**
 * Whether a key was pressed inside the history panel, where the canvas's
 * single-key shortcuts must do nothing.
 * @param {Event} event
 */
export const shortcutsInert = (event) =>
	event.target instanceof Element && event.target.closest('dialog.history') !== null;

/**
 * Opens the history of one story, non-modally beside the page it describes.
 * Only one panel is ever open.
 * @param {Document} doc
 * @param {{
 *   button: HTMLElement,
 *   story: { id: string, title: string, name: string },
 *   history: ReturnType<typeof import('./lib.js').historyOf>,
 *   copy: (text: string) => Promise<boolean>,
 *   say: (message: string) => void
 * }} options
 * @returns {HTMLDialogElement}
 */
export function openHistory(doc, { button, story, history, copy, say }) {
	closeHistory();
	const words = historyWords(story, history);
	const dialog = doc.createElement('dialog');
	dialog.className = 'history';
	const heading = doc.createElement('h2');
	heading.id = `history-${story.id}`;
	heading.tabIndex = -1;
	heading.textContent = words.heading;
	dialog.setAttribute('aria-labelledby', heading.id);
	dialog.append(heading);

	const list = doc.createElement('ol');
	for (const entry of words.entries) {
		const item = doc.createElement('li');
		const version = doc.createElement('div');
		version.className = 'version';
		version.textContent = entry.version;
		item.append(version);
		if (entry.mark) {
			const mark = doc.createElement('div');
			mark.className = 'kept';
			mark.textContent = entry.mark;
			item.append(mark);
		}
		for (const approval of entry.approvals) {
			const line = doc.createElement('div');
			line.textContent = approval.line;
			item.append(line);
			if (approval.note) {
				const note = doc.createElement('div');
				note.className = 'note';
				note.textContent = approval.note;
				item.append(note);
			}
		}
		list.append(item);
	}
	dialog.append(list);

	if (words.nobody) {
		const nobody = doc.createElement('p');
		nobody.textContent = words.nobody;
		dialog.append(nobody);
	}

	const exempt = history[0]?.mark === 'Implemented now';
	if (!exempt) {
		const intro = doc.createElement('p');
		intro.textContent = words.approveIntro;
		const command = doc.createElement('code');
		command.textContent = words.command;
		const copyButton = doc.createElement('button');
		copyButton.type = 'button';
		copyButton.textContent = 'Copy';
		copyButton.addEventListener('click', async () => {
			say((await copy(words.command)) ? 'Copied.' : 'Could not copy.');
		});
		dialog.append(intro, command, copyButton);
	}

	const close = doc.createElement('button');
	close.type = 'button';
	close.className = 'close';
	close.textContent = 'Close';
	close.addEventListener('click', closeHistory);
	dialog.append(close);
	dialog.addEventListener('keydown', (event) => {
		if (event.key !== 'Escape') return;
		event.preventDefault();
		closeHistory();
	});

	doc.body.append(dialog);
	const at = button.getBoundingClientRect();
	dialog.style.insetBlockStart = `${Math.max(0, at.bottom + 4)}px`;
	dialog.style.insetInlineStart = `${Math.max(0, at.left)}px`;
	dialog.show();
	// A label can run past the window's edge in a narrow window; the panel is
	// kept inside it rather than following the button off screen.
	const room = (doc.defaultView?.innerWidth ?? Infinity) - dialog.offsetWidth - 4;
	if (at.left > room) dialog.style.insetInlineStart = `${Math.max(0, room)}px`;
	open = { dialog, button };
	heading.focus();
	return dialog;
}

/**
 * The sidebar filter: a checkbox and the count, which is a polite live region
 * so turning the filter on is announced once.
 * @param {Document} doc
 * @param {{ checked: boolean, onChange: (checked: boolean) => void }} options
 */
export function filterControl(doc, { checked, onChange }) {
	const el = doc.createElement('div');
	el.className = 'needs';
	const label = doc.createElement('label');
	const input = doc.createElement('input');
	input.type = 'checkbox';
	input.checked = checked;
	const name = doc.createElement('span');
	name.textContent = 'Needs approval ';
	const count = doc.createElement('span');
	count.className = 'count';
	count.setAttribute('aria-live', 'polite');
	label.append(input, name, count);
	const empty = doc.createElement('p');
	empty.className = 'empty';
	empty.hidden = true;
	el.append(label, empty);

	let total = 0;
	const announce = () => {
		count.textContent = '';
		count.textContent = `(${total})`;
	};
	input.addEventListener('change', () => {
		announce();
		onChange(input.checked);
	});
	return {
		el,
		input,
		/**
		 * @param {number} needing every story that needs approval
		 * @param {number} shown how many the sidebar lists now
		 * @param {string} query what the find box holds
		 */
		update(needing, shown, query) {
			total = needing;
			const words = filterWords(needing, shown, query);
			count.textContent = `(${needing})`;
			empty.textContent = input.checked ? (words.empty ?? '') : '';
			empty.hidden = !(input.checked && words.empty);
		}
	};
}
