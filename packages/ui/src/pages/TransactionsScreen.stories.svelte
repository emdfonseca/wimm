<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import TransactionsScreen, { type LedgerDay } from './TransactionsScreen.svelte';

	/** Pinned, so "Today" reads `Today, 17 September` whenever the story runs. */
	const today = '2026-09-17';

	const days: LedgerDay[] = [
		{
			date: '2026-09-17',
			entries: [
				{
					id: '1',
					description: 'Pingo Doce',
					banksLine: 'COMPRA PINGO DOCE LISBOA 230002268342127',
					account: 'Current account · Monzo',
					amount: '−€42.18',
					date: '17 Sep',
					negative: true,
					initials: 'PD'
				},
				{
					id: '2',
					description: 'Transfer to Ana Reis',
					banksLine: 'Transfer to Ana Reis',
					account: 'Joint savings · Montepio',
					amount: '−€60.00',
					date: '17 Sep',
					negative: true,
					unsettled: true,
					initials: 'AR'
				}
			]
		},
		{
			date: '2026-09-16',
			entries: [
				{
					id: '3',
					description: 'Galp',
					banksLine: 'COMPRA GALP A5 OEIRAS 230002270158934',
					account: 'Current account · Monzo',
					amount: '−€71.40',
					date: '16 Sep',
					negative: true,
					unusual: true,
					initials: 'GA'
				},
				{
					id: '4',
					description: 'Salary',
					banksLine: 'Salary',
					account: 'Current account · Monzo',
					amount: '+€2,180.00',
					date: '16 Sep',
					initials: 'SA'
				},
				{
					id: '5',
					description: 'NOS',
					banksLine: 'DD NOS COMUNICACOES SA 000000234058260',
					account: 'Joint savings · Montepio',
					amount: '−€39.99',
					date: '16 Sep',
					negative: true,
					initials: 'NO'
				}
			]
		}
	];

	const olderDays: LedgerDay[] = [
		{
			date: '2026-08-04',
			entries: [
				{
					id: '6',
					description: 'Continente',
					account: 'Current account',
					amount: '−€118.72',
					date: '4 Aug',
					negative: true,
					initials: 'CO'
				},
				{
					id: '7',
					description: 'Águas de Lisboa',
					account: 'Joint savings',
					amount: '−€28.40',
					date: '4 Aug',
					negative: true,
					initials: 'AG'
				}
			]
		},
		{
			date: '2026-07-31',
			entries: [
				{
					id: '8',
					description: 'Salary',
					account: 'Current account',
					amount: '+€2,180.00',
					date: '31 Jul',
					initials: 'SA'
				}
			]
		}
	];

	const freshness = 'Updated at 09:14. Reaching back to 4 June 2026.';
	const span = '17 September to 15 September 2026';

	const pages = [
		{ key: 'p1', label: '17 September to 15 September 2026', href: '/transactions?page=p1' },
		{ key: 'p2', label: '12 August to 3 August 2026', href: '/transactions?page=p2' },
		{ key: 'p3', label: '2 August to 4 June 2026', href: '/transactions?page=p3' }
	];

	const { Story } = defineMeta({
		title: 'Pages/TransactionsScreen',
		component: TransactionsScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen', shell: '/transactions' },
		args: {
			days,
			count: 382,
			freshness,
			span,
			today
		}
	});
</script>

<!-- J07.A / 01 -->
<Story
	name="AsItOpens"
	tags={['kind-state']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByRole('heading', { name: 'Transactions', level: 1 })
		).toBeInTheDocument();
		await expect(canvas.getByText('Every account you own, newest first.')).toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Refresh' })).toBeInTheDocument();
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();

		for (const heading of ['Description', 'Account', 'Amount']) {
			await expect(canvas.getByText(heading)).toBeInTheDocument();
		}
		await expect(canvas.queryByText('Date')).not.toBeInTheDocument();

		// The date lives in the day header, not in every row.
		await expect(canvas.getByText('Today, 17 September')).toBeInTheDocument();
		await expect(canvas.getByText('Yesterday, 16 September')).toBeInTheDocument();
		await expect(canvas.queryByText('17 Sep')).not.toBeInTheDocument();

		// The bank's line sits under a name that differs from it.
		for (const [name, line, amount] of [
			['Pingo Doce', 'COMPRA PINGO DOCE LISBOA 230002268342127', '−€42.18'],
			['Galp', 'COMPRA GALP A5 OEIRAS 230002270158934', '−€71.40'],
			['NOS', 'DD NOS COMUNICACOES SA 000000234058260', '−€39.99']
		] as const) {
			const row = canvas.getByText(line).closest('.ledger-row') as HTMLElement;
			await expect(row).toHaveTextContent(name);
			await expect(row).toHaveTextContent(amount);
		}
		await expect(
			canvas.getByText('Pingo Doce').closest('.ledger-row') as HTMLElement
		).toHaveTextContent('Current account · Monzo');

		// Where the bank's line is the name, nothing is said twice.
		await expect(canvas.getAllByText('Transfer to Ana Reis')).toHaveLength(1);
		await expect(canvas.getByText('Not settled')).toBeInTheDocument();
		await expect(canvas.getByText('−€60.00')).toBeInTheDocument();
		await expect(canvas.getAllByText('Salary')).toHaveLength(1);
		await expect(canvas.getByText('+€2,180.00')).toBeInTheDocument();

		await expect(canvas.getAllByText(span)).toHaveLength(1);

		// Marked on the Galp row, once on the page, and never on a transfer that
		// has not settled.
		await expect(canvas.getAllByText('Unusual')).toHaveLength(1);
		await expect(
			(canvas.getByText('Galp').closest('.ledger-row') as HTMLElement).textContent
		).toContain('Unusual');
		const transfer = canvas.getByText('Transfer to Ana Reis').closest('.ledger-row') as HTMLElement;
		await expect(transfer).toHaveTextContent('Not settled');
		await expect(transfer).not.toHaveTextContent('Unusual');
	}}
