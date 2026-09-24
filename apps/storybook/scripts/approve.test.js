import { spawnSync } from 'node:child_process';
import {
	copyFileSync,
	existsSync,
	mkdirSync,
	mkdtempSync,
	readdirSync,
	readFileSync,
	writeFileSync
} from 'node:fs';
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
const SIZES = ['compact', 'medium', 'wide', 'ultra'];
/** The markup digest the stand-in run records; the fingerprint is FP. */
const DIGEST = 'd'.repeat(64);
/** A version approved before, whose pictures every test starts with. */
const OLD = 'b'.repeat(64);
const seeded = Object.fromEntries(SIZES.map((size) => [`/${POPULATED}/${OLD}-${size}.png`, `old ${size}`]));

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

/** What the approve run leaves behind: a run file and the pictures it took. */
const work = (name = '') => join(dir, 'work', name);
const approvedDir = (name = '') => join(dir, 'canvas', 'approved', name);

/**
 * Lays out what a regenerate step would leave, under `prepared/<name>`, and
 * returns the command that puts it in place.
 * @param {string} name
 * @param {{ story?: string, stories?: string[], digest?: string, pictured?: string, sizes?: string[] }} [run]
 */
function preparedRun(
	name,
	{ story = POPULATED, stories = [story], digest = DIGEST, pictured = digest, sizes = SIZES } = {}
) {
	const root = join(dir, 'prepared', name);
	mkdirSync(join(root, 'pictures'), { recursive: true });
	writeFileSync(
		join(root, 'run.jsonl'),
		stories.map((id) => `${JSON.stringify({ id, digest })}\n`).join('')
	);
	writeFileSync(
		join(root, 'pictures', 'pictures.jsonl'),
		stories
			.flatMap((id) => sizes.map((size) => `${JSON.stringify({ story: id, size, digest: pictured })}\n`))
			.join('')
	);
	for (const id of stories) {
		mkdirSync(join(root, 'pictures', id), { recursive: true });
		for (const size of sizes) writeFileSync(join(root, 'pictures', id, `${size}.png`), `${name} ${size}`);
	}
	return `cp -R ${root}/. ${work()}`;
}

/** Every file under canvas/approved/, with its contents. */
function approvedTree() {
	if (!existsSync(approvedDir())) return {};
	return Object.fromEntries(
		readdirSync(approvedDir(), { recursive: true, withFileTypes: true })
			.filter((e) => e.isFile())
			.map((e) => {
				const path = join(e.parentPath, e.name);
				return [path.slice(approvedDir().length), readFileSync(path, 'utf8')];
			})
	);
}

beforeEach(() => {
	dir = mkdtempSync(join(tmpdir(), 'approve-'));
	mkdirSync(work(), { recursive: true });
	mkdirSync(join(dir, 'canvas'));
	copyFileSync(join(fixtures, 'index.json'), at('index.json'));
	writeFileSync(at('versions.json'), versionsWith(FP, { [FOCUS]: entry('c'.repeat(64)) }));
	writeFileSync(at('approvals.jsonl'), '');
	mkdirSync(approvedDir(POPULATED), { recursive: true });
	for (const size of SIZES) writeFileSync(approvedDir(`${POPULATED}/${OLD}-${size}.png`), `old ${size}`);
	spawnSync('git', ['init', '-q'], { cwd: dir, env: hermetic });
	spawnSync('git', ['config', 'user.name', 'Emanuel Fonseca'], { cwd: dir, env: hermetic });
	spawnSync('git', ['config', 'user.email', 'e@example.com'], { cwd: dir, env: hermetic });
});

/**
 * @param {string[]} words the story, then the note
 * @param {{ answer?: string, tty?: boolean, env?: Record<string, string>, regenerate?: string }} [options]
 */
