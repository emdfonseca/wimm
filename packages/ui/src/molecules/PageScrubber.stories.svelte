<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, waitFor, within } from 'storybook/test';
	import PageScrubber from './PageScrubber.svelte';

	const pages = [
		{ key: '2026-09-15', label: '17 Sep to 4 Sep, 2026', href: '/transactions?page=2026-09-15.t1' },
		{ key: '2026-08-20', label: '3 Sep to 20 Aug, 2026', href: '/transactions?page=2026-08-20.t2' },
		{ key: '2026-07-11', label: '19 Aug to 11 Jul, 2026', href: '/transactions?page=2026-07-11.t3' },
		{ key: '2026-05-30', label: '10 Jul to 30 May, 2026', href: '/transactions?page=2026-05-30.t4' },
		{ key: '2026-04-02', label: '29 May to 2 Apr, 2026', href: '/transactions?page=2026-04-02.t5' },
		{ key: '2026-02-14', label: '1 Apr to 14 Feb, 2026', href: '/transactions?page=2026-02-14.t6' }
	];

	const { Story } = defineMeta({
		title: 'Molecules/PageScrubber',
		component: PageScrubber,
		tags: ['autodocs'],
		args: { pages, onSelectPage: fn(), onNewest: fn(), onOldest: fn() }
	});
</script>

<Story
	name="A middle page"
	args={{ current: '2026-07-11' }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByRole('link', { name: '19 Aug to 11 Jul, 2026' })
		).toHaveAttribute('aria-current', 'true');
	}}
/>

<Story name="The newest page" args={{ current: '2026-09-15' }} />

<Story name="The oldest page" args={{ current: '2026-02-14' }} />

<!-- Present, disabled — not hidden — at the end of the track: an arrow
     removed there shifts what sits beside it, which reads as the layout
     moving rather than as an edge reached. -->
<Story
	name="At the very ends there is nowhere further to jump"
	args={{
		current: '2026-09-15',
		onNewest: undefined,
		onOldest: undefined
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('button', { name: 'Newest' })).toBeDisabled();
		await expect(canvas.getByRole('button', { name: 'Oldest' })).toBeDisabled();
	}}
/>

<Story
	name="One page only"
	args={{
		pages: [pages[0]!],
		current: '2026-09-15'
	}}
/>

<!-- The span beside the scrubber names the current page, so the scrubber
     itself shows no label until a page is pointed at or focused. -->
<Story
	name="No label at rest"
	args={{ current: '2026-07-11' }}
	play={async ({ canvasElement }) => {
		await expect(canvasElement.querySelectorAll('.flag')).toHaveLength(0);
		await expect(canvasElement.querySelector('.current')).toBeNull();
	}}
/>

<Story
	name="Label under the pointer"
	args={{ current: '2026-07-11' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		canvas.getByRole('link', { name: '3 Sep to 20 Aug, 2026' }).focus();
		await waitFor(() => expect(canvasElement.querySelectorAll('.flag')).toHaveLength(1));
		await expect(canvasElement.querySelector('.flag')).toHaveTextContent('3 Sep to 20 Aug, 2026');
	}}
/>
