<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, userEvent, within } from 'storybook/test';
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

	/** Both halves of one movement between two accounts the member owns, on
	 *  the day it happened. Each is labelled; neither names the other. */
	const transferDay: LedgerDay = {
		date: '2026-09-01',
		entries: [
			{
				id: '9',
				description: 'Transfer to savings',
				account: 'Current account · Monzo',
				amount: '−€500.00',
				date: '1 Sep',
				negative: true,
				transfer: true,
				initials: 'TS'
			},
			{
				id: '10',
				description: 'Transfer from current account',
				account: 'Savings · Monzo',
				amount: '+€500.00',
				date: '1 Sep',
				transfer: true,
				initials: 'TC'
			}
		]
	};

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

	const accounts = [
		{ value: 'acc-current', label: 'Current account · Monzo' },
		{ value: 'acc-savings', label: 'Savings · Monzo' },
		{ value: 'acc-casa', label: 'Casa CC · Montepio' }
	];

	/** September 2026 back to September 2025, as the route labels them. */
	const months = Array.from({ length: 13 }, (_, i) => {
		const d = new Date(Date.UTC(2026, 8 - i, 1));
		return {
			value: d.toISOString().slice(0, 7),
			label: d.toLocaleDateString('en-GB', { month: 'long', year: 'numeric', timeZone: 'UTC' })
		};
	});

	const none = { account: '', q: '', month: '', direction: '' as const };

	function galp(id: string, day: number, amount: string): LedgerDay['entries'][number] {
		return {
			id,
			description: 'Galp',
			banksLine: `COMPRA GALP A5 OEIRAS 2300022701589${day}`,
			account: 'Current account · Monzo',
			amount,
			date: `${day} Sep`,
			negative: true,
			initials: 'GA'
		};
	}

	const galpDays: LedgerDay[] = [
		{ date: '2026-09-16', entries: [galp('g1', 16, '−€71.40')] },
		{ date: '2026-09-02', entries: [galp('g2', 2, '−€64.05'), galp('g3', 2, '−€12.30')] }
	];

	const augustDays: LedgerDay[] = [
		{
			date: '2026-08-31',
			entries: [
				{
					id: 'a1',
					description: 'Galp',
					banksLine: 'COMPRA GALP A5 OEIRAS 230002270158931',
					account: 'Current account · Monzo',
					amount: '−€58.20',
					date: '31 Aug',
					negative: true,
					initials: 'GA'
				},
				{
					id: 'a2',
					description: 'Salary',
					account: 'Current account · Monzo',
					amount: '+€2,180.00',
					date: '31 Aug',
					initials: 'SA'
				}
			]
		},
		{
			date: '2026-08-18',
			entries: [
				{
					id: 'a3',
					description: 'Continente',
					account: 'Current account · Monzo',
					amount: '−€118.72',
					date: '18 Aug',
					negative: true,
					initials: 'CO'
				}
			]
		}
	];

	const augustOldest: LedgerDay[] = [
		{
			date: '2026-08-01',
			entries: [
				{
					id: 'a9',
					description: 'EDP',
					banksLine: 'DD EDP COMERCIAL 000000118273645',
					account: 'Current account · Monzo',
					amount: '−€48.90',
					date: '1 Aug',
					negative: true,
					initials: 'ED'
				}
			]
		}
	];

	const augustPages = [
		{
			key: 'm1',
			label: '31 August to 18 August 2026',
			href: '/transactions?month=2026-08&page=m1'
		},
		{ key: 'm2', label: '17 August to 1 August 2026', href: '/transactions?month=2026-08&page=m2' }
	];

	const moneyInDays: LedgerDay[] = [
		{
			date: '2026-09-16',
			entries: [
				{
					id: 'i1',
					description: 'Salary',
					account: 'Current account · Monzo',
					amount: '+€2,180.00',
					date: '16 Sep',
					initials: 'SA'
				}
			]
		},
		transferDay.entries[1]
			? { date: '2026-09-01', entries: [transferDay.entries[1]] }
			: { date: '2026-09-01', entries: [] },
		{
			date: '2026-08-28',
			entries: [
				{
					id: 'i2',
					description: 'Refund from Worten',
					banksLine: 'DEVOLUCAO WORTEN 230002268110204',
					account: 'Current account · Monzo',
					amount: '+€34.99',
					date: '28 Aug',
					initials: 'WO'
				}
			]
		}
	];

	/** Every control of the filter bar, in the order it reads. */
	function filterControls(canvas: ReturnType<typeof within>) {
		const bar = within(canvas.getByRole('search', { name: 'Filter transactions' }));
		return {
			account: bar.getByRole('combobox', { name: 'Account' }) as HTMLSelectElement,
			search: bar.getByRole('searchbox', { name: 'Search' }) as HTMLInputElement,
			month: bar.queryByRole('combobox', { name: 'Month' }) as HTMLSelectElement | null,
			direction: bar.getByRole('radiogroup', { name: 'Direction' })
		};
	}

	/** What the totals under the count say, one entry per line, or null when
	 *  there are none. */
	function totalsOf(root: HTMLElement): string[] | null {
		const box = root.querySelector('.totals');
		if (!box) return null;
		return [...box.querySelectorAll('p')].map((p) => p.textContent!.replace(/\s+/g, ' ').trim());
	}

	/** The note on what the totals leave out, which opens on asking. */
	async function noteOf(canvas: ReturnType<typeof within>): Promise<string | null> {
		const why = canvas.queryByRole('button', { name: 'What these figures leave out' });
		if (!why) return null;
		await expect(why).toHaveAttribute('aria-expanded', 'false');
		await userEvent.click(why);
		await expect(why).toHaveAttribute('aria-expanded', 'true');
		return document.getElementById('totals-note')?.textContent?.trim() ?? null;
	}

	/** The words a select shows for what is chosen. */
	function chosen(select: HTMLSelectElement | null): string | undefined {
		return select?.selectedOptions[0]?.textContent?.trim();
	}

	function chosenDirection(group: HTMLElement): string | undefined {
		return within(group)
			.getAllByRole('radio')
			.find((r) => r.getAttribute('aria-checked') === 'true')
			?.textContent?.trim();
	}

	const { Story } = defineMeta({
		title: 'Pages/TransactionsScreen',
		component: TransactionsScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen', shell: '/transactions' },
		args: {
			days: [...days, transferDay],
			count: 384,
			freshness,
			span,
			today,
			accounts,
			months,
			clearHref: '/transactions'
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
		await expect(canvas.queryByText(/newest first/i)).not.toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Refresh' })).toBeInTheDocument();
		await expect(canvas.getByText('384 transactions')).toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();

		// Nothing is filtered, and nothing offers to clear a filter.
		const bar = filterControls(canvas);
		await expect(chosen(bar.account)).toBe('All accounts');
		await expect(chosen(bar.month)).toBe('All months');
		await expect(chosenDirection(bar.direction)).toBe('All');
		await expect(bar.search).toHaveValue('');
		await expect(bar.search).toHaveAttribute('placeholder', "Who it was with, or the bank's line");
		await expect(
			within(bar.account)
				.getAllByRole('option')
				.map((o) => o.textContent?.trim())
		).toEqual(['All accounts', 'Current account · Monzo', 'Savings · Monzo', 'Casa CC · Montepio']);
		await expect(within(bar.month!).getAllByRole('option')[2]).toHaveTextContent('August 2026');
		await expect(canvas.queryByRole('link', { name: 'Clear filters' })).toBeNull();
		// The whole ledger is not a period anybody asked about.
		await expect(totalsOf(canvasElement)).toBeNull();

		const columns = within(canvasElement.querySelector('.columns') as HTMLElement);
		for (const heading of ['Description', 'Account', 'Amount']) {
			await expect(columns.getByText(heading)).toBeInTheDocument();
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
		// A payment to somebody else is not a transfer between your accounts,
		// however much it reads like one.
		await expect(transfer).not.toHaveTextContent('Between your accounts');

		// Both halves of the pair, labelled, and nothing else on the page is.
		await expect(canvas.getAllByText('Between your accounts')).toHaveLength(2);
		for (const [name, amount] of [
			['Transfer to savings', '−€500.00'],
			['Transfer from current account', '+€500.00']
		] as const) {
			const row = canvas.getByText(name).closest('.ledger-row') as HTMLElement;
			await expect(row).toHaveTextContent(amount);
			await expect(row).toHaveTextContent('Between your accounts');
		}
	}}
/>

<!-- J07.A / 02 -->
<Story
	name="OneAccount"
	tags={['kind-state']}
	args={{
		count: 204,
		filters: { ...none, account: 'acc-savings' },
		days: [
			{
				date: '2026-09-01',
				entries: [transferDay.entries[1]!]
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: '204 transactions' })).toBeInTheDocument();
		await expect(chosen(filterControls(canvas).account)).toBe('Savings · Monzo');
		await expect(canvas.getByRole('link', { name: 'Clear filters' })).toBeInTheDocument();
		await expect(canvas.queryByText('Showing one account')).not.toBeInTheDocument();
		// A filter is not containment: there is no breadcrumb.
		await expect(canvas.queryByRole('navigation', { name: /breadcrumb/i })).not.toBeInTheDocument();

		// Reading one account still labels the half that is in it: pairing is
		// done across every account the member owns, and only the page is
		// narrowed. Its partner is not here, and nothing names it.
		const row = canvas
			.getByText('Transfer from current account')
			.closest('.ledger-row') as HTMLElement;
		await expect(row).toHaveTextContent('+€500.00');
		await expect(row).toHaveTextContent('Between your accounts');
		await expect(canvas.getAllByText('Between your accounts')).toHaveLength(1);
		await expect(canvas.queryByText('Transfer to savings')).not.toBeInTheDocument();
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
		await expect(canvas.queryByRole('search', { name: 'Filter transactions' })).toBeNull();
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
		await expect(canvas.queryByRole('search', { name: 'Filter transactions' })).toBeNull();
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
		await expect(canvas.getByText('384 transactions')).toBeInTheDocument();
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
		await expect(canvas.getByText('384 transactions')).toBeInTheDocument();
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
		await expect(canvas.getByText('384 transactions')).toBeInTheDocument();
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
		await expect(canvas.getByText('384 transactions')).toBeInTheDocument();
	}}
/>

<!-- Everything fits on one page: the span still shows, but no way to page is
     offered — the pager keeps the span rather than disappearing (5.5). -->
<Story
	name="EverythingFitsOnOnePage"
	tags={['kind-behaviour']}
	args={{ count: 5, months: [months[0]!] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByRole('navigation', { name: 'Pages of transactions' })
		).toBeInTheDocument();
		// Every row falls in one month, so there is nowhere for a month to go.
		await expect(filterControls(canvas).month).toBeNull();
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
		await expect(canvas.queryByText(/newest first/i)).not.toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
		await expect(canvas.getByText('384 transactions')).toBeInTheDocument();
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
	args={{ compact: true, count: 204, filters: { ...none, account: 'acc-savings' } }}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByText('Showing one account')).not.toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
		// On a phone the account waits in the sheet, and the button says one
		// filter is chosen.
		await userEvent.click(canvas.getByRole('button', { name: 'Filters 1 chosen' }));
		const sheet = within(canvas.getByRole('dialog', { name: 'Filters' }));
		await expect(
			chosen(sheet.getByRole('combobox', { name: 'Account' }) as HTMLSelectElement)
		).toBe('Savings · Monzo');
		await expect(sheet.getByRole('link', { name: 'Clear filters' })).toBeInTheDocument();
		await userEvent.click(sheet.getByRole('button', { name: 'Done' }));
		await expect(canvas.queryByRole('dialog')).toBeNull();
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

<!-- A search: the count and the rows are the matches, and the outcome is
     announced to a member whose focus stayed in the search box. -->
<Story
	name="Searched"
	tags={['kind-state']}
	args={{
		filters: { ...none, q: 'galp' },
		count: 12,
		days: galpDays,
		span: '16 September to 2 September 2026',
		totals: [{ moneyIn: '+€0.00', moneyOut: '−€612.40' }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: '12 transactions' })).toBeInTheDocument();
		await expect(filterControls(canvas).search).toHaveValue('galp');
		const rows = [...canvasElement.querySelectorAll('.ledger-row')];
		await expect(rows).toHaveLength(3);
		for (const row of rows) await expect(row).toHaveTextContent('Galp');
		await expect(canvas.getByRole('status')).toHaveTextContent('12 transactions');
		await expect(totalsOf(canvasElement)).toEqual(['In +€0.00 Out −€612.40']);
		await expect(await noteOf(canvas)).toBeNull();
		await expect(canvas.getByRole('link', { name: 'Clear filters' })).toBeInTheDocument();
	}}
