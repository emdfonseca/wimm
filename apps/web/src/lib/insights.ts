import { formatMoney, isNegative } from './money';
import { dayRange, shortDay } from './overview';
import {
	Cadence,
	type LateLedger,
	InsightScope,
	type Account,
	type CurrencyHistory,
	type CurrencyTrend,
	type GetMonthHistoryResponse,
	type RecurringPayment,
	type Transaction,
	type TrendPoint
} from '@wimm/contracts/banking';
import type {
	BalanceMover,
	BalancePoint,
	HistoryMonth,
	HistorySection,
	HistoryView,
	RecurringEntry,
	ScopeState,
	ScopeValue
} from '@wimm/ui';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
const LONG = [
	'January',
	'February',
	'March',
	'April',
	'May',
	'June',
	'July',
	'August',
	'September',
	'October',
	'November',
	'December'
];

type Money = { minor: bigint; currency?: string };

const date = (seconds: bigint | undefined) => new Date(Number(seconds ?? 0n) * 1000);
const minor = (money?: Money) => Number(money?.minor ?? 0n);
const withCode = (money?: Money) => (money ? { minor: money.minor, currency: money.currency ?? '' } : undefined);
// The screens' words write a minus as U+2212 whatever the locale's own is.
const minus = (text?: string) => (text ?? '').replace(/^-/, '\u2212');
const signed = (money?: Money) => minus(formatMoney(withCode(money), { signed: true }));
const plain = (money?: Money) => formatMoney(withCode(money)) ?? '';
const whole = (money?: Money) => minus(formatMoney(withCode(money), { signed: true, whole: true }));

/** `household` and `yours` are the address's words; `own` is the contract's. */
export function scopeFromAddress(url: URL): InsightScope {
	switch (url.searchParams.get('scope')) {
		case 'household':
			return InsightScope.HOUSEHOLD;
		case 'yours':
			return InsightScope.OWN;
		default:
			return InsightScope.ALL;
	}
}

/** The bare address is All. A parameter that names anything else the service
 *  did not answer, or names nothing known, is sent back to it. */
export function scopeNeedsRedirect(url: URL, answered: InsightScope): boolean {
	const named = url.searchParams.get('scope');
	if (named === null) return false;
	if (named !== 'household' && named !== 'yours' && named !== 'all') return true;
	return scopeFromAddress(url) !== answered;
}

const VALUE: Partial<Record<InsightScope, ScopeValue>> = {
	[InsightScope.HOUSEHOLD]: 'household',
	[InsightScope.OWN]: 'own',
	[InsightScope.ALL]: 'all'
};

function list(names: string[]): string {
	if (names.length <= 1) return names.join('');
	return `${names.slice(0, -1).join(', ')} and ${names[names.length - 1]}`;
}

/** Absent when the member has nothing to choose between. */
export function scopeState(history: GetMonthHistoryResponse): ScopeState | undefined {
	const choices = history.available.flatMap((s) => VALUE[s] ?? []);
	if (choices.length === 0) return undefined;

	let note: string | undefined;
	if (history.householdNotCounted.length > 0) {
		const many = history.householdNotCounted.length > 1;
		note =
			`Household counts ${list(history.householdCounted)}. ` +
			`${list(history.householdNotCounted)} ${many ? 'are' : 'is'} household money too, ` +
			`but ${many ? 'they are' : 'it is'} not yours, so ${many ? 'their' : 'its'} transactions are not counted.`;
	}
	return { choices, value: VALUE[history.scope] ?? 'all', note };
}

/** What follows the bare address: All is nothing at all. */
export const scopeQuery = (value: ScopeValue): string =>
	value === 'household' ? '?scope=household' : value === 'own' ? '?scope=yours' : '';

const monthLabel = (d: Date) => `${LONG[d.getUTCMonth()]} ${d.getUTCFullYear()}`;
const isoDay = (d: Date) => d.toISOString().slice(0, 10);

/** `/transactions?page=2026-03-12.<id>`: the page that starts at the payment. */
export function paymentAddress(transaction: Transaction): string {
	return `/transactions?page=${isoDay(date(transaction.bookingDate?.seconds))}.${transaction.id}`;
}

const fewer = (n: number, one: string, many: string) => (n === 1 ? one : many);

/** `In a typical month €189.40 more comes in than goes out.` */
export function typicalSentence(typical: Money): string {
	const n = minor(typical);
	const size = plain({ ...typical, minor: BigInt(Math.abs(n)) });
	if (n === 0) return 'In a typical month as much comes in as goes out.';
	return n > 0
		? `In a typical month ${size} more comes in than goes out.`
		: `More goes out than comes in. In a typical month ${size} more goes out than comes in.`;
}

interface Bar {
	start: Date;
	full: boolean;
	net: Money;
}

