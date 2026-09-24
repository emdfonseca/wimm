<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import { balanceChart, balancePoints, balanceSummary } from '../molecules/BalanceChart.fixture';
	import Overview, {
		type AccountEntry,
		type CurrencySection,
		type HistorySection,
		type RecentTransaction,
		type ScopeState
	} from './Overview.svelte';
	import {
		allView,
		bonusPayment,
		cleaning,
		fidelidade,
		history,
		historyMonths,
		netflix,
		nos,
		recurring,
		rent,
		spotify,
		usualView
	} from './Overview.fixture';

	const joint: AccountEntry = {
		id: 'a1',
		name: 'Joint account',
		bank: 'Monzo',
		balance: '€6,698.00',
		readAt: 'just now'
	};
	const current: AccountEntry = {
		id: 'a2',
		name: 'Current account',
		bank: 'Monzo',
		balance: '€2,480.55',
		readAt: 'just now'
	};
	const conta: AccountEntry = {
		id: 'a3',
		name: 'Conta à Ordem',
		bank: 'Montepio',
		balance: '€2,515.00',
		readAt: 'yesterday at 18:04'
	};
	const savings: AccountEntry = {
		id: 'a4',
		name: 'Savings',
		bank: 'Montepio',
		balance: '€3,200.00',
		readAt: 'just now'
	};

	const accounts = { household: [joint], own: [current, conta], shared: [savings] };

	// The two sentences the load writes. Which one is right depends on the
	// scope in force, so the screen is given one and holds neither.
	const transferNote =
		'Money moved between your accounts is left out where wimm holds both. Other transfers are counted.';
	const narrowTransferNote =
		'Money moved between accounts counted here is left out. Money moved to or from your other accounts is counted.';

	const recent: RecentTransaction[] = [
		{
			id: 't1',
			date: '17 Sep',
			description: 'Pingo Doce',
			account: 'Current account · Monzo',
			amount: '−€42.18',
			negative: true
		},
		{
			id: 't2',
			date: '15 Sep',
			description: 'Salary',
			account: 'Current account · Monzo',
			amount: '+€2,450.00'
		},
		{
			id: 't-transfer',
			date: '1 Sep',
			description: 'Transfer to savings',
			account: 'Current account · Monzo',
			amount: '−€500.00',
			negative: true,
			transfer: true
		},
		{
			id: 't3',
			date: '14 Sep',
			description: 'Netflix',
			account: 'Current account · Monzo',
			amount: '−€12.99',
			negative: true
		}
	];

	const largest: RecentTransaction[] = [
		{
			id: 'l1',
			date: '6 Sep',
			description: 'Rent',
			account: 'Joint account · Monzo',
			amount: '−€820.00',
			negative: true
		},
		{
			id: 'l2',
			date: '12 Sep',
			description: 'Galp',
			account: 'Current account · Monzo',
			amount: '−€92.10',
			negative: true
		},
		{
			id: 'l3',
			date: '14 Sep',
			description: 'Pingo Doce',
			account: 'Current account · Monzo',
			amount: '−€88.34',
			negative: true
		},
		{
			id: 'l4',
			date: '9 Sep',
			description: 'Amazon',
			account: 'Current account · Monzo',
			amount: '−€64.99',
			negative: true
		},
		{
			id: 'l5',
			date: '2 Sep',
			description: 'EDP',
			account: 'Joint account · Monzo',
			amount: '−€58.20',
			negative: true
		}
	];

	// Under Yours, ada's standing 800.00 to the joint account: money that
	// leaves the accounts Yours counts, so it is counted, recurring and large,
	// and still one half of a transfer.
	const jointPayment: RecentTransaction = {
		id: 'l-joint',
		date: '1 Sep',
		description: 'Joint account',
		account: 'Current account · Monzo',
		amount: '−€800.00',
		negative: true,
		transfer: true
	};

	const jointTransfer = {
		id: 'r-joint',
		name: 'Joint account',
		cadence: 'Monthly · Current account · Monzo',
		date: 'Expected 1 Oct',
		amount: '−€800.00'
	};

	const merchants = [
		{ name: 'Pingo Doce', value: '€412.60 · 9 payments', proportion: 1 },
		{ name: 'Galp', value: '€186.40 · 3 payments', proportion: 0.45 },
		{ name: 'Amazon', value: '€142.18 · 4 payments', proportion: 0.34 },
		{ name: 'Via Verde', value: '€64.30 · 12 payments', proportion: 0.16 },
		{ name: 'Netflix', value: '€12.99 · 1 payment', proportion: 0.03 }
	];

	const month = {
		moneyIn: {
			value: '€2,450.00',
			change: '€150.00 more',
			period: 'than 1 to 20 Aug',
			direction: 'up' as const
		},
		moneyOut: {
			value: '€1,812.64',
			change: '€182.40 more',
			period: 'than 1 to 20 Aug',
			direction: 'up' as const
		},
		net: {
			value: '+€637.36',
			change: '€32.40 less',
			period: 'than 1 to 20 Aug',
			direction: 'down' as const
		},
		note: '1 to 20 Sep, against 1 to 20 Aug.',
		href: '/transactions?month=2026-09'
	};

	const eur: CurrencySection = {
		currency: 'EUR',
		household: '€6,698.00',
		own: '€4,995.55',
		month,
		chart: {
			span: balanceChart.span,
			points: balancePoints,
			high: balanceChart.high,
			low: balanceChart.low,
			summary: balanceSummary
		},
		merchants,
		merchantsSpan: '1 to 20 Sep',
		largestPayments: largest,
		history,
		recurring,
		transferNote
	};

	const allScope: ScopeState = { choices: ['household', 'own', 'all'], value: 'all' };

	const gbpHistory: HistorySection = {
		span: 'July 2026 to September 2026',
		months: [
			{
				key: 'Sep',
				label: 'September 2026',
				short: 'Sep',
				state: 'So far',
				net: '+£489.70',
				value: 489.7,
				partial: true
			},
			{ key: 'Aug', label: 'August 2026', short: 'Aug', net: '+£120.00', value: 120 },
			{ key: 'Jul', label: 'July 2026', short: 'Jul', net: '−£45.00', value: -45 }
		],
		waiting: 'A typical month and an average month appear once three full months are held.',
		all: {
			summary:
				'Net by month from July 2026 to September 2026. 1 of 2 full months ended with more in than out. Highest +£120.00 in August 2026. Lowest −£45.00 in July 2026.',
			high: '+£490',
			low: '−£45'
		}
	};

	const gbp: CurrencySection = {
		currency: 'GBP',
		household: '£1,200.00',
		own: '£640.20',
		month: {
			moneyIn: { value: '£900.00' },
			moneyOut: { value: '£410.30' },
			net: { value: '+£489.70' },
			note: '1 to 20 Sep, against 1 to 20 Aug.',
			href: '/transactions?month=2026-09'
		},
		chart: {
			span: '22 Jun to 20 Sep',
			points: balancePoints,
			high: '£1,900',
			low: '£1,100',
			summary:
				'Balance from 22 Jun to 20 Sep. It starts at £1,100.00 and ends at £1,840.20. Highest £1,900.00 on 2 Sep. Lowest £1,100.00 on 22 Jun.'
		},
		history: gbpHistory,
		recurring: [
			{
				id: 'gr1',
				name: 'Council tax',
				cadence: 'Monthly · Sterling account · Monzo',
				date: 'Expected 1 Oct',
				amount: '−£142.00'
			}
		],
		merchants: [
			{ name: 'Tesco', value: '£210.10 · 5 payments', proportion: 1 },
			{ name: 'Boots', value: '£38.20 · 2 payments', proportion: 0.18 }
		],
		merchantsSpan: '1 to 20 Sep',
		transferNote,
		largestPayments: [
			{
				id: 'g1',
				date: '8 Sep',
				description: 'Hackney Council',
				account: 'Sterling account · Monzo',
				amount: '−£142.00',
				negative: true
			}
		]
	};

	const months = (canvas: ReturnType<typeof within>) => region(canvas, 'Months');

	const monthRow = (canvas: ReturnType<typeof within>, name: string) =>
		months(canvas).getByRole('button', { name: new RegExp(name) });

	const panelOf = (button: HTMLElement) =>
		within(document.getElementById(button.getAttribute('aria-controls')!)!);

	const liveOf = (root: Element) => root.querySelector('[aria-live="polite"]')!;

	const region = (canvas: ReturnType<typeof within>, name: string) =>
		within(canvas.getByRole('region', { name }));

	const tile = (canvas: ReturnType<typeof within>, label: string) =>
		canvas.getByText(label).closest('.tile') as HTMLElement;

	const { Story } = defineMeta({
		title: 'Pages/Overview',
		component: Overview,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen', shell: '/' },
		args: {
			hasAccounts: true,
			currencies: [eur],
			accounts,
			recentTransactions: recent,
			scope: allScope,
			onscope: fn(),
			onseeall: fn(),
			ongotoaccounts: fn()
		}
	});
