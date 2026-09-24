import { describe, expect, it } from 'vitest';
import {
	KINDS,
	kindOf,
	kindProblems,
	artboardsFrom,
	callbacksFor,
	branchBoards,
	connectorLabels,
	flowBoards,
	flowFrom,
	nextStory,
	placementProblems,
	routeFor,
	nearestLocation,
	normaliseMarkup,
	approvalLines,
	badgeWords,
	filterWords,
	flowStatus,
	flowWords,
	historyOf,
	historyWords,
	needsApproval,
	parseApprovals,
	readData,
	dataBase,
	statusOf,
	versionWords,
	approvalProblems,
	offersApprovedLook,
	parsePictureName,
	pictureLine,
	pictureName,
	pictureProblems,
	sitemapFrom,
	sizesFor,
	relativeToRepo,
	screenUrl,
	storyUrl
} from './lib.js';

const index = {
	v: 5,
	entries: {
		'atoms-button--primary': {
			id: 'atoms-button--primary',
			type: 'story',
			title: 'Atoms/Button',
			name: 'Primary'
		},
		'pages-overview--default': {
			id: 'pages-overview--default',
			type: 'story',
			title: 'Pages/Overview',
			name: 'Default'
		},
		'pages-overview--empty': {
			id: 'pages-overview--empty',
			type: 'story',
			title: 'Pages/Overview',
			name: 'Empty'
		},
		'pages-accounts--default': {
			id: 'pages-accounts--default',
			type: 'story',
			title: 'Pages/Accounts',
			name: 'Default'
		},
		'pages-overview--docs': {
			id: 'pages-overview--docs',
			type: 'docs',
			title: 'Pages/Overview',
			name: 'Docs'
		}
	}
};

describe('artboardsFrom', () => {
	it('shows every Pages story at compact and wide by default', () => {
		const boards = artboardsFrom(index, new URLSearchParams());
		expect(boards.map((b) => [b.id, b.size])).toEqual([
			['pages-overview--default', 'compact'],
			['pages-overview--default', 'wide'],
			['pages-overview--empty', 'compact'],
			['pages-overview--empty', 'wide'],
			['pages-accounts--default', 'compact'],
			['pages-accounts--default', 'wide']
		]);
	});

	it('ignores docs entries', () => {
		const ids = artboardsFrom(index, new URLSearchParams()).map((b) => b.id);
		expect(ids).not.toContain('pages-overview--docs');
	});

	it('narrows to one layer with ?layer, case-insensitively', () => {
		const ids = artboardsFrom(index, new URLSearchParams('layer=atoms&sizes=wide')).map(
			(b) => b.id
		);
		expect(ids).toEqual(['atoms-button--primary']);
		const again = artboardsFrom(index, new URLSearchParams('layer=Atoms&sizes=wide')).map(
			(b) => b.id
		);
		expect(again).toEqual(ids);
	});

	it('narrows to one component with ?title, and title wins over layer', () => {
		const ids = artboardsFrom(index, new URLSearchParams('title=Pages/Overview&sizes=wide')).map(
			(b) => b.id
		);
		expect(ids).toEqual(['pages-overview--default', 'pages-overview--empty']);
		const both = artboardsFrom(
			index,
			new URLSearchParams('layer=atoms&title=Pages/Accounts&sizes=wide')
		).map((b) => b.id);
		expect(both).toEqual(['pages-accounts--default']);
	});

	it('makes one artboard per story per size in ?sizes, story by story', () => {
		const boards = artboardsFrom(
			index,
			new URLSearchParams('title=Pages/Overview&sizes=compact,wide')
		);
		expect(boards.map((b) => [b.id, b.size])).toEqual([
			['pages-overview--default', 'compact'],
			['pages-overview--default', 'wide'],
			['pages-overview--empty', 'compact'],
			['pages-overview--empty', 'wide']
		]);
	});

	it('refuses an unknown size by name', () => {
		expect(() => artboardsFrom(index, new URLSearchParams('sizes=wide,huge'))).toThrow(/huge/);
	});

	it('carries theme and density onto every artboard, defaulting to light and comfortable', () => {
		const [plain] = artboardsFrom(index, new URLSearchParams());
		expect([plain?.theme, plain?.density]).toEqual(['light', 'comfortable']);
		const boards = artboardsFrom(index, new URLSearchParams('theme=dark&density=compact'));
		expect(boards.map((b) => [b.theme, b.density])).toEqual(boards.map(() => ['dark', 'compact']));
	});
});

describe('storyUrl', () => {
	it('points the story iframe at one story with its globals, the accessibility scan off', () => {
		expect(
			storyUrl('pages-overview--default', { theme: 'dark', density: 'compact', size: 'wide' })
		).toBe(
			'/iframe.html?id=pages-overview--default&viewMode=story&globals=theme:dark;density:compact;viewport.value:wide;a11y.manual:!true'
		);
	});
});

describe('relativeToRepo', () => {
	it('cuts an absolute path under /packages/ down to the repo root', () => {
		expect(relativeToRepo('/Users/ana/wimm/packages/ui/src/pages/Overview.svelte')).toEqual({
			path: 'packages/ui/src/pages/Overview.svelte',
			recognised: true
		});
	});

	it('cuts an absolute path under /apps/ the same way', () => {
		expect(relativeToRepo('/Users/ana/wimm/apps/web/src/routes/+page.svelte')).toEqual({
			path: 'apps/web/src/routes/+page.svelte',
			recognised: true
		});
	});

	it('returns a path with neither segment as it came, marked unrecognised', () => {
		expect(relativeToRepo('/opt/build/src/Thing.svelte')).toEqual({
			path: '/opt/build/src/Thing.svelte',
			recognised: false
		});
	});
});

