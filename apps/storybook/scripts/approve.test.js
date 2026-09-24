import { spawnSync } from 'node:child_process';
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { beforeEach, describe, expect, it } from 'vitest';

const script = join(import.meta.dirname, 'approvals.mjs');
const fixtures = join(import.meta.dirname, 'fixtures');
const FP = 'a'.repeat(64);
const MOVED = 'e'.repeat(64);
const POPULATED = 'pages-overview--populated';
const EMPTY = 'pages-overview--empty';
const FOCUS = 'pages-overview--focus';

const entry = (fp, firstSeen = '2026-09-20') => ({
	implemented: fp,
	versions: [{ fingerprint: fp, firstSeen }]
});
const versionsWith = (populated, more = {}) =>
	JSON.stringify({ [POPULATED]: entry(populated), ...more });

const hermetic = {
	...process.env,
	GIT_CONFIG_GLOBAL: '/dev/null',
	GIT_CONFIG_SYSTEM: '/dev/null'
};
delete hermetic.CLAUDECODE;
delete hermetic.CLAUDE_CODE_ENTRYPOINT;

/** Runs a command under a pseudo-terminal, answering the confirmation. */
const PTY = `
import os, pty, select, sys
pid, fd = pty.fork()
if pid == 0:
    os.execvpe(sys.argv[1], sys.argv[1:], os.environ)
out, sent = b'', False
while True:
    try:
        ready, _, _ = select.select([fd], [], [], 30)
        if not ready: break
        data = os.read(fd, 4096)
    except OSError:
        break
    if not data: break
    out += data
    if not sent and b'yes/no' in out:
        os.write(fd, os.environ['ANSWER'].encode() + b'\\n')
        sent = True
_, status = os.waitpid(pid, 0)
sys.stdout.write(out.decode(errors='replace'))
sys.exit(os.waitstatus_to_exitcode(status))
`;

/** @type {string} */
let dir;
const at = (name) => join(dir, 'canvas', name);
const approvals = () => readFileSync(at('approvals.jsonl'), 'utf8');

beforeEach(() => {
	dir = mkdtempSync(join(tmpdir(), 'approve-'));
	mkdirSync(join(dir, 'canvas'));
	copyFileSync(join(fixtures, 'index.json'), at('index.json'));
	writeFileSync(at('versions.json'), versionsWith(FP, { [FOCUS]: entry('c'.repeat(64)) }));
	writeFileSync(at('approvals.jsonl'), '');
	spawnSync('git', ['init', '-q'], { cwd: dir, env: hermetic });
	spawnSync('git', ['config', 'user.name', 'Emanuel Fonseca'], { cwd: dir, env: hermetic });
	spawnSync('git', ['config', 'user.email', 'e@example.com'], { cwd: dir, env: hermetic });
});

/**
 * @param {string[]} words the story, then the note
 * @param {{ answer?: string, tty?: boolean, env?: Record<string, string>, regenerate?: string }} [options]
 */
function approve(words, { answer = 'yes', tty = true, env = {}, regenerate = 'true' } = {}) {
	const args = [
		script,
		'--approve',
		'--file', at('approvals.jsonl'),
		'--versions', at('versions.json'),
		'--index', at('index.json'),
		'--regenerate', regenerate,
		'--today', '2026-09-20',
		...words
	];
	const options = { encoding: /** @type {const} */ ('utf8'), env: { ...hermetic, ...env } };
	return tty
		? spawnSync('python3', ['-c', PTY, 'node', ...args], {
				...options,
				env: { ...options.env, ANSWER: answer }
			})
		: spawnSync('node', args, { ...options, input: '' });
}

/** @param {ReturnType<typeof approve>} result @param {RegExp} message */
function refused(result, message) {
	expect(result.status).not.toBe(0);
	expect(`${result.stdout}${result.stderr}`).toMatch(message);
}

