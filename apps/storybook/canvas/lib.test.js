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
	sitemapFrom,
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
