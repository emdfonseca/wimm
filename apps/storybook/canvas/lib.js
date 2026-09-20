import { viewportOptions } from './viewports.js';

/** @param {URLSearchParams} query */
function viewFrom(query) {
	const theme = query.get('theme') ?? 'light';
	const density = query.get('density') ?? 'comfortable';
	const sizes = (query.get('sizes') ?? 'compact,wide').split(',');
	for (const size of sizes) {
		if (!(size in viewportOptions)) {
			throw new Error(
				`Unknown size "${size}". Sizes are ${Object.keys(viewportOptions).join(', ')}.`
			);
		}
	}
	return { theme, density, sizes };
}

/**
 * What a page story is for, and the row a screen's view draws it in, in this
 * order. A story says which with one tag, `kind-state` and so on:
 *
 *   state      what a member sees at rest, empty and mid-interaction included
 *   waiting    something is in progress
 *   error      this screen's own work failed, and it says so
 *   outcome    the screen on arrival, because of what happened somewhere else
 *   behaviour  it exists to assert a promise, and shows nothing a state does not
 */
export const KINDS = /** @type {const} */ ({
	state: 'States',
	waiting: 'Waiting',
	error: 'Errors',
	outcome: 'Outcomes',
	behaviour: 'Behaviour'
});

const KIND_TAG = 'kind-';

/** @param {{ tags?: string[] }} entry */
const kindTags = (entry) => (entry.tags ?? []).filter((t) => t.startsWith(KIND_TAG));

/**
 * @param {{ tags?: string[] }} entry
 * @returns {keyof typeof KINDS | undefined} undefined when it carries no kind the canvas knows
 */
export function kindOf(entry) {
	const kind = kindTags(entry)[0]?.slice(KIND_TAG.length);
	return kind && kind in KINDS ? /** @type {keyof typeof KINDS} */ (kind) : undefined;
}

/**
 * Whether every page story says what it is for, once, in a word the canvas
 * draws a row for. A story without one would be drawn in no row at all.
 * @param {{ id: string, tags?: string[] }[]} pages
 * @returns {string[]} what is wrong, empty when nothing is
 */
export function kindProblems(pages) {
	const problems = [];
	for (const entry of pages) {
		const tags = kindTags(entry);
		if (tags.length === 0) problems.push(`"${entry.id}" has no kind tag`);
		if (tags.length > 1)
			problems.push(`"${entry.id}" has more than one kind tag: ${tags.join(', ')}`);
		for (const tag of tags) {
			if (!(tag.slice(KIND_TAG.length) in KINDS)) {
				problems.push(
					`"${entry.id}" has the unknown kind tag "${tag}"; kinds are ${Object.keys(KINDS).join(', ')}`
				);
			}
		}
	}
	return problems;
}

/**
 * @param {{ entries: Record<string, { id: string, type: string, title: string, name: string, tags?: string[] }> }} index
 * @param {URLSearchParams} query
 */
export function artboardsFrom(index, query) {
	const title = query.get('title');
	const layer = (query.get('layer') ?? 'pages').toLowerCase();
	const { theme, density, sizes } = viewFrom(query);
	return Object.values(index.entries)
		.filter(
			(e) =>
				e.type === 'story' &&
				(title ? e.title === title : e.title.split('/')[0]?.toLowerCase() === layer)
		)
		.flatMap((e) =>
			sizes.map((size) => ({
				id: e.id,
				title: e.title,
				name: e.name,
				kind: kindOf(e),
				size,
				theme,
				density
			}))
		);
}

/**
 * The address of one screen's view, keeping how the canvas is being looked at.
 * @param {URLSearchParams} query
 * @param {string} title
 * @param {string} [id] the story to land on
 */
export function screenUrl(query, title, id) {
	const next = new URLSearchParams([['title', title]]);
	for (const key of ['theme', 'density', 'sizes']) {
		const value = query.get(key);
		if (value) next.set(key, value);
	}
	return `?${next}${id ? `#${encodeURIComponent(id)}` : ''}`;
}

