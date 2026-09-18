<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import LeftOutDialog from './LeftOutDialog.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/LeftOutDialog',
		component: LeftOutDialog,
		tags: ['autodocs'],
		args: {
			open: true,
			mode: 'leave-out',
			accountName: 'Conta Pessoal',
			onconfirm: fn(),
			oncancel: fn()
		}
	});
</script>

<Story
	name="Leave out"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByRole('heading', { name: 'Leave Conta Pessoal out of wimm?' })
		).toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Leave it out' }));
		await expect(args.onconfirm).toHaveBeenCalledOnce();
	}}
/>

<Story
	name="Keeping it, from the dialog"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Keep it' }));
		await expect(args.oncancel).toHaveBeenCalledOnce();
		await expect(args.onconfirm).not.toHaveBeenCalled();
	}}
/>

<Story
	name="Bring back, one grantee"
	args={{ mode: 'bring-back', grantees: [{ name: 'Grace', level: 'balance' }] }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(
				'Grace will see its balance again, which is what they had before you left it out. wimm starts reading it from now.'
			)
		).toBeInTheDocument();
	}}
/>

<Story
	name="Bring back, two grantees at different levels"
	args={{
		mode: 'bring-back',
		grantees: [
			{ name: 'Grace', level: 'balance' },
			{ name: 'Alan', level: 'details' }
		]
	}}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByText(
				'Grace will see its balance again and Alan will see its balance and its details, which is what each of them had before you left it out. wimm starts reading it from now.'
			)
		).toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Bring it back' }));
		await expect(args.onconfirm).toHaveBeenCalledOnce();
	}}
/>

<Story name="Bring back, nobody was granted anything" args={{ mode: 'bring-back', grantees: [] }} />

<Story
	name="Escape cancels"
	play={async ({ args }) => {
		await userEvent.keyboard('{Escape}');
		await expect(args.oncancel).toHaveBeenCalledOnce();
	}}
/>
