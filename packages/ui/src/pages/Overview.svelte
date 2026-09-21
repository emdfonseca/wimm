<script lang="ts" module>
	import type { BalancePoint } from '../molecules/BalanceChart.svelte';
	import type {
		MonthFigure as RowFigure,
		MonthPayment,
		MonthRiser
	} from '../molecules/MonthDetail.svelte';

	/**
	 * J12 · Overview. Where a member lands after signing in: the two figures
	 * that matter, how the month is going, how the balance has moved, where
	 * the most money went, the accounts behind the figures and a recent slice
	 * of their own transactions. It performs no connection action; where a
	 * member wants one, it leads to Accounts.
	 *
	 * Everything here arrives already grouped, computed and formatted. This
	 * screen classifies nothing and formats no money.
	 */
	export interface MonthFigure {
		/** Already formatted, carrying its currency and its sign. */
		value: string;
		/** `€150.00 more`. Absent where last month is not all held, and then
		 *  neither is `period`. */
		change?: string;
		/** `than 1 to 20 Aug`. */
		period?: string;
		direction?: 'up' | 'down';
	}

	export interface MonthSummary {
		moneyIn: MonthFigure;
		moneyOut: MonthFigure;
		net: MonthFigure;
		/** The dates the figures cover, as a sentence: `1 to 20 Sep, against
		 *  1 to 20 Aug.` or `Counted from 19 Sep.` The screen adds what it
		 *  always says about transfers. */
		note: string;
		/** Where the month's transactions are, for each of the three tiles. */
		href: string;
	}

	export interface ChartData {
		/** `22 Jun to 20 Sep`. */
		span?: string;
		points?: BalancePoint[];
		/** The top and bottom of the scale column, already formatted. */
		high?: string;
		low?: string;
		/** The chart in words: the span, both ends and the extremes. */
		summary?: string;
		/** Said under the dates when the chart leaves an account out. */
		coverage?: string;
		/** Less than a week of history anywhere: no line, and a sentence. */
		shortHistory?: boolean;
		/** The index of the day pointed at. */
		marked?: number | null;
	}

	export interface Merchant {
		name: string;
		/** `€412.60 · 9 payments`. One payment reads `1 payment`. */
		value: string;
		/** 0 to 1, against the largest merchant, which is always 1. */
		proportion: number;
	}

	export interface RecentTransaction {
		id: string;
		/** Short: `17 Sep`. */
		date: string;
		/** The merchant name. */
		description: string;
		/** Already formatted, carrying its currency and its sign. */
		amount: string;
		/** The account and its bank: `Current account · Monzo`. */
		account: string;
		negative?: boolean;
		/** One half of a movement between two accounts the member owns:
		 *  `Between your accounts`. */
		transfer?: boolean;
	}

	export type ScopeValue = 'household' | 'own' | 'all';

	/** Which owned accounts everything drawn from transactions counts. Absent
	 *  when the member has nothing to choose between. */
	export interface ScopeState {
		/** Only the choices on offer, in the order shown. */
		choices: ScopeValue[];
		value: ScopeValue;
		/** Said under the control where Household counts fewer accounts than
		 *  the household money figure does. */
		note?: string;
	}

	/** One calendar month of Month by month. */
	export interface HistoryMonth {
		key: string;
		/** `March 2026`. */
		label: string;
		/** `Mar`. Under its bar. */
		short: string;
		/** `So far`, `Held from 12 Apr`. Absent for a full month. */
		state?: string;
		/** Already formatted, carrying its sign. */
		net: string;
		/** What the bar is drawn from. */
		value: number;
		/** The net without unusual payments, and its bar; set only when the
		 *  month held one. */
		netUsual?: string;
		valueUsual?: number;
		/** `2 unusual · +€379.70 without them`. Absent when none. */
		unusualLine?: string;
		/** So far, or held from part way through. */
		partial?: boolean;
		span?: string;
		figures?: RowFigure[];
		risers?: MonthRiser[];
		nothingRose?: string;
		note?: string;
		/** `2 transfers between your accounts left out · €1,400.00`, written
		 *  by the load. Absent where the scope left none out. */
		leftOut?: string;
		/** Unusual payments, newest first. */
		payments?: MonthPayment[];
	}

	/** What one reading of the months says: all payments, or without the
	 *  unusual ones. */
	export interface HistoryView {
		/** `+€189.40`. Absent under three full months. */
		typical?: string;
		/** The typical month in the bars' unit. */
		typicalValue?: number;
		average?: string;
		/** `In a typical month €189.40 more comes in than goes out.` */
		sentence?: string;
		/** The chart in words. */
		summary: string;
		high: string;
		low: string;
	}

	export interface HistorySection {
		/** `September 2025 to September 2026`. */
		span: string;
		/** Newest first. */
		months: HistoryMonth[];
		/** `From 12 full months. The typical month is …` */
		basis?: string;
		/** Said in the tiles' place under three full months. */
		waiting?: string;
		/** One line per account missing from some full month: `Months before
		 *  19 Sep do not include CLASSIC CEMG.` */
		late?: string[];
		all: HistoryView;
		/** Absent when no month held an unusual payment: no view control. */
		usual?: HistoryView;
		/** `3 unusual payments are set aside. They are still listed in their
		 *  months.` */
		setAside?: string;
		/** The key of the month open as the screen loads. */
		open?: string;
	}

	export interface RecurringEntry {
		id: string;
		name: string;
		/** `Monthly · Current account · Monzo`. */
		cadence: string;
		/** `Expected 6 Oct`, `Was expected 18 Sep`. */
		date: string;
		/** Already formatted, carrying its sign. */
		amount: string;
	}

	/** Everything Overview shows for one currency. A currency is never
	 *  combined with another. */
	export interface CurrencySection {
		currency: string;
		/** Already formatted. Absent where no account counts as household
		 *  money in this currency, and then the other tile takes the row. */
		household?: string;
		/** Already formatted. Absent where the member owns no account of their
		 *  own in this currency. */
		own?: string;
		month?: MonthSummary;
		chart?: ChartData;
		merchants?: Merchant[];
		/** `1 to 20 Sep`. */
		merchantsSpan?: string;
		largestPayments?: RecentTransaction[];
		/** Absent where no full month is held. */
		history?: HistorySection;
		/** Said in Month by month's place where the scope's accounts hold no full
		 *  month. Set only while a scope control is on offer. */
		historyEmpty?: string;
		/** Soonest first. Absent where nothing recurs. */
		recurring?: RecurringEntry[];
		/** Said in Recurring payments' place where the scope's accounts hold
		 *  none. Set only while a scope control is on offer. */
		recurringEmpty?: string;
		/** How this scope treats movements between the member's own accounts,
		 *  written by the load because which sentence is right depends on the
		 *  scope in force. Shown under the month summary and under the months.
		 */
		transferNote?: string;
	}

	export interface AccountEntry {
		id: string;
		name: string;
		bank: string;
		/** Already formatted, carrying its currency and its sign. */
		balance: string;
		/** `just now`, `yesterday at 18:04`. */
		readAt: string;
		negative?: boolean;
		stale?: boolean;
		notUpdating?: boolean;
		/** Where the row leads. Accounts, always. */
		href?: string;
	}

	/** Which figure each account counts towards. Decided by wimmd. */
	export interface AccountGroups {
		household?: AccountEntry[];
		own?: AccountEntry[];
		shared?: AccountEntry[];
	}
