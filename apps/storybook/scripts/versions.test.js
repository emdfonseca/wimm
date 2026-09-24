import { spawnSync } from 'node:child_process';
import { copyFileSync, mkdtempSync, readFileSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { beforeEach, describe, expect, it } from 'vitest';
import { sha256Hex } from '../canvas/lib.js';

const script = join(import.meta.dirname, 'versions.mjs');
const fixtures = join(import.meta.dirname, 'fixtures');
const GLOBAL = 'd'.repeat(64);
const [A, B, C] = ['a', 'b', 'c'].map((c) => c.repeat(64));
const POPULATED = 'pages-overview--populated';
const EMPTY = 'pages-overview--empty';
const FOCUS = 'pages-overview--focus';

/** @param {string} digest */
const fingerprintOf = (digest) => sha256Hex(`${digest}\n${GLOBAL}`);

/** @type {string} */
let dir;
const path = (name) => join(dir, name);

beforeEach(() => {
	dir = mkdtempSync(join(tmpdir(), 'versions-'));
	copyFileSync(join(fixtures, 'index.json'), path('index.json'));
});

/**
 * @param {string} runFixture
 * @param {{ today?: string, index?: string }} [options]
 */
function generate(runFixture, { today = '2026-09-20', index = path('index.json') } = {}) {
	if (!runFixture.includes('\n')) copyFileSync(join(fixtures, runFixture), path('run.jsonl'));
	else writeFileSync(path('run.jsonl'), runFixture);
	return spawnSync(
		'node',
		[
			script,
			'--run', path('run.jsonl'),
			'--index', index,
			'--versions', path('versions.json'),
			'--approvals', path('approvals.jsonl'),
			'--global-digest', GLOBAL,
			'--today', today
		],
		{ encoding: 'utf8' }
	);
}

const runOf = (populated, extra = '') =>
	`{"id":"${POPULATED}","digest":"${populated}"}\n{"id":"${EMPTY}","digest":"${B}"}\n{"id":"${FOCUS}","digest":"${C}"}\n${extra}{"run":"passed"}\n`;

const read = () => JSON.parse(readFileSync(path('versions.json'), 'utf8'));

describe('a run it must refuse', () => {
	const SENTINEL = '{"sentinel":true}\n';
	it.each([
		['a run file missing a page story', 'run-missing.jsonl', /pages-overview--focus.*no record/],
		['a story recorded twice', 'run-duplicate.jsonl', /pages-overview--populated.*more than once/],
		['a digest that is not 64 hex', 'run-bad-digest.jsonl', /pages-overview--populated.*64 hex/],
		['a run marked failed', 'run-failed.jsonl', /run did not pass/]
	])('%s: fails and leaves versions.json alone', (_name, fixture, message) => {
		writeFileSync(path('versions.json'), SENTINEL);
		const result = generate(fixture);
		expect(result.status).not.toBe(0);
		expect(result.stderr).toMatch(message);
		expect(readFileSync(path('versions.json'), 'utf8')).toBe(SENTINEL);
	});

	it('does not create versions.json when it refuses', () => {
		const result = generate('run-missing.jsonl');
		expect(result.status).not.toBe(0);
		expect(() => statSync(path('versions.json'))).toThrow();
	});

	it('refuses an approvals record it cannot read, rather than dropping what it might name', () => {
		writeFileSync(path('versions.json'), SENTINEL);
		writeFileSync(path('approvals.jsonl'), 'not json\n');
		const result = generate('run-good.jsonl');
		expect(result.status).not.toBe(0);
		expect(result.stderr).toMatch(/approvals\.jsonl line 1/);
		expect(readFileSync(path('versions.json'), 'utf8')).toBe(SENTINEL);
	});
});

describe('writing versions', () => {
	it('gives every page story a version dated today and skips a molecule', async () => {
		const result = generate('run-good.jsonl');
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('versions.json changed, commit it');
		const versions = read();
		expect(Object.keys(versions)).toEqual([EMPTY, FOCUS, POPULATED]);
		expect(versions[POPULATED]).toEqual({
			implemented: await fingerprintOf(A),
			versions: [{ fingerprint: await fingerprintOf(A), firstSeen: '2026-09-20' }]
		});
	});

	it('writes nothing when nothing moved, and says so', () => {
		generate('run-good.jsonl');
		const before = statSync(path('versions.json'));
		const bytes = readFileSync(path('versions.json'), 'utf8');
		const result = generate('run-good.jsonl', { today: '2027-01-01' });
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('versions.json unchanged');
		expect(readFileSync(path('versions.json'), 'utf8')).toBe(bytes);
		expect(statSync(path('versions.json')).mtimeMs).toBe(before.mtimeMs);
	});

	it('names the stories that moved, and keeps firstSeen for the ones that did not', async () => {
		generate('run-good.jsonl');
		const result = generate(runOf('e'.repeat(64)), { today: '2026-10-01' });
		expect(result.stdout).toContain(POPULATED);
		expect(result.stdout).not.toContain(EMPTY);
		expect(result.stdout).toContain('versions.json changed, commit it');
		const versions = read();
		expect(versions[POPULATED].implemented).toBe(await fingerprintOf('e'.repeat(64)));
		expect(versions[POPULATED].versions[0].firstSeen).toBe('2026-10-01');
		expect(versions[EMPTY].versions[0].firstSeen).toBe('2026-09-20');
	});

	it('drops the replaced version when nobody approved it', () => {
		generate('run-good.jsonl');
		generate(runOf('e'.repeat(64)), { today: '2026-10-01' });
		expect(read()[POPULATED].versions).toHaveLength(1);
	});

	it('keeps the replaced version when the record approves it', async () => {
		generate('run-good.jsonl');
		const old = await fingerprintOf(A);
		writeFileSync(path('approvals.jsonl'), `${JSON.stringify({ story: POPULATED, fingerprint: old })}\n`);
		generate(runOf('e'.repeat(64)), { today: '2026-10-01' });
		const versions = read()[POPULATED];
		expect(versions.versions.map((v) => v.fingerprint)).toEqual([
			await fingerprintOf('e'.repeat(64)),
			old
		]);
		expect(versions.versions[1].firstSeen).toBe('2026-09-20');
	});

	it('returns to an approved version with its original date when the page goes back', async () => {
		generate('run-good.jsonl');
		const old = await fingerprintOf(A);
		writeFileSync(path('approvals.jsonl'), `${JSON.stringify({ story: POPULATED, fingerprint: old })}\n`);
		generate(runOf('e'.repeat(64)), { today: '2026-10-01' });
		generate(runOf(A), { today: '2026-10-02' });
		const versions = read()[POPULATED];
		expect(versions.implemented).toBe(old);
		expect(versions.versions[0]).toEqual({ fingerprint: old, firstSeen: '2026-09-20' });
	});
});

describe('a story gone from the index', () => {
	const without = () => {
		const index = JSON.parse(readFileSync(join(fixtures, 'index.json'), 'utf8'));
		delete index.entries[EMPTY];
		writeFileSync(path('index-without.json'), JSON.stringify(index));
		return path('index-without.json');
	};
	const runWithout = `{"id":"${POPULATED}","digest":"${A}"}\n{"id":"${FOCUS}","digest":"${C}"}\n{"run":"passed"}\n`;

	it('is dropped when nobody approved it', () => {
		generate('run-good.jsonl');
		generate(runWithout, { index: without() });
		expect(Object.keys(read())).toEqual([FOCUS, POPULATED]);
	});

	it('is kept as gone, with only what was approved, when somebody did', async () => {
		generate('run-good.jsonl');
		const fp = await fingerprintOf(B);
		writeFileSync(path('approvals.jsonl'), `${JSON.stringify({ story: EMPTY, fingerprint: fp })}\n`);
		generate(runWithout, { index: without() });
		expect(read()[EMPTY]).toEqual({
			gone: true,
			versions: [{ fingerprint: fp, firstSeen: '2026-09-20' }]
		});
	});
});