/>

<!-- One month, as a link from Overview opens it: the count is the month's and
     the dates on screen are inside it. -->
<Story
	name="OneMonth"
	tags={['kind-state']}
	args={{
		filters: { ...none, month: '2026-08' },
		count: 61,
		days: augustDays,
		span: '31 August to 18 August 2026',
		pages: augustPages,
		currentPage: 'm1',
		oldestHref: '/transactions?month=2026-08&oldest=1',
		totals: [{ moneyIn: '+€2,214.99', moneyOut: '−€1,204.55' }],
		transfersLeftOut: 1,
		notSettled: 1
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: '61 transactions' })).toBeInTheDocument();
		await expect(chosen(filterControls(canvas).month)).toBe('August 2026');
		// The same month on Overview adds up the same way: settled money, and a
		// transfer between the member's accounts moves nothing.
		await expect(totalsOf(canvasElement)).toEqual(['In +€2,214.99 Out −€1,204.55']);
		await expect(await noteOf(canvas)).toBe(
			'Leaves out 1 transfer between your accounts. Does not count 1 payment not yet settled.'
		);
		await expect(
			within(canvas.getByRole('navigation', { name: 'Pages of transactions' })).getAllByText(
				'31 August to 18 August 2026'
			)
		).toHaveLength(1);
	}}