function approve(words, { answer = 'yes', tty = true, env = {}, regenerate = preparedRun('current') } = {}) {
	const args = [
		script,
		'--approve',
		'--file', at('approvals.jsonl'),
		'--versions', at('versions.json'),
		'--index', at('index.json'),
		'--run', work('run.jsonl'),
		'--pictures', work('pictures'),
		'--approved', approvedDir(),
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
	expect(approvedTree()).toEqual(seeded);
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
		expect(approvedTree()).toEqual(seeded);
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
		expect(result.stdout).toContain(
			'Commit apps/storybook/canvas/approvals.jsonl and apps/storybook/canvas/approved/pages-overview--populated/ together, on their own.'
		);
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

describe('keeping pictures of the version approved', () => {
	it('keeps one picture per size, named by the fingerprint approved', () => {
		const result = approve([POPULATED]);
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('Recorded. Emanuel approved pages-overview--populated at aaaaaaa.');
		expect(result.stdout).toContain('Kept a picture at compact, medium, wide and ultra.');
		expect(result.stdout).toContain(
			'Commit apps/storybook/canvas/approvals.jsonl and apps/storybook/canvas/approved/pages-overview--populated/ together, on their own.'
		);
		expect(approvedTree()).toEqual(
			Object.fromEntries(SIZES.map((size) => [`/${POPULATED}/${FP}-${size}.png`, `current ${size}`]))
		);
		expect(approvals().trimEnd().split('\n')).toHaveLength(1);
	});

	it('refuses, recording nothing, when the run did not picture every size', () => {
		const result = approve([POPULATED], {
			regenerate: preparedRun('short', { sizes: ['compact', 'wide', 'ultra'] })
		});
		refused(result, /The approve run took no picture at medium\. Nothing recorded\./);
		expect(approvals()).toBe('');
	});

	it('refuses, recording nothing, a picture of a render other than the one recorded', () => {
		const result = approve([POPULATED], {
			regenerate: preparedRun('other', { pictured: 'f'.repeat(64) })
		});
		refused(result, /The approve run took no picture at compact, medium, wide and ultra\. Nothing recorded\./);
		expect(approvals()).toBe('');
	});

	it('leaves only the pictures of a newer version approved', () => {
		const line = `${JSON.stringify({ story: POPULATED, fingerprint: OLD, name: 'Grace Hopper', email: 'g@example.com', at: '2026-09-18T10:00:00Z', note: '' })}\n`;
		writeFileSync(at('approvals.jsonl'), line);
		writeFileSync(
			at('versions.json'),
			JSON.stringify({
				[POPULATED]: {
					implemented: FP,
					versions: [
						{ fingerprint: FP, firstSeen: '2026-09-20' },
						{ fingerprint: OLD, firstSeen: '2026-09-18' }
					]
				}
			})
		);
		expect(approve([POPULATED]).status).toBe(0);
		expect(Object.keys(approvedTree()).sort()).toEqual(
			SIZES.map((size) => `/${POPULATED}/${FP}-${size}.png`).sort()
		);
	});

	it('keeps one picture per size when a second person approves the same version', () => {
		expect(approve([POPULATED]).status).toBe(0);
		spawnSync('git', ['config', 'user.email', 'g@example.com'], { cwd: dir, env: hermetic });
		expect(approve([POPULATED], { regenerate: preparedRun('again') }).status).toBe(0);
		expect(approvedTree()).toEqual(
			Object.fromEntries(SIZES.map((size) => [`/${POPULATED}/${FP}-${size}.png`, `again ${size}`]))
		);
		expect(approvals().trimEnd().split('\n')).toHaveLength(2);
	});

	it('keeps one picture for a story drawn at compact only', () => {
		const NARROW = 'pages-overview--narrow';
		const index = JSON.parse(readFileSync(at('index.json'), 'utf8'));
		index.entries[NARROW] = {
			id: NARROW,
			type: 'story',
			title: 'Pages/Overview',
			name: 'Narrow',
			tags: ['kind-state', 'size-compact']
		};
		writeFileSync(at('index.json'), JSON.stringify(index));
		writeFileSync(at('versions.json'), versionsWith(FP, { [NARROW]: entry(FP) }));
		const result = approve([NARROW], {
			regenerate: preparedRun('narrow', { story: NARROW, sizes: ['compact'] })
		});
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('Kept a picture at compact.');
		expect(Object.keys(approvedTree())).toContain(`/${NARROW}/${FP}-compact.png`);
		expect(Object.keys(approvedTree()).filter((p) => p.startsWith(`/${NARROW}/`))).toHaveLength(1);
	});
});

describe('approving several page stories in one run', () => {
	const both = { [EMPTY]: entry(FP) };

	it('lists them, asks once, and records a line and pictures for each', () => {
		writeFileSync(at('versions.json'), versionsWith(FP, both));
		const result = approve([POPULATED, EMPTY, 'a', 'baseline'], {
			regenerate: preparedRun('two', { stories: [POPULATED, EMPTY] })
		});
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('Pages/Overview · Populated at aaaaaaa, first seen 20 Sep 2026');
		expect(result.stdout).toContain('Pages/Overview · Empty at aaaaaaa, first seen 20 Sep 2026');
		expect(result.stdout).toContain('Approve these 2 page stories? yes/no');
		expect(result.stdout).toContain('Recorded. Emanuel approved 2 page stories.');
		expect(result.stdout).toContain('Kept pictures of each at the sizes it is drawn at.');
		expect(result.stdout).toContain(
			'Commit apps/storybook/canvas/approvals.jsonl and apps/storybook/canvas/approved/ together, on their own.'
		);
		const lines = approvals().trimEnd().split('\n').map((l) => JSON.parse(l));
		expect(lines.map((l) => [l.story, l.note])).toEqual([
			[POPULATED, 'a baseline'],
			[EMPTY, 'a baseline']
		]);
		expect(Object.keys(approvedTree()).sort()).toEqual(
			[POPULATED, EMPTY].flatMap((id) => SIZES.map((size) => `/${id}/${FP}-${size}.png`)).sort()
		);
	});

	it('refuses the whole run when one story is refused, naming it', () => {
		writeFileSync(at('versions.json'), versionsWith(FP, both));
		const result = approve([POPULATED, 'pages-overview--nope', EMPTY], {
			regenerate: preparedRun('two', { stories: [POPULATED, EMPTY] })
		});
		refused(result, /No page story "pages-overview--nope"\..*Nothing recorded\./s);
		expect(approvals()).toBe('');
	});

	it('refuses named stories and --needs together', () => {
		refused(approve([POPULATED, '--needs']), /Name stories or use --needs, not both\. Nothing recorded\./);
		expect(approvals()).toBe('');
	});

	it('with --needs, approves what was changed or never approved, and nothing else', () => {
		// Populated changed since OLD was approved, Empty was never approved,
		// Focus only asserts something, and Narrow is approved as it is.
		const index = JSON.parse(readFileSync(at('index.json'), 'utf8'));
		index.entries['pages-overview--narrow'] = {
			id: 'pages-overview--narrow',
			type: 'story',
			title: 'Pages/Overview',
			name: 'Narrow',
			tags: ['kind-state']
		};
		writeFileSync(at('index.json'), JSON.stringify(index));
		writeFileSync(
			at('versions.json'),
			JSON.stringify({
				[POPULATED]: {
					implemented: FP,
					versions: [
						{ fingerprint: FP, firstSeen: '2026-09-20' },
						{ fingerprint: OLD, firstSeen: '2026-09-18' }
					]
				},
				[EMPTY]: entry(FP),
				[FOCUS]: entry(FP),
				'pages-overview--narrow': entry(FP)
			})
		);
		const approved = (story, fingerprint) =>
			`${JSON.stringify({ story, fingerprint, name: 'Grace Hopper', email: 'g@example.com', at: '2026-09-18T10:00:00Z', note: '' })}\n`;
		const before = approved(POPULATED, OLD) + approved('pages-overview--narrow', FP);
		writeFileSync(at('approvals.jsonl'), before);
		const result = approve(['--needs'], {
			regenerate: preparedRun('needs', { stories: [POPULATED, EMPTY] })
		});
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('Approve these 2 page stories? yes/no');
		const added = approvals().slice(before.length).trimEnd().split('\n').map((l) => JSON.parse(l).story);
		expect(added).toEqual([POPULATED, EMPTY]);
	});

	it('with --needs and nothing to approve, records nothing and says so', () => {
		const line = `${JSON.stringify({ story: POPULATED, fingerprint: FP, name: 'Grace Hopper', email: 'g@example.com', at: '2026-09-18T10:00:00Z', note: '' })}\n`;
		writeFileSync(at('approvals.jsonl'), line);
		writeFileSync(at('versions.json'), versionsWith(FP, { [FOCUS]: entry(FP) }));
		const index = JSON.parse(readFileSync(at('index.json'), 'utf8'));
		delete index.entries[EMPTY];
		writeFileSync(at('index.json'), JSON.stringify(index));
		const result = approve(['--needs']);
		expect(result.status).toBe(0);
		expect(result.stdout).toContain('Nothing needs approval. Nothing recorded.');
		expect(approvals()).toBe(line);
		expect(approvedTree()).toEqual(seeded);
	});
});

