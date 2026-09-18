<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import SeekPager from './SeekPager.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/SeekPager',
		component: SeekPager,
		tags: ['autodocs'],
		args: {
			span: '17 September to 15 September 2026'
		}
	});
</script>

<!-- Where the member is, is a date. There is no page number anywhere, and no
     Older/Newer here — both moved to the scrubber. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('17 September to 15 September 2026')).toBeInTheDocument();
		await expect(canvas.queryByText(/\bof\b/)).not.toBeInTheDocument();
		await expect(canvas.queryByText(/\d+\s*\/\s*\d+/)).not.toBeInTheDocument();
		await expect(canvas.queryByRole('link')).not.toBeInTheDocument();
	}}
/>

<!-- The oldest page states the reach in words. -->
<Story
	name="TheOldestPage"
	args={{
		span: '4 June 2026',
		nothingOlder: 'Nothing older. This is as far back as the bank would go.'
	}}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/Nothing older/)
		).toBeInTheDocument();
	}}
/>

<Story name="Compact" globals={{ viewport: { value: 'compact' } }} />