</script>

<Story
	name="Populated"
	tags={['kind-state']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: 'Overview', level: 1 })).toBeInTheDocument();

		await expect(tile(canvas, 'Household money')).toHaveTextContent('€6,698.00');
		await expect(tile(canvas, 'Your money')).toHaveTextContent('€4,995.55');

		const moneyIn = canvas.getByRole('link', { name: /Money in/ });
		await expect(moneyIn).toHaveTextContent('€2,450.00');
		await expect(moneyIn).toHaveTextContent('€150.00 more');
		await expect(moneyIn).toHaveTextContent('than 1 to 20 Aug');
		const moneyOut = canvas.getByRole('link', { name: /Money out/ });
		await expect(moneyOut).toHaveTextContent('€1,812.64');
		await expect(moneyOut).toHaveTextContent('€182.40 more');
		const net = canvas.getByRole('link', { name: /Net/ });
		await expect(net).toHaveTextContent('+€637.36');
		await expect(net).toHaveTextContent('€32.40 less');
		await expect(net).toHaveTextContent('than 1 to 20 Aug');

		await expect(
			canvas.getByText(`1 to 20 Sep, against 1 to 20 Aug. ${transferNote}`)
		).toBeInTheDocument();

		const chart = region(canvas, 'Balance · EUR');
		await expect(chart.getByText('22 Jun to 20 Sep')).toBeInTheDocument();
		await expect(chart.getByText('€12,118')).toBeInTheDocument();
		await expect(chart.getByText('€9,204')).toBeInTheDocument();
		await expect(chart.getByText('22 Jun')).toBeInTheDocument();
		await expect(chart.getByText('20 Sep')).toBeInTheDocument();
		await expect(chart.getByRole('img', { name: balanceSummary })).toBeInTheDocument();

		const top = region(canvas, 'Top merchants');
		await expect(top.getByText('1 to 20 Sep')).toBeInTheDocument();
		for (const [name, value] of [
			['Pingo Doce', '€412.60 · 9 payments'],
			['Galp', '€186.40 · 3 payments'],
			['Amazon', '€142.18 · 4 payments'],
			['Via Verde', '€64.30 · 12 payments'],
			['Netflix', '€12.99 · 1 payment']
		] as const) {
			await expect(top.getByText(name)).toBeInTheDocument();
			await expect(top.getByText(value)).toBeInTheDocument();
		}
		await expect(top.queryByText(/1 payments/)).not.toBeInTheDocument();

		const list = region(canvas, 'Accounts');
		await expect(canvas.getByRole('button', { name: 'Manage accounts' })).toBeInTheDocument();
		for (const heading of ['Household', 'Yours', 'Shared with you']) {
			await expect(list.getByRole('heading', { name: heading })).toBeInTheDocument();
		}
		const row = list.getByRole('link', { name: /Conta à Ordem/ });
		await expect(row).toHaveTextContent('Montepio');
		await expect(row).toHaveTextContent('€2,515.00');
		await expect(row).toHaveTextContent('Read yesterday at 18:04');
		await expect(list.getByRole('link', { name: /Joint account/ })).toHaveTextContent(
			'Read just now'
		);
		await expect(list.getByRole('link', { name: /Savings/ })).toHaveTextContent('€3,200.00');

		const biggest = region(canvas, 'Largest payments');
		for (const [name, account, date, amount] of [
			['Rent', 'Joint account · Monzo', '6 Sep', '−€820.00'],
			['Galp', 'Current account · Monzo', '12 Sep', '−€92.10'],
			['Pingo Doce', '', '14 Sep', '−€88.34'],
			['Amazon', '', '9 Sep', '−€64.99'],
			['EDP', '', '2 Sep', '−€58.20']
		] as const) {
			await expect(biggest.getByText(name)).toBeInTheDocument();
			await expect(biggest.getByText(date)).toBeInTheDocument();
			await expect(biggest.getByText(amount)).toBeInTheDocument();
			if (account) await expect(biggest.getAllByText(account).length).toBeGreaterThan(0);
		}

		const latest = region(canvas, 'Recent transactions');
		await expect(latest.getByRole('button', { name: 'See all' })).toBeInTheDocument();
		await expect(latest.getByText('Pingo Doce')).toBeInTheDocument();
		await expect(latest.getByText('17 Sep')).toBeInTheDocument();
		await expect(latest.getByText('−€42.18')).toBeInTheDocument();
		await expect(latest.getByText('Salary')).toBeInTheDocument();
		await expect(latest.getByText('+€2,450.00')).toBeInTheDocument();
		await expect(latest.getByText('Netflix')).toBeInTheDocument();

		// Never a connection action, and never the words of a budget.
		for (const name of [/Connect/, /Restore/, /Disconnect/, /Leave out/]) {
			await expect(canvas.queryByRole('button', { name })).not.toBeInTheDocument();
		}
		await expect(canvas.queryByText(/budget|income|expenses/i)).not.toBeInTheDocument();

		// One scope control, with All chosen and nothing said under it.
		const scope = canvas.getByRole('radiogroup', { name: 'Accounts counted' });
		for (const name of ['Household', 'Yours', 'All']) {
			await expect(within(scope).getByRole('radio', { name })).toBeInTheDocument();
		}
		await expect(within(scope).getByRole('radio', { name: 'All' })).toBeChecked();
		await expect(canvas.queryByText(/counts .* too/)).not.toBeInTheDocument();

		// Month by month.
		const by = months(canvas);
		await expect(
			canvas.getByRole('heading', { name: 'Month by month', level: 2 })
		).toBeInTheDocument();
		await expect(by.getByText('September 2025 to September 2026')).toBeInTheDocument();
		await expect(tile(canvas, 'Typical month')).toHaveTextContent('+€174.45');
		await expect(tile(canvas, 'Average month')).toHaveTextContent('−€110.24');
		await expect(
			by.getByText('In a typical month €174.45 more comes in than goes out.')
		).toBeInTheDocument();
		await expect(
			by.getByText(
				'From the last 6 full months. The typical month is the middle one, so one exceptional month barely moves it.'
			)
		).toBeInTheDocument();
		const view = canvas.getByRole('radiogroup', { name: 'Payments counted' });
		await expect(within(view).getByRole('radio', { name: 'All payments' })).toBeChecked();
		await expect(
			within(view).getByRole('radio', { name: 'Without unusual payments' })
		).toBeInTheDocument();
		await expect(by.getByRole('img', { name: allView.summary })).toBeInTheDocument();

		const rows = by.getAllByRole('button', {
			name: /^(January|February|March|April|May|June|July|August|September|October|November|December) 20\d\d/
		});
		await expect(rows).toHaveLength(13);
		const expected = [
			['September 2026', 'So far', '+€637.36'],
			['August 2026', '', '+€305.40'],
			['July 2026', '', '−€60.45'],
			['June 2026', '', '+€150.20'],
			['May 2026', '', '+€225.00'],
			['April 2026', '', '+€198.70'],
			['March 2026', '', '−€1,480.30', '2 unusual · +€379.70 without them'],
			['February 2026', '', '+€310.00'],
			['January 2026', '', '+€240.55'],
			['December 2025', '', '−€640.00', '1 unusual · −€160.00 without it'],
			['November 2025', '', '−€95.20'],
			['October 2025', '', '+€180.10'],
			['September 2025', '', '+€212.40']
		];
		for (const [i, [name, state, net, unusual]] of expected.entries()) {
			await expect(rows[i]).toHaveTextContent(name!);
			await expect(rows[i]).toHaveTextContent(net!);
			if (state) await expect(rows[i]).toHaveTextContent(state);
			if (unusual) await expect(rows[i]).toHaveTextContent(unusual);
			else await expect(rows[i]).not.toHaveTextContent(/unusual/);
		}

		// August is open: its figures and the merchants that rose.
		const august = monthRow(canvas, 'August 2026');
		await expect(august).toHaveAttribute('aria-pressed', 'true');
		const panel = panelOf(august);
		await expect(panel.getByText('€2,450.00')).toBeVisible();
		await expect(panel.getByText('€2,144.60')).toBeVisible();
		await expect(panel.getByText('More than usual')).toBeVisible();
		for (const words of [
			'€246.80 · €82.40 more than usual',
			'€188.20 · €64.10 more than usual',
			'€120.00 · not usually paid'
		]) {
			await expect(panel.getByText(words)).toBeInTheDocument();
		}
		await expect(monthRow(canvas, 'March 2026')).toHaveAttribute('aria-pressed', 'false');
		await expect(by.getByText(transferNote)).toBeInTheDocument();

		// A transfer among recent transactions is labelled, and the sentence
		// the screen used to hold is nowhere on the page.
		const seen = region(canvas, 'Recent transactions');
		await expect(seen.getByText('Transfer to savings')).toBeInTheDocument();
		await expect(seen.getByText('1 Sep')).toBeInTheDocument();
		await expect(seen.getByText('−€500.00')).toBeInTheDocument();
		await expect(seen.getByText('Between your accounts')).toBeInTheDocument();
		await expect(
			canvas.queryByText(/Money moved between your own accounts is counted\./)
		).not.toBeInTheDocument();

		// Recurring payments, soonest first, with no way to change them.
		const due = region(canvas, 'Recurring payments');
		await expect(
			canvas.getByRole('heading', { name: 'Recurring payments', level: 2 })
		).toBeInTheDocument();
		for (const [name, cadence, date, amount] of [
			['Spotify', 'Monthly · Current account · Monzo', 'Was expected 18 Sep', '−€9.99'],
			['Limpeza Casa', 'Weekly · Joint account · Monzo', 'Expected 24 Sep', '−€45.00'],
			['Rent', 'Monthly · Joint account · Monzo', 'Expected 6 Oct', '−€820.00'],
			['Netflix', 'Monthly · Current account · Monzo', 'Expected 14 Oct', '−€12.99'],
			['NOS', 'Monthly · Conta à Ordem · Montepio', 'Expected 17 Oct', '−€39.99'],
			['Fidelidade', 'Yearly · Joint account · Monzo', 'Expected 3 Mar 2027', '−€386.00']
		] as const) {
			const row = due.getByText(name).closest('.ledger-row') as HTMLElement;
			await expect(row).toHaveTextContent(cadence);
			await expect(row).toHaveTextContent(date);
			await expect(row).toHaveTextContent(amount);
		}
		await expect(
			canvas.queryByRole('button', { name: /Cancel|Pause|Hide|Dismiss/ })
		).not.toBeInTheDocument();
		await expect(due.queryByText(/total/i)).not.toBeInTheDocument();
		for (const banned of [
			/outlier/i,
			/anomal/i,
			/subscription/i,
			/\bearn/i,
			/probably/i,
			/windfall/i
		]) {
			await expect(canvas.queryByText(banned)).not.toBeInTheDocument();
		}
	}}