/**
 * A flow's steps as artboards, in flow order rather than index order.
 * @param {{ entries: Record<string, { id: string, type: string, title: string, name: string }> }} index
 * @param {Flow} flow
 * @param {URLSearchParams} query
 */
export function flowBoards(index, flow, query) {
	const { theme, density, sizes } = viewFrom(query);
	return flow.steps.flatMap((id) => {
		const e = index.entries[id];
		return sizes.map((size) => ({ id, title: e.title, name: e.name, size, theme, density }));
	});
}

/**
 * Where a flow leaves its happy path: for each branch, the step it leaves, what
 * happened, and the state it lands in as artboards at every size.
 * @param {{ entries: Record<string, { id: string, type: string, title: string, name: string }> }} index
 * @param {Flow} flow
 * @param {URLSearchParams} query
 */
export function branchBoards(index, flow, query) {
	const { theme, density, sizes } = viewFrom(query);
	return (flow.branches ?? []).map(({ from, outcome, to }) => {
		const e = index.entries[to];
		return {
			from,
			outcome,
			boards: sizes.map((size) => ({ id: to, title: e.title, name: e.name, size, theme, density }))
		};
	});
}

/**
 * What moves the member from each step to the next, one label per gap.
 * @param {Flow} flow
 */
export function connectorLabels(flow) {
	return flow.steps
		.slice(1)
		.map(
			(to, i) => flow.transitions.find((t) => t.from === flow.steps[i] && t.to === to)?.on ?? ''
		);
}

/**
 * @param {string} id
 * @param {{ theme: string, density: string, size: string }} globals
 *
 * The accessibility scan is off on the canvas: it is the dearest part of a
 * render, and `just check` runs it on every story already.
 */
export function storyUrl(id, { theme, density, size }) {
	return `/iframe.html?id=${encodeURIComponent(id)}&viewMode=story&globals=theme:${theme};density:${density};viewport.value:${size};a11y.manual:!true`;
}

/** @param {string} path */
export function relativeToRepo(path) {
	const cut = Math.max(path.lastIndexOf('/packages/'), path.lastIndexOf('/apps/'));
	return cut < 0 ? { path, recognised: false } : { path: path.slice(cut + 1), recognised: true };
}

const MAX_PARENTS = 3;

/** @param {{ file: string, line: number, column: number }} at */
const place = (at) => ({ ...relativeToRepo(at.file), line: at.line, column: at.column });

/**
 * The element pointing at `node` would report: itself or the nearest ancestor
 * that Svelte's dev build recorded a source location on.
 * @param {any} node
 */
export function nearestRecorded(node) {
	for (let at = node; at; at = at.parentElement) {
		if (at.__svelte_meta) return at;
	}
	return null;
}

/** @param {any} node */
export function nearestLocation(node) {
	const meta = nearestRecorded(node)?.__svelte_meta;
	if (!meta) return null;
	const parents = [];
	for (let p = meta.parent; p && parents.length < MAX_PARENTS; p = p.parent) {
		parents.push(place(p));
	}
	return { ...place(meta.loc), parents };
}
/**
 * The tree the sitemap panel shows: layer, then title, then story. A story
 * appears once however many sizes it is drawn at.
 * @param {{ id: string, title: string, name: string }[]} boards
 */
export function sitemapFrom(boards) {
	/** @type {Map<string, Map<string, Map<string, string>>>} */
	const layers = new Map();
	for (const { id, title, name } of boards) {
		const layer = title.split('/')[0] ?? title;
		const titles = layers.get(layer) ?? new Map();
		const stories = titles.get(title) ?? new Map();
		stories.set(id, name);
		titles.set(title, stories);
		layers.set(layer, titles);
	}
	return [...layers].map(([layer, titles]) => ({
		layer,
		titles: [...titles].map(([title, stories]) => ({
			title,
			stories: [...stories].map(([id, name]) => ({ id, name }))
		}))
	}));
}

/**
 * @typedef {{ from: string, on: string, to: string }} Transition
 * @typedef {{ from: string, outcome: string, to: string }} Branch a way off the happy path: what happened at a step, and the state it leaves the member in
 * @typedef {{ title: string, steps: string[], transitions: Transition[], branches?: Branch[] }} Flow
 */

