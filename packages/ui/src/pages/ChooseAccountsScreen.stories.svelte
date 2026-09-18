<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import ChooseAccountsScreen from './ChooseAccountsScreen.svelte';

	const members = [
		{ id: 'grace', name: 'Grace' },
		{ id: 'alan', name: 'Alan' }
	];

	// Named rather than indexed: a story that reaches for accounts[0] is a
	// possibly-undefined under this repo's strictness settings.
	const current = {
		id: 'a1',
		name: 'Conta à Ordem',
		bankAccountName: 'Conta à Ordem',
		bankName: 'Montepio',
		meta: 'Montepio · •••• 0538',
		balance: '€4,200.10',
		owned: true
	};
	const savings = {
		id: 'a2',
		name: 'Poupança',
		bankAccountName: 'Poupança',
		bankName: 'Montepio',
		meta: 'Montepio · •••• 7712',
		balance: '€11,930.00',
		owned: true
	};
	const personal = {
		id: 'a3',
		name: 'Conta Pessoal',
		bankAccountName: 'Conta Pessoal',
		bankName: 'Montepio',
		meta: 'Montepio · •••• 5594',
		balance: '€820.44',
		owned: true
	};

	const accounts = [current, savings, personal];

	const { Story } = defineMeta({
		title: 'Pages/ChooseAccountsScreen',
		component: ChooseAccountsScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: {
			bankName: 'Montepio',
			accounts,
			members,
			levels: {},
			onownedchange: fn(),
			onownerchange: fn(),
			onlevelchange: fn(),
			onleaveout: fn(),
			onbringback: fn(),
			onfinish: fn()
		}
	});
</script>

<!-- As it opens: every account owned by the connecting member, nobody else
     granted anything. -->
<Story
	name="As it opens"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const box of canvas.getAllByRole('checkbox', { name: /is mine/ })) {
			await expect(box).toBeChecked();
		}
		await expect(canvas.getByText(/Nobody sees anything until you say so/)).toBeInTheDocument();
	}}
/>

<!-- Finishing having granted nobody anything is legitimate, so Finish is never
     disabled. This is the behaviour that changed with the ownership model. -->
<Story
	name="Finish is always enabled"
	play={async ({ canvasElement, args }) => {
		const finish = within(canvasElement).getByRole('button', { name: 'Finish' });
		await expect(finish).toBeEnabled();
		await userEvent.click(finish);
		await expect(args.onfinish).toHaveBeenCalledOnce();
	}}
/>

<Story
	name="One granted at balance"
	args={{ levels: { a1: { grace: 'balance' } } }}
	play={async ({ canvasElement }) => {
		const group = within(canvasElement).getByRole('radiogroup', {
			name: 'What Grace sees of Conta à Ordem'
		});
		await expect(within(group).getByRole('radio', { name: 'Balance' })).toBeChecked();
	}}
/>

<Story
	name="Two members at different levels"
	args={{ levels: { a1: { grace: 'details', alan: 'balance' } } }}
/>

<!-- Disowning an account nobody else sees says plainly that no balance will be
     read for it — which is what makes disowning meaningful. -->
<Story
	name="One disowned"
	args={{ accounts: [current, savings, { ...personal, owned: false }] }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/account is nobody's\. wimm will not read a balance for it/)
		).toBeInTheDocument();
	}}
/>

<!-- A level change is announced by name and level, not as a running count. -->
<Story
	name="Changing a level is announced"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		const group = canvas.getByRole('radiogroup', { name: 'What Grace sees of Poupança' });
		await userEvent.click(within(group).getByRole('radio', { name: 'Details' }));

		await expect(args.onlevelchange).toHaveBeenCalledWith('a2', 'grace', 'details');
		await expect(canvas.getByRole('status')).toHaveTextContent('Grace now sees details of Poupança');
	}}
/>

<Story
	name="Disowning is announced"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('checkbox', { name: 'Conta Pessoal is mine' }));

		await expect(args.onownedchange).toHaveBeenCalledWith('a3', false);
		await expect(canvas.getByRole('status')).toHaveTextContent("Conta Pessoal is nobody's");
	}}
/>

<!-- Reopened later, showing the choice as it stands rather than as it started. -->
<Story
	name="Reopened with an existing choice"
	args={{ levels: { a1: { grace: 'balance' }, a2: { alan: 'details' } } }}
/>

<!-- Restoring: the accounts kept what they had, and a newly offered one is
     surfaced rather than quietly added. -->
