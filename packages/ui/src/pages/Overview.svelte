<script lang="ts" module>
	import type { BalancePoint } from '../molecules/BalanceChart.svelte';

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
		ongotoaccounts?: () => void;
		onseeall?: () => void;
	}

	let {
		hasAccounts = false,
		currencies = [],
		accounts = {},
		ownsNothing = false,
		recentTransactions = [],
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

	const several = $derived(currencies.length > 1);
	const suffix = (currency: string) => (several ? ` · ${currency}` : '');

	const groups = $derived(
		[
			{ key: 'household', title: 'Household', rows: accounts.household ?? [] },
			{ key: 'own', title: 'Yours', rows: accounts.own ?? [] },
			{ key: 'shared', title: 'Shared with you', rows: accounts.shared ?? [] }
		].filter((group) => group.rows.length > 0)
	);

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
				hideStatus
				{compact}
			/>
		{/each}
	</div>
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

					{#if section.month}
						<div class="tile-row month" class:compared={!!section.month.moneyIn.change}>
							{@render monthTile('Money in', section.month.moneyIn, section)}
							{@render monthTile('Money out', section.month.moneyOut, section)}
							{@render monthTile('Net', section.month.net, section)}
						</div>
						<p class="note">
							{section.month.note} Money moved between your own accounts is counted.
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
				</section>
			{/each}

			<div class="pair">
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
