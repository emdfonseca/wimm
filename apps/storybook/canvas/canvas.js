import {
	KINDS,
	artboardsFrom,
	callbacksFor,
	connectorLabels,
	branchBoards,
	flowBoards,
	flowFrom,
	nearestLocation,
	nearestRecorded,
	nextStory,
	routeFor,
	sitemapFrom,
	screenUrl,
	storyUrl
} from './lib.js';
import { flows, routes } from './flows.js';
import { viewportOptions } from './viewports.js';

const status = /** @type {HTMLElement} */ (document.getElementById('status'));
const stage = /** @type {HTMLElement} */ (document.getElementById('stage'));
const find = /** @type {HTMLInputElement} */ (document.getElementById('find'));
const tree = /** @type {HTMLElement} */ (document.getElementById('tree'));
const world = /** @type {HTMLElement} */ (document.getElementById('world'));

/** Taller than this is a runaway layout, not a screen. */
const MAX_HEIGHT = 20000;
const MIN_SCALE = 0.05;
const MAX_SCALE = 4;

document.documentElement.style.setProperty(
	'--bar-height',
	`${document.getElementById('bar')?.offsetHeight ?? 40}px`
);

/** @param {string} size */
const sizeOf = (size) => {
	const { width, height } =
		viewportOptions[/** @type {keyof typeof viewportOptions} */ (size)].styles;
	return { width: parseInt(width, 10), height: parseInt(height, 10) };
};

/** @param {string} message */
const say = (message) => {
	status.textContent = message;
};

/** True while a flow is being played: one live artboard, no pan or zoom. */
let playing = false;

// ---- Pan and zoom: one transform on one container ----

const view = { x: 0, y: 0, scale: 1 };

let frame = 0;
let settle = 0;

/** One transform write per frame, however many wheel events arrived. */
function apply() {
	if (frame) return;
	frame = requestAnimationFrame(() => {
		frame = 0;
		world.style.transform = `translate(${view.x}px, ${view.y}px) scale(${view.scale})`;
	});
	// While moving, the layer is promoted and only ever composited, never
	// re-rasterised. Once still, it is released so the page is redrawn sharp at
	// the new scale.
	world.classList.add('moving');
	clearTimeout(settle);
	settle = window.setTimeout(() => {
		world.classList.remove('moving');
		pump();
	}, 150);
}

/**
 * @param {number} factor
 * @param {number} px pointer position in stage coordinates
 * @param {number} py
 */
function zoomAbout(factor, px, py) {
	const next = Math.min(MAX_SCALE, Math.max(MIN_SCALE, view.scale * factor));
	view.x = px - ((px - view.x) * next) / view.scale;
	view.y = py - ((py - view.y) * next) / view.scale;
	view.scale = next;
	apply();
}

/** The part of the stage not under the bar or the sitemap. */
function usable() {
	const bar = document.getElementById('bar')?.offsetHeight ?? 0;
	const map = document.body.classList.contains('no-map')
		? 0
		: (document.getElementById('sitemap')?.offsetWidth ?? 0);
	return { left: map, top: bar, width: stage.clientWidth - map, height: stage.clientHeight - bar };
}

function fit() {
	const area = usable();
	const width = world.offsetWidth;
	const height = world.offsetHeight;
	if (!width || !height) return;
	view.scale = Math.min(area.width / width, area.height / height);
	view.x = area.left;
	view.y = area.top;
	apply();
}

function actualSize() {
	const area = usable();
	view.scale = 1;
	view.x = area.left;
	view.y = area.top;
	apply();
}

/** @type {Map<string, HTMLElement[]>} the artboards of each story, one per size */
const boardsByStory = new Map();

/**
 * Bring one story's artboards, every size of it, into the usable area: as large
 * as fits up to 100%, centred across and pinned to the top so a tall page is
 * read from its head.
 * @param {string} id
 */