/>

<!-- The end of a month is the end of what matches, not the end of what the
     bank would give. -->
<Story
	name="TheOldestPageOfAMonth"
	tags={['kind-state']}
	args={{
		filters: { ...none, month: '2026-08' },
		count: 61,
		days: augustOldest,
		span: '1 August 2026',
		atOldest: true,
		pages: augustPages,
		currentPage: 'm2',
		newestHref: '/transactions?month=2026-08'
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Nothing older matches these filters.')).toBeInTheDocument();
		await expect(
			canvas.queryByText('Nothing older. This is as far back as the bank would go.')
		).not.toBeInTheDocument();
	}}
/>

<!-- Money in only. The half of a transfer that arrived still reads as one. -->
<Story
	name="MoneyIn"
	tags={['kind-state']}
	args={{
		filters: { ...none, direction: 'in' as const },
		count: 38,
		days: moneyInDays,
		span: '16 September to 28 August 2026',
		totals: [{ moneyIn: '+€9,034.99', moneyOut: '−€0.00' }],
		transfersLeftOut: 2
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: '38 transactions' })).toBeInTheDocument();
		await expect(chosenDirection(filterControls(canvas).direction)).toBe('Money in');
		// Only the side that was asked for.
		await expect(totalsOf(canvasElement)).toEqual(['In +€9,034.99']);
		await expect(await noteOf(canvas)).toBe('Leaves out 2 transfers between your accounts.');
		const amounts = [...canvasElement.querySelectorAll('.ledger-row')].map((row) =>
			within(row as HTMLElement)
				.getByText(/€/)
				.textContent?.trim()
		);
		await expect(amounts).toHaveLength(3);
		for (const amount of amounts) await expect(amount?.startsWith('+')).toBe(true);
		await expect(
			canvas.getByText('Transfer from current account').closest('.ledger-row') as HTMLElement
		).toHaveTextContent('Between your accounts');
	}}
