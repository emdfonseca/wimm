<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import TransactionsScreen, { type LedgerDay } from './TransactionsScreen.svelte';

	/**
	 * The two newest days, relative to whenever the story runs. The screen
	 * computes "Today" and "Yesterday" from the date it is given, so a fixture
	 * pinned to a calendar date would render as a plain date the moment that
	 * date passed.
	 */
	function daysAgo(n: number): string {
		const d = new Date();
		d.setDate(d.getDate() - n);
		return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
	}

	const days: LedgerDay[] = [
		{
			date: daysAgo(0),
			entries: [
				{
					id: '1',
					description: 'Pingo Doce',
					account: 'Current account · Monzo',
					amount: '−€42.18',
					date: '17 Sep',
					negative: true,
					initials: 'PD'
				},
				{
					id: '2',
					description: 'Transfer to Ana Reis',
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
			date: daysAgo(1),
			entries: [
				{
					id: '3',
					description: 'Galp',
					account: 'Current account · Monzo',
					amount: '−€71.40',
					date: '16 Sep',
					negative: true,
					initials: 'GA'
				},
				{
					id: '4',
					description: 'Salary',
					account: 'Current account · Monzo',
					amount: '+€2,180.00',
					date: '16 Sep',
					initials: 'SA'
				},
				{
					id: '5',
					description: 'NOS',
					account: 'Joint savings · Montepio',
					amount: '−€39.99',
					date: '15 Sep',
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
					account: 'Current account · Monzo',
					amount: '−€118.72',
					date: '4 Aug',
					negative: true,
					initials: 'CO'
				},
				{
					id: '7',
					description: 'Águas de Lisboa',
					account: 'Joint savings · Montepio',
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
					account: 'Current account · Monzo',
					amount: '+€2,180.00',
					date: '31 Jul',
					initials: 'SA'
				}
			]
		}
	];

	const freshness = 'Updated at 09:14. Reaching back to 4 June 2026.';

	const { Story } = defineMeta({
		title: 'Pages/TransactionsScreen',
		component: TransactionsScreen,
		tags: ['autodocs'],
		args: {
			days,
			count: 382,
			freshness,
			span: '17 September to 15 September 2026',
			olderHref: '/transactions?before=abc'
		}
	});
</script>

<!-- J07.A / 01 -->
<Story
	name="AsItOpens"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: 'Transactions', level: 1 })).toBeInTheDocument();
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
		await expect(canvas.getByText('Today')).toBeInTheDocument();
		await expect(canvas.getByText('Yesterday')).toBeInTheDocument();
		await expect(canvas.getByText('Not settled')).toBeInTheDocument();
	}}
/>

<!-- J07.A / 02 -->
<Story
	name="OneAccount"
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
	args={{
		days: olderDays,
		span: '4 August to 31 July 2026',
		newerHref: '/transactions?after=def'
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		// The day heading and the span of dates both name it.
		await expect(canvas.getAllByText(/August.*2026/).length).toBeGreaterThan(0);
		await expect(canvas.getByRole('link', { name: 'Newer' })).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Older' })).toBeInTheDocument();
	}}
/>

<!-- J07.A / 04 -->
<Story
	name="TheOldestPage"
	args={{
		days: [olderDays[1]!],
		span: '4 June 2026',
		olderHref: undefined,
		newerHref: '/transactions?after=def',
		atOldest: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/Nothing older/)).toBeInTheDocument();
		await expect(canvas.queryByRole('link', { name: 'Older' })).not.toBeInTheDocument();
	}}
/>

<!-- J07.B / 01 -->
<Story
	name="NoBankConnected"
	args={{ days: [], count: 0, freshness: undefined, span: undefined, olderHref: undefined, noBank: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('No transactions yet')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Connect a bank' })).toBeInTheDocument();
	}}
/>

<!-- J07.B / 02 — the bank is not broken, and is never described as broken. -->
<Story
	name="BankNotSendingTransactions"
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		olderHref: undefined,
		narrow: [{ connectionId: 'c1', bankName: 'Monzo', widenHref: '/connect/widen/c1' }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo is not sending transactions yet')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Widen at Monzo' })).toBeInTheDocument();
		await expect(canvas.queryByText(/broken|expired|failing/i)).not.toBeInTheDocument();
	}}
/>

<!-- J07.B / 03 -->
<Story
	name="OwnsNoAccount"
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		olderHref: undefined,
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
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		olderHref: undefined,
		refreshing: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Nothing read yet')).toBeInTheDocument();
		await expect(canvas.getByText('The list fills as the bank answers.')).toBeInTheDocument();
		// Announced, because a member not looking at the toolbar has no other
		// way to know a request is in flight.
		await expect(canvas.getByRole('status')).toHaveTextContent(
			'Bringing transactions up to date'
		);
	}}
/>

<!-- Rows already held stay on screen while the banks are asked: no spinner over
     the list, and nothing moves when it lands. -->
<Story
	name="SyncingWithRowsAlreadyHeld"
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
	args={{
		days: [],
		count: 0,
		freshness: undefined,
		span: undefined,
		olderHref: undefined,
		problems: [{ connectionId: 'c1', bankName: 'Monzo', kind: 'first-read' as const }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Reading your transactions')).toBeInTheDocument();
		await expect(canvas.queryByText(/Reaching back to/)).not.toBeInTheDocument();
	}}
/>

<!-- J07.C / 01 — the rows stay, with their original time. -->
<Story
	name="RefreshRefused"
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
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
	}}
/>

<!-- J07.C / 02 -->
<Story
	name="OneBankDidNotAnswer"
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
	}}
/>

<!-- J07.D / 01 — the history stays, and the way back is offered. -->
<Story
	name="AccessHasRunOut"
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
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
	}}
/>

<!-- J07.D / 02 — disconnecting ends access; it does not destroy the record. -->
<Story
	name="BankDisconnected"
	args={{
		problems: [{ connectionId: 'c1', bankName: 'Monzo', kind: 'disconnected' as const }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo is no longer connected')).toBeInTheDocument();
		await expect(canvas.getByText(/still here/)).toBeInTheDocument();
		await expect(canvas.getByText('382 transactions')).toBeInTheDocument();
	}}
/>

<!-- Everything fits on one page: no way to page is offered. -->
<Story
	name="EverythingFitsOnOnePage"
	args={{ count: 5, olderHref: undefined, newerHref: undefined }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByRole('navigation', { name: 'Pages of transactions' })).not.toBeInTheDocument();
	}}
/>

<!-- J07.A / 01 · Compact — the rows stack; the bar is the shell's sibling. -->
<Story
	name="Compact"
	args={{ compact: true }}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
		await expect(canvas.getByText('NOS')).toBeInTheDocument();
	}}
/>

<!-- J07.A / 02 · Compact -->
<Story
	name="CompactOneAccount"
	args={{ compact: true, count: 204, filterAccount: 'Current account · Monzo' }}
	globals={{ viewport: { value: 'compact' } }}
/>

<!-- J08.A / 02 — the bank was widened and is sending now. -->
<Story
	name="BankNowIncluded"
	args={{
		problems: [{ connectionId: 'c1', bankName: 'Monzo', kind: 'widened' as const }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Monzo is sending transactions now')).toBeInTheDocument();
		await expect(canvas.getByText(/who owns these accounts has changed/)).toBeInTheDocument();
	}}
/>
