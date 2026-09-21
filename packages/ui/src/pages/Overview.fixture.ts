import { netMonths, netSummary } from '../molecules/MonthlyNetChart.fixture';
import type { HistoryMonth, HistorySection, RecurringEntry } from './Overview.svelte';

const euro = (n: number) =>
	`€${Math.abs(n).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

const signed = (n: number) => `${n < 0 ? '−' : '+'}${euro(n)}`;

/** A month's three figures where nothing else about it is under test. */
const figures = (net: number) => [
	{ label: 'Money in', value: euro(2450) },
	{ label: 'Money out', value: euro(2450 - net) },
	{ label: 'Net', value: signed(net) }
];

const galp = {
	description: 'Galp',
	account: 'Current account',
	amount: '−€210.00',
	date: '27 Mar',
	negative: true,
	unusual: true,
	note: 'usually about €60',
	href: '/transactions?page=2026-03-27.b2'
};

export const marchPayments = [
	{
		description: 'Auto Reparadora',
		account: 'Current account',
		amount: '−€1,650.00',
		date: '12 Mar',
		negative: true,
		unusual: true,
		note: 'first payment to Auto Reparadora',
		href: '/transactions?page=2026-03-12.a1'
	},
	galp
];

export const bonusPayment = {
	description: 'Employer Lda',
	account: 'Current account',
	amount: '+€9,804.00',
	date: '22 May',
	negative: false,
	unusual: true,
	note: 'usually about €2,450',
	href: '/transactions?page=2026-05-22.c3'
};

const details: Record<string, Partial<HistoryMonth>> = {
	'September 2026': {
		state: 'So far',
		span: '1 to 20 Sep',
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€1,812.64' },
			{ label: 'Net', value: '+€637.36' }
		],
		note: 'A month under way is not set against whole months.'
	},
	'August 2026': {
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€2,144.60' },
			{ label: 'Net', value: '+€305.40' }
		],
		risers: [
			{ label: 'Galp', value: '€246.80 · €82.40 more than usual', proportion: 1 },
			{ label: 'Amazon', value: '€188.20 · €64.10 more than usual', proportion: 0.78 },
			{ label: 'Zara', value: '€120.00 · not usually paid', proportion: 0.5 }
		]
	},
	'June 2026': {
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€2,299.80' },
			{ label: 'Net', value: '+€150.20' }
		],
		nothingRose: 'Nothing took more than usual this month.'
	},
	'March 2026': {
		unusualLine: '2 unusual · +€379.70 without them',
		netUsual: '+€379.70',
		valueUsual: 379.7,
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€3,930.30' },
			{ label: 'Net', value: '−€1,480.30' },
			{ label: 'Without unusual payments', value: '+€379.70' }
		],
		risers: [
			{ label: 'Auto Reparadora', value: '€1,650.00 · not usually paid', proportion: 1 },
			{ label: 'Galp', value: '€318.40 · €154.00 more than usual', proportion: 0.09 }
		],
		payments: marchPayments
	},
	'December 2025': {
		unusualLine: '1 unusual · −€160.00 without it',
		netUsual: '−€160.00',
		valueUsual: -160,
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€3,090.00' },
			{ label: 'Net', value: '−€640.00' },
			{ label: 'Without unusual payments', value: '−€160.00' }
		],
		payments: [
			{
				description: 'Worten',
				account: 'Current account',
				amount: '−€480.00',
				date: '18 Dec',
				negative: true,
				unusual: true,
				note: 'usually about €70',
				href: '/transactions?page=2025-12-18.d4'
			}
		]
	}
};

/** September 2025 to September 2026, newest first, as the load writes it. */
export const historyMonths: HistoryMonth[] = [...netMonths].reverse().map((m) => ({
	key: m.label,
	label: m.label,
	short: m.short,
	net: m.net,
	value: m.value,
	partial: m.partial,
	figures: figures(m.value),
	...details[m.label]
}));

export const allView = {
	typical: '+€189.40',
	typicalValue: 189.4,
	average: '−€37.80',
	sentence: 'In a typical month €189.40 more comes in than goes out.',
	summary: netSummary,
	high: '+€637',
	low: '−€1,480'
};

export const usualView = {
	typical: '+€205.55',
	typicalValue: 205.55,
	average: '+€157.20',
	sentence: 'In a typical month €205.55 more comes in than goes out.',
	summary:
		'Net by month from September 2025 to September 2026. 9 of 12 full months ended with more in than out. Highest +€310.00 in February 2026. Lowest −€160.00 in December 2025. A typical month is +€205.55.',
	high: '+€637',
	low: '−€160'
};

export const basis =
	'From 12 full months. The typical month is the middle one, so one exceptional month barely moves it.';

export const history: HistorySection = {
	span: 'September 2025 to September 2026',
	months: historyMonths,
	basis,
	all: allView,
	usual: usualView,
	setAside: '3 unusual payments are set aside. They are still listed in their months.',
	open: 'August 2026'
};

const entry = (
	id: string,
	name: string,
	cadence: string,
	date: string,
	amount: string
): RecurringEntry => ({ id, name, cadence, date, amount });

export const spotify = entry(
	'r1',
	'Spotify',
	'Monthly · Current account · Monzo',
	'Was expected 18 Sep',
	'−€9.99'
);
export const cleaning = entry(
	'r2',
	'Limpeza Casa',
	'Weekly · Joint account · Monzo',
	'Expected 24 Sep',
	'−€45.00'
);
export const rent = entry(
	'r3',
	'Rent',
	'Monthly · Joint account · Monzo',
	'Expected 6 Oct',
	'−€820.00'
);
export const netflix = entry(
	'r4',
	'Netflix',
	'Monthly · Current account · Monzo',
	'Expected 14 Oct',
	'−€12.99'
);
export const nos = entry(
	'r5',
	'NOS',
	'Monthly · Conta à Ordem · Montepio',
	'Expected 17 Oct',
	'−€39.99'
);
export const fidelidade = entry(
	'r6',
	'Fidelidade',
	'Yearly · Joint account · Monzo',
	'Expected 3 Mar 2027',
	'−€386.00'
);

export const recurring = [spotify, cleaning, rent, netflix, nos, fidelidade];