function jumpTo(id) {
	const els = boardsByStory.get(id);
	if (!els?.length) return;
	let [minX, minY, maxX, maxY] = [Infinity, Infinity, -Infinity, -Infinity];
	for (const el of els) {
		const box = el.getBoundingClientRect();
		minX = Math.min(minX, (box.left - view.x) / view.scale);
		minY = Math.min(minY, (box.top - view.y) / view.scale);
		maxX = Math.max(maxX, (box.right - view.x) / view.scale);
		maxY = Math.max(maxY, (box.bottom - view.y) / view.scale);
	}
	const pad = 32;
	const area = usable();
	const scale = Math.min(
		1,
		(area.width - pad * 2) / (maxX - minX),
		(area.height - pad * 2) / (maxY - minY)
	);
	view.scale = scale;
	view.x = area.left + (area.width - (maxX - minX) * scale) / 2 - minX * scale;
	view.y = area.top + pad - minY * scale;
	apply();
	history.replaceState(null, '', `#${id}`);
	for (const button of document.querySelectorAll('#tree button')) {
		button.setAttribute(
			'aria-current',
			String(/** @type {HTMLElement} */ (button).dataset.id === id)
		);
	}
	pump();
}

stage.addEventListener(
	'wheel',
	(event) => {
		if (playing) return;
		event.preventDefault();
		if (event.ctrlKey) {
			zoomAbout(
				Math.exp(-Math.max(-40, Math.min(40, event.deltaY)) * 0.01),
				event.clientX,
				event.clientY
			);
		} else {
			view.x -= event.deltaX;
			view.y -= event.deltaY;
			apply();
		}
	},
	{ passive: false }
);

/** @type {{ x: number, y: number } | null} */
let drag = null;
stage.addEventListener('pointerdown', (event) => {
	if (playing || event.altKey || event.button !== 0) return;
	drag = { x: event.clientX, y: event.clientY };
	stage.setPointerCapture(event.pointerId);
});
stage.addEventListener('pointermove', (event) => {
	if (!drag) return;
	view.x += event.clientX - drag.x;
	view.y += event.clientY - drag.y;
	drag = { x: event.clientX, y: event.clientY };
	apply();
});
stage.addEventListener('pointerup', () => {
	drag = null;
});

// A story that autofocuses would scroll the stage to reveal itself. Its
// position is the transform's alone.
stage.addEventListener('scroll', () => {
	stage.scrollTo(0, 0);
});

/** @param {KeyboardEvent} event */
function onKey(event) {
	document.body.classList.toggle('alt', event.altKey);
	if (event.target === find) return;
	if (event.type !== 'keydown' || event.ctrlKey || event.metaKey || event.altKey) return;
	if (playing) {
		playKey(event);
		return;
	}
	if (event.key === '0') fit();
	if (event.key === '1') actualSize();
	if (event.key === 'm') document.body.classList.toggle('no-map');
	if (event.key === '/') {
		event.preventDefault();
		find.focus();
	}
}
window.addEventListener('keydown', onKey);
window.addEventListener('keyup', onKey);
window.addEventListener('blur', () => document.body.classList.remove('alt'));

// ---- Pointing ----

const NO_LOCATIONS = 'Pointing needs the dev server: this build records no source locations.';

/** @param {Document} doc */
const recordsLocations = (doc) =>
	[...doc.querySelectorAll('body *')].slice(0, 200).some((el) => '__svelte_meta' in el);

/**
 * @param {ReturnType<typeof nearestLocation>} found
 * @returns {string}
 */
const describe = (found) => {
	if (!found) return '';
	const at = `${found.path}:${found.line}:${found.column}`;
	// Enclosing components from node_modules or a store are noise, not source.
	const inside = found.parents
		.filter((p) => p.recognised && !p.path.includes('node_modules'))
		.map((p) => `${p.path}:${p.line}`);
	return inside.length ? `${at}  in  ${inside.join('  ‹  ')}` : at;
};

/** @param {string} text */
async function copy(text) {
	try {
		await navigator.clipboard.writeText(text);
		return true;
	} catch {
		return false;
	}
}

/**
 * How much taller the artboard must be for nothing in it to sit behind a
 * scrollbar: the page's own overflow, or the deepest inner scroller's. An app
 * shell fills the viewport and scrolls its content inside, so the page has to
 * be as tall as the content for the shell to show all of it.
 * @param {Document} doc
 * @param {Window} win
 */
