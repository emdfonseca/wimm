import type { CurrencySection } from '@wimm/ui';
import { formatMoney } from './money';

type Money = { minor: bigint; currency: string };

// Fixed, because ICU spells September "Sept" in some locales and the screen's
// words are "Sep".
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

const day = (seconds: bigint | number) => new Date(Number(seconds) * 1000);

/** `3 Aug`, in UTC: the day a bank named, not an instant. */
export function shortDay(seconds: bigint | number): string {
	const d = day(seconds);
	return `${d.getUTCDate()} ${MONTHS[d.getUTCMonth()]}`;
}

/** `1 to 20 Sep`, or `30 Aug to 2 Sep` when it crosses a month. */
export function dayRange(from: Date, to: Date): string {
	const month = (d: Date) => MONTHS[d.getUTCMonth()];
	const d = (x: Date) => x.getUTCDate();
	return month(from) === month(to)
		? `${d(from)} to ${d(to)} ${month(to)}`
		: `${d(from)} ${month(from)} to ${d(to)} ${month(to)}`;
}

/** The chart in words: its span, both ends, the extremes and when they fell. */
export function chartSummary(points: { date: string; value: number; amount: string }[]): string {
	if (points.length === 0) return '';
	const first = points[0]!;
	const last = points[points.length - 1]!;
	let high = first;
	let low = first;
	for (const point of points) {
		if (point.value > high.value) high = point;
		if (point.value < low.value) low = point;
	}
	return (
		`Balance from ${first.date} to ${last.date}. ` +
		`It starts at ${first.amount} and ends at ${last.amount}. ` +
		`Highest ${high.amount} on ${high.date}. Lowest ${low.amount} on ${low.date}.`
	);
}

export const CHART_COVERAGE =
	'Not every account is in this chart. An account is left out when its bank shares balances only, or when its history is under 30 days.';

export const CHART_SHORT_HISTORY = 'A balance chart appears once there is a week of history.';

/**
 * A currency with zero figures and neither a trend nor a month summary is
 * unused, and gets no tile: `$0.00` beside a flat line says nothing.
 */
export function isQuiet(figures: (Money | undefined)[], hasTrend: boolean, hasMonth: boolean): boolean {
	return !hasTrend && !hasMonth && figures.every((figure) => !figure || figure.minor === 0n);
}

/** `€150.00 more` / `€150.00 less`; nothing where the two are equal. */
export function change(now?: Money, prior?: Money) {
	if (!now || !prior || now.minor === prior.minor) return {};
	const up = now.minor > prior.minor;
	const size = up ? now.minor - prior.minor : prior.minor - now.minor;
	return {
		change: `${formatMoney({ minor: size, currency: now.currency })} ${up ? 'more' : 'less'}`,
		direction: up ? ('up' as const) : ('down' as const)
	};
}

export interface MonthInput {
	in?: Money;
	out?: Money;
	net?: Money;
	priorIn?: Money;
	priorOut?: Money;
	priorNet?: Money;
	countedFrom?: { seconds: bigint };
	monthStart?: { seconds: bigint };
}

const TRANSFERS = 'Money moved between your own accounts is counted.';

/**
 * The month's three tiles and the sentence under them. `today` is the last
 * day of the window; the prior window is the same day number, capped at that
 * month's length, as wimmd computes it.
 */
export function monthSection(month: MonthInput, today: Date): NonNullable<CurrencySection['month']> {
	const start = new Date(Number(month.monthStart?.seconds ?? 0n) * 1000);
	const hasPrior = Boolean(month.priorIn && month.priorOut && month.priorNet);
	const priorStart = new Date(Date.UTC(start.getUTCFullYear(), start.getUTCMonth() - 1, 1));
	const priorLength = new Date(Date.UTC(start.getUTCFullYear(), start.getUTCMonth(), 0)).getUTCDate();
	const priorEnd = new Date(
		Date.UTC(priorStart.getUTCFullYear(), priorStart.getUTCMonth(), Math.min(today.getUTCDate(), priorLength))
	);
	const period = `than ${dayRange(priorStart, priorEnd)}`;

	const figure = (value: Money | undefined, prior: Money | undefined, signed = false) => ({
		value: formatMoney(value, { signed }) ?? '',
		...(hasPrior ? { ...change(value, prior), period } : {})
	});
	const figures = {
		moneyIn: figure(month.in, month.priorIn),
		moneyOut: figure(month.out, month.priorOut),
		net: figure(month.net, month.priorNet, true)
	};
	// A tile with no change to state has no period either.
	for (const tile of Object.values(figures) as { change?: string; period?: string }[]) {
		if (!tile.change) delete tile.period;
	}

	const window = month.countedFrom
		? `Counted from ${shortDay(month.countedFrom.seconds)}.`
		: `${dayRange(start, today)}${hasPrior ? `, against ${dayRange(priorStart, priorEnd)}` : ''}.`;

	return {
		...figures,
		note: `${window} ${TRANSFERS}`,
		href: `/transactions?month=${start.toISOString().slice(0, 7)}`
	};
}

/** `just now`, `12 minutes ago`, `yesterday at 18:04`, `17 Sep at 09:30`. */
export function readAt(seconds: bigint, nowMs: number = Date.now()): string {
	const then = day(seconds);
	const elapsed = Math.round(nowMs / 1000 - Number(seconds));
	if (elapsed < 60) return 'just now';
	const minutes = Math.round(elapsed / 60);
	if (minutes < 60) return `${minutes} ${minutes === 1 ? 'minute' : 'minutes'} ago`;

	const clock = then.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', hour12: false });
	const sameDay = (a: Date, b: Date) => a.toDateString() === b.toDateString();
	const now = new Date(nowMs);
	if (sameDay(then, now)) {
		const hours = Math.round(minutes / 60);
		return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
	}
	if (sameDay(then, new Date(nowMs - 86_400_000))) return `yesterday at ${clock}`;
	return `${then.getDate()} ${MONTHS[then.getMonth()]} at ${clock}`;
}
