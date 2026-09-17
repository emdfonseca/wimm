<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import BottomNav from './BottomNav.svelte';
	import { destinations } from '../destinations.js';

	const { Story } = defineMeta({
		title: 'Organisms/BottomNav',
		component: BottomNav,
		tags: ['autodocs'],
		globals: { viewport: { value: 'compact' } },
		args: { destinations: destinations('/') }
	});
</script>

<Story
	name="OverviewCurrent"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('link', { name: 'Overview' })).toHaveAttribute(
			'aria-current',
			'page'
		);
		await expect(canvas.getByRole('link', { name: 'Transactions' })).not.toHaveAttribute(
			'aria-current'
		);
	}}
/>

<Story
	name="TransactionsCurrent"
	args={{ destinations: destinations('/transactions') }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('link', { name: 'Transactions' })).toHaveAttribute(
			'aria-current',
			'page'
		);
	}}
/>

<!-- Each target is the whole tab. At 390 that is 195 x 60, and the floor this
     asserts is the 44 px one every pointer regime has to clear. -->
<Story
	name="TargetsClearTheFloor"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const label of ['Overview', 'Transactions']) {
			const tab = canvas.getByRole('link', { name: label });
			const box = tab.getBoundingClientRect();
			await expect(box.height).toBeGreaterThanOrEqual(44);
			await expect(box.width).toBeGreaterThanOrEqual(44);
		}
	}}
/>

<!-- A destination with no address is not rendered: the canvas disables it, and
     a disabled node in pen is absent rather than greyed. -->
<Story
	name="ADestinationThatDoesNotExistYet"
	args={{
		destinations: [
			{ label: 'Overview', href: '/', icon: 'layout-dashboard' as const, current: true },
			{ label: 'Budgets', icon: 'arrow-left-right' as const }
		]
	}}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).queryByText('Budgets')).not.toBeInTheDocument();
	}}
/>
