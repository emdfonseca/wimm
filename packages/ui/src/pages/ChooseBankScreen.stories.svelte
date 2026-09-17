<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import ChooseBankScreen from './ChooseBankScreen.svelte';

	const banks = [
		{ id: 'PT:Montepio', name: 'Caixa Económica Montepio Geral' },
		{ id: 'PT:ActivoBank', name: 'ActivoBank' },
		{ id: 'PT:Revolut', name: 'Revolut' },
		{ id: 'PT:CGD', name: 'Caixa Geral de Depósitos' }
	];

	const { Story } = defineMeta({
		title: 'Pages/ChooseBankScreen',
		component: ChooseBankScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		argTypes: { state: { control: 'inline-radio', options: ['default', 'loading', 'unavailable'] } },
		args: { banks, state: 'default', onselect: fn(), onretry: fn() }
	});
</script>

<Story
	name="Default"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getAllByRole('button')).toHaveLength(4);
		await userEvent.click(canvas.getByRole('button', { name: 'Revolut' }));
		await expect(args.onselect).toHaveBeenCalledWith('PT:Revolut');
	}}
/>

<!-- Typing narrows the list, and the count is announced: a list changing under
     a member's fingers is a change they cannot see if they are not looking. -->
<Story
	name="Narrowed"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.type(canvas.getByRole('searchbox'), 'caixa');
		await expect(canvas.getAllByRole('button')).toHaveLength(2);
		await expect(canvas.getByRole('status')).toHaveTextContent('2 banks match');
	}}
/>

<Story
	name="One match is singular"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.type(canvas.getByRole('searchbox'), 'revolut');
		await expect(canvas.getByRole('status')).toHaveTextContent('1 bank match');
	}}
/>

<!-- A bank wimm cannot reach is said plainly, rather than leaving a member
     searching for a spelling that would work. -->
<Story
	name="No match"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.type(canvas.getByRole('searchbox'), 'zzz');
		await expect(canvas.getByRole('status')).toHaveTextContent('No banks match');
		await expect(canvas.getByText(/cannot be connected yet/)).toBeInTheDocument();
	}}
/>

<Story name="Loading" args={{ state: 'loading', banks: [] }} />

<Story
	name="List unavailable"
	args={{ state: 'unavailable', banks: [] }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/Nothing has been connected/)).toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Try again' }));
		await expect(args.onretry).toHaveBeenCalledOnce();
	}}
/>

<!-- The page says who the member confirms with, before they choose. -->
<Story
	name="It says wimm never sees a password"
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/wimm never sees your banking password/)
		).toBeInTheDocument();
	}}
/>
