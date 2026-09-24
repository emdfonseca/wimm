import { spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { beforeEach, describe, expect, it } from 'vitest';

const script = join(import.meta.dirname, 'approvals.mjs');
const FP = 'a'.repeat(64);
const STORY = 'pages-overview--populated';
const VERSIONS = {
	[STORY]: { implemented: FP, versions: [{ fingerprint: FP, firstSeen: '2026-09-20' }] }
};

const hermetic = {
	...process.env,
	GIT_CONFIG_GLOBAL: '/dev/null',
	GIT_CONFIG_SYSTEM: '/dev/null',
	GIT_AUTHOR_NAME: 't',
	GIT_AUTHOR_EMAIL: 't@example.com',
	GIT_COMMITTER_NAME: 't',
	GIT_COMMITTER_EMAIL: 't@example.com'
};

/** @param {string} cwd @param {...string} args */
function git(cwd, ...args) {
	const result = spawnSync('git', args, { cwd, env: hermetic, encoding: 'utf8' });
	if (result.status !== 0) throw new Error(`git ${args.join(' ')}: ${result.stderr}`);
	return result.stdout.trim();
}

/** @param {string} email @param {string} at */
const approval = (email, at = '2026-09-20T18:04:11Z') =>
	`${JSON.stringify({ story: STORY, fingerprint: FP, name: email, email, at, note: '' })}\n`;

/** @type {string} */
let dir;
/** @type {string} */
let file;

beforeEach(() => {
	dir = mkdtempSync(join(tmpdir(), 'approvals-'));
	mkdirSync(join(dir, 'canvas'));
	file = join(dir, 'canvas', 'approvals.jsonl');
	writeFileSync(join(dir, 'canvas', 'versions.json'), JSON.stringify(VERSIONS));
});

const initRepo = () => {
	git(dir, 'init', '-q', '-b', 'main');
	git(dir, 'add', '.');
	git(dir, 'commit', '-q', '-m', 'first');
};

const commitAll = (message = 'next') => {
	git(dir, 'add', '.');
	git(dir, 'commit', '-q', '-m', message);
};

const check = () =>
	spawnSync(
		'node',
		[script, '--check', '--file', file, '--versions', join(dir, 'canvas', 'versions.json')],
		{ encoding: 'utf8', env: { ...hermetic, GIT_CEILING_DIRECTORIES: tmpdir() } }
	);

describe('--check content', () => {
	it('passes an empty record', () => {
		writeFileSync(file, '');
		expect(check().status).toBe(0);
	});

	it('fails on a line that names a story that does not exist, naming the line', () => {
		writeFileSync(file, approval('a@example.com').replace(STORY, 'pages-x--y'));
		const result = check();
		expect(result.status).not.toBe(0);
		expect(result.stderr).toContain('line 1: story "pages-x--y" is not in versions.json');
	});
});

describe('--check against what is committed', () => {
	const first = approval('a@example.com');
	const second = approval('b@example.com', '2026-09-21T09:00:00Z');

	beforeEach(() => {
		writeFileSync(file, first + second);
		initRepo();
	});

	it('passes a file that is as committed', () => {
		expect(check().status).toBe(0);
	});

	it('passes a line appended at the end', () => {
		writeFileSync(file, first + second + approval('c@example.com', '2026-09-22T09:00:00Z'));
		expect(check().status).toBe(0);
	});

	it('fails an edited committed line, naming it', () => {
		writeFileSync(file, first + second.replace('b@example.com', 'x@example.com'));
		const result = check();
		expect(result.status).not.toBe(0);
		expect(result.stderr).toContain('line 2: a committed approval was changed');
	});

	it('fails a removed line', () => {
		writeFileSync(file, first);
		const result = check();
		expect(result.status).not.toBe(0);
		expect(result.stderr).toContain('line 2: a committed approval is missing');
	});

	it('fails a line inserted in the middle', () => {
		writeFileSync(file, first + approval('c@example.com', '2026-09-22T09:00:00Z') + second);
		const result = check();
		expect(result.status).not.toBe(0);
		expect(result.stderr).toContain('line 2: a committed approval was changed');
	});

	it('fails a branch that rewrote a line the main branch already had', () => {
		git(dir, 'update-ref', 'refs/remotes/origin/main', 'HEAD');
		git(dir, 'checkout', '-q', '-b', 'topic');
		writeFileSync(file, first + second.replace('b@example.com', 'x@example.com'));
		commitAll('rewrite');
		const result = check();
		expect(result.status).not.toBe(0);
		expect(result.stderr).toContain('changed on this branch');
	});
});

describe('--check with nothing to compare against', () => {
	it('passes and says why when the file was never committed', () => {
		writeFileSync(file, approval('a@example.com'));
		writeFileSync(join(dir, 'other.txt'), 'x');
		git(dir, 'init', '-q', '-b', 'main');
		git(dir, 'add', 'other.txt', 'canvas/versions.json');
		git(dir, 'commit', '-q', '-m', 'first');
		const result = check();
		expect(result.status).toBe(0);
		expect(result.stdout).toMatch(/append-only comparison skipped: .*not committed/);
	});

	it('passes and says why when there is no git at all', () => {
		writeFileSync(file, approval('a@example.com'));
		const result = check();
		expect(result.status).toBe(0);
		expect(result.stdout).toMatch(/append-only comparison skipped: .*not in a git/);
		expect(readFileSync(file, 'utf8')).toBe(approval('a@example.com'));
	});
});
