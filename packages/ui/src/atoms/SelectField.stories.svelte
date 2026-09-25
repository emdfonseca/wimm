<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import SelectField from './SelectField.svelte';

	const months = [
		{ value: '', label: 'All months' },
		{ value: '2026-09', label: 'September 2026' },
		{ value: '2026-08', label: 'August 2026' },
		{ value: '2026-07', label: 'July 2026' }
	];

	const { Story } = defineMeta({
		title: 'Atoms/SelectField',
		component: SelectField,
		tags: ['autodocs'],
		args: { label: 'Month', name: 'month', options: months, onchange: fn() }
	});
</script>

<!-- The label names the select, and each choice is reported once. -->
<Story
	name="Default"
	play={async ({ canvasElement, args }) => {
		const select = within(canvasElement).getByRole('combobox', { name: 'Month' });
		await expect(select).toHaveValue('');

		await userEvent.selectOptions(select, '2026-08');
		await expect(args.onchange).toHaveBeenCalledTimes(1);
		await expect(args.onchange).toHaveBeenCalledWith('2026-08');

		await userEvent.selectOptions(select, '2026-07');
		await expect(args.onchange).toHaveBeenCalledTimes(2);
		await expect(args.onchange).toHaveBeenLastCalledWith('2026-07');
	}}
/>

<Story
	name="Chosen"
	args={{ value: '2026-08' }}
	play={async ({ canvasElement }) => {
		const select = within(canvasElement).getByRole('combobox', { name: 'Month' });
		await expect(select).toHaveValue('2026-08');
		await expect(within(select).getByRole('option', { name: 'August 2026' })).toHaveProperty(
			'selected',
			true
		);
	}}
/>

<Story
	name="LongOptions"
	args={{
		label: 'Account',
		name: 'account',
		value: 'a2',
		options: [
			{ value: '', label: 'All accounts' },
			{ value: 'a1', label: 'Current account · Monzo' },
			{ value: 'a2', label: 'Conta Poupança Habitação Jovem · Caixa Geral de Depósitos' }
		]
	}}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('combobox', { name: 'Account' })).toHaveValue(
			'a2'
		);
	}}
/>
