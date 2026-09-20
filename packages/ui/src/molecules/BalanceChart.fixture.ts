import type { BalancePoint } from './BalanceChart.svelte';

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
	return points;
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
