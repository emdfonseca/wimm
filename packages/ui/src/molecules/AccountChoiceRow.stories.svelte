<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import AccountChoiceRow, { type OtherMember } from './AccountChoiceRow.svelte';

	const household: OtherMember[] = [
		{ id: 'grace', name: 'Grace', isOwner: false, level: 'hidden' },
		{ id: 'alan', name: 'Alan', isOwner: false, level: 'hidden' }
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
			others: household,
			onownedchange: fn(),
			onownerchange: fn(),
			onlevelchange: fn(),
			onleaveout: fn()
		}
	});
</script>

<!-- As the chooser opens: the connecting member owns it, nobody else sees it. -->
<Story name="As it opens" />

<Story name="Disowned" args={{ owned: false }} />

<Story
	name="One member at balance"
	args={{ others: [household[0]!, { ...household[1]!, level: 'balance' }] }}
/>

<Story
	name="Two members at different levels"
	args={{
		others: [
			{ ...household[0]!, level: 'details' },
			{ ...household[1]!, level: 'balance' }
		]
	}}
/>

<!-- A co-owner does not also get a level control (ADR 0019): owning already
     means seeing everything. -->
<Story
	name="A co-owner"
	args={{ others: [{ ...household[0]!, isOwner: true }, household[1]!] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('checkbox', { name: 'Grace' })).toBeChecked();
		await expect(
			canvas.queryByRole('radiogroup', { name: 'What Grace sees of Conta à Ordem' })
		).toBeNull();
	}}
/>

<Story name="Newly offered on restore" args={{ newlyOffered: true }} />

<!-- An account with no other members is still a row: sole households exist,
     and the owners and levels sections simply have nothing to show. -->
<Story name="No one else in the household" args={{ others: [] }} />

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
		await expect(canvas.getByRole('checkbox', { name: 'Conta à Ordem is mine' })).toBeChecked();
		// One radiogroup per member who is not an owner, each with three options.
		await expect(canvas.getAllByRole('radiogroup')).toHaveLength(2);
		await expect(canvas.getAllByRole('radio')).toHaveLength(6);
	}}
/>

<Story
	name="Disowning"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('checkbox', { name: 'Conta à Ordem is mine' }));
		await expect(args.onownedchange).toHaveBeenCalledWith(false);
	}}
/>

<!-- The fix for the defect: ticking a co-owner box reports only that member,
     never a replacement list that would drop anyone else. -->
<Story
	name="Adding an owner reports only that member"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('checkbox', { name: 'Grace' }));
		await expect(args.onownerchange).toHaveBeenCalledWith('grace', true);
		await expect(args.onownedchange).not.toHaveBeenCalled();
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

<Story
	name="Leaving out"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Leave out' }));
		await expect(args.onleaveout).toHaveBeenCalled();
	}}
/>

<Story name="Left out, offering to bring back" args={{ leftOut: true }} />

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
		const checkbox = canvas.getByRole('checkbox', { name: 'Conta à Ordem is mine' });
		await expect(checkbox).toBeDisabled();
		await userEvent.click(checkbox, { pointerEventsCheck: 0 });
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
				others={[household[0]!, { ...household[1]!, level: 'balance' }]}
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
