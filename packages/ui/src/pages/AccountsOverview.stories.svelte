<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import AccountsOverview from './AccountsOverview.svelte';

	const current = {
		id: 'a1',
		connectionId: 'c1',
		name: 'Conta à Ordem',
		bank: 'Montepio',
		numberSuffix: '0538',
		balance: '€4,200.10',
		readAt: '2 minutes ago'
	};
	const savings = {
		id: 'a2',
		connectionId: 'c1',
		name: 'Poupança',
		bank: 'Montepio',
		numberSuffix: '7712',
		balance: '€11,930.00',
		readAt: '2 minutes ago'
	};

	const { Story } = defineMeta({
		title: 'Pages/AccountsOverview',
		component: AccountsOverview,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: {
			banks: [
				{
					connectionId: 'c1',
					bankId: 'PT:Montepio',
					bankName: 'Caixa Económica Montepio Geral',
					connectedBy: 'Emanuel',
					accessEndsOn: '16 December 2026',
					live: true,
					accountCount: 2,
					manageable: true
				}
			],
			accounts: [current, savings],
			totals: [{ total: '€16,130.10', currency: 'EUR' }],
			problems: [],
			onrefresh: fn(),
			onconnect: fn(),
			onrestore: fn()
		}
	});
</script>

<Story name="Connected" />

<!-- No total over nothing: a total of zero would say the household has no
     money rather than that wimm has not been told about any. -->
<Story
	name="Empty"
	args={{ accounts: [], totals: [] }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByText(/€0/)).not.toBeInTheDocument();
		await expect(
			canvas.queryByRole('button', { name: 'Refresh balances' })
		).not.toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Connect a bank' }));
		await expect(args.onconnect).toHaveBeenCalledOnce();
	}}
/>

<!-- wimm holds no rates, so currencies are never added and the screen says so. -->
<Story
	name="Two currencies"
	args={{
		accounts: [current, { ...savings, bank: 'Revolut', balance: '£2,500.00', id: 'a3' }],
		totals: [
			{ total: '€4,200.10', currency: 'EUR' },
			{ total: '£2,500.00', currency: 'GBP' }
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		// Scoped to the totals: the EUR total and the one EUR account carry the
		// same string, and a bare getByText would match both.
		const totals = within(canvas.getByRole('region', { name: 'Totals' }));
		await expect(totals.getByText('€4,200.10')).toBeInTheDocument();
		await expect(totals.getByText('£2,500.00')).toBeInTheDocument();
		await expect(canvas.getByText(/does not convert between them/)).toBeInTheDocument();
	}}
/>

<!-- The readings already on screen stay, with their original times: losing the
     previous number is the one thing that must not happen. -->
<Story
	name="Refresh refused"
	args={{
		accounts: [
			{ ...current, stale: true },
			{ ...savings, stale: true }
		],
		problems: [
			{
				connectionId: 'c1',
				bankName: 'Montepio',
				kind: 'rate-limited' as const,
				retryAfter: '14:20'
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		// Scoped to the banner: the same words are also in the live region that
		// announces the outcome, and a bare getByText matches both.
		const banner = within(canvasElement.querySelector('.banner') as HTMLElement);
		await expect(banner.getByText(/limits how often wimm may read balances/)).toBeInTheDocument();
		await expect(banner.getByText(/14:20/)).toBeInTheDocument();
		await expect(canvas.getByText('€4,200.10')).toBeInTheDocument();
	}}
/>

<!-- A bank that refused without saying when says so, never a zero rendered as
     an instruction. -->
<Story
	name="Rate limited with no retry time"
	args={{
		accounts: [{ ...current, stale: true }],
		problems: [{ connectionId: 'c1', bankName: 'Montepio', kind: 'rate-limited' as const }]
	}}
	play={async ({ canvasElement }) => {
		const banner = within(canvasElement.querySelector('.banner') as HTMLElement);
		await expect(
			banner.getByText(/did not say when the next one is possible/)
		).toBeInTheDocument();
	}}
/>

<!-- One bank failing leaves the other's figures untouched. -->
<Story
	name="One bank not answering"
	args={{
		accounts: [current, { ...savings, bank: 'ActivoBank', stale: true, id: 'a4' }],
		problems: [
			{ connectionId: 'c2', bankName: 'ActivoBank', kind: 'unreachable' as const }
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/ActivoBank did not answer/)).toBeInTheDocument();
		await expect(canvas.getByText('€11,930.00')).toBeInTheDocument();
	}}
/>

<!-- Access running out is not a failure to retry: it offers the flow again.
     At ActivoBank's one-day consent this is the normal resting state. -->
<Story
	name="Access run out"
	args={{
		accounts: [{ ...current, bank: 'ActivoBank', notUpdating: true }],
		problems: [{ connectionId: 'c3', bankName: 'ActivoBank', kind: 'access-ended' as const }]
	}}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/has stopped updating/)).toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Restore access' }));
		await expect(args.onrestore).toHaveBeenCalledWith('c3');
	}}