/>

<Story
	name="ADayPointedAt"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				chart: { ...eur.chart, marked: balancePoints.findIndex((p) => p.date === '3 Aug') }
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const chart = region(canvas, 'Balance · EUR');
		for (const words of [
			'3 Aug',
			'€9,870.12',
			'€805.79 less than 2 Aug',
			'Leroy Merlin',
			'−€640.00',
			'Unusual',
			'Galp',
			'−€92.10',
			'and 3 smaller'
		]) {
			await expect(chart.getAllByText(words).length).toBeGreaterThan(0);
		}
		await expect(canvas.queryByText('3 Aug · €9,870.12')).not.toBeInTheDocument();
		await expect(liveOf(canvasElement)).toHaveTextContent(
			'3 Aug. €9,870.12. €805.79 less than 2 Aug. Leroy Merlin −€640.00, Unusual. Galp −€92.10. And 3 smaller.'
		);
	}}
/>

<!-- One sentence covers both reasons an account is left out, because the
     member's question is the same: why the chart and the figures disagree. -->
<Story
	name="ChartDoesNotCoverEveryAccount"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				household: '€9,244.55',
				chart: {
					...eur.chart,
					coverage:
						'Not every account is in this chart. An account is left out when its bank shares balances only, or when its history is under 30 days.'
				}
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(tile(canvas, 'Household money')).toHaveTextContent('€9,244.55');
		await expect(
			canvas.getByText(
				'Not every account is in this chart. An account is left out when its bank shares balances only, or when its history is under 30 days.'
			)
		).toBeInTheDocument();
	}}
