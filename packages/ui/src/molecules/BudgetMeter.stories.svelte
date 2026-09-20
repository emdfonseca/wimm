<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import BudgetMeter from './BudgetMeter.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/BudgetMeter',
		component: BudgetMeter,
		tags: ['autodocs'],
		parameters: { layout: 'padded' },
		args: { label: 'Pingo Doce', value: '€412.60 · 9 payments', proportion: 1 }
	});
</script>

<!-- The largest merchant always fills the track: proportion is against it. -->
<Story
	name="Full"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Pingo Doce')).toBeInTheDocument();
		await expect(canvas.getByText('€412.60 · 9 payments')).toBeInTheDocument();
	}}
/>

<Story
	name="Partly"
	args={{ label: 'Galp', value: '€186.40 · 3 payments', proportion: 0.45 }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Galp')).toBeInTheDocument();
		await expect(canvas.getByText('€186.40 · 3 payments')).toBeInTheDocument();
	}}
/>

<Story
	name="Sliver"
	args={{ label: 'Netflix', value: '€12.99 · 1 payment', proportion: 0.04 }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Netflix')).toBeInTheDocument();
		await expect(canvas.getByText('€12.99 · 1 payment')).toBeInTheDocument();
	}}
/>