/** The chart in words for one reading of the months. */
function chartName(bars: Bar[], span: string, typical?: Money): string {
	const full = bars.filter((b) => b.full);
	const up = full.filter((b) => minor(b.net) > 0).length;
	let high = full[0];
	let low = full[0];
	for (const b of full) {
		if (minor(b.net) > minor(high?.net)) high = b;
		if (minor(b.net) < minor(low?.net)) low = b;
	}
	const at = (b?: Bar) => (b ? `${signed(b.net)} in ${monthLabel(b.start)}` : '');
	return (
		`Net by month from ${span}. ${up} of ${full.length} full months ended with more in than out. ` +
		`Highest ${at(high)}. Lowest ${at(low)}.` +
		(typical ? ` A typical month is ${signed(typical)}.` : '')
	);
}

function view(
	months: { start: Date; full: boolean; net: Money; drawn: Money }[],
	span: string,
	typical?: Money,
	average?: Money
): HistoryView {
	const drawn = months.map((m) => minor(m.drawn));
	return {
		typical: typical ? signed(typical) : undefined,
		typicalValue: typical ? minor(typical) : undefined,
		average: average ? signed(average) : undefined,
		sentence: typical ? typicalSentence(typical) : undefined,
		summary: chartName(
			months.map((m) => ({ start: m.start, full: m.full, net: m.net })),
			span,
			typical
		),
		high: whole({ ...months[0]!.drawn, minor: BigInt(Math.max(...drawn)) }),
		low: whole({ ...months[0]!.drawn, minor: BigInt(Math.min(...drawn)) })
	};
}

const WAITING = 'A typical month and an average month appear once three full months are held.';
export const HISTORY_EMPTY = 'No full month is held for these accounts yet.';
export const RECURRING_EMPTY = 'No recurring payments in these accounts.';
const NOTHING_TO_SET_ASIDE = 'No unusual payments to set aside.';
const NOTHING_ROSE = 'Nothing took more than usual this month.';
const UNDER_WAY = 'A month under way is not set against whole months.';

/** Month by month, for one currency. Absent where no full month is held. */
export function historySection(c: CurrencyHistory, today: Date): HistorySection | undefined {
	if (c.fullMonths === 0 || c.months.length === 0) return undefined;

	const months: HistoryMonth[] = c.months.map((m) => {
		const start = date(m.monthStart?.seconds);
		const partial = m.soFar || Boolean(m.heldFrom);
		const unusual = m.unusual.length;
		const state = m.soFar
			? 'So far'
			: m.heldFrom
				? `Held from ${shortDay(m.heldFrom.seconds)}`
				: undefined;

		const figures = [
			{ label: 'Money in', value: plain(m.in) },
			{ label: 'Money out', value: plain(m.out) },
			{ label: 'Net', value: signed(m.net) }
		];
		if (m.netUsual) figures.push({ label: 'Without unusual payments', value: signed(m.netUsual) });

		const largest = Math.max(...m.risers.map((r) => minor(r.total) - minor(r.usual)), 0);
		const full = !partial;
		return {
			key: `${start.getUTCFullYear()}-${String(start.getUTCMonth() + 1).padStart(2, '0')}`,
			label: monthLabel(start),
			short: MONTHS[start.getUTCMonth()]!,
			state,
			net: signed(m.net),
			value: minor(m.net),
			netUsual: m.netUsual ? signed(m.netUsual) : undefined,
			valueUsual: m.netUsual ? minor(m.netUsual) : undefined,
			unusualLine: m.netUsual
				? `${unusual} unusual · ${signed(m.netUsual)} without ${fewer(unusual, 'it', 'them')}`
				: undefined,
			partial,
			span: m.soFar ? dayRange(start, today) : undefined,
			figures,
			risers: m.risers.map((r) => {
				const rise = minor(r.total) - minor(r.usual);
				return {
					label: r.name,
					value:
						minor(r.usual) === 0
							? `${plain(r.total)} · not usually paid`
							: `${plain(r.total)} · ${plain({ ...r.total, minor: BigInt(rise) })} more than usual`,
					proportion: largest > 0 ? rise / largest : 0
				};
			}),
			nothingRose: full && c.fullMonths >= 3 && m.risers.length === 0 ? NOTHING_ROSE : undefined,
			note: m.soFar ? UNDER_WAY : undefined,
			payments: m.unusual.length
				? m.unusual.map((u) => {
						const tx = u.transaction!;
						return {
							description: tx.displayName,
							account: tx.accountName,
							amount: signed(tx.amount),
							date: shortDay(tx.bookingDate?.seconds ?? 0n),
							negative: isNegative(tx.amount),
							unusual: true,
							note: u.firstPayment
								? `first payment to ${tx.displayName}`
								: `usually about ${formatMoney(withCode(u.typical), { whole: true }) ?? ''}`,
							href: paymentAddress(tx)
						};
					})
				: undefined
		};
	});

	const oldest = c.months[c.months.length - 1]!;
	const newest = c.months[0]!;
	const span = `${monthLabel(date(oldest.monthStart?.seconds))} to ${monthLabel(date(newest.monthStart?.seconds))}`;
	const bars = c.months.map((m) => ({
		start: date(m.monthStart?.seconds),
		full: !m.soFar && !m.heldFrom,
		net: m.net!,
		drawn: m.net!,
		usual: m.netUsual ?? m.net!
	}));

	const usualCount = c.unusualCount;
	const openMonth = months.find((m) => !m.partial);
	return {
		span,
		months,
		basis:
			c.typicalNet && c.fullMonths >= 3
				? `From ${c.fullMonths} full months. The typical month is the middle one, so one exceptional month barely moves it.`
				: undefined,
		waiting: c.typicalNet ? undefined : WAITING,
		late: c.lateLedgers.length ? c.lateLedgers.map(lateLine) : undefined,
		all: view(bars, span, c.typicalNet, c.averageNet),
		// The control stays wherever a typical month is stated, so a scope with no
		// unusual payment does not lose it; it then says nothing is set aside.
		usual:
			usualCount > 0 || c.typicalNet
				? view(
						bars.map((b) => ({ ...b, net: b.usual, drawn: b.usual })),
						span,
						c.typicalNetUsual,
						c.averageNetUsual
					)
				: undefined,
		setAside:
			usualCount > 0
				? `${usualCount} unusual ${fewer(usualCount, 'payment is', 'payments are')} set aside. ` +
					`${fewer(usualCount, 'It is', 'They are')} still listed in ${fewer(usualCount, 'its month', 'their months')}.`
				: c.typicalNet
					? NOTHING_TO_SET_ASIDE
					: undefined,
		open: openMonth?.key
	};
}

