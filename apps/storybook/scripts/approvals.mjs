// The approvals record, apps/storybook/canvas/approvals.jsonl: --check holds it
// to its shape and to only ever growing.
import { spawnSync } from 'node:child_process';
import {
	appendFileSync,
	copyFileSync,
	existsSync,
	mkdirSync,
	readFileSync,
	readdirSync,
	renameSync,
	rmSync
} from 'node:fs';
import { basename, dirname, join, relative, resolve, sep } from 'node:path';
import { createInterface } from 'node:readline/promises';
import { parseArgs } from 'node:util';
import {
	PICTURES_DIR,
	RUN_FILE,
	approvalLines,
	approvalProblems,
	closestIds,
	formatDate,
	kindOf,
	needsApproval,
	parseApprovals,
	pictureName,
	pictureProblems,
	prefixProblems,
	shortVersion,
	sizesFor,
	statusOf
} from '../canvas/lib.js';
import { viewportOptions } from '../canvas/viewports.js';

const app = resolve(import.meta.dirname, '..');
const { values, positionals } = parseArgs({
	options: {
		check: { type: 'boolean', default: false },
		approve: { type: 'boolean', default: false },
		needs: { type: 'boolean', default: false },
		file: { type: 'string', default: join(app, 'canvas/approvals.jsonl') },
		versions: { type: 'string', default: join(app, 'canvas/versions.json') },
		index: { type: 'string' },
		run: { type: 'string', default: join(app, RUN_FILE) },
		pictures: { type: 'string', default: join(app, PICTURES_DIR) },
		approved: { type: 'string', default: join(app, 'canvas/approved') },
		regenerate: { type: 'string', default: 'just gen' },
		today: { type: 'string', default: new Date().toISOString().slice(0, 10) }
	},
	allowPositionals: true
});

/** @param {string} line */
const say = (line) => process.stdout.write(`${line}\n`);

/** @param {string} cwd @param {string[]} args */
function git(cwd, args) {
	const result = spawnSync('git', args, { cwd, encoding: 'utf8', maxBuffer: 64 * 1024 * 1024 });
	return result.status === 0 ? { ok: true, out: result.stdout } : { ok: false, out: '' };
}

/**
 * The problems with the record against its committed copies, or why there was
 * nothing to compare. Never a failure for a missing copy: a fresh clone is not a defect.
 * @param {string} file
 * @param {string[]} now
 * @returns {{ problems: string[], skipped?: string }}
 */
function againstGit(file, now) {
	const cwd = dirname(file);
	const prefix = git(cwd, ['rev-parse', '--show-prefix']);
	if (!prefix.ok) return { problems: [], skipped: 'not in a git repository' };
	const path = `${prefix.out.trim()}${basename(file)}`;
	const head = git(cwd, ['show', `HEAD:${path}`]);
	if (!head.ok) return { problems: [], skipped: 'the record is not committed yet' };
	const atHead = approvalLines(head.out);
	const problems = prefixProblems(atHead, now, {
		changed: 'a committed approval was changed; committed approvals are never changed',
		missing: 'a committed approval is missing; committed approvals are never removed'
	});

	// A rewriting commit on a branch is caught here: HEAD itself holds the
	// rewrite, so it can only be seen against where the branch left main.
	const main = git(cwd, ['rev-parse', '--verify', '--quiet', 'origin/main']);
	if (main.ok && !git(cwd, ['merge-base', '--is-ancestor', 'HEAD', 'origin/main']).ok) {
		const base = git(cwd, ['merge-base', 'HEAD', 'origin/main']);
		const atBase = base.ok ? git(cwd, ['show', `${base.out.trim()}:${path}`]) : { ok: false, out: '' };
		if (atBase.ok) {
			problems.push(
				...prefixProblems(approvalLines(atBase.out), atHead, {
					changed: 'an approval on main was changed on this branch',
					missing: 'an approval on main is missing on this branch'
				})
			);
		}
	}
	return { problems };
}

/**
 * Every file under the kept pictures, relative to them, and every `.next`
 * folder, even an empty one: an approval that did not finish can leave either.
 * @param {string} root
 */
function keptPaths(root) {
	if (!existsSync(root)) return [];
	return readdirSync(root, { recursive: true, withFileTypes: true }).flatMap((e) => {
		const path = relative(root, join(e.parentPath, e.name)).split(sep).join('/');
		if (e.isFile()) return [path];
		return e.isDirectory() && e.parentPath === root && e.name.endsWith('.next') ? [`${path}/`] : [];
	});
}