/** A node shaped like what Svelte's dev build leaves on an element. */
const el = (meta, parentElement = null) => ({ __svelte_meta: meta, parentElement });
const loc = (file, line, column) => ({ file, line, column });

describe('nearestLocation', () => {
	it('reads the location recorded on the node itself, repo-relative', () => {
		const node = el({ loc: loc('/Users/ana/wimm/packages/ui/src/pages/Overview.svelte', 142, 5) });
		expect(nearestLocation(node)).toEqual({
			path: 'packages/ui/src/pages/Overview.svelte',
			recognised: true,
			line: 142,
			column: 5,
			parents: []
		});
	});

	it('walks up parentElement to the nearest element that recorded one', () => {
		const outer = el({ loc: loc('/r/packages/ui/src/atoms/Heading.svelte', 9, 2) });
		const bare = el(undefined, outer);
		const target = el(undefined, bare);
		expect(nearestLocation(target)).toMatchObject({
			path: 'packages/ui/src/atoms/Heading.svelte',
			line: 9,
			column: 2
		});
	});

	it('returns nothing when no ancestor recorded a location', () => {
		expect(nearestLocation(el(undefined, el(undefined)))).toBeNull();
	});

	it('lists the enclosing components from meta.parent, nearest first, capped at three', () => {
		const chain = (n) =>
			n === 0
				? null
				: { file: `/r/apps/web/src/C${n}.svelte`, line: n, column: 1, parent: chain(n - 1) };
		const node = el({
			loc: loc('/r/packages/ui/src/pages/Overview.svelte', 1, 1),
			parent: chain(5)
		});
		expect(nearestLocation(node)?.parents).toEqual([
			{ path: 'apps/web/src/C5.svelte', recognised: true, line: 5, column: 1 },
			{ path: 'apps/web/src/C4.svelte', recognised: true, line: 4, column: 1 },
			{ path: 'apps/web/src/C3.svelte', recognised: true, line: 3, column: 1 }
		]);
	});
});

describe('sitemapFrom', () => {
	it('groups artboards into layer > title > story, once per story however many sizes', () => {
		const boards = artboardsFrom(index, new URLSearchParams('layer=pages'));
		expect(sitemapFrom(boards)).toEqual([
			{
				layer: 'Pages',
				titles: [
					{
						title: 'Pages/Overview',
						stories: [
							{ id: 'pages-overview--default', name: 'Default' },
							{ id: 'pages-overview--empty', name: 'Empty' }
						]
					},
					{
						title: 'Pages/Accounts',
						stories: [{ id: 'pages-accounts--default', name: 'Default' }]
					}
				]
			}
		]);
	});
});

describe('flowFrom', () => {
	const flows = {
		tour: {
			title: 'Tour',
			steps: ['pages-overview--default', 'pages-accounts--default'],
			transitions: [{ from: 'pages-overview--default', on: 'ongo', to: 'pages-accounts--default' }]
		}
	};

	it('returns a valid flow', () => {
		expect(flowFrom(flows, 'tour', index)).toBe(flows.tour);
	});

	it('refuses an unknown flow by name', () => {
		expect(() => flowFrom(flows, 'nope', index)).toThrow(/nope/);
	});

	it('reports every unknown story id at once', () => {
		const bad = {
			tour: {
				title: 'Tour',
				steps: ['pages-overview--default', 'gone--one'],
				transitions: [{ from: 'gone--two', on: 'ongo', to: 'gone--three' }]
			}
		};
		expect(() => flowFrom(bad, 'tour', index)).toThrow(
			/gone--one[\s\S]*gone--two[\s\S]*gone--three/
		);
	});

	it('refuses a transition with no callback name', () => {
		const bad = {
			tour: {
				title: 'Tour',
				steps: ['pages-overview--default'],
				transitions: [{ from: 'pages-overview--default', to: 'pages-accounts--default' }]
			}
		};
		expect(() => flowFrom(bad, 'tour', index)).toThrow(/callback/);
	});

	it('does not count a docs entry as a story', () => {
		const bad = {
			tour: { title: 'Tour', steps: ['pages-overview--docs'], transitions: [] }
		};
		expect(() => flowFrom(bad, 'tour', index)).toThrow(/pages-overview--docs/);
	});

	/** @param {object} branch */
	const branching = (branch) => ({
		tour: { title: 'Tour', steps: ['pages-overview--default'], transitions: [], branches: [branch] }
	});

	it('accepts a branch that leaves a step for a story off the happy path', () => {
		const good = branching({
			from: 'pages-overview--default',
			outcome: 'the bank said no',
			to: 'pages-accounts--default'
		});
		expect(flowFrom(good, 'tour', index)).toBe(good.tour);
	});

	it('refuses a branch to a story that does not exist', () => {
		const bad = branching({ from: 'pages-overview--default', outcome: 'x', to: 'gone--one' });
		expect(() => flowFrom(bad, 'tour', index)).toThrow(/no story "gone--one"/);
	});

	it('refuses a branch that leaves something other than a step', () => {
		const bad = branching({
			from: 'pages-accounts--default',
			outcome: 'x',
			to: 'pages-accounts--default'
		});
		expect(() => flowFrom(bad, 'tour', index)).toThrow(/not a step/);
	});

	it('refuses a branch that lands back on the happy path', () => {
		const bad = branching({
			from: 'pages-overview--default',
			outcome: 'x',
			to: 'pages-overview--default'
		});
		expect(() => flowFrom(bad, 'tour', index)).toThrow(/lands on a step/);
	});

	it('refuses a branch that does not say what happened', () => {
		const bad = branching({ from: 'pages-overview--default', to: 'pages-accounts--default' });
		expect(() => flowFrom(bad, 'tour', index)).toThrow(/names no outcome/);
	});
});