/**
 * A declared flow, or an error naming everything wrong with it at once.
 * @param {Record<string, Flow>} flows
 * @param {string} name
 * @param {{ entries: Record<string, { type: string }> }} index
 * @returns {Flow}
 */
export function flowFrom(flows, name, index) {
	const flow = flows[name];
	if (!flow) {
		throw new Error(`Unknown flow "${name}". Flows are ${Object.keys(flows).join(', ')}.`);
	}
	const problems = [];
	/** @param {string} id */
	const unknown = (id) => index.entries[id]?.type !== 'story';
	const branches = flow.branches ?? [];
	const ids = [
		...flow.steps,
		...flow.transitions.flatMap((t) => [t.from, t.to]),
		...branches.map((b) => b.to)
	];
	for (const id of new Set(ids)) {
		if (unknown(id)) problems.push(`no story "${id}"`);
	}
	for (const t of flow.transitions) {
		if (!t.on) problems.push(`transition from "${t.from}" to "${t.to}" names no callback`);
	}
	for (const b of branches) {
		// A branch is drawn under the step it leaves, so it has to leave one.
		if (!flow.steps.includes(b.from))
			problems.push(`branch to "${b.to}" leaves "${b.from}", which is not a step`);
		if (flow.steps.includes(b.to))
			problems.push(`branch to "${b.to}" lands on a step of the happy path`);
		if (!b.outcome) problems.push(`branch from "${b.from}" to "${b.to}" names no outcome`);
	}
	if (problems.length > 0) {
		throw new Error(`Flow "${name}": ${problems.join('; ')}.`);
	}
	return flow;
}

/**
 * Whether every page story that shows something has been put somewhere: on a
 * flow, or on the list of ones nobody has placed yet. A new story fails this
 * until someone decides. A behaviour story shows nothing a state does not, so
 * it belongs on neither.
 * @param {Record<string, Flow>} flows
 * @param {string[]} unplaced
 * @param {{ id: string, tags?: string[] }[]} pages every page story there is
 * @returns {string[]} what is wrong, empty when nothing is
 */
export function placementProblems(flows, unplaced, pages) {
	const placed = new Set(
		Object.values(flows).flatMap((f) => [...f.steps, ...(f.branches ?? []).map((b) => b.to)])
	);
	const listed = new Set(unplaced);
	const known = new Set(pages.map((e) => e.id));
	const problems = [];
	for (const entry of pages) {
		const { id } = entry;
		if (kindOf(entry) === 'behaviour') {
			if (placed.has(id)) problems.push(`"${id}" only asserts something and is on a flow`);
			if (listed.has(id)) problems.push(`"${id}" only asserts something and is listed as unplaced`);
		} else if (!placed.has(id) && !listed.has(id)) {
			problems.push(`"${id}" is on no flow and not listed as unplaced`);
		}
	}
	for (const id of listed) {
		if (placed.has(id)) problems.push(`"${id}" is on a flow and still listed as unplaced`);
		if (!known.has(id)) problems.push(`unplaced "${id}" is not a page story`);
	}
	return problems;
}

/**
 * @param {Flow} flow
 * @param {string} storyId
 * @param {string} callback
 * @returns {string | undefined}
 */
export function nextStory(flow, storyId, callback) {
	return flow.transitions.find((t) => t.from === storyId && t.on === callback)?.to;
}

/**
 * The callbacks play mode must listen for on one step.
 * @param {Flow} flow
 * @param {string} storyId
 */
export function callbacksFor(flow, storyId) {
	return [...new Set(flow.transitions.filter((t) => t.from === storyId).map((t) => t.on))];
}

/**
 * The story an href stands for while playing. Query and hash are ignored; an
 * absolute href counts only when it is on `origin`.
 * @param {Record<string, string>} routes
 * @param {string} href
 * @param {string} [origin]
 * @returns {string | undefined}
 */
export function routeFor(routes, href, origin = 'http://canvas.invalid') {
	const url = new URL(href, origin);
	if (url.origin !== origin) return undefined;
	return routes[url.pathname];
}
