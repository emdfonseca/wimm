<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import SeekPager from './SeekPager.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/SeekPager',
		component: SeekPager,
		tags: ['autodocs'],
		args: {
			span: '17 September to 15 September 2026',
			olderHref: '/transactions?before=abc'
		}
	});
</script>

<!-- The newest page: only Older is offered, because there is nothing newer to
     go back to. -->
<Story
	name="TheNewestPage"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('link', { name: 'Older' })).toBeInTheDocument();
		await expect(canvas.queryByRole('link', { name: 'Newer' })).not.toBeInTheDocument();
	}}
/>

<Story
	name="AMiddlePage"
	args={{ span: '4 August to 31 July 2026', newerHref: '/transactions?after=def' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('link', { name: 'Newer' })).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Older' })).toBeInTheDocument();
	}}
/>

<!-- Nothing older is a sentence, not a greyed control. -->
<Story
	name="TheOldestPage"
	args={{
		span: '4 June 2026',
		olderHref: undefined,
		newerHref: '/transactions?after=def',
		nothingOlder: 'Nothing older. This is as far back as the bank would go.'
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByRole('link', { name: 'Older' })).not.toBeInTheDocument();
		await expect(canvas.getByText(/Nothing older/)).toBeInTheDocument();
	}}
/>

<!-- Where the member is, is a date. There is no page number anywhere. -->
<Story
	name="NoPageNumbers"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('17 September to 15 September 2026')).toBeInTheDocument();
		await expect(canvas.queryByText(/\bof\b/)).not.toBeInTheDocument();
		await expect(canvas.queryByText(/\d+\s*\/\s*\d+/)).not.toBeInTheDocument();
	}}
/>

<Story name="Compact" globals={{ viewport: { value: 'compact' } }} />
