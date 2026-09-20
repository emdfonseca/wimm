<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import { balanceChart, balancePoints, balanceSummary } from '../molecules/BalanceChart.fixture';
	import Overview, {
		type AccountEntry,
		type CurrencySection,
		type RecentTransaction
	} from './Overview.svelte';

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
		largestPayments: largest
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
		merchants: [
			{ name: 'Tesco', value: '£210.10 · 5 payments', proportion: 1 },
			{ name: 'Boots', value: '£38.20 · 2 payments', proportion: 0.18 }
		],
		merchantsSpan: '1 to 20 Sep',
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

	const region = (canvas: ReturnType<typeof within>, name: string) =>
		within(canvas.getByRole('region', { name }));

	const tile = (canvas: ReturnType<typeof within>, label: string) =>
		canvas.getByText(label).closest('.tile') as HTMLElement;

	const transferNote = 'Money moved between your own accounts is counted.';

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
		await expect(canvas.getByText('3 Aug · €9,870.12')).toBeInTheDocument();
		await expect(canvas.getByText('€150.00 more')).toBeInTheDocument();
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
				merchantsSpan: '19 to 20 Sep'
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
	}}
/>

<Story
	name="NoTransactionHistory"
	tags={['kind-state']}
	args={{
		currencies: [{ currency: 'EUR', household: '€6,698.00', own: '€4,995.55' }],
		recentTransactions: []
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
		recentTransactions: []
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
	}}
/>

<Story
	name="HouseholdOfOne"
	tags={['kind-state']}
	args={{
		currencies: [{ ...eur, household: undefined, own: '€11,693.55' }],
		accounts: { own: [joint, current, conta] }
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
