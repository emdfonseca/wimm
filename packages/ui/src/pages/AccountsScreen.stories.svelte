<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import AccountsScreen from './AccountsScreen.svelte';

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
		title: 'Pages/AccountsScreen',
		component: AccountsScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen', shell: '/accounts' },
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
			problems: [],
			onrefresh: fn(),
			onconnect: fn(),
			onrestore: fn()
		}
	});
</script>

<Story
	name="Connected"
	tags={['kind-state']}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByRole('heading', { name: 'Accounts' })
		).toBeInTheDocument();
	}}
/>

<!-- Accounts shows no total at all. -->
<Story
	name="Empty"
	tags={['kind-state']}
	args={{ accounts: [] }}
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

<!-- The readings already on screen stay, with their original times: losing the
     previous number is the one thing that must not happen. -->
<Story
	name="Refresh refused"
	tags={['kind-error']}
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
		const banner = within(canvasElement.querySelector('.banner') as HTMLElement);
		await expect(banner.getByText(/limits how often wimm may read balances/)).toBeInTheDocument();
		await expect(banner.getByText(/14:20/)).toBeInTheDocument();
		await expect(canvas.getByText('Balances cannot be refreshed yet')).toBeInTheDocument();
		await expect(
			canvas.getByText(/The balances below are the ones already read, at\s+the times shown/)
		).toBeInTheDocument();
		await expect(canvas.getByText('€4,200.10')).toBeInTheDocument();
	}}
/>

<!-- A bank that refused without saying when says so, never a zero rendered as
     an instruction. -->
<Story
	name="Rate limited with no retry time"
	tags={['kind-error']}
	args={{
		accounts: [{ ...current, stale: true }],
		problems: [{ connectionId: 'c1', bankName: 'Montepio', kind: 'rate-limited' as const }]
	}}
	play={async ({ canvasElement }) => {
		const banner = within(canvasElement.querySelector('.banner') as HTMLElement);
		await expect(banner.getByText(/did not say when the next one is possible/)).toBeInTheDocument();
	}}
/>

<!-- One bank failing leaves the other's figures untouched. -->
<Story
	name="One bank not answering"
	tags={['kind-error']}
	args={{
		accounts: [current, { ...savings, bank: 'ActivoBank', stale: true, id: 'a4' }],
		problems: [{ connectionId: 'c2', bankName: 'ActivoBank', kind: 'unreachable' as const }]
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
	tags={['kind-error']}
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
	tags={['kind-state']}
	args={{ accounts: [{ ...current, numberSuffix: undefined }] }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).queryByText(/••••/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="Refreshing"
	tags={['kind-waiting']}
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
	tags={['kind-behaviour']}
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Refresh balances' }));
		await expect(args.onrefresh).toHaveBeenCalledOnce();
	}}
/>

<Story
	name="An overdrawn account"
	tags={['kind-state']}
	args={{ accounts: [{ ...current, balance: '−€312.40', negative: true }] }}
	play={async ({ canvasElement }) => {
		// The sign is what carries direction, never the colour alone (ADR 0006).
		await expect(within(canvasElement).getAllByText('−€312.40').length).toBeGreaterThan(0);
	}}
/>

<!-- canvas.md: returning to a failure moves focus to the page banner. Landing
     at the top of the page instead shows figures that look unchanged, which is
     exactly what a stale reading looks like. -->
<Story
	name="Focus lands on the banner"
	tags={['kind-behaviour']}
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
	tags={['kind-behaviour']}
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
	tags={['kind-behaviour']}
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
	tags={['kind-behaviour']}
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
	tags={['kind-outcome']}
	args={{ outcome: 'no-accounts' as const }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/granted access but offered no accounts/)
		).toBeInTheDocument();
	}}
/>

<Story
	name="Consent declined"
	tags={['kind-outcome']}
	args={{ outcome: 'declined' as const }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/nothing was added to the household/)).toBeInTheDocument();
		await expect(canvas.getByText(/pick a different bank/)).toBeInTheDocument();
	}}
/>

<Story
	name="A return that was already used"
	tags={['kind-outcome']}
	args={{ outcome: 'already-connected' as const }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText(/already connected/)).toBeInTheDocument();
	}}
/>

<!-- Frame J06.A / 03. A restore comes back through the same return as a first
     connection, so without this it was announced as one. -->
<Story
	name="Access restored"
	tags={['kind-outcome']}
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
	tags={['kind-outcome']}
	args={{ outcome: 'disconnected' as const, outcomeBank: 'Montepio', outcomeAccountCount: 3 }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Montepio was disconnected')).toBeInTheDocument();
		await expect(canvas.getByText(/Its 3 accounts are no longer shown/)).toBeInTheDocument();
	}}
/>

<Story
	name="A bank that could not be connected"
	tags={['kind-outcome']}
	args={{ outcome: 'bank-unavailable' as const }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText(/could not be reached/)).toBeInTheDocument();
		await expect(
			within(canvasElement).getByText(
				/wimm could not finish connecting, so nothing was added\. This is not something you did/
			)
		).toBeInTheDocument();
	}}
/>

<!-- And no outcome says nothing at all: a member arriving normally must not be
     told about a hand-off they did not make. -->
<Story
	name="No outcome is silent"
	tags={['kind-behaviour']}
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
	tags={['kind-outcome']}
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

<!-- Who connected a bank and when its access ends. -->
<Story
	name="Connected banks"
	tags={['kind-state']}
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
		await expect(canvas.getByText(/access ran out tomorrow/)).toBeInTheDocument();
	}}
/>

<!-- Disconnect was unreachable: the dialog existed and nothing opened it. -->
<Story
	name="Disconnect is reachable"
	tags={['kind-behaviour']}
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
	tags={['kind-state']}
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
     naming the bank and who connected it. -->
<Story
	name="Accounts sit under their bank"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);

		const card = canvas.getByRole('region', { name: 'Caixa Económica Montepio Geral' });
		await expect(card).toBeInTheDocument();

		await expect(within(card).getByText('Conta à Ordem')).toBeInTheDocument();
		await expect(within(card).getByText('Poupança')).toBeInTheDocument();

		await expect(within(card).getByText(/Connected by Emanuel/)).toBeInTheDocument();
		await expect(within(card).getByRole('button', { name: 'Disconnect' })).toBeInTheDocument();
	}}
/>

<!-- A reading says it was read: "Read just now", not a bare "just now". -->
<Story
	name="A reading says it was read"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getAllByText(/Read 2 minutes ago/)[0]).toBeInTheDocument();
	}}
/>

<!-- J09.A / 05: an account left out stays on Accounts, marked, with no figure
     — an owner who cannot see it has no way back to it. -->
<Story
	name="An account left out"
	tags={['kind-state']}
	args={{
		accounts: [current, { ...savings, leftOut: true, balance: undefined, readAt: undefined }]
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Left out')).toBeInTheDocument();
		await expect(canvas.queryByText('€11,930.00')).not.toBeInTheDocument();
	}}
/>