describe('branchBoards', () => {
	it('gives each branch its step, its outcome and its state at every size', () => {
		const flow = {
			title: 'Tour',
			steps: ['pages-overview--default'],
			transitions: [],
			branches: [
				{
					from: 'pages-overview--default',
					outcome: 'the bank said no',
					to: 'pages-accounts--default'
				}
			]
		};
		const [branch] = branchBoards(index, flow, new URLSearchParams('sizes=compact,wide'));
		expect(branch.from).toBe('pages-overview--default');
		expect(branch.outcome).toBe('the bank said no');
		expect(branch.boards.map((b) => [b.id, b.size])).toEqual([
			['pages-accounts--default', 'compact'],
			['pages-accounts--default', 'wide']
		]);
	});

	it('is empty for a flow with no branches', () => {
		const flow = { title: 'Tour', steps: ['pages-overview--default'], transitions: [] };
		expect(branchBoards(index, flow, new URLSearchParams())).toEqual([]);
	});
});

describe('placementProblems', () => {
	const flows = {
		tour: {
			title: 'Tour',
			steps: ['a--step'],
			transitions: [],
			branches: [{ from: 'a--step', outcome: 'x', to: 'a--branch' }]
		}
	};
	/** @param {string[]} ids */
	const states = (ids) => ids.map((id) => ({ id, tags: ['kind-state'] }));
	const asserts = { id: 'a--asserts', tags: ['kind-behaviour'] };

	it('is satisfied when every story that shows something is on a flow or listed as unplaced', () => {
		const pages = [...states(['a--step', 'a--branch', 'a--loose']), asserts];
		expect(placementProblems(flows, ['a--loose'], pages)).toEqual([]);
	});

	it('refuses a page story that is nowhere', () => {
		expect(placementProblems(flows, [], states(['a--step', 'a--branch', 'a--new']))).toEqual([
			'"a--new" is on no flow and not listed as unplaced'
		]);
	});

	it('refuses a story listed as unplaced once it is on a flow', () => {
		expect(placementProblems(flows, ['a--branch'], states(['a--step', 'a--branch']))).toEqual([
			'"a--branch" is on a flow and still listed as unplaced'
		]);
	});

	it('refuses an unplaced story that no longer exists', () => {
		expect(placementProblems(flows, ['a--gone'], states(['a--step', 'a--branch']))).toEqual([
			'unplaced "a--gone" is not a page story'
		]);
	});

	it('refuses a story that only asserts something on a flow or on the list', () => {
		const pages = [...states(['a--step']), { id: 'a--branch', tags: ['kind-behaviour'] }, asserts];
		expect(placementProblems(flows, ['a--asserts'], pages)).toEqual([
			'"a--branch" only asserts something and is on a flow',
			'"a--asserts" only asserts something and is listed as unplaced'
		]);
	});
});

describe('nextStory', () => {
	const flow = {
		title: 'Tour',
		steps: [],
		transitions: [
			{ from: 'a', on: 'ongo', to: 'b' },
			{ from: 'a', on: 'onsee', to: 'c' },
			{ from: 'b', on: 'ongo', to: 'c' }
		]
	};

	it('follows a mapped callback', () => {
		expect(nextStory(flow, 'a', 'onsee')).toBe('c');
		expect(nextStory(flow, 'b', 'ongo')).toBe('c');
	});

	it('returns nothing for an unmapped callback', () => {
		expect(nextStory(flow, 'a', 'onnope')).toBeUndefined();
		expect(nextStory(flow, 'z', 'ongo')).toBeUndefined();
	});
});

describe('callbacksFor', () => {
	const flow = {
		title: 'Tour',
		steps: [],
		transitions: [
			{ from: 'a', on: 'ongo', to: 'b' },
			{ from: 'a', on: 'ongo', to: 'c' },
			{ from: 'a', on: 'onsee', to: 'c' }
		]
	};

	it('lists each callback of a step once', () => {
		expect(callbacksFor(flow, 'a')).toEqual(['ongo', 'onsee']);
	});

	it('lists none for a step with no transitions', () => {
		expect(callbacksFor(flow, 'b')).toEqual([]);
	});
});

describe('routeFor', () => {
	const routes = { '/': 'home', '/transactions': 'ledger' };

	it('matches a path', () => {
		expect(routeFor(routes, '/transactions')).toBe('ledger');
		expect(routeFor(routes, '/')).toBe('home');
	});

	it('ignores query and hash', () => {
		expect(routeFor(routes, '/transactions?refresh=1#top')).toBe('ledger');
	});

	it('resolves an absolute same-origin href', () => {
		expect(routeFor(routes, 'http://localhost:6006/transactions', 'http://localhost:6006')).toBe(
			'ledger'
		);
	});

	it('returns nothing for an unmapped or foreign href', () => {
		expect(routeFor(routes, '/connect/monzo')).toBeUndefined();
		expect(routeFor(routes, 'https://elsewhere.test/transactions', 'http://localhost:6006')).toBe(
			undefined
		);
	});
});

describe('flowBoards', () => {
	const flow = {
		title: 'Tour',
		steps: ['pages-accounts--default', 'pages-overview--default'],
		transitions: []
	};

	it('draws the steps in flow order, each at every size', () => {
		const boards = flowBoards(index, flow, new URLSearchParams('sizes=compact,wide'));
		expect(boards.map((b) => [b.id, b.size])).toEqual([
			['pages-accounts--default', 'compact'],
			['pages-accounts--default', 'wide'],
			['pages-overview--default', 'compact'],
			['pages-overview--default', 'wide']
		]);
	});

	it('carries theme and density and refuses an unknown size', () => {
		const [board] = flowBoards(index, flow, new URLSearchParams('theme=dark&density=compact'));
		expect(board).toMatchObject({ theme: 'dark', density: 'compact', title: 'Pages/Accounts' });
		expect(() => flowBoards(index, flow, new URLSearchParams('sizes=huge'))).toThrow(/huge/);
	});
});