/>

<Story
	name="LedgerStartsThisMonth"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				month: {
					moneyIn: { value: '€2,450.00' },
					moneyOut: { value: '€1,812.64' },
					net: { value: '+€637.36' },
					note: 'Counted from 19 Sep.',
					href: '/transactions?month=2026-09'
				},
				chart: { shortHistory: true },
				merchantsSpan: '19 to 20 Sep',
				history: undefined,
				recurring: undefined
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const [label, value] of [
			['Money in', '€2,450.00'],
			['Money out', '€1,812.64'],
			['Net', '+€637.36']
		] as const) {
			await expect(canvas.getByRole('link', { name: new RegExp(label) })).toHaveTextContent(value);
		}
		await expect(canvas.queryByText(/\bmore\b|\bless\b|\bthan\b/)).not.toBeInTheDocument();
		await expect(canvas.getByText(`Counted from 19 Sep. ${transferNote}`)).toBeInTheDocument();

		const chart = region(canvas, 'Balance · EUR');
		await expect(chart.getByText('Balance · EUR')).toBeInTheDocument();
		await expect(
			chart.getByText('A balance chart appears once there is a week of history.')
		).toBeInTheDocument();
		await expect(chart.queryByRole('img')).not.toBeInTheDocument();
		await expect(region(canvas, 'Top merchants').getByText('19 to 20 Sep')).toBeInTheDocument();
		for (const absent of [
			'Month by month',
			'Recurring payments',
			'Typical month',
			'Payments counted'
		]) {
			await expect(canvas.queryByText(absent)).not.toBeInTheDocument();
		}
	}}
/>

<Story
	name="NoTransactionHistory"
	tags={['kind-state']}
	args={{
		currencies: [{ currency: 'EUR', household: '€6,698.00', own: '€4,995.55' }],
		recentTransactions: [],
		scope: undefined
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: 'Overview', level: 1 })).toBeInTheDocument();
		await expect(tile(canvas, 'Household money')).toHaveTextContent('€6,698.00');
		await expect(tile(canvas, 'Your money')).toHaveTextContent('€4,995.55');
		await expect(canvas.getByRole('heading', { name: 'Accounts' })).toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Manage accounts' })).toBeInTheDocument();
		for (const heading of ['Household', 'Yours', 'Shared with you']) {
			await expect(canvas.getByRole('heading', { name: heading })).toBeInTheDocument();
		}
		for (const absent of [
			'Money in',
			'Balance · EUR',
			'Top merchants',
			'Largest payments',
			'Recent transactions'
		]) {
			await expect(canvas.queryByText(absent)).not.toBeInTheDocument();
		}
		await expect(canvas.queryByText(/chart/i)).not.toBeInTheDocument();
		for (const absent of [
			'Month by month',
			'Recurring payments',
			'Typical month',
			'Payments counted',
			'Accounts counted'
		]) {
			await expect(canvas.queryByText(absent)).not.toBeInTheDocument();
		}
	}}
/>

<Story
	name="NothingSpentYet"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				month: { ...month, moneyOut: { value: '€0.00' } },
				merchants: [],
				largestPayments: []
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('link', { name: /Money out/ })).toHaveTextContent('€0.00');
		await expect(canvas.queryByText('Top merchants')).not.toBeInTheDocument();
		await expect(canvas.queryByText('Largest payments')).not.toBeInTheDocument();
	}}
/>