</script>

<script lang="ts">
	import Button from '../atoms/Button.svelte';
	import Icon from '../atoms/Icon.svelte';
	import AccountRow from '../molecules/AccountRow.svelte';
	import BalanceChart from '../molecules/BalanceChart.svelte';
	import BudgetMeter from '../molecules/BudgetMeter.svelte';
	import EmptyState from '../molecules/EmptyState.svelte';
	import LedgerRow from '../molecules/LedgerRow.svelte';
	import MetricTile from '../molecules/MetricTile.svelte';
	import MonthlyNetChart from '../molecules/MonthlyNetChart.svelte';
	import MonthDetail from '../molecules/MonthDetail.svelte';
	import MonthTable from '../molecules/MonthTable.svelte';
	import SegmentedControl from '../atoms/SegmentedControl.svelte';
	import Page from '../templates/Page.svelte';

	interface Props {
		/** False before any bank is connected, or where the member may see no
		 *  account — both get this screen's own explanation, never a total of
		 *  zero. */
		hasAccounts?: boolean;
		/** Largest total first. */
		currencies?: CurrencySection[];
		accounts?: AccountGroups;
		/** This member owns no account, whatever they were granted: nothing
		 *  here changes anything for them, so the way to Accounts management
		 *  is not offered. */
		ownsNothing?: boolean;
		/** Absent, not empty, where the member owns nothing or none of what
		 *  they own has a transaction read yet. */
		recentTransactions?: RecentTransaction[];
		/** Absent where the member has nothing to choose between. */
		scope?: ScopeState;
		onscope?: (value: ScopeValue) => void;
		ongotoaccounts?: () => void;
		onseeall?: () => void;
	}

	let {
		hasAccounts = false,
		currencies = [],
		accounts = {},
		ownsNothing = false,
		recentTransactions = [],
		scope,
		onscope,
		ongotoaccounts,
		onseeall
	}: Props = $props();

	/** `768px` is the declared Compact/Medium boundary (ADR 0002). Read
	 *  directly rather than passed down, because every caller would otherwise
	 *  need to know a layout breakpoint that is this screen's concern alone. */
	let compact = $state(false);

	$effect(() => {
		if (typeof window === 'undefined' || !window.matchMedia) return;
		const query = window.matchMedia('(max-width: 767px)');
		compact = query.matches;
		const onchange = (e: MediaQueryListEvent) => (compact = e.matches);
		query.addEventListener('change', onchange);
		return () => query.removeEventListener('change', onchange);
	});

	/** From `1200px` the month detail stands beside the table, in a panel too
	 *  narrow for a full ledger row, so its payments take the stacked form. */
	let beside = $state(false);

	$effect(() => {
		if (typeof window === 'undefined' || !window.matchMedia) return;
		const query = window.matchMedia('(min-width: 1200px)');
		beside = query.matches;
		const onchange = (e: MediaQueryListEvent) => (beside = e.matches);
		query.addEventListener('change', onchange);
		return () => query.removeEventListener('change', onchange);
	});

	const several = $derived(currencies.length > 1);
	const suffix = (currency: string) => (several ? ` · ${currency}` : '');

	const groups = $derived(
		[
			{ key: 'household', title: 'Household', rows: accounts.household ?? [] },
			{ key: 'own', title: 'Yours', rows: accounts.own ?? [] },
			{ key: 'shared', title: 'Shared with you', rows: accounts.shared ?? [] }
		].filter((group) => group.rows.length > 0)
	);

	const scopeLabels: Record<ScopeValue, string> = {
		household: 'Household',
		own: 'Yours',
		all: 'All'
	};

	/** Component state, gone on reload by design: the default reading is the
	 *  true one. */
	let views: Record<string, 'all' | 'usual'> = $state({});
	let picked: Record<string, string | null> = $state({});
	const uid = $props.id();

	/** One selection per currency, shared by the chart's marker, the table and
	 *  the detail panel. */
	const selectedKey = (section: CurrencySection) =>
		section.currency in picked ? picked[section.currency] : (section.history?.open ?? null);

	const selectedMonth = (section: CurrencySection) =>
		section.history?.months.find((m) => m.key === selectedKey(section));

	const figureOf = (month: HistoryMonth, label: string) =>
		month.figures?.find((f) => f.label === label)?.value;

	const markOf = (section: CurrencySection) => {
		const oldestFirst = [...(section.history?.months ?? [])].reverse();
		const at = oldestFirst.findIndex((m) => m.key === selectedKey(section));
		return at < 0 ? null : at;
	};

	const viewOf = (section: CurrencySection) => views[section.currency] ?? 'all';

	const chartMonths = (section: CurrencySection) => {
		const usual = viewOf(section) === 'usual';
		return [...(section.history?.months ?? [])].reverse().map((m) => ({
			label: m.label,
			short: m.short,
			value: usual ? (m.valueUsual ?? m.value) : m.value,
			net: usual ? (m.netUsual ?? m.net) : m.net,
			partial: m.partial
		}));
	};

	function point(section: CurrencySection, index: number | null) {
		if (index === null) return;
		const oldestFirst = [...(section.history?.months ?? [])].reverse();
		const month = oldestFirst[index];
		if (month) picked[section.currency] = month.key;
	}

	const withMerchants = $derived(currencies.filter((c) => (c.merchants?.length ?? 0) > 0));
	const withLargest = $derived(currencies.filter((c) => (c.largestPayments?.length ?? 0) > 0));