function hiddenHeight(doc, win) {
	let excess = Math.max(0, doc.documentElement.scrollHeight - doc.documentElement.clientHeight);
	for (const el of doc.body.querySelectorAll('*')) {
		const over = el.scrollHeight - el.clientHeight;
		if (over <= 1 || over <= excess) continue;
		if (/auto|scroll/.test(win.getComputedStyle(el).overflowY)) excess = over;
	}
	return excess;
}

/**
 * Wire one loaded artboard: measure it, and let it be pointed at.
 * @param {HTMLIFrameElement} iframe
 * @param {HTMLElement} sizeLabel
 * @param {HTMLElement} box
 * @param {boolean} [fixed] keep the frame's own height, as a played page scrolls
 */
function attach(iframe, sizeLabel, box, fixed = false) {
	const doc = iframe.contentDocument;
	const win = iframe.contentWindow;
	if (!doc || !win) return;

	// A story that focuses something as it renders takes the keyboard from the
	// find box, and with it whatever was being typed. While a person is in
	// find, an artboard's focus() does nothing; otherwise stories focus as usual.
	if (!playing) {
		for (const proto of [win.HTMLElement.prototype, win.SVGElement.prototype]) {
			const focus = proto.focus;
			proto.focus = function (...args) {
				if (document.activeElement !== find) focus.apply(this, args);
			};
		}
	}

	if (!fixed) {
		// Content height: the artboard is as tall as its story, never behind a
		// scrollbar. Re-measured whenever the story reflows.
		const measure = () => {
			const height = Math.min(MAX_HEIGHT, iframe.offsetHeight + hiddenHeight(doc, win));
			if (Math.abs(height - iframe.offsetHeight) < 1) return;
			iframe.style.height = `${height}px`;
			sizeLabel.textContent = `${iframe.offsetWidth} × ${height}`;
		};
		measure();
		// The story renders after the frame loads, and it is the body that grows
		// with it: the root element stays as tall as the frame.
		const observer = new win.ResizeObserver(measure);
		observer.observe(doc.documentElement);
		observer.observe(doc.body);
	}

	doc.addEventListener('keydown', onKey);
	doc.addEventListener('keyup', onKey);

	doc.addEventListener('mousemove', (event) => {
		const target = event.altKey ? nearestRecorded(event.target) : null;
		if (!target) {
			box.style.display = 'none';
			return;
		}
		const rect = target.getBoundingClientRect();
		Object.assign(box.style, {
			display: 'block',
			left: `${rect.left}px`,
			top: `${rect.top}px`,
			width: `${rect.width}px`,
			height: `${rect.height}px`
		});
	});
	doc.addEventListener('mouseleave', () => {
		box.style.display = 'none';
	});

	doc.addEventListener(
		'click',
		async (event) => {
			if (!event.altKey) return;
			event.preventDefault();
			event.stopPropagation();
			const found = nearestLocation(event.target);
			if (!found) {
				say(recordsLocations(doc) ? 'No source location recorded for that element.' : NO_LOCATIONS);
				return;
			}
			// The clipboard only ever gets a repo-relative path.
			if (!found.recognised) {
				say(`${describe(found)}  (outside packages/ and apps/, not copied)`);
				return;
			}
			const text = `${found.path}:${found.line}:${found.column}`;
			const copied = await copy(text);
			say(`${describe(found)}  ${copied ? '(copied)' : '(copy failed)'}`);
		},
		true
	);
}

// ---- Artboards ----

/**
 * Every artboard shares this page's one thread, so only the ones near the
 * viewport are live, nearest first, and never more than LIVE_MAX. One that has
 * been left behind gives its place to one that has come near.
 */
const CONCURRENT_LOADS = 1;
const LIVE_MAX = 24;
/** In viewports from the edge of the screen: 0 is on it. */
const NEAR = 0.5;
/** Zoomed out further than this a page cannot be read, so none is started. */
const READABLE_SCALE = 0.15;
let total = 0;
/** @type {Set<HTMLIFrameElement>} */
const waiting = new Set();
/** @type {Set<HTMLIFrameElement>} */
const live = new Set();
let loading = 0;

/** How far an artboard is off screen, in viewports. */
const away = (/** @type {HTMLIFrameElement} */ iframe) => {
	const box = iframe.getBoundingClientRect();
	return Math.max(
		Math.max(0, box.left - innerWidth, -box.right) / innerWidth,
		Math.max(0, box.top - innerHeight, -box.bottom) / innerHeight
	);
};

