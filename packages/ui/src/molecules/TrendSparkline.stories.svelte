<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import TrendSparkline from './TrendSparkline.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/TrendSparkline',
		component: TrendSparkline,
		tags: ['autodocs'],
		args: {
			caption: '90 days to 19 Sep',
			values: [
				3200, 3600, 3400, 4100, 3900, 4600, 4300, 5000, 4800, 5600, 5200, 6100
			]
		}
	});
</script>

<Story
	name="Populated"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Trend')).toBeInTheDocument();
		await expect(canvas.getByText('90 days to 19 Sep')).toBeInTheDocument();
	}}
/>

<Story
	name="EmptySeries"
	args={{ values: [] }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('90 days to 19 Sep')).toBeInTheDocument();
	}}
/>