describe('connectorLabels', () => {
	it('names the callback between neighbours, and nothing where none is declared', () => {
		const flow = {
			title: 'Tour',
			steps: ['a', 'b', 'c'],
			transitions: [{ from: 'a', on: 'ongo', to: 'b' }]
		};
		expect(connectorLabels(flow)).toEqual(['ongo', '']);
	});
});

describe('screenUrl', () => {
	it('addresses one screen and the story to land on', () => {
		expect(screenUrl(new URLSearchParams(), 'Pages/Overview', 'pages-overview--populated')).toBe(
			'?title=Pages%2FOverview#pages-overview--populated'
		);
	});

	it('keeps the theme, density and sizes and drops the view it came from', () => {
		const query = new URLSearchParams('flow=connect&theme=dark&density=compact&sizes=ultra');
		expect(screenUrl(query, 'Pages/Overview')).toBe(
			'?title=Pages%2FOverview&theme=dark&density=compact&sizes=ultra'
		);
	});

	it('keeps the fixture and the approval filter, so moving on does not leave them', () => {
		const query = new URLSearchParams('flow=connect&data=mixed&needs=1');
		expect(screenUrl(query, 'Pages/Overview')).toBe('?title=Pages%2FOverview&data=mixed&needs=1');
	});
});

describe('kinds', () => {
	it('draws rows in the order a designer reads a screen, assertions last', () => {
		expect(Object.keys(KINDS)).toEqual(['state', 'waiting', 'error', 'outcome', 'behaviour']);
	});

	it("reads a story's kind from its one kind tag", () => {
		expect(kindOf({ tags: ['dev', 'test', 'kind-error'] })).toBe('error');
		expect(kindOf({ tags: ['dev', 'test'] })).toBeUndefined();
		expect(kindOf({ tags: ['kind-mystery'] })).toBeUndefined();
	});

	it('is satisfied when every page story carries exactly one known kind', () => {
		expect(
			kindProblems([
				{ id: 'a--one', tags: ['dev', 'kind-state'] },
				{ id: 'a--two', tags: ['kind-behaviour', 'test'] }
			])
		).toEqual([]);
	});

	it('refuses a page story with no kind', () => {
		expect(kindProblems([{ id: 'a--one', tags: ['dev', 'test'] }])).toEqual([
			'"a--one" has no kind tag'
		]);
	});

	it('refuses a page story with two kinds', () => {
		expect(kindProblems([{ id: 'a--one', tags: ['kind-state', 'kind-error'] }])).toEqual([
			'"a--one" has more than one kind tag: kind-state, kind-error'
		]);
	});

	it('refuses a kind the canvas draws no row for', () => {
		expect(kindProblems([{ id: 'a--one', tags: ['kind-mystery'] }])).toEqual([
			'"a--one" has the unknown kind tag "kind-mystery"; kinds are state, waiting, error, outcome, behaviour'
		]);
	});

	it("gives each artboard its story's kind", () => {
		const tagged = {
			entries: {
				'pages-a--one': {
					id: 'pages-a--one',
					type: 'story',
					title: 'Pages/A',
					name: 'One',
					tags: ['kind-waiting']
				}
			}
		};
		const [board] = artboardsFrom(tagged, new URLSearchParams('title=Pages/A&sizes=wide'));
		expect(board.kind).toBe('waiting');
	});
});

describe('sizesFor', () => {
	it('draws a story at the sizes the view asks for', () => {
		expect(sizesFor({ tags: ['kind-state'] }, ['compact', 'wide'])).toEqual(['compact', 'wide']);
	});

	it('draws a story that pins its viewport at that size only, whatever the view asks for', () => {
		expect(sizesFor({ tags: ['kind-state', 'size-compact'] }, ['compact', 'wide'])).toEqual([
			'compact'
		]);
		expect(sizesFor({ tags: ['size-compact'] }, ['wide'])).toEqual(['compact']);
	});

	it('gives a pinned story one artboard on a screen and on a flow', () => {
		const pinned = {
			entries: {
				'pages-a--compact': {
					id: 'pages-a--compact',
					type: 'story',
					title: 'Pages/A',
					name: 'Compact',
					tags: ['kind-state', 'size-compact']
				}
			}
		};
		const query = new URLSearchParams('title=Pages/A&sizes=compact,wide');
		expect(artboardsFrom(pinned, query).map((b) => b.size)).toEqual(['compact']);
		const flow = { title: 'Tour', steps: ['pages-a--compact'], transitions: [] };
		expect(flowBoards(pinned, flow, query).map((b) => [b.step, b.size])).toEqual([[0, 'compact']]);
	});

	it('refuses a size tag naming no size', () => {
		expect(kindProblems([{ id: 'a--one', tags: ['kind-state', 'size-huge'] }])).toEqual([
			'"a--one" has the unknown size tag "size-huge"; sizes are compact, medium, wide, ultra'
		]);
	});
});

