import { redirect, type Cookies, type ServerLoad } from '@sveltejs/kit';
import { banking } from '$lib/server/banking';
import { call } from '$lib/server/call';
import { formatMoney, isNegative } from '$lib/money';
import {
	CHART_COVERAGE,
	chartSummary,
	dayRange,
	isQuiet,
	monthSection,
	readAt,
	shortDay
} from '$lib/overview';
import {
	chartPoints,
	HISTORY_EMPTY,
	RECURRING_EMPTY,
	historySection,
	recurringEntries,
	scopeNeedsRedirect,
	scopeFromAddress,
	scopeState
} from '$lib/insights';
import {
	AccountGroup,
	type GetMonthHistoryResponse,
	type InsightScope,
	type Account,
	type CurrencyMonth,
	type CurrencyTrend,
	type GetBalanceTrendResponse,
	type GetMonthSummaryResponse,
	type ListAccountsResponse,
	type Transaction
} from '@wimm/contracts/banking';
import type {
	AccountEntry,
	AccountGroups,
	ChartData,
	CurrencySection,
	Merchant,
	RecentTransaction
} from '@wimm/ui';

/**
 * Overview: household money and the member's own, the month so far, the
 * balance chart, top spending, a read-only accounts list and a recent slice of
 * the member's own transactions (`banking/overview`). Account management
 * lives on Accounts; this load carries none of its outcome/handoff wiring.
 */
export const load: ServerLoad = async ({ cookies, depends, url }) => {
	// Named so the page can re-read this data alone once the arrival's balance
	// read lands, without re-running the layout's session check.
	depends('wimm:accounts');

	const view = await overview(cookies, scopeFromAddress(url));
	// The address and the control must agree: an unknown or unavailable scope
	// was answered as All, so the address goes back to the bare one.
	if (view.answered !== undefined && scopeNeedsRedirect(url, view.answered)) redirect(303, '/');
	return view;
};

/**
 * **It never asks a bank.** A load function blocks navigation until it
 * returns, and reading balances is a round trip per account — so reading here
 * means clicking Overview and watching the previous page until every bank has
 * answered. The figures already held render with the time they were read, and
 * the page reads behind that arrival (`+page.svelte`'s own `read-balances`
 * call).
 */
async function overview(cookies: Cookies, scope: InsightScope) {
	try {
		const [accountsView, transactionsView, trendView, monthView, historyView] = await Promise.all([
			call(cookies, (options) => banking.listAccounts({ skipRead: true }, options)),
			// Stored rows only: a sync is Transactions' own concern, bound to its
			// own interval, and Overview's slice does not drive it.
			call(cookies, (options) =>
				banking.listTransactions({ accountId: '', skipSync: true }, options)
			),
			// Reads no bank; computed in wimmd from the same stored ledger
			// (design.md: "the trend is computed in wimmd").
			call(cookies, (options) => banking.getBalanceTrend({ scope }, options)),
			call(cookies, (options) => banking.getMonthSummary({ scope }, options)),
			call(cookies, (options) => banking.getMonthHistory({ scope }, options))
		]);

		const ledger = transactionsView.ledger;
		// A member who owns no account, or whose owned accounts have no
		// transactions read yet, sees no section at all — never one stating
		// there is nothing, which would tell an owner-less member something
		// they are not owed.
		const ownsNothing = ledger?.ownsNoAccount ?? false;
		const transactions = ledger?.transactions ?? [];

		return {
			answered: historyView.scope,
			scope: scopeState(historyView),
			// Before any bank is connected, or a member who may see no account:
			// both are an empty accounts list, and both get Overview's own
			// explanation rather than a total of zero.
			hasAccounts: accountsView.accounts.length > 0,
			ownsNothing,
			currencies: currenciesFrom(accountsView, trendView, monthView, historyView),
			accounts: groupAccounts(accountsView.accounts),
			recentTransactions:
				!ownsNothing && transactions.length > 0
					? transactions.slice(0, 5).map((transaction) => entryOf(transaction, false))
					: []
		};
	} catch {
		// No gateway, or the service behind it is down. Either way the member
		// sees Overview's own empty explanation rather than a failure.
		return {
			answered: undefined,
			scope: undefined,
			hasAccounts: false,
			ownsNothing: false,
			currencies: [],
			accounts: {},
			recentTransactions: []
		};
	}
}

/** One section per currency in use. Household and own figures come from
 *  wimmd already grouped; a currency with zero figures, no trend and no
 *  month summary is dropped here, the one place holding all three. */