describe('approving refuses, recording nothing', () => {
	it('inside an agent session', () => {
		refused(approve([POPULATED], { env: { CLAUDECODE: '1' } }), /recorded by a person in their own terminal\. Nothing recorded\./);
		expect(approvals()).toBe('');
	});

	it('inside an agent session by its entrypoint', () => {
		refused(approve([POPULATED], { env: { CLAUDE_CODE_ENTRYPOINT: 'cli' } }), /own terminal/);
		expect(approvals()).toBe('');
	});

	it('when stdin is not a terminal', () => {
		refused(approve([POPULATED], { tty: false }), /recorded by a person in their own terminal\. Nothing recorded\./);
		expect(approvals()).toBe('');
	});

	it('for a story that does not exist, offering the closest', () => {
		refused(
			approve(['pages-overveiw--populated']),
			/No page story "pages-overveiw--populated"\. Closest: pages-overview--populated\./
		);
		expect(approvals()).toBe('');
	});

	it('for a story the versions file did not have before the regenerate', () => {
		writeFileSync(at('after.json'), versionsWith(FP, { [EMPTY]: entry('b'.repeat(64)) }));
		refused(
			approve([EMPTY], { regenerate: `cp ${at('after.json')} ${at('versions.json')}` }),
			/pages-overview--empty has no version yet\. Run just gen apps\/storybook, then approve\./
		);
		expect(approvals()).toBe('');
	});

	it('when the regenerate step moved the version, naming the new one', () => {
		writeFileSync(at('after.json'), versionsWith(MOVED));
		refused(
			approve([POPULATED], { regenerate: `cp ${at('after.json')} ${at('versions.json')}` }),
			/pages-overview--populated now renders eeeeeee, first seen today\. Look again, then approve\. Nothing recorded\./
		);
		expect(approvals()).toBe('');
	});

	it('when the same person already approved that version', () => {
		const line = `${JSON.stringify({ story: POPULATED, fingerprint: FP, name: 'Emanuel Fonseca', email: 'e@example.com', at: '2026-09-18T10:00:00Z', note: '' })}\n`;
		writeFileSync(at('approvals.jsonl'), line);
		refused(approve([POPULATED]), /You approved this version on 18 Sep 2026\./);
		expect(approvals()).toBe(line);
	});

	it('for a story that only asserts something', () => {
		refused(approve([FOCUS]), /pages-overview--focus only asserts something and needs no approval\./);
		expect(approvals()).toBe('');
	});

	it('when git has no identity', () => {
		spawnSync('git', ['config', '--unset', 'user.email'], { cwd: dir, env: hermetic });
		refused(approve([POPULATED]), /git has no user\.name or user\.email set\. Nothing recorded\./);
		expect(approvals()).toBe('');
	});

	it('when the answer is no', () => {
		const result = approve([POPULATED], { answer: 'no' });
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('Nothing recorded.');
		expect(approvals()).toBe('');
	});
});

describe('approving as a person', () => {
	it('appends exactly one line and moves nothing before it', () => {
		const earlier = `${JSON.stringify({ story: POPULATED, fingerprint: FP, name: 'Grace Hopper', email: 'g@example.com', at: '2026-09-18T10:00:00Z', note: 'looks right' })}\n`;
		writeFileSync(at('approvals.jsonl'), earlier);
		const result = approve([POPULATED, 'after', 'the', 'rows', 'were', 'tightened']);
		expect(result.status).toBe(0);
		expect(result.stdout).toContain(
			'Approve Pages/Overview · Populated at aaaaaaa, first seen 20 Sep 2026? yes/no'
		);
		expect(result.stdout).toContain('Recorded. Emanuel approved pages-overview--populated at aaaaaaa.');
		expect(result.stdout).toContain('Commit apps/storybook/canvas/approvals.jsonl on its own.');
		const text = approvals();
		expect(text.startsWith(earlier)).toBe(true);
		const added = text.slice(earlier.length);
		expect(added.endsWith('\n')).toBe(true);
		expect(added.trimEnd().split('\n')).toHaveLength(1);
		const record = JSON.parse(added);
		expect(record).toMatchObject({
			story: POPULATED,
			fingerprint: FP,
			name: 'Emanuel Fonseca',
			email: 'e@example.com',
			note: 'after the rows were tightened'
		});
		expect(record.at).toMatch(/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$/);
	});
});
