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
		meta: 'Montepio · •••• 0538',
		balance: '€4,200.10',
		owned: true
	};
	const savings = {
		id: 'a2',
		name: 'Poupança',
		meta: 'Montepio · •••• 7712',
		balance: '€11,930.00',
		owned: true
	};
	const personal = {
		id: 'a3',
		name: 'Conta Pessoal',
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
			onlevelchange: fn(),
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
		for (const box of canvas.getAllByRole('checkbox')) {
			await expect(box).toBeChecked();
		}
		await expect(canvas.getByText(/nobody sees anything until you say so/)).toBeInTheDocument();
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