<Story
	name="Restoring with something new"
	args={{
		restoring: true,
		accounts: [
			current,
			{
				id: 'a9',
				name: 'Conta Nova',
				bankAccountName: 'Conta Nova',
				bankName: 'Montepio',
				meta: 'Montepio · •••• 9021',
				balance: '€64.00',
				owned: true,
				newlyOffered: true
			}
		],
		levels: { a1: { grace: 'balance' } }
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		// Where the canvas puts each sentence: the newly-offered account takes
		// the lede, and the reassurance about what was kept goes in the helper
		// (frame J06.A / 02).
		await expect(canvas.getByText(/offered an account it had not before/)).toBeInTheDocument();
		await expect(canvas.getByText(/kept their owners and levels/)).toBeInTheDocument();
	}}
/>

<!-- A member who owns none of these may look and change nothing. -->
<Story
	name="A member who owns none of it"
	args={{ readOnly: true, levels: { a1: { grace: 'balance' } } }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/not yours to change/)).toBeInTheDocument();
		for (const box of canvas.getAllByRole('checkbox')) {
			await expect(box).toBeDisabled();
		}
		await userEvent.click(canvas.getAllByRole('checkbox')[0]!, { pointerEventsCheck: 0 });
		await expect(args.onownedchange).not.toHaveBeenCalled();
	}}
/>

<!-- A household of one: no levels to set, and the screen is still the place
     that says the accounts are theirs. -->
<Story name="A household of one" args={{ members: [] }} />

<!-- canvas.md: returning from the bank moves focus to the chooser's heading.
     A member arrives here by a full navigation, and without this a screen
     reader lands at the top of a page it has already announced. -->
<Story
	name="Focus lands on the heading"
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('heading', { level: 1 })).toHaveFocus();
	}}
/>

<!-- J09.A: ticking a co-owner box reports only that member, never a
     replacement list — the fix for the defect the single "Mine" checkbox had. -->
<Story
	name="Making an owner"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getAllByRole('checkbox', { name: 'Grace' })[0]!);
		await expect(args.onownerchange).toHaveBeenCalledWith('a1', 'grace', true);
	}}
/>

<Story
	name="Handed on"
	args={{ owners: { a1: { grace: true } } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getAllByRole('checkbox', { name: 'Grace' })[0]).toBeChecked();
		await expect(
			canvas.queryByRole('radiogroup', { name: 'What Grace sees of Conta à Ordem' })
		).toBeNull();
	}}
/>

<!-- J09.A / 03: the last owner cannot step back. The database refuses it; the
     screen shows the refusal inline, beside the row, naming leaving the
     account out as the way to the same end. -->
<Story
	name="The last owner cannot step back"
	args={{ refusedAccountId: 'a1' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByText('An account has to belong to somebody')
		).toBeInTheDocument();
		await expect(
			canvas.getByText(/this would leave it with nobody/)
		).toBeInTheDocument();
	}}
/>

<!-- J09.A / 04: leaving an account out is offered as a confirmation, naming
     the account. -->
<Story
	name="Leaving an account out"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getAllByRole('button', { name: /Leave out|Bring back/ })[0]!);
		await expect(
			canvas.getByRole('heading', { name: 'Leave Conta à Ordem out of wimm?' })
		).toBeInTheDocument();

		await userEvent.click(canvas.getByRole('button', { name: 'Leave it out' }));
		await expect(args.onleaveout).toHaveBeenCalledWith('a1');
	}}
/>

<!-- J09.A / 05, 06: bringing an account back is required rather than offered,
     and names every member who will see it again and at what level — they
     were given that before it was left out and nobody is asking them again. -->
<Story
	name="Bringing it back"
	args={{
		accounts: [{ ...current, leftOut: true }, savings, personal],
		levels: { a1: { grace: 'balance', alan: 'details' } }
	}}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Bring back' }));
		await expect(
			canvas.getByRole('heading', { name: 'Bring Conta à Ordem back?' })
		).toBeInTheDocument();
		await expect(
			canvas.getByText(
				'Grace will see its balance again and Alan will see its balance and its details, which is what each of them had before you left it out. wimm starts reading it from now.'
			)
		).toBeInTheDocument();

		await userEvent.click(canvas.getByRole('button', { name: 'Bring it back' }));
		await expect(args.onbringback).toHaveBeenCalledWith('a1');
	}}
/>

<!-- J10.A / 01: naming an account, inline and ephemeral, on the row it
     belongs to. -->
<Story
	name="Naming an account"
	args={{ onrename: fn() }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getAllByRole('button', { name: 'Rename' })[0]!);

		const field = canvas.getByRole('textbox');
		await userEvent.type(field, 'Rent');
		await userEvent.click(canvas.getByRole('button', { name: 'Save' }));

		await expect(args.onrename).toHaveBeenCalledWith('a1', 'Rent');
	}}
/>
