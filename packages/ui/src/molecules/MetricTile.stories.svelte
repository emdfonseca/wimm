<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import MetricTile from './MetricTile.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/MetricTile',
		component: MetricTile,
		tags: ['autodocs'],
		args: {
			label: 'Household money',
			value: '€6,698.00'
		}
	});
</script>

<!-- A 82-high row: the tile fills it. -->
<Story
	name="NoDelta"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('€6,698.00')).toBeInTheDocument();
		await expect(canvas.queryByText(/^than /)).not.toBeInTheDocument();
		await expect(canvasElement.querySelector('.tile')!.getBoundingClientRect().height).toBe(82);
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 300px; block-size: 82px;"><MetricTile {...args} /></div>
	{/snippet}
</Story>

<!-- Both directions read in words and in the neutral text colour: green and
     red are for money direction only. -->
<Story
	name="MoreThanLastMonth"
	args={{
		label: 'Money out',
		value: '€1,812.64',
		delta: { change: '€182.40 more', period: 'than 1 to 20 Aug', direction: 'up' }
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('€182.40 more')).toBeInTheDocument();
		await expect(canvas.getByText('than 1 to 20 Aug')).toBeInTheDocument();
		await expect(canvasElement.querySelector('.tile')!.getBoundingClientRect().height).toBe(102);
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 300px; block-size: 102px;"><MetricTile {...args} /></div>
	{/snippet}
</Story>

<Story
	name="LessThanLastMonth"
	args={{
		label: 'Net',
		value: '+€637.36',
		delta: { change: '€32.40 less', period: 'than 1 to 20 Aug', direction: 'down' }
	}}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('€32.40 less')).toBeInTheDocument();
	}}
/>

<Story
	name="AsALink"
	args={{
		label: 'Money in',
		value: '€2,450.00',
		href: '/transactions',
		delta: { change: '€150.00 more', period: 'than 1 to 20 Aug', direction: 'up' }
	}}
	play={async ({ canvasElement }) => {
		const link = within(canvasElement).getByRole('link');
		await expect(link).toHaveAttribute('href', '/transactions');
		await expect(link).toHaveTextContent('Money in');
		await expect(link).toHaveTextContent('€2,450.00');
	}}
/>