/** Distance from an artboard's centre to the stage's centre, on screen. */
const distance = (/** @type {HTMLIFrameElement} */ iframe) => {
	const box = iframe.getBoundingClientRect();
	return Math.hypot(
		box.left + box.width / 2 - innerWidth / 2,
		box.top + box.height / 2 - innerHeight / 2
	);
};

/** Give up the live artboard furthest off screen. False when every one is near. */
function unloadFurthest() {
	let furthest = null;
	let most = NEAR;
	for (const iframe of live) {
		const a = away(iframe);
		if (a > most) [furthest, most] = [iframe, a];
	}
	if (!furthest) return false;
	live.delete(furthest);
	waiting.add(furthest);
	furthest.parentElement?.classList.remove('loaded');
	// Its height stays, so nothing on the canvas moves when it goes.
	furthest.src = 'about:blank';
	return true;
}

function report() {
	if (live.size === total) say(`${total} artboards`);
	else if (view.scale < READABLE_SCALE)
		say(`${live.size} of ${total} artboards rendered · zoom in to render more`);
	else say(`${live.size} of ${total} artboards rendered · the rest as they come near`);
}

function pump() {
	// A loading story blocks this thread, and so does a gesture. Loads wait for
	// the design canvas to be still.
	if (world.classList.contains('moving')) return;
	while (loading < CONCURRENT_LOADS && view.scale >= READABLE_SCALE) {
		let next = null;
		let best = Infinity;
		for (const iframe of waiting) {
			if (away(iframe) > NEAR) continue;
			const d = distance(iframe);
			if (d < best) [next, best] = [iframe, d];
		}
		if (!next) break;
		if (live.size + loading >= LIVE_MAX && !unloadFurthest()) break;
		waiting.delete(next);
		loading += 1;
		const iframe = next;
		const done = () => {
			loading -= 1;
			live.add(iframe);
			iframe.parentElement?.classList.add('loaded');
			pump();
		};
		iframe.addEventListener('load', () => setTimeout(done, 250), { once: true });
		iframe.src = /** @type {string} */ (iframe.dataset.src);
	}
	report();
}
window.addEventListener('resize', pump);

/** @param {ReturnType<typeof artboardsFrom>[number]} board */
function artboard(board) {
	const { width, height } = sizeOf(board.size);
	const el = document.createElement('section');
	el.className = 'board';
	el.id = `${board.id}--${board.size}`;
	boardsByStory.set(board.id, [...(boardsByStory.get(board.id) ?? []), el]);

	const label = document.createElement('div');
	label.className = 'label';
	const name = document.createElement('b');
	name.textContent = board.name;
	const size = document.createElement('span');
	size.textContent = `${width} × ${height}`;
	label.append(`${board.title} · `, name, ` · ${board.size} `, size);

	const frame = document.createElement('div');
	frame.className = 'frame';
	const iframe = document.createElement('iframe');
	iframe.title = `${board.title} — ${board.name} at ${board.size}`;
	iframe.width = String(width);
	iframe.height = String(height);
	iframe.dataset.src = storyUrl(board.id, board);

	const box = document.createElement('div');
	box.className = 'pointer';
	const cover = document.createElement('div');
	cover.className = 'cover';
	frame.append(iframe, box, cover);

	iframe.addEventListener('load', () => attach(iframe, size, box));
	waiting.add(iframe);

	el.append(label, frame);
	return el;
}

/**
 * Every screen and story, whichever of them is drawn. A screen's name opens
 * that screen; a story jumps to it, opening its screen first when it is not
 * the one on the canvas.
 * @param {ReturnType<typeof artboardsFrom>} boards every story of the layer
 * @param {string | null} shown the title on the canvas, if one is
 */
