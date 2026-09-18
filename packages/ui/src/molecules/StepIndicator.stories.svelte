<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import StepIndicator from './StepIndicator.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/StepIndicator',
		component: StepIndicator,
		tags: ['autodocs'],
		args: { current: 1, total: 2 }
	});
</script>

<Story
	name="Step 1 of 2"
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('Step 1 of 2')).toBeInTheDocument();
	}}
/>

<Story name="Step 2 of 2" args={{ current: 2, total: 2 }} />
