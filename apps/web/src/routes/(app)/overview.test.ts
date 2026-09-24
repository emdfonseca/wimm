import { beforeEach, describe, expect, it, vi } from 'vitest';
import { InsightScope, Cadence } from '@wimm/contracts/banking';
import type { HistorySection, RecurringEntry } from '@wimm/ui';

const listAccounts = vi.fn();
const listTransactions = vi.fn();
const getBalanceTrend = vi.fn();
const getMonthSummary = vi.fn();
const getMonthHistory = vi.fn();

vi.mock('$lib/server/banking', () => ({
	banking: {
		listAccounts: (...args: unknown[]) => listAccounts(...args),
		listTransactions: (...args: unknown[]) => listTransactions(...args),
		getBalanceTrend: (...args: unknown[]) => getBalanceTrend(...args),
		getMonthSummary: (...args: unknown[]) => getMonthSummary(...args),
		getMonthHistory: (...args: unknown[]) => getMonthHistory(...args)
	}
}));

vi.mock('$lib/server/call', () => ({
	call: <T>(_cookies: unknown, invoke: (o: unknown) => Promise<T>) =>
		invoke({ headers: {}, onHeader: () => {} })
}));

const cookies = { get: () => 'session', delete: () => {}, set: () => {} };
const money = (minor: bigint, currency: string) => ({ minor, currency });
const total = (minor: bigint, currency: string) => ({ total: money(minor, currency), accountCount: 1 });

const emptyHistory = {
	scope: InsightScope.ALL,
	available: [],
	householdCounted: [],
	householdNotCounted: [],
	histories: []
};

beforeEach(() => {
	for (const mock of [listAccounts, listTransactions, getBalanceTrend, getMonthSummary]) mock.mockReset();
	getMonthHistory.mockReset();
	getMonthHistory.mockResolvedValue(emptyHistory);
});

interface View {
	currencies: {
		currency: string;
		household?: string;
		own?: string;
		month?: unknown;
		chart?: unknown;
		history?: HistorySection;
		recurring?: RecurringEntry[];
	}[];
	recentTransactions: { description: string; amount: string }[];
	scope?: { choices: string[]; value: string; note?: string };
}

async function overview(search = '') {
	const { load } = await import('./+page.server');
	return (await load!({
		cookies,
		depends: () => {},
		url: new URL(`http://localhost/${search}`)
	} as never)) as View;
}

function sectionOf(view: View): HistorySection {
	const history = view.currencies[0]?.history;
	if (!history) throw new Error('no history section');
	return history;
}

const at = (iso: string) => ({ seconds: BigInt(Math.floor(Date.parse(iso) / 1000)), nanos: 0 });

describe('Overview load', () => {
	it('gives a currency with zero figures, no trend and no summary no tile', async () => {
		listAccounts.mockResolvedValue({
			accounts: [],
			householdTotals: [total(669_800n, 'EUR')],
			ownTotals: [total(0n, 'USD')],
			failures: []
		});
		listTransactions.mockResolvedValue({ ledger: { transactions: [] } });
		getBalanceTrend.mockResolvedValue({ trends: [], partialCoverage: false });
		getMonthSummary.mockResolvedValue({ months: [] });

		const view = await overview();

		expect(view.currencies.map((c) => c.currency)).toEqual(['EUR']);
	});

	it('keeps a zero currency that has a month summary, and reads the merchant name', async () => {
		listAccounts.mockResolvedValue({
			accounts: [],
			householdTotals: [total(669_800n, 'EUR')],
			ownTotals: [total(0n, 'GBP')],
			failures: []
		});
		listTransactions.mockResolvedValue({
			ledger: {
				transactions: [
					{
						id: 't1',
						displayName: 'Pingo Doce',
						counterpartyName: 'PINGO DOCE 123456789',
						accountName: 'Current account',
						amount: money(-4_218n, 'EUR'),
						bookingDate: { seconds: 1_789_000_000n }
					}
				]
			}
		});
		getBalanceTrend.mockResolvedValue({ trends: [], partialCoverage: false });
		getMonthSummary.mockResolvedValue({
			months: [
				{
					currency: 'GBP',
					in: money(10n, 'GBP'),
					out: money(0n, 'GBP'),
					net: money(10n, 'GBP'),
					monthStart: { seconds: 1_788_220_800n },
					topMerchants: [],
					largestPayments: []
				}
			]
		});

		const view = await overview();

		expect(view.currencies.map((c) => c.currency).sort()).toEqual(['EUR', 'GBP']);
		expect(view.recentTransactions[0]?.description).toBe('Pingo Doce');
	});
});


