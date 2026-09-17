<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import AccountChoiceRow from './AccountChoiceRow.svelte';

	const household = [
		{ id: 'grace', name: 'Grace' },
		{ id: 'alan', name: 'Alan' }
	];

	const { Story } = defineMeta({
		title: 'Molecules/AccountChoiceRow',
		component: AccountChoiceRow,
		tags: ['autodocs'],
		argTypes: { disabled: { control: 'boolean' }, newlyOffered: { control: 'boolean' } },
		args: {
			name: 'Conta à Ordem',
			meta: 'Montepio · •••• 0538',
			balance: '€4,200.10',
			owned: true,
			members: household,
			levels: {},
			onownedchange: fn(),
			onlevelchange: fn()
		}
	});
</script>

<!-- As the chooser opens: the connecting member owns it, nobody else sees it. -->
<Story name="As it opens" />

<Story name="Disowned" args={{ owned: false }} />

<Story name="One member at balance" args={{ levels: { grace: 'balance' } }} />

<Story
	name="Two members at different levels"
	args={{ levels: { grace: 'details', alan: 'balance' } }}
/>

<Story name="Newly offered on restore" args={{ newlyOffered: true }} />

<!-- An account with no other members is still a row: sole households exist,
     and the levels section simply has nothing to show. -->
<Story name="No one else in the household" args={{ members: [] }} />

<!-- The balance is present by design: choosing who sees an account by its name
     alone, with no figure, is choosing blind. -->
<Story
	name="It carries a balance"
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('€4,200.10')).toBeInTheDocument();
	}}
/>

<!-- Ownership is a checkbox and a level is one of three: the row is not a
     two-state control. -->
<Story
	name="Ownership and levels are different controls"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('checkbox')).toBeChecked();
		// One radiogroup per other member, each with three options.
		await expect(canvas.getAllByRole('radiogroup')).toHaveLength(2);
		await expect(canvas.getAllByRole('radio')).toHaveLength(6);
	}}
/>

<Story
	name="Disowning"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('checkbox'));
		await expect(args.onownedchange).toHaveBeenCalledWith(false);
	}}
/>

<Story
	name="Granting a level"
	play={async ({ canvasElement, args }) => {
		const group = within(canvasElement).getByRole('radiogroup', {
			name: 'What Grace sees of Conta à Ordem'
		});
		await userEvent.click(within(group).getByRole('radio', { name: 'Balance' }));
		await expect(args.onlevelchange).toHaveBeenCalledWith('grace', 'balance');
	}}
/>

<!-- Each control names the member and the account, so a screen reader hearing
     two of them in a list can tell which is which. -->
<Story
	name="Each level control names its member and account"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByRole('radiogroup', { name: 'What Grace sees of Conta à Ordem' })
		).toBeInTheDocument();
		await expect(
			canvas.getByRole('radiogroup', { name: 'What Alan sees of Conta à Ordem' })
		).toBeInTheDocument();
	}}
/>

<Story
	name="Disabled"
	args={{ disabled: true }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('checkbox')).toBeDisabled();
		await userEvent.click(canvas.getByRole('checkbox'), { pointerEventsCheck: 0 });
		await expect(args.onownedchange).not.toHaveBeenCalled();
	}}
/>

<Story name="A chooser of three" tags={['!test']}>
	{#snippet template(args)}
		<div style="display: flex; flex-direction: column; gap: 8px; inline-size: 620px;">
			<AccountChoiceRow {...args} />
			<AccountChoiceRow
				{...args}
				name="Poupança"
				meta="Montepio · •••• 7712"
				balance="€11,930.00"
				levels={{ grace: 'balance' }}
			/>
			<AccountChoiceRow
				{...args}
				name="Conta Pessoal"
				meta="Montepio · •••• 5594"
				balance="€820.44"
				owned={false}
			/>
		</div>
	{/snippet}
</Story>