<!-- Nothing stands in for the missing two. -->
<Story
	name="ThreeMerchants"
	tags={['kind-state']}
	args={{
		currencies: [{ ...eur, merchants: merchants.slice(0, 3), largestPayments: largest.slice(0, 3) }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const top = region(canvas, 'Top merchants');
		for (const name of ['Pingo Doce', 'Galp', 'Amazon']) {
			await expect(top.getByText(name)).toBeInTheDocument();
		}
		await expect(top.queryByText('Via Verde')).not.toBeInTheDocument();
		await expect(top.queryByText('Netflix')).not.toBeInTheDocument();
		await expect(region(canvas, 'Largest payments').getAllByText(/^−€/)).toHaveLength(3);
	}}
/>

<!-- A member who owns nothing is not told there is anything they cannot see:
     no section stands where a hidden one would. -->
<Story
	name="OwnsNoAccount"
	tags={['kind-state']}
	args={{
		currencies: [{ currency: 'EUR', household: '€4,120.00' }],
		accounts: {
			household: [{ ...joint, balance: '€4,120.00' }],
			shared: [savings]
		},
		ownsNothing: true,
		recentTransactions: [],
		scope: undefined
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: 'Overview', level: 1 })).toBeInTheDocument();
		await expect(tile(canvas, 'Household money')).toHaveTextContent('€4,120.00');
		const list = region(canvas, 'Accounts');
		await expect(list.getByRole('heading', { name: 'Household' })).toBeInTheDocument();
		await expect(list.getByRole('link', { name: /Joint account/ })).toHaveTextContent('Monzo');
		await expect(list.getByRole('link', { name: /Joint account/ })).toHaveTextContent(
			'Read just now'
		);
		await expect(list.getByRole('heading', { name: 'Shared with you' })).toBeInTheDocument();
		const shared = list.getByRole('link', { name: /Savings/ });
		await expect(shared).toHaveTextContent('Montepio');
		await expect(shared).toHaveTextContent('€3,200.00');
		await expect(shared).not.toHaveTextContent('••••');
		for (const absent of [
			'Your money',
			'Yours',
			'Money in',
			'Balance · EUR',
			'Top merchants',
			'Largest payments',
			'Recent transactions'
		]) {
			await expect(canvas.queryByText(absent)).not.toBeInTheDocument();
		}
		await expect(canvas.queryByRole('button', { name: 'Manage accounts' })).not.toBeInTheDocument();
		for (const absent of [
			'Month by month',
			'Recurring payments',
			'Typical month',
			'Payments counted',
			'Accounts counted'
		]) {
			await expect(canvas.queryByText(absent)).not.toBeInTheDocument();
		}
	}}
/>

<Story
	name="HouseholdOfOne"
	tags={['kind-state']}
	args={{
		currencies: [{ ...eur, household: undefined, own: '€11,693.55' }],
		accounts: { own: [joint, current, conta] },
		scope: undefined
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByText('Household money')).not.toBeInTheDocument();
		await expect(
			canvas.queryByRole('heading', { name: 'Shared with you' })
		).not.toBeInTheDocument();
		await expect(tile(canvas, 'Your money')).toHaveTextContent('€11,693.55');
		const list = region(canvas, 'Accounts');
		await expect(list.queryByRole('heading', { name: 'Household' })).not.toBeInTheDocument();
		await expect(list.getByRole('heading', { name: 'Yours' })).toBeInTheDocument();
		for (const name of [/Joint account/, /Current account/, /Conta à Ordem/]) {
			await expect(list.getByRole('link', { name })).toBeInTheDocument();
		}
		await expect(canvas.getByText('Month by month')).toBeInTheDocument();
		await expect(canvas.getByText('Recurring payments')).toBeInTheDocument();
		await expect(canvas.queryByText('Accounts counted')).not.toBeInTheDocument();
		await expect(
			canvas.queryByRole('radiogroup', { name: 'Accounts counted' })
		).not.toBeInTheDocument();
	}}
/>

<!-- wimm holds no rates, so currencies are never added and the screen says so. -->
<Story
	name="TwoCurrencies"
	tags={['kind-state']}
	args={{
		currencies: [eur, gbp],
		accounts: {
			...accounts,
			own: [
				...accounts.own,
				{
					id: 'a5',
					name: 'Sterling account',
					bank: 'Monzo',
					balance: '£640.20',
					readAt: 'just now'
				}
			]
		}
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const code of ['EUR', 'GBP']) {
			for (const label of ['Household money', 'Your money', 'Money in', 'Money out', 'Net']) {
				await expect(canvas.getByText(`${label} · ${code}`)).toBeInTheDocument();
			}
			for (const title of ['Balance', 'Top merchants', 'Largest payments']) {
				await expect(
					canvas.getByRole('region', { name: `${title} · ${code}` })
				).toBeInTheDocument();
			}
		}
		await expect(
			canvas.getByText('Shown per currency. wimm does not convert between them.')
		).toBeInTheDocument();
		await expect(canvas.getAllByRole('region', { name: 'Accounts' })).toHaveLength(1);
		await expect(canvas.getAllByRole('region', { name: 'Recent transactions' })).toHaveLength(1);

		for (const code of ['EUR', 'GBP']) {
			await expect(
				canvas.getByRole('region', { name: `Month by month · ${code}` })
			).toBeInTheDocument();
			await expect(
				canvas.getByRole('region', { name: `Recurring payments · ${code}` })
			).toBeInTheDocument();
		}
		await expect(canvas.getAllByRole('radiogroup', { name: 'Accounts counted' })).toHaveLength(1);
		await expect(
			region(canvas, 'Recurring payments · GBP').getByText('Council tax')
		).toBeInTheDocument();
	}}
/>

<!-- The service decides a currency is empty; the list still shows the
     account, so nothing the member may see is hidden from them. -->
<Story
	name="CurrencyHoldingNothing"
	tags={['kind-state']}
	args={{
		accounts: {
			...accounts,
			own: [
				...accounts.own,
				{ id: 'a6', name: 'Dollar account', bank: 'Monzo', balance: '$0.00', readAt: 'just now' }
			]
		}
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const dollar = region(canvas, 'Accounts').getByRole('link', { name: /Dollar account/ });
		await expect(dollar).toHaveTextContent('Monzo');
		await expect(dollar).toHaveTextContent('$0.00');
		await expect(dollar).toHaveTextContent('Read just now');
		await expect(canvas.getAllByText('$0.00')).toHaveLength(1);
		await expect(canvas.queryByText('Balance · USD')).not.toBeInTheDocument();
	}}
/>

<Story
	name="ReachesTheFullLedger"
	tags={['kind-behaviour']}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'See all' }));
		await expect(args.onseeall).toHaveBeenCalledOnce();
		await userEvent.click(canvas.getByRole('button', { name: 'Manage accounts' }));
		await expect(args.ongotoaccounts).toHaveBeenCalledOnce();
		await expect(canvas.getByRole('link', { name: /Joint account/ })).toHaveAttribute(
			'href',
			'/accounts'
		);
		await expect(canvas.getByRole('link', { name: /Money out/ })).toHaveAttribute(
			'href',
			'/transactions?month=2026-09'
		);
	}}
/>