/>

<!-- An account seen at balance shows no identifier, because the server sent
     none. -->
<Story
	name="An account seen at balance level"
	args={{
		accounts: [{ ...current, numberSuffix: undefined }],
		totals: [{ total: '€4,200.10', currency: 'EUR' }]
	}}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).queryByText(/••••/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="Refreshing"
	args={{ refreshing: true }}
	play={async ({ canvasElement, args }) => {
		const button = within(canvasElement).getByRole('button', { name: 'Refreshing…' });
		await expect(button).toBeDisabled();
		await userEvent.click(button, { pointerEventsCheck: 0 });
		await expect(args.onrefresh).not.toHaveBeenCalled();
	}}
/>

<Story
	name="Refreshing on request"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(
			within(canvasElement).getByRole('button', { name: 'Refresh balances' })
		);
		await expect(args.onrefresh).toHaveBeenCalledOnce();
	}}
/>

<Story
	name="An overdrawn account"
	args={{
		accounts: [{ ...current, balance: '−€312.40', negative: true }],
		totals: [{ total: '−€312.40', currency: 'EUR' }]
	}}
/>

<!-- canvas.md: returning to a failure moves focus to the page banner. Landing
     at the top of the page instead shows figures that look unchanged, which is
     exactly what a stale reading looks like. -->
<Story
	name="Focus lands on the banner"
	args={{
		accounts: [{ ...current, stale: true }],
		problems: [{ connectionId: 'c1', bankName: 'Montepio', kind: 'unreachable' as const }]
	}}
	play={async ({ canvasElement }) => {
		const banner = canvasElement.querySelector('.banner');
		await expect(banner).toHaveFocus();
	}}
/>

<!-- canvas.md: refreshing announces its outcome. The figures change in place,
     so a member not watching them has no other way to know it finished. -->
<Story
	name="A refusal is announced"
	args={{
		accounts: [{ ...current, stale: true }],
		problems: [
			{
				connectionId: 'c1',
				bankName: 'Montepio',
				kind: 'rate-limited' as const,
				retryAfter: '14:20'
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const announced = within(canvasElement)
			.getAllByRole('status')
			.map((node) => node.textContent)
			.join(' ');
		await expect(announced).toContain('Montepio limits how often wimm may read balances');
		await expect(announced).toContain('14:20');
	}}
/>

<Story
	name="Which bank did not answer is announced"
	args={{
		accounts: [current],
		problems: [{ connectionId: 'c2', bankName: 'ActivoBank', kind: 'unreachable' as const }]
	}}
	play={async ({ canvasElement }) => {
		const announced = within(canvasElement)
			.getAllByRole('status')
			.map((node) => node.textContent)
			.join(' ');
		await expect(announced).toContain('ActivoBank could not be updated');
	}}
/>

<Story
	name="Refreshing is announced while it runs"
	args={{ refreshing: true }}
	play={async ({ canvasElement }) => {
		const announced = within(canvasElement)
			.getAllByRole('status')
			.map((node) => node.textContent)
			.join(' ');
		await expect(announced).toContain('Refreshing balances');
	}}
/>

<!-- A hand-off that did not land. Every one of these used to redirect here and
     say nothing: PayPal granted access, exposed no accounts, and the member saw
     an unchanged screen with no explanation. -->
<Story
	name="A bank that offered no accounts"
	args={{ outcome: 'no-accounts' as const }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/granted access but offered no accounts/)
		).toBeInTheDocument();
	}}
