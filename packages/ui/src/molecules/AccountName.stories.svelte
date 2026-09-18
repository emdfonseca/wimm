<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import AccountName from './AccountName.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/AccountName',
		component: AccountName,
		tags: ['autodocs'],
		args: {
			householdName: '',
			bankName: 'Monzo',
			bankAccountName: 'Current Account',
			onSave: fn()
		}
	});
</script>

<Story name="Unnamed" />

<Story name="Named" args={{ householdName: 'Joint account' }} />

<Story
	name="Clearing"
	args={{ householdName: 'Joint account' }}
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Clear' }));
		await expect(args.onSave).toHaveBeenCalledWith('');
		await expect(within(canvasElement).getByRole('textbox')).toHaveValue('');
	}}
/>

<Story
	name="Saving"
	play={async ({ canvasElement, args }) => {
		const field = within(canvasElement).getByRole('textbox');
		await userEvent.type(field, 'Rent');
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Save' }));
		await expect(args.onSave).toHaveBeenCalledWith('Rent');
	}}
/>
