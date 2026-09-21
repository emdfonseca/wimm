import type { BalanceMover, BalancePoint } from './BalanceChart.svelte';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

function euro(value: number): string {
	return `€${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

/** 22 Jun to 20 Sep 2026, one point a day, straight between the five days the
 *  stories name: it starts at €9,640.00, bottoms out on 14 Jul, is €9,870.12
 *  on 3 Aug, peaks on 2 Sep and ends at €11,693.55. */
const anchors: [string, number][] = [
	['2026-06-22', 9640],
	['2026-07-14', 9204],
	['2026-08-03', 9870.12],
	['2026-09-02', 12118],
	['2026-09-20', 11693.55]
];

const DAY = 86_400_000;

const mover = (name: string, amount: string, extra: Partial<BalanceMover> = {}): BalanceMover => ({
	name,
	amount,
	negative: amount.startsWith('−'),
	...extra
});

/** What the stories point at, written out as the load would write it. */
const days: Record<string, Partial<BalancePoint>> = {
	'22 Jun': { change: undefined, empty: true },
	'3 Aug': {
		change: '€805.79 less than 2 Aug',
		movers: [mover('Leroy Merlin', '−€640.00', { unusual: true }), mover('Galp', '−€92.10')],
		smaller: 'and 3 smaller',
		empty: false
	},
	'9 Aug': { amount: '€9,064.33', change: 'No change from 8 Aug', empty: true },
	'10 Sep': {
		change: '€214.60 less than 9 Sep',
		movers: [
			mover('Pingo Doce', '−€92.40'),
			mover('Farmácia Central', '−€41.20'),
			mover('Galp', '−€38.00'),
			mover('Zara', '−€27.50'),
			mover('Uber', '−€15.50')
		],
		smaller: 'and 7 smaller',
		empty: false
	},
	'15 Sep': {
		amount: '€11,748.72',
		change: '€2,437.01 more than 14 Sep',
		movers: [mover('Salary', '+€2,450.00')],
		smaller: 'and 1 smaller',
		empty: false
	},
	'1 Sep': {
		amount: '€10,912.40',
		change: '€500.00 less than 31 Aug',
		movers: [mover('Transfer to savings', '−€500.00', { transfer: true })],
		empty: false
	},
	'28 Aug': {
		amount: '€12,640.18',
		change: '€9,804.00 more than 27 Aug',
		movers: [mover('Employer Lda', '+€9,804.00', { unusual: true })],
		empty: false
	}
};

export const balancePoints: BalancePoint[] = (() => {
	const points: BalancePoint[] = [];
	for (let a = 0; a < anchors.length - 1; a++) {
		const [fromDate, fromValue] = anchors[a]!;
		const [toDate, toValue] = anchors[a + 1]!;
		const from = Date.parse(fromDate);
		const days = Math.round((Date.parse(toDate) - from) / DAY);
		for (let d = a === 0 ? 0 : 1; d <= days; d++) {
			const at = new Date(from + d * DAY);
			const value = Math.round((fromValue + ((toValue - fromValue) * d) / days) * 100) / 100;
			points.push({
				date: `${at.getUTCDate()} ${MONTHS[at.getUTCMonth()]}`,
				value,
				amount: euro(value)
			});
		}
	}
	return points.map((point, i) => {
		const before = points[i - 1];
		const gap = before ? Math.round((point.value - before.value) * 100) / 100 : 0;
		const change = !before
			? undefined
			: gap === 0
				? `No change from ${before.date}`
				: `${euro(Math.abs(gap))} ${gap < 0 ? 'less' : 'more'} than ${before.date}`;
		return { ...point, change, empty: true, ...days[point.date] };
	});
})();

export const balanceSummary =
	'Balance from 22 Jun to 20 Sep. It starts at €9,640.00 and ends at €11,693.55. Highest €12,118.00 on 2 Sep. Lowest €9,204.00 on 14 Jul.';

export const balanceChart = {
	title: 'Balance · EUR',
	span: '22 Jun to 20 Sep',
	points: balancePoints,
	high: '€12,118',
	low: '€9,204',
	summary: balanceSummary
};