describe('normaliseMarkup', () => {
	it('drops hydration and block comments', () => {
		expect(normaliseMarkup('<div><!----><!--[--><p>a</p><!--]--><!--[!--></div>')).toBe(
			'<div><p>a</p></div>'
		);
	});

	it('collapses whitespace', () => {
		expect(normaliseMarkup('<ul>\n\t<li>a   \n b</li>\n  <li>c</li>\n</ul>')).toBe(
			'<ul><li>a b</li><li>c</li></ul>'
		);
	});

	it('treats two renders with different generated ids alike, and keeps the wiring', () => {
		const one = '<label for="input-a1">Name</label><input id="input-a1">';
		const two = '<label for="input-zz9">Name</label><input id="input-zz9">';
		expect(normaliseMarkup(one)).toBe(normaliseMarkup(two));
		expect(normaliseMarkup(one)).toBe('<label for="id-1">Name</label><input id="id-1">');
	});

	it('numbers ids by first appearance across aria references and fragment links', () => {
		const html =
			'<a href="#b7">x</a><div id="q1" aria-labelledby="q2 b7" aria-controls="q1"></div><p id="q2"></p>';
		expect(normaliseMarkup(html)).toBe(
			'<a href="#id-1">x</a><div aria-controls="id-2" aria-labelledby="id-3 id-1" id="id-2"></div><p id="id-3"></p>'
		);
	});

	it('leaves an unrelated pair of ids distinct', () => {
		expect(normaliseMarkup('<i id="a"></i><b for="b"></b>')).toBe('<i id="id-1"></i><b for="id-2"></b>');
	});

	it('ignores attribute order', () => {
		expect(normaliseMarkup('<a href="/x" class="c" title="t">a</a>')).toBe(
			normaliseMarkup('<a title="t" href="/x" class="c">a</a>')
		);
	});

	it('is changed by a word', () => {
		expect(normaliseMarkup('<p>Balance</p>')).not.toBe(normaliseMarkup('<p>Balances</p>'));
	});

	it('is changed by a scoped class', () => {
		expect(normaliseMarkup('<p class="svelte-a1b2c3">a</p>')).not.toBe(
			normaliseMarkup('<p class="svelte-d4e5f6">a</p>')
		);
	});

	it('keeps inline style', () => {
		expect(normaliseMarkup('<p style="margin: 0">a</p>')).not.toBe(
			normaliseMarkup('<p style="margin: 4px">a</p>')
		);
	});

	it('drops attributes the harness adds', () => {
		expect(normaliseMarkup('<div data-vitest-x="1" data-storybook-y="2"><p>a</p></div>')).toBe(
			'<div><p>a</p></div>'
		);
	});
});

describe('approvalProblems', () => {
	const FP = 'a'.repeat(64);
	const OTHER = 'b'.repeat(64);
	const versions = {
		'pages-overview--populated': {
			implemented: FP,
			versions: [
				{ fingerprint: FP, firstSeen: '2026-09-20' },
				{ fingerprint: OTHER, firstSeen: '2026-09-12' }
			]
		}
	};
	const record = (over = {}) => ({
		story: 'pages-overview--populated',
		fingerprint: FP,
		name: 'Ada Lovelace',
		email: 'e@example.com',
		at: '2026-09-20T18:04:11Z',
		note: '',
		...over
	});
	const line = (over) => JSON.stringify(record(over));
	const problems = (...lines) => approvalProblems(lines, versions);

	it('passes a well formed record, and two people on one version', () => {
		expect(problems(line(), line({ email: 'g@example.com', name: 'Grace' }))).toEqual([]);
		expect(problems()).toEqual([]);
	});

	it('refuses a line that is not JSON', () => {
		expect(problems(line(), '{oops')).toEqual(['line 2: not valid JSON']);
	});

	it('refuses a line that is not an object', () => {
		expect(problems('[1]')).toEqual(['line 1: not a record']);
	});

	it('refuses a field it does not know', () => {
		expect(problems(JSON.stringify({ ...record(), extra: 1 }))).toEqual([
			'line 1: unknown field "extra"'
		]);
	});

	it('refuses a missing field', () => {
		const rest = record();
		delete rest.email;
		expect(problems(JSON.stringify(rest))).toEqual(['line 1: missing field "email"']);
	});

	it('refuses a fingerprint that is not 64 hex', () => {
		expect(problems(line({ fingerprint: 'abc' }))).toContain(
			'line 1: fingerprint is not 64 hex characters'
		);
	});

	it('refuses a time that is not ISO', () => {
		expect(problems(line({ at: '20 Sep 2026' }))).toEqual([
			'line 1: time is not an ISO 8601 UTC time'
		]);
	});

	it('refuses a story the versions file has never held', () => {
		expect(problems(line({ story: 'pages-x--y' }))).toEqual([
			'line 1: story "pages-x--y" is not in versions.json'
		]);
	});

	it('refuses a version the story never had', () => {
		expect(problems(line({ fingerprint: 'c'.repeat(64) }))).toEqual([
			'line 1: story "pages-overview--populated" never had version ccccccc'
		]);
	});

	it('refuses one person approving one version twice', () => {
		expect(problems(line(), line({ at: '2026-09-21T09:00:00Z' }))).toEqual([
			'line 2: e@example.com already approved this version on line 1'
		]);
	});

	it('splits a file into lines without inventing a last one', () => {
		expect(approvalLines('a\nb\n')).toEqual(['a', 'b']);
		expect(approvalLines('')).toEqual([]);
	});
});

