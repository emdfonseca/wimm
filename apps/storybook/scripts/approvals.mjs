// The approvals record, apps/storybook/canvas/approvals.jsonl: --check holds it
// to its shape and to only ever growing.
import { spawnSync } from 'node:child_process';
import { appendFileSync, existsSync, readFileSync } from 'node:fs';
import { basename, dirname, join, resolve } from 'node:path';
import { createInterface } from 'node:readline/promises';
import { parseArgs } from 'node:util';
import {
	approvalLines,
	approvalProblems,
	closestIds,
	formatDate,
	kindOf,
	prefixProblems,
	shortVersion
} from '../canvas/lib.js';

const app = resolve(import.meta.dirname, '..');
const { values, positionals } = parseArgs({
	options: {
		check: { type: 'boolean', default: false },
		approve: { type: 'boolean', default: false },
		file: { type: 'string', default: join(app, 'canvas/approvals.jsonl') },
		versions: { type: 'string', default: join(app, 'canvas/versions.json') },
		index: { type: 'string' },
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

if (values.check) {
	const text = existsSync(values.file) ? readFileSync(values.file, 'utf8') : '';
	const lines = approvalLines(text);
	const versions = existsSync(values.versions) ? JSON.parse(readFileSync(values.versions, 'utf8')) : {};
	const { problems, skipped } = againstGit(values.file, lines);
	problems.unshift(...approvalProblems(lines, versions));
	if (skipped) say(`append-only comparison skipped: ${skipped}`);
	if (problems.length > 0) {
		process.stderr.write(`approvals.jsonl:\n${problems.map((p) => `  ${p}`).join('\n')}\n`);
		process.exit(1);
	}
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

async function approve() {
	const [story, ...noteWords] = positionals;
	if (!personAtKeyboard()) {
		stop('Approvals are recorded by a person in their own terminal. Nothing recorded.');
	}
	const cwd = dirname(values.file);
	const name = gitConfig(cwd, 'user.name');
	const email = gitConfig(cwd, 'user.email');
	if (!name || !email) stop('git has no user.name or user.email set. Nothing recorded.');

	const before = readVersions()[story]?.implemented;
	const regenerated = spawnSync(values.regenerate, {
		shell: true,
		cwd: app,
		stdio: 'inherit'
	});
	if (regenerated.status !== 0) stop('Could not regenerate the versions. Nothing recorded.');

	const versions = readVersions();
	const entries = await pageStories();
	const pages = Object.values(entries).filter(
		(e) => e.type === 'story' && e.title.startsWith('Pages/')
	);
	const page = pages.find((e) => e.id === story);
	if (!page) {
		const near = closestIds(story ?? '', pages.map((e) => e.id));
		stop(
			`No page story "${story}".${near.length > 0 ? ` Closest: ${near.join(', ')}.` : ''} Nothing recorded.`
		);
	}
	const now = versions[story];
	if (!now?.implemented || before === undefined) {
		stop(`${story} has no version yet. Run just gen apps/storybook, then approve. Nothing recorded.`);
	}
	if (kindOf(page) === 'behaviour') {
		stop(`${story} only asserts something and needs no approval. Nothing recorded.`);
	}
	if (now.implemented !== before) {
		const first = now.versions.find((v) => v.fingerprint === now.implemented)?.firstSeen;
		const when = first === values.today ? 'today' : formatDate(first ?? values.today);
		stop(
			`${story} now renders ${shortVersion(now.implemented)}, first seen ${when}. Look again, then approve. Nothing recorded.`
		);
	}

	const text = existsSync(values.file) ? readFileSync(values.file, 'utf8') : '';
	const mine = approvalLines(text)
		.map((line) => JSON.parse(line))
		.find((r) => r.story === story && r.fingerprint === now.implemented && r.email === email);
	if (mine) stop(`You approved this version on ${formatDate(mine.at)}. Nothing recorded.`);

	const firstSeen = now.versions.find((v) => v.fingerprint === now.implemented)?.firstSeen;
	const terminal = createInterface({ input: process.stdin, output: process.stdout });
	const answer = await terminal.question(
		`Approve ${page.title} · ${page.name} at ${shortVersion(now.implemented)}, first seen ${formatDate(firstSeen)}? yes/no `
	);
	terminal.close();
	if (answer.trim().toLowerCase() !== 'yes') {
		say('Nothing recorded.');
		return;
	}

	const line = JSON.stringify({
		story,
		fingerprint: now.implemented,
		name,
		email,
		at: new Date().toISOString().replace(/\.\d{3}Z$/, 'Z'),
		note: noteWords.join(' ')
	});
	const whole = approvalLines(`${text}${line}\n`);
	const problems = [
		...approvalProblems(whole, versions),
		...prefixProblems(approvalLines(text), whole, {
			changed: 'a committed approval was changed',
			missing: 'a committed approval is missing'
		})
	];
	if (problems.length > 0) stop(`The new approval is not valid:\n  ${problems.join('\n  ')}\nNothing recorded.`);
	appendFileSync(values.file, `${line}\n`);
	say(`Recorded. ${name.split(' ')[0]} approved ${story} at ${shortVersion(now.implemented)}.`);
	say('Commit apps/storybook/canvas/approvals.jsonl on its own.');
}

if (values.approve) await approve();