function buildSitemap(boards, shown) {
	const query = new URLSearchParams(location.search);
	for (const { layer, titles } of sitemapFrom(boards)) {
		const group = document.createElement('details');
		group.open = true;
		const summary = document.createElement('summary');
		summary.textContent = layer;
		group.append(summary);
		for (const { title, stories } of titles) {
			const section = document.createElement('div');
			section.className = 'entry';
			const heading = document.createElement('a');
			heading.className = 'title';
			heading.href = screenUrl(query, title);
			heading.textContent = title.split('/').slice(1).join('/') || title;
			if (title === shown) heading.setAttribute('aria-current', 'page');
			section.append(heading);
			for (const story of stories) {
				const button = document.createElement('button');
				button.type = 'button';
				button.dataset.id = story.id;
				button.dataset.search = `${title} ${story.name}`.toLowerCase();
				button.textContent = story.name;
				button.addEventListener('click', () => {
					if (boardsByStory.has(story.id)) jumpTo(story.id);
					else location.assign(screenUrl(query, title, story.id));
				});
				section.append(button);
			}
			group.append(section);
		}
		tree.append(group);
	}
}

/**
 * @param {'flow' | 'play'} mode
 * @param {string} name
 */
function flowUrl(mode, name) {
	const query = new URLSearchParams(location.search);
	const next = new URLSearchParams([[mode, name]]);
	for (const key of ['theme', 'density']) {
		const value = query.get(key);
		if (value) next.set(key, value);
	}
	return `${location.pathname}?${next}`;
}

/** @param {Parameters<typeof flowFrom>[2]} index */
function buildFlows(index) {
	const group = document.createElement('details');
	group.open = true;
	const summary = document.createElement('summary');
	summary.textContent = 'Flows';
	group.append(summary);
	for (const [name, flow] of Object.entries(flows)) {
		const section = document.createElement('div');
		section.className = 'entry';
		const heading = document.createElement('div');
		heading.className = 'title';
		heading.textContent = flow.title;
		section.append(heading);
		try {
			flowFrom(flows, name, index);
		} catch (error) {
			const problem = document.createElement('p');
			problem.className = 'problem';
			problem.textContent = error instanceof Error ? error.message : String(error);
			section.append(problem);
			group.append(section);
			continue;
		}
		for (const [mode, label] of /** @type {const} */ ([
			['flow', 'View'],
			['play', 'Play']
		])) {
			const link = document.createElement('a');
			link.href = flowUrl(mode, name);
			link.textContent = label;
			section.append(link);
		}
		group.append(section);
	}
	tree.prepend(group);
}

// ---- Screen view ----

/**
 * One screen: a row for each kind of story it has, what a member sees first and
 * the stories that only assert something last.
 * @param {ReturnType<typeof artboardsFrom>} boards
 */
function drawScreen(boards) {
	const title = document.createElement('h1');
	title.className = 'screen';
	title.textContent = boards[0]?.title ?? '';
	world.append(title);
	for (const [kind, label] of Object.entries(KINDS)) {
		const mine = boards.filter((b) => b.kind === kind);
		if (mine.length === 0) continue;
		const row = document.createElement('section');
		row.className = 'row';
		const heading = document.createElement('h2');
		heading.textContent = label;
		const strip = document.createElement('div');
		strip.className = 'boards';
		strip.append(...mine.map(artboard));
		row.append(heading, strip);
		world.append(row);
	}
}

// ---- Flow view ----

/**
 * The happy path runs left to right. Under each step hang the ways off it: what
 * happened there, and the state that leaves the member in.
 * @param {ReturnType<typeof flowFrom>} flow
 * @param {ReturnType<typeof flowBoards>} boards
 * @param {ReturnType<typeof branchBoards>} branches
 */
function drawFlow(flow, boards, branches) {
	const row = document.createElement('section');
	row.className = 'row';
	const heading = document.createElement('h2');
	heading.textContent = flow.title;
	const strip = document.createElement('div');
	strip.className = 'boards';
	row.append(heading, strip);
	world.append(row);
	const labels = connectorLabels(flow);
	for (const [i, id] of flow.steps.entries()) {
		const step = document.createElement('div');
		step.className = 'step';
		const happy = document.createElement('div');
		happy.className = 'boards';
		happy.append(...boards.filter((b) => b.step === i).map(artboard));
		step.append(happy);
		for (const branch of branches.filter((b) => b.from === id)) {
			const off = document.createElement('div');
			off.className = 'branch';
			const outcome = document.createElement('div');
			outcome.className = 'outcome';
			outcome.textContent = branch.outcome;
			const landed = document.createElement('div');
			landed.className = 'boards';
			landed.append(...branch.boards.map(artboard));
			off.append(outcome, landed);
			step.append(off);
		}
		strip.append(step);
		if (i < flow.steps.length - 1) {
			const connector = document.createElement('div');
			connector.className = 'connector';
			connector.textContent = labels[i] || '';
			connector.setAttribute('aria-hidden', 'true');
			strip.append(connector);
		}
	}
}