<Story
	name="BeforeAnyBankConnected"
	tags={['kind-state']}
	args={{ hasAccounts: false, currencies: [], accounts: {}, recentTransactions: [] }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByText(/€0/)).not.toBeInTheDocument();
		await expect(canvas.getByText('Connect a bank to see where you stand')).toBeInTheDocument();
		await expect(
			canvas.getByText(/Once a bank is connected, your total, your recent activity/)
		).toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Go to Accounts' }));
		await expect(args.ongotoaccounts).toHaveBeenCalledOnce();
	}}
/>

<!-- A member who may see no account gets the same explanation as no bank at
     all, not a total of zero. -->
<Story
	name="MemberMayNothing"
	tags={['kind-behaviour']}
	args={{ hasAccounts: false, currencies: [], accounts: {}, recentTransactions: [] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Connect a bank to see where you stand')).toBeInTheDocument();
		await expect(canvas.queryByText(/[€£$]\d/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="AMonthWithUnusualPaymentsOpened"
	tags={['kind-state']}
	args={{ currencies: [{ ...eur, history: { ...history, open: 'March 2026' } }] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const march = monthRow(canvas, 'March 2026');
		await expect(march).toHaveAttribute('aria-pressed', 'true');
		await expect(monthRow(canvas, 'August 2026')).toHaveAttribute('aria-pressed', 'false');
		const panel = panelOf(march);
		for (const words of [
			'€2,450.00',
			'€3,930.30',
			'−€1,480.30',
			'Without unusual payments',
			'+€379.70',
			'More than usual',
			'Unusual payments',
			'€1,650.00 · not usually paid',
			'€318.40 · €154.00 more than usual'
		]) {
			await expect(panel.getByText(words)).toBeVisible();
		}
		const [garage, fuel] = panel.getAllByRole('link');
		for (const words of [
			'12 Mar',
			'Auto Reparadora',
			'−€1,650.00',
			'Unusual',
			'first payment to Auto Reparadora'
		]) {
			await expect(garage).toHaveTextContent(words);
		}
		await expect(garage).not.toHaveTextContent('usually about');
		for (const words of ['27 Mar', 'Galp', '−€210.00', 'Unusual', 'usually about €60']) {
			await expect(fuel).toHaveTextContent(words);
		}
		await expect(panel.getAllByRole('link')).toHaveLength(2);
		await expect(garage).toHaveAttribute('href', '/transactions?page=2026-03-12.a1');
	}}
/>

<Story
	name="AMonthWithUnusualIncomeOpened"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				history: {
					...history,
					open: 'May 2026',
					all: { ...allView, average: '+€1,523.76' },
					usual: { ...usualView, average: '+€199.76' },
					months: historyMonths.map((m) =>
						m.label === 'May 2026'
							? {
									...m,
									net: '+€10,029.00',
									value: 10029,
									netUsual: '+€225.00',
									valueUsual: 225,
									unusualLine: '1 unusual · +€225.00 without it',
									figures: [
										{ label: 'Money in', value: '€12,254.00' },
										{ label: 'Money out', value: '€2,225.00' },
										{ label: 'Net', value: '+€10,029.00' },
										{ label: 'Without unusual payments', value: '+€225.00' }
									],
									payments: [bonusPayment]
								}
							: m
					)
				}
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(tile(canvas, 'Average month')).toHaveTextContent('+€1,523.76');
		await expect(tile(canvas, 'Typical month')).toHaveTextContent('+€174.45');
		const may = monthRow(canvas, 'May 2026');
		await expect(may).toHaveTextContent('+€10,029.00');
		await expect(may).toHaveTextContent('1 unusual · +€225.00 without it');
		const panel = panelOf(may);
		for (const words of [
			'€12,254.00',
			'€2,225.00',
			'+€10,029.00',
			'Without unusual payments',
			'+€225.00',
			'Unusual payments'
		]) {
			await expect(panel.getByText(words)).toBeVisible();
		}
		const bonus = panel.getByRole('link');
		for (const words of [
			'22 May',
			'Employer Lda',
			'+€9,804.00',
			'Unusual income',
			'usually about €2,450'
		]) {
			await expect(bonus).toHaveTextContent(words);
		}
		await userEvent.click(canvas.getByRole('radio', { name: 'Without unusual payments' }));
		await expect(tile(canvas, 'Average month')).toHaveTextContent('+€199.76');
	}}
/>

<Story
	name="ALikelyYearlyPayment"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				recurring: [
					...recurring.slice(0, -1),
					{ ...fidelidade, cadence: 'Likely yearly · Joint account · Monzo' }
				]
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const row = region(canvas, 'Recurring payments')
			.getByText('Fidelidade')
			.closest('.ledger-row') as HTMLElement;
		await expect(row).toHaveTextContent('Likely yearly · Joint account · Monzo');
		await expect(row).toHaveTextContent('Expected 3 Mar 2027');
		await expect(row).toHaveTextContent('−€386.00');
		await expect(row).not.toHaveTextContent(/(^|\s)Yearly ·/);
	}}
/>