/** `Months before 19 Sep do not include CLASSIC CEMG.`, with the year when it
 *  is not this one. */
function lateLine(l: LateLedger): string {
	const from = date(l.from?.seconds);
	const year = from.getUTCFullYear() === new Date().getUTCFullYear() ? '' : ` ${from.getUTCFullYear()}`;
	return `Months before ${shortDay(l.from?.seconds ?? 0n)}${year} do not include ${l.accountName}.`;
}

const CADENCE: Partial<Record<Cadence, string>> = {
	[Cadence.WEEKLY]: 'Weekly',
	[Cadence.MONTHLY]: 'Monthly',
	[Cadence.YEARLY]: 'Yearly'
};

/** Soonest first, as wimmd sends them. */
export function recurringEntries(
	payments: RecurringPayment[],
	accounts: Account[],
	today: Date
): RecurringEntry[] | undefined {
	if (payments.length === 0) return undefined;
	const byId = new Map(accounts.map((a) => [a.id, a]));

	return payments.map((p, i) => {
		const account = byId.get(p.accountId);
		const where = [account?.householdName || account?.name, account?.connection?.bankName].filter(Boolean);
		const cadence = p.likely ? 'Likely yearly' : (CADENCE[p.cadence] ?? '');
		const expected = date(p.expected?.seconds);
		const day =
			expected.getUTCFullYear() === today.getUTCFullYear()
				? shortDay(p.expected?.seconds ?? 0n)
				: `${shortDay(p.expected?.seconds ?? 0n)} ${expected.getUTCFullYear()}`;
		return {
			id: `${p.name}-${p.cadence}-${i}`,
			name: p.name,
			cadence: [cadence, ...where].join(' · '),
			date: `${p.late ? 'Was expected' : 'Expected'} ${day}`,
			amount: signed(p.amount)
		};
	});
}

function change(balance: Money | undefined, before: TrendPoint | undefined): string | undefined {
	if (!before) return undefined;
	const from = shortDay(before.date?.seconds ?? 0n);
	const diff = minor(balance) - minor(before.balance);
	if (diff === 0) return `No change from ${from}`;
	const size = plain({ minor: BigInt(Math.abs(diff)), currency: balance?.currency ?? '' });
	return `${size} ${diff > 0 ? 'more' : 'less'} than ${from}`;
}

/** A day of the balance chart: its balance, its change and what moved it. */
export function chartPoints(trend: CurrencyTrend): BalancePoint[] {
	return trend.points.map((point, i) => {
		const movers: BalanceMover[] = point.movers.map((m) => ({
			name: m.displayName,
			amount: signed(m.amount),
			negative: isNegative(m.amount),
			unusual: m.unusual
		}));
		return {
			date: shortDay(point.date?.seconds ?? 0n),
			value: minor(point.balance),
			amount: plain(point.balance),
			change: change(point.balance, trend.points[i - 1]),
			movers,
			smaller: point.smaller > 0 ? `and ${point.smaller} smaller` : undefined,
			empty: movers.length === 0 && point.smaller === 0
		};
	});
}