describe('the status of a page story', () => {
	const fp = (c) => c.repeat(64);
	const entry = {
		implemented: fp('a'),
		versions: [
			{ fingerprint: fp('a'), firstSeen: '2026-09-20' },
			{ fingerprint: fp('b'), firstSeen: '2026-09-12' }
		]
	};
	const by = (name, email, fingerprint, at, note = '') => ({
		story: 's',
		fingerprint,
		name,
		email,
		at,
		note
	});
	const ada = (f, at = '2026-09-20T18:04:11Z') => by('Ada Lovelace', 'e@x', f, at);
	const grace = (f, at = '2026-09-21T09:00:00Z') => by('Grace Hopper', 'g@x', f, at);

	it('is approved when the approved version is the implemented one', () => {
		const got = statusOf(entry, [ada(fp('a'))], 'state');
		expect(got.status).toBe('approved');
		expect(got.implemented?.fingerprint).toBe(fp('a'));
		expect(got.approved?.fingerprint).toBe(fp('a'));
		expect(got.approved?.approvals).toHaveLength(1);
	});

	it('is changed when only an earlier version was approved, and carries both versions', () => {
		const got = statusOf(entry, [ada(fp('b'), '2026-09-12T10:00:00Z')], 'state');
		expect(got.status).toBe('changed');
		expect(got.implemented?.fingerprint).toBe(fp('a'));
		expect(got.approved?.fingerprint).toBe(fp('b'));
		expect(got.approved?.firstSeen).toBe('2026-09-12');
	});

	it('is never approved when there are no approvals, carrying the implemented version alone', () => {
		const got = statusOf(entry, [], 'state');
		expect(got.status).toBe('never');
		expect(got.implemented?.fingerprint).toBe(fp('a'));
		expect(got.approved).toBeUndefined();
	});

	it('is exempt for a behaviour story, whatever was approved', () => {
		expect(statusOf(entry, [], 'behaviour').status).toBe('exempt');
		expect(statusOf(entry, [ada(fp('a'))], 'behaviour').status).toBe('exempt');
		expect(statusOf(entry, [], 'behaviour').implemented?.fingerprint).toBe(fp('a'));
	});

	it('has no status for a story the versions file lacks', () => {
		expect(statusOf(undefined, [], 'state').status).toBe('unversioned');
	});

	it('takes the newest approved version as the approved one', () => {
		const got = statusOf(entry, [ada(fp('b'))], 'state');
		expect(got.approved?.fingerprint).toBe(fp('b'));
		const both = statusOf(entry, [ada(fp('b')), ada(fp('a'))], 'state');
		expect(both.approved?.fingerprint).toBe(fp('a'));
	});

	it('lists approvals oldest first', () => {
		const got = statusOf(entry, [grace(fp('a')), ada(fp('a'))], 'state');
		expect(got.approved?.approvals.map((a) => a.name)).toEqual(['Ada Lovelace', 'Grace Hopper']);
	});

	it('needs approval when changed or never approved, and not otherwise', () => {
		expect(['changed', 'never'].every(needsApproval)).toBe(true);
		expect(['approved', 'exempt', 'unversioned'].some(needsApproval)).toBe(false);
	});

	describe('badge words', () => {
		const words = (approvals, kind = 'state') => badgeWords(statusOf(entry, approvals, kind));
		it('approved, one version, one person', () => {
			expect(words([ada(fp('a'))])).toEqual(['Approved aaaaaaa · 20 Sep 2026 by Ada']);
		});
		it('approved by two, dated by the later approval', () => {
			expect(words([ada(fp('a')), grace(fp('a'))])).toEqual([
				'Approved aaaaaaa · 21 Sep 2026 by Ada and Grace'
			]);
		});
		it('approved by three or more', () => {
			const third = by('Ada Lovelace', 'a@x', fp('a'), '2026-09-22T09:00:00Z');
			expect(words([ada(fp('a')), grace(fp('a')), third])).toEqual([
				'Approved aaaaaaa · 22 Sep 2026 by Ada, Grace and 1 other'
			]);
			const fourth = by('Alan Turing', 't@x', fp('a'), '2026-09-23T09:00:00Z');
			expect(words([ada(fp('a')), grace(fp('a')), third, fourth])[0]).toContain(
				'Ada, Grace and 2 others'
			);
		});
		it('changed, both versions', () => {
			expect(words([ada(fp('b'), '2026-09-12T10:00:00Z')])).toEqual([
				'Changed since approval',
				'Approved bbbbbbb · 12 Sep 2026 by Ada',
				'Implemented aaaaaaa · 20 Sep 2026'
			]);
		});
		it('never approved', () => {
			expect(words([])).toEqual(['Never approved', 'Implemented aaaaaaa · 20 Sep 2026']);
		});
		it('exempt shows the implemented version and no status', () => {
			expect(words([], 'behaviour')).toEqual(['Implemented aaaaaaa · 20 Sep 2026']);
		});
		it('unversioned shows nothing', () => {
			expect(badgeWords(statusOf(undefined, [], 'state'))).toEqual([]);
		});
		it('writes a version as seven characters, a dot and a date', () => {
			expect(versionWords({ fingerprint: `a3f9c21${'0'.repeat(57)}`, firstSeen: '2026-09-20' })).toBe(
				'a3f9c21 · 20 Sep 2026'
			);
		});
	});

	describe('history', () => {
		const three = {
			implemented: fp('c'),
			versions: [
				{ fingerprint: fp('c'), firstSeen: '2026-09-20' },
				{ fingerprint: fp('b'), firstSeen: '2026-09-15' },
				{ fingerprint: fp('a'), firstSeen: '2026-09-08' }
			]
		};
		it('marks the implemented version, the latest approved, and leaves the rest unmarked', () => {
			const got = historyOf(three, [ada(fp('a')), ada(fp('b')), grace(fp('b'))], 'state');
			expect(got.map((h) => h.mark)).toEqual([
				'Implemented now, not approved',
				'Latest approved',
				''
			]);
			expect(got[1].approvals.map((a) => a.name)).toEqual(['Ada Lovelace', 'Grace Hopper']);
		});
		it('carries both marks on one entry when the implemented version is the latest approved', () => {
			const got = historyOf(three, [ada(fp('c'))], 'state');
			expect(got[0].mark).toBe('Implemented now · Latest approved');
			expect(got).toHaveLength(3);
		});
		it('says nobody approved a page nobody approved', () => {
			const got = historyOf(three, [], 'state');
			expect(got[0].mark).toBe('Implemented now, not approved');
			expect(historyWords({ title: 'Pages/Overview', name: 'Populated', id: 'pages-overview--populated' }, got)).toMatchObject({
				heading: 'Overview · Populated',
				nobody: 'Nobody has approved this page yet.',
				approveIntro: 'To approve this version, run this in your own terminal:',
				command: 'just approve pages-overview--populated'
			});
		});
		it('writes each version and each approval with its note', () => {
			const got = historyOf(
				three,
				[by('Ada Lovelace', 'e@x', fp('b'), '2026-09-15T12:00:00Z', 'after the rows were tightened')],
				'state'
			);
			const words = historyWords({ title: 'Pages/Overview', name: 'Populated', id: 'pages-overview--populated' }, got);
			expect(words.entries[1]).toEqual({
				version: 'bbbbbbb · first seen 15 Sep 2026',
				mark: 'Latest approved',
				approvals: [{ line: 'Ada, 15 Sep 2026', note: 'after the rows were tightened' }]
			});
			expect(words.nobody).toBeUndefined();
		});
		it('marks an exempt story implemented, and does not ask for approval', () => {
			expect(historyOf(three, [], 'behaviour')[0].mark).toBe('Implemented now');
		});
	});

	describe('a flow', () => {
		const ids = (prefix, n) => Array.from({ length: n }, (_, i) => `${prefix}${i}`);
		const stepIds = ids('step', 8);
		const branchIds = ids('branch', 4);
		const flow = {
			title: 't',
			steps: stepIds,
			transitions: [],
			branches: branchIds.map((to) => ({ from: 'step0', outcome: 'o', to }))
		};
		const all = [...stepIds, ...branchIds];
		const versions = Object.fromEntries(
			all.map((id) => [
				id,
				{
					implemented: fp('a'),
					versions: [
						{ fingerprint: fp('a'), firstSeen: '2026-09-20' },
						{ fingerprint: fp('b'), firstSeen: '2026-09-12' }
					]
				}
			])
		);
		const entries = Object.fromEntries(all.map((id) => [id, { id, tags: ['kind-state'] }]));
		const approve = (list, f) =>
			list.map((id) => ({ ...ada(f), story: id }));

		it('counts approved, changed and never once each', () => {
			const approvals = [
				...approve(all.slice(0, 7), fp('a')),
				...approve(all.slice(7, 10), fp('b'))
			];
			const got = flowStatus(flow, { versions, approvals, entries });
			expect(got).toEqual({ total: 12, approved: 7, changed: 3, never: 2 });
			expect(flowWords(got)).toBe('7 of 12 approved · 3 changed since approval · 2 never approved');
		});
		it('leaves out a part that counts none', () => {
			const got = flowStatus(flow, { versions, approvals: approve(all.slice(0, 10), fp('a')), entries });
			expect(flowWords(got)).toBe('10 of 12 approved · 2 never approved');
		});
		it('reads Approved when every page is approved at its implemented version', () => {
			const got = flowStatus(flow, { versions, approvals: approve(all, fp('a')), entries });
			expect(flowWords(got)).toBe('Approved');
			const later = flowStatus(flow, {
				versions,
				approvals: [...approve(all.slice(1), fp('a')), ...approve(['step0'], fp('b'))],
				entries
			});
			expect(flowWords(later)).toBe('11 of 12 approved · 1 changed since approval');
		});
		it('does not count a behaviour story, and shows nothing when nothing needs approval', () => {
			const quiet = Object.fromEntries(all.map((id) => [id, { id, tags: ['kind-behaviour'] }]));
			const got = flowStatus(flow, { versions, approvals: [], entries: quiet });
			expect(got.total).toBe(0);
			expect(flowWords(got)).toBe('');
		});
		it('counts a story once when it is a step and a branch target', () => {
			const twice = { ...flow, branches: [{ from: 'step0', outcome: 'o', to: 'branch0' }, { from: 'step1', outcome: 'o', to: 'branch0' }] };
			expect(flowStatus(twice, { versions, approvals: [], entries }).total).toBe(9);
		});
	});

	describe('the sidebar filter', () => {
		it('writes the label with the count', () => {
			expect(filterWords(23, 23, '').label).toBe('Needs approval (23)');
		});
		it('says everything is approved when nothing needs a look', () => {
			expect(filterWords(0, 0, '').empty).toBe('Every page is approved at its implemented version.');
		});
		it('says nothing matches when a find leaves nothing', () => {
			expect(filterWords(5, 0, 'zzz').empty).toBe('No page needing approval matches that.');
		});
		it('says nothing when there is something to list', () => {
			expect(filterWords(5, 2, '').empty).toBeUndefined();
		});
	});

	it('reads approvals from a record, keeping the well formed lines', () => {
		const text = `${JSON.stringify(ada(fp('a')))}\nnot json\n`;
		expect(parseApprovals(text)).toHaveLength(1);
	});
});