/>

<Story
	name="Consent declined"
	args={{ outcome: 'declined' as const }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/nothing was added to the household/)).toBeInTheDocument();
		await expect(canvas.getByText(/pick a different bank/)).toBeInTheDocument();
	}}
/>

<Story
	name="A return that was already used"
	args={{ outcome: 'already-connected' as const }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/already connected/)
		).toBeInTheDocument();
	}}
/>

<!-- Frame J06.A / 03. A restore comes back through the same return as a first
     connection, so without this it was announced as one. -->
<Story
	name="Access restored"
	args={{
		outcome: 'restored' as const,
		outcomeBank: 'Montepio',
		outcomeAccessEndsOn: '11 September',
		accounts: [current]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Montepio is updating again')).toBeInTheDocument();
		await expect(canvas.getByText(/Access runs until 11 September/)).toBeInTheDocument();
	}}
/>

<!-- Frame J05.A / 02, which counts the accounts that went. -->
<Story
	name="Disconnected"
	args={{ outcome: 'disconnected' as const, outcomeBank: 'Montepio', outcomeAccountCount: 3 }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Montepio was disconnected')).toBeInTheDocument();
		await expect(canvas.getByText(/Its 3 accounts are no longer shown/)).toBeInTheDocument();
	}}
/>

<Story
	name="A bank that could not be connected"
	args={{ outcome: 'bank-unavailable' as const }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText(/could not be reached/)).toBeInTheDocument();
	}}
/>

<!-- And no outcome says nothing at all: a member arriving normally must not be
     told about a hand-off they did not make. -->
<Story
	name="No outcome is silent"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByText(/nothing was added to the household/)).not.toBeInTheDocument();
		await expect(canvas.queryByText(/could not be reached/)).not.toBeInTheDocument();
	}}
/>

<!-- The connect flow now ends here. The accounts are the confirmation; the
     notice says they are yours and nobody else's yet. -->
<Story
	name="Just connected"
	args={{
		outcome: 'connected' as const,
		banks: [
			{
				connectionId: 'c1',
				bankId: 'PT:Montepio',
				bankName: 'Caixa Económica Montepio Geral',
				connectedBy: 'Emanuel',
				accessEndsOn: '16 December 2026',
				live: true,
				accountCount: 2,
				manageable: true
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/That bank is connected/)).toBeInTheDocument();
		await expect(canvas.getByText(/nobody else in the household sees them/)).toBeInTheDocument();
	}}
/>

<!-- Who connected a bank and when its access ends, which the spec requires and
     Overview showed nowhere. -->
<Story
	name="Connected banks"
	args={{
		banks: [
			{
				connectionId: 'c1',
				bankId: 'PT:Montepio',
				bankName: 'Caixa Económica Montepio Geral',
				connectedBy: 'Emanuel',
				accessEndsOn: '16 December 2026',
				live: true,
				accountCount: 2,
				manageable: true
			},
			{
				connectionId: 'c2',
				bankId: 'PT:ActivoBank',
				bankName: 'ActivoBank',
				connectedBy: 'Grace',
				accessEndsOn: 'tomorrow',
				live: false,
				accountCount: 1,
				manageable: false
			}
		]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/Connected by Emanuel/)).toBeInTheDocument();
		await expect(canvas.getByText(/access ends 16 December 2026/)).toBeInTheDocument();
		// The frame's wording for a dead connection: "access ran out <date>".
		await expect(canvas.getByText(/access ran out tomorrow/)).toBeInTheDocument();
	}}