// ---- Play mode: one live artboard that follows the flow ----

/** @type {{ name: string, flow: ReturnType<typeof flowFrom>, index: any, id: string, size: string, theme: string, density: string, depth: number, iframe: HTMLIFrameElement } | null} */
let player = null;

/** @param {string} id */
const nameOf = (id) => {
	const entry = player?.index.entries[id];
	return entry ? `${entry.title} · ${entry.name}` : id;
};

function playStatus() {
	if (!player) return;
	const at = player.flow.steps.indexOf(player.id);
	const place = at < 0 ? 'side path' : `step ${at + 1} of ${player.flow.steps.length}`;
	say(`${player.flow.title} · ${place} · ${nameOf(player.id)} · fixture flow`);
}

/** @param {string} id */
function show(id) {
	if (!player) return;
	player.id = id;
	const { width } = sizeOf(player.size);
	const bar = document.getElementById('bar')?.offsetHeight ?? 0;
	player.iframe.width = String(width);
	player.iframe.height = String(Math.max(200, innerHeight - bar - 32));
	const heard = callbacksFor(player.flow, id);
	const url = storyUrl(id, player);
	const target = `${url}&play=${heard.join(',')}`;
	// Replace rather than assign: an assigned src adds a history entry of its
	// own, and Back would then undo the iframe instead of the step.
	if (player.iframe.src) player.iframe.contentWindow?.location.replace(target);
	else player.iframe.src = target;
	playStatus();
}

/** @param {string} id */
function go(id) {
	if (!player) return;
	player.depth += 1;
	history.pushState({ depth: player.depth }, '', `#${id}`);
	show(id);
}

/** @param {KeyboardEvent} event */
function playKey(event) {
	if (!player) return;
	if (event.key === 'Escape') {
		location.href = flowUrl('flow', player.name);
		return;
	}
	if (event.key === '[' && player.depth > 0) history.back();
	const size = { c: 'compact', m: 'medium', w: 'wide', u: 'ultra' }[event.key];
	if (size && size !== player.size) {
		player.size = size;
		show(player.id);
	}
}

/** @param {HTMLIFrameElement} iframe */
function attachPlay(iframe) {
	const doc = iframe.contentDocument;
	if (!doc || !player) return;
	const frame = /** @type {HTMLElement} */ (iframe.parentElement);
	frame.classList.add('loaded');
	attach(
		iframe,
		document.createElement('span'),
		/** @type {HTMLElement} */ (frame.querySelector('.pointer')),
		true
	);
	doc.addEventListener(
		'click',
		(event) => {
			if (event.altKey || !player) return;
			const link = /** @type {Element | null} */ (event.target)?.closest?.('a[href]');
			if (!link) return;
			event.preventDefault();
			const href = /** @type {HTMLAnchorElement} */ (link).getAttribute('href') ?? '';
			const to = routeFor(
				routes,
				/** @type {HTMLAnchorElement} */ (link).href,
				doc.location.origin
			);
			if (to) go(to);
			else say(`No route for ${href}  ·  ${nameOf(player.id)}`);
		},
		true
	);
}

/**
 * @param {string} name
 * @param {ReturnType<typeof flowFrom>} flow
 * @param {any} index
 * @param {URLSearchParams} query
 */