/>

<!-- Every filter at once. Clear filters is a link to the whole ledger, and it
     takes focus to the count so a keyboard lands where the list starts. -->
<Story
	name="FiltersCombined"
	tags={['kind-state']}
	args={{
		filters: { account: 'acc-current', q: 'galp', month: '2026-08', direction: 'out' as const },
		count: 3,
		days: [{ date: '2026-08-31', entries: [augustDays[0]!.entries[0]!] }],
		span: '31 August 2026',
		totals: [{ moneyIn: '+€0.00', moneyOut: '−€174.60' }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: '3 transactions' })).toBeInTheDocument();
		await expect(totalsOf(canvasElement)).toEqual(['Out −€174.60']);
		const bar = filterControls(canvas);
		await expect(chosen(bar.account)).toBe('Current account · Monzo');
		await expect(bar.search).toHaveValue('galp');
		await expect(chosen(bar.month)).toBe('August 2026');
		await expect(chosenDirection(bar.direction)).toBe('Money out');

		const clear = canvas.getByRole('link', { name: 'Clear filters' });
		await expect(clear).toHaveAttribute('href', '/transactions');
		// The story stays put; the link's own handler still runs.
		canvasElement.addEventListener('click', (event) => event.preventDefault(), { once: true });
		await userEvent.click(clear);
		await expect(canvas.getByRole('heading', { name: '3 transactions' })).toHaveFocus();
	}}