const quietAccounts = () => {
	listAccounts.mockResolvedValue({
		accounts: [
			{ id: 'a1', name: 'Current account', connection: { bankName: 'Monzo', live: true } }
		],
		householdTotals: [total(669_800n, 'EUR')],
		ownTotals: [],
		failures: []
	});
	listTransactions.mockResolvedValue({ ledger: { transactions: [] } });
	getBalanceTrend.mockResolvedValue({ trends: [], partialCoverage: false });
	getMonthSummary.mockResolvedValue({ months: [] });
};

describe('Overview scope', () => {
	it.each([
		['', InsightScope.ALL],
		['?scope=household', InsightScope.HOUSEHOLD],
		['?scope=yours', InsightScope.OWN]
	])('sends %s to all three calls', async (search, want) => {
		quietAccounts();
		getMonthHistory.mockResolvedValue({
			...emptyHistory,
			scope: want,
			available: [InsightScope.HOUSEHOLD, InsightScope.OWN, InsightScope.ALL]
		});

		await overview(search);

		expect(getBalanceTrend.mock.calls[0]![0]).toEqual({ scope: want });
		expect(getMonthSummary.mock.calls[0]![0]).toEqual({ scope: want });
		expect(getMonthHistory.mock.calls[0]![0]).toEqual({ scope: want });
	});

	it('redirects to the bare address when the scope asked for was not the one answered', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue({ ...emptyHistory, scope: InsightScope.ALL });

		await expect(overview('?scope=household')).rejects.toMatchObject({
			status: 303,
			location: '/'
		});
	});

	it('redirects an unknown scope too', async () => {
		quietAccounts();
		await expect(overview('?scope=everyone')).rejects.toMatchObject({ status: 303 });
	});

	it('offers no control when nothing is available', async () => {
		quietAccounts();
		expect((await overview()).scope).toBeUndefined();
	});

	it('offers the choices on offer, and names what Household does not count', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue({
			...emptyHistory,
			scope: InsightScope.HOUSEHOLD,
			available: [InsightScope.HOUSEHOLD, InsightScope.OWN, InsightScope.ALL],
			householdCounted: ['Joint account'],
			householdNotCounted: ['Joint savings']
		});

		const { scope } = await overview('?scope=household');

		expect(scope?.choices).toEqual(['household', 'own', 'all']);
		expect(scope?.value).toBe('household');
		expect(scope?.note).toBe(
			'Household counts Joint account. Joint savings is household money too, but it is not yours, so its transactions are not counted.'
		);
	});
});

const eur = (minor: bigint) => ({ minor, currency: 'EUR' });
const month = (iso: string, net: bigint, extra: Record<string, unknown> = {}) => ({
	monthStart: at(iso),
	in: eur(245_000n),
	out: eur(245_000n - net),
	net: eur(net),
	soFar: false,
	risers: [],
	unusual: [],
	...extra
});

function historyOf(months: unknown[], extra: Record<string, unknown> = {}) {
	return {
		...emptyHistory,
		histories: [
			{
				currency: 'EUR',
				months,
				fullMonths: 12,
				typicalMonths: 6,
				typicalNet: eur(18_940n),
				averageNet: eur(-3_780n),
				typicalNetUsual: eur(20_555n),
				averageNetUsual: eur(15_720n),
				unusualCount: 0,
				recurring: [],
				lateLedgers: [],
				...extra
			}
		]
	};
}