/>

<!-- J07.A / 02 -->
<Story
	name="OneAccount"
	tags={['kind-state']}
	args={{ count: 204, filterAccount: 'Current account · Monzo' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Showing one account')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Show all accounts' })).toBeInTheDocument();
		// A filter is not containment: there is no breadcrumb.
		await expect(canvas.queryByRole('navigation', { name: /breadcrumb/i })).not.toBeInTheDocument();
	}}
/>

<!-- J07.A / 03 — a day that straddled the boundary is named again. -->
<Story
	name="AnOlderPage"
	tags={['kind-state']}
	args={{
		days: olderDays,
		span: '12 August to 3 August 2026',
		pages,
		currentPage: 'p2',
		newestHref: '/transactions',
		oldestHref: '/transactions?page=p3'
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('4 August 2026')).toBeInTheDocument();
		// The dots carry the same words as accessible names, so the footer's own
		// span is counted where it lives.
		await expect(
			within(canvas.getByRole('navigation', { name: 'Pages of transactions' })).getAllByText(
				'12 August to 3 August 2026'
			)
		).toHaveLength(1);
		await expect(canvas.getByRole('link', { name: 'Oldest' })).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Newest' })).toBeInTheDocument();
	}}
/>

<!-- J07.A / 04 -->
<Story
	name="TheOldestPage"
	tags={['kind-state']}
	args={{
		days: [olderDays[1]!],
		span: '4 June 2026',
		atOldest: true,
		pages,
		currentPage: 'p3',
		newestHref: '/transactions'
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/Nothing older/)).toBeInTheDocument();
		await expect(
			within(canvas.getByRole('navigation', { name: 'Pages of transactions' })).getAllByText(
				'4 June 2026'
			)
		).toHaveLength(1);
		await expect(canvas.getByRole('button', { name: 'Oldest' })).toBeDisabled();
		await expect(canvas.getByRole('link', { name: 'Newest' })).toBeInTheDocument();
	}}
/>

<!-- J07.B / 01 -->
<Story
	name="NoBankConnected"
	tags={['kind-state']}
	args={{ days: [], count: 0, freshness: undefined, span: undefined, noBank: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('No transactions yet')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Connect a bank' })).toBeInTheDocument();
	}}
/>

<!-- J07.B / 02 — the bank is not broken, and is never described as broken. -->
<Story
	name="BankNotSendingTransactions"
	tags={['kind-state']}
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		narrow: [{ connectionId: 'c1', bankName: 'Monzo', widenHref: '/connect/widen/c1' }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo is not sending transactions yet')).toBeInTheDocument();
		await expect(canvas.getByText('Nothing read from Monzo')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Widen at Monzo' })).toBeInTheDocument();
		await expect(canvas.queryByText(/broken|expired|failing/i)).not.toBeInTheDocument();
	}}
/>

<!-- J07.B / 03 -->
<Story
	name="OwnsNoAccount"
	tags={['kind-state']}
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		ownsNothing: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('You do not own any accounts')).toBeInTheDocument();
		// Nothing says anything was withheld.
		await expect(canvas.queryByText(/withheld|hidden/i)).not.toBeInTheDocument();
	}}
/>

<!-- The first visit, while the banks are being asked. The rows the member
     already has would paint here; with none, the screen says the list fills as
     the bank answers rather than showing an empty list that looks settled. -->
<Story
	name="SyncingOnArrival"
	tags={['kind-waiting']}
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		refreshing: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Nothing read yet')).toBeInTheDocument();
		await expect(canvas.getByText('The list fills as the bank answers.')).toBeInTheDocument();
		// Announced, because a member not looking at the toolbar has no other
		// way to know a request is in flight.
		await expect(canvas.getByRole('status')).toHaveTextContent('Bringing transactions up to date');
	}}
/>

<!-- Rows already held stay on screen while the banks are asked: no spinner over
     the list, and nothing moves when it lands. -->
<Story
	name="SyncingWithRowsAlreadyHeld"
	tags={['kind-waiting']}
	args={{ refreshing: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
		await expect(canvas.getByText('Pingo Doce')).toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Refreshing…' })).toBeDisabled();
	}}