</script>

{#snippet moneyTiles(section: CurrencySection)}
	{@const tiles = [
		{ label: 'Household money', value: section.household },
		{ label: 'Your money', value: section.own }
	].filter((tile) => tile.value !== undefined)}
	{#if tiles.length > 0}
		<div class="tile-row money">
			{#each tiles as tile (tile.label)}
				<MetricTile label={`${tile.label}${suffix(section.currency)}`} value={tile.value!} />
			{/each}
		</div>
	{/if}
{/snippet}

{#snippet monthTile(label: string, figure: MonthFigure, section: CurrencySection)}
	<MetricTile
		label={`${label}${suffix(section.currency)}`}
		value={figure.value}
		href={section.month?.href}
		delta={figure.change && figure.period
			? { change: figure.change, period: figure.period, direction: figure.direction ?? 'up' }
			: undefined}
	/>
{/snippet}

{#snippet transactionRows(rows: RecentTransaction[])}
	<div class="rows">
		{#each rows as entry (entry.id)}
			<LedgerRow
				description={entry.description}
				account={entry.account}
				amount={entry.amount}
				date={entry.date}
				negative={entry.negative}
				transfer={entry.transfer}
				hideStatus
				{compact}
			/>
		{/each}
	</div>
{/snippet}

{#snippet insights(section: CurrencySection)}
	{@const history = section.history}
	{#if history}
		{@const view = viewOf(section)}
		{@const shown = view === 'usual' && history.usual ? history.usual : history.all}
		{@const detail = selectedMonth(section)}
		{@const detailId = `${uid}-${section.currency}`}
		<section class="history" aria-label={`Months${suffix(section.currency)}`}>
			{#if shown.typical}
				<div class="tile-row typical">
					<MetricTile label="Typical month" value={shown.typical} />
					<MetricTile label="Average month" value={shown.average ?? ''} />
				</div>
				{#if shown.sentence}<p class="note">{shown.sentence}</p>{/if}
				{#if history.basis}<p class="note">{history.basis}</p>{/if}
			{:else if history.waiting}
				<p class="note">{history.waiting}</p>
			{/if}
			{#each history.late ?? [] as line (line)}
				<p class="note">{line}</p>
			{/each}

			{#if history.usual}
				<div class="view">
					<SegmentedControl
						label="Payments counted"
						options={[
							{ value: 'all', label: 'All payments' },
							{ value: 'usual', label: 'Without unusual payments' }
						]}
						value={view}
						onchange={(value) => (views[section.currency] = value as 'all' | 'usual')}
					/>
					{#if view === 'usual' && history.setAside}
						<p class="note">{history.setAside}</p>
					{/if}
				</div>
			{/if}

			<MonthlyNetChart
				title={`Month by month${suffix(section.currency)}`}
				span={history.span}
				months={chartMonths(section)}
				typical={shown.typicalValue}
				high={shown.high}
				low={shown.low}
				summary={shown.summary}
				bind:marked={() => markOf(section), (index) => point(section, index)}
			/>

			<div class="month-layout">
				<!-- The detail comes first: stacked, it stays under the chart that
				     selects it rather than below thirteen rows. -->
				{#if detail}
					<div class="month-detail">
						<MonthDetail
							id={detailId}
							month={detail.label}
							span={detail.span}
							figures={detail.figures}
							risers={detail.risers}
							nothingRose={detail.nothingRose}
							note={detail.note}
							leftOut={detail.leftOut}
							payments={detail.payments}
							compact={compact || beside}
						/>
					</div>
				{/if}
				<div class="month-table">
					<MonthTable
						label={`Months${suffix(section.currency)}`}
						rows={history.months.map((month) => ({
							key: month.key,
							label: month.label,
							state: month.state,
							moneyIn: figureOf(month, 'Money in'),
							moneyOut: figureOf(month, 'Money out'),
							net: view === 'usual' ? (month.netUsual ?? month.net) : month.net,
							unusualLine: month.unusualLine
						}))}
						controls={detail ? detailId : undefined}
						{compact}
						bind:selected={() => selectedKey(section) ?? null, (key) => (picked[section.currency] = key)}
					/>
				</div>
			</div>
			{#if section.transferNote}
				<p class="note">{section.transferNote}</p>
			{/if}
		</section>
	{:else if section.historyEmpty}
		<section class="card" aria-label={`Months${suffix(section.currency)}`}>
			<header class="card-header">
				<h2 class="card-title">Month by month{suffix(section.currency)}</h2>
			</header>
			<p class="note">{section.historyEmpty}</p>
		</section>
	{/if}

	{#if (section.recurring?.length ?? 0) === 0 && section.recurringEmpty}
		<section class="card" aria-label={`Recurring payments${suffix(section.currency)}`}>
			<header class="card-header">
				<h2 class="card-title">Recurring payments{suffix(section.currency)}</h2>
			</header>
			<p class="note">{section.recurringEmpty}</p>
		</section>
	{/if}

	{#if (section.recurring?.length ?? 0) > 0}
		<section class="card" aria-label={`Recurring payments${suffix(section.currency)}`}>
			<header class="card-header">
				<h2 class="card-title">Recurring payments{suffix(section.currency)}</h2>
			</header>
			<div class="rows">
				{#each section.recurring ?? [] as entry (entry.id)}
					<LedgerRow
						description={entry.name}
						account={entry.cadence}
						date={entry.date}
						amount={entry.amount}
						negative={entry.amount.startsWith('−')}
						hideStatus
						{compact}
					/>
				{/each}
			</div>
		</section>
	{/if}
{/snippet}

{#snippet merchantsBlock()}
	{#if withMerchants.length > 0}
		<div class="stack merchants">
			{#each withMerchants as section (section.currency)}
				<section class="card" aria-label={`Top merchants${suffix(section.currency)}`}>
					<header class="card-header">
						<h2 class="card-title">Top merchants{suffix(section.currency)}</h2>
						<span class="spacer"></span>
						{#if section.merchantsSpan}
							<span class="card-span">{section.merchantsSpan}</span>
						{/if}
					</header>
					<div class="meters">
						{#each section.merchants ?? [] as merchant (merchant.name)}
							<BudgetMeter
								label={merchant.name}
								value={merchant.value}
								proportion={merchant.proportion}
							/>
						{/each}
					</div>
				</section>
			{/each}
		</div>
	{/if}
{/snippet}

{#snippet accountsBlock()}
	{#if groups.length > 0}
		<section class="card accounts" aria-label="Accounts">
			<header class="card-header">
				<h2 class="card-title">Accounts</h2>
				<span class="spacer"></span>
				{#if !ownsNothing}
					<Button variant="secondary" onclick={ongotoaccounts}>Manage accounts</Button>
				{/if}
			</header>
			{#each groups as group (group.key)}
				<div class="group" role="group" aria-labelledby="group-{group.key}">
					<h3 class="group-heading" id="group-{group.key}">{group.title}</h3>
					<div class="rows inset">
						{#each group.rows as account (account.id)}
							<AccountRow
								name={account.name}
								bank={account.bank}
								balance={account.balance}
								readAt={account.readAt}
								negative={account.negative}
								stale={account.stale}
								notUpdating={account.notUpdating}
								href={account.href ?? '/accounts'}
							/>
						{/each}
					</div>
				</div>
			{/each}
		</section>
	{/if}
{/snippet}

<Page title="Overview">
	<div class="screen" class:compact>
		{#if !hasAccounts}
			<EmptyState title="Connect a bank to see where you stand" elevated>
				Once a bank is connected, your total, your recent activity and how it is trending show up
				here.
				{#snippet icon()}<Icon name="landmark" size={24} />{/snippet}
				{#snippet action()}
					<Button onclick={ongotoaccounts}>Go to Accounts</Button>
				{/snippet}
			</EmptyState>
		{:else}
			{#each currencies as section, index (section.currency)}
				<section class="currency" aria-label={several ? section.currency : undefined}>
					{@render moneyTiles(section)}
					{#if index === 0 && several}
						<!-- wimm holds no rates, and a converted figure would invent the
						     number a household trusts most. -->
						<p class="note">Shown per currency. wimm does not convert between them.</p>
					{/if}

					{#if index === 0 && scope && scope.choices.length > 0}
						<div class="scope">
							<SegmentedControl
								label="Accounts counted"
								options={scope.choices.map((value) => ({ value, label: scopeLabels[value] }))}
								value={scope.value}
								onchange={(value) => onscope?.(value as ScopeValue)}
							/>
							{#if scope.note}
								<p class="note">{scope.note}</p>
							{/if}
						</div>
					{/if}

					{#if section.month}
						<div class="tile-row month" class:compared={!!section.month.moneyIn.change}>
							{@render monthTile('Money in', section.month.moneyIn, section)}
							{@render monthTile('Money out', section.month.moneyOut, section)}
							{@render monthTile('Net', section.month.net, section)}
						</div>
						<p class="note">
							{section.month.note}
							{section.transferNote}
						</p>
					{/if}

					{#if section.chart}
						<BalanceChart
							title={`Balance · ${section.currency}`}
							span={section.chart.span}
							points={section.chart.points}
							high={section.chart.high}
							low={section.chart.low}
							summary={section.chart.summary}
							coverage={section.chart.coverage}
							shortHistory={section.chart.shortHistory}
							marked={section.chart.marked ?? null}
						/>
					{/if}

					{#if !compact}
						{@render insights(section)}
					{/if}
				</section>
			{/each}

			<div class="pair">
				{#if compact}
					{@render accountsBlock()}
					{#each currencies as section (section.currency)}
						{@render insights(section)}
					{/each}
					{@render merchantsBlock()}
				{:else}
					{@render merchantsBlock()}
					{@render accountsBlock()}
				{/if}
			</div>

			{#if withLargest.length > 0 || recentTransactions.length > 0}
				<div class="pair">
					{#if withLargest.length > 0}
						<div class="stack">
							{#each withLargest as section (section.currency)}
								<section class="card" aria-label={`Largest payments${suffix(section.currency)}`}>
									<header class="card-header">
										<h2 class="card-title">Largest payments{suffix(section.currency)}</h2>
									</header>
									{@render transactionRows(section.largestPayments ?? [])}
								</section>
							{/each}
						</div>
					{/if}

					{#if recentTransactions.length > 0}
						<section class="card" aria-label="Recent transactions">
							<header class="card-header">
								<h2 class="card-title">Recent transactions</h2>
								<span class="spacer"></span>
								<Button variant="secondary" onclick={onseeall}>See all</Button>
							</header>
							{@render transactionRows(recentTransactions)}
						</section>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</Page>

<style>
	.screen {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-block-size: 0;
		gap: var(--space-6);
		inline-size: 100%;
		font-family: var(--type-family-body);
	}

	.compact {
		gap: var(--space-4);
	}

	.currency {
		display: flex;
		flex-direction: column;
		gap: var(--space-6);
	}

	.compact .currency {
		gap: var(--space-4);
	}

	.tile-row {
		display: flex;
		gap: 16px;
	}

	.tile-row > :global(*) {
		flex: 1 1 0;
		min-inline-size: 0;
	}

	.money {
		block-size: 82px;
	}

	.month {
		block-size: 82px;
	}

	.month.compared {
		block-size: 102px;
	}

	.compact .month {
		flex-direction: column;
		block-size: auto;
	}

	.scope,
	.view,
	.history {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.scope :global(.segmented),
	.view :global(.segmented) {
		align-self: flex-start;
	}

	.typical {
		block-size: 82px;
	}

	.compact .typical {
		flex-direction: column;
		block-size: auto;
	}

	/* The detail and the table stack, detail first, until there is room to
	   stand the table beside it. */
	.month-layout {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 16px;
		align-items: start;
	}

	.note {
		margin: 0;
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
	}

	/* Each pair sits side by side from Wide up. Below that they stack, and at
	   Compact the accounts come before the merchants: the balance is the more
	   common reason to open the app on a phone. */
	.pair {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 24px;
		align-items: start;
	}

	.compact .pair {
		gap: 16px;
	}

	.compact .accounts {
		order: -1;
	}

	@media (min-width: 1200px) {
		.month-layout {
			grid-template-columns: minmax(0, 3fr) minmax(0, 2fr);
		}

		.month-table {
			grid-column: 1;
			grid-row: 1;
		}

		.month-table:only-child {
			grid-column: 1 / -1;
		}

		.month-detail {
			grid-column: 2;
			grid-row: 1;
			position: sticky;
			inset-block-start: 16px;
		}

		.pair {
			grid-template-columns: repeat(2, minmax(0, 1fr));
			/* Half the column is too narrow for a ledger row's account column. */
			--ledger-account-basis: auto;
		}

		/* A card with nothing beside it takes the row, as a lone money tile does:
		   half a row of Accounts beside half a row of nothing reads as missing. */
		.pair > :only-child {
			grid-column: 1 / -1;
		}
	}

	.stack {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.compact .stack {
		gap: 16px;
	}

	.card {
		display: flex;
		flex-direction: column;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: 12px;
		min-block-size: 48px;
		padding-inline: 16px;
		border-block-end: 1px solid var(--color-border-subtle);
	}

	.card-title {
		margin: 0;
		color: var(--color-text-primary);
		font-size: var(--type-size-body-md);
		font-weight: 700;
	}

	.card-span {
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
	}

	.spacer {
		flex: 1 1 auto;
	}

	/* A card holding only a sentence: the sentence takes the card's own inset. */
	.card > .note {
		padding: 16px;
	}

	.rows {
		display: flex;
		flex-direction: column;
	}

	.rows > :global(* + *) {
		border-block-start: 1px solid var(--color-border-subtle);
	}

	.inset {
		padding-inline: 6px;
	}

	.group-heading {
		display: flex;
		align-items: center;
		margin: 0;
		block-size: 28px;
		padding-inline: 16px;
		color: var(--color-text-secondary);
		font-size: 11px;
		font-weight: 600;
	}

	.meters {
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 16px;
	}
</style>