describe('Month by month', () => {
	it('words the months, the typical month and what set aside means', async () => {
		quietAccounts();
		const bonus = {
			transaction: {
				id: 'b1',
				displayName: 'Employer Lda',
				accountName: 'Current account',
				amount: eur(980_400n),
				bookingDate: at('2026-05-22T00:00:00Z')
			},
			typical: eur(245_000n),
			firstPayment: false
		};
		const garage = {
			transaction: {
				id: 'a1',
				displayName: 'Auto Reparadora',
				accountName: 'Current account',
				amount: eur(-165_000n),
				bookingDate: at('2026-03-12T00:00:00Z')
			},
			firstPayment: true
		};
		getMonthHistory.mockResolvedValue(
			historyOf(
				[
					month('2026-09-01T00:00:00Z', 63_736n, { soFar: true }),
					month('2026-08-01T00:00:00Z', 30_540n, {
						risers: [
							{ name: 'Galp', total: eur(24_680n), usual: eur(16_440n), payments: 2 },
							{ name: 'Zara', total: eur(12_000n), usual: eur(0n), payments: 1 }
						]
					}),
					month('2026-06-01T00:00:00Z', 15_020n),
					month('2026-05-01T00:00:00Z', 1_002_900n, { netUsual: eur(22_500n), unusual: [bonus] }),
					month('2026-04-01T00:00:00Z', 4_110n, { heldFrom: at('2026-04-12T00:00:00Z') }),
					month('2026-03-01T00:00:00Z', -148_030n, {
						netUsual: eur(37_970n),
						unusual: [garage, garage]
					})
				],
				{ unusualCount: 3 }
			)
		);

		const history = sectionOf(await overview());
		const byKey = (key: string) => history.months.find((m) => m.key === key);

		expect(history.span).toBe('March 2026 to September 2026');
		expect(byKey('2026-09')).toMatchObject({ state: 'So far', net: '+€637.36', partial: true });
		expect(byKey('2026-04')?.state).toBe('Held from 12 Apr');
		expect(byKey('2026-03')?.unusualLine).toBe('2 unusual · +€379.70 without them');
		expect(byKey('2026-05')?.unusualLine).toBe('1 unusual · +€225.00 without it');
		expect(byKey('2026-08')?.risers?.map((r) => r.value)).toEqual([
			'€246.80 · €82.40 more than usual',
			'€120.00 · not usually paid'
		]);
		expect(byKey('2026-06')?.nothingRose).toBe('Nothing took more than usual this month.');
		expect(byKey('2026-09')?.nothingRose).toBeUndefined();
		expect(history.open).toBe('2026-08');
		expect(history.all.typical).toBe('+€189.40');
		expect(history.all.average).toBe('−€37.80');
		expect(history.all.sentence).toBe('In a typical month €189.40 more comes in than goes out.');
		expect(history.usual?.typical).toBe('+€205.55');
		expect(history.usual?.average).toBe('+€157.20');
		expect(history.setAside).toBe(
			'3 unusual payments are set aside. They are still listed in their months.'
		);
		expect(history.basis).toBe(
			'From the last 6 full months. The typical month is the middle one, so one exceptional month barely moves it.'
		);
	});

	it.each([
		[4, 4, 'From 4 full months.'],
		[6, 6, 'From 6 full months.']
	])(
		'says every full month is drawn from when %i are held and %i used',
		async (fullMonths, typicalMonths, first) => {
			quietAccounts();
			getMonthHistory.mockResolvedValue(
				historyOf([month('2026-08-01T00:00:00Z', 100n)], { fullMonths, typicalMonths })
			);

			expect(sectionOf(await overview()).basis).toBe(
				`${first} The typical month is the middle one, so one exceptional month barely moves it.`
			);
		}
	);

	it('reads a typical month below zero in words, and links each unusual payment', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue(
			historyOf(
				[
					month('2026-08-01T00:00:00Z', 100n, {
						netUsual: eur(50n),
						unusual: [
							{
								transaction: {
									id: 'g1',
									displayName: 'Galp',
									accountName: 'Current account',
									amount: eur(-21_000n),
									bookingDate: at('2026-03-27T00:00:00Z')
								},
								typical: eur(6_000n),
								firstPayment: false
							}
						]
					})
				],
				{ typicalNet: eur(-21_240n), unusualCount: 1 }
			)
		);

		const history = sectionOf(await overview());

		expect(history.all.sentence).toBe(
			'More goes out than comes in. In a typical month €212.40 more goes out than comes in.'
		);
		expect(history.setAside).toBe(
			'1 unusual payment is set aside. It is still listed in its month.'
		);
		expect(history.months[0]!.payments).toEqual([
			expect.objectContaining({
				description: 'Galp',
				amount: '−€210.00',
				note: 'usually about €60',
				unusual: true,
				href: '/transactions?page=2026-03-27.g1'
			})
		]);
	});

	it('says a first payment is one, never what is usual', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue(
			historyOf([
				month('2026-08-01T00:00:00Z', 100n, {
					netUsual: eur(50n),
					unusual: [
						{
							transaction: {
								id: 'a1',
								displayName: 'Auto Reparadora',
								accountName: 'Current account',
								amount: eur(-165_000n),
								bookingDate: at('2026-03-12T00:00:00Z')
							},
							firstPayment: true
						}
					]
				})
			])
		);

		const [payment] = sectionOf(await overview()).months[0]!.payments ?? [];

		expect(payment?.note).toBe('first payment to Auto Reparadora');
	});

	it('states no typical month under three full months, and offers no view control', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue(
			historyOf([month('2026-08-01T00:00:00Z', 100n)], {
				fullMonths: 2,
				typicalNet: undefined,
				averageNet: undefined
			})
		);

		const history = sectionOf(await overview());

		expect(history.waiting).toBe(
			'A typical month and an average month appear once three full months are held.'
		);
		expect(history.all.typical).toBeUndefined();
		expect(history.usual).toBeUndefined();
		expect(history.basis).toBeUndefined();
	});

	it('has no section at all before a full month is held', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue(
			historyOf([month('2026-09-01T00:00:00Z', 100n, { soFar: true })], { fullMonths: 0 })
		);

		expect((await overview()).currencies[0]?.history).toBeUndefined();
	});

	it('says which account the earlier months leave out', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue(
			historyOf([month('2026-08-01T00:00:00Z', 30_540n)], {
				lateLedgers: [{ accountName: 'CLASSIC CEMG', from: at('2026-09-19T00:00:00Z') }]
			})
		);
		expect(sectionOf(await overview()).late).toEqual([
			'Months before 19 Sep do not include CLASSIC CEMG.'
		]);
	});

	it('says nothing where no account is missing', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue(historyOf([month('2026-08-01T00:00:00Z', 30_540n)]));
		expect(sectionOf(await overview()).late).toBeUndefined();
	});
});

