import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * No load function asks a bank.
 *
 * A SvelteKit load blocks the navigation until it returns. Reading balances is
 * a round trip per account and a first fill is unbounded pages per account, so
 * a bank call in a load means clicking a destination and watching the previous
 * page until every bank has answered — with no feedback, because as far as the
 * browser is concerned nothing has happened yet.
 *
 * This shipped twice: once on Transactions, and once on Overview after the
 * first was fixed in place rather than as a class. Hence a check rather than a
 * comment.
 *
 * What is allowed instead: the load returns what is stored, every figure
 * carrying the time it was read, and the page asks the banks behind the arrival
 * and re-reads itself when they answer.
 */

const ROUTES = new URL('.', import.meta.url).pathname;

/** Files SvelteKit runs as part of resolving a navigation. */
const BLOCKING = /(\+page\.server\.ts|\+layout\.server\.ts|\+page\.ts|\+layout\.ts)$/;

/** Calls that reach a bank unless told not to, and the flag that tells them. */
const GATED: Record<string, string> = {
	'banking.listAccounts': 'skipRead: true',
	'banking.listTransactions': 'skipSync: true'
};

/** Calls that exist only to reach a bank. No flag makes them safe here. */
const ALWAYS_READS = ['banking.refreshBalances', 'banking.refreshTransactions'];

function walk(dir: string): string[] {
	return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
		const path = join(dir, entry.name);
		if (entry.isDirectory()) return walk(path);
		return BLOCKING.test(entry.name) ? [path] : [];
	});
}

/**
 * The call and its arguments, up to the balanced closing bracket. A regex over
 * the whole file would match the flag from a different call three lines down,
 * which is the way a check like this passes while the defect is present.
 */
function callArguments(source: string, callee: string): string[] {
	const found: string[] = [];
	let from = 0;

	for (;;) {
		const at = source.indexOf(callee + '(', from);
		if (at === -1) return found;

		let depth = 0;
		let end = at + callee.length;
		for (; end < source.length; end++) {
			const char = source[end];
			if (char === '(') depth++;
			else if (char === ')' && --depth === 0) break;
		}
		found.push(source.slice(at, end + 1));
		from = end + 1;
	}
}

export function offences(source: string): string[] {
	const problems: string[] = [];

	for (const [callee, flag] of Object.entries(GATED)) {
		for (const call of callArguments(source, callee)) {
			if (!call.includes(flag)) {
				problems.push(`${callee} without ${flag}`);
			}
		}
	}
	for (const callee of ALWAYS_READS) {
		if (source.includes(callee + '(')) problems.push(`${callee} in a load`);
	}
	return problems;
}

describe('no load function asks a bank', () => {
	const files = walk(ROUTES);

	it('finds the load functions to check', () => {
		// A check that silently scanned nothing would pass forever.
		expect(files.length).toBeGreaterThan(0);
	});

	for (const file of files) {
		it(file.slice(ROUTES.length), () => {
			expect(offences(readFileSync(file, 'utf8'))).toEqual([]);
		});
	}
});

/**
 * The check's own fixtures. Asserting that today's routes pass proves nothing
 * about the check, because today's routes are correct — these assert the
 * opposite, that the shapes it must refuse are refused.
 */
describe('the check refuses what it must', () => {
	for (const [name, source, want] of [
		[
			'a blocking balance read',
			`const view = await call(cookies, (o) => banking.listAccounts({ skipRead: false }, o));`,
			'banking.listAccounts without skipRead: true'
		],
		[
			'a balance read that says nothing',
			`await call(cookies, (o) => banking.listAccounts({}, o));`,
			'banking.listAccounts without skipRead: true'
		],
		[
			'a blocking ledger sync',
			`await call(cookies, (o) => banking.listTransactions({ skipSync: false }, o));`,
			'banking.listTransactions without skipSync: true'
		],
		[
			'a refresh in a load',
			`await call(cookies, (o) => banking.refreshBalances({}, o));`,
			'banking.refreshBalances in a load'
		]
	] as const) {
		it(`refuses ${name}`, () => {
			expect(offences(source)).toContain(want);
		});
	}

	// The trap a looser check falls into: a safe call followed by an unsafe one
	// whose arguments happen to sit near the safe one's flag.
	it('refuses an unsafe call standing beside a safe one', () => {
		const source = `
			await banking.listAccounts({ skipRead: true }, o);
			await banking.listTransactions({ accountId, cursor }, o);
		`;
		expect(offences(source)).toEqual(['banking.listTransactions without skipSync: true']);
	});

	it('accepts the shape the routes actually use', () => {
		const source = `
			const [ledger, accounts] = await Promise.all([
				call(cookies, (o) => banking.listTransactions({ accountId, cursor, skipSync: true }, o)),
				call(cookies, (o) => banking.listAccounts({ skipRead: true }, o))
			]);
		`;
		expect(offences(source)).toEqual([]);
	});
});