describe('reading the two files', () => {
	const ok = (body) => ({ ok: true, json: async () => JSON.parse(body), text: async () => body });
	const fail = { ok: false, json: async () => ({}), text: async () => '' };
	const line = JSON.stringify({ story: 's', fingerprint: 'a'.repeat(64) });

	it('reads versions and approvals from beside the page', async () => {
		const seen = [];
		const got = await readData(async (url) => {
			seen.push(url);
			return url.endsWith('versions.json') ? ok('{"s":{"versions":[]}}') : ok(`${line}\n`);
		}, '.');
		expect(seen).toEqual(['./versions.json', './approvals.jsonl']);
		expect(got).toMatchObject({ ok: true, versions: { s: { versions: [] } } });
		expect(got.ok && got.approvals).toHaveLength(1);
	});

	it('says versions could not be read when they are missing or not JSON', async () => {
		const message = 'Versions could not be read. Run just gen apps/storybook.';
		expect(await readData(async () => fail, '.')).toEqual({ ok: false, message });
		expect(await readData(async () => ok('<html>'), '.')).toEqual({ ok: false, message });
		expect(await readData(async () => { throw new Error('network'); }, '.')).toEqual({ ok: false, message });
	});

	it('says approvals could not be read when only they are missing', async () => {
		const got = await readData(async (url) => (url.endsWith('versions.json') ? ok('{}') : fail), '.');
		expect(got).toEqual({ ok: false, message: 'Approvals could not be read. Statuses are hidden.' });
	});

	it('reads a fixture pair for ?data=<name>, and refuses a name that is not a fixture name', () => {
		expect(dataBase(new URLSearchParams(''))).toBe('.');
		expect(dataBase(new URLSearchParams('data=never-some'))).toBe('./fixtures/never-some');
		expect(dataBase(new URLSearchParams('data=../x'))).toBe('./fixtures/invalid');
	});
});