describe('Recurring payments', () => {
	it('says what is expected, what is late, and what is likely yearly', async () => {
		quietAccounts();
		const payment = (name: string, cadence: Cadence, iso: string, extra = {}) => ({
			name,
			amount: eur(-999n),
			cadence,
			likely: false,
			expected: at(iso),
			late: false,
			accountId: 'a1',
			...extra
		});
		getMonthHistory.mockResolvedValue(
			historyOf([month('2026-08-01T00:00:00Z', 100n)], {
				recurring: [
					payment('Spotify', Cadence.MONTHLY, '2026-09-18T00:00:00Z', { late: true }),
					payment('Rent', Cadence.MONTHLY, '2026-10-06T00:00:00Z'),
					payment('Fidelidade', Cadence.YEARLY, '2027-03-03T00:00:00Z', { likely: true })
				]
			})
		);

		const recurring = (await overview()).currencies[0]!.recurring!;

		expect(recurring.map((r) => [r.name, r.cadence, r.date, r.amount])).toEqual([
			['Spotify', 'Monthly · Current account · Monzo', 'Was expected 18 Sep', '−€9.99'],
			['Rent', 'Monthly · Current account · Monzo', 'Expected 6 Oct', '−€9.99'],
			['Fidelidade', 'Likely yearly · Current account · Monzo', 'Expected 3 Mar 2027', '−€9.99']
		]);
	});

	it('shows no section when nothing recurs', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue(historyOf([month('2026-08-01T00:00:00Z', 100n)]));

		expect((await overview()).currencies[0]?.recurring).toBeUndefined();
	});
});