/>

<!-- Disconnect was unreachable: the dialog existed and nothing opened it. -->
<Story
	name="Disconnect is reachable"
	args={{
		banks: [
			{
				connectionId: 'c1',
				bankId: 'PT:Montepio',
				bankName: 'Montepio',
				connectedBy: 'Emanuel',
				accessEndsOn: '16 December 2026',
				live: true,
				accountCount: 2,
				manageable: true
			}
		],
		ondisconnect: fn(),
		onmanage: fn()
	}}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);

		await userEvent.click(canvas.getByRole('button', { name: 'Disconnect' }));
		await expect(args.ondisconnect).toHaveBeenCalledOnce();

		await userEvent.click(canvas.getByRole('button', { name: 'Who sees these' }));
		await expect(args.onmanage).toHaveBeenCalledOnce();
	}}
/>

<!-- A bank whose accounts the member does not own offers no management control:
     one that would be refused is not shown. -->
<Story
	name="A bank the member cannot manage"
	args={{
		banks: [
			{
				connectionId: 'c2',
				bankId: 'PT:ActivoBank',
				bankName: 'ActivoBank',
				connectedBy: 'Grace',
				accessEndsOn: 'tomorrow',
				live: true,
				accountCount: 1,
				manageable: false
			}
		]
	}}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).queryByRole('button', { name: 'Who sees these' })
		).not.toBeInTheDocument();
	}}
/>

<!-- Frame J05.A / 01: accounts sit inside their bank's card, under a header
     naming the bank and who connected it. They were rendering as bare rows on
     the page background because nothing set connectionId, so the grouping never
     ran at all. -->
<Story
	name="Accounts sit under their bank"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);

		const card = canvas.getByRole('region', { name: 'Caixa Económica Montepio Geral' });
		await expect(card).toBeInTheDocument();

		// Both rows are inside the card, not siblings of it.
		await expect(within(card).getByText('Conta à Ordem')).toBeInTheDocument();
		await expect(within(card).getByText('Poupança')).toBeInTheDocument();

		// And the header the frame draws is there with them.
		await expect(within(card).getByText(/Connected by Emanuel/)).toBeInTheDocument();
		await expect(within(card).getByRole('button', { name: 'Disconnect' })).toBeInTheDocument();
	}}
/>

<!-- The frame's tile is labelled "Household total". -->
<Story
	name="The total is a labelled tile"
	play={async ({ canvasElement }) => {
		const totals = within(
			within(canvasElement).getByRole('region', { name: 'Totals' })
		);
		await expect(totals.getByText('Household total')).toBeInTheDocument();
		await expect(totals.getByText('€16,130.10')).toBeInTheDocument();
	}}
/>

<!-- A reading says it was read: "Read just now", not a bare "just now". -->
<Story
	name="A reading says it was read"
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getAllByText(/Read 2 minutes ago/)[0]
		).toBeInTheDocument();
	}}
/>

<!-- J09.A / 05: an account left out stays on Overview, marked, with no figure
     — an owner who cannot see it has no way back to it. The total passed in
     is already the household's, excluding it: that arithmetic is the
     backend's (ADR 0022), this only asserts the row itself. -->
<Story
	name="An account left out"
	args={{
		accounts: [current, { ...savings, leftOut: true, balance: undefined, readAt: undefined }],
		totals: [{ total: '€4,200.10', currency: 'EUR' }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Left out')).toBeInTheDocument();
		await expect(canvas.queryByText('€11,930.00')).not.toBeInTheDocument();
		await expect(canvas.getAllByText('€4,200.10')).toHaveLength(2);
	}}
/>