if (values.check) {
	const text = existsSync(values.file) ? readFileSync(values.file, 'utf8') : '';
	const lines = approvalLines(text);
	const versions = existsSync(values.versions) ? JSON.parse(readFileSync(values.versions, 'utf8')) : {};
	const { problems, skipped } = againstGit(values.file, lines);
	problems.unshift(...approvalProblems(lines, versions));
	if (skipped) say(`append-only comparison skipped: ${skipped}`);
	const pictures = pictureProblems(
		keptPaths(values.approved),
		parseApprovals(text),
		Object.keys(viewportOptions)
	);
	if (problems.length > 0) {
		process.stderr.write(`approvals.jsonl:\n${problems.map((p) => `  ${p}`).join('\n')}\n`);
	}
	if (pictures.length > 0) process.stderr.write(`${pictures.join('\n')}\n`);
	if (problems.length > 0 || pictures.length > 0) process.exit(1);
	say(`approvals.jsonl: ${lines.length} approval(s), all in order`);
}

/** @param {string} why never returns: nothing is recorded */
function stop(why) {
	process.stderr.write(`${why}\n`);
	process.exit(1);
}

/** @param {string} cwd @param {string} key */
const gitConfig = (cwd, key) => git(cwd, ['config', key]).out.trim();

const readVersions = () =>
	existsSync(values.versions) ? JSON.parse(readFileSync(values.versions, 'utf8')) : {};

async function pageStories() {
	const index = values.index
		? JSON.parse(readFileSync(values.index, 'utf8'))
		: await (await import('storybook/internal/core-server')).buildIndex({
				configDir: join(app, '.storybook')
			});
	return index.entries;
}

/** Whether a person is at the keyboard: no agent session, and a terminal to answer from. */
const personAtKeyboard = () =>
	!process.env.CLAUDECODE && !process.env.CLAUDE_CODE_ENTRYPOINT && Boolean(process.stdin.isTTY);

/** Story ids look like `pages-overview--populated`; the first argument that does not starts the note. */
const STORY_ID = /^[a-z0-9][a-z0-9-]*--[a-z0-9-]+$/;