<Story
	name="AMonthWithNoneOpened"
	tags={['kind-state']}
	args={{ currencies: [{ ...eur, history: { ...history, open: 'June 2026' } }] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const june = monthRow(canvas, 'June 2026');
		await expect(june).toHaveAttribute('aria-pressed', 'true');
		await expect(june).not.toHaveTextContent(/unusual/i);
		const panel = panelOf(june);
		for (const words of [
			'€2,450.00',
			'€2,299.80',
			'+€150.20',
			'Nothing took more than usual this month.'
		]) {
			await expect(panel.getByText(words)).toBeVisible();
		}
		for (const absent of [/Unusual payments/, /unusual/i, /Without unusual payments/]) {
			await expect(panel.queryByText(absent)).not.toBeInTheDocument();
		}
		// A month that left nothing out says nothing about transfers.
		await expect(panel.queryByText(/left out/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="TheMonthSoFarOpened"
	tags={['kind-state']}
	args={{ currencies: [{ ...eur, history: { ...history, open: 'September 2026' } }] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const sep = monthRow(canvas, 'September 2026');
		await expect(sep).toHaveAttribute('aria-pressed', 'true');
		await expect(sep).toHaveTextContent('So far');
		const panel = panelOf(sep);
		for (const words of [
			'1 to 20 Sep',
			'€2,450.00',
			'€1,812.64',
			'+€637.36',
			'A month under way is not set against whole months.'
		]) {
			await expect(panel.getByText(words)).toBeVisible();
		}
		await expect(panel.queryByText('More than usual')).not.toBeInTheDocument();
	}}
/>

<Story
	name="WithoutUnusualPayments"
	tags={['kind-state']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('radio', { name: 'Without unusual payments' }));
		await expect(canvas.getByRole('radio', { name: 'Without unusual payments' })).toBeChecked();
		await expect(tile(canvas, 'Typical month')).toHaveTextContent('+€211.85');
		await expect(tile(canvas, 'Average month')).toHaveTextContent('+€199.76');
		const by = months(canvas);
		await expect(
			by.getByText('In a typical month €211.85 more comes in than goes out.')
		).toBeInTheDocument();
		await expect(
			by.getByText('3 unusual payments are set aside. They are still listed in their months.')
		).toBeInTheDocument();
		await expect(monthRow(canvas, 'March 2026')).toHaveTextContent(
			'2 unusual · +€379.70 without them'
		);
		await expect(by.getByRole('img', { name: usualView.summary })).toBeInTheDocument();
		await expect(
			usualView.summary.endsWith('Lowest −€160.00 in December 2025. A typical month is +€211.85.')
		).toBe(true);
	}}
/>

<Story
	name="MoreGoesOutThanComesIn"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				history: {
					...history,
					usual: undefined,
					all: {
						...allView,
						typical: '−€212.40',
						typicalValue: -212.4,
						average: '−€260.15',
						sentence:
							'More goes out than comes in. In a typical month €212.40 more goes out than comes in.'
					}
				}
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(tile(canvas, 'Typical month')).toHaveTextContent('−€212.40');
		await expect(tile(canvas, 'Average month')).toHaveTextContent('−€260.15');
		await expect(
			canvas.getByText(
				'More goes out than comes in. In a typical month €212.40 more goes out than comes in.'
			)
		).toBeInTheDocument();
	}}
/>

<Story
	name="TwoFullMonths"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				recurring: undefined,
				history: {
					span: 'July 2026 to September 2026',
					open: 'August 2026',
					months: historyMonths
						.slice(0, 3)
						.map((m) => ({ ...m, risers: undefined, open: undefined })),
					waiting: 'A typical month and an average month appear once three full months are held.',
					all: {
						summary:
							'Net by month from July 2026 to September 2026. 1 of 2 full months ended with more in than out. Highest +€305.40 in August 2026. Lowest −€60.45 in July 2026.',
						high: '+€637',
						low: '−€60'
					}
				}
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Month by month')).toBeInTheDocument();
		await expect(canvas.getByText('July 2026 to September 2026')).toBeInTheDocument();
		await expect(months(canvas).getAllByRole('button')).toHaveLength(3);
		await expect(
			canvas.getByText(
				'A typical month and an average month appear once three full months are held.'
			)
		).toBeInTheDocument();
		for (const absent of [
			'Typical month',
			'Average month',
			'Payments counted',
			'More than usual',
			'Recurring payments'
		]) {
			await expect(canvas.queryByText(absent)).not.toBeInTheDocument();
		}
	}}
/>

<Story
	name="AYoungAccountBesideOlderOnes"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				history: { ...history, late: ['Months before 19 Sep do not include CLASSIC CEMG.'] }
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByText('Months before 19 Sep do not include CLASSIC CEMG.')
		).toBeInTheDocument();
		await expect(canvas.getByText('Typical month')).toBeInTheDocument();
		await expect(canvas.queryByText(/No full month is held/)).toBeNull();
	}}
/>

<Story
	name="LedgerBeginsPartWayThrough"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				history: {
					...history,
					span: 'April 2026 to September 2026',
					months: historyMonths
						.slice(0, 6)
						.map((m) =>
							m.label === 'April 2026' ? { ...m, state: 'Held from 12 Apr', partial: true } : m
						),
					basis:
						'From 4 full months. The typical month is the middle one, so one exceptional month barely moves it.'
				}
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('April 2026 to September 2026')).toBeInTheDocument();
		await expect(monthRow(canvas, 'April 2026')).toHaveTextContent('Held from 12 Apr');
		await expect(months(canvas).getAllByRole('button')).toHaveLength(6);
		await expect(canvas.getByText(/^From 4 full months\./)).toBeInTheDocument();
	}}
/>

<Story
	name="HouseholdScope"
	tags={['kind-state']}
	args={{
		scope: { choices: ['household', 'own', 'all'], value: 'household' },
		currencies: [
			{ ...eur, recurring: [cleaning, rent, fidelidade], transferNote: narrowTransferNote }
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('radio', { name: 'Household' })).toBeChecked();
		const due = region(canvas, 'Recurring payments');
		for (const name of ['Limpeza Casa', 'Rent', 'Fidelidade']) {
			await expect(due.getByText(name)).toBeInTheDocument();
		}
		for (const name of ['Spotify', 'Netflix', 'NOS']) {
			await expect(due.queryByText(name)).not.toBeInTheDocument();
		}
		await expect(tile(canvas, 'Household money')).toHaveTextContent('€6,698.00');
		await expect(tile(canvas, 'Your money')).toHaveTextContent('€4,995.55');
		await expect(canvas.queryByText(/Household counts/)).not.toBeInTheDocument();
		// A narrower scope counts what crosses it, and says so.
		await expect(canvas.getAllByText(narrowTransferNote)[0]).toBeInTheDocument();
		await expect(canvas.queryByText(transferNote)).not.toBeInTheDocument();
	}}
/>