describe('The balance chart', () => {
	it('says what moved a day, in words', async () => {
		quietAccounts();
		const point = (iso: string, minor: bigint, extra = {}) => ({
			date: at(iso),
			balance: eur(minor),
			movers: [],
			smaller: 0,
			...extra
		});
		getBalanceTrend.mockResolvedValue({
			partialCoverage: false,
			trends: [
				{
					currency: 'EUR',
					shortHistory: false,
					points: [
						point('2026-08-02T23:59:59Z', 1_067_591n),
						point('2026-08-03T23:59:59Z', 987_012n, {
							movers: [
								{ displayName: 'Leroy Merlin', amount: eur(-64_000n), unusual: true },
								{ displayName: 'Galp', amount: eur(-9_210n), unusual: false }
							],
							smaller: 3
						}),
						point('2026-08-04T23:59:59Z', 987_012n),
						point('2026-08-05T23:59:59Z', 1_231_712n, {
							movers: [{ displayName: 'Salary', amount: eur(245_000n), unusual: false }],
							smaller: 1
						})
					]
				}
			]
		});

		const points = (await overview()).currencies[0]!.chart as {
			points: { date: string; change?: string; movers: unknown[]; smaller?: string; empty: boolean }[];
		};

		expect(points.points[0]!.change).toBeUndefined();
		expect(points.points[1]).toMatchObject({
			change: '€805.79 less than 2 Aug',
			smaller: 'and 3 smaller',
			empty: false
		});
		expect(points.points[1]!.movers).toEqual([
			{ name: 'Leroy Merlin', amount: '−€640.00', negative: true, unusual: true },
			{ name: 'Galp', amount: '−€92.10', negative: true, unusual: false }
		]);
		expect(points.points[2]).toMatchObject({ change: 'No change from 3 Aug', empty: true });
		expect(points.points[3]).toMatchObject({ smaller: 'and 1 smaller' });
	});
});

describe('Sections that stay in place across scopes', () => {
	const offered = [InsightScope.HOUSEHOLD, InsightScope.OWN, InsightScope.ALL];

	it('says why Month by month and Recurring payments are empty while a control is on offer', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue({
			...historyOf([month('2026-09-01T00:00:00Z', 100n, { soFar: true })], {
				fullMonths: 0,
				typicalNet: undefined,
				averageNet: undefined
			}),
			scope: InsightScope.HOUSEHOLD,
			available: offered
		});

		const section = (await overview('?scope=household')).currencies[0]! as Record<string, unknown>;

		expect(section.history).toBeUndefined();
		expect(section.historyEmpty).toBe('No full month is held for these accounts yet.');
		expect(section.recurringEmpty).toBe('No recurring payments in these accounts.');
	});

	it('says the same where the scope returned no history for the currency at all', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue({ ...emptyHistory, scope: InsightScope.OWN, available: offered });

		const section = (await overview('?scope=yours')).currencies[0]! as Record<string, unknown>;

		expect(section.historyEmpty).toBe('No full month is held for these accounts yet.');
		expect(section.recurringEmpty).toBe('No recurring payments in these accounts.');
	});

	it('says nothing where no control is on offer', async () => {
		quietAccounts();

		const section = (await overview()).currencies[0]! as Record<string, unknown>;

		expect(section.historyEmpty).toBeUndefined();
		expect(section.recurringEmpty).toBeUndefined();
	});

	it('keeps the view control, and says there is nothing to set aside', async () => {
		quietAccounts();
		getMonthHistory.mockResolvedValue({
			...historyOf([month('2026-08-01T00:00:00Z', 100n)]),
			scope: InsightScope.HOUSEHOLD,
			available: offered
		});

		const history = sectionOf(await overview('?scope=household'));

		expect(history.usual).toBeDefined();
		expect(history.setAside).toBe('No unusual payments to set aside.');
	});
});