async function approve() {
	if (!personAtKeyboard()) {
		stop('Approvals are recorded by a person in their own terminal. Nothing recorded.');
	}
	const count = positionals.findIndex((word) => !STORY_ID.test(word));
	const named = count === -1 ? positionals : positionals.slice(0, count);
	const note = (count === -1 ? [] : positionals.slice(count)).join(' ');
	if (values.needs && named.length > 0) stop('Name stories or use --needs, not both. Nothing recorded.');
	if (!values.needs && named.length === 0) stop('Name a page story, or use --needs. Nothing recorded.');

	const cwd = dirname(values.file);
	const name = gitConfig(cwd, 'user.name');
	const email = gitConfig(cwd, 'user.email');
	if (!name || !email) stop('git has no user.name or user.email set. Nothing recorded.');

	const text = existsSync(values.file) ? readFileSync(values.file, 'utf8') : '';
	const earlier = parseApprovals(text);
	const beforeVersions = readVersions();
	const entries = await pageStories();
	const pages = Object.values(entries).filter(
		(e) => e.type === 'story' && e.title.startsWith('Pages/')
	);
	// What needs a look is what the canvas showed: the versions as they stood
	// before this run regenerated them.
	const stories = values.needs
		? pages
				.filter((page) =>
					needsApproval(
						statusOf(
							beforeVersions[page.id],
							earlier.filter((a) => a.story === page.id),
							kindOf(page)
						).status
					)
				)
				.map((page) => page.id)
		: named;
	if (stories.length === 0) {
		say('Nothing needs approval. Nothing recorded.');
		return;
	}

	rmSync(values.pictures, { recursive: true, force: true });
	const regenerated = spawnSync(values.regenerate, {
		shell: true,
		cwd: app,
		stdio: 'inherit',
		env: { ...process.env, WIMM_PICTURE_STORIES: stories.join(',') }
	});
	if (regenerated.status !== 0) stop('Could not regenerate the versions. Nothing recorded.');

	const versions = readVersions();
	/** @type {string[]} */
	const refusals = [];
	const approving = stories.flatMap((story) => {
		const refuse = (/** @type {string} */ why) => (refusals.push(why), []);
		const page = pages.find((e) => e.id === story);
		if (!page) {
			const near = closestIds(story, pages.map((e) => e.id));
			return refuse(`No page story "${story}".${near.length > 0 ? ` Closest: ${near.join(', ')}.` : ''}`);
		}
		const now = versions[story];
		const before = beforeVersions[story]?.implemented;
		if (!now?.implemented || before === undefined) {
			return refuse(`${story} has no version yet. Run just gen apps/storybook, then approve.`);
		}
		if (kindOf(page) === 'behaviour') {
			return refuse(`${story} only asserts something and needs no approval.`);
		}
		const firstSeen = now.versions.find((v) => v.fingerprint === now.implemented)?.firstSeen;
		if (now.implemented !== before) {
			const when = firstSeen === values.today ? 'today' : formatDate(firstSeen ?? values.today);
			return refuse(
				`${story} now renders ${shortVersion(now.implemented)}, first seen ${when}. Look again, then approve.`
			);
		}
		const mine = earlier.find(
			(r) => r.story === story && r.fingerprint === now.implemented && r.email === email
		);
		if (mine) {
			return refuse(
				stories.length === 1
					? `You approved this version on ${formatDate(mine.at)}.`
					: `You approved ${story} at this version on ${formatDate(mine.at)}.`
			);
		}
		const sizes = sizesFor(page, Object.keys(viewportOptions));
		const taken = picturesTaken(story);
		const missing = sizes.filter((size) => !taken.has(size));
		if (missing.length > 0) {
			return refuse(
				stories.length === 1
					? `The approve run took no picture at ${listWords(missing)}.`
					: `The approve run took no picture of ${story} at ${listWords(missing)}.`
			);
		}
		return [{ story, page, fingerprint: now.implemented, firstSeen, sizes, taken }];
	});
	if (refusals.length > 0) {
		stop(refusals.length === 1 ? `${refusals[0]} Nothing recorded.` : `${refusals.join('\n')}\nNothing recorded.`);
	}

	const describe = (/** @type {typeof approving[number]} */ a) =>
		`${a.page.title} · ${a.page.name} at ${shortVersion(a.fingerprint)}, first seen ${formatDate(a.firstSeen)}`;
	if (approving.length > 1) for (const a of approving) say(describe(a));
	const terminal = createInterface({ input: process.stdin, output: process.stdout });
	const answer = await terminal.question(
		approving.length === 1
			? `Approve ${describe(approving[0])}? yes/no `
			: `Approve these ${approving.length} page stories? yes/no `
	);
	terminal.close();
	if (answer.trim().toLowerCase() !== 'yes') {
		say('Nothing recorded.');
		return;
	}

	const at = new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
	const added = approving
		.map((a) => `${JSON.stringify({ story: a.story, fingerprint: a.fingerprint, name, email, at, note })}\n`)
		.join('');
	const whole = approvalLines(`${text}${added}`);
	const problems = [
		...approvalProblems(whole, versions),
		...prefixProblems(approvalLines(text), whole, {
			changed: 'a committed approval was changed',
			missing: 'a committed approval is missing'
		})
	];
	if (problems.length > 0) stop(`The new approval is not valid:\n  ${problems.join('\n  ')}\nNothing recorded.`);

	// Staged beside the kept pictures, appended, then swapped in: a failure
	// before the append leaves `.next` folders the check refuses by name.
	for (const { story, fingerprint, taken } of approving) {
		rmSync(join(values.approved, `${story}.next`), { recursive: true, force: true });
		for (const [size, file] of taken) {
			const to = join(values.approved, pictureName(`${story}.next`, fingerprint, size));
			mkdirSync(dirname(to), { recursive: true });
			copyFileSync(file, to);
		}
	}
	appendFileSync(values.file, added);
	for (const { story } of approving) {
		const kept = join(values.approved, story);
		rmSync(kept, { recursive: true, force: true });
		renameSync(`${kept}.next`, kept);
	}

	const first = name.split(' ')[0];
	if (approving.length === 1) {
		const [{ story, fingerprint, sizes }] = approving;
		say(`Recorded. ${first} approved ${story} at ${shortVersion(fingerprint)}.`);
		say(`Kept a picture at ${listWords(sizes)}.`);
		say(
			`Commit apps/storybook/canvas/approvals.jsonl and apps/storybook/canvas/approved/${story}/ together, on their own.`
		);
	} else {
		say(`Recorded. ${first} approved ${approving.length} page stories.`);
		say('Kept pictures of each at the sizes it is drawn at.');
		say('Commit apps/storybook/canvas/approvals.jsonl and apps/storybook/canvas/approved/ together, on their own.');
	}
}

/** @param {string[]} words `a`, `a and b`, `a, b and c` */
const listWords = (words) =>
	words.length < 2 ? words.join('') : `${words.slice(0, -1).join(', ')} and ${words.at(-1)}`;

/** @param {string} file @returns {any[]} one record per line, none when the file is absent */
const records = (file) =>
	existsSync(file) ? approvalLines(readFileSync(file, 'utf8')).map((raw) => JSON.parse(raw)) : [];

/**
 * The pictures the regenerate run took of `story`, by size. Only a picture of
 * the markup that run recorded counts: the fingerprint approved is built from
 * that digest, so no other picture can show the version approved.
 * @param {string} story
 * @returns {Map<string, string>}
 */
function picturesTaken(story) {
	const recorded = records(values.run).find((r) => r.id === story)?.digest;
	/** @type {Map<string, string>} */
	const taken = new Map();
	for (const { story: of, size, digest } of records(join(values.pictures, 'pictures.jsonl'))) {
		const file = join(values.pictures, story, `${size}.png`);
		if (of === story && recorded && digest === recorded && existsSync(file)) taken.set(size, file);
	}
	return taken;
}

if (values.approve) await approve();