function currenciesFrom(
	accountsView: ListAccountsResponse,
	trendView: GetBalanceTrendResponse,
	monthView: GetMonthSummaryResponse,
	historyView: GetMonthHistoryResponse
): CurrencySection[] {
	const household = new Map(accountsView.householdTotals.map((t) => [t.total?.currency ?? '', t]));
	const own = new Map(accountsView.ownTotals.map((t) => [t.total?.currency ?? '', t]));
	const trends = new Map(trendView.trends.map((t) => [t.currency, t]));
	const months = new Map(monthView.months.map((m) => [m.currency, m]));
	const offered = historyView.available.length > 0;
	const histories = new Map(historyView.histories.map((h) => [h.currency, h]));
	const today = new Date();

	const codes = new Set([...household.keys(), ...own.keys(), ...trends.keys(), ...months.keys(), ...histories.keys()]);
	const sections: { section: CurrencySection; weight: number }[] = [];

	for (const currency of codes) {
		if (!currency) continue;
		const h = household.get(currency)?.total;
		const o = own.get(currency)?.total;
		const trend = trends.get(currency);
		const month = months.get(currency);
		const history = histories.get(currency);
		if (isQuiet([h, o], Boolean(trend), Boolean(month))) continue;

		const historyShown = history ? historySection(history, today) : undefined;
		const recurringShown = history
			? recurringEntries(history.recurring, accountsView.accounts, today)
			: undefined;

		sections.push({
			weight: Math.abs(Number(h?.minor ?? 0n)) + Math.abs(Number(o?.minor ?? 0n)),
			section: {
				currency,
				household: household.has(currency) ? formatMoney(h) : undefined,
				own: own.has(currency) ? formatMoney(o) : undefined,
				month: month ? monthSection(month, today) : undefined,
				chart: trend ? chartOf(trend, trendView.partialCoverage) : undefined,
				merchants: month ? merchantsOf(month) : undefined,
				merchantsSpan: month
					? month.countedFrom
						? dayRange(new Date(Number(month.countedFrom.seconds) * 1000), today)
						: dayRange(new Date(Number(month.monthStart?.seconds ?? 0n) * 1000), today)
					: undefined,
				largestPayments: month?.largestPayments.map((transaction) => entryOf(transaction, true)),
				history: historyShown,
				historyEmpty: offered && !historyShown ? HISTORY_EMPTY : undefined,
				recurring: recurringShown,
				recurringEmpty: offered && !recurringShown ? RECURRING_EMPTY : undefined
			}
		});
	}

	return sections.sort((a, b) => b.weight - a.weight).map(({ section }) => section);
}

function chartOf(trend: CurrencyTrend, partialCoverage: boolean): ChartData {
	if (trend.shortHistory || trend.points.length === 0) return { shortHistory: true };

	const points = chartPoints(trend);
	const values = points.map((p) => p.value);
	const at = (value: number) => trend.points.find((p) => Number(p.balance?.minor ?? 0n) === value)?.balance;

	return {
		span: `${points[0]?.date} to ${points[points.length - 1]?.date}`,
		points,
		high: formatMoney(at(Math.max(...values)), { whole: true }),
		low: formatMoney(at(Math.min(...values)), { whole: true }),
		summary: chartSummary(points),
		coverage: partialCoverage ? CHART_COVERAGE : undefined
	};
}

function merchantsOf(month: CurrencyMonth): Merchant[] {
	const top = Number(month.topMerchants[0]?.total?.minor ?? 0n);
	return month.topMerchants.map((merchant) => ({
		name: merchant.name,
		value: `${formatMoney(merchant.total)} · ${merchant.payments} ${merchant.payments === 1 ? 'payment' : 'payments'}`,
		proportion: top > 0 ? Number(merchant.total?.minor ?? 0n) / top : 0
	}));
}

/** The groups are wimmd's decision (`Account.group`); this only sorts them
 *  into the three lists and formats what each row shows. */
function groupAccounts(accounts: Account[]): AccountGroups {
	const groups: Record<'household' | 'own' | 'shared', AccountEntry[]> = {
		household: [],
		own: [],
		shared: []
	};
	const key = {
		[AccountGroup.HOUSEHOLD]: 'household',
		[AccountGroup.OWN]: 'own',
		[AccountGroup.SHARED]: 'shared'
	} as Partial<Record<AccountGroup, 'household' | 'own' | 'shared'>>;

	for (const account of accounts) {
		const group = key[account.group];
		if (!group || account.leftOutAt) continue;
		groups[group].push({
			id: account.id,
			name: account.householdName || account.name,
			bank: account.connection?.bankName ?? '',
			balance: formatMoney(account.balance?.money) ?? '',
			readAt: account.balance?.readAt ? readAt(account.balance.readAt.seconds) : '',
			negative: isNegative(account.balance?.money),
			stale: account.balance?.stale ?? false,
			notUpdating: account.connection ? !account.connection.live : false,
			href: '/accounts'
		});
	}

	return {
		household: groups.household.length ? groups.household : undefined,
		own: groups.own.length ? groups.own : undefined,
		shared: groups.shared.length ? groups.shared : undefined
	};
}

/** A transaction as Overview lists it: the merchant name wimmd derived, and
 *  the account with its bank when `withBank`. */
function entryOf(transaction: Transaction, withBank: boolean): RecentTransaction {
	return {
		id: transaction.id,
		date: transaction.bookingDate ? shortDay(transaction.bookingDate.seconds) : '',
		description: transaction.displayName,
		amount: formatMoney(transaction.amount, { signed: true }) ?? '',
		account:
			withBank && transaction.bankName
				? `${transaction.accountName} · ${transaction.bankName}`
				: transaction.accountName,
		negative: isNegative(transaction.amount)
	};
}
