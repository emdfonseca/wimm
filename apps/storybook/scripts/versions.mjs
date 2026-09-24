// Writes apps/storybook/canvas/versions.json from what the headless run left.
// It validates first and refuses to write anything from a run it cannot trust;
// it never compares and never fails because the file was out of date.
import { createHash } from 'node:crypto';
import { existsSync, readFileSync, readdirSync, writeFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { parseArgs } from 'node:util';
import { RUN_FILE, nextVersions, runProblems, serialiseVersions, sha256Hex } from '../canvas/lib.js';

const app = resolve(import.meta.dirname, '..');
const { values } = parseArgs({
	options: {
		run: { type: 'string', default: join(app, RUN_FILE) },
		index: { type: 'string' },
		versions: { type: 'string', default: join(app, 'canvas/versions.json') },
		approvals: { type: 'string', default: join(app, 'canvas/approvals.jsonl') },
		'global-digest': { type: 'string' },
		today: { type: 'string', default: new Date().toISOString().slice(0, 10) }
	}
});

/** @param {string} line */
const say = (line) => process.stdout.write(`${line}\n`);

/** @param {string[]} problems */
function refuse(problems) {
	process.stderr.write(`versions.json not written:\n${problems.map((p) => `  ${p}`).join('\n')}\n`);
	process.exit(1);
}

/** One digest over the styles every page shares: tokens, base, fonts and the font files. */
function globalDigest() {
	const ui = resolve(app, '../../packages/ui/src');
	const files = [
		'tokens.css',
		'base.css',
		'fonts.css',
		...readdirSync(join(ui, 'fonts')).sort().map((f) => `fonts/${f}`)
	];
	const hash = createHash('sha256');
	for (const file of files) hash.update(file).update('\0').update(readFileSync(join(ui, file)));
	return hash.digest('hex');
}

async function pageIds() {
	const index = values.index
		? JSON.parse(readFileSync(values.index, 'utf8'))
		: await (await import('storybook/internal/core-server')).buildIndex({
				configDir: join(app, '.storybook')
			});
	return Object.values(index.entries)
		.filter((e) => e.type === 'story' && e.title.startsWith('Pages/'))
		.map((e) => e.id)
		.sort();
}

if (!existsSync(values.run)) refuse([`no run file at ${values.run}; run the tests first`]);
const ids = await pageIds();
const { problems, digests } = runProblems(readFileSync(values.run, 'utf8'), ids);

/** @type {{ story: string, fingerprint: string }[]} */
const approvals = [];
if (existsSync(values.approvals)) {
	readFileSync(values.approvals, 'utf8')
		.split('\n')
		.forEach((raw, i) => {
			if (raw.trim() === '') return;
			try {
				const { story, fingerprint } = JSON.parse(raw);
				if (typeof story !== 'string' || typeof fingerprint !== 'string') throw new Error();
				approvals.push({ story, fingerprint });
			} catch {
				problems.push(`approvals.jsonl line ${i + 1} cannot be read; run the approvals check`);
			}
		});
}
if (problems.length > 0) refuse(problems);

const shared = values['global-digest'] ?? globalDigest();
/** @type {Map<string, string>} */
const fingerprints = new Map();
for (const [id, digest] of digests) fingerprints.set(id, await sha256Hex(`${digest}\n${shared}`));

const before = existsSync(values.versions) ? readFileSync(values.versions, 'utf8') : '';
const { versions, moved, added, removed } = nextVersions({
	previous: before ? JSON.parse(before) : {},
	fingerprints,
	approvals,
	today: values.today
});
const after = serialiseVersions(versions);

if (after === before) {
	say('versions.json unchanged');
} else {
	writeFileSync(values.versions, after);
	if (moved.length > 0) say(`new version:\n${moved.map((id) => `  ${id}`).join('\n')}`);
	if (added.length > 0) say(`${added.length} story(ies) added`);
	if (removed.length > 0) say(`${removed.length} story(ies) removed`);
	say('versions.json changed, commit it');
}
