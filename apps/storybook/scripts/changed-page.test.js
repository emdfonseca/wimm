import { spawnSync } from 'node:child_process';
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { expect, it } from 'vitest';
import { kindOf, parseApprovals, sha256Hex, statusOf } from '../canvas/lib.js';

const scripts = import.meta.dirname;
const GLOBAL = 'd'.repeat(64);
const STORY = 'pages-overview--populated';

it('a page that changed after it was approved passes the check and reads as changed', async () => {
	const dir = mkdtempSync(join(tmpdir(), 'changed-'));
	mkdirSync(join(dir, 'canvas'));
	const at = (name) => join(dir, 'canvas', name);
	copyFileSync(join(scripts, 'fixtures/index.json'), at('index.json'));
	const runOf = (digest) =>
		[
			{ id: STORY, digest },
			{ id: 'pages-overview--empty', digest: 'b'.repeat(64) },
			{ id: 'pages-overview--focus', digest: 'c'.repeat(64) },
			{ run: 'passed' }
		]
			.map((r) => `${JSON.stringify(r)}\n`)
			.join('');
	const generate = (digest) => {
		writeFileSync(at('run.jsonl'), runOf(digest));
		return spawnSync(
			'node',
			[
				join(scripts, 'versions.mjs'),
				'--run', at('run.jsonl'),
				'--index', at('index.json'),
				'--versions', at('versions.json'),
				'--approvals', at('approvals.jsonl'),
				'--global-digest', GLOBAL,
				'--today', '2026-09-20'
			],
			{ encoding: 'utf8' }
		);
	};

	writeFileSync(at('approvals.jsonl'), '');
	expect(generate('a'.repeat(64)).status).toBe(0);
	const approved = JSON.parse(readFileSync(at('versions.json'), 'utf8'))[STORY].implemented;
	writeFileSync(
		at('approvals.jsonl'),
		`${JSON.stringify({ story: STORY, fingerprint: approved, name: 'Emanuel Fonseca', email: 'e@x.io', at: '2026-09-20T18:04:11Z', note: '' })}\n`
	);

	const moved = generate('e'.repeat(64));
	expect(moved.status).toBe(0);
	const check = spawnSync(
		'node',
		[join(scripts, 'approvals.mjs'), '--check', '--file', at('approvals.jsonl'), '--versions', at('versions.json')],
		{ encoding: 'utf8', env: { ...process.env, GIT_CEILING_DIRECTORIES: tmpdir() } }
	);
	expect(check.status).toBe(0);

	const versions = JSON.parse(readFileSync(at('versions.json'), 'utf8'));
	const got = statusOf(
		versions[STORY],
		parseApprovals(readFileSync(at('approvals.jsonl'), 'utf8')),
		kindOf({ tags: ['kind-state'] })
	);
	expect(got.status).toBe('changed');
	expect(got.approved?.fingerprint).toBe(approved);
	expect(got.implemented?.fingerprint).toBe(await sha256Hex(`${'e'.repeat(64)}\n${GLOBAL}`));
	expect(got.implemented?.fingerprint).not.toBe(got.approved?.fingerprint);
});