/>

<!-- J07.B / 04 — no period is promised before the first read. -->
<Story
	name="ReadingForTheFirstTime"
	tags={['kind-waiting']}
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		problems: [{ connectionId: 'c1', bankName: 'Monzo', kind: 'first-read' as const }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Reading your transactions')).toBeInTheDocument();
		await expect(
			canvas.getByText(/is sending them now\. This is the first time/)
		).toBeInTheDocument();
		await expect(canvas.queryByText(/Reaching back to/)).not.toBeInTheDocument();
	}}
/>

<!-- J07.C / 01 — the rows stay, with their original time. -->
<Story
	name="RefreshRefused"
	tags={['kind-error']}
	args={{
		problems: [
			{ connectionId: 'c1', bankName: 'Monzo', kind: 'rate-limited' as const, retryAfter: '11:20' }
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo has been asked too often')).toBeInTheDocument();
		// The notice says it and the live region announces it, in the same words.
		await expect(canvas.getAllByText(/after 11:20/).length).toBeGreaterThan(1);
		await expect(
			canvas.getAllByText(/What is below was\s+already read, and it has not changed/).length
		).toBeGreaterThan(0);
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
	}}
/>

<!-- J07.C / 02 -->
<Story
	name="OneBankDidNotAnswer"
	tags={['kind-error']}
	args={{
		problems: [
			{
				connectionId: 'c2',
				bankName: 'Montepio',
				kind: 'unreachable' as const,
				lastReadAt: '08:02'
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Montepio did not answer')).toBeInTheDocument();
		await expect(canvas.getByText(/last read at 08:02/)).toBeInTheDocument();
		await expect(canvas.getAllByText(/Everything else is up to date/).length).toBeGreaterThan(0);
	}}
/>

<!-- J07.D / 01 — the history stays, and the way back is offered. -->
<Story
	name="AccessHasRunOut"
	tags={['kind-error']}
	args={{
		problems: [
			{
				connectionId: 'c1',
				bankName: 'Monzo',
				kind: 'access-ended' as const,
				restoreHref: '/connect/restore/c1'
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo has stopped sending transactions')).toBeInTheDocument();
		await expect(canvas.getByText(/The access you gave Monzo has run out/)).toBeInTheDocument();
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
	}}
/>

<!-- J07.D / 02 — disconnecting ends access; it does not destroy the record. -->
<Story
	name="BankDisconnected"
	tags={['kind-outcome']}
	args={{
		problems: [{ connectionId: 'c1', bankName: 'Monzo', kind: 'disconnected' as const }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo is no longer connected')).toBeInTheDocument();
		await expect(
			canvas.getByText(/Its accounts have left Overview and its balances are gone/)
		).toBeInTheDocument();
		await expect(canvas.getByText(/still here/)).toBeInTheDocument();
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
	}}
/>

<!-- Everything fits on one page: the span still shows, but no way to page is
     offered — the pager keeps the span rather than disappearing (5.5). -->
<Story
	name="EverythingFitsOnOnePage"
	tags={['kind-behaviour']}
	args={{ count: 5 }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByRole('navigation', { name: 'Pages of transactions' })
		).toBeInTheDocument();
	}}
/>

<!-- J07.A / 01 · Compact — the rows stack; the bar is the shell's sibling. -->
<Story
	name="Compact"
	tags={['kind-state', 'size-compact']}
	args={{ compact: true }}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Every account you own, newest first.')).toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
		await expect(canvas.queryByText('Description')).not.toBeInTheDocument();
		await expect(canvas.getByText('Today, 17 September')).toBeInTheDocument();
		await expect(canvas.getByText('DD NOS COMUNICACOES SA 000000234058260')).toBeInTheDocument();
		await expect(canvas.getByText('NOS')).toBeInTheDocument();
		await expect(canvas.getAllByText(span)).toHaveLength(1);
	}}
/>

<!-- J07.A / 02 · Compact -->
<Story
	name="CompactOneAccount"
	tags={['kind-state', 'size-compact']}
	args={{ compact: true, count: 204, filterAccount: 'Current account · Monzo' }}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Newest first.')).toBeInTheDocument();
		await expect(canvas.getByText('Showing one account')).toBeInTheDocument();
		await expect(canvas.getAllByText('Current account · Monzo').length).toBeGreaterThan(0);
		await expect(canvas.getByRole('link', { name: 'Show all accounts' })).toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
	}}
/>

<!-- J08.A / 02 — the bank was widened and is sending now. -->
<Story
	name="BankNowIncluded"
	tags={['kind-outcome']}
	args={{
		problems: [{ connectionId: 'c1', bankName: 'Monzo', kind: 'widened' as const }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo is sending transactions now')).toBeInTheDocument();
		await expect(canvas.getByText(/who owns these accounts has changed/)).toBeInTheDocument();
		await expect(
			canvas.getByText(/This is the first read, so it goes back as far as/)
		).toBeInTheDocument();
	}}
/>
