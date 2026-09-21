import type { NetMonth } from './MonthlyNetChart.svelte';

const month = (
	label: string,
	short: string,
	value: number,
	extra: Partial<NetMonth> = {}
): NetMonth => ({
	label,
	short,
	value,
	net: `${value < 0 ? '−' : '+'}€${Math.abs(value).toLocaleString('en-US', {
		minimumFractionDigits: 2,
		maximumFractionDigits: 2
	})}`,
	...extra
});

/** September 2025 to September 2026, oldest first: twelve full months and the
 *  one under way. March and December ended below zero. */
export const netMonths: NetMonth[] = [
	month('September 2025', 'Sep', 212.4),
	month('October 2025', 'Oct', 180.1),
	month('November 2025', 'Nov', -95.2),
	month('December 2025', 'Dec', -640),
	month('January 2026', 'Jan', 240.55),
	month('February 2026', 'Feb', 310),
	month('March 2026', 'Mar', -1480.3),
	month('April 2026', 'Apr', 198.7),
	month('May 2026', 'May', 225),
	month('June 2026', 'Jun', 150.2),
	month('July 2026', 'Jul', -60.45),
	month('August 2026', 'Aug', 305.4),
	month('September 2026', 'Sep', 637.36, { partial: true })
];

export const netSummary =
	'Net by month from September 2025 to September 2026. 8 of 12 full months ended with more in than out. Highest +€310.00 in February 2026. Lowest −€1,480.30 in March 2026. A typical month is +€189.40.';

export const netChart = {
	title: 'Month by month · EUR',
	span: 'September 2025 to September 2026',
	months: netMonths,
	typical: 189.4,
	high: '+€637',
	low: '−€1,480',
	summary: netSummary
};

/** A ledger from 12 April: April is held from part way through. */
export const heldMonths: NetMonth[] = netMonths.map((m) =>
	m.label === 'April 2026' ? { ...m, partial: true } : m
);

export const twoFullMonths: NetMonth[] = [
	month('July 2026', 'Jul', -60.45),
	month('August 2026', 'Aug', 305.4),
	month('September 2026', 'Sep', 637.36, { partial: true })
];

export const twoFullMonthsSummary =
	'Net by month from July 2026 to September 2026. 1 of 2 full months ended with more in than out. Highest +€305.40 in August 2026. Lowest −€60.45 in July 2026.';