function play(name, flow, index, query) {
	const size = query.get('size') ?? 'wide';
	if (!(size in viewportOptions)) {
		say(`Unknown size "${size}". Sizes are ${Object.keys(viewportOptions).join(', ')}.`);
		return;
	}
	playing = true;
	document.body.classList.add('play', 'no-map');
	const hint = document.getElementById('hint');
	if (hint) {
		hint.textContent = 'Click through · [ back · Esc flow view · C M W U size · Alt-click points';
	}
	const frame = document.createElement('div');
	frame.className = 'frame';
	const iframe = document.createElement('iframe');
	iframe.title = `${flow.title}, played`;
	const box = document.createElement('div');
	box.className = 'pointer';
	frame.append(iframe, box);
	const holder = document.createElement('div');
	holder.id = 'player';
	holder.append(frame);
	stage.append(holder);

	const wanted = decodeURIComponent(location.hash.slice(1));
	const start = index.entries[wanted]?.type === 'story' ? wanted : flow.steps[0];
	player = {
		name,
		flow,
		index,
		id: start,
		size,
		theme: query.get('theme') ?? 'light',
		density: query.get('density') ?? 'comfortable',
		depth: 0,
		iframe
	};
	history.replaceState({ depth: 0 }, '', `#${start}`);
	iframe.addEventListener('load', () => attachPlay(iframe));
	window.addEventListener('popstate', (event) => {
		if (!player) return;
		player.depth = event.state?.depth ?? 0;
		const id = decodeURIComponent(location.hash.slice(1));
		if (index.entries[id]?.type === 'story') show(id);
	});
	window.addEventListener('message', (event) => {
		if (!player || event.origin !== location.origin) return;
		if (event.source !== player.iframe.contentWindow) return;
		if (event.data?.type !== 'wimm-canvas-callback') return;
		const to = nextStory(player.flow, player.id, event.data.name);
		if (to) go(to);
		else say(`No transition for ${event.data.name} on ${nameOf(player.id)}`);
	});
	show(start);
}

find.addEventListener('input', () => {
	const words = find.value.toLowerCase().split(/\s+/).filter(Boolean);
	for (const button of tree.querySelectorAll('button')) {
		const hit = words.every((w) => (button.dataset.search ?? '').includes(w));
		button.hidden = !hit;
	}
	for (const section of tree.querySelectorAll('.entry')) {
		section.toggleAttribute('hidden', !section.querySelector('button:not([hidden])'));
	}
	for (const group of tree.querySelectorAll('details')) {
		group.toggleAttribute('hidden', !group.querySelector('.entry:not([hidden])'));
	}
});
// A story that autofocuses pulls focus into its iframe as it renders. A
// person cannot click into an artboard (the cover takes the click), so focus
// landing in an iframe while browsing is never theirs: hand it back.
find.addEventListener('blur', () => {
	setTimeout(() => {
		if (!playing && document.activeElement?.tagName === 'IFRAME') {
			find.focus({ preventScroll: true });
		}
	});
});
find.addEventListener('keydown', (event) => {
	if (event.key === 'Escape') find.blur();
	if (event.key !== 'Enter') return;
	const first = /** @type {HTMLElement | null} */ (tree.querySelector('button:not([hidden])'));
	first?.click();
});

async function main() {
	let boards;
	let index;
	/** @type {ReturnType<typeof artboardsFrom>} */
	let every;
	/** @type {ReturnType<typeof branchBoards>} */
	let branches = [];
	const query = new URLSearchParams(location.search);
	/** @type {ReturnType<typeof flowFrom> | null} */
	let flow = null;
	try {
		index = await (await fetch('/index.json')).json();
		const name = query.get('play') ?? query.get('flow');
		if (name) {
			flow = flowFrom(flows, name, index);
			if (query.has('play')) {
				play(name, flow, index, query);
				return;
			}
			boards = flowBoards(index, flow, query);
			branches = branchBoards(index, flow, query);
		} else {
			// One screen at a time: every artboard shares this page's one thread,
			// and a canvas of all of them takes minutes to be looked at.
			boards = query.has('title') ? artboardsFrom(index, query) : [];
		}
		const layer = new URLSearchParams(query);
		layer.delete('title');
		every = artboardsFrom(index, layer);
	} catch (error) {
		say(error instanceof Error ? error.message : String(error));
		return;
	}
	buildSitemap(every, flow ? null : query.get('title'));
	buildFlows(index);
	if (boards.length === 0) {
		say(
			query.has('title') ? 'No stories match this view.' : 'Choose a screen, or find one with /.'
		);
		return;
	}

	if (flow) drawFlow(flow, boards, branches);
	else drawScreen(boards);
	total = boards.length + branches.reduce((n, b) => n + b.boards.length, 0);
	const wanted = decodeURIComponent(location.hash.slice(1));
	if (boardsByStory.has(wanted)) jumpTo(wanted);
	else actualSize();
	pump();
}

main();