describe('pictures of the last approved version', () => {
	const a = 'a'.repeat(64);

	it('names a picture by story, fingerprint and size, and reads the name back', () => {
		const name = pictureName('pages-overview--populated', a, 'wide');
		expect(name).toBe(`pages-overview--populated/${a}-wide.png`);
		expect(parsePictureName(name)).toEqual({
			story: 'pages-overview--populated',
			fingerprint: a,
			size: 'wide'
		});
	});

	const b = 'b'.repeat(64);
	const sizes = ['compact', 'medium', 'wide', 'ultra'];
	const story = 'pages-overview--populated';
	/** @param {string} fingerprint @param {string} at */
	const approval = (fingerprint, at, email = 'e@x.test') => ({
		story,
		fingerprint,
		name: 'E',
		email,
		at,
		note: ''
	});
	const dir = 'apps/storybook/canvas/approved';

	it('refuses a picture left from an earlier approved version', () => {
		const approvals = [approval(a, '2026-09-20T10:00:00Z'), approval(b, '2026-09-21T10:00:00Z')];
		expect(pictureProblems([pictureName(story, a, 'wide')], approvals, sizes)).toEqual([
			`${dir}/${story}/${a}-wide.png is not of the last approved version, bbbbbbb.`
		]);
		expect(pictureProblems([pictureName(story, b, 'wide')], approvals, sizes)).toEqual([]);
	});

	it('refuses a picture of a story nobody approved', () => {
		const approvals = [approval(a, '2026-09-20T10:00:00Z')];
		expect(pictureProblems([pictureName('pages-x--y', a, 'wide')], approvals, sizes)).toEqual([
			`${dir}/pages-x--y/${a}-wide.png belongs to no approved story.`
		]);
	});

	it('refuses a picture at a size the canvas does not have', () => {
		const approvals = [approval(a, '2026-09-20T10:00:00Z')];
		expect(pictureProblems([pictureName(story, a, 'tablet')], approvals, sizes)).toEqual([
			`${dir}/${story}/${a}-tablet.png names a size the canvas does not have.`
		]);
	});

	it('refuses a folder left from an approval that did not finish, once per folder', () => {
		const approvals = [approval(a, '2026-09-20T10:00:00Z')];
		const next = [`${story}.next/${a}-compact.png`, `${story}.next/${a}-wide.png`];
		expect(pictureProblems(next, approvals, sizes)).toEqual([
			`${dir}/${story}.next/ is left from an approval that did not finish. Delete it.`
		]);
	});

	it('refuses a file not named for a fingerprint and a size', () => {
		const approvals = [approval(a, '2026-09-20T10:00:00Z')];
		expect(pictureProblems([`${story}/wide.png`], approvals, sizes)).toEqual([
			`${dir}/${story}/wide.png is not named <fingerprint>-<size>.png.`
		]);
	});

	it('passes a story with no pictures, and one with a picture of its last approval', () => {
		const approvals = [approval(a, '2026-09-20T10:00:00Z'), approval(b, '2026-09-21T10:00:00Z')];
		expect(pictureProblems([], approvals, sizes)).toEqual([]);
		const current = sizes.map((size) => pictureName(story, b, size));
		expect(pictureProblems(current, approvals, sizes)).toEqual([]);
	});
});

describe('the approved look on an artboard', () => {
	const light = { theme: 'light', density: 'comfortable' };

	it('is offered on a changed artboard, light and comfortable', () => {
		expect(offersApprovedLook('changed', light)).toBe(true);
	});

	it.each([
		['approved', light],
		['never', light],
		['exempt', light],
		['unversioned', light],
		['changed', { theme: 'dark', density: 'comfortable' }],
		['changed', { theme: 'light', density: 'compact' }]
	])('is not offered on a %s artboard viewed %o', (status, view) => {
		expect(offersApprovedLook(/** @type {any} */ (status), view)).toBe(false);
	});

	it('says which approved version the picture shows, or that there is none', () => {
		const fingerprint = `77b0d4a${'0'.repeat(57)}`;
		expect(pictureLine('showing', fingerprint)).toBe('Showing approved 77b0d4a');
		expect(pictureLine('missing', fingerprint)).toBe('No picture of the approved version');
	});
});