<Story
	name="HouseholdScopeWithNoHistory"
	tags={['kind-state']}
	args={{
		scope: { choices: ['household', 'own', 'all'], value: 'household' },
		currencies: [
			{
				...eur,
				history: undefined,
				historyEmpty: 'No full month is held for these accounts yet.',
				recurring: undefined,
				recurringEmpty: 'No recurring payments in these accounts.'
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('radio', { name: 'Household' })).toBeChecked();
		const months = region(canvas, 'Months');
		await expect(months.getByText('Month by month')).toBeInTheDocument();
		await expect(
			months.getByText('No full month is held for these accounts yet.')
		).toBeInTheDocument();
		const due = region(canvas, 'Recurring payments');
		await expect(due.getByText('No recurring payments in these accounts.')).toBeInTheDocument();
		await expect(canvas.queryByText('Typical month')).not.toBeInTheDocument();
		await expect(canvas.queryByText('Average month')).not.toBeInTheDocument();
		await expect(canvas.queryByRole('radiogroup', { name: 'Payments counted' })).not.toBeInTheDocument();
	}}
/>

<Story
	name="HouseholdScopeWithNothingToSetAside"
	tags={['kind-state']}
	args={{
		scope: { choices: ['household', 'own', 'all'], value: 'household' },
		currencies: [
			{
				...eur,
				history: {
					...history,
					usual: history.all,
					setAside: 'No unusual payments to set aside.'
				}
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('radio', { name: 'Without unusual payments' }));
		await expect(canvas.getByText('No unusual payments to set aside.')).toBeInTheDocument();
		await expect(canvas.queryByText(/unusual payments? (is|are) set aside/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="HouseholdScopeIsNarrower"
	tags={['kind-state']}
	args={{
		scope: {
			choices: ['household', 'own', 'all'],
			value: 'household',
			note: 'Household counts Joint account. Joint savings is household money too, but it is not yours, so its transactions are not counted.'
		},
		currencies: [{ ...eur, recurring: [cleaning, rent, fidelidade] }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('radio', { name: 'Household' })).toBeChecked();
		await expect(
			canvas.getByText(
				'Household counts Joint account. Joint savings is household money too, but it is not yours, so its transactions are not counted.'
			)
		).toBeInTheDocument();
	}}
/>

<Story
	name="YoursScope"
	tags={['kind-state']}
	args={{
		scope: { choices: ['household', 'own', 'all'], value: 'own' },
		currencies: [
			{
				...eur,
				// Under Yours the standing transfer to the joint account is
				// money that leaves every month, so it recurs here.
				recurring: [spotify, netflix, nos, jointTransfer],
				// It is also counted, and large, so it is among the largest
				// payments — and still labelled.
				largestPayments: [jointPayment, ...largest.slice(1)],
				transferNote: narrowTransferNote,
				history: { ...history, open: 'June 2026' }
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('radio', { name: 'Yours' })).toBeChecked();
		const due = region(canvas, 'Recurring payments');
		for (const name of ['Spotify', 'Netflix', 'NOS', 'Joint account']) {
			await expect(due.getAllByText(name)[0]).toBeInTheDocument();
		}
		for (const name of ['Limpeza Casa', 'Rent', 'Fidelidade']) {
			await expect(due.queryByText(name)).not.toBeInTheDocument();
		}
		const row = due.getAllByText('Joint account')[0]!.closest('.ledger-row') as HTMLElement;
		await expect(row).toHaveTextContent('Monthly · Current account · Monzo');
		await expect(row).toHaveTextContent('Expected 1 Oct');
		await expect(row).toHaveTextContent('−€800.00');

		// Counted here, and labelled: the narrower sentence says exactly that.
		const biggest = region(canvas, 'Largest payments');
		await expect(biggest.getByText('Joint account')).toBeInTheDocument();
		await expect(biggest.getByText('1 Sep')).toBeInTheDocument();
		await expect(biggest.getByText('−€800.00')).toBeInTheDocument();
		await expect(biggest.getByText('Between your accounts')).toBeInTheDocument();

		await expect(canvas.getAllByText(narrowTransferNote)[0]).toBeInTheDocument();
		await expect(canvas.queryByText(transferNote)).not.toBeInTheDocument();
		// Under Yours the pair crosses the scope, so the month left none out.
		await expect(panelOf(monthRow(canvas, 'June 2026')).queryByText(/left out/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="ChoosingAScope"
	tags={['kind-behaviour']}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('radio', { name: 'All' }));
		await expect(args.onscope).not.toHaveBeenCalled();
		await userEvent.click(canvas.getByRole('radio', { name: 'Yours' }));
		await expect(args.onscope).toHaveBeenCalledOnce();
		await expect(args.onscope).toHaveBeenCalledWith('own');
	}}
/>

<Story
	name="ReadingMonthsByPointing"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const march = monthRow(canvas, 'March 2026');
		await userEvent.click(march);
		await expect(march).toHaveAttribute('aria-pressed', 'true');
		await expect(monthRow(canvas, 'August 2026')).toHaveAttribute('aria-pressed', 'false');
		// One detail panel, and it follows the selection.
		await expect(months(canvas).getByRole('region', { name: 'March 2026' })).toBeVisible();
		await expect(months(canvas).queryByRole('region', { name: 'August 2026' })).toBeNull();
		await expect(march.getAttribute('aria-controls')).toBe(
			months(canvas).getByRole('region', { name: 'March 2026' }).id
		);
		const plot = months(canvas).getByRole('img', { name: allView.summary });
		plot.focus();
		await userEvent.keyboard('{Home}');
		await expect(monthRow(canvas, 'September 2025')).toHaveAttribute('aria-pressed', 'true');
		await expect(march).toHaveAttribute('aria-pressed', 'false');
		await expect(
			months(canvas).getByRole('region', { name: 'September 2025' })
		).toBeVisible();
	}}
/>

<!-- June left two transfers out, so the month says how many and what they came
     to, under the figures it changed. -->
<Story
	name="AMonthWithTransfersLeftOut"
	tags={['kind-state']}
	args={{
		currencies: [
			{
				...eur,
				history: {
					...history,
					open: 'June 2026',
					months: history.months.map((m) =>
						m.label === 'June 2026'
							? { ...m, leftOut: '2 transfers between your accounts left out · €1,400.00' }
							: m
					)
				}
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const june = monthRow(canvas, 'June 2026');
		await expect(june).toHaveAttribute('aria-pressed', 'true');
		const panel = panelOf(june);
		for (const words of ['€2,450.00', '€2,299.80', '+€150.20']) {
			await expect(panel.getByText(words)).toBeVisible();
		}
		await expect(
			panel.getByText('2 transfers between your accounts left out · €1,400.00')
		).toBeVisible();
	}}
/>

<!-- The day the money left. The line moved, so the chart keeps the row and
     says what kind of movement it was. -->
<Story
	name="ADayATransferLeft"
	tags={['kind-state']}
	args={{ currencies: [{ ...eur, chart: { ...eur.chart, marked: balancePoints.findIndex((p) => p.date === '1 Sep') } }] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const words of [
			'1 Sep',
			'€10,912.40',
			'€500.00 less than 31 Aug',
			'Transfer to savings',
			'−€500.00',
			'Between your accounts'
		]) {
			await expect(canvas.getAllByText(words)[0]).toBeInTheDocument();
		}
		await expect(liveOf(canvasElement)).toHaveTextContent(
			'1 Sep. €10,912.40. €500.00 less than 31 Aug. Transfer to savings −€500.00, Between your accounts.'
		);
	}}
/>
