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
const SIZE_TAG = 'size-';

/** @param {{ tags?: string[] }} entry */
const sizeTags = (entry) => (entry.tags ?? []).filter((t) => t.startsWith(SIZE_TAG));

/**
 * The sizes one story is drawn at. A story that pins its own viewport says so
 * with a tag, `size-compact`, and is drawn at that size whatever the view asks
 * for: its compact form stretched across a wide artboard is a page no member
 * sees.
 * @param {{ tags?: string[] }} entry
 * @param {string[]} sizes the sizes the view asked for
 */
export function sizesFor(entry, sizes) {
	const own = sizeTags(entry)
		.map((t) => t.slice(SIZE_TAG.length))
		.filter((size) => size in viewportOptions);
	return own.length > 0 ? own : sizes;
}

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
		for (const tag of sizeTags(entry)) {
			if (!(tag.slice(SIZE_TAG.length) in viewportOptions)) {
				problems.push(
					`"${entry.id}" has the unknown size tag "${tag}"; sizes are ${Object.keys(viewportOptions).join(', ')}`
				);
			}
		}
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
			sizesFor(e, sizes).map((size) => ({
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
	for (const key of ['theme', 'density', 'sizes', 'data', 'needs']) {
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
	return flow.steps.flatMap((id, step) => {
		const e = index.entries[id];
		return sizesFor(e, sizes).map((size) => ({
			id,
			step,
			title: e.title,
			name: e.name,
			size,
			theme,
			density
		}));
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
			boards: sizesFor(e, sizes).map((size) => ({
				id: to,
				title: e.title,
				name: e.name,
				size,
				theme,
				density
			}))
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

const ID_LISTS = new Set(['aria-labelledby', 'aria-describedby', 'aria-controls']);
const HARNESS_ATTRIBUTES = ['data-storybook', 'data-vitest'];
const TAG = /<[^>]*>/g;
const ATTRIBUTE = /([^\s=/>"']+)(?:=(?:"([^"]*)"|'([^']*)'|([^\s>"']+)))?/g;

/**
 * One tag with its attributes in name order, harness attributes removed and
 * every id it names replaced by `ids`' number for it.
 * @param {string} tag
 * @param {Map<string, string>} ids
 */
function normaliseTag(tag, ids) {
	if (tag.startsWith('</') || tag.startsWith('<!')) return tag;
	const name = /^<([^\s/>]+)/.exec(tag)?.[1];
	if (!name) return tag;
	const rest = tag.slice(name.length + 1).replace(/\/?>$/, '');
	const selfClosing = tag.endsWith('/>');
	/** @param {string} id */
	const numbered = (id) => {
		if (!ids.has(id)) ids.set(id, `id-${ids.size + 1}`);
		return ids.get(id);
	};
	/** @type {[string, string | undefined][]} */
	const attributes = [];
	for (const [, attr, dq, sq, bare] of rest.matchAll(ATTRIBUTE)) {
		if (HARNESS_ATTRIBUTES.some((prefix) => attr.startsWith(prefix))) continue;
		attributes.push([attr, dq ?? sq ?? bare]);
	}
	attributes.sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0));
	const written = attributes.map(([attr, value]) => {
		if (value === undefined) return attr;
		let out = value;
		if (attr === 'id' || attr === 'for') out = numbered(value);
		else if (ID_LISTS.has(attr)) out = value.split(/\s+/).filter(Boolean).map(numbered).join(' ');
		else if (attr === 'href' && value.startsWith('#')) out = `#${numbered(value.slice(1))}`;
		return `${attr}="${out}"`;
	});
	return `<${[name, ...written].join(' ')}${selfClosing ? '/' : ''}>`;
}

/**
 * The markup of a rendered story with everything that varies from run to run
 * taken out and everything a person can see kept: scoped classes, inline style
 * and every word. Generated ids are numbered by first appearance so the wiring
 * between a label and its input survives and the counter does not.
 * @param {string} html
 * @returns {string}
 */
export function normaliseMarkup(html) {
	const bare = html.replace(/<!--(?:\[!?|\]|)-->/g, '');
	/** @type {Map<string, string>} */
	const ids = new Map();
	let out = '';
	let last = 0;
	for (const match of bare.matchAll(TAG)) {
		out += bare.slice(last, match.index).replace(/\s+/g, ' ');
		out += normaliseTag(match[0], ids);
		last = match.index + match[0].length;
	}
	out += bare.slice(last).replace(/\s+/g, ' ');
	return out.replace(/>\s+</g, '><').trim();
}

/** Where the headless run leaves one `{ id, digest }` line per page story, then `{ run }`. */
export const RUN_FILE = 'node_modules/.cache/wimm-canvas/run.jsonl';

/**
 * @param {string} text
 * @returns {Promise<string>} 64 lower-case hex characters
 */
export async function sha256Hex(text) {
	const bytes = await globalThis.crypto.subtle.digest('SHA-256', new TextEncoder().encode(text));
	return [...new Uint8Array(bytes)].map((b) => b.toString(16).padStart(2, '0')).join('');
}

export const HEX64 = /^[0-9a-f]{64}$/;

/**
 * @param {string} text JSON Lines
 * @returns {{ line: number, value: any }[]} the lines that parse, numbered from 1
 */
function jsonLines(text) {
	return text
		.split('\n')
		.map((raw, i) => ({ raw, line: i + 1 }))
		.filter(({ raw }) => raw.trim() !== '')
		.flatMap(({ raw, line }) => {
			try {
				return [{ line, value: JSON.parse(raw) }];
			} catch {
				return [{ line, value: undefined }];
			}
		});
}

/**
 * What the headless run left, held against the page stories that exist: a clean
 * result, exactly one record per page story, and a real digest in each.
 * @param {string} text the run file
 * @param {string[]} pageIds every page story in Storybook's index
 * @returns {{ problems: string[], digests: Map<string, string> }}
 */
export function runProblems(text, pageIds) {
	const problems = [];
	/** @type {Map<string, string>} */
	const digests = new Map();
	const seen = new Set();
	const known = new Set(pageIds);
	const results = [];
	for (const { line, value } of jsonLines(text)) {
		if (value === undefined || typeof value !== 'object' || value === null) {
			problems.push(`run file line ${line} is not a record`);
		} else if ('run' in value) {
			results.push(value.run);
		} else if (typeof value.id !== 'string') {
			problems.push(`run file line ${line} names no story`);
		} else if (!known.has(value.id)) {
			problems.push(`"${value.id}" is recorded but is not a page story`);
		} else if (seen.has(value.id)) {
			problems.push(`"${value.id}" is recorded more than once`);
		} else {
			seen.add(value.id);
			if (typeof value.digest === 'string' && HEX64.test(value.digest)) {
				digests.set(value.id, value.digest);
			} else {
				problems.push(`"${value.id}" has a digest that is not 64 hex characters`);
			}
		}
	}
	for (const id of pageIds) {
		if (!seen.has(id)) problems.push(`"${id}" has no record in the run`);
	}
	if (results.length !== 1) problems.push('the run did not pass (it left no result)');
	else if (results[0] !== 'passed') problems.push('the run did not pass (it is marked failed)');
	return { problems, digests };
}

/**
 * @typedef {{ fingerprint: string, firstSeen: string }} Version
 * @typedef {{ implemented?: string, gone?: true, versions: Version[] }} StoryVersions
 * @typedef {Record<string, StoryVersions>} Versions
 */

/**
 * The versions file after a regenerate. A fingerprint that did not move changes
 * nothing, which is what makes regeneration a no-op. A version that has been
 * replaced survives only when an approval names it.
 * @param {{
 *   previous: Versions,
 *   fingerprints: Map<string, string>,
 *   approvals: { story: string, fingerprint: string }[],
 *   today: string
 * }} input
 * @returns {{ versions: Versions, moved: string[], added: string[], removed: string[] }}
 */
export function nextVersions({ previous, fingerprints, approvals, today }) {
	const approved = new Set(approvals.map((a) => `${a.story}\n${a.fingerprint}`));
	/** @param {string} id @param {Version[]} list */
	const approvedOnly = (id, list) => list.filter((v) => approved.has(`${id}\n${v.fingerprint}`));
	/** @type {Versions} */
	const next = {};
	const moved = [];
	const added = [];
	const removed = [];
	for (const id of [...new Set([...Object.keys(previous), ...fingerprints.keys()])].sort()) {
		const before = previous[id];
		const fingerprint = fingerprints.get(id);
		if (fingerprint === undefined) {
			const kept = approvedOnly(id, before?.versions ?? []);
			if (kept.length > 0) next[id] = { gone: true, versions: kept };
			else removed.push(id);
		} else if (before?.implemented === fingerprint) {
			next[id] = before;
		} else {
			const others = (before?.versions ?? []).filter((v) => v.fingerprint !== fingerprint);
			const own = (before?.versions ?? []).find((v) => v.fingerprint === fingerprint);
			next[id] = {
				implemented: fingerprint,
				versions: [
					own ?? { fingerprint, firstSeen: today },
					...approvedOnly(id, others)
				]
			};
			(before && !before.gone ? moved : added).push(id);
		}
	}
	return { versions: next, moved, added, removed };
}

/** @param {Versions} versions */
export const serialiseVersions = (versions) => `${JSON.stringify(versions, null, '\t')}\n`;

/** The fields of one approval, and no others. `note` may be empty. */
export const APPROVAL_FIELDS = ['story', 'fingerprint', 'name', 'email', 'at', 'note'];

const ISO_UTC = /^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$/;

/** @param {string} text the approvals record */
export const approvalLines = (text) => {
	const lines = text.split('\n');
	if (lines.at(-1) === '') lines.pop();
	return lines;
};

/**
 * What is wrong with the approvals record, each problem naming its line.
 * @param {string[]} lines
 * @param {Versions} versions
 * @returns {string[]}
 */
export function approvalProblems(lines, versions) {
	const problems = [];
	/** @type {Map<string, number>} */
	const seen = new Map();
	lines.forEach((raw, i) => {
		const n = i + 1;
		/** @param {string} why */
		const bad = (why) => problems.push(`line ${n}: ${why}`);
		/** @type {any} */
		let record;
		try {
			record = JSON.parse(raw);
		} catch {
			return bad('not valid JSON');
		}
		if (typeof record !== 'object' || record === null || Array.isArray(record)) {
			return bad('not a record');
		}
		let shaped = true;
		for (const key of Object.keys(record)) {
			if (!APPROVAL_FIELDS.includes(key)) {
				bad(`unknown field "${key}"`);
				shaped = false;
			}
		}
		for (const key of APPROVAL_FIELDS) {
			if (typeof record[key] !== 'string') {
				bad(`missing field "${key}"`);
				shaped = false;
			}
		}
		if (!shaped) return;
		if (!HEX64.test(record.fingerprint)) bad('fingerprint is not 64 hex characters');
		if (!ISO_UTC.test(record.at) || Number.isNaN(Date.parse(record.at))) {
			bad('time is not an ISO 8601 UTC time');
		}
		const story = versions[record.story];
		if (!story) {
			bad(`story "${record.story}" is not in versions.json`);
		} else if (
			HEX64.test(record.fingerprint) &&
			!story.versions.some((v) => v.fingerprint === record.fingerprint)
		) {
			bad(`story "${record.story}" never had version ${record.fingerprint.slice(0, 7)}`);
		}
		const key = `${record.story}\n${record.fingerprint}\n${record.email}`;
		if (seen.has(key)) bad(`${record.email} already approved this version on line ${seen.get(key)}`);
		else seen.set(key, n);
	});
	return problems;
}

/**
 * Whether `before` is still the start of `after`, line for line. Lines added at
 * the end are the only change an approvals record may take.
 * @param {string[]} before
 * @param {string[]} after
 * @param {{ changed: string, missing: string }} words how each problem is put
 * @returns {string[]}
 */
export function prefixProblems(before, after, words) {
	const problems = [];
	before.forEach((line, i) => {
		if (after[i] === undefined) problems.push(`line ${i + 1}: ${words.missing}`);
		else if (after[i] !== line) problems.push(`line ${i + 1}: ${words.changed}`);
	});
	return problems;
}

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

/** @param {string} value `YYYY-MM-DD` or an ISO time; written `20 Sep 2026` */
export function formatDate(value) {
	const [year, month, day] = value.slice(0, 10).split('-').map(Number);
	return `${day} ${MONTHS[month - 1]} ${year}`;
}

/** @param {string} fingerprint the seven characters a person sees */
export const shortVersion = (fingerprint) => fingerprint.slice(0, 7);

/**
 * The story ids nearest to one a person typed, for a typo.
 * @param {string} wanted
 * @param {string[]} ids
 * @param {number} [count]
 */
export function closestIds(wanted, ids, count = 3) {
	/** @param {string} a @param {string} b */
	const distance = (a, b) => {
		let row = Array.from({ length: b.length + 1 }, (_, i) => i);
		for (let i = 1; i <= a.length; i++) {
			const next = [i];
			for (let j = 1; j <= b.length; j++) {
				next[j] = Math.min(row[j] + 1, next[j - 1] + 1, row[j - 1] + (a[i - 1] === b[j - 1] ? 0 : 1));
			}
			row = next;
		}
		return row[b.length];
	};
	return ids
		.map((id) => ({ id, d: distance(wanted, id) }))
		.filter(({ d }) => d <= Math.max(3, Math.floor(wanted.length / 3)))
		.sort((a, b) => a.d - b.d || (a.id < b.id ? -1 : 1))
		.slice(0, count)
		.map(({ id }) => id);
}

/**
 * @typedef {{ story: string, fingerprint: string, name: string, email: string, at: string, note: string }} Approval
 * @typedef {'approved' | 'changed' | 'never' | 'exempt' | 'unversioned'} Status
 */

/**
 * The approvals that can be read, in the order written. A line that cannot be
 * read is left out here; the check is what refuses it.
 * @param {string} text
 * @returns {Approval[]}
 */
export function parseApprovals(text) {
	return approvalLines(text).flatMap((raw) => {
		try {
			const record = JSON.parse(raw);
			return typeof record?.story === 'string' && typeof record?.fingerprint === 'string'
				? [record]
				: [];
		} catch {
			return [];
		}
	});
}

/** @param {Approval[]} approvals @param {string} fingerprint */
const approvalsOf = (approvals, fingerprint) =>
	approvals.filter((a) => a.fingerprint === fingerprint).sort((a, b) => (a.at < b.at ? -1 : a.at > b.at ? 1 : 0));

/**
 * Both versions travel together: the one implemented now, always, and the newest
 * one anybody approved, when anybody did.
 * @param {StoryVersions | undefined} entry
 * @param {Approval[]} approvals the approvals of this story
 * @param {string | undefined} kind
 * @returns {{
 *   status: Status,
 *   implemented?: Version,
 *   approved?: Version & { approvals: Approval[] }
 * }}
 */
export function statusOf(entry, approvals, kind) {
	if (!entry?.implemented) return { status: 'unversioned' };
	const implemented = entry.versions.find((v) => v.fingerprint === entry.implemented);
	const latest = entry.versions.find((v) => approvalsOf(approvals, v.fingerprint).length > 0);
	const approved = latest
		? { ...latest, approvals: approvalsOf(approvals, latest.fingerprint) }
		: undefined;
	if (kind === 'behaviour') return { status: 'exempt', implemented };
	if (!approved) return { status: 'never', implemented };
	return {
		status: approved.fingerprint === entry.implemented ? 'approved' : 'changed',
		implemented,
		approved
	};
}

/** @param {string} status */
export const needsApproval = (status) => status === 'changed' || status === 'never';

/** @param {Version} version */
export const versionWords = (version) =>
	`${shortVersion(version.fingerprint)} · ${formatDate(version.firstSeen)}`;

/** @param {string} name the first word of the name a person commits under */
const firstName = (name) => name.split(' ')[0];

/** @param {Approval[]} approvals oldest first */
function byWords(approvals) {
	const names = [...new Set(approvals.map((a) => firstName(a.name)))];
	if (names.length === 1) return names[0];
	if (names.length === 2) return `${names[0]} and ${names[1]}`;
	const others = names.length - 2;
	return `${names[0]}, ${names[1]} and ${others} other${others === 1 ? '' : 's'}`;
}

/**
 * @param {ReturnType<typeof statusOf>} status
 * @returns {string[]} one line each, the two versions when they differ
 */
export function badgeWords({ status, implemented, approved }) {
	if (status === 'unversioned' || !implemented) return [];
	const implementedLine = `Implemented ${versionWords(implemented)}`;
	if (status === 'exempt') return [implementedLine];
	if (status === 'never') return ['Never approved', implementedLine];
	if (!approved) return [];
	const last = approved.approvals.at(-1);
	const approvedLine = `Approved ${shortVersion(approved.fingerprint)} · ${formatDate(last ? last.at : approved.firstSeen)} by ${byWords(approved.approvals)}`;
	return status === 'approved' ? [approvedLine] : ['Changed since approval', approvedLine, implementedLine];
}

/**
 * A story's versions, implemented one first, each with its approvals and its mark.
 * @param {StoryVersions | undefined} entry
 * @param {Approval[]} approvals the approvals of this story
 * @param {string | undefined} kind
 */
export function historyOf(entry, approvals, kind) {
	const latest = entry?.versions.find((v) => approvalsOf(approvals, v.fingerprint).length > 0);
	return (entry?.versions ?? []).map((version) => {
		const list = approvalsOf(approvals, version.fingerprint);
		const isImplemented = version.fingerprint === entry?.implemented;
		const isLatest = version.fingerprint === latest?.fingerprint;
		let mark = '';
		if (isImplemented && kind === 'behaviour') mark = 'Implemented now';
		else if (isImplemented && isLatest) mark = 'Implemented now · Latest approved';
		else if (isImplemented) mark = 'Implemented now, not approved';
		else if (isLatest) mark = 'Latest approved';
		return { ...version, approvals: list, mark };
	});
}

/**
 * The words of a history panel.
 * @param {{ id: string, title: string, name: string }} story
 * @param {ReturnType<typeof historyOf>} history
 */
export function historyWords(story, history) {
	const screen = story.title.split('/').slice(1).join('/') || story.title;
	const nobody = history.every((h) => h.approvals.length === 0);
	return {
		heading: `${screen} · ${story.name}`,
		entries: history.map((h) => ({
			version: `${shortVersion(h.fingerprint)} · first seen ${formatDate(h.firstSeen)}`,
			mark: h.mark,
			approvals: h.approvals.map((a) => ({
				line: `${firstName(a.name)}, ${formatDate(a.at)}`,
				note: a.note
			}))
		})),
		...(nobody ? { nobody: 'Nobody has approved this page yet.' } : {}),
		approveIntro: 'To approve this version, run this in your own terminal:',
		command: `just approve ${story.id}`
	};
}

/**
 * How many of a flow's steps and branches are approved at their implemented
 * version, changed since, or never approved. A story counts once.
 * @param {Flow} flow
 * @param {{ versions: Versions, approvals: Approval[], entries: Record<string, { tags?: string[] }> }} data
 */
export function flowStatus(flow, { versions, approvals, entries }) {
	const counts = { total: 0, approved: 0, changed: 0, never: 0 };
	for (const id of new Set([...flow.steps, ...(flow.branches ?? []).map((b) => b.to)])) {
		const { status } = statusOf(
			versions[id],
			approvals.filter((a) => a.story === id),
			entries[id] ? kindOf(entries[id]) : undefined
		);
		if (status === 'exempt' || status === 'unversioned') continue;
		counts.total += 1;
		counts[status] += 1;
	}
	return counts;
}

/** @param {{ total: number, approved: number, changed: number, never: number }} counts */
export function flowWords({ total, approved, changed, never }) {
	if (total === 0) return '';
	if (approved === total) return 'Approved';
	return [
		`${approved} of ${total} approved`,
		...(changed > 0 ? [`${changed} changed since approval`] : []),
		...(never > 0 ? [`${never} never approved`] : [])
	].join(' · ');
}

/**
 * @param {number} count every story that needs approval
 * @param {number} shown how many the sidebar lists now
 * @param {string} query what the find box holds
 */
export function filterWords(count, shown, query) {
	let empty;
	if (shown === 0) {
		empty = query.trim()
			? 'No page needing approval matches that.'
			: 'Every page is approved at its implemented version.';
	}
	return { label: `Needs approval (${count})`, empty };
}

/**
 * The address the two files are read from: beside the page, or a fixture pair
 * under `?data=<name>` so a state can be looked at without approving anything.
 * @param {URLSearchParams} query
 */
export function dataBase(query) {
	const name = query.get('data');
	if (!name) return '.';
	return `./fixtures/${/^[a-z0-9-]+$/.test(name) ? name : 'invalid'}`;
}

/**
 * Both files, or the sentence saying which could not be read. A canvas that
 * cannot know does not guess "Never approved".
 * @param {(url: string) => Promise<{ ok: boolean, json: () => Promise<any>, text: () => Promise<string> }>} load
 * @param {string} base
 * @returns {Promise<{ ok: true, versions: Versions, approvals: Approval[] } | { ok: false, message: string }>}
 */
export async function readData(load, base) {
	/** @type {Versions} */
	let versions;
	try {
		const response = await load(`${base}/versions.json`);
		if (!response.ok) throw new Error('missing');
		versions = await response.json();
	} catch {
		return { ok: false, message: 'Versions could not be read. Run just gen apps/storybook.' };
	}
	try {
		const response = await load(`${base}/approvals.jsonl`);
		if (!response.ok) throw new Error('missing');
		return { ok: true, versions, approvals: parseApprovals(await response.text()) };
	} catch {
		return { ok: false, message: 'Approvals could not be read. Statuses are hidden.' };
	}
}