/>

<!-- Filters that match nothing. The ledger is not empty, a bank is not
     missing, and the one way out is to clear them. -->
<Story
	name="NothingMatches"
	tags={['kind-state']}
	args={{
		filters: { ...none, q: 'plumber' },
		days: [],
		count: 0,
		span: undefined,
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
		await expect(
			canvas.getByText('Nothing matches these filters', { ignore: '[role=status]' })
		).toBeInTheDocument();
		await expect(
			canvas.getByText('Change a filter above, or clear them to see every transaction.')
		).toBeInTheDocument();
		await expect(canvas.getAllByRole('link', { name: 'Clear filters' })).toHaveLength(1);
		await expect(filterControls(canvas).search).toHaveValue('plumber');
		await expect(canvas.getByRole('button', { name: 'Refresh' })).toBeInTheDocument();
		await expect(canvas.getByText('Montepio did not answer')).toBeInTheDocument();
		await expect(canvas.getByText(freshness)).toBeInTheDocument();
		await expect(canvas.queryByRole('link', { name: 'Connect a bank' })).toBeNull();
		await expect(canvas.queryByText('Nothing read yet')).not.toBeInTheDocument();
		// The screen's own live region, not the bank's notice.
		await expect(canvasElement.querySelector('.sr-only[role="status"]')).toHaveTextContent(
			'Nothing matches these filters'
		);
	}}
/>

<!-- One account and nothing else in force: that account has nothing, which is
     a different sentence from nothing matching. -->
<Story
	name="AnAccountWithNothing"
	tags={['kind-state']}
	args={{ filters: { ...none, account: 'acc-casa' }, days: [], count: 0, span: undefined }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByText('Casa CC · Montepio has no transactions', { ignore: '[role=status]' })
		).toBeInTheDocument();
		await expect(canvas.getByRole('status')).toHaveTextContent(
			'Casa CC · Montepio has no transactions'
		);
		await expect(canvas.getByText('The bank has sent none for this account.')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Show all accounts' })).toHaveAttribute(
			'href',
			'/transactions'
		);
		await expect(chosen(filterControls(canvas).account)).toBe('Casa CC · Montepio');
	}}
/>

<!-- On a phone the bar is the search and a Filters button, under the
     freshness line, so the list starts on the first screen. -->
<Story
	name="CompactFiltered"
	tags={['kind-state', 'size-compact']}
	args={{
		compact: true,
		filters: { ...none, q: 'galp', month: '2026-08' },
		count: 3,
		days: [{ date: '2026-08-31', entries: [augustDays[0]!.entries[0]!] }],
		span: '31 August 2026'
	}}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const search = canvas.getByRole('searchbox', { name: 'Search' });
		const open = canvas.getByRole('button', { name: 'Filters 1 chosen' });
		await expect(search).toHaveValue('galp');
		// The freshness line, then the search and the button on one row.
		await expect(search.getBoundingClientRect().top).toBeGreaterThan(
			canvas.getByText(freshness).getBoundingClientRect().top
		);
		await expect(
			Math.abs(open.getBoundingClientRect().top - search.getBoundingClientRect().top)
		).toBeLessThan(8);
		await expect(canvas.queryByRole('combobox')).toBeNull();
		await expect(canvas.getByRole('heading', { name: '3 transactions' })).toBeInTheDocument();
	}}
/>

<!-- Two currencies are two lines and never one sum: wimm holds no rates. -->
<Story
	name="TotalsInTwoCurrencies"
	tags={['kind-state']}
	args={{
		filters: { ...none, month: '2026-08' },
		count: 61,
		days: augustDays,
		span: '31 August to 18 August 2026',
		totals: [
			{ moneyIn: '+€2,214.99', moneyOut: '−€1,204.55' },
			{ moneyIn: '+US$0.00', moneyOut: '−US$42.00' }
		]
	}}
	play={async ({ canvasElement }) => {
		await expect(totalsOf(canvasElement)).toEqual([
			'In +€2,214.99 Out −€1,204.55',
			'In +US$0.00 Out −US$42.00'
		]);
	}}
/>
